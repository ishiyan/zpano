//nolint:testpackage
package coronatrendvigor

import "testing"

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o    Output
		text string
	}{
		{Value, valueStr},
		{TrendVigor, trendVigorStr},
		{outputLast, unknownStr},
		{Output(0), unknownStr},
		{Output(9999), unknownStr},
		{Output(-9999), unknownStr},
	}

	for _, tt := range tests {
		if tt.text != tt.o.String() {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.o, tt.text, tt.o.String())
		}
	}
}

func TestOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		o  Output
		ok bool
	}{
		{Value, true},
		{TrendVigor, true},
		{outputLast, false},
		{Output(0), false},
		{Output(9999), false},
		{Output(-9999), false},
	}

	for _, tt := range tests {
		if tt.ok != tt.o.IsKnown() {
			t.Errorf("'%v'.IsKnown(): expected %v, actual %v", tt.o, tt.ok, tt.o.IsKnown())
		}
	}
}

func TestOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		o         Output
		json      string
		succeeded bool
	}{
		{Value, dqs + valueStr + dqs, true},
		{TrendVigor, dqs + trendVigorStr + dqs, true},
		{outputLast, nilstr, false},
		{Output(9999), nilstr, false},
		{Output(-9999), nilstr, false},
		{Output(0), nilstr, false},
	}

	for _, tt := range tests {
		bs, err := tt.o.MarshalJSON()

		if err != nil && tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success, got error %v", tt.o, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tt.o)

			continue
		}

		if string(bs) != tt.json {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tt.o, tt.json, string(bs))
		}
	}
}

func TestOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero Output
	tests := []struct {
		o         Output
		json      string
		succeeded bool
	}{
		{Value, dqs + valueStr + dqs, true},
		{TrendVigor, dqs + trendVigorStr + dqs, true},
		{zero, dqs + unknownStr + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		var o Output

		err := o.UnmarshalJSON([]byte(tt.json))
		if err != nil && tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success, got error %v", tt.json, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected error, got success", tt.json)

			continue
		}

		if tt.o != o {
			t.Errorf("UnmarshalJSON('%v'): expected %v, actual %v", tt.json, tt.o, o)
		}
	}
}
