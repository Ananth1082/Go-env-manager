package env_manager

import "testing"

func assertEqual[T comparable](t *testing.T, actual, expected T, msg string) {
	if actual != expected {
		t.Errorf("Assertion error\nExpected %v got %v\n%s", expected, actual, msg)
	}
}

func assertCondition(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Errorf("Assertion error\nCondition failed: %s", msg)
	}
}
