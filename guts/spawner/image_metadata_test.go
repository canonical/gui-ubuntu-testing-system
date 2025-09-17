package main

import (
  "testing"
)

func TestGetRemoteShaSum(t *testing.T) {
  ServeDirectory()
  imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
  shasum, err := GetRemoteShaSum(imageUrl)
  CheckError(err)
  expectedShasum := "a52d5d22d71375efae79de6cf8a125228ac19356c84f4af17ef5955147be7ef5"
  if shasum != expectedShasum {
    t.Errorf("Parsed shasum not same as expected!\nExpected: %v\nActual: %v", expectedShasum, shasum)
  }
}

func TestCdImageGetShasumOfImage(t *testing.T) {
  ServeDirectory()
  imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
  shasum, err := CdImageGetShasumOfImage(imageUrl)
  CheckError(err)
  expectedShasum := "a52d5d22d71375efae79de6cf8a125228ac19356c84f4af17ef5955147be7ef5"
  if shasum != expectedShasum {
    t.Errorf("Parsed shasum not same as expected!\nExpected: %v\nActual: %v", expectedShasum, shasum)
  }
}

func TestCdImageDownloadCheckSumFileForImage(t *testing.T) {
  ServeDirectory()
  imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
  shasums, err :=  CdImageDownloadCheckSumFileForImage(imageUrl)
  expectedShasumFile := "a52d5d22d71375efae79de6cf8a125228ac19356c84f4af17ef5955147be7ef5 *questing-mini-iso-amd64.iso\n"
  if expectedShasumFile != shasums {
    t.Errorf("Parsed shasum file not the same as expected!\nExpected: %v\nActual: %v", expectedShasumFile, shasums)
  }
}

func TestCdImageParseShasumForImage(t *testing.T) {
  shasumFile := "a4310d26648801af766733bf01845d7d2f4d26a96a02ea45ee532d74068999a0 *questing-desktop-amd64.iso\nf425a4872fdd163f38ec785eaa1bdab1f1bdae202b67247f57ed1aa96a3a20a4 *questing-desktop-arm64.iso\n"
  desiredImage := "questing-desktop-amd64.iso"
  parsedShasum, err := CdImageParseShasumForImage(shasumFile, desiredImage)
  CheckError(err)
  expectedShasum := "a4310d26648801af766733bf01845d7d2f4d26a96a02ea45ee532d74068999a0"
  if parsedShasum != expectedShasum {
    t.Errorf("Parsed shasum not the same as expected!\nExpected: %v\nActual: %v", expectedShasum, parsedShasum)
  }
}

