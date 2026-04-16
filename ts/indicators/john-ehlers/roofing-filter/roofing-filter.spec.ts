import { } from 'jasmine';

import { RoofingFilter } from './roofing-filter';
import { RoofingFilterParams } from './roofing-filter-params';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { RoofingFilterOutput } from './roofing-filter-output';

// Test data from CSV reference files.

const input = [
  1065.25, 1065.25, 1063.75, 1059.25, 1059.25, 1057.75, 1054, 1056.25, 1058.5, 1059.5,
  1064.75, 1063, 1062.5, 1065, 1061.5, 1058.25, 1058.25, 1061.75, 1062, 1061.25,
  1062.5, 1066.5, 1066.5, 1069.25, 1074.75, 1075, 1076, 1078, 1079.25, 1079.75,
  1078, 1078.75, 1078.25, 1076.5, 1075.75, 1075.75, 1075, 1073.25, 1071, 1083,
  1082.25, 1084, 1085.75, 1085.25, 1085.75, 1087.25, 1089, 1089, 1090, 1095,
  1097.25, 1097.25, 1099, 1098.25, 1093.75, 1095, 1097.25, 1099.25, 1097.5, 1096,
  1095, 1094, 1095.75, 1095.75, 1093.75, 1100.5, 1102.25, 1102, 1102.75, 1105.75,
  1108.25, 1109.5, 1107.25, 1102.5, 1104.75, 1099.25, 1102.75, 1099.5, 1096.75, 1098.25,
  1095.25, 1097, 1097.75, 1100.5, 1099.5, 1101.75, 1101.75, 1102.75, 1099.75, 1097,
  1100.75, 1105.75, 1104.5, 1108.5, 1111.25, 1112.25, 1110, 1109.75, 1108.25, 1106,
];

// Expected: test 7-1, 1-pole HP, no zero-mean (shortest=10, longest=48).
const expected71 = [
  0, 0, 0, -0.53, -1.62, -2.72, -4.03, -5.09, -5.05, -4.09,
  -2.20, -0.05, 1.29, 2.14, 2.39, 1.46, -0.05, -0.90, -0.80, -0.41,
  0.03, 0.99, 2.30, 3.60, 5.39, 7.33, 8.69, 9.52, 10.00, 10.11,
  9.59, 8.58, 7.46, 6.12, 4.61, 3.26, 2.16, 1.12, -0.11, 0.12,
  2.14, 4.27, 6.08, 7.22, 7.54, 7.48, 7.46, 7.43, 7.29, 7.64,
  8.69, 9.68, 10.26, 10.32, 9.23, 7.38, 5.98, 5.47, 5.30, 4.74,
  3.77, 2.58, 1.66, 1.28, 0.92, 1.21, 2.62, 4.12, 5.14, 5.97,
  6.95, 7.94, 8.26, 7.16, 5.36, 3.27, 1.36, 0.07, -1.34, -2.48,
  -3.29, -3.79, -3.61, -2.72, -1.53, -0.40, 0.67, 1.49, 1.70, 0.89,
  0.04, 0.47, 1.66, 3.05, 4.81, 6.48, 7.28, 7.00, 5.99, 4.62,
];

// Expected: test 7-2, 1-pole HP, zero-mean (shortest=10, longest=48), Filt2 column.
const expected72 = [
  0, 0, 0, -0.50, -1.46, -2.31, -3.26, -3.85, -3.34, -2.02,
  -0.01, 2.01, 3.02, 3.45, 3.26, 1.99, 0.33, -0.52, -0.35, 0.05,
  0.46, 1.31, 2.37, 3.30, 4.57, 5.84, 6.39, 6.38, 6.05, 5.41,
  4.26, 2.79, 1.39, -0.04, -1.45, -2.54, -3.26, -3.83, -4.51, -3.74,
  -1.39, 0.78, 2.39, 3.16, 3.07, 2.64, 2.29, 1.98, 1.61, 1.74,
  2.51, 3.13, 3.29, 2.94, 1.56, -0.37, -1.64, -1.91, -1.84, -2.14,
  -2.79, -3.56, -3.98, -3.85, -3.72, -2.99, -1.29, 0.27, 1.19, 1.83,
  2.53, 3.14, 3.06, 1.65, -0.25, -2.17, -3.70, -4.46, -5.23, -5.66,
  -5.72, -5.49, -4.65, -3.23, -1.72, -0.45, 0.61, 1.30, 1.34, 0.42,
  -0.43, 0.02, 1.14, 2.30, 3.67, 4.79, 4.95, 4.08, 2.63, 0.80,
];

// Expected: test 7-3, 2-pole HP (shortest=40, longest=80), Filt column.
const expected73 = [
  0, 0, 0, -0.03, -0.10, -0.17, -0.28, -0.37, -0.38, -0.27,
  0.03, 0.52, 1.13, 1.85, 2.62, 3.37, 4.04, 4.69, 5.35, 6.00,
  6.63, 7.29, 7.99, 8.71, 9.52, 10.42, 11.34, 12.27, 13.19, 14.07,
  14.85, 15.49, 15.99, 16.32, 16.45, 16.40, 16.19, 15.82, 15.27, 14.69,
  14.20, 13.79, 13.45, 13.16, 12.90, 12.66, 12.45, 12.26, 12.07, 11.93,
  11.88, 11.88, 11.91, 11.94, 11.88, 11.69, 11.41, 11.11, 10.76, 10.33,
  9.80, 9.17, 8.47, 7.75, 6.99, 6.26, 5.64, 5.14, 4.71, 4.37,
  4.16, 4.07, 4.02, 3.93, 3.77, 3.50, 3.13, 2.68, 2.13, 1.49,
  0.79, 0.05, -0.67, -1.31, -1.86, -2.31, -2.65, -2.89, -3.06, -3.24,
  -3.40, -3.46, -3.39, -3.21, -2.88, -2.41, -1.89, -1.37, -0.91, -0.51,
];

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

    expect(meta.type).toBe(IndicatorType.RoofingFilter);
    expect(meta.mnemonic).toBe('roof1hp(10, 48, hl/2)');
    expect(meta.description).toBe('Roofing Filter roof1hp(10, 48, hl/2)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(RoofingFilterOutput.RoofingFilterValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
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
