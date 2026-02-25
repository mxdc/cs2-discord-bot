package parser

import (
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

type Match struct {
	GameID         string
	GameMode       string
	GameFinishedAt time.Time
	MapName        string
	OwnTeam        Team
	EnemyTeam      Team
	Winner         int
}

func ParseMatchDetails(
	game *leetify.MatchDetailsResponse,
	steamPlayers []steam.SteamPlayer,
	players []config.Player,
) Match {
	gameTime, _ := time.Parse(time.RFC3339, game.FinishedAt)

	mode := "unknown"
	if game.DataSource == "matchmaking_competitive" {
		mode = "Competitive"
	} else if game.DataSource == "matchmaking" {
		mode = "Premier"
	} else if game.DataSource == "faceit" {
		mode = "Faceit"
	}

	ownTeamPlayers, enemyTeamPlayers := getTeamsFromPlayerStats(game.PlayerStats, steamPlayers, players)

	match := Match{
		GameID:         game.ID,
		GameFinishedAt: gameTime,
		MapName:        game.MapName,
		OwnTeam: Team{
			Players: ownTeamPlayers,
		},
		EnemyTeam: Team{
			Players: enemyTeamPlayers,
		},
		GameMode: mode,
	}

	return match
}

func ParseMatchResult(
	matchSummary leetify.MatchResult,
	matchDetails *leetify.MatchDetailsResponse,
	steamPlayers []steam.SteamPlayer,
	players []config.Player,
) Match {
	match := Match{
		GameID:         matchSummary.GameID,
		GameMode:       matchSummary.GameMode,
		GameFinishedAt: matchSummary.GameFinishedAt,
		MapName:        matchSummary.MapName,
		OwnTeam: Team{
			Score:   matchSummary.OwnTeam.Score,
			Players: parsePlayers(matchSummary.OwnTeam.Players, matchDetails, steamPlayers, players),
		},
		EnemyTeam: Team{
			Score:   matchSummary.EnemyTeam.Score,
			Players: parsePlayers(matchSummary.EnemyTeam.Players, matchDetails, steamPlayers, []config.Player{}),
		},
		Winner: matchSummary.Winner,
	}

	return match
}

func parsePlayers(
	players []leetify.Player,
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
				updatedPlayer.Name = sp.PersonaName
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
					break
				}
			}
		}

		updatedPlayers = append(updatedPlayers, updatedPlayer)
	}

	return updatedPlayers
}

// sortPlayersByMates puts known players from config first, then unknown players
func sortPlayersByMates(players []leetify.Player, configPlayers []config.Player) []leetify.Player {
	// myClanPlayers will contain known players defined in config file
	var myClanPlayers []leetify.Player
	// unknownPlayers will contain players absent from config file
	var unknownPlayers []leetify.Player

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

func getTeamsFromPlayerStats(
	players []leetify.LeetifyPlayerStats,
	steamPlayers []steam.SteamPlayer,
	configPlayers []config.Player,
) ([]Player, []Player) {
	var ownTeamPlayers []Player
	var enemyTeamPlayers []Player

	for _, player := range players {
		parsedPlayer := Player{
			SteamID:     player.Steam64ID,
			Kills:       player.TotalKills,
			Deaths:      player.TotalDeaths,
			Mvps:        player.Mvps,
			KdRatio:     player.KdRatio,
			TotalDamage: player.TotalDamage,
		}

		// Update with steam data if available
		for _, sp := range steamPlayers {
			if sp.SteamID == parsedPlayer.SteamID {
				parsedPlayer.Name = sp.PersonaName
				parsedPlayer.CountryCode = sp.CountryCode
				break
			}
		}

		// Determine if player is on own team or enemy team based on config
		isOwnTeam := false
		for _, configPlayer := range configPlayers {
			if parsedPlayer.SteamID == configPlayer.SteamID {
				isOwnTeam = true
				break
			}
		}

		if isOwnTeam {
			ownTeamPlayers = append(ownTeamPlayers, parsedPlayer)
		} else {
			enemyTeamPlayers = append(enemyTeamPlayers, parsedPlayer)
		}
	}

	return ownTeamPlayers, enemyTeamPlayers
}
