package main

import (
  "database/sql"
  "fmt"
  "reflect"
  "os/exec"
  _ "github.com/lib/pq"
)

func CheckPostgresServiceUp() error {
  systemctlCommand := "systemctl status postgresql.service"
  systemctlCommand := exec.Command("systemctl", "status", "postgresql.service")
  if err := systemctlCommand.Run(); err != nil {
    return PostgresServiceNotUpError()
  }
  return nil
}

func PostgresConnect() (*sql.DB, error) {
  err := CheckPostgresServiceUp()
  if err != nil {
    return err
  }
  ConnectString := GutsCfg.PostgresConnectString()
  db, err := sql.Open("postgres", ConnectString)
  return db, err
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
