package main

import (
  "testing"
  "reflect"
  "slices"
  "os"
  "os/exec"
  "strings"
  "errors"
  "fmt"
)

func TestFindArtifactUrlsByUuid(t *testing.T) {
  Uuid := "eccd3988-490d-4414-be97-605d1ac81073"
  gutsCfg, err := ParseConfig("./guts-api.yaml")
  CheckError(err)
  db, err = PostgresConnect(gutsCfg)
  if SkipTestIfPostgresInactive(err) {
    t.Skip("Skipping test as postgresql service is not up")
  } else {
    CheckError(err)
  }
  result_urls, err := FindArtifactUrlsByUuid(Uuid, db)
  CheckError(err)
  expectedUrls := make([]string, 3)
  expectedUrls[0] = "https://guts.ubuntu.com/artifacts/eccd3988-490d-4414-be97-605d1ac81073/"
  expectedUrls[1] = "https://guts.ubuntu.com/artifacts/eccd3988-490d-4414-be97-605d1ac81073/"
  expectedUrls[2] = "https://guts.ubuntu.com/artifacts/eccd3988-490d-4414-be97-605d1ac81073/"

  if !reflect.DeepEqual(result_urls, expectedUrls) {
    t.Errorf("Results url as expected!\nExpected: %v\nActual: %v", expectedUrls, result_urls)
  }
}

func TestCollateArtifacts(t *testing.T) {
  ServeDirectory()

  // Get output artifacts for given uuid
  Uuid := "27549483-e8f5-497f-a05d-e6d8e67a8e8a"
  gutsCfg, err := ParseConfig("./guts-api.yaml")
  CheckError(err)
  db, err = PostgresConnect(gutsCfg)
  if SkipTestIfPostgresInactive(err) {
    t.Skip("Skipping test as postgresql service is not up")
  } else {
    CheckError(err)
  }
  artifactsTarGz, err := CollateArtifacts(Uuid, db)
  CheckError(err)

  // Write the tar to a file
  tempTarName := "/tmp/test-tar.tar.gz"
  err = os.WriteFile(tempTarName, artifactsTarGz, 0644)
  CheckError(err)
  // defer os.Remove(tempTarName)

  if _, err := os.Stat(tempTarName); errors.Is(err, os.ErrNotExist) {
    t.Errorf("Couldn't write tarfile: %v", tempTarName)
    // path/to/whatever does not exist
  }


  // Get the list of files included in the tar
  out, err := exec.Command("tar", "ztf", tempTarName).Output()
  CheckError(err)
  tarFiles := strings.Split(string(out), "\n")

  // Create the expected list of files
  expectedFiles := make([]string, 9)
  expectedFiles[0] = "res-1/output.xml"
  expectedFiles[1] = "res-1/log.html"
  expectedFiles[2] = "res-1/report.html"
  expectedFiles[3] = "res-2/output.xml"
  expectedFiles[4] = "res-2/log.html"
  expectedFiles[5] = "res-2/report.html"
  expectedFiles[6] = "res-3/output.xml"
  expectedFiles[7] = "res-3/log.html"
  expectedFiles[8] = "res-3/report.html"

  // the tar'ing up files doesn't guarantee the files will always be in the same order
  // in the tarball - I originally was going to check they're the same with shasums
  // but this ordering issue changes the shasum, making it inconsistent
  slices.Sort(expectedFiles)
  slices.Sort(tarFiles)

  // the first element of tarFiles is empty, quirk of exec.Command
  if !reflect.DeepEqual(expectedFiles, tarFiles[1:]) {
    t.Errorf("Expected tar files not the same as actual tarfiles:\nExpected: %v\nActual: %v", expectedFiles, tarFiles)
  }
}

