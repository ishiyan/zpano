import { } from 'jasmine';

import { SuperSmoother } from './super-smoother';
import { SuperSmootherParams } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { SuperSmootherOutput } from './output';
import { input, expected } from './testdata';

// Test data: first 500 rows from test_3-3_Supersmoother.csv
// (Julia reference implementation, 10-period super smoother).

describe('SuperSmoother', () => {
  const skipRows = 60;
  const tolerance = 2.5; // MBST seeds filter to first sample, Julia to zero; convergence takes many samples.

  it('should return expected mnemonic', () => {
    const ss = new SuperSmoother({ shortestCyclePeriod: 10 });
    expect(ss.metadata().mnemonic).toBe('ss(10, hl/2)');
  });

  it('should throw if shortest cycle period is less than 2', () => {
    expect(() => { new SuperSmoother({ shortestCyclePeriod: 1 }); }).toThrow();
  });

  it('should calculate expected output with CSV reference data', () => {
    const ss = new SuperSmoother({ shortestCyclePeriod: 10 });

    for (let i = 0; i < input.length; i++) {
      const act = ss.update(input[i]);

      if (i < 2) {
        expect(act).toBeNaN();
        expect(ss.isPrimed()).toBe(false);
        continue;
      }

      expect(ss.isPrimed()).toBe(true);

      // Skip early rows where MBST and Julia priming differ.
      if (i < skipRows) {
        continue;
      }

      expect(Math.abs(act - expected[i])).toBeLessThan(tolerance);
    }

    expect(ss.update(Number.NaN)).toBeNaN();
  });

  it('should report correct primed state', () => {
    const ss = new SuperSmoother({ shortestCyclePeriod: 10 });

    expect(ss.isPrimed()).toBe(false);

    ss.update(100);
    expect(ss.isPrimed()).toBe(false);

    ss.update(100);
    expect(ss.isPrimed()).toBe(false);

    ss.update(100);
    expect(ss.isPrimed()).toBe(true);
  });

  it('should return correct metadata', () => {
    const ss = new SuperSmoother({ shortestCyclePeriod: 10 });
    const meta = ss.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.SuperSmoother);
    expect(meta.mnemonic).toBe('ss(10, hl/2)');
    expect(meta.description).toBe('Super Smoother ss(10, hl/2)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(SuperSmootherOutput.SuperSmootherValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });
});
