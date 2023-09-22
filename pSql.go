package main

import (
  "database/sql"
  "fmt"
)

func createPassrf() error {
  db, err := sql.Open("sqlite3", "wallx.db")
  if err != nil {
    fmt.Println(" Err when 'open'")
    return err
  }
  defer db.Close()
  query4 := `create table if not exists passrf(serialnum text primary key, date text, whom text, code text)`
  _, err = db.Exec(query4)
  
  if err != nil {
    fmt.Println(" Err when exec 'create'")
    return err
  }
  fmt.Println(" Created!")
  return nil
}

func dropPassrf() error {
  db, err := sql.Open("sqlite3", "wallx.db")
  if err != nil {
    fmt.Println(" Err when 'open'")
    return err
  }
  query6 := `drop table if exists passrf`
  _, err = db.Exec(query6)
  if err != nil {
    fmt.Println(" Err when drop ")
    return err
  }
  return nil
}

func renameWhen() error {
  query5 := `alter table passrf rename column 'when' to date`
  db, err := sql.Open("sqlite3", "wallx.db")
  if err != nil {
    fmt.Println(" Err when 'open'")
    return err
  }
  defer db.Close()
  _, err = db.Exec(query5)
  if err != nil {
    fmt.Println(" Err when 'Exec'")
    return err
  }
  fmt.Println(" Renamed!")
  return nil
} 