package locales

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Translations struct {
	Lang                       string `yaml:"lang"`
	BotUsername                string `yaml:"bot_username"`
	ListSeparator              string `yaml:"list_separator"`
	ListConjunction            string `yaml:"list_conjunction"`
	MatchFinished              string `yaml:"match_finished"`
	WinSingle                  string `yaml:"win_single"`
	WinSingleRank              string `yaml:"win_single_rank"`
	LossSingle                 string `yaml:"loss_single"`
	LossSingleRank             string `yaml:"loss_single_rank"`
	TieSingle                  string `yaml:"tie_single"`
	WinMultiple                string `yaml:"win_multiple"`
	LossMultiple               string `yaml:"loss_multiple"`
	TieMultiple                string `yaml:"tie_multiple"`
	SessionAllLosses           string `yaml:"session_all_losses"`
	SessionAllWins             string `yaml:"session_all_wins"`
	SessionSingleAllWins       string `yaml:"session_single_all_wins"`
	SessionMoreWins            string `yaml:"session_more_wins"`
	SessionMoreLosses          string `yaml:"session_more_losses"`
	SessionTie                 string `yaml:"session_tie"`
	SessionSingleTie           string `yaml:"session_single_tie"`
	SessionSingleAllLossesRank string `yaml:"session_single_all_losses_rank"`
	SessionSingleAllWinsRank   string `yaml:"session_single_all_wins_rank"`
}

type TranslationConfigFile struct {
	Keys []Translations `yaml:"keys"`
}

func MustLoadTranslations(path, lang string) Translations {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read translation file at: %s", path)
	}

	// Unmarshal the YAML into our global map
	var translations TranslationConfigFile
	err = yaml.Unmarshal(data, &translations)
	if err != nil {
		log.Fatalf("Unable to parse translation file at: %s", path)
	}

	// find the translations for the specified language
	for _, t := range translations.Keys {
		if t.Lang == lang {
			return t
		}
	}

	log.Fatalf("No translations found for language: %s", lang)
	return Translations{}
}
