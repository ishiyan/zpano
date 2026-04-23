import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { coronaTestInput } from '../corona/test-input';
import { CoronaTrendVigor } from './corona-trend-vigor';
import { CoronaTrendVigorOutput } from './output';

const tolerance = 1e-4;

describe('CoronaTrendVigor', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(CoronaTrendVigorOutput.Value).toBe(0);
    expect(CoronaTrendVigorOutput.TrendVigor).toBe(1);
  });

  it('should return expected mnemonic for default params', () => {
    const x = CoronaTrendVigor.default();
    expect(x.metadata().mnemonic).toBe('ctv(50, 20, -10, 10, 30, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = CoronaTrendVigor.default();
    const meta = x.metadata();
    const mn = 'ctv(50, 20, -10, 10, 30, hl/2)';
    const mnTV = 'ctv-tv(30, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.CoronaTrendVigor);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.outputs.length).toBe(2);
    expect(meta.outputs[0].kind).toBe(CoronaTrendVigorOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[1].kind).toBe(CoronaTrendVigorOutput.TrendVigor);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnTV);
  });

  it('should throw for invalid params', () => {
    expect(() => CoronaTrendVigor.fromParams({ rasterLength: 1 })).toThrowError(/RasterLength/);
    expect(() => CoronaTrendVigor.fromParams({ minParameterValue: 5, maxParameterValue: 5 }))
      .toThrowError(/MaxParameterValue/);
    expect(() => CoronaTrendVigor.fromParams({ highPassFilterCutoff: 1 })).toThrowError(/HighPassFilterCutoff/);
    expect(() => CoronaTrendVigor.fromParams({ minimalPeriod: 1 })).toThrowError(/MinimalPeriod/);
  });

  it('should return empty heatmap and NaN for NaN input', () => {
    const x = CoronaTrendVigor.default();
    const [h, tv] = x.update(Number.NaN, time);
    expect(h.isEmpty()).toBe(true);
    expect(Number.isNaN(tv)).toBe(true);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 11', () => {
    const x = CoronaTrendVigor.default();
    let primedAt = -1;
    for (let i = 0; i < coronaTestInput.length; i++) {
      x.update(coronaTestInput[i], new Date(time.getTime() + i * 60_000));
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(11);
  });

  it('should match reference tv / vmin / vmax snapshots', () => {
    const snapshots: Array<[number, number, number, number]> = [
      [11, 5.6512200755, 20.0000000000, 20.0000000000],
      [12, 6.8379492897, 20.0000000000, 20.0000000000],
      [50, 2.6145116709, 2.3773561485, 20.0000000000],
      [100, 2.7536803664, 2.4892742850, 20.0000000000],
      [150, -6.4606404251, 20.0000000000, 20.0000000000],
      [200, -10.0000000000, 20.0000000000, 20.0000000000],
      [251, -0.1894989954, 0.5847573715, 20.0000000000],
    ];

    const x = CoronaTrendVigor.default();
    let si = 0;
    for (let i = 0; i < coronaTestInput.length; i++) {
      const t = new Date(time.getTime() + i * 60_000);
      const [h, tv] = x.update(coronaTestInput[i], t);

      expect(h.parameterFirst).toBe(-10);
      expect(h.parameterLast).toBe(10);
      expect(Math.abs(h.parameterResolution - 2.45)).toBeLessThan(1e-9);

      if (!x.isPrimed()) {
        expect(h.isEmpty()).toBe(true);
        expect(Number.isNaN(tv)).toBe(true);
        continue;
      }

      expect(h.values.length).toBe(50);

      if (si < snapshots.length && snapshots[si][0] === i) {
        expect(Math.abs(snapshots[si][1] - tv))
          .withContext(`tv[${i}]: expected ${snapshots[si][1]}, actual ${tv}`).toBeLessThan(tolerance);
        expect(Math.abs(snapshots[si][2] - h.valueMin))
          .withContext(`vmin[${i}]: expected ${snapshots[si][2]}, actual ${h.valueMin}`).toBeLessThan(tolerance);
        expect(Math.abs(snapshots[si][3] - h.valueMax))
          .withContext(`vmax[${i}]: expected ${snapshots[si][3]}, actual ${h.valueMax}`).toBeLessThan(tolerance);
        si++;
      }
    }
    expect(si).toBe(snapshots.length);
  });

  const primeCount = 50;
  const value = 100.0;

  function prime(x: CoronaTrendVigor): void {
    for (let i = 0; i < primeCount; i++) {
      x.update(coronaTestInput[i % coronaTestInput.length], time);
    }
  }

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(2);
    expect(out[0] instanceof Heatmap).toBe(true);
    expect((out[0] as Heatmap).time).toEqual(time);
    expect((out[1] as Scalar).time).toEqual(time);
  }

  it('should produce 2-element output via updateScalar', () => {
    const x = CoronaTrendVigor.default();
    prime(x);
    const s = new Scalar(); s.time = time; s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 2-element output via updateBar', () => {
    const x = CoronaTrendVigor.default();
    prime(x);
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 2-element output via updateQuote', () => {
    const x = CoronaTrendVigor.default();
    prime(x);
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 2-element output via updateTrade', () => {
    const x = CoronaTrendVigor.default();
    prime(x);
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
