package main

import (
  "strings"
  "reflect"
  "regexp"
)

func ValidateUuid(Uuid string) error {
  return regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).MatchString(Uuid)
}

