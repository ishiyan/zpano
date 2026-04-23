import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { ChandeMomentumOscillatorParams } from './params';

const epsilon = 1e-12;

/**
 * __Chande Momentum Oscillator__ (__CMO__) is a momentum indicator based on the average
 * of up samples and down samples over a specified length ℓ.
 *
 * The calculation formula is:
 *
 *     CMOᵢ = 100 (SUᵢ - SDᵢ) / (SUᵢ + SDᵢ),
 *
 * where SUᵢ (sum up) is the sum of gains and SDᵢ (sum down)
 * is the sum of losses over the chosen length [i-ℓ, i].
 *
 * The indicator is not primed during the first ℓ updates.
 */
export class ChandeMomentumOscillator extends LineIndicator {
  private readonly length: number;
  private readonly ringBuffer: number[];
  private ringHead = 0;
  private count = 0;
  private previousSample = 0;
  private gainSum = 0;
  private lossSum = 0;

  /** Constructs an instance given the parameters. */
  public constructor(params: ChandeMomentumOscillatorParams) {
    super();
    const length = Math.floor(params.length);
    if (length < 1) {
      throw new Error('length should be positive');
    }

    this.length = length;
    this.ringBuffer = new Array<number>(length).fill(0);

    const mn = 'cmo(' + length.toString() +
      componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent) + ')';

    this.mnemonic = mn;
    this.description = 'Chande Momentum Oscillator ' + mn;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.ChandeMomentumOscillator,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /** Updates the value of the Chande momentum oscillator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    this.count++;
    if (this.count === 1) {
      this.previousSample = sample;

      return Number.NaN;
    }

    // New delta
    const delta = sample - this.previousSample;
    this.previousSample = sample;

    if (!this.primed) {
      // Fill until we have this.length deltas (i.e., this.length+1 samples)
      this.ringBuffer[this.ringHead] = delta;
      this.ringHead = (this.ringHead + 1) % this.length;

      if (delta > 0) {
        this.gainSum += delta;
      } else if (delta < 0) {
        this.lossSum += -delta;
      }

      if (this.count <= this.length) {
        return Number.NaN;
      }

      // Now we have exactly this.length deltas in the buffer
      this.primed = true;
    } else {
      // Remove oldest delta and add the new one
      const old = this.ringBuffer[this.ringHead];
      if (old > 0) {
        this.gainSum -= old;
      } else if (old < 0) {
        this.lossSum -= -old;
      }

      this.ringBuffer[this.ringHead] = delta;
      this.ringHead = (this.ringHead + 1) % this.length;

      if (delta > 0) {
        this.gainSum += delta;
      } else if (delta < 0) {
        this.lossSum += -delta;
      }

      // Clamp to avoid tiny negative sums from FP noise
      if (this.gainSum < 0) {
        this.gainSum = 0;
      }

      if (this.lossSum < 0) {
        this.lossSum = 0;
      }
    }

    const den = this.gainSum + this.lossSum;
    if (Math.abs(den) < epsilon) {
      return 0;
    }

    return 100.0 * (this.gainSum - this.lossSum) / den;
  }
}
