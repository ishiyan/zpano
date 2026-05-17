import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { HurstDifferenceParams } from './params';

/** Function to calculate mnemonic of a __HurstDifference__ indicator. */
export const hurstDifferenceMnemonic = (params: HurstDifferenceParams): string =>
    'hurdif('.concat(params.period.toString(), componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Hurst Difference indicator. */
export class HurstDifference extends LineIndicator {
    private window: Array<number>;
    private period: number;
    private windowCount: number;
    private nMinus1: number;
    private log2PM1: number;
    private ln2: number;
    private invNSq: number;
    private prevFgdi: number;
    private _lastFgdi: number;

    /**
     * Constructs an instance given a period.
     * The period should be an integer greater than 1.
     **/
    public constructor(params: HurstDifferenceParams) {
        super();
        const period = Math.floor(params.period);
        if (period < 2) {
            throw new Error('period should be greater than 1');
        }

        this.mnemonic = hurstDifferenceMnemonic(params);
        this.description = 'Hurst difference ' + this.mnemonic;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
        this.window = new Array<number>(period + 1);
        this.period = period;
        this.nMinus1 = period - 1;
        this.windowCount = 0;
        this.primed = false;
        this.log2PM1 = Math.log(2.0 * (period - 1));
        this.ln2 = Math.log(2.0);
        this.invNSq = 1.0 / (period * period);
        this.prevFgdi = Number.NaN;
        this._lastFgdi = Number.NaN;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.HurstDifference,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
                { mnemonic: this.mnemonic + ' fgdi', description: this.description + ' FGDI' },
            ],
        );
    }

    /** Updates the indicator and returns the hurst_diff value. Use fgdiValue for the FGDI output. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const period = this.period;

        if (this.primed) {
            for (let i = 0; i < period; i++) {
                this.window[i] = this.window[i + 1];
            }

            this.window[period] = sample;
        } else {
            this.window[this.windowCount] = sample;
            this.windowCount++;

            if (this.windowCount <= period) {
                return Number.NaN;
            }

            this.primed = true;
        }

        // Use the last `period` elements of the window (indices 1..period inclusive).
        // Find min/max for normalization.
        let priceMax = this.window[1];
        let priceMin = this.window[1];

        for (let k = 2; k <= period; k++) {
            if (this.window[k] > priceMax) { priceMax = this.window[k]; }
            if (this.window[k] < priceMin) { priceMin = this.window[k]; }
        }

        const priceRange = priceMax - priceMin;

        let fgdi: number;

        if (priceRange <= 0.0) {
            fgdi = 0.0;
        } else {
            // Normalize and compute path length.
            let priorNorm = (this.window[1] - priceMin) / priceRange;
            let length = 0.0;

            for (let k = 2; k <= period; k++) {
                const currNorm = (this.window[k] - priceMin) / priceRange;
                const diff = currNorm - priorNorm;
                length += Math.sqrt(diff * diff + this.invNSq);
                priorNorm = currNorm;
            }

            if (length > 0.0) {
                fgdi = 1.0 + (Math.log(length) + this.ln2) / this.log2PM1;
            } else {
                fgdi = 0.0;
            }
        }

        // First difference.
        let hurstDiff: number;
        if (Number.isNaN(this.prevFgdi)) {
            hurstDiff = Number.NaN;
        } else {
            hurstDiff = fgdi - this.prevFgdi;
        }

        this.prevFgdi = fgdi;
        this._lastFgdi = fgdi;

        return hurstDiff;
    }

    /** Updates and returns both outputs: [hurstDiff, fgdi]. */
    public updateAll(sample: number): [number, number] {
        const hurstDiff = this.update(sample);
        return [hurstDiff, this._lastFgdi];
    }

    /** Returns the last computed FGDI value. */
    public get fgdiValue(): number { return this._lastFgdi; }
}
