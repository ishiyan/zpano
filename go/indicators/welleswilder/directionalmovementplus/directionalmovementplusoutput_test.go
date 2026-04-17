//nolint:testpackage,dupl
package directionalmovementplus

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    DirectionalMovementPlusOutput
		text string
	}{
		{DirectionalMovementPlusValue, directionalMovementPlusOutputValue},
		{directionalMovementPlusLast, directionalMovementPlusOutputUnknown},
		{DirectionalMovementPlusOutput(0), directionalMovementPlusOutputUnknown},
		{DirectionalMovementPlusOutput(9999), directionalMovementPlusOutputUnknown},
		{DirectionalMovementPlusOutput(-9999), directionalMovementPlusOutputUnknown},
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
		o       DirectionalMovementPlusOutput
		boolean bool
	}{
		{DirectionalMovementPlusValue, true},
		{directionalMovementPlusLast, false},
		{DirectionalMovementPlusOutput(0), false},
		{DirectionalMovementPlusOutput(9999), false},
		{DirectionalMovementPlusOutput(-9999), false},
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
		o         DirectionalMovementPlusOutput
		json      string
		succeeded bool
	}{
		{DirectionalMovementPlusValue, dqs + directionalMovementPlusOutputValue + dqs, true},
		{directionalMovementPlusLast, nilstr, false},
		{DirectionalMovementPlusOutput(9999), nilstr, false},
		{DirectionalMovementPlusOutput(-9999), nilstr, false},
		{DirectionalMovementPlusOutput(0), nilstr, false},
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

	var zero DirectionalMovementPlusOutput
	tests := []struct {
		o         DirectionalMovementPlusOutput
		json      string
		succeeded bool
	}{
		{DirectionalMovementPlusValue, dqs + directionalMovementPlusOutputValue + dqs, true},
		{zero, dqs + directionalMovementPlusOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o DirectionalMovementPlusOutput

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
