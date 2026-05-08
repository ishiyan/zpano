//nolint:testpackage
package newmovingaverage

import "testing"

func TestNewMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		output Output
		want   string
	}{
		{Value, "value"},
		{outputLast, "unknown"},
		{Output(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.output.String(); got != tt.want {
			t.Errorf("Output(%d).String() = %q, want %q", tt.output, got, tt.want)
		}
	}
}

func TestNewMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	if !Value.IsKnown() {
		t.Error("Value should be known")
	}

	if outputLast.IsKnown() {
		t.Error("outputLast should not be known")
	}

	if Output(99).IsKnown() {
		t.Error("Output(99) should not be known")
	}
}

func TestNewMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	b, err := Value.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(b) != `"value"` {
		t.Errorf("expected \"value\", got %s", string(b))
	}

	_, err = outputLast.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown output")
	}
}

func TestNewMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var o Output

	err := o.UnmarshalJSON([]byte(`"value"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if o != Value {
		t.Errorf("expected Value, got %v", o)
	}

	err = o.UnmarshalJSON([]byte(`"unknown"`))
	if err == nil {
		t.Error("expected error for unknown output")
	}
}
