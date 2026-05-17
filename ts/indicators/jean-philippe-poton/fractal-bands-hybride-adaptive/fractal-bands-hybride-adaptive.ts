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
import { FractalBandsHybrideAdaptiveParams } from './params';

/** Function to calculate mnemonic of a __FractalBandsHybrideAdaptive__ indicator. */
export const fractalBandsHybrideAdaptiveMnemonic = (params: FractalBandsHybrideAdaptiveParams): string =>
    'fbanha('.concat(
        params.period.toString(), ',',
        params.normalSpeedFallback.toString(), ',',
        params.alpha.toString(), ',',
        params.nyquist.toString(), ',',
        params.alphaHP.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Fractal Bands Hybride Adaptive indicator. */
export class FractalBandsHybrideAdaptive extends LineIndicator {
    private window: Array<number>;
    private _closes: Array<number>;
    private _period: number;
    private _windowSize: number;
    private _normalSpeedFallback: number;
    private _alpha: number;
    private _nyquist: number;
    private _alphaHP: number;
    private windowCount: number;
    private logDenom: number;
    private ln2: number;
    private invPeriodSq: number;
    private _frasma2: number;
    private _upperBand: number;
    private _lowerBand: number;
    // Ehlers CyclePeriod buffers.
    private smoothBuf: Array<number>;
    private cycleBuf: Array<number>;
    private q1Buf: Array<number>;
    private i1Buf: Array<number>;
    private dpBuf: Array<number>;
    private instPeriodBuf: Array<number>;

    /**
     * Constructs an instance given parameters.
     **/
    public constructor(params: FractalBandsHybrideAdaptiveParams) {
        super();
        const period = Math.floor(params.period);
        const normalSpeedFallback = Math.floor(params.normalSpeedFallback);
        const alpha = params.alpha;
        const nyquist = params.nyquist;
        const alphaHP = params.alphaHP;

        if (period < 2) {
            throw new Error('period should be greater than 1');
        }

        if (normalSpeedFallback < 1) {
            throw new Error('normal_speed_fallback should be greater than 0');
        }

        if (alpha <= 0.0) {
            throw new Error('alpha should be greater than 0');
        }

        if (nyquist <= 0.0) {
            throw new Error('nyquist should be greater than 0');
        }

        if (alphaHP <= 0.0 || alphaHP >= 1.0) {
            throw new Error('alpha_hp should be between 0 and 1');
        }

        this.mnemonic = fractalBandsHybrideAdaptiveMnemonic(params);
        this.description = 'Fractal bands hybride adaptive ' + this.mnemonic;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
        this.window = new Array<number>(period + 1);
        this._closes = [];
        this._period = period;
        this._windowSize = period + 1;
        this._normalSpeedFallback = normalSpeedFallback;
        this._alpha = alpha;
        this._nyquist = nyquist;
        this._alphaHP = alphaHP;
        this.windowCount = 0;
        this.primed = false;
        this.logDenom = Math.log(2.0 * (period - 1));
        this.ln2 = Math.log(2.0);
        this.invPeriodSq = 1.0 / (period * period);
        this._frasma2 = Number.NaN;
        this._upperBand = Number.NaN;
        this._lowerBand = Number.NaN;
        this.smoothBuf = [];
        this.cycleBuf = [];
        this.q1Buf = [];
        this.i1Buf = [];
        this.dpBuf = [];
        this.instPeriodBuf = [];
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.FractalBandsHybrideAdaptive,
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

    /** Compute Ehlers CyclePeriod for the current bar. */
    private getCyclePeriod(): number {
        const t = this._closes.length - 1;
        const prices = this._closes;

        // Extend buffers to index t.
        while (this.smoothBuf.length <= t) { this.smoothBuf.push(0.0); }
        while (this.cycleBuf.length <= t) { this.cycleBuf.push(0.0); }
        while (this.q1Buf.length <= t) { this.q1Buf.push(0.0); }
        while (this.i1Buf.length <= t) { this.i1Buf.push(0.0); }
        while (this.dpBuf.length <= t) { this.dpBuf.push(0.0); }
        while (this.instPeriodBuf.length <= t) { this.instPeriodBuf.push(6.0); }

        if (t < 6) {
            return Number.NaN;
        }

        // 4-bar weighted smoother.
        this.smoothBuf[t] = (prices[t] + 2.0 * prices[t - 1] +
            2.0 * prices[t - 2] + prices[t - 3]) / 6.0;

        // High-pass filter.
        const alphaHP = this._alphaHP;
        const hpCoeff = (1.0 - 0.5 * alphaHP) * (1.0 - 0.5 * alphaHP);
        const oneMinusAlpha = 1.0 - alphaHP;

        this.cycleBuf[t] = hpCoeff * (this.smoothBuf[t] - 2.0 * this.smoothBuf[t - 1] + this.smoothBuf[t - 2]) +
            2.0 * oneMinusAlpha * this.cycleBuf[t - 1] - oneMinusAlpha * oneMinusAlpha * this.cycleBuf[t - 2];

        // Quadrature component.
        this.q1Buf[t] = (0.0962 * this.cycleBuf[t] + 0.5769 * this.cycleBuf[t - 2] -
            0.5769 * this.cycleBuf[t - 4] - 0.0962 * this.cycleBuf[t - 6]) *
            (0.5 + 0.08 * this.instPeriodBuf[t - 1]);

        // In-phase component.
        this.i1Buf[t] = this.cycleBuf[t - 3];

        // Smooth I and Q with EMA.
        if (t > 6) {
            this.i1Buf[t] = 0.15 * this.i1Buf[t] + 0.85 * this.i1Buf[t - 1];
            this.q1Buf[t] = 0.15 * this.q1Buf[t] + 0.85 * this.q1Buf[t - 1];
        }

        // Compute delta phase.
        let dp: number;
        if (Math.abs(this.i1Buf[t]) > 1e-10) {
            dp = Math.atan(this.q1Buf[t] / this.i1Buf[t]);
        } else {
            dp = this.dpBuf[t - 1];
        }

        // Clamp delta phase.
        if (dp < 0.1) { dp = 0.1; }
        if (dp > 1.1) { dp = 1.1; }
        this.dpBuf[t] = dp;

        // Median delta phase over 5 bars.
        let medianDP: number;
        if (t >= 10) {
            const w = [this.dpBuf[t - 4], this.dpBuf[t - 3], this.dpBuf[t - 2],
                this.dpBuf[t - 1], this.dpBuf[t]];
            w.sort((a, b) => a - b);
            medianDP = w[2];
        } else {
            medianDP = dp;
        }

        // Instantaneous period.
        let dc: number;
        if (Math.abs(medianDP) > 1e-10) {
            dc = 6.2832 / medianDP + 0.5;
        } else {
            dc = this.instPeriodBuf[t - 1];
        }

        // Clamp and smooth.
        if (dc < 6.0) { dc = 6.0; }
        if (dc > 50.0) { dc = 50.0; }
        this.instPeriodBuf[t] = 0.33 * dc + 0.67 * this.instPeriodBuf[t - 1];

        return this.instPeriodBuf[t];
    }

    /** Updates the indicator and returns the FRASMA2 value. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const period = this._period;
        const windowSize = this._windowSize;

        // Accumulate close history.
        this._closes.push(sample);

        // Update Ehlers CyclePeriod.
        const cp = this.getCyclePeriod();

        // Fill the FGDI window (period+1 elements).
        if (this.primed) {
            for (let i = 0; i < windowSize - 1; i++) {
                this.window[i] = this.window[i + 1];
            }
            this.window[windowSize - 1] = sample;
        } else {
            this.window[this.windowCount] = sample;
            this.windowCount++;

            if (this.windowCount < windowSize) {
                return Number.NaN;
            }

            this.primed = true;
        }

        // FGDI computation over period+1 points.
        let priceMax = this.window[0];
        let priceMin = this.window[0];

        for (let k = 1; k < windowSize; k++) {
            if (this.window[k] > priceMax) { priceMax = this.window[k]; }
            if (this.window[k] < priceMin) { priceMin = this.window[k]; }
        }

        const priceRange = priceMax - priceMin;
        let fgdi: number;

        if (priceRange < 1e-10) {
            fgdi = 1.0;
        } else {
            let length = 0.0;
            for (let i = 1; i < windowSize; i++) {
                const normCur = (this.window[i] - priceMin) / priceRange;
                const normPrev = (this.window[i - 1] - priceMin) / priceRange;
                const diff = normCur - normPrev;
                length += Math.sqrt(diff * diff + this.invPeriodSq);
            }
            fgdi = 1.0 + (Math.log(length) + this.ln2) / this.logDenom;
        }

        // Hurst exponent.
        let hurst = 2.0 - fgdi;
        if (hurst < 0.01) { hurst = 0.01; }

        const trailDim = 1.0 / hurst;
        const beta = trailDim / 2.0;

        // Adaptive normal_speed from CyclePeriod.
        let ns: number;
        if (Number.isNaN(cp) || cp < 1.0) {
            ns = this._normalSpeedFallback;
        } else {
            ns = cp * this._nyquist;
        }

        const speed = Math.max(Math.round(ns * beta), 1);

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

        // Deviation over the last period closes.
        let sqSum = 0.0;
        const devStart = Math.max(nCloses - period, 0);
        for (let k = devStart; k < nCloses; k++) {
            const res = this._closes[k] - frasma2Val;
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
