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
import { Levels, newValueLevel } from '../../core/outputs/levels';
import { QuantumPriceLevelsParams } from './params';

/** Signed real cube root via pow (matches the reference implementation). */
function cbrt(x: number): number {
  return x >= 0 ? Math.pow(x, 1 / 3) : -Math.pow(-x, 1 / 3);
}

/** K0 constant for energy level n (Dasgupta et al. 2007). */
function computeK0(n: number): number {
  const numerator = 1.1924 + 33.2383 * n + 56.2169 * n * n;
  const denominator = 1.0 + 43.6106 * n;
  return Math.pow(numerator / denominator, 1 / 3);
}

/** A computed QPL result set. */
interface QplResult {
  lambda: number;
  sigma: number;
  nqpr: number[];
  resistances: number[];
  supports: number[];
  valid: boolean;
}

/** Function to calculate the mnemonic of a __QuantumPriceLevels__ indicator. */
export const quantumPriceLevelsMnemonic = (params: QuantumPriceLevelsParams): string => {
  const lookback = Math.floor(params.lookback ?? 2048);
  const numLevels = Math.floor(params.numLevels ?? 21);
  const numBins = Math.floor(params.numBins ?? 100);
  const scaleFactor = params.scaleFactor ?? 0.21;
  const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);

  return `qpl(${lookback},${numLevels},${numBins},${formatScale(scaleFactor)}${cm})`;
};

function formatScale(v: number): string {
  return `${v}`;
}

/**
 * QuantumPriceLevels is Raymond Lee's Quantum Price Levels (QPL) indicator.
 *
 * It computes discrete support/resistance price levels from a quantum-finance analogy:
 * the market is modelled as a quantum anharmonic oscillator, and the discrete energy
 * eigenvalues of the system map to price levels above and below the current price.
 *
 * Reference:
 *
 * Lee, R. S. T. (2021). Quantum Finance Forecast System with Quantum Anharmonic
 * Oscillator Model for Quantum Price Level Modeling. IAJER, 4(02), 1-21.
 */
export class QuantumPriceLevels implements Indicator {

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly lookback: number;
  private readonly numLevels: number;
  private readonly numBins: number;
  private readonly scaleFactor: number;
  private readonly k: number[];

