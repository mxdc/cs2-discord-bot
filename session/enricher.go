package session

import (
	"log"
	"os"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/parser"
	"github.com/mxdc/cs2-discord-bot/steam"
)

type MatchEnricher struct {
	leetifyClient  *leetify.Client
	steamClient    *steam.Client
	trackedPlayers []config.Player
	log            *log.Logger
}

func NewMatchEnricher(
	leetifyClient *leetify.Client,
	steamClient *steam.Client,
	trackedPlayers []config.Player,
) *MatchEnricher {
	return &MatchEnricher{
		leetifyClient:  leetifyClient,
		steamClient:    steamClient,
		trackedPlayers: trackedPlayers,
		log:            log.New(os.Stderr, "MatchEnricher: ", log.LstdFlags),
	}
}

func (e *MatchEnricher) Enrich(game leetify.Game) parser.MatchWithDetails {
	details := e.fetchDetails(game.GameId)

	return parser.ParseMatchResultWithDetails(game, details, nil, e.trackedPlayers)
}

func (e *MatchEnricher) EnrichWithProfiles(game leetify.Game) parser.MatchWithDetails {
	details := e.fetchDetails(game.GameId)
	profiles := e.fetchProfiles(details)

	return parser.ParseMatchResultWithDetails(game, details, profiles, e.trackedPlayers)
}

func (e *MatchEnricher) fetchDetails(gameID string) *leetify.MatchDetailsResponse {
	details, err := e.leetifyClient.GetMatchDetails(gameID)
	if err != nil {
		e.log.Printf("Warning: failed to get match details: %v", err)
		return nil
	}

	return details
}

func (e *MatchEnricher) fetchProfiles(details *leetify.MatchDetailsResponse) []steam.SteamPlayer {
	if details == nil {
		return nil
	}

	var steamIDs []string
	for _, ps := range details.PlayerStats {
		steamIDs = append(steamIDs, ps.Steam64ID)
	}

	profiles, err := e.steamClient.GetSteamPlayers(steamIDs)
	if err != nil {
		e.log.Printf("Warning: failed to get steam players: %v", err)
		return nil
	}

	return profiles
}
