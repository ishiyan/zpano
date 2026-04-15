//nolint:testpackage,dupl
package doubleexponentialmovingaverage

import (
	"testing"
)

func TestDoubleExponentialMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    DoubleExponentialMovingAverageOutput
		text string
	}{
		{DoubleExponentialMovingAverageValue, doubleExponentialMovingAverageValueStr},
		{doubleExponentialMovingAverageLast, doubleExponentialMovingAverageUnknownStr},
		{DoubleExponentialMovingAverageOutput(0), doubleExponentialMovingAverageUnknownStr},
		{DoubleExponentialMovingAverageOutput(9999), doubleExponentialMovingAverageUnknownStr},
		{DoubleExponentialMovingAverageOutput(-9999), doubleExponentialMovingAverageUnknownStr},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestDoubleExponentialMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       DoubleExponentialMovingAverageOutput
		boolean bool
	}{
		{DoubleExponentialMovingAverageValue, true},
		{doubleExponentialMovingAverageLast, false},
		{DoubleExponentialMovingAverageOutput(0), false},
		{DoubleExponentialMovingAverageOutput(9999), false},
		{DoubleExponentialMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestDoubleExponentialMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         DoubleExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{DoubleExponentialMovingAverageValue, dqs + doubleExponentialMovingAverageValueStr + dqs, true},
		{doubleExponentialMovingAverageLast, nilstr, false},
		{DoubleExponentialMovingAverageOutput(9999), nilstr, false},
		{DoubleExponentialMovingAverageOutput(-9999), nilstr, false},
		{DoubleExponentialMovingAverageOutput(0), nilstr, false},
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

func TestDoubleExponentialMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero DoubleExponentialMovingAverageOutput
	tests := []struct {
		o         DoubleExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{DoubleExponentialMovingAverageValue, dqs + doubleExponentialMovingAverageValueStr + dqs, true},
		{zero, dqs + doubleExponentialMovingAverageUnknownStr + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o DoubleExponentialMovingAverageOutput

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
