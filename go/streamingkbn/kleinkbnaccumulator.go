// Package streamingkbn implements streaming (online) statistical accumulators
// using Klein second-order Kahan-Babuška-Neumaier (KBN) compensated summation
// for improved numerical stability.
//
// Kahan (1965) introduced single-level compensated summation.
// Neumaier (1974) improved it with a branch on |sum| >= |x|
// (the KBN algorithm proper). Klein (2006) generalised KBN
// to arbitrary order; this is the second-order variant, which
// applies the same KBN trick to the correction term itself.
//
// Level 1 (KBN): t = sum + x; if |sum|>=|x|: c=(sum-t)+x
//
//	else:                c=(x-t)+sum
//
// Level 2 (Klein): same correction applied to cs + c
//
// The corrected sum is: sum + cs + ccs.
//
// References:
//   - https://github.com/kuiperzone/Compensated-Accumulators
//   - https://en.wikipedia.org/wiki/Kahan_summation_algorithm
package streamingkbn

import "math"

// KleinKBNAccumulator implements Klein second-order Kahan-Babuška-Neumaier
// (KBN) floating-point compensated summation.
//
// Maintains sum + cs + ccs where sum is the primary sum, cs is the
// first-level KBN correction, and ccs is a second-level KBN correction
// applied to the first correction term (Klein's generalisation).
//
// Unlike naive summation, KBN correctly sums sequences with extreme
// magnitude differences (e.g. Peters' example [1.0, 1e100, 1.0, -1e100]
// → 2.0, while naive and standard Kahan return 0.0).
//
// Use Set(x) to overwrite the accumulator value (resets both
// compensation terms to zero). Prefer Set over constructing a
// new instance when the accumulator is stored in an object slot.
type KleinKBNAccumulator struct {
	sum  float64
	cs   float64
	ccs  float64
}

// Set overwrites the accumulator value and resets both compensation terms to zero.
func (a *KleinKBNAccumulator) Set(x float64) {
	a.sum = x
	a.cs = 0
	a.ccs = 0
}

// Reset resets the accumulator to zero.
func (a *KleinKBNAccumulator) Reset() {
	a.Set(0)
}

// Revert removes x from the accumulator by adding -x.
func (a *KleinKBNAccumulator) Revert(x float64) {
	a.Update(-x)
}

// Update adds x to the accumulator using Klein second-order KBN compensated summation.
func (a *KleinKBNAccumulator) Update(x float64) {
	sum := a.sum
	t := sum + x

	var c float64
	if math.Abs(sum) >= math.Abs(x) {
		c = (sum - t) + x
	} else {
		c = (x - t) + sum
	}
	a.sum = t

	cs := a.cs
	t = cs + c
	var cc float64
	if math.Abs(cs) >= math.Abs(c) {
		cc = (cs - t) + c
	} else {
		cc = (c - t) + cs
	}
	a.cs = t
	a.ccs = cc
}

// Value returns the current compensated sum: sum + cs + ccs.
func (a *KleinKBNAccumulator) Value() float64 {
	return a.sum + a.cs + a.ccs
}
