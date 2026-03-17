package scenario

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name               string                 `yaml:"name"`
	TargetURL          string                 `yaml:"target_url"`
	Tool               string                 `yaml:"tool"`
	Payload            map[string]interface{} `yaml:"payload"`
	AttackVector       string                 `yaml:"attack_vector"`
	IdempotencyKeyPath string                 `yaml:"idempotency_key_path"`
	AuthMutatePath     string                 `yaml:"auth_mutate_path"`
	AuthMutateValue    interface{}            `yaml:"auth_mutate_value"`
	DelayMs            int                    `yaml:"delay_ms"`
	RaceCount          int                    `yaml:"race_count"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}
	
	// Set defaults
	if cfg.RaceCount == 0 {
		cfg.RaceCount = 5
	}
	if cfg.DelayMs == 0 {
		cfg.DelayMs = 1000
	}

	return cfg, nil
}
