package main

import (
  "database/sql"
  "fmt"
  _ "github.com/lib/pq"
)

func PostgresConnect(config GutsApiConfig) *sql.DB {
  ConnectString := config.PostgresConnectString()
  db, err := sql.Open("postgres", ConnectString)
  CheckError(err)
  return db
}

func (g GutsApiConfig) PostgresConnectString() string {
  psqlInfo := fmt.Sprintf("host=%s ", g.Postgres.Host)
  psqlInfo += fmt.Sprintf("port=%d ", g.Postgres.Port)
  psqlInfo += fmt.Sprintf("user=%s ", g.Postgres.User)
  psqlInfo += fmt.Sprintf("password=%s ", g.Postgres.Password)
  psqlInfo += fmt.Sprintf("dbname=%s ", g.Postgres.DbName)
  psqlInfo += fmt.Sprintf("sslmode=%s", "disable")
  return psqlInfo
}
