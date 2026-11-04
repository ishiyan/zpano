package harmonicenergysuperpositionwave

import (
	"io"
	"math"
	"testing"

	"zpano/indicators/larsvonthienen/annualcyclemodel"
)

func historicalModels(t *testing.T) []annualcyclemodel.Model {
	t.Helper()
	models := make([]annualcyclemodel.Model, 6)
	for j, lag := range MetonicLags {
		values := make([]float64, 252)
		for i := range values {
			values[i] = (1 + 0.1*float64(j)) * testInput[(i+7*j)%252]
		}
		p := annualcyclemodel.DefaultParams()
		p.MaxPeriod = 80
		model, err := annualcyclemodel.FitAnnualValues(values, 2021-lag, p)
		if err != nil {
			t.Fatal(err)
		}
		models[j] = model
	}
	return models
}
func TestPreparedCase0(t *testing.T) {
	models := historicalModels(t)
	reader, err := NewSuperpositionWaveReader(models, 2021, 252, "none", "strict", false)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedN0_C0_I0_WAVE
	for index, reference := range expected {
		actual, err := reader.Update()
		if err != nil || math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	if _, err := reader.Update(); err != io.EOF {
		t.Fatal("declared end")
	}
	reader.Reset()
	actual, err := reader.Update()
	if err != nil || math.Abs(actual-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
func TestPreparedCase1(t *testing.T) {
	models := historicalModels(t)
	reader, err := NewSuperpositionWaveReader(models, 2021, 252, "rms", "strict", false)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedN1_C0_I0_WAVE
	for index, reference := range expected {
		actual, err := reader.Update()
		if err != nil || math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	if _, err := reader.Update(); err != io.EOF {
		t.Fatal("declared end")
	}
	reader.Reset()
	actual, err := reader.Update()
	if err != nil || math.Abs(actual-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
func TestPreparedCase2(t *testing.T) {
	models := historicalModels(t)
	reader, err := NewSuperpositionWaveReader(models, 2021, 252, "rms", "strict", true)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedN1_C0_I1_WAVE
	for index, reference := range expected {
		actual, err := reader.Update()
		if err != nil || math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	if _, err := reader.Update(); err != io.EOF {
		t.Fatal("declared end")
	}
	reader.Reset()
	actual, err := reader.Update()
	if err != nil || math.Abs(actual-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
func TestPreparedCase3(t *testing.T) {
	models := historicalModels(t)
	for i := range models {
		models[i].SampleCount = 250
		models[i].Dates = nil
	}
	reader, err := NewSuperpositionWaveReader(models, 2021, 252, "rms", "extend", false)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedN1_C1_I0_WAVE
	for index, reference := range expected {
		actual, err := reader.Update()
		if err != nil || math.Abs(actual-reference) > math.Max(1e-10, math.Abs(reference)*5e-15) {
			t.Fatalf("index %d expected %.17g got %.17g", index, reference, actual)
		}
	}
	if _, err := reader.Update(); err != io.EOF {
		t.Fatal("declared end")
	}
	reader.Reset()
	actual, err := reader.Update()
	if err != nil || math.Abs(actual-expected[0]) > 1e-10 {
		t.Fatal("reset")
	}
}
