package main

import (
	"fmt"
	"reflect"
	"testing"
)

func makeDummyJobReq() JobRequest {
	var expectedJobReq JobRequest
	url := "myurl"
	expectedJobReq.ArtifactUrl = &url
	expectedJobReq.TestsRepo = "myrepo"
	expectedJobReq.TestsRepoBranch = "main"
	expectedJobReq.TestsPlans = []string{"plan1", "plan2"}
	expectedJobReq.TestBed = "mytestbedurl"
	expectedJobReq.Debug = false
	expectedJobReq.Priority = 1
	expectedJobReq.Reporter = ""
	return expectedJobReq
}

func TestParseJobFromJsonSuccess(t *testing.T) {
	inputJson := `{"artifact_url": "myurl", "tests_repo": "myrepo", "tests_repo_branch": "main", "tests_plans": ["plan1", "plan2"], "testbed": "mytestbedurl", "debug": false, "priority": 1, "reporter": ""}`
	actualJobReq, err := ParseJobFromJson([]byte(inputJson))
	CheckError(err)
	expectedJobReq := makeDummyJobReq()
	if !reflect.DeepEqual(actualJobReq, expectedJobReq) {
		t.Errorf("Parsed job not same as expected!\nExpected: %v\nActual: %v", expectedJobReq, actualJobReq)
	}
}

func TestParseJobFromJsonFails(t *testing.T) {
	inputJson := `{"tests_repo": "myrepo", "tests_repo_branch": "main", "tests_plans": ["plan1", "plan2"], "testbed": "mytestbedurl", "debug": 0, "priority": 1, "reporter": ""}`
	_, err := ParseJobFromJson([]byte(inputJson))
	if err == nil {
		t.Errorf("Parsing job from:\n%v\nshould fail, but didn't!", inputJson)
	}
}

func TestGetAuthDataForKeySuccess(t *testing.T) {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	andersson123KeyPreSha := "4c126f75-c7d8-4a89-9370-f065e7ff4208"
	andersson123Key := Sha256sumOfString(andersson123KeyPreSha)
	var expectedTimData UserData
	expectedTimData.Username = "andersson123"
	expectedTimData.Key = "ba580bf88cfbc949f4894c85f65e65932872073105cb79d44caafa416452fbf2"
	expectedTimData.MaxPriority = 10
	timData, err := GetAuthDataForKey(andersson123Key)
	CheckError(err)
	if !reflect.DeepEqual(expectedTimData, timData) {
		t.Errorf("Expected userdata not the same as actual:\nExpected: %v\nActual: %v", expectedTimData, timData)
	}
}

func TestGetAuthDataForKeyUnknownUser(t *testing.T) {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}
	farnsworthKeyPreSha := "good-news-everyone"
	farnsworthKey := Sha256sumOfString(farnsworthKeyPreSha)
	_, err = GetAuthDataForKey(farnsworthKey)
	expectedErrString := fmt.Sprintf("key %v doesn't exist", farnsworthKey)
	if !reflect.DeepEqual(err.Error(), expectedErrString) {
		t.Errorf("Unexpected error: %v\nExpected error: %v", err.Error(), expectedErrString)
	}
}

func TestAuthorizeUserAndAssignPriorityReqUnderMaxPrio(t *testing.T) {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}

	andersson123KeyPreSha := "4c126f75-c7d8-4a89-9370-f065e7ff4208"
	andersson123Key := Sha256sumOfString(andersson123KeyPreSha)

	dummyJobReq := makeDummyJobReq()

	_, alteredJobReq, err := AuthorizeUserAndAssignPriority(andersson123Key, dummyJobReq)
	CheckError(err)
	if !reflect.DeepEqual(alteredJobReq, dummyJobReq) {
		t.Errorf("Job request unintentionally altered!\nIntended: %v\nActual: %v", dummyJobReq, alteredJobReq)
	}
}

func TestAuthorizeUserAndAssignPriorityReqMaxPrio(t *testing.T) {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}

	andersson123KeyPreSha := "4c126f75-c7d8-4a89-9370-f065e7ff4208"
	andersson123Key := Sha256sumOfString(andersson123KeyPreSha)

	dummyJobReq := makeDummyJobReq()
	dummyJobReq.Priority = 10

	_, alteredJobReq, err := AuthorizeUserAndAssignPriority(andersson123Key, dummyJobReq)
	CheckError(err)
	if !reflect.DeepEqual(alteredJobReq, dummyJobReq) {
		t.Errorf("Job request unintentionally altered!\nIntended: %v\nActual: %v", dummyJobReq, alteredJobReq)
	}
}

func TestAuthorizeUserAndAssignPriorityReqOverMaxPrio(t *testing.T) {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}

	andersson123KeyPreSha := "4c126f75-c7d8-4a89-9370-f065e7ff4208"
	andersson123Key := Sha256sumOfString(andersson123KeyPreSha)

	dummyJobReq := makeDummyJobReq()
	dummyJobReq.Priority = 11

	timData, alteredJobReq, err := AuthorizeUserAndAssignPriority(andersson123Key, dummyJobReq)
	CheckError(err)
	if alteredJobReq.Priority != timData.MaxPriority {
		t.Errorf("Request priority is %v and should have been reduced to %v", alteredJobReq.Priority, timData.MaxPriority)
	}
}

