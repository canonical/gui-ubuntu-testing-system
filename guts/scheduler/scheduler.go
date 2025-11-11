package scheduler

import (
	"fmt"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/storage"
	"guts.ubuntu.com/v2/utils"
	"log"
	"os"
	"strings"
	"time"
)

type TestsEntry struct {
	Uuid       string
	TestCase   string
	VncAddress string
	State      string
	ResultsUrl string
	UpdatedAt  time.Time
	Tpm        bool
	CommitHash string
	Plan       string
}

func GetNewJobsUuids(Driver database.DbDriver) ([]string, error) {
	var uuids []string

	uuidQuery := `SELECT uuid FROM jobs EXCEPT SELECT uuid FROM tests`
	stmt, err := Driver.PrepareQuery(uuidQuery)
	if err != nil { // coverage-ignore
		return uuids, err
	}
	defer utils.DeferredErrCheck(stmt.Close)

	rows, err := stmt.Query()
	if err != nil { // coverage-ignore
		return uuids, err
	}

	for rows.Next() {
		var thisUuid string
		err = rows.Scan(&thisUuid)
		if err != nil { // coverage-ignore
			return uuids, err
		}
		uuids = append(uuids, thisUuid)
	}

	if err = rows.Err(); err != nil { // coverage-ignore
		return uuids, err
	}

	return uuids, nil
}

func GetTestData(Driver database.DbDriver, Uuid string) (string, string, []string, error) {
	testsRepo := ""
	testsRepoBranch := ""
	var testsPlans []string

	var plansBytes []uint8

	columns := []string{"tests_repo", "tests_repo_branch", "tests_plans"}

	row, err := Driver.QueryRow("jobs", "uuid", Uuid, columns)
	if err != nil { // coverage-ignore
		return testsRepo, testsRepoBranch, testsPlans, err
	}

	err = row.Scan(
		&testsRepo,
		&testsRepoBranch,
		&plansBytes,
	)
	if err != nil { // coverage-ignore
		return testsRepo, testsRepoBranch, testsPlans, err
	}

	plansString := string(plansBytes)
	plansString = strings.Replace(plansString, "{", "", -1)
	plansString = strings.Replace(plansString, "}", "", -1)
	testsPlans = strings.Split(plansString, ",")

	return testsRepo, testsRepoBranch, testsPlans, nil
}

///////////////////////////////////////////////////////////////////////////
// tested up to here

func WriteTestsForJob(Driver database.DbDriver, Uuid string) error {
	cloneDirName, err := os.MkdirTemp("", "gitrepo")
	if err != nil { // coverage-ignore
		return err
	}

	defer utils.DeferredErrCheckStringArg(os.RemoveAll, cloneDirName)

	testsRepo, testsRepoBranch, testPlanPaths, err := GetTestData(Driver, Uuid)
	if err != nil { // coverage-ignore
		return err
	}

	err = utils.GitCloneToDir(testsRepo, testsRepoBranch, cloneDirName)
	if err != nil { // coverage-ignore
		return err
	}

	for _, planPath := range testPlanPaths {
		// create full plan path
		fullPlanPath := fmt.Sprintf("%v/%v", cloneDirName, planPath)
		// parse test plan
		testPlan, err := utils.ParsePlan(fullPlanPath)
		if err != nil { // coverage-ignore
			return err
		}

		for _, testCase := range testPlan.Tests {
			// create test entry
			var tEntry TestsEntry
			tEntry.Uuid = Uuid
			tEntry.TestCase = testCase.Name
			tEntry.VncAddress = ""
			tEntry.State = "requested"
			tEntry.ResultsUrl = ""
			tEntry.UpdatedAt = time.Now()
			tEntry.Tpm = testCase.Data.Requirements.Tpm
			tEntry.CommitHash = ""
			tEntry.Plan = planPath

			// write test entry to database
			err = WriteTestToDb(Driver, tEntry)
			if err != nil { // coverage-ignore
				return err
			}
		}
	}
	return nil
}

