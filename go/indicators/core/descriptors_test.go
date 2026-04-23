//nolint:testpackage
package core

import (
	"testing"
)

// TestDescriptorCoverage asserts that every known Identifier has a corresponding
// entry in the static descriptors registry. It is skipped until Phase 4 populates
// the table for all ~70 implemented indicators.
func TestDescriptorCoverage(t *testing.T) {
	t.Parallel()

	for id := SimpleMovingAverage; id.IsKnown(); id++ {
		if _, ok := descriptors[id]; !ok {
			t.Errorf("descriptor missing for identifier '%s'", id)
		}
	}
}

// TestDescriptorOfUnregistered verifies that DescriptorOf reports a miss for
// unknown identifiers.
func TestDescriptorOfUnregistered(t *testing.T) {
	t.Parallel()

	if _, ok := DescriptorOf(Identifier(0)); ok {
		t.Errorf("DescriptorOf(0): expected miss, got hit")
	}

	if _, ok := DescriptorOf(Identifier(-1)); ok {
		t.Errorf("DescriptorOf(-1): expected miss, got hit")
	}
}

// TestDescriptorsReturnsCopy verifies that mutating the returned map does not
// affect the static registry.
func TestDescriptorsReturnsCopy(t *testing.T) {
	t.Parallel()

	snapshot := Descriptors()
	snapshot[Identifier(9999)] = Descriptor{}

	if _, ok := DescriptorOf(Identifier(9999)); ok {
		t.Errorf("mutation of returned map leaked into registry")
	}
}
