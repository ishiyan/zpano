import { } from 'jasmine';

import { MovingAverageConvergenceDivergence } from './moving-average-convergence-divergence';
import { MovingAverageConvergenceDivergenceOutput } from './output';
import { MovingAverageType } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import {
    input,
    expectedMACD,
    expectedSignal,
    expectedHistogram,
} from './testdata';

// Test data from TA-Lib (252 entries), used by MBST C# tests.
// Expected MACD values (EMA12 - EMA26) for default params.
// Expected Signal values for default params.
// Expected Histogram values for default params.
describe('MovingAverageConvergenceDivergence', () => {

  it('should calculate default EMA(12,26,9) against all 252 expected values', () => {
    const tolerance = 1e-8;
    const ind = new MovingAverageConvergenceDivergence();

    for (let i = 0; i < 252; i++) {
      const [macd, signal, histogram] = ind.update(input[i]);

      if (isNaN(expectedMACD[i])) {
        expect(macd).withContext(`[${i}] macd`).toBeNaN();
      } else {
        expect(Math.abs(macd - expectedMACD[i])).withContext(`[${i}] macd`).toBeLessThan(tolerance);
      }

      if (isNaN(expectedSignal[i])) {
        expect(signal).withContext(`[${i}] signal`).toBeNaN();
      } else {
        expect(Math.abs(signal - expectedSignal[i])).withContext(`[${i}] signal`).toBeLessThan(tolerance);
      }

      if (isNaN(expectedHistogram[i])) {
        expect(histogram).withContext(`[${i}] histogram`).toBeNaN();
      } else {
        expect(Math.abs(histogram - expectedHistogram[i])).withContext(`[${i}] histogram`).toBeLessThan(tolerance);
      }
    }
  });

  it('should throw if fast length is less than 2', () => {
    expect(() => { new MovingAverageConvergenceDivergence({ fastLength: 1 }); }).toThrow();
  });

  it('should throw if slow length is less than 2', () => {
    expect(() => { new MovingAverageConvergenceDivergence({ slowLength: 1 }); }).toThrow();
  });

  it('should throw if signal length is less than 1', () => {
    expect(() => { new MovingAverageConvergenceDivergence({ signalLength: 0 }); }).toThrow();
  });

  it('should report correct primed state', () => {
    const ind = new MovingAverageConvergenceDivergence();

    expect(ind.isPrimed()).toBe(false);

    // Default: fast=12, slow=26, signal=9. Primed at index 33 (26-1 + 9-1 = 33).
    for (let i = 0; i < 33; i++) {
      ind.update(input[i]);
      expect(ind.isPrimed()).withContext(`index ${i}`).toBe(false);
    }

    ind.update(input[33]);
    expect(ind.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const ind = new MovingAverageConvergenceDivergence();
    const [macd, signal, histogram] = ind.update(NaN);
    expect(macd).toBeNaN();
    expect(signal).toBeNaN();
    expect(histogram).toBeNaN();
  });

  it('should return correct metadata (default EMA)', () => {
    const ind = new MovingAverageConvergenceDivergence();
    const meta = ind.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.MovingAverageConvergenceDivergence);
    expect(meta.mnemonic).toBe('macd(12,26,9)');
    expect(meta.description).toBe('Moving Average Convergence Divergence macd(12,26,9)');
    expect(meta.outputs.length).toBe(3);
    expect(meta.outputs[0].kind).toBe(MovingAverageConvergenceDivergenceOutput.MACDValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].kind).toBe(MovingAverageConvergenceDivergenceOutput.SignalValue);
    expect(meta.outputs[2].kind).toBe(MovingAverageConvergenceDivergenceOutput.HistogramValue);
  });

  it('should return correct metadata (SMA + EMA signal)', () => {
    const ind = new MovingAverageConvergenceDivergence({
      movingAverageType: MovingAverageType.SMA,
      signalMovingAverageType: MovingAverageType.EMA,
    });
    const meta = ind.metadata();

    expect(meta.mnemonic).toBe('macd(12,26,9,SMA,EMA)');
  });

  it('should auto-swap fast and slow when slow < fast', () => {
    const ind = new MovingAverageConvergenceDivergence({ fastLength: 26, slowLength: 12 });
    const meta = ind.metadata();

    // Should be swapped to fast=12, slow=26.
    expect(meta.mnemonic).toBe('macd(12,26,9)');
  });
});
