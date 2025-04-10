package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config Command

func (config Config) Select(args []string) Selection {
	selected := NewSelection(Command(config))
	return selected.Select(args)
}

func ParseConfig(content []byte) (Config, error) {
	var config Config

	if err := yaml.Unmarshal(content, &config); err != nil {
		return config, err
	}

	return config, nil
}

func LoadConfig(path string) (Config, error) {
	logger.Printf("Attempting to load config file: %s", path)
	if content, err := os.ReadFile(path); err != nil {
		return Config{}, err
	} else {
		return ParseConfig(content)
	}
}
