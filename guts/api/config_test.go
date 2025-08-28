package main

import (
  "testing"
  "reflect"
  "strings"
  "os"
  "io/fs"
  "gopkg.in/yaml.v3"
)


func TestParseConfigSuccess(t *testing.T) {
  err := ParseConfig("./guts-api.yaml")
  CheckError(err)
  var wanted GutsApiConfig
  wanted.Database.ConnectionString = "host=localhost port=5432 user=guts_api password=guts_api dbname=guts sslmode=disable"
  wanted.Api.Hostname = "localhost"
  wanted.Api.Port = 8080
  if !reflect.DeepEqual(GutsCfg, wanted) {
    t.Errorf("Parsed config not the same as wanted config!\nExpected:\n%v\nActual:\n%v", cfg, wanted)
  }
}

func TestParseConfigFileNotFound(t *testing.T) {
  err := ParseConfig("./guts-api-no-exist.yaml")
  var ExpectedType *fs.PathError
  if reflect.TypeOf(err) != reflect.TypeOf(ExpectedType) {
    t.Errorf("Error type not as expected!\nExpected: %v\nActual: %v", ExpectedType, reflect.TypeOf(err))
  }
  expected_string :=  "guts-api-no-exist.yaml: no such file or directory"
  if !strings.Contains(err.Error(), expected_string) {
    t.Errorf("Error string doesn't contain expected substring!\nExpected:\n%v\nActual:\n%v", expected_string, err.Error())
  }
}

func TestParseConfigYamlParsingFailure(t *testing.T) {
  f, err := os.CreateTemp(".", "not-a-yaml-file")
  CheckError(err)
  defer DeferredErrCheck(f.Close)
  defer DeferredErrCheckStringArg(os.Remove, f.Name())

  data := []byte("This is not a yaml file.")
  _, err = f.Write(data)
  CheckError(err)

  err = ParseConfig(f.Name())

  var ExpectedType *yaml.TypeError
  if reflect.TypeOf(err) != reflect.TypeOf(ExpectedType) {
    t.Errorf("Error type not as expected!\nExpected: %v\nActual: %v", reflect.TypeOf(ExpectedType), reflect.TypeOf(err))
  }
  expected_string := "cannot unmarshal !!str `This is...` into main.GutsApiConfig"
  if !strings.Contains(err.Error(), expected_string) {
    t.Errorf("Error string doesn't contain expected substring!\nExpected:\n%v\nActual:\n%v", expected_string, err.Error())
  }
}
