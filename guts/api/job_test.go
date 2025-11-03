package api

import (
	"fmt"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"reflect"
	"testing"
	"time"
)

func TestGetCompleteResultsForUuidFailure(t *testing.T) {
	Uuid := "21a57878-3307-449c-9f71-9f3f5d11f41c"
	_, Driver, _, err := Setup()
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	_, err = GetCompleteResultsForUuid(Uuid, Driver)
	expectedErrString := fmt.Sprintf("No jobs with uuid %v found!", Uuid)
	if err == nil {
		t.Errorf("Expected failure for uuid %v", Uuid)
	}
	if err.Error() != expectedErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", expectedErrString, err.Error())
	}
}

func TestGetCompleteResultsForUuidSuccess(t *testing.T) {
	Uuid := "4ce9189f-561a-4886-aeef-1836f28b073b"
	_, Driver, _, err := Setup()
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}

	servingProcess := utils.ServeRelativeDirectory("/../../postgres/test-data/test-files/")
	defer utils.DeferredErrCheck(servingProcess.Kill)

	job, err := GetCompleteResultsForUuid(Uuid, Driver)
	utils.CheckError(err)
	var expectedJob JobWithTestsDetails
	var TestJob JobEntry
	TestJob.Uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
	TestJob.ArtifactUrl = nil
	TestJob.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	TestJob.TestsRepoBranch = "main"
	TestJob.TestsPlans = []string{"tests/firefox-example/plans/extended.yaml", "tests/firefox-example/plans/regular.yaml"}
	TestJob.ImageUrl = "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	TestJob.Reporter = "test_observer"
	TestJob.Status = "running"
	layout := "2006-01-02T15:04:05.999999Z"
	value := "2025-07-23T14:17:14.632177Z"
	parsedTime, err := time.Parse(layout, value)
	utils.CheckError(err)
	TestJob.SubmittedAt = parsedTime
	TestJob.Requester = "andersson123"
	TestJob.Debug = false
	TestJob.Priority = 11
	expectedJob.Job = TestJob
	expectedJob.Results = make(map[string]string)
	// what?
	// expectedJob.Results["Firefox-Example-Basic"] = "running"
	expectedJob.Results["Firefox-Example-Basic"] = "requested"
	expectedJob.Results["Firefox-Example-New-Tab"] = "spawning"
	if !reflect.DeepEqual(job, expectedJob) {
		t.Errorf("expected job not the same as actual\nexpected: %v\nactual: %v", expectedJob, job)
	}
}

func TestFindJobByUuid(t *testing.T) {
	Uuid := "4ce9189f-561a-4886-aeef-1836f28b073b"
	_, Driver, _, err := Setup()
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	job, err := FindJobByUuid(Uuid, Driver)
	utils.CheckError(err)
	var TestJob JobEntry
	TestJob.Uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
	TestJob.ArtifactUrl = nil
	TestJob.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	TestJob.TestsRepoBranch = "main"
	TestJob.TestsPlans = []string{"tests/firefox-example/plans/extended.yaml", "tests/firefox-example/plans/regular.yaml"}
	TestJob.ImageUrl = "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	TestJob.Reporter = "test_observer"
	TestJob.Status = "running"
	layout := "2006-01-02T15:04:05.999999"
	value := "2025-07-23T14:17:14.632177"
	TestJob.SubmittedAt, err = time.Parse(layout, value)
	utils.CheckError(err)
	TestJob.Requester = "andersson123"
	TestJob.Debug = false
	TestJob.Priority = 11
	if !reflect.DeepEqual(job, TestJob) {
		t.Errorf("Expected job not the same as actual:\n%v\n%v", TestJob, job)
	}
}

func TestJobToJson(t *testing.T) {
	var TestJob JobEntry
	TestJob.Uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
	TestJob.ArtifactUrl = nil
	TestJob.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	TestJob.TestsRepoBranch = "main"
	TestJob.TestsPlans = []string{"tests/firefox-example/plans/extended.yaml", "tests/firefox-example/plans/regular.yaml"}
	TestJob.ImageUrl = "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	TestJob.Reporter = "test_observer"
	TestJob.Status = "running"
	layout := "2006-01-02T15:04:05.999999Z"
	value := "2025-07-23T14:17:14.632177Z"
	parsedTime, err := time.Parse(layout, value)
	utils.CheckError(err)
	TestJob.SubmittedAt = parsedTime
	TestJob.Requester = "andersson123"
	TestJob.Debug = false
	TestJob.Priority = 8
	ExpectedJson := `{"uuid":"4ce9189f-561a-4886-aeef-1836f28b073b","artifact_url":null,"tests_repo":"https://github.com/canonical/ubuntu-gui-testing.git","tests_repo_branch":"main","tests_plans":["tests/firefox-example/plans/extended.yaml","tests/firefox-example/plans/regular.yaml"],"image_url":"https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso","reporter":"test_observer","status":"running","submitted_at":"2025-07-23T14:17:14.632177Z","requester":"andersson123","debug":false,"priority":8}`
	ConvertedJson := TestJob.ToJson()
	if !reflect.DeepEqual(ExpectedJson, ConvertedJson) {
		t.Errorf("json conversion not as expected!\nExpected: %v\nActual: %v", ExpectedJson, ConvertedJson)
	}
}

func TestJobWithTestsDetailsToJson(t *testing.T) {
	var jobwDetails JobWithTestsDetails
	var TestJob JobEntry
	TestJob.Uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
	TestJob.ArtifactUrl = nil
	TestJob.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	TestJob.TestsRepoBranch = "main"
	TestJob.TestsPlans = []string{"tests/firefox-example/plans/extended.yaml", "tests/firefox-example/plans/regular.yaml"}
	TestJob.ImageUrl = "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	TestJob.Reporter = "test_observer"
	TestJob.Status = "running"
	layout := "2006-01-02T15:04:05.999999Z"
	value := "2025-07-23T14:17:14.632177Z"
	parsedTime, err := time.Parse(layout, value)
	utils.CheckError(err)
	TestJob.SubmittedAt = parsedTime
	TestJob.Requester = "andersson123"
	TestJob.Debug = false
	TestJob.Priority = 8
	jobwDetails.Job = TestJob
	expectedJson := `{"Job":{"uuid":"4ce9189f-561a-4886-aeef-1836f28b073b","artifact_url":null,"tests_repo":"https://github.com/canonical/ubuntu-gui-testing.git","tests_repo_branch":"main","tests_plans":["tests/firefox-example/plans/extended.yaml","tests/firefox-example/plans/regular.yaml"],"image_url":"https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso","reporter":"test_observer","status":"running","submitted_at":"2025-07-23T14:17:14.632177Z","requester":"andersson123","debug":false,"priority":8},"results":null}`
	convertedJson := jobwDetails.ToJson()
	if !reflect.DeepEqual(expectedJson, convertedJson) {
		t.Errorf("expected json not same as actual\nexpected: %v\nactual: %v", expectedJson, convertedJson)
	}
}
