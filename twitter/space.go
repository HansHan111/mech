package twitter

import (
   "encoding/json"
   "github.com/89z/rosso/http"
   "net/url"
   "path"
   "strings"
   "time"
)

const bearer =
   "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs=" +
   "1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"

var Client = http.Default_Client

type Guest struct {
   Guest_Token string
}

func New_Guest() (*Guest, error) {
   req, err := http.NewRequest(
      "POST", "https://api.twitter.com/1.1/guest/activate.json", nil,
   )
   if err != nil {
      return nil, err
   }
   req.Header.Set("Authorization", "Bearer " + bearer)
   res, err := Client.Do(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   guest := new(Guest)
   if err := json.NewDecoder(res.Body).Decode(guest); err != nil {
      return nil, err
   }
   return guest, nil
}

func (a Audio_Space) Duration() time.Duration {
   meta := a.Metadata
   if meta.Ended_At == 0 {
      return 0
   }
   return time.Duration(meta.Ended_At - meta.Started_At) * time.Millisecond
}

type errorString string

func (e errorString) Error() string {
   return string(e)
}

type Audio_Space struct {
   Metadata struct {
      Media_Key string
      Title string
      State string
      Started_At int64
      Ended_At int64 `json:"ended_at,string"`
   }
   Participants struct {
      Admins []struct {
         Display_Name string
      }
   }
}

// https://twitter.com/i/spaces/1jMJgednpreKL?s=20
func SpaceID(addr string) (string, error) {
   parse, err := url.Parse(addr)
   if err != nil {
      return "", err
   }
   return path.Base(parse.Path), nil
}

const spacePersistedQuery = "lFpix9BgFDhAMjn9CrW6jQ"

func (a Audio_Space) Time() time.Time {
   return time.UnixMilli(a.Metadata.Started_At)
}

func (g Guest) Audio_Space(id string) (*Audio_Space, error) {
   var str strings.Builder
   str.WriteString("https://twitter.com/i/api/graphql/")
   str.WriteString(spacePersistedQuery)
   str.WriteString("/AudioSpaceById")
   req, err := http.NewRequest("GET", str.String(), nil)
   if err != nil {
      return nil, err
   }
   req.Header = http.Header{
      "Authorization": {"Bearer " + bearer},
      "X-Guest-Token": {g.Guest_Token},
   }
   buf, err := json.Marshal(spaceRequest{ID: id})
   if err != nil {
      return nil, err
   }
   req.URL.RawQuery = "variables=" + url.QueryEscape(string(buf))
   res, err := Client.Do(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   var space struct {
      Data struct {
         AudioSpace Audio_Space
      }
   }
   if err := json.NewDecoder(res.Body).Decode(&space); err != nil {
      return nil, err
   }
   return &space.Data.AudioSpace, nil
}

func (g Guest) Source(space *Audio_Space) (*Source, error) {
   var str strings.Builder
   str.WriteString("https://twitter.com/i/api/1.1/live_video_stream/status/")
   str.WriteString(space.Metadata.Media_Key)
   req, err := http.NewRequest("GET", str.String(), nil)
   if err != nil {
      return nil, err
   }
   req.Header = http.Header{
      "Authorization": {"Bearer " + bearer},
      "X-Guest-Token": {g.Guest_Token},
   }
   res, err := Client.Do(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   var video struct {
      Source Source
   }
   if err := json.NewDecoder(res.Body).Decode(&video); err != nil {
      return nil, err
   }
   return &video.Source, nil
}

type Source struct {
   Location string // Segment
}

type spaceRequest struct {
   ID string `json:"id"`
   IsMetatagsQuery bool `json:"isMetatagsQuery"`
   WithBirdwatchPivots bool `json:"withBirdwatchPivots"`
   WithDownvotePerspective bool `json:"withDownvotePerspective"`
   WithReactionsMetadata bool `json:"withReactionsMetadata"`
   WithReactionsPerspective bool `json:"withReactionsPerspective"`
   WithReplays bool `json:"withReplays"`
   WithScheduledSpaces bool `json:"withScheduledSpaces"`
   WithSuperFollowsTweetFields bool `json:"withSuperFollowsTweetFields"`
   WithSuperFollowsUserFields bool `json:"withSuperFollowsUserFields"`
}

func (a Audio_Space) String() string {
   var buf strings.Builder
   buf.WriteString("Key: ")
   buf.WriteString(a.Metadata.Media_Key)
   buf.WriteString("\nTitle: ")
   buf.WriteString(a.Metadata.Title)
   buf.WriteString("\nState: ")
   buf.WriteString(a.Metadata.State)
   if a.Metadata.Started_At >= 1 {
      buf.WriteString("\nStarted: ")
      buf.WriteString(a.Time().String())
   }
   if a.Metadata.Ended_At >= 1 {
      buf.WriteString("\nDuration: ")
      buf.WriteString(a.Duration().String())
   }
   for _, admin := range a.Participants.Admins {
      buf.WriteString("\nAdmin: ")
      buf.WriteString(admin.Display_Name)
   }
   return buf.String()
}

func (a Audio_Space) Base() string {
   var buf strings.Builder
   for _, admin := range a.Participants.Admins {
      buf.WriteString(admin.Display_Name)
      break
   }
   buf.WriteByte('-')
   buf.WriteString(a.Metadata.Title)
   return buf.String()
}
