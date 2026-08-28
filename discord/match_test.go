package discord

import (
	"testing"

	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/parser"
)

func testTranslations() locales.Translations {
	return locales.Translations{
		ListSeparator:   ",",
		ListConjunction: "and",
		MatchFinished:   "Match finished",
		WinSingle:       "%s won",
		LossSingle:      "%s lost",
		TieSingle:       "%s tied",
		WinMultiple:     "%s won",
		LossMultiple:    "%s lost",
		TieMultiple:     "%s tied",
	}
}

func TestMatchResultBuilder_WinSingle(t *testing.T) {
	match := parser.MatchWithDetails{
		Winner: 1,
		OwnTeam: parser.Team{
			Score:        16,
			KnownPlayers: []parser.Player{{Name: "player1"}},
		},
		EnemyTeam: parser.Team{Score: 10},
	}

	message := NewMatchResultBuilder(match, testTranslations(), false).BuildMessage()

	if message.Content != "Player1 won" {
		t.Errorf("expected 'Player1 won', got %q", message.Content)
	}
	if len(message.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(message.Embeds))
	}
	if message.Embeds[0].Color != ColorGreen {
		t.Errorf("expected green embed for a win, got %d", message.Embeds[0].Color)
	}
}

func TestMatchResultBuilder_MatchFinishedWhenNoScore(t *testing.T) {
	match := parser.MatchWithDetails{}
	message := NewMatchResultBuilder(match, testTranslations(), false).BuildMessage()

	if message.Content != "Match finished" {
		t.Errorf("expected 'Match finished', got %q", message.Content)
	}
}

func TestMatchResultBuilder_Defeat(t *testing.T) {
	match := parser.MatchWithDetails{
		Winner:    2,
		OwnTeam:   parser.Team{Score: 5, KnownPlayers: []parser.Player{{Name: "player1"}}},
		EnemyTeam: parser.Team{Score: 16},
	}

	message := NewMatchResultBuilder(match, testTranslations(), false).BuildMessage()

	if message.Content != "Player1 lost" {
		t.Errorf("expected 'Player1 lost', got %q", message.Content)
	}
	if message.Embeds[0].Color != ColorRed {
		t.Errorf("expected red embed for a loss, got %d", message.Embeds[0].Color)
	}
}
