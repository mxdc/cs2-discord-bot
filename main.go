package main

import (
	"flag"
	"log"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/crawler"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/session"
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

func main() {
	configFile := flag.String("config.file", "config.yml", "Path to the configuration file")
	sessionMode := flag.Bool("session", false, "Enable session mode (groups matches into sessions)")
	flag.Parse()

	cfg := config.MustLoadConfig(*configFile)
	client := leetify.NewLeetifyClient(cfg.LeetifyAPIURL)

	log.Println("CS2: Starting crawler")

	if *sessionMode {
		startSessionNotifier(cfg, client)
	} else {
		startMatchNotifier(cfg, client)
	}

	log.Printf("CS2: Discord webhook configured: %t", cfg.DiscordHook != "")
	select {} // block forever
}
