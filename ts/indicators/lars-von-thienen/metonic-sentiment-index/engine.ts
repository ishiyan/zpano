import * as zerolagexponentialmovingaverage from '../zero-lag-exponential-moving-average/smoother.js';
import * as metonic from '../metonic/calendar.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { MetonicSentimentIndexParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    dailySentiment: number = 0;
    smoothedClose: number = 0;
    smoothedOpen: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, dailySentiment: nan, smoothedClose: nan, smoothedOpen: nan });
}
export class Engine {
    p: MetonicSentimentIndexParams = undefined as unknown as MetonicSentimentIndexParams;
    calendar: metonic.Calendar = undefined as unknown as metonic.Calendar;
    delay: metonic.DelayMean = undefined as unknown as metonic.DelayMean;
    lastValue: number = 0;
    samplesSeen: number = 0;
    closeSmoother: zerolagexponentialmovingaverage.ZLEMA = undefined as unknown as zerolagexponentialmovingaverage.ZLEMA;
    openSmoother: zerolagexponentialmovingaverage.ZLEMA = undefined as unknown as zerolagexponentialmovingaverage.ZLEMA;
    previousClose: number = 0;
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.calendar = new metonic.Calendar();
        this.delay = metonic.newDelayMean([this.p.delay1, this.p.delay2, this.p.delay3]);
        this.lastValue = NaN;
        this.samplesSeen = 0;
        this.closeSmoother = zerolagexponentialmovingaverage.newZLEMA(this.p.length);
        this.openSmoother = zerolagexponentialmovingaverage.newZLEMA(this.p.length);
        this.previousClose = NaN;
    }
    primed(): boolean {
        return !Number.isNaN(this.lastValue);
    }
    warmup(): number {
        return ((Math.max(this.p.delay1, this.p.delay2, this.p.delay3) + Math.trunc(((this.p.length - 1)) / 2)) + this.p.mode);
    }
    update(price: number, day: Date | undefined, openPrice: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        if ((this.p.mode === 0)) {
            harmoniccycle.validateSample(openPrice);
            if (Number.isNaN(openPrice)) {
                return emptySnapshot();
            }
        }
        {
            let err = this.calendar.validate(day);
            if ((err !== null)) {
                throwError(err);
            }
        }
        let smoothedClose = this.closeSmoother.update(price);
        let smoothedOpen = NaN;
        let comparison = this.previousClose;
        if ((this.p.mode === 0)) {
            smoothedOpen = this.openSmoother.update(openPrice);
            comparison = smoothedOpen;
        }
        let sentiment = NaN;
        if (((this.samplesSeen >= (Math.trunc(((this.p.length - 1)) / 2) + this.p.mode)) && !Number.isNaN(comparison))) {
            sentiment = -1;
            if ((smoothedClose > comparison)) {
                sentiment = 1;
            }
        }
        this.previousClose = smoothedClose;
        let mean = this.delay.update(sentiment);
        this.lastValue = (3 * mean);
        this.samplesSeen++;
        this.calendar.commit(day);
        return new Snapshot({ value: this.lastValue, dailySentiment: sentiment, smoothedClose: smoothedClose, smoothedOpen: smoothedOpen });
    }
}
