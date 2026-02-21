package main

import (
	"flag"
	"log"
	"sort"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/discord"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/parser"
	"github.com/mxdc/cs2-discord-bot/steam"
)

type MatchDetected struct {
	Match  leetify.MatchResult
	Player config.Player
}

func matchManager(
	cfg *config.AppConfig,
	client *leetify.LeetifyClient,
	in <-chan MatchDetected,
) {
	seen := make(map[string]bool)
	discordClient := discord.NewWebhookClient(cfg.DiscordHook)
	steamClient := steam.NewSteamClient(cfg.SteamAPIKey)
	log.Println("Manager: Started manager, waiting for matches...")

	for msg := range in {
		if seen[msg.Match.GameID] {
			continue
		}

		seen[msg.Match.GameID] = true
		log.Println("Manager: New match detected:", msg.Match.GameID)

		time.Sleep(5 * time.Second)

		// Get all Steam IDs from both teams
		allSteamIDs := append(msg.Match.OwnTeamSteam64Ids, msg.Match.EnemyTeamSteam64Ids...)

		// Get Steam player data (names and countries)
		steamPlayers, err := steamClient.GetSteamPlayers(allSteamIDs)
		if err != nil {
			log.Printf("Manager: Warning: failed to get steam players: %v", err)
			// Continue without steam data
			steamPlayers = steam.SteamPlayers{}
		}

		details, err := client.GetMatchDetails(msg.Match.GameID)
		if err != nil {
			log.Printf("Manager: Warning: failed to get match details: %v", err)
			// Continue without match details
			details = nil
		}

		match := parser.ParseMatchResult(msg.Match, details, steamPlayers, cfg.Players)

		// Send Discord webhook
		discordClient.SendMatchResult(match)
	}
}

func crawler(
	client *leetify.LeetifyClient,
	player config.Player,
	out chan<- MatchDetected,
) {
	log.Printf("%s: Crawler started", player.AccountName)
	lastMatches, err := client.GetPlayerMatches(player)
	if err != nil {
		log.Fatalf("%s: Error: %v", player.AccountName, err)
	}
	log.Printf("%s: %d previous matches", player.AccountName, len(lastMatches))
	time.Sleep(2 * time.Minute)
	// lastMatches := []leetify.MatchResult{}

	for {
		matches, err := client.GetPlayerMatches(player)
		if err != nil {
			log.Printf("%s: Error: %v", player.AccountName, err)
			time.Sleep(1 * time.Minute)
			continue
		}

		if len(matches) == 0 {
			log.Printf("%s: No matches found, retrying in 2 minutes", player.AccountName)
			time.Sleep(2 * time.Minute)
			continue
		}

		// Check for new matches
		newMatches := findNewMatches(lastMatches, matches)
		for _, match := range newMatches {
			log.Printf("%s: New match found: %s", player.AccountName, match.GameID)
			out <- MatchDetected{Match: match, Player: player}
		}

		lastMatches = matches
		log.Printf("%s: Checked %d matches, found %d new", player.AccountName, len(matches), len(newMatches))
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

func getTrackedPlayers(players []config.Player) []config.Player {
	tracked := []config.Player{}

	for _, player := range players {
		if player.Track {
			tracked = append(tracked, player)
		}
	}
	return tracked
}

func main() {
	configFile := flag.String("config.file", "config.yml", "Path to the configuration file")
	flag.Parse()

	cfg := config.MustLoadConfig(*configFile)
	client := leetify.NewLeetifyClient(cfg.LeetifyAPIURL)

	// Channels
	matchChan := make(chan MatchDetected, 1024)

	log.Println("CS2: Starting crawler")

	// Start manager
	go matchManager(cfg, client, matchChan)

	// Start crawler for each tracked player
	trackedPlayers := getTrackedPlayers(cfg.Players)
	for i, player := range trackedPlayers {
		go crawler(client, player, matchChan)
		if i < len(trackedPlayers)-1 {
			time.Sleep(20 * time.Second)
		}
	}

	log.Println("CS2: Crawler started")
	log.Printf("CS2: Tracking matches for %d player(s)", len(trackedPlayers))
	log.Printf("CS2: Discord webhook configured: %t", cfg.DiscordHook != "")
	select {} // block forever
}
