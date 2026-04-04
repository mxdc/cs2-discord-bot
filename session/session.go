package session

import (
	"time"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

type GameSession struct {
	Matches                []leetify.LeetifyGameResponse
	LastMatchEndTime       time.Time
	LastMatchDetectionTime time.Time
	sessionDuration        time.Duration
	sessionTimeout         time.Duration
	debugMode              bool
}

func NewSession(game leetify.LeetifyGameResponse, detectedAt time.Time, debugMode bool) *GameSession {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)

	return &GameSession{
		Matches:                []leetify.LeetifyGameResponse{game},
		LastMatchEndTime:       matchEndTime,
		LastMatchDetectionTime: detectedAt,
		sessionDuration:        3*time.Hour + 15*time.Minute,
		sessionTimeout:         3*time.Hour + 30*time.Minute,
		debugMode:              debugMode,
	}
}

func (s *GameSession) AddMatch(game leetify.LeetifyGameResponse, detectedAt time.Time) {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)

	s.Matches = append(s.Matches, game)
	s.LastMatchEndTime = matchEndTime
	s.LastMatchDetectionTime = detectedAt
}

func (s *GameSession) IsSessionTimeout() bool {
	if s.debugMode {
		return time.Since(s.LastMatchEndTime) > s.sessionTimeout
	}

	return time.Since(s.LastMatchDetectionTime) > s.sessionTimeout
}

func (s *GameSession) IsMatchPartOfSession(game leetify.LeetifyGameResponse) bool {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)
	diff := matchEndTime.Sub(s.LastMatchEndTime).Abs()

	return diff <= s.sessionDuration
}

func (s *GameSession) GetSteamIDs() []string {
	allSteamIDs := []string{}
	for _, game := range s.Matches {
		allSteamIDs = append(allSteamIDs, game.OwnTeamSteam64Ids...)
		allSteamIDs = append(allSteamIDs, game.EnemyTeamSteam64Ids...)
	}
	return allSteamIDs
}

func (s *GameSession) IsMatchBeforeCurrentSession(game leetify.LeetifyGameResponse) bool {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)
	return matchEndTime.Before(s.LastMatchEndTime)
}
