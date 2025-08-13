package main

import (
  "encoding/json"
  "database/sql"
  "github.com/lib/pq"
)

type SingleJob struct {
  Uuid string `json:"uuid"`
  ArtifactUrl *string `json:"artifact_url"`
  TestsRepo string `json:"tests_repo"`
  TestsRepoBranch string `json:"tests_repo_branch"`
  TestsPlans []string `json:"tests_plans"`
  ImageUrl string `json:"image_url"`
  Reporter string `json:"reporter"`
  Status string `json:"status"`
  SubmittedAt string `json:"submitted_at"`
  Requester string `json:"requester"`
  Debug bool `json:"debug"`
  Priority int `json:"priority"`
}

type JobWithTestsDetails struct {
  Job SingleJob
  Results map[string]string `json:"results"`
}

type ReturnableJson interface {
  toJson()
}

func (j SingleJob) toJson() string {
  b, err := json.Marshal(j)
  if err != nil { // coverage-ignore
    return ""
  }
  return string(b)
}

func (j JobWithTestsDetails) toJson() string {
  b, err := json.Marshal(j)
  if err != nil { // coverage-ignore
    return ""
  }
  return string(b)
}

func GetCompleteResultsForUuid(uuidToFind string, db *sql.DB) (JobWithTestsDetails, error) {
  var completeJob JobWithTestsDetails
  job, err := FindJobByUuid(uuidToFind, db)
  if err != nil {
    return completeJob, err
  }

  testResults, err := CollateUuidTestResults(uuidToFind, db)
  if err != nil { // coverage-ignore
    return completeJob, err
  }

  completeJob.Job = job
  completeJob.Results = testResults

  return completeJob, nil
}

func CollateUuidTestResults(uuidToFind string, db *sql.DB) (map[string]string, error) {
  testResults := make(map[string]string)

  stmt, err := db.Prepare("SELECT test_case, state FROM tests WHERE uuid = $1")
  if err != nil { // coverage-ignore
    return testResults, err
  }
  defer DeferredErrCheck(stmt.Close)

  rows, err := stmt.Query(uuidToFind)
  if err != nil { // coverage-ignore
    return testResults, err
  }
  defer DeferredErrCheck(rows.Close)

  for rows.Next() {
    var testCase string
    var state string
    err := rows.Scan(&testCase, &state)
    if err != nil { // coverage-ignore
      return testResults, err
    }
    testResults[testCase] = state
  }
  return testResults, nil
}

func FindJobByUuid(uuidToFind string, db *sql.DB) (SingleJob, error) {
  var job SingleJob

  stmt, err := db.Prepare("SELECT uuid, artifact_url, tests_repo, tests_repo_branch, tests_plans, image_url, reporter, status, submitted_at, requester, debug, priority FROM jobs WHERE uuid=$1")
  if err != nil { // coverage-ignore
    return job, err
  }
  defer DeferredErrCheck(stmt.Close)

  err = stmt.QueryRow(uuidToFind).Scan(
    &job.Uuid,
    &job.ArtifactUrl,
    &job.TestsRepo,
    &job.TestsRepoBranch,
    pq.Array(&job.TestsPlans),
    &job.ImageUrl,
    &job.Reporter,
    &job.Status,
    &job.SubmittedAt,
    &job.Requester,
    &job.Debug,
    &job.Priority,
  )
  if err != nil {
    if err == sql.ErrNoRows {
      return job, UuidNotFoundError{uuid: uuidToFind}
    }
    return job, err // coverage-ignore
  }
  return job, nil
}

