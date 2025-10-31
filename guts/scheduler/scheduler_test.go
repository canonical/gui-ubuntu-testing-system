package scheduler

import (
	"fmt"
	"guts.ubuntu.com/v2/api"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/storage"
	"guts.ubuntu.com/v2/utils"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestGetNewJobsUuids(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	uuids, err := GetNewJobsUuids(Driver)
	utils.CheckError(err)

	expectedUuids := []string{
		"035a731b-9138-47d3-9f03-d7647186c7a1",
		"1e1de355-0805-470a-9594-9d55d059cfc3",
		"25f93036-17b2-417a-a123-1f25b79afe33",
		"43d8b125-802c-4398-b9f3-7b977d5f2251",
		"60cf4a1f-a26b-461c-a970-71fc14402904",
		"8a4f22aa-fb17-485a-933d-b426eeba0735",
		"a2212936-3e04-486c-9a05-47a60c80971f",
		"a5f0dd5a-46fa-4a24-8244-7ddd849d91a4",
		"daaf4391-4496-4ea8-922f-9ce93af8c851",
	}

	slices.Sort(expectedUuids)
	slices.Sort(uuids)

	if !reflect.DeepEqual(uuids, expectedUuids) {
		t.Errorf("uuids not as expected!\nexpected: %v\nactual: %v", expectedUuids, uuids)
	}
}

func TestGetTestData(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	testUuid := "4ce9189f-561a-4886-aeef-1836f28b073b"

	repo, branch, plans, err := GetTestData(Driver, testUuid)
	utils.CheckError(err)

	expectedRepo := "https://github.com/canonical/ubuntu-gui-testing.git"
	expectedBranch := "main"
	expectedPlans := []string{
		"tests/firefox-example/plans/extended.yaml",
		"tests/firefox-example/plans/regular.yaml",
	}

	if repo != expectedRepo {
		t.Errorf("unexpected repo!\nexpected: %v\nactual: %v", expectedRepo, repo)
	}
	if branch != expectedBranch {
		t.Errorf("unexpected branch!\nexpected: %v\nactual: %v", expectedBranch, branch)
	}
	if !reflect.DeepEqual(plans, expectedPlans) {
		t.Errorf("unexpected plans!\nexpected: %v\nactual: %v", expectedPlans, plans)
	}
}

func TestWriteTestsForJob(t *testing.T) {
	// create a job with some stuff from the API
	Driver, err := database.TestDbDriver("guts_api", "guts_api")
	utils.CheckError(err)

	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}

	andersson123KeyPreSha := "4c126f75-c7d8-4a89-9370-f065e7ff4208"
	andersson123Key := utils.Sha256sumOfString(andersson123KeyPreSha)
	timData, err := api.GetAuthDataForKey(andersson123Key, Driver)
	utils.CheckError(err)

	dummyJobReq := api.MakeDummyJobReq()
	dummyJobReq.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	dummyJobReq.TestsRepoBranch = "main"
	dummyJobReq.TestsPlans = []string{"tests/firefox-example/plans/regular.yaml", "tests/firefox-example/plans/extended.yaml"}
	jobEntry := api.CreateJobEntry(dummyJobReq, timData)

	err = api.WriteJobEntryToDb(jobEntry, Driver)
	utils.CheckError(err)

	// init scheduler driver
	Driver, err = database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	// test the WriteTestsForJob function
	err = WriteTestsForJob(Driver, jobEntry.Uuid)
	utils.CheckError(err)

	// nuke the uuid
	err = Driver.NukeUuid(jobEntry.Uuid)
	utils.CheckError(err)
}

func TestGetUpdatedJobState(t *testing.T) {
	// init scheduler driver
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	// First, let's check for a pass/pass/pass
	Uuid := "505af468-13b4-405f-a384-273a31c60e6a"
	state, err := GetUpdatedJobState(Driver, Uuid)
	utils.CheckError(err)
	expectedState := "pass"
	if expectedState != state {
		t.Errorf("unexpected state output!\nexpected: %v\nactual: %v", expectedState, state)
	}

	// Now, let's check for a pass/pass/fail
	Uuid = "a052e18b-4c33-42b3-aa29-e5e0f2c4ad43"
	state, err = GetUpdatedJobState(Driver, Uuid)
	utils.CheckError(err)
	expectedState = "fail"
	if expectedState != state {
		t.Errorf("unexpected state output!\nexpected: %v\nactual: %v", expectedState, state)
	}

	// Now, let's check for a pass/pass/running
	Uuid = "b28d5289-0b2b-4aa0-996b-cfa07101035b"
	state, err = GetUpdatedJobState(Driver, Uuid)
	utils.CheckError(err)
	expectedState = "running"
	if expectedState != state {
		t.Errorf("unexpected state output!\nexpected: %v\nactual: %v", expectedState, state)
	}
}

func TestHandleNewJobRequests(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	err = HandleNewJobRequests(Driver)
	utils.CheckError(err)
}

