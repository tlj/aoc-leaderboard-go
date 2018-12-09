
Advent of Code Leaderboard Day-by-day
=====================================

Preparations
------------

Find you session cookie from the AoC website. Find the leaderboard ID.

Usage
-----

```
docker run -p 80:8081 \
    -e AOC_LEADERBOARD_ID=[YOUR LEADERBOARD ID] \
    -e AOC_SESSION_COOKIE=[YOUR SESSION COOKIE] \
    thomaslandro/aoc-leaderboard-go:latest
```

Open a web browser and point it to http://localhost
