package main

import (
  "testing"
  "reflect"
)

func TestUuidNotFoundError(t *testing.T) {
  var UuidError UuidNotFoundError
  UuidError.uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
  ExpectedString := "No jobs with uuid 4ce9189f-561a-4886-aeef-1836f28b073b found!"
  if !reflect.DeepEqual(UuidError.Error(), ExpectedString) {
    t.Errorf("Uuid failure string not as expected!\nExpected: %v\nActual: %v", ExpectedString, UuidError.Error())
  }
}

