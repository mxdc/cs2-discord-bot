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
	client    *leetify.LeetifyClient
	player    config.Player
	out       chan<- session.MatchDetected
	debugMode bool
}

func NewCrawler(
	client *leetify.LeetifyClient,
	player config.Player,
	out chan<- session.MatchDetected,
	debugMode bool,
) *Crawler {
	return &Crawler{
		client:    client,
		player:    player,
		out:       out,
		debugMode: debugMode,
	}
}

func (c *Crawler) StartCrawling() {
	log.Printf("%s: Crawler started", c.player.PlayerID())
	before, err := c.client.GetPlayerMatches(c.player)
	if err != nil {
		log.Fatalf("%s: Error: %v", c.player.PlayerID(), err)
	}
	// In debug mode, we want to process the last 5 matches on startup
	// to test the notifier without having to play new matches
	if c.debugMode {
		log.Printf("%s: Debug mode enabled", c.player.PlayerID())
		if len(before.Games) > 5 {
			before.Games = before.Games[5:]
		}
	}
	log.Printf("%s: %d previous matches", c.player.PlayerID(), len(before.Games))
	if !c.debugMode {
		time.Sleep(10 * time.Minute)
	}

	for {
		after, err := c.client.GetPlayerMatches(c.player)
		if err != nil {
			log.Printf("%s: Error: %v", c.player.PlayerID(), err)
			time.Sleep(15 * time.Minute)
			continue
		}

		if len(after.Games) == 0 {
			log.Printf("%s: No matches found, retrying in 1 hour", c.player.PlayerID())
			time.Sleep(1 * time.Hour)
			continue
		}

		// Check for new matches
		newMatches := findNewMatches(before.Games, after.Games)
		for _, match := range newMatches {
			log.Printf("%s: New match found: %s", c.player.PlayerID(), match.GameId)
			c.out <- session.MatchDetected{Match: match, Player: c.player, DetectedAt: time.Now()}
		}

		before = after

		if len(newMatches) > 0 {
			log.Printf("%s: found %d new", c.player.PlayerID(), len(newMatches))
		}

		time.Sleep(1 * time.Hour)
	}
}

// findNewMatches returns matches that are in current but not in previous
func findNewMatches(previous, current []leetify.LeetifyGameResponse) []leetify.LeetifyGameResponse {
	prevSet := make(map[string]bool)
	for _, match := range previous {
		prevSet[match.GameId] = true
	}

	var newMatches []leetify.LeetifyGameResponse
	for _, match := range current {
		if !prevSet[match.GameId] {
			newMatches = append(newMatches, match)
		}
	}

	// Sort newMatches by GameFinishedAt field, from oldest to newest
	sort.Slice(newMatches, func(i, j int) bool {
		currentTimeI, errI := time.Parse(time.RFC3339, newMatches[i].GameFinishedAt)
		currentTimeJ, errJ := time.Parse(time.RFC3339, newMatches[j].GameFinishedAt)
		if errI != nil || errJ != nil {
			return false
		}
		return currentTimeI.Before(currentTimeJ)
	})

	return newMatches
}
