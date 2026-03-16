package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kybernis/kybernis-audit/pkg/evaluator"
	"github.com/kybernis/kybernis-audit/pkg/proxy"
	"github.com/kybernis/kybernis-audit/pkg/runner"
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
	scenarioFile := fs.String("scenario", "default.yaml", "Path to the chaos config")
	targetURL := fs.String("target", "http://localhost:8080", "Target backend URL")
	port := fs.Int("port", 9999, "Proxy port")
	fs.Parse(args)

	startProxyServer(*scenarioFile, *targetURL, *port, tracker)
}

func runWrapper(args []string, tracker *telemetry.Tracker) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	scenarioFile := fs.String("scenario", "default.yaml", "Path to the chaos config")
	targetURL := fs.String("target", "http://localhost:8080", "Target backend URL")
	port := fs.Int("port", 9999, "Proxy port")
	fs.Parse(args)

	cmdArgs := fs.Args()
	if len(cmdArgs) > 0 && cmdArgs[0] == "--" {
		cmdArgs = cmdArgs[1:]
	}
	if len(cmdArgs) == 0 {
		log.Fatal("You must provide an agent command to run. E.g.: kybernis-audit run -- python agent.py")
	}

	// 1. Setup proxy in background
	srv, eval, err := startProxyBackground(*scenarioFile, *targetURL, *port, tracker)
	if err != nil {
		log.Fatalf("Failed to start background proxy: %v", err)
	}

	// 2. Execute the user's agent script
	runErr := runner.Execute(*port, cmdArgs)

	// 3. Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	// 4. Summarize Evaluation Findings
	fmt.Println("\n📊 === KYBERNIS AUDIT REPORT ===")
	eval.PrintSummary()

	if runErr != nil {
		log.Fatalf("\n❌ Run finished with agent execution error: %v", runErr)
	}
}

func startProxyServer(scenarioFile, target string, port int, tracker *telemetry.Tracker) {
	cfg, err := scenario.LoadConfig(scenarioFile)
	if err != nil {
		log.Fatalf("Failed to load scenario: %v", err)
	}

	fmt.Printf("🎯 Scenario: %s\n", cfg.Name)
	eval := evaluator.NewTracker()
	interceptor, err := proxy.NewChaosInterceptor(target, tracker, cfg, eval)
	if err != nil {
		log.Fatalf("Failed to initialize proxy: %v", err)
	}

	listenAddr := fmt.Sprintf(":%d", port)
	fmt.Printf("🎧 Proxy listening on http://localhost%s\n", listenAddr)
	if err := http.ListenAndServe(listenAddr, interceptor); err != nil {
		log.Fatalf("Proxy server failed: %v", err)
	}
}

func startProxyBackground(scenarioFile, target string, port int, tracker *telemetry.Tracker) (*http.Server, *evaluator.Tracker, error) {
	cfg, err := scenario.LoadConfig(scenarioFile)
	if err != nil {
		return nil, nil, err
	}
	eval := evaluator.NewTracker()
	interceptor, err := proxy.NewChaosInterceptor(target, tracker, cfg, eval)
	if err != nil {
		return nil, nil, err
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: interceptor,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Background proxy server failed: %v", err)
		}
	}()

	// Wait briefly for proxy to bind
	time.Sleep(100 * time.Millisecond)
	return srv, eval, nil
}
