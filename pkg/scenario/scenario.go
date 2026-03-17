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
	DelayMs            int                    `yaml:"delay_ms"`
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
	return cfg, nil
}