func TestCollateArtifactsDownloadFails(t *testing.T) {
  Uuid := "44eea936-1e4a-4e20-b25d-ab0df9978ada"
  gutsCfg, err := ParseConfig("./guts-api.yaml")
  CheckError(err)
  db, err = PostgresConnect(gutsCfg)
  if SkipTestIfPostgresInactive(err) {
    t.Skip("Skipping test as postgresql service is not up")
  } else {
    CheckError(err)
  }
  artifactsTarGz, err := CollateArtifacts(Uuid, db)
  if len(artifactsTarGz) != 0 {
    t.Errorf("Variable should have length 0 but instead has length %v", len(artifactsTarGz))
  }
  expectedErrString := "gzip: invalid header"
  if !reflect.DeepEqual(err.Error(), expectedErrString) {
    t.Errorf("Unexpected error!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
  }
}

func TestCreateOutputDirectoriesFromUrls(t *testing.T) {
  artifactUrls := []string {
    "http://localhost:9999/res-1.tar.gz",
    "http://localhost:9999/res-2.tar.gz",
    "http://localhost:9999/res-3.tar.gz",
  }
  expectedDirNames := []string {
    "res-1",
    "res-2",
    "res-3",
  }
  directoryNames, err := CreateOutputDirectoriesFromUrls(artifactUrls)
  CheckError(err)

  if !reflect.DeepEqual(expectedDirNames, directoryNames) {
    t.Errorf("Expected directory names not the same as actual!\nExpected: %v\nActual: %v", expectedDirNames, directoryNames)
  }

}

func TestDownloadTarFiles(t *testing.T) {
  ServeDirectory()
  artifactUrls := []string {
    "http://localhost:9999/res-1.tar.gz",
    "http://localhost:9999/res-2.tar.gz",
    "http://localhost:9999/res-3.tar.gz",
  }
  _, err := DownloadTarFiles(artifactUrls)
  CheckError(err)
}

func TestCreateOutputDirectoriesFromUrlsEmptyUrls(t *testing.T) {
  var urls []string
  dirs, err := CreateOutputDirectoriesFromUrls(urls)
  if len(dirs) != 0 {
    t.Errorf("Length of %v should be 0 but it isn't!", dirs)
  }
  expectedErrString := "list of urls is empty! can't create output directories"
  if !reflect.DeepEqual(expectedErrString, err.Error()) {
    t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
  }
}

func TestIsValidUrl(t *testing.T) {
  validUrl := "http://localhost:9999/res-1.tar.gz"
  valid := IsValidUrl(validUrl)
  if !valid {
    t.Errorf("%v misidentified as an invalid url", validUrl)
  }
}

func TestIsValidUrlInvalid(t *testing.T) {
  invalidUrl := "\b Farnsworth is not a url."
  valid := IsValidUrl(invalidUrl)
  if valid {
    t.Errorf("%v misidentified as a valid url", invalidUrl)
  }
}

func TestCreateOutputDirectoriesFromUrlsUnparseable(t *testing.T) {
  urls := []string {
    "asdf-this-is-not-a-url",
  }
  _, err := CreateOutputDirectoriesFromUrls(urls)
  expectedErrString :=  fmt.Sprintf("%v is not a valid url", urls[0])
  if !reflect.DeepEqual(err.Error(), expectedErrString) {
    t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
  }
}

func TestDownloadTarFilesFails(t *testing.T) {
  artifactUrls := []string {
    "http://localhost:9998/does-not-exist.tar.gz",
  }
  _, err := DownloadTarFiles(artifactUrls)
  expectedErrString := "connect: connection refused"
  if err == nil {
    t.Errorf("Downloading non-existent tar files unexpectedly succeeded!")
  }
  if !strings.Contains(err.Error(), expectedErrString) {
    t.Errorf("Unexpected err string!\nExpected substring: %v\nActual: %v", expectedErrString, err.Error())
  }
}

func TestTarUpFilesInGivenDirectoriesInputValidation(t *testing.T) {
  allFilesMaps := make([]map[string][]byte, 10)
  dirsForFiles := []string {
    "Philip",
    "Hubert",
  }
  _, err :=  TarUpFilesInGivenDirectories(dirsForFiles, allFilesMaps)
  expectedErrString := "length of variables doesn't add up"
  if !reflect.DeepEqual(err.Error(), expectedErrString) {
    t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
  }
}

func TestFindArtifactUrlsByUuidFails(t *testing.T) {
  Uuid := "?"
  gutsCfg, err := ParseConfig("./guts-api.yaml")
  CheckError(err)
  db, err = PostgresConnect(gutsCfg)
  if SkipTestIfPostgresInactive(err) {
    t.Skip("Skipping test as postgresql service is not up")
  } else {
    CheckError(err)
  }
  _, err = FindArtifactUrlsByUuid(Uuid, db)
  expectedErrString := "No jobs with uuid ? found!"
  if !reflect.DeepEqual(err.Error(), expectedErrString) {
    t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
  }
}

func TestDownloadFile(t *testing.T) {
  ServeDirectory()
  artifactUrl := "http://localhost:9999/res-1.tar.gz"
  _, err := DownloadFile(artifactUrl)
  CheckError(err)
}

func TestDownloadFileEmpty(t *testing.T) {
  artifactUrl := "http://localhost:9999/empty.tar.gz"
  _, err := DownloadFile(artifactUrl)
  expectedErrString := "file at http://localhost:9999/empty.tar.gz is empty"
  if !reflect.DeepEqual(err.Error(), expectedErrString) {
    t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
  }
}
