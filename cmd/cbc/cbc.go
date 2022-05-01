package main

import (
   "fmt"
   "github.com/89z/format/hls"
   "github.com/89z/mech/cbc"
   "net/http"
   "os"
   "sort"
)

func doManifest(id, address string, bandwidth int, info bool) error {
   cache, err := os.UserCacheDir()
   if err != nil {
      return err
   }
   profile, err := cbc.OpenProfile(cache, "mech/cbc.json")
   if err != nil {
      return err
   }
   if id == "" {
      id = cbc.GetID(address)
   }
   asset, err := cbc.NewAsset(id)
   if err != nil {
      return err
   }
   media, err := profile.Media(asset)
   if err != nil {
      return err
   }
   fmt.Println("GET", media.URL)
   res, err := http.Get(media.URL)
   if err != nil {
      return err
   }
   defer res.Body.Close()
   master, err := hls.NewScanner(res.Body).Master(res.Request.URL)
   if err != nil {
      return err
   }
   if bandwidth >= 1 {
      sort.Sort(hls.Bandwidth{master, bandwidth})
   }
   for _, stream := range master.Stream {
      if info {
         fmt.Println(stream)
      } else {
         /*
         video, err := cbc.NewVideo(guid)
         if err != nil {
            return err
         }
         return download(stream, video)
         */
      }
   }
   return nil
}

func doProfile(email, password string) error {
   cache, err := os.UserCacheDir()
   if err != nil {
      return err
   }
   login, err := cbc.NewLogin(email, password)
   if err != nil {
      return err
   }
   web, err := login.WebToken()
   if err != nil {
      return err
   }
   top, err := web.OverTheTop()
   if err != nil {
      return err
   }
   profile, err := top.Profile()
   if err != nil {
      return err
   }
   return profile.Create(cache, "mech/cbc.json")
}

/*
func download(stream hls.Stream, video *cbc.Video) error {
   fmt.Println("GET", stream.URI)
   res, err := http.Get(stream.URI.String())
   if err != nil {
      return err
   }
   defer res.Body.Close()
   seg, err := hls.NewScanner(res.Body).Segment(res.Request.URL)
   if err != nil {
      return err
   }
   file, err := os.Create(video.Base() + seg.Ext())
   if err != nil {
      return err
   }
   defer file.Close()
   pro := format.ProgressChunks(file, len(seg.Info))
   for _, info := range seg.Info {
      res, err := http.Get(info.URI.String())
      if err != nil {
         return err
      }
      pro.AddChunk(res.ContentLength)
      if _, err := io.Copy(pro, res.Body); err != nil {
         return err
      }
      if err := res.Body.Close(); err != nil {
         return err
      }
   }
   return nil
}
*/
