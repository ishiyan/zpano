import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { GoertzelSpectrum } from './goertzel-spectrum';
import { GoertzelSpectrumOutput } from './output';
import { testInput, snapshots } from './testdata';
import type { Snapshot, Spot } from './testdata';

const tolerance = 1e-10;
const minMaxTolerance = 1e-9;

// 252-entry TA-Lib MAMA reference series (Price D5..D256). Mirrors the Go test input.
// Snapshots were captured from the Go implementation and hand-verified at i=63 against an
// independent Python implementation of the Goertzel spectrum (match better than 1e-14).
describe('GoertzelSpectrum', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(GoertzelSpectrumOutput.Value).toBe(0);
  });

  it('should return expected mnemonic for default params', () => {
    const x = GoertzelSpectrum.default();
    expect(x.metadata().mnemonic).toBe('gspect(64, 2, 64, 1, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = GoertzelSpectrum.default();
    const meta = x.metadata();
    const mn = 'gspect(64, 2, 64, 1, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.GoertzelSpectrum);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Goertzel spectrum ' + mn);
    expect(meta.outputs.length).toBe(1);

    expect(meta.outputs[0].kind).toBe(GoertzelSpectrumOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Goertzel spectrum ' + mn);
  });

  it('should return expected mnemonics for flag overrides', () => {
    const cases: Array<[any, string]> = [
      [{}, 'gspect(64, 2, 64, 1, hl/2)'],
      [{ isFirstOrder: true }, 'gspect(64, 2, 64, 1, fo, hl/2)'],
      [{ disableSpectralDilationCompensation: true }, 'gspect(64, 2, 64, 1, no-sdc, hl/2)'],
      [{ disableAutomaticGainControl: true }, 'gspect(64, 2, 64, 1, no-agc, hl/2)'],
      [{ automaticGainControlDecayFactor: 0.8 }, 'gspect(64, 2, 64, 1, agc=0.8, hl/2)'],
      [{ fixedNormalization: true }, 'gspect(64, 2, 64, 1, no-fn, hl/2)'],
      [
        {
          isFirstOrder: true,
          disableSpectralDilationCompensation: true,
          disableAutomaticGainControl: true,
          fixedNormalization: true,
        },
        'gspect(64, 2, 64, 1, fo, no-sdc, no-agc, no-fn, hl/2)',
      ],
    ];
    for (const [p, mn] of cases) {
      const x = GoertzelSpectrum.fromParams(p);
      expect(x.metadata().mnemonic).withContext(`params=${JSON.stringify(p)}`).toBe(mn);
    }
  });

  it('should throw for invalid params', () => {
    expect(() => GoertzelSpectrum.fromParams({ length: 1 })).toThrowError(/Length/);
    expect(() => GoertzelSpectrum.fromParams({ minPeriod: 1 })).toThrowError(/MinPeriod/);
    expect(() => GoertzelSpectrum.fromParams({ minPeriod: 10, maxPeriod: 10 })).toThrowError(/MaxPeriod should be > MinPeriod/);
    expect(() => GoertzelSpectrum.fromParams({ length: 16, minPeriod: 2, maxPeriod: 64 })).toThrowError(/MaxPeriod should be <= 2 \* Length/);
    expect(() => GoertzelSpectrum.fromParams({ automaticGainControlDecayFactor: -0.1 })).toThrowError(/AutomaticGainControlDecayFactor/);
    expect(() => GoertzelSpectrum.fromParams({ automaticGainControlDecayFactor: 1.0 })).toThrowError(/AutomaticGainControlDecayFactor/);
  });

  it('should return empty heatmap for NaN input and not prime', () => {
    const x = GoertzelSpectrum.default();
    const h = x.update(Number.NaN, time);
    expect(h instanceof Heatmap).toBe(true);
    expect(h.isEmpty()).toBe(true);
    expect(h.parameterFirst).toBe(2);
    expect(h.parameterLast).toBe(64);
    expect(h.parameterResolution).toBe(1);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 63 (length=64)', () => {
    const x = GoertzelSpectrum.default();
    expect(x.isPrimed()).toBe(false);
    let primedAt = -1;
    for (let i = 0; i < testInput.length; i++) {
      x.update(testInput[i], new Date(time.getTime() + i * 60_000));
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(63);
  });

  it('should match reference snapshots', () => {
    const x = GoertzelSpectrum.default();
    let si = 0;
    for (let i = 0; i < testInput.length; i++) {
      const t = new Date(time.getTime() + i * 60_000);
      const h = x.update(testInput[i], t);

      expect(h.parameterFirst).toBe(2);
      expect(h.parameterLast).toBe(64);
      expect(h.parameterResolution).toBe(1);

      if (!x.isPrimed()) {
        expect(h.isEmpty()).withContext(`bar ${i}`).toBe(true);
        continue;
      }

      expect(h.values.length).toBe(63);

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

  const primeCount = 70;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(1);
    expect(out[0] instanceof Heatmap).toBe(true);
    expect((out[0] as Heatmap).time).toEqual(time);
  }

  it('should produce 1-element output via updateScalar', () => {
    const x = GoertzelSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 1-element output via updateBar', () => {
    const x = GoertzelSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 1-element output via updateQuote', () => {
    const x = GoertzelSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 1-element output via updateTrade', () => {
    const x = GoertzelSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
