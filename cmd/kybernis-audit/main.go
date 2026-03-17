package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kybernis/kybernis-audit/pkg/fuzzer"
	"github.com/kybernis/kybernis-audit/pkg/scenario"
	"github.com/kybernis/kybernis-audit/pkg/telemetry"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	tracker := telemetry.NewTracker()
	defer tracker.Close() // Ensure telemetry fires before exit

	tracker.TrackEvent("cli_started", map[string]interface{}{"command": os.Args[1], "version": "v1.1.0-alpha"})

	switch os.Args[1] {
	case "fuzz":
		runFuzzer(os.Args[2:], tracker)
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("🛡️  Kybernis Audit Engine")
	fmt.Println("Usage:")
	fmt.Println("  kybernis-audit fuzz [flags]             Run deterministic agent fuzzing against your tool API")
}

func runFuzzer(args []string, tracker *telemetry.Tracker) {
	fs := flag.NewFlagSet("fuzz", flag.ExitOnError)
	scenarioFile := fs.String("config", "scenario.yaml", "Path to the chaos config")
	targetURL := fs.String("target", "", "Target backend URL (overrides config)")
	fs.Parse(args)

	cfg, err := scenario.LoadConfig(*scenarioFile)
	if err != nil {
		log.Fatalf("Failed to load scenario: %v", err)
	}

	url := cfg.TargetURL
	if *targetURL != "" {
		url = *targetURL
	}

	if url == "" {
		log.Fatalf("Target URL must be specified in config or via --target flag")
	}

	fuzz := fuzzer.NewFuzzer(url, cfg)
	err = fuzz.Run()
	
	// Track the fuzzer outcome
	tracker.TrackEvent("audit_completed", map[string]interface{}{
		"attack_vector": cfg.AttackVector,
		"variant":       cfg.Variant,
		"vulnerable":    err != nil, // If err != nil, backend failed (it's vulnerable)
	})

	if err != nil {
		log.Fatalf("\nAudit Failed: %v", err)
	}
	fmt.Println("\n✅ Audit Completed Successfully.")
}
