package purecrypt

import (
  "fmt"
  "os"
  "encoding/json"
  "encoding/base64"
  "crypto/sha256"
)

type Checkword struct {
  Enc string `json: "enc"`
}

func ChWordExists() bool {
  _, err := os.Stat("checkword.json")
  if os.IsNotExist(err) {
    return false
  }
  return true
}

func HashStr(ps string) string {
	h := sha256.New()
	h.Write([]byte(ps))
	hh := h.Sum(nil)
	hhs := base64.StdEncoding.EncodeToString(hh)
	return string(hhs)
}


func WriteCheckword(pw string) error {
  phr := "Password is Correct"
  chw := Checkword{HashStr(Symcode(phr, pw))}
  f, err := os.Create("checkword.json")
  if err != nil {
    return err
  }
  defer f.Close()
  jsw, err := json.Marshal(chw)
  if err != nil {
    return err
  }
  _, err = f.Write(jsw)
  if err != nil {
    return err
  }
  return nil
}

func IsCorrect(pw string) bool {
  f, err := os.Open("checkword.json")
  if err != nil {
    return false
  }
  defer f.Close()
  var chw Checkword
  err = json.NewDecoder(f).Decode(&chw)
  if err != nil {
    fmt.Println(err)
    return false
  }
  phr := "Password is Correct"
  if HashStr(Symcode(phr, pw)) == chw.Enc {
    return true
  }
  return false
}

