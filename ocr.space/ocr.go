// OCR
package ocr

import (
   "bytes"
   "encoding/json"
   "github.com/89z/mech"
   "io"
   "mime/multipart"
   "os"
   "strings"
)

const (
   API = "http://api.ocr.space/parse/image"
)

func createForm(form map[string]string) (string, io.Reader, error) {
   body := new(bytes.Buffer)
   mp := multipart.NewWriter(body)
   defer mp.Close()
   for key, val := range form {
      if strings.HasPrefix(val, "@") {
         val = val[1:]
         file, err := os.Open(val)
         if err != nil {
            return "", nil, err
         }
         defer file.Close()
         part, err := mp.CreateFormFile(key, val)
         if err != nil {
            return "", nil, err
         }
         io.Copy(part, file)
      } else {
         mp.WriteField(key, val)
      }
   }
   return mp.FormDataContentType(), body, nil
}

type Image struct {
   ParsedResults []struct {
      ParsedText string
   }
}

func NewImage(name string) (*Image, error) {
   form := map[string]string{"OCREngine": "2", "file": "@" + name}
   ct, body, err := createForm(form)
   if err != nil {
      return nil, err
   }
   req, err := mech.NewRequest("POST", API, body)
   if err != nil {
      return nil, err
   }
   req.Header.Set("Content-Type", ct)
   req.Header.Set("apikey", "helloworld")
   res, err := new(mech.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   img := new(Image)
   json.NewDecoder(res.Body).Decode(img)
   return img, nil
}
