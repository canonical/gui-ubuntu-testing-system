package main

import (
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/runner"
	"guts.ubuntu.com/v2/utils"
	"math/rand/v2"
	"time"
)

func main() { // coverage-ignore
	// parse config path
	cfgPath := runner.ParseArgs()

	// parse the runner config
	RunnerCfg, err := runner.ParseConfig(cfgPath)
	utils.CheckError(err)

	// Initialise the database driver
	Driver, err := database.NewDbDriver(RunnerCfg.Database.Driver, RunnerCfg.Database.ConnectionString)
	utils.CheckError(err)

	for {
		// perform the regular loop
		err = runner.RunnerLoop(Driver, RunnerCfg)
		utils.CheckError(err)
		// wait somewhere between 30 and 90 seconds before checking for new jobs
		pollSleepDuration := time.Second * time.Duration(rand.IntN(60)+30)
		time.Sleep(pollSleepDuration)
	}
}
