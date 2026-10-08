/**
 * Compute the q-th percentile using NumPy's `method="linear"`
 * definition.
 *
 * The input is not modified; a sorted copy is used.
 *
 * @param window Input data.
 * @param q Percentile in the range [0, 1].
 * @returns The interpolated percentile value.
 * @throws Error if q is outside [0, 1] or the window is empty.
 */
export function percentile(window: Iterable<number>, q: number): number {
    if (!(0 <= q && q <= 1)) {
        throw new Error('q must be between 0 and 1');
    }

    const values = Array.from(window).sort((a, b) => a - b);
    if (values.length === 0) {
        throw new Error('window must not be empty');
    }
    const n = values.length;

    if (n === 1) {
        return values[0];
    }

    const idx = q * (n - 1);
    const lo = Math.trunc(idx);

    if (lo >= n - 1) {
        return values[n - 1];
    }

    const hi = lo + 1;
    const frac = idx - lo;

    return values[lo] + frac * (values[hi] - values[lo]);
}
