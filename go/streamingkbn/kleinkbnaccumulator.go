package streamingkbn

import "math"

// Klein second-order Kahan-Babuška-Neumaier (KBN) compensated summation.
//
// Kahan (1965) introduced single-level compensated summation.
// Neumaier (1974) improved it with a branch on |sum| >= |x|
// (the KBN algorithm proper). Klein (2006) generalised KBN
// to arbitrary order; this is the second-order variant, which
// applies the same KBN trick to the correction term itself.
//
//	Level 1 (KBN):      t = sum + x
//	                    if |sum| >= |x|: c = (sum - t) + x
//	                    else:            c = (x - t) + sum
//	                    sum = t
//	Level 2 (Klein):    t = cs + c
//	                    if |cs| >= |c|:  cc = (cs - t) + c
//	                    else:            cc = (c - t) + cs
//	                    cs = t
//	                    ccs += cc
//
// The corrected sum is: sum + cs + ccs.
//
// References:
//   - A. Klein, "A Generalized Kahan-Babuška-Summation-Algorithm",
//     Computing 76, 279-293 (2006).
//   - https://github.com/kuiperzone/Compensated-Accumulators
//   - https://en.wikipedia.org/wiki/Kahan_summation_algorithm

// KleinKBNAccumulator is a Klein second-order Kahan-Babuška-Neumaier (KBN)
// floating-point accumulator. The zero value is an empty accumulator ready
// to use.
//
// It maintains three terms whose sum is the corrected total:
//
//   - sum: the primary (naive) running sum;
//   - cs:  the running sum of first-level KBN corrections;
//   - ccs: the running sum of second-level corrections, i.e. the
//     rounding errors made while accumulating cs (Klein's
//     generalisation).
//
// Unlike naive summation, KBN correctly sums sequences with extreme
// magnitude differences (e.g. Peters' example [1.0, 1e100, 1.0, -1e100]
// → 2.0, while naive and standard Kahan summation return 0.0).
//
// Level 1 (Kahan-Babuška-Neumaier):
//
//	t = sum + x
//	if |sum| >= |x|:  c = (sum - t) + x
//	else:             c = (x - t) + sum
//	sum = t
//
// The branch makes sure the larger operand comes first, so the
// expression recovers exactly the low-order bits that were lost
// when rounding sum + x to t.
//
// Level 2 (Klein generalisation) applies the same technique to the
// addition cs + c and accumulates its rounding error cc into ccs.
//
// The accumulator only stores sums, so Revert(x) (adding -x)
// removes any previously added value, not only the most recent one.
// This makes it suitable for FIFO rolling windows.
type KleinKBNAccumulator struct {
	sum float64
	cs  float64
	ccs float64
}

// Reset sets the accumulator to zero.
func (a *KleinKBNAccumulator) Reset() {
	a.Set(0)
}

// Set overwrites the accumulated value with x and clears both
// compensation terms.
//
// Prefer Set over constructing a new instance when the
// accumulator is stored in an object slot.
func (a *KleinKBNAccumulator) Set(x float64) {
	a.sum = x
	a.cs = 0
	a.ccs = 0
}

// Revert removes a previously added value x (equivalent to Update(-x)).
func (a *KleinKBNAccumulator) Revert(x float64) {
	a.Update(-x)
}

// Update adds x to the accumulator.
func (a *KleinKBNAccumulator) Update(x float64) {
	s := a.sum
	t := s + x
	var c float64
	if math.Abs(s) >= math.Abs(x) {
		c = (s - t) + x
	} else {
		c = (x - t) + s
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
	a.ccs += cc
}

// Value returns the compensated sum of all added values.
func (a *KleinKBNAccumulator) Value() float64 {
	return a.sum + a.cs + a.ccs
}
