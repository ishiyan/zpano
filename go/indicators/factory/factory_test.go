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

func TestFactoryDiscreteFourierTransformSpectrum(t *testing.T) {
	t.Parallel()

	defaultIndicator, err := factory.New(core.DiscreteFourierTransformSpectrum, "{}")
	if err != nil {
		t.Fatalf("default factory construction failed: %v", err)
	}
	if got := defaultIndicator.Metadata().Mnemonic; got != "dftps(10, 48, 48, 48, 10)" {
		t.Fatalf("default mnemonic = %q, want %q", got, "dftps(10, 48, 48, 48, 10)")
	}
	if got := len(defaultIndicator.Metadata().Outputs); got != 2 {
		t.Fatalf("len(outputs) = %d, want 2", got)
	}

	customIndicator, err := factory.New(
		core.DiscreteFourierTransformSpectrum,
		`{"minPeriod": 8, "maxPeriod": 32, "windowLength": 24, "highpassPeriod": 40, "smoothingPeriod": 10.5, "spectralDilationCompensation": false, "enhanceResolution": true, "barComponent": "median"}`,
	)
	if err != nil {
		t.Fatalf("custom factory construction failed: %v", err)
	}
	if got := customIndicator.Metadata().Mnemonic; got != "dftps(8, 32, 24, 40, 10.5, no-sdc, enh, hl/2)" {
		t.Fatalf("custom mnemonic = %q, want %q", got, "dftps(8, 32, 24, 40, 10.5, no-sdc, enh, hl/2)")
	}
	if got := customIndicator.Metadata().Outputs[1].Mnemonic; got != "dftps-dc(8, 32, 24, 40, 10.5, no-sdc, hl/2)" {
		t.Fatalf("custom dc mnemonic = %q, want %q", got, "dftps-dc(8, 32, 24, 40, 10.5, no-sdc, hl/2)")
	}

	invalidJSONCases := []string{
		`{"length": 48}`,
		`{"spectrumResolution": 1}`,
		`{"disableSpectralDilationCompensation": true}`,
		`{"disableAutomaticGainControl": true}`,
		`{"automaticGainControlDecayFactor": 0.995}`,
		`{"fixedNormalization": true}`,
		`{"unknownField": 1}`,
		`{"minPeriod": 10.5}`,
		`{"minPeriod": 2}`,
		`{"maxPeriod": 9}`,
		`{"windowLength": 1}`,
		`{"windowLength": 12.5}`,
		`{"highpassPeriod": 4}`,
		`{"smoothingPeriod": 1}`,
		`{"spectralDilationCompensation": 1}`,
		`{"enhanceResolution": "yes"}`,
		`{"minPeriod": null}`,
	}
	for _, payload := range invalidJSONCases {
		if _, err := factory.New(core.DiscreteFourierTransformSpectrum, payload); err == nil {
			t.Fatalf("expected error for payload %s", payload)
		}
	}
}
