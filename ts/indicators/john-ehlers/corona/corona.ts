import { CoronaParams, DefaultCoronaParams } from './params';

const DEFAULT_HIGH_PASS_FILTER_CUTOFF = 30;
const DEFAULT_MINIMAL_PERIOD = 6;
const DEFAULT_MAXIMAL_PERIOD = 30;
const DEFAULT_DECIBELS_LOWER_THRESHOLD = 6;
const DEFAULT_DECIBELS_UPPER_THRESHOLD = 20;

const HIGH_PASS_FILTER_BUFFER_SIZE = 6;
const FIR_COEF_SUM = 12;

const DELTA_LOWER_THRESHOLD = 0.1;
const DELTA_FACTOR = -0.015;
const DELTA_SUMMAND = 0.5;

const DOMINANT_CYCLE_BUFFER_SIZE = 5;
const DOMINANT_CYCLE_MEDIAN_INDEX = 2;

const DECIBELS_SMOOTHING_ALPHA = 0.33;
const DECIBELS_SMOOTHING_ONE_MINUS = 0.67;

const NORMALIZED_AMPLITUDE_FACTOR = 0.99;
const DECIBELS_FLOOR = 0.01;
const DECIBELS_GAIN = 10;

/** Per-bin state of a single bandpass filter in the Corona bank. */
export interface CoronaFilter {
  inPhase: number;
  inPhasePrevious: number;
  quadrature: number;
  quadraturePrevious: number;
  real: number;
  realPrevious: number;
  imaginary: number;
  imaginaryPrevious: number;

  /** |Real + j·Imaginary|² of the most recent update. */
  amplitudeSquared: number;

  /** Smoothed dB value of the most recent update. */
  decibels: number;
}

/** Shared spectral-analysis engine consumed by the CoronaSpectrum, CoronaSignalToNoiseRatio,
 * CoronaSwingPosition and CoronaTrendVigor indicators. It is not an indicator on its own.
 *
 * Call _update(sample)_ once per bar; then read _isPrimed_, _dominantCycle_,
 * _dominantCycleMedian_, _maximalAmplitudeSquared_ and the _filterBank_ array.
 *
 * Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities, November 2008. */
export class Corona {
  public readonly minimalPeriod: number;
  public readonly maximalPeriod: number;
  public readonly minimalPeriodTimesTwo: number;
  public readonly maximalPeriodTimesTwo: number;
  public readonly filterBankLength: number;

  private readonly decibelsLowerThreshold: number;
  private readonly decibelsUpperThreshold: number;
  private readonly alpha: number;
  private readonly halfOnePlusAlpha: number;
  private readonly preCalculatedBeta: number[];

  public readonly filterBank: CoronaFilter[];
  private readonly highPassBuffer: number[];
  private readonly dominantCycleBuffer: number[];

  private samplePrevious = 0;
  private smoothHPPrevious = 0;
  private _maximalAmplitudeSquared = 0;
  private sampleCount = 0;
  private _dominantCycle: number;
  private _dominantCycleMedian: number;
  private _primed = false;

