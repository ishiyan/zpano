//nolint:testpackage,dupl
package simplemovingaverage

import (
	"testing"
)

func TestSimpleMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    SimpleMovingAverageOutput
		text string
	}{
		{SimpleMovingAverageValue, simpleMovingAverageValue},
		{simpleMovingAverageLast, simpleMovingAverageUnknown},
		{SimpleMovingAverageOutput(0), simpleMovingAverageUnknown},
		{SimpleMovingAverageOutput(9999), simpleMovingAverageUnknown},
		{SimpleMovingAverageOutput(-9999), simpleMovingAverageUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestSimpleMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       SimpleMovingAverageOutput
		boolean bool
	}{
		{SimpleMovingAverageValue, true},
		{simpleMovingAverageLast, false},
		{SimpleMovingAverageOutput(0), false},
		{SimpleMovingAverageOutput(9999), false},
		{SimpleMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestSimpleMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         SimpleMovingAverageOutput
		json      string
		succeeded bool
	}{
		{SimpleMovingAverageValue, dqs + simpleMovingAverageValue + dqs, true},
		{simpleMovingAverageLast, nilstr, false},
		{SimpleMovingAverageOutput(9999), nilstr, false},
		{SimpleMovingAverageOutput(-9999), nilstr, false},
		{SimpleMovingAverageOutput(0), nilstr, false},
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

func TestSimpleMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero SimpleMovingAverageOutput
	tests := []struct {
		o         SimpleMovingAverageOutput
		json      string
		succeeded bool
	}{
		{SimpleMovingAverageValue, dqs + simpleMovingAverageValue + dqs, true},
		{zero, dqs + simpleMovingAverageUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o SimpleMovingAverageOutput

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
