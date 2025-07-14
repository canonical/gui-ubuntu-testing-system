<!-- markdownlint-disable MD013 -->

# Basic Architecture and Functionality

## Entity Relationship Diagram

```mermaid

 erDiagram
     Postgres {
        table jobs
        table tests
        table users
        table reporter
    }
    API {
        endpoint request
        endpoint job
        endpoint artifacts
    }
    "User/Automation" }|..|| API : "job request (API key in headers)"
    API }|..|| Postgres : "validates job request and writes to db, expands testbed shorthand"
    Spawner }|..|{ Postgres: "Checks for new test requests and spawns VM, kills VM when test is complete"
    Scheduler }|..|{ Postgres: "|-Checks for any new jobs
    |-Writes n individual test requests for new jobs
    |-Checks for incomplete jobs
    |-Checks to see if all individual tests for a job are complete
    |-Marks job as pass or fail when all tests complete
    |-Checks for dead VMs, re-request if so
    |-Checks for dead yarf processes, re-request if so"
    Runner }|..|{ Postgres: "Runs test via yarf on waiting VMs, writes results"
    Reporter }|..|{ Postgres: "Reads tables and writes to external service (can only write to reporter table)"

```

## Schemas

### Job Request Schema

```mermaid

erDiagram
    "Job request (json)" {
        string artifact_url "[url to artifact to test in testbed] (leave empty to test the testbed)"
        string tests_repo "url to a git repo, defaults to gh/canonical/ubuntu-gui-testing"
        string tests_repo_branch "branch of tests_repo to test from, defaults to main"
        string test_plans "['tests/$dir/plan.yaml', 'tests/$other_dir/plan.yaml'], # test plan includes path to testdir e.g. tests/$dir, and other needed information"
        string testbed "points to a url for a .img or .iso, or a shorthand for an image, e.g. ubuntu-daily"
        string reporter "one of [test observer]"
        bool debug "add debug test artifacts"
        int priority "integer to indicate job queue hierarchy"
    }

```

## Database schemas

### 'jobs' table

```mermaid

erDiagram
    "'jobs' table" {
        string artifact_url "url to artifact to be tested"
        string tests_repo "repository containing yarf suitable tests"
        string tests_repo_branch "branch of tests_repo"
        string tests_plans "list of paths to .yml files detailing a suite of tests"
        string image_url "expanded from the shorthand provided in the test request, can also be a url to internally stored images"
        string uuid "primary key"
        string reporter "one of [test_observer]"
        string status "one of [pending, running, pass, fail]"
        datetime submitted_at "datetime of job request"
        string requester "username of requester"
        bool debug "add debug test artifacts"
        int priority "integer to indicate job queue hierarchy"
    }

```

### 'tests' table

```mermaid

erDiagram
    "'tests' table" {
        string test_case "a test case in the test plan"
        string uuid "foreign key to jobs table"
        string vnc_address "vnc host & port assigned this individual test case"
        string state "one of [requested/spawning/spawned/running/pass/fail]"
        string results_url "Either none or a URL, populated only when test case has finished"
        datetime updated_at "This must be modified on every update to an entry"
    }

```

### 'users' table

```mermaid

erDiagram
    "'users' table" {
        string username "LP username for developers, can be bots without LP accounts, or usernames not tied to LP"
        string key "the api key"
        int maximum_priority "integer describing the maximum allowed priority for the user"
    }

```

### 'reporter' table

```mermaid

erDiagram
    "'reporter' table" {
        string uuid "foreign key to jobs table"
        string base_reporting_url "base api url to report test results to for this uuid"
    }

```
