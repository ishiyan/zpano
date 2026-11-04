import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class ZLEMA {
    alpha: number = 0;
    first: number = 0;
    value: number = 0;
    lag: number = 0;
    history: number[] = [];
    constructor(values: Partial<ZLEMA> = {}) { Object.assign(this, values); }
    update(price: number): number {
        if (Number.isNaN(this.first)) {
            this.first = price;
            this.value = price;
            this.history = harmoniccycle.push(this.history, price, (this.lag + 1));
            return price;
        }
        let history = [...[...[], ...this.history], price];
        let lagged = this.first;
        if ((history.length > this.lag)) {
            lagged = history[((history.length - 1) - this.lag)];
        }
        let adjusted = harmoniccycle.finite(((2 * price) - lagged));
        let value = harmoniccycle.finite(((this.alpha * adjusted) + (((1 - this.alpha)) * this.value)));
        this.history = harmoniccycle.push(this.history, price, (this.lag + 1));
        this.value = value;
        return value;
    }
}
export function newZLEMA(length: number): ZLEMA {
    {
        let err = harmoniccycle.validateInteger(length, "length", 1);
        if ((err !== null)) {
            throwError(err);
        }
    }
    return new ZLEMA({ alpha: (2 / (length + 1)), lag: Math.trunc(((length - 1)) / 2), first: NaN, value: NaN, history: [] as number[] });
}
