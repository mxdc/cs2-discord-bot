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
	Match      leetify.Game
	Player     config.Player
	DetectedAt time.Time
}

func (md *MatchDetected) IsTooOld() bool {
	return time.Since(md.Match.FinishedAt()) > 24*time.Hour
}

type MatchNotifier struct {
	cfg           *config.AppConfig
	client        *leetify.Client
	mistralClient *mistral.Client
	translations  locales.Translations
	in            <-chan MatchDetected
}

func NewMatchNotifier(
	cfg *config.AppConfig,
	client *leetify.Client,
	mistralClient *mistral.Client,
	translations locales.Translations,
	in <-chan MatchDetected,
) *MatchNotifier {
	return &MatchNotifier{
		cfg:           cfg,
		client:        client,
		mistralClient: mistralClient,
		translations:  translations,
		in:            in,
	}
}

func (mm *MatchNotifier) HandleMatch() {
	log.Println("Notifier: Started notifier, waiting for matches...")
	seenGames := &SeenGames{games: []SeenGame{}}
	discordClient := discord.NewWebhookClient(mm.cfg.DiscordHook, mm.mistralClient, mm.translations, false)
	steamClient := steam.New(mm.cfg.SteamAPIKey)

	for msg := range mm.in {
		if !seenGames.ShouldNotify(msg.Player.SteamID, msg.Match) {
			continue
		}

		seenGames.AddGame(msg.Player.SteamID, msg.Match.GameId, msg.Match.GameFinishedAt)
		log.Println("Manager: New match detected:", msg.Match.GameId)

		time.Sleep(5 * time.Minute)
		matchDetails, err := mm.client.GetMatchDetails(msg.Match.GameId)
		if err != nil {
			// Continue without match details
			log.Printf("Manager: Warning: failed to get match details: %v", err)
		}

		// Get all Steam IDs from match details
		var allSteamIDs []string
		if matchDetails != nil {
			for _, ps := range matchDetails.PlayerStats {
				allSteamIDs = append(allSteamIDs, ps.Steam64ID)
			}
		}

		// Get Steam player data (names and countries)
		steamPlayers, err := steamClient.GetSteamPlayers(allSteamIDs)
		if err != nil {
			// Continue without steam data
			log.Printf("Manager: Warning: failed to get steam players: %v", err)
		}

		matchWithDetails := parser.ParseMatchResultWithDetails(msg.Match, matchDetails, steamPlayers, mm.cfg.Players)

		// Send Discord webhook
		discordClient.SendMatchResult(matchWithDetails)
	}
}

type SessionNotifier struct {
	client        *leetify.Client
	cfg           *config.AppConfig
	mistralClient *mistral.Client
	translations  locales.Translations
	in            <-chan GameSession
	withRank      bool
}

func NewSessionNotifier(
	cfg *config.AppConfig,
	leetifyClient *leetify.Client,
	mistralClient *mistral.Client,
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
	steamClient := steam.New(sn.cfg.SteamAPIKey)

	for completedSession := range sn.in {
		log.Printf("SessionNotifier: New session received with %d matches", len(completedSession.Matches))

		sessionWithDetails := parser.SessionWithDetails{
			TrackedPlayers: sn.cfg.Players,
			IsFresh:        completedSession.IsFresh,
		}

		// Players flags are used for single match session only
		steamPlayers := []steam.SteamPlayer{}
		if len(completedSession.Matches) == 1 {
			// Fetch match details first to get Steam IDs
			matchDetails, err := sn.client.GetMatchDetails(completedSession.Matches[0].GameId)
			if err == nil && matchDetails != nil {
				var allSteamIDs []string
				for _, ps := range matchDetails.PlayerStats {
					allSteamIDs = append(allSteamIDs, ps.Steam64ID)
				}
				steamPlayers, err = steamClient.GetSteamPlayers(allSteamIDs)
				if err != nil {
					// Continue without steam data
					log.Printf("SessionNotifier: Warning: failed to get steam players: %v", err)
				}
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
