import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikMovingAverage } from '../jurik-moving-average/jurik-moving-average';
import { JurikCommodityChannelIndexParams } from './params';

/** Function to calculate mnemonic of a JurikCommodityChannelIndex indicator. */
export const jurikCommodityChannelIndexMnemonic = (params: JurikCommodityChannelIndexParams): string =>
    'jccx('.concat(
        params.length.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Jurik Commodity Channel Index line indicator. */
export class JurikCommodityChannelIndex extends LineIndicator {
    private readonly fastJMA: JurikMovingAverage;
    private readonly slowJMA: JurikMovingAverage;
    private readonly diffBufSize: number;
    private diffBuffer: number[] = [];

    /**
     * Constructs an instance using given parameters.
     */
    public constructor(params: JurikCommodityChannelIndexParams) {
        super();

        const length = params.length;
        if (length < 2) {
            throw new Error('invalid jurik commodity channel index parameters: length must be >= 2');
        }

        this.mnemonic = jurikCommodityChannelIndexMnemonic(params);
        this.description = 'Jurik commodity channel index ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;

        this.fastJMA = new JurikMovingAverage({ length: 4, phase: 0 });
        this.slowJMA = new JurikMovingAverage({ length, phase: 0 });
        this.diffBufSize = 3 * length;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikCommodityChannelIndex,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
            ],
        );
    }

    /** Updates the value of the JCCX indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const fastVal = this.fastJMA.update(sample);
        const slowVal = this.slowJMA.update(sample);

        if (Number.isNaN(fastVal) || Number.isNaN(slowVal)) {
            return NaN;
        }

        const diff = fastVal - slowVal;

        this.diffBuffer.push(diff);
        if (this.diffBuffer.length > this.diffBufSize) {
            this.diffBuffer = this.diffBuffer.slice(1);
        }

        this.primed = true;

        // Compute MAD.
        const n = this.diffBuffer.length;
        let mad = 0;

        for (const d of this.diffBuffer) {
            mad += Math.abs(d);
        }

        mad /= n;

        if (mad < 0.00001) {
            return 0.0;
        }

        return diff / (1.5 * mad);
    }
}
