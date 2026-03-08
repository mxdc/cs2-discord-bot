package discord

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mxdc/cs2-discord-bot/parser"
)

var positionEmoji = map[int]string{
	0: "🥇",
	1: "🥈",
	2: "🥉",
}

type EmbedFieldFormatter struct {
	fields []EmbedField
}

func NewEmbedFieldFormatter() *EmbedFieldFormatter {
	return &EmbedFieldFormatter{fields: []EmbedField{}}
}

func (f *EmbedFieldFormatter) GetFields() []EmbedField {
	return f.fields
}

func (f *EmbedFieldFormatter) addMatchOneLinerField(match parser.MatchWithDetails) {
	matchLink := match.GetMatchLink()
	matchResult := match.GetOneLinerResult()
	resultEmoji := getResultPrefixEmoji(match.Winner)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("%s [**%s**](%s)", resultEmoji, matchResult, matchLink),
		Inline: false,
	}

	f.fields = append(f.fields, field)
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
	if matchMVP.IsNameInvisible() {
		return
	}

	playerLink := matchMVP.FormatPlayerLink(true, false)

	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("⭐ *Match MVP*\u00A0\u00A0\u00A0\u00A0**%s**", playerLink),
		Inline: false,
	}

	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addMatchLinkField(match parser.MatchWithDetails) {
	if len(match.GameID) == 0 {
		return
	}

	matchLink := match.GetMatchLink()
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
		matchLink := match.GetMatchLink()
		matchResult := match.GetOneLinerResult()
		matchResultWithLink := fmt.Sprintf("`%s` [**%s**](%s)", resultEmoji, matchResult, matchLink)
		lines[i] = matchResultWithLink
	}

	field := EmbedField{
		Name:   "",
		Value:  strings.Join(lines, "\n"),
		Inline: false,
	}
	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addSessionTeammatesField(session parser.SessionWithDetails, withWorst bool) {
	best := session.BestKillDeathTeammate()
	worst := session.WorstKillDeathTeammate()

	bestPlayerLink := best.FormatPlayerLink(false, false)
	worstPlayerLink := worst.FormatPlayerLink(false, false)

	bestTeammateKey := ":star: *Best Buddy*"
	bestTeammateValue := fmt.Sprintf("**%s** · **%d**K/**%d**D", bestPlayerLink, best.Kills, best.Deaths)
	bestTeammateStr := fmt.Sprintf("%s\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0%s", bestTeammateKey, bestTeammateValue)

	worstTeammateKey := ":poop: *Worst Buddy*"
	worstTeammateValue := fmt.Sprintf("**%s** · **%d**K/**%d**D", worstPlayerLink, worst.Kills, worst.Deaths)
	worstTeammateStr := fmt.Sprintf("%s\u00A0\u00A0\u00A0\u00A0%s", worstTeammateKey, worstTeammateValue)

	var toDisplay string
	if withWorst {
		toDisplay = fmt.Sprintf("%s\n%s", bestTeammateStr, worstTeammateStr)
	} else {
		toDisplay = bestTeammateStr
	}

	field := EmbedField{
		Name:   "",
		Value:  toDisplay,
		Inline: false,
	}
	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addSessionRankUpdate(session parser.SessionWithDetails) {
	players := session.KnownPlayersSortedByRank()
	_, nameW, _, _, rankW := computeColumnWidths(players)

	var lines []string
	for _, p := range players {
		// 11 for Premier Rank, 12 for Classic Matchmaking
		if p.RankStats.RankType != 11 || p.RankStats.Rank == 0 {
			continue
		}

		medal := "🔰"
		currPos := len(lines) + 1
		if currPos < 4 {
			medal = positionEmoji[currPos-1]
		}
		playerLink := p.FormatPlayerLink(false, false)
		lines = append(lines, fmt.Sprintf(
			"%s `%*d` **%-*s**",
			medal,
			rankW, p.RankStats.Rank,
			nameW, playerLink,
		))
	}

	if len(lines) == 0 {
		return
	}

	headerStr := "*New MMRs*"
	field := EmbedField{
		Name:   "",
		Value:  fmt.Sprintf("%s\n%s", headerStr, strings.Join(lines, "\n")),
		Inline: false,
	}
	f.fields = append(f.fields, field)
}

func (f *EmbedFieldFormatter) addSessionCumulatedScoresField(session parser.SessionWithDetails) {
	players := session.KnownPlayersSortedByKills()
	posW, nameW, killsW, deathsW, _ := computeColumnWidths(players)
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
		medal := "🔰"
		if i < 3 {
			medal = positionEmoji[i]
		}

		lines[i+2] = fmt.Sprintf(
			"%-*s  %-*s  %*d  %*d",
			posW, medal,
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

func computeColumnWidths(players []parser.Player) (int, int, int, int, int) {
	posW := 1
	nameW := 1
	killsW := 1
	deathsW := 1
	rankW := 1

	for _, p := range players {
		posW = max(posW, len(players)%2)
		nameW = max(nameW, len(p.Name))
		killsW = max(killsW, len(strconv.Itoa(p.Kills)))
		deathsW = max(deathsW, len(strconv.Itoa(p.Deaths)))
		rankW = max(rankW, len(strconv.Itoa(p.RankStats.Rank)))
	}

	return posW, nameW, killsW, deathsW, rankW
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
