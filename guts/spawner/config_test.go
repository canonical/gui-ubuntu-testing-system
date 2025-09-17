package main

import (
  "reflect"
  "testing"
)

func TestParseConfigSuccess(t *testing.T) {
  err := ParseConfig("./guts-spawner.yaml")
  CheckError(err)
  var testCfg GutsSpawnerConfig
  testCfg.Database.Driver = "postgres"
  testCfg.Database.ConnectionString = "host=localhost port=5432 user=guts_spawner password=guts_spawner dbname=guts sslmode=disable"
  testCfg.Virtualisation.Memory = 4096
  testCfg.Virtualisation.Cores = 8
  testCfg.General.ImageCachePath = "/srv/guts/images"
  if !reflect.DeepEqual(SpawnerCfg, testCfg) {
    t.Errorf("parsed config not the same as expected!\nExpected: %v\nActual: %v", testCfg, SpawnerCfg)
  }
}

