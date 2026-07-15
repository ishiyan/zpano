package streamingkbn

import (
	"math"
	"testing"
)

func TestRawMomentsKleinKBN_SimpleUpdate(t *testing.T) {
	t.Parallel()

	m := NewRawMomentsKleinKBN(0, true, true)
	for _, x := range []float64{1.0, 2.0, 3.0, 4.0} {
		m.Update(x)
	}
	if !almostEqual(m.Mean(), 2.5, 1e-15) {
		t.Errorf("mean = %v, want 2.5", m.Mean())
	}
	if !almostEqual(m.Variance(), 1.25, 1e-15) {
		t.Errorf("var = %v, want 1.25", m.Variance())
	}
	if !almostEqual(m.Skewness(), 0.0, 1e-14) {
		t.Errorf("skew = %v, want 0.0", m.Skewness())
	}
	if !almostEqual(m.Kurtosis(), -1.36, 1e-13) {
		t.Errorf("kurt = %v, want -1.36", m.Kurtosis())
	}
}

func TestRawMomentsKleinKBN_StandardDeviation(t *testing.T) {
	t.Parallel()

	m := NewRawMomentsKleinKBN(0, true, true)
	for _, x := range []float64{1.0, 2.0, 3.0, 4.0} {
		m.Update(x)
	}
	expected := math.Sqrt(m.Variance())
	if !almostEqual(m.StandardDeviation(), expected, 1e-15) {
		t.Errorf("std = %v, want %v", m.StandardDeviation(), expected)
	}
}

func TestRawMomentsKleinKBN_Ddof(t *testing.T) {
	t.Parallel()

	m := NewRawMomentsKleinKBN(1, true, true)
	for _, x := range []float64{1.0, 2.0, 3.0} {
		m.Update(x)
	}
	if !almostEqual(m.Variance(), 1.0, 1e-15) {
		t.Errorf("var = %v, want 1.0", m.Variance())
	}
}

func TestRawMomentsKleinKBN_CompareScipy(t *testing.T) {
	t.Parallel()

	baconData := []float64{
		0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067,
		-0.014, 0.040, -0.005, 0.081, 0.040, -0.037, -0.061, 0.017,
		-0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009,
	}
	m := NewRawMomentsKleinKBN(0, true, true)
	for _, x := range baconData {
		m.Update(x)
	}
	if !almostEqual(m.Mean(), 0.009000000000000001, 1e-15) {
		t.Errorf("mean = %v, want 0.009000000000000001", m.Mean())
	}
	if !almostEqual(m.Variance(), 0.0014989166666666666, 1e-14) {
		t.Errorf("var = %v, want 0.0014989166666666666", m.Variance())
	}
	if !almostEqual(m.Skewness(), -0.08256245520856798, 1e-14) {
		t.Errorf("skew = %v, want -0.08256245520856798", m.Skewness())
	}
	if !almostEqual(m.Kurtosis(), -0.5675462058921257, 1e-13) {
		t.Errorf("kurt = %v, want -0.5675462058921257", m.Kurtosis())
	}
}

func TestRawMomentsKleinKBN_CompareScipyBiasFalse(t *testing.T) {
	t.Parallel()

	baconData := []float64{
		0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067,
		-0.014, 0.040, -0.005, 0.081, 0.040, -0.037, -0.061, 0.017,
		-0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009,
	}
	m := NewRawMomentsKleinKBN(0, false, true)
	for _, x := range baconData {
		m.Update(x)
	}
	if !almostEqual(m.Skewness(), -0.08817174934967527, 1e-14) {
		t.Errorf("skew = %v, want -0.08817174934967527", m.Skewness())
	}
	if !almostEqual(m.Kurtosis(), -0.40766032118608714, 1e-13) {
		t.Errorf("kurt = %v, want -0.40766032118608714", m.Kurtosis())
	}
}

func TestRawMomentsKleinKBN_RevertRoundtrip(t *testing.T) {
	t.Parallel()

	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	m := NewRawMomentsKleinKBN(0, true, true)
	for _, x := range data {
		m.Update(x)
	}
	for i := len(data) - 2; i >= 0; i-- {
		m.Revert(data[i+1])
	}
	if m.n != 1 {
		t.Errorf("n = %d, want 1", m.n)
	}
	if !almostEqual(m.Mean(), 1.0, 1e-15) {
		t.Errorf("mean = %v, want 1.0", m.Mean())
	}
}

func TestRawMomentsKleinKBN_RevertPartial(t *testing.T) {
	t.Parallel()

	data := []float64{10.0, 18.0, 5.0, 12.0, 7.0}
	mFull := NewRawMomentsKleinKBN(0, true, true)
	mPart := NewRawMomentsKleinKBN(0, true, true)
	for _, x := range data {
		mFull.Update(x)
	}
	for _, x := range data[:4] {
		mPart.Update(x)
	}
	mFull.Revert(data[4])

	if !almostEqual(mFull.Mean(), mPart.Mean(), 1e-15) {
		t.Errorf("mean full=%v part=%v", mFull.Mean(), mPart.Mean())
	}
	if !almostEqual(mFull.Variance(), mPart.Variance(), 1e-15) {
		t.Errorf("var full=%v part=%v", mFull.Variance(), mPart.Variance())
	}
	if !almostEqual(mFull.Skewness(), mPart.Skewness(), 1e-14) {
		t.Errorf("skew full=%v part=%v", mFull.Skewness(), mPart.Skewness())
	}
	if !almostEqual(mFull.Kurtosis(), mPart.Kurtosis(), 1e-13) {
		t.Errorf("kurt full=%v part=%v", mFull.Kurtosis(), mPart.Kurtosis())
	}
}

func TestRawMomentsKleinKBN_Reset(t *testing.T) {
	t.Parallel()

	m := NewRawMomentsKleinKBN(0, true, true)
	m.Update(10.0)
	m.Reset()
	if m.n != 0 {
		t.Errorf("n = %d, want 0", m.n)
	}
	if !math.IsNaN(m.Variance()) {
		t.Errorf("var = %v, want NaN", m.Variance())
	}
}
