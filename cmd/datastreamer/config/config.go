package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	L2Url         string `yaml:"l2Url"`
	DatastreamUrl string `yaml:"datastreamUrl"`
	Version       int    `yaml:"version"`
	LogLevel      string `yaml:"logLevel"`
}

func GetConf(path string) (Config, error) {
	if path == "" {
		path = "datastreamerConfig.yaml"
	}
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	c := Config{}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
