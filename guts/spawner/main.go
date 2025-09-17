package main

import (
	"fmt"
	"log"
  "time"
  "math/rand/v2"
)

func CheckError(err error) { // coverage-ignore
	if err != nil {
		log.Fatal(err.Error())
	}
}

func DeferredErrCheck(f func() error) {
	err := f()
	CheckError(err)
}

func main() {
	ParseArgs()
	fmt.Println(VncHost)
	fmt.Println(VncPort)
	err := ParseConfig(ConfigFilePath)
	CheckError(err)
  err = CreateCacheIfNotExists()
	CheckError(err)
	fmt.Println(SpawnerCfg)
	Driver, err = NewDbDriver(SpawnerCfg)
	CheckError(err)
	fmt.Println(Driver)

  for {
    // perform the regular loop
    err = SpawnerLoop()
    // wait somewhere between 30 and 90 seconds before checking for new jobs
    pollSleepDuration := time.Second * (rand.IntN(60) + 30)
    time.Sleep(pollSleepDuration)
  }
}
