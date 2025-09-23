package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os/exec"
	"reflect"
	"strings"
)

func CheckError(err error) { // coverage-ignore
	if err != nil {
		log.Fatal(err.Error())
	}
}

func DeferredErrCheck(f func() error) { // coverage-ignore
	err := f()
	CheckError(err)
}

func DeferredErrCheckStringArg(f func(s string) error, s string) { // coverage-ignore
	err := f(s)
	CheckError(err)
}

type PostgresServiceNotUpError struct{}

func (e PostgresServiceNotUpError) Error() string {
	return "Unit postgresql.service is not active."
}

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

func (d DbDriver) QueryRow(table, queryField, queryValue string, fields []string) (*sql.Row, error) { // coverage-ignore
	row, err := d.Interface.InterfaceQueryRow(table, queryField, queryValue, fields)
	return row, err
}

func (d DbDriver) Query(table, queryField, queryValue string, fields []string) (*sql.Rows, error) { // coverage-ignore
	rows, err := d.Interface.InterfaceQuery(table, queryField, queryValue, fields)
	return rows, err
}

func (d DbDriver) PrepareQuery(query string) (*sql.Stmt, error) { // coverage-ignore
	stmt, err := d.Interface.InterfacePrepareQuery(query)
	return stmt, err
}

////////////////////////////////////////////////////////////////////////////////
// section with interfaces and functionality for different engines

type DbOperationInterface interface {
	DbAvailable() error
	InterfaceQueryRow(table, queryField, queryValue string, fields []string) (*sql.Row, error)
	InterfaceQuery(table, queryField, queryValue string, fields []string) (*sql.Rows, error)
	InterfacePrepareQuery(queryString string) (*sql.Stmt, error)
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

func (p PgOperationInterface) InterfaceQueryRow(table, queryField, queryValue string, fields []string) (*sql.Row, error) { // coverage-ignore
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

func (p PgOperationInterface) InterfaceQuery(table, queryField, queryValue string, fields []string) (*sql.Rows, error) { // coverage-ignore
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

func (p PgOperationInterface) InterfacePrepareQuery(queryString string) (*sql.Stmt, error) { // coverage-ignore
	stmt, err := p.Db.Prepare(queryString)
	return stmt, err
}

// just used for testing, so we don't test
func SkipTestIfPostgresInactive(PgError error) bool { // coverage-ignore
	var expectedType PostgresServiceNotUpError
	if PgError != nil {
		if reflect.DeepEqual(reflect.TypeOf(PgError), reflect.TypeOf(expectedType)) {
			return true
		}
	}
	return false
}

func TestDbDriver() (DbDriver, error) {
	var driver DbDriver
	driver.Driver = "postgres"
	driver.ConnectionString = "host=localhost port=5432 user=guts_api password=guts_api dbname=guts sslmode=disable"
	driver.SupportedDrivers = []string{"postgres"}
	driver, err := NewOperationInterface(driver)
	return driver, err
}
