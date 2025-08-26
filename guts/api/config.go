package main

import (
  "gopkg.in/yaml.v3"
  "os"
  "path/filepath"
)


type GutsApiConfig struct {
  Postgres struct {
    Host string `yaml:"hostname"`
    Port int `yaml:"port"`
    User string `yaml:"username"`
    Password string `yaml:"password"`
    DbName string `yaml:"dbname"`
  }
  Api struct {
    Hostname string `yaml:"hostname"`
    Port int `yaml:"port"`
  }
  Tarball struct {
    TarBallCachePath string `yaml:"tarball_cache_path"`
    TarBallCacheMaxSize int `yaml:"tarball_cache_max_size"`  // in bytes
    TarBallCacheReductionThreshold int `yaml:"tarball_cache_reduction_threshold"`  // in bytes
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
