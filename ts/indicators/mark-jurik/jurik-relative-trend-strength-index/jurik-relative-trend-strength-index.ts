import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikRelativeTrendStrengthIndexParams } from './params';

/** Function to calculate mnemonic of a JurikRelativeTrendStrengthIndex indicator. */
export const jurikRelativeTrendStrengthIndexMnemonic = (params: JurikRelativeTrendStrengthIndexParams): string =>
    'jrsx('.concat(
        params.length.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/**
 * JurikRelativeTrendStrengthIndex computes the Jurik RSX indicator.
 * RSX is a noise-free version of RSI based on triple-smoothed EMA of momentum and absolute momentum.
 */
export class JurikRelativeTrendStrengthIndex extends LineIndicator {
    private readonly paramLen: number;

    // State variables.
    private f0 = 0;
    private f88 = 0;
    private f90 = 0;

    private f8 = 0;
    private f10 = 0;
    private f18 = 0;
    private f20 = 0;
    private f28 = 0;
    private f30 = 0;
    private f38 = 0;
    private f40 = 0;
    private f48 = 0;
    private f50 = 0;
    private f58 = 0;
    private f60 = 0;
    private f68 = 0;
    private f70 = 0;
    private f78 = 0;
    private f80 = 0;

    constructor(params: JurikRelativeTrendStrengthIndexParams) {
        super();

        const length = params.length;

        if (length < 2) {
            throw new Error('invalid jurik relative trend strength index parameters: length should be at least 2');
        }

        this.mnemonic = jurikRelativeTrendStrengthIndexMnemonic(params);
        this.description = 'Jurik relative trend strength index ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;

        this.paramLen = length;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikRelativeTrendStrengthIndex,
            this.mnemonic,
            this.description,
            [{ mnemonic: this.mnemonic, description: this.description }],
        );
    }

    /** Updates the value of the RSX indicator given the next sample. */
    public update(sample: number): number {
        if (isNaN(sample)) {
            return sample;
        }

        const hundred = 100.0;
        const fifty = 50.0;
        const oneFive = 1.5;
        const half = 0.5;
        const minLen = 5;
        const eps = 1e-10;

        const length = this.paramLen;

        if (this.f90 === 0) {
            // First call: initialize.
            this.f90 = 1;
            this.f0 = 0;

            if (length - 1 >= minLen) {
                this.f88 = length - 1;
            } else {
                this.f88 = minLen;
            }

            this.f8 = hundred * sample;
            this.f18 = 3.0 / (length + 2);
            this.f20 = 1 - this.f18;
        } else {
            if (this.f88 <= this.f90) {
                this.f90 = this.f88 + 1;
            } else {
                this.f90++;
            }

            this.f10 = this.f8;
            this.f8 = hundred * sample;
            const v8 = this.f8 - this.f10;

            this.f28 = this.f20 * this.f28 + this.f18 * v8;
            this.f30 = this.f18 * this.f28 + this.f20 * this.f30;
            const vC = this.f28 * oneFive - this.f30 * half;

            this.f38 = this.f20 * this.f38 + this.f18 * vC;
            this.f40 = this.f18 * this.f38 + this.f20 * this.f40;
            const v10 = this.f38 * oneFive - this.f40 * half;

            this.f48 = this.f20 * this.f48 + this.f18 * v10;
            this.f50 = this.f18 * this.f48 + this.f20 * this.f50;
            const v14 = this.f48 * oneFive - this.f50 * half;

            this.f58 = this.f20 * this.f58 + this.f18 * Math.abs(v8);
            this.f60 = this.f18 * this.f58 + this.f20 * this.f60;
            const v18 = this.f58 * oneFive - this.f60 * half;

            this.f68 = this.f20 * this.f68 + this.f18 * v18;
            this.f70 = this.f18 * this.f68 + this.f20 * this.f70;
            const v1C = this.f68 * oneFive - this.f70 * half;

            this.f78 = this.f20 * this.f78 + this.f18 * v1C;
            this.f80 = this.f18 * this.f78 + this.f20 * this.f80;
            const v20 = this.f78 * oneFive - this.f80 * half;

            if (this.f88 >= this.f90 && this.f8 !== this.f10) {
                this.f0 = 1;
            }

            if (this.f88 === this.f90 && this.f0 === 0) {
                this.f90 = 0;
            }

            if (this.f88 < this.f90 && v20 > eps) {
                let v4 = (v14 / v20 + 1) * fifty;
                if (v4 > hundred) {
                    v4 = hundred;
                }

                if (v4 < 0) {
                    v4 = 0;
                }

                this.primed = true;

                return v4;
            }
        }

        // During warmup or when denominator is too small.
        if (this.f88 < this.f90) {
            this.primed = true;
        }

        if (!this.primed) {
            return NaN;
        }

        return fifty;
    }
}
