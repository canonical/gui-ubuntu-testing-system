package api

import (
	"flag"
	"os"
)

type ApiArgs struct {
	ConfigFilePath string
}

func ParseArgs() ApiArgs {
	var args ApiArgs

	cfgPathFromEnv := os.Getenv("GUTS_CFG_PATH")
	if cfgPathFromEnv == "" { // coverage-ignore
		cfgPathFromEnv = "./guts-api.yaml"
	}

	if flag.Lookup("cfg-path") == nil {
		flag.StringVar(&args.ConfigFilePath, "cfg-path", cfgPathFromEnv, "Path to config file for the guts api")
	} else {
		args.ConfigFilePath = flag.Lookup("cfg-path").Value.(flag.Getter).Get().(string)
	}
	flag.Parse()
	return args
}
