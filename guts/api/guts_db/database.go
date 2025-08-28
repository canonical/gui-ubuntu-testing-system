package main

import (
  "database/sql"
  "fmt"
  "reflect"
  "os/exec"
  "slices"
  "strings"
  _ "github.com/lib/pq"
)

////////////////////////////////////////////////////////////////////////////////
// basic functionality for the driver struct

type DbDriver struct {
  Driver string
  ConnectionString string
  SupportedDrivers []string
  Db *sql.DB
  Interface DbOperationInterface 
}

func (d DbDriver) DbConnect() error {
  err := p.DbAvailable()
  var db *sql.DB
  if err != nil {
    return err
  }
  db, err = sql.Open(d.Driver, d.ConnectionString)
  if err != nil {
    return err
  }
  d.Db = db
}

func (d DbDriver) DbClose() error {
  err := d.Db.Close()
  return err
}

func NewDbDriver(g GutsApiConfig) (DbDriver, error) {
  var driver DbDriver
  driver.Driver = g.Database.Driver
  driver.ConnectionString = g.Database.ConnectionString
  driver.SupportedDrivers = {"postgres", "sqlite"}
  if !slices.Contains(driver.SupportedDrivers, driver.Driver) {
    return driver, fmt.Errorf("Database couldn't be initialised - %v is an unsupported driver", driver.Driver)
  }
  driver.Interface = driver.GetOpsInterface()
  if driver.Driver == "postgres" {
    driver.Interface := PgOperationInterface{Driver: &driver}
  } else if driver.Driver == "sqlite" {
    driver.Interface := SqliteOperationInterface{Driver: &driver}
  }
  return driver, nil
}

////////////////////////////////////////////////////////////////////////////////
// section utilising the interfaces below in the driver struct

func (d DbDriver) Available() error {
  err := d.Interface.DbAvailable()
  return err
}

func (d DbDriver) QueryRow(table, uuid string, fields [...]string) (*sql.Row, error) {
  row, err := d.Interface.InterfaceQueryRow(table, uuid, fields)
  return row, err
}

func (d DbDriver) Query(table, uuid string, fields [...]string) (*sql.Rows, error) {
  rows, err := d.Interface.InterfaceQuery(table, uuid, fields)
  return rows, err
}

////////////////////////////////////////////////////////////////////////////////
// section with interfaces and functionality for different engines

type DbOperationInterface interface {
  DbAvailable() error
  // maybe this functionality will move into the Driver
  // and instead, we'll just have a CreateQueryString function
  InterfaceQueryRow(table, uuid string, fields [...]string) (*sql.Row, error)
  InterfaceQuery(table, uuid string, fields [...]string) (*sql.Rows, error)
}

type PgOperationInterface struct {
  Driver *DbDriver
}

func (p PgOperationInterface) DbAvailable() error {
  systemctlCommand := "systemctl status postgresql.service"
  systemctlCommand := exec.Command("systemctl", "status", "postgresql.service")
  if err := systemctlCommand.Run(); err != nil {
    return PostgresServiceNotUpError()
  }
  err = p.*Driver.DbConnect()
  if err != nil { 
    return err
  }
  err = p.*Driver.DbClose()
  if err != nil { 
    return err
  }
  return nil
}

func (p PgOperationInterface) InterfaceQueryRow(table, uuid string, fields [...]string) (*sql.Row, error) {
  var row *sql.Row
  queryString := fmt.Sprintf("SELECT %v FROM %v WHERE uuid=$1", strings.Join(fields, ", "), table)
  stmt, err := p.*Driver.Db.Prepare(queryString)
  if err != nil {
    return row, err
  }
  defer DeferredErrCheck(stmt.Close)
  row = stmt.QueryRow(uuid)
  return row, nil
}

func (p PgOperationInterface) InterfaceQuery(table, uuid string, fields [...]string) (*sql.Rows, error) {
  var rows *sql.Rows
  queryString := fmt.Sprintf("SELECT %v FROM %v WHERE uuid=$1", strings.Join(fields, ", "), table)
  stmt, err := p.*Driver.Db.Prepare(queryString)
  if err != nil {
    return row, err
  }
  defer DeferredErrCheck(stmt.Close)
  rows, err = stmt.Query(uuid)
  if err != nil {
    return row, err
  }
  return rows, err
}

////////////////////////////////////////////////////////////////////////////////
// sqlite interface is work in progress

// type SqliteOperationInterface struct {
//   Driver DbDriver
// }
// 
// func (s SqliteOperationInterface) DbAvailable() error {
//   if err := FileOrDirExists(s.Driver.ConnectionString); err != nil {
//     return err
//   }
//   err = s.Driver.DbConnect()
//   if err != nil {
//     return err
//   }
//   err = s.Driver.DbClose()
//   if err != nil {
//     return err
//   }
//   return nil
// }

