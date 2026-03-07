package discord

import (
	"github.com/mxdc/cs2-discord-bot/parser"
)

func formatPlayerNamesAsTitle(players []parser.Player) string {
	header := ""

	for i, player := range players {
		playerName := player.FormatPlayerLink(true, true)
		header += playerName
		if i < len(players)-2 {
			header += ", "
		} else if i < len(players)-1 {
			header += " and "
		}
	}

	return header
}

func getResultPrefixEmoji(winner int) string {
	if winner == 1 {
		return "🏆"
	}
	if winner == 2 {
		return "💀"
	}
	return "🤝"
}
