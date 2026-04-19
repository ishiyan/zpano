//nolint:testpackage,dupl
package stochastic

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    StochasticOutput
		text string
	}{
		{StochasticFastK, stochasticOutputFastK},
		{StochasticSlowK, stochasticOutputSlowK},
		{StochasticSlowD, stochasticOutputSlowD},
		{stochasticLast, stochasticOutputUnknown},
		{StochasticOutput(0), stochasticOutputUnknown},
		{StochasticOutput(9999), stochasticOutputUnknown},
		{StochasticOutput(-9999), stochasticOutputUnknown},
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
		o       StochasticOutput
		boolean bool
	}{
		{StochasticFastK, true},
		{StochasticSlowK, true},
		{StochasticSlowD, true},
		{stochasticLast, false},
		{StochasticOutput(0), false},
		{StochasticOutput(9999), false},
		{StochasticOutput(-9999), false},
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
		o         StochasticOutput
		json      string
		succeeded bool
	}{
		{StochasticFastK, dqs + stochasticOutputFastK + dqs, true},
		{StochasticSlowK, dqs + stochasticOutputSlowK + dqs, true},
		{StochasticSlowD, dqs + stochasticOutputSlowD + dqs, true},
		{stochasticLast, nilstr, false},
		{StochasticOutput(9999), nilstr, false},
		{StochasticOutput(-9999), nilstr, false},
		{StochasticOutput(0), nilstr, false},
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

	var zero StochasticOutput
	tests := []struct {
		o         StochasticOutput
		json      string
		succeeded bool
	}{
		{StochasticFastK, dqs + stochasticOutputFastK + dqs, true},
		{StochasticSlowK, dqs + stochasticOutputSlowK + dqs, true},
		{StochasticSlowD, dqs + stochasticOutputSlowD + dqs, true},
		{zero, dqs + stochasticOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o StochasticOutput

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
