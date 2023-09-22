package database

import (
	"fmt"
	"os"
	"net/http"
	"regexp"
	"strings"
	"time"
	"bufio"
	"database/sql"
  "github.com/yurajp/bridge/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/PuerkitoBio/goquery"
)

type Link struct {
  Title string
  Url string
  Date string
}

var (
  dbFile string
  lkTable string
)

func PrepareDb() error {
  dbFile = config.Conf.Db.SqltPath
  lkTable = config.Conf.Db.SqltTable
  if _, err := os.Stat(dbFile); err == nil {
    return nil
  }
  db, err := sql.Open("sqlite3", dbFile)
  if err != nil {
    return err
  }
  defer db.Close()
  create := fmt.Sprintf(`create table if not exists %s (title text, url text, date text)`, lkTable)
  _, err = db.Exec(create)
  if err != nil {
    return err
  }
  return nil
}

func ScrapeTitle(url string) string {
  cl := http.Client{Timeout: 10 * time.Second}
  resp, err := cl.Get(url)
  if err != nil {
    return ""
  }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    return ""
  }
  doc, err := goquery.NewDocumentFromReader(resp.Body)
  if err != nil {
    return ""
  }
  var title string
	doc.Find("head").Each(func(_ int, s *goquery.Selection) {
		title = s.Find("title").Text()
	})
	return title
}

func LinkScanner(text string) (int, error) {
  url := regexp.MustCompile(`http(s)?://*`)
  sc := bufio.NewScanner(strings.NewReader(text))
  linksDb := []Link{}
  nosuccess := 0
  for sc.Scan() {
    line := sc.Text()
    if url.MatchString(line) {
      if title := ScrapeTitle(line); title != "" {
        link := Link{title, line, time.Now().Format("2006-01-02")}
        linksDb = append(linksDb, link)
      } else {
        nosuccess++
      }
    }
  }
  if nosuccess > 0 {
    fmt.Printf("  %d link(s) will NOT be stored\n", nosuccess)
  }
  if len(linksDb) == 0 {
    return 0, nil
  }
  return handleDb(linksDb)
}


func handleDb(links []Link) (int, error) {
  dbFile = config.Conf.Db.SqltPath
  lkTable = config.Conf.Db.SqltTable
  err := PrepareDb()
  if err != nil {
    return 0, fmt.Errorf("Error when creating table in db: %w", err)
  }
  db, err := sql.Open("sqlite3", dbFile)
  if err != nil {
    return 0, fmt.Errorf("Cannot open database: %w", err)
  }
  defer db.Close()
  dedup := fmt.Sprintf(`DELETE FROM %s WHERE url = ?`, lkTable)
  insert := fmt.Sprintf(`INSERT INTO %s VALUES(?, ?, ?)`, lkTable)
  for _, lk := range links {
    _, err = db.Exec(dedup, lk.Url)
    if err != nil {
      return 0, fmt.Errorf("Error when delete duplicate from db: %w", err)
    }
    _, err = db.Exec(insert, lk.Title, lk.Url, lk.Date)
    if err != nil {
      return 0, fmt.Errorf("Error when insert into db: %w", err)
    }
  }
  return len(links), nil
}