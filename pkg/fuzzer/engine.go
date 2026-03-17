package fuzzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kybernis/kybernis-audit/pkg/scenario"
)

type Fuzzer struct {
	TargetURL string
	Scenario  scenario.Config
	Client    *http.Client
}

func NewFuzzer(target string, cfg scenario.Config) *Fuzzer {
	return &Fuzzer{
		TargetURL: target,
		Scenario:  cfg,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *Fuzzer) Run() error {
	fmt.Printf("🔥 Starting Deterministic Agent Fuzzer against %s\n", f.TargetURL)
	fmt.Printf("🛡️  Scenario: %s\n", f.Scenario.Name)

	switch f.Scenario.AttackVector {
	case "drift":
		return f.executeDriftAttack()
	case "dare":
		return f.executeDareAttack()
	case "ghost":
		return f.executeGhostAttack()
	case "race":
		return f.executeRaceAttack()
	default:
		return fmt.Errorf("unknown attack vector: %s", f.Scenario.AttackVector)
	}
}

// executeDriftAttack simulates an LLM hallucinating a new idempotency key on retry.
func (f *Fuzzer) executeDriftAttack() error {
	fmt.Println("⚡ [DRIFT] Initiating Semantic Idempotency Bypass...")

	// 1. Prepare initial payload
	payloadA := f.injectKey(f.Scenario.Payload, f.Scenario.IdempotencyKeyPath, uuid.New().String())

	// 2. Fire Request A
	resA, err := f.sendPayload(payloadA)
	if err != nil || (resA.StatusCode != 200 && resA.StatusCode != 201) {
		return fmt.Errorf("❌ Target endpoint rejected the baseline payload. Ensure the endpoint is active.")
	}
	fmt.Println("✅ [BASELINE] Initial tool execution succeeded.")

	// 3. Simulate Agent Network Timeout / Replanning phase
	time.Sleep(time.Duration(f.Scenario.DelayMs) * time.Millisecond)

	// 4. Fire Request B (The Drift)
	driftedKey := uuid.New().String()
	payloadB := f.injectKey(f.Scenario.Payload, f.Scenario.IdempotencyKeyPath, driftedKey)
	
	fmt.Printf("⚠️  [AGENT RETRY] Firing duplicate semantic payload with drifted key: %s\n", driftedKey)
	resB, err := f.sendPayload(payloadB)
	if err != nil {
		return err
	}

	// 5. Evaluate backend resilience
	if resB.StatusCode == 200 || resB.StatusCode == 201 {
		fmt.Println("\n❌ [FAILED] Backend is bleeding. Semantic double-spend executed.")
		fmt.Printf("    ↳ The agent bypassed standard idempotency by generating a new key (%s).\n", driftedKey)
		fmt.Println("    ↳ Fix this by implementing a persistent State-Machine Architecture (Kybernis SDK).")
		return fmt.Errorf("vulnerability_detected: drift")
	}

	fmt.Println("\n✅ [PASSED] Backend caught the semantic drift and blocked the execution.")
	return nil
}

// TODO: Implement Dare, Ghost, Race

func (f *Fuzzer) executeDareAttack() error {
	fmt.Println("⚡ [DARE] Initiating Duplicate Action Replay...")
	return fmt.Errorf("not implemented")
}

func (f *Fuzzer) executeGhostAttack() error {
	fmt.Println("⚡ [GHOST] Initiating Ghost Execution...")
	return fmt.Errorf("not implemented")
}

func (f *Fuzzer) executeRaceAttack() error {
	fmt.Println("⚡ [RACE] Initiating Parallel Duplicate Race...")
	return fmt.Errorf("not implemented")
}

func (f *Fuzzer) sendPayload(payload map[string]interface{}) (*http.Response, error) {
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", f.TargetURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return f.Client.Do(req)
}

func (f *Fuzzer) injectKey(payload map[string]interface{}, path string, key string) map[string]interface{} {
	// Deep copy
	out := make(map[string]interface{})
	data, _ := json.Marshal(payload)
	json.Unmarshal(data, &out)

	// Simplified single-level path injection for the prototype
	out[path] = key
	return out
}
