package sandbox

// Verdict is the result of a permission check.
type Verdict int

const (
	VerdictAllow  Verdict = iota
	VerdictDeny
	VerdictPrompt
)

// String returns the verdict as a human-readable string.
func (v Verdict) String() string {
	switch v {
	case VerdictAllow:
		return "ALLOW"
	case VerdictDeny:
		return "DENY"
	case VerdictPrompt:
		return "PROMPT"
	default:
		return "UNKNOWN"
	}
}

// PermissionDecision is the result of a PolicyEngine check.
type PermissionDecision struct {
	Verdict Verdict
	Reason  string
}

// Allow creates an ALLOW decision.
func Allow(reason string) PermissionDecision {
	return PermissionDecision{Verdict: VerdictAllow, Reason: reason}
}

// Deny creates a DENY decision.
func Deny(reason string) PermissionDecision {
	return PermissionDecision{Verdict: VerdictDeny, Reason: reason}
}

// Prompt creates a PROMPT decision requiring user escalation.
func Prompt(reason string) PermissionDecision {
	return PermissionDecision{Verdict: VerdictPrompt, Reason: reason}
}
