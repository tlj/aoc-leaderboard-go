package main

import (
	"fmt"
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvNumeric(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatalf("Error getting env %s as numeric: %v\n", key, err)
		}
		return v
	}
	return fallback
}

func main() {
	cookie := getEnv("AOC_SESSION_COOKIE", "")
	year := getEnvNumeric("AOC_YEAR", int64(time.Now().Year()))
	id := getEnvNumeric("AOC_LEADERBOARD_ID", 0)
	debug := getEnvNumeric("AOC_DEBUG", 0)
	port := getEnvNumeric("HTTP_PORT", 8080)

	if cookie == "" || id == 0 {
		log.Fatal("AOC_SESSION_COOKIE and AOC_LEADERBOARD_ID env variables required.")
	}

	log.Printf("Starting leaderboard %d year %d.", id, year)

	leaderboard.CurrentBoard = leaderboard.LeaderBoard{
		SessionCookie: cookie,
		Year: year,
		Id: id,
		Debug: debug == 1,
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

	log.Printf("Listening to port %d.\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
