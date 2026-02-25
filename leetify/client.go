package leetify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
)

type LeetifyClient struct {
	httpClient *http.Client
	BaseURL    string
}

func NewLeetifyClient(baseURL string) *LeetifyClient {
	return &LeetifyClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: baseURL,
	}
}

type Player struct {
	SteamID string
}

type Team struct {
	Score   int
	Players []Player
}

type MatchResult struct {
	GameID              string    `json:"gameId"`
	OwnTeamSteam64Ids   []string  `json:"ownTeamSteam64Ids"`
	EnemyTeamSteam64Ids []string  `json:"enemyTeamSteam64Ids"`
	DataSource          string    `json:"dataSource"`
	GameFinishedAt      time.Time `json:"gameFinishedAt"`
	IsCs2               bool      `json:"isCs2"`
	MapName             string    `json:"mapName"`
	MatchResult         string    `json:"matchResult"`
	RankType            int       `json:"rankType"`
	Scores              []int     `json:"scores"`
	// Computed fields for compatibility
	OwnTeam   Team
	EnemyTeam Team
	Winner    int
	GameMode  string
}

type LeetifyGameResponse struct {
	EnemyTeamSteam64Ids []string `json:"enemyTeamSteam64Ids"`
	OwnTeamSteam64Ids   []string `json:"ownTeamSteam64Ids"`
	DataSource          string   `json:"dataSource"`
	GameFinishedAt      string   `json:"gameFinishedAt"`
	GameId              string   `json:"gameId"`
	IsCs2               bool     `json:"isCs2"`
	MapName             string   `json:"mapName"`
	MatchResult         string   `json:"matchResult"`
	RankType            int      `json:"rankType"`
	Scores              []int    `json:"scores"`
}

type ProfileResponse struct {
	Games []LeetifyGameResponse `json:"games"`
}

func (c *LeetifyClient) GetPlayerMatches(playerConfig config.Player) ([]MatchResult, error) {
	url := fmt.Sprintf("%s/api/profile/id/%s", c.BaseURL, playerConfig.SteamID)
	if playerConfig.AccountName != "" {
		url = fmt.Sprintf("%s/api/profile/vanity-url/%s", c.BaseURL, playerConfig.AccountName)
	}
	log.Printf("Leetify: Fetching matches from %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Origin", "https://leetify.com")
	req.Header.Set("Referer", "https://leetify.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var profileResp ProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profileResp); err != nil {
		return nil, fmt.Errorf("Failed to decode response: %w", err)
	}

	// Parse games into MatchResult format using existing function
	matches := parseStatsFromLeetify(profileResp.Games)
	return matches, nil
}

type MatchDetailsResponse struct {
	PlayerStats    []LeetifyPlayerStats `json:"playerStats"`
	SteamShareCode string               `json:"steamShareCode"`
	ID             string               `json:"id"`
	DataSource     string               `json:"dataSource"`
	FinishedAt     string               `json:"gameFinishedAt"`
	IsCs2          bool                 `json:"isCs2"`
	MapName        string               `json:"mapName"`
	TeamScores     []int                `json:"teamScores"`
}

type LeetifyPlayerStats struct {
	ID                string    `json:"id"`
	GameID            string    `json:"gameId"`
	GameFinishedAt    time.Time `json:"gameFinishedAt"`
	Steam64ID         string    `json:"steam64Id"`
	Name              string    `json:"name"`
	Score             int       `json:"score"`
	InitialTeamNumber int       `json:"initialTeamNumber"`
	Mvps              int       `json:"mvps"`
	TotalKills        int       `json:"totalKills"`
	TotalDeaths       int       `json:"totalDeaths"`
	KdRatio           float64   `json:"kdRatio"`
	TotalDamage       int       `json:"totalDamage"`
}

func (c *LeetifyClient) GetMatchDetails(gameID string) (*MatchDetailsResponse, error) {
	url := fmt.Sprintf("%s/api/games/%s", c.BaseURL, gameID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", "https://leetify.com")
	req.Header.Set("Referer", "https://leetify.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("match details request failed with status: %d", resp.StatusCode)
	}

	var details MatchDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode match details response: %w", err)
	}

	return &details, nil
}

func parseStatsFromLeetify(games []LeetifyGameResponse) []MatchResult {
	var matches []MatchResult

	for _, game := range games {
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

		matches = append(matches, match)
	}

	return matches
}
