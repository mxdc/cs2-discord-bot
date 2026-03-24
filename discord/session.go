package discord

import (
	"fmt"

	"github.com/mxdc/cs2-discord-bot/parser"
)

type SessionResultBuilder struct {
	session parser.SessionWithDetails
}

func NewSessionResultBuilder(session parser.SessionWithDetails) *SessionResultBuilder {
	return &SessionResultBuilder{
		session: session,
	}
}

func (b *SessionResultBuilder) BuildMessage() WebhookMessage {
	content := b.formatSessionHeader()
	embed := b.createSessionEmbed()

	return WebhookMessage{
		Content:  content,
		TTS:      false,
		Embeds:   []Embed{embed},
		Username: "CS2",
	}
}

func (b *SessionResultBuilder) formatSessionHeader() string {
	knownPlayers := b.session.KnownPlayers()

	names := formatPlayerNamesAsTitle(knownPlayers)
	if b.session.AllMatchDefeats() {
		if len(knownPlayers) == 1 {
			return fmt.Sprintf("%s is on a losing streak.", names)
		}
		return fmt.Sprintf("%s are on a losing streak.", names)
	}

	if b.session.AllMatchVictories() {
		if len(knownPlayers) == 1 {
			return fmt.Sprintf("%s is on a winning streak.", names)
		}
		return fmt.Sprintf("%s are on a winning streak.", names)
	}

	return fmt.Sprintf("%s played %d matches.", names, len(b.session.Matches))
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
