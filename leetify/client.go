package leetify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func (c *LeetifyClient) GetPlayerMatches(playerConfig config.Player) ([]LeetifyGameResponse, error) {
	url := fmt.Sprintf("%s/api/profile/id/%s", c.BaseURL, playerConfig.SteamID)
	if playerConfig.AccountName != "" {
		url = fmt.Sprintf("%s/api/profile/vanity-url/%s", c.BaseURL, playerConfig.AccountName)
	}
	log.Printf("Leetify: Fetching matches from %s\n", url)

	var emptyGames []LeetifyGameResponse

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return emptyGames, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Origin", "https://leetify.com")
	req.Header.Set("Referer", "https://leetify.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return emptyGames, fmt.Errorf("Failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return emptyGames, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var profileResp ProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profileResp); err != nil {
		return emptyGames, fmt.Errorf("Failed to decode response: %w", err)
	}

	// Parse games into MatchResult format using existing function
	// matches := parseStatsFromLeetify(profileResp.Games)
	return profileResp.Games, nil
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
