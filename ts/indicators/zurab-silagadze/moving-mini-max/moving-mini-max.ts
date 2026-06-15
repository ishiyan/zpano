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
import { Levels, newLevel } from '../../core/outputs/levels';
import { Polyline, Point } from '../../core/outputs/polyline';
import { MovingMiniMaxParams } from './params';

/** A detected support/resistance level. */
interface MiniMaxLevel {
  price: number;
  offset: number;
  strength: number;
}

/** A detected peak as a (strength, index) pair. */
interface Peak {
  strength: number;
  index: number;
}

/** A computed MMM result set. */
interface MiniMaxResult {
  up: number;
  down: number;
  resistances: MiniMaxLevel[];
  supports: MiniMaxLevel[];
  upDist: number[];
  downDist: number[];
  valid: boolean;
}

/** Computes Q_{i,i+1} and Q_{i,i-1} for each position i = 0..n-1. */
function calcQValues(window: number[], n: number, m: number, negate: boolean): [number[], number[]] {
  const sign = negate ? -1.0 : 1.0;
  const qPlus = new Array<number>(n).fill(0);
  const qMinus = new Array<number>(n).fill(0);

  for (let i = 0; i < n; i++) {
    const si = window[i];
    let sumPlus = 0;
    let sumMinus = 0;

    for (let k = 1; k <= m; k++) {
      const sForward = i + k < n ? window[i + k] : window[n - 1];
      const sBackward = i - k >= 0 ? window[i - k] : window[0];

      const denomPlus = sForward + si;
      const argPlus = denomPlus === 0 ? 0 : (sign * 2.0 * (sForward - si)) / denomPlus;

      const denomMinus = sBackward + si;
      const argMinus = denomMinus === 0 ? 0 : (sign * 2.0 * (sBackward - si)) / denomMinus;

      sumPlus += Math.exp(argPlus);
      sumMinus += Math.exp(argMinus);
    }

    qPlus[i] = sumPlus;
    qMinus[i] = sumMinus;
  }

  return [qPlus, qMinus];
}

/** Computes transition probabilities P_{i,i+1} and P_{i,i-1} from Q-values. */
function calcPValues(qPlus: number[], qMinus: number[], n: number): [number[], number[]] {
  const pPlus = new Array<number>(n).fill(0);
  const pMinus = new Array<number>(n).fill(0);

  for (let i = 0; i < n; i++) {
    const denom = qPlus[i] + qMinus[i];
    if (denom === 0) {
      pPlus[i] = 0.5;
      pMinus[i] = 0.5;
    } else {
      pPlus[i] = qPlus[i] / denom;
      pMinus[i] = qMinus[i] / denom;
    }
  }

  return [pPlus, pMinus];
}

/** Computes the normalized mini-max series from transition probabilities. */
function calcMiniMax(pPlus: number[], pMinus: number[], n: number): number[] {
  const u = new Array<number>(n).fill(0);
  u[0] = 1.0;

  for (let i = 1; i < n; i++) {
    const pPrevToI = pPlus[i - 1];
    const pIToPrev = pMinus[i];
    if (pIToPrev === 0) {
      u[i] = u[i - 1] * 1e10;
    } else {
      u[i] = (pPrevToI / pIToPrev) * u[i - 1];
    }
  }

  let total = 0;
  for (let i = 0; i < n; i++) {
    total += u[i];
  }

  const minimax = new Array<number>(n);
  if (total === 0) {
    for (let i = 0; i < n; i++) {
      minimax[i] = 1.0 / n;
    }
    return minimax;
  }

  for (let i = 0; i < n; i++) {
    minimax[i] = u[i] / total;
  }
  return minimax;
}

/** Finds distinct local peaks, returned sorted by strength descending. */
function findPeaks(values: number[], numPeaks: number, minSeparation: number): Peak[] {
  const n = values.length;
  const candidates: Peak[] = [];

  for (let i = 0; i < n; i++) {
    let isPeak: boolean;
    if (i === 0) {
      isPeak = n <= 1 || values[i] >= values[i + 1];
    } else if (i === n - 1) {
      isPeak = values[i] >= values[i - 1];
    } else {
      isPeak = values[i] >= values[i - 1] && values[i] >= values[i + 1];
    }
    if (isPeak) {
      candidates.push({ strength: values[i], index: i });
    }
  }

  // Sort by strength descending; ties break on the larger index first (matches the
  // reference, which sorts (value, index) tuples in reverse).
  candidates.sort((a, b) => {
    if (a.strength !== b.strength) {
      return b.strength - a.strength;
    }
    return b.index - a.index;
  });

  const selected: Peak[] = [];
  for (const c of candidates) {
    if (selected.length >= numPeaks) {
      break;
    }
    let tooClose = false;
    for (const sel of selected) {
      if (Math.abs(c.index - sel.index) < minSeparation) {
        tooClose = true;
        break;
      }
    }
    if (!tooClose) {
      selected.push(c);
    }
  }

  return selected;
}

/** Function to calculate the mnemonic of a __MovingMiniMax__ indicator. */
export const movingMiniMaxMnemonic = (params: MovingMiniMaxParams): string => {
  const m = Math.floor(params.m ?? 5);
  const n = Math.floor(params.n ?? 50);
  const numExtrema = Math.floor(params.numExtrema ?? 3);
  const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);

  return `mmm(${m},${n},${numExtrema}${cm})`;
};

