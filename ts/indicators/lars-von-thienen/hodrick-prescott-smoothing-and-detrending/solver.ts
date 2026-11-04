import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export function hpTrend(values: number[], smoothing: number): number[] {
    let n = values.length;
    if ((n < 6)) {
        throwError("HP requires at least six observations");
    }
    {
        let err = harmoniccycle.validateReal(smoothing, "smoothing", 0);
        if ((err !== null)) {
            throwError(err);
        }
    }
    let a = Array.from({ length: n }, () => 0);
    let b = Array.from({ length: n }, () => 0);
    let c = Array.from({ length: n }, () => 0);
    for (const [i, value] of values.entries()) {
        harmoniccycle.validateSample(value);
        if (Number.isNaN(value)) {
            throwError("HP batch data must be finite");
        }
        a[i] = ((6 * smoothing) + 1);
        b[i] = (-4 * smoothing);
        c[i] = smoothing;
    }
    a[0] = (1 + smoothing);
    a[(n - 1)] = (1 + smoothing);
    a[1] = (1 + (5 * smoothing));
    a[(n - 2)] = (1 + (5 * smoothing));
    b[0] = (-2 * smoothing);
    b[(n - 2)] = (-2 * smoothing);
    b[(n - 1)] = 0;
    c[(n - 1)] = 0;
    c[(n - 2)] = 0;
    let h1: number = 0;
    let h2: number = 0;
    let h3: number = 0;
    let h4: number = 0;
    let h5: number = 0;
    let hh2: number = 0;
    let hh3: number = 0;
    let hh5: number = 0;
    for (let i = 0; i < n; i++) {
        let z = harmoniccycle.finite(((a[i] - (h4 * h1)) - (hh5 * hh2)));
        if ((z <= 0)) {
            throwError("HP elimination lost positive definiteness");
        }
        let hb = b[i];
        let hh1 = h1;
        h1 = (((hb - (h4 * h2))) / z);
        b[i] = h1;
        let hc = c[i];
        hh2 = h2;
        h2 = (hc / z);
        c[i] = h2;
        a[i] = harmoniccycle.finite(((((values[i] - (hh3 * hh5)) - (h3 * h4))) / z));
        hh3 = h3;
        h3 = a[i];
        h4 = (hb - (h5 * hh1));
        hh5 = h5;
        h5 = hc;
    }
    let output = Array.from({ length: n }, () => 0);
    h2 = 0;
    h1 = a[(n - 1)];
    output[(n - 1)] = h1;
    for (let i = (n - 1); (i > 0); i--) {
        output[(i - 1)] = harmoniccycle.finite(((a[(i - 1)] - (b[(i - 1)] * h1)) - (c[(i - 1)] * h2)));
        [h2, h1] = [h1, output[(i - 1)]];
    }
    return output;
}
export function cycleData(values: number[], smoothing: number): number[] {
    if ((smoothing === 0)) {
        return [...[], ...values];
    }
    let trend = hpTrend(values, smoothing);
    for (let i = 0; i < trend.length; i++) {
        trend[i] = harmoniccycle.finite((values[i] - trend[i]));
    }
    return trend;
}
