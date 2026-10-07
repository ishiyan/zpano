package streamingkbn

import "testing"

func TestKleinKBNSummator_Empty(t *testing.T) {
	t.Parallel()

	s := NewKleinKBNSummator()
	assertEqualInt(t, "n", s.N(), 0)
	assertEqual(t, "value", s.Value(), 0.0)
	assertNaN(t, "mean", s.Mean())
}

func TestKleinKBNSummator_Update(t *testing.T) {
	t.Parallel()

	s := NewKleinKBNSummator()
	for _, x := range []float64{1.0, 2.0, 3.0, 4.0} {
		s.Update(x)
	}
	assertEqualInt(t, "n", s.N(), 4)
	assertEqual(t, "value", s.Value(), 10.0)
	assertEqual(t, "mean", s.Mean(), 2.5)
}

func TestKleinKBNSummator_ZeroIsCounted(t *testing.T) {
	t.Parallel()

	s := NewKleinKBNSummator()
	for _, x := range []float64{0.0, 3.0, 0.0} {
		s.Update(x)
	}
	assertEqualInt(t, "n", s.N(), 3)
	assertEqual(t, "value", s.Value(), 3.0)
	assertEqual(t, "mean", s.Mean(), 1.0)
}

func TestKleinKBNSummator_Compensated(t *testing.T) {
	t.Parallel()

	// Peters' example: naive summation yields 0.0.
	s := NewKleinKBNSummator()
	for _, x := range []float64{1.0, 1e100, 1.0, -1e100} {
		s.Update(x)
	}
	assertEqual(t, "value", s.Value(), 2.0)
	assertEqual(t, "mean", s.Mean(), 0.5)
}

func TestKleinKBNSummator_Revert(t *testing.T) {
	t.Parallel()

	s := NewKleinKBNSummator()
	for _, x := range []float64{1.0, 2.0, 0.0, 4.0} {
		s.Update(x)
	}
	s.Revert(0.0)
	assertEqualInt(t, "n", s.N(), 3)
	assertEqual(t, "value", s.Value(), 7.0)
	s.Revert(1.0) // not the most recent value
	assertEqualInt(t, "n", s.N(), 2)
	assertEqual(t, "value", s.Value(), 6.0)
	assertEqual(t, "mean", s.Mean(), 3.0)
}

func TestKleinKBNSummator_RollingWindow(t *testing.T) {
	t.Parallel()

	data := []float64{0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067}
	const w = 3
	s := NewKleinKBNSummator()
	for i, x := range data {
		s.Update(x)
		if i >= w {
			s.Revert(data[i-w])
		}
		win := window(data, i, w)
		assertEqualInt(t, "n", s.N(), len(win))
		assertAlmostEqual(t, "value", s.Value(), fsum(win), 17)
	}
}

func TestKleinKBNSummator_RevertToEmpty(t *testing.T) {
	t.Parallel()

	s := NewKleinKBNSummator()
	s.Update(5.0)
	s.Revert(5.0)
	assertEqualInt(t, "n", s.N(), 0)
	assertEqual(t, "value", s.Value(), 0.0)
	assertNaN(t, "mean", s.Mean())
}

func TestKleinKBNSummator_RevertEmptyRaises(t *testing.T) {
	t.Parallel()

	s := NewKleinKBNSummator()
	assertPanics(t, "Cannot revert from an empty summator", func() { s.Revert(1.0) })
}

func TestKleinKBNSummator_RevertToEmptyClearsCompensation(t *testing.T) {
	t.Parallel()
	for _, finalZero := range []bool{false, true} {
		s := NewKleinKBNSummator()
		for _, x := range []float64{0.1, 1e16, 1e32, 1e48} {
			s.Update(x)
		}
		if finalZero {
			s.Update(0)
		}
		for _, x := range []float64{1e16, 0.1, 1e32, 1e48} {
			s.Revert(x)
		}
		if finalZero {
			assertEqualInt(t, "n before removing zero", s.N(), 1)
			s.Revert(0)
		}
		assertEqualInt(t, "n", s.N(), 0)
		assertEqual(t, "value", s.Value(), 0)
		assertNaN(t, "mean", s.Mean())
		s.Update(3)
		assertEqualInt(t, "n after reuse", s.N(), 1)
		assertEqual(t, "value after reuse", s.Value(), 3)
		assertEqual(t, "mean after reuse", s.Mean(), 3)
	}
}

func TestKleinKBNSummator_Reset(t *testing.T) {
	t.Parallel()

	s := NewKleinKBNSummator()
	for _, x := range []float64{1.0, 2.0} {
		s.Update(x)
	}
	s.Reset()
	assertEqualInt(t, "n", s.N(), 0)
	assertEqual(t, "value", s.Value(), 0.0)
	assertNaN(t, "mean", s.Mean())
	s.Update(3.0)
	assertEqual(t, "mean", s.Mean(), 3.0)
}
