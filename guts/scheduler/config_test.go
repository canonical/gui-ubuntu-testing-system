package scheduler

import (
	"guts.ubuntu.com/v2/utils"
	"reflect"
	"testing"
)

func TestParseConfigLocal(t *testing.T) {
	cfgPath := "./guts-scheduler-local.yaml"
	schedulerCfg, err := ParseConfig(cfgPath)
	utils.CheckError(err)

	var expectedCfg GutsSchedulerConfig
	expectedCfg.Storage = make(map[string]string)
	expectedCfg.Storage["provider"] = "local"
	expectedCfg.Storage["object_path"] = "/srv/data/"
	expectedCfg.Storage["object_port"] = "9999"
	expectedCfg.Storage["object_host"] = "http://localhost"
	expectedCfg.Database.Driver = "postgres"
	expectedCfg.Database.ConnectionString = "host=localhost port=5432 user=guts_api password=guts_api dbname=guts sslmode=disable"
	expectedCfg.TestInactiveResetTime = "2 minutes"
	expectedCfg.ArtifactRetentionDays = 180

	if !reflect.DeepEqual(expectedCfg, schedulerCfg) {
		t.Errorf("unexpected parsed config!\nexpected: %v\nactual: %v", expectedCfg, schedulerCfg)
	}
}
