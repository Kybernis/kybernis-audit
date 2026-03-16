package scenario

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name           string `yaml:"name"`
	TargetEndpoint string `yaml:"target_endpoint"`
	FaultInjection string `yaml:"fault_injection"`
	Invariant      string `yaml:"invariant"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse scenario yaml: %w", err)
	}

	return &cfg, nil
}
