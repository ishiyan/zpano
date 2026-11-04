import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class BartelsOutput {
    score: number = 0;
    genuinePercent: number = 0;
    segments: number = 0;
    constructor(values: Partial<BartelsOutput> = {}) { Object.assign(this, values); }
}
export function bartelsScore(values: number[], period: number, maxSegments: number, minSegments: number): BartelsOutput {
    {
        let err = harmoniccycle.validateInteger(period, "period", 3);
        if ((err !== null)) {
            throwError(err);
        }
    }
    if ((((maxSegments < 2) || (minSegments < 2)) || (minSegments > maxSegments))) {
        throwError("invalid Bartels segments");
    }
    for (const value of values) {
        harmoniccycle.validateSample(value);
        if (Number.isNaN(value)) {
            throwError("Bartels batch data must be finite");
        }
    }
    let segments = Math.min(Math.trunc(values.length / period), maxSegments);
    let nan = NaN;
    if ((segments < minSegments)) {
        return new BartelsOutput({ score: nan, genuinePercent: nan, segments: segments });
    }
    let sumA: number = 0;
    let sumB: number = 0;
    let sumEnergy: number = 0;
    for (let segment = 0; segment < segments; segment++) {
        let a: number = 0;
        let b: number = 0;
        for (let j = 0; j < period; j++) {
            let x = ((segment * period) + j);
            let angle = (((((360 / period) * (x + 1))) / 180) * Math.PI);
            a += (Math.sin(angle) * values[x]);
            b += (Math.cos(angle) * values[x]);
        }
        sumA += a;
        sumB += b;
        sumEnergy = harmoniccycle.finite(((sumEnergy + (a * a)) + (b * b)));
    }
    if ((sumEnergy === 0)) {
        return new BartelsOutput({ score: nan, genuinePercent: nan, segments: segments });
    }
    let [averageA, averageB] = [(sumA / segments), (sumB / segments)];
    let amplitude = Math.sqrt(((averageA * averageA) + (averageB * averageB)));
    let rmsAmplitude = Math.sqrt((sumEnergy / segments));
    let reference = (rmsAmplitude / Math.sqrt(segments));
    let ratio = (amplitude / reference);
    let score = Math.exp((-ratio * ratio));
    return new BartelsOutput({ score: score, genuinePercent: (100 * ((1 - score))), segments: segments });
}
