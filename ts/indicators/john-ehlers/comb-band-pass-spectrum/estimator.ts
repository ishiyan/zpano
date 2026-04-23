/** Internal Ehlers Comb Band-Pass power spectrum estimator. Ports the Go `estimator`
 * in `go/indicators/johnehlers/combbandpassspectrum/estimator.go`, which follows
 * EasyLanguage listing 10-1. Not exported from the module barrel. */
export class CombBandPassSpectrumEstimator {
  public readonly minPeriod: number;
  public readonly maxPeriod: number;
  public readonly lengthSpectrum: number;
  public readonly isSpectralDilationCompensation: boolean;
  public readonly isAutomaticGainControl: boolean;
  public readonly automaticGainControlDecayFactor: number;

  // Pre-filter coefficients (scalar).
  private readonly coeffHP0: number;
  private readonly coeffHP1: number;
  private readonly coeffHP2: number;
  private readonly ssC1: number;
  private readonly ssC2: number;
  private readonly ssC3: number;

  // Per-bin band-pass coefficients, indexed [0..lengthSpectrum).
  // Bin i corresponds to period N = minPeriod + i.
  private readonly periods: number[];
  private readonly beta: number[];
  private readonly alpha: number[];
  private readonly comp: number[];

  // Pre-filter state.
  private close0 = 0;
  private close1 = 0;
  private close2 = 0;
  private hp0 = 0;
  private hp1 = 0;
  private hp2 = 0;
  private filt0 = 0;
  private filt1 = 0;
  private filt2 = 0;

  // Band-pass filter state. bp[i][m] holds band-pass output for bin i at lag m
  // (m=0 current, ..., m=maxPeriod-1 oldest tracked).
  private readonly bp: number[][];

  public readonly spectrum: number[];
  public spectrumMin = 0;
  public spectrumMax = 0;
  public previousSpectrumMax = 0;

  constructor(
    minPeriod: number,
    maxPeriod: number,
    bandwidth: number,
    isSpectralDilationCompensation: boolean,
    isAutomaticGainControl: boolean,
    automaticGainControlDecayFactor: number,
  ) {
    const twoPi = 2 * Math.PI;

    const lengthSpectrum = maxPeriod - minPeriod + 1;

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
    this.lengthSpectrum = lengthSpectrum;
    this.isSpectralDilationCompensation = isSpectralDilationCompensation;
    this.isAutomaticGainControl = isAutomaticGainControl;
    this.automaticGainControlDecayFactor = automaticGainControlDecayFactor;
    this.coeffHP0 = cHP0;
    this.coeffHP1 = cHP1;
    this.coeffHP2 = cHP2;
    this.ssC1 = ssC1;
    this.ssC2 = ssC2;
    this.ssC3 = ssC3;

    this.periods = new Array<number>(lengthSpectrum);
    this.beta = new Array<number>(lengthSpectrum);
    this.alpha = new Array<number>(lengthSpectrum);
    this.comp = new Array<number>(lengthSpectrum);
    this.bp = new Array<number[]>(lengthSpectrum);
    this.spectrum = new Array<number>(lengthSpectrum).fill(0);

    for (let i = 0; i < lengthSpectrum; i++) {
      const n = minPeriod + i;
      const beta = Math.cos(twoPi / n);
      const gamma = 1 / Math.cos(twoPi * bandwidth / n);
      const alpha = gamma - Math.sqrt(gamma * gamma - 1);

      this.periods[i] = n;
      this.beta[i] = beta;
      this.alpha[i] = alpha;
      this.comp[i] = isSpectralDilationCompensation ? n : 1;

      this.bp[i] = new Array<number>(maxPeriod).fill(0);
    }
  }

  /** Advances the estimator by one input sample and evaluates the spectrum.
   * Callers are responsible for gating on priming; update is safe to call from
   * the first bar (the BP history just carries zeros until the pre-filters settle). */
  public update(sample: number): void {
    // Shift close history.
    this.close2 = this.close1;
    this.close1 = this.close0;
    this.close0 = sample;

    // Shift HP history and compute new HP.
    this.hp2 = this.hp1;
    this.hp1 = this.hp0;
    this.hp0 = this.coeffHP0 * (this.close0 - 2 * this.close1 + this.close2)
      + this.coeffHP1 * this.hp1
      - this.coeffHP2 * this.hp2;

    // Shift Filt history and compute new Filt (SuperSmoother on HP).
    this.filt2 = this.filt1;
    this.filt1 = this.filt0;
    this.filt0 = this.ssC1 * (this.hp0 + this.hp1) / 2 + this.ssC2 * this.filt1 + this.ssC3 * this.filt2;

    const diffFilt = this.filt0 - this.filt2;

    // AGC seeds the running max with the decayed previous max; floating max
    // starts at -inf.
    this.spectrumMin = Number.MAX_VALUE;
    if (this.isAutomaticGainControl) {
      this.spectrumMax = this.automaticGainControlDecayFactor * this.previousSpectrumMax;
    } else {
      this.spectrumMax = -Number.MAX_VALUE;
    }

    const maxPeriod = this.maxPeriod;

    for (let i = 0; i < this.lengthSpectrum; i++) {
      const bpRow = this.bp[i];

      // Rightward shift.
      for (let m = maxPeriod - 1; m >= 1; m--) {
        bpRow[m] = bpRow[m - 1];
      }

      const a = this.alpha[i];
      const b = this.beta[i];
      bpRow[0] = 0.5 * (1 - a) * diffFilt + b * (1 + a) * bpRow[1] - a * bpRow[2];

      // Pwr[i] = Σ over m in [0..N) of (BP[i,m] / Comp[i])^2.
      const n = this.periods[i];
      const c = this.comp[i];
      let pwr = 0;

      for (let m = 0; m < n; m++) {
        const v = bpRow[m] / c;
        pwr += v * v;
      }

      this.spectrum[i] = pwr;

      if (this.spectrumMax < pwr) this.spectrumMax = pwr;
      if (this.spectrumMin > pwr) this.spectrumMin = pwr;
    }

    this.previousSpectrumMax = this.spectrumMax;
  }
}
