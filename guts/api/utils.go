package main

import (
  "strings"
  "reflect"
  "regexp"
  "os"
  "crypto/sha256"
  "encoding/base64"
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

func Sha256sumOfString(inputString string) (string, error) {
  hasher := sha256.New()
  hasher.Write([]byte(inputString))
  return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func GetStatusUrlForUuid(uuid string) string {
  statusUrl := fmt.Sprintf("%v%v/status/%v", GetProtocolPrefix(), GutsCfg.Api.Hostname, uuid)
  return statusUrl
}

func GetProtocolPrefix() string {
  if GutsCfg.Api.Port == 8080 {
    return "http://"
  } else if GutsCfg.Api.Port == 443 {
    return "https://"
  } else {
    return ""
  }
}

