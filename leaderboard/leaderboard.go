package leaderboard

import (
	"encoding/json"
	"fmt"
	"github.com/tlj/aoc-leaderboard-go/member_score"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"
)

var CurrentBoard LeaderBoard

type LeaderBoard struct {
	Event *Event
	Year int64
	Id int64
	SessionCookie string
	LastSyncedAt time.Time
	Debug bool
	MaxDay int64
	Days map[int]*Day
	TopScores []*member_score.MemberScore
	Totals map[int]*member_score.MemberScore
}

type Day struct {
	Day int
	Year int64
	MemberScores map[int]*member_score.MemberScore
}

func (d Day) DayStartsAt() int64 {
	layout := "2006-01-02T15:04:05.000Z"
	str := fmt.Sprintf("%d-12-%02dT05:00:00.000Z", d.Year, d.Day)

	t, _ := time.Parse(layout, str)

	return t.Unix()
}

func (l *LeaderBoard) ApplyPenalties() {
	for _, day := range l.Days {
		fastestMember := member_score.MemberScore{Part2: 1000000, Part1: 1000000}
		for _, member := range day.MemberScores {
			if member.Part1 < fastestMember.Part1 && member.Part2 > 0 && member.Part1 > 0 {
				fastestMember = *member
			}
		}

		penaltyTreshold := fastestMember.Part1 * 2
		maxPenalty := fastestMember.Part2Diff() / 2

		for _, member := range day.MemberScores {
			if member.Part2 == 0 {
				continue
			}
			if member.Id != fastestMember.Id {
				penalty := int64(0)
				if member.Part1 > penaltyTreshold {
					startDiff := member.Part1 - penaltyTreshold
					if startDiff > 0 {
						penalty = startDiff / 20

						if penalty > maxPenalty {
							penalty = maxPenalty
						}

					}
				}
				member.WTime = member.Part2Diff() + penalty
			} else {
				member.WTime = member.Part2Diff()
			}
			if _, ok := l.Totals[member.Id]; ok {
				l.Totals[member.Id].WTime += member.WTime
			}
		}
	}
}

func (l *LeaderBoard) UpdateScores() {
	days := make(map[int]*Day)
	totals := make(map[int]*member_score.MemberScore)
	var topScores []*member_score.MemberScore
	maxDay := 0

	for _, member := range l.Event.Members {
		if member.Name == "" {
			member.Name = strconv.Itoa(member.Id)
		}
		for idx, day := range member.CompletionDayLevels {
			if _, ok := days[idx]; !ok {
				days[idx] = &Day{Year: l.Year, Day: idx, MemberScores:make(map[int]*member_score.MemberScore)}
			}
			dayStartsAt := days[idx].DayStartsAt()

			ms := member_score.MemberScore{
				Id:    member.Id,
				Name:  member.Name,
				Day:   idx,
				Count: 1,
			}

			if _, ok := day[1]; ok {
				ms.Part1 = int64(day[1].GetStarTs) - dayStartsAt
				if _, ok := day[2]; ok {
					ms.Part2 = int64(day[2].GetStarTs) - dayStartsAt
					if _, ok := totals[member.Id]; !ok {
						totals[member.Id] = &member_score.MemberScore{
							Id: member.Id,
							Name: member.Name,
							AocLocalScore: member.LocalScore,
							AocGlobalScore: member.GlobalScore,
							Count: 0,
							Part1: 0,
							Part2: 0,
						}
					}
					totals[member.Id].Part1 += ms.Part1
					totals[member.Id].Part2 += ms.Part2
					totals[member.Id].Count += 1

					topScores = append(topScores, &ms)
				}
			}

			days[idx].MemberScores[ms.Id] = &ms

			if idx > maxDay {
				maxDay = idx
			}
		}
	}

	completedTotals := make(map[int]*member_score.MemberScore)
	for id, member := range totals {
		if member.Count == int64(maxDay) {
			completedTotals[id] = member
		}
	}

	l.MaxDay = int64(maxDay)
	l.Days = days
	l.Totals = completedTotals
	l.TopScores = topScores
	l.ApplyPenalties()

	sort.Sort(member_score.ByWTime(topScores))
}

