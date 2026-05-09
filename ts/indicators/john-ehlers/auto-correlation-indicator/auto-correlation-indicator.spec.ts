import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { AutoCorrelationIndicator } from './auto-correlation-indicator';
import { AutoCorrelationIndicatorOutput } from './output';
import { testInput, snapshots } from './testdata';
import type { Snapshot, Spot } from './testdata';

const tolerance = 1e-12;
const minMaxTolerance = 1e-10;

// 252-entry TA-Lib MAMA reference series. Mirrors the Go test input.
// Snapshots mirror the Go implementation's reference values.
describe('AutoCorrelationIndicator', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(AutoCorrelationIndicatorOutput.Value).toBe(0);
  });

  it('should return expected mnemonic for default params', () => {
    const x = AutoCorrelationIndicator.default();
    expect(x.metadata().mnemonic).toBe('aci(3, 48, 10, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = AutoCorrelationIndicator.default();
    const meta = x.metadata();
    const mn = 'aci(3, 48, 10, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.AutoCorrelationIndicator);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Autocorrelation indicator ' + mn);
    expect(meta.outputs.length).toBe(1);

    expect(meta.outputs[0].kind).toBe(AutoCorrelationIndicatorOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Autocorrelation indicator ' + mn);
  });

  it('should return expected mnemonics for flag overrides', () => {
    const cases: Array<[any, string]> = [
      [{}, 'aci(3, 48, 10, hl/2)'],
      [{ averagingLength: 5 }, 'aci(3, 48, 10, average=5, hl/2)'],
      [{ minLag: 5, maxLag: 30, smoothingPeriod: 8 }, 'aci(5, 30, 8, hl/2)'],
    ];
    for (const [p, mn] of cases) {
      const x = AutoCorrelationIndicator.fromParams(p);
      expect(x.metadata().mnemonic).withContext(`params=${JSON.stringify(p)}`).toBe(mn);
    }
  });

  it('should throw for invalid params', () => {
    expect(() => AutoCorrelationIndicator.fromParams({ minLag: -1 })).toThrowError(/MinLag should be >= 1/);
    expect(() => AutoCorrelationIndicator.fromParams({ minLag: 10, maxLag: 10 })).toThrowError(/MaxLag should be > MinLag/);
    expect(() => AutoCorrelationIndicator.fromParams({ smoothingPeriod: 1 })).toThrowError(/SmoothingPeriod should be >= 2/);
    expect(() => AutoCorrelationIndicator.fromParams({ averagingLength: -1 })).toThrowError(/AveragingLength should be >= 0/);
  });

  it('should return empty heatmap for NaN input and not prime', () => {
    const x = AutoCorrelationIndicator.default();
    const h = x.update(Number.NaN, time);
    expect(h instanceof Heatmap).toBe(true);
    expect(h.isEmpty()).toBe(true);
    expect(h.parameterFirst).toBe(3);
    expect(h.parameterLast).toBe(48);
    expect(h.parameterResolution).toBe(1);
    expect(x.isPrimed()).toBe(false);
  });

  it('should match reference snapshots', () => {
    const x = AutoCorrelationIndicator.default();
    let si = 0;
    for (let i = 0; i < testInput.length; i++) {
      const t = new Date(time.getTime() + i * 60_000);
      const h = x.update(testInput[i], t);

      expect(h.parameterFirst).toBe(3);
      expect(h.parameterLast).toBe(48);
      expect(h.parameterResolution).toBe(1);

      if (!x.isPrimed()) {
        expect(h.isEmpty()).withContext(`bar ${i}`).toBe(true);
        continue;
      }

      expect(h.values.length).toBe(46);

      if (si < snapshots.length && snapshots[si].i === i) {
        const snap = snapshots[si];
        expect(Math.abs(h.valueMin - snap.valueMin))
          .withContext(`valueMin[${i}]`).toBeLessThan(minMaxTolerance);
        expect(Math.abs(h.valueMax - snap.valueMax))
          .withContext(`valueMax[${i}]`).toBeLessThan(minMaxTolerance);
        for (const [index, v] of snap.spots) {
          expect(Math.abs(h.values[index] - v))
            .withContext(`values[${i}][${index}]`).toBeLessThan(tolerance);
        }
        si++;
      }
    }
    expect(si).toBe(snapshots.length);
  });

  it('should place peak bin at injected period for synthetic sine', () => {
    // period=35: only multiple in [MinLag=3, MaxLag=48] is 35 itself.
    const period = 35;
    const bars = 600;

    const x = AutoCorrelationIndicator.default();

    let last: Heatmap | undefined;
    for (let i = 0; i < bars; i++) {
      const sample = 100 + Math.sin(2 * Math.PI * i / period);
      last = x.update(sample, new Date(time.getTime() + i * 60_000));
    }

    expect(last).toBeTruthy();
    expect(last!.isEmpty()).toBe(false);

    let peakBin = 0;
    for (let i = 0; i < last!.values.length; i++) {
      if (last!.values[i] > last!.values[peakBin]) peakBin = i;
    }

    const expectedBin = period - last!.parameterFirst;
    expect(peakBin).toBe(expectedBin);
  });

  const primeCount = 200;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(1);
    expect(out[0] instanceof Heatmap).toBe(true);
    expect((out[0] as Heatmap).time).toEqual(time);
  }

  it('should produce 1-element output via updateScalar', () => {
    const x = AutoCorrelationIndicator.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 1-element output via updateBar', () => {
    const x = AutoCorrelationIndicator.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 1-element output via updateQuote', () => {
    const x = AutoCorrelationIndicator.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 1-element output via updateTrade', () => {
    const x = AutoCorrelationIndicator.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
