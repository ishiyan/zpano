import * as metonic from '../metonic/calendar.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { MetonicWilliamsPercentRParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    linearPercentR: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, linearPercentR: nan });
}
export class Engine {
    p: MetonicWilliamsPercentRParams = undefined as unknown as MetonicWilliamsPercentRParams;
    calendar: metonic.Calendar = undefined as unknown as metonic.Calendar;
    delay: metonic.DelayMean = undefined as unknown as metonic.DelayMean;
    lastValue: number = 0;
    samplesSeen: number = 0;
    highs: number[] = [];
    lows: number[] = [];
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.calendar = new metonic.Calendar();
        this.delay = metonic.newDelayMean([this.p.delay1, this.p.delay2, this.p.delay3]);
        this.lastValue = NaN;
        this.samplesSeen = 0;
        this.highs = [];
        this.lows = [];
    }
    primed(): boolean {
        return !Number.isNaN(this.lastValue);
    }
    warmup(): number {
        return ((Math.max(this.p.delay1, this.p.delay2, this.p.delay3) + this.p.length) - 1);
    }
    update(price: number, day: Date | undefined, high: number, low: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        harmoniccycle.validateSample(high);
        harmoniccycle.validateSample(low);
        if ((Number.isNaN(high) || Number.isNaN(low))) {
            return emptySnapshot();
        }
        if (((low > price) || (price > high))) {
            throwError("completed HLC bar must satisfy low <= close <= high");
        }
        {
            let err = this.calendar.validate(day);
            if ((err !== null)) {
                throwError(err);
            }
        }
        this.highs = harmoniccycle.push(this.highs, high, this.p.length);
        this.lows = harmoniccycle.push(this.lows, low, this.p.length);
        let linear = NaN;
        if ((this.highs.length === this.p.length)) {
            let [maximum, minimum] = [this.highs[0], this.lows[0]];
            for (const value of this.highs) {
                maximum = Math.max(maximum, value);
            }
            for (const value of this.lows) {
                minimum = Math.min(minimum, value);
            }
            if ((maximum > minimum)) {
                linear = ((-100 * ((maximum - price))) / ((maximum - minimum)));
            }
        }
        this.lastValue = this.delay.update(linear);
        this.samplesSeen++;
        this.calendar.commit(day);
        return new Snapshot({ value: this.lastValue, linearPercentR: linear });
    }
}
