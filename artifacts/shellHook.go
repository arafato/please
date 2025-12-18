package artifacts

import (
	"context"
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

	cmd := exec.Command("bash", "-c", s.script)
	cmd.Env = append(os.Environ(),
		"GOMODCACHE="+"value1",
		"GOCACHE="+"value 2",
	)
	return nil
}
