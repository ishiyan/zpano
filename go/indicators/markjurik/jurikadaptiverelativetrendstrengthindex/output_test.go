//nolint:testpackage,dupl
package jurikadaptiverelativetrendstrengthindex

import (
	"testing"
)

func TestAdaptiveRelativeTrendStrengthIndexOutputString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		o    Output
		text string
	}{
		{Value, adaptiveRelativeTrendStrengthIndexValue},
		{adaptiveRelativeTrendStrengthIndexLast, adaptiveRelativeTrendStrengthIndexUnknown},
		{Output(0), adaptiveRelativeTrendStrengthIndexUnknown},
		{Output(9999), adaptiveRelativeTrendStrengthIndexUnknown},
		{Output(-9999), adaptiveRelativeTrendStrengthIndexUnknown},
	}
	for _, tt := range tests {
		exp := tt.text
		act := tt.o.String()
		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestAdaptiveRelativeTrendStrengthIndexOutputIsKnown(t *testing.T) {
	t.Parallel()
	tests := []struct {
		o       Output
		boolean bool
	}{
		{Value, true},
		{adaptiveRelativeTrendStrengthIndexLast, false},
		{Output(0), false},
		{Output(9999), false},
		{Output(-9999), false},
	}
	for _, tt := range tests {
		exp := tt.boolean
		act := tt.o.IsKnown()
		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.o, exp, act)
		}
	}
}

func TestAdaptiveRelativeTrendStrengthIndexOutputMarshalJSON(t *testing.T) {
	t.Parallel()
	const dqs = "\""
	var nilstr string
	tests := []struct {
		o         Output
		json      string
		succeeded bool
	}{
		{Value, dqs + adaptiveRelativeTrendStrengthIndexValue + dqs, true},
		{adaptiveRelativeTrendStrengthIndexLast, nilstr, false},
		{Output(9999), nilstr, false},
		{Output(-9999), nilstr, false},
		{Output(0), nilstr, false},
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

func TestAdaptiveRelativeTrendStrengthIndexOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()
	const dqs = "\""
	var zero Output
	tests := []struct {
		o         Output
		json      string
		succeeded bool
	}{
		{Value, dqs + adaptiveRelativeTrendStrengthIndexValue + dqs, true},
		{zero, dqs + adaptiveRelativeTrendStrengthIndexUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}
	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)
		var o Output
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
