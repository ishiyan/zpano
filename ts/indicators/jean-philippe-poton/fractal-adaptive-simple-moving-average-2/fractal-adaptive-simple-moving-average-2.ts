import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { FractalAdaptiveSimpleMovingAverage2Params } from './params';

/** Function to calculate mnemonic of a __FractalAdaptiveSimpleMovingAverage2__ indicator. */
export const fractalAdaptiveSimpleMovingAverage2Mnemonic = (params: FractalAdaptiveSimpleMovingAverage2Params): string =>
    'frasma2('.concat(
        params.period.toString(), ',',
        params.normalSpeed.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Fractal Adaptive Simple Moving Average 2 (FRASMA2) line indicator. */
export class FractalAdaptiveSimpleMovingAverage2 extends LineIndicator {
    private _window: Array<number>;
    private _closes: Array<number>;
    private _period: number;
    private _normalSpeed: number;
    private _windowCount: number;
    private _log2Pm1: number;
    private _ln2: number;
    private _invPSq: number;

    /**
     * Constructs an instance given parameters.
     * Period should be an integer greater than 1.
     * NormalSpeed should be an integer greater than 0.
     **/
    public constructor(params: FractalAdaptiveSimpleMovingAverage2Params) {
        super();
        const period = Math.floor(params.period);
        const normalSpeed = Math.floor(params.normalSpeed);

        if (period < 2) {
            throw new Error('period should be greater than 1');
        }

        if (normalSpeed < 1) {
            throw new Error('normal_speed should be greater than 0');
        }

        this.mnemonic = fractalAdaptiveSimpleMovingAverage2Mnemonic(params);
        this.description = 'Fractal adaptive simple moving average 2 ' + this.mnemonic;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
        this._window = new Array<number>(period);
        this._closes = [];
        this._period = period;
        this._normalSpeed = normalSpeed;
        this._windowCount = 0;
        this.primed = false;
        this._log2Pm1 = Math.log(2.0 * (period - 1));
        this._ln2 = Math.log(2.0);
        this._invPSq = 1.0 / (period * period);
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.FractalAdaptiveSimpleMovingAverage2,
            this.mnemonic,
            this.description,
            [{ mnemonic: this.mnemonic, description: this.description }],
        );
    }

    /** Updates the value of the indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const period = this._period;

        // Accumulate close history for SMA computation.
        this._closes.push(sample);

        // Fill the FGDI window. First `period` values are NaN.
        if (this._windowCount < period) {
            this._window[this._windowCount] = sample;
            this._windowCount++;

            if (this._windowCount < period) {
                return Number.NaN;
            }

            this.primed = true;

            return Number.NaN;
        } else {
            for (let i = 0; i < period - 1; i++) {
                this._window[i] = this._window[i + 1];
            }

            this._window[period - 1] = sample;
        }

        // --- Compute FGDI using corrected formula (N-1 segments) ---
        let priceMax = this._window[0];
        let priceMin = this._window[0];

        for (let k = 1; k < period; k++) {
            if (this._window[k] > priceMax) { priceMax = this._window[k]; }
            if (this._window[k] < priceMin) { priceMin = this._window[k]; }
        }

        const priceRange = priceMax - priceMin;
        if (priceRange <= 0.0) {
            return Number.NaN;
        }

        // N-1 segments: iterate from index 1 to period-1.
        let priorNorm = (this._window[0] - priceMin) / priceRange;
        let length = 0.0;

        for (let k = 1; k < period; k++) {
            const currNorm = (this._window[k] - priceMin) / priceRange;
            const diff = currNorm - priorNorm;
            length += Math.sqrt(diff * diff + this._invPSq);
            priorNorm = currNorm;
        }

        if (length <= 0.0) {
            return Number.NaN;
        }

        const fgdi = 1.0 + (Math.log(length) + this._ln2) / this._log2Pm1;

        // --- Adaptive speed ---
        const denom = 2.0 - fgdi;
        if (Math.abs(denom) < 1e-10) {
            return Number.NaN;
        }

        const trailDim = 1.0 / denom;
        const alpha = trailDim / 2.0;
        let speed = Math.max(1, Math.round(this._normalSpeed * alpha));

        // --- SMA of length `speed` ending at current position ---
        const nCloses = this._closes.length;
        if (speed > nCloses) {
            return Number.NaN;
        }

        let smaSum = 0.0;
        for (let k = nCloses - speed; k < nCloses; k++) {
            smaSum += this._closes[k];
        }

        return smaSum / speed;
    }
}
