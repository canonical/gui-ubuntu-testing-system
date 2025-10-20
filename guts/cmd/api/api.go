package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"guts.ubuntu.com/v2/api"
	"guts.ubuntu.com/v2/utils"
	"net/http"
)

// ignore coverage here - it's not smart enough for gin contexts
func RequestEndpoint(c *gin.Context) { // coverage-ignore
	_, Driver, args, err := api.Setup()
	utils.CheckError(err)
	bareKey := c.GetHeader("X-Api-Key")
	var jobReq api.JobRequest
	if err := c.BindJSON(&jobReq); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Internal server error:\n%v", err.Error())})
		return
	}
	retJson, err := api.ProcessJobRequest(args.ConfigFilePath, bareKey, jobReq, Driver)
	if err != nil {
		switch t := err.(type) {
		default: // coverage-ignore
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error of type %v:\n%v", t, err.Error())})
		case api.EmptyApiKeyError:
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		case api.ApiKeyNotAcceptedError:
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		case api.BadUrlError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		case api.InvalidArtifactTypeError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		case api.NonWhitelistedDomainError:
			c.IndentedJSON(http.StatusForbidden, gin.H{"message": err.Error()})
		case utils.GenericGitError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		case api.PlanFileNonexistentError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
		return
	}
	c.IndentedJSON(http.StatusOK, retJson)
}

// ignore coverage here - it's not smart enough for gin contexts
func JobEndpoint(c *gin.Context) { // coverage-ignore
	_, Driver, _, err := api.Setup()
	utils.CheckError(err)
	uuid := c.Param("uuid")
	err = utils.ValidateUuid(uuid)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	job, err := api.GetCompleteResultsForUuid(uuid, Driver)
	if err != nil {
		switch t := err.(type) {
		default: // coverage-ignore
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error of type %v:\n%v", t, err.Error())})
		case api.UuidNotFoundError:
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		}
	}
	c.IndentedJSON(http.StatusOK, job.ToJson())
}

// ignore coverage here - it's not smart enough for gin contexts
func ArtifactsEndpoint(c *gin.Context) { // coverage-ignore
	GutsCfg, Driver, _, err := api.Setup()
	utils.CheckError(err)
	uuid := c.Param("uuid")
	err = utils.ValidateUuid(uuid)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	artifactsTarGz, err := api.CollateArtifacts(uuid, Driver, GutsCfg)
	if err != nil {
		switch t := err.(type) {
		default: // coverage-ignore
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error of type %v:\n%v", t, err.Error())})
		case api.UuidNotFoundError:
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		}
	}
	c.Data(http.StatusOK, "application/x-tar", artifactsTarGz)
}

func main() { // coverage-ignore
	router := gin.Default()
	router.GET("/job/:uuid", JobEndpoint)
	router.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
	router.POST("/request/", RequestEndpoint)
	args := api.ParseArgs()
	GutsCfg, err := api.ParseConfig(args.ConfigFilePath)
	utils.CheckError(err)
	router_address := fmt.Sprintf("%v:%v", GutsCfg.Api.Hostname, GutsCfg.Api.Port)
	err = router.Run(router_address)
	utils.CheckError(err)
}
