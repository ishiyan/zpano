import { } from 'jasmine';

import { PercentagePriceOscillator } from './percentage-price-oscillator';
import { MovingAverageType } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { PercentagePriceOscillatorOutput } from './output';
import { input } from './testdata';

// Test data from TA-Lib (252 entries), used by MBST C# tests.
describe('PercentagePriceOscillator', () => {

  it('should throw if fast length is less than 2', () => {
    expect(() => { new PercentagePriceOscillator({ fastLength: 1, slowLength: 26 }); }).toThrow();
  });

  it('should throw if slow length is less than 2', () => {
    expect(() => { new PercentagePriceOscillator({ fastLength: 12, slowLength: 1 }); }).toThrow();
  });

  it('should calculate SMA(2,3) correctly', () => {
    const tolerance = 5e-4;
    const ppo = new PercentagePriceOscillator({ fastLength: 2, slowLength: 3 });

    // First 2 values NaN.
    for (let i = 0; i < 2; i++) {
      expect(ppo.update(input[i])).toBeNaN();
    }

    // Index 2: first value.
    let v = ppo.update(input[2]);
    expect(Math.abs(v - 1.10264)).toBeLessThan(tolerance);

    // Index 3: second value.
    v = ppo.update(input[3]);
    expect(Math.abs(v - (-0.02813))).toBeLessThan(tolerance);

    // Feed remaining.
    for (let i = 4; i < 251; i++) {
      ppo.update(input[i]);
    }

    // Last value.
    v = ppo.update(input[251]);
    expect(Math.abs(v - (-0.21191))).toBeLessThan(tolerance);
    expect(ppo.isPrimed()).toBe(true);
  });

  it('should calculate SMA(12,26) correctly', () => {
    const tolerance = 5e-4;
    const ppo = new PercentagePriceOscillator({ fastLength: 12, slowLength: 26 });

    // First 25 values NaN.
    for (let i = 0; i < 25; i++) {
      expect(ppo.update(input[i])).toBeNaN();
    }

    // Index 25: first value.
    let v = ppo.update(input[25]);
    expect(Math.abs(v - (-3.6393))).toBeLessThan(tolerance);

    // Index 26: second value.
    v = ppo.update(input[26]);
    expect(Math.abs(v - (-3.9534))).toBeLessThan(tolerance);

    // Feed remaining.
    for (let i = 27; i < 251; i++) {
      ppo.update(input[i]);
    }

    // Last value.
    v = ppo.update(input[251]);
    expect(Math.abs(v - (-0.15281))).toBeLessThan(tolerance);
  });

  it('should calculate EMA(12,26) correctly', () => {
    const tolerance = 5e-3;
    const ppo = new PercentagePriceOscillator({
      fastLength: 12,
      slowLength: 26,
      movingAverageType: MovingAverageType.EMA,
      firstIsAverage: false,
    });

    // First 25 values NaN.
    for (let i = 0; i < 25; i++) {
      expect(ppo.update(input[i])).toBeNaN();
    }

    // Index 25: first value.
    let v = ppo.update(input[25]);
    expect(Math.abs(v - (-2.7083))).toBeLessThan(tolerance);

    // Index 26: second value.
    v = ppo.update(input[26]);
    expect(Math.abs(v - (-2.7390))).toBeLessThan(tolerance);

    // Feed remaining.
    for (let i = 27; i < 251; i++) {
      ppo.update(input[i]);
    }

    // Last value.
    v = ppo.update(input[251]);
    expect(Math.abs(v - 0.83644)).toBeLessThan(tolerance);
  });

  it('should report correct primed state', () => {
    const ppo = new PercentagePriceOscillator({ fastLength: 3, slowLength: 5 });

    expect(ppo.isPrimed()).toBe(false);

    for (let i = 1; i < 5; i++) {
      ppo.update(i);
      expect(ppo.isPrimed()).toBe(false);
    }

    ppo.update(5);
    expect(ppo.isPrimed()).toBe(true);

    for (let i = 6; i < 10; i++) {
      ppo.update(i);
      expect(ppo.isPrimed()).toBe(true);
    }
  });

  it('should pass NaN through', () => {
    const ppo = new PercentagePriceOscillator({ fastLength: 2, slowLength: 3 });
    expect(ppo.update(Number.NaN)).toBeNaN();
  });

  it('should return correct metadata (SMA)', () => {
    const ppo = new PercentagePriceOscillator({ fastLength: 12, slowLength: 26 });
    const meta = ppo.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.PercentagePriceOscillator);
    expect(meta.mnemonic).toBe('ppo(SMA12/SMA26)');
    expect(meta.description).toBe('Percentage Price Oscillator ppo(SMA12/SMA26)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(PercentagePriceOscillatorOutput.PercentagePriceOscillatorValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should return correct metadata (EMA)', () => {
    const ppo = new PercentagePriceOscillator({
      fastLength: 12,
      slowLength: 26,
      movingAverageType: MovingAverageType.EMA,
    });
    const meta = ppo.metadata();

    expect(meta.mnemonic).toBe('ppo(EMA12/EMA26)');
  });
});
