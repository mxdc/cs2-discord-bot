package parser

import (
	"slices"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/steam"
)

type Player struct {
	SteamID     string
	Name        string
	CountryCode string
	Mvps        int
	Kills       int
	Deaths      int
	KdRatio     float64
	TotalDamage int
}

type Team struct {
	Score   int
	Players []Player
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

type MatchResult struct {
	GameID              string
	OwnTeamSteam64Ids   []string
	EnemyTeamSteam64Ids []string
	DataSource          string
	GameFinishedAt      time.Time
	IsCs2               bool
	MapName             string
	MatchResult         string
	RankType            int
	Scores              []int
	// Computed fields for compatibility
	OwnTeam   Team
	EnemyTeam Team
	Winner    int
	GameMode  string
}

func parseGameResponseFromLeetify(game leetify.LeetifyGameResponse) MatchResult {
	gameTime, _ := time.Parse(time.RFC3339, game.GameFinishedAt)

	mode := "unknown"
	if game.DataSource == "matchmaking_competitive" {
		mode = "Competitive"
	} else if game.DataSource == "matchmaking" {
		mode = "Premier"
	} else if game.DataSource == "faceit" {
		mode = "Faceit"
	}

	match := MatchResult{
		GameID:              game.GameId,
		OwnTeamSteam64Ids:   game.OwnTeamSteam64Ids,
		EnemyTeamSteam64Ids: game.EnemyTeamSteam64Ids,
		DataSource:          game.DataSource,
		GameFinishedAt:      gameTime,
		IsCs2:               game.IsCs2,
		MapName:             game.MapName,
		MatchResult:         game.MatchResult,
		RankType:            game.RankType,
		Scores:              game.Scores,
		// Computed
		GameMode: mode,
	}

	// Create team structures based on Leetify's own/enemy team distinction
	var ownTeamPlayers []Player
	for _, steamID := range game.OwnTeamSteam64Ids {
		ownTeamPlayers = append(ownTeamPlayers, Player{
			SteamID: steamID,
		})
	}

	var enemyTeamPlayers []Player
	for _, steamID := range game.EnemyTeamSteam64Ids {
		enemyTeamPlayers = append(enemyTeamPlayers, Player{
			SteamID: steamID,
		})
	}

	// Determine winner based on match result from Leetify
	var ownTeamScore, enemyTeamScore int
	switch game.MatchResult {
	case "win":
		match.Winner = 1 // Own team won
		// Own team has higher score, enemy team has lower score
		ownTeamScore = slices.Max(game.Scores)
		enemyTeamScore = slices.Min(game.Scores)
	case "loss":
		match.Winner = 2 // Enemy team won
		// Enemy team has higher score, own team has lower score
		enemyTeamScore = slices.Max(game.Scores)
		ownTeamScore = slices.Min(game.Scores)
	default:
		match.Winner = 0 // tie or unknown
		// Assign in array order since scores are equal or unknown
		ownTeamScore, enemyTeamScore = game.Scores[0], game.Scores[1]
	}

	match.OwnTeam = Team{
		Score:   ownTeamScore,
		Players: ownTeamPlayers,
	}
	match.EnemyTeam = Team{
		Score:   enemyTeamScore,
		Players: enemyTeamPlayers,
	}

	return match
}

func ParseMatchResultWithDetails(
	game leetify.LeetifyGameResponse,
	matchDetails *leetify.MatchDetailsResponse,
	steamPlayers []steam.SteamPlayer,
	players []config.Player,
) MatchWithDetails {
	match := parseGameResponseFromLeetify(game)

	matchWithDetails := MatchWithDetails{
		GameID:         match.GameID,
		GameMode:       match.GameMode,
		GameFinishedAt: match.GameFinishedAt,
		MapName:        match.MapName,
		OwnTeam: Team{
			Score:   match.OwnTeam.Score,
			Players: parsePlayers(match.OwnTeam.Players, matchDetails, steamPlayers, players),
		},
		EnemyTeam: Team{
			Score:   match.EnemyTeam.Score,
			Players: parsePlayers(match.EnemyTeam.Players, matchDetails, steamPlayers, []config.Player{}),
		},
		Winner: match.Winner,
	}

	return matchWithDetails
}

func parsePlayers(
	players []Player,
	matchDetails *leetify.MatchDetailsResponse,
	steamPlayers []steam.SteamPlayer,
	configPlayers []config.Player,
) []Player {
	updatedPlayers := []Player{}
	sortedPlayers := sortPlayersByMates(players, configPlayers)

	for _, player := range sortedPlayers {
		updatedPlayer := Player{SteamID: player.SteamID}

		// Update with steam data if available
		for _, sp := range steamPlayers {
			if sp.SteamID == player.SteamID {
				// updatedPlayer.Name = sp.PersonaName
				updatedPlayer.CountryCode = sp.CountryCode
				break
			}
		}

		// Update with match details data if available
		if matchDetails != nil {
			for _, p := range matchDetails.PlayerStats {
				if p.Steam64ID == updatedPlayer.SteamID {
					updatedPlayer.Kills = p.TotalKills
					updatedPlayer.Deaths = p.TotalDeaths
					updatedPlayer.Mvps = p.Mvps
					updatedPlayer.KdRatio = p.KdRatio
					updatedPlayer.TotalDamage = p.TotalDamage
					updatedPlayer.Name = p.Name
					break
				}
			}
		}

		updatedPlayers = append(updatedPlayers, updatedPlayer)
	}

	return updatedPlayers
}

// sortPlayersByMates puts known players from config first, then unknown players
func sortPlayersByMates(players []Player, configPlayers []config.Player) []Player {
	// myClanPlayers will contain known players defined in config file
	var myClanPlayers []Player
	// unknownPlayers will contain players absent from config file
	var unknownPlayers []Player

	for _, player := range players {
		isKnown := false
		for _, configPlayer := range configPlayers {
			if player.SteamID == configPlayer.SteamID {
				isKnown = true
				break
			}
		}

		if isKnown {
			myClanPlayers = append(myClanPlayers, player)
		} else {
			unknownPlayers = append(unknownPlayers, player)
		}
	}

	// Return known players first, then unknown players
	return append(myClanPlayers, unknownPlayers...)
}
