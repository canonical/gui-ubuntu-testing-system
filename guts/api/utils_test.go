package main

import (
  "testing"
  "reflect"
)

func TestValidateUuid(t *testing.T) {
  Uuid := "27549483-e8f5-497f-a05d-e6d8e67a8e8a"
  err := ValidateUuid(Uuid)
  reflect.DeepEqual(Uuid, "")
  CheckError(err)
}

func TestValidateUuidFails(t *testing.T) {
  Uuid := "farnsworth"
  err := ValidateUuid(Uuid)
  var expectedType InvalidUuidError
  if !reflect.DeepEqual(reflect.TypeOf(err), reflect.TypeOf(expectedType)) {
    t.Errorf("Unexpected error type!")
  }
}

func TestValidateUuidIncorrectSegmentLengthFailure(t *testing.T) {
  Uuid := "farnsworth-farnsworth-farnsworth-farnsworth-farnsworth"
  err := ValidateUuid(Uuid)
  var expectedType InvalidUuidError
  if !reflect.DeepEqual(reflect.TypeOf(err), reflect.TypeOf(expectedType)) {
    t.Errorf("Unexpected error type!")
  }
}

func TestValidateUuidNonAlphanumericFailure(t *testing.T) {
  Uuid := "!!!!!!!!-!!!!-!!!!-!!!!-!!!!!!!!!!!!"
  err := ValidateUuid(Uuid)
  var expectedType InvalidUuidError
  if !reflect.DeepEqual(reflect.TypeOf(err), reflect.TypeOf(expectedType)) {
    t.Errorf("Unexpected error type!")
  }
}

func TestCheckStringIsAlphanumeric(t *testing.T) {
  myTestString := "Farnsworth2841"
  if !CheckStringIsAlphanumeric(myTestString) {
    t.Errorf("%v misidentified as NOT being an alphanumeric string", myTestString)
  }
}

func TestCheckStringIsAlphanumericFails(t *testing.T) {
  myTestString := "Farnsworth2841!!!"
  if CheckStringIsAlphanumeric(myTestString) {
    t.Errorf("%v misidentified as being an alphanumeric string", myTestString)
  }
}

func InitTestPgSkippableReady() (bool, error) {
  ParseArgs()
  err := ParseConfig(configFilePath)
  if err != nil {
    CheckError(err)
  }
  Driver, err = NewDbDriver(GutsApiConfig)
  if SkipTestIfPostgresInactive(err) {
    return false, "Skipping test as postgresql service is not up"
  } else {
    CheckError(err)
  }
  return true, nil
}

