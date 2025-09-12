package main

import (
	"flag"
	"os"
)

var (
	configFilePath string
)

func ParseArgs() {
	cfgPathFromEnv := os.Getenv("GUTS_CFG_PATH")
	if cfgPathFromEnv == "" {
		cfgPathFromEnv = "./guts-api.yaml"
	}

	if flag.Lookup("cfg-path") == nil {
		flag.StringVar(&configFilePath, "cfg-path", cfgPathFromEnv, "Path to config file for the guts api")
		flag.Parse()
	}
}