func (l *LeaderBoard) UpdateFromSource() {
	url := fmt.Sprintf("https://adventofcode.com/%d/leaderboard/private/view/%d.json", l.Year, l.Id)

	if l.Debug {
		log.Printf("Fake updating from rawData.")
		event := Event{}
		err := json.Unmarshal([]byte(rawData), &event)
		if err != nil {
			log.Fatal(err)
		}
		l.Event = &event
		l.LastSyncedAt = time.Now()
		l.UpdateScores()
		return
	}

	log.Printf("Updating from %s.", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: l.SessionCookie})
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Status: %d.", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	event := Event{}
	err = json.Unmarshal(body, &event)
	if err != nil {
		log.Fatal(err)
	}

	l.Event = &event
	l.LastSyncedAt = time.Now()
	l.UpdateScores()
}

func ReadableTime(timeSpent int64) string {
	var hours int64
	var minutes int64
	var timeString string

	if timeSpent > 60 * 60 {
		hours = int64(math.Floor(float64(timeSpent / (60 * 60))))
		timeString += fmt.Sprintf("%02d:", hours)

		timeSpent -= 60*60*hours
	}

	if timeSpent > 60 {
		minutes = int64(math.Floor(float64(timeSpent / 60)))
		timeString += fmt.Sprintf("%02d:", minutes)

		timeSpent -= 60 * minutes
	}

	timeString += fmt.Sprintf("%02d", timeSpent)

	return timeString
}

