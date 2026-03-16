package telemetry

import (
	"log"
	"os"
)

type Tracker struct {
	enabled bool
}

func NewTracker() *Tracker {
	// Respect opt-out like PostHog / Next.js
	enabled := true
	if os.Getenv("KYBERNIS_TELEMETRY") == "0" {
		enabled = false
	}
	return &Tracker{enabled: enabled}
}

func (t *Tracker) TrackEvent(eventName string, properties map[string]interface{}) {
	if !t.enabled {
		return
	}
	// TODO: Send to external analytics endpoint (e.g., PostHog) asynchronously
	// This does NOT collect PII, agent code, or payload data.
	log.Printf("[Telemetry] Sent anonymous event: %s %v\n", eventName, properties)
}
