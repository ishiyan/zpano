import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikTurningPointOscillatorParams } from './params';

/** Function to calculate mnemonic of a JurikTurningPointOscillator indicator. */
export const jurikTurningPointOscillatorMnemonic = (params: JurikTurningPointOscillatorParams): string =>
    'jtpo('.concat(
        params.length.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Jurik Turning Point Oscillator computes Spearman rank correlation between price ranks and time positions. Output is in [-1, +1]. */
export class JurikTurningPointOscillator extends LineIndicator {
    private readonly length: number;
    private readonly buffer: number[];
    private bufIdx = 0;
    private count = 0;
    private readonly f18: number;
    private readonly mid: number;

    /**
     * Constructs an instance using given parameters.
     */
    public constructor(params: JurikTurningPointOscillatorParams) {
        super();

        const length = params.length;
        if (length < 2) {
            throw new Error('invalid jurik turning point oscillator parameters: length should be at least 2');
        }

        this.length = length;
        this.buffer = new Array(length).fill(0);
        const n = length;
        this.f18 = 12.0 / (n * (n - 1) * (n + 1));
        this.mid = (n + 1) / 2.0;

        this.mnemonic = jurikTurningPointOscillatorMnemonic(params);
        this.description = 'Jurik turning point oscillator ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikTurningPointOscillator,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
            ],
        );
    }

    /** Updates the indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const length = this.length;

        this.buffer[this.bufIdx] = sample;
        this.bufIdx = (this.bufIdx + 1) % length;
        this.count++;

        if (this.count < length) {
            return NaN;
        }

        // Extract window in chronological order.
        const window: number[] = new Array(length);
        for (let i = 0; i < length; i++) {
            window[i] = this.buffer[(this.bufIdx + i) % length];
        }

        // Check if all values are identical.
        let allSame = true;
        for (let i = 1; i < length; i++) {
            if (window[i] !== window[0]) {
                allSame = false;
                break;
            }
        }

        if (allSame) {
            if (!this.primed) {
                this.primed = true;
            }
            return NaN;
        }

        // Build indices sorted by price (stable sort).
        const items: { idx: number; price: number }[] = new Array(length);
        for (let i = 0; i < length; i++) {
            items[i] = { idx: i, price: window[i] };
        }

        items.sort((a, b) => {
            if (a.price < b.price) { return -1; }
            if (a.price > b.price) { return 1; }
            return 0;
        });

        // arr2[i] = original time position (1-based) of the i-th sorted element.
        const arr2: number[] = new Array(length);
        for (let i = 0; i < length; i++) {
            arr2[i] = items[i].idx + 1;
        }

        // Assign fractional ranks for ties.
        const arr3: number[] = new Array(length);
        let i = 0;
        while (i < length) {
            let j = i;
            while (j < length - 1 && items[j + 1].price === items[j].price) {
                j++;
            }
            const avgRank = (i + 1 + j + 1) / 2.0;
            for (let k = i; k <= j; k++) {
                arr3[k] = avgRank;
            }
            i = j + 1;
        }

        // Compute correlation sum.
        let corrSum = 0;
        for (let i = 0; i < length; i++) {
            corrSum += (arr3[i] - this.mid) * (arr2[i] - this.mid);
        }

        if (!this.primed) {
            this.primed = true;
        }

        return this.f18 * corrSum;
    }
}
