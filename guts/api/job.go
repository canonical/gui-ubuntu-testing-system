package main

import (
  "encoding/json"
  "database/sql"
  "github.com/lib/pq"
  "time"
)

var (
  AllJobColumns = [...]string{"uuid", "artifact_url", "tests_repo", "tests_repo_branch", "tests_plans", "image_url", "reporter", "status", "submitted_at", "requester", "debug", "priority"}
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
  SubmittedAt time.Time `json:"submitted_at"`
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

func GetCompleteResultsForUuid(uuidToFind string) (JobWithTestsDetails, error) {
  var completeJob JobWithTestsDetails
  job, err := FindJobByUuid(uuidToFind)
  if err != nil {
    return completeJob, err
  }

  testResults, err := CollateUuidTestResults(uuidToFind)
  if err != nil { // coverage-ignore
    return completeJob, err
  }

  completeJob.Job = job
  completeJob.Results = testResults

  return completeJob, nil
}

func CollateUuidTestResults(uuidToFind string) (map[string]string, error) {
  testResults := make(map[string]string)

  var params = [...]string{"test_case", "state"}
  rows, err := Driver.Query("tests", "uuid", uuidToFind, params)
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

func FindJobByUuid(uuidToFind string) (SingleJob, error) {
  var job SingleJob

  rows, err := Driver.QueryRow("tests", "uuid", uuidToFind, AllJobColumns)
  if err != nil { // coverage-ignore
    return job, err
  }
  err = row.Scan(
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

