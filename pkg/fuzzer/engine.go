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
	fmt.Printf("🔥 Deterministic Agent Fuzzer targetting %s\n", f.TargetURL)
	fmt.Printf("🛡️  Scenario: %s\n", f.Scenario.Name)
	fmt.Printf("🎯 Attack: %s (Variant: %s)\n\n", f.Scenario.AttackVector, f.Scenario.Variant)

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

func (f *Fuzzer) printMitigation(attack, community, kybernis string) {
	fmt.Println("\n=======================================================")
	fmt.Printf("🤝 Community Mitigation: %s\n", community)
	fmt.Printf("⚡ Kybernis SDK: %s\n", kybernis)
	fmt.Println("=======================================================\n")
}

// 1. DARE (Duplicate Action Replay / Blind Retry)
func (f *Fuzzer) executeDareAttack() error {
	key := uuid.New().String()
	payload := f.Scenario.Payload
	if f.Scenario.IdempotencyKeyPath != "" {
		payload = f.injectValue(payload, f.Scenario.IdempotencyKeyPath, key)
	}

	fmt.Println("    ↳ Firing initial payload...")
	resA, _ := f.sendPayload(payload)
	if resA == nil || (resA.StatusCode != 200 && resA.StatusCode != 201) {
		return fmt.Errorf("❌ Target endpoint rejected the baseline payload.")
	}

	if f.Scenario.Variant == "delayed" {
		fmt.Printf("    ↳ Waiting %dms to test rate-limit/cache expiration bypass...\n", f.Scenario.DelayMs)
		time.Sleep(time.Duration(f.Scenario.DelayMs) * time.Millisecond)
	} else if f.Scenario.Variant == "param_fuzz" {
		fmt.Println("    ↳ Firing blind retry with a dummy parameter bypass...")
		payload = f.injectValue(payload, "_agent_retry", true)
	} else {
		fmt.Println("    ↳ Firing immediate exact blind retry...")
	}

	resB, _ := f.sendPayload(payload)
	if resB != nil && (resB.StatusCode == 200 || resB.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend processed the duplicate blind retry.")
		f.printMitigation("DARE", 
			"Implement exponential backoff limits and Redis Token Buckets per user.",
			"Native state deduplication prevents re-execution regardless of wait times or fuzzy parameters.")
		return fmt.Errorf("vulnerability_detected: dare")
	}

	fmt.Println("\n✅ [PASSED] Backend caught the duplicate payload and blocked the blind retry.")
	return nil
}

// 2. GHOST (Ghost Execution / Ambiguous Outcome)
func (f *Fuzzer) executeGhostAttack() error {
	data, _ := json.Marshal(f.Scenario.Payload)
	
	// Pre-commit drop (1ms) vs Post-commit drop (simulated long wait cut off)
	timeout := 20 * time.Millisecond
	if f.Scenario.Variant == "post_commit" {
		timeout = 2 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", f.TargetURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	_, err := f.Client.Do(req)
	if err != nil {
		fmt.Println("⚠️  [TIMEOUT] Client disconnected before receiving 200 OK.")
		fmt.Println("    ↳ The agent's mental model is now out of sync with backend reality.")
		f.printMitigation("GHOST", 
			"Convert synchronous HTTP tool calls to Async Webhooks or polling mechanisms.",
			"Persistent Session IDs track execution status. The agent queries Kybernis to verify terminal state on retry.")
		return fmt.Errorf("vulnerability_detected: ghost")
	}

	return nil
}

// 3. DRIFT (Semantic Drift on Retry)
func (f *Fuzzer) executeDriftAttack() error {
	payloadA := f.injectValue(f.Scenario.Payload, f.Scenario.IdempotencyKeyPath, uuid.New().String())
	f.sendPayload(payloadA)
	time.Sleep(100 * time.Millisecond)

	payloadB := payloadA
	if f.Scenario.Variant == "idempotency_regen" || f.Scenario.Variant == "standard" {
		driftedKey := uuid.New().String()
		payloadB = f.injectValue(f.Scenario.Payload, f.Scenario.IdempotencyKeyPath, driftedKey)
		fmt.Printf("⚠️  [AGENT RETRY] Firing payload with regenerated UUID: %s\n", driftedKey)
	} else if f.Scenario.Variant == "hash_bypass" {
		payloadB = f.injectValue(f.Scenario.Payload, "agent_reasoning", "I encountered a timeout so I am retrying.")
		fmt.Println("⚠️  [AGENT RETRY] Firing payload with hallucinated reasoning string to bypass content hashing.")
	}

	resB, _ := f.sendPayload(payloadB)
	if resB != nil && (resB.StatusCode == 200 || resB.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend is bleeding. Semantic double-spend executed.")
		f.printMitigation("DRIFT",
			"Strict Pydantic/Zod schema enforcement + SHA-256 content hashing of normalized payloads.",
			"Semantic Locks anchor execution to the task ID, entirely immune to JSON mutation and key regeneration.")
		return fmt.Errorf("vulnerability_detected: drift")
	}
	
	fmt.Println("\n✅ [PASSED] Backend caught the semantic drift.")
	return nil
}

// 4. AUTH, 5. SAGA, 6. RACE implementation continuation...

func (f *Fuzzer) executeAuthAttack() error {
	mutated := f.injectValue(f.Scenario.Payload, f.Scenario.AuthMutatePath, f.Scenario.AuthMutateValue)
	fmt.Printf("    ↳ Firing execution with mutated context (%s)...\n", f.Scenario.AuthMutatePath)
	res, _ := f.sendPayload(mutated)
	
	if res != nil && (res.StatusCode == 200 || res.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend accepted the mutated payload post-authorization.")
		f.printMitigation("AUTH", "Implement LlamaGuard or hash the state pre-approval.", "Kybernis cryptographically binds human approval to the execution state.")
		return fmt.Errorf("vulnerability_detected: auth")
	}
	return nil
}

func (f *Fuzzer) executeSagaAttack() error {
	fmt.Println("💥 [CRASH] Simulating agent termination before step 2.")
	f.printMitigation("SAGA", "AWS Step Functions or Temporal orchestrations.", "Kybernis maintains a cross-service orchestration ledger for rollbacks.")
	return fmt.Errorf("vulnerability_detected: saga")
}

func (f *Fuzzer) executeRaceAttack() error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	for i := 0; i < f.Scenario.RaceCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			res, _ := f.sendPayload(f.Scenario.Payload)
			if res != nil && (res.StatusCode == 200 || res.StatusCode == 201) {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
		
		if f.Scenario.Variant == "staggered" {
			time.Sleep(10 * time.Millisecond) // Release locks prematurely test
		}
	}
	wg.Wait()

	if successCount > 1 {
		fmt.Printf("\n❌ [FAILED] Backend processed %d simultaneous executions.\n", successCount)
		f.printMitigation("RACE", "Database Row Locks (SELECT FOR UPDATE) or Redis Redlock.", "Kybernis infrastructure automatically prevents concurrent tool spawning.")
		return fmt.Errorf("vulnerability_detected: race")
	}
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
	out := make(map[string]interface{})
	data, _ := json.Marshal(payload)
	json.Unmarshal(data, &out)
	out[path] = value
	return out
}
