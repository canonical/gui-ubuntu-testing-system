package utils

import (
	"archive/tar"
	"bytes"
	"fmt"
	//"io"
	"gopkg.in/yaml.v3"
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
	servingProcess := ServeRelativeDirectory("/../../postgres/test-data/test-files/")
	defer DeferredErrCheck(servingProcess.Kill)

	artifactUrl := "http://localhost:9999/res-1.tar.gz"
	_, err := DownloadFile(artifactUrl)
	CheckError(err)
}

func TestDownloadFileEmpty(t *testing.T) {
	servingProcess := ServeRelativeDirectory("/../../postgres/test-data/test-files/")
	defer DeferredErrCheck(servingProcess.Kill)

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

func TestGetFileNameFromUrl(t *testing.T) {
	url := "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	expectedName := "questing-desktop-amd64.iso"
	actualName := GetFileNameFromUrl(url)
	if expectedName != actualName {
		t.Errorf("image name parsed from %v should be %v but is %v", url, expectedName, actualName)
	}
}

func TestGenericGitError(t *testing.T) {
	gitCmd := []string{"git", "status"}
	gitErr := GenericGitError{Command: gitCmd}
	desiredErrString := fmt.Sprintf("Git operation failed:\n%v", gitCmd)
	if gitErr.Error() != desiredErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, gitErr.Error())
	}
}

func TestGzipTarArchiveBytes(t *testing.T) {
	myBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	compressedBytes, err := GzipTarArchiveBytes(myBytes)
	CheckError(err)
	targetBytes := []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 98, 100, 98, 102, 97, 101, 99, 231, 224, 228, 2, 4, 0, 0, 255, 255, 123, 87, 32, 37, 10, 0, 0, 0}
	if !reflect.DeepEqual(compressedBytes, targetBytes) {
		t.Errorf("Unexpected output bytes after gzip compression!\nExpected: %v\nActual: %v", targetBytes, compressedBytes)
	}
}

func TestMd5SumOfBytes(t *testing.T) {
	myBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	md5Sum := Md5SumOfBytes(myBytes)
	targetSum := "70903e79b7575e3f4e7ffa15c2608ac7"
	if md5Sum != targetSum {
		t.Errorf("unexpected md5 sum of %v!\nExpected: %v\nActual: %v", myBytes, targetSum, md5Sum)
	}
}

func TestTarUpDirectory(t *testing.T) {
	// Create test tar directory
	tarDir, err := os.MkdirTemp("", "tartest")
	CheckError(err)
	asdfFile := fmt.Sprintf("%v/asdf", tarDir)
	err = os.MkdirAll(tarDir, 0755)
	CheckError(err)
	err = os.WriteFile(asdfFile, []byte("asdf"), 0644)
	CheckError(err)
	// now directly test TarUpDirectory
	tarBytes, err := TarUpDirectory(tarDir)
	CheckError(err)
	err = os.RemoveAll(tarDir)
	CheckError(err)
	expectedLen := 2048
	if len(tarBytes) != expectedLen {
		t.Errorf("unexpected length of tardata!\nExpected: %v\nActual: %v", expectedLen, len(tarBytes))
	}
}

func TestWriteToTarFile(t *testing.T) {
	var tarData []byte
	tarBuffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(tarBuffer)
	// create temporary file
	tarDir, err := os.MkdirTemp("", "tartest")
	CheckError(err)
	asdfFile := fmt.Sprintf("%v/asdf", tarDir)
	err = os.MkdirAll(tarDir, 0755)
	CheckError(err)
	err = os.WriteFile(asdfFile, []byte("asdf"), 0644)
	CheckError(err)
	// get file info of file
	f, err := os.Open(asdfFile)
	CheckError(err)
	fi, err := f.Stat()
	CheckError(err)
	err = f.Close()
	CheckError(err)
	// pass it to WriteToTar
	err = WriteToTar(asdfFile, tarWriter, fi)
	// then:
	err = tarWriter.Close()
	CheckError(err)

	tarData = tarBuffer.Bytes()
	err = os.RemoveAll(tarDir)
	CheckError(err)
	expectedLen := 2048
	if len(tarData) != expectedLen {
		t.Errorf("unexpected length of tardata!\nExpected: %v\nActual: %v", expectedLen, len(tarData))
	}
}

func TestWriteToTarDir(t *testing.T) {
	var tarData []byte
	tarBuffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(tarBuffer)
	// create temporary file
	tarDir, err := os.MkdirTemp("", "tartest")
	CheckError(err)
	asdfFile := fmt.Sprintf("%v/asdf", tarDir)
	err = os.MkdirAll(tarDir, 0755)
	CheckError(err)
	err = os.WriteFile(asdfFile, []byte("asdf"), 0644)
	CheckError(err)
	// get file info of file
	f, err := os.Open(tarDir)
	CheckError(err)
	fi, err := f.Stat()
	CheckError(err)
	err = f.Close()
	CheckError(err)
	// pass it to WriteToTar
	err = WriteToTar(tarDir, tarWriter, fi)
	// then:
	err = tarWriter.Close()
	CheckError(err)

	tarData = tarBuffer.Bytes()
	err = os.RemoveAll(tarDir)
	CheckError(err)

	expectedLen := 1536
	if len(tarData) != expectedLen {
		t.Errorf("unexpected length of tardata!\nExpected: %v\nActual: %v", expectedLen, len(tarData))
	}
}

