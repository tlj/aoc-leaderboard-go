# build stage
FROM golang:1.11 AS build-env

RUN go get github.com/tools/godep

ADD Godeps/ /go/src/github.com/tlj/aoc-leaderboard-go/Godeps/

WORKDIR /go/src/github.com/tlj/aoc-leaderboard-go

RUN godep restore

ADD . /go/src/github.com/tlj/aoc-leaderboard-go

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/aoc-leaderboard main.go

# build
FROM iron/go

WORKDIR /app

# Now just add the binary
COPY --from=build-env /go/src/github.com/tlj/aoc-leaderboard-go/bin/aoc-leaderboard /app/
COPY --from=build-env /go/src/github.com/tlj/aoc-leaderboard-go/css/ /app/css/
COPY --from=build-env /go/src/github.com/tlj/aoc-leaderboard-go/templates/ /app/templates/

ENTRYPOINT ["./aoc-leaderboard"]
