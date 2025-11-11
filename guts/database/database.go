package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"guts.ubuntu.com/v2/utils"
	"log"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"time"
)

type PostgresServiceNotUpError struct{}

func (e PostgresServiceNotUpError) Error() string {
	return "Unit postgresql.service is not active."
}

// //////////////////////////////////////////////////////////////////////////////
// Helper functions outside of the driver
func UpdateUpdatedAt(id int, Driver DbDriver) error {
	err := Driver.TestsUpdateUpdatedAt(id)
	return err
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
	// rename interface to client
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

func (d DbDriver) RunQueryRow(query string) (*sql.Row, error) {
	row, err := d.Interface.InterfaceRunQueryRow(query)
	return row, err
}

func (d DbDriver) UpdateRow(query string) error {
	err := d.Interface.InterfaceRunRowUpdate(query)
	return err
}

func (d DbDriver) TestsUpdateUpdatedAt(id int) error {
	err := d.Interface.UpdateUpdatedAt(id)
	return err
}

func (d DbDriver) SetTestStateTo(id int, state string) error {
	stateUpdateQuery := fmt.Sprintf(`UPDATE tests SET state='%v' WHERE id=%v`, state, id)
	err := d.UpdateRow(stateUpdateQuery)
	return err
}

func (d DbDriver) NukeUuid(uuid string) error {
	return d.Interface.RemoveUuidFromAllTables(uuid)
}

////////////////////////////////////////////////////////////////////////////////
// section with interfaces and functionality for different engines

type DbOperationInterface interface {
	DbAvailable() error
	InterfaceQueryRow(table, queryField, queryValue string, fields []string) (*sql.Row, error)
	InterfaceQuery(table, queryField, queryValue string, fields []string) (*sql.Rows, error)
	InterfacePrepareQuery(queryString string) (*sql.Stmt, error)
	InterfaceRunQueryRow(queryString string) (*sql.Row, error)
	InterfaceRunRowUpdate(query string) error
	UpdateUpdatedAt(id int) error
	RemoveUuidFromAllTables(uuid string) error
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
	defer utils.DeferredErrCheck(stmt.Close)
	row = stmt.QueryRow(queryValue)
	return row, nil
}

func (p PgOperationInterface) InterfaceQuery(table, queryField, queryValue string, fields []string) (*sql.Rows, error) { // coverage-ignore
  //////////
  // What happening?
	var rows *sql.Rows
	queryString := fmt.Sprintf("SELECT %v FROM %v WHERE %v='$1'", strings.Join(fields, ", "), table, queryField)
	log.Printf("running query %v with query parameter %v\n", queryString, queryValue)
	stmt, err := p.Db.Prepare(queryString)
	if err != nil { // coverage-ignore
		return rows, err
	}
	defer utils.DeferredErrCheck(stmt.Close)
  log.Printf("Statement prepared")
	rows, err = stmt.Query(queryValue)
	if err != nil { // coverage-ignore
		return rows, err
	}
	return rows, nil
}

func (p PgOperationInterface) InterfacePrepareQuery(queryString string) (*sql.Stmt, error) { // coverage-ignore
	stmt, err := p.Db.Prepare(queryString)
	return stmt, err
}

func (p PgOperationInterface) InterfaceRunQueryRow(queryString string) (*sql.Row, error) {
	var row *sql.Row
	stmt, err := p.Db.Prepare(queryString)
	if err != nil { // coverage-ignore
		return row, err
	}
	defer utils.DeferredErrCheck(stmt.Close)
	row = stmt.QueryRow()
	return row, nil
}

func (p PgOperationInterface) InterfaceRunRowUpdate(query string) error {
	stmt, err := p.Db.Prepare(query)
	if err != nil { // coverage-ignore
		return err
	}
	defer utils.DeferredErrCheck(stmt.Close)
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
	defer utils.DeferredErrCheck(stmt.Close)
	_, err = stmt.Exec(ts, id)
	return err
}

func (p PgOperationInterface) DeleteUuidFromTable(uuid, table string) error {
	removeQuery := fmt.Sprintf(`DELETE FROM %v WHERE uuid='%v'`, table, uuid)
	stmt, err := p.Db.Prepare(removeQuery)
	if err != nil { // coverage-ignore
		return err
	}
	defer utils.DeferredErrCheck(stmt.Close)
	_, err = stmt.Exec()
	return err
}

func (p PgOperationInterface) RemoveUuidFromAllTables(uuid string) error {
	// order must be preserved as uuid is a primary key in jobs
	err := p.DeleteUuidFromTable(uuid, "reporter")
	if err != nil { // coverage-ignore
		return err
	}

	err = p.DeleteUuidFromTable(uuid, "tests")
	if err != nil { // coverage-ignore
		return err
	}

	err = p.DeleteUuidFromTable(uuid, "jobs")
	if err != nil { // coverage-ignore
		return err
	}

	return nil
}

/////////////////////////////////////////////////////////////////////////////////
// Helper functions

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

func TestDbDriver(username, password string) (DbDriver, error) { // coverage-ignore
	var driver DbDriver
	driver.Driver = "postgres"
	driver.ConnectionString = fmt.Sprintf("host=localhost port=5432 user=%v password=%v dbname=guts sslmode=disable", username, password)
	driver.SupportedDrivers = []string{"postgres"}
	driver, err := NewOperationInterface(driver)
	return driver, err
}

func NewDbDriver(desiredDriver, desiredConnectionString string) (DbDriver, error) {
	var driver DbDriver
	driver.Driver = desiredDriver
	driver.ConnectionString = desiredConnectionString
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
