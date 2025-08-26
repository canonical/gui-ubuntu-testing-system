package main

import (
  "testing"
  "reflect"
)

func TestFindJobByUuid(t *testing.T) {
  Uuid := "4ce9189f-561a-4886-aeef-1836f28b073b"
  ParseArgs()
  gutsCfg, err := ParseConfig(configFilePath)
  CheckError(err)
  db, err = PostgresConnect(gutsCfg)
  CheckError(err)
  job, err := FindJobByUuid(Uuid, db)
  CheckError(err)
  var TestJob SingleJob
  TestJob.Uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
  TestJob.ArtifactUrl = nil
  TestJob.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
  TestJob.TestsRepoBranch = "main"
  TestJob.TestsPlans = []string{"tests/firefox-example/plans/extended.yaml", "tests/firefox-example/plans/regular.yaml"}
  TestJob.ImageUrl = "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
  TestJob.Reporter = "test_observer"
  TestJob.Status = "running"
  TestJob.SubmittedAt = "2025-07-23T14:17:14.632177"
  TestJob.Requester = "andersson123"
  TestJob.Debug = false
  TestJob.Priority = 8
  if !reflect.DeepEqual(job, TestJob) {
    t.Errorf("Expected job not the same as actual:\n%v\n%v", TestJob, job)
  }
}

func TestJobToJson(t *testing.T) {
  var TestJob SingleJob
  TestJob.Uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
  TestJob.ArtifactUrl = nil
  TestJob.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
  TestJob.TestsRepoBranch = "main"
  TestJob.TestsPlans = []string{"tests/firefox-example/plans/extended.yaml", "tests/firefox-example/plans/regular.yaml"}
  TestJob.ImageUrl = "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
  TestJob.Reporter = "test_observer"
  TestJob.Status = "running"
  TestJob.SubmittedAt = "2025-07-23T14:17:14.632177"
  TestJob.Requester = "andersson123"
  TestJob.Debug = false
  TestJob.Priority = 8
  ExpectedJson := `{"uuid":"4ce9189f-561a-4886-aeef-1836f28b073b","artifact_url":null,"tests_repo":"https://github.com/canonical/ubuntu-gui-testing.git","tests_repo_branch":"main","tests_plans":["tests/firefox-example/plans/extended.yaml","tests/firefox-example/plans/regular.yaml"],"image_url":"https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso","reporter":"test_observer","status":"running","submitted_at":"2025-07-23T14:17:14.632177","requester":"andersson123","debug":false,"priority":8}`
  ConvertedJson := TestJob.toJson()
  if !reflect.DeepEqual(ExpectedJson, ConvertedJson) {
    t.Errorf("json conversion not as expected!\nExpected: %v\nActual: %v", ExpectedJson, ConvertedJson)
  }
}


