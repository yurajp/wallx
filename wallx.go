package main


import (
    "fmt"
    "net/http"
    "log"
    "sort"
    "strings"
    "html/template"
    "database/sql"
    "os/exec"
    "time"
    "context"
    "errors"
  
    _ "github.com/mattn/go-sqlite3"
    "github.com/yurajp/wallx/purecrypt"
    "github.com/yurajp/wallx/conf"
)

type App struct {
  web *Web
  db *sql.DB
}

type Web struct {
  server *http.Server
  ctx context.Context
  templs map[string]*template.Template
  word string
}

type Site struct {
    Name string
    Login string
    Pass string
    Link string
}

type Card struct {
    Name string
    Number string
    Expire string
    Cvc string
}


var app *App
var port string
var livetime time.Duration

func makeTempls() (map[string]*template.Template, error) {
  templs := map[string]*template.Template{}
  routes := []string{"wellcome",
  "home", "allSites", "allCards", "oneSite", "oneCard", "createSite",
  "createCard", "delSite", "delCard", "allDocs", "createDoc", "passrf",
   "createPassrf", "message",
}
  for _, r := range routes {
    t, err := template.ParseFiles("templates/" + r + ".html")
    if err != nil {
      return templs, err
    }
    templs[r] = t
  }
  return templs, nil
}

func NewWeb() *Web{
    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir("./static"))
    mux.Handle("/static/", http.StripPrefix("/static/", fs))
    muxMap := map[string]func(http.ResponseWriter, *http.Request){"/": wellcome,
      "/home": homeHandler,
      "/sites": allSitesWeb,
      "/cards": allCardsWeb,
      "/createSite": createSiteWeb,
      "/createCard": createCardWeb,
      "/site": showSiteWeb,
      "/card": showCardWeb,
      "/deleteSite": deleteSiteWeb,
      "/deleteCard": deleteCardWeb,
      "/docs": showDocsWeb,
      "/createDoc": createDocWeb,
      "/passrf": showPassrfWeb,
      "/createPassrf": createPassrfWeb,
      "/recode": RecodeWeb,
      "/backup": BackupWeb,
      "exit": exitWeb,
    }
    for p, fn := range muxMap {
      mux.HandleFunc(p, fn)
    }
    
    server := &http.Server{
      Addr: port,
      Handler: mux,
    }
    templs, err := makeTempls()
    if err != nil {
      fmt.Println(err)
      return &Web{}
    }
    ctx := context.Background()
    web := Web{server, ctx, templs, ""}
    return &web
}

func check(err error) {
  if err != nil {
    panic(err)
  }
}

func (app *App) IsDead() bool {
  return app.web.word == ""
}

func (app *App) Dies() {
  app.web.word = ""
}

func (app *App) execTempl(w http.ResponseWriter, t string, data any) error {
  if app.IsDead() {
    return errors.New("Password not set")
  }
  if tmp, ok := app.web.templs[t]; ok {
    err := tmp.Execute(w, data)
    if err != nil {
      return err
    }
  } else {
    return errors.New("Template does not exists")
  }
  return nil
}

func wellcome(w http.ResponseWriter, r *http.Request) {
  exists := purecrypt.ChWordExists()
  if !exists {
    fmt.Println("\tINITIAL SETUP")
  }
  if r.Method == http.MethodGet {
    if wc, ok := app.web.templs["wellcome"]; ok {
      wc.Execute(w, exists)
    } 
  }
  if r.Method == http.MethodPost {
    if !exists {
      app.createTables()
      err := r.ParseForm()
      check(err)
      word1 := r.FormValue("word1")
      word2 := r.FormValue("word2")
      if len(word1) < 5 || word1 != word2 {
        http.Redirect(w, r, "/", 302)
        return
      }
      err = purecrypt.WriteCheckword(word1)
      check(err)
      app.web.word = word1
      http.Redirect(w, r, "/home", 302)
    } else {
      err := r.ParseForm()
      check(err)
      word := r.FormValue("word")
      if purecrypt.IsCorrect(word) {
        app.web.word = word
        http.Redirect(w, r, "/home", 303)
      } else {
        fmt.Println("  Password not match!")
        http.Redirect(w, r, "/", 302)
      }
    }
  }
}


func exitWeb(w http.ResponseWriter, r *http.Request) {
  app.web.word = ""
  app.execTempl(w, "/", true)
}


func (app *App) createTables() {
  query1 := `create table if not exists sites(name text primary key, login text, pass text, link text)`
  _, err := app.db.Exec(query1)
  check(err)
  query2 := `create table if not exists cards(name text primary key, number text, expire text, cvc text, unique(name, number))`
  _, err = app.db.Exec(query2)
  check(err)
  query3 := `create table if not exists docs(name text primary key, value text)`
  _, err = app.db.Exec(query3)
  check(err)
  query4 := `create table if not exists passrf(serialnum text primary key, date text, whom text, code text)`
  _, err = app.db.Exec(query4)
  check(err)
}

func deleteSiteWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  ctx, _ := context.WithTimeout(context.Background(), livetime)
  
  go func() {
    for {
      select{
        case <-ctx.Done():
          app.Dies()
          return
        default:
      }
    }
  }()
  name := r.URL.Query().Get("name")
  err := app.RemoveSiteFromDb(strings.Trim(name, "\""))
  if err != nil {
    panic(err)
  }
  err = app.execTempl(w, "deleteSite", name)
  if err != nil {
    fmt.Println(err)
  }
}

func (app *App) RemoveSiteFromDb(s string) error {
  query := `delete from sites where name=?`
  _, err := app.db.Exec(query, s)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) Run() {
    go func() {
        err := app.web.server.ListenAndServe()
        check(err)
    }()
    cmd := exec.Command("epiphany", "http://localhost:8686/")
    go cmd.Run()
    
    var q string
    fmt.Println("\n\t WALLAP RUNNING\n\t Enter any to quit")
    fmt.Scanf("%s", q)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
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
  if app.IsDead() {
    fmt.Println("\t Oops!")
    http.Redirect(w, r, "/", 302)
    return
  }
  err := app.execTempl(w, "home", nil)
  if err != nil {
      fmt.Println(err)
  }
}

func createSiteWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createSite", nil)
    if err != nil {
      fmt.Println(err)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      fmt.Println(err)
    }
    name := r.FormValue("name")
    login := r.FormValue("login")
    pass := r.FormValue("pass")
    link := r.FormValue("link")
    s := &Site{name, purecrypt.Symcode(login, app.web.word), purecrypt.Symcode(pass, app.web.word), link}
    err = app.AddSiteToDb(s)
    if err != nil {
      fmt.Println(err)
    }
    http.Redirect(w, r, "  /sites", 303)
  } 
}

func (app *App) AddSiteToDb(s *Site) error {
    query := `insert into sites(name, login, pass, link) values(?, ?, ?, ?) on conflict(name) do update set pass=excluded.pass, login=excluded.login, link=excluded.link`
    _, err := app.db.Exec(query, s.Name, s.Login, s.Pass, s.Link)
    if err != nil {
      return err
    }
    return nil
}

func showSiteWeb(w http.ResponseWriter, r *http.Request) {
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
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  name := r.URL.Query().Get("name")
  s, err := app.GetSiteFromDb(name)
  if err != nil {
    log.Println(err)
  }
  dl := purecrypt.Desymcode(s.Login, app.web.word)
  dp := purecrypt.Desymcode(s.Pass, app.web.word)
  sw := Site{s.Name, dl, dp, s.Link}
  err = app.execTempl(w, "oneSite", sw)
  if err != nil {
    fmt.Println(err)
  }
}

func (app *App) GetSiteFromDb(q string) (Site, error) {
  query := `select name, login, pass, link from sites where lower(name) like ?`
  row := app.db.QueryRow(query, q + "%")
  var s Site
  err := row.Scan(&s.Name, &s.Login, &s.Pass, &s.Link)
  if err != nil {
    return Site{}, err
  }
  return s, nil
}

func makeSiteLink(nm string) template.HTML {
  addr := app.web.server.Addr
  url := fmt.Sprintf(`<a href="http://localhost%s/site?name=%s">%s</a>`, addr, nm, nm)
  return template.HTML(url)
}

func allSitesWeb(w http.ResponseWriter, r *http.Request) {
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
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  query := `select name from sites`
  rows, err := app.db.Query(query)
  if err != nil {
    log.Println(err)
  }
  list := []string{}
  for rows.Next() {
    var s string
    rows.Scan(&s)
    list = append(list, s)
  }
  names := []template.HTML{}
  sort.Slice(list, func(i, j int) bool { 
    return strings.ToLower(list[i]) < strings.ToLower(list[j]) 
  })
  for _, s := range list {
    names = append(names, makeSiteLink(s))
  }
  err = app.execTempl(w, "allSites", names)
  if err != nil {
    fmt.Println(err)
  }
}


func main() {
    err := conf.Prepare()
    if err != nil {
      fmt.Println(err)
      return
    }
    cfg, err := conf.GetConfig()
    if err != nil {
      fmt.Println(err)
      return
    }
    port = cfg.Port
    livetime = cfg.Livetime

    db, err := sql.Open("sqlite3", "wallx.db")
    if err != nil {
        fmt.Println(err)
       return
    }
    defer db.Close()
    // defer func() {
    //   clpb := exec.Command("termux-clipboard-set", " ")
    //   err := clpb.Run()
    //   if err != nil {
    //     fmt.Println(err)
    //   }
    // }()
    web := NewWeb()
    app = &App{web, db}
    _, cancel := context.WithCancel(app.web.ctx)
    defer cancel()
    
    app.Run()
}

