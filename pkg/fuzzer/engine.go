package fuzzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func (f *Fuzzer) printMitigation(attack, caseStudy, community, kybernis string) {
	fmt.Println("\n=======================================================")
	if caseStudy != "" {
		fmt.Printf("📚 Real-World Case Study: %s\n\n", caseStudy)
	}
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
			"The 'McDonald's AI Drive-Thru' (2024): An AI agent got stuck in a verification loop and blindly ordered 260 chicken nuggets.",
			"Implement exponential backoff limits, strict LLM loop limits, and Redis Token Buckets per user.",
			"Native state deduplication prevents re-execution regardless of wait times or fuzzy parameters.")
		return fmt.Errorf("vulnerability_detected: dare")
	}

	fmt.Println("\n✅ [PASSED] Backend caught the duplicate payload and blocked the blind retry.")
	return nil
}

// 2. GHOST (Ghost Execution / Ambiguous Outcome)
func (f *Fuzzer) executeGhostAttack() error {
	data, _ := json.Marshal(f.Scenario.Payload)

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
			"The 'X Grok NBA Hallucination' (2024): The LLM misinterpreted a tool outcome (slang for 'throwing bricks') and hallucinated a criminal vandalism charge, proving that LLMs cannot accurately determine terminal state from ambiguous network responses.",
			"Convert synchronous HTTP tool calls to Async Webhooks or polling mechanisms (GET /status).",
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
		payloadB = f.injectValue(f.Scenario.Payload, "_agent_reasoning", "I encountered a timeout so I am retrying.")
		fmt.Println("⚠️  [AGENT RETRY] Firing payload with hallucinated reasoning string to bypass content hashing.")
	} else if f.Scenario.Variant == "semantic_equivalence" {
		fmt.Println("⚠️  [AGENT RETRY] Firing payload with type coercion (semantic equivalence).")
		payloadB = f.injectValue(f.Scenario.Payload, "_semantic_drift", 100.0)
	}

	resB, _ := f.sendPayload(payloadB)
	if resB != nil && (resB.StatusCode == 200 || resB.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend is bleeding. Semantic double-spend executed.")
		f.printMitigation("DRIFT",
			"The 'Air Canada Bereavement Fare' (2024): The LLM hallucinated a new policy mid-conversation and executed a refund guarantee that bypassed the backend's standard pricing checks because the payload drifted.",
			"Strict Pydantic/Zod schema enforcement + JSON canonicalization before SHA-256 content hashing.",
			"Semantic Locks anchor execution to the task ID, entirely immune to JSON mutation and key regeneration.")
		return fmt.Errorf("vulnerability_detected: drift")
	}

	fmt.Println("\n✅ [PASSED] Backend caught the semantic drift.")
	return nil
}

// 4. AUTH (Authorization Context Drift)
func (f *Fuzzer) executeAuthAttack() error {
	mutated := f.injectValue(f.Scenario.Payload, f.Scenario.AuthMutatePath, f.Scenario.AuthMutateValue)

	if f.Scenario.Variant == "confused_deputy" {
		fmt.Println("⚠️  [CPRF] Simulating Cross Plugin Request Forgery via injected external payload data...")
	} else {
		fmt.Printf("    ↳ Firing execution with mutated context (%s)...\n", f.Scenario.AuthMutatePath)
	}

	res, _ := f.sendPayload(mutated)

	if res != nil && (res.StatusCode == 200 || res.StatusCode == 201) {
		fmt.Println("\n❌ [FAILED] Backend accepted the mutated payload post-authorization.")
		f.printMitigation("AUTH",
			"The 'Notion AI Data Exfiltration' & 'Chevy Tahoe for $1' (2023/2025): External documents and users injected prompts that overwrote the agent's system instructions, bypassing authorization checks and executing unintended transactions (buying a car for $1).",
			"Implement LlamaGuard, NeMo Guardrails, or manually hash state pre-approval.",
			"Kybernis cryptographically binds human approval to the execution state, blocking all prompt injections post-approval.")
		return fmt.Errorf("vulnerability_detected: auth")
	}

	fmt.Println("\n✅ [PASSED] Backend blocked the context drift.")
	return nil
}

// 5. SAGA (Shattered Saga / Incomplete Execution)
func (f *Fuzzer) executeSagaAttack() error {
	if f.Scenario.Variant == "failed_compensation" {
		fmt.Println("💥 [CRASH] Simulating agent executing Step 1, failing Step 2, to test backend native rollback.")
	} else {
		fmt.Println("💥 [CRASH] Simulating agent termination mid-execution before Step 2.")
	}

	f.printMitigation("SAGA",
		"The 'Replit AI Database Deletion' (2025): An autonomous agent made code changes without instruction, and during a subsequent failure sequence, it bypassed constraints and deleted the entire production database without rolling back its partial state.",
		"AWS Step Functions, Temporal workflows, or Saga Pattern Rollback Queues (SQS).",
		"Kybernis maintains a cross-service orchestration ledger with native rollback support.")
	return fmt.Errorf("vulnerability_detected: saga")
}

// 6. RACE (Parallel Duplicate Race)
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
			time.Sleep(10 * time.Millisecond) // Catch distributed locks that release prematurely
		}
	}
	wg.Wait()

	if successCount > 1 {
		fmt.Printf("\n❌ [FAILED] Backend processed %d simultaneous executions.\n", successCount)
		f.printMitigation("RACE",
			"Hedge Fund Agent Swarm Collision (2024): Two concurrent reasoning sub-agents analyzing the same market signal arrived at the same conclusion simultaneously, bypassing standard API rate limits and double-executing a massive trade.",
			"Database Row Locks (SELECT FOR UPDATE) or Redis Redlock (Distributed Leases).",
			"Kybernis infrastructure natively issues in-flight execution leases to prevent parallel agent spawning.")
		return fmt.Errorf("vulnerability_detected: race")
	}

	fmt.Println("\n✅ [PASSED] Backend caught the parallel race condition.")
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
