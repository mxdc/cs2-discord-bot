package session

import (
	"sort"
	"time"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

const (
	sessionDuration = 3*time.Hour + 15*time.Minute
	sessionTimeout  = 3*time.Hour + 30*time.Minute
)

type GameSession struct {
	Matches           []leetify.Game
	LastMatchEndTime  time.Time
	LastDetectionTime time.Time
	IsFresh           bool
	debugMode         bool
}

func NewSession(game leetify.Game, detectedAt time.Time, debugMode bool) *GameSession {
	return &GameSession{
		Matches:           []leetify.Game{game},
		LastMatchEndTime:  game.FinishedAt(),
		LastDetectionTime: detectedAt,
		IsFresh:           false,
		debugMode:         debugMode,
	}
}

func (s *GameSession) AddMatch(game leetify.Game, detectedAt time.Time) {
	s.Matches = append(s.Matches, game)

	// Sort matches chronologically from oldest to newest
	sort.Slice(s.Matches, func(i, j int) bool {
		return s.Matches[i].FinishedAt().Before(s.Matches[j].FinishedAt())
	})

	if len(s.Matches) > 0 {
		s.LastMatchEndTime = s.Matches[len(s.Matches)-1].FinishedAt()
	}

	s.LastDetectionTime = detectedAt
}

func (s *GameSession) IsSessionTimeout() bool {
	if s.debugMode {
		return time.Since(s.LastMatchEndTime) > sessionTimeout
	}

	return time.Since(s.LastDetectionTime) > sessionTimeout
}

func (s *GameSession) IsMatchPartOfSession(game leetify.Game) bool {
	diff := game.FinishedAt().Sub(s.LastMatchEndTime).Abs()

	return diff <= sessionDuration
}

func (s *GameSession) IsMatchBeforeCurrentSession(game leetify.Game) bool {
	return game.FinishedAt().Before(s.LastMatchEndTime)
}

func (s *GameSession) LastMatch() leetify.Game {
	if len(s.Matches) == 0 {
		return leetify.Game{}
	}

	return s.Matches[len(s.Matches)-1]
}
