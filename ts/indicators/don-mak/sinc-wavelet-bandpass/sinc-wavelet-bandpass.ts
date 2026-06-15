import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { SincWaveletBandpassParams, Band } from './params';

const VELOCITY_TAPS = 4;

// Cubic velocity kernel (PFD degree=3, order=1, smoothing=0).
const VELOCITY_KERNEL = [11 / 6, -3, 3 / 2, -1 / 3];

/** Returns [omega0, omega1, numTaps] for a band, or undefined if unknown. */
function bandParams(band: Band): [number, number, number] | undefined {
  switch (band) {
    case Band.High:
      return [Math.PI / 4, Math.PI / 8, 121];
    case Band.Mid:
      return [Math.PI / 8, Math.PI / 16, 121];
    case Band.Low:
      return [Math.PI / 16, Math.PI / 32, 201];
    case Band.Full:
      return [Math.PI / 4, Math.PI / 32, 201];
    default:
      return undefined;
  }
}

/** Computes sinc band-pass filter coefficients (difference of two sinc functions). */
function computeCoefficients(omega0: number, omega1: number, numTaps: number): number[] {
  const coeffs = new Array<number>(numTaps).fill(0);
  coeffs[0] = (omega0 - omega1) / Math.PI;
  for (let k = 1; k < numTaps; k++) {
    const piK = Math.PI * k;
    coeffs[k] = Math.sin(omega0 * k) / piK - Math.sin(omega1 * k) / piK;
  }
  return coeffs;
}

const BAND_NAMES: Record<Band, string> = {
  [Band.High]: 'high',
  [Band.Mid]: 'mid',
  [Band.Low]: 'low',
  [Band.Full]: 'full',
};

/**
 * SincWaveletBandpass is Don Mak's Sinc Wavelet Band-Pass (SWB) filter.
 *
 * It is a causal FIR band-pass filter derived from the sinc wavelet system,
 * decomposing price into frequency bands (HIGH, MID, LOW, FULL). Optionally a
 * cubic velocity kernel is applied to produce a momentum oscillator.
 *
 * Reference:
 *
 * Mak, D.K. (2003). The Science of Financial Market Trading. Ch 9, Appendix 7.
 */
export class SincWaveletBandpass extends LineIndicator {

  private readonly velocity: boolean;
  private readonly coefficients: number[];
  private readonly numTaps: number;

  private readonly priceBuffer: number[];
  private priceCount = 0;
  private priceIndex = 0;

  private readonly velBuffer: number[];
  private velCount = 0;
  private velIndex = 0;

  public constructor(params: SincWaveletBandpassParams) {
    super();

    const band = params.band ?? Band.Mid;
    const velocity = params.velocity ?? false;

    const bp = bandParams(band);
    if (bp === undefined) {
      throw new Error('unknown band');
    }
    const [omega0, omega1, numTaps] = bp;

    const cfg = BAND_NAMES[band] + (velocity ? ',v' : '');
    this.mnemonic = `swb(${cfg}${componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent)})`;
    this.description = 'Sinc wavelet band-pass ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.velocity = velocity;
    this.numTaps = numTaps;
    this.coefficients = computeCoefficients(omega0, omega1, numTaps);
    this.priceBuffer = new Array<number>(numTaps).fill(0);
    this.velBuffer = new Array<number>(VELOCITY_TAPS).fill(0);
    this.primed = false;
  }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.SincWaveletBandpass,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  public update(sample: number): number {
    // Store price in the ring buffer.
    this.priceBuffer[this.priceIndex] = sample;
    this.priceIndex = (this.priceIndex + 1) % this.numTaps;
    this.priceCount++;

    if (this.priceCount < this.numTaps) {
      this.primed = false;
      return Number.NaN;
    }

    // Band-pass convolution: coefficients[k] multiplies the k-th most recent price.
    let bpValue = 0;
    let idx = this.priceIndex - 1;
    for (let k = 0; k < this.numTaps; k++) {
      const bufIdx = ((idx % this.numTaps) + this.numTaps) % this.numTaps;
      bpValue += this.coefficients[k] * this.priceBuffer[bufIdx];
      idx--;
    }

    if (!this.velocity) {
      this.primed = true;
      return bpValue;
    }

    // Store band-pass output in the velocity ring buffer.
    this.velBuffer[this.velIndex] = bpValue;
    this.velIndex = (this.velIndex + 1) % VELOCITY_TAPS;
    this.velCount++;

    if (this.velCount < VELOCITY_TAPS) {
      this.primed = false;
      return Number.NaN;
    }

    // Cubic velocity: kernel[k] multiplies the k-th most recent band-pass value.
    let velValue = 0;
    idx = this.velIndex - 1;
    for (let k = 0; k < VELOCITY_TAPS; k++) {
      const bufIdx = ((idx % VELOCITY_TAPS) + VELOCITY_TAPS) % VELOCITY_TAPS;
      velValue += VELOCITY_KERNEL[k] * this.velBuffer[bufIdx];
      idx--;
    }

    this.primed = true;
    return velValue;
  }
}
