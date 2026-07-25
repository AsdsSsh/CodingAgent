package sandbox

// Sandbox is the abstraction for execution environments.
// Planned implementations: SubprocessSandbox (current), DockerSandbox (future).
type Sandbox interface {
	// PolicyEngine returns the permission engine for this sandbox.
	PolicyEngine() *PolicyEngine

	// ExecuteCommand runs a shell command under sandbox constraints.
	ExecuteCommand(command string) CommandResult

	// IsInWorkspace checks whether a path is within the allowed workspace.
	IsInWorkspace(path string) bool
}

// CommandResult is the result of a sandbox command execution.
type CommandResult struct {
	Success bool
	Output  string
	Error   string
}

// CommandOk creates a successful command result.
func CommandOk(output string) CommandResult {
	return CommandResult{Success: true, Output: output}
}

// CommandFail creates a failed command result.
func CommandFail(err string) CommandResult {
	return CommandResult{Success: false, Error: err}
}
