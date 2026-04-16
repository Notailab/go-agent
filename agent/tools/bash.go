package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Notailab/go-agent/agent/core"
)

type BashTool struct{}

func (t *BashTool) Name() string {
	return "Bash"
}

func (t *BashTool) Description() string {
	return "Run a bash command in an isolated shell session."
}

func (t *BashTool) Parameters() core.Parameters {
	return core.Parameters{
		Type: "object",
		Properties: map[string]core.Param{
			"command": {
				Type:        "string",
				Description: "The bash command to execute",
			},
			"cwd": {
				Type:        "string",
				Description: "Optional working directory for the command",
			},
			"timeout_seconds": {
				Type:        "integer",
				Description: "Optional timeout in seconds",
			},
		},
		Required: []string{"command"},
	}
}

func (t *BashTool) Execute(paramsJson string) (string, error) {
	params, err := core.ParseToolParams(paramsJson, t.Parameters())
	if err != nil {
		return "", err
	}

	rawCommand := params["command"].(string)
	if strings.TrimSpace(rawCommand) == "" {
		return "", fmt.Errorf("missing required parameter: command")
	}
	command := rawCommand

	timeout := 30 * time.Second
	if rawTimeout, ok := params["timeout_seconds"]; ok && rawTimeout != nil {
		if timeoutValue, ok := rawTimeout.(float64); ok && timeoutValue > 0 {
			timeout = time.Duration(timeoutValue) * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "--noprofile", "--norc", "-lc", command)

	if rawCwd, ok := params["cwd"].(string); ok && strings.TrimSpace(rawCwd) != "" {
		cwd := strings.TrimSpace(rawCwd)
		if info, err := os.Stat(cwd); err != nil {
			return "", fmt.Errorf("invalid cwd: %w", err)
		} else if !info.IsDir() {
			return "", fmt.Errorf("invalid cwd: not a directory: %s", cwd)
		}
		cmd.Dir = cwd
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		combined := strings.TrimSpace(stdout.String())
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText != "" {
			if combined != "" {
				combined += "\n"
			}
			combined += stderrText
		}
		if ctx.Err() == context.DeadlineExceeded {
			if combined != "" {
				return combined, fmt.Errorf("command timed out after %s", timeout)
			}
			return "", fmt.Errorf("command timed out after %s", timeout)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if combined != "" {
				return combined, fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
			}
			return "", fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		if combined != "" {
			return combined, err
		}
		return "", err
	}

	return stdout.String(), nil
}