func TestGetRunningJobs(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	uuids, err := GetRunningJobs(Driver)
	utils.CheckError(err)

	expectedUuids := []string{
		"4ce9189f-561a-4886-aeef-1836f28b073b",
		"2afd5896-9203-4c87-8790-a40091557d8d",
		"57c743a2-97a0-42be-86be-e659439d9e0e",
		"bc0b65b1-97d2-4be8-a472-d68d2a24f006",
		"9b72a160-584e-4c14-a87f-34dcdd346da4",
		"69972298-0729-47b3-ba28-2ce5292fae74",
		"724d8077-bfbe-4cf4-b2d3-ea3d84dc55c3",
		"35e28f6e-94f4-4a70-8556-d4d024893722",
		"78a39d4c-d8ca-4a7d-8cb5-206aca36b6a5",
		"724254b8-5d51-42f5-8394-99976f87e520",
		"2c03d81c-e321-41f0-a857-c9138bff70ee",
		"134c827f-e758-47f4-a7a0-fa13fe8d02d1",
		"7964da5e-0300-4634-a745-5aca7d365e55",
		"505af468-13b4-405f-a384-273a31c60e6a",
		"a052e18b-4c33-42b3-aa29-e5e0f2c4ad43",
		"b28d5289-0b2b-4aa0-996b-cfa07101035b",
	}

	if !reflect.DeepEqual(uuids, expectedUuids) {
		t.Errorf("uuids not as expected!\nexpected: %v\nactual: %v", expectedUuids, uuids)
	}
}

func TestUpdateJobStatus(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	testUuid := "4ce9189f-561a-4886-aeef-1836f28b073b"

	err = UpdateJobStatus(Driver, "pending", testUuid)
	utils.CheckError(err)

	err = UpdateJobStatus(Driver, "running", testUuid)
	utils.CheckError(err)
}

func TestUpdateCompleteJobs(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	err = UpdateCompleteJobs(Driver)
	utils.CheckError(err)
}

func TestGetFailedRowIdsForState(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	rowIds, err := GetFailedRowIdsForState(Driver, "2 minutes", "spawning")
	utils.CheckError(err)

	expectedRowIds := []string{
		"2",
		"41",
		"52",
		"66",
		"67",
	}

	if !reflect.DeepEqual(rowIds, expectedRowIds) {
		t.Errorf("unexpected row ids!\nexpected: %v\nactual: %v", expectedRowIds, rowIds)
	}

	rowIds, err = GetFailedRowIdsForState(Driver, "2 minutes", "running")
	utils.CheckError(err)

	expectedRowIds = []string{
		"39",
		"40",
		"93",
		"96",
	}

	if !reflect.DeepEqual(rowIds, expectedRowIds) {
		t.Errorf("unexpected row ids!\nexpected: %v\nactual: %v", expectedRowIds, rowIds)
	}
}

func TestBatchUpdateTestsWithRowIds(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	spawningRowIds := []string{
		"2",
		"41",
		"52",
		"66",
		"67",
	}

	err = BatchUpdateTestsWithRowIds(Driver, "state", "requested", spawningRowIds)
	utils.CheckError(err)
	err = BatchUpdateTestsWithRowIds(Driver, "state", "spawning", spawningRowIds)
	utils.CheckError(err)
}

func TestFixFailedSpawns(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	spawningRowIds := []string{
		"2",
		"41",
		"52",
		"66",
		"67",
	}

	err = FixFailedSpawns(Driver, "2 minutes")
	utils.CheckError(err)
	err = BatchUpdateTestsWithRowIds(Driver, "state", "spawning", spawningRowIds)
	utils.CheckError(err)
}

func TestFixFailedRuns(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	runningRowIds := []string{
		"39",
		"40",
		"93",
	}

	err = FixFailedRuns(Driver, "2 minutes")
	utils.CheckError(err)
	err = BatchUpdateTestsWithRowIds(Driver, "state", "running", runningRowIds)
	utils.CheckError(err)
}

// 5b45f42a-3508-40c2-b619-eb42ccf49d84
func TestDataRetentionPolicy(t *testing.T) {
	testUuid := "5b45f42a-3508-40c2-b619-eb42ccf49d84"

	Driver, err := database.TestDbDriver("guts_scheduler", "guts_scheduler")
	utils.CheckError(err)

	cfgPath := "./guts-scheduler-local.yaml"
	schedulerCfg, err := ParseConfig(cfgPath)
	utils.CheckError(err)

	// create backend
	backend, err := storage.GetStorageBackend(schedulerCfg.Storage)
	utils.CheckError(err)

	retentionDuration, err := time.ParseDuration("3s")
	utils.CheckError(err)

	// set up the test directory
	objectDir := fmt.Sprintf("%v/%v/", schedulerCfg.Storage["object_path"], testUuid)
	err = os.MkdirAll(objectDir, 0755)
	utils.CheckError(err)
	time.Sleep(time.Second * 4)

	err = DataRetentionPolicy(Driver, backend, retentionDuration)
	utils.CheckError(err)
}
