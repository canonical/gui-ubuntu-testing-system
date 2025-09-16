package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

func ValidateUuid(Uuid string) error {
	if regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).MatchString(Uuid) {
		return nil
	}
	return InvalidUuidError{uuid: Uuid}
}

func FileOrDirExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%v doesn't exist", path)
		} else { // coverage-ignore
			return err
		}
	}
	return nil
}

func AllFilesExist(paths ...string) bool {
	for i := 0; i < len(paths); i++ {
		err := FileOrDirExists(paths[i])
		if err != nil {
			return false
		}
	}
	return true
}

func AtomicWrite(data []byte, filename string) error {
	newFile := fmt.Sprintf("%v.new", filename)
	err := os.WriteFile(newFile, data, 0644)
	if err != nil { // coverage-ignore
		return err
	}
	err = os.Rename(newFile, filename)
	if err != nil { // coverage-ignore
		return err
	}
	return nil
}

func CreateDirIfNotExists(directory string) error {
	// Creates a directory if it doesn't exist
	_, err := os.Open(directory)
	if err != nil {
		err = os.Mkdir(directory, 0755)
		return err
	}
	return nil
}

func GetDirSize(directory string) (int, error) {
	// Gets the total size in bytes of a directory
	DirSize = 0
	err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil { // coverage-ignore
				return err
			}
			DirSize += int(info.Size())
			return nil
		})
	if err != nil { // coverage-ignore
		return DirSize, err
	}
	return DirSize, err
}

func IsValidUrl(urlToTest string) bool {
	u, err := url.Parse(urlToTest)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func DownloadFile(url string) ([]byte, error) {
	var b []byte
	resp, err := http.Get(url)
	if err != nil {
		return b, err
	}
	defer DeferredErrCheck(resp.Body.Close)
	b, err = io.ReadAll(resp.Body)
	if err != nil { // coverage-ignore
		return b, err
	}
	if len(b) == 0 {
		return b, fmt.Errorf("file at %v is empty", url)
	}
	return b, nil
}

func Sha256sumOfString(inputString string) string {
	hasher := sha256.New()
	hasher.Write([]byte(inputString))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetStatusUrlForUuid(uuid string) string {
	statusUrl := fmt.Sprintf("%v%v/status/%v", GetProtocolPrefix(), GutsCfg.Api.Hostname, uuid)
	return statusUrl
}

func GetProtocolPrefix() string {
	switch GutsCfg.Api.Port {
	default:
		return ""
	case 8080:
		return "http://"
	case 443:
		return "https://"
	}
}
