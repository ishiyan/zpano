// ########################################################
// Klein second-order Kahan-Babuška-Neumaier (KBN) compensated summation.
//
// Kahan (1965) introduced single-level compensated summation.
// Neumaier (1974) improved it with a branch on |sum| >= |x|
// (the KBN algorithm proper).  Klein (2006) generalised KBN
// to arbitrary order; this is the second-order variant, which
// applies the same KBN trick to the correction term itself.
//
// Level 1 (KBN):      t = sum + x
//                     if |sum| >= |x|: c = (sum - t) + x
//                     else:            c = (x - t) + sum
//                     sum = t
// Level 2 (Klein):    t = cs + c
//                     if |cs| >= |c|:  cc = (cs - t) + c
//                     else:            cc = (c - t) + cs
//                     cs = t
//                     ccs += cc
//
// The corrected sum is: sum + cs + ccs.
//
// References:
//   A. Klein, "A Generalized Kahan-Babuška-Summation-Algorithm",
//     Computing 76, 279-293 (2006).
//   https://github.com/kuiperzone/Compensated-Accumulators
//   https://en.wikipedia.org/wiki/Kahan_summation_algorithm
// ########################################################

/**
 * Klein second-order Kahan-Babuška-Neumaier (KBN) floating-point accumulator.
 *
 * Maintains three terms whose sum is the corrected total:
 *
 * - `_sum`: the primary (naive) running sum;
 * - `_cs`:  the running sum of first-level KBN corrections;
 * - `_ccs`: the running sum of second-level corrections, i.e. the
 *   rounding errors made while accumulating `_cs` (Klein's
 *   generalisation).
 *
 * Unlike naive summation, KBN correctly sums sequences with extreme
 * magnitude differences (e.g. Peters' example [1.0, 1e100, 1.0, -1e100]
 * → 2.0, while naive and standard Kahan summation return 0.0).
 *
 * Level 1 (Kahan-Babuška-Neumaier):
 *
 *     t = sum + x
 *     if |sum| >= |x|:  c = (sum - t) + x
 *     else:             c = (x - t) + sum
 *     sum = t
 *
 * The branch makes sure the larger operand comes first, so the
 * expression recovers exactly the low-order bits that were lost
 * when rounding `sum + x` to `t`.
 *
 * Level 2 (Klein generalisation) applies the same technique to the
 * addition `cs + c` and accumulates its rounding error `cc`
 * into `ccs`.
 *
 * The accumulator only stores sums, so `revert(x)` (adding `-x`)
 * removes any previously added value, not only the most recent one.
 * This makes it suitable for FIFO rolling windows.
 */
export class KleinKbnAccumulator {
    private _sum = 0.0;
    private _cs = 0.0;
    private _ccs = 0.0;

    /** Sets the accumulator to zero. */
    reset(): void {
        this.set(0.0);
    }

    /**
     * Overwrites the accumulated value with x and clears both
     * compensation terms.
     *
     * Prefer set() over constructing a new instance when the
     * accumulator is stored in an object field.
     */
    set(x: number): void {
        this._sum = x;
        this._cs = 0.0;
        this._ccs = 0.0;
    }

    /** Removes a previously added value x (equivalent to update(-x)). */
    revert(x: number): void {
        this.update(-x);
    }

    /** Adds x to the accumulator. */
    update(x: number): void {
        const s = this._sum;
        let t = s + x;
        let c: number;
        if (Math.abs(s) >= Math.abs(x)) {
            c = (s - t) + x;
        } else {
            c = (x - t) + s;
        }
        this._sum = t;

        const cs = this._cs;
        t = cs + c;
        let cc: number;
        if (Math.abs(cs) >= Math.abs(c)) {
            cc = (cs - t) + c;
        } else {
            cc = (c - t) + cs;
        }
        this._cs = t;
        this._ccs += cc;
    }

    /** The compensated sum of all added values. */
    get value(): number {
        return this._sum + this._cs + this._ccs;
    }
}
