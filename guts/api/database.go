package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os/exec"
	"slices"
	"strings"
)

var (
	Driver DbDriver
)

////////////////////////////////////////////////////////////////////////////////
// basic functionality for the driver struct

type DbDriver struct {
	Driver           string
	ConnectionString string
	SupportedDrivers []string
	Interface        DbOperationInterface
}

func (d DbDriver) DbConnect() (*sql.DB, error) {
	var db *sql.DB
	db, err := sql.Open(d.Driver, d.ConnectionString)
	if err != nil {
		return db, err
	}
	return db, nil
}

func NewDbDriver(g GutsApiConfig) (DbDriver, error) {
	var driver DbDriver
	driver.Driver = g.Database.Driver
	driver.ConnectionString = g.Database.ConnectionString
	driver.SupportedDrivers = []string{"postgres"}
	if !slices.Contains(driver.SupportedDrivers, driver.Driver) {
		return driver, fmt.Errorf("database couldn't be initialised - %v is an unsupported driver", driver.Driver)
	}
	driver, err := NewOperationInterface(driver)
	if err != nil { // coverage-ignore
		return driver, err
	}
	return driver, nil
}

func NewOperationInterface(driver DbDriver) (DbDriver, error) {
	if driver.Driver == "postgres" {
		var thisInterface PgOperationInterface
		thisInterface.Driver = driver
		db, err := driver.DbConnect()
		if err != nil { // coverage-ignore
			return driver, err
		}
		thisInterface.Db = db
		driver.Interface = thisInterface
		return driver, nil
	}
	return driver, fmt.Errorf("%v is an invalid driver", driver.Driver)
}

////////////////////////////////////////////////////////////////////////////////
// section utilising the interfaces below in the driver struct

func (d DbDriver) QueryRow(table, queryField, queryValue string, fields []string) (*sql.Row, error) {
	row, err := d.Interface.InterfaceQueryRow(table, queryField, queryValue, fields)
	return row, err
}

func (d DbDriver) Query(table, queryField, queryValue string, fields []string) (*sql.Rows, error) {
	rows, err := d.Interface.InterfaceQuery(table, queryField, queryValue, fields)
	return rows, err
}

func (d DbDriver) InsertJobsRow(job JobEntry) error {
	err := d.Interface.InterfaceInsertJobsRow(job)
	return err
}

////////////////////////////////////////////////////////////////////////////////
// section with interfaces and functionality for different engines

type DbOperationInterface interface {
	DbAvailable() error
	InterfaceQueryRow(table, queryField, queryValue string, fields []string) (*sql.Row, error)
	InterfaceQuery(table, queryField, queryValue string, fields []string) (*sql.Rows, error)
	InterfaceInsertJobsRow(job JobEntry) error
}

type PgOperationInterface struct {
	Driver DbDriver
	Db     *sql.DB
}

func (p PgOperationInterface) DbAvailable() error {
	systemctlCommand := exec.Command("systemctl", "status", "postgresql.service")
	if err := systemctlCommand.Run(); err != nil { // coverage-ignore
		return PostgresServiceNotUpError{}
	}
	_, err := p.Driver.DbConnect()
	if err != nil { // coverage-ignore
		return err
	}
	return nil
}

func (p PgOperationInterface) InterfaceQueryRow(table, queryField, queryValue string, fields []string) (*sql.Row, error) {
	var row *sql.Row
	queryString := fmt.Sprintf("SELECT %v FROM %v WHERE %v=$1", strings.Join(fields, ", "), table, queryField)
	stmt, err := p.Db.Prepare(queryString)
	if err != nil { // coverage-ignore
		return row, err
	}
	defer DeferredErrCheck(stmt.Close)
	row = stmt.QueryRow(queryValue)
	return row, nil
}

func (p PgOperationInterface) InterfaceQuery(table, queryField, queryValue string, fields []string) (*sql.Rows, error) {
	var rows *sql.Rows
	queryString := fmt.Sprintf("SELECT %v FROM %v WHERE %v=$1", strings.Join(fields, ", "), table, queryField)
	stmt, err := p.Db.Prepare(queryString)
	if err != nil { // coverage-ignore
		return rows, err
	}
	defer DeferredErrCheck(stmt.Close)
	rows, err = stmt.Query(queryValue)
	if err != nil { // coverage-ignore
		return rows, err
	}
	return rows, err
}

func (p PgOperationInterface) InterfaceInsertJobsRow(job JobEntry) error {
	queryString := fmt.Sprintf(
		`INSERT INTO jobs (%v) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		strings.Join(AllJobColumns, ", "),
	)
	stmt, err := p.Db.Prepare(queryString)
	if err != nil { // coverage-ignore
		return err
	}
	defer DeferredErrCheck(stmt.Close)
	_, err = stmt.Exec(
		job.Uuid,
		job.ArtifactUrl,
		job.TestsRepo,
		job.TestsRepoBranch,
		fmt.Sprintf(`{%v}`, strings.Join(job.TestsPlans, ",")),
		job.ImageUrl,
		job.Reporter,
		job.Status,
		job.SubmittedAt,
		job.Requester,
		job.Debug,
		job.Priority,
	)
	return err
}
