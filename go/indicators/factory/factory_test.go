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

func TestFactoryCombBandPassSpectrum(t *testing.T) {
	t.Parallel()

	defaultIndicator, err := factory.New(core.CombBandPassSpectrum, "{}")
	if err != nil {
		t.Fatalf("default factory construction failed: %v", err)
	}
	if got := defaultIndicator.Metadata().Mnemonic; got != "cbps(10, 48, 0.3, 48, 10)" {
		t.Fatalf("default mnemonic = %q, want %q", got, "cbps(10, 48, 0.3, 48, 10)")
	}
	if got := len(defaultIndicator.Metadata().Outputs); got != 2 {
		t.Fatalf("len(outputs) = %d, want 2", got)
	}
	if got := defaultIndicator.Metadata().Outputs[1].Mnemonic; got != "cbps-dc(10, 48, 0.3, 48, 10)" {
		t.Fatalf("default dc mnemonic = %q, want %q", got, "cbps-dc(10, 48, 0.3, 48, 10)")
	}

	customIndicator, err := factory.New(
		core.CombBandPassSpectrum,
		`{"minPeriod": 8, "maxPeriod": 32, "bandwidth": 0.4, "highpassPeriod": 40, "smoothingPeriod": 10.5, "spectralDilationCompensation": false, "enhanceResolution": true, "barComponent": "median"}`,
	)
	if err != nil {
		t.Fatalf("custom factory construction failed: %v", err)
	}
	if got := customIndicator.Metadata().Mnemonic; got != "cbps(8, 32, 0.4, 40, 10.5, no-sdc, enh, hl/2)" {
		t.Fatalf("custom mnemonic = %q, want %q", got, "cbps(8, 32, 0.4, 40, 10.5, no-sdc, enh, hl/2)")
	}
	if got := customIndicator.Metadata().Outputs[1].Mnemonic; got != "cbps-dc(8, 32, 0.4, 40, 10.5, no-sdc, hl/2)" {
		t.Fatalf("custom dc mnemonic = %q, want %q", got, "cbps-dc(8, 32, 0.4, 40, 10.5, no-sdc, hl/2)")
	}

	invalidJSONCases := []string{
		`{"length": 48}`,
		`{"windowLength": 48}`,
		`{"disableSpectralDilationCompensation": true}`,
		`{"disableAutomaticGainControl": true}`,
		`{"automaticGainControlDecayFactor": 0.995}`,
		`{"fixedNormalization": true}`,
		`{"unknownField": 1}`,
		`{"minPeriod": 10.5}`,
		`{"minPeriod": 2}`,
		`{"maxPeriod": 9}`,
		`{"bandwidth": 0}`,
		`{"bandwidth": 2.5}`,
		`{"highpassPeriod": 4}`,
		`{"smoothingPeriod": 1}`,
		`{"spectralDilationCompensation": 1}`,
		`{"enhanceResolution": "yes"}`,
		`{"minPeriod": null}`,
	}
	for _, payload := range invalidJSONCases {
		if _, err := factory.New(core.CombBandPassSpectrum, payload); err == nil {
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

func TestFactoryAutoCorrelationPeriodogram(t *testing.T) {
	t.Parallel()

	defaultIndicator, err := factory.New(core.AutoCorrelationPeriodogram, "{}")
	if err != nil {
		t.Fatalf("default factory construction failed: %v", err)
	}
	if got := defaultIndicator.Metadata().Mnemonic; got != "acp(8, 48, 3)" {
		t.Fatalf("default mnemonic = %q, want %q", got, "acp(8, 48, 3)")
	}
	if got := len(defaultIndicator.Metadata().Outputs); got != 2 {
		t.Fatalf("len(outputs) = %d, want 2", got)
	}
	if got := defaultIndicator.Metadata().Outputs[1].Mnemonic; got != "acp-dc(8, 48, 3)" {
		t.Fatalf("default dc mnemonic = %q, want %q", got, "acp-dc(8, 48, 3)")
	}

	bookDefaultIndicator, err := factory.New(core.AutoCorrelationPeriodogram, `{"source": "book", "enhanceResolution": true}`)
	if err != nil {
		t.Fatalf("book factory construction failed: %v", err)
	}
	if got := bookDefaultIndicator.Metadata().Mnemonic; got != "acp(10, 48, 3, book, enh)" {
		t.Fatalf("book mnemonic = %q, want %q", got, "acp(10, 48, 3, book, enh)")
	}
	if got := bookDefaultIndicator.Metadata().Outputs[1].Mnemonic; got != "acp-dc(10, 48, 3, book)" {
		t.Fatalf("book dc mnemonic = %q, want %q", got, "acp-dc(10, 48, 3, book)")
	}

	customIndicator, err := factory.New(
		core.AutoCorrelationPeriodogram,
		`{"source": "article", "minPeriod": 10, "maxPeriod": 40, "averagingLength": 0, "enhanceResolution": true, "barComponent": "median"}`,
	)
	if err != nil {
		t.Fatalf("custom factory construction failed: %v", err)
	}
	if got := customIndicator.Metadata().Mnemonic; got != "acp(10, 40, 0, enh, hl/2)" {
		t.Fatalf("custom mnemonic = %q, want %q", got, "acp(10, 40, 0, enh, hl/2)")
	}
	if got := customIndicator.Metadata().Outputs[1].Mnemonic; got != "acp-dc(10, 40, 0, enh, hl/2)" {
		t.Fatalf("custom dc mnemonic = %q, want %q", got, "acp-dc(10, 40, 0, enh, hl/2)")
	}

	invalidJSONCases := []string{
		`{"minPeriod": 0}`,
		`{"minPeriod": 2}`,
		`{"minPeriod": 8.5}`,
		`{"minPeriod": null}`,
		`{"maxPeriod": 4}`,
		`{"maxPeriod": 8}`,
		`{"maxPeriod": 48.5}`,
		`{"averagingLength": 1}`,
		`{"averagingLength": -1}`,
		`{"averagingLength": 3.5}`,
		`{"source": "literal370"}`,
		`{"source": 1}`,
		`{"enhanceResolution": 1}`,
		`{"disableAutomaticGainControl": true}`,
		`{"automaticGainControlDecayFactor": 0.995}`,
		`{"fixedNormalization": true}`,
		`{"unknownField": 1}`,
	}
	for _, payload := range invalidJSONCases {
		if _, err := factory.New(core.AutoCorrelationPeriodogram, payload); err == nil {
			t.Fatalf("expected error for payload %s", payload)
		}
	}
}

func TestFactoryTransformedDiscreteFourierTransformSpectrumStrict(t *testing.T) {
	t.Parallel()

	defaultIndicator, err := factory.New(core.TransformedDiscreteFourierTransformSpectrum, `{}`)
	if err != nil {
		t.Fatalf("default factory construction failed: %v", err)
	}
	if got := defaultIndicator.Metadata().Mnemonic; got != "tdfts(8, 50, 50, 40)" {
		t.Fatalf("default mnemonic = %q, want %q", got, "tdfts(8, 50, 50, 40)")
	}
	if got := len(defaultIndicator.Metadata().Outputs); got != 2 {
		t.Fatalf("len(outputs) = %d, want 2", len(defaultIndicator.Metadata().Outputs))
	}
	if got := defaultIndicator.Metadata().Outputs[1].Mnemonic; got != "tdfts-dc(8, 50, 50, 40)" {
		t.Fatalf("default dc mnemonic = %q, want %q", got, "tdfts-dc(8, 50, 50, 40)")
	}

	customIndicator, err := factory.New(
		core.TransformedDiscreteFourierTransformSpectrum,
		`{"minPeriod": 10, "maxPeriod": 30, "windowLength": 48, "highpassPeriod": 40.5, "barComponent": "median"}`,
	)
	if err != nil {
		t.Fatalf("custom factory construction failed: %v", err)
	}
	if got := customIndicator.Metadata().Mnemonic; got != "tdfts(10, 30, 48, 40.5, hl/2)" {
		t.Fatalf("custom mnemonic = %q, want %q", got, "tdfts(10, 30, 48, 40.5, hl/2)")
	}

	invalidJSONCases := []string{
		`{"minPeriod": 7}`,
		`{"minPeriod": 8.5}`,
		`{"minPeriod": null}`,
		`{"maxPeriod": 7}`,
		`{"windowLength": 0}`,
		`{"windowLength": 10.5}`,
		`{"highpassPeriod": 4.0}`,
		`{"unknownField": 1}`,
	}
	for _, payload := range invalidJSONCases {
		if _, err := factory.New(core.TransformedDiscreteFourierTransformSpectrum, payload); err == nil {
			t.Fatalf("expected error for payload %s", payload)
		}
	}
}

func TestFactoryAutoCorrelationHeatmapStrict(t *testing.T) {
	t.Parallel()

	defaultIndicator, err := factory.New(core.AutoCorrelationHeatmap, `{}`)
	if err != nil {
		t.Fatalf("default factory construction failed: %v", err)
	}
	if got := defaultIndicator.Metadata().Mnemonic; got != "ach(20, 20, 99)" {
		t.Fatalf("default mnemonic = %q, want %q", got, "ach(20, 20, 99)")
	}
	if got := len(defaultIndicator.Metadata().Outputs); got != 1 {
		t.Fatalf("len(outputs) = %d, want 1", got)
	}

	customIndicator, err := factory.New(
		core.AutoCorrelationHeatmap,
		`{"length": 20, "smoothingPeriod": 10.5, "maxLag": 0, "barComponent": "median"}`,
	)
	if err != nil {
		t.Fatalf("custom factory construction failed: %v", err)
	}
	if got := customIndicator.Metadata().Mnemonic; got != "ach(20, 10.5, 0, hl/2)" {
		t.Fatalf("custom mnemonic = %q, want %q", got, "ach(20, 10.5, 0, hl/2)")
	}

	invalidJSONCases := []string{
		`{"length": 1}`,
		`{"length": 2.5}`,
		`{"length": null}`,
		`{"smoothingPeriod": 0.5}`,
		`{"maxLag": -1}`,
		`{"maxLag": 100}`,
		`{"maxLag": 1.5}`,
		`{"unknownField": 1}`,
	}
	for _, payload := range invalidJSONCases {
		if _, err := factory.New(core.AutoCorrelationHeatmap, payload); err == nil {
			t.Fatalf("expected error for payload %s", payload)
		}
	}
}
