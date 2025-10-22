package scheduler

import (
  "testing"
  "reflect"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
)

func TestGetNewJobsUuids(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

  uuids, err := GetNewJobsUuids(Driver)
	utils.CheckError(err)

  expectedUuids := []string{
    "035a731b-9138-47d3-9f03-d7647186c7a1",
    "0c547d31-fd0d-4f41-a5a2-edd0aa7f0cd3",
    "1e1de355-0805-470a-9594-9d55d059cfc3",
    "1f9a3fd3-56d3-4d12-bddf-13f92965d94b",
    "43d8b125-802c-4398-b9f3-7b977d5f2251",
    "60cf4a1f-a26b-461c-a970-71fc14402904",
    "8a4f22aa-fb17-485a-933d-b426eeba0735",
    "8ae88e70-a48f-4625-9c0e-b42c237daa9d",
    "a2212936-3e04-486c-9a05-47a60c80971f",
    "a5f0dd5a-46fa-4a24-8244-7ddd849d91a4",
    "daaf4391-4496-4ea8-922f-9ce93af8c851",
    "eb11357d-aca8-44e1-9cca-eeb119003108",
    "f7475f6b-aa42-4869-b8a6-80156c6bed2f",
    "f8ba9d03-8633-40fc-b872-be209ad45369",
  }

  if !reflect.DeepEqual(uuids, expectedUuids) {
    t.Errorf("uuids not as expected!\nexpected: %v\nactual: %v", expectedUuids, uuids)
  }
}

func TestGetTestData(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_runner", "guts_runner")
	utils.CheckError(err)

  testUuid := "4ce9189f-561a-4886-aeef-1836f28b073b"

  repo, branch, plans, err := GetTestData(Driver, testUuid)
	utils.CheckError(err)

  expectedRepo := "https://github.com/canonical/ubuntu-gui-testing.git"
  expectedBranch := "main"
  expectedPlans := []string{
    "tests/firefox-example/plans/extended.yaml",
    "tests/firefox-example/plans/regular.yaml",
  }

  if repo != expectedRepo {
    t.Errorf("unexpected repo!\nexpected: %v\nactual: %v", expectedRepo, repo)
  }
  if branch != expectedBranch {
    t.Errorf("unexpected branch!\nexpected: %v\nactual: %v", expectedBranch, branch)
  }
  if !reflect.DeepEqual(plans, expectedPlans) {
    t.Errorf("unexpected plans!\nexpected: %v\nactual: %v", expectedPlans, plans)
  }
}
