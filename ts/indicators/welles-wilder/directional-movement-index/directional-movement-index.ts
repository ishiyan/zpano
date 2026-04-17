import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorType } from '../../core/indicator-type';
import { OutputMetadata } from '../../core/outputs/output-metadata';
import { OutputType } from '../../core/outputs/output-type';
import { DirectionalIndicatorMinus } from '../directional-indicator-minus/directional-indicator-minus';
import { DirectionalIndicatorPlus } from '../directional-indicator-plus/directional-indicator-plus';
import { DirectionalMovementIndexOutput } from './directional-movement-index-output';

const dxMnemonic = 'dx';
const dxDescription = 'Directional Movement Index';
const epsilon = 1e-8;

/**
 * Welles Wilder's Directional Movement Index (DX).
 *
 * The directional movement index measures the strength of a trend by comparing
 * the positive and negative directional indicators. It is calculated as:
 *
 *   DX = 100 * |+DI - -DI| / (+DI + -DI)
 *
 * where +DI is the directional indicator plus and -DI is the directional
 * indicator minus, both computed over the same length.
 *
 * The indicator requires close, high, and low values.
 */
export class DirectionalMovementIndex implements Indicator {

  private readonly length_: number;
  private value_ = NaN;
  private readonly directionalIndicatorPlus: DirectionalIndicatorPlus;
  private readonly directionalIndicatorMinus: DirectionalIndicatorMinus;

  constructor(length: number) {
    if (length < 1) {
      throw new Error(`invalid length ${length}: must be >= 1`);
    }

    this.length_ = length;
    this.directionalIndicatorPlus = new DirectionalIndicatorPlus(length);
    this.directionalIndicatorMinus = new DirectionalIndicatorMinus(length);
  }

  /** Returns the length parameter. */
  public get length(): number {
    return this.length_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.directionalIndicatorPlus.isPrimed() && this.directionalIndicatorMinus.isPrimed();
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    const outputs: OutputMetadata[] = [
      {
        kind: DirectionalMovementIndexOutput.DirectionalMovementIndexValue,
        type: OutputType.Scalar,
        mnemonic: dxMnemonic,
        description: dxDescription,
      },
      {
        kind: DirectionalMovementIndexOutput.DirectionalIndicatorPlusValue,
        type: OutputType.Scalar,
        mnemonic: '+di',
        description: 'Directional Indicator Plus',
      },
      {
        kind: DirectionalMovementIndexOutput.DirectionalIndicatorMinusValue,
        type: OutputType.Scalar,
        mnemonic: '-di',
        description: 'Directional Indicator Minus',
      },
      {
        kind: DirectionalMovementIndexOutput.DirectionalMovementPlusValue,
        type: OutputType.Scalar,
        mnemonic: '+dm',
        description: 'Directional Movement Plus',
      },
      {
        kind: DirectionalMovementIndexOutput.DirectionalMovementMinusValue,
        type: OutputType.Scalar,
        mnemonic: '-dm',
        description: 'Directional Movement Minus',
      },
      {
        kind: DirectionalMovementIndexOutput.AverageTrueRangeValue,
        type: OutputType.Scalar,
        mnemonic: 'atr',
        description: 'Average True Range',
      },
      {
        kind: DirectionalMovementIndexOutput.TrueRangeValue,
        type: OutputType.Scalar,
        mnemonic: 'tr',
        description: 'True Range',
      },
    ];

    return {
      type: IndicatorType.DirectionalMovementIndex,
      mnemonic: dxMnemonic,
      description: dxDescription,
      outputs,
    };
  }

  /** Updates the Directional Movement Index given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    const dipValue = this.directionalIndicatorPlus.update(close, high, low);
    const dimValue = this.directionalIndicatorMinus.update(close, high, low);

    if (this.directionalIndicatorPlus.isPrimed() && this.directionalIndicatorMinus.isPrimed()) {
      const sum = dipValue + dimValue;

      if (Math.abs(sum) < epsilon) {
        this.value_ = 0;
      } else {
        this.value_ = 100 * Math.abs(dipValue - dimValue) / sum;
      }

      return this.value_;
    }

    return NaN;
  }

  /** Updates the Directional Movement Index using a single sample value as a substitute for close, high, and low. */
  public updateSample(sample: number): number {
    return this.update(sample, sample, sample);
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(sample.close, sample.high, sample.low);
    return [scalar];
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bid + sample.ask) / 2;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = sample.price;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }
}
