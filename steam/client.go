package steam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// PlayerSummaryResponse represents the Steam API response structure
type PlayerSummaryResponse struct {
	Response struct {
		Players []struct {
			SteamID        string `json:"steamid"`
			PersonaName    string `json:"personaname"`
			LocCountryCode string `json:"loccountrycode"`
		} `json:"players"`
	} `json:"response"`
}

// SteamPlayer represents a player's Steam ID, country code and persona name
type SteamPlayer struct {
	SteamID     string
	CountryCode string
	PersonaName string
}

// SteamPlayers is a collection of steam players with lookup methods
type SteamPlayers []SteamPlayer

// Client wraps the Steam Web API client
type Client struct {
	apiKey string
}

// NewSteamClient creates a new Steam API client
func NewSteamClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// GetSteamPlayers gets country codes and persona names for multiple players
func (c *Client) GetSteamPlayers(steamIDs []string) (SteamPlayers, error) {
	if len(steamIDs) == 0 {
		return SteamPlayers{}, nil
	}

	// Filter out invalid IDs
	var validIDs []string
	for _, steamID := range steamIDs {
		if _, err := strconv.ParseUint(steamID, 10, 64); err == nil {
			validIDs = append(validIDs, steamID)
		}
	}

	if len(validIDs) == 0 {
		return SteamPlayers{}, nil
	}

	// Initialize result with empty data
	var result SteamPlayers
	for _, steamID := range validIDs {
		result = append(result, SteamPlayer{
			SteamID: steamID,
			// Empty until we get API response
			CountryCode: "",
			PersonaName: "",
		})
	}

	// Call Steam API with comma-separated Steam IDs
	steamIDsStr := strings.Join(validIDs, ",")
	url := fmt.Sprintf(
		"https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s",
		c.apiKey,
		steamIDsStr,
	)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Steam: Call failed: %v\n", err)
		return result, nil
	}
	defer resp.Body.Close()

	var data PlayerSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Steam: Failed to decode response: %v\n", err)
		return result, nil
	}

	// Update result with actual data from API response
	for _, apiPlayer := range data.Response.Players {
		for i := range result {
			if result[i].SteamID == apiPlayer.SteamID {
				result[i].CountryCode = apiPlayer.LocCountryCode
				result[i].PersonaName = apiPlayer.PersonaName
				break
			}
		}
	}

	return result, nil
}
