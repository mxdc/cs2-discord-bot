package parser

import (
	"sort"

	"github.com/mxdc/cs2-discord-bot/config"
)

type SessionWithDetails struct {
	Matches        []MatchWithDetails
	TrackedPlayers []config.Player
	IsFresh        bool
}

func (s *SessionWithDetails) BestRatioTeammate() Player {
	var bestRatioPlayer Player
	var bestRatio float64

	players := s.KnownPlayersWithCumulatedStats()
	for _, player := range players {
		ratio := player.KdRatio
		if ratio > bestRatio {
			bestRatio = ratio
			bestRatioPlayer = player
		}
	}

	return bestRatioPlayer

}

func (s *SessionWithDetails) WorstRatioTeammate() Player {
	var worstRatioPlayer Player
	var worstRatio float64 = 1e9

	players := s.KnownPlayersWithCumulatedStats()
	for _, player := range players {
		ratio := player.KdRatio
		if ratio < worstRatio {
			worstRatio = ratio
			worstRatioPlayer = player
		}
	}

	return worstRatioPlayer

}

func (s *SessionWithDetails) BestKillDeathTeammate() Player {
	var bestKiller Player

	players := s.KnownPlayersWithCumulatedStats()
	for _, player := range players {
		if player.Kills > bestKiller.Kills {
			bestKiller = player
		}
	}

	return bestKiller
}

func (s *SessionWithDetails) WorstKillDeathTeammate() Player {
	players := s.KnownPlayersWithCumulatedStats()
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

func (s *SessionWithDetails) KnownPlayersWithCumulatedStats() []Player {
	statsMap := make(map[string]*Player)

	// Matches are supposed to be in chronological order
	for _, match := range s.Matches {
		allMatchPlayers := append(match.OwnTeam.Players, match.EnemyTeam.Players...)

		for _, p := range allMatchPlayers {
			for _, tracked := range s.TrackedPlayers {
				if p.SteamID == tracked.SteamID {
					if playerSession, found := statsMap[p.SteamID]; found {
						// Update cumulative stats
						playerSession.Mvps += p.Mvps
						playerSession.Kills += p.Kills
						playerSession.Deaths += p.Deaths
						playerSession.TotalDamage += p.TotalDamage

						// Only update if rank is Premier
						if p.RankStats.RankType == 11 {
							playerSession.RankStats.Rank = p.RankStats.Rank
							playerSession.RankStats.Wins = p.RankStats.Wins
						}
					} else {
						// Initialize on first encountered match
						playerForSession := Player{
							SteamID:     p.SteamID,
							Name:        p.Name,
							CountryCode: p.CountryCode,
							Mvps:        p.Mvps,
							Kills:       p.Kills,
							Deaths:      p.Deaths,
							TotalDamage: p.TotalDamage,
						}
						if p.RankStats.RankType == 11 {
							playerForSession.RankStats = PlayerRankStats{
								Rank:     p.RankStats.Rank,
								Wins:     p.RankStats.Wins,
								OldRank:  p.RankStats.OldRank,
								RankType: p.RankStats.RankType,
							}
						}
						statsMap[p.SteamID] = &playerForSession
					}
				}
			}
		}
	}

	uniquePlayers := make([]Player, 0, len(statsMap))
	for _, session := range statsMap {
		finalPlayer := *session

		// Recalculate K/D Ratio for the entire session
		if finalPlayer.Deaths > 0 {
			finalPlayer.KdRatio = float64(finalPlayer.Kills) / float64(finalPlayer.Deaths)
		} else {
			finalPlayer.KdRatio = float64(finalPlayer.Kills)
		}

		// The rank has changed if the final rank is different from the very first OldRank
		finalPlayer.RankStats.RankChanged = (finalPlayer.RankStats.Rank != session.RankStats.OldRank)

		uniquePlayers = append(uniquePlayers, finalPlayer)
	}

	// Sort players by K/D ratio in descending order
	sort.Slice(uniquePlayers, func(i, j int) bool {
		return uniquePlayers[i].KdRatio > uniquePlayers[j].KdRatio
	})

	return uniquePlayers
}

func (s *SessionWithDetails) KnownPlayersSortedByKills() []Player {
	players := s.KnownPlayersWithCumulatedStats()
	sort.Slice(players, func(i, j int) bool {
		return players[i].Kills > players[j].Kills
	})
	return players
}

func (s *SessionWithDetails) KnownPlayersSortedByRank() []Player {
	players := s.KnownPlayersWithCumulatedStats()
	sort.Slice(players, func(i, j int) bool {
		return players[i].RankStats.Rank > players[j].RankStats.Rank
	})
	return players
}

// MatchResults holds the counted results of all matches in a session
type MatchResults struct {
	Victories int
	Defeats   int
	Ties      int
	Total     int
}

// countMatchResults efficiently counts all match results in a single pass
func (s *SessionWithDetails) countMatchResults() MatchResults {
	results := MatchResults{}

	for _, match := range s.Matches {
		results.Total++

		switch match.Winner {
		case 1:
			results.Victories++
		case 2:
			results.Defeats++
		default:
			results.Ties++
		}
	}

	return results
}

func (s *SessionWithDetails) AllMatchDefeats() bool {
	results := s.countMatchResults()
	return results.Total > 0 && results.Victories == 0 && results.Ties == 0
}

func (s *SessionWithDetails) AllMatchVictories() bool {
	results := s.countMatchResults()
	return results.Total > 0 && results.Defeats == 0 && results.Ties == 0
}

func (s *SessionWithDetails) MoreVictoriesThanDefeats() bool {
	results := s.countMatchResults()
	return results.Victories > results.Defeats
}

func (s *SessionWithDetails) MoreDefeatsThanVictories() bool {
	results := s.countMatchResults()
	return results.Defeats > results.Victories
}

func (s *SessionWithDetails) SortMatchesByEndTime() {
	sort.Slice(s.Matches, func(i, j int) bool {
		return s.Matches[i].GameFinishedAt.Before(s.Matches[j].GameFinishedAt)
	})
}

func (s *SessionWithDetails) AllMatchsInPremierMode() bool {
	for _, match := range s.Matches {
		if match.IsPremierMode() == false {
			return false
		}
	}

	return true
}
