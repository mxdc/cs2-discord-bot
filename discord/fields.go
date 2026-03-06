package discord

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mxdc/cs2-discord-bot/parser"
)

type EmbedFieldFormatter struct {
	fields []EmbedField
}

func NewEmbedFieldFormatter() *EmbedFieldFormatter {
	return &EmbedFieldFormatter{fields: []EmbedField{}}
}

func (f *EmbedFieldFormatter) GetFields() []EmbedField {
	return f.fields
}

func (f *EmbedFieldFormatter) addGameModeField(gameMode string) {
	if len(gameMode) == 0 {
		return
	}

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("*Mode* **%s**", gameMode),
		Inline: true,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addScoreField(match parser.MatchWithDetails) {
	if match.OwnTeam.Score == 0 && match.EnemyTeam.Score == 0 {
		return
	}

	score := fmt.Sprintf("%d - %d", match.OwnTeam.Score, match.EnemyTeam.Score)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("*Score* **%s**", score),
		Inline: true,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addMapNameField(mapName string) {
	if len(mapName) == 0 {
		return
	}

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("*Map* **%s**", mapName),
		Inline: true,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addPlayerMVPField(match parser.MatchWithDetails) {
	if len(match.OwnTeam.Players) == 0 || len(match.EnemyTeam.Players) == 0 {
		return
	}

	matchMVP := findMVP(match)
	playerLink := formatPlayerLink(matchMVP, true)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("⭐ *Match MVP*\u00A0\u00A0\u00A0\u00A0**%s**", playerLink),
		Inline: false,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addMatchLinkField(gameID string) {
	if len(gameID) == 0 {
		return
	}

	matchLink := fmt.Sprintf("https://leetify.com/public/match-details/%s/details-general", gameID)
	matchLinkLabel := fmt.Sprintf("▸ [View match details on Leetify](%s)", matchLink)

	field := EmbedField{
		Name:   "",
		Value:  matchLinkLabel,
		Inline: false,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addSessionMatchesField(matches []parser.MatchWithDetails) {
	if len(matches) == 0 {
		return
	}

	lines := make([]string, len(matches))
	for i, match := range matches {
		resultEmoji := getResultPrefixEmoji(match.Winner)
		matchLink := fmt.Sprintf("https://leetify.com/public/match-details/%s/details-general", match.GameID)
		matchResult := fmt.Sprintf(
			"%s · %d-%d · %s",
			match.GameMode,
			match.OwnTeam.Score,
			match.EnemyTeam.Score,
			match.MapName,
		)
		matchResultWithLink := fmt.Sprintf("%s [**%s**](%s)", resultEmoji, matchResult, matchLink)
		lines[i] = matchResultWithLink
	}

	field := EmbedField{
		Name:   "",
		Value:  strings.Join(lines, "\n"),
		Inline: false,
	}
	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addSessionTeammatesField(session parser.SessionWithDetails) {
	best := session.BestKillDeathTeammate()
	worst := session.WorstKillDeathTeammate()

	bestPlayerLink := formatPlayerLink(best, false)
	worstPlayerLink := formatPlayerLink(worst, false)

	bestTeammateKey := ":star: *Best Player*"
	bestTeammateValue := fmt.Sprintf("**%s** · **%d**K/**%d**D", bestPlayerLink, best.Kills, best.Deaths)
	bestTeammateStr := fmt.Sprintf("%s\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0%s", bestTeammateKey, bestTeammateValue)

	worstTeammateKey := ":poop: *Worst Player*"
	worstTeammateValue := fmt.Sprintf("**%s** · **%d**K/**%d**D", worstPlayerLink, worst.Kills, worst.Deaths)
	worstTeammateStr := fmt.Sprintf("%s\u00A0\u00A0\u00A0\u00A0%s", worstTeammateKey, worstTeammateValue)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("%s\n%s", bestTeammateStr, worstTeammateStr),
		Inline: false,
	}
	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addSessionCumulatedScoresField(session parser.SessionWithDetails) {
	posW, nameW, killsW, deathsW := computeColumnWidths(session.KnownPlayers())
	players := session.KnownPlayersSortedByKills()
	lines := make([]string, len(players)+2)

	lines[0] = fmt.Sprintf(
		"%-*s  %-*s  %*s  %*s",
		posW, "🏅",
		nameW, "👤",
		killsW-2, "🔫",
		deathsW-1, "💀",
	)

	// 6 spaces between columns
	total := posW + nameW + killsW + deathsW + 6
	lines[1] = fmt.Sprintf("-%*s", total, strings.Repeat("-", total))

	for i, p := range players {
		pos := "🔰"
		if i == 0 {
			pos = "🥇"
		} else if i == 1 {
			pos = "🥈"
		} else if i == 2 {
			pos = "🥉"
		}

		lines[i+2] = fmt.Sprintf(
			"%-*s  %-*s  %*d  %*d",
			posW, pos,
			nameW, p.Name,
			killsW, p.Kills,
			deathsW, p.Deaths,
		)
	}

	headerStr := "*Cumulated Scores*"

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("%s\n```%s```", headerStr, strings.Join(lines, "\n")),
		Inline: false,
	}
	f.fields = append(f.fields, field)
}

func computeColumnWidths(players []parser.Player) (int, int, int, int) {
	posW := 1
	nameW := 1
	killsW := 1
	deathsW := 1

	for _, p := range players {
		posW = max(posW, len(players)%2)
		nameW = max(nameW, len(p.Name))
		killsW = max(killsW, len(strconv.Itoa(p.Kills)))
		deathsW = max(deathsW, len(strconv.Itoa(p.Deaths)))
	}

	return posW, nameW, killsW, deathsW
}

func findMVP(match parser.MatchWithDetails) parser.Player {
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
