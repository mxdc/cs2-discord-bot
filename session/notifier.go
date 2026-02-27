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
	cfg    *config.AppConfig
	client *leetify.LeetifyClient
	in     <-chan MatchDetected
}

func NewMatchNotifier(
	cfg *config.AppConfig,
	client *leetify.LeetifyClient,
	in <-chan MatchDetected,
) *MatchNotifier {
	return &MatchNotifier{
		cfg:    cfg,
		client: client,
		in:     in,
	}
}

func (mm *MatchNotifier) HandleMatch() {
	log.Println("Notifier: Started notifier, waiting for matches...")
	seen := make(map[string]bool)
	discordClient := discord.NewWebhookClient(mm.cfg.DiscordHook)
	steamClient := steam.NewSteamClient(mm.cfg.SteamAPIKey)

	for msg := range mm.in {
		if seen[msg.Match.GameId] {
			continue
		}

		seen[msg.Match.GameId] = true
		log.Println("Manager: New match detected:", msg.Match.GameId)

		// Get all Steam IDs from both teams
		allSteamIDs := append(msg.Match.OwnTeamSteam64Ids, msg.Match.EnemyTeamSteam64Ids...)

		// Get Steam player data (names and countries)
		steamPlayers, err := steamClient.GetSteamPlayers(allSteamIDs)
		if err != nil {
			// Continue without steam data
			log.Printf("Manager: Warning: failed to get steam players: %v", err)
		}

		time.Sleep(5 * time.Second)
		matchDetails, err := mm.client.GetMatchDetails(msg.Match.GameId)
		if err != nil {
			// Continue without match details
			log.Printf("Manager: Warning: failed to get match details: %v", err)
		}
		matchWithDetails := parser.ParseMatchResultWithDetails(msg.Match, matchDetails, steamPlayers, mm.cfg.Players)

		// Send Discord webhook
		discordClient.SendMatchResult(matchWithDetails)
	}
}

type SessionNotifier struct {
	client *leetify.LeetifyClient
	cfg    *config.AppConfig
	in     <-chan GameSession
}

func NewSessionNotifier(
	client *leetify.LeetifyClient,
	cfg *config.AppConfig,
	in <-chan GameSession,
) *SessionNotifier {
	return &SessionNotifier{
		client: client,
		cfg:    cfg,
		in:     in,
	}
}

func (sn *SessionNotifier) HandleSession() {
	log.Println("SessionNotifier: Started sessionNotifier, waiting for completed sessions...")

	discordClient := discord.NewWebhookClient(sn.cfg.DiscordHook)
	steamClient := steam.NewSteamClient(sn.cfg.SteamAPIKey)

	for completedSession := range sn.in {
		log.Printf("SessionNotifier: New session received with %d matches", len(completedSession.Matches))

		allSteamIDs := completedSession.GetSteamIDs()
		// Get Steam player data (names and countries)
		steamPlayers, err := steamClient.GetSteamPlayers(allSteamIDs)
		if err != nil {
			// Continue without steam data
			log.Printf("SessionNotifier: Warning: failed to get steam players: %v", err)
		}

		sessionWithDetails := parser.SessionWithDetails{TrackedPlayers: sn.cfg.Players}

		for i, game := range completedSession.Matches {
			matchDetails, err := sn.client.GetMatchDetails(game.GameId)
			if err != nil {
				// Continue without match details
				log.Printf("SessionNotifier: Warning: failed to get match details: %v", err)
			}

			matchWithDetails := parser.ParseMatchResultWithDetails(game, matchDetails, steamPlayers, sn.cfg.Players)
			sessionWithDetails.Matches = append(sessionWithDetails.Matches, matchWithDetails)

			// Avoid rate limit failure
			if i < len(completedSession.Matches)-2 {
				time.Sleep(10 * time.Second)
			}
		}

		// Send Discord webhook
		discordClient.SendSessionResult(sessionWithDetails)
	}
}
