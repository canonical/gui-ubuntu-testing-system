package main

import (
  "flag"
  "database/sql"
  "os"
)

var (
  configFilePath string
)

func ParseArgs() {
  cfgPathFromEnv := os.GetEnv("GUTS_CFG_PATH")
  if cfgPathFromEnv == "" {
    cfgPathFromEnv = "/etc/guts-api.yaml"
  }
  flag.StringVar(&configFilePath, "cfg-path", cfgPathFromEnv, "Path to config file for the guts api")
  flag.Parse()
}

