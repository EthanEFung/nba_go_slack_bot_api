package schedule

import (
  "fmt"
  "net/http"
  "log"
  "encoding/json"
  "time"
  "strconv"
)

/*
Team contains the TriCode as 3 capital letter code and the number of wins and losses
*/
type Team struct {
  TeamID string
  TriCode string
  Win string
  Loss string
}

/*
Game contains the start time in utc and eastern, as well as the home and visiting teams
*/
type Game struct {
  StartTimeUTC string
  StartTimeEastern string
  HTeam Team
  VTeam Team
}

/*
Message contains the games for the date coming from uri "{{ date }}/scorboard.json"
*/
type Message struct {
  NumGames int
  Games []Game
}

func deriveNBADate(t time.Time) string {
  year, month, day := t.Date()
  result := strconv.Itoa(year)
  result += fmt.Sprintf("%02d", month)
  result += fmt.Sprintf("%02d", day)
  return result
}

/*
GetSchedule will get the times of the games given a date
*/
func GetSchedule(t time.Time) string {
  d := deriveNBADate(t)
  resp, err := http.Get("http://data.nba.net/10s/prod/v1/" + d + "/scoreboard.json")

  if err != nil {
    log.Fatal(err)
  }

  defer resp.Body.Close()

  decoder := json.NewDecoder(resp.Body)

  var m Message
  decoder.Decode(&m)
  if err != nil {
    log.Fatal(err)
  }
  
  result := strconv.Itoa(m.NumGames) + " games\n"
  for _, game := range m.Games {
    t, err := time.Parse(time.RFC3339, game.StartTimeUTC)
    if err != nil {
      log.Fatal(err)
    }
    
    pst, err := time.LoadLocation("America/Los_Angeles")
    if err != nil {
      log.Fatal(err)
    }
    result += ":" + game.HTeam.TriCode + ": - :" + game.VTeam.TriCode + ": @ " + t.In(pst).Format(time.Kitchen) + " PST\n"
  }

  return result
}

