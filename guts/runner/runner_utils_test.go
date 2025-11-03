package runner

import (
	"fmt"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestGetPartialGitData(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	var gitData TestGitData
	gitData.TestCase = "Firefox-Example-Basic"
	gitData.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	gitData.TestsRepoBranch = "main"

	rowId := 1

	accData, err := GetPartialGitData(rowId, Driver)
	utils.CheckError(err)

	if !reflect.DeepEqual(gitData, accData) {
		t.Errorf("unexpected git data!\nexpected: %v\nactual: %v", gitData, accData)
	}
}

func TestCloneTestsData(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 1

	var gitData TestGitData
	gitData.TestCase = "Firefox-Example-Basic"
	gitData.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	gitData.TestsRepoBranch = "main"

	accData, err := CloneTestsData(rowId, Driver)
	utils.CheckError(err)

	utils.FileOrDirExists(fmt.Sprintf("%v/.git", accData.RepoDir))
	utils.CheckError(err)

	os.RemoveAll(accData.RepoDir)
	utils.CheckError(err)

	if gitData.TestCase != accData.TestCase {
		t.Errorf("unexpected test case!\nexpected: %v\nactual: %v", gitData.TestCase, accData.TestCase)
	}
	if gitData.TestsRepo != accData.TestsRepo {
		t.Errorf("unexpected test case!\nexpected: %v\nactual: %v", gitData.TestsRepo, accData.TestsRepo)
	}
	if gitData.TestsRepoBranch != accData.TestsRepoBranch {
		t.Errorf("unexpected test case!\nexpected: %v\nactual: %v", gitData.TestsRepoBranch, accData.TestsRepoBranch)
	}
	if !regexp.MustCompile(`\b[0-9a-f]{40}\b`).MatchString(accData.CommitHash) {
		t.Errorf("seems like the commit has was parsed incorrectly!\ncommit hash: %v", accData.CommitHash)
	}
}

func TestCloneTestsDataNoExistRow(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 999

	var gitData TestGitData
	gitData.TestCase = "Firefox-Example-Basic"
	gitData.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	gitData.TestsRepoBranch = "main"

	_, err = CloneTestsData(rowId, Driver)
	if err == nil {
		t.Errorf("unexpected success in calling CloneTestsData for row id %v", rowId)
	}

	expectedErrString := "couldn't find any git data for row 999"
	if expectedErrString != err.Error() {
		t.Errorf("unexpected error string\nexpected: %v\nactual: %v", expectedErrString, err.Error())
	}
}

func TestCloneTestsDataBadGitUrl(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 84
	expectedErr := utils.GenericGitError{}

	_, err = CloneTestsData(rowId, Driver)
	if err == nil {
		t.Errorf("Git clone should have failed but didn't!")
	}

	errType := reflect.TypeOf(err)
	expectedType := reflect.TypeOf(expectedErr)
	if errType != expectedType {
		t.Errorf("unexpected error type!\nexpected: %v\nactual: %v", expectedType, errType)
	}
}

func TestFindJobForRunner(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	expectedRow := 65
	expectedUuid := "724254b8-5d51-42f5-8394-99976f87e520"

	accRow, accUuid, err := FindJobForRunner(Driver)
	utils.CheckError(err)

	if expectedRow != accRow {
		t.Errorf("unexpected row response!\nexpected: %v\nactual: %v", expectedRow, accRow)
	}
	if expectedUuid != accUuid {
		t.Errorf("unexpected uuid response!\nexpected: %v\nactual: %v", expectedUuid, accUuid)
	}
}

func TestSetCommitHashForTest(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 23
	commitHash := "1665bf0f817a795ffbc0a9f6d000def16ace544c"

	err = SetCommitHashForTest(rowId, commitHash, Driver)
	utils.CheckError(err)

	err = SetCommitHashForTest(rowId, "", Driver)
	utils.CheckError(err)
}

func TestSetResultsUrlForTest(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 23
	artifactUrl := "http://localhost:9999/my-file.txt"

	err = SetResultsUrlForTest(rowId, artifactUrl, Driver)
	utils.CheckError(err)

	err = SetResultsUrlForTest(rowId, "", Driver)
	utils.CheckError(err)
}

func TestGetPlanAndTestCase(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 1

	plan, testCase, err := GetPlanAndTestCase(rowId, Driver)
	utils.CheckError(err)

	expectedPlan := "tests/firefox-example/plans/extended.yaml"
	expectedTc := "Firefox-Example-Basic"

	if expectedPlan != plan {
		t.Errorf("unexpected plan!\nexpected: %v\nactual: %v", expectedPlan, plan)
	}
	if expectedTc != testCase {
		t.Errorf("unexpected test case!\nexpected: %v\nactual: %v", expectedTc, testCase)
	}
}

func TestGetYarfCommandLine(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	fullPlan := `---
tests:
  Firefox-Example-Basic:
    entrypoint: tests/firefox-example
  Firefox-Example-New-Tab:
    entrypoint: tests/firefox-example`

	dname, err := os.MkdirTemp("", "sampledir")
	utils.CheckError(err)

	planDir := fmt.Sprintf("%v/%v", dname, "tests/firefox-example/plans/")
	planName := "extended.yaml"

	err = os.MkdirAll(planDir, 0755)
	utils.CheckError(err)

	testPlanFn := fmt.Sprintf("%v%v", planDir, planName)
	err = os.WriteFile(testPlanFn, []byte(fullPlan), 0644)
	utils.CheckError(err)

	defer utils.DeferredErrCheckStringArg(os.RemoveAll, dname)

	var testData TestGitData
	testData.CommitHash = "98c3b02b214190f58b317c8e32c0f3a0efd926bc"
	testData.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	testData.TestsRepoBranch = "main"
	testData.RepoDir = dname

	rowId := 1

	yarfCmdLine, err := GetYarfCommandLine(testData, rowId, dname, Driver)
	utils.CheckError(err)

	cmdLineRegex := `\byarf --platform=Vnc tests\/firefox-example --outdir \/tmp\/sampledir[0-9]{1,10} -- --suite \"Firefox-Example-Basic\"`
	strCmdLine := strings.Join(yarfCmdLine, " ")
	if !regexp.MustCompile(cmdLineRegex).MatchString(strCmdLine) {
		t.Errorf("yarf cmd line didn't match expected regex!\nCommand line: %v\nregex: %v", yarfCmdLine, cmdLineRegex)
	}
}

func TestGetYarfCommandLineInvalidId(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	var testData TestGitData

	rowId := 999

	_, err = GetYarfCommandLine(testData, rowId, "", Driver)

	if err == nil {
		t.Errorf("Getting yarf command line succeeded where it should have failed!")
	}

	expectedErrString := "sql: no rows in result set"

	if err.Error() != expectedErrString {
		t.Errorf("unexpected error string!\nexpected: %v\nactual: %v", expectedErrString, err.Error())
	}
}

func TestGetYarfCommandLineBadPlan(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	fullPlan := `---
tests:
  asdf1:
    entrypoint: tests/firefox-example
  asdf2:
    entrypoint: tests/firefox-example`

	dname, err := os.MkdirTemp("", "sampledir")
	utils.CheckError(err)

	planDir := fmt.Sprintf("%v/%v", dname, "tests/firefox-example/plans/")
	planName := "extended.yaml"

	err = os.MkdirAll(planDir, 0755)
	utils.CheckError(err)

	testPlanFn := fmt.Sprintf("%v%v", planDir, planName)
	err = os.WriteFile(testPlanFn, []byte(fullPlan), 0644)
	utils.CheckError(err)

	defer utils.DeferredErrCheckStringArg(os.RemoveAll, dname)

	var testData TestGitData
	testData.CommitHash = "98c3b02b214190f58b317c8e32c0f3a0efd926bc"
	testData.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	testData.TestsRepoBranch = "main"
	testData.RepoDir = dname

	rowId := 1

	_, err = GetYarfCommandLine(testData, rowId, dname, Driver)
	if err == nil {
		t.Errorf("getting yarf command line didn't fail when it should have!")
	}

	expectedErrString := "couldn't parse test entrypoint for test Firefox-Example-Basic and plan {[{asdf1 {tests/firefox-example {false}}} {asdf2 {tests/firefox-example {false}}}]}"

	if err.Error() != expectedErrString {
		t.Errorf("unexpected error string!\nexpected: %v\nactual: %v", expectedErrString, err.Error())
	}
}

// func TestRemoveVncAddress(t *testing.T) {
//   Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
//   utils.CheckError(err)
//
//   rowId := 20
//   err := RemoveVncAddress()
//   // need to set this back straight after!
// }

func TestRemoveVncAddress(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 1

	host, port, err := GetHostAndPort(rowId, Driver)
	utils.CheckError(err)

	existingVncAddress := fmt.Sprintf("%v:%v", host, port)

	err = RemoveVncAddress(rowId, Driver)
	utils.CheckError(err)

	updateAddrQuery := fmt.Sprintf(`UPDATE tests SET vnc_address='%v' WHERE id='%v'`, existingVncAddress, rowId)
	err = Driver.UpdateRow(updateAddrQuery)
	utils.CheckError(err)
}

func TestGetHostAndPort(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

	rowId := 1

	host, port, err := GetHostAndPort(rowId, Driver)
	utils.CheckError(err)

	expectedHost := "127.0.0.1"
	if host != expectedHost {
		t.Errorf("unexpected host!\nexpected: %v\nactual: %v", expectedHost, host)
	}

	expectedPort := "5968"
	if port != expectedPort {
		t.Errorf("unexpected port!\nexpected: %v\nactual: %v", expectedPort, port)
	}
}
