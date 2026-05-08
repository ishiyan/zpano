import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikWaveletSamplerParams } from './params';

/** Function to calculate mnemonic of a JurikWaveletSampler indicator. */
export const jurikWaveletSamplerMnemonic = (params: JurikWaveletSamplerParams): string =>
    'jwav('.concat(
        params.index.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** (n, M) parameters for each wavelet column. */
const nmTable: Array<{ n: number; m: number }> = [
    { n: 1, m: 0 }, { n: 2, m: 0 }, { n: 3, m: 0 }, { n: 4, m: 0 }, { n: 5, m: 0 },
    { n: 7, m: 2 }, { n: 10, m: 2 }, { n: 14, m: 4 }, { n: 19, m: 4 }, { n: 26, m: 8 },
    { n: 35, m: 8 }, { n: 48, m: 16 }, { n: 65, m: 16 }, { n: 90, m: 32 }, { n: 123, m: 32 },
    { n: 172, m: 64 }, { n: 237, m: 64 }, { n: 334, m: 128 },
];

/** Jurik Wavelet Sampler (WAV) line indicator. */
export class JurikWaveletSampler extends LineIndicator {
    private readonly index: number;
    private readonly maxLookback: number;
    private prices: number[] = [];
    private barCount = 0;
    private readonly cols: number[];

    public constructor(params: JurikWaveletSamplerParams) {
        super();

        if (params.index < 1 || params.index > 18) {
            throw new Error('invalid jurik wavelet sampler parameters: index must be in range [1, 18]');
        }

        this.index = params.index;

        // Compute max lookback.
        let maxLookback = 0;
        for (let c = 0; c < params.index; c++) {
            const lb = nmTable[c].n + Math.trunc(nmTable[c].m / 2);
            if (lb > maxLookback) { maxLookback = lb; }
        }
        this.maxLookback = maxLookback;
        this.cols = new Array(params.index).fill(0);
        this.mnemonic = jurikWaveletSamplerMnemonic(params);
        this.description = 'Jurik wavelet sampler ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikWaveletSampler,
            this.mnemonic,
            this.description,
            [{ mnemonic: this.mnemonic, description: this.description }],
        );
    }

    /** Returns the current column values after the last update. */
    public columns(): number[] {
        return this.cols.slice();
    }

    /** Updates the value of the WAV indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        this.prices.push(sample);
        this.barCount++;

        let allValid = true;

        for (let c = 0; c < this.index; c++) {
            const n = nmTable[c].n;
            const m = nmTable[c].m;
            const deadZone = n + Math.trunc(m / 2);

            if (this.barCount <= deadZone) {
                this.cols[c] = NaN;
                allValid = false;
            } else {
                if (m === 0) {
                    // Simple lag.
                    this.cols[c] = this.prices[this.barCount - 1 - n];
                } else {
                    // Mean of (M+1) prices centered at lag n.
                    const half = Math.trunc(m / 2);
                    const centerIdx = this.barCount - 1 - n;
                    let total = 0.0;

                    for (let k = centerIdx - half; k <= centerIdx + half; k++) {
                        total += this.prices[k];
                    }

                    this.cols[c] = total / (m + 1);
                }
            }
        }

        if (allValid) {
            this.primed = true;
        }

        // Return first column as the framework output.
        return this.cols[0];
    }
}
