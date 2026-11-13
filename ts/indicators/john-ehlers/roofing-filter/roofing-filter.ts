import { buildMetadata } from '../../core/build-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { LineIndicator } from '../../core/line-indicator';
import { RoofingFilterParams, defaultParams, roofingFilterMnemonic } from './params';

export { RoofingFilterParams, defaultParams, roofingFilterMnemonic };

function smootherCoefficients(period: number): [number, number, number] {
    const smootherAngle = (Math.SQRT2 * Math.PI) / period;
    const a1 = Math.exp(-smootherAngle);
    const c2 = 2 * a1 * Math.cos(smootherAngle);
    const c3 = -a1 * a1;
    return [1 - c2 - c3, c2, c3];
}

/**
 * Ehlers' Roofing Filter: high-pass detrending followed by a two-pole SuperSmoother.
 */
export class RoofingFilter extends LineIndicator {
    private readonly poles: number;
    private readonly zeroMean: boolean;
    private readonly warmupCount: number;
    private readonly gain: number;
    private readonly decay: number;
    private readonly c1: number;
    private readonly c2: number;
    private readonly c3: number;

    private count = 0;
    private input1 = 0;
    private input2 = 0;
    private hp1 = 0;
    private hp2 = 0;
    private filt1 = 0;
    private filt2 = 0;
    private zero1 = 0;
    private currentValue = Number.NaN;

    public static default(): RoofingFilter {
        return new RoofingFilter(defaultParams());
    }

    public static fromParams(params?: RoofingFilterParams): RoofingFilter {
        return new RoofingFilter(params);
    }

    public constructor(params?: RoofingFilterParams) {
        super();

        const shortest = params?.shortestCyclePeriod ?? 10;
        if (typeof shortest !== 'number' || !Number.isFinite(shortest) || shortest < 2) {
            throw new Error('shortestCyclePeriod must be finite and at least 2 bars');
        }

        const longest = params?.longestCyclePeriod ?? 48;
        if (typeof longest !== 'number' || !Number.isFinite(longest) || longest <= shortest) {
            throw new Error('longestCyclePeriod must be finite and greater than shortestCyclePeriod');
        }

        const poles = params?.highpassPoles ?? 2;
        if (typeof poles !== 'number' || !Number.isInteger(poles) || (poles !== 1 && poles !== 2)) {
            throw new Error('highpassPoles must be integer 1 or 2');
        }

        const zeroMean = params?.zeroMean ?? false;
        if (typeof zeroMean !== 'boolean' || (zeroMean && poles !== 1)) {
            throw new Error('zeroMean must be a boolean and requires highpassPoles === 1');
        }

        this.mnemonic = roofingFilterMnemonic(params);
        this.description = 'Roofing Filter ' + this.mnemonic;
        this.barComponent = params?.barComponent;
        this.quoteComponent = params?.quoteComponent;
        this.tradeComponent = params?.tradeComponent;

        this.poles = poles;
        this.zeroMean = zeroMean;
        this.warmupCount = poles === 2 || zeroMean ? 4 : 3;

        const angle = poles === 1 ? (2 * Math.PI) / longest : (Math.SQRT2 * Math.PI) / longest;
        const tangent = Math.tan(angle / 2);
        const alpha = (2 * tangent) / (1 + tangent);
        this.gain = 1 - alpha / 2;
        this.decay = 1 - alpha;
        [this.c1, this.c2, this.c3] = smootherCoefficients(shortest);

        this.reset();
    }

    /** Clears all streaming histories and priming state while preserving configuration. */
    public reset(): void {
        this.count = 0;
        this.input1 = 0;
        this.input2 = 0;
        this.hp1 = 0;
        this.hp2 = 0;
        this.filt1 = 0;
        this.filt2 = 0;
        this.zero1 = 0;
        this.currentValue = Number.NaN;
        this.primed = false;
    }

    /** Returns the number of initial NaN outputs before priming. */
    public warmupBars(): number {
        return this.warmupCount;
    }

    /** Returns the latest calculated scalar value (NaN before priming). */
    public value(): number {
        return this.currentValue;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(IndicatorIdentifier.RoofingFilter, this.mnemonic, this.description, [
            { mnemonic: this.mnemonic, description: this.description },
        ]);
    }

    /** Consumes one scalar sample and returns the current RoofingFilter value. */
    public update(sample: number): number {
        if (typeof sample !== 'number') {
            throw new Error('sample must be finite or NaN');
        }
        if (Number.isNaN(sample)) {
            return Number.NaN;
        }
        if (!Number.isFinite(sample)) {
            throw new Error('sample must be finite or NaN');
        }

        let hp = 0;
        let filt = 0;
        let zero = 0;

        if (this.count >= this.poles) {
            if (this.poles === 2) {
                const difference = sample - 2 * this.input1 + this.input2;
                hp = this.gain * this.gain * difference + 2 * this.decay * this.hp1 - this.decay * this.decay * this.hp2;
            } else {
                hp = this.gain * (sample - this.input1) + this.decay * this.hp1;
            }
            filt = (this.c1 * (hp + this.hp1)) / 2 + this.c2 * this.filt1 + this.c3 * this.filt2;
            if (this.zeroMean) {
                zero = this.gain * (filt - this.filt1) + this.decay * this.zero1;
            }
        }

        this.input2 = this.input1;
        this.input1 = sample;
        this.hp2 = this.hp1;
        this.hp1 = hp;
        this.filt2 = this.filt1;
        this.filt1 = filt;
        this.zero1 = zero;
        this.count += 1;

        const output = this.zeroMean ? zero : filt;
        if (this.count > this.warmupCount) {
            this.primed = true;
            this.currentValue = output;
            return output;
        }
        return Number.NaN;
    }
}
