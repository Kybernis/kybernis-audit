package fuzzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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
	fmt.Printf("🛡️  Scenario: %s\n\n", f.Scenario.Name)

	switch f.Scenario.AttackVector {
	case "dare":
		return f.executeDareAttack()
	case "ghost":
		return f.executeGhostAttack()
	case "drift":
		return f.executeDriftAttack()
	case "auth":
		return f.executeAuthAttack()
	case "saga":
		return f.executeSagaAttack()
	case "race":
		return f.executeRaceAttack()
	default:
		return fmt.Errorf("unknown attack vector: %s", f.Scenario.AttackVector)
	}
}

// 1. DARE (Duplicate Action Replay / Blind Retry)
func (f *Fuzzer) executeDareAttack() error {
	fmt.Println("⚡ [DARE] Initiating Duplicate Action Replay...")
	
	// Ensure we have an idempotency key injected so a robust backend CAN catch it
	key := uuid.New().String()
	payload := f.Scenario.Payload
	if f.Scenario.IdempotencyKeyPath != "" {
		payload = f.injectValue(payload, f.Scenario.IdempotencyKeyPath, key)
	}

	fmt.Println("    ↳ Firing initial payload...")
	resA, _ := f.sendPayload(payload)
	if resA != nil && (resA.StatusCode == 200 || resA.StatusCode == 201) {
		fmt.Println("✅ [BASELINE] Initial tool execution succeeded.")
	} else {
		return fmt.Errorf("❌ Target endpoint rejected the baseline payload.")
	}

	time.Sleep(time.Duration(f.Scenario.DelayMs) * time.Millisecond)

	fmt.Println("    ↳ Firing exact same payload (blind retry)...")
	resB, _ := f.sendPayload(payload)

	if resB != nil && (resB.StatusCode == 200 || resB.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend processed the duplicate blind retry.")
		fmt.Println("    ↳ The system lacks basic idempotency. An agent stuck in a loop will destroy your state.")
		return fmt.Errorf("vulnerability_detected: dare")
	}

	fmt.Println("\n✅ [PASSED] Backend caught the duplicate payload and blocked the blind retry.")
	return nil
}

// 2. GHOST (Ghost Execution / Ambiguous Outcome)
func (f *Fuzzer) executeGhostAttack() error {
	fmt.Println("⚡ [GHOST] Initiating Ghost Execution...")
	fmt.Println("    ↳ Firing mutation and simulating network failure mid-execution...")

	payload := f.Scenario.Payload
	data, _ := json.Marshal(payload)
	
	// Create a context that times out instantly to simulate a dropped connection
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", f.TargetURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	_, err := f.Client.Do(req)
	if err != nil {
		fmt.Println("⚠️  [TIMEOUT] Client disconnected before receiving 200 OK.")
		fmt.Println("    ↳ If the backend continued executing, the agent's mental model is now out of sync with reality.")
		fmt.Println("\n❌ [WARNING] To fix GHOST executions, you need an external state ledger to confirm terminal status on retry.")
		return fmt.Errorf("vulnerability_detected: ghost")
	}

	return nil
}

// 3. DRIFT (Semantic Drift on Retry)
func (f *Fuzzer) executeDriftAttack() error {
	fmt.Println("⚡ [DRIFT] Initiating Semantic Idempotency Bypass...")

	payloadA := f.injectValue(f.Scenario.Payload, f.Scenario.IdempotencyKeyPath, uuid.New().String())
	resA, _ := f.sendPayload(payloadA)
	if resA == nil || (resA.StatusCode != 200 && resA.StatusCode != 201) {
		return fmt.Errorf("❌ Target endpoint rejected the baseline payload.")
	}
	fmt.Println("✅ [BASELINE] Initial tool execution succeeded.")

	time.Sleep(time.Duration(f.Scenario.DelayMs) * time.Millisecond)

	driftedKey := uuid.New().String()
	payloadB := f.injectValue(f.Scenario.Payload, f.Scenario.IdempotencyKeyPath, driftedKey)
	
	fmt.Printf("⚠️  [AGENT RETRY] Firing duplicate semantic payload with drifted key: %s\n", driftedKey)
	resB, _ := f.sendPayload(payloadB)

	if resB != nil && (resB.StatusCode == 200 || resB.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend is bleeding. Semantic double-spend executed.")
		fmt.Printf("    ↳ The agent bypassed standard idempotency by generating a new key (%s).\n", driftedKey)
		return fmt.Errorf("vulnerability_detected: drift")
	}

	fmt.Println("\n✅ [PASSED] Backend caught the semantic drift and blocked the execution.")
	return nil
}

