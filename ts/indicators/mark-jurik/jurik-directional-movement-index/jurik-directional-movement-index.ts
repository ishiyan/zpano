import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { JurikMovingAverage } from '../jurik-moving-average/jurik-moving-average';
import { JurikDirectionalMovementIndexParams } from './params';

/**
 * Jurik Directional Movement Index (DMX).
 *
 * Produces three output lines:
 *   - Bipolar: 100*(Plus-Minus)/(Plus+Minus)
 *   - Plus: JMA(upward) / JMA(TrueRange)
 *   - Minus: JMA(downward) / JMA(TrueRange)
 *
 * The internal JMA instances use phase=-100 (maximum lag, no overshoot).
 */
export class JurikDirectionalMovementIndex implements Indicator {
    private readonly _mnemonic: string;
    private readonly _description: string;
    private _primed = false;
    private bar = 0;
    private prevHigh = NaN;
    private prevLow = NaN;
    private prevClose = NaN;
    private readonly jmaPlus: JurikMovingAverage;
    private readonly jmaMinus: JurikMovingAverage;
    private readonly jmaDenom: JurikMovingAverage;
    private plusVal = NaN;
    private minusVal = NaN;
    private bipolarVal = NaN;

    constructor(params: JurikDirectionalMovementIndexParams) {
        const length = params.length;

        if (length < 1) {
            throw new Error('invalid jurik directional movement index parameters: length should be positive');
        }

        this._mnemonic = `dmx(${length})`;
        this._description = 'Jurik directional movement index ' + this._mnemonic;

        const jmaParams = { length, phase: -100 };
        this.jmaPlus = new JurikMovingAverage(jmaParams);
        this.jmaMinus = new JurikMovingAverage(jmaParams);
        this.jmaDenom = new JurikMovingAverage(jmaParams);
    }

    /** The indicator mnemonic. */
    public get mnemonic(): string {
        return this._mnemonic;
    }

    /** The indicator description. */
    public get description(): string {
        return this._description;
    }

    /** Indicates whether the indicator is primed. */
    public isPrimed(): boolean {
        return this._primed;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikDirectionalMovementIndex,
            this._mnemonic,
            this._description,
            [
                { mnemonic: this._mnemonic + ':bipolar', description: this._description + ' bipolar' },
                { mnemonic: this._mnemonic + ':plus', description: this._description + ' plus' },
                { mnemonic: this._mnemonic + ':minus', description: this._description + ' minus' },
            ],
        );
    }

    /** Updates the indicator given the next high, low, and close values. */
    public update(high: number, low: number, close: number): [number, number, number] {
        const warmup = 41;
        const epsilon = 0.00001;
        const hundred = 100.0;

        this.bar++;

        let trueRange = 0;
        let upward = 0;
        let downward = 0;

        if (this.bar >= 2) {
            const v1 = hundred * (high - this.prevHigh);
            const v2 = hundred * (this.prevLow - low);

            if (v1 > v2 && v1 > 0) {
                upward = v1;
            }

            if (v2 > v1 && v2 > 0) {
                downward = v2;
            }
        }

        if (this.bar >= 3) {
            const m1 = Math.abs(high - low);
            const m2 = Math.abs(high - this.prevClose);
            const m3 = Math.abs(low - this.prevClose);
            trueRange = Math.max(Math.max(m1, m2), m3);
        }

        this.prevHigh = high;
        this.prevLow = low;
        this.prevClose = close;

        // Feed into JMA instances.
        const numerPlus = this.jmaPlus.update(upward);
        const numerMinus = this.jmaMinus.update(downward);
        const denom = this.jmaDenom.update(trueRange);

        if (this.bar <= warmup) {
            this.bipolarVal = NaN;
            this.plusVal = NaN;
            this.minusVal = NaN;

            return [NaN, NaN, NaN];
        }

        this._primed = true;

        // Compute Plus and Minus.
        if (denom > epsilon) {
            this.plusVal = numerPlus / denom;
        } else {
            this.plusVal = 0;
        }

        if (denom > epsilon) {
            this.minusVal = numerMinus / denom;
        } else {
            this.minusVal = 0;
        }

        // Compute Bipolar.
        const sum = this.plusVal + this.minusVal;
        if (sum > epsilon) {
            this.bipolarVal = hundred * (this.plusVal - this.minusVal) / sum;
        } else {
            this.bipolarVal = 0;
        }

        return [this.bipolarVal, this.plusVal, this.minusVal];
    }

    /** Updates the indicator given the next scalar sample. */
    public updateScalar(sample: Scalar): IndicatorOutput {
        const v = sample.value;
        const [bipolar, plus, minus] = this.update(v, v, v);

        const s1 = new Scalar();
        s1.time = sample.time;
        s1.value = bipolar;
        const s2 = new Scalar();
        s2.time = sample.time;
        s2.value = plus;
        const s3 = new Scalar();
        s3.time = sample.time;
        s3.value = minus;

        return [s1, s2, s3];
    }

    /** Updates the indicator given the next bar sample. */
    public updateBar(sample: Bar): IndicatorOutput {
        const [bipolar, plus, minus] = this.update(sample.high, sample.low, sample.close);

        const s1 = new Scalar();
        s1.time = sample.time;
        s1.value = bipolar;
        const s2 = new Scalar();
        s2.time = sample.time;
        s2.value = plus;
        const s3 = new Scalar();
        s3.time = sample.time;
        s3.value = minus;

        return [s1, s2, s3];
    }

    /** Updates the indicator given the next quote sample. */
    public updateQuote(sample: Quote): IndicatorOutput {
        const [bipolar, plus, minus] = this.update(sample.askPrice, sample.bidPrice, (sample.askPrice + sample.bidPrice) / 2);

        const s1 = new Scalar();
        s1.time = sample.time;
        s1.value = bipolar;
        const s2 = new Scalar();
        s2.time = sample.time;
        s2.value = plus;
        const s3 = new Scalar();
        s3.time = sample.time;
        s3.value = minus;

        return [s1, s2, s3];
    }

    /** Updates the indicator given the next trade sample. */
    public updateTrade(sample: Trade): IndicatorOutput {
        const v = sample.price;
        const [bipolar, plus, minus] = this.update(v, v, v);

        const s1 = new Scalar();
        s1.time = sample.time;
        s1.value = bipolar;
        const s2 = new Scalar();
        s2.time = sample.time;
        s2.value = plus;
        const s3 = new Scalar();
        s3.time = sample.time;
        s3.value = minus;

        return [s1, s2, s3];
    }
}
