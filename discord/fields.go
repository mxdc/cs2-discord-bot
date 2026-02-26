package discord

import (
	"fmt"

	"github.com/mxdc/cs2-discord-bot/parser"
)

type EmbedFieldFormatter struct {
	fields []EmbedField
}

func NewEmbedFieldFormatter() *EmbedFieldFormatter {
	return &EmbedFieldFormatter{fields: []EmbedField{}}
}

func (f *EmbedFieldFormatter) GetFields() []EmbedField {
	return f.fields
}

func (f *EmbedFieldFormatter) addGameModeField(gameMode string) {
	if len(gameMode) == 0 {
		return
	}

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("*Mode* **%s**", gameMode),
		Inline: true,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addScoreField(match parser.MatchWithDetails) {
	if match.OwnTeam.Score == 0 && match.EnemyTeam.Score == 0 {
		return
	}

	score := fmt.Sprintf("%d - %d", match.OwnTeam.Score, match.EnemyTeam.Score)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("*Score* **%s**", score),
		Inline: true,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addMapNameField(mapName string) {
	if len(mapName) == 0 {
		return
	}

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("*Map* **%s**", mapName),
		Inline: true,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addPlayerMVPField(match parser.MatchWithDetails) {
	if len(match.OwnTeam.Players) == 0 || len(match.EnemyTeam.Players) == 0 {
		return
	}

	matchMVP := findMVP(match)
	playerLink := formatPlayerLink(matchMVP)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("⭐ *Match MVP*\u00A0\u00A0\u00A0\u00A0**%s**", playerLink),
		Inline: false,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addMatchLinkField(gameID string) {
	if len(gameID) == 0 {
		return
	}

	matchLink := fmt.Sprintf("▸ [View match details on Leetify](https://leetify.com/public/match-details/%s/details-general)", gameID)

	field := EmbedField{
		Name:   "",
		Value:  matchLink,
		Inline: false,
	}

	f.fields = append(f.fields, field)
}
