package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"
)

type JobRequest struct {
	ArtifactUrl     *string  `json:"artifact_url"` // has to be a pointer because it can be empty
	TestsRepo       string   `json:"tests_repo"`
	TestsRepoBranch string   `json:"tests_repo_branch"`
	TestsPlans      []string `json:"tests_plans"`
	TestBed         string   `json:"testbed"`
	Debug           bool     `json:"debug"`
	Priority        int      `json:"priority"`
	Reporter        string   `json:"reporter"`
}

func (j JobRequest) toJson() string {
	b, err := json.Marshal(j)
	if err != nil { // coverage-ignore
		return ""
	}
	return string(b)
}

type UserData struct {
	Username    string
	Key         string
	MaxPriority int
}

func ParseJobFromJson(jsonData []byte) (JobRequest, error) {
	var thisJob JobRequest
	err := json.Unmarshal(jsonData, &thisJob)
	return thisJob, err
}

func ProcessJobRequest(apiKey string, jobReq JobRequest) (string, error) {
	if apiKey == "" {
		return "", EmptyApiKeyError{}
	}
	shakey := Sha256sumOfString(apiKey)
	userData, jobReq, err := AuthorizeUserAndAssignPriority(shakey, jobReq)
	if err != nil {
		return "", ApiKeyNotAcceptedError{}
	}
	if err = ValidateArtifactUrl(*jobReq.ArtifactUrl); err != nil {
		return "", err
	}
	if err = ValidateTestbedUrl(jobReq.TestBed); err != nil {
		return "", err
	}
	if err = ValidateTestData(jobReq.TestsRepoBranch, jobReq.TestsRepo, jobReq.TestsPlans); err != nil {
		return "", err
	}
	jobRow := CreateJobEntry(jobReq, userData)
	if err = WriteJobEntryToDb(jobRow); err != nil { // coverage-ignore
		return "", err
	}
	returnJson := fmt.Sprintf(`{"uuid": "%v", "status_url": "%v"}`, jobRow.Uuid, GetStatusUrlForUuid(jobRow.Uuid))
	return returnJson, nil
}

func GetAuthDataForKey(key string) (UserData, error) {
	var user UserData
	var params = []string{"username", "key", "maximum_priority"}
	row, err := Driver.QueryRow("users", "key", key, params)
	if err != nil { // coverage-ignore
		return user, err
	}
	err = row.Scan(
		&user.Username,
		&user.Key,
		&user.MaxPriority,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("key %v doesn't exist", key)
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
	CheckError(err)
	types := []string{"snap", "deb"}
	err = ValidateUrlAgainstDomainsAndTypes(artifactUrl, GutsCfg.Api.ArtifactDomains, types)
	return err
}

func ValidateTestbedUrl(testbedUrl string) error {
	err := ParseConfig(configFilePath)
	CheckError(err)
	types := []string{"img", "iso"}
	err = ValidateUrlAgainstDomainsAndTypes(testbedUrl, GutsCfg.Api.TestbedDomains, types)
	return err
}

func ValidateUrlAgainstDomainsAndTypes(url string, domains []string, artifactTypes []string) error {
	orRegex := strings.Join(artifactTypes, "|")
	for _, entry := range domains {
		thisRegex := fmt.Sprintf(`(http|https):\/\/%v\/(.*)\.(%v)`, entry, orRegex)
		match := regexp.MustCompile(thisRegex).MatchString(url)
		if match {
			response, err := http.Get(url)
			if err != nil { // coverage-ignore
				return err
			}
			defer DeferredErrCheck(response.Body.Close)
			if response.StatusCode < 300 && response.StatusCode >= 200 {
				return nil
			} else {
				return BadUrlError{url: url, code: response.StatusCode}
			}
		} else if !match && strings.Contains(url, entry) {
			return InvalidArtifactTypeError{url: url}
		}
	}
	return NonWhitelistedDomainError{url: url}
}

func ValidateTestData(testsRepoBranch, testsRepo string, testPlans []string) error {
	repoDirName := "repoDir"

	tempDirName, err := os.MkdirTemp("", "gitrepo")
	if err != nil { // coverage-ignore
		rmErr := os.RemoveAll(tempDirName)
		if rmErr != nil {
			return rmErr
		}
		return err
	}

	repoDir := fmt.Sprintf("%v/%v", tempDirName, repoDirName)
	err = os.MkdirAll(repoDir, 0755)
	if err != nil { // coverage-ignore
		rmErr := os.RemoveAll(tempDirName)
		if rmErr != nil {
			return rmErr
		}
		return err
	}

	shallowCloneCmd := exec.Command(
		"git",
		"clone",
		"--branch",
		testsRepoBranch,
		"--no-checkout",
		"--depth=1",
		"--filter=tree:0",
		testsRepo,
		repoDir,
	)
	if err := shallowCloneCmd.Run(); err != nil {
		err = os.RemoveAll(tempDirName)
		if err != nil { // coverage-ignore
			return err
		}
		return GenericGitError{command: shallowCloneCmd.Args}
	}

	sparseCheckoutCmd := exec.Command(
		"git",
		"sparse-checkout",
		"set",
		"--no-cone",
	)
	sparseCheckoutCmd.Args = slices.Concat(sparseCheckoutCmd.Args, testPlans)
	sparseCheckoutCmd.Dir = repoDir
	if err := sparseCheckoutCmd.Run(); err != nil { // coverage-ignore
		rmErr := os.RemoveAll(tempDirName)
		if rmErr != nil {
			return rmErr
		}
		return GenericGitError{command: sparseCheckoutCmd.Args}
	}

	gitCheckoutCmd := exec.Command("git", "checkout")
	gitCheckoutCmd.Dir = repoDir
	if err := gitCheckoutCmd.Run(); err != nil { // coverage-ignore
		rmErr := os.RemoveAll(tempDirName)
		if rmErr != nil {
			return rmErr
		}
		return GenericGitError{command: gitCheckoutCmd.Args}
	}

	for _, testPlan := range testPlans {
		planFile := fmt.Sprintf("%v/%v", repoDir, testPlan)
		if _, err := os.Stat(planFile); err != nil {
			rmErr := os.RemoveAll(tempDirName)
			if rmErr != nil { // coverage-ignore
				return rmErr
			}
			return PlanFileNonexistentError{planFile: planFile}
		}
	}
	rmErr := os.RemoveAll(tempDirName)
	if rmErr != nil { // coverage-ignore
		return rmErr
	}

	return nil
}

func CreateJobEntry(job JobRequest, uData UserData) JobEntry { // coverage-ignore
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
