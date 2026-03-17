package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Tracker struct {
	Enabled   bool
	Endpoint  string
	MachineID string
	wg        sync.WaitGroup
	client    *http.Client
}

func NewTracker() *Tracker {
	enabled := true
	if os.Getenv("KYBERNIS_TELEMETRY") == "0" || os.Getenv("KYBERNIS_TELEMETRY") == "false" {
		enabled = false
	}

	// Fetch or generate anonymous machine ID
	machineID := getOrGenerateMachineID()

	return &Tracker{
		Enabled:   enabled,
		Endpoint:  "https://api.kybernis.dev/v1/telemetry", // Telemetry ingestion endpoint
		MachineID: machineID,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (t *Tracker) TrackEvent(eventName string, properties map[string]interface{}) {
	if !t.Enabled {
		return
	}

	payload := map[string]interface{}{
		"event":      eventName,
		"machine_id": t.MachineID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"properties": properties,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		req, err := http.NewRequest("POST", t.Endpoint, bytes.NewBuffer(data))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "kybernis-audit/cli")
			// Fire and forget
			t.client.Do(req)
		}
	}()
}

// Close waits for all pending telemetry requests to complete, up to a timeout.
func (t *Tracker) Close() {
	if !t.Enabled {
		return
	}
	
	c := make(chan struct{})
	go func() {
		defer close(c)
		t.wg.Wait()
	}()
	
	select {
	case <-c:
	case <-time.After(1 * time.Second): // don't block CLI exit for long
	}
}

func getOrGenerateMachineID() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return uuid.New().String()
	}

	configDir := home + "/.kybernis"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}

	idFile := configDir + "/anon_id"
	data, err := os.ReadFile(idFile)
	if err == nil && len(data) > 0 {
		return string(data)
	}

	newID := uuid.New().String()
	os.WriteFile(idFile, []byte(newID), 0644)
	return newID
}
