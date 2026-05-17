import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { Band } from '../../core/outputs/band';
import { Scalar } from '../../../entities/scalar';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Trade } from '../../../entities/trade';
import { FractalGraphDimensionIndexParams } from './params';

/** Function to calculate mnemonic of a __FractalGraphDimensionIndex__ indicator. */
export const fractalGraphDimensionIndexMnemonic = (params: FractalGraphDimensionIndexParams): string =>
    'fgdi('.concat(params.period.toString(), componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Fractal Graph Dimension Index indicator. */
export class FractalGraphDimensionIndex extends LineIndicator {
    private window: Array<number>;
    private period: number;
    private windowCount: number;
    private nMinus1: number;
    private log2N1: number;
    private ln2: number;
    private invNSq: number;
    private _fgdi: number;
    private _upper: number;
    private _lower: number;
    private _stddev: number;

    /**
     * Constructs an instance given a period.
     * The period should be an integer greater than 1.
     **/
    public constructor(params: FractalGraphDimensionIndexParams) {
        super();
        const period = Math.floor(params.period);
        if (period < 2) {
            throw new Error('period should be greater than 1');
        }

        this.mnemonic = fractalGraphDimensionIndexMnemonic(params);
        this.description = 'Fractal graph dimension index ' + this.mnemonic;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
        this.window = new Array<number>(period);
        this.period = period;
        this.nMinus1 = period - 1;
        this.windowCount = 0;
        this.primed = false;
        this.log2N1 = Math.log(2.0 * (period - 1));
        this.ln2 = Math.log(2.0);
        this.invNSq = 1.0 / (period * period);
        this._fgdi = Number.NaN;
        this._upper = Number.NaN;
        this._lower = Number.NaN;
        this._stddev = Number.NaN;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.FractalGraphDimensionIndex,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
                { mnemonic: this.mnemonic + ' upper', description: this.description + ' Upper' },
                { mnemonic: this.mnemonic + ' lower', description: this.description + ' Lower' },
                { mnemonic: this.mnemonic + ' stddev', description: this.description + ' Stddev' },
                { mnemonic: this.mnemonic + ' band', description: this.description + ' Band' },
            ],
        );
    }

    /** Updates the indicator and returns the FGDI value. Use fgdiValue, upperValue, lowerValue, stddevValue for all outputs. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const period = this.period;
        const nMinus1 = this.nMinus1;

        if (this.primed) {
            for (let i = 0; i < nMinus1; i++) {
                this.window[i] = this.window[i + 1];
            }

            this.window[nMinus1] = sample;
        } else {
            this.window[this.windowCount] = sample;
            this.windowCount++;

            if (this.windowCount < period) {
                return Number.NaN;
            }

            this.primed = true;
        }

        // Find min/max for normalization.
        let priceMax = this.window[0];
        let priceMin = this.window[0];

        for (let k = 1; k < period; k++) {
            if (this.window[k] > priceMax) { priceMax = this.window[k]; }
            if (this.window[k] < priceMin) { priceMin = this.window[k]; }
        }

        const priceRange = priceMax - priceMin;
        if (priceRange < 1e-10) {
            this._fgdi = 1.0;
            this._stddev = 0.0;
            this._upper = 1.0;
            this._lower = 1.0;
            return 1.0;
        }

        // Normalize and compute path segments.
        let priorNorm = (this.window[0] - priceMin) / priceRange;
        let length = 0.0;
        const segments = new Array<number>(nMinus1);

        for (let k = 1; k < period; k++) {
            const currNorm = (this.window[k] - priceMin) / priceRange;
            const diff = currNorm - priorNorm;
            const seg = Math.sqrt(diff * diff + this.invNSq);
            segments[k - 1] = seg;
            length += seg;
            priorNorm = currNorm;
        }

        // FGDI = 1 + (ln(L) + ln(2)) / ln(2*(N-1))
        const fgdi = 1.0 + (Math.log(length) + this.ln2) / this.log2N1;

        // Standard deviation of the estimate.
        const meanSeg = length / nMinus1;
        let sumSq = 0.0;

        for (let k = 0; k < nMinus1; k++) {
            const d = segments[k] - meanSeg;
            sumSq += d * d;
        }

        const variance = sumSq / (length * length * this.log2N1 * this.log2N1);
        const stddev = Math.sqrt(variance);

        this._fgdi = fgdi;
        this._upper = fgdi + stddev;
        this._lower = fgdi - stddev;
        this._stddev = stddev;

        return fgdi;
    }

    /** Updates and returns all four outputs: [fgdi, upper, lower, stddev]. */
    public updateAll(sample: number): [number, number, number, number] {
        const fgdi = this.update(sample);
        return [fgdi, this._upper, this._lower, this._stddev];
    }

    /** Returns the last computed FGDI value. */
    public get fgdiValue(): number { return this._fgdi; }

    /** Returns the last computed upper band value. */
    public get upperValue(): number { return this._upper; }

    /** Returns the last computed lower band value. */
    public get lowerValue(): number { return this._lower; }

    /** Returns the last computed stddev value. */
    public get stddevValue(): number { return this._stddev; }

    /** Updates the indicator given the next scalar sample. */
    public updateScalar(sample: Scalar): IndicatorOutput {
        const [fgdi, upper, lower, stddev] = this.updateAll(sample.value);

        const s0 = new Scalar(); s0.time = sample.time; s0.value = fgdi;
        const s1 = new Scalar(); s1.time = sample.time; s1.value = upper;
        const s2 = new Scalar(); s2.time = sample.time; s2.value = lower;
        const s3 = new Scalar(); s3.time = sample.time; s3.value = stddev;

        const band = new Band();
        band.time = sample.time;
        if (isNaN(lower) || isNaN(upper)) {
            band.lower = NaN;
            band.upper = NaN;
        } else {
            band.lower = lower;
            band.upper = upper;
        }

        return [s0, s1, s2, s3, band];
    }

    /** Updates the indicator given the next bar sample. */
    public updateBar(sample: Bar): IndicatorOutput {
        const v = this.barComponentFunc(sample);
        const scalar = new Scalar();
        scalar.time = sample.time;
        scalar.value = v;
        return this.updateScalar(scalar);
    }

    /** Updates the indicator given the next quote sample. */
    public updateQuote(sample: Quote): IndicatorOutput {
        const v = this.quoteComponentFunc(sample);
        const scalar = new Scalar();
        scalar.time = sample.time;
        scalar.value = v;
        return this.updateScalar(scalar);
    }

    /** Updates the indicator given the next trade sample. */
    public updateTrade(sample: Trade): IndicatorOutput {
        const v = this.tradeComponentFunc(sample);
        const scalar = new Scalar();
        scalar.time = sample.time;
        scalar.value = v;
        return this.updateScalar(scalar);
    }
}
