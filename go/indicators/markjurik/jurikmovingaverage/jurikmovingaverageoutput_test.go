//nolint:testpackage
package jurikmovingaverage

import (
	"testing"
)

func TestMJurikovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    JurikMovingAverageOutput
		text string
	}{
		{MovingAverageValue, movingAverageValue},
		{movingAverageLast, movingAverageUnknown},
		{JurikMovingAverageOutput(0), movingAverageUnknown},
		{JurikMovingAverageOutput(9999), movingAverageUnknown},
		{JurikMovingAverageOutput(-9999), movingAverageUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       JurikMovingAverageOutput
		boolean bool
	}{
		{MovingAverageValue, true},
		{movingAverageLast, false},
		{JurikMovingAverageOutput(0), false},
		{JurikMovingAverageOutput(9999), false},
		{JurikMovingAverageOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         JurikMovingAverageOutput
		json      string
		succeeded bool
	}{
		{MovingAverageValue, dqs + movingAverageValue + dqs, true},
		{movingAverageLast, nilstr, false},
		{JurikMovingAverageOutput(9999), nilstr, false},
		{JurikMovingAverageOutput(-9999), nilstr, false},
		{JurikMovingAverageOutput(0), nilstr, false},
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

func TestMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero JurikMovingAverageOutput
	tests := []struct {
		o         JurikMovingAverageOutput
		json      string
		succeeded bool
	}{
		{MovingAverageValue, dqs + movingAverageValue + dqs, true},
		{zero, dqs + movingAverageUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o JurikMovingAverageOutput

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
