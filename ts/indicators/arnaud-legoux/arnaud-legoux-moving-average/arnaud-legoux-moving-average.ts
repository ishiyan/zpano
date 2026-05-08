import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { ArnaudLegouxMovingAverageParams } from './params';

/** Arnaud Legoux Moving Average (ALMA) line indicator. */
export class ArnaudLegouxMovingAverage extends LineIndicator {
    private weights: number[];
    private windowLength: number;
    private buffer: number[];
    private bufferCount: number;
    private bufferIndex: number;

    /**
     * Constructs an instance given parameters.
     **/
    public constructor(params: ArnaudLegouxMovingAverageParams) {
        super();
        const window = Math.floor(params.window);
        if (window < 1) {
            throw new Error('window should be greater than 0');
        }

        const sigma = params.sigma;
        if (sigma <= 0) {
            throw new Error('sigma should be greater than 0');
        }

        const offset = params.offset;
        if (offset < 0 || offset > 1) {
            throw new Error('offset should be between 0 and 1');
        }

        this.mnemonic = 'alma(' + window + ', ' + sigma + ', ' + offset +
            componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent) + ')';
        this.description = 'Arnaud Legoux moving average ' + this.mnemonic;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;

        // Precompute Gaussian weights.
        const m = offset * (window - 1);
        const s = window / sigma;
        const weights: number[] = new Array(window);
        let norm = 0;

        for (let i = 0; i < window; i++) {
            const diff = i - m;
            const w = Math.exp(-(diff * diff) / (2.0 * s * s));
            weights[i] = w;
            norm += w;
        }

        for (let i = 0; i < window; i++) {
            weights[i] /= norm;
        }

        this.weights = weights;
        this.windowLength = window;
        this.buffer = new Array(window);
        this.bufferCount = 0;
        this.bufferIndex = 0;
        this.primed = false;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.ArnaudLegouxMovingAverage,
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

        const window = this.windowLength;

        if (window === 1) {
            this.primed = true;
            return sample;
        }

        // Fill the circular buffer.
        this.buffer[this.bufferIndex] = sample;
        this.bufferIndex = (this.bufferIndex + 1) % window;

        if (!this.primed) {
            this.bufferCount++;
            if (this.bufferCount < window) {
                return Number.NaN;
            }

            this.primed = true;
        }

        // Compute weighted sum.
        let result = 0;
        let index = this.bufferIndex;

        for (let i = 0; i < window; i++) {
            result += this.weights[i] * this.buffer[index];
            index = (index + 1) % window;
        }

        return result;
    }
}
