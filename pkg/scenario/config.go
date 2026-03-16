package scenario

import (
	"log"
)

type Config struct {
	Name           string `yaml:"name"`
	TargetEndpoint string `yaml:"target_endpoint"`
	FaultInjection string `yaml:"fault_injection"`
	Invariant      string `yaml:"invariant"`
}

func LoadConfig(path string) (*Config, error) {
	// TODO: Implement actual YAML unmarshalling here
	log.Printf("[Scenario] Loaded scenario from %s", path)

	// Mock returning a double-spend timeout injection
	return &Config{
		Name:           "Double Spend Timeout Test",
		TargetEndpoint: "/refund",
		FaultInjection: "timeout_after_success",
		Invariant:      "no_duplicate_mutations",
	}, nil
}
