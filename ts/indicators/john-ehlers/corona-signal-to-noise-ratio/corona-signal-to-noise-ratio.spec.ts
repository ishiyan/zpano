import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { coronaTestInput } from '../corona/test-input';
import { CoronaSignalToNoiseRatio } from './corona-signal-to-noise-ratio';
import { CoronaSignalToNoiseRatioOutput } from './output';

const tolerance = 1e-4;

// Produce synthetic High/Low around the sample matching the Go reference.
function makeHL(i: number, sample: number): [number, number] {
  const frac = 0.005 + 0.03 * (1 + Math.sin(i * 0.37));
  const half = sample * frac;
  return [sample - half, sample + half];
}

describe('CoronaSignalToNoiseRatio', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(CoronaSignalToNoiseRatioOutput.Value).toBe(0);
    expect(CoronaSignalToNoiseRatioOutput.SignalToNoiseRatio).toBe(1);
  });

  it('should return expected mnemonic for default params', () => {
    const x = CoronaSignalToNoiseRatio.default();
    expect(x.metadata().mnemonic).toBe('csnr(50, 20, 1, 11, 30, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = CoronaSignalToNoiseRatio.default();
    const meta = x.metadata();
    const mn = 'csnr(50, 20, 1, 11, 30, hl/2)';
    const mnSNR = 'csnr-snr(30, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.CoronaSignalToNoiseRatio);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Corona signal to noise ratio ' + mn);
    expect(meta.outputs.length).toBe(2);

    expect(meta.outputs[0].kind).toBe(CoronaSignalToNoiseRatioOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[0].mnemonic).toBe(mn);

    expect(meta.outputs[1].kind).toBe(CoronaSignalToNoiseRatioOutput.SignalToNoiseRatio);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnSNR);
  });

  it('should throw for invalid params', () => {
    expect(() => CoronaSignalToNoiseRatio.fromParams({ rasterLength: 1 })).toThrowError(/RasterLength/);
    expect(() => CoronaSignalToNoiseRatio.fromParams({ minParameterValue: 5, maxParameterValue: 5 }))
      .toThrowError(/MaxParameterValue/);
    expect(() => CoronaSignalToNoiseRatio.fromParams({ highPassFilterCutoff: 1 })).toThrowError(/HighPassFilterCutoff/);
    expect(() => CoronaSignalToNoiseRatio.fromParams({ minimalPeriod: 1 })).toThrowError(/MinimalPeriod/);
    expect(() => CoronaSignalToNoiseRatio.fromParams({ minimalPeriod: 10, maximalPeriod: 10 }))
      .toThrowError(/MaximalPeriod/);
  });

  it('should return empty heatmap and NaN for NaN input', () => {
    const x = CoronaSignalToNoiseRatio.default();
    const [h, snr] = x.update(Number.NaN, Number.NaN, Number.NaN, time);
    expect(h instanceof Heatmap).toBe(true);
    expect(h.isEmpty()).toBe(true);
    expect(Number.isNaN(snr)).toBe(true);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 11', () => {
    const x = CoronaSignalToNoiseRatio.default();
    expect(x.isPrimed()).toBe(false);
    let primedAt = -1;
    for (let i = 0; i < coronaTestInput.length; i++) {
      const [low, high] = makeHL(i, coronaTestInput[i]);
      x.update(coronaTestInput[i], low, high, new Date(time.getTime() + i * 60_000));
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(11);
  });

  it('should match reference snr / vmin / vmax snapshots', () => {
    const snapshots: Array<[number, number, number, number]> = [
      [11, 1.0000000000, 0.0000000000, 20.0000000000],
      [12, 1.0000000000, 0.0000000000, 20.0000000000],
      [50, 1.0000000000, 0.0000000000, 20.0000000000],
      [100, 2.9986583538, 4.2011609652, 20.0000000000],
      [150, 1.0000000000, 0.0000000035, 20.0000000000],
      [200, 1.0000000000, 0.0000000000, 20.0000000000],
      [251, 1.0000000000, 0.0000000026, 20.0000000000],
    ];

    const x = CoronaSignalToNoiseRatio.default();
    let si = 0;
    for (let i = 0; i < coronaTestInput.length; i++) {
      const t = new Date(time.getTime() + i * 60_000);
      const [low, high] = makeHL(i, coronaTestInput[i]);
      const [h, snr] = x.update(coronaTestInput[i], low, high, t);

      expect(h.parameterFirst).toBe(1);
      expect(h.parameterLast).toBe(11);
      expect(Math.abs(h.parameterResolution - 4.9)).toBeLessThan(1e-9);

      if (!x.isPrimed()) {
        expect(h.isEmpty()).withContext(`bar ${i}`).toBe(true);
        expect(Number.isNaN(snr)).toBe(true);
        continue;
      }

      expect(h.values.length).toBe(50);

      if (si < snapshots.length && snapshots[si][0] === i) {
        expect(Math.abs(snapshots[si][1] - snr))
          .withContext(`snr[${i}]: expected ${snapshots[si][1]}, actual ${snr}`)
          .toBeLessThan(tolerance);
        expect(Math.abs(snapshots[si][2] - h.valueMin))
          .withContext(`vmin[${i}]: expected ${snapshots[si][2]}, actual ${h.valueMin}`)
          .toBeLessThan(tolerance);
        expect(Math.abs(snapshots[si][3] - h.valueMax))
          .withContext(`vmax[${i}]: expected ${snapshots[si][3]}, actual ${h.valueMax}`)
          .toBeLessThan(tolerance);
        si++;
      }
    }
    expect(si).toBe(snapshots.length);
  });

  const primeCount = 50;
  const value = 100.0;

  function prime(x: CoronaSignalToNoiseRatio): void {
    for (let i = 0; i < primeCount; i++) {
      const s = coronaTestInput[i % coronaTestInput.length];
      const [low, high] = makeHL(i, s);
      x.update(s, low, high, time);
    }
  }

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(2);
    expect(out[0] instanceof Heatmap).toBe(true);
    expect((out[0] as Heatmap).time).toEqual(time);
    const s = out[1] as Scalar;
    expect(s.time).toEqual(time);
  }

  it('should produce 2-element output via updateScalar', () => {
    const x = CoronaSignalToNoiseRatio.default();
    prime(x);
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 2-element output via updateBar', () => {
    const x = CoronaSignalToNoiseRatio.default();
    prime(x);
    const bar = new Bar({ time, open: value, high: value * 1.005, low: value * 0.995, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 2-element output via updateQuote', () => {
    const x = CoronaSignalToNoiseRatio.default();
    prime(x);
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 2-element output via updateTrade', () => {
    const x = CoronaSignalToNoiseRatio.default();
    prime(x);
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
