//nolint:testpackage
package roofedautocorrelationheatmap

import "testing"

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		output Output
		text   string
	}{
		{Value, valueStr},
		{outputLast, unknownStr},
		{Output(0), unknownStr},
		{Output(9999), unknownStr},
		{Output(-9999), unknownStr},
	}

	for _, tc := range tests {
		actual := tc.output.String()
		if actual != tc.text {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tc.output, tc.text, actual)
		}
	}
}

func TestOutputIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		output Output
		known  bool
	}{
		{Value, true},
		{outputLast, false},
		{Output(0), false},
		{Output(9999), false},
		{Output(-9999), false},
	}

	for _, tc := range tests {
		actual := tc.output.IsKnown()
		if actual != tc.known {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tc.output, tc.known, actual)
		}
	}
}

func TestOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	tests := []struct {
		output    Output
		jsonText  string
		succeeded bool
	}{
		{Value, dqs + valueStr + dqs, true},
		{outputLast, "", false},
		{Output(0), "", false},
		{Output(9999), "", false},
		{Output(-9999), "", false},
	}

	for _, tc := range tests {
		data, err := tc.output.MarshalJSON()
		if err != nil && tc.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success '%v', got error %v", tc.output, tc.jsonText, err)
			continue
		}
		if err == nil && !tc.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tc.output)
			continue
		}
		actual := string(data)
		if actual != tc.jsonText {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tc.output, tc.jsonText, actual)
		}
	}
}

func TestOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	tests := []struct {
		output    Output
		jsonText  string
		succeeded bool
	}{
		{Value, dqs + valueStr + dqs, true},
		{Output(0), dqs + unknownStr + dqs, false},
		{Output(0), dqs + "foobar" + dqs, false},
	}

	for _, tc := range tests {
		var actual Output
		err := actual.UnmarshalJSON([]byte(tc.jsonText))
		if err != nil && tc.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success '%v', got error %v", tc.jsonText, tc.output, err)
			continue
		}
		if err == nil && !tc.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected error, got success", tc.jsonText)
			continue
		}
		if actual != tc.output {
			t.Errorf("UnmarshalJSON('%v'): expected '%v', actual '%v'", tc.jsonText, tc.output, actual)
		}
	}
}
