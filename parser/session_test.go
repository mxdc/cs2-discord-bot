package parser

import (
	"testing"

	"github.com/mxdc/cs2-discord-bot/config"
)

func makeMatch(winner int, trackedKills, trackedDeaths int) MatchWithDetails {
	return MatchWithDetails{
		Winner: winner,
		OwnTeam: Team{
			Players: []Player{
				{SteamID: "tracked", Kills: trackedKills, Deaths: trackedDeaths},
			},
		},
		EnemyTeam: Team{},
	}
}

func TestKnownPlayersWithCumulatedStats_SumsAcrossMatches(t *testing.T) {
	session := SessionWithDetails{
		TrackedPlayers: []config.Player{{SteamID: "tracked"}},
		Matches: []MatchWithDetails{
			makeMatch(1, 10, 5),
			makeMatch(2, 15, 8),
		},
	}

	players := session.KnownPlayersWithCumulatedStats()

	if len(players) != 1 {
		t.Fatalf("expected 1 known player, got %d", len(players))
	}
	if players[0].Kills != 25 || players[0].Deaths != 13 {
		t.Errorf("expected cumulated 25 kills / 13 deaths, got %d/%d", players[0].Kills, players[0].Deaths)
	}
}

func TestAllMatchDefeats(t *testing.T) {
	session := SessionWithDetails{Matches: []MatchWithDetails{{Winner: 2}, {Winner: 2}}}
	if !session.AllMatchDefeats() {
		t.Error("expected AllMatchDefeats true")
	}
}

func TestAllMatchVictories(t *testing.T) {
	session := SessionWithDetails{Matches: []MatchWithDetails{{Winner: 1}, {Winner: 1}}}
	if !session.AllMatchVictories() {
		t.Error("expected AllMatchVictories true")
	}
}

func TestMoreVictoriesThanDefeats(t *testing.T) {
	session := SessionWithDetails{Matches: []MatchWithDetails{{Winner: 1}, {Winner: 1}, {Winner: 2}}}
	if !session.MoreVictoriesThanDefeats() {
		t.Error("expected MoreVictoriesThanDefeats true")
	}
	if session.MoreDefeatsThanVictories() {
		t.Error("expected MoreDefeatsThanVictories false")
	}
}

func TestMoreDefeatsThanVictories(t *testing.T) {
	session := SessionWithDetails{Matches: []MatchWithDetails{{Winner: 2}, {Winner: 2}, {Winner: 1}}}
	if !session.MoreDefeatsThanVictories() {
		t.Error("expected MoreDefeatsThanVictories true")
	}
}
