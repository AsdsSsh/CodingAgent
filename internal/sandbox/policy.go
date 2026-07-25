package sandbox

import (
	"sync"
	"time"
)

// PolicyEngine is the central permission decision engine with three-tier checking:
// 1. Permanent user overrides (allow/deny by tool name)
// 2. Permission level coverage + frequency check
// 3. Permission level escalation prompt
type PolicyEngine struct {
	mu            sync.RWMutex
	currentLevel  PermissionLevel
	maxFrequency  int
	windowDuration time.Duration
	userOverrides map[string]bool       // tool name → allowed
	useTimestamps map[string][]time.Time // tool name → usage history
}

// NewPolicyEngine creates a PolicyEngine with the given level and frequency limits.
func NewPolicyEngine(level PermissionLevel, maxFrequency int, windowSeconds int) *PolicyEngine {
	return &PolicyEngine{
		currentLevel:   level,
		maxFrequency:   maxFrequency,
		windowDuration: time.Duration(windowSeconds) * time.Second,
		userOverrides:  make(map[string]bool),
		useTimestamps:  make(map[string][]time.Time),
	}
}

// NewPolicyEngineDefault creates a PolicyEngine with default frequency limits (10 per 300s).
func NewPolicyEngineDefault(level PermissionLevel) *PolicyEngine {
	return NewPolicyEngine(level, 10, 300)
}

// Check evaluates whether a tool call should be allowed, denied, or prompted.
func (pe *PolicyEngine) Check(toolName string, required PermissionLevel) PermissionDecision {
	// Tier 1: permanent user overrides
	pe.mu.RLock()
	override, hasOverride := pe.userOverrides[toolName]
	pe.mu.RUnlock()

	if hasOverride {
		if override {
			pe.recordUse(toolName)
			return Allow("permanently allowed by user")
		}
		return Deny("permanently denied by user")
	}

	// Tier 2: permission level sufficient → check frequency
	pe.mu.RLock()
	level := pe.currentLevel
	pe.mu.RUnlock()

	if level.Covers(required) {
		if pe.shouldPrompt(toolName) {
			return Prompt("Tool '" + toolName + "' has been used " + itoa(pe.maxFrequency) + " times in the last " + itoa(int(pe.windowDuration.Seconds())) + " seconds, please re-authorize")
		}
		pe.recordUse(toolName)
		return Allow("permission level " + level.String())
	}

	// Tier 3: insufficient permission
	return Prompt("Tool '" + toolName + "' requires " + required.String() + " level, current level is " + level.String())
}

// shouldPrompt checks if the tool has exceeded frequency limits in the current window.
func (pe *PolicyEngine) shouldPrompt(toolName string) bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	timestamps, ok := pe.useTimestamps[toolName]
	if !ok {
		return false
	}

	cutoff := time.Now().Add(-pe.windowDuration)
	count := 0
	for _, t := range timestamps {
		if t.After(cutoff) {
			count++
		}
	}
	return count >= pe.maxFrequency
}

// recordUse logs a tool usage timestamp for frequency tracking.
func (pe *PolicyEngine) recordUse(toolName string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.useTimestamps[toolName] = append(pe.useTimestamps[toolName], time.Now())
}

// Escalate raises the current permission level (monotonic — cannot be lowered).
func (pe *PolicyEngine) Escalate(newLevel PermissionLevel) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	if newLevel > pe.currentLevel {
		pe.currentLevel = newLevel
	}
}

// AddOverride permanently allows or denies a specific tool.
func (pe *PolicyEngine) AddOverride(toolName string, allow bool) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.userOverrides[toolName] = allow
}

// RemoveOverride removes a permanent override, restoring normal policy evaluation.
func (pe *PolicyEngine) RemoveOverride(toolName string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	delete(pe.userOverrides, toolName)
}

// CurrentLevel returns the current permission level.
func (pe *PolicyEngine) CurrentLevel() PermissionLevel {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.currentLevel
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0'+n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
