package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os/exec"
	"slices"
  "time"
)

var (
	Driver DbDriver
)

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

func NewDbDriver(g GutsSpawnerConfig) (DbDriver, error) {
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

func (d DbDriver) RunQueryRow(query string) (*sql.Row, error) {
	row, err := d.Interface.InterfaceRunQueryRow(query)
	return row, err
}

func (d DbDriver) UpdateRow(query string) error {
	err := d.Interface.InterfaceRunRowUpdate(query)
	return err
}

func (d DbDriver) TestsUpdateUpdatedAt(id int) error {
  err := d.Interface.UpdateUpdatedAt(id int)
}

type DbOperationInterface interface {
	DbAvailable() error
	InterfaceRunQueryRow(query string) (*sql.Row, error)
	InterfaceRunRowUpdate(query string) error
  UpdateUpdatedAt(id int) error
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

func (p PgOperationInterface) InterfaceRunQueryRow(query string) (*sql.Row, error) {
	var row *sql.Row
	stmt, err := p.Db.Prepare(query)
	if err != nil { // coverage-ignore
		return row, err
	}
	defer DeferredErrCheck(stmt.Close)
	row = stmt.QueryRow()
	return row, nil
}

func (p PgOperationInterface) InterfaceRunRowUpdate(query string) error {
	fmt.Println("Running:")
	fmt.Println(query)
	stmt, err := p.Db.Prepare(query)
	if err != nil { // coverage-ignore
		return err
	}
	defer DeferredErrCheck(stmt.Close)
	_, err = stmt.Query()
	return err
}

func (p PgOperationInterface) UpdateUpdatedAt(id int) error {
  ts := time.Now()
  updateCmd := `UPDATE tests SET updated_at=$1 WHERE id=$2`
	stmt, err := p.Db.Prepare(updateCmd)
	if err != nil { // coverage-ignore
		return err
	}
	defer DeferredErrCheck(stmt.Close)
  _, err = stmt.Exec(ts, id)
  return err
}
