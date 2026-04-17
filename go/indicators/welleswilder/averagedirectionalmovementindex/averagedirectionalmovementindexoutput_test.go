//nolint:testpackage,dupl
package averagedirectionalmovementindex

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    AverageDirectionalMovementIndexOutput
		text string
	}{
		{AverageDirectionalMovementIndexValue, averageDirectionalMovementIndexOutputValue},
		{DirectionalMovementIndexValue, averageDirectionalMovementIndexOutputDirectionalMovementIndex},
		{DirectionalIndicatorPlusValue, averageDirectionalMovementIndexOutputDirectionalIndicatorPlus},
		{DirectionalIndicatorMinusValue, averageDirectionalMovementIndexOutputDirectionalIndicatorMinus},
		{DirectionalMovementPlusValue, averageDirectionalMovementIndexOutputDirectionalMovementPlus},
		{DirectionalMovementMinusValue, averageDirectionalMovementIndexOutputDirectionalMovementMinus},
		{AverageTrueRangeValue, averageDirectionalMovementIndexOutputAverageTrueRange},
		{TrueRangeValue, averageDirectionalMovementIndexOutputTrueRange},
		{averageDirectionalMovementIndexLast, averageDirectionalMovementIndexOutputUnknown},
		{AverageDirectionalMovementIndexOutput(0), averageDirectionalMovementIndexOutputUnknown},
		{AverageDirectionalMovementIndexOutput(9999), averageDirectionalMovementIndexOutputUnknown},
		{AverageDirectionalMovementIndexOutput(-9999), averageDirectionalMovementIndexOutputUnknown},
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
		o       AverageDirectionalMovementIndexOutput
		boolean bool
	}{
		{AverageDirectionalMovementIndexValue, true},
		{DirectionalMovementIndexValue, true},
		{DirectionalIndicatorPlusValue, true},
		{DirectionalIndicatorMinusValue, true},
		{DirectionalMovementPlusValue, true},
		{DirectionalMovementMinusValue, true},
		{AverageTrueRangeValue, true},
		{TrueRangeValue, true},
		{averageDirectionalMovementIndexLast, false},
		{AverageDirectionalMovementIndexOutput(0), false},
		{AverageDirectionalMovementIndexOutput(9999), false},
		{AverageDirectionalMovementIndexOutput(-9999), false},
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
		o         AverageDirectionalMovementIndexOutput
		json      string
		succeeded bool
	}{
		{AverageDirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexOutputValue + dqs, true},
		{DirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexOutputDirectionalMovementIndex + dqs, true},
		{DirectionalIndicatorPlusValue, dqs + averageDirectionalMovementIndexOutputDirectionalIndicatorPlus + dqs, true},
		{DirectionalIndicatorMinusValue, dqs + averageDirectionalMovementIndexOutputDirectionalIndicatorMinus + dqs, true},
		{DirectionalMovementPlusValue, dqs + averageDirectionalMovementIndexOutputDirectionalMovementPlus + dqs, true},
		{DirectionalMovementMinusValue, dqs + averageDirectionalMovementIndexOutputDirectionalMovementMinus + dqs, true},
		{AverageTrueRangeValue, dqs + averageDirectionalMovementIndexOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + averageDirectionalMovementIndexOutputTrueRange + dqs, true},
		{averageDirectionalMovementIndexLast, nilstr, false},
		{AverageDirectionalMovementIndexOutput(9999), nilstr, false},
		{AverageDirectionalMovementIndexOutput(-9999), nilstr, false},
		{AverageDirectionalMovementIndexOutput(0), nilstr, false},
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

	var zero AverageDirectionalMovementIndexOutput
	tests := []struct {
		o         AverageDirectionalMovementIndexOutput
		json      string
		succeeded bool
	}{
		{AverageDirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexOutputValue + dqs, true},
		{DirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexOutputDirectionalMovementIndex + dqs, true},
		{DirectionalIndicatorPlusValue, dqs + averageDirectionalMovementIndexOutputDirectionalIndicatorPlus + dqs, true},
		{DirectionalIndicatorMinusValue, dqs + averageDirectionalMovementIndexOutputDirectionalIndicatorMinus + dqs, true},
		{DirectionalMovementPlusValue, dqs + averageDirectionalMovementIndexOutputDirectionalMovementPlus + dqs, true},
		{DirectionalMovementMinusValue, dqs + averageDirectionalMovementIndexOutputDirectionalMovementMinus + dqs, true},
		{AverageTrueRangeValue, dqs + averageDirectionalMovementIndexOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + averageDirectionalMovementIndexOutputTrueRange + dqs, true},
		{zero, dqs + averageDirectionalMovementIndexOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o AverageDirectionalMovementIndexOutput

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
