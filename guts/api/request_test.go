package main

import (
  "testing"
  "reflect"
)


func TestParseJobFromJsonSuccess(t *testing.T) {
  inputJson := `{"artifact_url": "myurl", "tests_repo": "myrepo", "tests_repo_branch": "main", "tests_plans": ["plan1", "plan2"], "testbed": "mytestbedurl", "debug": 0, "priority": 1, "reporter": ""}`
  actualJobReq, err := ParseJobFromJson([]byte(inputJson))
  var expectedJobReq JobRequest
  expectedJobReq.ArtifactUrl = "myurl"
  expectedJobReq.TestsRepo = "myrepo"
  expectedJobReq.TestsRepoBranch = "main"
  expectedJobReq.TestsPlans = {"plan1", "plan2"}
  expectedJobReq.TestBed = "mytestbedurl"
  expectedJobReq.Debug = 0
  expectedJobReq.Priority = 1
  ...
}

