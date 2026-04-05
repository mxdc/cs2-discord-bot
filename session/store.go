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

	return sg.alreadyNotified(gameID) == false && sg.isMostRecentForPlayer(steamID, game) == true
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

func (sg *SeenGames) isMostRecentForPlayer(steamID string, game leetify.LeetifyGameResponse) bool {
	if len(sg.games) == 0 {
		return true
	}

	gameTime, err := time.Parse(time.RFC3339, game.GameFinishedAt)
	if err != nil {
		return false
	}

	// Find the most recent game in the store for the specific player
	var mostRecentStoreTime time.Time
	hasMostRecent := false

	for _, storedGame := range sg.games {
		// Check if this stored game involves the specific player
		if storedGame.SteamID == steamID {
			storedTime, err := time.Parse(time.RFC3339, storedGame.GameFinishedAt)
			if err != nil {
				continue
			}

			if !hasMostRecent || storedTime.After(mostRecentStoreTime) {
				mostRecentStoreTime = storedTime
				hasMostRecent = true
			}
		}
	}

	// If no games found in store for this player, this game is the most recent
	if !hasMostRecent {
		return true
	}

	// Compare the game time with the most recent stored game time for this player
	return gameTime.After(mostRecentStoreTime) || gameTime.Equal(mostRecentStoreTime)
}
