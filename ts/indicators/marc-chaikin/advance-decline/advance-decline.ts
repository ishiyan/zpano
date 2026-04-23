import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Scalar } from '../../../entities/scalar';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';

/**
 * AdvanceDecline is Marc Chaikin's Advance-Decline (A/D) Line.
 *
 * The Accumulation/Distribution Line is a cumulative indicator that uses volume
 * and price to assess whether a stock is being accumulated or distributed.
 * The A/D line seeks to identify divergences between the stock price and volume flow.
 *
 * The value is calculated as:
 *
 *   CLV = ((Close - Low) - (High - Close)) / (High - Low)
 *   AD  = AD_previous + CLV x Volume
 *
 * When High equals Low, the A/D value is unchanged (no division by zero).
 *
 * Reference:
 *
 * Chaikin, Marc. "Chaikin Accumulation/Distribution Line".
 */
export class AdvanceDecline extends LineIndicator {
  private ad: number;
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor() {
    super();

    this.ad = 0;
    this.value = Number.NaN;
    this.primed = false;

    this.mnemonic = 'ad';
    this.description = 'Advance-Decline';
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.AdvanceDecline,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /** Updates the indicator with the given sample (H=L=C, volume=1, so AD is unchanged). */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }

    return this.updateHLCV(sample, sample, sample, 1);
  }

  /** Updates the indicator with the given high, low, close, and volume values. */
  public updateHLCV(high: number, low: number, close: number, volume: number): number {
    if (Number.isNaN(high) || Number.isNaN(low) || Number.isNaN(close) || Number.isNaN(volume)) {
      return Number.NaN;
    }

    const temp = high - low;
    if (temp > 0) {
      this.ad += ((close - low) - (high - close)) / temp * volume;
    }

    this.value = this.ad;
    this.primed = true;

    return this.value;
  }

  /** Updates the indicator given the next bar sample, extracting HLCV. */
  public override updateBar(sample: Bar): IndicatorOutput {
    const v = this.updateHLCV(sample.high, sample.low, sample.close, sample.volume);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return [scalar];
  }
}
