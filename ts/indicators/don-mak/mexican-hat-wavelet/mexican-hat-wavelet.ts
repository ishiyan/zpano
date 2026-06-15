import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { MexicanHatWaveletParams, Band } from './params';

// Preset dilation values (a_f) for the three standard bands (Table 5.2).
const DILATION_HIGH = 1.483;
const DILATION_MID = 4.048;
const DILATION_LOW = 15.97;

/** Rounds half to even (banker's rounding), matching Python's round(). */
function roundHalfEven(x: number): number {
  if (Math.abs(x - Math.trunc(x)) === 0.5) {
    const f = Math.floor(x);
    return f % 2 === 0 ? f : f + 1;
  }
  return Math.round(x);
}

/** Computes dilation a_f from a desired center period in bars (Eq 5.11). */
function dilationFromPeriod(period: number): number {
  const omega0 = (2 * Math.PI) / period;
  const twoOverA = 1.091 * omega0 - 0.071 * omega0 * omega0;
  if (twoOverA <= 0) {
    throw new Error('period is too large for the fitting formula (2/a <= 0)');
  }
  return 2 / twoOverA;
}

/**
 * Computes normalized Mexican Hat wavelet FIR coefficients for dilation a_f.
 *
 * psi(t) = (1 - 2*t^2) * exp(-t^2); h(n) = psi(n / a_f) for n = 0..K,
 * K = 4 * round(a_f), normalized by 0.488 + 0.646*a_f + 0.0001*a_f^2.
 */
function computeCoefficients(aF: number): number[] {
  let k = 4 * roundHalfEven(aF);
  if (k < 1) {
    k = 1;
  }

  const norm = 0.488 + 0.646 * aF + 0.0001 * aF * aF;

  const coeffs: number[] = [];
  for (let n = 0; n <= k; n++) {
    const t = n / aF;
    const t2 = t * t;
    const hN = (1 - 2 * t2) * Math.exp(-t2);
    coeffs.push(hN / norm);
  }

  return coeffs;
}

/**
 * MexicanHatWavelet is Don Mak's Mexican Hat Wavelet (MHW) bandpass filter.
 *
 * It is a causal bandpass FIR filter derived from the Mexican Hat wavelet (the
 * second derivative of a Gaussian), decomposing price into frequency bands with
 * zero phase shift at the filter's center frequency.
 *
 * Reference:
 *
 * Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 5.
 */
export class MexicanHatWavelet extends LineIndicator {

  private readonly coefficients: number[];
  private readonly numTaps: number;

  private readonly buffer: number[];
  private count = 0;

  public constructor(params: MexicanHatWaveletParams) {
    super();

    const band = params.band ?? Band.Mid;
    const dilation = params.dilation;
    const period = params.period;

    let aF: number;
    let cfg: string;

    switch (band) {
      case Band.High:
        aF = DILATION_HIGH;
        cfg = 'high';
        break;
      case Band.Mid:
        aF = DILATION_MID;
        cfg = 'mid';
        break;
      case Band.Low:
        aF = DILATION_LOW;
        cfg = 'low';
        break;
      case Band.Custom: {
        const hasDilation = dilation !== undefined && dilation !== 0;
        const hasPeriod = period !== undefined && period !== 0;
        if (hasDilation && hasPeriod) {
          throw new Error('provide only one of dilation or period, not both');
        }
        if (!hasDilation && !hasPeriod) {
          throw new Error('band=custom requires either dilation or period');
        }
        if (hasPeriod) {
          if ((period as number) <= 2) {
            throw new Error('period must be > 2');
          }
          aF = dilationFromPeriod(period as number);
          cfg = `p${(period as number).toFixed(2)}`;
        } else {
          if ((dilation as number) <= 0) {
            throw new Error('dilation must be > 0');
          }
          aF = dilation as number;
          cfg = `d${(dilation as number).toFixed(2)}`;
        }
        break;
      }
      default:
        throw new Error('unknown band');
    }

    this.mnemonic = `mhw(${cfg}${componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent)})`;
    this.description = 'Mexican hat wavelet ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.coefficients = computeCoefficients(aF);
    this.numTaps = this.coefficients.length;
    this.buffer = new Array<number>(this.numTaps).fill(0);
    this.primed = false;
  }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.MexicanHatWavelet,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  public update(sample: number): number {
    // Shift buffer right and insert the new price at position 0.
    for (let i = this.numTaps - 1; i > 0; i--) {
      this.buffer[i] = this.buffer[i - 1];
    }
    this.buffer[0] = sample;
    this.count++;

    if (this.count < this.numTaps) {
      this.primed = false;
      return Number.NaN;
    }

    // FIR convolution: y = sum(coefficients[k] * buffer[k]).
    let y = 0;
    for (let k = 0; k < this.numTaps; k++) {
      y += this.coefficients[k] * this.buffer[k];
    }

    this.primed = true;
    return y;
  }
}
