package crawler

import (
	"log"
	"sort"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/session"
)

type Crawler struct {
	client *leetify.LeetifyClient
	player config.Player
	out    chan<- session.MatchDetected
}

func NewCrawler(
	client *leetify.LeetifyClient,
	player config.Player,
	out chan<- session.MatchDetected,
) *Crawler {
	return &Crawler{
		client: client,
		player: player,
		out:    out,
	}
}

func (c *Crawler) StartCrawling() {
	log.Printf("%s: Crawler started", c.player.AccountName)
	lastMatches, err := c.client.GetPlayerMatches(c.player)
	if err != nil {
		log.Fatalf("%s: Error: %v", c.player.AccountName, err)
	}
	log.Printf("%s: %d previous matches", c.player.AccountName, len(lastMatches))
	time.Sleep(2 * time.Minute)
	// lastMatches := []leetify.MatchResult{}

	for {
		matches, err := c.client.GetPlayerMatches(c.player)
		if err != nil {
			log.Printf("%s: Error: %v", c.player.AccountName, err)
			time.Sleep(1 * time.Minute)
			continue
		}

		if len(matches) == 0 {
			log.Printf("%s: No matches found, retrying in 2 minutes", c.player.AccountName)
			time.Sleep(2 * time.Minute)
			continue
		}

		// Check for new matches
		newMatches := findNewMatches(lastMatches, matches)
		for _, match := range newMatches {
			log.Printf("%s: New match found: %s", c.player.AccountName, match.GameID)
			c.out <- session.MatchDetected{Match: match, Player: c.player}
		}

		lastMatches = matches
		log.Printf("%s: Checked %d matches, found %d new", c.player.AccountName, len(matches), len(newMatches))
		time.Sleep(2 * time.Minute)
	}
}

// findNewMatches returns matches that are in current but not in previous
func findNewMatches(previous, current []leetify.MatchResult) []leetify.MatchResult {
	prevSet := make(map[string]bool)
	for _, match := range previous {
		prevSet[match.GameID] = true
	}

	var newMatches []leetify.MatchResult
	for _, match := range current {
		if !prevSet[match.GameID] {
			newMatches = append(newMatches, match)
		}
	}

	// Sort newMatches by GameFinishedAt field, from oldest to newest
	sort.Slice(newMatches, func(i, j int) bool {
		return newMatches[i].GameFinishedAt.Before(newMatches[j].GameFinishedAt)
	})

	return newMatches
}
