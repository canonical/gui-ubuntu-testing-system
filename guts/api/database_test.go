package main

import (
	"testing"
)

func TestDbConnect(t *testing.T) {
	err := Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	_, err = Driver.DbConnect()
	if err != nil {
		t.Errorf("Connecting to db shouldn't have failed!\nError: %v\nConfig: %v", err.Error(), Driver)
	}
}

func TestDbConnectBadDriver(t *testing.T) {
	err := Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	savedDriver := Driver.Driver
	Driver.Driver = "not-a-db"
	_, err = Driver.DbConnect()
	Driver.Driver = savedDriver
	expectedErrString := `sql: unknown driver "not-a-db" (forgotten import?)`
	if err.Error() != expectedErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestNewDbDriver(t *testing.T) {
	err := Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	dummyCfg := GutsCfg
	_, err = NewDbDriver(dummyCfg)
	if err != nil {
		t.Errorf("Unexpected error initialising db driver: %v", err.Error())
	}
}

func TestNewDbDriverBadDriver(t *testing.T) {
	err := Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	dummyCfg := GutsCfg
	dummyCfg.Database.Driver = "not-a-db"
	_, err = NewDbDriver(dummyCfg)
	expectedErrString := "database couldn't be initialised - not-a-db is an unsupported driver"
	if err.Error() != expectedErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestNewOperationInterfaceSuccess(t *testing.T) {
	err := Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	_, err = NewOperationInterface(Driver)
	if err != nil {
		t.Errorf("Unexpected error creating db interface: %v", err.Error())
	}
}

func TestNewOperationInterfaceFails(t *testing.T) {
	err := Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	dummyDriver := Driver
	dummyDriver.Driver = "not-a-db"
	_, err = NewOperationInterface(dummyDriver)
	expectedErrString := "not-a-db is an invalid driver"
	if expectedErrString != err.Error() {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestDbAvailableSuccess(t *testing.T) {
	err := Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	Driver, err := NewOperationInterface(Driver)
	CheckError(err)
	err = Driver.Interface.DbAvailable()
	if err != nil {
		t.Errorf("Unexpected error creating db interface: %v", err.Error())
	}
}
