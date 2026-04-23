/** Internal discrete Fourier transform power spectrum estimator. Port of MBST's
 * DiscreteFourierTransformSpectrumEstimator used only by the DiscreteFourierTransformSpectrum
 * indicator; not exported from the module barrel. */
export class DiscreteFourierTransformSpectrumEstimator {
  public readonly length: number;
  public readonly spectrumResolution: number;
  public readonly lengthSpectrum: number;
  public readonly maxOmegaLength: number;
  public readonly minPeriod: number;
  public readonly maxPeriod: number;
  public readonly isSpectralDilationCompensation: boolean;
  public readonly isAutomaticGainControl: boolean;
  public readonly automaticGainControlDecayFactor: number;

  public readonly inputSeries: number[];
  public readonly inputSeriesMinusMean: number[];
  public readonly spectrum: number[];
  public readonly period: number[];

  // Pre-computed trigonometric tables, size [lengthSpectrum][maxOmegaLength].
  // maxOmegaLength equals length (full-window DFT).
  private readonly frequencySinOmega: number[][];
  private readonly frequencyCosOmega: number[][];

  public mean = 0;
  public spectrumMin = 0;
  public spectrumMax = 0;
  public previousSpectrumMax = 0;

  constructor(
    length: number,
    minPeriod: number,
    maxPeriod: number,
    spectrumResolution: number,
    isSpectralDilationCompensation: boolean,
    isAutomaticGainControl: boolean,
    automaticGainControlDecayFactor: number,
  ) {
    const twoPi = 2 * Math.PI;

    const lengthSpectrum = Math.trunc((maxPeriod - minPeriod) * spectrumResolution) + 1;
    const maxOmegaLength = length;

    this.length = length;
    this.spectrumResolution = spectrumResolution;
    this.lengthSpectrum = lengthSpectrum;
    this.maxOmegaLength = maxOmegaLength;
    this.minPeriod = minPeriod;
    this.maxPeriod = maxPeriod;
    this.isSpectralDilationCompensation = isSpectralDilationCompensation;
    this.isAutomaticGainControl = isAutomaticGainControl;
    this.automaticGainControlDecayFactor = automaticGainControlDecayFactor;

    this.inputSeries = new Array<number>(length).fill(0);
    this.inputSeriesMinusMean = new Array<number>(length).fill(0);
    this.spectrum = new Array<number>(lengthSpectrum).fill(0);
    this.period = new Array<number>(lengthSpectrum).fill(0);

    this.frequencySinOmega = new Array<number[]>(lengthSpectrum);
    this.frequencyCosOmega = new Array<number[]>(lengthSpectrum);

    const result = spectrumResolution;

    // Spectrum is evaluated from MaxPeriod down to MinPeriod with the configured resolution.
    for (let i = 0; i < lengthSpectrum; i++) {
      const p = maxPeriod - i / result;
      this.period[i] = p;
      const theta = twoPi / p;

      const sinRow = new Array<number>(maxOmegaLength);
      const cosRow = new Array<number>(maxOmegaLength);
      for (let j = 0; j < maxOmegaLength; j++) {
        const omega = j * theta;
        sinRow[j] = Math.sin(omega);
        cosRow[j] = Math.cos(omega);
      }
      this.frequencySinOmega[i] = sinRow;
      this.frequencyCosOmega[i] = cosRow;
    }
  }

  /** Fills mean, inputSeriesMinusMean, spectrum, spectrumMin, spectrumMax
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

    // Evaluate the DFT power spectrum.
    this.spectrumMin = Number.MAX_VALUE;
    if (this.isAutomaticGainControl) {
      this.spectrumMax = this.automaticGainControlDecayFactor * this.previousSpectrumMax;
    } else {
      this.spectrumMax = -Number.MAX_VALUE;
    }

    for (let i = 0; i < this.lengthSpectrum; i++) {
      const sinRow = this.frequencySinOmega[i];
      const cosRow = this.frequencyCosOmega[i];

      let sumSin = 0;
      let sumCos = 0;

      for (let j = 0; j < this.maxOmegaLength; j++) {
        const sample = this.inputSeriesMinusMean[j];
        sumSin += sample * sinRow[j];
        sumCos += sample * cosRow[j];
      }

      let s = sumSin * sumSin + sumCos * sumCos;
      if (this.isSpectralDilationCompensation) {
        s /= this.period[i];
      }

      this.spectrum[i] = s;

      if (this.spectrumMax < s) this.spectrumMax = s;
      if (this.spectrumMin > s) this.spectrumMin = s;
    }

    this.previousSpectrumMax = this.spectrumMax;
  }
}
