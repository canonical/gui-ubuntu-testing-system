package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TestCaseData struct {
	EntryPoint   string
	Requirements struct {
		Tpm bool
	}
}

type TestCase struct {
	Name string
	Data TestCaseData
}

type TestCases []TestCase

type TestPlan struct {
	Tests TestCases
}

type GenericGitError struct {
	Command []string
}

func (g GenericGitError) Error() string {
	return fmt.Sprintf("Git operation failed:\n%v", g.Command)
}

type InvalidUuidError struct {
	uuid string
}

func (e InvalidUuidError) Error() string {
	return fmt.Sprintf("%v isn't a valid uuid!", e.uuid)
}

func CheckError(err error) { // coverage-ignore
	if err != nil {
		panic(err)
		// log.Fatal(err.Error())
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

func PidActive(pid int) bool { // coverage-ignore
  _, err := os.FindProcess(pid)
  log.Printf("checking status of pid %v", pid)
  if err != nil {
    log.Printf(err.Error())
    return false
  }
  return true
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
		err = os.MkdirAll(directory, 0755)
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

func Md5SumOfBytes(inputBytes []byte) string {
	hasher := md5.New()
	hasher.Write(inputBytes)
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

func GetFileNameFromUrl(url string) string {
	splitUrl := strings.Split(url, "/")
	fileName := splitUrl[len(splitUrl)-1]
	return fileName
}

func ServeRelativeDirectory(relativeDir string) *os.Process { // coverage-ignore
	pwd, err := os.Getwd()
	CheckError(err)
	testFilesDir := pwd + relativeDir
	process := ServeDirectory(testFilesDir)
	return process
}

// this function is just used for testing, so we don't test it
func ServeDirectory(testFilesDir string) *os.Process { // coverage-ignore
	port := "9999"

	serveCmd := exec.Command("php", "-S", "localhost:"+port)
	serveCmd.Dir = testFilesDir

	err := serveCmd.Start()
	CheckError(err)

	for i := 0; i < 300; i++ {
		// I think this long timeout was leftover from debugging.
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), timeout)
		if err != nil {
			time.Sleep(time.Millisecond * 250)
		} else {
			if conn != nil {
				err := conn.Close()
				CheckError(err)
				time.Sleep(time.Millisecond * 250)
				return serveCmd.Process
			}
		}
	}

	CheckError(fmt.Errorf("Port never came up when trying to serve directory with command:\n%v", serveCmd))
	return serveCmd.Process
}

func WriteToTar(path string, tw *tar.Writer, fi os.FileInfo) error {
	// Open the path
	fr, err := os.Open(path)
	if err != nil { // coverage-ignore
		return err
	}
	defer DeferredErrCheck(fr.Close)

	copyData := true

	h := new(tar.Header)
	h.Name = path
	h.Size = fi.Size()
	h.Mode = int64(fi.Mode())
	h.ModTime = fi.ModTime()
	h.Typeflag = tar.TypeReg
	if fi.IsDir() {
		h.Typeflag = tar.TypeDir
		copyData = false
	}
	err = tw.WriteHeader(h)
	if err != nil { // coverage-ignore
		return err
	}

	if copyData {
		_, err = io.Copy(tw, fr)
		if err != nil { // coverage-ignore
			return err
		}
	}
	return nil
}

func TraverseDirectory(dirPath string, tw *tar.Writer) error {
	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil { // coverage-ignore
		return err
	}

	// read all the files/dir in it
	fis, err := dir.Readdir(0)
	if err != nil { // coverage-ignore
		return err
	}

	DeferredErrCheck(dir.Close)

	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()
		err = WriteToTar(curPath, tw, fi)
		// typically, we wouldn't ignore this err handling,
		// but WriteToTar only fails due to golang std
		// library calls failing, so it's idiomatic
		// to not test these kinds of errors in this case.
		if err != nil { // coverage-ignore
			return err
		}
		if fi.IsDir() {
			err = TraverseDirectory(curPath, tw)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func TarUpDirectory(dirToTar string) ([]byte, error) {
	var tarData []byte
	tarBuffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(tarBuffer)
	err := TraverseDirectory(dirToTar, tarWriter)

	// close tar writer
	err = tarWriter.Close()
	if err != nil { // coverage-ignore
		return tarData, err
	}

	tarData = tarBuffer.Bytes()
	return tarData, err
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

func EnsureGetEnv(envVar string) (string, error) {
	envValue := os.Getenv(envVar)
	if envValue == "" { // coverage-ignore
		return envValue, fmt.Errorf("tried to get %v environment variable, but it was empty or unset", envVar)
	}
	return envValue, nil
}

func StartProcess(processArgs []string, envVars *[]string) (*exec.Cmd, error) {
	cmd := exec.Command(processArgs[0], processArgs[1:]...)
	cmd.Env = os.Environ()
	if envVars != nil {
		for _, entry := range *envVars {
			cmd.Env = append(cmd.Env, entry)
		}
	}
  log.Printf("running command:\n%v", cmd)
	err := cmd.Start()
	return cmd, err
}

func GitCloneToDir(repository, branch, directory string) error {
	// leave directory as empty to just clone with repo name to current directory
	cloneCmd := exec.Command(
		"git",
		"clone",
		"--branch",
		branch,
		repository,
		directory,
	)
	if err := cloneCmd.Run(); err != nil { // coverage-ignore
		if directory != "" {
			err = os.RemoveAll(directory)
			if err != nil { // coverage-ignore
				return err
			}
		}
		return GenericGitError{Command: cloneCmd.Args}
	}
	return nil
}

func (p *TestCases) UnmarshalYAML(value *yaml.Node) error { // coverage-ignore
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("`tests` must contain YAML mapping, has %v", value.Kind)
	}
	*p = make([]TestCase, len(value.Content)/2)
	for i := 0; i < len(value.Content); i += 2 {
		var res = &(*p)[i/2]
		if err := value.Content[i].Decode(&res.Name); err != nil {
			return err
		}
		if err := value.Content[i+1].Decode(&res.Data); err != nil {
			return err
		}
	}
	return nil
}

func ParsePlan(planPath string) (TestPlan, error) {
	var testPlan TestPlan
	dat, err := os.ReadFile(planPath)
	if err != nil { // coverage-ignore
		return testPlan, err
	}

	err = yaml.Unmarshal(dat, &testPlan)
	if err != nil {
		return testPlan, err
	}
	return testPlan, nil
}
