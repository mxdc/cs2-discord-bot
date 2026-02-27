package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mxdc/cs2-discord-bot/parser"
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
	ColorBlue  = 3447003  // Information blue
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
func (c *WebhookClient) SendMatchResult(match parser.MatchWithDetails) {
	content := formatMatchHeader(match)
	embed := createMatchEmbed(match)
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

// SendSessionResult sends a Discord embed message for a session result
func (c *WebhookClient) SendSessionResult(session parser.SessionWithDetails) {
	if len(session.Matches) == 1 {
		c.SendMatchResult(session.Matches[0])
		return
	}

	sessionResultBuiler := NewSessionResultBuilder(session)
	message := sessionResultBuiler.BuildMessage()
	log.Println("Discord: Sending Discord notification...")

	if err := c.sendWebhook(message); err != nil {
		log.Printf("Discord: Error sending Discord webhook: %v", err)
	} else {
		log.Println("Discord: Discord notification sent successfully")
	}
}

func formatMatchHeader(match parser.MatchWithDetails) string {
	if match.OwnTeam.Score == 0 && match.EnemyTeam.Score == 0 {
		return "A match has finished."
	}

	players := match.OwnTeam.Players

	resultEmoji := getResultPrefixEmoji(match.Winner)

	// Build the player names string
	names := formatPlayerNamesAsTitle(players)
	header := fmt.Sprintf("%s %s", resultEmoji, names)

	if match.Winner == 1 {
		return fmt.Sprintf("%s won the match!", header)
	}
	if match.Winner == 2 {
		return fmt.Sprintf("%s lost the match.", header)
	}

	return fmt.Sprintf("%s finished in a tie.", header)
}

func createMatchEmbed(match parser.MatchWithDetails) Embed {
	var color int

	if match.Winner == 1 {
		color = ColorGreen
	} else if match.Winner == 2 {
		color = ColorRed
	} else {
		color = ColorGray
	}

	fieldsFormatter := NewEmbedFieldFormatter()
	fieldsFormatter.addGameModeField(match.GameMode)
	fieldsFormatter.addScoreField(match)
	fieldsFormatter.addMapNameField(match.MapName)
	fieldsFormatter.addPlayerMVPField(match)
	fieldsFormatter.addMatchLinkField(match.GameID)

	formattedFields := fieldsFormatter.GetFields()

	embed := Embed{
		Title:  "",
		Color:  color,
		Fields: formattedFields,
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
