package ilc

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config Command

func (config *Config) UnmarshalYAML(node *yaml.Node) error {
	type tempConfig Config
	var temp tempConfig
	if err := node.Decode(&temp); err != nil {
		return err
	}
	*config = Config(temp)
	config.Description = strings.TrimSpace(config.Description)
	return nil
}

func (config Config) Select(args []string) Selection {
	selected := NewSelection(Command(config))
	return selected.Select(args)
}

func (config Config) Validate() error {
	return Command(config).Validate()
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
