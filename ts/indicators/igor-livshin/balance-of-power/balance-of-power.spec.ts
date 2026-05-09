import { } from 'jasmine';

import { BalanceOfPower } from './balance-of-power';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { BalanceOfPowerOutput } from './output';
import {
    inputOpen,
    inputHigh,
    inputLow,
    inputClose,
    expectedBop,
} from './testdata';

// First 20 entries from the C# TA-Lib test data.
function roundTo(v: number, digits: number): number {
  const p = Math.pow(10, digits);
  return Math.round(v * p) / p;
}

describe('BalanceOfPower', () => {

  it('should calculate BOP from OHLC correctly', () => {
    const digits = 9;
    const bop = new BalanceOfPower();

    for (let i = 0; i < inputOpen.length; i++) {
      const v = bop.updateOHLC(inputOpen[i], inputHigh[i], inputLow[i], inputClose[i]);
      expect(v).not.toBeNaN();
      expect(bop.isPrimed()).toBe(true);
      expect(roundTo(v, digits)).toBe(roundTo(expectedBop[i], digits));
    }
  });

  it('should always be primed', () => {
    const bop = new BalanceOfPower();
    expect(bop.isPrimed()).toBe(true);

    bop.updateOHLC(92.5, 93.25, 90.75, 91.5);
    expect(bop.isPrimed()).toBe(true);
  });

  it('should pass NaN through', () => {
    const bop = new BalanceOfPower();
    expect(bop.update(Number.NaN)).toBeNaN();
    expect(bop.updateOHLC(Number.NaN, 1.0, 2.0, 3.0)).toBeNaN();
    expect(bop.updateOHLC(1.0, Number.NaN, 2.0, 3.0)).toBeNaN();
    expect(bop.updateOHLC(1.0, 2.0, Number.NaN, 3.0)).toBeNaN();
    expect(bop.updateOHLC(1.0, 2.0, 3.0, Number.NaN)).toBeNaN();
  });

  it('should return correct metadata', () => {
    const bop = new BalanceOfPower();
    const meta = bop.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.BalanceOfPower);
    expect(meta.mnemonic).toBe('bop');
    expect(meta.description).toBe('Balance of Power');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(BalanceOfPowerOutput.BalanceOfPowerValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should return 0 for zero range', () => {
    const bop = new BalanceOfPower();
    expect(bop.updateOHLC(0.001, 0.001, 0.001, 0.001)).toBe(0);
  });

  it('should return 0 for scalar update', () => {
    const bop = new BalanceOfPower();
    expect(bop.update(50.0)).toBe(0);
    expect(bop.update(100.0)).toBe(0);
  });
});
