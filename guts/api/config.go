package main

import (
  "gopkg.in/yaml.v3"
  "os"
  "path/filepath"
)


type GutsApiConfig struct {
  Database struct {
    Driver string `yaml:"driver"`
    ConnectionString string `yaml:"connection_string"`
  }
  Api struct {
    Hostname string `yaml:"hostname"`
    Port int `yaml:"port"`
  }
  Tarball struct {
    TarballCachePath string `yaml:"tarball_cache_path"`
    TarballCacheMaxSize int `yaml:"tarball_cache_max_size"`  // in bytes
    TarballCacheReductionThreshold int `yaml:"tarball_cache_reduction_threshold"`  // in bytes
  }
}

var (
  GutsCfg GutsApiConfig
)

func ParseConfig(filePath string) error {
  filename, err := filepath.Abs(filePath)
  if err != nil {
    return err
  }
  yamlFile, err := os.ReadFile(filename)
  if err != nil {
    return err
  }
  err = yaml.Unmarshal(yamlFile, &GutsCfg)
  return err
}
