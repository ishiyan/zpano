import { } from 'jasmine';

import { RoofingFilter } from './roofing-filter';
import { RoofingFilterParams } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { RoofingFilterOutput } from './output';
import {
    input,
    expected71,
    expected72,
    expected73,
} from './testdata';

// Test data from CSV reference files.

// Expected: test 7-1, 1-pole HP, no zero-mean (shortest=10, longest=48).
// Expected: test 7-2, 1-pole HP, zero-mean (shortest=10, longest=48), Filt2 column.
// Expected: test 7-3, 2-pole HP (shortest=40, longest=80), Filt column.
describe('RoofingFilter', () => {
  const skipRows = 30;
  const tolerance = 0.5; // MBST priming differs from Julia zero-init; convergence takes many samples.

  it('should throw if shortest cycle period is less than 2', () => {
    expect(() => { new RoofingFilter({ shortestCyclePeriod: 1, longestCyclePeriod: 48 }); }).toThrow();
  });

  it('should throw if longest cycle period is not greater than shortest', () => {
    expect(() => { new RoofingFilter({ shortestCyclePeriod: 10, longestCyclePeriod: 10 }); }).toThrow();
  });

  it('should calculate expected output for 1-pole HP (test 7-1)', () => {
    const rf = new RoofingFilter({ shortestCyclePeriod: 10, longestCyclePeriod: 48 });

    for (let i = 0; i < input.length; i++) {
      const act = rf.update(input[i]);

      if (i < 3) {
        expect(act).toBeNaN();
        expect(rf.isPrimed()).toBe(false);
        continue;
      }

      expect(rf.isPrimed()).toBe(true);

      if (i < skipRows) {
        continue;
      }

      expect(Math.abs(act - expected71[i])).toBeLessThan(tolerance);
    }

    expect(rf.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output for 1-pole HP zero-mean (test 7-2)', () => {
    const rf = new RoofingFilter({ shortestCyclePeriod: 10, longestCyclePeriod: 48, hasZeroMean: true });

    for (let i = 0; i < input.length; i++) {
      const act = rf.update(input[i]);

      if (i < 4) {
        expect(act).toBeNaN();
        expect(rf.isPrimed()).toBe(false);
        continue;
      }

      expect(rf.isPrimed()).toBe(true);

      if (i < skipRows) {
        continue;
      }

      expect(Math.abs(act - expected72[i])).toBeLessThan(tolerance);
    }
  });

  it('should calculate expected output for 2-pole HP (test 7-3)', () => {
    const rf = new RoofingFilter({
      shortestCyclePeriod: 40, longestCyclePeriod: 80, hasTwoPoleHighpassFilter: true,
    });

    for (let i = 0; i < input.length; i++) {
      const act = rf.update(input[i]);

      if (i < 4) {
        expect(act).toBeNaN();
        expect(rf.isPrimed()).toBe(false);
        continue;
      }

      expect(rf.isPrimed()).toBe(true);

      if (i < skipRows) {
        continue;
      }

      expect(Math.abs(act - expected73[i])).toBeLessThan(tolerance);
    }
  });

  it('should report correct primed state for 1-pole', () => {
    const rf = new RoofingFilter({ shortestCyclePeriod: 10, longestCyclePeriod: 48 });

    expect(rf.isPrimed()).toBe(false);

    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(true);
  });

  it('should report correct primed state for 1-pole zero-mean', () => {
    const rf = new RoofingFilter({ shortestCyclePeriod: 10, longestCyclePeriod: 48, hasZeroMean: true });

    expect(rf.isPrimed()).toBe(false);

    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(true);
  });

  it('should report correct primed state for 2-pole', () => {
    const rf = new RoofingFilter({
      shortestCyclePeriod: 10, longestCyclePeriod: 48, hasTwoPoleHighpassFilter: true,
    });

    expect(rf.isPrimed()).toBe(false);

    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(false);
    rf.update(100); expect(rf.isPrimed()).toBe(true);
  });

  it('should return correct metadata for 1-pole', () => {
    const rf = new RoofingFilter({ shortestCyclePeriod: 10, longestCyclePeriod: 48 });
    const meta = rf.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.RoofingFilter);
    expect(meta.mnemonic).toBe('roof1hp(10, 48, hl/2)');
    expect(meta.description).toBe('Roofing Filter roof1hp(10, 48, hl/2)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(RoofingFilterOutput.RoofingFilterValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should return correct mnemonic for 2-pole', () => {
    const rf = new RoofingFilter({
      shortestCyclePeriod: 10, longestCyclePeriod: 48, hasTwoPoleHighpassFilter: true,
    });
    expect(rf.metadata().mnemonic).toBe('roof2hp(10, 48, hl/2)');
  });

  it('should return correct mnemonic for zero-mean', () => {
    const rf = new RoofingFilter({
      shortestCyclePeriod: 10, longestCyclePeriod: 48, hasZeroMean: true,
    });
    expect(rf.metadata().mnemonic).toBe('roof1hpzm(10, 48, hl/2)');
  });
});
