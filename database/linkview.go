package database

import (
  "database/sql"
  "strings"
  "github.com/yurajp/bridge/config"
  
	_ "github.com/mattn/go-sqlite3"
)

type LinkView struct {
  Source string
  Title string
  Url string
}


func lkShort(tt string) string {
  return strings.Split(tt, " |")[0]
}

func noMedium(u string) string {
  spl := strings.Split(u, ".")
  if len(spl) > 1 && spl[len(spl) - 2] == "medium" {
    return "NO ACCESS (medium.com)"
  }
  return u
}

func lkSource(lk string) string {
  nm := strings.Split(strings.TrimPrefix(lk, "https://"), "/")[0]
  return strings.TrimPrefix(noMedium(nm), "www.")
}

func MakeView() ([]LinkView, error) {
  dbFile = config.Conf.Db.SqltPath
  lkTable = config.Conf.Db.SqltTable
  db, err := sql.Open("sqlite3", dbFile)
  if err != nil {
    return []LinkView{}, err
  }
  defer db.Close()
  query := "SELECT * FROM " + lkTable + " ORDER BY date DESC"
  rows, err := db.Query(query)
  if err != nil {
    return []LinkView{}, err
  }
  defer rows.Close()
  lvs := []LinkView{}
  for rows.Next() {
    var lk Link
    rows.Scan(&lk.Title, &lk.Url, &lk.Date)
    lv := LinkView{lkSource(lk.Url), lkShort(lk.Title), lk.Url}
    lvs = append(lvs, lv)
  }
  return lvs, nil
}

