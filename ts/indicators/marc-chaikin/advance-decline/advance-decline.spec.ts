import { } from 'jasmine';

import { AdvanceDecline } from './advance-decline';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { AdvanceDeclineOutput } from './output';
import {
    testHighs,
    testLows,
    testCloses,
    testVolumes,
    testExpectedAD,
} from './testdata';

// High test data, 252 entries. From TA-Lib excel-sma3-sma10-chaikin.csv.
// Low test data, 252 entries.
// Close test data, 252 entries.
// Volume test data, 252 entries.
// Expected AD output, 252 entries. All valid (lookback = 0).
// From TA-Lib excel-sma3-sma10-chaikin.csv, column F (AD).
function roundTo(v: number, digits: number): number {
  const p = Math.pow(10, digits);
  return Math.round(v * p) / p;
}

describe('AdvanceDecline', () => {

  it('should calculate AD with HLCV correctly', () => {
    const digits = 2;
    const ad = new AdvanceDecline();

    for (let i = 0; i < testHighs.length; i++) {
      const v = ad.updateHLCV(testHighs[i], testLows[i], testCloses[i], testVolumes[i]);
      expect(v).not.toBeNaN();
      expect(ad.isPrimed()).toBe(true);
      expect(roundTo(v, digits)).toBe(roundTo(testExpectedAD[i], digits));
    }
  });

  it('should match TA-Lib spot checks', () => {
    const digits = 2;
    const ad = new AdvanceDecline();

    const values: number[] = [];
    for (let i = 0; i < testHighs.length; i++) {
      values.push(ad.updateHLCV(testHighs[i], testLows[i], testCloses[i], testVolumes[i]));
    }

    // TA-Lib spot checks from test_per_hlcv.c.
    expect(roundTo(values[0], digits)).toBe(roundTo(-1631000.00, digits));
    expect(roundTo(values[1], digits)).toBe(roundTo(2974412.02, digits));
    expect(roundTo(values[250], digits)).toBe(roundTo(8707691.07, digits));
    expect(roundTo(values[251], digits)).toBe(roundTo(8328944.54, digits));
  });

  it('should not be primed initially', () => {
    const ad = new AdvanceDecline();
    expect(ad.isPrimed()).toBe(false);
  });

  it('should become primed after first valid update', () => {
    const ad = new AdvanceDecline();
    ad.updateHLCV(100, 90, 95, 1000);
    expect(ad.isPrimed()).toBe(true);
  });

  it('should not become primed after NaN update', () => {
    const ad = new AdvanceDecline();
    ad.update(Number.NaN);
    expect(ad.isPrimed()).toBe(false);
  });

  it('should pass NaN through', () => {
    const ad = new AdvanceDecline();
    expect(ad.update(Number.NaN)).toBeNaN();
    expect(ad.updateHLCV(Number.NaN, 1, 2, 3)).toBeNaN();
    expect(ad.updateHLCV(1, Number.NaN, 2, 3)).toBeNaN();
    expect(ad.updateHLCV(1, 2, Number.NaN, 3)).toBeNaN();
    expect(ad.updateHLCV(1, 2, 3, Number.NaN)).toBeNaN();
  });

  it('should return 0 for scalar update (H=L=C)', () => {
    const ad = new AdvanceDecline();
    const v = ad.update(100.0);
    expect(v).toBe(0);
    expect(ad.isPrimed()).toBe(true);
  });

  it('should return correct metadata', () => {
    const ad = new AdvanceDecline();
    const meta = ad.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.AdvanceDecline);
    expect(meta.mnemonic).toBe('ad');
    expect(meta.description).toBe('Advance-Decline');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(AdvanceDeclineOutput.AdvanceDeclineValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should not change AD when high equals low', () => {
    const ad = new AdvanceDecline();
    ad.updateHLCV(100, 90, 95, 1000);
    const v1 = ad.updateHLCV(100, 100, 100, 5000);
    const v2 = ad.updateHLCV(100, 90, 95, 1000);
    // v1 should equal the previous AD (unchanged when H=L)
    expect(v1).toBe(ad.updateHLCV(100, 100, 100, 0) - 0 + v1); // just check it didn't change
  });
});
