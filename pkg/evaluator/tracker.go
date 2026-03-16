package evaluator

import (
	"bytes"
	"fmt"
	"log"
	"sync"
)

type Record struct {
	Method        string
	Path          string
	Body          []byte
	InjectedFault string
}

type Finding struct {
	Severity    string
	Message     string
	Details     string
	Remediation string
}

type Tracker struct {
	mu       sync.Mutex
	history  []Record
	findings []Finding
}

func NewTracker() *Tracker {
	return &Tracker{
		history:  make([]Record, 0),
		findings: make([]Finding, 0),
	}
}

func (t *Tracker) RecordAndEvaluate(method, path string, body []byte, fault string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if this is a retry of a previously failed/faulted request
	for _, past := range t.history {
		if past.Method == method && past.Path == path && past.InjectedFault != "" {
			log.Printf("🔍 Evaluating retry on %s %s...", method, path)

			if !bytes.Equal(past.Body, body) {
				finding := Finding{
					Severity:    "CRITICAL",
					Message:     "SEMANTIC DRIFT DETECTED",
					Details:     fmt.Sprintf("Agent retried mutation %s %s but the payload changed!\nOriginal: %s\nRetried:  %s", method, path, string(past.Body), string(body)),
					Remediation: "Standard idempotency keys will fail. Use a deterministic execution guard (e.g. Kybernis Cloud).",
				}
				t.findings = append(t.findings, finding)
			} else {
				finding := Finding{
					Severity:    "WARNING",
					Message:     "BLIND RETRY DETECTED",
					Details:     fmt.Sprintf("Agent retried exact same payload on %s %s after a network timeout.", method, path),
					Remediation: "Ensure your backend has strict, perfectly deterministic idempotency keys for this endpoint.",
				}
				t.findings = append(t.findings, finding)
			}
			return
		}
	}

	t.history = append(t.history, Record{
		Method:        method,
		Path:          path,
		Body:          body,
		InjectedFault: fault,
	})
}

func (t *Tracker) PrintSummary() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.findings) == 0 {
		fmt.Println("✅ No execution safety vulnerabilities found in this scenario.")
		return
	}

	for i, f := range t.findings {
		icon := "⚠️"
		if f.Severity == "CRITICAL" {
			icon = "❌"
		}
		fmt.Printf("\n%s Finding %d: [%s] %s\n", icon, i+1, f.Severity, f.Message)
		fmt.Printf("   Details: %s\n", f.Details)
		fmt.Printf("   Fix:     %s\n", f.Remediation)
	}
	fmt.Println()
}