var rawData = `{"members":{"61663":{"stars":2,"name":"Torbjørn Alvestrand","global_score":0,"last_star_ts":"1544132297","local_score":47,"completion_day_level":{"1":{"1":{"get_star_ts":"1544044342"}},"2":{"1":{"get_star_ts":"1544132297"}}},"id":"61663"},"149805":{"local_score":433,"completion_day_level":{"6":{"2":{"get_star_ts":"1544088967"},"1":{"get_star_ts":"1544088252"}},"8":{"1":{"get_star_ts":"1544273310"}},"1":{"2":{"get_star_ts":"1543650055"},"1":{"get_star_ts":"1543648680"}},"2":{"1":{"get_star_ts":"1543742406"},"2":{"get_star_ts":"1543742985"}},"3":{"2":{"get_star_ts":"1543854108"},"1":{"get_star_ts":"1543816292"}},"5":{"2":{"get_star_ts":"1543989158"},"1":{"get_star_ts":"1543988385"}},"4":{"2":{"get_star_ts":"1543912603"},"1":{"get_star_ts":"1543910478"}}},"name":"Lukasz Chmielewski","stars":13,"last_star_ts":"1544273310","global_score":0,"id":"149805"},"201396":{"name":"Bjørnar Heggset Nes","stars":0,"global_score":0,"last_star_ts":0,"local_score":0,"completion_day_level":{},"id":"201396"},"145167":{"id":"145167","local_score":0,"completion_day_level":{},"name":"Geir Morten “Oyvang” Larsen","stars":0,"last_star_ts":0,"global_score":0},"200921":{"id":"200921","local_score":216,"completion_day_level":{"2":{"1":{"get_star_ts":"1543739664"},"2":{"get_star_ts":"1543740259"}},"1":{"1":{"get_star_ts":"1543663117"},"2":{"get_star_ts":"1543664501"}},"5":{"1":{"get_star_ts":"1543991436"},"2":{"get_star_ts":"1543992011"}}},"stars":6,"name":"hackon","last_star_ts":"1543992011","global_score":0},"246027":{"id":"246027","global_score":0,"last_star_ts":0,"name":null,"stars":0,"completion_day_level":{},"local_score":0},"184510":{"name":"czeski","stars":0,"global_score":0,"last_star_ts":0,"local_score":0,"completion_day_level":{},"id":"184510"},"229475":{"completion_day_level":{"7":{"2":{"get_star_ts":"1544181242"},"1":{"get_star_ts":"1544179334"}},"2":{"1":{"get_star_ts":"1543828026"},"2":{"get_star_ts":"1543828839"}},"1":{"2":{"get_star_ts":"1543825342"},"1":{"get_star_ts":"1543823513"}},"4":{"1":{"get_star_ts":"1543929014"},"2":{"get_star_ts":"1543931207"}},"5":{"1":{"get_star_ts":"1543997234"},"2":{"get_star_ts":"1543998768"}},"3":{"1":{"get_star_ts":"1543841363"},"2":{"get_star_ts":"1543842165"}},"6":{"1":{"get_star_ts":"1544088500"},"2":{"get_star_ts":"1544089699"}}},"local_score":404,"last_star_ts":"1544181242","global_score":0,"stars":14,"name":"Saradha Kannan","id":"229475"},"145084":{"stars":18,"name":"Torunn Sæther","global_score":0,"last_star_ts":"1544335188","local_score":608,"completion_day_level":{"2":{"2":{"get_star_ts":"1543943046"},"1":{"get_star_ts":"1543941688"}},"5":{"1":{"get_star_ts":"1543987081"},"2":{"get_star_ts":"1543987788"}},"9":{"2":{"get_star_ts":"1544335188"},"1":{"get_star_ts":"1544333287"}},"6":{"1":{"get_star_ts":"1544075658"},"2":{"get_star_ts":"1544077390"}},"3":{"1":{"get_star_ts":"1543988718"},"2":{"get_star_ts":"1543989413"}},"4":{"2":{"get_star_ts":"1543908059"},"1":{"get_star_ts":"1543907280"}},"1":{"1":{"get_star_ts":"1543940633"},"2":{"get_star_ts":"1543940914"}},"7":{"1":{"get_star_ts":"1544161276"},"2":{"get_star_ts":"1544163402"}},"8":{"2":{"get_star_ts":"1544251019"},"1":{"get_star_ts":"1544248323"}}},"id":"145084"},"200948":{"id":"200948","stars":18,"name":"Trond Isak Alseth","last_star_ts":"1544359382","global_score":0,"local_score":641,"completion_day_level":{"3":{"2":{"get_star_ts":"1543866742"},"1":{"get_star_ts":"1543864348"}},"4":{"1":{"get_star_ts":"1543940778"},"2":{"get_star_ts":"1543940951"}},"1":{"2":{"get_star_ts":"1543641704"},"1":{"get_star_ts":"1543640874"}},"7":{"2":{"get_star_ts":"1544216537"},"1":{"get_star_ts":"1544206991"}},"8":{"2":{"get_star_ts":"1544249160"},"1":{"get_star_ts":"1544248178"}},"2":{"2":{"get_star_ts":"1543730398"},"1":{"get_star_ts":"1543727678"}},"5":{"1":{"get_star_ts":"1543989527"},"2":{"get_star_ts":"1543990417"}},"6":{"2":{"get_star_ts":"1544115883"},"1":{"get_star_ts":"1544115580"}},"9":{"1":{"get_star_ts":"1544335827"},"2":{"get_star_ts":"1544359382"}}}},"246876":{"name":"Mixa Ksh","stars":0,"last_star_ts":0,"global_score":0,"local_score":0,"completion_day_level":{},"id":"246876"},"40128":{"id":"40128","completion_day_level":{},"local_score":0,"global_score":0,"last_star_ts":0,"name":"Lars Jølsum","stars":0},"246294":{"completion_day_level":{},"local_score":0,"last_star_ts":0,"global_score":0,"name":"Wuxscho","stars":0,"id":"246294"},"40808":{"id":"40808","last_star_ts":0,"global_score":0,"name":"Runar Os Mathisen","stars":0,"completion_day_level":{},"local_score":0},"352156":{"id":"352156","name":"Kai Rune Orten","stars":10,"last_star_ts":"1543995337","global_score":0,"local_score":361,"completion_day_level":{"4":{"1":{"get_star_ts":"1543942554"},"2":{"get_star_ts":"1543944241"}},"1":{"1":{"get_star_ts":"1543654595"},"2":{"get_star_ts":"1543656001"}},"2":{"2":{"get_star_ts":"1543736672"},"1":{"get_star_ts":"1543733279"}},"5":{"1":{"get_star_ts":"1543991281"},"2":{"get_star_ts":"1543995337"}},"3":{"1":{"get_star_ts":"1543867072"},"2":{"get_star_ts":"1543867874"}}}},"202546":{"completion_day_level":{"3":{"1":{"get_star_ts":"1543819286"},"2":{"get_star_ts":"1543819612"}},"1":{"2":{"get_star_ts":"1543650944"},"1":{"get_star_ts":"1543650543"}},"4":{"1":{"get_star_ts":"1543910088"},"2":{"get_star_ts":"1543910307"}},"7":{"2":{"get_star_ts":"1544224145"},"1":{"get_star_ts":"1544221750"}},"8":{"1":{"get_star_ts":"1544254760"},"2":{"get_star_ts":"1544255905"}},"2":{"1":{"get_star_ts":"1543739385"},"2":{"get_star_ts":"1543740257"}},"5":{"1":{"get_star_ts":"1543991544"},"2":{"get_star_ts":"1543991935"}},"6":{"1":{"get_star_ts":"1544218675"},"2":{"get_star_ts":"1544219228"}},"9":{"1":{"get_star_ts":"1544372797"},"2":{"get_star_ts":"1544372918"}}},"local_score":636,"global_score":0,"last_star_ts":"1544372918","stars":18,"name":"Niko Pavlica","id":"202546"},"302967":{"global_score":0,"last_star_ts":0,"name":null,"stars":0,"completion_day_level":{},"local_score":0,"id":"302967"},"371735":{"completion_day_level":{"2":{"1":{"get_star_ts":"1543759749"},"2":{"get_star_ts":"1543781400"}},"5":{"2":{"get_star_ts":"1544001881"},"1":{"get_star_ts":"1543998588"}},"1":{"1":{"get_star_ts":"1543651397"},"2":{"get_star_ts":"1543671219"}},"3":{"1":{"get_star_ts":"1543826320"},"2":{"get_star_ts":"1543867749"}},"4":{"2":{"get_star_ts":"1543966036"},"1":{"get_star_ts":"1543962829"}}},"local_score":325,"last_star_ts":"1544001881","global_score":0,"name":"Blaž Maležič","stars":10,"id":"371735"},"12047":{"id":"12047","local_score":0,"completion_day_level":{},"name":"Jure Prevc","stars":0,"global_score":0,"last_star_ts":0},"13966":{"id":"13966","stars":17,"name":"kratskij","global_score":0,"last_star_ts":"1544335700","local_score":670,"completion_day_level":{"6":{"2":{"get_star_ts":"1544083493"},"1":{"get_star_ts":"1544082920"}},"9":{"1":{"get_star_ts":"1544335700"}},"2":{"2":{"get_star_ts":"1543733486"},"1":{"get_star_ts":"1543732063"}},"5":{"2":{"get_star_ts":"1543988766"},"1":{"get_star_ts":"1543986993"}},"8":{"1":{"get_star_ts":"1544247337"},"2":{"get_star_ts":"1544248885"}},"3":{"2":{"get_star_ts":"1543814413"},"1":{"get_star_ts":"1543813804"}},"1":{"2":{"get_star_ts":"1543641459"},"1":{"get_star_ts":"1543640637"}},"4":{"1":{"get_star_ts":"1543901940"},"2":{"get_star_ts":"1543902151"}},"7":{"2":{"get_star_ts":"1544274371"},"1":{"get_star_ts":"1544167438"}}}},"128375":{"global_score":0,"last_star_ts":"1544370016","stars":12,"name":"Thomas Isaksen","completion_day_level":{"2":{"1":{"get_star_ts":"1544217644"},"2":{"get_star_ts":"1544219134"}},"3":{"1":{"get_star_ts":"1544278782"},"2":{"get_star_ts":"1544284404"}},"5":{"2":{"get_star_ts":"1544310524"},"1":{"get_star_ts":"1544309075"}},"1":{"1":{"get_star_ts":"1544205728"},"2":{"get_star_ts":"1544214849"}},"4":{"2":{"get_star_ts":"1544291801"},"1":{"get_star_ts":"1544291693"}},"6":{"1":{"get_star_ts":"1544367089"},"2":{"get_star_ts":"1544370016"}}},"local_score":247,"id":"128375"},"245916":{"completion_day_level":{"5":{"1":{"get_star_ts":"1543986371"},"2":{"get_star_ts":"1543986693"}},"2":{"2":{"get_star_ts":"1543727890"},"1":{"get_star_ts":"1543727347"}},"9":{"2":{"get_star_ts":"1544336084"},"1":{"get_star_ts":"1544335195"}},"6":{"1":{"get_star_ts":"1544082412"},"2":{"get_star_ts":"1544083095"}},"3":{"1":{"get_star_ts":"1543813773"},"2":{"get_star_ts":"1543813862"}},"1":{"1":{"get_star_ts":"1543640525"},"2":{"get_star_ts":"1543641057"}},"7":{"1":{"get_star_ts":"1544159755"},"2":{"get_star_ts":"1544161152"}},"4":{"2":{"get_star_ts":"1543901207"},"1":{"get_star_ts":"1543901011"}},"8":{"1":{"get_star_ts":"1544246427"},"2":{"get_star_ts":"1544246958"}}},"local_score":761,"last_star_ts":"1544336084","global_score":0,"name":"Nejc Ramovs","stars":18,"id":"245916"},"202227":{"id":"202227","completion_day_level":{"9":{"1":{"get_star_ts":"1544340408"},"2":{"get_star_ts":"1544343568"}},"6":{"2":{"get_star_ts":"1544075403"},"1":{"get_star_ts":"1544074791"}},"5":{"1":{"get_star_ts":"1543988638"},"2":{"get_star_ts":"1543989047"}},"2":{"1":{"get_star_ts":"1543744177"},"2":{"get_star_ts":"1543745946"}},"8":{"1":{"get_star_ts":"1544260297"},"2":{"get_star_ts":"1544261118"}},"4":{"1":{"get_star_ts":"1543912270"},"2":{"get_star_ts":"1543912517"}},"1":{"2":{"get_star_ts":"1543655811"},"1":{"get_star_ts":"1543655175"}},"3":{"1":{"get_star_ts":"1543820642"},"2":{"get_star_ts":"1543822341"}},"7":{"2":{"get_star_ts":"1544165024"},"1":{"get_star_ts":"1544161782"}}},"local_score":641,"global_score":0,"last_star_ts":"1544343568","stars":18,"name":"Gisle"},"392678":{"id":"392678","completion_day_level":{"5":{"1":{"get_star_ts":"1543987755"},"2":{"get_star_ts":"1543988322"}},"2":{"1":{"get_star_ts":"1543734510"},"2":{"get_star_ts":"1543735914"}},"6":{"2":{"get_star_ts":"1544076838"},"1":{"get_star_ts":"1544074784"}},"9":{"2":{"get_star_ts":"1544343058"},"1":{"get_star_ts":"1544339228"}},"3":{"1":{"get_star_ts":"1543818911"},"2":{"get_star_ts":"1543835180"}},"1":{"1":{"get_star_ts":"1543655674"},"2":{"get_star_ts":"1543657042"}},"7":{"2":{"get_star_ts":"1544166053"},"1":{"get_star_ts":"1544163506"}},"4":{"2":{"get_star_ts":"1543905419"},"1":{"get_star_ts":"1543904653"}},"8":{"1":{"get_star_ts":"1544260301"},"2":{"get_star_ts":"1544260999"}}},"local_score":659,"global_score":0,"last_star_ts":"1544343058","stars":18,"name":"Stian Fredrik Aune"},"139793":{"last_star_ts":"1544162986","global_score":0,"stars":13,"name":"Felix Rabe","completion_day_level":{"1":{"1":{"get_star_ts":"1543655180"},"2":{"get_star_ts":"1543655893"}},"2":{"2":{"get_star_ts":"1543727950"},"1":{"get_star_ts":"1543727518"}},"3":{"1":{"get_star_ts":"1543814093"},"2":{"get_star_ts":"1543814531"}},"7":{"1":{"get_star_ts":"1544162986"}},"5":{"2":{"get_star_ts":"1543989022"},"1":{"get_star_ts":"1543987347"}},"4":{"2":{"get_star_ts":"1543920506"},"1":{"get_star_ts":"1543919658"}},"6":{"2":{"get_star_ts":"1544081817"},"1":{"get_star_ts":"1544081329"}}},"local_score":460,"id":"139793"},"116603":{"completion_day_level":{"6":{"2":{"get_star_ts":"1544203588"},"1":{"get_star_ts":"1544202802"}},"9":{"2":{"get_star_ts":"1544338429"},"1":{"get_star_ts":"1544338410"}},"2":{"2":{"get_star_ts":"1543872895"},"1":{"get_star_ts":"1543872320"}},"5":{"2":{"get_star_ts":"1543996380"},"1":{"get_star_ts":"1543994909"}},"8":{"1":{"get_star_ts":"1544259759"},"2":{"get_star_ts":"1544260530"}},"1":{"2":{"get_star_ts":"1543871377"},"1":{"get_star_ts":"1543870885"}},"7":{"1":{"get_star_ts":"1544205872"},"2":{"get_star_ts":"1544210548"}},"3":{"1":{"get_star_ts":"1543875852"},"2":{"get_star_ts":"1543876709"}},"4":{"2":{"get_star_ts":"1543904956"},"1":{"get_star_ts":"1543904739"}}},"local_score":572,"global_score":0,"last_star_ts":"1544338429","stars":18,"name":"Thomas L. Johnsen","id":"116603"},"152261":{"local_score":0,"completion_day_level":{},"name":"Thord Setsaas","stars":0,"last_star_ts":0,"global_score":0,"id":"152261"},"139369":{"completion_day_level":{"8":{"1":{"get_star_ts":"1544356432"},"2":{"get_star_ts":"1544356983"}},"1":{"2":{"get_star_ts":"1543641104"},"1":{"get_star_ts":"1543640628"}},"4":{"1":{"get_star_ts":"1543934875"},"2":{"get_star_ts":"1543935151"}},"3":{"1":{"get_star_ts":"1543814145"},"2":{"get_star_ts":"1543815156"}},"7":{"2":{"get_star_ts":"1544352731"},"1":{"get_star_ts":"1544160319"}},"6":{"1":{"get_star_ts":"1544076289"},"2":{"get_star_ts":"1544077016"}},"9":{"1":{"get_star_ts":"1544361100"}},"5":{"2":{"get_star_ts":"1543988824"},"1":{"get_star_ts":"1543987129"}},"2":{"1":{"get_star_ts":"1543734275"},"2":{"get_star_ts":"1543735005"}}},"local_score":619,"global_score":0,"last_star_ts":"1544361100","name":"Jørgen Myklebust","stars":17,"id":"139369"},"191176":{"local_score":112,"completion_day_level":{"2":{"1":{"get_star_ts":"1543839816"},"2":{"get_star_ts":"1543841052"}},"1":{"2":{"get_star_ts":"1543837610"},"1":{"get_star_ts":"1543825360"}}},"stars":4,"name":"dpuric3","last_star_ts":"1543841052","global_score":0,"id":"191176"},"437241":{"id":"437241","global_score":0,"last_star_ts":"1544039307","stars":2,"name":"Ove Stavnås","completion_day_level":{"5":{"2":{"get_star_ts":"1544039307"},"1":{"get_star_ts":"1544037555"}}},"local_score":56},"371746":{"completion_day_level":{"2":{"2":{"get_star_ts":"1543746920"},"1":{"get_star_ts":"1543745853"}},"3":{"2":{"get_star_ts":"1543820023"},"1":{"get_star_ts":"1543819911"}},"1":{"1":{"get_star_ts":"1543691561"},"2":{"get_star_ts":"1543692401"}},"4":{"2":{"get_star_ts":"1543962209"},"1":{"get_star_ts":"1543961324"}},"7":{"1":{"get_star_ts":"1544352529"}},"5":{"1":{"get_star_ts":"1544046126"},"2":{"get_star_ts":"1544047492"}},"6":{"1":{"get_star_ts":"1544114570"},"2":{"get_star_ts":"1544121295"}}},"local_score":362,"global_score":0,"last_star_ts":"1544352529","name":null,"stars":13,"id":"371746"},"145234":{"id":"145234","last_star_ts":"1544373701","global_score":0,"stars":15,"name":"kalvatn","completion_day_level":{"1":{"1":{"get_star_ts":"1543650351"},"2":{"get_star_ts":"1543653295"}},"3":{"1":{"get_star_ts":"1543815987"},"2":{"get_star_ts":"1543817366"}},"7":{"1":{"get_star_ts":"1544263183"}},"4":{"2":{"get_star_ts":"1543903457"},"1":{"get_star_ts":"1543903408"}},"9":{"1":{"get_star_ts":"1544361628"},"2":{"get_star_ts":"1544373701"}},"6":{"2":{"get_star_ts":"1544119147"},"1":{"get_star_ts":"1544107288"}},"2":{"2":{"get_star_ts":"1543757477"},"1":{"get_star_ts":"1543756597"}},"5":{"1":{"get_star_ts":"1544000580"},"2":{"get_star_ts":"1544001621"}}},"local_score":501},"2793":{"last_star_ts":"1544169543","global_score":0,"name":"alrasch","stars":14,"completion_day_level":{"2":{"1":{"get_star_ts":"1543768365"},"2":{"get_star_ts":"1543768416"}},"1":{"1":{"get_star_ts":"1543751280"},"2":{"get_star_ts":"1543753593"}},"4":{"2":{"get_star_ts":"1543915726"},"1":{"get_star_ts":"1543915414"}},"3":{"1":{"get_star_ts":"1543827598"},"2":{"get_star_ts":"1543827648"}},"5":{"2":{"get_star_ts":"1544001028"},"1":{"get_star_ts":"1544000888"}},"7":{"2":{"get_star_ts":"1544169543"},"1":{"get_star_ts":"1544168889"}},"6":{"2":{"get_star_ts":"1544086906"},"1":{"get_star_ts":"1544085815"}}},"local_score":418,"id":"2793"},"211848":{"id":"211848","completion_day_level":{},"local_score":0,"last_star_ts":0,"global_score":0,"name":null,"stars":0},"415858":{"id":"415858","completion_day_level":{"1":{"1":{"get_star_ts":"1543705639"},"2":{"get_star_ts":"1543707326"}},"3":{"1":{"get_star_ts":"1543831507"},"2":{"get_star_ts":"1543833253"}},"4":{"1":{"get_star_ts":"1543963933"},"2":{"get_star_ts":"1543965514"}},"2":{"2":{"get_star_ts":"1543735948"},"1":{"get_star_ts":"1543734613"}}},"local_score":274,"global_score":0,"last_star_ts":"1543965514","stars":8,"name":"Thomas Doorman"},"231460":{"id":"231460","local_score":419,"completion_day_level":{"8":{"1":{"get_star_ts":"1544364781"},"2":{"get_star_ts":"1544365882"}},"1":{"1":{"get_star_ts":"1543658328"},"2":{"get_star_ts":"1543659211"}},"2":{"1":{"get_star_ts":"1543731447"},"2":{"get_star_ts":"1543749686"}},"4":{"2":{"get_star_ts":"1543959706"},"1":{"get_star_ts":"1543959172"}},"3":{"2":{"get_star_ts":"1543833711"},"1":{"get_star_ts":"1543831570"}},"5":{"2":{"get_star_ts":"1544001804"},"1":{"get_star_ts":"1544000641"}}},"stars":12,"name":"Michelle M. Ludvigsen","global_score":0,"last_star_ts":"1544365882"},"201909":{"id":"201909","last_star_ts":0,"global_score":0,"stars":0,"name":"ivarkrabol","completion_day_level":{},"local_score":0},"125387":{"id":"125387","stars":0,"name":"Lukasz Chmielewski","global_score":0,"last_star_ts":0,"local_score":0,"completion_day_level":{}},"136721":{"id":"136721","last_star_ts":"1543906880","global_score":0,"name":"Tonny Hovdal","stars":1,"completion_day_level":{"1":{"1":{"get_star_ts":"1543906880"}}},"local_score":26},"201972":{"local_score":0,"completion_day_level":{},"name":"aldochan","stars":0,"global_score":0,"last_star_ts":0,"id":"201972"},"253396":{"id":"253396","last_star_ts":0,"global_score":0,"name":"Jan Sochor","stars":0,"completion_day_level":{},"local_score":0},"276702":{"id":"276702","last_star_ts":"1544274770","global_score":0,"name":"Andreas Kalvå","stars":16,"completion_day_level":{"8":{"1":{"get_star_ts":"1544265446"},"2":{"get_star_ts":"1544274770"}},"3":{"1":{"get_star_ts":"1544191919"},"2":{"get_star_ts":"1544192786"}},"7":{"2":{"get_star_ts":"1544181901"},"1":{"get_star_ts":"1544175628"}},"1":{"2":{"get_star_ts":"1544022461"},"1":{"get_star_ts":"1544020640"}},"4":{"1":{"get_star_ts":"1544197487"},"2":{"get_star_ts":"1544198201"}},"6":{"2":{"get_star_ts":"1544110974"},"1":{"get_star_ts":"1544110502"}},"5":{"2":{"get_star_ts":"1544024547"},"1":{"get_star_ts":"1544023367"}},"2":{"2":{"get_star_ts":"1544128092"},"1":{"get_star_ts":"1544125624"}}},"local_score":423},"211347":{"id":"211347","last_star_ts":0,"global_score":0,"stars":0,"name":"the vlado","completion_day_level":{},"local_score":0},"201045":{"last_star_ts":"1544335145","global_score":0,"name":"George Samidare","stars":18,"completion_day_level":{"3":{"2":{"get_star_ts":"1543814167"},"1":{"get_star_ts":"1543813911"}},"1":{"2":{"get_star_ts":"1543642086"},"1":{"get_star_ts":"1543640609"}},"4":{"2":{"get_star_ts":"1543903497"},"1":{"get_star_ts":"1543903091"}},"7":{"1":{"get_star_ts":"1544160020"},"2":{"get_star_ts":"1544161092"}},"8":{"2":{"get_star_ts":"1544247441"},"1":{"get_star_ts":"1544246651"}},"2":{"1":{"get_star_ts":"1543727210"},"2":{"get_star_ts":"1543727662"}},"5":{"2":{"get_star_ts":"1543986839"},"1":{"get_star_ts":"1543986504"}},"6":{"2":{"get_star_ts":"1544073952"},"1":{"get_star_ts":"1544073682"}},"9":{"2":{"get_star_ts":"1544335145"},"1":{"get_star_ts":"1544334964"}}},"local_score":749,"id":"201045"},"204810":{"id":"204810","stars":12,"name":"Simen Brenne Wigtil","global_score":0,"last_star_ts":"1544377206","local_score":303,"completion_day_level":{"2":{"2":{"get_star_ts":"1543750702"},"1":{"get_star_ts":"1543747060"}},"1":{"2":{"get_star_ts":"1543662914"},"1":{"get_star_ts":"1543660413"}},"3":{"1":{"get_star_ts":"1543868248"},"2":{"get_star_ts":"1543869705"}},"4":{"1":{"get_star_ts":"1544131392"},"2":{"get_star_ts":"1544131826"}},"5":{"2":{"get_star_ts":"1544191397"},"1":{"get_star_ts":"1544187331"}},"6":{"2":{"get_star_ts":"1544377206"},"1":{"get_star_ts":"1544368633"}}}},"40888":{"id":"40888","local_score":0,"completion_day_level":{},"stars":0,"name":"Jon Terje Tvergrov Kalvatn","global_score":0,"last_star_ts":0},"259736":{"completion_day_level":{},"local_score":0,"global_score":0,"last_star_ts":0,"stars":0,"name":"Jørgen Lien Sellæg","id":"259736"},"46209":{"local_score":0,"completion_day_level":{},"stars":0,"name":"Haakon Heiberg","last_star_ts":0,"global_score":0,"id":"46209"}},"event":"2018","owner_id":"116603"}`