package main

import (
  "fmt"
)


type UuidNotFoundError struct {
    uuid string
}

func (e UuidNotFoundError) Error() string {
    return fmt.Sprintf("No jobs with uuid %v found!", e.uuid)
}

type InvalidUuidError struct {
  uuid string
}

func (e InvalidUuidError) Error() string {
    return fmt.Sprintf("%v isn't a valid uuid!", e.uuid)
}

type PostgresServiceNotUpError struct {}

func (e PostgresServiceNotUpError) Error() string {
  return fmt.Sprintf("Unit postgresql.service is not active.")
}

type BadUrlError struct {
  url string
  code int
}

func (b BadUrlError) Error() string {
  return fmt.Sprintf("Url %v returned %v", b.url, b.code)
}

type NonWhitelistedDomainError struct {
  url string
}

func (n NonWhitelistedDomainError) Error() string {
  return fmt.Sprintf("Url %v not from accepted list of domains", n.url)
}

type ApiKeyNotAcceptedError struct {}

func (a ApiKeyNotAcceptedError) Error() string {
  return "Api key not accepted!"
}

type EmptyApiKeyError struct {}

func (e EmptyApiKeyError) Error() string {
  return "Api key passed is empty"
}

type GenericGitError struct {
  command []string
}

func (g GenericGitError) Error() string {
  return fmt.Sprintf("Git operation failed:\n%v", g.command)
}

type PlanFileNonexistentError struct {
  planFile string
}

func (p PlanFileNonexistentError) Error() string {
  return fmt.Sprintf("Plan file %v doesn't exist!", p.planFile)
}

