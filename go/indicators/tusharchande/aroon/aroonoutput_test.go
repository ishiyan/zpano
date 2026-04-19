//nolint:testpackage,dupl
package aroon

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    AroonOutput
		text string
	}{
		{AroonUp, aroonOutputUp},
		{AroonDown, aroonOutputDown},
		{AroonOsc, aroonOutputOsc},
		{aroonLast, aroonOutputUnknown},
		{AroonOutput(0), aroonOutputUnknown},
		{AroonOutput(9999), aroonOutputUnknown},
		{AroonOutput(-9999), aroonOutputUnknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o       AroonOutput
		boolean bool
	}{
		{AroonUp, true},
		{AroonDown, true},
		{AroonOsc, true},
		{aroonLast, false},
		{AroonOutput(0), false},
		{AroonOutput(9999), false},
		{AroonOutput(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         AroonOutput
		json      string
		succeeded bool
	}{
		{AroonUp, dqs + aroonOutputUp + dqs, true},
		{AroonDown, dqs + aroonOutputDown + dqs, true},
		{AroonOsc, dqs + aroonOutputOsc + dqs, true},
		{aroonLast, nilstr, false},
		{AroonOutput(9999), nilstr, false},
		{AroonOutput(-9999), nilstr, false},
		{AroonOutput(0), nilstr, false},
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

func TestOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero AroonOutput
	tests := []struct {
		o         AroonOutput
		json      string
		succeeded bool
	}{
		{AroonUp, dqs + aroonOutputUp + dqs, true},
		{AroonDown, dqs + aroonOutputDown + dqs, true},
		{AroonOsc, dqs + aroonOutputOsc + dqs, true},
		{zero, dqs + aroonOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o AroonOutput

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
