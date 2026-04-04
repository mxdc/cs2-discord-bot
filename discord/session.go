package discord

import (
	"fmt"

	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/parser"
)

type SessionResultBuilder struct {
	session      parser.SessionWithDetails
	translations locales.Translations
	withRank     bool
}

func NewSessionResultBuilder(session parser.SessionWithDetails, translations locales.Translations, withRank bool) *SessionResultBuilder {
	return &SessionResultBuilder{
		session:      session,
		translations: translations,
		withRank:     withRank,
	}
}

func (b *SessionResultBuilder) BuildMessage() WebhookMessage {
	content := b.formatSessionHeader()
	embed := b.createSessionEmbed()

	return WebhookMessage{
		Content:  content,
		TTS:      false,
		Embeds:   []Embed{embed},
		Username: b.translations.BotUsername,
	}
}

func (b *SessionResultBuilder) formatSessionHeader() string {
	knownPlayers := b.session.KnownPlayersWithCumulatedStats()

	names := formatPlayerNamesAsTitle(knownPlayers, b.translations)

	if len(knownPlayers) == 1 {
		return formatSessionHeaderForSinglePlayer(b.translations, b.session, names, knownPlayers[0], b.withRank)
	}

	return formatSessionHeaderForMultiplePlayers(b.translations, b.session, names)
}

func (b *SessionResultBuilder) createSessionEmbed() Embed {
	fieldsFormatter := NewEmbedFieldFormatter()
	fieldsFormatter.addSessionMatchesField(b.session.Matches)
	// fieldsFormatter.addSessionTeammatesField(b.session, false)
	// fieldsFormatter.addSessionCumulatedScoresField(b.session)
	// fieldsFormatter.addSessionRankUpdate(b.session)
	fields := fieldsFormatter.GetFields()

	color := ColorBlue
	if b.session.AllMatchDefeats() {
		color = ColorRed
	} else if b.session.AllMatchVictories() {
		color = ColorGreen
	}

	return Embed{
		Title:  "",
		Fields: fields,
		Color:  color,
	}
}

func formatSessionHeaderForSinglePlayer(
	translations locales.Translations,
	session parser.SessionWithDetails,
	playerNameHeader string,
	knownPlayer parser.Player,
	withRank bool,
) string {
	t := translations

	rankStats := knownPlayer.RankStats
	newRank := 0
	if rankStats.RankType == 11 && rankStats.RankChanged && rankStats.Rank > 0 {
		newRank = rankStats.Rank
	}

	if session.AllMatchDefeats() {
		if withRank && newRank > 0 {
			return fmt.Sprintf(t.SessionSingleAllLossesRank, playerNameHeader, newRank)
		}
		return fmt.Sprintf(t.SessionAllLosses, playerNameHeader)
	}

	if session.AllMatchVictories() {
		if withRank && newRank > 0 {
			return fmt.Sprintf(t.SessionSingleAllWinsRank, playerNameHeader, newRank)
		}
		return fmt.Sprintf(t.SessionSingleAllWins, playerNameHeader)
	}

	if session.MoreVictoriesThanDefeats() {
		return fmt.Sprintf(t.SessionMoreWins, playerNameHeader)
	}

	if session.MoreDefeatsThanVictories() {
		return fmt.Sprintf(t.SessionMoreLosses, playerNameHeader)
	}

	return fmt.Sprintf(t.SessionSingleTie, playerNameHeader)
}

func formatSessionHeaderForMultiplePlayers(translations locales.Translations, session parser.SessionWithDetails, playerNamesHeader string) string {
	t := translations

	if session.AllMatchDefeats() {
		return fmt.Sprintf(t.SessionAllLosses, playerNamesHeader)
	}

	if session.AllMatchVictories() {
		return fmt.Sprintf(t.SessionAllWins, playerNamesHeader)
	}

	if session.MoreVictoriesThanDefeats() {
		return fmt.Sprintf(t.SessionMoreWins, playerNamesHeader)
	}

	if session.MoreDefeatsThanVictories() {
		return fmt.Sprintf(t.SessionMoreLosses, playerNamesHeader)
	}

	return fmt.Sprintf(t.SessionTie, playerNamesHeader)
}
