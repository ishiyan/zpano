import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Scalar } from '../../../entities/scalar';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';

const epsilon = 1e-8;

/**
 * BalanceOfPower is Igor Livshin's Balance of Power (BOP).
 *
 * The Balance of Market Power captures the struggles of bulls vs. bears
 * throughout the trading day. It assigns scores to both bulls and bears
 * based on how much they were able to move prices throughout the trading day.
 *
 * The value is calculated as:
 *
 *   BOP = (Close - Open) / (High - Low)
 *
 * When the range (High - Low) is less than epsilon, the value is 0.
 *
 * Reference:
 *
 * Livshin, Igor. "Balance of Market Power".
 */
export class BalanceOfPower extends LineIndicator {
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor() {
    super();

    this.value = Number.NaN;
    this.primed = true;

    this.mnemonic = 'bop';
    this.description = 'Balance of Power';
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.BalanceOfPower,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /** Updates the indicator with the given sample (O=H=L=C, so result is always 0). */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }

    return this.updateOHLC(sample, sample, sample, sample);
  }

  /** Updates the indicator with the given OHLC values. */
  public updateOHLC(open: number, high: number, low: number, close: number): number {
    if (Number.isNaN(open) || Number.isNaN(high) || Number.isNaN(low) || Number.isNaN(close)) {
      return Number.NaN;
    }

    const range = high - low;
    if (range < epsilon) {
      this.value = 0;
    } else {
      this.value = (close - open) / range;
    }

    return this.value;
  }

  /** Updates the indicator given the next bar sample, extracting OHLC. */
  public override updateBar(sample: Bar): IndicatorOutput {
    const v = this.updateOHLC(sample.open, sample.high, sample.low, sample.close);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return [scalar];
  }
}