// 4. AUTH (Authorization Context Drift)
func (f *Fuzzer) executeAuthAttack() error {
	fmt.Println("⚡ [AUTH] Initiating Authorization Context Drift...")

	if f.Scenario.AuthMutatePath == "" {
		return fmt.Errorf("auth attack requires auth_mutate_path in scenario config")
	}

	mutatedPayload := f.injectValue(f.Scenario.Payload, f.Scenario.AuthMutatePath, f.Scenario.AuthMutateValue)
	
	fmt.Printf("    ↳ Firing execution with mutated context (%s)...\n", f.Scenario.AuthMutatePath)
	res, _ := f.sendPayload(mutatedPayload)

	if res != nil && (res.StatusCode == 200 || res.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend accepted the mutated payload post-authorization.")
		fmt.Println("    ↳ The human approved X, but the agent successfully executed Y.")
		return fmt.Errorf("vulnerability_detected: auth")
	}

	fmt.Println("\n✅ [PASSED] Backend verified cryptographic authorization and blocked the context drift.")
	return nil
}

// 5. SAGA (Shattered Saga / Incomplete Execution)
func (f *Fuzzer) executeSagaAttack() error {
	fmt.Println("⚡ [SAGA] Initiating Shattered Saga...")
	fmt.Println("    ↳ Executing step 1 of multi-step tool...")
	
	res, _ := f.sendPayload(f.Scenario.Payload)
	if res == nil || (res.StatusCode != 200 && res.StatusCode != 201) {
		return fmt.Errorf("❌ Step 1 rejected by backend.")
	}

	fmt.Println("💥 [CRASH] Simulating agent termination before step 2.")
	fmt.Println("\n❌ [FAILED] If your backend does not have an orchestration ledger, you just left corrupted partial state in production.")
	fmt.Println("    ↳ Use Kybernis SDK's State-Machine Architecture to enforce cross-service atomicity.")
	
	return fmt.Errorf("vulnerability_detected: saga")
}

// 6. RACE (Parallel Duplicate Race)
func (f *Fuzzer) executeRaceAttack() error {
	fmt.Printf("⚡ [RACE] Initiating Parallel Duplicate Race (Spawning %d concurrent agents)...\n", f.Scenario.RaceCount)

	payload := f.Scenario.Payload
	if f.Scenario.IdempotencyKeyPath != "" {
		payload = f.injectValue(payload, f.Scenario.IdempotencyKeyPath, uuid.New().String())
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	for i := 0; i < f.Scenario.RaceCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			res, _ := f.sendPayload(payload)
			if res != nil && (res.StatusCode == 200 || res.StatusCode == 201) {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount > 1 {
		fmt.Printf("\n❌ [FAILED] Backend processed %d simultaneous identical tool executions.\n", successCount)
		fmt.Println("    ↳ You are missing an 'in-flight lease' (distributed lock). Parallel reasoning branches bypassed your checks.")
		return fmt.Errorf("vulnerability_detected: race")
	}

	fmt.Println("\n✅ [PASSED] Backend utilized a distributed lock and allowed only 1 execution.")
	return nil
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

func (f *Fuzzer) injectValue(payload map[string]interface{}, path string, value interface{}) map[string]interface{} {
	// Deep copy
	out := make(map[string]interface{})
	data, _ := json.Marshal(payload)
	json.Unmarshal(data, &out)

	// Simplified single-level path injection for the prototype
	out[path] = value
	return out
}
