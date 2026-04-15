package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/mistral"
	"github.com/mxdc/cs2-discord-bot/parser"
)

type WebhookClient struct {
	webhookURL    string
	mistralClient *mistral.MistralClient
	httpClient    *http.Client
	translations  locales.Translations
	withRank      bool
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

func NewWebhookClient(
	webhookURL string,
	mistralClient *mistral.MistralClient,
	translations locales.Translations,
	withRank bool,
) *WebhookClient {
	return &WebhookClient{
		translations:  translations,
		webhookURL:    webhookURL,
		mistralClient: mistralClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		withRank: withRank,
	}
}

func (c *WebhookClient) SendMatchResult(match parser.MatchWithDetails) {
	message := NewMatchResultBuilder(match, c.translations, c.withRank).BuildMessage()
	if c.mistralClient != nil {
		result := c.mistralClient.GetGeneratedTitles(message.Content)
		message.Content = result
	}

	log.Println("Discord: Sending Discord notification...")

	if err := c.sendWebhook(message); err != nil {
		log.Printf("Discord: Error sending Discord webhook: %v", err)
	} else {
		log.Println("Discord: Discord notification sent successfully")
	}
}

func (c *WebhookClient) SendSessionResult(session parser.SessionWithDetails) {
	if len(session.Matches) == 1 {
		c.SendMatchResult(session.Matches[0])
		return
	}

	withRank := c.withRank && session.IsFresh
	sessionResultBuiler := NewSessionResultBuilder(session, c.translations, withRank)
	message := sessionResultBuiler.BuildMessage()
	if c.mistralClient != nil {
		result := c.mistralClient.GetGeneratedTitles(message.Content)
		message.Content = result
	}

	log.Println("Discord: Sending Discord notification...")

	if err := c.sendWebhook(message); err != nil {
		log.Printf("Discord: Error sending Discord webhook: %v", err)
	} else {
		log.Println("Discord: Discord notification sent successfully")
	}
}

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
