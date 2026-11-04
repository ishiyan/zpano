package annualcyclepatterndecoder

import (
	"math"
	"testing"

	"zpano/indicators/larsvonthienen/annualcyclemodel"
)

func TestPreparedCase0(t *testing.T) {
	p := annualcyclemodel.DefaultParams()
	p.CandidatePool = 10
	p.GenuineThreshold = 49.0
	p.MaxComponents = 7
	p.MaxPeriod = 125
	p.MinPeriod = 5
	p.PeakDistance = 2
	p.Smoothing = 1600.0
	model, err := annualcyclemodel.FitAnnualValues(testInput, 2020, p)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := annualcyclemodel.NewAnnualWaveReader(model)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedP05_P1125_L1600_G49_C10_K7_D2_WAVE
	if math.Abs(model.Correlation-expectedP05_P1125_L1600_G49_C10_K7_D2_CORRELATION[0]) > 1e-13 || float64(len(model.Components)) != expectedP05_P1125_L1600_G49_C10_K7_D2_COMPONENT_COUNT[0] {
		t.Fatal("model diagnostics", model.Correlation, len(model.Components))
	}
	for index, reference := range expected {
		actual := reader.Update()
		if math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	reader.Reset()
	if math.Abs(reader.Update()-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
func TestPreparedCase1(t *testing.T) {
	p := annualcyclemodel.DefaultParams()
	p.CandidatePool = 10
	p.GenuineThreshold = 49.0
	p.MaxComponents = 1
	p.MaxPeriod = 125
	p.MinPeriod = 5
	p.PeakDistance = 2
	p.Smoothing = 1600.0
	model, err := annualcyclemodel.FitAnnualValues(testInput, 2020, p)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := annualcyclemodel.NewAnnualWaveReader(model)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedP05_P1125_L1600_G49_C10_K1_D2_WAVE
	if math.Abs(model.Correlation-expectedP05_P1125_L1600_G49_C10_K1_D2_CORRELATION[0]) > 1e-13 || float64(len(model.Components)) != expectedP05_P1125_L1600_G49_C10_K1_D2_COMPONENT_COUNT[0] {
		t.Fatal("model diagnostics", model.Correlation, len(model.Components))
	}
	for index, reference := range expected {
		actual := reader.Update()
		if math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	reader.Reset()
	if math.Abs(reader.Update()-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
func TestPreparedCase2(t *testing.T) {
	p := annualcyclemodel.DefaultParams()
	p.CandidatePool = 7
	p.GenuineThreshold = 49.0
	p.MaxComponents = 3
	p.MaxPeriod = 80
	p.MinPeriod = 5
	p.PeakDistance = 2
	p.Smoothing = 1600.0
	model, err := annualcyclemodel.FitAnnualValues(testInput, 2020, p)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := annualcyclemodel.NewAnnualWaveReader(model)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedP05_P180_L1600_G49_C7_K3_D2_WAVE
	if math.Abs(model.Correlation-expectedP05_P180_L1600_G49_C7_K3_D2_CORRELATION[0]) > 1e-13 || float64(len(model.Components)) != expectedP05_P180_L1600_G49_C7_K3_D2_COMPONENT_COUNT[0] {
		t.Fatal("model diagnostics", model.Correlation, len(model.Components))
	}
	for index, reference := range expected {
		actual := reader.Update()
		if math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	reader.Reset()
	if math.Abs(reader.Update()-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
func TestPreparedCase3(t *testing.T) {
	p := annualcyclemodel.DefaultParams()
	p.CandidatePool = 10
	p.GenuineThreshold = 49.0
	p.MaxComponents = 7
	p.MaxPeriod = 80
	p.MinPeriod = 5
	p.PeakDistance = 2
	p.Smoothing = 16000.0
	model, err := annualcyclemodel.FitAnnualValues(testInput, 2020, p)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := annualcyclemodel.NewAnnualWaveReader(model)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedP05_P180_L16000_G49_C10_K7_D2_WAVE
	if math.Abs(model.Correlation-expectedP05_P180_L16000_G49_C10_K7_D2_CORRELATION[0]) > 1e-13 || float64(len(model.Components)) != expectedP05_P180_L16000_G49_C10_K7_D2_COMPONENT_COUNT[0] {
		t.Fatal("model diagnostics", model.Correlation, len(model.Components))
	}
	for index, reference := range expected {
		actual := reader.Update()
		if math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	reader.Reset()
	if math.Abs(reader.Update()-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
