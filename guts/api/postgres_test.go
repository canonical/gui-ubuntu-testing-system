package main

import (
  "testing"
  "reflect"
)

func TestPostgresConnect(t *testing.T) {
  err := ParseConfig(configFilePath)
  db, err = PostgresConnect()
  if SkipTestIfPostgresInactive(err) {
    t.Skip("Skipping test as postgresql service is not up")
  } else {
    CheckError(err)
  }
  if err != nil {
    t.Errorf("Postgres connection failed with creds:\n%v", GutsCfg.Postgres)
  }
}

func TestPostgresConnectString(t *testing.T) {
  err := ParseConfig("./guts-api.yaml")
  CheckError(err)
  connect_string := GutsCfg.PostgresConnectString()
  expected_string := "host=localhost port=5432 user=guts_api password=guts_api dbname=guts sslmode=disable"
  if !reflect.DeepEqual(connect_string, expected_string) {
    t.Errorf("Postgres connection string not the same as expected!\nExpected:\n%v\nActual:\n%v", expected_string, connect_string)
  }
}
