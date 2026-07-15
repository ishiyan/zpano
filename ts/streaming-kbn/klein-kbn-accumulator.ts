/**
 * Klein second-order Kahan-Babuška-Neumaier (KBN) floating-point compensated summation.
 *
 * Maintains _sum + _cs + _ccs where _sum is the primary sum, _cs is the
 * first-level KBN correction, and _ccs is a second-level KBN correction
 * applied to the first correction term (Klein's generalisation).
 *
 * Level 1 (KBN):
 *   t = sum + x;  if |sum| >= |x|:  c = (sum - t) + x
 *                  else:             c = (x - t) + sum
 * Level 2 (Klein): same correction applied to cs + c
 *
 * Reference: https://en.wikipedia.org/wiki/Kahan_summation_algorithm
 */
export class KleinKbnAccumulator {
    private _sum = 0.0;
    private _cs = 0.0;
    private _ccs = 0.0;

    /** Overwrites the accumulator value and resets both compensation terms to zero. */
    set(x: number): void {
        this._sum = x;
        this._cs = 0.0;
        this._ccs = 0.0;
    }

    /** Resets the accumulator to zero. */
    reset(): void {
        this.set(0.0);
    }

    /** Returns the current compensated sum: _sum + _cs + _ccs. */
    get value(): number {
        return this._sum + this._cs + this._ccs;
    }

    /** Adds x to the accumulator using Klein second-order KBN compensated summation. */
    update(x: number): void {
        const s = this._sum;
        const t = s + x;
        const c = Math.abs(s) >= Math.abs(x)
            ? (s - t) + x
            : (x - t) + s;
        this._sum = t;

        const cs = this._cs;
        const t2 = cs + c;
        const cc = Math.abs(cs) >= Math.abs(c)
            ? (cs - t2) + c
            : (c - t2) + cs;
        this._cs = t2;
        this._ccs = cc;
    }

    /** Removes x from the accumulator by adding -x. */
    revert(x: number): void {
        this.update(-x);
    }
}
