import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { CombBandPassSpectrum } from './comb-band-pass-spectrum';
import { CombBandPassSpectrumOutput } from './output';
import { testInput, snapshots } from './testdata';
import type { Snapshot, Spot } from './testdata';

const tolerance = 1e-12;
const minMaxTolerance = 1e-10;

// 252-entry TA-Lib MAMA reference series. Mirrors the Go test input.
// Snapshots mirror the Go implementation's reference values.
describe('CombBandPassSpectrum', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(CombBandPassSpectrumOutput.Value).toBe(0);
  });

  it('should return expected mnemonic for default params', () => {
    const x = CombBandPassSpectrum.default();
    expect(x.metadata().mnemonic).toBe('cbps(10, 48, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = CombBandPassSpectrum.default();
    const meta = x.metadata();
    const mn = 'cbps(10, 48, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.CombBandPassSpectrum);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Comb band-pass spectrum ' + mn);
    expect(meta.outputs.length).toBe(1);

    expect(meta.outputs[0].kind).toBe(CombBandPassSpectrumOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Comb band-pass spectrum ' + mn);
  });

  it('should return expected mnemonics for flag overrides', () => {
    const cases: Array<[any, string]> = [
      [{}, 'cbps(10, 48, hl/2)'],
      [{ bandwidth: 0.5 }, 'cbps(10, 48, bw=0.5, hl/2)'],
      [{ disableSpectralDilationCompensation: true }, 'cbps(10, 48, no-sdc, hl/2)'],
      [{ disableAutomaticGainControl: true }, 'cbps(10, 48, no-agc, hl/2)'],
      [{ automaticGainControlDecayFactor: 0.8 }, 'cbps(10, 48, agc=0.8, hl/2)'],
      [{ fixedNormalization: true }, 'cbps(10, 48, no-fn, hl/2)'],
      [
        {
          bandwidth: 0.5,
          disableSpectralDilationCompensation: true,
          disableAutomaticGainControl: true,
          fixedNormalization: true,
        },
        'cbps(10, 48, bw=0.5, no-sdc, no-agc, no-fn, hl/2)',
      ],
    ];
    for (const [p, mn] of cases) {
      const x = CombBandPassSpectrum.fromParams(p);
      expect(x.metadata().mnemonic).withContext(`params=${JSON.stringify(p)}`).toBe(mn);
    }
  });

  it('should throw for invalid params', () => {
    expect(() => CombBandPassSpectrum.fromParams({ minPeriod: 1 })).toThrowError(/MinPeriod should be >= 2/);
    expect(() => CombBandPassSpectrum.fromParams({ minPeriod: 10, maxPeriod: 10 })).toThrowError(/MaxPeriod should be > MinPeriod/);
    expect(() => CombBandPassSpectrum.fromParams({ bandwidth: -0.1 })).toThrowError(/Bandwidth should be in \(0, 1\)/);
    expect(() => CombBandPassSpectrum.fromParams({ bandwidth: 1.0 })).toThrowError(/Bandwidth should be in \(0, 1\)/);
    expect(() => CombBandPassSpectrum.fromParams({ automaticGainControlDecayFactor: -0.1 })).toThrowError(/AutomaticGainControlDecayFactor/);
    expect(() => CombBandPassSpectrum.fromParams({ automaticGainControlDecayFactor: 1.0 })).toThrowError(/AutomaticGainControlDecayFactor/);
  });

  it('should return empty heatmap for NaN input and not prime', () => {
    const x = CombBandPassSpectrum.default();
    const h = x.update(Number.NaN, time);
    expect(h instanceof Heatmap).toBe(true);
    expect(h.isEmpty()).toBe(true);
    expect(h.parameterFirst).toBe(10);
    expect(h.parameterLast).toBe(48);
    expect(h.parameterResolution).toBe(1);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 47 (maxPeriod=48)', () => {
    const x = CombBandPassSpectrum.default();
    expect(x.isPrimed()).toBe(false);
    let primedAt = -1;
    for (let i = 0; i < testInput.length; i++) {
      x.update(testInput[i], new Date(time.getTime() + i * 60_000));
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(47);
  });

  it('should match reference snapshots', () => {
    const x = CombBandPassSpectrum.default();
    let si = 0;
    for (let i = 0; i < testInput.length; i++) {
      const t = new Date(time.getTime() + i * 60_000);
      const h = x.update(testInput[i], t);

      expect(h.parameterFirst).toBe(10);
      expect(h.parameterLast).toBe(48);
      expect(h.parameterResolution).toBe(1);

      if (!x.isPrimed()) {
        expect(h.isEmpty()).withContext(`bar ${i}`).toBe(true);
        continue;
      }

      expect(h.values.length).toBe(39);

      if (si < snapshots.length && snapshots[si].i === i) {
        const snap = snapshots[si];
        expect(Math.abs(h.valueMin - snap.valueMin))
          .withContext(`valueMin[${i}]: expected ${snap.valueMin}, actual ${h.valueMin}`)
          .toBeLessThan(minMaxTolerance);
        expect(Math.abs(h.valueMax - snap.valueMax))
          .withContext(`valueMax[${i}]: expected ${snap.valueMax}, actual ${h.valueMax}`)
          .toBeLessThan(minMaxTolerance);
        for (const [index, v] of snap.spots) {
          expect(Math.abs(h.values[index] - v))
            .withContext(`values[${i}][${index}]: expected ${v}, actual ${h.values[index]}`)
            .toBeLessThan(tolerance);
        }
        si++;
      }
    }
    expect(si).toBe(snapshots.length);
  });

  it('should place peak bin at injected period for synthetic sine', () => {
    const period = 20;
    const bars = 400;

    // Disable AGC/SDC/FloatingNormalization so the peak reflects raw BP power.
    const x = CombBandPassSpectrum.fromParams({
      disableSpectralDilationCompensation: true,
      disableAutomaticGainControl: true,
      fixedNormalization: true,
    });

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

  const primeCount = 60;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(1);
    expect(out[0] instanceof Heatmap).toBe(true);
    expect((out[0] as Heatmap).time).toEqual(time);
  }

  it('should produce 1-element output via updateScalar', () => {
    const x = CombBandPassSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 1-element output via updateBar', () => {
    const x = CombBandPassSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 1-element output via updateQuote', () => {
    const x = CombBandPassSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 1-element output via updateTrade', () => {
    const x = CombBandPassSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
