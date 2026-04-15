package session

import (
	"log"
	"time"
)

type SessionManager struct {
	in        <-chan MatchDetected
	out       chan<- GameSession
	seenGames *SeenGames
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
	seenGames := &SeenGames{games: []SeenGame{}}

	return &SessionManager{
		in:        in,
		out:       out,
		seenGames: seenGames,
		debugMode: debugMode,
	}
}

func (sm *SessionManager) HandleIncomingMatches() {
	var currentSession *GameSession
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

			if !sm.seenGames.ShouldNotify(msg.Player.SteamID, msg.Match) {
				log.Printf("SessionManager: Match %s has already been seen, ignoring", msg.Match.GameId)
				continue
			}

			sm.seenGames.AddGame(msg.Player.SteamID, msg.Match.GameId, msg.Match.GameFinishedAt)

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
			sm.flush(currentSession)

			currentSession = NewSession(msg.Match, msg.DetectedAt, sm.debugMode)
			log.Printf("SessionManager: Started new session with match %s", msg.Match.GameId)

		case <-ticker.C:
			if currentSession == nil {
				continue
			}

			if currentSession.IsSessionTimeout() {
				log.Printf("SessionManager: Inactivity timeout reached, flushing session")
				sm.flush(currentSession)
				currentSession = nil
			}
		}
	}
}

func (sm *SessionManager) flush(currentSession *GameSession) {
	if currentSession == nil {
		return
	}

	last := currentSession.LastMatch()
	recent := sm.seenGames.MostRecentGame()
	if len(last.GameId) > 0 && len(recent.GameID) > 0 && last.GameId == recent.GameID {
		currentSession.IsFresh = true
	}
	sm.out <- *currentSession
}
