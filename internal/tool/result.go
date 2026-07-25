package tool

import (
	"fmt"
	"strings"
)

const maxOutputLength = 100_000

// ToolResult is the outcome of a tool execution.
type ToolResult struct {
	Success   bool
	Data      string
	Error     string
	Truncated bool
}

// Ok creates a successful result.
func Ok(data string) ToolResult {
	return ToolResult{Success: true, Data: data}
}

// OkTruncated creates a successful but truncated result.
func OkTruncated(data string) ToolResult {
	return ToolResult{Success: true, Data: data, Truncated: true}
}

// Fail creates a failure result with an error message.
func Fail(err string) ToolResult {
	return ToolResult{Success: false, Error: err}
}

// FromCommandOutput creates a result from shell command stdout/stderr.
func FromCommandOutput(exitCode int, stdout, stderr string) ToolResult {
	var sb strings.Builder
	if stdout != "" {
		sb.WriteString(stdout)
	}
	if stderr != "" {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(stderr)
	}
	output := sb.String()
	truncated := len(output) > maxOutputLength
	if truncated {
		output = output[:maxOutputLength] + "\n... [output truncated]"
	}
	if exitCode != 0 {
		return ToolResult{Success: false, Data: output, Error: fmt.Sprintf("Exit code: %d", exitCode), Truncated: truncated}
	}
	return ToolResult{Success: true, Data: output, Truncated: truncated}
}

// FormatForLLM formats the result for insertion into LLM conversation.
func (r ToolResult) FormatForLLM() string {
	if !r.Success {
		if r.Error != "" {
			return "Error: " + r.Error
		}
		return "Error: Unknown error"
	}
	if r.Truncated {
		return r.Data + "\n\n[Output truncated — use more specific parameters]"
	}
	return r.Data
}
