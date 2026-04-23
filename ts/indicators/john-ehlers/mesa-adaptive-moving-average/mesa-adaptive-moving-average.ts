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
import { Band } from '../../core/outputs/band';
import { HilbertTransformerCycleEstimator } from '../hilbert-transformer/cycle-estimator';
import { HilbertTransformerCycleEstimatorParams } from '../hilbert-transformer/cycle-estimator-params';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { createEstimator, estimatorMoniker } from '../hilbert-transformer/common';
import {
  MesaAdaptiveMovingAverageLengthParams,
  MesaAdaptiveMovingAverageSmoothingFactorParams,
} from './params';

const DEFAULT_FAST_LIMIT_LENGTH = 3;
const DEFAULT_SLOW_LIMIT_LENGTH = 39;
const DEFAULT_SMOOTHING_LENGTH = 4;
const DEFAULT_ALPHA_EMA_QI = 0.2;
const DEFAULT_ALPHA_EMA_PERIOD = 0.2;
const EPSILON = 1e-8;

/** __Mesa Adaptive Moving Average__ (_MAMA_, or Ehler's Mother of All Moving Averages)
 * is an EMA with the smoothing factor, α, being changed with each new sample within the fast and the slow
 * limit boundaries which are the constant parameters of MAMA:
 *
 *	MAMAᵢ = αᵢPᵢ + (1 - αᵢ)*MAMAᵢ₋₁,  αs ≤ αᵢ ≤ αf
 *
 * The αf is the α of the fast (shortest, default suggested value 0.5 or 3 samples) limit boundary.
 *
 * The αs is the α of the slow (longest, default suggested value 0.05 or 39 samples) limit boundary.
 *
 * The concept of MAMA is to relate the phase rate of change, as measured by a Hilbert Transformer
 * estimator, to the EMA smoothing factor α, thus making the EMA adaptive.
 *
 * The Following Adaptive Moving Average (FAMA) is produced by applying the MAMA to the first
 * MAMA indicator.
 *
 * Reference:
 *
 *	John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 177-184.
 */
export class MesaAdaptiveMovingAverage implements Indicator {
  private readonly htce: HilbertTransformerCycleEstimator;
  private readonly alphaFastLimit: number;
  private readonly alphaSlowLimit: number;
  private previousPhase = 0;
  private mama = Number.NaN;
  private fama = Number.NaN;
  private isPhaseCached = false;
  private primed = false;

