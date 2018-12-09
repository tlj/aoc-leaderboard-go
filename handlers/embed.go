package handlers

import (
	"github.com/bradfitz/iter"
	"github.com/tlj/aoc-leaderboard-go/leaderboard"
	"github.com/tlj/aoc-leaderboard-go/member_score"
	"html/template"
	"log"
	"net/http"
	"sort"
)

func Embed(w http.ResponseWriter, r *http.Request) {
	maxDay := int(leaderboard.CurrentBoard.MaxDay)

	var totalMemberScores []*member_score.MemberScore
	var dailyMemberScores []*member_score.MemberScore

	for _, memberScore := range leaderboard.CurrentBoard.Totals {
		totalMemberScores = append(totalMemberScores, memberScore)
	}

	for _, memberScore := range leaderboard.CurrentBoard.Days[maxDay].MemberScores {
		dailyMemberScores = append(dailyMemberScores, memberScore)
	}

	sort.Sort(member_score.ByWTime(totalMemberScores))
	sort.Sort(member_score.ByWTime(dailyMemberScores))

	if len(dailyMemberScores) > 10 {
		dailyMemberScores = dailyMemberScores[:10]
	}

	if len(totalMemberScores) > 10 {
		totalMemberScores = totalMemberScores[:10]
	}

	topScores := leaderboard.CurrentBoard.TopScores
	if len(topScores) > 10 {
		topScores = topScores[:10]
	}

	type DayScores map[string]interface{}
	type Context map[string]interface{}
	c := Context{
		"day": leaderboard.CurrentBoard.Days[maxDay].Day,
		"year": leaderboard.CurrentBoard.Days[maxDay].Year,
		"dayScores": dailyMemberScores,
		"totalScores": totalMemberScores,
		"topScores": topScores,
	}

	funcMap := template.FuncMap{
		"readableTime": leaderboard.ReadableTime,
		"N": iter.N,
	}

	tmpl := template.Must(template.New("embed.html").Funcs(funcMap).ParseGlob("templates/*.html"))
	err := tmpl.ExecuteTemplate(w, "embed.html", c)
	if err != nil {
		log.Printf("Error executin template: %v", err)
	}
}

