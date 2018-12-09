package main

import (
	handlers2 "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/tlj/aoc-leaderboard-go/handlers"
	"github.com/tlj/aoc-leaderboard-go/leaderboard"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	cookie := os.Getenv("AOC_SESSION_COOKIE")
	year := os.Getenv("AOC_YEAR")
	id := os.Getenv("AOC_LEADERBOARD_ID")
	debug := os.Getenv("AOC_DEBUG")

	if cookie == "" || id == "" {
		log.Fatal("AOC_SESSION_COOKIE, AOC_YEAR and AOC_LEADERBOARD_ID env variables required.")
	}

	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}

	var err error
	var yearInt int64
	if yearInt, err = strconv.ParseInt(year, 10, 64); err != nil {
		log.Fatal("AOC_YEAR has to be int.")
	}
	var idInt int64
	if idInt, err = strconv.ParseInt(id, 10, 64); err != nil {
		log.Fatal("AOC_YEAR has to be int.")
	}

	log.Printf("Starting leaderboard %d year %d.", idInt, yearInt)

	leaderboard.CurrentBoard = leaderboard.LeaderBoard{
		SessionCookie: cookie,
		Year: yearInt,
		Id: idInt,
		Debug: debug == "1",
	}
	leaderboard.CurrentBoard.UpdateFromSource()

	go func() {
		for range time.NewTicker(120 * time.Second).C {
			leaderboard.CurrentBoard.UpdateFromSource()
		}
	}()

	r := mux.NewRouter()

	cssHandler := http.FileServer(http.Dir("./css/"))
	http.Handle("/css/", http.StripPrefix("/css/", cssHandler))
	r.HandleFunc("/day/{day:[0-9]+}/{orderBy}", handlers.Day)
	r.HandleFunc("/day/{day:[0-9]+}", handlers.Day)
	r.HandleFunc("/day/{day:[0-9]+}/", handlers.Day)
	r.HandleFunc("/day", handlers.Day)
	r.HandleFunc("/embed", handlers.Embed)
	r.HandleFunc("/topscores", handlers.TopScores)
	r.HandleFunc("/", handlers.Day)
	http.Handle("/", handlers2.CombinedLoggingHandler(os.Stdout, r))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
