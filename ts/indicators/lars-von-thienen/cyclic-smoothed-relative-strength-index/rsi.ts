import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class WilderRSI {
    length: number = 0;
    count: number = 0;
    previous: number = 0;
    up: number = 0;
    down: number = 0;
    constructor(values: Partial<WilderRSI> = {}) { Object.assign(this, values); }
    update(price: number): number {
        if (Number.isNaN(this.previous)) {
            this.previous = price;
            return NaN;
        }
        let change = harmoniccycle.finite((price - this.previous));
        let [gain, loss] = [Math.max(change, 0), Math.max(-change, 0)];
        if ((this.count < this.length)) {
            this.up = harmoniccycle.finite((this.up + gain));
            this.down = harmoniccycle.finite((this.down + loss));
            this.count++;
            this.previous = price;
            if ((this.count < this.length)) {
                return NaN;
            }
            this.up /= this.length;
            this.down /= this.length;
        }
        else {
            this.up = harmoniccycle.finite(((((this.up * (this.length - 1)) + gain)) / this.length));
            this.down = harmoniccycle.finite(((((this.down * (this.length - 1)) + loss)) / this.length));
            this.previous = price;
        }
        if ((this.down === 0)) {
            return 100;
        }
        if ((this.up === 0)) {
            return 0;
        }
        return (100 - (100 / ((1 + (this.up / this.down)))));
    }
}
export function newWilderRSI(length: number): WilderRSI {
    {
        let err = harmoniccycle.validateInteger(length, "length", 1);
        if ((err !== null)) {
            throwError(err);
        }
    }
    return new WilderRSI({ length: length, previous: NaN });
}
