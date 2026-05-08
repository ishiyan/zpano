import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikAdaptiveRelativeTrendStrengthIndexParams } from './params';

/** Function to calculate mnemonic of a JurikAdaptiveRelativeTrendStrengthIndex indicator. */
export const jurikAdaptiveRelativeTrendStrengthIndexMnemonic = (params: JurikAdaptiveRelativeTrendStrengthIndexParams): string =>
    'jarsx('.concat(
        params.loLength.toString(),
        ', ',
        params.hiLength.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Jurik Adaptive Relative Trend Strength Index line indicator. */
export class JurikAdaptiveRelativeTrendStrengthIndex extends LineIndicator {
    private readonly loLength: number;
    private readonly hiLength: number;
    private readonly eps = 0.001;

    private barCount = 0;
    private previousPrice = 0;

    // Rolling buffers for adaptive length.
    private readonly longBuffer = new Array<number>(100).fill(0);
    private longIndex = 0;
    private longSum = 0;
    private longCount = 0;
    private readonly shortBuffer = new Array<number>(10).fill(0);
    private shortIndex = 0;
    private shortSum = 0;
    private shortCount = 0;

    // RSX core state.
    private kg = 0;
    private c = 0;
    private warmup = 0;
    // Signal path (3 cascaded stages).
    private sig1A = 0;
    private sig1B = 0;
    private sig2A = 0;
    private sig2B = 0;
    private sig3A = 0;
    private sig3B = 0;
    // Denominator path (3 cascaded stages).
    private den1A = 0;
    private den1B = 0;
    private den2A = 0;
    private den2B = 0;
    private den3A = 0;
    private den3B = 0;

    /**
     * Constructs an instance using given parameters.
     */
    public constructor(params: JurikAdaptiveRelativeTrendStrengthIndexParams) {
        super();

        const loLength = params.loLength;
        const hiLength = params.hiLength;

        if (loLength < 2) {
            throw new Error('invalid jurik adaptive relative trend strength index parameters: lo_length should be at least 2');
        }

        if (hiLength < loLength) {
            throw new Error('invalid jurik adaptive relative trend strength index parameters: hi_length should be at least lo_length');
        }

        this.loLength = loLength;
        this.hiLength = hiLength;
        this.mnemonic = jurikAdaptiveRelativeTrendStrengthIndexMnemonic(params);
        this.description = 'Jurik adaptive relative trend strength index ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikAdaptiveRelativeTrendStrengthIndex,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
            ],
        );
    }

    /** Updates the value of the JARSX indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const bar = this.barCount;
        this.barCount++;

        if (bar === 0) {
            this.previousPrice = sample;

            // First bar: add 0 to both buffers.
            this.longBuffer[0] = 0.0;
            this.longSum = 0.0;
            this.longCount = 1;
            this.shortBuffer[0] = 0.0;
            this.shortSum = 0.0;
            this.shortCount = 1;

            // Compute adaptive length from bar 0.
            const avg1 = 0.0;
            const avg2 = 0.0;
            const value2 = Math.log((this.eps + avg1) / (this.eps + avg2));
            const value3 = value2 / (1.0 + Math.abs(value2));
            const adaptiveLength = this.loLength +
                (this.hiLength - this.loLength) * (1.0 + value3) / 2.0;
            let length = Math.trunc(adaptiveLength);
            if (length < 2) {
                length = 2;
            }

            this.kg = 3.0 / (length + 2);
            this.c = 1.0 - this.kg;
            this.warmup = length - 1;
            if (this.warmup < 5) {
                this.warmup = 5;
            }

            return NaN;
        }

        // Bars 1+
        const oldPrice = this.previousPrice;
        this.previousPrice = sample;
        const value1 = Math.abs(sample - oldPrice);

        // Update long rolling buffer.
        if (this.longCount < 100) {
            this.longBuffer[this.longCount] = value1;
            this.longSum += value1;
            this.longCount++;
        } else {
            this.longSum -= this.longBuffer[this.longIndex];
            this.longBuffer[this.longIndex] = value1;
            this.longSum += value1;
            this.longIndex = (this.longIndex + 1) % 100;
        }

        // Update short rolling buffer.
        if (this.shortCount < 10) {
            this.shortBuffer[this.shortCount] = value1;
            this.shortSum += value1;
            this.shortCount++;
        } else {
            this.shortSum -= this.shortBuffer[this.shortIndex];
            this.shortBuffer[this.shortIndex] = value1;
            this.shortSum += value1;
            this.shortIndex = (this.shortIndex + 1) % 10;
        }

        // RSX core computation.
        const mom = 100.0 * (sample - oldPrice);
        const absMom = Math.abs(mom);

        const kg = this.kg;
        const c = this.c;

        // Signal path — Stage 1.
        this.sig1A = c * this.sig1A + kg * mom;
        this.sig1B = kg * this.sig1A + c * this.sig1B;
        const s1 = 1.5 * this.sig1A - 0.5 * this.sig1B;

        // Signal path — Stage 2.
        this.sig2A = c * this.sig2A + kg * s1;
        this.sig2B = kg * this.sig2A + c * this.sig2B;
        const s2 = 1.5 * this.sig2A - 0.5 * this.sig2B;

        // Signal path — Stage 3.
        this.sig3A = c * this.sig3A + kg * s2;
        this.sig3B = kg * this.sig3A + c * this.sig3B;
        const numerator = 1.5 * this.sig3A - 0.5 * this.sig3B;

        // Denominator path — Stage 1.
        this.den1A = c * this.den1A + kg * absMom;
        this.den1B = kg * this.den1A + c * this.den1B;
        const d1 = 1.5 * this.den1A - 0.5 * this.den1B;

        // Denominator path — Stage 2.
        this.den2A = c * this.den2A + kg * d1;
        this.den2B = kg * this.den2A + c * this.den2B;
        const d2 = 1.5 * this.den2A - 0.5 * this.den2B;

        // Denominator path — Stage 3.
        this.den3A = c * this.den3A + kg * d2;
        this.den3B = kg * this.den3A + c * this.den3B;
        const denominator = 1.5 * this.den3A - 0.5 * this.den3B;

        // Output after warmup.
        if (bar >= this.warmup) {
            this.primed = true;

            let value: number;
            if (denominator !== 0.0) {
                value = (numerator / denominator + 1.0) * 50.0;
            } else {
                value = 50.0;
            }

            return Math.max(0.0, Math.min(100.0, value));
        }

        return NaN;
    }
}
