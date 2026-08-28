package crawler

import (
	"testing"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

func TestFindNewMatches_DedupAndSort(t *testing.T) {
	previous := []leetify.Game{
		{GameId: "a", GameFinishedAt: "2024-01-01T10:00:00Z"},
	}
	current := []leetify.Game{
		{GameId: "a", GameFinishedAt: "2024-01-01T10:00:00Z"},
		{GameId: "c", GameFinishedAt: "2024-01-03T10:00:00Z"},
		{GameId: "b", GameFinishedAt: "2024-01-02T10:00:00Z"},
	}

	newMatches := findNewMatches(previous, current)

	if len(newMatches) != 2 {
		t.Fatalf("expected 2 new matches, got %d", len(newMatches))
	}
	if newMatches[0].GameId != "b" || newMatches[1].GameId != "c" {
		t.Errorf("expected chronological order [b, c], got [%s, %s]", newMatches[0].GameId, newMatches[1].GameId)
	}
}

func TestFindNewMatches_NoNewMatches(t *testing.T) {
	games := []leetify.Game{{GameId: "a", GameFinishedAt: "2024-01-01T10:00:00Z"}}
	newMatches := findNewMatches(games, games)
	if len(newMatches) != 0 {
		t.Errorf("expected 0 new matches, got %d", len(newMatches))
	}
}
