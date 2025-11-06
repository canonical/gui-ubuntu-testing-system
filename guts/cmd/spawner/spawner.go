package main

import (
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/spawner"
	"guts.ubuntu.com/v2/utils"
	"math/rand/v2"
	"time"
)

func main() { // coverage-ignore
	spawner.ParseArgs()
	SpawnerCfg, err := spawner.ParseConfig(spawner.ConfigFilePath)
	utils.CheckError(err)
	err = spawner.CreateCacheIfNotExists(SpawnerCfg)
	utils.CheckError(err)
	Driver, err := database.NewDbDriver(SpawnerCfg.Database.Driver, SpawnerCfg.Database.ConnectionString)
	utils.CheckError(err)

	for {
		// perform the regular loop
		err = spawner.SpawnerLoop(Driver, SpawnerCfg)
		utils.CheckError(err)
		// wait somewhere between 30 and 90 seconds before checking for new jobs
		pollSleepDuration := time.Second * time.Duration(rand.IntN(10)+5)
		time.Sleep(pollSleepDuration)
	}
}
