import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Band } from '../../core/outputs/band';
import { DominantCycle } from '../dominant-cycle/dominant-cycle';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { createEstimator, estimatorMoniker } from '../hilbert-transformer/common';
import { SineWaveParams } from './params';

const DEFAULT_ALPHA_EMA_PERIOD_ADDITIONAL = 0.33;
const DEFAULT_SMOOTHING_LENGTH = 4;
const DEFAULT_ALPHA_EMA_QI = 0.2;
const DEFAULT_ALPHA_EMA_PERIOD = 0.2;
// MBST's DominantCyclePeriod default warm-up is MaxPeriod * 2 = 100.
const DEFAULT_WARM_UP_PERIOD = 100;

const DEG2RAD = Math.PI / 180.0;
const LEAD_OFFSET = 45.0;

/** __Sine Wave__ (Ehlers) computes a clear sine wave representation of the dominant cycle phase.
 *
 * It exposes five outputs:
 *
 *	- Value: the sine wave value, sin(phase·Deg2Rad).
 *	- Lead: the sine wave lead value, sin((phase+45)·Deg2Rad).
 *	- Band: a band with Value as the upper line and Lead as the lower line.
 *	- DominantCyclePeriod: the smoothed dominant cycle period.
 *	- DominantCyclePhase: the dominant cycle phase, in degrees.
 *
 * Reference:
 *
 *	John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 95-105.
 */
export class SineWave implements Indicator {
  private readonly dc: DominantCycle;
  private primed = false;
  private value = Number.NaN;
  private lead = Number.NaN;

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;
  private readonly mnemonicLead: string;
  private readonly descriptionLead: string;
  private readonly mnemonicBand: string;
  private readonly descriptionBand: string;
  private readonly mnemonicDCP: string;
  private readonly descriptionDCP: string;
  private readonly mnemonicDCPhase: string;
  private readonly descriptionDCPhase: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Creates an instance using default parameters (α=0.33, HomodyneDiscriminator cycle
   * estimator with smoothingLength=4, αq=0.2, αp=0.2, warmUpPeriod=100, BarComponent.Median). */
  public static default(): SineWave {
    return new SineWave({
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
  public static fromParams(params: SineWaveParams): SineWave {
    return new SineWave(params);
  }

  private constructor(params: SineWaveParams) {
    const alpha = params.alphaEmaPeriodAdditional;
    if (alpha <= 0 || alpha > 1) {
      throw new Error('invalid sine wave parameters: α for additional smoothing should be in range (0, 1]');
    }

    // SineWave defaults to BarComponent.Median (not the framework default), so it always
    // shows in the mnemonic.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // Inner DominantCycle built with explicit components so its mnemonic contains hl/2.
    this.dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: alpha,
      estimatorType: params.estimatorType,
      estimatorParams: params.estimatorParams,
      barComponent: bc,
      quoteComponent: qc,
      tradeComponent: tc,
    });

    // Compose estimator moniker (same logic as DominantCycle).
    const effectiveType = params.estimatorType ?? HilbertTransformerCycleEstimatorType.HomodyneDiscriminator;
    const htce = createEstimator(params.estimatorType, params.estimatorParams);
    let em = '';
    const isDefaultHd = effectiveType === HilbertTransformerCycleEstimatorType.HomodyneDiscriminator
      && htce.smoothingLength === DEFAULT_SMOOTHING_LENGTH
      && htce.alphaEmaQuadratureInPhase === DEFAULT_ALPHA_EMA_QI
      && htce.alphaEmaPeriod === DEFAULT_ALPHA_EMA_PERIOD;
    if (!isDefaultHd) {
      const moniker = estimatorMoniker(effectiveType, htce);
      if (moniker.length > 0) {
        em = ', ' + moniker;
      }
    }

    const cm = componentTripleMnemonic(bc, qc, tc);
    const a = alpha.toFixed(3);

    this.mnemonicValue = `sw(${a}${em}${cm})`;
    this.mnemonicLead = `sw-lead(${a}${em}${cm})`;
    this.mnemonicBand = `sw-band(${a}${em}${cm})`;
    this.mnemonicDCP = `dcp(${a}${em}${cm})`;
    this.mnemonicDCPhase = `dcph(${a}${em}${cm})`;

    this.descriptionValue = 'Sine wave ' + this.mnemonicValue;
    this.descriptionLead = 'Sine wave lead ' + this.mnemonicLead;
    this.descriptionBand = 'Sine wave band ' + this.mnemonicBand;
    this.descriptionDCP = 'Dominant cycle period ' + this.mnemonicDCP;
    this.descriptionDCPhase = 'Dominant cycle phase ' + this.mnemonicDCPhase;

    // Silence unused-variable warning for DefaultBarComponent import (kept for consistency).
    void DefaultBarComponent;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.SineWave,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
        { mnemonic: this.mnemonicLead, description: this.descriptionLead },
        { mnemonic: this.mnemonicBand, description: this.descriptionBand },
        { mnemonic: this.mnemonicDCP, description: this.descriptionDCP },
        { mnemonic: this.mnemonicDCPhase, description: this.descriptionDCPhase },
      ],
    );
  }

  /** Updates the indicator given the next sample value. Returns the quadruple
   * (value, lead, period, phase). Returns all-NaN if not yet primed. */
  public update(sample: number): [number, number, number, number] {
    if (Number.isNaN(sample)) {
      return [sample, sample, sample, sample];
    }

    const [, period, phase] = this.dc.update(sample);

    if (Number.isNaN(phase)) {
      return [Number.NaN, Number.NaN, Number.NaN, Number.NaN];
    }

    this.primed = true;
    this.value = Math.sin(phase * DEG2RAD);
    this.lead = Math.sin((phase + LEAD_OFFSET) * DEG2RAD);
    return [this.value, this.lead, period, phase];
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
    const [value, lead, period, phase] = this.update(sample);

    const sv = new Scalar();
    sv.time = time;
    sv.value = value;

    const sl = new Scalar();
    sl.time = time;
    sl.value = lead;

    const band = new Band();
    band.time = time;
    band.upper = value;
    band.lower = lead;

    const sp = new Scalar();
    sp.time = time;
    sp.value = period;

    const sph = new Scalar();
    sph.time = time;
    sph.value = phase;

    return [sv, sl, band, sp, sph];
  }
}
