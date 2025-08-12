package main

import (
  "flag"
  "database/sql"
)

var (
  configFilePath string
  db *sql.DB
)

func ParseArgs() {
  flag.StringVar(&configFilePath, "cfg-path", "./guts-api.yaml", "Path to config file for the guts api")
  flag.Parse()
}