  private readonly returns: number[];
  private bufPos = 0;
  private count = 0;
  private prevPrice = NaN;
  private havePrev = false;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  public constructor(params?: QuantumPriceLevelsParams) {
    const p = params ?? {};

    const lookback = Math.floor(p.lookback ?? 2048);
    const numLevels = Math.floor(p.numLevels ?? 21);
    const numBins = Math.floor(p.numBins ?? 100);
    const scaleFactor = p.scaleFactor ?? 0.21;

    if (lookback < 2) {
      throw new Error('lookback should be >= 2');
    }
    if (numLevels < 1) {
      throw new Error('num levels should be >= 1');
    }
    if (numBins < 2) {
      throw new Error('num bins should be >= 2');
    }
    if (scaleFactor <= 0) {
      throw new Error('scale factor should be > 0');
    }

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.lookback = lookback;
    this.numLevels = numLevels;
    this.numBins = numBins;
    this.scaleFactor = scaleFactor;

    this.k = new Array<number>(numLevels);
    for (let n = 0; n < numLevels; n++) {
      this.k[n] = computeK0(n);
    }

    this.returns = new Array<number>(lookback).fill(0);

    this.mnemonic_ = quantumPriceLevelsMnemonic(p);
    this.description_ = 'Quantum price levels ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.QuantumPriceLevels,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' lambda', description: this.description_ + ' anharmonic coefficient' },
        { mnemonic: this.mnemonic_ + ' stddev', description: this.description_ + ' return standard deviation' },
        { mnemonic: this.mnemonic_ + ' nqpr', description: this.description_ + ' normalized multipliers' },
        { mnemonic: this.mnemonic_ + ' resistances', description: this.description_ + ' resistance levels' },
        { mnemonic: this.mnemonic_ + ' supports', description: this.description_ + ' support levels' },
      ],
    );
  }

  /** Updates the indicator and returns the computed QPL result set. */
  public update(sample: number): QplResult {
    const empty: QplResult = { lambda: NaN, sigma: NaN, nqpr: [], resistances: [], supports: [], valid: false };

    if (!this.havePrev) {
      this.prevPrice = sample;
      this.havePrev = true;
      this.primed_ = false;
      return empty;
    }

    const newReturn = sample > 0 ? this.prevPrice / sample : 1.0;
    this.prevPrice = sample;

    if (this.count < this.lookback) {
      this.returns[this.count] = newReturn;
      this.count++;
    } else {
      this.returns[this.bufPos] = newReturn;
      this.bufPos = (this.bufPos + 1) % this.lookback;
    }

    if (this.count < this.lookback) {
      this.primed_ = false;
      return empty;
    }

    this.primed_ = true;

    const lookback = this.lookback;
    const numBins = this.numBins;
    const numLevels = this.numLevels;
    const scaleFactor = this.scaleFactor;

    // Statistics (population mu, sigma).
    let sumR = 0;
    for (let i = 0; i < lookback; i++) {
      sumR += this.returns[i];
    }
    const mu = sumR / lookback;

    let sumVar = 0;
    for (let i = 0; i < lookback; i++) {
      const diff = this.returns[i] - mu;
      sumVar += diff * diff;
    }
    const sigma = Math.sqrt(sumVar / lookback);
    if (sigma === 0) {
      return empty;
    }

    // Histogram centred at r = 1.
    const halfBins = Math.floor(numBins / 2);
    const dr = (3.0 * sigma) / halfBins;
    const leftBoundary = 1.0 - halfBins * dr;

    const q = new Array<number>(numBins).fill(0);
    let totalCount = 0;
    for (let i = 0; i < lookback; i++) {
      const r = this.returns[i];
      const binIndex = Math.floor((r - leftBoundary) / dr);
      if (binIndex >= 0 && binIndex < numBins) {
        q[binIndex]++;
        totalCount++;
      }
    }

    if (totalCount === 0) {
      return empty;
    }

    // Ground state (peak bin).
    let maxQ = 0;
    let maxQno = 0;
    for (let k = 0; k < numBins; k++) {
      const nq = q[k] / totalCount;
      if (nq > maxQ) {
        maxQ = nq;
        maxQno = k;
      }
    }

    if (maxQno === 0 || maxQno === numBins - 1) {
      return empty;
    }

    // lambda via FDM.
    const phiPlus1 = q[maxQno + 1] / totalCount;
    const phiMinus1 = q[maxQno - 1] / totalCount;

    const rPeak = leftBoundary + maxQno * dr;
    const r0 = rPeak - dr / 2.0;
    const rPlus1 = r0 + dr;
    const rMinus1 = r0 - dr;

    const lUp = rMinus1 * rMinus1 * phiMinus1 - rPlus1 * rPlus1 * phiPlus1;
    const lDw = rPlus1 * rPlus1 * rPlus1 * rPlus1 * phiPlus1 - rMinus1 * rMinus1 * rMinus1 * rMinus1 * phiMinus1;

    if (lDw === 0) {
      return empty;
    }

    const lambda = Math.abs(lUp / lDw);

    // Energy levels via Cardano.
    const qfel = new Array<number>(numLevels);
    for (let n = 0; n < numLevels; n++) {
      const twoNPlus1 = 2 * n + 1;
      const pp = -(twoNPlus1 * twoNPlus1);
      const qCoef = -lambda * (twoNPlus1 * twoNPlus1 * twoNPlus1) * (this.k[n] * this.k[n] * this.k[n]);
      const discriminant = (qCoef * qCoef) / 4.0 + (pp * pp * pp) / 27.0;
      if (discriminant < 0) {
        return empty;
      }
      const sqrtD = Math.sqrt(discriminant);
      const u = cbrt(-qCoef / 2.0 + sqrtD);
      const v = cbrt(-qCoef / 2.0 - sqrtD);
      qfel[n] = u + v;
    }

    if (qfel[0] === 0) {
      return empty;
    }

    // NQPR and projection from the current price.
    const nqpr = new Array<number>(numLevels);
    const resistances = new Array<number>(numLevels);
    const supports = new Array<number>(numLevels);
    for (let n = 0; n < numLevels; n++) {
      const qpr = qfel[n] / qfel[0];
      nqpr[n] = 1.0 + scaleFactor * sigma * qpr;
      resistances[n] = sample * nqpr[n];
      supports[n] = sample / nqpr[n];
    }

    return { lambda, sigma, nqpr, resistances, supports, valid: true };
  }

  private wrap(time: Date, r: QplResult): IndicatorOutput {
    const s0 = new Scalar(); s0.time = time; s0.value = r.valid ? r.lambda : NaN;
    const s1 = new Scalar(); s1.time = time; s1.value = r.valid ? r.sigma : NaN;
    return [
      s0,
      s1,
      this.levelsOf(time, r.nqpr),
      this.levelsOf(time, r.resistances),
      this.levelsOf(time, r.supports),
    ];
  }

  private levelsOf(time: Date, values: number[]): Levels {
    if (values.length === 0) {
      return Levels.newEmptyLevels(time);
    }
    return Levels.newLevels(time, values.map((v) => newValueLevel(v)));
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.wrap(sample.time, this.update(sample.value));
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.wrap(sample.time, this.update(this.barComponentFunc(sample)));
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.wrap(sample.time, this.update(this.quoteComponentFunc(sample)));
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.wrap(sample.time, this.update(this.tradeComponentFunc(sample)));
  }
}
