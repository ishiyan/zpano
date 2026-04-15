//nolint:testpackage,dupl
package standarddeviation

import (
	"testing"
)

func TestStandardDeviationOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    StandardDeviationOutput
		text string
	}{
		{StandardDeviationValue, standardDeviationValue},
		{standardDeviationLast, standardDeviationUnknown},
		{StandardDeviationOutput(0), standardDeviationUnknown},
		{StandardDeviationOutput(9999), standardDeviationUnknown},
		{StandardDeviationOutput(-9999), standardDeviationUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestStandardDeviationOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       StandardDeviationOutput
		boolean bool
	}{
		{StandardDeviationValue, true},
		{standardDeviationLast, false},
		{StandardDeviationOutput(0), false},
		{StandardDeviationOutput(9999), false},
		{StandardDeviationOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestStandardDeviationOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         StandardDeviationOutput
		json      string
		succeeded bool
	}{
		{StandardDeviationValue, dqs + standardDeviationValue + dqs, true},
		{standardDeviationLast, nilstr, false},
		{StandardDeviationOutput(9999), nilstr, false},
		{StandardDeviationOutput(-9999), nilstr, false},
		{StandardDeviationOutput(0), nilstr, false},
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

func TestStandardDeviationOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero StandardDeviationOutput
	tests := []struct {
		o         StandardDeviationOutput
		json      string
		succeeded bool
	}{
		{StandardDeviationValue, dqs + standardDeviationValue + dqs, true},
		{zero, dqs + standardDeviationUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o StandardDeviationOutput

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
