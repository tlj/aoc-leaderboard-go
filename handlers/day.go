package handlers

import (
	"github.com/bradfitz/iter"
	"github.com/gorilla/mux"
	"github.com/tlj/aoc-leaderboard-go/leaderboard"
	"github.com/tlj/aoc-leaderboard-go/member_score"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
)

func Day(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var day int64
	var err error

	_, ok := vars["day"]
	if !ok {
		day = leaderboard.CurrentBoard.MaxDay
	} else {
		day, err = strconv.ParseInt(vars["day"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var memberScores []*member_score.MemberScore
	if day == 0 {
		for _, memberScore := range leaderboard.CurrentBoard.Totals {
			memberScores = append(memberScores, memberScore)
		}
	} else {
		for _, memberScore := range leaderboard.CurrentBoard.Days[int(day)].MemberScores {
			memberScores = append(memberScores, memberScore)
		}
	}

	orderBy, ok := vars["orderBy"]
	if !ok {
		orderBy = "part2diff"
	}

	if orderBy == "part2diff" {
		sort.Sort(member_score.ByPart2Diff(memberScores))
	} else if orderBy == "part1" {
		sort.Sort(member_score.ByPart1(memberScores))
	} else if orderBy == "part2" {
		sort.Sort(member_score.ByPart2(memberScores))
	} else if orderBy == "ogscore" {
		sort.Sort(member_score.ByAocGlobalScore(memberScores))
	} else if orderBy == "olscore" {
		sort.Sort(member_score.ByAocLocalScore(memberScores))
	} else if orderBy == "name" {
		sort.Sort(member_score.ByName(memberScores))
	} else {
		sort.Sort(member_score.ByPart2Diff(memberScores))
		orderBy = "part2diff"
	}

	type DayScores map[string]interface{}

	type Context map[string]interface{}
	c := Context{
		"day": day,
		"year": leaderboard.CurrentBoard.Year,
		"orderBy": orderBy,
		"dayScores": DayScores{
			"day": day,
			"scores": memberScores,
			"orderBy": orderBy,
		},
		"topScores": leaderboard.CurrentBoard.TopScores[:20],
		"maxDay" : int(leaderboard.CurrentBoard.MaxDay) + 1,
	}

	funcMap := template.FuncMap{
		"readableTime": leaderboard.ReadableTime,
		"N": iter.N,
	}

	tmpl := template.Must(template.New("day.html").Funcs(funcMap).ParseGlob("templates/*.html"))
	err = tmpl.ExecuteTemplate(w, "day.html", c)
	if err != nil {
		log.Printf("Error executin template: %v", err)
	}
}
