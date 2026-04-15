package session

import (
	"time"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

type SeenGame struct {
	SteamID        string
	GameID         string
	GameFinishedAt string
}

type SeenGames struct {
	games []SeenGame
}

func (sg *SeenGames) ShouldNotify(steamID string, game leetify.LeetifyGameResponse) bool {
	gameID := game.GameId

	return sg.alreadyNotified(gameID) == false
}

func (sg *SeenGames) alreadyNotified(gameId string) bool {
	for _, game := range sg.games {
		if game.GameID == gameId {
			return true
		}
	}

	return false
}

func (sg *SeenGames) AddGame(steamID, gameID, gameFinishedAt string) {
	seenGame := SeenGame{
		SteamID:        steamID,
		GameID:         gameID,
		GameFinishedAt: gameFinishedAt,
	}
	sg.games = append(sg.games, seenGame)
}

func (sg *SeenGames) MostRecentGame() SeenGame {
	if len(sg.games) == 0 {
		return SeenGame{}
	}

	var mostRecentGame SeenGame
	var mostRecentTime time.Time

	for i, game := range sg.games {
		gameTime, err := time.Parse(time.RFC3339, game.GameFinishedAt)
		if err != nil {
			continue
		}

		if len(mostRecentGame.GameID) == 0 || gameTime.After(mostRecentTime) {
			mostRecentGame = sg.games[i]
			mostRecentTime = gameTime
		}
	}

	return mostRecentGame
}
