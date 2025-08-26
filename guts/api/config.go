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
}

func ParseConfig(filePath string) (GutsApiConfig, error) {
  var config GutsApiConfig
  filename, err := filepath.Abs(filePath)
  if err != nil {
    return config, err
  }
  yamlFile, err := os.ReadFile(filename)
  if err != nil {
    return config, err
  }
  err = yaml.Unmarshal(yamlFile, &config)
  return config, err
}
