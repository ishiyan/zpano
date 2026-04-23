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
import { ExponentialMovingAverage } from '../../common/exponential-moving-average/exponential-moving-average';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';
import { MovingAverageConvergenceDivergenceParams, MovingAverageType } from './params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Function to calculate mnemonic of a __MovingAverageConvergenceDivergence__ indicator. */
export const movingAverageConvergenceDivergenceMnemonic = (
  fastLen: number, slowLen: number, signalLen: number,
  maType: MovingAverageType, signalMAType: MovingAverageType,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);
  const maLabel = (t: MovingAverageType): string => t === MovingAverageType.SMA ? 'SMA' : 'EMA';

  let suffix = '';
  if (maType !== MovingAverageType.EMA || signalMAType !== MovingAverageType.EMA) {
    suffix = `,${maLabel(maType)},${maLabel(signalMAType)}`;
  }

  return `macd(${fastLen},${slowLen},${signalLen}${suffix}${cm})`;
};

/**
 * MovingAverageConvergenceDivergence is Gerald Appel's MACD indicator.
 *
 * MACD is calculated by subtracting the slow moving average from the fast moving average.
 * A signal line (moving average of MACD) and histogram (MACD minus signal) are also produced.
 *
 * The indicator produces three outputs:
 *   - MACD: fast MA - slow MA
 *   - Signal: MA of the MACD line
 *   - Histogram: MACD - Signal
 *
 * Reference:
 *
 * Appel, Gerald (2005). Technical Analysis: Power Tools for Active Investors.
 */
export class MovingAverageConvergenceDivergence implements Indicator {

  private readonly fastMA: LineUpdater;
  private readonly slowMA: LineUpdater;
  private readonly signalMA: LineUpdater;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private macdValue_ = NaN;
  private signalValue_ = NaN;
  private histogramValue_ = NaN;
  private primed_ = false;

  private readonly fastDelay: number;
  private fastCount = 0;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: MovingAverageConvergenceDivergenceParams) {
    const p = params ?? {};

    const defaultFastLength = 12;
    const defaultSlowLength = 26;
    const defaultSignalLength = 9;

    let fastLength = Math.floor(p.fastLength ?? defaultFastLength);
    let slowLength = Math.floor(p.slowLength ?? defaultSlowLength);
    const signalLength = Math.floor(p.signalLength ?? defaultSignalLength);

    if (fastLength < 2) {
      throw new Error('fast length should be greater than 1');
    }

    if (slowLength < 2) {
      throw new Error('slow length should be greater than 1');
    }

    if (signalLength < 1) {
      throw new Error('signal length should be greater than 0');
    }

    // Auto-swap fast/slow if needed (matches TaLib behavior).
    if (slowLength < fastLength) {
      [fastLength, slowLength] = [slowLength, fastLength];
    }

    const maType = p.movingAverageType ?? MovingAverageType.EMA;
    const signalMAType = p.signalMovingAverageType ?? MovingAverageType.EMA;
    const firstIsAverage = p.firstIsAverage ?? true;

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    const createMA = (type: MovingAverageType, length: number): LineUpdater => {
      if (type === MovingAverageType.SMA) {
        return new SimpleMovingAverage({ length });
      }
      return new ExponentialMovingAverage({ length, firstIsAverage });
    };

    this.fastMA = createMA(maType, fastLength);
    this.slowMA = createMA(maType, slowLength);
    this.signalMA = createMA(signalMAType, signalLength);

    this.fastDelay = slowLength - fastLength;

    this.mnemonic_ = movingAverageConvergenceDivergenceMnemonic(
      fastLength, slowLength, signalLength, maType, signalMAType,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'Moving Average Convergence Divergence ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.MovingAverageConvergenceDivergence,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' macd', description: this.description_ + ' MACD' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' Signal' },
        { mnemonic: this.mnemonic_ + ' histogram', description: this.description_ + ' Histogram' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [macd, signal, histogram].
   */
  public update(sample: number): [number, number, number] {
    if (isNaN(sample)) {
      return [NaN, NaN, NaN];
    }

    // Feed the slow MA every sample.
    const slow = this.slowMA.update(sample);

    // Delay the fast MA to align SMA seed windows (matches TaLib batch algorithm).
    let fast: number;

    if (this.fastCount < this.fastDelay) {
      this.fastCount++;
      fast = NaN;
    } else {
      fast = this.fastMA.update(sample);
    }

    if (isNaN(fast) || isNaN(slow)) {
      this.macdValue_ = NaN;
      this.signalValue_ = NaN;
      this.histogramValue_ = NaN;
      return [NaN, NaN, NaN];
    }

    const macd = fast - slow;
    this.macdValue_ = macd;

    const signal = this.signalMA.update(macd);

    if (isNaN(signal)) {
      this.signalValue_ = NaN;
      this.histogramValue_ = NaN;
      return [macd, NaN, NaN];
    }

    this.signalValue_ = signal;
    const histogram = macd - signal;
    this.histogramValue_ = histogram;
    this.primed_ = this.fastMA.isPrimed() && this.slowMA.isPrimed() && this.signalMA.isPrimed();

    return [macd, signal, histogram];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [macd, signal, histogram] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = macd;
    const s1 = new Scalar(); s1.time = sample.time; s1.value = signal;
    const s2 = new Scalar(); s2.time = sample.time; s2.value = histogram;

    return [s0, s1, s2];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const v = this.barComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = this.quoteComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = this.tradeComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }
}
