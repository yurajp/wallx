package main

import (
  "fmt"
  "time"
  "os"
  "os/exec" 
  "github.com/yurajp/bridge/config"
  "github.com/yurajp/bridge/database"
  "github.com/yurajp/bridge/server"
  "github.com/yurajp/bridge/client"
  "github.com/yurajp/bridge/web"
)


func iserr(err error) bool {
  if err != nil {
    fmt.Println(err)
    return true
  }
  return false
}

func main() {
  exec.Command("termux-wifi-enable", "true").Run()
  err := config.LoadConf()
  if iserr(err) {
    return
  }
  if len(os.Args) > 1 && os.Args[1] == "sync" {
    fmt.Println("\n\tSYNC DB")
    err = database.SyncerDb()
    if iserr(err) {
      return
    }
    return
  }
  err = database.PrepareDb()
  if iserr(err) {
    return
  }
  go web.Launcher()
  Main:
  for {
    select {
    case mode := <-web.Cmode:
      if mode == "server" {
        go server.AsServer()
        fmt.Println("\n\t BRIDGE server running\n")
      } 
      if mode == "text" {
        go func() {
          fmt.Println("\n\t TEXT")
          err := client.AsClient("text")
          if err != nil {
            fmt.Println(err)
          }
        }()
      }
      if mode == "files" {
        go func() {
          fmt.Println("\n\t FILES")
          err := client.AsClient("files")
          if err != nil {
            fmt.Println(err)
          }
        }()
      }
      if mode == "config" {
        err := config.TerminalConfig()
        if err != nil {
          fmt.Println(err)
        }
      }
      if mode == "migrate" {
        err := database.MigratePgToSqlt()
        if err != nil {
          fmt.Println(err)
        }
      }
    case <-web.Q:
      break Main
  //
    }
  }
  fmt.Println("\t CLOSED")
  time.Sleep(3 * time.Second)
}