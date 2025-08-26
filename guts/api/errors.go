package main

import (
  "fmt"
)


type UuidNotFoundError struct {
    uuid string
}

func (e UuidNotFoundError) Error() string {
    return fmt.Sprintf("No jobs with uuid %v found!", e.uuid)
}

type InvalidUuidError struct {
  uuid string
}

func (e InvalidUuidError) Error() string {
    return fmt.Sprintf("%v isn't a valid uuid!", e.uuid)
}

type PostgresServiceNotUpError struct {}

func (e PostgresServiceNotUpError) Error() string {
  return fmt.Sprintf("Unit postgresql.service is not active.")
}

