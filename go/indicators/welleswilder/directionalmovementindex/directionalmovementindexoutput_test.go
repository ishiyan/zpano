//nolint:testpackage,dupl
package directionalmovementindex

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    DirectionalMovementIndexOutput
		text string
	}{
		{DirectionalMovementIndexValue, directionalMovementIndexOutputValue},
		{DirectionalIndicatorPlusValue, directionalMovementIndexOutputDirectionalIndicatorPlus},
		{DirectionalIndicatorMinusValue, directionalMovementIndexOutputDirectionalIndicatorMinus},
		{DirectionalMovementPlusValue, directionalMovementIndexOutputDirectionalMovementPlus},
		{DirectionalMovementMinusValue, directionalMovementIndexOutputDirectionalMovementMinus},
		{AverageTrueRangeValue, directionalMovementIndexOutputAverageTrueRange},
		{TrueRangeValue, directionalMovementIndexOutputTrueRange},
		{directionalMovementIndexLast, directionalMovementIndexOutputUnknown},
		{DirectionalMovementIndexOutput(0), directionalMovementIndexOutputUnknown},
		{DirectionalMovementIndexOutput(9999), directionalMovementIndexOutputUnknown},
		{DirectionalMovementIndexOutput(-9999), directionalMovementIndexOutputUnknown},
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
		o       DirectionalMovementIndexOutput
		boolean bool
	}{
		{DirectionalMovementIndexValue, true},
		{DirectionalIndicatorPlusValue, true},
		{DirectionalIndicatorMinusValue, true},
		{DirectionalMovementPlusValue, true},
		{DirectionalMovementMinusValue, true},
		{AverageTrueRangeValue, true},
		{TrueRangeValue, true},
		{directionalMovementIndexLast, false},
		{DirectionalMovementIndexOutput(0), false},
		{DirectionalMovementIndexOutput(9999), false},
		{DirectionalMovementIndexOutput(-9999), false},
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
		o         DirectionalMovementIndexOutput
		json      string
		succeeded bool
	}{
		{DirectionalMovementIndexValue, dqs + directionalMovementIndexOutputValue + dqs, true},
		{DirectionalIndicatorPlusValue, dqs + directionalMovementIndexOutputDirectionalIndicatorPlus + dqs, true},
		{DirectionalIndicatorMinusValue, dqs + directionalMovementIndexOutputDirectionalIndicatorMinus + dqs, true},
		{DirectionalMovementPlusValue, dqs + directionalMovementIndexOutputDirectionalMovementPlus + dqs, true},
		{DirectionalMovementMinusValue, dqs + directionalMovementIndexOutputDirectionalMovementMinus + dqs, true},
		{AverageTrueRangeValue, dqs + directionalMovementIndexOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + directionalMovementIndexOutputTrueRange + dqs, true},
		{directionalMovementIndexLast, nilstr, false},
		{DirectionalMovementIndexOutput(9999), nilstr, false},
		{DirectionalMovementIndexOutput(-9999), nilstr, false},
		{DirectionalMovementIndexOutput(0), nilstr, false},
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

	var zero DirectionalMovementIndexOutput
	tests := []struct {
		o         DirectionalMovementIndexOutput
		json      string
		succeeded bool
	}{
		{DirectionalMovementIndexValue, dqs + directionalMovementIndexOutputValue + dqs, true},
		{DirectionalIndicatorPlusValue, dqs + directionalMovementIndexOutputDirectionalIndicatorPlus + dqs, true},
		{DirectionalIndicatorMinusValue, dqs + directionalMovementIndexOutputDirectionalIndicatorMinus + dqs, true},
		{DirectionalMovementPlusValue, dqs + directionalMovementIndexOutputDirectionalMovementPlus + dqs, true},
		{DirectionalMovementMinusValue, dqs + directionalMovementIndexOutputDirectionalMovementMinus + dqs, true},
		{AverageTrueRangeValue, dqs + directionalMovementIndexOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + directionalMovementIndexOutputTrueRange + dqs, true},
		{zero, dqs + directionalMovementIndexOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o DirectionalMovementIndexOutput

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
