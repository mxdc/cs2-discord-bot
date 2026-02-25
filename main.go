package main

import (
	"flag"
	"log"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/crawler"
	"github.com/mxdc/cs2-discord-bot/discord"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/parser"
	"github.com/mxdc/cs2-discord-bot/session"
	"github.com/mxdc/cs2-discord-bot/steam"
)

func getTrackedPlayers(players []config.Player) []config.Player {
	tracked := []config.Player{}

	for _, player := range players {
		if player.Track {
			tracked = append(tracked, player)
		}
	}
	return tracked
}

func startMatchNotifier(cfg *config.AppConfig, client *leetify.LeetifyClient) {
	log.Println("CS2: Running in match mode")

	matchChan := make(chan session.MatchDetected, 1024)

	matchNotifier := session.NewMatchNotifier(cfg, client, matchChan)
	go matchNotifier.HandleMatch()

	startCrawlers(client, cfg, matchChan)
}

func startSessionNotifier(cfg *config.AppConfig, client *leetify.LeetifyClient) {
	log.Println("CS2: Running in session mode")

	matchChan := make(chan session.MatchDetected, 1024)
	sessionChan := make(chan session.GameSession, 256)

	sessionMgr := session.NewSessionManager(cfg, client, matchChan, sessionChan)
	go sessionMgr.HandleIncomingMatches()

	sessionNotifier := session.NewSessionNotifier(cfg, sessionChan)
	go sessionNotifier.HandleSession()

	startCrawlers(client, cfg, matchChan)
}

func startCrawlers(client *leetify.LeetifyClient, cfg *config.AppConfig, matchChan chan<- session.MatchDetected) {
	trackedPlayers := getTrackedPlayers(cfg.Players)
	for i, player := range trackedPlayers {
		crawler := crawler.NewCrawler(client, player, matchChan)
		go crawler.StartCrawling()
		if i < len(trackedPlayers)-1 {
			time.Sleep(20 * time.Second)
		}
	}

	log.Println("CS2: Crawler started")
	log.Printf("CS2: Tracking matches for %d player(s)", len(trackedPlayers))
}

func notifyMatch(cfg *config.AppConfig, client *leetify.LeetifyClient, matchId string) {
	discordClient := discord.NewWebhookClient(cfg.DiscordHook)
	steamClient := steam.NewSteamClient(cfg.SteamAPIKey)

	details, err := client.GetMatchDetails(matchId)
	if err != nil {
		log.Fatalf("Manager: Warning: failed to get match details: %v", err)
	}

	allSteamIDs := []string{}
	for _, pl := range details.PlayerStats {
		allSteamIDs = append(allSteamIDs, pl.Steam64ID)
	}

	// Get Steam player data (names and countries)
	steamPlayers, err := steamClient.GetSteamPlayers(allSteamIDs)
	if err != nil {
		log.Printf("Manager: Warning: failed to get steam players: %v", err)
		// Continue without steam data
		steamPlayers = steam.SteamPlayers{}
	}

	match := parser.ParseMatchDetails(details, steamPlayers, cfg.Players)

	// Send Discord webhook
	discordClient.SendMatchResult(match)

}

func main() {
	configFile := flag.String("config.file", "config.yml", "Path to the configuration file")
	sessionMode := flag.Bool("session", false, "Enable session mode (groups matches into sessions)")
	oneshotId := flag.String("oneshot", "", "Enable one-shot mode by processing one match once and exit")
	flag.Parse()

	cfg := config.MustLoadConfig(*configFile)
	client := leetify.NewLeetifyClient(cfg.LeetifyAPIURL)

	log.Println("CS2: Starting crawler")

	if *oneshotId != "" {
		notifyMatch(cfg, client, *oneshotId)
	} else if *sessionMode {
		startSessionNotifier(cfg, client)
	} else {
		startMatchNotifier(cfg, client)
	}

	log.Printf("CS2: Discord webhook configured: %t", cfg.DiscordHook != "")
	select {} // block forever
}
