package database

import (
  "database/sql"
  "os"
  "fmt"
  "time"
  "github.com/yurajp/bridge/config"
  
  _ "github.com/mattn/go-sqlite3"
  _ "github.com/lib/pq"
)

type SqLink struct {
  Title string
  Url string
  Date string
}

type PgLink struct {
  Id int
  Title string
  Link string
  Date time.Time
}

var (
  sqltPath string
  sqltTable string
  pgHost string
  pgPort string
  pgUser string
  pgPswd string
  pgName string
  pgTable string
)


func (pl PgLink) ToSqlite() SqLink {
  sdate := pl.Date.Format("2006-01-02")
  return SqLink{pl.Title, pl.Link, sdate}
}

func PrepareSqlt() error {
  sqltPath = config.Conf.Db.SqltPath
  sqltTable = config.Conf.Db.SqltTable
  pgHost = config.Conf.Db.PgHost
  pgPort = config.Conf.Db.PgPort
  pgUser = config.Conf.Db.PgUser
  pgPswd = config.Conf.Db.PgPswd
  pgName = config.Conf.Db.PgName
  pgTable = config.Conf.Db.PgTable
  if _, err := os.Stat(sqltPath); err == nil {
    return nil
  }
  db, err := sql.Open("sqlite3", sqltPath)
  if err != nil {
    return err
  }
  defer db.Close()
  create := fmt.Sprintf(`create table if not exists %s (title text, url text, date text) without rowid`, sqltTable)
  _, err = db.Exec(create)
  if err != nil {
    return err
  }
  return nil
}

func MigratePgToSqlt() error {
  err := PrepareSqlt()
  if err != nil {
    return fmt.Errorf("Prepare: %w", err)
  }
  pgConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPswd, pgName)
	
	pgdb, err := sql.Open("postgres", pgConn)
	if err != nil {
	  return fmt.Errorf("Postgres: %w", err)
	}
	defer pgdb.Close()
	sqLinks := []SqLink{}
	query := fmt.Sprintf(`select * from %s`, pgTable)
	rows, err := pgdb.Query(query)
	if err != nil {
	  return fmt.Errorf("Postgres query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
	  var pl PgLink
	  err = rows.Scan(&pl.Id, &pl.Title, &pl.Link, &pl.Date)
	  if err != nil {
	    return fmt.Errorf("Postgres rows scan: %w", err)
	  }
	  sl := pl.ToSqlite()
	  sqLinks = append(sqLinks, sl)
	}
	if len(sqLinks) == 0 {
	  return nil
	}
	sdb, err := sql.Open("sqlite3", sqltPath)
	if err != nil {
	  return fmt.Errorf("Open sqlite db: %w", err)
	}
	insStat := fmt.Sprintf(`insert into %s values (?, ?, ?)`, sqltTable)
	for _, slk := range sqLinks {
	   _, err := sdb.Exec(insStat, slk.Title, slk.Url, slk.Date)
	   if err != nil {
	     return fmt.Errorf("Insert into sqlite db: %w", err)
	   }
	 }
	 return nil
}