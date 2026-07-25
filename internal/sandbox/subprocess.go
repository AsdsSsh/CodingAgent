package sandbox

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Default blocked commands — always blocked regardless of permission level.
var defaultBlockedCommands = []string{
	"rm -rf /", "sudo", "dd if=", "mkfs",
	":(){ :|:& };:", "shutdown", "reboot", "chmod 777 /",
}

// SubprocessSandbox executes commands in OS subprocesses with resource limits.
type SubprocessSandbox struct {
	workspace      string
	timeoutSeconds int
	blockedCmds    []string
	policyEngine   *PolicyEngine
}

// NewSubprocessSandbox creates a sandbox for the given workspace and permission level.
func NewSubprocessSandbox(workspace string, timeoutSeconds int, blockedCmds []string, level PermissionLevel) *SubprocessSandbox {
	if blockedCmds == nil {
		blockedCmds = defaultBlockedCommands
	}
	return &SubprocessSandbox{
		workspace:      workspace,
		timeoutSeconds: timeoutSeconds,
		blockedCmds:    blockedCmds,
		policyEngine:   NewPolicyEngineDefault(level),
	}
}

func (s *SubprocessSandbox) PolicyEngine() *PolicyEngine { return s.policyEngine }

func (s *SubprocessSandbox) ExecuteCommand(command string) CommandResult {
	// Blacklist check before spawning process
	for _, blocked := range s.blockedCmds {
		if strings.Contains(command, blocked) {
			return CommandFail("Command blocked by security policy: contains '" + blocked + "'")
		}
	}

	timeout := time.Duration(s.timeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}
	cmd.Dir = s.workspace

	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return CommandFail("Command timed out after " + itoa(s.timeoutSeconds) + " seconds")
	}

	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		result := fromCommandOutput(exitCode, string(output))
		return CommandResult{Success: result.Success, Output: result.Output, Error: result.Error}
	}

	result := fromCommandOutput(0, string(output))
	return CommandResult{Success: result.Success, Output: result.Output, Error: result.Error}
}

func (s *SubprocessSandbox) IsInWorkspace(path string) bool {
	absPath, err := filepathAbs(path)
	if err != nil {
		return false
	}
	absWS, err := filepathAbs(s.workspace)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(absWS))
}

// filepathAbs is a thin wrapper for testing.
var filepathAbs = func(path string) (string, error) {
	return strings.ToLower(path), nil
}

func init() {
	// Use the real filepath.Abs at runtime
	filepathAbs = func(path string) (string, error) {
		// Clean the path first
		cleaned := strings.TrimSpace(path)
		if cleaned == "" {
			cleaned = "."
		}
		return filepathAbsReal(cleaned)
	}
}

func filepathAbsReal(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return path, nil
	}
	abs := path
	if !strings.HasPrefix(path, "/") && !(len(path) >= 2 && path[1] == ':') {
		abs = wd + "/" + path
	}
	return strings.ToLower(abs), nil
}

// fromCommandOutput combines stdout/stderr into a single result.
func fromCommandOutput(exitCode int, output string) CommandResult {
	const maxOutput = 100_000
	truncated := len(output) > maxOutput
	if truncated {
		output = output[:maxOutput] + "\n... [output truncated]"
	}
	if exitCode != 0 {
		return CommandResult{Success: false, Output: output, Error: "Exit code: " + itoa(exitCode)}
	}
	return CommandResult{Success: true, Output: output}
}
