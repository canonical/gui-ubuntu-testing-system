package main

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type GutsSpawnerConfig struct {
	Database struct {
		Driver           string `yaml:"driver"`
		ConnectionString string `yaml:"connection_string"`
	}
  Virtualisation struct {
    Memory        int `yaml:"memory"`
    Cores         int `yaml:"cores"`
  }
  General struct {
    ImageCachePath    int `yaml:"image_cache_path"`
  }
}

var (
	SpawnerCfg GutsSpawnerConfig
)

func ParseConfig(filePath string) error {
	filename, err := filepath.Abs(filePath)
	if err != nil { // coverage-ignore
		return err
	}
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &SpawnerCfg)
	return err
}
