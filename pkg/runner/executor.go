package runner

import (
	"fmt"
	"os"
	"os/exec"
)

func Execute(proxyPort int, cmdArgs []string) error {
	if len(cmdArgs) == 0 {
		return fmt.Errorf("no command provided to run")
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Inject the proxy variables natively into the child process
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", proxyPort)
	env := os.Environ()
	env = append(env, fmt.Sprintf("HTTP_PROXY=%s", proxyURL))
	env = append(env, fmt.Sprintf("HTTPS_PROXY=%s", proxyURL))
	cmd.Env = env

	fmt.Printf("🚀 Starting agent process: %v\n", cmdArgs)
	fmt.Printf("🔗 Network calls will automatically route through Kybernis proxy at %s\n\n", proxyURL)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("agent process exited with error: %w", err)
	}

	fmt.Println("\n✅ Agent process completed naturally.")
	return nil
}
