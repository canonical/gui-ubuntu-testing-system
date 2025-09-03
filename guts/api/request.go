package main

import (
  "fmt"
  "encoding/json"
  "regexp"
  "net/http"
  "strings"
  "slices"
  "os"
  "os/exec"
  "github.com/google/uuid"
  "time"
)

type JobRequest struct {
  ArtifactUrl *string `json:"artifact_url"`  // has to be a pointer because it can be empty
  TestsRepo string `json:"tests_repo"`
  TestsRepoBranch string `json:"tests_repo"`
  TestsPlans []string `json:"tests_plans"`
  TestBed string `json:"testbed"`
  Debug bool `json:"debug"`
  Priority int `json:"priority"`
  Reporter string `json:"reporter"`
}

type UserData struct {
  Username string
  Key string
  MaxPriority int
}

func ParseJobFromJson(jsonData []byte) (JobRequest, error) {
  var thisJob JobRequest
  err := json.Unmarshal(jsonData, thisJob)
  return thisJob, err
}

func ProcessJobRequest(apiKey string, jobReq JobRequest) (string, error) {
  if apiKey == "" {
    return EmptyApiKeyError
  }
  shakey := Sha256sumOfString(apiKey)
  userData, jobReq, err := AuthorizeUserAndAssignPriority(shakey, jobReq)
  if err != nil {
    return ApiKeyNotAcceptedError
  }
  if err = ValidateArtifactUrl(jobReq.ArtifactUrl); err != nil {
    return err
  }
  if err = ValidateTestbedUrl(jobReq.TestbedUrl); err != nil {
    return err
  }
  if err = ValidateTestData(jobReq.TestsRepoBranch, jobReq.TestsRepo, jobReq.TestsPlans); err != nil {
    return err
  }
  jobRow := CreateJobEntry(jobReq, userData)
  if err = WriteJobEntryToDb(jobRow); err != nil {
    return err
  }
  returnJson := fmt.Sprintf(`{"uuid": "%v", "status_url": "%v"}`, jobRow.Uuid, GetStatusUrlForUuid(jobRow.Uuid))
  return returnJson, nil
}

func GetAuthDataForKey(key string) (UserData, error) {
  var user UserData
  var params = [...]string{"username", "key", "maximum_priority"}
  row, err := Driver.QueryRow("users", "user", user, params)
  if err != nil {
    return user, err
  }
  err = row.Scan(
    &user.Username,
    &user.Key,
    &user.MaxPriority
  )
  if err != nil {
    if err == sql.ErrNoRows {
      return user, fmt.Errorf("User %v doesn't exist!")
    }
  }
  return user, err
}

func AuthorizeUserAndAssignPriority(shadKey string, jobReq JobRequest) (UserData, JobRequest, error) {
  userData, err := GetAuthDataForKey(shadKey)
  if err != nil {
    return userData, jobReq, err
  }
  if jobReq.Priority > userData.MaxPriority {
    jobReq.Priority = userData.MaxPriority
  }
  return userData, jobReq, nil
}

func ValidateArtifactUrl(artifactUrl string) error {
  err := ParseConfig(configFilePath)
  var types []string = {"snap", "deb"}
  err = ValidateUrlAgainstDomainsAndTypes(artifactUrl, GutsCfg.Api.ArtifactDomains, types)
  return err
}

func ValidateTestbedUrl(testbedUrl string) error {
  err := ParseConfig(configFilePath)
  var types []string = {"img", "iso"}
  err = ValidateUrlAgainstDomainsAndTypes(testbedUrl, GutsCfg.Api.TestbedDomains, types)
  return err
}

func ValidateUrlAgainstDomainsAndTypes(url string, domains []string, artifactTypes []string) error {
  orRegex := strings.Join(artifactTypes, "|")
  for idx, entry := range(domains) {
    thisRegex := fmt.Sprintf(`https:\/\/%v\/(.*)\.(%v)`, entry, orRegex)
    err := regexp.MustCompile(thisRegex).MatchString(url)
    if err == nil {
      response, err := http.Get(url)
      if err != nil {
        return err
      }
      defer DeferredErrCheck(response.Body.Close)
      if response.StatusCode < 300 && response.StatusCode >= 200 {
        return nil
      } else {
        return BadUrlError{url: url, code: response.StatusCode}
      }
    }
  }
  return NonWhitelistedDomainError{url: url}
}

func ValidateTestData(testsRepoBranch, testsRepo string, testPlans []string) error {
  repoDirName := "repoDir"
  var gitShallowCloneCmdString []string = {
    "git",
    "clone",
    "--branch",
    testsRepoBranch,
    "--no-checkout",
    "--depth=1",
    "--filter=tree:0",
    testsRepo,
    repoDir
  }
  var gitSparseCheckoutCmdString []string = {
    "git",
    "sparse-checkout",
    "set",
    "--no-cone"
  }
  gitSparseCheckoutCmd = slices.Concat(gitSparseCheckoutCmd, testPlans)
  var gitCheckoutCmdString []string = {
    "git",
    "checkout"
  }
  tempDirName, err := os.MkdirTemp("", "gitrepo")
  if err != nil {
    return err
  }
  repoDir := fmt.Sprintf("%v/%v", tempDirName, repoDirName)
  DeferredErrCheckStringArg(os.RemoveAll, dirName)
  shallowCloneCmd := exec.Command(gitShallowCloneCmdString)
  if err := shallowCloneCmd.Run(); err != nil {
    return GenericGitError{command: gitShallowCloneCmdString}
  }
  sparseCheckoutCmd := exec.Command(gitSparseCheckoutCmdString)
  if err := sparseCheckoutCmd.Run(); err != nil {
    return GenericGitError{command: gitSparseCheckoutCmdString}
  }
  gitCheckoutCmd := exec.Command(gitCheckoutCmdString)
  if err := gitCheckoutCmd.Run(); err != nil {
    return GenericGitError{command: gitCheckoutCmdString}
  }
  for idx, testPlan := range(testPlans) {
    planFile := strings.Join(repoDir, testPlan)
    if _, err := os.Stat(planFile); err != nil {
      return PlanFileNonexistentError{planFile: planFile}
    }
  }
  return nil
}

func CreateJobEntry(job JobRequest, uData UserData) JobEntry {
  var thisJob JobEntry
  thisJob.Uuid = uuid.New().String()
  thisJob.ArtifactUrl = job.ArtifactUrl
  thisJob.TestsRepo = job.TestsRepo
  thisJob.TestsRepoBranch = job.TestsRepoBranch
  thisJob.TestsPlans = job.TestsPlans
  thisJob.ImageUrl = job.TestBed
  thisJob.Reporter = job.Reporter
  thisJob.Status = "pending"
  thisJob.SubmittedAt = time.Now()
  thisJob.Requester = uData.Username
  thisJob.Debug = job.Debug
  thisJob.Priority = job.Priority
  return thisJob
}

func WriteJobEntryToDb(job JobEntry) error {
  err := Driver.InsertJobsRow(job)
  return err
}

