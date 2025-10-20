package spawner

import (
	"flag"
	"log"
	"os"
)

var (
	VncHost        string
	VncPort        uint
	ConfigFilePath string
)

func ParseArgs() { // coverage-ignore
	// Vnc options
	if flag.Lookup("host") == nil {
		flag.StringVar(&VncHost, "host", "", "Vnc host to advertise to the runner")
	}
	if flag.Lookup("port") == nil {
		flag.UintVar(&VncPort, "port", 0, "Vnc port to advertise to the runner")
	}

	// Config file path
	cfgPathFromEnv := os.Getenv("GUTS_CFG_PATH")
	if cfgPathFromEnv == "" {
		cfgPathFromEnv = "./guts-spawner.yaml"
	}
	if flag.Lookup("cfg-path") == nil {
		flag.StringVar(&ConfigFilePath, "cfg-path", cfgPathFromEnv, "Path to config file for the guts spawner")
	}

	// Parse
	flag.Parse()

	// Ensure required params are present
	if VncHost == "" {
		log.Fatal("-host arg must be set")
	}
	if VncPort == 0 {
		log.Fatal("-port arg must be set")
	}
}
