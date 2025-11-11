package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"guts.ubuntu.com/v2/utils"
	"log"
	"net/http"
)

// ignore coverage here - it's not smart enough for gin contexts
func RequestEndpoint(c *gin.Context) { // coverage-ignore
	_, Driver, args, err := Setup()
	utils.CheckError(err)
	bareKey := c.GetHeader("X-Api-Key")
	var jobReq JobRequest
	if err := c.BindJSON(&jobReq); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Internal server error:\n%v", err.Error())})
		return
	}
	log.Printf("Processing job request: %v\n", jobReq)
	retJson, err := ProcessJobRequest(args.ConfigFilePath, bareKey, jobReq, Driver)
	if err != nil {
		log.Printf("job request failed! error:\n%v", err.Error())
		switch t := err.(type) {
		default: // coverage-ignore
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error of type %v:\n%v", t, err.Error())})
		case EmptyApiKeyError:
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		case ApiKeyNotAcceptedError:
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		case BadUrlError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		case InvalidArtifactTypeError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		case NonWhitelistedDomainError:
			c.IndentedJSON(http.StatusForbidden, gin.H{"message": err.Error()})
		case utils.GenericGitError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		case PlanFileNonexistentError:
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
		return
	}
	c.IndentedJSON(http.StatusOK, retJson)
}

// ignore coverage here - it's not smart enough for gin contexts
func JobEndpoint(c *gin.Context) { // coverage-ignore
	_, Driver, _, err := Setup()
	utils.CheckError(err)
	uuid := c.Param("uuid")
  log.Printf("finding uuid: %v", uuid)
	err = utils.ValidateUuid(uuid)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	job, err := GetCompleteResultsForUuid(uuid, Driver)
	if err != nil {
    log.Printf("error:\n%v", err.Error())
		switch t := err.(type) {
		default: // coverage-ignore
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error of type %v:\n%v", t, err.Error())})
		case UuidNotFoundError:
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		}
	}
	c.IndentedJSON(http.StatusOK, job.ToJson())
}

// ignore coverage here - it's not smart enough for gin contexts
func ArtifactsEndpoint(c *gin.Context) { // coverage-ignore
	GutsCfg, Driver, _, err := Setup()
	utils.CheckError(err)
	uuid := c.Param("uuid")
	err = utils.ValidateUuid(uuid)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	artifactsTarGz, err := CollateArtifacts(uuid, Driver, GutsCfg)
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
