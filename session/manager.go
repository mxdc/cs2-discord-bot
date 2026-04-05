package session

import (
	"log"
	"time"
)

type SessionManager struct {
	in        <-chan MatchDetected
	out       chan<- GameSession
	debugMode bool
}

const (
	tickerInterval = 5 * time.Minute
)

func NewSessionManager(
	in <-chan MatchDetected,
	out chan<- GameSession,
	debugMode bool,
) *SessionManager {
	return &SessionManager{
		in:        in,
		out:       out,
		debugMode: debugMode,
	}
}

func (sm *SessionManager) HandleIncomingMatches() {
	var currentSession *GameSession
	seenGames := &SeenGames{games: []SeenGame{}}
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	log.Println("SessionManager: Started, waiting for matches...")

	for {
		select {

		case msg := <-sm.in:
			// drop old matches
			if msg.IsTooOld() {
				log.Printf("SessionManager: Match %s is too old, ignoring", msg.Match.GameId)
				continue
			}

			if seenGames.AlreadyNotified(msg.Match.GameId) {
				continue
			}
			seenGames.AddGame(msg.Player.SteamID, msg.Match.GameId, msg.Match.GameFinishedAt)

			log.Printf("SessionManager: New match detected: %s", msg.Match.GameId)

			if currentSession == nil {
				currentSession = NewSession(msg.Match, msg.DetectedAt, sm.debugMode)
				log.Printf("SessionManager: Started new session with match %s", msg.Match.GameId)
				continue
			}

			if currentSession.IsMatchPartOfSession(msg.Match) {
				currentSession.AddMatch(msg.Match, msg.DetectedAt)
				log.Printf("SessionManager: Added match %s to current session", msg.Match.GameId)
				continue
			}

			if currentSession.IsMatchBeforeCurrentSession(msg.Match) {
				log.Printf("SessionManager: Match %s is before current session, ignoring", msg.Match.GameId)
				continue
			}

			log.Printf("SessionManager: Match too far in time, flushing session")
			sm.out <- *currentSession

			currentSession = NewSession(msg.Match, msg.DetectedAt, sm.debugMode)
			log.Printf("SessionManager: Started new session with match %s", msg.Match.GameId)

		case <-ticker.C:
			if currentSession == nil {
				continue
			}

			if currentSession.IsSessionTimeout() {
				log.Printf("SessionManager: Inactivity timeout reached, flushing session")
				sm.out <- *currentSession
				currentSession = nil
			}
		}
	}
}
