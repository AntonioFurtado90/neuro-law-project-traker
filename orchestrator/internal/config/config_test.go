package config

import "testing"

func TestRequireEnv_ReturnsValueWhenSet(t *testing.T) {
	t.Setenv("NEUROLAW_TEST_VAR", "hello")

	value, err := RequireEnv("NEUROLAW_TEST_VAR")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != "hello" {
		t.Fatalf("expected %q, got %q", "hello", value)
	}
}

func TestRequireEnv_ErrorsWhenUnset(t *testing.T) {
	t.Setenv("NEUROLAW_TEST_VAR", "")

	_, err := RequireEnv("NEUROLAW_TEST_VAR")

	if err == nil {
		t.Fatal("expected an error for an unset variable, got nil")
	}
	if _, ok := err.(*MissingConfigError); !ok {
		t.Fatalf("expected *MissingConfigError, got %T", err)
	}
}
