package scheduler

import (
	"flag"
	"os"
)

func ParseArgs() string { // coverage-ignore
	var configFilePath string
	// Config file path
	cfgPathFromEnv := os.Getenv("GUTS_SCHEDULER_CFG_PATH")
	if cfgPathFromEnv == "" {
		cfgPathFromEnv = "./guts-scheduler-local.yaml"
	}
	if flag.Lookup("cfg-path") == nil {
		flag.StringVar(&configFilePath, "cfg-path", cfgPathFromEnv, "Path to config file for the guts runner")
	}

	// Parse
	flag.Parse()

	return configFilePath
}
