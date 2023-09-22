package database

import (
  "os"
  "fmt"
  "database/sql"
  "github.com/yurajp/bridge/config"
  "github.com/melbahja/goph"
  _ "github.com/mattn/go-sqlite3"
)

var (
  locpath = "database/remote.db"
  lockdb = config.Conf.Db.SqltPath
)

func getRemoteDb() error {
  if _, err := os.Stat(locpath); err != nil {
    if os.IsNotExist(err) {
      f, er := os.Create(locpath)
      if er != nil {
        return er
      }
      defer f.Close()
    }
  }
  auth, err := goph.Key(config.Conf.Remote.KeyPath, "")
  if err != nil {
    return fmt.Errorf("key: %s", err)
  }
	client, err := goph.New(config.Conf.Remote.User, config.Conf.Remote.Addr, auth)
	if err != nil {
		return fmt.Errorf("ssh: %s", err)
	}
	defer client.Close()
  
  err = client.Download(config.Conf.Remote.DbPath , locpath)
  if err != nil {
    return fmt.Errorf("download db: %s", err)
  }
  return nil
}

func loadLinks(name string) ([]Link, error) {
  lL := []Link{}
  db, err := sql.Open("sqlite3", fmt.Sprintf("database/%s.db", name))
  if err != nil {
    return lL, fmt.Errorf("open %s db: %s", name, err)
  }
  defer db.Close()
  lkTable = config.Conf.Db.SqltTable
  query := "SELECT * FROM " + lkTable
  rows, err := db.Query(query)
  if err != nil {
    return []Link{}, fmt.Errorf("db query: %s", err)
  }
  defer rows.Close()
  for rows.Next() {
    var lk Link
    rows.Scan(&lk.Title, &lk.Url, &lk.Date)
    lL = append(lL, lk)
  }
  return lL, nil
}

func missingLinks() ([]Link, int, int, error) {
  mls := []Link{}
  err := getRemoteDb()
  if err != nil {
    return mls, 0, 0, err
  }
  rems, err := loadLinks("remote")
  if err != nil {
    return mls, 0, 0, err
  }
  fmt.Printf(" %d links in remote DB\n", len(rems))
  locs, err := loadLinks("bridge")
  if err!= nil {
    return mls, 0, 0, err
  }
  fmt.Printf(" %d links in local DB\n", len(locs))
  Remote:
  for _, rlk := range rems {
    for _, llk := range locs {
      if llk.Url == rlk.Url {
        continue Remote
      }
    }
    mls = append(mls, rlk)
  }
  return mls, len(rems), len(locs), nil
}

func completeDb(ls []Link) error {
  if len(ls) == 0 {
    fmt.Println("No missing links")
    return nil
  }
  db, err := sql.Open("sqlite3", config.Conf.Db.SqltPath)
  if err != nil {
    return fmt.Errorf("open bridge.db: %s", err)
  }
  defer db.Close()
  cmd := fmt.Sprintf(`INSERT INTO %s VALUES (?, ?, ?)`, lkTable)
  n := 0
  for i, lk := range ls {
    _, er := db.Exec(cmd, lk.Title, lk.Url, lk.Date)
    if er != nil {
      return fmt.Errorf("when insert into db: %s", er)
    }
    n = i
  }
  wd := "s were"
  if n == 1 {
    wd = " was"
  }
  fmt.Printf(" %d missing link%s added \n", n, wd)
  return nil
}

func SyncerDb() error {
  err := PrepareDb()
  if err != nil {
    return fmt.Errorf("prepareDb: %s", err)
  }
  missing, nr, nl, err := missingLinks()
  if err != nil {
    return fmt.Errorf("missingLinks: %s", err)
  }
  if len(missing) != 0 {
    err = completeDb(missing)
    if err != nil {
      return fmt.Errorf("completeDb: %s", err)
    }
    if len(missing)+nl > nr {
      fmt.Println(" Should you upload joined database to remote machine? \n    [y/n]")
      var y string
      fmt.Scanf("%s", &y)
      if y == "y" {
        err = UploadDb()
        if err != nil {
          return fmt.Errorf("uploadDb: %s", err)
        }
      }
    }
  } else {
    fmt.Println(" No missing links in local database")
    if nl > nr {
      fmt.Println(" But local DB is larger than remote.\n Would you like to upload local DB to remote machine?\n    [y/n]")
      var y string
      fmt.Scanf("%s", &y)
      if y == "y" {
        err = UploadDb()
        if err != nil {
          return fmt.Errorf("uploadDb: %s", err)
        }
      }
    }
  }
  return nil
}

func UploadDb() error {
  auth, err := goph.Key(config.Conf.Remote.KeyPath, "")
  if err != nil {
    return fmt.Errorf("key: %s", err)
  }
  client, err := goph.New(config.Conf.Remote.User, config.Conf.Remote.Addr, auth)
  if err != nil {
    return fmt.Errorf("goph client: %s", err)
  }
  err = client.Upload(config.Conf.Db.SqltPath, config.Conf.Remote.DbPath)
  if err != nil {
    return fmt.Errorf("client upload: %s", err)
  }
  fmt.Println(" Database was uploaded")
  return nil
}