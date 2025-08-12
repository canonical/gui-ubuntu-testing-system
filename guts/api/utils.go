package main

import (
  "strings"
  "reflect"
  "regexp"
)

func CheckStringIsAlphanumeric(testString string) bool {
  return regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(testString)
}

func ValidateUuid(Uuid string) error {
  parts := strings.Split(Uuid, `-`)
  lengths := []int{8, 4, 4, 4, 12}

  if !reflect.DeepEqual(len(lengths), len(parts)) {
    return InvalidUuidError{uuid: Uuid}
  }

  for idx, length := range(lengths) {
    thisSegment := parts[idx]
    if !reflect.DeepEqual(len(thisSegment), length) {
      return InvalidUuidError{uuid: Uuid}
    }
    if !CheckStringIsAlphanumeric(thisSegment) {
      return InvalidUuidError{uuid: Uuid}
    }
  }
  return nil
}

