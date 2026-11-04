import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export function gridBands(values: number[], leveling: number): readonly [
    number,
    number
] {
    let [low, high] = [values[0], values[0]];
    for (const value of values) {
        low = Math.min(low, value);
        high = Math.max(high, value);
    }
    let step = (((high - low)) / 100);
    let fraction = (leveling / 100);
    let [lower, upper] = [low, high];
    for (let k = 0; k < 101; k++) {
        let test = (low + (step * k));
        let below = 0;
        for (const value of values) {
            if ((value < test)) {
                below++;
            }
        }
        if (((below / values.length) >= fraction)) {
            lower = test;
            break;
        }
    }
    for (let k = 0; k < 101; k++) {
        let test = (high - (step * k));
        let above = 0;
        for (const value of values) {
            if ((value >= test)) {
                above++;
            }
        }
        if (((above / values.length) >= fraction)) {
            upper = test;
            break;
        }
    }
    return [lower, upper] as const;
}
