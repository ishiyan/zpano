//nolint:testpackage,dupl
package tripleexponentialmovingaverage

import (
	"testing"
)

func TestTripleExponentialMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    TripleExponentialMovingAverageOutput
		text string
	}{
		{TripleExponentialMovingAverageValue, tripleExponentialMovingAverageValueStr},
		{tripleExponentialMovingAverageLast, tripleExponentialMovingAverageUnknownStr},
		{TripleExponentialMovingAverageOutput(0), tripleExponentialMovingAverageUnknownStr},
		{TripleExponentialMovingAverageOutput(9999), tripleExponentialMovingAverageUnknownStr},
		{TripleExponentialMovingAverageOutput(-9999), tripleExponentialMovingAverageUnknownStr},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestTripleExponentialMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       TripleExponentialMovingAverageOutput
		boolean bool
	}{
		{TripleExponentialMovingAverageValue, true},
		{tripleExponentialMovingAverageLast, false},
		{TripleExponentialMovingAverageOutput(0), false},
		{TripleExponentialMovingAverageOutput(9999), false},
		{TripleExponentialMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestTripleExponentialMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         TripleExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{TripleExponentialMovingAverageValue, dqs + tripleExponentialMovingAverageValueStr + dqs, true},
		{tripleExponentialMovingAverageLast, nilstr, false},
		{TripleExponentialMovingAverageOutput(9999), nilstr, false},
		{TripleExponentialMovingAverageOutput(-9999), nilstr, false},
		{TripleExponentialMovingAverageOutput(0), nilstr, false},
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

func TestTripleExponentialMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero TripleExponentialMovingAverageOutput
	tests := []struct {
		o         TripleExponentialMovingAverageOutput
		json      string
		succeeded bool
	}{
		{TripleExponentialMovingAverageValue, dqs + tripleExponentialMovingAverageValueStr + dqs, true},
		{zero, dqs + tripleExponentialMovingAverageUnknownStr + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o TripleExponentialMovingAverageOutput

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
