package api

import (
	"fmt"
	"github.com/lib/pq"
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
  log.Printf("Statement:\n%v", stmt)
	if err != nil { // coverage-ignore
		log.Printf("failed to prepare query:\n%v\nwith:\n%v", queryString, err.Error())
		return err
	}
	defer utils.DeferredErrCheck(stmt.Close)

  plansArr := pq.Array(job.TestsPlans)
  // plansArr := fmt.Sprintf(`'\{\"%v\"\}'`, strings.Join(job.TestsPlans, `\",\"`))
  // plansArr := fmt.Sprintf(`ARRAY ['%v']`, strings.Join(job.TestsPlans, `','`))
  arrVal, _ := plansArr.Value()
  inputString := arrVal.(string)
  log.Printf("***********************************************************")
  // inputString = strings.Replace(inputString, `"`, "$", -1)
  // inputString = strings.Replace(inputString, `/`, `\\/`, -1)
  log.Printf(inputString)
  log.Printf("***********************************************************")

  // log.Printf("*************************\ntest plans:\n%v\n", plansArr)
	_, err = stmt.Exec(
		job.Uuid,
		job.ArtifactUrl,
		job.TestsRepo,
		job.TestsRepoBranch,
    // plansArr,
    fmt.Sprintf(`'%v'::string[]`, inputString),
    // fmt.Sprintf(`'{%v}'`, inputString),
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
