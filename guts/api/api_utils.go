package api

import (
	"fmt"
	// "github.com/lib/pq"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"log"
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
	log.Printf("%v\n", job)
	allJobColumns := []string{
		"uuid",
		"artifact_url",
		"tests_repo",
		"tests_repo_branch",
		"tests_plans",
		"image_url",
		"reporter",
		"status",
		"submitted_at",
		"requester",
		"debug",
		"priority",
	}
	queryString := fmt.Sprintf(
		`INSERT INTO jobs (%v) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		strings.Join(allJobColumns, ", "),
	)
	stmt, err := driver.PrepareQuery(queryString)
	if err != nil { // coverage-ignore
		log.Printf("failed to prepare query:\n%v\nwith:\n%v", queryString, err.Error())
		return err
	}
	defer utils.DeferredErrCheck(stmt.Close)
  // plansArr := strings.Replace(fmt.Sprintf(`'{E"%v"}'`, strings.Join(job.TestsPlans, `",E"`)), "/", `\/`, -1),
  // plansArr := fmt.Sprintf(`'{E'%v'}'`, strings.Join(job.TestsPlans, `",E"`))
  // plansArr = strings.Replace(plansArr, "/", `\/`, -1)

  // plansArr := pq.Array(job.TestsPlans)
  plansArr := fmt.Sprintf(`ARRAY ['%v']`, strings.Join(job.TestsPlans, `','`))

  log.Printf("%v\n", plansArr)
	_, err = stmt.Exec(
		job.Uuid,
		job.ArtifactUrl,
		job.TestsRepo,
		job.TestsRepoBranch,
    // string interpolation, replace `/` with `\/` ?
    // '{"tests/firefox-example/plans/regular.yaml"}'
    // fmt.Sprintf(`'{E"%v"}'`, strings.Join(job.TestsPlans, `",E"`)),
    // strings.Replace(fmt.Sprintf(`'{E"%v"}'`, strings.Join(job.TestsPlans, `",E"`)), "/", "\/", -1),
    plansArr,
    // strings.Replace(fmt.Sprintf(`'{E"%v"}'`, strings.Join(job.TestsPlans, `",E"`)), "/", `\/`, -1),
		// pq.Array(job.TestsPlans),
		// job.TestsPlans,
		job.ImageUrl,
		job.Reporter,
		job.Status,
		job.SubmittedAt,
		job.Requester,
		job.Debug,
		job.Priority,
	)
	if err != nil { // coverage-ignore
		log.Printf("failed to execute statement %v", stmt)
	}
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