func TestTraverseDirectory(t *testing.T) {
	var tarData []byte
	tarBuffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(tarBuffer)
	// create temporary file
	tarDir, err := os.MkdirTemp("", "tartest")
	CheckError(err)
	asdfFile := fmt.Sprintf("%v/asdf", tarDir)
	err = os.MkdirAll(tarDir, 0755)
	CheckError(err)
	err = os.WriteFile(asdfFile, []byte("asdf"), 0644)
	CheckError(err)

	err = TraverseDirectory(tarDir, tarWriter)
	CheckError(err)
	err = tarWriter.Close()
	CheckError(err)

	tarData = tarBuffer.Bytes()
	err = os.RemoveAll(tarDir)
	CheckError(err)

	expectedLen := 2048
	if len(tarData) != expectedLen {
		t.Errorf("unexpected length of tardata!\nExpected: %v\nActual: %v", expectedLen, len(tarData))
	}
}

func TestEnsureGetEnvSuccess(t *testing.T) {
	testString := "roswell-that-ends-well"
	envVar := "DELTA"
	os.Setenv(envVar, testString)
	accString, err := EnsureGetEnv(envVar)
	os.Unsetenv(envVar)
	CheckError(err)
	if accString != testString {
		t.Errorf("Env var not as expected!\nExpected: %v\nActual: %v", testString, accString)
	}
}

func TestEnsureGetEnvFailure(t *testing.T) {
	envVar := "DELTA"
	_, err := EnsureGetEnv(envVar)
	expectedErrString := fmt.Sprintf("tried to get %v environment variable, but it was empty or unset", envVar)
	if err == nil {
		t.Errorf("Getting env var %v should have failed but did not", envVar)
	}
	if err.Error() != expectedErrString {
		t.Errorf("unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestStartProcess(t *testing.T) {
	cmdArgs := []string{
		"sleep",
		"1",
	}
	process, err := StartProcess(cmdArgs, nil)
	CheckError(err)

	err = process.Wait()
	CheckError(err)
}

func TestStartProcessWithEnv(t *testing.T) {
	cmdArgs := []string{
		"sleep",
		"1",
	}
	envVars := []string{
		"VNC_HOST=asdf",
		"VNC_PORT=12345",
	}
	process, err := StartProcess(cmdArgs, &envVars)
	CheckError(err)

	err = process.Wait()
	CheckError(err)
}

func TestParsePlan(t *testing.T) {
	fullPlan := `---
tests:
  Firefox-Example-Basic:
    entrypoint: tests/firefox-example
  Firefox-Example-New-Tab:
    entrypoint: tests/firefox-example`

	dname, err := os.MkdirTemp("", "sampledir")
	CheckError(err)

	defer DeferredErrCheckStringArg(os.RemoveAll, dname)

	testPlanFn := fmt.Sprintf("%v/testPlan.yaml", dname)
	err = os.WriteFile(testPlanFn, []byte(fullPlan), 0644)
	CheckError(err)

	parsedPlan, err := ParsePlan(testPlanFn)
	CheckError(err)

	expectedCases := make(TestCases, 2)
	expectedCases[0].Name = "Firefox-Example-Basic"
	expectedCases[0].Data.EntryPoint = "tests/firefox-example"
	expectedCases[0].Data.Requirements.Tpm = false
	expectedCases[1].Name = "Firefox-Example-New-Tab"
	expectedCases[1].Data.EntryPoint = "tests/firefox-example"
	expectedCases[1].Data.Requirements.Tpm = false

	if len(parsedPlan.Tests) != len(expectedCases) {
		t.Errorf("unexpected number of plans parsed!\nexpected: %v\nactual: %v", len(expectedCases), len(parsedPlan.Tests))
	}

	for idx, entry := range parsedPlan.Tests {
		if entry != expectedCases[idx] {
			t.Errorf("unexpected test case!\nexpected: %v\nactual: %v", expectedCases[idx], entry)
		}
	}
}

func TestParsePlanBadFile(t *testing.T) {
	fullPlan := `this-is-not-a-yaml-file`

	dname, err := os.MkdirTemp("", "sampledir")
	CheckError(err)

	defer DeferredErrCheckStringArg(os.RemoveAll, dname)

	testPlanFn := fmt.Sprintf("%v/testPlan.yaml", dname)
	err = os.WriteFile(testPlanFn, []byte(fullPlan), 0644)
	CheckError(err)

	parsedPlan, err := ParsePlan(testPlanFn)
	if err == nil {
		t.Errorf("parsing plan %v succeeded where it should have failed!\nparsed plan is: %v", fullPlan, parsedPlan)
	}

	expectedErrType := yaml.TypeError{}

	if !reflect.DeepEqual(reflect.TypeOf(err), reflect.TypeOf(&expectedErrType)) {
		t.Errorf("unexpected error type!\nexpected: %v\nactual: %v", reflect.TypeOf(err), reflect.TypeOf(&expectedErrType))
	}
}

func TestGitCloneToDir(t *testing.T) {
	repo := "https://github.com/canonical/ubuntu-gui-testing.git"
	branch := "main"
	tmpDir, err := os.MkdirTemp("", "clonetest")
	CheckError(err)

	defer DeferredErrCheckStringArg(os.RemoveAll, tmpDir)

	err = GitCloneToDir(repo, branch, tmpDir)
	CheckError(err)
}
