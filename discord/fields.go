package discord

import (
	"fmt"
	"strings"

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

func (f *EmbedFieldFormatter) addMatchOneLinerField(match parser.MatchWithDetails) {
	matchLink := match.GetMatchLink()
	matchResult := match.GetOneLinerResult()
	resultEmoji := getResultPrefixEmoji(match.Winner)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("%s [**%s**](%s)", resultEmoji, matchResult, matchLink),
		Inline: false,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addSessionMatchesField(matches []parser.MatchWithDetails) {
	if len(matches) == 0 {
		return
	}

	lines := make([]string, len(matches))
	for i, match := range matches {
		resultEmoji := getResultPrefixEmoji(match.Winner)
		matchLink := match.GetMatchLink()
		matchResult := match.GetOneLinerResult()
		matchResultWithLink := fmt.Sprintf("%s [**%s**](%s)", resultEmoji, matchResult, matchLink)
		lines[i] = matchResultWithLink
	}

	field := EmbedField{
		Name:   "",
		Value:  strings.Join(lines, "\n"),
		Inline: false,
	}
	f.fields = append(f.fields, field)
}
