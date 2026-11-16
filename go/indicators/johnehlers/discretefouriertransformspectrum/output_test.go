//nolint:testpackage
package discretefouriertransformspectrum

import "testing"

func TestOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		output Output
		text   string
	}{
		{Value, valueStr},
		{DominantCycle, dominantCycleStr},
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
		{DominantCycle, true},
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
		{DominantCycle, dqs + dominantCycleStr + dqs, true},
		{outputLast, "", false},
		{Output(9999), "", false},
		{Output(-9999), "", false},
		{Output(0), "", false},
	}

	for _, tc := range tests {
		encoded, err := tc.output.MarshalJSON()
		if err != nil && tc.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success '%v', got error %v", tc.output, tc.jsonText, err)
			continue
		}
		if err == nil && !tc.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tc.output)
			continue
		}
		if string(encoded) != tc.jsonText {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tc.output, tc.jsonText, string(encoded))
		}
	}
}

func TestOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero Output
	tests := []struct {
		expected  Output
		jsonText  string
		succeeded bool
	}{
		{Value, dqs + valueStr + dqs, true},
		{DominantCycle, dqs + dominantCycleStr + dqs, true},
		{zero, dqs + unknownStr + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
		{zero, "", false},
	}

	for _, tc := range tests {
		var actual Output
		err := actual.UnmarshalJSON([]byte(tc.jsonText))
		if err != nil && tc.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success '%v', got error %v", tc.jsonText, tc.expected, err)
			continue
		}
		if err == nil && !tc.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected error, got '%v'", tc.jsonText, actual)
			continue
		}
		if actual != tc.expected {
			t.Errorf("UnmarshalJSON('%v'): expected '%v', actual '%v'", tc.jsonText, tc.expected, actual)
		}
	}
}
