//nolint:testpackage,dupl
package directionalmovementminus

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    DirectionalMovementMinusOutput
		text string
	}{
		{DirectionalMovementMinusValue, directionalMovementMinusOutputValue},
		{directionalMovementMinusLast, directionalMovementMinusOutputUnknown},
		{DirectionalMovementMinusOutput(0), directionalMovementMinusOutputUnknown},
		{DirectionalMovementMinusOutput(9999), directionalMovementMinusOutputUnknown},
		{DirectionalMovementMinusOutput(-9999), directionalMovementMinusOutputUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       DirectionalMovementMinusOutput
		boolean bool
	}{
		{DirectionalMovementMinusValue, true},
		{directionalMovementMinusLast, false},
		{DirectionalMovementMinusOutput(0), false},
		{DirectionalMovementMinusOutput(9999), false},
		{DirectionalMovementMinusOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         DirectionalMovementMinusOutput
		json      string
		succeeded bool
	}{
		{DirectionalMovementMinusValue, dqs + directionalMovementMinusOutputValue + dqs, true},
		{directionalMovementMinusLast, nilstr, false},
		{DirectionalMovementMinusOutput(9999), nilstr, false},
		{DirectionalMovementMinusOutput(-9999), nilstr, false},
		{DirectionalMovementMinusOutput(0), nilstr, false},
	}

	for _, tt := range tests {
		exp := tt.json
		bs, err := tt.o.MarshalJSON()

		if err != nil && tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success '%v', got error %v", tt.o, exp, err)
			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tt.o)
			continue
		}

		act := string(bs)
		if exp != act {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero DirectionalMovementMinusOutput
	tests := []struct {
		o         DirectionalMovementMinusOutput
		json      string
		succeeded bool
	}{
		{DirectionalMovementMinusValue, dqs + directionalMovementMinusOutputValue + dqs, true},
		{zero, dqs + directionalMovementMinusOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o DirectionalMovementMinusOutput

		err := o.UnmarshalJSON(bs)
		if err != nil && tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success '%v', got error %v", tt.json, exp, err)
			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("MarshalJSON('%v'): expected error, got success", tt.json)
			continue
		}

		if exp != o {
			t.Errorf("MarshalJSON('%v'): expected '%v', actual '%v'", tt.json, exp, o)
		}
	}
}
