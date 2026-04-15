package session

import (
	"sort"
	"time"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

type GameSession struct {
	Matches           []leetify.LeetifyGameResponse
	LastMatchEndTime  time.Time
	LastDetectionTime time.Time
	sessionDuration   time.Duration
	sessionTimeout    time.Duration
	IsFresh           bool
	debugMode         bool
}

func NewSession(game leetify.LeetifyGameResponse, detectedAt time.Time, debugMode bool) *GameSession {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)

	return &GameSession{
		Matches:           []leetify.LeetifyGameResponse{game},
		LastMatchEndTime:  matchEndTime,
		LastDetectionTime: detectedAt,
		sessionDuration:   3*time.Hour + 15*time.Minute,
		sessionTimeout:    3*time.Hour + 30*time.Minute,
		IsFresh:           false,
		debugMode:         debugMode,
	}
}

func (s *GameSession) AddMatch(game leetify.LeetifyGameResponse, detectedAt time.Time) {
	s.Matches = append(s.Matches, game)

	// Sort matches chronologically from oldest to newest
	sort.Slice(s.Matches, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, s.Matches[i].GameFinishedAt)
		timeJ, _ := time.Parse(time.RFC3339, s.Matches[j].GameFinishedAt)
		return timeI.Before(timeJ)
	})

	if len(s.Matches) > 0 {
		lastMatchTime, _ := time.Parse(time.RFC3339, s.Matches[len(s.Matches)-1].GameFinishedAt)
		s.LastMatchEndTime = lastMatchTime
	}

	s.LastDetectionTime = detectedAt
}

func (s *GameSession) IsSessionTimeout() bool {
	if s.debugMode {
		return time.Since(s.LastMatchEndTime) > s.sessionTimeout
	}

	return time.Since(s.LastDetectionTime) > s.sessionTimeout
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

func (s *GameSession) LastMatch() leetify.LeetifyGameResponse {
	if len(s.Matches) == 0 {
		return leetify.LeetifyGameResponse{}
	}

	return s.Matches[len(s.Matches)-1]
}
