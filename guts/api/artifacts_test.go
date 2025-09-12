package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestFindArtifactUrlsByUuid(t *testing.T) {
	Uuid := "eccd3988-490d-4414-be97-605d1ac81073"
	err := ParseConfig("./guts-api.yaml")
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	result_urls, err := FindArtifactUrlsByUuid(Uuid)
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
	err := ParseConfig("./guts-api.yaml")
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	artifactsTarGz, err := CollateArtifacts(Uuid)
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
	err := ParseConfig("./guts-api.yaml")
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	artifactsTarGz, err := CollateArtifacts(Uuid)
	if len(artifactsTarGz) != 0 {
		t.Errorf("Variable should have length 0 but instead has length %v", len(artifactsTarGz))
	}
	expectedErrString := "gzip: invalid header"
	if !reflect.DeepEqual(err.Error(), expectedErrString) {
		t.Errorf("Unexpected error!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestCreateOutputDirectoriesFromUrls(t *testing.T) {
	artifactUrls := []string{
		"http://localhost:9999/res-1.tar.gz",
		"http://localhost:9999/res-2.tar.gz",
		"http://localhost:9999/res-3.tar.gz",
	}
	expectedDirNames := []string{
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
	artifactUrls := []string{
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

func TestCreateOutputDirectoriesFromUrlsUnparseable(t *testing.T) {
	urls := []string{
		"asdf-this-is-not-a-url",
	}
	_, err := CreateOutputDirectoriesFromUrls(urls)
	expectedErrString := fmt.Sprintf("%v is not a valid url", urls[0])
	if !reflect.DeepEqual(err.Error(), expectedErrString) {
		t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestDownloadTarFilesFails(t *testing.T) {
	artifactUrls := []string{
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
	dirsForFiles := []string{
		"Philip",
		"Hubert",
	}
	_, err := TarUpFilesInGivenDirectories(dirsForFiles, allFilesMaps)
	expectedErrString := "length of variables doesn't add up"
	if !reflect.DeepEqual(err.Error(), expectedErrString) {
		t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestFindArtifactUrlsByUuidFails(t *testing.T) {
	Uuid := "?"
	err := ParseConfig("./guts-api.yaml")
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	_, err = FindArtifactUrlsByUuid(Uuid)
	expectedErrString := "No jobs with uuid ? found!"
	if !reflect.DeepEqual(err.Error(), expectedErrString) {
		t.Errorf("Unexpected err string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestCacheRetentionPolicyDirNoExist(t *testing.T) {
	fakeDir := "/srv/this-dir-noexist"
	err := CacheRetentionPolicy(fakeDir)
	if err == nil {
		t.Errorf("The cache retention policy didn't fail for a non-existent directory: %v", fakeDir)

	}
}

func TestCacheRetentionPolicySuccess(t *testing.T) {
	// Set necessary variables for a small cache
	savedCacheMaxSize := GutsCfg.Tarball.TarballCacheMaxSize
	savedCacheReductionThreshold := GutsCfg.Tarball.TarballCacheReductionThreshold
	savedCachePath := GutsCfg.Tarball.TarballCachePath
	GutsCfg.Tarball.TarballCacheMaxSize = 100000
	GutsCfg.Tarball.TarballCacheReductionThreshold = 90000
	GutsCfg.Tarball.TarballCachePath = "/srv/dummy-cache/"
	// create the cache dir
	err := CreateDirIfNotExists(GutsCfg.Tarball.TarballCachePath)
	CheckError(err)
	// Populate the cache until above the MaxSize
	err = PopulateCacheDummyData(GutsCfg.Tarball.TarballCacheMaxSize)
	CheckError(err)
	dirSize, err := GetDirSize(GutsCfg.Tarball.TarballCachePath)
	CheckError(err)
	if dirSize < GutsCfg.Tarball.TarballCacheMaxSize {
		t.Errorf("Dirsize %v should be greater than %v", dirSize, GutsCfg.Tarball.TarballCacheMaxSize)
	}
	// Run the CacheRetentionPolicy
	err = CacheRetentionPolicy(GutsCfg.Tarball.TarballCachePath)
	CheckError(err)
	dirSize, err = GetDirSize(GutsCfg.Tarball.TarballCachePath)
	CheckError(err)
	if dirSize > GutsCfg.Tarball.TarballCacheMaxSize {
		t.Errorf("Dirsize %v should be less than %v", dirSize, GutsCfg.Tarball.TarballCacheMaxSize)
	}
	// remove the cache-dir
	err = os.RemoveAll(GutsCfg.Tarball.TarballCachePath)
	CheckError(err)
	// Set necessary variables back
	GutsCfg.Tarball.TarballCacheMaxSize = savedCacheMaxSize
	GutsCfg.Tarball.TarballCacheReductionThreshold = savedCacheReductionThreshold
	GutsCfg.Tarball.TarballCachePath = savedCachePath
}

func PopulateCacheDummyData(threshold int) error {
	dirSize := 0
	for dirSize < threshold {
		thisUuid := uuid.New().String()
		thisDir := fmt.Sprintf("%v%v", GutsCfg.Tarball.TarballCachePath, thisUuid)
		err := os.Mkdir(thisDir, 0755)
		if err != nil {
			return err
		}
		tarball := make([]byte, 999)
		_, err = rand.Read(tarball)
		CheckError(err)
		err = WriteTarballToCache(tarball, thisUuid, thisDir, fmt.Sprintf("%v/results.tar.gz", thisDir), fmt.Sprintf("%v/%v.last_downloaded", thisDir, thisUuid))
		CheckError(err)
		dirSize, err = GetDirSize(GutsCfg.Tarball.TarballCachePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestWriteTarballToCacheAlreadyExists(t *testing.T) {
	thisUuid := uuid.New().String()
	thisDir := fmt.Sprintf("%v%v", GutsCfg.Tarball.TarballCachePath, thisUuid)
	err := os.Mkdir(thisDir, 0755)
	CheckError(err)
	tarball := make([]byte, 999)
	_, err = rand.Read(tarball)
	CheckError(err)
	err = WriteTarballToCache(tarball, thisUuid, thisDir, fmt.Sprintf("%v/results.tar.gz", thisDir), fmt.Sprintf("%v/%v.last_downloaded", thisDir, thisUuid))
	CheckError(err)
	// okay, now it's written ... run again and ensure no failure?
	err = WriteTarballToCache(tarball, thisUuid, thisDir, fmt.Sprintf("%v/results.tar.gz", thisDir), fmt.Sprintf("%v/%v.last_downloaded", thisDir, thisUuid))
	CheckError(err)
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

func TestTarUpFilesInGivenDirectories(t *testing.T) {
	// create the desired directories
	fileDirs := []string{"dir1", "dir2", "dir3"}
	// create the slice of maps of files for each directory
	myMap := make([]map[string][]byte, 3)
	// Create a map for each directory
	// var mapFiles map[string][]byte
	mapFiles := make(map[string][]byte)
	mapFiles["file1"] = make([]byte, 99)
	_, err := rand.Read(mapFiles["file1"])
	CheckError(err)
	mapFiles["file2"] = make([]byte, 99)
	_, err = rand.Read(mapFiles["file2"])
	CheckError(err)
	mapFiles["file3"] = make([]byte, 99)
	_, err = rand.Read(mapFiles["file3"])
	CheckError(err)
	// Insert map into slice
	myMap[0] = mapFiles
	myMap[1] = mapFiles
	myMap[2] = mapFiles
	// Tar up the files
	tarBytes, err := TarUpFilesInGivenDirectories(fileDirs, myMap)
	CheckError(err)
	// write to a tempfile
	f, err := os.CreateTemp("", "tarfile")
	CheckError(err)
	defer DeferredErrCheckStringArg(os.Remove, f.Name())
	_, err = f.Write(tarBytes)
	CheckError(err)
	// extract list of files in archive
	out, err := exec.Command("tar", "-tf", f.Name()).Output()
	CheckError(err)
	// compare to expected archive
	// The order of output from tar -tf isn't consistent, so we can't rely on the string
	expectedArchiveStr := "\ndir1/file1\ndir1/file2\ndir1/file3\ndir2/file1\ndir2/file2\ndir2/file3\ndir3/file1\ndir3/file2\ndir3/file3"
	actualArchiveStr := string(out)
	expectedArchive := strings.Split(expectedArchiveStr, "\n")
	actualArchive := strings.Split(actualArchiveStr, "\n")
	slices.Sort(actualArchive)
	if !reflect.DeepEqual(expectedArchive, actualArchive) {
		t.Errorf("Expected archive not the same as actual!\nExpected: %v\nActual: %v", expectedArchive, actualArchive)
	}
}
