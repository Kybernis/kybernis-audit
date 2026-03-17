package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
	defer tracker.Close() // Ensure we flush telemetry before exit
	tracker.TrackEvent("cli_started", map[string]interface{}{"command": os.Args[1], "version": "v1.2.0-alpha"})

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
	targetURL := fs.String("target", "", "Global target backend URL (overrides config)")
	fs.Parse(args)

	manifest, err := scenario.LoadConfig(*scenarioFile)
	if err != nil {
		log.Fatalf("Failed to load scenario manifest: %v", err)
	}

	if len(manifest.Scenarios) == 0 {
		log.Fatalf("No scenarios found in the config file. Add a 'scenarios' list.")
	}

	fmt.Printf("\n🚀 Launching Kybernis Audit: %s\n", manifest.Name)
	fmt.Println("=======================================================")

	failedCount := 0
	passedCount := 0
	var failedNames []string

	for i, cfg := range manifest.Scenarios {
		fmt.Printf("\n🧪 [Test %d/%d] %s\n", i+1, len(manifest.Scenarios), cfg.Name)
		fmt.Println(strings.Repeat("-", 55))

		url := cfg.TargetURL
		if *targetURL != "" {
			url = *targetURL
		}

		if url == "" {
			log.Fatalf("Target URL must be specified globally or per scenario")
		}

		fuzz := fuzzer.NewFuzzer(url, cfg, tracker)
		err = fuzz.Run()
		if err != nil {
			failedCount++
			failedNames = append(failedNames, fmt.Sprintf("%s (%s)", strings.ToUpper(cfg.AttackVector), cfg.Variant))
		} else {
			passedCount++
		}
	}

	// The Psychological Summary Table
	fmt.Println("\n\n=======================================================")
	fmt.Println("🛑 KYBERNIS AUDIT SUMMARY REPORT")
	fmt.Println("=======================================================")
	fmt.Printf("Total Tests Run: %d\n", len(manifest.Scenarios))
	fmt.Printf("✅ Passed: %d\n", passedCount)
	fmt.Printf("❌ Failed: %d\n\n", failedCount)

	if failedCount > 0 {
		fmt.Println("🚨 CRITICAL VULNERABILITIES DETECTED 🚨")
		for _, name := range failedNames {
			fmt.Printf("  - ❌ %s\n", name)
		}
		fmt.Println("\nYour agent infrastructure is critically exposed to semantic double-spends and partial state corruption.")
		fmt.Println("Patching these individually requires Distributed Redis Locks, Temporal workflows, and strict payload hashing.")
		fmt.Println("\nOr, fix all of them permanently in 2 lines of code.")
		fmt.Println("⚡ Use the Kybernis SDK State-Machine Architecture: https://kybernis.io")
	} else {
		fmt.Println("🎉 ALL TESTS PASSED. Your agent infrastructure is resilient.")
	}
	fmt.Println("=======================================================\n")

	if failedCount > 0 {
		os.Exit(1)
	}
}
