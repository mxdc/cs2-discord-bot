package session

import (
	"time"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

type GameSession struct {
	Matches          []leetify.LeetifyGameResponse
	LastMatchEndTime time.Time
}

func NewSession(game leetify.LeetifyGameResponse) *GameSession {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)

	return &GameSession{
		Matches:          []leetify.LeetifyGameResponse{game},
		LastMatchEndTime: matchEndTime,
	}
}

func (s *GameSession) AddMatch(game leetify.LeetifyGameResponse) {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)

	s.Matches = append(s.Matches, game)
	s.LastMatchEndTime = matchEndTime
}

func (s *GameSession) IsMatchWithinSession(game leetify.LeetifyGameResponse) bool {
	matchEndTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)
	return matchEndTime.Sub(s.LastMatchEndTime) <= 30*time.Minute
}

func (s *GameSession) GetSteamIDs() []string {
	allSteamIDs := []string{}
	for _, game := range s.Matches {
		allSteamIDs = append(allSteamIDs, game.OwnTeamSteam64Ids...)
		allSteamIDs = append(allSteamIDs, game.EnemyTeamSteam64Ids...)
	}
	return allSteamIDs
}
