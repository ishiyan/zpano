import { } from 'jasmine';

import { BollingerBandsTrend } from './bollinger-bands-trend';
import { BollingerBandsTrendOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { Scalar } from '../../../entities/scalar';
import { closingPrices, sampleExpected, populationExpected } from './testdata';

describe('BollingerBandsTrend', () => {

  it('should calculate sample (unbiased) full data against all 252 expected values', () => {
    const tolerance = 1e-8;
    const ind = new BollingerBandsTrend({ fastLength: 20, slowLength: 50, upperMultiplier: 2, lowerMultiplier: 2 });

    for (let i = 0; i < 252; i++) {
      const v = ind.update(closingPrices[i]);

      if (isNaN(sampleExpected[i])) {
        expect(v).withContext(`[${i}]`).toBeNaN();
        continue;
      }

      expect(Math.abs(v - sampleExpected[i])).withContext(`[${i}] got ${v} want ${sampleExpected[i]}`).toBeLessThan(tolerance);
    }
  });

  it('should calculate population (biased) full data against all 252 expected values', () => {
    const tolerance = 1e-8;
    const ind = new BollingerBandsTrend({ fastLength: 20, slowLength: 50, upperMultiplier: 2, lowerMultiplier: 2, isUnbiased: false });

    for (let i = 0; i < 252; i++) {
      const v = ind.update(closingPrices[i]);

      if (isNaN(populationExpected[i])) {
        expect(v).withContext(`[${i}]`).toBeNaN();
        continue;
      }

      expect(Math.abs(v - populationExpected[i])).withContext(`[${i}] got ${v} want ${populationExpected[i]}`).toBeLessThan(tolerance);
    }
  });

  it('should report correct primed state', () => {
    const ind = new BollingerBandsTrend({ fastLength: 20, slowLength: 50, upperMultiplier: 2, lowerMultiplier: 2 });

    expect(ind.isPrimed()).toBe(false);

    // Not primed until slowLength (50) samples have been fed
    for (let i = 0; i < 49; i++) {
      ind.update(closingPrices[i]);
      expect(ind.isPrimed()).withContext(`index ${i}`).toBe(false);
    }

    ind.update(closingPrices[49]);
    expect(ind.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const ind = new BollingerBandsTrend({ fastLength: 20, slowLength: 50, upperMultiplier: 2, lowerMultiplier: 2 });

    const v = ind.update(NaN);
    expect(v).toBeNaN();
  });

  it('should return correct metadata', () => {
    const ind = new BollingerBandsTrend({ fastLength: 20, slowLength: 50, upperMultiplier: 2, lowerMultiplier: 2 });
    const meta = ind.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.BollingerBandsTrend);
    expect(meta.mnemonic).toBe('bbtrend(20,50,2,2)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(BollingerBandsTrendOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should validate parameters', () => {
    expect(() => new BollingerBandsTrend({ fastLength: 1, slowLength: 50 })).toThrow();
    expect(() => new BollingerBandsTrend({ fastLength: 20, slowLength: 1 })).toThrow();
    expect(() => new BollingerBandsTrend({ fastLength: 20, slowLength: 20 })).toThrow();
    expect(() => new BollingerBandsTrend({ fastLength: 20, slowLength: 10 })).toThrow();
    expect(() => new BollingerBandsTrend({ fastLength: 20, slowLength: 50 })).not.toThrow();
  });

  it('should work with updateScalar', () => {
    const ind = new BollingerBandsTrend({ fastLength: 20, slowLength: 50, upperMultiplier: 2, lowerMultiplier: 2 });

    for (let i = 0; i < 252; i++) {
      const s = new Scalar();
      s.value = closingPrices[i];
      s.time = new Date(2020, 0, i + 1);
      const output = ind.updateScalar(s);

      expect(output.length).toBe(1);

      const result = output[0] as Scalar;
      expect(result.time).toEqual(s.time);

      if (isNaN(sampleExpected[i])) {
        expect(result.value).withContext(`[${i}]`).toBeNaN();
      } else {
        expect(Math.abs(result.value - sampleExpected[i])).withContext(`[${i}]`).toBeLessThan(1e-8);
      }
    }
  });
});
