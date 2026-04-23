import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { MomentumParams } from './params';

/**
 * __Momentum__ (__MOM__) is the absolute (not normalized) difference between today's sample
 * and the sample l periods ago.
 *
 * This implementation calculates the value of the MOM using the formula:
 *
 *     MOMi = Pi - Pi-l,
 *
 * where l is the length.
 *
 * The indicator is not primed during the first l updates.
 */
export class Momentum extends LineIndicator {
  private readonly window: number[];
  private readonly windowLength: number;
  private readonly lastIndex: number;
  private windowCount = 0;

  /** Constructs an instance given the parameters. */
  public constructor(params: MomentumParams) {
    super();
    const length = Math.floor(params.length);
    if (length < 1) {
      throw new Error('length should be positive');
    }

    this.window = new Array<number>(length + 1).fill(0);
    this.windowLength = length + 1;
    this.lastIndex = length;

    const mn = 'mom(' + length.toString() +
      componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent) + ')';

    this.mnemonic = mn;
    this.description = 'Momentum ' + mn;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.Momentum,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /** Updates the value of the momentum given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    if (this.primed) {
      for (let i = 0; i < this.lastIndex; i++) {
        this.window[i] = this.window[i + 1];
      }

      this.window[this.lastIndex] = sample;

      return sample - this.window[0];
    }

    this.window[this.windowCount] = sample;
    this.windowCount++;

    if (this.windowLength === this.windowCount) {
      this.primed = true;

      return sample - this.window[0];
    }

    return Number.NaN;
  }
}
