//nolint:testpackage,dupl
package triangularmovingaverage

import (
	"testing"
)

func TestTriangularMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    TriangularMovingAverageOutput
		text string
	}{
		{TriangularMovingAverageValue, triangularMovingAverageValue},
		{triangularMovingAverageLast, triangularMovingAverageUnknown},
		{TriangularMovingAverageOutput(0), triangularMovingAverageUnknown},
		{TriangularMovingAverageOutput(9999), triangularMovingAverageUnknown},
		{TriangularMovingAverageOutput(-9999), triangularMovingAverageUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestTriangularMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       TriangularMovingAverageOutput
		boolean bool
	}{
		{TriangularMovingAverageValue, true},
		{triangularMovingAverageLast, false},
		{TriangularMovingAverageOutput(0), false},
		{TriangularMovingAverageOutput(9999), false},
		{TriangularMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestTriangularMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         TriangularMovingAverageOutput
		json      string
		succeeded bool
	}{
		{TriangularMovingAverageValue, dqs + triangularMovingAverageValue + dqs, true},
		{triangularMovingAverageLast, nilstr, false},
		{TriangularMovingAverageOutput(9999), nilstr, false},
		{TriangularMovingAverageOutput(-9999), nilstr, false},
		{TriangularMovingAverageOutput(0), nilstr, false},
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

func TestTriangularMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero TriangularMovingAverageOutput
	tests := []struct {
		o         TriangularMovingAverageOutput
		json      string
		succeeded bool
	}{
		{TriangularMovingAverageValue, dqs + triangularMovingAverageValue + dqs, true},
		{zero, dqs + triangularMovingAverageUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o TriangularMovingAverageOutput

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
