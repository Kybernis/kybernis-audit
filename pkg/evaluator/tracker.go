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

type Tracker struct {
	mu      sync.Mutex
	history []Record
}

func NewTracker() *Tracker {
	return &Tracker{
		history: make([]Record, 0),
	}
}

func (t *Tracker) RecordAndEvaluate(method, path string, body []byte, fault string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if this is a retry of a previously failed/faulted request
	for _, past := range t.history {
		if past.Method == method && past.Path == path && past.InjectedFault != "" {
			// This is a retry of a faulted endpoint!
			log.Printf("🔍 Evaluating retry on %s %s...", method, path)

			if !bytes.Equal(past.Body, body) {
				fmt.Println("\n❌ [KYBERNIS AUDIT FINDING] SEMANTIC DRIFT DETECTED ❌")
				fmt.Println("   The agent experienced a failure and retried the mutation, but the payload changed!")
				fmt.Println("   Original Payload: ", string(past.Body))
				fmt.Printf("   Retried Payload:  %s\n", string(body))
				fmt.Println("   Vulnerability:    Standard backend idempotency keys will fail. This is a potential double-spend.")
				fmt.Println("   Fix:              Use a deterministic execution guard (e.g. Kybernis Cloud).")
				fmt.Println()
			} else {
				fmt.Println("\n⚠️ [KYBERNIS AUDIT FINDING] BLIND RETRY DETECTED ⚠️")
				fmt.Println("   The agent retried the exact same mutation after a timeout.")
				fmt.Println("   If the original request succeeded on the backend, this will cause a double-execution unless your backend idempotency keys are perfect.")
				fmt.Println()
			}
			return
		}
	}

	// First time seeing this request (or not a retry of a faulted one)
	t.history = append(t.history, Record{
		Method:        method,
		Path:          path,
		Body:          body,
		InjectedFault: fault,
	})
}
