package main

import (
  "testing"
)

func TestPostgresServiceNotUpError(t *testing.T) {
	var pgError PostgresServiceNotUpError
	desiredErrString := "Unit postgresql.service is not active."
	if pgError.Error() != desiredErrString {
		t.Errorf("PostgresServiceNotUpError giving unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, pgError.Error())
	}
}
