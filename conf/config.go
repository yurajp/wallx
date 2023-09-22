package conf

import (
  "fmt"
  "encoding/json"
  "os"
  "time"
  "regexp"
  "strconv"
)

type Config struct {
  Port string `json: "port"`
  Livetime time.Duration `json: "livetime"`
}

var Cfg *Config

func ConfigExists() bool {
  _, err := os.Stat("conf/Config.json")
  if os.IsNotExist(err) {
    return false
  }
  return true
}

func SetConfigTerm() *Config {
  var port, valtime string
  done := false
  for !done {
    fmt.Print("\n  Print port number\n  or enter for default (8686)\n  ")
    var p string
    fmt.Scanf("%s", &p)
    if p == "" {
      port = ":8686"
      done = true
    } else {
      if PortIsCorrect(p) {
        port = ":" + p
        done = true
      } else {
        fmt.Println(" Port should be number between \n  1024 and 49151")
      }
    }
  }
  done = false
  var t string
  for !done {
    fmt.Println("\n  Print time in minutes\n  when password is valid\n  or enter for default (3 min)")
    fmt.Scanf("%s", &t)
    if t == "" {
      valtime = "3"
      done = true
    } else {
      if TimeIsCorrect(t) {
        valtime = t
        done = true
      } else {
        fmt.Println("  Time must be number")
      }
    }
  }
  livetime, _ := time.ParseDuration(valtime + "m")
  return &Config{port, livetime}
}

func PortIsCorrect(p string) bool {
  re := regexp.MustCompile(`^\d\d\d\d(\d)?$`)
  if !re.MatchString(p) {
    return false
  }
  d, _ := strconv.Atoi(p)
  if d < 1024 || d > 49151 {
    return false
  }
  return true
}

func TimeIsCorrect(t string) bool {
  re := regexp.MustCompile(`^\d{1,}$`)
  if !re.MatchString(t) {
    return false
  }
  return true
}

func WriteConfig(cfg *Config) error {
  js, err := json.Marshal(cfg)
  if err != nil {
    return err
  }
  err = os.WriteFile("conf/Config.json", js, 0640)
  if err != nil {
    return err
  }
  return nil
}

func GetConfig() (*Config, error) {
  if !ConfigExists() {
    return &Config{}, os.ErrNotExist
  }
  js, err := os.ReadFile("conf/Config.json")
  if err != nil {
    return &Config{}, err
  }
  var cfg Config
  err = json.Unmarshal(js, &cfg)
  if err != nil {
    return &Config{}, err
  }
  return &cfg, nil
}

func Prepare() error {
  if ConfigExists() {
    return nil
  }
  cfg := SetConfigTerm()
  err := WriteConfig(cfg)
  if err != nil {
    return err
  }
  return nil
}
