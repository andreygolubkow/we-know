package areas

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Area struct {
	Title    string   `yaml:"title"`
	Patterns []string `yaml:"patterns"`
}

type Config map[string]Area

func ReadConfig(filepath string) (Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
