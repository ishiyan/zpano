// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicwilliamspercentr

import (
	"encoding/json"
	"testing"
)

func TestOutput(t *testing.T) {
	for value := Output(1); value < outputLast; value++ {
		if !value.IsKnown() || value.String() == "unknown" {
			t.Fatal(value)
		}
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var decoded Output
		if err = json.Unmarshal(data, &decoded); err != nil || decoded != value {
			t.Fatal(value, err)
		}
	}
	if Output(0).IsKnown() {
		t.Fatal("zero output is unknown")
	}
	if _, err := Output(0).MarshalJSON(); err == nil {
		t.Fatal("unknown marshal must fail")
	}
	var value Output
	if err := value.UnmarshalJSON([]byte(`"invalid"`)); err == nil {
		t.Fatal("unknown unmarshal must fail")
	}
}
