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

	return knownPlayers
}