/**
 * MovingMiniMax is Zurab Silagadze's Moving Mini-Max (MMM) indicator.
 *
 * A nonlinear indicator for technical analysis that emphasizes local maximums and minimums
 * in a price series with inherent smoothing. The algorithm is borrowed from gamma-ray
 * spectroscopy peak finding and models price exploration as a quantum particle that can
 * tunnel through small noise barriers but is stopped by genuine trend reversals.
 *
 * Reference:
 *
 * Silagadze, Z. K. (2011). Moving Mini-Max -- a new indicator for technical analysis.
 * IFTA Journal 11, 46-49. arXiv:0802.0984v2.
 */
export class MovingMiniMax implements Indicator {

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly m: number;
  private readonly n: number;
  private readonly numExtrema: number;

  private readonly window: number[];
  private bufPos = 0;
  private count = 0;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  public constructor(params?: MovingMiniMaxParams) {
    const p = params ?? {};

    const m = Math.floor(p.m ?? 5);
    const n = Math.floor(p.n ?? 50);
    const numExtrema = Math.floor(p.numExtrema ?? 3);

    if (m < 1) {
      throw new Error('m should be >= 1');
    }
    if (n <= 2 * m) {
      throw new Error('n should be > 2*m');
    }
    if (numExtrema < 1) {
      throw new Error('num extrema should be >= 1');
    }

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.m = m;
    this.n = n;
    this.numExtrema = numExtrema;

    this.window = new Array<number>(n).fill(0);

    this.mnemonic_ = movingMiniMaxMnemonic(p);
    this.description_ = 'Moving mini-max ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.MovingMiniMax,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' up', description: this.description_ + ' up value' },
        { mnemonic: this.mnemonic_ + ' down', description: this.description_ + ' down value' },
        { mnemonic: this.mnemonic_ + ' resistances', description: this.description_ + ' resistances' },
        { mnemonic: this.mnemonic_ + ' supports', description: this.description_ + ' supports' },
        { mnemonic: this.mnemonic_ + ' up dist', description: this.description_ + ' up distribution' },
        { mnemonic: this.mnemonic_ + ' down dist', description: this.description_ + ' down distribution' },
      ],
    );
  }

  /** Updates the indicator and returns the computed MMM result set. */
  public update(sample: number): MiniMaxResult {
    const empty: MiniMaxResult = {
      up: NaN, down: NaN, resistances: [], supports: [], upDist: [], downDist: [], valid: false,
    };

    if (this.count < this.n) {
      this.window[this.count] = sample;
      this.count++;
    } else {
      this.window[this.bufPos] = sample;
      this.bufPos = (this.bufPos + 1) % this.n;
    }

    if (this.count < this.n) {
      this.primed_ = false;
      return empty;
    }

    this.primed_ = true;

    const n = this.n;
    const m = this.m;

    // Reconstruct the window in chronological order (oldest -> newest).
    const window = new Array<number>(n);
    for (let i = 0; i < n; i++) {
      window[i] = this.window[(this.bufPos + i) % n];
    }

    const [qUpPlus, qUpMinus] = calcQValues(window, n, m, false);
    const [qDnPlus, qDnMinus] = calcQValues(window, n, m, true);

    const [pUpPlus, pUpMinus] = calcPValues(qUpPlus, qUpMinus, n);
    const [pDnPlus, pDnMinus] = calcPValues(qDnPlus, qDnMinus, n);

    const upDist = calcMiniMax(pUpPlus, pUpMinus, n);
    const dnDist = calcMiniMax(pDnPlus, pDnMinus, n);

    const minSep = Math.max(m, 2);

    const uPeaks = findPeaks(upDist, this.numExtrema, minSep);
    const dPeaks = findPeaks(dnDist, this.numExtrema, minSep);

    const resistances: MiniMaxLevel[] = uPeaks.map((pk) => ({
      price: window[pk.index], offset: n - 1 - pk.index, strength: pk.strength,
    }));
    const supports: MiniMaxLevel[] = dPeaks.map((pk) => ({
      price: window[pk.index], offset: n - 1 - pk.index, strength: pk.strength,
    }));

    return {
      up: upDist[n - 1],
      down: dnDist[n - 1],
      resistances,
      supports,
      upDist,
      downDist: dnDist,
      valid: true,
    };
  }

  private wrap(time: Date, r: MiniMaxResult): IndicatorOutput {
    const s0 = new Scalar(); s0.time = time; s0.value = r.valid ? r.up : NaN;
    const s1 = new Scalar(); s1.time = time; s1.value = r.valid ? r.down : NaN;
    return [
      s0,
      s1,
      this.levelsOf(time, r.resistances),
      this.levelsOf(time, r.supports),
      this.polylineOf(time, r.upDist),
      this.polylineOf(time, r.downDist),
    ];
  }

  private levelsOf(time: Date, levels: MiniMaxLevel[]): Levels {
    if (levels.length === 0) {
      return Levels.newEmptyLevels(time);
    }
    return Levels.newLevels(time, levels.map((lv) => newLevel(lv.price, lv.offset, lv.strength)));
  }

  private polylineOf(time: Date, values: number[]): Polyline {
    if (values.length === 0) {
      return Polyline.newEmptyPolyline(time);
    }
    const points: Point[] = values.map((v, i) => ({ offset: i, value: v }));
    return Polyline.newPolyline(time, points);
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
