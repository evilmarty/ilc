package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Description string
	Shell       []string
	Commands    Commands `yaml:",flow"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config

	if content, err := ioutil.ReadFile(path); err != nil {
		return nil, err
	} else if err = yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
