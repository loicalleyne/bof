// Package main — crush run subprocess management.
package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// NOTE: bof-mcp passes promptContent via strings.NewReader assigned to cmd.Stdin.
// This intentionally differs from esquisse-mcp which passes a temp file path.
// Both patterns satisfy the security invariant (no shell interpolation).

// RunResult holds the output of a crush run invocation.
type RunResult struct {
	Output   string
	ExitCode int
}

// RunCrush invokes crush run --model {model} --quiet with the given prompt
// passed via stdin as a strings.NewReader. It does NOT interpolate model or
// prompt content into a shell command string — this is the security invariant
// for shell-injection prevention.
func RunCrush(ctx context.Context, model, promptContent string) (RunResult, error) {
	crushPath, err := exec.LookPath("crush")
	if err != nil {
		return RunResult{}, fmt.Errorf("crush binary not found in PATH: %w", err)
	}

	cmd := exec.CommandContext(ctx, crushPath, "run", "--model", model, "--quiet")
	cmd.Stdin = strings.NewReader(promptContent)
	out, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return RunResult{}, fmt.Errorf("crush run: %w", err)
		}
	}
	return RunResult{Output: string(out), ExitCode: exitCode}, nil
}
