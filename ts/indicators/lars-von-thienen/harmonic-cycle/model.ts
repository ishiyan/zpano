import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class CycleModel {
    period: number = 0;
    amplitude: number = 0;
    phase: number = 0;
    troughOffset: number = 0;
    constructor(values: Partial<CycleModel> = {}) { Object.assign(this, values); }
}
export function goertzel(values: number[], period: number): CycleModel {
    {
        let err = validateReal(period, "period", 2.0000000000000004);
        if ((err !== null)) {
            throwError(err);
        }
    }
    let n = values.length;
    if ((n < 1)) {
        throwError("Goertzel needs observations");
    }
    let omega = ((2 * Math.PI) / period);
    let [sine, cosine] = [Math.sin(omega), Math.cos(omega)];
    let coefficient = (2 * cosine);
    let q1: number = 0;
    let q2: number = 0;
    for (const value of values) {
        validateSample(value);
        if (Number.isNaN(value)) {
            throwError("Goertzel batch input must be finite");
        }
        let q0 = finite((((coefficient * q1) - q2) + value));
        [q2, q1] = [q1, q0];
    }
    let [r1, i1] = [(q1 - (q2 * cosine)), (q2 * sine)];
    let [cn, sn] = [Math.cos(((n - 1) * omega)), Math.sin(((n - 1) * omega))];
    let real = finite(((r1 * cn) + (i1 * sn)));
    let imaginary = finite(((i1 * cn) - (r1 * sn)));
    let amplitude = finite(((2 * Math.sqrt(finite(((real * real) + (imaginary * imaginary))))) / n));
    if ((amplitude === 0)) {
        return new CycleModel({ period: period, amplitude: 0, phase: NaN, troughOffset: NaN });
    }
    let phase = ((Math.PI / 2) + Math.atan2(imaginary, real));
    let offset = (period * ((0.75 - (phase / ((2 * Math.PI))))));
    return new CycleModel({ period: period, amplitude: amplitude, phase: phase, troughOffset: offset });
}
export function cycleValue(model: CycleModel, offset: number): number {
    if ((model.amplitude === 0)) {
        return 0;
    }
    return (model.amplitude * Math.sin(((((2 * Math.PI) * offset) / model.period) + model.phase)));
}
export function cycleSlope(model: CycleModel, offset: number): number {
    if ((model.amplitude === 0)) {
        return 0;
    }
    let omega = ((2 * Math.PI) / model.period);
    return ((model.amplitude * omega) * Math.cos(((omega * offset) + model.phase)));
}
export function compositeValue(models: CycleModel[], offset: number): number {
    let sum: number = 0;
    for (const model of models) {
        sum += cycleValue(model, offset);
    }
    return finite(sum);
}
export function cycleStrength(amplitude: number, period: number): number {
    {
        let err = validateReal(amplitude, "amplitude", 0);
        if ((err !== null)) {
            throwError(err);
        }
    }
    {
        let err = validateReal(period, "period", 2.0000000000000004);
        if ((err !== null)) {
            throwError(err);
        }
    }
    return (amplitude / period);
}
export function cyclePolarity(shortSlope: number, longSlope: number, tolerance: number): number {
    validateSample(shortSlope);
    validateSample(longSlope);
    {
        let err = validateReal(tolerance, "tolerance", 0);
        if ((err !== null)) {
            throwError(err);
        }
    }
    if ((Number.isNaN(shortSlope) || Number.isNaN(longSlope))) {
        return NaN;
    }
    if (((shortSlope > tolerance) && (longSlope > tolerance))) {
        return 1;
    }
    if (((shortSlope < -tolerance) && (longSlope < -tolerance))) {
        return -1;
    }
    return 0;
}
export function phaseState(phase: number): number {
    validateSample(phase);
    return (100 * Math.sin(phase));
}
export { validateInteger, validateReal, finite, validateSample, push } from "./validation.js";
export { sum } from "./summation.js";
