package session

import (
	"testing"
	"time"

	"github.com/mxdc/cs2-discord-bot/leetify"
)

func gameFinishedAt(t time.Time) leetify.Game {
	return leetify.Game{GameFinishedAt: t.Format(time.RFC3339)}
}

func TestIsMatchPartOfSession(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := NewSession(gameFinishedAt(base), base, false)

	within := gameFinishedAt(base.Add(2 * time.Hour))
	if !s.IsMatchPartOfSession(within) {
		t.Error("expected match 2h later to be part of the session")
	}

	outside := gameFinishedAt(base.Add(4 * time.Hour))
	if s.IsMatchPartOfSession(outside) {
		t.Error("expected match 4h later to not be part of the session")
	}
}

func TestIsMatchBeforeCurrentSession(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := NewSession(gameFinishedAt(base), base, false)

	earlier := gameFinishedAt(base.Add(-1 * time.Hour))
	if !s.IsMatchBeforeCurrentSession(earlier) {
		t.Error("expected earlier match to be before the current session")
	}

	later := gameFinishedAt(base.Add(1 * time.Hour))
	if s.IsMatchBeforeCurrentSession(later) {
		t.Error("expected later match to not be before the current session")
	}
}

func TestIsSessionTimeout_NormalMode(t *testing.T) {
	base := time.Now().Add(-4 * time.Hour)
	s := NewSession(gameFinishedAt(base), base, false)

	if !s.IsSessionTimeout() {
		t.Error("expected timeout in normal mode when LastDetectionTime is 4h old")
	}
}

func TestIsSessionTimeout_DebugMode(t *testing.T) {
	base := time.Now().Add(-4 * time.Hour)
	s := NewSession(gameFinishedAt(base), time.Now(), true)

	if !s.IsSessionTimeout() {
		t.Error("expected timeout in debug mode when LastMatchEndTime is 4h old, even though LastDetectionTime is recent")
	}
}
