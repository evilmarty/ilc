package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func loadConfig(args []string) (*Config, error) {
	filePath, err := getConfigPath(args)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading ilc file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling ilc file: %w", err)
	}

	cfg.FilePath = filePath
	return &cfg, nil
}

func getConfigPath(args []string) (string, error) {
	var filePath string
	fs := flag.NewFlagSet("ilc", flag.ContinueOnError)
	fs.StringVar(&filePath, "f", "", "Path to the ilc file")
	fs.Parse(args)

	if filePath != "" {
		return filePath, nil
	}

	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		if _, err := os.Stat(args[0]); err == nil {
			return args[0], nil
		}
	}

	if envPath := os.Getenv("ILC_FILE"); envPath != "" {
		return envPath, nil
	}

	if _, err := os.Stat("ilc.yml"); err == nil {
		return "ilc.yml", nil
	}
	if _, err := os.Stat("ilc.yaml"); err == nil {
		return "ilc.yaml", nil
	}

	return "", fmt.Errorf("no ilc file found")
}
