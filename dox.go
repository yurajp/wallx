package main

import (
    "fmt"
    "net/http"
    "context"
    
    "github.com/yurajp/wallx/purecrypt"
)

type PassRF struct {
  SerialNum string
  Date string
  Whom string
  Code string
}

type Doc struct {
  Name string
  Value string
}


func (app *App) AddDocToDb(d *Doc) error {
  query := `insert into docs(name, value) values(?, ?)`
  _, err := app.db.Exec(query, d.Name, d.Value)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) AddPassrfToDb(p *PassRF) error {
  query := `insert into passrf(serialnum, date, whom, code) values(?, ?, ?, ?)`
  sn := p.SerialNum
  wn := p.Date
  wm := p.Whom
  cd := p.Code
  _, err := app.db.Exec(query, sn, wn, wm, cd)
  if err != nil {
    return err
  }
  return nil
}

func createDocWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createDoc", nil)
    if err != nil {
      fmt.Println(err)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      fmt.Println(err)
    }
    nm := r.FormValue("name")
    val := r.FormValue("value")
    d := &Doc{nm, purecrypt.Symcode(val, app.web.word)}
    err = app.AddDocToDb(d)
    if err != nil {
      fmt.Println(err)
    }
    http.Redirect(w, r, "/docs", 303)
  }
}

func showDocsWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  ctx, _ := context.WithTimeout(context.Background(), livetime)
  go func() {
    for {
      select {
        case <-ctx.Done():
          app.Dies()
          return
        default:
      }
    }
  }()
  dcs, err := app.GetDocsFromDb() 
  if err != nil {
    fmt.Println(err)
    return
  }
  ddx := []Doc{}
  for _, d := range dcs {
    dd := Doc{d.Name, purecrypt.Desymcode(d.Value, app.web.word)}
    ddx = append(ddx, dd)
  }
  err = app.execTempl(w, "allDocs", ddx)
  if err != nil {
    fmt.Println(err)
  }
}

func (app *App) GetDocsFromDb() ([]Doc, error) {
  query := `select * from docs`
  rows, err := app.db.Query(query)
  if err != nil {
    fmt.Println(err)
    return []Doc{}, err
  }
  dcs := []Doc{}
  for rows.Next() {
    var d Doc
    rows.Scan(&d.Name, &d.Value)
    dcs = append(dcs, d)
  }
  return dcs, nil
}

func createPassrfWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createPassrf", nil)
    if err != nil {
      fmt.Println(err)
      return
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      fmt.Println(err)
    }
    sn := r.FormValue("serialnum")
    wn := r.FormValue("date")
    wm := r.FormValue("whom")
    cd := r.FormValue("code")
    wd := app.web.word
    ps := &PassRF{purecrypt.Symcode(sn, wd), purecrypt.Symcode(wn, wd),
      purecrypt.Symcode(wm, wd), purecrypt.Symcode(cd, wd)}
    err = app.AddPassrfToDb(ps)
    if err != nil {
        fmt.Println(" Error when adding psrf to db")
        fmt.Println(err)
    }
    http.Redirect(w, r, "/home", 303)
  }
}

func (app *App) GetPassrfFromDb() (PassRF, error) {
  query := `select * from passrf`
  rows, err := app.db.Query(query)
  if err != nil {
    return PassRF{}, err
  }
  var p PassRF
  for rows.Next() {
    var tp PassRF
    rows.Scan(&tp.SerialNum, &tp.Date, &tp.Whom, &tp.Code)
    p = tp
  }
  return p, nil
}

func showPassrfWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
  }
  p, err := app.GetPassrfFromDb()
  wd := app.web.word
  dp := PassRF{purecrypt.Desymcode(p.SerialNum, wd), purecrypt.Desymcode(p.Date, wd), 
    purecrypt.Desymcode(p.Whom, wd), purecrypt.Desymcode(p.Code, wd)}
    if dp.SerialNum == "" {
      http.Redirect(w, r, "/createPassrf", 302)
      return
    }
  err = app.execTempl(w, "passrf", dp)
  if err != nil {
    fmt.Println(err)
  }
}

