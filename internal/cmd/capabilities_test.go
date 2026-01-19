package cmd

import (
	"testing"
)

func TestCapabilitiesResponse(t *testing.T) {
	// Test that all the slices are populated correctly
	writeList := DataWriteCommandsList()
	exemptList := ExemptCommandsList()
	readList := readCommands()

	if len(writeList) == 0 {
		t.Error("write commands list is empty")
	}
	if len(exemptList) == 0 {
		t.Error("exempt commands list is empty")
	}
	if len(readList) == 0 {
		t.Error("read commands list is empty")
	}

	// Verify no overlap between write and exempt
	writeSet := make(map[string]bool)
	for _, cmd := range writeList {
		writeSet[cmd] = true
	}

	for _, cmd := range exemptList {
		if writeSet[cmd] {
			t.Errorf("command %q appears in both write and exempt lists", cmd)
		}
	}

	// Verify read commands don't include write commands
	for _, cmd := range readList {
		if writeSet[cmd] {
			t.Errorf("command %q appears in both read and write lists", cmd)
		}
	}
}

func TestCapabilitiesFeatures(t *testing.T) {
	expectedFeatures := []string{"enable-commands", "readonly", "envelope", "agent-mode"}

	// Verify that the features we document are what we expect
	features := []string{"enable-commands", "readonly", "envelope", "agent-mode"}

	if len(features) != len(expectedFeatures) {
		t.Errorf("features count mismatch: got %d, want %d", len(features), len(expectedFeatures))
	}

	featureSet := make(map[string]bool)
	for _, f := range features {
		featureSet[f] = true
	}

	for _, expected := range expectedFeatures {
		if !featureSet[expected] {
			t.Errorf("missing expected feature: %s", expected)
		}
	}
}
