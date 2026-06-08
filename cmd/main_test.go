package main

import "testing"

func TestShouldRunFullResearchHarnessDefaultsFalse(t *testing.T) {
	t.Setenv("RUN_FULL_RESEARCH_HARNESS", "")
	if shouldRunFullResearchHarness() {
		t.Fatal("expected research harness to stay disabled by default")
	}
}

func TestShouldRunFullResearchHarnessTrueWhenSet(t *testing.T) {
	t.Setenv("RUN_FULL_RESEARCH_HARNESS", "true")
	if !shouldRunFullResearchHarness() {
		t.Fatal("expected research harness to run when explicitly enabled")
	}
}
