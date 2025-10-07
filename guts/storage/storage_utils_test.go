package storage

import (
	"fmt"
	"guts.ubuntu.com/v2/utils"
	"os"
	"testing"
)

func TestGetStorageBackendLocalSuccess(t *testing.T) {
	testPath := "/srv/data"

	// remove object path
	err := os.RemoveAll(testPath)
	utils.CheckError(err)

	// create object path to test
	err = os.MkdirAll(testPath, 0755)
	utils.CheckError(err)

	// test pre-existing object_path
	thisStrgCfg := make(map[string]string)
	thisStrgCfg["provider"] = "local"
	thisStrgCfg["object_path"] = testPath
	thisStrgCfg["object_port"] = "9999"
	thisStrgCfg["object_host"] = "http://localhost"
	_, err = GetStorageBackend(thisStrgCfg)
	utils.CheckError(err)

	// remove object path
	err = os.RemoveAll(testPath)
	utils.CheckError(err)

	// test with non-existent path
	bknd, err := GetStorageBackend(thisStrgCfg)
	utils.CheckError(err)

	// test backend upload
	container := "mycont"
	remotePath := "path/to/my/file.txt"
	fileContent := "hello"
	dataBytes := []byte(fileContent)
	fullUrl, err := bknd.Upload(container, remotePath, dataBytes)
	utils.CheckError(err)

	// check file is written

	servingProcess := utils.ServeDirectory(testPath)
	defer utils.DeferredErrCheck(servingProcess.Kill)

	realPath := fmt.Sprintf("%v/%v/%v", testPath, container, remotePath)
	err = utils.FileOrDirExists(realPath)
	utils.CheckError(err)

	// check if we can download the file?
	fileBytes, err := utils.DownloadFile(fullUrl)
	utils.CheckError(err)

	if string(fileBytes) != fileContent {
		t.Errorf("downloading file from server resulted in unexpected content!\nExpected: %v\nActual: %v", fileContent, string(fileBytes))
	}

	// check path is as expected
	expectedPath := fmt.Sprintf("%v:%v/%v/%v", thisStrgCfg["object_host"], thisStrgCfg["object_port"], container, remotePath)
	if expectedPath != fullUrl {
		t.Errorf("unexpected object path!\nExpected: %v\nActual: %v", expectedPath, fullUrl)
	}
}

func TestGetStorageBackendLocalFailure(t *testing.T) {
	thisStrgCfg := make(map[string]string)
	thisStrgCfg["provider"] = "local"
	_, err := GetStorageBackend(thisStrgCfg)
	expectedErrString := "object_path not set"
	if err.Error() != expectedErrString {
		t.Errorf("unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}

	thisStrgCfg["object_path"] = "/srv/data/"
	_, err = GetStorageBackend(thisStrgCfg)
	expectedErrString = "object_port not set"
	if err.Error() != expectedErrString {
		t.Errorf("unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}

	thisStrgCfg["object_port"] = "9999"
	_, err = GetStorageBackend(thisStrgCfg)
	expectedErrString = "object_host not set"
	if err.Error() != expectedErrString {
		t.Errorf("unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}

	thisStrgCfg["object_host"] = "http://localhost"
	_, err = GetStorageBackend(thisStrgCfg)
	utils.CheckError(err)
}

func TestGetStorageBackendAsdf(t *testing.T) {
	thisStrgCfg := make(map[string]string)
	thisStrgCfg["provider"] = "asdf"
	_, err := GetStorageBackend(thisStrgCfg)
	if err == nil {
		t.Errorf("getting storage backend for backend %v succeeded where it should have failed", thisStrgCfg["provider"])
	}
}

func TestGetStorageBackendSwiftSuccess(t *testing.T) {
	thisStrgCfg := make(map[string]string)
	thisStrgCfg["provider"] = "swift"
	thisStrgCfg["swift_user"] = "user"
	thisStrgCfg["swift_api_key"] = "api-key"
	thisStrgCfg["swift_auth_url"] = "auth-url"
	thisStrgCfg["swift_domain"] = "Default"
	thisStrgCfg["swift_tenant"] = "Default"
	_, err := GetStorageBackend(thisStrgCfg)
	expectedErrString := "Can't find AuthVersion in AuthUrl - set explicitly"
	if err.Error() != expectedErrString {
		t.Errorf("GetStorageBackend for provider %v failed with unexpected error!\nExpected: %v\nActual: %v", thisStrgCfg["provider"], expectedErrString, err.Error())
	}
}

func TestGetStorageBackendSwiftFailures(t *testing.T) {
	thisStrgCfg := make(map[string]string)
	thisStrgCfg["provider"] = "swift"
	thisStrgCfg["mph"] = "user"
	_, err := GetStorageBackend(thisStrgCfg)
	expectedErrString := "swift_user not set"
	if err.Error() != expectedErrString {
		t.Errorf("GetStorageBackend for provider %v failed with unexpected error!\nExpected: %v\nActual: %v", thisStrgCfg["provider"], expectedErrString, err.Error())
	}

	thisStrgCfg = make(map[string]string)
	thisStrgCfg["provider"] = "swift"
	thisStrgCfg["swift_user"] = "user"
	_, err = GetStorageBackend(thisStrgCfg)
	expectedErrString = "swift_api_key not set"
	if err.Error() != expectedErrString {
		t.Errorf("GetStorageBackend for provider %v failed with unexpected error!\nExpected: %v\nActual: %v", thisStrgCfg["provider"], expectedErrString, err.Error())
	}

	thisStrgCfg["swift_api_key"] = "api-key"
	_, err = GetStorageBackend(thisStrgCfg)
	expectedErrString = "swift_auth_url not set"
	if err.Error() != expectedErrString {
		t.Errorf("GetStorageBackend for provider %v failed with unexpected error!\nExpected: %v\nActual: %v", thisStrgCfg["provider"], expectedErrString, err.Error())
	}

	thisStrgCfg["swift_auth_url"] = "auth-url"
	_, err = GetStorageBackend(thisStrgCfg)
	expectedErrString = "swift_domain not set"
	if err.Error() != expectedErrString {
		t.Errorf("GetStorageBackend for provider %v failed with unexpected error!\nExpected: %v\nActual: %v", thisStrgCfg["provider"], expectedErrString, err.Error())
	}

	thisStrgCfg["swift_domain"] = "Default"
	_, err = GetStorageBackend(thisStrgCfg)
	expectedErrString = "swift_tenant not set"
	if err.Error() != expectedErrString {
		t.Errorf("GetStorageBackend for provider %v failed with unexpected error!\nExpected: %v\nActual: %v", thisStrgCfg["provider"], expectedErrString, err.Error())
	}

	thisStrgCfg["swift_tenant"] = "Default"
	_, err = GetStorageBackend(thisStrgCfg)
	expectedErrString = "Can't find AuthVersion in AuthUrl - set explicitly"
	if err.Error() != expectedErrString {
		t.Errorf("GetStorageBackend for provider %v failed with unexpected error!\nExpected: %v\nActual: %v", thisStrgCfg["provider"], expectedErrString, err.Error())
	}
}
