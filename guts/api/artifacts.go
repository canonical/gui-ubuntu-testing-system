package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	DirSize int
)

func CollateArtifacts(uuidToFind string) ([]byte, error) { // coverage-ignore
	// Collates several .tar.gz files from separate tests into one .tar.gz file
	var gzippedTarBytes []byte
	uuidCacheDir := fmt.Sprintf("%v%v", GutsCfg.Tarball.TarballCachePath, uuidToFind)
	cachedTarFile := fmt.Sprintf("%v/results.tar.gz", uuidCacheDir)
	cachedLastDownloadedFile := fmt.Sprintf("%v/%v.last_downloaded", uuidCacheDir, uuidToFind)

	if AllFilesExist(uuidCacheDir, cachedTarFile, cachedLastDownloadedFile) {
		dat, err := os.ReadFile(cachedTarFile)
		if err != nil {
			return dat, err
		}
		err = RefreshLastDownloadedFile(cachedLastDownloadedFile)
		if err != nil {
			return dat, err
		}
		return dat, err
	}

	urls, err := FindArtifactUrlsByUuid(uuidToFind)
	if err != nil {
		return gzippedTarBytes, err
	}

	directoryNames, err := CreateOutputDirectoriesFromUrls(urls)
	if err != nil {
		return gzippedTarBytes, err
	}

	allTarFilesBytes, err := DownloadTarFiles(urls)
	if err != nil {
		return gzippedTarBytes, err
	}

	allFilesInTarBytes, err := TarUpFilesInGivenDirectories(directoryNames, allTarFilesBytes)
	if err != nil {
		return gzippedTarBytes, err
	}

	gzippedTarBytes, err = GzipTarArchiveBytes(allFilesInTarBytes)
	if err != nil {
		return gzippedTarBytes, err
	}

	err = CreateDirIfNotExists(GutsCfg.Tarball.TarballCachePath)
	if err != nil {
		return gzippedTarBytes, err
	}

	err = WriteTarballToCache(gzippedTarBytes, uuidToFind, uuidCacheDir, cachedTarFile, cachedLastDownloadedFile)
	if err != nil {
		return gzippedTarBytes, err
	}

	// The error is here!
	err = CacheRetentionPolicy(GutsCfg.Tarball.TarballCachePath)
	if err != nil {
		return gzippedTarBytes, err
	}

	return gzippedTarBytes, nil
}

func CacheRetentionPolicy(cacheDirectory string) error {
	// Implements a cache retention policy, ensuring the cache stays below a certain size

	entries, err := os.ReadDir(cacheDirectory)
	if err != nil {
		return err
	}

	uuidEpochMap := make(map[int]string)
	sliceOfStamps := make([]int, len(entries))

	for idx, e := range entries {
		entryLastUpdatedFile := fmt.Sprintf("%v/%v/%v.last_downloaded", cacheDirectory, e.Name(), e.Name())
		dat, err := os.ReadFile(entryLastUpdatedFile)
		if err != nil { // coverage-ignore
			return err
		}
		lastUpdated, err := strconv.Atoi(string(dat))
		if err != nil { // coverage-ignore
			return err
		}
		uuidEpochMap[lastUpdated] = e.Name()
		sliceOfStamps[idx] = lastUpdated
	}

	sort.Ints(sliceOfStamps)

	for _, lastUpdated := range sliceOfStamps {
		thisSpecificDir := fmt.Sprintf("%v/%v", cacheDirectory, uuidEpochMap[lastUpdated])
		dirSize, err := GetDirSize(cacheDirectory)
		if err != nil { // coverage-ignore
			return err
		}
		if dirSize < GutsCfg.Tarball.TarballCacheReductionThreshold {
			return nil
		}
		err = os.RemoveAll(thisSpecificDir)
		if err != nil { // coverage-ignore
			return err
		}
	}

	return nil
}

func RefreshLastDownloadedFile(cachedLastDownloadedFile string) error {
	now := time.Now().Unix()
	lastDownloaded := []byte(strconv.FormatInt(now, 10))
	err := AtomicWrite(lastDownloaded, cachedLastDownloadedFile)
	if err != nil { // coverage-ignore
		return err
	}
	return nil
}

func WriteTarballToCache(tarBall []byte, uuid string, uuidCacheDir string, cachedTarFile string, cachedLastDownloadedFile string) error {
	if AllFilesExist(uuidCacheDir, cachedTarFile, cachedLastDownloadedFile) {
		err := RefreshLastDownloadedFile(cachedLastDownloadedFile)
		if err != nil { // coverage-ignore
			return err
		}
		return nil
	}

	err := os.RemoveAll(uuidCacheDir)
	if err != nil { // coverage-ignore
		return err
	}

	err = os.Mkdir(uuidCacheDir, 0755)
	if err != nil { // coverage-ignore
		return err
	}

	err = AtomicWrite(tarBall, cachedTarFile)
	if err != nil { // coverage-ignore
		return err
	}

	err = RefreshLastDownloadedFile(cachedLastDownloadedFile)
	if err != nil { //coverage-ignore
		return err
	}

	return nil
}

