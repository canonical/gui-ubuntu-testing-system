package utils

import (
	"os"
	"reflect"
	"testing"
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

func TestCreateDirIfNotExists(t *testing.T) {
	dummyDir := "/tmp/my-dummy-directory/"
	err := CreateDirIfNotExists(dummyDir)
	CheckError(err)
	_, err = os.Open(dummyDir)
	CheckError(err)
}

func TestCreateDirIfNotExistsDirExistsAlready(t *testing.T) {
	dummyDir := "/tmp/my-dummy-directory/"
	_, err := os.Open(dummyDir)
	if err == nil {
		err = os.RemoveAll(dummyDir)
		CheckError(err)
	}

	err = CreateDirIfNotExists(dummyDir)
	CheckError(err)
	_, err = os.Open(dummyDir)
	CheckError(err)

	err = CreateDirIfNotExists(dummyDir)
	CheckError(err)
	_, err = os.Open(dummyDir)
	CheckError(err)
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

func TestDownloadFile(t *testing.T) {
	ServeDirectory("/../../postgres/test-data/test-files/")
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

func TestGetProtocolPrefix(t *testing.T) {
	port := 8080
	expectedReturn := "http://"
	if expectedReturn != GetProtocolPrefix(port) {
		t.Errorf("%v is expected for port %v", expectedReturn, port)
	}
	port = 443
	expectedReturn = "https://"
	if expectedReturn != GetProtocolPrefix(port) {
		t.Errorf("%v is expected for port %v", expectedReturn, port)
	}
	port = 123
	expectedReturn = ""
	if expectedReturn != GetProtocolPrefix(port) {
		t.Errorf("%v is expected for port %v", expectedReturn, port)
	}
}

func TestInvalidUuidError(t *testing.T) {
	thisUuid := "576d00b4-bfbc-45fa-bfa6-736c46df81de"
	thisError := InvalidUuidError{uuid: thisUuid}
	expectedErrString := "576d00b4-bfbc-45fa-bfa6-736c46df81de isn't a valid uuid!"
	if thisError.Error() != expectedErrString {
		t.Errorf("expected err string not the same as actual\nexpected: %v\nactual: %v", expectedErrString, thisError.Error())
	}
}

func TestFileOrDirExistsSuccess(t *testing.T) {
	testDir := "/home"
	err := FileOrDirExists(testDir)
	if err != nil {
		t.Errorf("%v is a directory which should exist but apparently doesn't", testDir)
	}
}

func TestFileOrDirExistsFailure(t *testing.T) {
	testDir := "/home2"
	err := FileOrDirExists(testDir)
	if err == nil {
		t.Errorf("%v is a directory which shouldn't exist but apparently does", testDir)
	}
}

func TestAllFilesExistsSuccess(t *testing.T) {
	testFiles := []string{"/home/", "/root/", "/sys/"}
	if !AllFilesExist(testFiles[0], testFiles[1], testFiles[2]) {
		t.Errorf("apparently one of %v doesn't exist", testFiles)
	}
}

func TestAllFilesExistsFailure(t *testing.T) {
	testFiles := []string{"/home/", "/root/", "/dummy-dir/"}
	if AllFilesExist(testFiles[0], testFiles[1], testFiles[2]) {
		t.Errorf("apparently all of %v exist!", testFiles)
	}
}

func TestSha256sumOfString(t *testing.T) {
	myString := "inspector-5"
	mySha := "100e8c42469b60454bfb16a298ac5d3b700ff3c859b0c65ae551d929bdd00c37"
	shadString := Sha256sumOfString(myString)
	if shadString != mySha {
		t.Errorf("sha256sum of %v should be %v, but is instead %v", myString, mySha, shadString)
	}
}
