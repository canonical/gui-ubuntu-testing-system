package main

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

func TestGetProtocolPrefix(t *testing.T) {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	GutsCfg.Api.Port = 8080
	expectedReturn := "http://"
	if expectedReturn != GetProtocolPrefix() {
		t.Errorf("%v is expected for port %v", expectedReturn, GutsCfg.Api.Port)
	}
	GutsCfg.Api.Port = 443
	expectedReturn = "https://"
	if expectedReturn != GetProtocolPrefix() {
		t.Errorf("%v is expected for port %v", expectedReturn, GutsCfg.Api.Port)
	}
	GutsCfg.Api.Port = 123
	expectedReturn = ""
	if expectedReturn != GetProtocolPrefix() {
		t.Errorf("%v is expected for port %v", expectedReturn, GutsCfg.Api.Port)
	}
	ParseArgs()
	err = ParseConfig(configFilePath)
	CheckError(err)
}
