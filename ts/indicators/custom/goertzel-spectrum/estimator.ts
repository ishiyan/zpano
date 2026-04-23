/** Internal Goertzel spectrum estimator. Port of MBST's GoertzelSpectrumEstimator used only
 * by the GoertzelSpectrum indicator; not exported from the module barrel. */
export class GoertzelSpectrumEstimator {
  public readonly length: number;
  public readonly spectrumResolution: number;
  public readonly lengthSpectrum: number;
  public readonly minPeriod: number;
  public readonly maxPeriod: number;
  public readonly isFirstOrder: boolean;
  public readonly isSpectralDilationCompensation: boolean;
  public readonly isAutomaticGainControl: boolean;
  public readonly automaticGainControlDecayFactor: number;

  public readonly inputSeries: number[];
  public readonly inputSeriesMinusMean: number[];
  public readonly spectrum: number[];
  public readonly period: number[];

  // Pre-computed trigonometric tables.
  private readonly frequencySin: number[]; // first-order only
  private readonly frequencyCos: number[]; // first-order only
  private readonly frequencyCos2: number[]; // second-order only

  public mean = 0;
  public spectrumMin = 0;
  public spectrumMax = 0;
  public previousSpectrumMax = 0;

  constructor(
    length: number,
    minPeriod: number,
    maxPeriod: number,
    spectrumResolution: number,
    isFirstOrder: boolean,
    isSpectralDilationCompensation: boolean,
    isAutomaticGainControl: boolean,
    automaticGainControlDecayFactor: number,
  ) {
    const twoPi = 2 * Math.PI;

    const lengthSpectrum = Math.trunc((maxPeriod - minPeriod) * spectrumResolution) + 1;

    this.length = length;
    this.spectrumResolution = spectrumResolution;
    this.lengthSpectrum = lengthSpectrum;
    this.minPeriod = minPeriod;
    this.maxPeriod = maxPeriod;
    this.isFirstOrder = isFirstOrder;
    this.isSpectralDilationCompensation = isSpectralDilationCompensation;
    this.isAutomaticGainControl = isAutomaticGainControl;
    this.automaticGainControlDecayFactor = automaticGainControlDecayFactor;

    this.inputSeries = new Array<number>(length).fill(0);
    this.inputSeriesMinusMean = new Array<number>(length).fill(0);
    this.spectrum = new Array<number>(lengthSpectrum).fill(0);
    this.period = new Array<number>(lengthSpectrum).fill(0);

    this.frequencySin = [];
    this.frequencyCos = [];
    this.frequencyCos2 = [];

    const result = spectrumResolution;

    if (isFirstOrder) {
      this.frequencySin = new Array<number>(lengthSpectrum).fill(0);
      this.frequencyCos = new Array<number>(lengthSpectrum).fill(0);

      for (let i = 0; i < lengthSpectrum; i++) {
        const period = maxPeriod - i / result;
        this.period[i] = period;
        const theta = twoPi / period;
        this.frequencySin[i] = Math.sin(theta);
        this.frequencyCos[i] = Math.cos(theta);
      }
    } else {
      this.frequencyCos2 = new Array<number>(lengthSpectrum).fill(0);

      for (let i = 0; i < lengthSpectrum; i++) {
        const period = maxPeriod - i / result;
        this.period[i] = period;
        this.frequencyCos2[i] = 2 * Math.cos(twoPi / period);
      }
    }
  }

  /** Fills mean, inputSeriesMinusMean, spectrum, spectrumMin, spectrumMax from the current
   * inputSeries contents. */
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

    // Seed with the first bin.
    let spectrum = this.goertzelEstimate(0);
    if (this.isSpectralDilationCompensation) {
      spectrum /= this.period[0];
    }

    this.spectrum[0] = spectrum;
    this.spectrumMin = spectrum;

    if (this.isAutomaticGainControl) {
      this.spectrumMax = this.automaticGainControlDecayFactor * this.previousSpectrumMax;
      if (this.spectrumMax < spectrum) {
        this.spectrumMax = spectrum;
      }
    } else {
      this.spectrumMax = spectrum;
    }

    for (let i = 1; i < this.lengthSpectrum; i++) {
      spectrum = this.goertzelEstimate(i);
      if (this.isSpectralDilationCompensation) {
        spectrum /= this.period[i];
      }

      this.spectrum[i] = spectrum;

      if (this.spectrumMax < spectrum) {
        this.spectrumMax = spectrum;
      } else if (this.spectrumMin > spectrum) {
        this.spectrumMin = spectrum;
      }
    }

    this.previousSpectrumMax = this.spectrumMax;
  }

  private goertzelEstimate(j: number): number {
    return this.isFirstOrder ? this.goertzelFirstOrderEstimate(j) : this.goertzelSecondOrderEstimate(j);
  }

  private goertzelSecondOrderEstimate(j: number): number {
    const cos2 = this.frequencyCos2[j];

    let s1 = 0;
    let s2 = 0;

    for (let i = 0; i < this.length; i++) {
      const s0 = this.inputSeriesMinusMean[i] + cos2 * s1 - s2;
      s2 = s1;
      s1 = s0;
    }

    const spectrum = s1 * s1 + s2 * s2 - cos2 * s1 * s2;
    return spectrum < 0 ? 0 : spectrum;
  }

  private goertzelFirstOrderEstimate(j: number): number {
    const cosTheta = this.frequencyCos[j];
    const sinTheta = this.frequencySin[j];

    let yre = 0;
    let yim = 0;

    for (let i = 0; i < this.length; i++) {
      const re = this.inputSeriesMinusMean[i] + cosTheta * yre - sinTheta * yim;
      const im = this.inputSeriesMinusMean[i] + cosTheta * yim + sinTheta * yre;
      yre = re;
      yim = im;
    }

    return yre * yre + yim * yim;
  }
}