func TestValidateArtifactUrlDeb(t *testing.T) {
	// serve a deb
	ServeDirectory()
	// create the url
	testUrl := "http://localhost:9999/hello_2.10-3build1_amd64.deb"
	// validate the url
	err := ValidateArtifactUrl(testUrl)
	CheckError(err)
}

func TestValidateArtifactUrlSnap(t *testing.T) {
	// serve a snap
	ServeDirectory()
	// create the url
	testUrl := "http://localhost:9999/hello_42.snap"
	// validate the url
	err := ValidateArtifactUrl(testUrl)
	CheckError(err)
}

func TestValidateArtifactUrlInvalidArtifactType(t *testing.T) {
	// serve a snap
	ServeDirectory()
	// create the url
	testUrl := "http://localhost:9999/hello_42.rpm"
	// validate the url
	err := ValidateArtifactUrl(testUrl)
	if err == nil {
		t.Errorf("Validating %v threw no error when it should have!", testUrl)
	}
}

func TestValidateArtifactUrlNonexistentUrl(t *testing.T) {
	// serve a snap
	ServeDirectory()
	// create the url
	testUrl := "http://localhost:9999/no-exist.deb"
	// validate the url
	err := ValidateArtifactUrl(testUrl)
	if err == nil {
		t.Errorf("Validating %v threw no error when it should have!", testUrl)
	}
}

func TestValidateArtifactUrlUnacceptableDomain(t *testing.T) {
	// serve a snap
	ServeDirectory()
	// create the url
	testUrl := "http://farnsworth:9999/no-exist.deb"
	// validate the url
	err := ValidateArtifactUrl(testUrl)
	if err == nil {
		t.Errorf("Validating %v threw no error when it should have!", testUrl)
	}
}

func TestValidateTestbedUrlIso(t *testing.T) {
	// serve an iso
	ServeDirectory()
	// create the url
	testUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	// validate the url
	err := ValidateTestbedUrl(testUrl)
	CheckError(err)
}

func TestValidateTestbedUrlImg(t *testing.T) {
	// serve an iso
	ServeDirectory()
	// create the url
	testUrl := "http://localhost:9999/testimg.img"
	// validate the url
	err := ValidateTestbedUrl(testUrl)
	CheckError(err)
}

func TestValidateTestData(t *testing.T) {
	branch := "main"
	repo := "https://github.com/canonical/ubuntu-gui-testing.git"
	plans := []string{
		"tests/firefox-example/plans/regular.yaml",
		"tests/firefox-example/plans/extended.yaml",
	}
	err := ValidateTestData(branch, repo, plans)
	CheckError(err)
}

func TestValidateTestDataBadRemote(t *testing.T) {
	branch := "main"
	repo := "https://github.com/farnsworth-gui-testing.git"
	plans := []string{
		"tests/firefox-example/plans/regular.yaml",
		"tests/firefox-example/plans/extended.yaml",
	}
	err := ValidateTestData(branch, repo, plans)
	if err == nil {
		t.Errorf("Something is very wrong - %v was incorrectly identified as a functional remote", repo)
	}
}

func TestValidateTestDataBadBranch(t *testing.T) {
	branch := "farnsworth"
	repo := "https://github.com/canonical/ubuntu-gui-testing.git"
	plans := []string{
		"tests/firefox-example/plans/regular.yaml",
		"tests/firefox-example/plans/extended.yaml",
	}
	err := ValidateTestData(branch, repo, plans)
	if err == nil {
		t.Errorf("Something is very wrong - %v was incorrectly identified as an existing branch", branch)
	}
}

func TestValidateTestDataBadPlans(t *testing.T) {
	branch := "main"
	repo := "https://github.com/canonical/ubuntu-gui-testing.git"
	plans := []string{
		"tests/firefox-example/plans/farnsworth.yaml",
		"tests/firefox-example/plans/leela.yaml",
	}
	err := ValidateTestData(branch, repo, plans)
	if err == nil {
		t.Errorf("Something is very wrong - %v were incorrectly identified as existing plans", plans)
	}
}

func TestWriteJobEntryToDbSucceeds(t *testing.T) {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	err = Setup()
	if SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		CheckError(err)
	}

	andersson123KeyPreSha := "4c126f75-c7d8-4a89-9370-f065e7ff4208"
	andersson123Key := Sha256sumOfString(andersson123KeyPreSha)
	timData, err := GetAuthDataForKey(andersson123Key)
	CheckError(err)

	dummyJobReq := makeDummyJobReq()
	jobEntry := CreateJobEntry(dummyJobReq, timData)

	err = WriteJobEntryToDb(jobEntry)
	CheckError(err)
}
