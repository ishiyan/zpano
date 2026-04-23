import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { KaufmanAdaptiveMovingAverageLengthParams } from './params';
import { KaufmanAdaptiveMovingAverageSmoothingFactorParams } from './params';

const guardLength = (object: any): object is KaufmanAdaptiveMovingAverageLengthParams => 'fastestLength' in object;

/** Function to calculate mnemonic of a __KaufmanAdaptiveMovingAverage__ indicator. */
export const kaufmanAdaptiveMovingAverageMnemonic =
  (params: KaufmanAdaptiveMovingAverageLengthParams | KaufmanAdaptiveMovingAverageSmoothingFactorParams): string => {
    if (guardLength(params)) {
      const p = params as KaufmanAdaptiveMovingAverageLengthParams;
      return 'kama('.concat(Math.floor(p.efficiencyRatioLength).toString(),
        ', ', Math.floor(p.fastestLength).toString(), ', ', Math.floor(p.slowestLength).toString(),
        componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent), ')');
    } else {
      const p = params as KaufmanAdaptiveMovingAverageSmoothingFactorParams;
      return 'kama('.concat(Math.floor(p.efficiencyRatioLength).toString(),
        ', ', p.fastestSmoothingFactor.toFixed(4), ', ', p.slowestSmoothingFactor.toFixed(4),
        componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent), ')');
    }
  };

/** __Kaufman Adaptive Moving Average__ (_KAMA_) is an EMA with the smoothing
 * factor, α, being changed with each new sample within the fastest and the slowest boundaries:
 *
 * KAMAᵢ = αPᵢ + (1 - α)*KAMAᵢ₋₁,  α = (αs + (αf - αs)ε)²
 *
 * where the αf is the α of the fastest (shortest, default 2 samples) period boundary,
 * the αs is the α of the slowest (longest, default 30 samples) period boundary,
 * and ε is the efficiency ratio:
 *
 * ε = |P - Pℓ| / ∑|Pᵢ - Pᵢ₊₁|,  i ≤ ℓ-1
 *
 * where ℓ is a number of samples used to calculate the ε.
 * The recommended values of ℓ are in the range of 8 to 10.
 *
 * The efficiency ratio has the value of 1 when samples move in the same direction for
 * the full ℓ periods, and a value of 0 when samples are unchanged over the ℓ periods.
 * When samples move in wide swings within the interval, the sum of the denominator
 * becomes very large compared with the numerator and the ε approaches 0.
 * Smaller values of ε result in a smaller smoothing constant and a slower trend.
 *
 * The indicator is not primed during the first ℓ updates.
 *
 * Reference:
 * Perry J. Kaufman, Smarter Trading, McGraw-Hill, Ney York, 1995, pp. 129-153.
 */
export class KaufmanAdaptiveMovingAverage extends LineIndicator {
  private readonly efficiencyRatioLength: number;
  private readonly alphaFastest: number;
  private readonly alphaSlowest: number;
  private readonly alphaDiff: number;
  private readonly window: Array<number>;
  private readonly absoluteDelta: Array<number>;
  private absoluteDeltaSum = 0;
  private value = 0;
  private efficiencyRatioValue = 0;
  private windowCount = 0;

  /**
   * Constructs an instance given a length in samples or a smoothing factor in (0, 1).
   **/
  public constructor(params: KaufmanAdaptiveMovingAverageLengthParams | KaufmanAdaptiveMovingAverageSmoothingFactorParams) {
    super();

    if (guardLength(params)) {
      const p = params as KaufmanAdaptiveMovingAverageLengthParams;

      this.efficiencyRatioLength = Math.floor(p.efficiencyRatioLength);
      if (this.efficiencyRatioLength < 2) {
        throw new Error('efficiency ratio length should be greater than 1');
      }

      const fastestLen = Math.floor(p.fastestLength);
      if (fastestLen < 2) {
        throw new Error('fastest length should be greater than 1');
      }

      const slowestLen = Math.floor(p.slowestLength);
      if (slowestLen < 2) {
        throw new Error('slowest length should be greater than 1');
      }

      this.alphaFastest = 2 / (fastestLen + 1);
      this.alphaSlowest = 2 / (slowestLen + 1);
    } else {
      const p = params as KaufmanAdaptiveMovingAverageSmoothingFactorParams;

      this.efficiencyRatioLength = Math.floor(p.efficiencyRatioLength);
      if (this.efficiencyRatioLength < 2) {
        throw new Error('efficiency ratio length should be greater than 1');
      }

      if (p.fastestSmoothingFactor <= 0 || p.fastestSmoothingFactor >= 1) {
        throw new Error('fastest smoothing factor should be in range (0, 1)');
      }

      if (p.slowestSmoothingFactor <= 0 || p.slowestSmoothingFactor >= 1) {
        throw new Error('slowest smoothing factor should be in range (0, 1)');
      }

      this.alphaFastest = p.fastestSmoothingFactor;
      this.alphaSlowest = p.slowestSmoothingFactor;
    }

    this.alphaDiff = this.alphaFastest - this.alphaSlowest;
    this.window = new Array<number>(this.efficiencyRatioLength + 1);
    this.absoluteDelta = new Array<number>(this.efficiencyRatioLength + 1);

    this.mnemonic = kaufmanAdaptiveMovingAverageMnemonic(params);
    this.description = 'Kaufman adaptive moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** The value of the efficiency ratio component of the indicator. */
  public getEfficiencyRatio(): number { return this.primed ? this.efficiencyRatioValue : Number.NaN; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.KaufmanAdaptiveMovingAverage,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    const epsilon = 0.00000001;

    let temp;

    if (this.primed) {
      temp = Math.abs(sample - this.window[this.efficiencyRatioLength]);
      this.absoluteDeltaSum += temp - this.absoluteDelta[1];

      for (let i = 0; i < this.efficiencyRatioLength; i++) {
        const j = i + 1;
        this.window[i] = this.window[j];
        this.absoluteDelta[i] = this.absoluteDelta[j];
      }

      this.window[this.efficiencyRatioLength] = sample;
      this.absoluteDelta[this.efficiencyRatioLength] = temp;
      const delta = Math.abs(sample - this.window[0]);

      if (this.absoluteDeltaSum <= delta || this.absoluteDeltaSum < epsilon) {
        temp = 1;
      } else {
        temp = delta / this.absoluteDeltaSum;
      }

      this.efficiencyRatioValue = temp;
      temp = this.alphaSlowest + temp * this.alphaDiff;
      this.value += (sample - this.value) * temp * temp;

      return this.value;
    } else { // Not primed.
      this.window[this.windowCount] = sample;

      if (0 < this.windowCount) {
        temp = Math.abs(sample - this.window[this.windowCount - 1]);
        this.absoluteDelta[this.windowCount] = temp;
        this.absoluteDeltaSum += temp;
      }

      if (this.efficiencyRatioLength === this.windowCount) {
        this.primed = true;
        const delta = Math.abs(sample - this.window[0]);

        if (this.absoluteDeltaSum <= delta || this.absoluteDeltaSum < epsilon) {
          temp = 1;
        } else {
          temp = delta / this.absoluteDeltaSum;
        }

        this.efficiencyRatioValue = temp;
        temp = this.alphaSlowest + temp * this.alphaDiff;
        this.value = this.window[this.efficiencyRatioLength - 1];
        this.value += (sample - this.value) * temp * temp;

        return this.value;
      } else {
        this.windowCount++;
      }
    }

    return Number.NaN;
  }
}
