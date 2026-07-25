package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	// Blacklist check
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
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}
	cmd.Dir = s.workspace

	// Use process group so we can kill children on timeout
	cmd.SysProcAttr = newProcessGroupAttr()

	// Pipe stdout and stderr manually (avoid CombinedOutput pipe deadlock)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return CommandFail("Failed to create stdout pipe: " + err.Error())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return CommandFail("Failed to create stderr pipe: " + err.Error())
	}

	if err := cmd.Start(); err != nil {
		return CommandFail("Failed to start command: " + err.Error())
	}

	// Read stdout + stderr in goroutines
	const maxOutput = 500_000
	var outBuf, errBuf bytes.Buffer
	outDone := make(chan struct{})
	errDone := make(chan struct{})

	go func() {
		io.CopyN(&outBuf, stdout, maxOutput)
		close(outDone)
	}()
	go func() {
		io.CopyN(&errBuf, stderr, maxOutput)
		close(errDone)
	}()

	// Wait for command to finish or timeout
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Timeout: kill the entire process tree
		killProcessTree(cmd.Process.Pid)
		return CommandFail("Command timed out after " + itoa(s.timeoutSeconds) + " seconds")

	case err := <-cmdDone:
		// Command finished — wait for all output to be read
		<-outDone
		<-errDone

		var output string
		if outBuf.Len()+errBuf.Len() > maxOutput {
			output = outBuf.String() + errBuf.String()
			if len(output) > maxOutput {
				output = output[:maxOutput] + "\n... [output truncated]"
			}
			if err != nil {
				return CommandResult{
					Success:   false,
					Output:    output,
					Error:     "Command failed and output truncated",
					Truncated: true,
				}
			}
			return CommandResult{Success: true, Output: output, Truncated: true}
		}

		output = outBuf.String()

		if errBuf.Len() > 0 {
			if len(output) > 0 {
				output += "\n"
			}
			output += errBuf.String()
		}

		if err != nil {
			exitCode := 1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
			return CommandResult{
				Success: false,
				Output:  output,
				Error:   fmt.Sprintf("Exit code: %d", exitCode),
			}
		}

		return CommandResult{Success: true, Output: output}
	}
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

var filepathAbs = func(path string) (string, error) {
	return strings.ToLower(path), nil
}

func init() {
	filepathAbs = func(path string) (string, error) {
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

