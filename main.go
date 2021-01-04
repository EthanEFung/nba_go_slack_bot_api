package main

import (
  "os"
  "io/ioutil"
  "fmt"
  "bytes"
  "net/http"
  "crypto/sha256"
  "crypto/hmac"
  "net/url"
  "log"
  "encoding/hex"
  "encoding/json"
  "strings"
  "strconv"
  "time"

  "go_nba/schedule"
  "github.com/gorilla/mux"
)

func init() {
  if os.Getenv("SLACK_SS") == "" {
    log.Fatal("Must have a slack ss")
  }
}

/*
Payload is the JSON sent to the slack api that is used to send a
response to the user in the channel
*/
type Payload struct {
  ResponseType string `json:"response_type"`
  Text string `json:"text"`
}

/* writeError will will send the error a 200 status response */
func writeError(w http.ResponseWriter, e error) bool {
  if e != nil {
    fmt.Println(e)
    w.Write([]byte("Error: " + e.Error()))
  }
  return e != nil
}

func writePayload(w http.ResponseWriter, p Payload) {
  w.Header().Set("Content-Type", "application/json")
  b, err := json.Marshal(p)
  if writeError(w, err) {
    return
  }
  w.Write(b)
}

func postMessage(v url.Values) {
  http.PostForm("https://slack.com/api/chat.postMessage", v)
}

func handleChallenge(w http.ResponseWriter, r *http.Request) {
  type event struct {
    Type string
    EventTS string
    User string
    Text string
    Channel string
  }

  type message struct {
    Challenge string 
    Token string
    TeamID string
    Event event
    Type string
    EventContext string
    EventID string
    EventTime int
  }

  decoder := json.NewDecoder(r.Body)
  var m message
  err := decoder.Decode(&m)
  if writeError(w, err) {
    return
  }
  if m.Challenge != "" {
    w.Write([]byte(m.Challenge))
  }
}

func handleGames(w http.ResponseWriter, r *http.Request) {
  fmt.Println("Handling Games")
  defer r.Body.Close()
  err := r.ParseForm()
  if writeError(w, err) {
    return
  }
  
  words := strings.Fields(r.Form.Get("text"))
  if len(words) == 0 {
    writePayload(w, Payload{
      ResponseType: "ephemeral",
      Text: "Get the nba schedule with the slash command `/nba_games today` or `/nba_games tomorrow`",
    })
    return;
  }

  loc, err := time.LoadLocation("America/Los_Angeles");
  if  writeError(w, err) {
    return
  }
  now := time.Now().In(loc);

  switch words[0] {
  case "today":
    writePayload(w, Payload{
      ResponseType: "in_channel",
      Text: schedule.GetSchedule(now),
    })
  case "tomorrow":
    writePayload(w, Payload{
      ResponseType: "in_channel",
      Text: schedule.GetSchedule(now.AddDate(0,0,1)),
    })
  case "help":
    writePayload(w, Payload{
      ResponseType: "ephemeral",
      Text: "Get the nba schedule with the slash command `/nba_games today` or `/nba_games tomorrow`",
    })
  default:
    writePayload(w, Payload{
      ResponseType: "ephemeral",
      Text: "Hmmm, I don't understand.\nGet the nba schedule with the slash command `/nba_games today` or `/nba_games tomorrow`",
    })
  }
}


func main() {
  router := mux.NewRouter()
  router.StrictSlash(true).Use(
    mux.CORSMethodMiddleware(router),
    // verify slack secret
    func(next http.Handler) http.Handler {
      fmt.Println("Authorizing")
      return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        key := []byte(os.Getenv("SLACK_SS"))
        buf, err := ioutil.ReadAll(r.Body)
        if writeError(w, err) {
          return
        }
        rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
        if writeError(w, err) {
          return
        }
        v, t := "v0", r.Header.Get("X-Slack-Request-Timestamp")
        stamp, err := strconv.Atoi(t)
        if (time.Now().Unix() - int64(stamp)) > 60 * 5 {
          http.Error(w, "Authorized Time Expired", 403)
          return
        }
        message := []byte(fmt.Sprintf("%s:%s:%s", v, t, buf))
        mac := hmac.New(sha256.New, key)
        mac.Write(message)
        expectedMAC := []byte(v + "=" + hex.EncodeToString(mac.Sum(nil)))
        messageMAC := []byte(r.Header.Get("X-Slack-Signature"))
        
        if hmac.Equal(messageMAC, expectedMAC) == false {
          http.Error(w, "Forbidden", 403)
          return
        }
        r.Body = rdr2
        next.ServeHTTP(w,r)
      })
    },
  )
  router.HandleFunc("/", handleChallenge).Methods("POST")
  router.HandleFunc("/games", handleGames).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}