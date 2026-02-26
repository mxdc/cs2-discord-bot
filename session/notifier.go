package session

import (
	"log"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/discord"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/parser"
	"github.com/mxdc/cs2-discord-bot/steam"
)

type MatchDetected struct {
	Match  leetify.LeetifyGameResponse
	Player config.Player
}

type MatchNotifier struct {
	cfg           *config.AppConfig
	client        *leetify.LeetifyClient
	discordClient *discord.WebhookClient
	steamClient   *steam.Client
	in            <-chan MatchDetected
}

func NewMatchNotifier(
	cfg *config.AppConfig,
	client *leetify.LeetifyClient,
	in <-chan MatchDetected,
) *MatchNotifier {
	return &MatchNotifier{
		cfg:           cfg,
		client:        client,
		discordClient: discord.NewWebhookClient(cfg.DiscordHook),
		steamClient:   steam.NewSteamClient(cfg.SteamAPIKey),
		in:            in,
	}
}

func (mm *MatchNotifier) HandleMatch() {
	log.Println("Notifier: Started notifier, waiting for matches...")
	seen := make(map[string]bool)

	for msg := range mm.in {
		if seen[msg.Match.GameId] {
			continue
		}

		seen[msg.Match.GameId] = true
		match := parser.ParseGameResponseFromLeetify(msg.Match)
		log.Println("Manager: New match detected:", match.GameID)

		time.Sleep(5 * time.Second)

		// Get all Steam IDs from both teams
		allSteamIDs := append(match.OwnTeamSteam64Ids, match.EnemyTeamSteam64Ids...)

		// Get Steam player data (names and countries)
		steamPlayers, err := mm.steamClient.GetSteamPlayers(allSteamIDs)
		if err != nil {
			// Continue without steam data
			log.Printf("Manager: Warning: failed to get steam players: %v", err)
		}

		details, err := mm.client.GetMatchDetails(match.GameID)
		if err != nil {
			// Continue without match details
			log.Printf("Manager: Warning: failed to get match details: %v", err)
		}

		matchWithDetails := parser.ParseMatchResultWithDetails(match, details, steamPlayers, mm.cfg.Players)

		// Send Discord webhook
		mm.discordClient.SendMatchResult(matchWithDetails)
	}
}

type SessionNotifier struct {
	cfg *config.AppConfig
	in  <-chan GameSession
}

func NewSessionNotifier(
	cfg *config.AppConfig,
	in <-chan GameSession,
) *SessionNotifier {
	return &SessionNotifier{
		cfg: cfg,
		in:  in,
	}
}

func (sn *SessionNotifier) HandleSession() {
	log.Println("SessionNotifier: Started sessionNotifier, waiting for completed sessions...")

	for completedSession := range sn.in {
		log.Printf("SessionNotifier: New session received with %d matches", len(completedSession.Matches))
		// TODO: Implement session notification logic here
		// This is where you would process the completed session
		// (e.g., create Discord embed, send webhook, etc.)
	}
}
