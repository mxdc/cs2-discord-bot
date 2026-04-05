package session

type SeenGame struct {
	SteamID        string
	GameID         string
	GameFinishedAt string
}

type SeenGames struct {
	games []SeenGame
}

func (sg *SeenGames) AlreadyNotified(gameId string) bool {
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
