package parser

import (
	"testing"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/leetify"
)

func TestParseGameResponseFromLeetify_Win(t *testing.T) {
	game := leetify.Game{
		GameId:         "g1",
		GameFinishedAt: "2024-01-01T10:00:00Z",
		MapName:        "de_dust2",
		DataSource:     "matchmaking",
		MatchResult:    "win",
		Scores:         []int{16, 10},
	}

	match := parseGameResponseFromLeetify(game)

	if match.Winner != 1 {
		t.Errorf("expected Winner 1, got %d", match.Winner)
	}
	if match.GameMode != "Premier" {
		t.Errorf("expected GameMode Premier, got %s", match.GameMode)
	}
	if match.OwnTeam.Score != 16 || match.EnemyTeam.Score != 10 {
		t.Errorf("expected scores 16-10, got %d-%d", match.OwnTeam.Score, match.EnemyTeam.Score)
	}
}

func TestParseGameResponseFromLeetify_LossAndModes(t *testing.T) {
	cases := []struct {
		dataSource string
		wantMode   string
	}{
		{"matchmaking_competitive", "Competitive"},
		{"matchmaking", "Premier"},
		{"faceit", "Faceit"},
		{"unknown_source", "unknown"},
	}

	for _, c := range cases {
		game := leetify.Game{DataSource: c.dataSource, MatchResult: "loss", Scores: []int{5, 16}}
		match := parseGameResponseFromLeetify(game)
		if match.GameMode != c.wantMode {
			t.Errorf("DataSource %q: expected mode %q, got %q", c.dataSource, c.wantMode, match.GameMode)
		}
		if match.Winner != 2 {
			t.Errorf("DataSource %q: expected Winner 2 (loss), got %d", c.dataSource, match.Winner)
		}
	}
}

func TestParseGameResponseFromLeetify_Tie(t *testing.T) {
	game := leetify.Game{MatchResult: "tie", Scores: []int{12, 12}}
	match := parseGameResponseFromLeetify(game)
	if match.Winner != 0 {
		t.Errorf("expected Winner 0 (tie), got %d", match.Winner)
	}
}

func TestParseMatchResultWithDetails_AssignsTeamsByTrackedPlayer(t *testing.T) {
	game := leetify.Game{GameId: "g1", MatchResult: "win", Scores: []int{16, 10}}

	details := &leetify.MatchDetailsResponse{
		PlayerStats: []leetify.PlayerStats{
			{Steam64ID: "1", InitialTeamNumber: 2, Name: "Tracked"},
			{Steam64ID: "2", InitialTeamNumber: 3, Name: "Opponent"},
		},
	}

	players := []config.Player{{SteamID: "1", Track: true}}

	result := ParseMatchResultWithDetails(game, details, nil, players)

	if len(result.OwnTeam.Players) != 1 || result.OwnTeam.Players[0].SteamID != "1" {
		t.Fatalf("expected tracked player 1 on own team, got %+v", result.OwnTeam.Players)
	}
	if len(result.EnemyTeam.Players) != 1 || result.EnemyTeam.Players[0].SteamID != "2" {
		t.Fatalf("expected player 2 on enemy team, got %+v", result.EnemyTeam.Players)
	}
	if len(result.OwnTeam.KnownPlayers) != 1 {
		t.Fatalf("expected 1 known player, got %d", len(result.OwnTeam.KnownPlayers))
	}
}

func TestParseMatchResultWithDetails_NilDetails(t *testing.T) {
	game := leetify.Game{GameId: "g1", MatchResult: "win", Scores: []int{16, 10}}
	result := ParseMatchResultWithDetails(game, nil, nil, nil)

	if len(result.OwnTeam.Players) != 0 || len(result.EnemyTeam.Players) != 0 {
		t.Errorf("expected no players when matchDetails is nil, got own=%d enemy=%d",
			len(result.OwnTeam.Players), len(result.EnemyTeam.Players))
	}
}

func TestSortPlayersByMates_KnownFirst(t *testing.T) {
	players := []Player{
		{SteamID: "unknown1"},
		{SteamID: "known1"},
		{SteamID: "unknown2"},
	}
	configPlayers := []config.Player{{SteamID: "known1"}}

	sorted := sortPlayersByMates(players, configPlayers)

	if sorted[0].SteamID != "known1" {
		t.Errorf("expected known1 first, got %s", sorted[0].SteamID)
	}
	if len(sorted) != 3 {
		t.Errorf("expected 3 players, got %d", len(sorted))
	}
}
