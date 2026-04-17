import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { CommodityChannelIndexOutput } from './commodity-channel-index-output';
import { CommodityChannelIndexParams, DefaultInverseScalingFactor } from './commodity-channel-index-params';

/** Function to calculate mnemonic of a __CommodityChannelIndex__ indicator. */
export const commodityChannelIndexMnemonic = (params: CommodityChannelIndexParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  return `cci(${params.length}${cm})`;
};

/**
 * CommodityChannelIndex is Donald Lambert's Commodity Channel Index (CCI).
 *
 * CCI measures the deviation of the price from its statistical mean. High values
 * indicate that prices are unusually high compared to average, and low values
 * indicate that prices are unusually low.
 *
 *   CCI = (typicalPrice - SMA) / (scalingFactor * meanDeviation)
 *
 * where scalingFactor defaults to 0.015 so that approximately 70-80% of CCI values
 * fall between -100 and +100.
 *
 * Reference:
 *
 * Lambert, Donald (1980). "Commodity Channel Index: Tools for Trading Cyclic Trends".
 * Commodities (now Futures) magazine.
 */
export class CommodityChannelIndex extends LineIndicator {
  private readonly length: number;
  private readonly scalingFactor: number;
  private readonly window: number[];
  private windowCount: number;
  private windowSum: number;
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params: CommodityChannelIndexParams) {
    super();

    const length = Math.floor(params.length);

    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    const inverseFactor = params.inverseScalingFactor ?? DefaultInverseScalingFactor;

    this.length = length;
    this.scalingFactor = length / inverseFactor;
    this.window = new Array<number>(length).fill(0);
    this.windowCount = 0;
    this.windowSum = 0;
    this.value = Number.NaN;
    this.primed = false;

    this.mnemonic = commodityChannelIndexMnemonic(params);
    this.description = 'Commodity Channel Index ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.CommodityChannelIndex,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: CommodityChannelIndexOutput.CommodityChannelIndexValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    const lastIndex = this.length - 1;

    if (this.primed) {
      this.windowSum += sample - this.window[0];

      for (let i = 0; i < lastIndex; i++) {
        this.window[i] = this.window[i + 1];
      }

      this.window[lastIndex] = sample;

      const average = this.windowSum / this.length;

      let temp = 0;
      for (let i = 0; i < this.length; i++) {
        temp += Math.abs(this.window[i] - average);
      }

      if (Math.abs(temp) < Number.EPSILON) {
        this.value = 0;
      } else {
        this.value = this.scalingFactor * (sample - average) / temp;
      }
    } else {
      this.windowSum += sample;
      this.window[this.windowCount] = sample;
      this.windowCount++;

      if (this.windowCount === this.length) {
        this.primed = true;

        const average = this.windowSum / this.length;

        let temp = 0;
        for (let i = 0; i < this.length; i++) {
          temp += Math.abs(this.window[i] - average);
        }

        if (Math.abs(temp) < Number.EPSILON) {
          this.value = 0;
        } else {
          this.value = this.scalingFactor * (sample - average) / temp;
        }
      }
    }

    return this.value;
  }
}
