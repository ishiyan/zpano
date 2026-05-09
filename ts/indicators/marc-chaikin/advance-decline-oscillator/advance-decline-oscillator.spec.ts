import { } from 'jasmine';

import { AdvanceDeclineOscillator } from './advance-decline-oscillator';
import { AdvanceDeclineOscillatorOutput } from './output';
import { MovingAverageType } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import {
    testHighs,
    testLows,
    testCloses,
    testVolumes,
    testExpectedEMA,
    testExpectedSMA,
} from './testdata';

// High test data, 252 entries. From TA-Lib excel-sma3-sma10-chaikin.csv.
// Low test data, 252 entries.
// Close test data, 252 entries.
// Volume test data, 252 entries.
// Expected EMA ADOSC output, 252 entries. First 9 are NaN (lookback = 9 for EMA(10)).
// Expected SMA ADOSC output, 252 entries. First 9 are NaN (lookback = 9 for SMA(10)).
function roundTo(v: number, digits: number): number {
  const p = Math.pow(10, digits);
  return Math.round(v * p) / p;
}

describe('AdvanceDeclineOscillator', () => {

  it('should calculate ADOSC with EMA correctly', () => {
    const digits = 2;
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.EMA,
    });

    for (let i = 0; i < testHighs.length; i++) {
      const v = adosc.updateHLCV(testHighs[i], testLows[i], testCloses[i], testVolumes[i]);

      if (i < 9) {
        expect(v).toBeNaN();
        expect(adosc.isPrimed()).toBe(false);
        continue;
      }

      expect(v).not.toBeNaN();
      expect(adosc.isPrimed()).toBe(true);
      expect(roundTo(v, digits)).toBe(roundTo(testExpectedEMA[i], digits));
    }
  });

  it('should calculate ADOSC with SMA correctly', () => {
    const digits = 2;
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.SMA,
    });

    for (let i = 0; i < testHighs.length; i++) {
      const v = adosc.updateHLCV(testHighs[i], testLows[i], testCloses[i], testVolumes[i]);

      if (i < 9) {
        expect(v).toBeNaN();
        expect(adosc.isPrimed()).toBe(false);
        continue;
      }

      expect(v).not.toBeNaN();
      expect(adosc.isPrimed()).toBe(true);
      expect(roundTo(v, digits)).toBe(roundTo(testExpectedSMA[i], digits));
    }
  });

  it('should match TA-Lib spot checks', () => {
    const digits = 2;
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.EMA,
    });

    const values: number[] = [];
    for (let i = 0; i < testHighs.length; i++) {
      values.push(adosc.updateHLCV(testHighs[i], testLows[i], testCloses[i], testVolumes[i]));
    }

    // TA-Lib spot checks from test_per_hlcv.c.
    expect(roundTo(values[9], digits)).toBe(roundTo(841238.33, digits));
    expect(roundTo(values[10], digits)).toBe(roundTo(2255663.07, digits));
    expect(roundTo(values[250], digits)).toBe(roundTo(-526700.32, digits));
    expect(roundTo(values[251], digits)).toBe(roundTo(-1139932.73, digits));
  });

  it('should match ADOSC(5,2) spot check', () => {
    const digits = 2;
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 5,
      slowLength: 2,
      movingAverageType: MovingAverageType.EMA,
    });

    const values: number[] = [];
    for (let i = 0; i < testHighs.length; i++) {
      values.push(adosc.updateHLCV(testHighs[i], testLows[i], testCloses[i], testVolumes[i]));
    }

    // begIndex=4, output[0] at index 4.
    expect(roundTo(values[4], digits)).toBe(roundTo(585361.29, digits));
  });

  it('should not be primed initially', () => {
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.EMA,
    });
    expect(adosc.isPrimed()).toBe(false);
  });

  it('should not become primed after NaN update', () => {
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.EMA,
    });
    adosc.update(Number.NaN);
    expect(adosc.isPrimed()).toBe(false);
  });

  it('should pass NaN through', () => {
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.EMA,
    });
    expect(adosc.update(Number.NaN)).toBeNaN();
    expect(adosc.updateHLCV(Number.NaN, 1, 2, 3)).toBeNaN();
    expect(adosc.updateHLCV(1, Number.NaN, 2, 3)).toBeNaN();
    expect(adosc.updateHLCV(1, 2, Number.NaN, 3)).toBeNaN();
    expect(adosc.updateHLCV(1, 2, 3, Number.NaN)).toBeNaN();
  });

  it('should return correct metadata', () => {
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.EMA,
    });
    const meta = adosc.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.AdvanceDeclineOscillator);
    expect(meta.mnemonic).toBe('adosc(EMA3/EMA10)');
    expect(meta.description).toBe('Chaikin Advance-Decline Oscillator adosc(EMA3/EMA10)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(AdvanceDeclineOscillatorOutput.AdvanceDeclineOscillatorValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should return SMA mnemonic', () => {
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
      movingAverageType: MovingAverageType.SMA,
    });
    const meta = adosc.metadata();

    expect(meta.mnemonic).toBe('adosc(SMA3/SMA10)');
  });

  it('should default to EMA when movingAverageType is undefined', () => {
    const adosc = new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 10,
    });
    const meta = adosc.metadata();

    expect(meta.mnemonic).toBe('adosc(EMA3/EMA10)');
  });

  it('should throw for fast length < 2', () => {
    expect(() => new AdvanceDeclineOscillator({
      fastLength: 1,
      slowLength: 10,
      movingAverageType: MovingAverageType.EMA,
    })).toThrow();
  });

  it('should throw for slow length < 2', () => {
    expect(() => new AdvanceDeclineOscillator({
      fastLength: 3,
      slowLength: 1,
      movingAverageType: MovingAverageType.EMA,
    })).toThrow();
  });
});
