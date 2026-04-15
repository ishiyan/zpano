//nolint:testpackage,dupl
package exponentialmovingaverage

import (
	"testing"
)

func TestExponentialMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    ExponentialMovingAverageOutput
		text string
	}{
		{ExponentialMovingAverageValue, exponentialMovingAverageValue},
		{exponentialMovingAverageLast, exponentialMovingAverageUnknown},
		{ExponentialMovingAverageOutput(0), exponentialMovingAverageUnknown},
		{ExponentialMovingAverageOutput(9999), exponentialMovingAverageUnknown},
		{ExponentialMovingAverageOutput(-9999), exponentialMovingAverageUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestExponentialMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       ExponentialMovingAverageOutput
		boolean bool
	}{
		{ExponentialMovingAverageValue, true},
		{exponentialMovingAverageLast, false},
		{ExponentialMovingAverageOutput(0), false},
		{ExponentialMovingAverageOutput(9999), false},
		{ExponentialMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestExponentialMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         ExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{ExponentialMovingAverageValue, dqs + exponentialMovingAverageValue + dqs, true},
		{exponentialMovingAverageLast, nilstr, false},
		{ExponentialMovingAverageOutput(9999), nilstr, false},
		{ExponentialMovingAverageOutput(-9999), nilstr, false},
		{ExponentialMovingAverageOutput(0), nilstr, false},
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

func TestExponentialMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero ExponentialMovingAverageOutput
	tests := []struct {
		o         ExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{ExponentialMovingAverageValue, dqs + exponentialMovingAverageValue + dqs, true},
		{zero, dqs + exponentialMovingAverageUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o ExponentialMovingAverageOutput

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
