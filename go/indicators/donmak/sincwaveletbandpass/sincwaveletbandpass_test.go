//nolint:testpackage
package sincwaveletbandpass

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

func checkSeries(t *testing.T, name string, params *Params, inputs, expected []float64) {
	t.Helper()

	swb, err := NewSincWaveletBandpass(params)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}

	if len(inputs) != len(expected) {
		t.Fatalf("%s: length mismatch", name)
	}

	for i := 0; i < len(inputs); i++ {
		value := swb.Update(inputs[i])
		exp := expected[i]

		if math.IsNaN(exp) {
			if !math.IsNaN(value) {
				t.Errorf("%s[%d]: expected NaN, got %v", name, i, value)
			}

			continue
		}

		if math.Abs(value-exp) > tolerance {
			t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, value)
		}
	}
}

func TestSincWaveletBandpassData(t *testing.T) {
	t.Parallel()

	checkSeries(t, "HIGH", &Params{Band: BandHigh}, testInput, expectedHIGH)
	checkSeries(t, "MID", &Params{Band: BandMid}, testInput, expectedMID)
	checkSeries(t, "LOW", &Params{Band: BandLow}, testInput, expectedLOW)
	checkSeries(t, "FULL", &Params{Band: BandFull}, testInput, expectedFULL)
	checkSeries(t, "HIGH_V", &Params{Band: BandHigh, Velocity: true}, testInput, expectedHIGH_V)
	checkSeries(t, "MID_V", &Params{Band: BandMid, Velocity: true}, testInput, expectedMID_V)
	checkSeries(t, "LOW_V", &Params{Band: BandLow, Velocity: true}, testInput, expectedLOW_V)
	checkSeries(t, "FULL_V", &Params{Band: BandFull, Velocity: true}, testInput, expectedFULL_V)

	checkSeries(t, "TEST1_MID", &Params{Band: BandMid}, test1InputSine, test1ExpectedMID)

	checkSeries(t, "TEST2_HIGH_V", &Params{Band: BandHigh, Velocity: true}, test2InputMixed, test2ExpectedHIGH_V)
	checkSeries(t, "TEST2_MID_V", &Params{Band: BandMid, Velocity: true}, test2InputMixed, test2ExpectedMID_V)
	checkSeries(t, "TEST2_LOW_V", &Params{Band: BandLow, Velocity: true}, test2InputMixed, test2ExpectedLOW_V)
}

func TestSincWaveletBandpassMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		params *Params
		want   string
	}{
		{DefaultParams(), "swb(mid)"},
		{&Params{Band: BandHigh}, "swb(high)"},
		{&Params{Band: BandFull}, "swb(full)"},
		{&Params{Band: BandMid, Velocity: true}, "swb(mid,v)"},
		{&Params{Band: BandFull, Velocity: true}, "swb(full,v)"},
	}

	for _, tt := range tests {
		swb, err := NewSincWaveletBandpass(tt.params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if swb.LineIndicator.Mnemonic != tt.want {
			t.Errorf("mnemonic: expected '%s', got '%s'", tt.want, swb.LineIndicator.Mnemonic)
		}
	}
}

func TestSincWaveletBandpassMetadata(t *testing.T) {
	t.Parallel()

	swb, err := NewSincWaveletBandpass(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := swb.Metadata()

	if meta.Identifier != core.SincWaveletBandpass {
		t.Errorf("identifier: expected SincWaveletBandpass, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "swb(mid)" {
		t.Errorf("mnemonic: expected 'swb(mid)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Errorf("outputs: expected 1, got %d", len(meta.Outputs))
	}
}

func TestSincWaveletBandpassInvalidParams(t *testing.T) {
	t.Parallel()

	if _, err := NewSincWaveletBandpass(&Params{Band: Band(99)}); err == nil {
		t.Errorf("expected error for unknown band, got nil")
	}
}