func WriteTestToDb(Driver database.DbDriver, test TestsEntry) error {
	columns := []string{"uuid", "test_case", "vnc_address", "state", "results_url", "updated_at", "tpm", "commit_hash", "plan"}
	queryString := fmt.Sprintf(
		`INSERT INTO tests (%v) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		strings.Join(columns, ", "),
	)

	stmt, err := Driver.PrepareQuery(queryString)
	if err != nil { // coverage-ignore
		return err
	}
	defer utils.DeferredErrCheck(stmt.Close)

	_, err = stmt.Exec(
		test.Uuid,
		test.TestCase,
		test.VncAddress,
		test.State,
		test.ResultsUrl,
		test.UpdatedAt,
		test.Tpm,
		test.CommitHash,
		test.Plan,
	)

	return err
}

func GetUpdatedJobState(Driver database.DbDriver, Uuid string) (string, error) {
	newState := ""

	rows, err := Driver.Query("tests", "uuid", Uuid, []string{"state"})
	if err != nil { // coverage-ignore
		return "", err
	}

	log.Printf("Getting updated job state for %v\n", Uuid)

	for rows.Next() {
		var thisState string

		err = rows.Scan(&thisState)
		if err != nil { // coverage-ignore
			return newState, err
		}

		log.Printf("state=%v\n", thisState)

		if thisState != "pass" && thisState != "fail" {
			return "running", nil
		}
		if thisState != newState {
			if newState != "fail" {
				newState = thisState
			}
		}
	}

	if err = rows.Err(); err != nil { // coverage-ignore
		return "", err
	}

	return newState, nil
}

func HandleNewJobRequests(Driver database.DbDriver) error {
	currUuids, err := GetNewJobsUuids(Driver)
	if err != nil { // coverage-ignore
		return err
	}

	log.Printf("new job uuids: %v\n", currUuids)

	for _, thisUuid := range currUuids {
		log.Printf("Writing tests to tests table for job %v\n", thisUuid)
		err = WriteTestsForJob(Driver, thisUuid)
		// As per the dogma through the rest of this repo - WriteTestsForJob
		// only returns non-nil in the event of standard library errors,
		// so we ignore this block
		if err != nil { // coverage-ignore
			return err
		}
	}
	return nil
}

func GetRunningJobs(Driver database.DbDriver) ([]string, error) {
	var uuids []string

	rows, err := Driver.Query("jobs", "status", "running", []string{"uuid"})
	if err != nil { // coverage-ignore
		return uuids, err
	}

	for rows.Next() {
		var thisUuid string
		err = rows.Scan(&thisUuid)
		if err != nil { // coverage-ignore
			return uuids, err
		}
		uuids = append(uuids, thisUuid)
	}

	if err = rows.Err(); err != nil { // coverage-ignore
		return uuids, err
	}

	return uuids, nil
}

func UpdateJobStatus(Driver database.DbDriver, status, uuid string) error {
	log.Printf("setting job %v status to %v", uuid, status)
	updateQuery := fmt.Sprintf(`UPDATE jobs SET status='%v' WHERE uuid='%v'`, status, uuid)
	err := Driver.UpdateRow(updateQuery)
	return err
}

func UpdateCompleteJobs(Driver database.DbDriver) error {
	log.Println("checking for complete jobs")
	// find jobs in state running
	runningUuids, err := GetRunningJobs(Driver)
	if err != nil { // coverage-ignore
		return err
	}

  if len(runningUuids) < 1 {
    return nil
  }
	log.Printf("checking if any of the following jobs have finished running yet:\n%v", runningUuids)

	// check results of accompanying tests
	for _, Uuid := range runningUuids {
		newState, err := GetUpdatedJobState(Driver, Uuid)
		if err != nil { // coverage-ignore
			return err
		}
		log.Printf("updated job state for %v is %v", Uuid, newState)

		// if all tests in pass or fail, update the job entry accordingly
		if newState != "running" {
			err = UpdateJobStatus(Driver, newState, Uuid)
			if err != nil { // coverage-ignore
				return err
			}
		}
	}
	return nil
}

func GetFailedRowIdsForState(Driver database.DbDriver, interval, state string) ([]string, error) {
	var ids []string

	idQuery := fmt.Sprintf(`SELECT id FROM tests WHERE state='%s' AND updated_at < (now() - interval '%s')`, state, interval)
	stmt, err := Driver.PrepareQuery(idQuery)
	if err != nil { // coverage-ignore
		return ids, err
	}
	defer utils.DeferredErrCheck(stmt.Close)

	rows, err := stmt.Query()
	if err != nil { // coverage-ignore
		return ids, err
	}

	for rows.Next() {
		var thisId string
		err = rows.Scan(&thisId)
		if err != nil { // coverage-ignore
			return ids, err
		}
		ids = append(ids, thisId)
	}

	if err = rows.Err(); err != nil { // coverage-ignore
		return ids, err
	}

	return ids, nil
}

func BatchUpdateTestsWithRowIds(Driver database.DbDriver, field, value string, ids []string) error {
	query := fmt.Sprintf(`UPDATE tests SET %s='%s' WHERE id IN (%s)`, field, value, strings.Join(ids, ", "))
	err := Driver.UpdateRow(query)
	return err
}

func FixFailedSpawns(Driver database.DbDriver, interval string) error {
	ids, err := GetFailedRowIdsForState(Driver, interval, "spawning")
	if err != nil { // coverage-ignore
		return err
	}
	return BatchUpdateTestsWithRowIds(Driver, "state", "requested", ids)
}

func FixFailedRuns(Driver database.DbDriver, interval string) error {
	ids, err := GetFailedRowIdsForState(Driver, interval, "running")
	if err != nil { // coverage-ignore
		return err
	}
	return BatchUpdateTestsWithRowIds(Driver, "state", "requested", ids)
}

func DataRetentionPolicy(Driver database.DbDriver, backend storage.StorageBackend, duration time.Duration) error {
	// clear the object storage
	uuids, err := backend.RemoveObjectsOlderThan(duration)
	if err != nil { // coverage-ignore
		return err
	}

	// iterate through uuids and nuke each of them in the db
	for _, uuid := range uuids {
		err = Driver.NukeUuid(uuid)
		if err != nil { // coverage-ignore
			return err
		}
	}
	return nil
}

func SchedulerLoop(Driver database.DbDriver, SchedulerCfg GutsSchedulerConfig) error { // coverage-ignore

	// Scheduler step 1: handle new job requests
	err := HandleNewJobRequests(Driver)
	if err != nil {
		return err
	}

	// Scheduler step 2: Update complete jobs
	err = UpdateCompleteJobs(Driver)
	if err != nil {
		return err
	}

	// Scheduler step 3: Check for failed spawner processes
	err = FixFailedSpawns(Driver, SchedulerCfg.TestInactiveResetTime)
	if err != nil {
		return err
	}

	// Scheduler step 4: Check for failed runner processes
	err = FixFailedRuns(Driver, SchedulerCfg.TestInactiveResetTime)
	if err != nil {
		return err
	}

	// get the storage backend
	backend, err := storage.GetStorageBackend(SchedulerCfg.Storage)
	if err != nil {
		return err
	}

	// Scheduler step 5: Remove old objects and db entries
	retentionDuration, err := time.ParseDuration(fmt.Sprintf("%vd", SchedulerCfg.ArtifactRetentionDays))
	if err != nil {
		return err
	}
	err = DataRetentionPolicy(Driver, backend, retentionDuration)
	return err
}
