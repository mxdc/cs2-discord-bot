package session

import (
	"log"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/parser"
)

type SessionManager struct {
	cfg    *config.AppConfig
	client *leetify.LeetifyClient
	in     <-chan MatchDetected
	out    chan<- GameSession
}

const (
	sessionTimeout = 45 * time.Minute
	tickerInterval = 5 * time.Minute
)

func NewSessionManager(
	cfg *config.AppConfig,
	client *leetify.LeetifyClient,
	in <-chan MatchDetected,
	out chan<- GameSession,
) *SessionManager {
	return &SessionManager{
		cfg:    cfg,
		client: client,
		in:     in,
		out:    out,
	}
}

func (sm *SessionManager) HandleIncomingMatches() {
	var currentSession *GameSession
	seen := make(map[string]bool)
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	log.Println("SessionManager: Started, waiting for matches...")

	for {
		select {

		case msg := <-sm.in:
			if seen[msg.Match.GameId] {
				continue
			}
			seen[msg.Match.GameId] = true

			match := parser.ParseGameResponseFromLeetify(msg.Match)
			log.Printf("SessionManager: New match detected: %s", match.GameID)

			if currentSession == nil {
				currentSession = NewSession(match)
				log.Printf("SessionManager: Started new session with match %s", match.GameID)
				continue
			}

			if currentSession.IsMatchWithinSession(match) {
				currentSession.AddMatch(match)
				log.Printf("SessionManager: Added match %s to current session", match.GameID)
				continue
			}

			log.Printf("SessionManager: Match too far in time, flushing session")
			sm.out <- *currentSession

			currentSession = NewSession(match)
			log.Printf("SessionManager: Started new session with match %s", match.GameID)

		case <-ticker.C:
			if currentSession == nil {
				continue
			}

			if time.Since(currentSession.LastMatchEndTime) > sessionTimeout {
				log.Printf("SessionManager: Inactivity timeout reached, flushing session")
				sm.out <- *currentSession
				currentSession = nil
			}
		}
	}
}
