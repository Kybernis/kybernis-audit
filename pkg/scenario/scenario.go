package scenario

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name      string   `yaml:"name"`
	TargetURL string   `yaml:"target_url"`
	Scenarios []Config `yaml:"scenarios"`
}

type Config struct {
	Name               string                 `yaml:"name"`
	TargetURL          string                 `yaml:"target_url"`
	Tool               string                 `yaml:"tool"`
	Payload            map[string]interface{} `yaml:"payload"`
	AttackVector       string                 `yaml:"attack_vector"`
	Variant            string                 `yaml:"variant"`
	IdempotencyKeyPath string                 `yaml:"idempotency_key_path"`
	AuthMutatePath     string                 `yaml:"auth_mutate_path"`
	AuthMutateValue    interface{}            `yaml:"auth_mutate_value"`
	DelayMs            int                    `yaml:"delay_ms"`
	RaceCount          int                    `yaml:"race_count"`
}

func LoadConfig(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}

	var m Manifest
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return Manifest{}, err
	}
	
	// Apply global TargetURL and defaults to each scenario
	for i := range m.Scenarios {
		if m.Scenarios[i].TargetURL == "" {
			m.Scenarios[i].TargetURL = m.TargetURL
		}
		if m.Scenarios[i].RaceCount == 0 {
			m.Scenarios[i].RaceCount = 5
		}
		if m.Scenarios[i].DelayMs == 0 {
			m.Scenarios[i].DelayMs = 1000
		}
		if m.Scenarios[i].Variant == "" {
			m.Scenarios[i].Variant = "standard"
		}
	}

	return m, nil
}
