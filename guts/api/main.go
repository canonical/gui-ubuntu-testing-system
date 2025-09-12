package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// var (
//   Driver DbDriver
// )

func CheckError(err error) { // coverage-ignore
	if err != nil {
		log.Fatal(err.Error())
	}
}

func DeferredErrCheck(f func() error) {
	err := f()
	CheckError(err)
}

func DeferredErrCheckStringArg(f func(s string) error, s string) {
	err := f(s)
	CheckError(err)
}

func JobEndpoint(c *gin.Context) {
	uuid := c.Param("uuid")
	err := ValidateUuid(uuid)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	job, err := GetCompleteResultsForUuid(uuid)
	if err != nil {
		switch t := err.(type) {
		default: // coverage-ignore
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error of type %v:\n%v", t, err.Error())})
		case UuidNotFoundError:
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		}
	}
	// not sure if this will marshal the time field properly. need to check
	c.IndentedJSON(http.StatusOK, job.toJson())
}

func ArtifactsEndpoint(c *gin.Context) {
	uuid := c.Param("uuid")
	err := ValidateUuid(uuid)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	artifactsTarGz, err := CollateArtifacts(uuid)
	if err != nil {
		switch t := err.(type) {
		default: // coverage-ignore
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error of type %v:\n%v", t, err.Error())})
		case UuidNotFoundError:
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		}
	}
	c.Data(http.StatusOK, "application/x-tar", artifactsTarGz)
}

func main() { // coverage-ignore
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	Driver, err = NewDbDriver(GutsCfg)
	CheckError(err)
	router := gin.Default()
	router.GET("/job/:uuid", JobEndpoint)
	router.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
	router_address := fmt.Sprintf("%v:%v", GutsCfg.Api.Hostname, GutsCfg.Api.Port)
	err = router.Run(router_address)
	CheckError(err)
}
