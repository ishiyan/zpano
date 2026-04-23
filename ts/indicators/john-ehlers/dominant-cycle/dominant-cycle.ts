import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent, DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent, DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { HilbertTransformerCycleEstimator } from '../hilbert-transformer/cycle-estimator';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { createEstimator, estimatorMoniker } from '../hilbert-transformer/common';
import { DominantCycleParams } from './params';

const DEFAULT_ALPHA_EMA_PERIOD_ADDITIONAL = 0.33;
const DEFAULT_SMOOTHING_LENGTH = 4;
const DEFAULT_ALPHA_EMA_QI = 0.2;
const DEFAULT_ALPHA_EMA_PERIOD = 0.2;
// MBST's DominantCyclePeriod default warm-up is MaxPeriod * 2 = 100. The HTCE's internal
// auto-default (smoothingLengthPlus3HtLength, = 25 for smoothingLength=4) is shorter and would
// prime earlier than the MBST reference. We pass it explicitly so `default()` matches MBST.
const DEFAULT_WARM_UP_PERIOD = 100;

/** __Dominant Cycle__ (Ehlers) computes the instantaneous cycle period and phase derived
 * from a Hilbert transformer cycle estimator.
 *
 * It exposes three outputs:
 *
 *	- RawPeriod: the raw instantaneous cycle period produced by the Hilbert transformer estimator.
 *	- Period: the dominant cycle period obtained by additional EMA smoothing of the raw period.
 *	  Periodᵢ = α·RawPeriodᵢ + (1 − α)·Periodᵢ₋₁, 0 < α ≤ 1.
 *	- Phase: the dominant cycle phase, in degrees.
 *
 * The smoothed data are multiplied by the real (cosine) component of the dominant cycle and
 * independently by the imaginary (sine) component of the dominant cycle. The products are
 * summed then over one full dominant cycle. The phase angle is computed as the arctangent of
 * the ratio of the real part to the imaginary part.
 *
 * Reference:
 *
 *	John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
 */
export class DominantCycle implements Indicator {
  private readonly htce: HilbertTransformerCycleEstimator;
  private readonly alphaEmaPeriodAdditional: number;
  private readonly oneMinAlphaEmaPeriodAdditional: number;
  private readonly smoothedInput: number[];
  private readonly smoothedInputLengthMin1: number;
  private smoothedPeriod = 0;
  private smoothedPhase = 0;
  private primed = false;

  private readonly mnemonicRawPeriod: string;
  private readonly descriptionRawPeriod: string;
  private readonly mnemonicPeriod: string;
  private readonly descriptionPeriod: string;
  private readonly mnemonicPhase: string;
  private readonly descriptionPhase: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Creates an instance using default parameters (α=0.33, HomodyneDiscriminator cycle
   * estimator with smoothingLength=4, αq=0.2, αp=0.2, warmUpPeriod=100). */
  public static default(): DominantCycle {
    return new DominantCycle({
      alphaEmaPeriodAdditional: DEFAULT_ALPHA_EMA_PERIOD_ADDITIONAL,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: {
        smoothingLength: DEFAULT_SMOOTHING_LENGTH,
        alphaEmaQuadratureInPhase: DEFAULT_ALPHA_EMA_QI,
        alphaEmaPeriod: DEFAULT_ALPHA_EMA_PERIOD,
        warmUpPeriod: DEFAULT_WARM_UP_PERIOD,
      },
    });
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: DominantCycleParams): DominantCycle {
    return new DominantCycle(params);
  }

  private constructor(params: DominantCycleParams) {
    const alpha = params.alphaEmaPeriodAdditional;
    if (alpha <= 0 || alpha > 1) {
      throw new Error('invalid dominant cycle parameters: α for additional smoothing should be in range (0, 1]');
    }

    this.alphaEmaPeriodAdditional = alpha;
    this.oneMinAlphaEmaPeriodAdditional = 1 - alpha;

    this.htce = createEstimator(params.estimatorType, params.estimatorParams);

    const effectiveType = params.estimatorType ?? HilbertTransformerCycleEstimatorType.HomodyneDiscriminator;
    let em = '';
    const isDefaultHd = effectiveType === HilbertTransformerCycleEstimatorType.HomodyneDiscriminator
      && this.htce.smoothingLength === DEFAULT_SMOOTHING_LENGTH
      && this.htce.alphaEmaQuadratureInPhase === DEFAULT_ALPHA_EMA_QI
      && this.htce.alphaEmaPeriod === DEFAULT_ALPHA_EMA_PERIOD;
    if (!isDefaultHd) {
      const moniker = estimatorMoniker(effectiveType, this.htce);
      if (moniker.length > 0) {
        em = ', ' + moniker;
      }
    }

    const bc = params.barComponent ?? DefaultBarComponent;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
    const a = alpha.toFixed(3);

    this.mnemonicRawPeriod = `dcp-raw(${a}${em}${cm})`;
    this.mnemonicPeriod = `dcp(${a}${em}${cm})`;
    this.mnemonicPhase = `dcph(${a}${em}${cm})`;

    this.descriptionRawPeriod = 'Dominant cycle raw period ' + this.mnemonicRawPeriod;
    this.descriptionPeriod = 'Dominant cycle period ' + this.mnemonicPeriod;
    this.descriptionPhase = 'Dominant cycle phase ' + this.mnemonicPhase;

    const maxPeriod = this.htce.maxPeriod;
    this.smoothedInput = new Array<number>(maxPeriod).fill(0);
    this.smoothedInputLengthMin1 = maxPeriod - 1;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** The current WMA-smoothed price value produced by the underlying Hilbert transformer
   * cycle estimator. Returns NaN if the indicator is not yet primed.
   *
   * This accessor is intended for composite indicators (e.g. TrendCycleMode) that wrap a
   * DominantCycle and need to consult the same smoothed input stream that drives the
   * dominant-cycle computation. */
  public get smoothedPrice(): number {
    return this.primed ? this.htce.smoothed : Number.NaN;
  }

  /** The maximum cycle period supported by the underlying Hilbert transformer cycle
   * estimator (also the size of the internal smoothed-input buffer). */
  public get maxPeriod(): number {
    return this.htce.maxPeriod;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.DominantCycle,
      this.mnemonicPeriod,
      this.descriptionPeriod,
      [
        { mnemonic: this.mnemonicRawPeriod, description: this.descriptionRawPeriod },
        { mnemonic: this.mnemonicPeriod, description: this.descriptionPeriod },
        { mnemonic: this.mnemonicPhase, description: this.descriptionPhase },
      ],
    );
  }

