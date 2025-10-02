package spawner

import (
	"fmt"
	"guts.ubuntu.com/v2/utils"
	"os"
	"testing"
)

func TestGetRemoteShaSum(t *testing.T) {
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	shasum, err := GetRemoteShaSum(imageUrl)
	utils.CheckError(err)
	expectedShasum := "a52d5d22d71375efae79de6cf8a125228ac19356c84f4af17ef5955147be7ef5"
	if shasum != expectedShasum {
		t.Errorf("Parsed shasum not same as expected!\nExpected: %v\nActual: %v", expectedShasum, shasum)
	}
}

func TestGetRemoteShaSumFailure(t *testing.T) {
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://planetexpress.com/questing-mini-iso-amd64.iso"
	_, err := GetRemoteShaSum(imageUrl)
	expectedErrString := fmt.Sprintf("Couldn't acquire shasum of image at %v", imageUrl)
	if err.Error() != expectedErrString {
		t.Errorf("unexpected err string!\nexpected: %v\nactual: %v", expectedErrString, err.Error())
	}
}

func TestCdImageGetShasumOfImage(t *testing.T) {
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	shasum, err := CdImageGetShasumOfImage(imageUrl)
	utils.CheckError(err)
	expectedShasum := "a52d5d22d71375efae79de6cf8a125228ac19356c84f4af17ef5955147be7ef5"
	if shasum != expectedShasum {
		t.Errorf("Parsed shasum not same as expected!\nExpected: %v\nActual: %v", expectedShasum, shasum)
	}
}

func TestCdImageGetShasumOfImageFailure(t *testing.T) {
	imageUrl := "http://planetexpress.com/questing-mini-iso-amd64.iso"
	shasum, err := CdImageGetShasumOfImage(imageUrl)
	if err == nil {
		t.Errorf("acquiring shasum file from %v succeeded where it should have failed", imageUrl)
	}
	if shasum != "" {
		t.Errorf("parsed shasum should be empty but is instead: %v", shasum)
	}
}

func TestCdImageDownloadCheckSumFileForImage(t *testing.T) {
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	shasums, err := CdImageDownloadCheckSumFileForImage(imageUrl)
	utils.CheckError(err)
	expectedShasumFile := "a52d5d22d71375efae79de6cf8a125228ac19356c84f4af17ef5955147be7ef5 *questing-mini-iso-amd64.iso\n"
	if expectedShasumFile != shasums {
		t.Errorf("Parsed shasum file not the same as expected!\nExpected: %v\nActual: %v", expectedShasumFile, shasums)
	}
}

func TestCdImageParseShasumForImage(t *testing.T) {
	shasumFile := "a4310d26648801af766733bf01845d7d2f4d26a96a02ea45ee532d74068999a0 *questing-desktop-amd64.iso\nf425a4872fdd163f38ec785eaa1bdab1f1bdae202b67247f57ed1aa96a3a20a4 *questing-desktop-arm64.iso\n"
	desiredImage := "questing-desktop-amd64.iso"
	parsedShasum, err := CdImageParseShasumForImage(shasumFile, desiredImage)
	utils.CheckError(err)
	expectedShasum := "a4310d26648801af766733bf01845d7d2f4d26a96a02ea45ee532d74068999a0"
	if parsedShasum != expectedShasum {
		t.Errorf("Parsed shasum not the same as expected!\nExpected: %v\nActual: %v", expectedShasum, parsedShasum)
	}
}

func TestCdImageParseShasumForImageFails(t *testing.T) {
	shasumFile := "asdf *questing-desktop-amd64.iso\nasdf *questing-desktop-arm64.iso\n"
	desiredImage := "questing-desktop-amd64.iso"
	parsedShasum, err := CdImageParseShasumForImage(shasumFile, desiredImage)
	if err == nil {
		t.Errorf("Parsing shasum for %v from %v succeeded where it should have failed", desiredImage, shasumFile)
	}
	if parsedShasum != "" {
		t.Errorf("parsed shasum should be empty but is instead: %v", parsedShasum)
	}
}

func TestGetLocalShaSum(t *testing.T) {
	testString := "delta-brainwave"
	f, err := os.CreateTemp("", "testshafile")
	utils.CheckError(err)
	defer utils.DeferredErrCheckStringArg(os.Remove, f.Name())
	err = os.WriteFile(f.Name(), []byte(testString), 0644)
	utils.CheckError(err)
	expectedShasum := "efe716a6fedbbc1ace5186dd81b5308435360ce7113bc0dfba539a1c0bc79907"
	actualShasum, err := GetLocalShaSum(f.Name())
	utils.CheckError(err)
	if actualShasum != expectedShasum {
		t.Errorf("expected shasum not the same as actual\nexpected: %v\nactual: %v", expectedShasum, actualShasum)
	}
}

func TestGetLocalShaSumNoFile(t *testing.T) {
	dummyFile := "/file/does/not/exist"
	actualShasum, err := GetLocalShaSum(dummyFile)
	if err == nil {
		t.Errorf("Getting shasum for %v should have failed but didn't", dummyFile)
	}
	if actualShasum != "" {
		t.Errorf("parsed shasum should be empty but is instead: %v", actualShasum)
	}
}
