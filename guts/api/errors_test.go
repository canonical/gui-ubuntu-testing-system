package main

import (
	"reflect"
	"testing"
)

func TestUuidNotFoundError(t *testing.T) {
	var UuidError UuidNotFoundError
	UuidError.uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
	ExpectedString := "No jobs with uuid 4ce9189f-561a-4886-aeef-1836f28b073b found!"
	if !reflect.DeepEqual(UuidError.Error(), ExpectedString) {
		t.Errorf("Uuid failure string not as expected!\nExpected: %v\nActual: %v", ExpectedString, UuidError.Error())
	}
}

func SkipTestIfPostgresInactive(PgError error) bool {
	var expectedType PostgresServiceNotUpError
	if PgError != nil {
		if reflect.DeepEqual(reflect.TypeOf(PgError), reflect.TypeOf(expectedType)) {
			return true
		}
	}
	return false
}

func TestPostgresServiceNotUpError(t *testing.T) {
	var pgError PostgresServiceNotUpError
	desiredErrString := "Unit postgresql.service is not active."
	if pgError.Error() != desiredErrString {
		t.Errorf("PostgresServiceNotUpError giving unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, pgError.Error())
	}
}
