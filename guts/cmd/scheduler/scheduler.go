package main

import (
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/scheduler"
	"guts.ubuntu.com/v2/utils"
	"math/rand/v2"
	"time"
)

func main() { // coverage-ignore
	cfgPath := scheduler.ParseArgs()

	schedulerCfg, err := scheduler.ParseConfig(cfgPath)
	utils.CheckError(err)

	Driver, err := database.NewDbDriver(schedulerCfg.Database.Driver, schedulerCfg.Database.ConnectionString)
	utils.CheckError(err)

	for {
		// perform the regular loop
		err = scheduler.SchedulerLoop(Driver, schedulerCfg)
		utils.CheckError(err)
		// wait somewhere between 15 and 45 seconds before starting the main scheduler loop again
		pollSleepDuration := time.Second * time.Duration(rand.IntN(30)+15)
		time.Sleep(pollSleepDuration)
	}
}
