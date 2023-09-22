package main

import (
    "fmt"
    _ "github.com/mattn/go-sqlite3"
    "regexp"
    "net/http"
    "html/template"
    "sort"
    "log"
    "strings"
    "strconv"
    "context"

    "github.com/yurajp/wallx/purecrypt"
)

type CardName struct {
  Name string
  Num string
}

func allCardsWeb(w http.ResponseWriter, r *http.Request) {
  ctx, _ := context.WithTimeout(context.Background(), livetime)
 // defer stop()
  
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
  query := `select name, number from cards`
  rows, err := app.db.Query(query)
  if err != nil {
    log.Println(err)
  }
  list := []CardName{}
  for rows.Next() {
    var cn CardName
    rows.Scan(&cn.Name, &cn.Num)
    list = append(list, cn)
  }
  names := []template.HTML{}
  sort.Slice(list, func(i, j int) bool { 
    return strings.ToLower(list[i].Name)[0] < strings.ToLower(list[j].Name)[0] 
  })
  for _, c := range list {
    names = append(names, makeCardLink(c))
  }
  err = app.execTempl(w, "allCards", names)
  if err != nil {
    fmt.Println(err)
  }
}

func cleanNum(n string) string {
  return strings.Replace(n, " ", "", -1)
}

func createCardWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  mistake := ""
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createCard", mistake)
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
    number := r.FormValue("number")
    month := r.FormValue("month")
    year := r.FormValue("year")
    expire := fmt.Sprintf("%s / %s", month, year)
    cvc := r.FormValue("cvc")
    c := &Card{name, number, expire, cvc}
    mistake = c.CheckCard()
    if mistake != "" {
      err = app.execTempl(w, "createCard", mistake)
      if err != nil {
        fmt.Println(err)
      }
      return
    }
    encNum := purecrypt.Symcode(cleanNum(c.Number), app.web.word)
    encExp := purecrypt.Symcode(c.Expire, app.web.word)
    encCvc := purecrypt.Symcode(c.Cvc, app.web.word)
    ccr := Card{name, encNum, encExp, encCvc}
    err = app.AddCardToDb(ccr)
    if err != nil {
      fmt.Println(err)
    }
    http.Redirect(w, r, "/cards", 303)
  } 
}

func (c *Card) CheckCard() string {
  if !checkNum(c.Number) {
    return "INCORRECT CARD NUMBER"
  }
  if !checkDate(strings.Split(c.Expire, " / ")) {
    return "INCORRECT EXPIRE DATE"
  }
  if !checkCvc(c.Cvc) {
    return "INCORRECT CVC"
  }
  return ""
}
  
func makeCardLink(cn CardName) template.HTML {
  if cn.Num == "" {
    return template.HTML("")
  }
  url := fmt.Sprintf(`<a href="http://localhost:8686/card?name=%s">%s</a>`, cn.Name, cn.Name)
  dcN := purecrypt.Desymcode(cn.Num, app.web.word)
  shN := "*" + dcN[12:]
  span := fmt.Sprintf(`<span>%s</span>`, shN)
  return template.HTML(url + span)
}

func spaceNum(n string) string {
  s := " "
  return n[:4] + s + n[4:8] + s + n[8:12] + s + n[12:]
}

func showCardWeb(w http.ResponseWriter, r *http.Request) {
  ctx, _ := context.WithTimeout(context.Background(), livetime)
 // defer stop()
  
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
  c, err := app.GetCardFromDb(name)
  if err != nil {
    log.Println(err)
  }
  dn := purecrypt.Desymcode(c.Number, app.web.word)
  de := purecrypt.Desymcode(c.Expire, app.web.word)
  dc := purecrypt.Desymcode(c.Cvc, app.web.word)
  dcd := Card{c.Name, spaceNum(dn), de, dc}
  err = app.execTempl(w, "oneCard", dcd)
  if err != nil {
    fmt.Println(err)
  }
}

func (app *App) AddCardToDb(c Card) error {
    query := `insert into cards(name, number, expire, cvc) values(?, ?, ?, ?) on conflict(name) do update set number=excluded.number, expire=excluded.expire, cvc=excluded.cvc`
    _, err := app.db.Exec(query, c.Name, c.Number, c.Expire, c.Cvc)
    if err != nil {
      return err
    }
    return nil
}

func makeExpire(n string) (string, bool) {
  re := regexp.MustCompile(`\d\d[/ -\.]\d\d`)
  if !re.MatchString(n) {
    return "", false
  }
  sp := regexp.MustCompile(`[\./ -]`)
  exs := sp.Split(n, -1)
  return fmt.Sprintf("%s / %s", exs[0], exs[1]), true
}
 
func (app *App) GetCardFromDb(q string) (Card, error) {
  query := `select name, number, expire, cvc from cards where name=?`
  row := app.db.QueryRow(query, q)
  var c Card
  err := row.Scan(&c.Name, &c.Number, &c.Expire, &c.Cvc)
  if err != nil {
    return Card{}, err
  }
  return c, nil
}

func deleteCardWeb(w http.ResponseWriter, r *http.Request) {
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
  err := app.RemoveCardFromDb(strings.Trim(name, "\""))
  if err != nil {
    panic(err)
  }
  err = app.execTempl(w, "delCard", name)
  if err != nil {
    fmt.Println(err)
  }
}

func (app *App) RemoveCardFromDb(c string) error {
  query := `delete from cards where name=?`
  _, err := app.db.Exec(query, c)
  if err != nil {
    return err
  }
  return nil
}

func checkNum(n string) bool {
  re := regexp.MustCompile(`\d{4}\s?\d{4}\s?\d{4}\s?\d{4}`)
  return re.MatchString(n)
}  

func checkCvc(n string) bool {
  re := regexp.MustCompile(`\d\d\d`)
  return re.MatchString(n)
}

func checkDate(my []string) bool {
  if len(my) != 2 {
    return false
  }
  re := regexp.MustCompile(`\d\d`)
  if !re.MatchString(my[0]) || !re.MatchString(my[1]) {
    return false
  }
  dm, _ := strconv.Atoi(my[0])
  dy, _ := strconv.Atoi(my[1])
  if dm < 1 || dm > 12 || dy < 23 || dy > 35 {
    return false
  }
  return true
}
