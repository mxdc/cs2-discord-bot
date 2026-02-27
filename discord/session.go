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
	return fmt.Sprintf(":video_game: %s played %d matches.", names, len(b.session.Matches))
}

func (b *SessionResultBuilder) createSessionEmbed() Embed {
	fieldsFormatter := NewEmbedFieldFormatter()
	fieldsFormatter.addSessionMatchesField(b.session.Matches)
	fieldsFormatter.addSessionTeammatesField(b.session)
	fields := fieldsFormatter.GetFields()

	return Embed{
		Title:  "Session summary",
		Fields: fields,
		Color:  ColorBlue,
	}
}
