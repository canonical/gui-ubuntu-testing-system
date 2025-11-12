package main

import (
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/runner"
	"guts.ubuntu.com/v2/utils"
	"math/rand/v2"
  "log"
	"time"
)

func main() { // coverage-ignore
  log.Printf("starting guts runner...")
	// parse config path
	cfgPath := runner.ParseArgs()
  log.Printf("loaded config path: %v", cfgPath)

	// parse the runner config
	RunnerCfg, err := runner.ParseConfig(cfgPath)
	utils.CheckError(err)
  log.Printf("loaded config:\n%v", RunnerCfg)

	// Initialise the database driver
	Driver, err := database.NewDbDriver(RunnerCfg.Database.Driver, RunnerCfg.Database.ConnectionString)
	utils.CheckError(err)

	for {
		// perform the regular loop
    log.Printf("performing runner loop...")
		err = runner.RunnerLoop(Driver, RunnerCfg)
		utils.CheckError(err)
    log.Printf("runner loop complete...")
		// wait somewhere between 30 and 90 seconds before checking for new jobs
		pollSleepDuration := time.Second * time.Duration(rand.IntN(60)+30)
    log.Printf("sleeping for %v seconds...", pollSleepDuration)
		time.Sleep(pollSleepDuration)
	}
}
