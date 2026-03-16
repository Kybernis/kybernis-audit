package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/kybernis/kybernis-audit/pkg/proxy"
	"github.com/kybernis/kybernis-audit/pkg/scenario"
	"github.com/kybernis/kybernis-audit/pkg/telemetry"
)

func main() {
	scenarioFile := flag.String("scenario", "default.yaml", "Path to the chaos scenario config")
	targetURL := flag.String("target", "http://localhost:8080", "The mock backend API the agent interacts with")
	port := flag.Int("port", 9999, "Local port to listen on for the proxy")

	flag.Parse()

	tracker := telemetry.NewTracker()
	tracker.TrackEvent("cli_started", map[string]interface{}{"version": "v1.0.0-alpha"})

	fmt.Println("🛡️  Kybernis Audit Engine Started")

	cfg, err := scenario.LoadConfig(*scenarioFile)
	if err != nil {
		log.Fatalf("Failed to load scenario: %v", err)
	}

	fmt.Printf("🎯 Scenario: %s\n", cfg.Name)
	fmt.Printf("💥 Injecting Fault: %s on %s\n", cfg.FaultInjection, cfg.TargetEndpoint)

	interceptor, err := proxy.NewChaosInterceptor(*targetURL, tracker, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize proxy: %v", err)
	}

	listenAddr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🎧 Proxy listening on http://localhost%s\n", listenAddr)
	fmt.Printf("👉 Point your agent's API URL to this proxy to begin testing.\n\n")

	if err := http.ListenAndServe(listenAddr, interceptor); err != nil {
		log.Fatalf("Proxy server failed: %v", err)
	}
}
