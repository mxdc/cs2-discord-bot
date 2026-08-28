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
	content := b.formatMatchHeader()
	embed := b.createMatchEmbed()

	return WebhookMessage{
		Content:  content,
		TTS:      false,
		Embeds:   []Embed{embed},
		Username: b.translations.BotUsername,
	}
}

func (b *MatchResultBuilder) formatMatchHeader() string {
	match := b.match

	if match.OwnTeam.Score == 0 && match.EnemyTeam.Score == 0 {
		return b.translations.MatchFinished
	}

	knownPlayers := match.OwnTeam.KnownPlayers
	header := formatPlayerNamesAsTitle(knownPlayers, b.translations)

	if len(knownPlayers) == 1 {
		return formatMatchHeaderForSinglePlayer(b.translations, match, header, knownPlayers[0], b.withRank)
	}

	return formatMatchHeaderForMultiplePlayers(b.translations, match, header)
}

func (b *MatchResultBuilder) createMatchEmbed() Embed {
	match := b.match

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

	return Embed{
		Title:  "",
		Color:  color,
		Fields: fieldsFormatter.GetFields(),
	}
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
