package factory_test

import (
	"testing"

	"zpano/indicators/core"
	"zpano/indicators/factory"
)

func TestFactoryZeroLagErrorCorrectingExponentialMovingAverage(t *testing.T) {
	t.Parallel()

	defaultIndicator, err := factory.New(core.ZeroLagErrorCorrectingExponentialMovingAverage, "{}")
	if err != nil {
		t.Fatalf("default factory construction failed: %v", err)
	}
	if got := defaultIndicator.Metadata().Mnemonic; got != "zecema(20, 50)" {
		t.Fatalf("default mnemonic = %q, want %q", got, "zecema(20, 50)")
	}
	if got := len(defaultIndicator.Metadata().Outputs); got != 2 {
		t.Fatalf("len(outputs) = %d, want 2", got)
	}

	zeroGainIndicator, err := factory.New(core.ZeroLagErrorCorrectingExponentialMovingAverage, `{"length": 12, "gainLimit": 0, "barComponent": "median"}`)
	if err != nil {
		t.Fatalf("custom length factory construction failed: %v", err)
	}
	if got := zeroGainIndicator.Metadata().Mnemonic; got != "zecema(12, 0, hl/2)" {
		t.Fatalf("custom length mnemonic = %q, want %q", got, "zecema(12, 0, hl/2)")
	}

	alphaIndicator, err := factory.New(core.ZeroLagErrorCorrectingExponentialMovingAverage, `{"smoothingFactor": 0.2, "gainLimit": 22}`)
	if err != nil {
		t.Fatalf("smoothingFactor factory construction failed: %v", err)
	}
	if got := alphaIndicator.Metadata().Mnemonic; got != "zecema(9, 0.20000000, 22)" {
		t.Fatalf("smoothingFactor mnemonic = %q, want %q", got, "zecema(9, 0.20000000, 22)")
	}

	invalidJSONCases := []string{
		`{"gainStep": 0.1}`,
		`{"length": 20, "smoothingFactor": 0.1}`,
		`{"length": 0}`,
		`{"length": 1.5}`,
		`{"gainLimit": -1}`,
		`{"gainLimit": 2.5}`,
		`{"smoothingFactor": 0}`,
	}
	for _, payload := range invalidJSONCases {
		if _, err := factory.New(core.ZeroLagErrorCorrectingExponentialMovingAverage, payload); err == nil {
			t.Fatalf("expected error for payload %s", payload)
		}
	}
}
