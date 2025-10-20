package database

import (
	"guts.ubuntu.com/v2/utils"
	"testing"
)

func TestDbConnect(t *testing.T) {
	Driver, err := TestDbDriver("guts_api", "guts_api")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	_, err = Driver.DbConnect()
	if err != nil {
		t.Errorf("Connecting to db shouldn't have failed!\nError: %v\nConfig: %v", err.Error(), Driver)
	}
}

func TestDbConnectBadDriver(t *testing.T) {
	Driver, err := TestDbDriver("guts_api", "guts_api")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
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
	Driver, err := TestDbDriver("guts_api", "guts_api")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	_, err = NewOperationInterface(Driver)
	if err != nil {
		t.Errorf("Unexpected error creating db interface: %v", err.Error())
	}
}

func TestNewOperationInterfaceFails(t *testing.T) {
	Driver, err := TestDbDriver("guts_api", "guts_api")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
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
	Driver, err := TestDbDriver("guts_api", "guts_api")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
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
		t.Errorf("utils.PostgresServiceNotUpError giving unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, pgError.Error())
	}
}

func TestRunQueryRow(t *testing.T) {
	Driver, err := TestDbDriver("guts_api", "guts_api")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	_, err = Driver.RunQueryRow("SELECT * FROM jobs WHERE uuid='4ce9189f-561a-4886-aeef-1836f28b073b';")
	utils.CheckError(err)
}

func TestUpdateRow(t *testing.T) {
	Driver, err := TestDbDriver("guts_spawner", "guts_spawner")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	err = Driver.UpdateRow(`UPDATE tests SET state='spawning' WHERE id=4`)
	utils.CheckError(err)
	err = Driver.UpdateRow(`UPDATE tests SET state='requested' WHERE id=4`)
	utils.CheckError(err)
}

func TestTestsUpdateUpdatedAt(t *testing.T) {
	Driver, err := TestDbDriver("guts_spawner", "guts_spawner")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	err = Driver.TestsUpdateUpdatedAt(4)
	utils.CheckError(err)
}

func TestNewDbDriver(t *testing.T) {
	_, err := NewDbDriver("postgres", "host=localhost port=5432 user=guts_api password=guts_api dbname=guts sslmode=disable")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		if err != nil {
			t.Errorf("Unexpected error initialising db driver: %v", err.Error())
		}
	}
}

func TestNewDbDriverBadDriver(t *testing.T) {
	_, err := NewDbDriver("not-a-db", "")
	expectedErrString := "database couldn't be initialised - not-a-db is an unsupported driver"
	if err.Error() != expectedErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestSetTestStateTo(t *testing.T) {
	Driver, err := TestDbDriver("guts_spawner", "guts_spawner")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	rowId := 13
	err = Driver.SetTestStateTo(rowId, "requested")
	utils.CheckError(err)
	err = Driver.SetTestStateTo(rowId, "running")
	utils.CheckError(err)
	err = Driver.SetTestStateTo(rowId, "requested")
	utils.CheckError(err)
}

func TestUpdateUpdatedAt(t *testing.T) {
	Driver, err := TestDbDriver("guts_spawner", "guts_spawner")
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	rowId := 1
	err = UpdateUpdatedAt(rowId, Driver)
	utils.CheckError(err)
}
