import { buildMetadata } from '../../core/build-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { LineIndicator } from '../../core/line-indicator';
import { SuperSmootherParams, defaultParams, superSmootherMnemonic } from './params';

export { SuperSmootherParams, defaultParams, superSmootherMnemonic };

function smootherCoefficients(period: number): [number, number, number] {
    const smootherAngle = (Math.SQRT2 * Math.PI) / period;
    const a1 = Math.exp(-smootherAngle);
    const c2 = 2 * a1 * Math.cos(smootherAngle);
    const c3 = -a1 * a1;
    return [1 - c2 - c3, c2, c3];
}

/**
 * Ehlers' two-pole SuperSmoother low-pass filter with a zero at the two-bar Nyquist period.
 */
export class SuperSmoother extends LineIndicator {
    private readonly warmupCount = 2;
    private readonly c1: number;
    private readonly c2: number;
    private readonly c3: number;

    private count = 0;
    private input1 = 0;
    private output1 = 0;
    private output2 = 0;
    private currentValue = Number.NaN;

    public static default(): SuperSmoother {
        return new SuperSmoother(defaultParams());
    }

    public static fromParams(params?: SuperSmootherParams): SuperSmoother {
        return new SuperSmoother(params);
    }

    public constructor(params?: SuperSmootherParams) {
        super();

        const shortest = params?.shortestCyclePeriod ?? 10;
        if (typeof shortest !== 'number' || !Number.isFinite(shortest) || shortest < 2) {
            throw new Error('shortestCyclePeriod must be finite and at least 2 bars');
        }

        this.mnemonic = superSmootherMnemonic(params);
        this.description = 'Super Smoother ' + this.mnemonic;
        this.barComponent = params?.barComponent;
        this.quoteComponent = params?.quoteComponent;
        this.tradeComponent = params?.tradeComponent;

        [this.c1, this.c2, this.c3] = smootherCoefficients(shortest);
        this.reset();
    }

    /** Clears all streaming histories and priming state while preserving configuration. */
    public reset(): void {
        this.count = 0;
        this.input1 = 0;
        this.output1 = 0;
        this.output2 = 0;
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
        return buildMetadata(IndicatorIdentifier.SuperSmoother, this.mnemonic, this.description, [
            { mnemonic: this.mnemonic, description: this.description },
        ]);
    }

    /** Consumes one scalar sample and returns the current SuperSmoother value. */
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

        if (this.count === 0) {
            this.input1 = sample;
            this.output1 = sample;
            this.output2 = sample;
        }

        const output = (this.c1 * (sample + this.input1)) / 2 + this.c2 * this.output1 + this.c3 * this.output2;
        this.input1 = sample;
        this.output2 = this.output1;
        this.output1 = output;
        this.count += 1;

        if (this.count > this.warmupCount) {
            this.primed = true;
            this.currentValue = output;
            return output;
        }
        return Number.NaN;
    }
}
