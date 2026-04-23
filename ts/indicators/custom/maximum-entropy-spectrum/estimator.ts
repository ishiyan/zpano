/** Internal maximum-entropy spectrum estimator. Port of MBST's
 * MaximumEntropySpectrumEstimator used only by the MaximumEntropySpectrum indicator;
 * not exported from the module barrel. */
export class MaximumEntropySpectrumEstimator {
  public readonly length: number;
  public readonly degree: number;
  public readonly spectrumResolution: number;
  public readonly lengthSpectrum: number;
  public readonly minPeriod: number;
  public readonly maxPeriod: number;
  public readonly isAutomaticGainControl: boolean;
  public readonly automaticGainControlDecayFactor: number;

  public readonly inputSeries: number[];
  public readonly inputSeriesMinusMean: number[];
  public readonly coefficients: number[];
  public readonly spectrum: number[];
  public readonly period: number[];

  // Pre-computed trigonometric tables, size [lengthSpectrum][degree].
  private readonly frequencySinOmega: number[][];
  private readonly frequencyCosOmega: number[][];

  // Burg working buffers.
  private readonly h: number[];
  private readonly g: number[];
  private readonly per: number[];
  private readonly pef: number[];

  public mean = 0;
  public spectrumMin = 0;
  public spectrumMax = 0;
  public previousSpectrumMax = 0;

  constructor(
    length: number,
    degree: number,
    minPeriod: number,
    maxPeriod: number,
    spectrumResolution: number,
    isAutomaticGainControl: boolean,
    automaticGainControlDecayFactor: number,
  ) {
    const twoPi = 2 * Math.PI;

    const lengthSpectrum = Math.trunc((maxPeriod - minPeriod) * spectrumResolution) + 1;

    this.length = length;
    this.degree = degree;
    this.spectrumResolution = spectrumResolution;
    this.lengthSpectrum = lengthSpectrum;
    this.minPeriod = minPeriod;
    this.maxPeriod = maxPeriod;
    this.isAutomaticGainControl = isAutomaticGainControl;
    this.automaticGainControlDecayFactor = automaticGainControlDecayFactor;

    this.inputSeries = new Array<number>(length).fill(0);
    this.inputSeriesMinusMean = new Array<number>(length).fill(0);
    this.coefficients = new Array<number>(degree).fill(0);
    this.spectrum = new Array<number>(lengthSpectrum).fill(0);
    this.period = new Array<number>(lengthSpectrum).fill(0);

    this.frequencySinOmega = new Array<number[]>(lengthSpectrum);
    this.frequencyCosOmega = new Array<number[]>(lengthSpectrum);

    this.h = new Array<number>(degree + 1).fill(0);
    this.g = new Array<number>(degree + 2).fill(0);
    this.per = new Array<number>(length + 1).fill(0);
    this.pef = new Array<number>(length + 1).fill(0);

    const result = spectrumResolution;

    // Spectrum is evaluated from MaxPeriod down to MinPeriod with the configured resolution.
    for (let i = 0; i < lengthSpectrum; i++) {
      const period = maxPeriod - i / result;
      this.period[i] = period;
      const theta = twoPi / period;

      const sinRow = new Array<number>(degree);
      const cosRow = new Array<number>(degree);
      for (let j = 0; j < degree; j++) {
        const omega = -(j + 1) * theta;
        sinRow[j] = Math.sin(omega);
        cosRow[j] = Math.cos(omega);
      }
      this.frequencySinOmega[i] = sinRow;
      this.frequencyCosOmega[i] = cosRow;
    }
  }

  /** Fills mean, inputSeriesMinusMean, coefficients, spectrum, spectrumMin, spectrumMax
   * from the current inputSeries contents. */
  public calculate(): void {
    // Subtract the mean from the input series.
    let mean = 0;
    for (let i = 0; i < this.length; i++) {
      mean += this.inputSeries[i];
    }
    mean /= this.length;

    for (let i = 0; i < this.length; i++) {
      this.inputSeriesMinusMean[i] = this.inputSeries[i] - mean;
    }
    this.mean = mean;

    this.burgEstimate(this.inputSeriesMinusMean);

    // Evaluate the spectrum from the AR coefficients.
    this.spectrumMin = Number.MAX_VALUE;
    if (this.isAutomaticGainControl) {
      this.spectrumMax = this.automaticGainControlDecayFactor * this.previousSpectrumMax;
    } else {
      this.spectrumMax = -Number.MAX_VALUE;
    }

    for (let i = 0; i < this.lengthSpectrum; i++) {
      let real = 1.0;
      let imag = 0.0;

      const cosRow = this.frequencyCosOmega[i];
      const sinRow = this.frequencySinOmega[i];

      for (let j = 0; j < this.degree; j++) {
        real -= this.coefficients[j] * cosRow[j];
        imag -= this.coefficients[j] * sinRow[j];
      }

      const s = 1.0 / (real * real + imag * imag);
      this.spectrum[i] = s;

      if (this.spectrumMax < s) this.spectrumMax = s;
      if (this.spectrumMin > s) this.spectrumMin = s;
    }

    this.previousSpectrumMax = this.spectrumMax;
  }

  /** Estimates auto-regression coefficients of the configured degree using the Burg maximum
   * entropy method. Direct port of Paul Bourke's zero-based `ar.c` reference, matching MBST. */
  private burgEstimate(series: number[]): void {
    for (let i = 1; i <= this.length; i++) {
      this.pef[i] = 0;
      this.per[i] = 0;
    }

    for (let i = 1; i <= this.degree; i++) {
      let sn = 0;
      let sd = 0;

      let jj = this.length - i;

      for (let j = 0; j < jj; j++) {
        const t1 = series[j + i] + this.pef[j];
        const t2 = series[j] + this.per[j];
        sn -= 2.0 * t1 * t2;
        sd += t1 * t1 + t2 * t2;
      }

      const t = sn / sd;
      this.g[i] = t;

      if (i !== 1) {
        for (let j = 1; j < i; j++) {
          this.h[j] = this.g[j] + t * this.g[i - j];
        }
        for (let j = 1; j < i; j++) {
          this.g[j] = this.h[j];
        }
        jj--;
      }

      for (let j = 0; j < jj; j++) {
        this.per[j] += t * this.pef[j] + t * series[j + i];
        this.pef[j] = this.pef[j + 1] + t * this.per[j + 1] + t * series[j + 1];
      }
    }

    for (let i = 0; i < this.degree; i++) {
      this.coefficients[i] = -this.g[i + 1];
    }
  }
}