func GzipTarArchiveBytes(tarArchive []byte) ([]byte, error) {
	var buf bytes.Buffer
	var returnBytes []byte
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(tarArchive)
	if err != nil { // coverage-ignore
		return returnBytes, err
	}
	err = gz.Close()
	if err != nil { // coverage-ignore
		return returnBytes, err
	}
	returnBytes = buf.Bytes()
	return returnBytes, nil
}

func CreateOutputDirectoriesFromUrls(urls []string) ([]string, error) {
	var listOfFilenames []string
	if len(urls) == 0 {
		return listOfFilenames, errors.New("list of urls is empty! can't create output directories")
	}
	for _, thisUrl := range urls {
		if !IsValidUrl(thisUrl) {
			return listOfFilenames, fmt.Errorf("%v is not a valid url", thisUrl)
		}
		splitUrl := strings.Split(thisUrl, `/`)
		filename := splitUrl[len(splitUrl)-1]
		directory := strings.Split(filename, `.`)[0]
		listOfFilenames = append(listOfFilenames, directory)
	}
	return listOfFilenames, nil
}

func DownloadTarFiles(tarfileUrls []string) ([]map[string][]byte, error) {
	var allFilesMaps []map[string][]byte

	for _, url := range tarfileUrls {
		var b []byte
		b, err := DownloadFile(url)
		if err != nil {
			return allFilesMaps, err
		}
		tarFiles, err := ExtractTarfiles(b)
		if err != nil {
			return allFilesMaps, err
		}
		allFilesMaps = append(allFilesMaps, tarFiles)
	}
	return allFilesMaps, nil
}

func TarUpFilesInGivenDirectories(dirsForFiles []string, inputFiles []map[string][]byte) ([]byte, error) {
	var returnBytes []byte
	if !reflect.DeepEqual(len(dirsForFiles), len(inputFiles)) {
		return returnBytes, errors.New("length of variables doesn't add up")
	}

	// initialise tar writer
	tarBuffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(tarBuffer)

	for idx, entry := range dirsForFiles {
		tarFiles := inputFiles[idx]
		for fileName, fileBytes := range tarFiles {
			hdr := &tar.Header{
				Name: entry + "/" + fileName,
				Mode: 0644,
				Size: int64(len(fileBytes)),
			}
			// write tar header
			err := tarWriter.WriteHeader(hdr)
			if err != nil { // coverage-ignore
				return returnBytes, err
			}
			// write tar data
			_, err = tarWriter.Write(fileBytes)
			if err != nil { // coverage-ignore
				return returnBytes, err
			}
		}
	}

	// close tar writer
	err := tarWriter.Close()
	if err != nil { // coverage-ignore
		return returnBytes, err
	}

	returnBytes = tarBuffer.Bytes()

	return returnBytes, nil
}

func FindArtifactUrlsByUuid(uuidToFind string) ([]string, error) {
	var result_urls []string
	var params = []string{"results_url"}
	rows, err := Driver.Query("tests", "uuid", uuidToFind, params)
	if err != nil { // coverage-ignore
		return result_urls, err
	}
	defer DeferredErrCheck(rows.Close)
	for rows.Next() {
		var result_url string
		err := rows.Scan(&result_url)
		if err != nil { // coverage-ignore
			return result_urls, err
		}
		result_urls = append(result_urls, result_url)
	}

	if len(result_urls) == 0 {
		return result_urls, UuidNotFoundError{uuid: uuidToFind}
	}

	return result_urls, nil
}

func ExtractTarfiles(tarFileBytes []byte) (map[string][]byte, error) {
	returnData := make(map[string][]byte)

	r := bytes.NewReader(tarFileBytes)
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return returnData, err
	}
	defer DeferredErrCheck(gzipReader.Close)
	tarReader := tar.NewReader(gzipReader)
	var maxBytes int64 = 4000000
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil { // coverage-ignore
			return returnData, err
		}
		limFileReader := io.LimitReader(tarReader, maxBytes)
		data, err := io.ReadAll(limFileReader)
		if err != nil {
			return returnData, err
		}
		if len(data) != 0 {
			returnData[filepath.Base(header.Name)] = data
		}
	}
	return returnData, nil
}
