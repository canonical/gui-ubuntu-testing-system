package api

import (
	"gopkg.in/yaml.v3"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"os"
	"path/filepath"
)

type GutsApiConfig struct {
	Database struct {
		Driver           string `yaml:"driver"`
		ConnectionString string `yaml:"connection_string"`
	}
	Api struct {
		Hostname        string   `yaml:"hostname"`
		Port            int      `yaml:"port"`
		ArtifactDomains []string `yaml:"artifact_domains"`
		TestbedDomains  []string `yaml:"testbed_domains"`
		GitDomains      []string `yaml:"git_domains"`
	}
	Tarball struct {
		TarballCachePath               string `yaml:"tarball_cache_path"`
		TarballCacheMaxSize            int    `yaml:"tarball_cache_max_size"`            // in bytes
		TarballCacheReductionThreshold int    `yaml:"tarball_cache_reduction_threshold"` // in bytes
	}
}

func Setup() (GutsApiConfig, database.DbDriver, ApiArgs, error) {
	args := ParseArgs()
	gutsCfg, err := ParseConfig(args.ConfigFilePath)
	utils.CheckError(err)
	driver, err := database.NewDbDriver(gutsCfg.Database.Driver, gutsCfg.Database.ConnectionString)
	utils.CheckError(err)
	return gutsCfg, driver, args, err
}

func ParseConfig(filePath string) (GutsApiConfig, error) {
	var GutsCfg GutsApiConfig
	filename, err := filepath.Abs(filePath)
	if err != nil { // coverage-ignore
		return GutsCfg, err
	}
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return GutsCfg, err
	}
	err = yaml.Unmarshal(yamlFile, &GutsCfg)
	return GutsCfg, err
}