  /** Creates a Corona engine. Zero/undefined parameters are replaced with the Ehlers defaults. */
  public constructor(params?: CoronaParams) {
    const hp = params?.highPassFilterCutoff;
    const minP = params?.minimalPeriod;
    const maxP = params?.maximalPeriod;
    const dbLo = params?.decibelsLowerThreshold;
    const dbHi = params?.decibelsUpperThreshold;

    const cfg = {
      highPassFilterCutoff: hp !== undefined && hp > 0 ? hp : DEFAULT_HIGH_PASS_FILTER_CUTOFF,
      minimalPeriod: minP !== undefined && minP > 0 ? minP : DEFAULT_MINIMAL_PERIOD,
      maximalPeriod: maxP !== undefined && maxP > 0 ? maxP : DEFAULT_MAXIMAL_PERIOD,
      decibelsLowerThreshold: dbLo !== undefined && dbLo !== 0 ? dbLo : DEFAULT_DECIBELS_LOWER_THRESHOLD,
      decibelsUpperThreshold: dbHi !== undefined && dbHi !== 0 ? dbHi : DEFAULT_DECIBELS_UPPER_THRESHOLD,
    };

    if (cfg.highPassFilterCutoff < 2) {
      throw new Error('invalid corona parameters: HighPassFilterCutoff should be >= 2');
    }
    if (cfg.minimalPeriod < 2) {
      throw new Error('invalid corona parameters: MinimalPeriod should be >= 2');
    }
    if (cfg.maximalPeriod <= cfg.minimalPeriod) {
      throw new Error('invalid corona parameters: MaximalPeriod should be > MinimalPeriod');
    }
    if (cfg.decibelsLowerThreshold < 0) {
      throw new Error('invalid corona parameters: DecibelsLowerThreshold should be >= 0');
    }
    if (cfg.decibelsUpperThreshold <= cfg.decibelsLowerThreshold) {
      throw new Error('invalid corona parameters: DecibelsUpperThreshold should be > DecibelsLowerThreshold');
    }

    this.minimalPeriod = cfg.minimalPeriod;
    this.maximalPeriod = cfg.maximalPeriod;
    this.minimalPeriodTimesTwo = cfg.minimalPeriod * 2;
    this.maximalPeriodTimesTwo = cfg.maximalPeriod * 2;
    this.decibelsLowerThreshold = cfg.decibelsLowerThreshold;
    this.decibelsUpperThreshold = cfg.decibelsUpperThreshold;

    this.filterBankLength = this.maximalPeriodTimesTwo - this.minimalPeriodTimesTwo + 1;

    this.filterBank = new Array<CoronaFilter>(this.filterBankLength);
    for (let i = 0; i < this.filterBankLength; i++) {
      this.filterBank[i] = {
        inPhase: 0, inPhasePrevious: 0,
        quadrature: 0, quadraturePrevious: 0,
        real: 0, realPrevious: 0,
        imaginary: 0, imaginaryPrevious: 0,
        amplitudeSquared: 0,
        decibels: 0,
      };
    }

    this.highPassBuffer = new Array<number>(HIGH_PASS_FILTER_BUFFER_SIZE).fill(0);

    // MBST initializes the dominant cycle buffer with MaxValue sentinels.
    this.dominantCycleBuffer = new Array<number>(DOMINANT_CYCLE_BUFFER_SIZE).fill(Number.MAX_VALUE);
    this._dominantCycle = Number.MAX_VALUE;
    this._dominantCycleMedian = Number.MAX_VALUE;

    const phi = (2 * Math.PI) / cfg.highPassFilterCutoff;
    this.alpha = (1 - Math.sin(phi)) / Math.cos(phi);
    this.halfOnePlusAlpha = 0.5 * (1 + this.alpha);

    this.preCalculatedBeta = new Array<number>(this.filterBankLength);
    for (let index = 0; index < this.filterBankLength; index++) {
      const n = this.minimalPeriodTimesTwo + index;
      this.preCalculatedBeta[index] = Math.cos((4 * Math.PI) / n);
    }
  }

  /** Indicates whether the engine has seen enough samples to produce meaningful output. */
  public isPrimed(): boolean { return this._primed; }

  /** The most recent weighted-center-of-gravity estimate of the dominant cycle period. */
  public get dominantCycle(): number { return this._dominantCycle; }

  /** The 5-sample median of the most recent dominant cycle estimates. */
  public get dominantCycleMedian(): number { return this._dominantCycleMedian; }

  /** The maximum amplitude-squared observed across the filter bank for the most recently
   * processed sample. Reset to zero at the start of every update call. */
  public get maximalAmplitudeSquared(): number { return this._maximalAmplitudeSquared; }

