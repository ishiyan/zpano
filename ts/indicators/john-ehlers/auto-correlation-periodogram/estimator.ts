/** Internal Ehlers Autocorrelation Periodogram estimator. Ports the Go `estimator`
 * in `go/indicators/johnehlers/autocorrelationperiodogram/estimator.go`, which
 * follows EasyLanguage listing 8-3. Not exported from the module barrel. */
export class AutoCorrelationPeriodogramEstimator {
  private static readonly DFT_LAG_START = 3;

  public readonly minPeriod: number;
  public readonly maxPeriod: number;
  public readonly averagingLength: number;
  public readonly lengthSpectrum: number;
  public readonly filtBufferLen: number;

  public readonly isSpectralSquaring: boolean;
  public readonly isSmoothing: boolean;
  public readonly isAutomaticGainControl: boolean;
  public readonly automaticGainControlDecayFactor: number;

  // Pre-filter coefficients (scalar).
  private readonly coeffHP0: number;
  private readonly coeffHP1: number;
  private readonly coeffHP2: number;
  private readonly ssC1: number;
  private readonly ssC2: number;
  private readonly ssC3: number;

  // DFT basis tables.
  private readonly cosTab: number[][];
  private readonly sinTab: number[][];

  // Pre-filter state.
  private close0 = 0;
  private close1 = 0;
  private close2 = 0;
  private hp0 = 0;
  private hp1 = 0;
  private hp2 = 0;

  private readonly filt: number[];
  private readonly corr: number[];
  private readonly rPrevious: number[];

  public readonly spectrum: number[];
  public spectrumMin = 0;
  public spectrumMax = 0;
  public previousSpectrumMax = 0;

  constructor(
    minPeriod: number,
    maxPeriod: number,
    averagingLength: number,
    isSpectralSquaring: boolean,
    isSmoothing: boolean,
    isAutomaticGainControl: boolean,
    automaticGainControlDecayFactor: number,
  ) {
    const twoPi = 2 * Math.PI;
    const dftLagStart = AutoCorrelationPeriodogramEstimator.DFT_LAG_START;

    const lengthSpectrum = maxPeriod - minPeriod + 1;
    const filtBufferLen = maxPeriod + averagingLength;
    const corrLen = maxPeriod + 1;

    // Highpass coefficients, cutoff at MaxPeriod.
    const omegaHP = 0.707 * twoPi / maxPeriod;
    const alphaHP = (Math.cos(omegaHP) + Math.sin(omegaHP) - 1) / Math.cos(omegaHP);
    const cHP0 = (1 - alphaHP / 2) * (1 - alphaHP / 2);
    const cHP1 = 2 * (1 - alphaHP);
    const cHP2 = (1 - alphaHP) * (1 - alphaHP);

    // SuperSmoother coefficients, period = MinPeriod.
    const a1 = Math.exp(-1.414 * Math.PI / minPeriod);
    const b1 = 2 * a1 * Math.cos(1.414 * Math.PI / minPeriod);
    const ssC2 = b1;
    const ssC3 = -a1 * a1;
    const ssC1 = 1 - ssC2 - ssC3;

    this.minPeriod = minPeriod;
    this.maxPeriod = maxPeriod;
    this.averagingLength = averagingLength;
    this.lengthSpectrum = lengthSpectrum;
    this.filtBufferLen = filtBufferLen;
    this.isSpectralSquaring = isSpectralSquaring;
    this.isSmoothing = isSmoothing;
    this.isAutomaticGainControl = isAutomaticGainControl;
    this.automaticGainControlDecayFactor = automaticGainControlDecayFactor;
    this.coeffHP0 = cHP0;
    this.coeffHP1 = cHP1;
    this.coeffHP2 = cHP2;
    this.ssC1 = ssC1;
    this.ssC2 = ssC2;
    this.ssC3 = ssC3;

    this.cosTab = new Array<number[]>(lengthSpectrum);
    this.sinTab = new Array<number[]>(lengthSpectrum);

    for (let i = 0; i < lengthSpectrum; i++) {
      const period = minPeriod + i;
      const cosRow = new Array<number>(corrLen).fill(0);
      const sinRow = new Array<number>(corrLen).fill(0);
      for (let n = dftLagStart; n < corrLen; n++) {
        const angle = twoPi * n / period;
        cosRow[n] = Math.cos(angle);
        sinRow[n] = Math.sin(angle);
      }
      this.cosTab[i] = cosRow;
      this.sinTab[i] = sinRow;
    }

    this.filt = new Array<number>(filtBufferLen).fill(0);
    this.corr = new Array<number>(corrLen).fill(0);
    this.rPrevious = new Array<number>(lengthSpectrum).fill(0);
    this.spectrum = new Array<number>(lengthSpectrum).fill(0);
  }

