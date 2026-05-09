import { } from 'jasmine';

import { ParabolicStopAndReverse } from './parabolic-stop-and-reverse';
import { ParabolicStopAndReverseOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import {
    testHighs,
    testLows,
    testExpected,
    wilderHighs,
    wilderLows,
} from './testdata';

// High test data, 252 entries. Standard TA-Lib test dataset.
// Low test data, 252 entries.
// Expected SAREXT output for 252-bar dataset with default parameters.
// Positive = long position, negative = short position.
// Index 0 = NaN (lookback = 1), indices 1-251 = valid.
// Wilder's original SAR test data (38 bars).
describe('ParabolicStopAndReverse', () => {

  it('should calculate SAREXT with default params for 252-bar dataset', () => {
    const tol = 1e-6;
    const sar = new ParabolicStopAndReverse();

    for (let i = 0; i < testHighs.length; i++) {
      const result = sar.updateHL(testHighs[i], testLows[i]);

      if (Number.isNaN(testExpected[i])) {
        expect(result).toBeNaN();
        continue;
      }

      expect(Math.abs(result - testExpected[i])).toBeLessThan(tol);
    }
  });

  it('should match Wilder spot checks', () => {
    const tol = 1e-3;
    const sar = new ParabolicStopAndReverse();

    const results: number[] = [];
    for (let i = 0; i < wilderHighs.length; i++) {
      results.push(sar.updateHL(wilderHighs[i], wilderLows[i]));
    }

    // Wilder spot checks (TA_SAR, absolute values). Output[0] = results[1].
    const spotChecks = [
      { outIndex: 0, expected: 50.00 },
      { outIndex: 1, expected: 50.047 },
      { outIndex: 4, expected: 50.182 },
      { outIndex: 35, expected: 52.93 },
      { outIndex: 36, expected: 50.00 },
    ];

    for (const sc of spotChecks) {
      const actual = Math.abs(results[sc.outIndex + 1]); // +1 because results[0] = NaN
      expect(Math.abs(actual - sc.expected)).toBeLessThan(tol);
    }
  });

  it('should be primed after 2 bars', () => {
    const sar = new ParabolicStopAndReverse();

    expect(sar.isPrimed()).toBe(false);
    sar.updateHL(93.25, 90.75);
    expect(sar.isPrimed()).toBe(false);
    sar.updateHL(94.94, 91.405);
    expect(sar.isPrimed()).toBe(true);
  });

  it('should return correct metadata', () => {
    const sar = new ParabolicStopAndReverse();
    const meta = sar.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.ParabolicStopAndReverse);
    expect(meta.mnemonic).toBe('sar()');
    expect(meta.description).toBe('Parabolic Stop And Reverse sar()');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(ParabolicStopAndReverseOutput.ParabolicStopAndReverseValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should pass NaN through', () => {
    const sar = new ParabolicStopAndReverse();
    expect(sar.update(Number.NaN)).toBeNaN();
    expect(sar.updateHL(Number.NaN, 1)).toBeNaN();
    expect(sar.updateHL(1, Number.NaN)).toBeNaN();
  });

  it('should handle NaN without corrupting state', () => {
    const sar = new ParabolicStopAndReverse();

    sar.updateHL(93.25, 90.75);
    sar.updateHL(94.94, 91.405);

    // Feed NaN.
    const nanResult = sar.updateHL(Number.NaN, 92.0);
    expect(nanResult).toBeNaN();

    // Feed valid data — should still work.
    const validResult = sar.updateHL(96.375, 94.25);
    expect(validResult).not.toBeNaN();
  });

  it('should force long start with positive startValue', () => {
    const sar = new ParabolicStopAndReverse({ startValue: 85.0 });

    const r1 = sar.updateHL(testHighs[0], testLows[0]);
    expect(r1).toBeNaN();

    const r2 = sar.updateHL(testHighs[1], testLows[1]);
    expect(r2).toBeGreaterThan(0);
  });

  it('should force short start with negative startValue', () => {
    const sar = new ParabolicStopAndReverse({ startValue: -100.0 });

    const r1 = sar.updateHL(testHighs[0], testLows[0]);
    expect(r1).toBeNaN();

    const r2 = sar.updateHL(testHighs[1], testLows[1]);
    expect(r2).toBeLessThan(0);
  });

  it('should throw for negative long acceleration', () => {
    expect(() => new ParabolicStopAndReverse({ accelerationInitLong: -0.01 })).toThrow();
  });

  it('should throw for negative short acceleration', () => {
    expect(() => new ParabolicStopAndReverse({ accelerationShort: -0.01 })).toThrow();
  });

  it('should throw for negative offset on reverse', () => {
    expect(() => new ParabolicStopAndReverse({ offsetOnReverse: -0.01 })).toThrow();
  });
});
