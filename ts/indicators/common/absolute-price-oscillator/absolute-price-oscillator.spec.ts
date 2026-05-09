import { } from 'jasmine';

import { AbsolutePriceOscillator } from './absolute-price-oscillator';
import { MovingAverageType } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { AbsolutePriceOscillatorOutput } from './output';
import { input } from './testdata';

// Test data from TA-Lib (252 entries), used by MBST C# tests.
describe('AbsolutePriceOscillator', () => {

  it('should throw if fast length is less than 2', () => {
    expect(() => { new AbsolutePriceOscillator({ fastLength: 1, slowLength: 26 }); }).toThrow();
  });

  it('should throw if slow length is less than 2', () => {
    expect(() => { new AbsolutePriceOscillator({ fastLength: 12, slowLength: 1 }); }).toThrow();
  });

  it('should calculate SMA(12,26) correctly', () => {
    const tolerance = 5e-4;
    const apo = new AbsolutePriceOscillator({ fastLength: 12, slowLength: 26 });

    // First 25 values NaN.
    for (let i = 0; i < 25; i++) {
      expect(apo.update(input[i])).toBeNaN();
    }

    // Index 25: first value.
    let v = apo.update(input[25]);
    expect(Math.abs(v - (-3.3124))).toBeLessThan(tolerance);

    // Index 26: second value.
    v = apo.update(input[26]);
    expect(Math.abs(v - (-3.5876))).toBeLessThan(tolerance);

    // Feed remaining.
    for (let i = 27; i < 251; i++) {
      apo.update(input[i]);
    }

    // Last value.
    v = apo.update(input[251]);
    expect(Math.abs(v - (-0.1667))).toBeLessThan(tolerance);
    expect(apo.isPrimed()).toBe(true);
  });

  it('should calculate EMA(12,26) correctly', () => {
    const tolerance = 5e-4;
    const apo = new AbsolutePriceOscillator({
      fastLength: 12,
      slowLength: 26,
      movingAverageType: MovingAverageType.EMA,
      firstIsAverage: false,
    });

    // First 25 values NaN.
    for (let i = 0; i < 25; i++) {
      expect(apo.update(input[i])).toBeNaN();
    }

    // Index 25: first value.
    let v = apo.update(input[25]);
    expect(Math.abs(v - (-2.4193))).toBeLessThan(tolerance);

    // Index 26: second value.
    v = apo.update(input[26]);
    expect(Math.abs(v - (-2.4367))).toBeLessThan(tolerance);

    // Feed remaining.
    for (let i = 27; i < 251; i++) {
      apo.update(input[i]);
    }

    // Last value.
    v = apo.update(input[251]);
    expect(Math.abs(v - 0.90401)).toBeLessThan(tolerance);
  });

  it('should report correct primed state', () => {
    const apo = new AbsolutePriceOscillator({ fastLength: 3, slowLength: 5 });

    expect(apo.isPrimed()).toBe(false);

    for (let i = 1; i < 5; i++) {
      apo.update(i);
      expect(apo.isPrimed()).toBe(false);
    }

    apo.update(5);
    expect(apo.isPrimed()).toBe(true);

    for (let i = 6; i < 10; i++) {
      apo.update(i);
      expect(apo.isPrimed()).toBe(true);
    }
  });

  it('should pass NaN through', () => {
    const apo = new AbsolutePriceOscillator({ fastLength: 2, slowLength: 3 });
    expect(apo.update(Number.NaN)).toBeNaN();
  });

  it('should return correct metadata (SMA)', () => {
    const apo = new AbsolutePriceOscillator({ fastLength: 12, slowLength: 26 });
    const meta = apo.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.AbsolutePriceOscillator);
    expect(meta.mnemonic).toBe('apo(SMA12/SMA26)');
    expect(meta.description).toBe('Absolute Price Oscillator apo(SMA12/SMA26)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(AbsolutePriceOscillatorOutput.AbsolutePriceOscillatorValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should return correct metadata (EMA)', () => {
    const apo = new AbsolutePriceOscillator({
      fastLength: 12,
      slowLength: 26,
      movingAverageType: MovingAverageType.EMA,
    });
    const meta = apo.metadata();

    expect(meta.mnemonic).toBe('apo(EMA12/EMA26)');
  });
});
