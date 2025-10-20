package runner

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type GutsRunnerConfig struct {
	Storage  map[string]string `yaml:"storage"`
	Database struct {
		Driver           string `yaml:"driver"`
		ConnectionString string `yaml:"connection_string"`
	}
}

func ParseConfig(cfgPath string) (GutsRunnerConfig, error) {
	var RunnerCfg GutsRunnerConfig
	filename, err := filepath.Abs(cfgPath)
	if err != nil { // coverage-ignore
		return RunnerCfg, err
	}
	yamlFile, err := os.ReadFile(filename)
	if err != nil { // coverage-ignore
		return RunnerCfg, err
	}
	err = yaml.Unmarshal(yamlFile, &RunnerCfg)
	return RunnerCfg, err
}
