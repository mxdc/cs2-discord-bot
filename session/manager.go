package session

import (
	"log"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/leetify"
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

			log.Printf("SessionManager: New match detected: %s", msg.Match.GameId)

			if currentSession == nil {
				currentSession = NewSession(msg.Match)
				log.Printf("SessionManager: Started new session with match %s", msg.Match.GameId)
				continue
			}

			if currentSession.IsMatchWithinSession(msg.Match) {
				currentSession.AddMatch(msg.Match)
				log.Printf("SessionManager: Added match %s to current session", msg.Match.GameId)
				continue
			}

			log.Printf("SessionManager: Match too far in time, flushing session")
			sm.out <- *currentSession

			currentSession = NewSession(msg.Match)
			log.Printf("SessionManager: Started new session with match %s", msg.Match.GameId)

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
