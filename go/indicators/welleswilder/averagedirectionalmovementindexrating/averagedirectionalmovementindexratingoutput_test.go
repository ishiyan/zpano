//nolint:testpackage,dupl
package averagedirectionalmovementindexrating

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    AverageDirectionalMovementIndexRatingOutput
		text string
	}{
		{AverageDirectionalMovementIndexRatingValue, averageDirectionalMovementIndexRatingOutputValue},
		{AverageDirectionalMovementIndexValue, averageDirectionalMovementIndexRatingOutputAverageDirectionalMovementIndex},
		{DirectionalMovementIndexValue, averageDirectionalMovementIndexRatingOutputDirectionalMovementIndex},
		{DirectionalIndicatorPlusValue, averageDirectionalMovementIndexRatingOutputDirectionalIndicatorPlus},
		{DirectionalIndicatorMinusValue, averageDirectionalMovementIndexRatingOutputDirectionalIndicatorMinus},
		{DirectionalMovementPlusValue, averageDirectionalMovementIndexRatingOutputDirectionalMovementPlus},
		{DirectionalMovementMinusValue, averageDirectionalMovementIndexRatingOutputDirectionalMovementMinus},
		{AverageTrueRangeValue, averageDirectionalMovementIndexRatingOutputAverageTrueRange},
		{TrueRangeValue, averageDirectionalMovementIndexRatingOutputTrueRange},
		{averageDirectionalMovementIndexRatingLast, averageDirectionalMovementIndexRatingOutputUnknown},
		{AverageDirectionalMovementIndexRatingOutput(0), averageDirectionalMovementIndexRatingOutputUnknown},
		{AverageDirectionalMovementIndexRatingOutput(9999), averageDirectionalMovementIndexRatingOutputUnknown},
		{AverageDirectionalMovementIndexRatingOutput(-9999), averageDirectionalMovementIndexRatingOutputUnknown},
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
		o       AverageDirectionalMovementIndexRatingOutput
		boolean bool
	}{
		{AverageDirectionalMovementIndexRatingValue, true},
		{AverageDirectionalMovementIndexValue, true},
		{DirectionalMovementIndexValue, true},
		{DirectionalIndicatorPlusValue, true},
		{DirectionalIndicatorMinusValue, true},
		{DirectionalMovementPlusValue, true},
		{DirectionalMovementMinusValue, true},
		{AverageTrueRangeValue, true},
		{TrueRangeValue, true},
		{averageDirectionalMovementIndexRatingLast, false},
		{AverageDirectionalMovementIndexRatingOutput(0), false},
		{AverageDirectionalMovementIndexRatingOutput(9999), false},
		{AverageDirectionalMovementIndexRatingOutput(-9999), false},
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
		o         AverageDirectionalMovementIndexRatingOutput
		json      string
		succeeded bool
	}{
		{AverageDirectionalMovementIndexRatingValue, dqs + averageDirectionalMovementIndexRatingOutputValue + dqs, true},
		{AverageDirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexRatingOutputAverageDirectionalMovementIndex + dqs, true},
		{DirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalMovementIndex + dqs, true},
		{DirectionalIndicatorPlusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalIndicatorPlus + dqs, true},
		{DirectionalIndicatorMinusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalIndicatorMinus + dqs, true},
		{DirectionalMovementPlusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalMovementPlus + dqs, true},
		{DirectionalMovementMinusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalMovementMinus + dqs, true},
		{AverageTrueRangeValue, dqs + averageDirectionalMovementIndexRatingOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + averageDirectionalMovementIndexRatingOutputTrueRange + dqs, true},
		{averageDirectionalMovementIndexRatingLast, nilstr, false},
		{AverageDirectionalMovementIndexRatingOutput(9999), nilstr, false},
		{AverageDirectionalMovementIndexRatingOutput(-9999), nilstr, false},
		{AverageDirectionalMovementIndexRatingOutput(0), nilstr, false},
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

	var zero AverageDirectionalMovementIndexRatingOutput
	tests := []struct {
		o         AverageDirectionalMovementIndexRatingOutput
		json      string
		succeeded bool
	}{
		{AverageDirectionalMovementIndexRatingValue, dqs + averageDirectionalMovementIndexRatingOutputValue + dqs, true},
		{AverageDirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexRatingOutputAverageDirectionalMovementIndex + dqs, true},
		{DirectionalMovementIndexValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalMovementIndex + dqs, true},
		{DirectionalIndicatorPlusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalIndicatorPlus + dqs, true},
		{DirectionalIndicatorMinusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalIndicatorMinus + dqs, true},
		{DirectionalMovementPlusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalMovementPlus + dqs, true},
		{DirectionalMovementMinusValue, dqs + averageDirectionalMovementIndexRatingOutputDirectionalMovementMinus + dqs, true},
		{AverageTrueRangeValue, dqs + averageDirectionalMovementIndexRatingOutputAverageTrueRange + dqs, true},
		{TrueRangeValue, dqs + averageDirectionalMovementIndexRatingOutputTrueRange + dqs, true},
		{zero, dqs + averageDirectionalMovementIndexRatingOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o AverageDirectionalMovementIndexRatingOutput

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
