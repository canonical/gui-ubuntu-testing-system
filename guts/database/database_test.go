package database

import (
	"testing"
)

func TestDbConnect(t *testing.T) {
	Driver, err := TestDbDriver()
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
	Driver, err := TestDbDriver()
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

func TestNewOperationInterfaceSuccess(t *testing.T) {
	Driver, err := TestDbDriver()
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
	Driver, err := TestDbDriver()
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
	Driver, err := TestDbDriver()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	err = Driver.Interface.DbAvailable()
	if err != nil {
		t.Errorf("Unexpected error creating db interface: %v", err.Error())
	}
}

func TestPostgresServiceNotUpError(t *testing.T) {
	var pgError PostgresServiceNotUpError
	desiredErrString := "Unit postgresql.service is not active."
	if pgError.Error() != desiredErrString {
		t.Errorf("PostgresServiceNotUpError giving unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, pgError.Error())
	}
}
