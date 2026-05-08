//nolint:testpackage
package advancedecline

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestAdvanceDeclineWithVolume(t *testing.T) {
	t.Parallel()

	const digits = 2

	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()
	expected := testExpectedAD()
	count := len(highs)

	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < count; i++ {
		v := ad.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
		if math.IsNaN(v) {
			t.Errorf("[%d] expected non-NaN, got NaN", i)
			continue
		}

		if !ad.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}

		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestAdvanceDeclineTaLibSpotChecks(t *testing.T) {
	t.Parallel()

	const digits = 2

	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()
	count := len(highs)

	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
	if err != nil {
		t.Fatal(err)
	}

	var values []float64
	for i := 0; i < count; i++ {
		v := ad.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
		values = append(values, v)
	}

	// TA-Lib spot checks from test_per_hlcv.c.
	spotChecks := []struct {
		index    int
		expected float64
	}{
		{0, -1631000.00},
		{1, 2974412.02},
		{250, 8707691.07},
		{251, 8328944.54},
	}

	for _, sc := range spotChecks {
		got := roundTo(values[sc.index], digits)
		exp := roundTo(sc.expected, digits)

		if got != exp {
			t.Errorf("spot check [%d]: expected %v, got %v", sc.index, exp, got)
		}
	}
}

func TestAdvanceDeclineUpdateBar(t *testing.T) {
	t.Parallel()

	const digits = 2

	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()
	expected := testExpectedAD()

	for i := 0; i < 10; i++ {
		bar := &entities.Bar{
			Time:   tm,
			Open:   highs[i], // Open is not used by AD.
			High:   highs[i],
			Low:    lows[i],
			Close:  closes[i],
			Volume: volumes[i],
		}

		output := ad.UpdateBar(bar)
		scalar := output[0].(entities.Scalar)

		got := roundTo(scalar.Value, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] bar: expected %v, got %v", i, exp, got)
		}
	}
}

func TestAdvanceDeclineScalarUpdate(t *testing.T) {
	t.Parallel()

	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
	if err != nil {
		t.Fatal(err)
	}

	// Scalar update: H=L=C, so range=0, AD unchanged (remains 0 after primed).
	v := ad.Update(100.0)
	if v != 0 {
		t.Errorf("expected 0 after scalar update, got %v", v)
	}

	if !ad.IsPrimed() {
		t.Error("expected primed after first update")
	}
}

func TestAdvanceDeclineNaN(t *testing.T) {
	t.Parallel()

	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
	if err != nil {
		t.Fatal(err)
	}

	if v := ad.Update(math.NaN()); !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}

	if v := ad.UpdateHLCV(math.NaN(), 1, 2, 3); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN high, got %v", v)
	}

	if v := ad.UpdateHLCV(1, math.NaN(), 2, 3); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN low, got %v", v)
	}

	if v := ad.UpdateHLCV(1, 2, math.NaN(), 3); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN close, got %v", v)
	}

	if v := ad.UpdateHLCV(1, 2, 3, math.NaN()); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN volume, got %v", v)
	}
}

func TestAdvanceDeclineNotPrimedInitially(t *testing.T) {
	t.Parallel()

	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
	if err != nil {
		t.Fatal(err)
	}

	if ad.IsPrimed() {
		t.Error("expected not primed initially")
	}

	if v := ad.Update(math.NaN()); !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}

	// Still not primed after NaN.
	if ad.IsPrimed() {
		t.Error("expected not primed after NaN")
	}
}

func TestAdvanceDeclineMetadata(t *testing.T) {
	t.Parallel()

	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
	if err != nil {
		t.Fatal(err)
	}

	meta := ad.Metadata()

	if meta.Identifier != core.AdvanceDecline {
		t.Errorf("expected identifier AdvanceDecline, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "ad" {
		t.Errorf("expected mnemonic 'ad', got '%v'", meta.Mnemonic)
	}

	if meta.Description != "Advance-Decline" {
		t.Errorf("expected description 'Advance-Decline', got '%v'", meta.Description)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected output type Scalar, got %v", meta.Outputs[0].Shape)
	}
}
