package llm

import "testing"

func TestDetectPresetExactMatch(t *testing.T) {
	preset := DetectPreset("claude-sonnet-4-20250514")
	if preset == nil {
		t.Fatal("expected preset for exact model match")
	}
	if preset.Name != "Anthropic" {
		t.Errorf("expected Anthropic, got %s", preset.Name)
	}
}

func TestDetectPresetPrefixMatch(t *testing.T) {
	preset := DetectPreset("gpt-4o-mini")
	if preset == nil {
		t.Fatal("expected preset for gpt- prefix match")
	}
	if preset.Name != "OpenAI" {
		t.Errorf("expected OpenAI, got %s", preset.Name)
	}
}

func TestDetectPresetDeepSeek(t *testing.T) {
	preset := DetectPreset("deepseek-chat")
	if preset == nil {
		t.Fatal("expected preset for deepseek- prefix match")
	}
	if preset.Name != "DeepSeek" {
		t.Errorf("expected DeepSeek, got %s", preset.Name)
	}
}

func TestDetectPresetUnknown(t *testing.T) {
	preset := DetectPreset("unknown-model-xyz")
	if preset != nil {
		t.Errorf("expected nil for unknown model, got %s", preset.Name)
	}
}

func TestDetectPresetEmpty(t *testing.T) {
	preset := DetectPreset("")
	if preset != nil {
		t.Error("expected nil for empty model name")
	}
}
