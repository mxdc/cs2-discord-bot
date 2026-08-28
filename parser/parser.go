package parser

import (
	"slices"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/steam"
)

func parseGameResponseFromLeetify(game leetify.Game) MatchResult {
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
		GameID:         game.GameId,
		GameFinishedAt: gameTime,
		MapName:        game.MapName,
		GameMode:       mode,
	}

	// Determine winner based on match result from match history
	// scores[0] = own team, scores[1] = enemy team
	var ownTeamScore, enemyTeamScore int
	if len(game.Scores) >= 2 {
		ownTeamScore = game.Scores[0]
		enemyTeamScore = game.Scores[1]
	}

	switch game.MatchResult {
	case "win":
		match.Winner = 1 // Own team won
	case "loss":
		match.Winner = 2 // Enemy team won
	case "tie":
		match.Winner = 0 // Tie
	default:
		match.Winner = 0 // Unknown
	}

	// Initialize empty teams - will be populated from match details
	match.OwnTeam = Team{
		Score:   ownTeamScore,
		Players: []Player{},
	}
	match.EnemyTeam = Team{
		Score:   enemyTeamScore,
		Players: []Player{},
	}

	return match
}

func ParseMatchResultWithDetails(
	game leetify.Game,
	matchDetails *leetify.MatchDetailsResponse,
	steamPlayers []steam.SteamPlayer,
	players []config.Player,
) MatchWithDetails {
	match := parseGameResponseFromLeetify(game)

	// Find tracked player's team number from match details
	trackedPlayerTeamNumber := 0
	if matchDetails != nil {
		for _, ps := range matchDetails.PlayerStats {
			for _, configPlayer := range players {
				if ps.Steam64ID == configPlayer.SteamID {
					trackedPlayerTeamNumber = ps.InitialTeamNumber
					break
				}
			}
			if trackedPlayerTeamNumber != 0 {
				break
			}
		}
	}

	// Group players by team number
	ownTeamPlayers := []Player{}
	enemyTeamPlayers := []Player{}

	if matchDetails != nil {
		for _, ps := range matchDetails.PlayerStats {
			player := Player{SteamID: ps.Steam64ID}

			if ps.InitialTeamNumber == trackedPlayerTeamNumber {
				ownTeamPlayers = append(ownTeamPlayers, player)
			} else {
				enemyTeamPlayers = append(enemyTeamPlayers, player)
			}
		}
	}

	matchWithDetails := MatchWithDetails{
		GameID:         match.GameID,
		GameMode:       match.GameMode,
		GameFinishedAt: match.GameFinishedAt,
		MapName:        match.MapName,
		OwnTeam: Team{
			Score:        match.OwnTeam.Score,
			Players:      parsePlayers(ownTeamPlayers, matchDetails, steamPlayers, players),
			KnownPlayers: []Player{},
		},
		EnemyTeam: Team{
			Score:        match.EnemyTeam.Score,
			Players:      parsePlayers(enemyTeamPlayers, matchDetails, steamPlayers, []config.Player{}),
			KnownPlayers: []Player{},
		},
		Winner: match.Winner,
	}

	ownTeamKnownPlayers := parseKnownPlayers(matchWithDetails.OwnTeam.Players, players)
	matchWithDetails.OwnTeam.KnownPlayers = ownTeamKnownPlayers

	return matchWithDetails
}

func parseKnownPlayers(players []Player, configPlayers []config.Player) []Player {
	var knownPlayers []Player

	for _, player := range players {
		for _, configPlayer := range configPlayers {
			if player.SteamID == configPlayer.SteamID {
				knownPlayers = append(knownPlayers, player)
				break
			}
		}
	}

	// Sort known players by kills in descending order
	slices.SortFunc(knownPlayers, func(a, b Player) int {
		return b.Kills - a.Kills
	})

	return knownPlayers
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

		// Update with Steam data if available
		for _, sp := range steamPlayers {
			if sp.SteamID == player.SteamID {
				updatedPlayer.Name = sp.PersonaName
				updatedPlayer.CountryCode = sp.CountryCode
				break
			}
		}

		// Update with match details data if available
		if matchDetails != nil {
			// Find the player stats and update
			for _, p := range matchDetails.PlayerStats {
				name := updatedPlayer.Name
				if len(name) == 0 {
					name = p.Name
				}

				if p.Steam64ID == updatedPlayer.SteamID {
					updatedPlayer.Kills = p.TotalKills
					updatedPlayer.Deaths = p.TotalDeaths
					updatedPlayer.Mvps = p.Mvps
					updatedPlayer.KdRatio = p.KdRatio
					updatedPlayer.TotalDamage = p.TotalDamage
					updatedPlayer.Name = name
					break
				}
			}

			// Find the player rank and update
			for _, p := range matchDetails.MatchmakingGameStats {
				if p.SteamID == updatedPlayer.SteamID {
					updatedPlayer.RankStats = PlayerRankStats{
						Rank:        p.Rank,
						OldRank:     p.OldRank,
						RankType:    p.RankType,
						RankChanged: p.RankChanged,
						Wins:        p.Wins,
					}
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
