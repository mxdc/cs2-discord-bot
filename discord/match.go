package discord

import (
	"fmt"

	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/parser"
)

type MatchResultBuilder struct {
	match        parser.MatchWithDetails
	translations locales.Translations
	withRank     bool
}

func NewMatchResultBuilder(
	match parser.MatchWithDetails,
	translations locales.Translations,
	withRank bool,
) *MatchResultBuilder {
	return &MatchResultBuilder{
		match:        match,
		translations: translations,
		withRank:     withRank,
	}
}

func (b *MatchResultBuilder) BuildMessage() WebhookMessage {
	content := formatMatchHeader(b.match, b.translations, b.withRank)
	embed := createMatchEmbed(b.match)

	return WebhookMessage{
		Content:  content,
		TTS:      false,
		Embeds:   []Embed{embed},
		Username: b.translations.BotUsername,
	}
}

func formatMatchHeader(match parser.MatchWithDetails, translations locales.Translations, withRank bool) string {
	if match.OwnTeam.Score == 0 && match.EnemyTeam.Score == 0 {
		return translations.MatchFinished
	}

	knownPlayers := match.OwnTeam.KnownPlayers
	header := formatPlayerNamesAsTitle(knownPlayers, translations)

	if len(knownPlayers) == 1 {
		return formatMatchHeaderForSinglePlayer(translations, match, header, knownPlayers[0], withRank)
	}

	return formatMatchHeaderForMultiplePlayers(translations, match, header)
}

func createMatchEmbed(match parser.MatchWithDetails) Embed {
	var color int

	if match.Winner == 1 {
		color = ColorGreen
	} else if match.Winner == 2 {
		color = ColorRed
	} else {
		color = ColorGray
	}

	fieldsFormatter := NewEmbedFieldFormatter()
	fieldsFormatter.addMatchOneLinerField(match)
	// fieldsFormatter.addGameModeField(match.GameMode)
	// fieldsFormatter.addScoreField(match)
	// fieldsFormatter.addMapNameField(match.MapName)
	// fieldsFormatter.addPlayerMVPField(match)
	// fieldsFormatter.addMatchLinkField(match)

	formattedFields := fieldsFormatter.GetFields()

	embed := Embed{
		Title:  "",
		Color:  color,
		Fields: formattedFields,
	}

	return embed
}

func formatMatchHeaderForSinglePlayer(
	translations locales.Translations,
	match parser.MatchWithDetails,
	playerNameHeader string,
	knownPlayer parser.Player,
	withRank bool,
) string {
	t := translations

	_, newRank := knownPlayer.GetRecentPremierRank()
	displayRank := newRank > 0 && withRank && match.IsPremierMode()

	switch match.Winner {
	case 1:
		if displayRank {
			return fmt.Sprintf(t.WinSingleRank, playerNameHeader, newRank)
		}
		return fmt.Sprintf(t.WinSingle, playerNameHeader)
	case 2:
		if displayRank {
			return fmt.Sprintf(t.LossSingleRank, playerNameHeader, newRank)
		}
		return fmt.Sprintf(t.LossSingle, playerNameHeader)
	default:
		return fmt.Sprintf(t.TieSingle, playerNameHeader)
	}
}

func formatMatchHeaderForMultiplePlayers(
	translations locales.Translations,
	match parser.MatchWithDetails,
	playerNamesHeader string,
) string {
	t := translations

	switch match.Winner {
	case 1:
		return fmt.Sprintf(t.WinMultiple, playerNamesHeader)
	case 2:
		return fmt.Sprintf(t.LossMultiple, playerNamesHeader)
	default:
		return fmt.Sprintf(t.TieMultiple, playerNamesHeader)
	}
}
