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
  // modify this!
  // only return 404 when that's the error.
  // when something else goes wrong, we need 500.
  if err != nil {
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
  } else {
    c.IndentedJSON(http.StatusOK, job.toJson())
  }
}

func ArtifactsEndpoint(c *gin.Context) {
  uuid := c.Param("uuid")
  err := ValidateUuid(uuid)
  if err != nil {
    c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
  }
  artifactsTarGz, err := CollateArtifacts(uuid, db)
  // modify this!
  // only return 404 when that's the error.
  // when something else goes wrong, we need 500.
  if err != nil {
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
  } else {
    c.Data(http.StatusOK, "application/x-tar", artifactsTarGz)
  }
}

func main() { // coverage-ignore
  ParseArgs()
  gutsCfg, err := ParseConfig(configFilePath)
  CheckError(err)
  db = PostgresConnect(gutsCfg)
  defer DeferredErrCheck(db.Close)
  router := gin.Default()
  router.GET("/job/:uuid", JobEndpoint)
  router.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
  router_address := fmt.Sprintf("%v:%v", gutsCfg.Api.Hostname, gutsCfg.Api.Port)
  err = router.Run(router_address)
  CheckError(err)
}

