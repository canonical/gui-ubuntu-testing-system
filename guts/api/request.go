package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"log"
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

func (j JobRequest) ToJson() string {
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

// Don't need to test this directly, it's tested by api_test.go
func ProcessJobRequest(cfgPath, apiKey string, jobReq JobRequest, driver database.DbDriver) (string, error) { // coverage-ignore
	log.Printf("Processing job request %v", jobReq)
	if apiKey == "" {
		return "", EmptyApiKeyError{}
	}
	log.Printf("given api key: %v", apiKey)
	shakey := utils.Sha256sumOfString(apiKey)
	log.Printf("shakey: %v", shakey)
	userData, jobReq, err := AuthorizeUserAndAssignPriority(shakey, jobReq, driver)
	if err != nil {
		return "", ApiKeyNotAcceptedError{}
	}
	log.Printf("Parsed user data: %v", userData)
	if jobReq.ArtifactUrl != nil {
		if err = ValidateArtifactUrl(*jobReq.ArtifactUrl, cfgPath); err != nil {
			return "", err
		}
	}
	log.Printf("Validated artifact url")
	if err = ValidateTestbedUrl(jobReq.TestBed, cfgPath); err != nil {
		return "", err
	}
	log.Printf("Validated testbed url")
	if err = ValidateTestData(jobReq.TestsRepoBranch, jobReq.TestsRepo, jobReq.TestsPlans); err != nil {
		return "", err
	}
	jobRow := CreateJobEntry(jobReq, userData)
	log.Printf("Writing the following row to the jobs table:\n%v", jobRow)
	if err = WriteJobEntryToDb(jobRow, driver); err != nil { // coverage-ignore
		return "", err
	}
	log.Printf("Written to db!")
	returnJson := fmt.Sprintf(`{"uuid": "%v", "status_url": "%v"}`, jobRow.Uuid, GetStatusUrlForUuid(jobRow.Uuid, cfgPath))
	log.Printf("Returning the following json:\n%v", returnJson)
	return returnJson, nil
}

func GetAuthDataForKey(key string, driver database.DbDriver) (UserData, error) {
	var user UserData
	var params = []string{"username", "key", "maximum_priority"}
	row, err := driver.QueryRow("users", "key", key, params)
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

func AuthorizeUserAndAssignPriority(shadKey string, jobReq JobRequest, driver database.DbDriver) (UserData, JobRequest, error) {
	userData, err := GetAuthDataForKey(shadKey, driver)
	if err != nil {
		return userData, jobReq, err
	}
	if jobReq.Priority > userData.MaxPriority {
		jobReq.Priority = userData.MaxPriority
	}
	return userData, jobReq, nil
}

func ValidateArtifactUrl(artifactUrl, cfgPath string) error {
	gutsCfg, err := ParseConfig(cfgPath)
	utils.CheckError(err)
	types := []string{"snap", "deb"}
	err = ValidateUrlAgainstDomainsAndTypes(artifactUrl, gutsCfg.Api.ArtifactDomains, types)
	return err
}

func ValidateTestbedUrl(testbedUrl, cfgPath string) error {
	gutsCfg, err := ParseConfig(cfgPath)
	utils.CheckError(err)
	types := []string{"img", "iso"}
	err = ValidateUrlAgainstDomainsAndTypes(testbedUrl, gutsCfg.Api.TestbedDomains, types)
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
			defer utils.DeferredErrCheck(response.Body.Close)
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
		return utils.GenericGitError{Command: shallowCloneCmd.Args}
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
		return utils.GenericGitError{Command: sparseCheckoutCmd.Args}
	}

	gitCheckoutCmd := exec.Command("git", "checkout")
	gitCheckoutCmd.Dir = repoDir
	if err := gitCheckoutCmd.Run(); err != nil { // coverage-ignore
		rmErr := os.RemoveAll(tempDirName)
		if rmErr != nil {
			return rmErr
		}
		return utils.GenericGitError{Command: gitCheckoutCmd.Args}
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
	log.Printf("Created job entry:\n%v\n", thisJob)
	return thisJob
}

func WriteJobEntryToDb(job JobEntry, driver database.DbDriver) error {
	err := InsertJobsRow(job, driver)
	return err
}

// We don't test this function because it's only used for unit tests
func MakeDummyJobReq() JobRequest { // coverage-ignore
	var expectedJobReq JobRequest
	url := "myurl"
	expectedJobReq.ArtifactUrl = &url
	expectedJobReq.TestsRepo = "myrepo"
	expectedJobReq.TestsRepoBranch = "main"
	expectedJobReq.TestsPlans = []string{"plan1", "plan2"}
	expectedJobReq.TestBed = "mytestbedurl"
	expectedJobReq.Debug = false
	expectedJobReq.Priority = 1
	expectedJobReq.Reporter = ""
	return expectedJobReq
}
