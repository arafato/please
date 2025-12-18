package container

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var containerBinaryName = func() string {
	switch runtime.GOOS {
	case "darwin":
		return "container"
	case "linux":
		return "docker"
	default:
		return "container"
	}
}()

type Client struct {
	path string
}

func NewClient() (*Client, error) {
	binary := containerBinaryName
	if path, err := exec.LookPath(binary); err == nil {
		return &Client{path: path}, nil
	}
	return nil, fmt.Errorf("failed to discover binary '%s'", binary)
}

// TODO: image and (default) version is missing
func (c *Client) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.path, args...)
	return cmd.CombinedOutput()
}

func (c *Client) Install(ctx context.Context, image string, version string, platform string) error {
	var cmd *exec.Cmd

	if platform != "" {
		cmd = exec.CommandContext(ctx, c.path, "image", "pull", "--platform", fmt.Sprintf("%s", platform), fmt.Sprintf("%s:%s", image, version))
	} else {
		cmd = exec.CommandContext(ctx, c.path, "image", "pull", fmt.Sprintf("%s:%s", image, version))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
