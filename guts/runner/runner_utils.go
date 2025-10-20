package runner

import (
	"database/sql"
	"fmt"
	"gopkg.in/yaml.v3"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/storage"
	"guts.ubuntu.com/v2/utils"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TestGitData struct {
	TestCase        string
	CommitHash      string
	TestsRepo       string
	TestsRepoBranch string
	RepoDir         string
}

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

func GetPartialGitData(rowId int, Driver database.DbDriver) (TestGitData, error) {
	var testGitData TestGitData

	testQuery := fmt.Sprintf(`SELECT tests.test_case, jobs.tests_repo, jobs.tests_repo_branch FROM tests JOIN jobs ON jobs.uuid=tests.uuid WHERE tests.id=%v`, rowId)
	row, err := Driver.RunQueryRow(testQuery)

	if err != nil { // coverage-ignore
		return testGitData, err
	}
	err = row.Scan(
		&testGitData.TestCase,
		&testGitData.TestsRepo,
		&testGitData.TestsRepoBranch,
	)

	if err != nil { // coverage-ignore
		if err == sql.ErrNoRows {
			return testGitData, nil
		} else {
			return testGitData, err
		}
	}

	return testGitData, nil
}

func CloneTestsData(rowId int, Driver database.DbDriver) (TestGitData, error) {
	testData, err := GetPartialGitData(rowId, Driver)
	if err != nil { // coverage-ignore
		return testData, err
	}
	if testData == (TestGitData{}) {
		return testData, fmt.Errorf("couldn't find any git data for row %v", rowId)
	}
	// create temp dir
	cloneDirName, err := os.MkdirTemp("", "gitrepo")
	if err != nil { // coverage-ignore
		return testData, err
	}
	// clone the repository to a directory
	cloneCmd := exec.Command(
		"git",
		"clone",
		"--branch",
		testData.TestsRepoBranch,
		testData.TestsRepo,
		cloneDirName,
	)
	if err := cloneCmd.Run(); err != nil {
		err = os.RemoveAll(cloneDirName)
		if err != nil { // coverage-ignore
			return testData, err
		}
		return testData, utils.GenericGitError{Command: cloneCmd.Args}
	}

	// get the commit hash
	gitHash, err := exec.Command(
		"git",
		"log",
		"-1",
		`--pretty=format:%H`,
	).Output()
	if err != nil { // coverage-ignore
		err = os.RemoveAll(cloneDirName)
		return testData, err
	}

	testData.CommitHash = string(gitHash)
	testData.RepoDir = cloneDirName
	return testData, nil
}

func FindJobForRunner(Driver database.DbDriver) (int, string, error) {
	var id int
	var Uuid string

	testQuery := `SELECT tests.id, tests.uuid FROM tests JOIN jobs ON jobs.uuid=tests.uuid WHERE state='spawned' AND vnc_address!='' ORDER BY priority DESC LIMIT 1`
	row, err := Driver.RunQueryRow(testQuery)
	if err != nil { // coverage-ignore
		return id, Uuid, err
	}
	err = row.Scan(
		&id,
		&Uuid,
	)

	if err != nil { // coverage-ignore
		if err == sql.ErrNoRows {
			return id, Uuid, nil
		} else {
			return id, Uuid, err
		}
	}

	return id, Uuid, nil
}

func SetCommitHashForTest(id int, hash string, Driver database.DbDriver) error {
	updateQuery := fmt.Sprintf(`UPDATE tests SET commit_hash='%v' WHERE id=%v`, hash, id)
	err := Driver.UpdateRow(updateQuery)
	return err
}

func SetResultsUrlForTest(id int, resultsUrl string, Driver database.DbDriver) error {
	updateQuery := fmt.Sprintf(`UPDATE tests SET results_url='%v' WHERE id=%v`, resultsUrl, id)
	err := Driver.UpdateRow(updateQuery)
	return err
}

func GetPlanAndTestCase(rowId int, Driver database.DbDriver) (string, string, error) {
	plan := ""
	testCase := ""

	testQuery := fmt.Sprintf(`SELECT plan, test_case FROM tests WHERE id=%v`, rowId)
	row, err := Driver.RunQueryRow(testQuery)
	if err != nil { // coverage-ignore
		return plan, testCase, err
	}
	err = row.Scan(
		&plan,
		&testCase,
	)

	if err != nil { // coverage-ignore
		return plan, testCase, err
	}

	return plan, testCase, nil
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

func GetYarfCommandLine(TestData TestGitData, rowId int, artifactsDir string, Driver database.DbDriver) ([]string, error) {
	var cmdLine []string
	plan, testCase, err := GetPlanAndTestCase(rowId, Driver)

	if err != nil {
		return cmdLine, err
	}

	fullPlanPath := fmt.Sprintf("%v/%v", TestData.RepoDir, plan)

	testPlan, err := ParsePlan(fullPlanPath)

	entrypoint := ""
	for _, entry := range testPlan.Tests {
		if entry.Name == testCase {
			entrypoint = entry.Data.EntryPoint
		}
	}

	if entrypoint == "" {
		return cmdLine, fmt.Errorf("couldn't parse test entrypoint for test %v and plan %v", testCase, testPlan)
	}

	cmdLine = []string{
		"yarf",
		"--platform=Vnc",
		entrypoint,
		"--outdir",
		artifactsDir,
		"--",
		"--suite",
		fmt.Sprintf(`"%v"`, testCase),
	}
	return cmdLine, nil
}

func RemoveVncAddress(id int, Driver database.DbDriver) error {
	updateQuery := fmt.Sprintf(`UPDATE tests SET vnc_address='' WHERE id='%v'`, id)
	err := Driver.UpdateRow(updateQuery)
	return err
}

func GetHostAndPort(id int, Driver database.DbDriver) (string, string, error) {
	vncAddress := ""
	row, err := Driver.QueryRow("tests", "id", fmt.Sprintf("%v", id), []string{"vnc_address"})
	if err != nil { // coverage-ignore
		return "", "", err
	}
	err = row.Scan(
		&vncAddress,
	)
	if err != nil { // coverage-ignore
		if err == sql.ErrNoRows {
			return "", "", nil
		} else {
			return "", "", err
		}
	}
	splitAddr := strings.Split(vncAddress, ":")
	// This is overkill checking because the rows are written deterministically
	// But might as well check
	if len(splitAddr) != 2 { // coverage-ignore
		return "", "", fmt.Errorf("vnc address %v doesn't conform to expected syntax", vncAddress)
	}
	return splitAddr[0], splitAddr[1], nil
}

// don't bother testing the main loop, that's for integration testing
func RunnerLoop(Driver database.DbDriver, RunnerCfg GutsRunnerConfig) error { // coverage-ignore
	// ensure we have a functional storage backend
	backend, err := storage.GetStorageBackend(RunnerCfg.Storage)
	if err != nil {
		return err
	}

	//.get row id and uuid
	rowId, Uuid, err := FindJobForRunner(Driver)
	if err != nil {
		return err
	}

	// - set state to `running`
	err = Driver.SetTestStateTo(rowId, "running")
	if err != nil {
		return err
	}

	// - clone the tests repo
	GitData, err := CloneTestsData(rowId, Driver)
	if err != nil {
		return err
	}

	// - update the `commit_hash` column
	err = SetCommitHashForTest(rowId, GitData.CommitHash, Driver)
	if err != nil {
		return err
	}

	// create temp dir for artifacts
	artifactDirName, err := os.MkdirTemp("", "artifacts")
	if err != nil {
		return err
	}

	// create yarf command line
	yarfCmdLine, err := GetYarfCommandLine(GitData, rowId, artifactDirName, Driver)
	if err != nil {
		return err
	}

	host, port, err := GetHostAndPort(rowId, Driver)
	if err != nil {
		return err
	}

	envVars := []string{
		fmt.Sprintf("VNC_HOST=%v", host),
		fmt.Sprintf("VNC_PORT=%v", port),
	}
	yarfProcess, err := utils.StartProcess(yarfCmdLine, &envVars)
	if err != nil {
		return err
	}

	yarfTempFailCode := 999
	heartbeatDuration := time.Second * 5

	for !yarfProcess.ProcessState.Exited() {
		err = database.UpdateUpdatedAt(rowId, Driver)
		if err != nil {
			return err
		}
		time.Sleep(heartbeatDuration)
	}

	// Test must have now completed.
	exitCode := yarfProcess.ProcessState.ExitCode()
	if exitCode == yarfTempFailCode {
		// this means that the test run was a tempfail
		// here, unset the vnc_address and set the state back to requested
		// doing this means the test will be retried
		err = RemoveVncAddress(rowId, Driver)
		if err != nil {
			return err
		}
		err = Driver.SetTestStateTo(rowId, "requested")
		if err != nil {
			return err
		}
		return nil
	}

	// Bundle up test artifacts and result - which is artifactDirName
	tarBytes, err := utils.TarUpDirectory(artifactDirName)
	if err != nil {
		return err
	}

	// gzip the tarBytes
	gzippedTarBytes, err := utils.GzipTarArchiveBytes(tarBytes)
	if err != nil {
		return err
	}

	// upload the test artifacts to the storage backend
	storageUrl, err := backend.Upload(Uuid, fmt.Sprintf("%v-%v.tar.gz", Uuid, rowId), gzippedTarBytes)
	if err != nil {
		return err
	}

	// write artifact_url to tests table
	err = SetResultsUrlForTest(rowId, storageUrl, Driver)
	if err != nil {
		return err
	}

	// set state string
	finalState := "pass"
	if exitCode != 0 {
		finalState = "fail"
	}

	// update test state
	err = Driver.SetTestStateTo(rowId, finalState)
	if err != nil {
		return err
	}

	return nil
}
