package session

import (
	"log"
	"time"
)

type SessionManager struct {
	in  <-chan MatchDetected
	out chan<- GameSession
}

const (
	tickerInterval = 5 * time.Minute
)

func NewSessionManager(
	in <-chan MatchDetected,
	out chan<- GameSession,
) *SessionManager {
	return &SessionManager{
		in:  in,
		out: out,
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

			if currentSession.IsMatchPartOfSession(msg.Match) {
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

			if currentSession.IsSessionFinished() {
				log.Printf("SessionManager: Inactivity timeout reached, flushing session")
				sm.out <- *currentSession
				currentSession = nil
			}
		}
	}
}
