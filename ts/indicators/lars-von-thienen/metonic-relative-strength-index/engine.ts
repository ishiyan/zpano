import * as cyclicsmoothedrelativestrengthindex from '../cyclic-smoothed-relative-strength-index/rsi.js';
import * as metonic from '../metonic/calendar.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { MetonicRelativeStrengthIndexParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    linearRsi: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, linearRsi: nan });
}
export class Engine {
    p: MetonicRelativeStrengthIndexParams = undefined as unknown as MetonicRelativeStrengthIndexParams;
    calendar: metonic.Calendar = undefined as unknown as metonic.Calendar;
    delay: metonic.DelayMean = undefined as unknown as metonic.DelayMean;
    lastValue: number = 0;
    samplesSeen: number = 0;
    rsi: cyclicsmoothedrelativestrengthindex.WilderRSI = undefined as unknown as cyclicsmoothedrelativestrengthindex.WilderRSI;
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.calendar = new metonic.Calendar();
        this.delay = metonic.newDelayMean([this.p.delay1, this.p.delay2, this.p.delay3]);
        this.lastValue = NaN;
        this.samplesSeen = 0;
        this.rsi = cyclicsmoothedrelativestrengthindex.newWilderRSI(this.p.length);
    }
    primed(): boolean {
        return !Number.isNaN(this.lastValue);
    }
    warmup(): number {
        return (Math.max(this.p.delay1, this.p.delay2, this.p.delay3) + this.p.length);
    }
    update(price: number, day: Date | undefined): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        {
            let err = this.calendar.validate(day);
            if ((err !== null)) {
                throwError(err);
            }
        }
        let linear = this.rsi.update(price);
        this.lastValue = this.delay.update(linear);
        this.samplesSeen++;
        this.calendar.commit(day);
        return new Snapshot({ value: this.lastValue, linearRsi: linear });
    }
}