  /** Advances the estimator by one input sample and evaluates the spectrum. */
  public update(sample: number): void {
    const dftLagStart = AutoCorrelationPeriodogramEstimator.DFT_LAG_START;

    this.close2 = this.close1;
    this.close1 = this.close0;
    this.close0 = sample;

    this.hp2 = this.hp1;
    this.hp1 = this.hp0;
    this.hp0 = this.coeffHP0 * (this.close0 - 2 * this.close1 + this.close2)
      + this.coeffHP1 * this.hp1
      - this.coeffHP2 * this.hp2;

    // Shift Filt history rightward.
    const filt = this.filt;
    const filtLen = this.filtBufferLen;
    for (let k = filtLen - 1; k >= 1; k--) {
      filt[k] = filt[k - 1];
    }

    filt[0] = this.ssC1 * (this.hp0 + this.hp1) / 2 + this.ssC2 * filt[1] + this.ssC3 * filt[2];

    // Pearson correlation per lag [0..maxPeriod], fixed M = averagingLength.
    const m = this.averagingLength;
    const corr = this.corr;

    for (let lag = 0; lag <= this.maxPeriod; lag++) {
      let sx = 0;
      let sy = 0;
      let sxx = 0;
      let syy = 0;
      let sxy = 0;

      for (let c = 0; c < m; c++) {
        const x = filt[c];
        const y = filt[lag + c];
        sx += x;
        sy += y;
        sxx += x * x;
        syy += y * y;
        sxy += x * y;
      }

      const denom = (m * sxx - sx * sx) * (m * syy - sy * sy);
      let r = 0;
      if (denom > 0) {
        r = (m * sxy - sx * sy) / Math.sqrt(denom);
      }
      corr[lag] = r;
    }

    // DFT, smoothing, and AGC.
    this.spectrumMin = Number.MAX_VALUE;
    if (this.isAutomaticGainControl) {
      this.spectrumMax = this.automaticGainControlDecayFactor * this.previousSpectrumMax;
    } else {
      this.spectrumMax = -Number.MAX_VALUE;
    }

    for (let i = 0; i < this.lengthSpectrum; i++) {
      const cosRow = this.cosTab[i];
      const sinRow = this.sinTab[i];

      let cosPart = 0;
      let sinPart = 0;

      for (let n = dftLagStart; n <= this.maxPeriod; n++) {
        cosPart += corr[n] * cosRow[n];
        sinPart += corr[n] * sinRow[n];
      }

      const sqSum = cosPart * cosPart + sinPart * sinPart;
      const raw = this.isSpectralSquaring ? sqSum * sqSum : sqSum;
      const r = this.isSmoothing ? 0.2 * raw + 0.8 * this.rPrevious[i] : raw;

      this.rPrevious[i] = r;
      this.spectrum[i] = r;

      if (this.spectrumMax < r) this.spectrumMax = r;
    }

    this.previousSpectrumMax = this.spectrumMax;

    // Normalize against the running max.
    if (this.spectrumMax > 0) {
      for (let i = 0; i < this.lengthSpectrum; i++) {
        const v = this.spectrum[i] / this.spectrumMax;
        this.spectrum[i] = v;
        if (this.spectrumMin > v) this.spectrumMin = v;
      }
    } else {
      for (let i = 0; i < this.lengthSpectrum; i++) {
        this.spectrum[i] = 0;
      }
      this.spectrumMin = 0;
    }
  }
}
