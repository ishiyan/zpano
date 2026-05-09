import { } from 'jasmine';

import { MoneyFlowIndex } from './money-flow-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { MoneyFlowIndexOutput } from './output';
import {
    typicalPrices,
    volumes,
    expectedMfi,
    expectedMfiVolume1,
} from './testdata';

// Typical price test data: (high + low + close) / 3, 252 entries.
// Same typical prices as CCI test data, from TA-Lib test_MF.xls.
// Volume test data, 252 entries. From TA-Lib test_MF.xls, column E.
// Expected MFI(14) output with real volume, 252 entries. 14 NaN then 238 values.
// From TA-Lib test_MF.xls, column K.
// Expected MFI(14) output with volume=1, 252 entries. 14 NaN then 238 values.
// From TA-Lib test_MF.xls, column S.
function roundTo(v: number, digits: number): number {
  const p = Math.pow(10, digits);
  return Math.round(v * p) / p;
}

describe('MoneyFlowIndex', () => {

  it('should throw if length is less than 1', () => {
    expect(() => { new MoneyFlowIndex({ length: 0 }); }).toThrow();
  });

  it('should throw if length is negative', () => {
    expect(() => { new MoneyFlowIndex({ length: -8 }); }).toThrow();
  });

  it('should calculate MFI(14) with real volume correctly', () => {
    const digits = 9;
    const mfi = new MoneyFlowIndex({ length: 14 });

    for (let i = 0; i < 14; i++) {
      const v = mfi.updateWithVolume(typicalPrices[i], volumes[i]);
      expect(v).toBeNaN();
      expect(mfi.isPrimed()).toBe(false);
    }

    for (let i = 14; i < typicalPrices.length; i++) {
      const v = mfi.updateWithVolume(typicalPrices[i], volumes[i]);
      expect(v).not.toBeNaN();
      expect(mfi.isPrimed()).toBe(true);
      expect(roundTo(v, digits)).toBe(roundTo(expectedMfi[i], digits));
    }
  });

  it('should calculate MFI(14) with volume=1 correctly', () => {
    const digits = 9;
    const mfi = new MoneyFlowIndex({ length: 14 });

    for (let i = 0; i < 14; i++) {
      const v = mfi.update(typicalPrices[i]);
      expect(v).toBeNaN();
    }

    for (let i = 14; i < typicalPrices.length; i++) {
      const v = mfi.update(typicalPrices[i]);
      expect(v).not.toBeNaN();
      expect(roundTo(v, digits)).toBe(roundTo(expectedMfiVolume1[i], digits));
    }
  });

  it('should report correct primed state', () => {
    const mfi = new MoneyFlowIndex({ length: 5 });
    expect(mfi.isPrimed()).toBe(false);

    for (let i = 1; i <= 5; i++) {
      mfi.update(i);
      expect(mfi.isPrimed()).toBe(false);
    }

    mfi.update(5);
    expect(mfi.isPrimed()).toBe(true);

    mfi.update(6);
    expect(mfi.isPrimed()).toBe(true);
  });

  it('should pass NaN through', () => {
    const mfi = new MoneyFlowIndex({ length: 5 });
    expect(mfi.update(Number.NaN)).toBeNaN();
    expect(mfi.updateWithVolume(1.0, Number.NaN)).toBeNaN();
    expect(mfi.updateWithVolume(Number.NaN, Number.NaN)).toBeNaN();
  });

  it('should return correct metadata', () => {
    const mfi = new MoneyFlowIndex({ length: 14 });
    const meta = mfi.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.MoneyFlowIndex);
    expect(meta.mnemonic).toBe('mfi(14, hlc/3)');
    expect(meta.description).toBe('Money Flow Index mfi(14, hlc/3)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(MoneyFlowIndexOutput.MoneyFlowIndexValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should return 0 when sum is less than 1', () => {
    const mfi = new MoneyFlowIndex({ length: 2 });

    for (let i = 0; i < 10; i++) {
      mfi.updateWithVolume(0.001, 0.5);
    }

    expect(mfi.isPrimed()).toBe(true);

    const v = mfi.updateWithVolume(0.001, 0.5);
    expect(v).toBe(0);
  });
});
