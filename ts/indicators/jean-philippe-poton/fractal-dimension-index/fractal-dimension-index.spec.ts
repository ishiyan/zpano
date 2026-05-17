import { } from 'jasmine';

import { FractalDimensionIndex } from './fractal-dimension-index';
import { testInput, expected_P5, expected_P10, expected_P15, expected_P20, expected_P30, expected_P50, expected_P80, expected_P120 } from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('FractalDimensionIndex', () => {

  it('should return expected mnemonic', () => {
    const fdi = new FractalDimensionIndex({period: 30});
    expect(fdi.metadata().mnemonic).toBe('fdi(30)');
  });

  it('should throw if period is less than 2', () => {
    expect(() => { new FractalDimensionIndex({period: 1}); }).toThrow();
  });

  it('should have correct identifier in metadata', () => {
    const fdi = new FractalDimensionIndex({period: 30});
    expect(fdi.metadata().identifier).toBe(IndicatorIdentifier.FractalDimensionIndex);
  });

  it('should calculate expected output for period 5', () => {
    const fdi = new FractalDimensionIndex({period: 5});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P5[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P5[i], 13);
      }
    }
  });

  it('should calculate expected output for period 10', () => {
    const fdi = new FractalDimensionIndex({period: 10});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P10[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P10[i], 13);
      }
    }
  });

  it('should calculate expected output for period 15', () => {
    const fdi = new FractalDimensionIndex({period: 15});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P15[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P15[i], 13);
      }
    }
  });

  it('should calculate expected output for period 20', () => {
    const fdi = new FractalDimensionIndex({period: 20});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P20[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P20[i], 13);
      }
    }
  });

  it('should calculate expected output for period 30', () => {
    const fdi = new FractalDimensionIndex({period: 30});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P30[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P30[i], 13);
      }
    }
  });

  it('should calculate expected output for period 50', () => {
    const fdi = new FractalDimensionIndex({period: 50});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P50[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P50[i], 13);
      }
    }
  });

  it('should calculate expected output for period 80', () => {
    const fdi = new FractalDimensionIndex({period: 80});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P80[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P80[i], 13);
      }
    }
  });

  it('should calculate expected output for period 120', () => {
    const fdi = new FractalDimensionIndex({period: 120});

    for (let i = 0; i < testInput.length; i++) {
      const result = fdi.update(testInput[i]);
      if (Number.isNaN(expected_P120[i])) {
        expect(result).toBeNaN();
      } else {
        expect(result).toBeCloseTo(expected_P120[i], 13);
      }
    }
  });

  it('should report primed state correctly', () => {
    const fdi = new FractalDimensionIndex({period: 30});
    for (let i = 0; i < 30; i++) {
      fdi.update(testInput[i]);
      expect(fdi.isPrimed()).toBe(false);
    }
    fdi.update(testInput[30]);
    expect(fdi.isPrimed()).toBe(true);
  });

  it('should pass through NaN', () => {
    const fdi = new FractalDimensionIndex({period: 5});
    expect(fdi.update(Number.NaN)).toBeNaN();
  });
});
