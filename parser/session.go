package parser

import (
	"github.com/mxdc/cs2-discord-bot/config"
)

type SessionWithDetails struct {
	Matches        []MatchWithDetails
	TrackedPlayers []config.Player
}

func (s *SessionWithDetails) BestKillDeathTeammate() Player {
	var bestKiller Player

	players := s.KnownPlayers()
	for _, player := range players {
		if player.Kills > bestKiller.Kills {
			bestKiller = player
		}
	}

	// compute the total kills and deaths for the best killer across all matches
	totalKills := 0
	totalDeaths := 0
	for _, match := range s.Matches {
		for _, p := range match.OwnTeam.Players {
			if p.SteamID == bestKiller.SteamID {
				totalKills += p.Kills
				totalDeaths += p.Deaths
			}
		}
	}

	bestKiller.Kills = totalKills
	bestKiller.Deaths = totalDeaths

	return bestKiller
}

func (s *SessionWithDetails) WorstKillDeathTeammate() Player {
	players := s.KnownPlayers()
	if len(players) == 0 {
		return Player{}
	}

	worstKiller := players[0]
	for _, player := range players {
		if player.Kills < worstKiller.Kills {
			worstKiller = player
		}
	}

	// compute the total kills and deaths for the worst killer across all matches
	totalKills := 0
	totalDeaths := 0
	for _, match := range s.Matches {
		for _, p := range match.OwnTeam.Players {
			if p.SteamID == worstKiller.SteamID {
				totalKills += p.Kills
				totalDeaths += p.Deaths
			}
		}
	}

	worstKiller.Kills = totalKills
	worstKiller.Deaths = totalDeaths

	return worstKiller
}

func (s *SessionWithDetails) KnownPlayers() []Player {
	knownPlayers := []Player{}

	for _, match := range s.Matches {
		for _, player := range match.OwnTeam.Players {
			for _, trackedPlayer := range s.TrackedPlayers {
				if player.SteamID == trackedPlayer.SteamID {
					knownPlayers = append(knownPlayers, player)
				}
			}
		}
	}

	uniquePlayers := []Player{}

	for _, player := range knownPlayers {
		duplicate := false
		for _, uniquePlayer := range uniquePlayers {
			if player.SteamID == uniquePlayer.SteamID {
				duplicate = true
				break
			}
		}
		if !duplicate {
			uniquePlayers = append(uniquePlayers, player)
		}
	}

	return uniquePlayers
}
