//nolint:testpackage,dupl
package onbalancevolume

import (
	"testing"
)

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    OnBalanceVolumeOutput
		text string
	}{
		{OnBalanceVolumeValue, onBalanceVolumeOutputValue},
		{onBalanceVolumeLast, onBalanceVolumeOutputUnknown},
		{OnBalanceVolumeOutput(0), onBalanceVolumeOutputUnknown},
		{OnBalanceVolumeOutput(9999), onBalanceVolumeOutputUnknown},
		{OnBalanceVolumeOutput(-9999), onBalanceVolumeOutputUnknown},
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
		o       OnBalanceVolumeOutput
		boolean bool
	}{
		{OnBalanceVolumeValue, true},
		{onBalanceVolumeLast, false},
		{OnBalanceVolumeOutput(0), false},
		{OnBalanceVolumeOutput(9999), false},
		{OnBalanceVolumeOutput(-9999), false},
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
		o         OnBalanceVolumeOutput
		json      string
		succeeded bool
	}{
		{OnBalanceVolumeValue, dqs + onBalanceVolumeOutputValue + dqs, true},
		{onBalanceVolumeLast, nilstr, false},
		{OnBalanceVolumeOutput(9999), nilstr, false},
		{OnBalanceVolumeOutput(-9999), nilstr, false},
		{OnBalanceVolumeOutput(0), nilstr, false},
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

	var zero OnBalanceVolumeOutput
	tests := []struct {
		o         OnBalanceVolumeOutput
		json      string
		succeeded bool
	}{
		{OnBalanceVolumeValue, dqs + onBalanceVolumeOutputValue + dqs, true},
		{zero, dqs + onBalanceVolumeOutputUnknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.o
		bs := []byte(tt.json)

		var o OnBalanceVolumeOutput

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