  /** Feeds the next sample to the engine. NaN samples are a no-op. Returns true once primed. */
  public update(sample: number): boolean {
    if (Number.isNaN(sample)) {
      return this._primed;
    }

    this.sampleCount++;

    // First sample: store as prior reference and return without further processing.
    if (this.sampleCount === 1) {
      this.samplePrevious = sample;
      return false;
    }

    // Step 1: High-pass filter.
    const hp = this.alpha * this.highPassBuffer[HIGH_PASS_FILTER_BUFFER_SIZE - 1]
      + this.halfOnePlusAlpha * (sample - this.samplePrevious);
    this.samplePrevious = sample;

    for (let i = 0; i < HIGH_PASS_FILTER_BUFFER_SIZE - 1; i++) {
      this.highPassBuffer[i] = this.highPassBuffer[i + 1];
    }
    this.highPassBuffer[HIGH_PASS_FILTER_BUFFER_SIZE - 1] = hp;

    // Step 2: 6-tap FIR smoothing {1, 2, 3, 3, 2, 1} / 12.
    const smoothHP = (this.highPassBuffer[0]
      + 2 * this.highPassBuffer[1]
      + 3 * this.highPassBuffer[2]
      + 3 * this.highPassBuffer[3]
      + 2 * this.highPassBuffer[4]
      + this.highPassBuffer[5]) / FIR_COEF_SUM;

    // Step 3: Momentum.
    const momentum = smoothHP - this.smoothHPPrevious;
    this.smoothHPPrevious = smoothHP;

    // Step 4: Adaptive delta.
    let delta = DELTA_FACTOR * this.sampleCount + DELTA_SUMMAND;
    if (delta < DELTA_LOWER_THRESHOLD) {
      delta = DELTA_LOWER_THRESHOLD;
    }

    // Step 5: Filter-bank update.
    this._maximalAmplitudeSquared = 0;
    for (let index = 0; index < this.filterBankLength; index++) {
      const n = this.minimalPeriodTimesTwo + index;
      const nf = n;

      const gamma = 1 / Math.cos((8 * Math.PI * delta) / nf);
      const a = gamma - Math.sqrt(gamma * gamma - 1);

      const quadrature = momentum * (nf / (4 * Math.PI));
      const inPhase = smoothHP;

      const halfOneMinA = 0.5 * (1 - a);
      const beta = this.preCalculatedBeta[index];
      const betaOnePlusA = beta * (1 + a);

      const f = this.filterBank[index];

      const real = halfOneMinA * (inPhase - f.inPhasePrevious) + betaOnePlusA * f.real - a * f.realPrevious;
      const imag = halfOneMinA * (quadrature - f.quadraturePrevious) + betaOnePlusA * f.imaginary - a * f.imaginaryPrevious;

      const ampSq = real * real + imag * imag;

      f.inPhasePrevious = f.inPhase;
      f.inPhase = inPhase;
      f.quadraturePrevious = f.quadrature;
      f.quadrature = quadrature;
      f.realPrevious = f.real;
      f.real = real;
      f.imaginaryPrevious = f.imaginary;
      f.imaginary = imag;
      f.amplitudeSquared = ampSq;

      if (ampSq > this._maximalAmplitudeSquared) {
        this._maximalAmplitudeSquared = ampSq;
      }
    }

    // Step 6: dB normalization and dominant-cycle weighted average.
    let numerator = 0;
    let denominator = 0;
    this._dominantCycle = 0;

    for (let index = 0; index < this.filterBankLength; index++) {
      const f = this.filterBank[index];

      let decibels = 0;
      if (this._maximalAmplitudeSquared > 0) {
        const normalized = f.amplitudeSquared / this._maximalAmplitudeSquared;
        if (normalized > 0) {
          const arg = (1 - NORMALIZED_AMPLITUDE_FACTOR * normalized) / DECIBELS_FLOOR;
          if (arg > 0) {
            decibels = DECIBELS_GAIN * Math.log10(arg);
          }
        }
      }

      decibels = DECIBELS_SMOOTHING_ALPHA * decibels + DECIBELS_SMOOTHING_ONE_MINUS * f.decibels;
      if (decibels > this.decibelsUpperThreshold) {
        decibels = this.decibelsUpperThreshold;
      }
      f.decibels = decibels;

      if (decibels <= this.decibelsLowerThreshold) {
        const n = this.minimalPeriodTimesTwo + index;
        const adjusted = this.decibelsUpperThreshold - decibels;
        numerator += n * adjusted;
        denominator += adjusted;
      }
    }

    if (denominator !== 0) {
      this._dominantCycle = (0.5 * numerator) / denominator;
    }
    if (this._dominantCycle < this.minimalPeriod) {
      this._dominantCycle = this.minimalPeriod;
    }

    // Step 7: 5-sample median of dominant cycle.
    for (let i = 0; i < DOMINANT_CYCLE_BUFFER_SIZE - 1; i++) {
      this.dominantCycleBuffer[i] = this.dominantCycleBuffer[i + 1];
    }
    this.dominantCycleBuffer[DOMINANT_CYCLE_BUFFER_SIZE - 1] = this._dominantCycle;

    const sorted = [...this.dominantCycleBuffer].sort((a, b) => a - b);
    this._dominantCycleMedian = sorted[DOMINANT_CYCLE_MEDIAN_INDEX];
    if (this._dominantCycleMedian < this.minimalPeriod) {
      this._dominantCycleMedian = this.minimalPeriod;
    }

    if (this.sampleCount < this.minimalPeriodTimesTwo) {
      return false;
    }
    this._primed = true;
    return true;
  }
}

// Re-exports avoid unused-import warning in bundlers.
export { DefaultCoronaParams };
