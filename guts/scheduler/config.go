package scheduler

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type GutsSchedulerConfig struct {
	Storage  map[string]string `yaml:"storage"`
	Database struct {
		Driver           string `yaml:"driver"`
		ConnectionString string `yaml:"connection_string"`
	}
	TestInactiveResetTime string `yaml:"test_inactive_reset_time"` // like '2 minutes'
	ArtifactRetentionDays int    `yaml:"artifact_retention_days"`
}

func ParseConfig(cfgPath string) (GutsSchedulerConfig, error) {
	var SchedulerConfig GutsSchedulerConfig
	filename, err := filepath.Abs(cfgPath)
	if err != nil { // coverage-ignore
		return SchedulerConfig, err
	}
	yamlFile, err := os.ReadFile(filename)
	if err != nil { // coverage-ignore
		return SchedulerConfig, err
	}
	err = yaml.Unmarshal(yamlFile, &SchedulerConfig)
	return SchedulerConfig, err
}
