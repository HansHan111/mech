package musicbrainz

import (
   "encoding/json"
   "github.com/89z/mech"
   "net/http"
   "sort"
   "strconv"
)

var Verbose = mech.Verbose

type Group struct {
   ReleaseCount int `json:"release-count"`
   Releases []*Release
}

func GroupFromArtist(artistID string, offset int) (*Group, error) {
   req, err := http.NewRequest("GET", API, nil)
   if err != nil {
      return nil, err
   }
   val := req.URL.Query()
   val.Set("artist", artistID)
   val.Set("fmt", "json")
   val.Set("inc", "release-groups")
   val.Set("limit", "100")
   val.Set("status", "official")
   val.Set("type", "album")
   if offset > 0 {
      val.Set("offset", strconv.Itoa(offset))
   }
   req.URL.RawQuery = val.Encode()
   res, err := mech.RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   grp := new(Group)
   if err := json.NewDecoder(res.Body).Decode(grp); err != nil {
      return nil, err
   }
   return grp, nil
}

func NewGroup(groupID string) (*Group, error) {
   req, err := http.NewRequest("GET", API, nil)
   if err != nil {
      return nil, err
   }
   val := req.URL.Query()
   val.Set("fmt", "json")
   val.Set("inc", "artist-credits recordings")
   val.Set("release-group", groupID)
   req.URL.RawQuery = val.Encode()
   res, err := mech.RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   grp := new(Group)
   if err := json.NewDecoder(res.Body).Decode(grp); err != nil {
      return nil, err
   }
   return grp, nil
}

func (g Group) Sort() {
   releaseFns := []releaseFn{
      func(a, b *Release) bool {
         status := map[string]int{"Official": 0, "Bootleg": 1}
         return status[a.Status] < status[b.Status]
      },
      func(a, b *Release) bool {
         return a.date(4) < b.date(4)
      },
      func(a, b *Release) bool {
         return a.trackLen() < b.trackLen()
      },
      func(a, b *Release) bool {
         return a.date(10) < b.date(10)
      },
   }
   sort.Slice(g.Releases, func(a, b int) bool {
      ra, rb := g.Releases[a], g.Releases[b]
      for _, fn := range releaseFns {
         if fn(ra, rb) {
            return true
         }
         if fn(rb, ra) {
            break
         }
      }
      return false
   })
}

type releaseFn func(a, b *Release) bool
