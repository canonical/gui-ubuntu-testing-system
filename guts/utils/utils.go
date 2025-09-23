package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

type InvalidUuidError struct {
	uuid string
}

func (e InvalidUuidError) Error() string {
	return fmt.Sprintf("%v isn't a valid uuid!", e.uuid)
}

func CheckError(err error) { // coverage-ignore
	if err != nil {
		log.Fatal(err.Error())
	}
}

func DeferredErrCheck(f func() error) { // coverage-ignore
	err := f()
	CheckError(err)
}

func DeferredErrCheckStringArg(f func(s string) error, s string) { // coverage-ignore
	err := f(s)
	CheckError(err)
}

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

func AtomicWrite(data []byte, filename string) error { // coverage-ignore
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

func GetDirSize(directory string) (int, error) { // coverage-ignore
	// Gets the total size in bytes of a directory
	DirSize := 0
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
	if err != nil { // coverage-ignore
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

func GetProtocolPrefix(port int) string {
	switch port {
	default:
		return ""
	case 8080:
		return "http://"
	case 443:
		return "https://"
	}
}

// this function is just used for testing, so we don't test it
func ServeDirectory(relativeDirToServe string) { // coverage-ignore
	pwd, err := os.Getwd()
	CheckError(err)
	port := "9999"
	testFilesDir := pwd + relativeDirToServe
	serveCmd := exec.Command("php", "-S", "localhost:"+port)
	serveCmd.Dir = testFilesDir
	go serveCmd.Run() //nolint:all
	for i := 0; i < 60; i++ {
		timeout := time.Second * 5
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), timeout)
		if err != nil {
			time.Sleep(timeout)
		} else {
			if conn != nil {
				err := conn.Close()
				CheckError(err)
				return
			}
		}
	}
	CheckError(fmt.Errorf("Port never came up when trying to serve directory with command:\n%v", serveCmd))
}
