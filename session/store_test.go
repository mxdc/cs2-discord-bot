package session

import "testing"

func TestSeenGames_ShouldNotifyAndAddGame(t *testing.T) {
	sg := NewSeenGames()

	if !sg.ShouldNotify("game1") {
		t.Error("expected ShouldNotify true for a game never seen")
	}

	sg.AddGame("steam1", "game1", "2024-01-01T10:00:00Z")

	if sg.ShouldNotify("game1") {
		t.Error("expected ShouldNotify false after AddGame")
	}
	if !sg.ShouldNotify("game2") {
		t.Error("expected ShouldNotify true for a different game")
	}
}

func TestSeenGames_MostRecentGame(t *testing.T) {
	sg := NewSeenGames()
	sg.AddGame("steam1", "older", "2024-01-01T10:00:00Z")
	sg.AddGame("steam1", "newer", "2024-01-02T10:00:00Z")

	recent := sg.MostRecentGame()
	if recent.GameID != "newer" {
		t.Errorf("expected most recent game to be 'newer', got %q", recent.GameID)
	}
}

func TestSeenGames_MostRecentGame_Empty(t *testing.T) {
	sg := NewSeenGames()
	recent := sg.MostRecentGame()
	if recent.GameID != "" {
		t.Errorf("expected empty SeenGame, got %+v", recent)
	}
}
