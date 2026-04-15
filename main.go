package main

import (
	"flag"
	"log"
	"time"

	"github.com/mxdc/cs2-discord-bot/config"
	"github.com/mxdc/cs2-discord-bot/crawler"
	"github.com/mxdc/cs2-discord-bot/leetify"
	"github.com/mxdc/cs2-discord-bot/locales"
	"github.com/mxdc/cs2-discord-bot/mistral"
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

func startMatchNotifier(
	cfg *config.AppConfig,
	client *leetify.LeetifyClient,
	mistralClient *mistral.MistralClient,
	translations locales.Translations,
	debugMode bool,
) {
	log.Printf("CS2: Running in match mode with lang: %s", cfg.Lang)

	matchChan := make(chan session.MatchDetected, 1024)

	matchNotifier := session.NewMatchNotifier(cfg, client, mistralClient, translations, matchChan)
	go matchNotifier.HandleMatch()

	startCrawlers(client, cfg, matchChan, debugMode)
}

func startSessionNotifier(
	cfg *config.AppConfig,
	client *leetify.LeetifyClient,
	mistralClient *mistral.MistralClient,
	translations locales.Translations,
	withRank bool,
	debugMode bool,
) {
	log.Printf("CS2: Running in session mode with lang: %s", cfg.Lang)

	matchChan := make(chan session.MatchDetected, 1024)
	sessionChan := make(chan session.GameSession, 256)

	sessionMgr := session.NewSessionManager(matchChan, sessionChan, debugMode)
	go sessionMgr.HandleIncomingMatches()

	sessionNotifier := session.NewSessionNotifier(cfg, client, mistralClient, translations, sessionChan, withRank)
	go sessionNotifier.HandleSession()

	startCrawlers(client, cfg, matchChan, debugMode)
}

func startCrawlers(client *leetify.LeetifyClient, cfg *config.AppConfig, matchChan chan<- session.MatchDetected, debugMode bool) {
	log.Println("CS2: Starting crawler")

	trackedPlayers := getTrackedPlayers(cfg.Players)
	for i, player := range trackedPlayers {
		crawler := crawler.NewCrawler(client, player, matchChan, debugMode)
		go crawler.StartCrawling()
		if i < len(trackedPlayers)-1 {
			time.Sleep(5 * time.Minute)
		}
	}

	log.Println("CS2: Crawler started")
	log.Printf("CS2: Tracking matches for %d player(s)", len(trackedPlayers))
}

func main() {
	configFile := flag.String("config.file", "config.yml", "Path to the configuration file")
	sessionMode := flag.Bool("session", false, "Enable session mode (groups matches into sessions)")
	debugMode := flag.Bool("debug", false, "Enable debug mode")
	withAi := flag.Bool("with.ai", false, "Enable AI mode")
	withRank := flag.Bool("with.rank", false, "Display new rank after each match")
	promptFilePath := flag.String("prompt.file", "prompts/system.md", "Path to the system prompt file")
	translationFilePath := flag.String("translation.file", "translations.yml", "Path to the translation file")
	flag.Parse()

	cfg := config.MustLoadConfig(*configFile)
	translations := locales.MustLoadTranslations(*translationFilePath, cfg.Lang)
	client := leetify.NewLeetifyClient(cfg.LeetifyAPIURL)

	var mistralClient *mistral.MistralClient
	if *withAi {
		mistralClient = mistral.NewMistralClient(cfg.MistralAPIKey, *promptFilePath)
	}

	if *sessionMode {
		startSessionNotifier(cfg, client, mistralClient, translations, *withRank, *debugMode)
	} else {
		startMatchNotifier(cfg, client, mistralClient, translations, *debugMode)
	}

	log.Printf("CS2: Discord webhook configured: %t", cfg.DiscordHook != "")

	select {} // block forever
}
