package sandbox

import "testing"

func TestPolicyEngineAllowByLevel(t *testing.T) {
	pe := NewPolicyEngine(LevelRead, 10, 300)
	decision := pe.Check("read", LevelRead)
	if decision.Verdict != VerdictAllow {
		t.Errorf("expected ALLOW for READ tool with READ level, got %s", decision.Verdict)
	}
}

func TestPolicyEnginePromptEscalation(t *testing.T) {
	pe := NewPolicyEngine(LevelRead, 10, 300)
	decision := pe.Check("bash", LevelExecute)
	if decision.Verdict != VerdictPrompt {
		t.Errorf("expected PROMPT for EXECUTE tool with READ level, got %s", decision.Verdict)
	}
}

func TestPolicyEngineCoversHigher(t *testing.T) {
	pe := NewPolicyEngine(LevelExecute, 10, 300)
	decision := pe.Check("read", LevelRead)
	if decision.Verdict != VerdictAllow {
		t.Errorf("expected ALLOW for READ tool with EXECUTE level, got %s", decision.Verdict)
	}
	decision2 := pe.Check("write", LevelWrite)
	if decision2.Verdict != VerdictAllow {
		t.Errorf("expected ALLOW for WRITE tool with EXECUTE level, got %s", decision2.Verdict)
	}
}

func TestPolicyEngineUserOverride(t *testing.T) {
	pe := NewPolicyEngine(LevelRead, 10, 300)
	pe.AddOverride("bash", true)
	decision := pe.Check("bash", LevelExecute)
	if decision.Verdict != VerdictAllow {
		t.Errorf("expected ALLOW for bash with user override, got %s", decision.Verdict)
	}
}

func TestPolicyEngineUserDenyOverride(t *testing.T) {
	pe := NewPolicyEngine(LevelExecute, 10, 300)
	pe.AddOverride("bash", false)
	decision := pe.Check("bash", LevelExecute)
	if decision.Verdict != VerdictDeny {
		t.Errorf("expected DENY for bash with user deny override, got %s", decision.Verdict)
	}
}

func TestPolicyEngineEscalate(t *testing.T) {
	pe := NewPolicyEngine(LevelRead, 10, 300)
	pe.Escalate(LevelExecute)
	if pe.CurrentLevel() != LevelExecute {
		t.Errorf("expected EXECUTE after escalation, got %s", pe.CurrentLevel())
	}
	// Cannot downgrade
	pe.Escalate(LevelRead)
	if pe.CurrentLevel() != LevelExecute {
		t.Error("level should not downgrade")
	}
}

func TestPermissionLevelCovers(t *testing.T) {
	if !LevelExecute.Covers(LevelRead) {
		t.Error("EXECUTE should cover READ")
	}
	if !LevelWrite.Covers(LevelWrite) {
		t.Error("WRITE should cover WRITE")
	}
	if LevelRead.Covers(LevelExecute) {
		t.Error("READ should not cover EXECUTE")
	}
}
