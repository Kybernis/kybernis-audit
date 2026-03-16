package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kybernis/kybernis-audit/pkg/evaluator"
	"github.com/kybernis/kybernis-audit/pkg/proxy"
	"github.com/kybernis/kybernis-audit/pkg/scenario"
	"github.com/kybernis/kybernis-audit/pkg/telemetry"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	tracker := telemetry.NewTracker()
	tracker.TrackEvent("cli_started", map[string]interface{}{"command": os.Args[1], "version": "v1.0.0-alpha"})

	switch os.Args[1] {
	case "proxy":
		runProxy(os.Args[2:], tracker)
	case "run":
		runWrapper(os.Args[2:], tracker)
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("🛡️  Kybernis Audit Engine")
	fmt.Println("Usage:")
	fmt.Println("  kybernis-audit proxy [flags]             Run the proxy standalone")
	fmt.Println("  kybernis-audit run [flags] -- <cmd>      Run an agent script through the proxy (CI mode)")
}

func runProxy(args []string, tracker *telemetry.Tracker) {
	fs := flag.NewFlagSet("proxy", flag.ExitOnError)
	scenarioFile := fs.String("scenario", "default.yaml", "Path to the chaos scenario config")
	targetURL := fs.String("target", "http://localhost:8080", "The mock backend API the agent interacts with")
	port := fs.Int("port", 9999, "Local port to listen on for the proxy")
	fs.Parse(args)

	cfg, err := scenario.LoadConfig(*scenarioFile)
	if err != nil {
		log.Fatalf("Failed to load scenario: %v", err)
	}

	fmt.Printf("🎯 Scenario: %s\n", cfg.Name)
	fmt.Printf("💥 Injecting Fault: %s on %s\n", cfg.FaultInjection, cfg.TargetEndpoint)

	eval := evaluator.NewTracker()
	interceptor, err := proxy.NewChaosInterceptor(*targetURL, tracker, cfg, eval)
	if err != nil {
		log.Fatalf("Failed to initialize proxy: %v", err)
	}

	listenAddr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🎧 Proxy listening on http://localhost%s\n", listenAddr)
	if err := http.ListenAndServe(listenAddr, interceptor); err != nil {
		log.Fatalf("Proxy server failed: %v", err)
	}
}

func runWrapper(args []string, tracker *telemetry.Tracker) {
	// To be implemented in next step
	fmt.Println("The 'run' wrapper is not yet implemented.")
}
