import { } from 'jasmine';

import { BollingerBands } from './bollinger-bands';
import { BollingerBandsOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import {
    closingPrices,
    sma20Expected,
    sampleLowerExpected,
    sampleUpperExpected,
    sampleBandWidthExpected,
    samplePercentBandExpected,
    populationLowerExpected,
    populationUpperExpected,
    populationBandWidthExpected,
    populationPercentBandExpected,
} from './testdata';

// Population (unbiased=false) expected values — middle is same as sample.
describe('BollingerBands', () => {

  it('should calculate sample (unbiased) full data against all 252 expected values', () => {
    const tolerance = 1e-8;
    const ind = new BollingerBands({ length: 20, upperMultiplier: 2, lowerMultiplier: 2 });

    for (let i = 0; i < 252; i++) {
      const [lower, middle, upper, bw, pctB] = ind.update(closingPrices[i]);

      if (isNaN(sma20Expected[i])) {
        expect(lower).withContext(`[${i}] lower`).toBeNaN();
        expect(middle).withContext(`[${i}] middle`).toBeNaN();
        expect(upper).withContext(`[${i}] upper`).toBeNaN();
        expect(bw).withContext(`[${i}] bandWidth`).toBeNaN();
        expect(pctB).withContext(`[${i}] percentBand`).toBeNaN();
        continue;
      }

      expect(Math.abs(lower - sampleLowerExpected[i])).withContext(`[${i}] lower`).toBeLessThan(tolerance);
      expect(Math.abs(middle - sma20Expected[i])).withContext(`[${i}] middle`).toBeLessThan(tolerance);
      expect(Math.abs(upper - sampleUpperExpected[i])).withContext(`[${i}] upper`).toBeLessThan(tolerance);
      expect(Math.abs(bw - sampleBandWidthExpected[i])).withContext(`[${i}] bandWidth`).toBeLessThan(tolerance);
      expect(Math.abs(pctB - samplePercentBandExpected[i])).withContext(`[${i}] percentBand`).toBeLessThan(tolerance);
    }
  });

  it('should calculate population full data against all 252 expected values', () => {
    const tolerance = 1e-8;
    const ind = new BollingerBands({ length: 20, upperMultiplier: 2, lowerMultiplier: 2, isUnbiased: false });

    for (let i = 0; i < 252; i++) {
      const [lower, middle, upper, bw, pctB] = ind.update(closingPrices[i]);

      if (isNaN(sma20Expected[i])) {
        expect(lower).withContext(`[${i}] lower`).toBeNaN();
        continue;
      }

      expect(Math.abs(lower - populationLowerExpected[i])).withContext(`[${i}] lower`).toBeLessThan(tolerance);
      expect(Math.abs(middle - sma20Expected[i])).withContext(`[${i}] middle`).toBeLessThan(tolerance);
      expect(Math.abs(upper - populationUpperExpected[i])).withContext(`[${i}] upper`).toBeLessThan(tolerance);
      expect(Math.abs(bw - populationBandWidthExpected[i])).withContext(`[${i}] bandWidth`).toBeLessThan(tolerance);
      expect(Math.abs(pctB - populationPercentBandExpected[i])).withContext(`[${i}] percentBand`).toBeLessThan(tolerance);
    }
  });

  it('should report correct primed state', () => {
    const ind = new BollingerBands({ length: 20, upperMultiplier: 2, lowerMultiplier: 2 });

    expect(ind.isPrimed()).toBe(false);

    for (let i = 0; i < 19; i++) {
      ind.update(closingPrices[i]);
      expect(ind.isPrimed()).withContext(`index ${i}`).toBe(false);
    }

    ind.update(closingPrices[19]);
    expect(ind.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const ind = new BollingerBands({ length: 20, upperMultiplier: 2, lowerMultiplier: 2 });

    const [lower, middle, upper, bw, pctB] = ind.update(NaN);
    expect(lower).toBeNaN();
    expect(middle).toBeNaN();
    expect(upper).toBeNaN();
    expect(bw).toBeNaN();
    expect(pctB).toBeNaN();
  });

  it('should return correct metadata', () => {
    const ind = new BollingerBands({ length: 20, upperMultiplier: 2, lowerMultiplier: 2 });
    const meta = ind.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.BollingerBands);
    expect(meta.mnemonic).toBe('bb(20,2,2)');
    expect(meta.outputs.length).toBe(6);
    expect(meta.outputs[0].kind).toBe(BollingerBandsOutput.LowerValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].kind).toBe(BollingerBandsOutput.MiddleValue);
    expect(meta.outputs[2].kind).toBe(BollingerBandsOutput.UpperValue);
    expect(meta.outputs[3].kind).toBe(BollingerBandsOutput.BandWidth);
    expect(meta.outputs[4].kind).toBe(BollingerBandsOutput.PercentBand);
    expect(meta.outputs[5].kind).toBe(BollingerBandsOutput.Band);
    expect(meta.outputs[5].shape).toBe(Shape.Band);
  });

  it('should validate parameters', () => {
    expect(() => new BollingerBands({ length: 1 })).toThrow();
    expect(() => new BollingerBands({ length: 0 })).toThrow();
    expect(() => new BollingerBands({ length: -1 })).toThrow();
    expect(() => new BollingerBands({ length: 2 })).not.toThrow();
    expect(() => new BollingerBands({ length: 20 })).not.toThrow();
  });
});
