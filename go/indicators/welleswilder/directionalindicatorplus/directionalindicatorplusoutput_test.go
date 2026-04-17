//nolint:testpackage,dupl
package directionalindicatorplus

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    DirectionalIndicatorPlusOutput
		text string
	}{
		{DirectionalIndicatorPlusValue, directionalIndicatorPlusOutputValue},
		{DirectionalMovementPlusValue, directionalIndicatorPlusOutputDirectionalMovement},
		{AverageTrueRangeValue, directionalIndicatorPlusOutputAverageTrueRange},
		{TrueRangeValue, directionalIndicatorPlusOutputTrueRange},
		{directionalIndicatorPlusLast, directionalIndicatorPlusOutputUnknown},
		{DirectionalIndicatorPlusOutput(0), directionalIndicatorPlusOutputUnknown},
		{DirectionalIndicatorPlusOutput(9999), directionalIndicatorPlusOutputUnknown},
		{DirectionalIndicatorPlusOutput(-9999), directionalIndicatorPlusOutputUnknown},
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
		o       DirectionalIndicatorPlusOutput
		boolean bool
	}{
		{DirectionalIndicatorPlusValue, true},
		{DirectionalMovementPlusValue, true},
		{AverageTrueRangeValue, true},
		{TrueRangeValue, true},
		{directionalIndicatorPlusLast, false},
		{DirectionalIndicatorPlusOutput(0), false},
		{DirectionalIndicatorPlusOutput(9999), false},
		{DirectionalIndicatorPlusOutput(-9999), false},
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
		o         DirectionalIndicatorPlusOutput
		json      string
		succeeded bool
	}{
		{DirectionalIndicatorPlusValue, dqs + directionalIndicatorPlusOutputValue + dqs, true},
		{DirectionalMovementPlusValue, dqs + directionalIndicatorPlusOutputDirectionalMovement + dqs, true},
		{AverageTrueRangeValue, dqs + directionalIndicatorPlusOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + directionalIndicatorPlusOutputTrueRange + dqs, true},
		{directionalIndicatorPlusLast, nilstr, false},
		{DirectionalIndicatorPlusOutput(9999), nilstr, false},
		{DirectionalIndicatorPlusOutput(-9999), nilstr, false},
		{DirectionalIndicatorPlusOutput(0), nilstr, false},
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

	var zero DirectionalIndicatorPlusOutput
	tests := []struct {
		o         DirectionalIndicatorPlusOutput
		json      string
		succeeded bool
	}{
		{DirectionalIndicatorPlusValue, dqs + directionalIndicatorPlusOutputValue + dqs, true},
		{DirectionalMovementPlusValue, dqs + directionalIndicatorPlusOutputDirectionalMovement + dqs, true},
		{AverageTrueRangeValue, dqs + directionalIndicatorPlusOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + directionalIndicatorPlusOutputTrueRange + dqs, true},
		{zero, dqs + directionalIndicatorPlusOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o DirectionalIndicatorPlusOutput

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
