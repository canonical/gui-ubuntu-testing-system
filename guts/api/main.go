package main

import (
  "fmt"
  "net/http"
  "github.com/gin-gonic/gin"
  "log"
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
  CheckError(err)
  defer DeferredErrCheck(db.Close)
  router := gin.Default()
  router.GET("/job/:uuid", JobEndpoint)
  router.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
  router_address := fmt.Sprintf("%v:%v", gutsCfg.Api.Hostname, gutsCfg.Api.Port)
  err = router.Run(router_address)
  CheckError(err)
}

