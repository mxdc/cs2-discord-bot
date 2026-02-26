package session

import (
	"time"

	"github.com/mxdc/cs2-discord-bot/parser"
)

type GameSession struct {
	Matches          []parser.MatchResult
	LastMatchEndTime time.Time
}

func NewSession(match parser.MatchResult) *GameSession {
	return &GameSession{
		Matches:          []parser.MatchResult{match},
		LastMatchEndTime: match.GameFinishedAt,
	}
}

func (s *GameSession) AddMatch(match parser.MatchResult) {
	s.Matches = append(s.Matches, match)
	s.LastMatchEndTime = match.GameFinishedAt
}

func (s *GameSession) IsMatchWithinSession(match parser.MatchResult) bool {
	return match.GameFinishedAt.Sub(s.LastMatchEndTime) <= 30*time.Minute
}
