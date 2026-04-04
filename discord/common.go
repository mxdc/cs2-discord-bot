package discord

import (
	"fmt"

	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/parser"
)

func formatPlayerNamesAsTitle(players []parser.Player, translations locales.Translations) string {
	header := ""
	separator := fmt.Sprintf("%s ", translations.ListSeparator)
	conjuction := fmt.Sprintf(" %s ", translations.ListConjunction)

	for i, player := range players {
		playerName := player.FormatPlayerTitle()
		header += playerName
		if i < len(players)-2 {
			header += separator
		} else if i < len(players)-1 {
			header += conjuction
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
