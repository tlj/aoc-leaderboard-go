package leaderboard

import (
	"encoding/json"
	"strconv"
)

// A FlexInt is an int that can be unmarshalled from a JSON field
// that has either a number or a string value.
// E.g. if the json field contains an string "42", the
// FlexInt value will be "42".
type FlexInt int64

func (fi *FlexInt) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		return json.Unmarshal(b, (*int64)(fi))
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*fi = FlexInt(i)
	return nil
}

type Event struct {
	Year string
	Members map[string]struct {
		GlobalScore int `json:"global_score"`
		LocalScore int `json:"local_score"`
		Name string `json:"name"`
		Stars int `json:"stars"`
		CompletionDayLevels map[int]map[int]struct {
			GetStarTs FlexInt `json:"get_star_ts"`
		} `json:"completion_day_level"`
		LastStarTs FlexInt `json:"last_star_ts"`
		Id int `json:"id,string"`
	}
	OwnerId string `json:"owner_id"`
}