  /** Updates the indicator given the next sample value. Returns the triple
   * (rawPeriod, period, phase). Returns (NaN, NaN, NaN) if not yet primed. */
  public update(sample: number): [number, number, number] {
    if (Number.isNaN(sample)) {
      return [sample, sample, sample];
    }

    this.htce.update(sample);
    this.pushSmoothedInput(this.htce.smoothed);

    if (this.primed) {
      this.smoothedPeriod = this.alphaEmaPeriodAdditional * this.htce.period
        + this.oneMinAlphaEmaPeriodAdditional * this.smoothedPeriod;
      this.calculateSmoothedPhase();
      return [this.htce.period, this.smoothedPeriod, this.smoothedPhase];
    }

    if (this.htce.primed) {
      this.primed = true;
      this.smoothedPeriod = this.htce.period;
      this.calculateSmoothedPhase();
      return [this.htce.period, this.smoothedPeriod, this.smoothedPhase];
    }

    return [Number.NaN, Number.NaN, Number.NaN];
  }

  /** Updates an indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value);
  }

  /** Updates an indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, this.barComponentFunc(sample));
  }

  /** Updates an indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(sample.time, this.quoteComponentFunc(sample));
  }

  /** Updates an indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, this.tradeComponentFunc(sample));
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const [rawPeriod, period, phase] = this.update(sample);

    const s1 = new Scalar();
    s1.time = time;
    s1.value = rawPeriod;

    const s2 = new Scalar();
    s2.time = time;
    s2.value = period;

    const s3 = new Scalar();
    s3.time = time;
    s3.value = phase;

    return [s1, s2, s3];
  }

  private pushSmoothedInput(value: number): void {
    for (let i = this.smoothedInputLengthMin1; i > 0; i--) {
      this.smoothedInput[i] = this.smoothedInput[i - 1];
    }
    this.smoothedInput[0] = value;
  }

  private calculateSmoothedPhase(): void {
    const rad2deg = 180.0 / Math.PI;
    const twoPi = 2.0 * Math.PI;
    const epsilon = 0.01;

    // The smoothed data are multiplied by the real (cosine) component of the dominant cycle
    // and independently by the imaginary (sine) component of the dominant cycle. The products
    // are summed over one full dominant cycle.
    let length = Math.floor(this.smoothedPeriod + 0.5);
    if (length > this.smoothedInputLengthMin1) {
      length = this.smoothedInputLengthMin1;
    }

    let realPart = 0;
    let imagPart = 0;

    for (let i = 0; i < length; i++) {
      const temp = (twoPi * i) / length;
      const smoothed = this.smoothedInput[i];
      realPart += smoothed * Math.sin(temp);
      imagPart += smoothed * Math.cos(temp);
    }

    // We compute the phase angle as the arctangent of the ratio of the real part to the imaginary part.
    // The phase increases from left to right.
    const previous = this.smoothedPhase;
    let phase = Math.atan(realPart / imagPart) * rad2deg;
    if (Number.isNaN(phase) || !Number.isFinite(phase)) {
      phase = previous;
    }

    if (Math.abs(imagPart) <= epsilon) {
      if (realPart > 0) {
        phase += 90;
      } else if (realPart < 0) {
        phase -= 90;
      }
    }

    // 90 degree reference shift.
    phase += 90;
    // Compensate for one-bar lag of the WMA-smoothed input price by adding the phase
    // corresponding to a 1-bar lag of the smoothed dominant cycle period.
    phase += 360 / this.smoothedPeriod;
    // Resolve phase ambiguity when the imaginary part is negative.
    if (imagPart < 0) {
      phase += 180;
    }
    // Cycle wraparound.
    if (phase > 360) {
      phase -= 360;
    }

    this.smoothedPhase = phase;
  }
}
