package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"guts.ubuntu.com/v2/api"
	"guts.ubuntu.com/v2/utils"
)

func main() { // coverage-ignore
	router := gin.Default()
	router.GET("/job/:uuid", api.JobEndpoint)
	router.GET("/artifacts/:uuid/results.tar.gz", api.ArtifactsEndpoint)
	router.POST("/request/", api.RequestEndpoint)
	args := api.ParseArgs()
	GutsCfg, err := api.ParseConfig(args.ConfigFilePath)
	utils.CheckError(err)
	router_address := fmt.Sprintf("%v:%v", GutsCfg.Api.Hostname, GutsCfg.Api.Port)
	err = router.Run(router_address)
	utils.CheckError(err)
}
