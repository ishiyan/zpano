import { } from 'jasmine';

import { Stochastic } from './stochastic';
import { StochasticOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { highs, lows, closes } from './testdata';

// Standard HLC test data (252 entries) from TA-Lib.
describe('Stochastic', () => {

  // Test 1: fastK=5, slowK=3/SMA, slowD=4/SMA.
  // begIndex=9, SlowK[0]=38.139, SlowD[0]=36.725.
  it('should calculate 5/SMA3/SMA4 single value at index 9', () => {
    const tolerance = 1e-2;
    const ind = new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 4 });

    for (let i = 0; i < 9; i++) {
      ind.update(closes[i], highs[i], lows[i]);
    }

    const [, slowK, slowD] = ind.update(closes[9], highs[9], lows[9]);
    expect(Math.abs(slowK - 38.139)).withContext('SlowK').toBeLessThan(tolerance);
    expect(Math.abs(slowD - 36.725)).withContext('SlowD').toBeLessThan(tolerance);
    expect(ind.isPrimed()).toBe(true);
  });

  // Test 2: fastK=5, slowK=3/SMA, slowD=3/SMA.
  // begIndex=8, first: SlowK[0]=24.0128, SlowD[0]=36.254.
  it('should calculate 5/SMA3/SMA3 first value at index 8', () => {
    const tolerance = 1e-2;
    const ind = new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 3 });

    for (let i = 0; i < 8; i++) {
      ind.update(closes[i], highs[i], lows[i]);
    }

    const [, slowK, slowD] = ind.update(closes[8], highs[8], lows[8]);
    expect(Math.abs(slowK - 24.0128)).withContext('SlowK').toBeLessThan(tolerance);
    expect(Math.abs(slowD - 36.254)).withContext('SlowD').toBeLessThan(tolerance);
    expect(ind.isPrimed()).toBe(true);
  });

  // Test 3: fastK=5, slowK=3/SMA, slowD=4/SMA.
  // Last: SlowK=30.194, SlowD=46.641.
  it('should calculate 5/SMA3/SMA4 last value', () => {
    const tolerance = 1e-2;
    const ind = new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 4 });

    let slowK = NaN;
    let slowD = NaN;

    for (let i = 0; i < 252; i++) {
      [, slowK, slowD] = ind.update(closes[i], highs[i], lows[i]);
    }

    expect(Math.abs(slowK - 30.194)).withContext('SlowK').toBeLessThan(tolerance);
    expect(Math.abs(slowD - 46.641)).withContext('SlowD').toBeLessThan(tolerance);
  });

  // Test 4: fastK=5, slowK=3/SMA, slowD=3/SMA.
  // Last: SlowK=30.194, SlowD=43.69.
  it('should calculate 5/SMA3/SMA3 last value', () => {
    const tolerance = 1e-2;
    const ind = new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 3 });

    let slowK = NaN;
    let slowD = NaN;

    for (let i = 0; i < 252; i++) {
      [, slowK, slowD] = ind.update(closes[i], highs[i], lows[i]);
    }

    expect(Math.abs(slowK - 30.194)).withContext('SlowK').toBeLessThan(tolerance);
    expect(Math.abs(slowD - 43.69)).withContext('SlowD').toBeLessThan(tolerance);
  });

  it('should report correct primed state', () => {
    const ind = new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 3 });

    expect(ind.isPrimed()).toBe(false);

    for (let i = 0; i < 8; i++) {
      ind.update(closes[i], highs[i], lows[i]);
      expect(ind.isPrimed()).withContext(`index ${i}`).toBe(false);
    }

    ind.update(closes[8], highs[8], lows[8]);
    expect(ind.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const ind = new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 3 });

    const [fastK, slowK, slowD] = ind.update(NaN, 1.0, 1.0);
    expect(fastK).toBeNaN();
    expect(slowK).toBeNaN();
    expect(slowD).toBeNaN();
  });

  it('should return correct metadata', () => {
    const ind = new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 3 });
    const meta = ind.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.Stochastic);
    expect(meta.mnemonic).toBe('stoch(5/SMA3/SMA3)');
    expect(meta.outputs.length).toBe(3);
    expect(meta.outputs[0].kind).toBe(StochasticOutput.FastK);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].kind).toBe(StochasticOutput.SlowK);
    expect(meta.outputs[2].kind).toBe(StochasticOutput.SlowD);
  });

  it('should validate parameters', () => {
    expect(() => new Stochastic({ fastKLength: 0, slowKLength: 3, slowDLength: 3 })).toThrow();
    expect(() => new Stochastic({ fastKLength: 5, slowKLength: 0, slowDLength: 3 })).toThrow();
    expect(() => new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 0 })).toThrow();
    expect(() => new Stochastic({ fastKLength: 5, slowKLength: 3, slowDLength: 3 })).not.toThrow();
  });
});
