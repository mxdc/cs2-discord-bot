package parser

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type PlayerRankStats struct {
	Rank        int
	OldRank     int
	RankType    int // 11 for Premier Rank, 12 for Classic Matchmaking
	RankChanged bool
	Wins        int
}

type Player struct {
	SteamID     string
	Name        string
	CountryCode string
	Mvps        int
	Kills       int
	Deaths      int
	KdRatio     float64
	TotalDamage int
	RankStats   PlayerRankStats
}

func (p *Player) GetRecentPremierRank() (int, int) {
	rankStats := p.RankStats
	newRank := 0
	oldRank := 0

	if rankStats.RankType == 11 && rankStats.RankChanged && rankStats.Rank > 0 {
		newRank = rankStats.Rank
		oldRank = rankStats.OldRank

	}

	return oldRank, newRank
}

func (p *Player) FormatPlayerLink(withFlag, asTitle bool) string {
	var playerName string

	if asTitle {
		playerName = p.FormatPlayerTitle()
	} else {
		playerName = p.Name
	}

	playerNameWithLink := fmt.Sprintf("[%s](https://leetify.com/public/profile/%s)", playerName, p.SteamID)

	if withFlag && p.CountryCode != "" {
		flag := CountryCodeToFlag(p.CountryCode)
		playerNameWithLink = fmt.Sprintf("%s %s", flag, playerNameWithLink)
	}

	return playerNameWithLink
}

func (p *Player) FormatPlayerTitle() string {
	return cases.Title(language.English).String(strings.ToLower(p.Name))
}

type Team struct {
	Score        int
	Players      []Player
	KnownPlayers []Player
}

type MatchWithDetails struct {
	GameID         string
	GameMode       string
	GameFinishedAt time.Time
	MapName        string
	OwnTeam        Team
	EnemyTeam      Team
	Winner         int
}

func (m *MatchWithDetails) Defeat() bool {
	return m.Winner == 2
}

func (m *MatchWithDetails) Victory() bool {
	return m.Winner == 1
}

func (m *MatchWithDetails) Tie() bool {
	return m.Winner == 0
}

func (m *MatchWithDetails) IsPremierMode() bool {
	return m.GameMode == "Premier"
}

func (m *MatchWithDetails) GetMatchLink() string {
	return fmt.Sprintf("https://leetify.com/public/match-details/%s/details-general", m.GameID)
}

func (m *MatchWithDetails) GetOneLinerResult() string {
	return fmt.Sprintf(
		"%s · %d-%d · %s",
		m.GameMode,
		m.OwnTeam.Score,
		m.EnemyTeam.Score,
		m.MapName,
	)
}

type MatchResult struct {
	GameID         string
	GameFinishedAt time.Time
	MapName        string
	OwnTeam        Team
	EnemyTeam      Team
	Winner         int
	GameMode       string
}
