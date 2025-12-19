package artifacts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type ShellHook struct {
	script      string
	hostEnvVars map[string]string
}

func NewShellHook(script string, hostEnvVars map[string]string) *ShellHook {
	return &ShellHook{
		script:      script,
		hostEnvVars: hostEnvVars,
	}
}

func (s *ShellHook) Execute(ctx context.Context) error {
	if s.script == "" {
		// NOOP
		return nil
	}

	fmt.Println("⚙ Executing install hook...")
	cmd := exec.Command("bash", "-c", s.script)
	vars := make([]string, 0, len(s.hostEnvVars))
	for k, v := range s.hostEnvVars {
		vars = append(vars, fmt.Sprintf("%s=%s", k, os.ExpandEnv(v)))
	}
	cmd.Env = append(os.Environ(), vars...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute install hook: %w", err)
	}
	return nil
}
