package api

import (
	"gopkg.in/yaml.v3"
	"guts.ubuntu.com/v2/utils"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParseConfigSuccess(t *testing.T) {
	GutsCfg, _, _, err := Setup()
	utils.CheckError(err)
	var wanted GutsApiConfig
	wanted.Database.Driver = "postgres"
	wanted.Database.ConnectionString = "host=localhost port=5432 user=guts_api password=guts_api dbname=guts sslmode=disable"
	wanted.Api.Hostname = "localhost"
	wanted.Api.Port = 8080
	domains := []string{"launchpad.net", "localhost:9999"}
	wanted.Api.ArtifactDomains = domains
	domains = []string{"cdimage.ubuntu.com", "releases.ubuntu.com", "localhost:9999"}
	wanted.Api.TestbedDomains = domains
	domains = []string{"git.launchpad.net", "github.com"}
	wanted.Api.GitDomains = domains
	wanted.Tarball.TarballCachePath = "/srv/tarball-cache/"
	wanted.Tarball.TarballCacheMaxSize = 10737418240
	wanted.Tarball.TarballCacheReductionThreshold = 9663676416
	if !reflect.DeepEqual(GutsCfg, wanted) {
		t.Errorf("Parsed config not the same as wanted config!\nExpected:\n%v\nActual:\n%v", GutsCfg, wanted)
	}
}

func TestParseConfigFileNotFound(t *testing.T) {
	// _, _, _, err := Setup()
	_, err := ParseConfig("./guts-api-no-exist.yaml")
	var ExpectedType *fs.PathError

	if reflect.TypeOf(err) != reflect.TypeOf(ExpectedType) {
		t.Errorf("Error type not as expected!\nExpected: %v\nActual: %v", ExpectedType, reflect.TypeOf(err))
	}
	expected_string := "guts-api-no-exist.yaml: no such file or directory"
	if !strings.Contains(err.Error(), expected_string) {
		t.Errorf("Error string doesn't contain expected substring!\nExpected:\n%v\nActual:\n%v", expected_string, err.Error())
	}
}

func TestParseConfigYamlParsingFailure(t *testing.T) {
	f, err := os.CreateTemp(".", "not-a-yaml-file")
	utils.CheckError(err)
	defer utils.DeferredErrCheck(f.Close)
	defer utils.DeferredErrCheckStringArg(os.Remove, f.Name())

	data := []byte("This is not a yaml file.")
	_, err = f.Write(data)
	utils.CheckError(err)

	_, err = ParseConfig(f.Name())

	var ExpectedType *yaml.TypeError
	if reflect.TypeOf(err) != reflect.TypeOf(ExpectedType) {
		t.Errorf("Error type not as expected!\nExpected: %v\nActual: %v", reflect.TypeOf(ExpectedType), reflect.TypeOf(err))
	}
	expected_string := "cannot unmarshal !!str `This is...` into api.GutsApiConfig"
	if !strings.Contains(err.Error(), expected_string) {
		t.Errorf("Error string doesn't contain expected substring!\nExpected:\n%v\nActual:\n%v", expected_string, err.Error())
	}
}
