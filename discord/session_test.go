package discord

import (
	"testing"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/parser"
)

func sessionTestTranslations() locales.Translations {
	return locales.Translations{
		ListSeparator:        ",",
		ListConjunction:      "and",
		SessionAllLosses:     "%s lost every match",
		SessionAllWins:       "%s won every match",
		SessionSingleAllWins: "%s won every match",
		SessionMoreWins:      "%s won more than they lost",
		SessionMoreLosses:    "%s lost more than they won",
		SessionTie:           "%s tied",
		SessionSingleTie:     "%s tied",
	}
}

func TestSessionResultBuilder_SingleAllWins(t *testing.T) {
	session := parser.SessionWithDetails{
		TrackedPlayers: []config.Player{{SteamID: "p1"}},
		Matches: []parser.MatchWithDetails{
			{Winner: 1, OwnTeam: parser.Team{Players: []parser.Player{{SteamID: "p1", Name: "player1"}}}},
			{Winner: 1, OwnTeam: parser.Team{Players: []parser.Player{{SteamID: "p1", Name: "player1"}}}},
		},
	}

	message := NewSessionResultBuilder(session, sessionTestTranslations(), false).BuildMessage()

	if message.Content != "Player1 won every match" {
		t.Errorf("expected 'Player1 won every match', got %q", message.Content)
	}
	if message.Embeds[0].Color != ColorGreen {
		t.Errorf("expected green embed for an all-wins session, got %d", message.Embeds[0].Color)
	}
}

func TestSessionResultBuilder_AllDefeats(t *testing.T) {
	session := parser.SessionWithDetails{
		TrackedPlayers: []config.Player{{SteamID: "p1"}},
		Matches: []parser.MatchWithDetails{
			{Winner: 2, OwnTeam: parser.Team{Players: []parser.Player{{SteamID: "p1", Name: "player1"}}}},
		},
	}

	message := NewSessionResultBuilder(session, sessionTestTranslations(), false).BuildMessage()

	if message.Content != "Player1 lost every match" {
		t.Errorf("expected 'Player1 lost every match', got %q", message.Content)
	}
	if message.Embeds[0].Color != ColorRed {
		t.Errorf("expected red embed for an all-defeats session, got %d", message.Embeds[0].Color)
	}
}
