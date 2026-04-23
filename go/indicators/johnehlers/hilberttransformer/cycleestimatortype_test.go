//nolint:testpackage
package hilberttransformer

import (
	"testing"
)

func TestCycleEstimatorTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		t    CycleEstimatorType
		text string
	}{
		{HomodyneDiscriminator, homodyneDiscriminator},
		{HomodyneDiscriminatorUnrolled, homodyneDiscriminatorUnrolled},
		{PhaseAccumulator, phaseAccumulator},
		{DualDifferentiator, dualDifferentiator},
		{last, unknown},
		{CycleEstimatorType(0), unknown},
		{CycleEstimatorType(9999), unknown},
		{CycleEstimatorType(-9999), unknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.t.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.t, exp, act)
		}
	}
}

func TestCycleEstimatorTypeIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		t       CycleEstimatorType
		boolean bool
	}{
		{HomodyneDiscriminator, true},
		{HomodyneDiscriminatorUnrolled, true},
		{PhaseAccumulator, true},
		{DualDifferentiator, true},
		{last, false},
		{CycleEstimatorType(0), false},
		{CycleEstimatorType(9999), false},
		{CycleEstimatorType(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.t.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.t, exp, act)
		}
	}
}

func TestCycleEstimatorTypeMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		t         CycleEstimatorType
		json      string
		succeeded bool
	}{
		{HomodyneDiscriminator, dqs + homodyneDiscriminator + dqs, true},
		{HomodyneDiscriminatorUnrolled, dqs + homodyneDiscriminatorUnrolled + dqs, true},
		{PhaseAccumulator, dqs + phaseAccumulator + dqs, true},
		{DualDifferentiator, dqs + dualDifferentiator + dqs, true},
		{last, nilstr, false},
		{CycleEstimatorType(9999), nilstr, false},
		{CycleEstimatorType(-9999), nilstr, false},
		{CycleEstimatorType(0), nilstr, false},
	}

	for _, tt := range tests {
		exp := tt.json
		bs, err := tt.t.MarshalJSON()

		if err != nil && tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success '%v', got error %v", tt.t, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tt.t)

			continue
		}

		act := string(bs)
		if exp != act {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tt.t, exp, act)
		}
	}
}

func TestCycleEstimatorTypeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero CycleEstimatorType
	tests := []struct {
		t         CycleEstimatorType
		json      string
		succeeded bool
	}{
		{HomodyneDiscriminator, dqs + homodyneDiscriminator + dqs, true},
		{HomodyneDiscriminatorUnrolled, dqs + homodyneDiscriminatorUnrolled + dqs, true},
		{PhaseAccumulator, dqs + phaseAccumulator + dqs, true},
		{DualDifferentiator, dqs + dualDifferentiator + dqs, true},
		{zero, "\"unknown\"", false},
		{zero, "\"foobar\"", false},
	}

	for _, tt := range tests {
		exp := tt.t
		bs := []byte(tt.json)

		var act CycleEstimatorType

		err := act.UnmarshalJSON(bs)
		if err != nil && tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success '%v', got error %v", tt.json, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("MarshalJSON('%v'): expected error, got success", tt.json)

			continue
		}

		if exp != act {
			t.Errorf("MarshalJSON('%v'): expected '%v', actual '%v'", tt.json, exp, act)
		}
	}
}
