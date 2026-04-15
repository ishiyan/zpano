//nolint:testpackage,dupl
package kaufmanadaptivemovingaverage

import (
	"testing"
)

func TestKaufmanAdaptiveMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    KaufmanAdaptiveMovingAverageOutput
		text string
	}{
		{KaufmanAdaptiveMovingAverageValue, kaufmanAdaptiveMovingAverageValue},
		{kaufmanAdaptiveMovingAverageLast, kaufmanAdaptiveMovingAverageUnknown},
		{KaufmanAdaptiveMovingAverageOutput(0), kaufmanAdaptiveMovingAverageUnknown},
		{KaufmanAdaptiveMovingAverageOutput(9999), kaufmanAdaptiveMovingAverageUnknown},
		{KaufmanAdaptiveMovingAverageOutput(-9999), kaufmanAdaptiveMovingAverageUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestKaufmanAdaptiveMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       KaufmanAdaptiveMovingAverageOutput
		boolean bool
	}{
		{KaufmanAdaptiveMovingAverageValue, true},
		{kaufmanAdaptiveMovingAverageLast, false},
		{KaufmanAdaptiveMovingAverageOutput(0), false},
		{KaufmanAdaptiveMovingAverageOutput(9999), false},
		{KaufmanAdaptiveMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestKaufmanAdaptiveMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         KaufmanAdaptiveMovingAverageOutput
		json      string
		succeeded bool
	}{
		{KaufmanAdaptiveMovingAverageValue, dqs + kaufmanAdaptiveMovingAverageValue + dqs, true},
		{kaufmanAdaptiveMovingAverageLast, nilstr, false},
		{KaufmanAdaptiveMovingAverageOutput(9999), nilstr, false},
		{KaufmanAdaptiveMovingAverageOutput(-9999), nilstr, false},
		{KaufmanAdaptiveMovingAverageOutput(0), nilstr, false},
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

func TestKaufmanAdaptiveMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero KaufmanAdaptiveMovingAverageOutput
	tests := []struct {
		o         KaufmanAdaptiveMovingAverageOutput
		json      string
		succeeded bool
	}{
		{KaufmanAdaptiveMovingAverageValue, dqs + kaufmanAdaptiveMovingAverageValue + dqs, true},
		{zero, dqs + kaufmanAdaptiveMovingAverageUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o KaufmanAdaptiveMovingAverageOutput

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
