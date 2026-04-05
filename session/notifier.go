package session

import (
	"log"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/discord"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/mistral"
	"github.com/mxdc/cs2-discord-bot/parser"
	"github.com/mxdc/cs2-discord-bot/steam"
)

type MatchDetected struct {
	Match      leetify.LeetifyGameResponse
	Player     config.Player
	DetectedAt time.Time
}

func (md *MatchDetected) IsTooOld() bool {
	matchEndTime, _ := time.Parse(time.RFC3339, md.Match.GameFinishedAt)
	return time.Since(matchEndTime) > 24*time.Hour
}

type MatchNotifier struct {
	cfg           *config.AppConfig
	client        *leetify.LeetifyClient
	mistralClient *mistral.MistralClient
	translations  locales.Translations
	in            <-chan MatchDetected
	withRank      bool
}

func NewMatchNotifier(
	cfg *config.AppConfig,
	client *leetify.LeetifyClient,
	mistralClient *mistral.MistralClient,
	translations locales.Translations,
	in <-chan MatchDetected,
	withRank bool,
) *MatchNotifier {
	return &MatchNotifier{
		cfg:           cfg,
		client:        client,
		mistralClient: mistralClient,
		translations:  translations,
		in:            in,
		withRank:      withRank,
	}
}

func (mm *MatchNotifier) HandleMatch() {
	log.Println("Notifier: Started notifier, waiting for matches...")
	seenGames := &SeenGames{games: []SeenGame{}}
	discordClient := discord.NewWebhookClient(mm.cfg.DiscordHook, mm.mistralClient, mm.translations, mm.withRank)
	steamClient := steam.NewSteamClient(mm.cfg.SteamAPIKey)

	for msg := range mm.in {
		if seenGames.AlreadyNotified(msg.Match.GameId) {
			continue
		}

		seenGames.AddGame(msg.Player.SteamID, msg.Match.GameId, msg.Match.GameFinishedAt)
		log.Println("Manager: New match detected:", msg.Match.GameId)

		// Get all Steam IDs from both teams
		allSteamIDs := append(msg.Match.OwnTeamSteam64Ids, msg.Match.EnemyTeamSteam64Ids...)

		// Get Steam player data (names and countries)
		steamPlayers, err := steamClient.GetSteamPlayers(allSteamIDs)
		if err != nil {
			// Continue without steam data
			log.Printf("Manager: Warning: failed to get steam players: %v", err)
		}

		time.Sleep(5 * time.Minute)
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
	client        *leetify.LeetifyClient
	cfg           *config.AppConfig
	mistralClient *mistral.MistralClient
	translations  locales.Translations
	in            <-chan GameSession
	withRank      bool
}

func NewSessionNotifier(
	cfg *config.AppConfig,
	leetifyClient *leetify.LeetifyClient,
	mistralClient *mistral.MistralClient,
	translations locales.Translations,
	in <-chan GameSession,
	withRank bool,
) *SessionNotifier {
	return &SessionNotifier{
		cfg:           cfg,
		client:        leetifyClient,
		mistralClient: mistralClient,
		translations:  translations,
		in:            in,
		withRank:      withRank,
	}
}

func (sn *SessionNotifier) HandleSession() {
	log.Println("SessionNotifier: Started sessionNotifier, waiting for completed sessions...")

	discordClient := discord.NewWebhookClient(sn.cfg.DiscordHook, sn.mistralClient, sn.translations, sn.withRank)
	steamClient := steam.NewSteamClient(sn.cfg.SteamAPIKey)

	for completedSession := range sn.in {
		log.Printf("SessionNotifier: New session received with %d matches", len(completedSession.Matches))

		sessionWithDetails := parser.SessionWithDetails{TrackedPlayers: sn.cfg.Players}

		// Players flags are used for single match session only
		var err error
		steamPlayers := []steam.SteamPlayer{}
		if len(completedSession.Matches) == 1 {
			allSteamIDs := completedSession.GetSteamIDs()
			steamPlayers, err = steamClient.GetSteamPlayers(allSteamIDs)
			if err != nil {
				// Continue without steam data
				log.Printf("SessionNotifier: Warning: failed to get steam players: %v", err)
			}
		}

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
				time.Sleep(3 * time.Minute)
			}
		}

		// sort matches by chronological order from oldest to newest
		sessionWithDetails.SortMatchesByEndTime()

		// Send Discord webhook
		discordClient.SendSessionResult(sessionWithDetails)
	}
}
