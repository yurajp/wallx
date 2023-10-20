package main

import (
  "time"
  "os"
  "os/exec" 
  "github.com/yurajp/bridge/config"
  "github.com/yurajp/bridge/database"
  "github.com/yurajp/bridge/server"
  "github.com/yurajp/bridge/client"
  "github.com/yurajp/bridge/web"
  "github.com/sirupsen/logrus"
)

var log *logrus.Logger

func iserr(s string, err error) bool {
  if err != nil {
    log.WithError(err).Error(s)
    return true
  }
  return false
}

func main() {
	log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Level = logrus.InfoLevel
	log.Out = os.Stdout
  exec.Command("termux-wifi-enable", "true").Run()
  err := config.LoadConf()
  if iserr("Cannot enable wi-fi", err) {
    return
  }
  if len(os.Args) > 1 && os.Args[1] == "sync" {
    log.Info("SYNC DB")
    err = database.SyncerDb()
    if iserr("Cannot sync DB", err) {
      return
    }
    return
  }
  err = database.PrepareDb()
  if iserr("PrepareDB failed", err) {
    return
  }
  go web.Launcher()
  Main:
  for {
    select {
    case mode := <-web.Cmode:
      if mode == "server" {
        go server.AsServer()
      } 
      if mode == "text" {
        go func() {
          log.Info("Sending text")
          err := client.AsClient("text")
          if iserr("Client for text failed", err) {
            return
          }
        }()
      }
      if mode == "files" {
        go func() {
          log.Info("Sending files")
          err := client.AsClient("files")
          if iserr("Client for files failed", err) {
            return
          }
        }()
      }
      if mode == "config" {
        err := config.TerminalConfig()
        if iserr("Config failed", err) {
          return
        }
      }
      if mode == "migrate" {
        err := database.MigratePgToSqlt()
        if iserr("Migration failed", err) {
          return
        }
      }
    case <-web.Q:
      break Main
    }
  }
  log.Info("BRIDGE CLOSED")
  time.Sleep(3 * time.Second)
}