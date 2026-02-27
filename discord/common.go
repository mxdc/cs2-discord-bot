package discord

import (
	"fmt"
	"strings"

	"github.com/mxdc/cs2-discord-bot/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func formatPlayerNamesAsTitle(players []parser.Player) string {
	header := ""

	for i, player := range players {
		playerName := cases.Title(language.English).String(strings.ToLower(player.Name))
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

func formatPlayerLink(player parser.Player) string {
	playerName := fmt.Sprintf("[%s](https://leetify.com/public/profile/%s)", player.Name, player.SteamID)

	if player.CountryCode != "" {
		flag := CountryCodeToFlag(player.CountryCode)
		playerName = fmt.Sprintf("%s %s", flag, playerName)
	}

	return playerName
}
