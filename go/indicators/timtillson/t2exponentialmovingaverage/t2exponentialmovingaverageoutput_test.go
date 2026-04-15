//nolint:testpackage,dupl
package t2exponentialmovingaverage

import (
	"testing"
)

func TestT2ExponentialMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    T2ExponentialMovingAverageOutput
		text string
	}{
		{T2ExponentialMovingAverageValue, t2ExponentialMovingAverageValueStr},
		{t2ExponentialMovingAverageLast, t2ExponentialMovingAverageUnknownStr},
		{T2ExponentialMovingAverageOutput(0), t2ExponentialMovingAverageUnknownStr},
		{T2ExponentialMovingAverageOutput(9999), t2ExponentialMovingAverageUnknownStr},
		{T2ExponentialMovingAverageOutput(-9999), t2ExponentialMovingAverageUnknownStr},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestT2ExponentialMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       T2ExponentialMovingAverageOutput
		boolean bool
	}{
		{T2ExponentialMovingAverageValue, true},
		{t2ExponentialMovingAverageLast, false},
		{T2ExponentialMovingAverageOutput(0), false},
		{T2ExponentialMovingAverageOutput(9999), false},
		{T2ExponentialMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestT2ExponentialMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         T2ExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{T2ExponentialMovingAverageValue, dqs + t2ExponentialMovingAverageValueStr + dqs, true},
		{t2ExponentialMovingAverageLast, nilstr, false},
		{T2ExponentialMovingAverageOutput(9999), nilstr, false},
		{T2ExponentialMovingAverageOutput(-9999), nilstr, false},
		{T2ExponentialMovingAverageOutput(0), nilstr, false},
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

func TestT2ExponentialMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero T2ExponentialMovingAverageOutput
	tests := []struct {
		o         T2ExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{T2ExponentialMovingAverageValue, dqs + t2ExponentialMovingAverageValueStr + dqs, true},
		{zero, dqs + t2ExponentialMovingAverageUnknownStr + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o T2ExponentialMovingAverageOutput

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
