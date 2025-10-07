package runner

import (
	"guts.ubuntu.com/v2/utils"
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	var DummyCfg GutsRunnerConfig
	DummyCfg.Database.Driver = "postgres"
	DummyCfg.Database.ConnectionString = "host=localhost port=5432 user=guts_api password=guts_api dbname=guts sslmode=disable"
	DummyCfg.Storage = make(map[string]string)
	DummyCfg.Storage["provider"] = "local"
	DummyCfg.Storage["object_path"] = "/srv/data/"
	DummyCfg.Storage["object_port"] = "9999"
	DummyCfg.Storage["object_host"] = "http://localhost"

	cfgPath := "./guts-runner-local.yaml"
	accCfg, err := ParseConfig(cfgPath)
	utils.CheckError(err)

	if !reflect.DeepEqual(accCfg, DummyCfg) {
		t.Errorf("unexpected config struct!\nexpected: %v\nactual: %v", DummyCfg, accCfg)
	}
}
