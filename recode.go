package main

import (
  "net/http"
  "database/sql"
  "fmt"
  "github.com/yurajp/wallx/purecrypt"
  "time"
  "os"
  "io"
  "errors"
  
)

func RecodeWeb(w http.ResponseWriter, r *http.Request) {
  if r.Method == http.MethodGet {
    if wc, ok := app.web.templs["wellcome"]; ok {
      wc.Execute(w, false)
    } 
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    check(err)
    word1 := r.FormValue("word1")
    word2 := r.FormValue("word2")
    if len(word1) < 5 || word1 != word2 {
      http.Redirect(w, r, "/", 401)
    }
    err = app.RecodeDb(word1)
    check(err)

    err = app.execTempl(w, "message", "Password was changed")
  }
}


func (app *App) RecodeDb(newWord string) error {
    nDb, err := sql.Open("sqlite3", "temp.db")
    if err != nil {
       return err
    }
    defer nDb.Close()
    nWb := &Web{}
    nAp := &App{nWb, nDb}
    nAp.createTables()
    
    querySs := `select * from sites`
    queryA := `insert in sites(name, login, pass, link) values(?, ?, ?, ?)`
    rowsSs, err := app.db.Query(querySs)
    if err != nil {
      return err
    }
    for rowsSs.Next() {
      var nm, lg, ps, lk string
      rowsSs.Scan(&nm, &lg, &ps, &lk)
      nlg := purecrypt.Symcode(purecrypt.Desymcode(lg, app.web.word), newWord)
      nps := purecrypt.Symcode(purecrypt.Desymcode(ps, app.web.word), newWord)
      _, err = nDb.Exec(queryA, nm, nlg, nps, lk)
      if err != nil {
        return err
      }
    }
    queryCs := `select * from cards`
    queryB := `insert in cards(name, number, expire, cvc) values(?, ?, ?, ?)`
    rowsCs, err := app.db.Query(queryCs)
    if err != nil {
      return err
    }
    for rowsCs.Next() {
      var nc, nb, ex, cv string
      rowsCs.Scan(&nc, &nb, &ex, &cv)
      nnb := purecrypt.Symcode(purecrypt.Desymcode(nb, app.web.word), newWord)
      nex := purecrypt.Symcode(purecrypt.Desymcode(ex, app.web.word), newWord)
      ncv := purecrypt.Symcode(purecrypt.Desymcode(cv, app.web.word), newWord)
      _, err = nDb.Exec(queryB, nc, nnb, nex, ncv)
      if err != nil {
        return err
      }
    }
    nDb.Close()
    app.db.Close()
    app.db = nil
    err = os.Rename("temp.db", "wallx.db")
    if err != nil {
      return err
    }
    db, err := sql.Open("sqlite3", "wallx.db")
    if err != nil {
      return err
    }
    app.db = db
    app.web.word = newWord
    err = purecrypt.WriteCheckword(newWord)
    if err != nil {
      return err
    }
    return nil
}

func (app *App) BackupDb() error {
  ty, tm, td := time.Now().Date()
  date := fmt.Sprintf("%v%v%v", ty - 2000, tm, td)
  i, err := os.Stat("archive")
  if errors.Is(err, os.ErrNotExist) || !i.IsDir() {
    os.Mkdir("archive", 0750)
  }
  fpath := fmt.Sprintf("archive/%s.wallx.db", date)
  f, err := os.Create(fpath)
  if err != nil {
    return fmt.Errorf("Cannot create db for backup:\n %s", err)
  }
  defer f.Close()
  dbf, err := os.Open("wallx.db")
  if err != nil {
    return fmt.Errorf("Cannot open db for backup:\n %s", err)
  }
  defer dbf.Close()
  _, err = io.Copy(f, dbf)
  if err != nil {
    return fmt.Errorf("Cannot copy db to backup:\n %s", err)
  }
  return nil
}

func BackupWeb(w http.ResponseWriter, r *http.Request) {
  err := app.BackupDb()
  if err != nil {
    terr := app.execTempl(w, "message", fmt.Sprintf("%s", err))
    check(terr)
  }
  err = app.execTempl(w, "message", "Backup done")
  check(err)
}