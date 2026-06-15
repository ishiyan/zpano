//nolint:testpackage
package mexicanhatwavelet

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

func checkSeries(t *testing.T, name string, params *Params, inputs, expected []float64) {
	t.Helper()

	mhw, err := NewMexicanHatWavelet(params)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}

	if len(inputs) != len(expected) {
		t.Fatalf("%s: length mismatch", name)
	}

	for i := 0; i < len(inputs); i++ {
		value := mhw.Update(inputs[i])
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

func TestMexicanHatWaveletData(t *testing.T) {
	t.Parallel()

	checkSeries(t, "HIGH", &Params{Band: BandHigh}, testInput, expectedHIGH)
	checkSeries(t, "MID", &Params{Band: BandMid}, testInput, expectedMID)
	checkSeries(t, "LOW", &Params{Band: BandLow}, testInput, expectedLOW)
	checkSeries(t, "P8", &Params{Band: BandCustom, Period: 8.0}, testInput, expectedP8)
	checkSeries(t, "P20", &Params{Band: BandCustom, Period: 20.0}, testInput, expectedP20)
	checkSeries(t, "P32", &Params{Band: BandCustom, Period: 32.0}, testInput, expectedP32)
	checkSeries(t, "D2_0", &Params{Band: BandCustom, Dilation: 2.0}, testInput, expectedD2_0)
	checkSeries(t, "D8_0", &Params{Band: BandCustom, Dilation: 8.0}, testInput, expectedD8_0)

	checkSeries(t, "TEST1_MID", &Params{Band: BandMid}, test1InputSine, test1ExpectedMID)

	checkSeries(t, "TEST2_HIGH", &Params{Band: BandHigh}, test2InputMixed, test2ExpectedHIGH)
	checkSeries(t, "TEST2_MID", &Params{Band: BandMid}, test2InputMixed, test2ExpectedMID)
	checkSeries(t, "TEST2_LOW", &Params{Band: BandLow}, test2InputMixed, test2ExpectedLOW)
}

func TestMexicanHatWaveletMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		params *Params
		want   string
	}{
		{DefaultParams(), "mhw(mid)"},
		{&Params{Band: BandHigh}, "mhw(high)"},
		{&Params{Band: BandLow}, "mhw(low)"},
		{&Params{Band: BandCustom, Dilation: 2.0}, "mhw(d2.00)"},
		{&Params{Band: BandCustom, Period: 20.0}, "mhw(p20.00)"},
	}

	for _, tt := range tests {
		mhw, err := NewMexicanHatWavelet(tt.params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mhw.LineIndicator.Mnemonic != tt.want {
			t.Errorf("mnemonic: expected '%s', got '%s'", tt.want, mhw.LineIndicator.Mnemonic)
		}
	}
}

func TestMexicanHatWaveletMetadata(t *testing.T) {
	t.Parallel()

	mhw, err := NewMexicanHatWavelet(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := mhw.Metadata()

	if meta.Identifier != core.MexicanHatWavelet {
		t.Errorf("identifier: expected MexicanHatWavelet, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "mhw(mid)" {
		t.Errorf("mnemonic: expected 'mhw(mid)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Errorf("outputs: expected 1, got %d", len(meta.Outputs))
	}
}

func TestMexicanHatWaveletInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"custom no params", &Params{Band: BandCustom}},
		{"custom both params", &Params{Band: BandCustom, Dilation: 2.0, Period: 20.0}},
		{"custom period too small", &Params{Band: BandCustom, Period: 2.0}},
		{"custom dilation nonpositive", &Params{Band: BandCustom, Dilation: -1.0}},
		{"unknown band", &Params{Band: Band(99)}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewMexicanHatWavelet(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
