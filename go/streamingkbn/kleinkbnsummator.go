package streamingkbn

import "math"

// KleinKBNSummator is a counted compensated sum: a [KleinKBNAccumulator]
// plus a sample count, which also gives the arithmetic mean.
//
// See [KleinKBNAccumulator] for the summation algorithm.
//
// Every call to Update increments the count, including calls with
// x == 0; every call to Revert decrements it. Because only sums are
// stored, Revert may remove any previously added value (not only
// the most recent one), so the summator works for FIFO rolling windows.
type KleinKBNSummator struct {
	n   int
	sum KleinKBNAccumulator
}

// NewKleinKBNSummator returns a new empty summator.
func NewKleinKBNSummator() *KleinKBNSummator {
	return &KleinKBNSummator{}
}

// Reset clears the count and the sum.
func (s *KleinKBNSummator) Reset() {
	s.n = 0
	s.sum.Reset()
}

// Revert removes a previously added value x.
// Removing the final sample clears the sum and its compensation terms.
//
// Panics if the summator is empty.
func (s *KleinKBNSummator) Revert(x float64) {
	if s.n <= 0 {
		panic("Cannot revert from an empty summator")
	}
	if s.n == 1 {
		s.Reset()
		return
	}
	s.n--
	// Adding zero leaves the accumulator unchanged, so skip it.
	if x != 0 {
		s.sum.Revert(x)
	}
}

// Update adds a value x.
func (s *KleinKBNSummator) Update(x float64) {
	s.n++
	// Adding zero leaves the accumulator unchanged, so skip it.
	if x != 0 {
		s.sum.Update(x)
	}
}

// Value returns the compensated sum of all added values (0.0 when empty).
func (s *KleinKBNSummator) Value() float64 {
	return s.sum.Value()
}

// Mean returns the arithmetic mean, sum / n (NaN when empty).
func (s *KleinKBNSummator) Mean() float64 {
	n := s.n
	if n <= 0 {
		return math.NaN()
	}
	return s.sum.Value() / float64(n)
}

// N returns the number of added values.
func (s *KleinKBNSummator) N() int {
	return s.n
}
