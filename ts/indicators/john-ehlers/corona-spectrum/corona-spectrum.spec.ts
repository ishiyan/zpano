import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { coronaTestInput } from '../corona/test-input';
import { CoronaSpectrum } from './corona-spectrum';
import { CoronaSpectrumOutput } from './output';

const tolerance = 1e-4;

describe('CoronaSpectrum', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(CoronaSpectrumOutput.Value).toBe(0);
    expect(CoronaSpectrumOutput.DominantCycle).toBe(1);
    expect(CoronaSpectrumOutput.DominantCycleMedian).toBe(2);
  });

  it('should return expected mnemonic for default params', () => {
    const x = CoronaSpectrum.default();
    expect(x.metadata().mnemonic).toBe('cspect(6, 20, 6, 30, 30, hl/2)');
  });

  it('should round custom param ranges (ceil min, floor max)', () => {
    const x = CoronaSpectrum.fromParams({
      minRasterValue: 4,
      maxRasterValue: 25,
      minParameterValue: 8.7,   // ceils to 9
      maxParameterValue: 40.4,  // floors to 40
      highPassFilterCutoff: 20,
    });
    expect(x.metadata().mnemonic).toBe('cspect(4, 25, 9, 40, 20, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = CoronaSpectrum.default();
    const meta = x.metadata();
    const mn = 'cspect(6, 20, 6, 30, 30, hl/2)';
    const mnDC = 'cspect-dc(30, hl/2)';
    const mnDCM = 'cspect-dcm(30, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.CoronaSpectrum);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Corona spectrum ' + mn);
    expect(meta.outputs.length).toBe(3);

    expect(meta.outputs[0].kind).toBe(CoronaSpectrumOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[0].mnemonic).toBe(mn);

    expect(meta.outputs[1].kind).toBe(CoronaSpectrumOutput.DominantCycle);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnDC);
    expect(meta.outputs[1].description).toBe('Corona spectrum dominant cycle ' + mnDC);

    expect(meta.outputs[2].kind).toBe(CoronaSpectrumOutput.DominantCycleMedian);
    expect(meta.outputs[2].shape).toBe(Shape.Scalar);
    expect(meta.outputs[2].mnemonic).toBe(mnDCM);
    expect(meta.outputs[2].description).toBe('Corona spectrum dominant cycle median ' + mnDCM);
  });

  it('should throw for invalid params', () => {
    expect(() => CoronaSpectrum.fromParams({ minRasterValue: -1 })).toThrowError(/MinRasterValue/);
    expect(() => CoronaSpectrum.fromParams({ minRasterValue: 10, maxRasterValue: 10 })).toThrowError(/MaxRasterValue/);
    expect(() => CoronaSpectrum.fromParams({ minParameterValue: 1 })).toThrowError(/MinParameterValue/);
    expect(() => CoronaSpectrum.fromParams({ minParameterValue: 20, maxParameterValue: 20 })).toThrowError(/MaxParameterValue/);
    expect(() => CoronaSpectrum.fromParams({ highPassFilterCutoff: 1 })).toThrowError(/HighPassFilterCutoff/);
  });

  it('should return empty heatmap and NaN scalars for NaN input', () => {
    const x = CoronaSpectrum.default();
    const [h, dc, dcm] = x.update(Number.NaN, time);
    expect(h instanceof Heatmap).toBe(true);
    expect(h.isEmpty()).toBe(true);
    expect(h.parameterFirst).toBe(6);
    expect(h.parameterLast).toBe(30);
    expect(h.parameterResolution).toBe(2);
    expect(Number.isNaN(dc)).toBe(true);
    expect(Number.isNaN(dcm)).toBe(true);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 11 (MinimalPeriodTimesTwo = 12)', () => {
    const x = CoronaSpectrum.default();
    expect(x.isPrimed()).toBe(false);
    let primedAt = -1;
    for (let i = 0; i < coronaTestInput.length; i++) {
      x.update(coronaTestInput[i], new Date(time.getTime() + i * 60_000));
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(11);
  });

  it('should match reference dc / dcm snapshots', () => {
    const snapshots: Array<[number, number, number]> = [
      [11, 17.7604672565, 17.7604672565],
      [12, 6.0000000000, 6.0000000000],
      [50, 15.9989078712, 15.9989078712],
      [100, 14.7455497547, 14.7455497547],
      [150, 17.5000000000, 17.2826036069],
      [200, 19.7557338512, 20.0000000000],
      [251, 6.0000000000, 6.0000000000],
    ];

    const x = CoronaSpectrum.default();
    let si = 0;
    for (let i = 0; i < coronaTestInput.length; i++) {
      const t = new Date(time.getTime() + i * 60_000);
      const [h, dc, dcm] = x.update(coronaTestInput[i], t);

      expect(h.parameterFirst).toBe(6);
      expect(h.parameterLast).toBe(30);
      expect(h.parameterResolution).toBe(2);

      if (!x.isPrimed()) {
        expect(h.isEmpty()).withContext(`bar ${i}`).toBe(true);
        expect(Number.isNaN(dc)).toBe(true);
        expect(Number.isNaN(dcm)).toBe(true);
        continue;
      }

      expect(h.values.length).toBe(49);

      if (si < snapshots.length && snapshots[si][0] === i) {
        expect(Math.abs(snapshots[si][1] - dc))
          .withContext(`dc[${i}]: expected ${snapshots[si][1]}, actual ${dc}`)
          .toBeLessThan(tolerance);
        expect(Math.abs(snapshots[si][2] - dcm))
          .withContext(`dcm[${i}]: expected ${snapshots[si][2]}, actual ${dcm}`)
          .toBeLessThan(tolerance);
        si++;
      }
    }
    expect(si).toBe(snapshots.length);
  });

  const primeCount = 50;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(3);
    expect(out[0] instanceof Heatmap).toBe(true);
    expect((out[0] as Heatmap).time).toEqual(time);
    for (let i = 1; i < 3; i++) {
      const s = out[i] as Scalar;
      expect(s.time).toEqual(time);
    }
  }

  it('should produce 3-element output via updateScalar', () => {
    const x = CoronaSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(coronaTestInput[i % coronaTestInput.length], time);
    }
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 3-element output via updateBar', () => {
    const x = CoronaSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(coronaTestInput[i % coronaTestInput.length], time);
    }
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 3-element output via updateQuote', () => {
    const x = CoronaSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(coronaTestInput[i % coronaTestInput.length], time);
    }
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 3-element output via updateTrade', () => {
    const x = CoronaSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(coronaTestInput[i % coronaTestInput.length], time);
    }
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
