package api

import (
	"fmt"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"reflect"
	"strings"
)

func GetStatusUrlForUuid(uuid, cfgPath string) string {
	gutsCfg, err := ParseConfig(cfgPath)
	if err != nil { // coverage-ignore
		return ""
	}
	statusUrl := fmt.Sprintf("%v%v/status/%v", utils.GetProtocolPrefix(gutsCfg.Api.Port), gutsCfg.Api.Hostname, uuid)
	return statusUrl
}

func InsertJobsRow(job JobEntry, driver database.DbDriver) error {
	allJobColumns := []string{"uuid", "artifact_url", "tests_repo", "tests_repo_branch", "tests_plans", "image_url", "reporter", "status", "submitted_at", "requester", "debug", "priority"}
	queryString := fmt.Sprintf(
		`INSERT INTO jobs (%v) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		strings.Join(allJobColumns, ", "),
	)
	stmt, err := driver.PrepareQuery(queryString)
	if err != nil { // coverage-ignore
		return err
	}
	defer utils.DeferredErrCheck(stmt.Close)
	_, err = stmt.Exec(
		job.Uuid,
		job.ArtifactUrl,
		job.TestsRepo,
		job.TestsRepoBranch,
		fmt.Sprintf(`{%v}`, strings.Join(job.TestsPlans, ",")),
		job.ImageUrl,
		job.Reporter,
		job.Status,
		job.SubmittedAt,
		job.Requester,
		job.Debug,
		job.Priority,
	)
	return err
}

// This function is just used for tests, so we don't test it.
func SkipTestIfPostgresInactive(PgError error) bool { // coverage-ignore
	var expectedType database.PostgresServiceNotUpError
	if PgError != nil {
		if reflect.DeepEqual(reflect.TypeOf(PgError), reflect.TypeOf(expectedType)) {
			return true
		}
	}
	return false
}
