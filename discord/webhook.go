package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mxdc/cs2-discord-bot/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type WebhookClient struct {
	webhookURL string
	httpClient *http.Client
}

type Embed struct {
	Title  string       `json:"title"`
	Color  int          `json:"color"`
	Fields []EmbedField `json:"fields"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type WebhookMessage struct {
	Content   string  `json:"content"`
	TTS       bool    `json:"tts"`
	Embeds    []Embed `json:"embeds"`
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url,omitempty"`
}

// Colors for Discord embeds
const (
	ColorGreen = 3066993  // Victory green
	ColorRed   = 15158332 // Defeat red
	ColorGray  = 9807270  // Neutral gray
)

func NewWebhookClient(webhookURL string) *WebhookClient {
	return &WebhookClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendMatchResult sends a Discord embed message for a match result
func (c *WebhookClient) SendMatchResult(match parser.Match) {
	content := c.formatHeader(match)
	embed := c.createMatchEmbed(match)
	message := WebhookMessage{
		Content:  content,
		TTS:      false,
		Embeds:   []Embed{embed},
		Username: "CS2",
	}

	log.Println("Discord: Sending Discord notification...")

	if err := c.sendWebhook(message); err != nil {
		log.Printf("Discord: Error sending Discord webhook: %v", err)
	} else {
		log.Println("Discord: Discord notification sent successfully")
	}
}

func (c *WebhookClient) formatHeader(match parser.Match) string {
	players := match.OwnTeam.Players

	var header string
	if match.Winner == 1 {
		header += "🏆 "
	} else if match.Winner == 2 {
		header += "💀 "
	} else {
		header += "🤝 "
	}

	// Build the player names string
	for i, player := range players {
		playerName := cases.Title(language.English).String(strings.ToLower(player.Name))
		header += playerName
		if i < len(players)-2 {
			header += ", "
		} else if i < len(players)-1 {
			header += " and "
		}
	}

	if match.Winner == 1 {
		return fmt.Sprintf("%s won the match!", header)
	}
	if match.Winner == 2 {
		return fmt.Sprintf("%s lost the match.", header)
	}

	return fmt.Sprintf("%s finished in a tie.", header)
}

// createMatchEmbed creates a simplified Discord embed for the match result
func (c *WebhookClient) createMatchEmbed(match parser.Match) Embed {
	var color int

	if match.Winner == 1 {
		color = ColorGreen
	} else if match.Winner == 2 {
		color = ColorRed
	} else {
		color = ColorGray
	}

	// Format scores
	score := fmt.Sprintf("%d - %d", match.OwnTeam.Score, match.EnemyTeam.Score)

	// Create match link using Leetify format
	matchLink := fmt.Sprintf("▸ [View match details on Leetify](https://leetify.com/public/match-details/%s/details-general)", match.GameID)
	matchMVP := findMVP(match)
	playerLink := formatPlayerLink(matchMVP)

	// team1 := formatTeamMembers(match.OwnTeam)
	// team2 := formatTeamMembers(match.EnemyTeam)

	embed := Embed{
		Title: "",
		Color: color,
		Fields: []EmbedField{
			{
				Name:   "",
				Value:  fmt.Sprintf("*Mode* **%s**", match.GameMode),
				Inline: true,
			},
			{
				Name:   "",
				Value:  fmt.Sprintf("*Score* **%s**", score),
				Inline: true,
			},
			{
				Name:   "",
				Value:  fmt.Sprintf("*Map* **%s**", match.MapName),
				Inline: true,
			},
			// {
			// 	Name:   "",
			// 	Value:  fmt.Sprintf("*Your Team*\n%s", team1),
			// 	Inline: true,
			// },
			// {
			// 	Name:   "",
			// 	Value:  fmt.Sprintf("*Opponents*\n%s", team2),
			// 	Inline: true,
			// },
			{
				Name:   "",
				Value:  fmt.Sprintf("⭐ *Match MVP*\u00A0\u00A0\u00A0\u00A0**%s**", playerLink),
				Inline: true,
			},
			// {
			// 	Name:   "",
			// 	Value:  formatPlayerLink(matchMVP),
			// 	Inline: true,
			// },
			{
				Name:   "",
				Value:  matchLink,
				Inline: false,
			},
		},
	}

	return embed
}

// sendWebhook sends the webhook message to Discord
func (c *WebhookClient) sendWebhook(message WebhookMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook message: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// func formatTeamMembers(team parser.Team) string {
// 	var names string

// 	for i, player := range team.Players {
// 		playerName := cases.Title(language.English).String(strings.ToLower(player.Name))
// 		playerWithLink := fmt.Sprintf("[%s](https://leetify.com/public/profile/%s)", playerName, player.SteamID)
// 		playerFlag := CountryCodeToFlag(player.CountryCode)
// 		playerFull := fmt.Sprintf("%s %s", playerFlag, playerWithLink)
// 		names += playerFull
// 		if i < len(team.Players)-1 {
// 			names += "\n"
// 		}
// 	}

// 	return names
// }

func formatPlayerLink(player parser.Player) string {
	playerName := fmt.Sprintf("[%s](https://leetify.com/public/profile/%s)", player.Name, player.SteamID)

	if player.CountryCode != "" {
		flag := CountryCodeToFlag(player.CountryCode)
		playerName = fmt.Sprintf("%s %s", flag, playerName)
	}

	return playerName
}

func findMVP(match parser.Match) parser.Player {
	var mvp parser.Player

	for _, player := range match.OwnTeam.Players {
		if player.Mvps > mvp.Mvps || (player.Mvps == mvp.Mvps && player.Kills > mvp.Kills) {
			mvp = player
		}
	}
	for _, player := range match.EnemyTeam.Players {
		if player.Mvps > mvp.Mvps || (player.Mvps == mvp.Mvps && player.Kills > mvp.Kills) {
			mvp = player
		}
	}

	return mvp
}
