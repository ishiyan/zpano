//nolint:testpackage,dupl
package variance

import (
	"testing"
)

func TestVarianceOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    VarianceOutput
		text string
	}{
		{VarianceValue, varianceValue},
		{varianceLast, varianceUnknown},
		{VarianceOutput(0), varianceUnknown},
		{VarianceOutput(9999), varianceUnknown},
		{VarianceOutput(-9999), varianceUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestVarianceOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       VarianceOutput
		boolean bool
	}{
		{VarianceValue, true},
		{varianceLast, false},
		{VarianceOutput(0), false},
		{VarianceOutput(9999), false},
		{VarianceOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestVarianceOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         VarianceOutput
		json      string
		succeeded bool
	}{
		{VarianceValue, dqs + varianceValue + dqs, true},
		{varianceLast, nilstr, false},
		{VarianceOutput(9999), nilstr, false},
		{VarianceOutput(-9999), nilstr, false},
		{VarianceOutput(0), nilstr, false},
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

func TestVarianceOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero VarianceOutput
	tests := []struct {
		o         VarianceOutput
		json      string
		succeeded bool
	}{
		{VarianceValue, dqs + varianceValue + dqs, true},
		{zero, dqs + varianceUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o VarianceOutput

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
