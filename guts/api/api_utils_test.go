package api

import (
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"testing"
)

func TestNewDbDriver(t *testing.T) {
	GutsCfg, _, _, err := Setup()
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	_, err = database.NewDbDriver(GutsCfg.Database.Driver, GutsCfg.Database.ConnectionString)
	if err != nil {
		t.Errorf("Unexpected error initialising db driver: %v", err.Error())
	}
}

func TestNewDbDriverBadDriver(t *testing.T) {
	GutsCfg, _, _, err := Setup()
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	dummyCfg := GutsCfg
	dummyCfg.Database.Driver = "not-a-db"
	_, err = database.NewDbDriver(dummyCfg.Database.Driver, dummyCfg.Database.ConnectionString)
	expectedErrString := "database couldn't be initialised - not-a-db is an unsupported driver"
	if err.Error() != expectedErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestGetStatusUrlForUuid(t *testing.T) {
	thisUuid := "2366a0e6-ba55-48bc-8fd6-6ed92a4c0e1a"
	configPath := "../guts-api.yaml"
	url := GetStatusUrlForUuid(thisUuid, configPath)
	expectedUrl := "http://localhost/status/2366a0e6-ba55-48bc-8fd6-6ed92a4c0e1a"
	if url != expectedUrl {
		t.Errorf("Unexpected url!\nExpected: %v\nActual: %v", expectedUrl, url)
	}
}
