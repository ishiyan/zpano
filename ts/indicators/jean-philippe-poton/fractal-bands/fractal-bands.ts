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
import { FractalBandsParams } from './params';

/** Function to calculate mnemonic of a __FractalBands__ indicator. */
export const fractalBandsMnemonic = (params: FractalBandsParams): string =>
    'fban('.concat(
        params.period.toString(), ',',
        params.normalSpeed.toString(), ',',
        params.alpha.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Fractal Bands indicator. */
export class FractalBands extends LineIndicator {
    private window: Array<number>;
    private _closes: Array<number>;
    private _period: number;
    private _periodMinus1: number;
    private _normalSpeed: number;
    private _alpha: number;
    private windowCount: number;
    private logDenom: number;
    private ln2: number;
    private invPeriodSq: number;
    private _frasma2: number;
    private _upperBand: number;
    private _lowerBand: number;

    /**
     * Constructs an instance given parameters.
     * Period should be greater than 1, normalSpeed greater than 0, alpha greater than 0.
     **/
    public constructor(params: FractalBandsParams) {
        super();
        const period = Math.floor(params.period);
        const normalSpeed = Math.floor(params.normalSpeed);
        const alpha = params.alpha;

        if (period < 2) {
            throw new Error('period should be greater than 1');
        }

        if (normalSpeed < 1) {
            throw new Error('normal_speed should be greater than 0');
        }

        if (alpha <= 0.0) {
            throw new Error('alpha should be greater than 0');
        }

        this.mnemonic = fractalBandsMnemonic(params);
        this.description = 'Fractal bands ' + this.mnemonic;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
        this.window = new Array<number>(period);
        this._closes = [];
        this._period = period;
        this._periodMinus1 = period - 1;
        this._normalSpeed = normalSpeed;
        this._alpha = alpha;
        this.windowCount = 0;
        this.primed = false;
        this.logDenom = Math.log(2.0 * (period - 1));
        this.ln2 = Math.log(2.0);
        this.invPeriodSq = 1.0 / (period * period);
        this._frasma2 = Number.NaN;
        this._upperBand = Number.NaN;
        this._lowerBand = Number.NaN;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.FractalBands,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
                { mnemonic: this.mnemonic + ' upper', description: this.description + ' Upper Band' },
                { mnemonic: this.mnemonic + ' lower', description: this.description + ' Lower Band' },
                { mnemonic: this.mnemonic + ' band', description: this.description + ' Band' },
            ],
        );
    }

    /** Updates the indicator and returns the FRASMA2 value. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const period = this._period;
        const periodMinus1 = this._periodMinus1;

        // Accumulate close history for SMA computation.
        this._closes.push(sample);

        // Fill the FGDI window.
        if (this.primed) {
            for (let i = 0; i < periodMinus1; i++) {
                this.window[i] = this.window[i + 1];
            }

            this.window[periodMinus1] = sample;
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

        let fgdi: number;

        if (priceRange <= 0.0) {
            fgdi = 0.0;
        } else {
            // Compute normalized path length: period points, period-1 segments.
            let priorNorm = (this.window[0] - priceMin) / priceRange;
            let length = 0.0;

            for (let k = 1; k < period; k++) {
                const currNorm = (this.window[k] - priceMin) / priceRange;
                const diff = currNorm - priorNorm;
                length += Math.sqrt(diff * diff + this.invPeriodSq);
                priorNorm = currNorm;
            }

            if (length > 0.0) {
                fgdi = 1.0 + (Math.log(length) + this.ln2) / this.logDenom;
            } else {
                fgdi = 0.0;
            }
        }

        // Hurst exponent.
        let hurst = 2.0 - fgdi;
        if (hurst < 0.01) {
            hurst = 0.01;
        }

        const trailDim = 1.0 / hurst;
        const beta = trailDim / 2.0;
        let speed = Math.max(Math.round(this._normalSpeed * beta), 1);

        // FRASMA2: SMA of close over 'speed' bars ending at current position.
        const nCloses = this._closes.length;
        if (speed > nCloses) {
            this._frasma2 = Number.NaN;
            this._upperBand = Number.NaN;
            this._lowerBand = Number.NaN;
            return Number.NaN;
        }

        let smaSum = 0.0;
        for (let k = nCloses - speed; k < nCloses; k++) {
            smaSum += this._closes[k];
        }

        const frasma2Val = smaSum / speed;

        // Deviation over the FGDI lookback window (period bars).
        let sqSum = 0.0;
        for (let k = 0; k < period; k++) {
            const res = this.window[k] - frasma2Val;
            sqSum += res * res;
        }

        const deviation = 2.0 * Math.sqrt(sqSum / period);

        // Fractal bands.
        const bandMult = deviation * Math.pow(this._alpha, hurst);
        const upperBand = frasma2Val + bandMult;
        const lowerBand = frasma2Val - bandMult;

        this._frasma2 = frasma2Val;
        this._upperBand = upperBand;
        this._lowerBand = lowerBand;

        return frasma2Val;
    }

    /** Updates and returns all three outputs: [frasma2, upperBand, lowerBand]. */
    public updateAll(sample: number): [number, number, number] {
        const frasma2 = this.update(sample);
        return [frasma2, this._upperBand, this._lowerBand];
    }

    /** Returns the last computed FRASMA2 value. */
    public get frasma2Value(): number { return this._frasma2; }

    /** Returns the last computed upper band value. */
    public get upperBandValue(): number { return this._upperBand; }

    /** Returns the last computed lower band value. */
    public get lowerBandValue(): number { return this._lowerBand; }

    /** Updates the indicator given the next scalar sample. */
    public updateScalar(sample: Scalar): IndicatorOutput {
        const [frasma2, upper, lower] = this.updateAll(sample.value);

        const s0 = new Scalar(); s0.time = sample.time; s0.value = frasma2;
        const s1 = new Scalar(); s1.time = sample.time; s1.value = upper;
        const s2 = new Scalar(); s2.time = sample.time; s2.value = lower;

        const band = new Band();
        band.time = sample.time;
        if (isNaN(lower) || isNaN(upper)) {
            band.lower = NaN;
            band.upper = NaN;
        } else {
            band.lower = lower;
            band.upper = upper;
        }

        return [s0, s1, s2, band];
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
