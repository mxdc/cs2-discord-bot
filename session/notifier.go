package session

import (
	"log"
	"os"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/discord"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/parser"
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
	enricher      *MatchEnricher
	discordClient *discord.WebhookClient
	seenGames     *SeenGames
	in            <-chan MatchDetected
	log           *log.Logger
}

func NewMatchNotifier(
	enricher *MatchEnricher,
	discordClient *discord.WebhookClient,
	seenGames *SeenGames,
	in <-chan MatchDetected,
) *MatchNotifier {
	return &MatchNotifier{
		enricher:      enricher,
		discordClient: discordClient,
		seenGames:     seenGames,
		in:            in,
		log:           log.New(os.Stderr, "MatchNotifier: ", log.LstdFlags),
	}
}

func (mm *MatchNotifier) HandleMatch() {
	mm.log.Println("Started, waiting for matches...")

	for msg := range mm.in {
		if !mm.seenGames.ShouldNotify(msg.Match.GameId) {
			continue
		}

		mm.seenGames.AddGame(msg.Player.SteamID, msg.Match.GameId, msg.Match.GameFinishedAt)
		mm.log.Printf("New match detected: %s", msg.Match.GameId)

		// Avoid rate limit failure
		time.Sleep(5 * time.Minute)

		mm.processMatch(msg.Match)
	}
}

// processMatch enriches and sends one match, recovering from any panic so a
// single bad match doesn't permanently kill this notifier's goroutine.
func (mm *MatchNotifier) processMatch(game leetify.Game) {
	defer func() {
		if r := recover(); r != nil {
			mm.log.Printf("recovered from panic while processing match %s: %v", game.GameId, r)
		}
	}()

	matchWithDetails := mm.enricher.EnrichWithProfiles(game)
	mm.discordClient.SendMatchResult(matchWithDetails)
}

type SessionNotifier struct {
	enricher       *MatchEnricher
	discordClient  *discord.WebhookClient
	trackedPlayers []config.Player
	in             <-chan GameSession
	log            *log.Logger
}

func NewSessionNotifier(
	enricher *MatchEnricher,
	discordClient *discord.WebhookClient,
	trackedPlayers []config.Player,
	in <-chan GameSession,
) *SessionNotifier {
	return &SessionNotifier{
		enricher:       enricher,
		discordClient:  discordClient,
		trackedPlayers: trackedPlayers,
		in:             in,
		log:            log.New(os.Stderr, "SessionNotifier: ", log.LstdFlags),
	}
}

func (sn *SessionNotifier) HandleSession() {
	sn.log.Println("Started, waiting for completed sessions...")

	for completedSession := range sn.in {
		sn.log.Printf("New session received with %d matches", len(completedSession.Matches))
		sn.processSession(completedSession)
	}
}

// processSession enriches and sends one completed session, recovering from
// any panic so a single bad session doesn't permanently kill this
// notifier's goroutine.
func (sn *SessionNotifier) processSession(completedSession GameSession) {
	defer func() {
		if r := recover(); r != nil {
			sn.log.Printf("recovered from panic while processing session: %v", r)
		}
	}()

	sessionWithDetails := parser.SessionWithDetails{
		TrackedPlayers: sn.trackedPlayers,
		IsFresh:        completedSession.IsFresh,
	}

	// Player flags are only fetched for single-match sessions
	isSingleMatchSession := len(completedSession.Matches) == 1

	for i, game := range completedSession.Matches {
		var matchWithDetails parser.MatchWithDetails
		if isSingleMatchSession {
			matchWithDetails = sn.enricher.EnrichWithProfiles(game)
		} else {
			matchWithDetails = sn.enricher.Enrich(game)
		}
		sessionWithDetails.Matches = append(sessionWithDetails.Matches, matchWithDetails)

		// Avoid rate limit failure
		if i < len(completedSession.Matches)-2 {
			time.Sleep(3 * time.Minute)
		}
	}

	// sort matches by chronological order from oldest to newest
	sessionWithDetails.SortMatchesByEndTime()

	sn.discordClient.SendSessionResult(sessionWithDetails)
}
