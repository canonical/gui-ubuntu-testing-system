package main

import (
  "strings"
  "reflect"
  "regexp"
  "os"
)

func ValidateUuid(Uuid string) error {
  return regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).MatchString(Uuid)
}

func FileOrDirExists(path string) error {
  if _, err := os.Stat(path); err != nil {
    if os.IsNotExist(err) {
      return false
    } else {
      CheckError(err)
    }
  }
  return true
}

func AllFilesExist(paths ...string) error {
  for i := 0; i < len(paths); i++ {
    err := FileOrDirExists(paths[i])
    if err != nil {
      return err
    }
  }
  return nil
}

func RemoveFiles(paths ...string) error {
  for i := 0; i < len(paths); i++ {
    if err := FileOrDirExists(paths[i]); err != nil {
      os.
    }

  }
}

func AtomicWrite(data []byte, filename string) error {
  newFile := fmt.Sprintf("%v.new", filename)
  err := os.WriteFile(newFile)
  if err != nil {
    return err
  }
  err = os.Rename(newFile, filename)
  if err != nil {
    return err
  }
  return nil
}

