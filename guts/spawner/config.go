package spawner

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
		Memory int `yaml:"memory"`
		Cores  int `yaml:"cores"`
	}
	General struct {
		ImageCachePath string `yaml:"image_cache_path"`
	}
}

func ParseConfig(filePath string) (GutsSpawnerConfig, error) {
	var SpawnerCfg GutsSpawnerConfig
	filename, err := filepath.Abs(filePath)
	if err != nil { // coverage-ignore
		return SpawnerCfg, err
	}
	yamlFile, err := os.ReadFile(filename)
	if err != nil { // coverage-ignore
		return SpawnerCfg, err
	}
	err = yaml.Unmarshal(yamlFile, &SpawnerCfg)
	return SpawnerCfg, err
}
