/** Internal Ehlers Autocorrelation Indicator estimator. Ports the Go `estimator`
 * in `go/indicators/johnehlers/autocorrelationindicator/estimator.go`, which
 * follows EasyLanguage listing 8-2. Not exported from the module barrel. */
export class AutoCorrelationIndicatorEstimator {
  public readonly minLag: number;
  public readonly maxLag: number;
  public readonly averagingLength: number;
  public readonly lengthSpectrum: number;
  public readonly filtBufferLen: number;

  // Pre-filter coefficients (scalar).
  private readonly coeffHP0: number;
  private readonly coeffHP1: number;
  private readonly coeffHP2: number;
  private readonly ssC1: number;
  private readonly ssC2: number;
  private readonly ssC3: number;

  // Pre-filter state.
  private close0 = 0;
  private close1 = 0;
  private close2 = 0;
  private hp0 = 0;
  private hp1 = 0;
  private hp2 = 0;

  // Filt history: filt[k] = Filt k bars ago (0 = current).
  private readonly filt: number[];

  public readonly spectrum: number[];
  public spectrumMin = 0;
  public spectrumMax = 0;

  constructor(minLag: number, maxLag: number, smoothingPeriod: number, averagingLength: number) {
    const twoPi = 2 * Math.PI;

    const lengthSpectrum = maxLag - minLag + 1;
    const mMax = averagingLength === 0 ? maxLag : averagingLength;
    const filtBufferLen = maxLag + mMax;

    // Highpass coefficients, cutoff at MaxLag.
    const omegaHP = 0.707 * twoPi / maxLag;
    const alphaHP = (Math.cos(omegaHP) + Math.sin(omegaHP) - 1) / Math.cos(omegaHP);
    const cHP0 = (1 - alphaHP / 2) * (1 - alphaHP / 2);
    const cHP1 = 2 * (1 - alphaHP);
    const cHP2 = (1 - alphaHP) * (1 - alphaHP);

    // SuperSmoother coefficients, period = SmoothingPeriod.
    const a1 = Math.exp(-1.414 * Math.PI / smoothingPeriod);
    const b1 = 2 * a1 * Math.cos(1.414 * Math.PI / smoothingPeriod);
    const ssC2 = b1;
    const ssC3 = -a1 * a1;
    const ssC1 = 1 - ssC2 - ssC3;

    this.minLag = minLag;
    this.maxLag = maxLag;
    this.averagingLength = averagingLength;
    this.lengthSpectrum = lengthSpectrum;
    this.filtBufferLen = filtBufferLen;
    this.coeffHP0 = cHP0;
    this.coeffHP1 = cHP1;
    this.coeffHP2 = cHP2;
    this.ssC1 = ssC1;
    this.ssC2 = ssC2;
    this.ssC3 = ssC3;

    this.filt = new Array<number>(filtBufferLen).fill(0);
    this.spectrum = new Array<number>(lengthSpectrum).fill(0);
  }

  /** Advances the estimator by one input sample and evaluates the spectrum.
   * Callers are responsible for gating on priming. */
  public update(sample: number): void {
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
    const len = this.filtBufferLen;
    for (let k = len - 1; k >= 1; k--) {
      filt[k] = filt[k - 1];
    }

    filt[0] = this.ssC1 * (this.hp0 + this.hp1) / 2 + this.ssC2 * filt[1] + this.ssC3 * filt[2];

    // Pearson correlation per lag.
    this.spectrumMin = Number.MAX_VALUE;
    this.spectrumMax = -Number.MAX_VALUE;

    for (let i = 0; i < this.lengthSpectrum; i++) {
      const lag = this.minLag + i;
      const m = this.averagingLength === 0 ? lag : this.averagingLength;

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

      const v = 0.5 * (r + 1);
      this.spectrum[i] = v;

      if (v < this.spectrumMin) this.spectrumMin = v;
      if (v > this.spectrumMax) this.spectrumMax = v;
    }
  }
}