  private readonly mnemonic: string;
  private readonly description: string;
  private readonly mnemonicFama: string;
  private readonly descriptionFama: string;
  private readonly mnemonicBand: string;
  private readonly descriptionBand: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Creates an instance using default parameters (fastLimitLength=3, slowLimitLength=39,
   * HomodyneDiscriminator cycle estimator with smoothingLength=4, αq=0.2, αp=0.2). */
  public static default(): MesaAdaptiveMovingAverage {
    return MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: DEFAULT_FAST_LIMIT_LENGTH,
      slowLimitLength: DEFAULT_SLOW_LIMIT_LENGTH,
    });
  }

  /** Creates an instance based on length parameters. */
  public static fromLength(params: MesaAdaptiveMovingAverageLengthParams): MesaAdaptiveMovingAverage {
    const fastLen = Math.floor(params.fastLimitLength);
    if (fastLen < 2) {
      throw new Error('invalid mesa adaptive moving average parameters: fast limit length should be larger than 1');
    }

    const slowLen = Math.floor(params.slowLimitLength);
    if (slowLen < 2) {
      throw new Error('invalid mesa adaptive moving average parameters: slow limit length should be larger than 1');
    }

    const alphaFast = 2 / (fastLen + 1);
    const alphaSlow = 2 / (slowLen + 1);

    return new MesaAdaptiveMovingAverage(
      alphaFast, alphaSlow,
      fastLen, slowLen, true,
      params.estimatorType, params.estimatorParams,
      params.barComponent, params.quoteComponent, params.tradeComponent,
    );
  }

  /** Creates an instance based on smoothing factor parameters. */
  public static fromSmoothingFactor(
    params: MesaAdaptiveMovingAverageSmoothingFactorParams,
  ): MesaAdaptiveMovingAverage {
    let alphaFast = params.fastLimitSmoothingFactor;
    if (alphaFast < 0 || alphaFast > 1) {
      throw new Error('invalid mesa adaptive moving average parameters: fast limit smoothing factor should be in range [0, 1]');
    }

    let alphaSlow = params.slowLimitSmoothingFactor;
    if (alphaSlow < 0 || alphaSlow > 1) {
      throw new Error('invalid mesa adaptive moving average parameters: slow limit smoothing factor should be in range [0, 1]');
    }

    if (alphaFast < EPSILON) {
      alphaFast = EPSILON;
    }
    if (alphaSlow < EPSILON) {
      alphaSlow = EPSILON;
    }

    return new MesaAdaptiveMovingAverage(
      alphaFast, alphaSlow,
      0, 0, false,
      params.estimatorType, params.estimatorParams,
      params.barComponent, params.quoteComponent, params.tradeComponent,
    );
  }

  private constructor(
    alphaFastLimit: number, alphaSlowLimit: number,
    fastLimitLength: number, slowLimitLength: number, usesLength: boolean,
    estimatorType: HilbertTransformerCycleEstimatorType | undefined,
    estimatorParams: HilbertTransformerCycleEstimatorParams | undefined,
    barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
  ) {
    this.alphaFastLimit = alphaFastLimit;
    this.alphaSlowLimit = alphaSlowLimit;

    this.htce = createEstimator(estimatorType, estimatorParams);

    const effectiveType = estimatorType ?? HilbertTransformerCycleEstimatorType.HomodyneDiscriminator;
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

    const bc = barComponent ?? DefaultBarComponent;
    const qc = quoteComponent ?? DefaultQuoteComponent;
    const tc = tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

    if (usesLength) {
      this.mnemonic = `mama(${fastLimitLength}, ${slowLimitLength}${em}${cm})`;
      this.mnemonicFama = `fama(${fastLimitLength}, ${slowLimitLength}${em}${cm})`;
      this.mnemonicBand = `mama-fama(${fastLimitLength}, ${slowLimitLength}${em}${cm})`;
    } else {
      const f = alphaFastLimit.toFixed(3);
      const s = alphaSlowLimit.toFixed(3);
      this.mnemonic = `mama(${f}, ${s}${em}${cm})`;
      this.mnemonicFama = `fama(${f}, ${s}${em}${cm})`;
      this.mnemonicBand = `mama-fama(${f}, ${s}${em}${cm})`;
    }

    const descr = 'Mesa adaptive moving average ';
    this.description = descr + this.mnemonic;
    this.descriptionFama = descr + this.mnemonicFama;
    this.descriptionBand = descr + this.mnemonicBand;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.MesaAdaptiveMovingAverage,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
        { mnemonic: this.mnemonicFama, description: this.descriptionFama },
        { mnemonic: this.mnemonicBand, description: this.descriptionBand },
      ],
    );
  }

  /** Updates the indicator given the next sample value. Returns the MAMA value. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    this.htce.update(sample);

    if (this.primed) {
      return this.calculate(sample);
    }

    if (this.htce.primed) {
      if (this.isPhaseCached) {
        this.primed = true;
        return this.calculate(sample);
      }

      this.isPhaseCached = true;
      this.previousPhase = this.calculatePhase();
      this.mama = sample;
      this.fama = sample;
    }

    return Number.NaN;
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
    const mama = this.update(sample);
    const fama = Number.isNaN(mama) ? Number.NaN : this.fama;

    const scalarMama = new Scalar();
    scalarMama.time = time;
    scalarMama.value = mama;

    const scalarFama = new Scalar();
    scalarFama.time = time;
    scalarFama.value = fama;

    const band = new Band();
    band.time = time;
    band.upper = mama;
    band.lower = fama;

    return [scalarMama, scalarFama, band];
  }

  private calculatePhase(): number {
    if (this.htce.inPhase === 0) {
      return this.previousPhase;
    }

    const rad2deg = 180 / Math.PI;

    // The cycle phase is computed from the arctangent of the ratio
    // of the Quadrature component to the InPhase component.
    // const phase = Math.atan2(this.htce.inPhase, this.htce.quadrature) * rad2deg;
    const phase = Math.atan(this.htce.quadrature / this.htce.inPhase) * rad2deg;
    if (!Number.isNaN(phase) && Number.isFinite(phase)) {
      return phase;
    }

    return this.previousPhase;
  }

  private calculateMama(sample: number): number {
    const phase = this.calculatePhase();

    // The phase rate of change is obtained by taking the
    // difference of successive previousPhase measurements.
    let phaseRateOfChange = this.previousPhase - phase;
    this.previousPhase = phase;

    // Any negative rate change is theoretically impossible
    // because phase must advance as the time increases.
    if (phaseRateOfChange < 1) {
      phaseRateOfChange = 1;
    }

    // The α is computed as the fast limit divided
    // by the phase rate of change.
    let alpha = this.alphaFastLimit / phaseRateOfChange;
    if (alpha < this.alphaSlowLimit) {
      alpha = this.alphaSlowLimit;
    }
    if (alpha > this.alphaFastLimit) {
      alpha = this.alphaFastLimit;
    }

    this.mama = alpha * sample + (1 - alpha) * this.mama;
    return alpha;
  }

  private calculate(sample: number): number {
    const alpha = this.calculateMama(sample) / 2;
    this.fama = alpha * this.mama + (1 - alpha) * this.fama;
    return this.mama;
  }
}
