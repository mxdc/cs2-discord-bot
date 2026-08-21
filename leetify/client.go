package leetify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
)

type LeetifyClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewLeetifyClient(baseURL string) *LeetifyClient {
	return &LeetifyClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

type LeetifyGameResponse struct {
	DataSource          string  `json:"dataSource"`
	Deaths              int     `json:"deaths"`
	GameFinishedAt      string  `json:"gameFinishedAt"`
	GameId              string  `json:"gameId"`
	Kills               int     `json:"kills"`
	LeetifyRating       float64 `json:"leetifyRating"`
	MapName             string  `json:"mapName"`
	MatchmakingRankType int     `json:"matchmakingRankType"`
	MatchResult         string  `json:"matchResult"`
	Rank                int     `json:"rank"`
	Scores              []int   `json:"scores"`
}

type ProfileResponse struct {
	Games []LeetifyGameResponse `json:"games"`
}

func (c *LeetifyClient) GetPlayerMatches(playerConfig config.Player) (ProfileResponse, error) {
	u := c.getUrlForPlayer(playerConfig)

	log.Printf("Leetify: Fetching matches from %s\n", u.Path)

	var emptyProfile ProfileResponse

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return emptyProfile, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Origin", "https://leetify.com")
	req.Header.Set("Referer", "https://leetify.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return emptyProfile, fmt.Errorf("Failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return emptyProfile, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var profileResp ProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profileResp); err != nil {
		return emptyProfile, fmt.Errorf("Failed to decode response: %w", err)
	}

	return profileResp, nil
}

type MatchDetailsResponse struct {
	PlayerStats          []LeetifyPlayerStats `json:"playerStats"`
	SteamShareCode       string               `json:"steamShareCode"`
	ID                   string               `json:"id"`
	DataSource           string               `json:"dataSource"`
	FinishedAt           string               `json:"gameFinishedAt"`
	IsCs2                bool                 `json:"isCs2"`
	MapName              string               `json:"mapName"`
	TeamScores           []int                `json:"teamScores"`
	MatchmakingGameStats []struct {
		ID             string    `json:"id"`
		GameID         string    `json:"gameId"`
		SteamID        string    `json:"steam64Id"`
		GameFinishedAt time.Time `json:"gameFinishedAt"`
		Rank           int       `json:"rank"`
		OldRank        int       `json:"oldRank"`
		RankType       int       `json:"rankType"`
		RankChanged    bool      `json:"rankChanged"`
		Wins           int       `json:"wins"`
	} `json:"matchmakingGameStats"`
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
	u := c.getUrlForGameID(gameID)

	req, err := http.NewRequest("GET", u.String(), nil)
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

func (c *LeetifyClient) getUrlForPlayer(playerConfig config.Player) *url.URL {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		log.Fatalf("failed to parse base URL: %v", err)
	}

	u.Path = "/api/profile/" + playerConfig.SteamID + "/match-history"

	return u
}

func (c *LeetifyClient) getUrlForGameID(gameID string) *url.URL {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		log.Fatalf("failed to parse base URL: %v", err)
	}

	u.Path = "/api/games/" + gameID

	return u
}
