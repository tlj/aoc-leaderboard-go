package handlers

import (
	"github.com/bradfitz/iter"
	"github.com/tlj/aoc-leaderboard-go/leaderboard"
	"html/template"
	"log"
	"net/http"
)

func TopScores(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"readableTime": leaderboard.ReadableTime,
		"N": iter.N,
	}

	type Context map[string]interface{}
	c := Context{
		"day": -1,
		"maxDay": int(leaderboard.CurrentBoard.MaxDay),
		"year": leaderboard.CurrentBoard.Year,
		"topScores": leaderboard.CurrentBoard.TopScores,
	}

	tmpl := template.Must(template.New("topscores.html").Funcs(funcMap).ParseGlob("templates/*.html"))
	err := tmpl.ExecuteTemplate(w, "topscores.html", c)
	if err != nil {
		log.Printf("Error executin template: %v", err)
	}
}


