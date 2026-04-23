import { } from 'jasmine';

import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { VarianceOutput } from './output';
import { Variance } from './variance';

// Variance input test data.
const input = [
  1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12
];

describe('Variance', () => {
  const epsilon = 10e-2;

  it('should return expected metadata for sample variance', () => {
    const v = new Variance({length: 7, unbiased: true});
    const m = v.metadata();

    expect(m.identifier).toBe(IndicatorIdentifier.Variance);
    expect(m.mnemonic).toBe('var.s(7)');
    expect(m.description).toBe('Unbiased estimation of the sample variance var.s(7)');
    expect(m.outputs.length).toBe(1);
    expect(m.outputs[0].kind).toBe(VarianceOutput.VarianceValue);
    expect(m.outputs[0].shape).toBe(Shape.Scalar);
    expect(m.outputs[0].mnemonic).toBe('var.s(7)');
    expect(m.outputs[0].description).toBe('Unbiased estimation of the sample variance var.s(7)');
  });

  it('should return expected metadata for population variance', () => {
    const v = new Variance({length: 7, unbiased: false});
    const m = v.metadata();

    expect(m.identifier).toBe(IndicatorIdentifier.Variance);
    expect(m.mnemonic).toBe('var.p(7)');
    expect(m.description).toBe('Estimation of the population variance var.p(7)');
    expect(m.outputs.length).toBe(1);
    expect(m.outputs[0].kind).toBe(VarianceOutput.VarianceValue);
    expect(m.outputs[0].shape).toBe(Shape.Scalar);
    expect(m.outputs[0].mnemonic).toBe('var.p(7)');
    expect(m.outputs[0].description).toBe('Estimation of the population variance var.p(7)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new Variance({length: 1, unbiased: true}); }).toThrow();
  });

  it('should calculate expected Excel (VAR.P) output of population variance of length 3', () => {
    const expected = [
      Number.NaN, Number.NaN,
      9.55555555555556, 6.22222222222222, 4.66666666666667, 4.22222222222222, 1.55555555555556,
      9.55555555555556, 6.22222222222222, 2.88888888888889, 9.55555555555556, 14.88888888888890
    ];
    const len = 3;
    const varp = new Variance({length: len, unbiased: false});

    for (let i = 0; i < len - 1; i++) {
      expect(varp.update(input[i])).toBeNaN();
      expect(varp.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      expect(varp.update(input[i])).toBeCloseTo(expected[i], epsilon);
      expect(varp.isPrimed()).toBe(true);
    }

    expect(varp.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected Excel (VAR.P) output of population variance of length 5', () => {
    const expected = [
      Number.NaN, Number.NaN, Number.NaN, Number.NaN,
      10.16000, 6.56000, 2.96000, 9.36000, 5.76000, 6.00000, 11.04000, 12.24000
    ];
    const len = 5;
    const varp = new Variance({length: len, unbiased: false});

    for (let i = 0; i < len - 1; i++) {
      expect(varp.update(input[i])).toBeNaN();
      expect(varp.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      expect(varp.update(input[i])).toBeCloseTo(expected[i], epsilon);
      expect(varp.isPrimed()).toBe(true);
    }

    expect(varp.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected Excel (VAR.S) output of sample variance of length 3', () => {
    const expected = [
      Number.NaN, Number.NaN,
      14.33333333333330, 9.33333333333334, 7.00000000000000, 6.33333333333334, 2.33333333333333,
      14.33333333333330, 9.33333333333334, 4.33333333333334, 14.33333333333330, 22.33333333333330
    ];
    const len = 3;
    const vars = new Variance({length: len, unbiased: true});

    for (let i = 0; i < len - 1; i++) {
      expect(vars.update(input[i])).toBeNaN();
      expect(vars.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      expect(vars.update(input[i])).toBeCloseTo(expected[i], epsilon);
      expect(vars.isPrimed()).toBe(true);
    }

    expect(vars.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected Excel (VAR.S) output of sample variance of length 5', () => {
    const expected = [
      Number.NaN, Number.NaN, Number.NaN, Number.NaN,
      12.7000, 8.2000, 3.7000, 11.7000, 7.2000, 7.5000, 13.8000, 15.3000
    ];
    const len = 5;
    const vars = new Variance({length: len, unbiased: true});

    for (let i = 0; i < len - 1; i++) {
      expect(vars.update(input[i])).toBeNaN();
      expect(vars.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      expect(vars.update(input[i])).toBeCloseTo(expected[i], epsilon);
      expect(vars.isPrimed()).toBe(true);
    }

    expect(vars.update(Number.NaN)).toBeNaN();
  });
});
