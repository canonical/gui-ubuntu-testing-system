package scheduler

import (
  "net/http"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/storage"
	"guts.ubuntu.com/v2/utils"
)


func GetNewJobsUuids(Driver database.DbDriver) ([]string, error) {
  var uuids []string

  uuidQuery := `SELECT DISTINCT uuid FROM jobs t1 LEFT JOIN tests t2 USING (uuid) WHERE t2.uuid IS NULL`
  stmt, err := Driver.PrepareQuery(uuidQuery)
  if err != nil { // coverage-ignore
    return err
  }
  defer utils.DeferredErrCheck(stmt.Close)

  rows, err := stmt.Query()
  if err != nil { // coverage-ignore
    return uuids, err
  }

  for rows.Next() {
    var thisUuid string
    err = rows.Scan(&thisUuid)
    if err != nil {
      return uuids, err
    }
    uuids = append(uuids, thisUuid)
  }

  if err = rows.Err(); err != nil {
    return uuids, err
  }

  return uuids, nil
}

// WHAT ARE WE DOING!
// - CONFIG! obvs
// - handle new job requests in the jobs table
//   - poll jobs table for jobs that don't exist in the tests table - done
//   - clone the git data for this job
//   - with the git data, write individual tests to the tests table
//   - clear the git data
//   - continue
// - check for complete jobs/sets of tests and update jobs table
//   - find jobs in state running
//   - find accompanying tests
//   - if all tests in pass or fail, update the job entry accordingly
// - check for failed vm spawning
//   - check tests table for tests in state "spawning", with updated_at more than 3 minutes ago
//   - set test state back to "requested"
// - check for failed runner runs
//   - check tests table for tests in state "running", with updated_at more than 3 minutes ago
//   - set test state back to "requested"
// - data retention policy
//   - remove swift storage objects that were last updated more than 180 days ago
// - fin



