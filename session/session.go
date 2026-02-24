package session

import (
	"time"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

type GameSession struct {
	Matches          []leetify.MatchResult
	LastMatchEndTime time.Time
}

func NewSession(match leetify.MatchResult) *GameSession {
	return &GameSession{
		Matches:          []leetify.MatchResult{match},
		LastMatchEndTime: match.GameFinishedAt,
	}
}

func (s *GameSession) AddMatch(match leetify.MatchResult) {
	s.Matches = append(s.Matches, match)
	s.LastMatchEndTime = match.GameFinishedAt
}

func (s *GameSession) IsMatchWithinSession(match leetify.MatchResult) bool {
	return match.GameFinishedAt.Sub(s.LastMatchEndTime) <= 30*time.Minute
}
