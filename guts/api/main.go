package main

import (
  "fmt"
  "net/http"
  "github.com/gin-gonic/gin"
  "log"
  "guts_db"
)

var (
  Driver DbDriver
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

func DeferredErrCheckStringArg(f func(s string) error, s string) {
  err := f(s)
  CheckError(err)
}

func RequestEndpoint(c *gin.Context) {
  bareKey := c.GetHeader("X-Api-Key")
  var jobReq JobRequest
  if err := c.BindJson(&jobReq); err != nil {
    c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error:\n%v", err.Error())})
  }
  retJson, err := ProcessJobRequest(bareKey, jobReq)
  if err != nil {
    switch t := err.(type) {
    default:
      c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error:\n%v", err.Error())})
    case *EmptyApiKeyError:
      c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
    case: *ApiKeyNotAcceptedError:
      c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
    case *url.Error:
      c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
    case: *BadUrlError:
      c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
    case: *NonWhitelistedDomainError:
      c.IndentedJSON(http.StatusForbidden, gin.H{"message": err.Error()})
    case: *GenericGitError:
      c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
    case: *PlanFileNonexistentError:
      c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
    }
  }
  c.IndentedJSON(http.StatusOK, retJson)
}

func JobEndpoint(c *gin.Context) {
  uuid := c.Param("uuid")
  err := ValidateUuid(uuid)
  if err != nil {
    c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
  }
  job, err := GetCompleteResultsForUuid(uuid, db)
  if err != nil {
    switch t := err.(type) {
    default:
      c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error:\n%v", err.Error())})
    case *UuidNotFoundError:
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
  artifactsTarGz, err := CollateArtifacts(uuid, db)
  if err != nil {
    switch t := err.(type) {
    default:
      c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Internal server error:\n%v", err.Error())})
    case *UuidNotFoundError:
      c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
    }
  }
  c.Data(http.StatusOK, "application/x-tar", artifactsTarGz)
}

func main() { // coverage-ignore
  ParseArgs()
  err := ParseConfig(configFilePath)
  CheckError(err)
  db, err = PostgresConnect()
  Driver, err = NewDbDriver(GutsApiConfig)
  CheckError(err)
  defer DeferredErrCheck(db.Close)
  router := gin.Default()
  router.GET("/job/:uuid", JobEndpoint)
  router.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
  router_address := fmt.Sprintf("%v:%v", GutsCfg.Api.Hostname, GutsCfg.Api.Port)
  err = router.Run(router_address)
  CheckError(err)
}

