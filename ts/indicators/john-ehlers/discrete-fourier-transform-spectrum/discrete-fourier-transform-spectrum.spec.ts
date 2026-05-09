import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { DiscreteFourierTransformSpectrum } from './discrete-fourier-transform-spectrum';
import { DiscreteFourierTransformSpectrumOutput } from './output';
import { testInput, snapshots } from './testdata';
import type { Snapshot, Spot } from './testdata';

const tolerance = 1e-12;
const minMaxTolerance = 1e-10;

// 252-entry TA-Lib MAMA reference series. Mirrors the Go test input.
// Snapshots mirror the Go implementation's reference values.
describe('DiscreteFourierTransformSpectrum', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(DiscreteFourierTransformSpectrumOutput.Value).toBe(0);
  });

  it('should return expected mnemonic for default params', () => {
    const x = DiscreteFourierTransformSpectrum.default();
    expect(x.metadata().mnemonic).toBe('dftps(48, 10, 48, 1, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = DiscreteFourierTransformSpectrum.default();
    const meta = x.metadata();
    const mn = 'dftps(48, 10, 48, 1, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.DiscreteFourierTransformSpectrum);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Discrete Fourier transform spectrum ' + mn);
    expect(meta.outputs.length).toBe(1);

    expect(meta.outputs[0].kind).toBe(DiscreteFourierTransformSpectrumOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Discrete Fourier transform spectrum ' + mn);
  });

  it('should return expected mnemonics for flag overrides', () => {
    const cases: Array<[any, string]> = [
      [{}, 'dftps(48, 10, 48, 1, hl/2)'],
      [{ disableSpectralDilationCompensation: true }, 'dftps(48, 10, 48, 1, no-sdc, hl/2)'],
      [{ disableAutomaticGainControl: true }, 'dftps(48, 10, 48, 1, no-agc, hl/2)'],
      [{ automaticGainControlDecayFactor: 0.8 }, 'dftps(48, 10, 48, 1, agc=0.8, hl/2)'],
      [{ fixedNormalization: true }, 'dftps(48, 10, 48, 1, no-fn, hl/2)'],
      [
        {
          disableSpectralDilationCompensation: true,
          disableAutomaticGainControl: true,
          fixedNormalization: true,
        },
        'dftps(48, 10, 48, 1, no-sdc, no-agc, no-fn, hl/2)',
      ],
    ];
    for (const [p, mn] of cases) {
      const x = DiscreteFourierTransformSpectrum.fromParams(p);
      expect(x.metadata().mnemonic).withContext(`params=${JSON.stringify(p)}`).toBe(mn);
    }
  });

  it('should throw for invalid params', () => {
    expect(() => DiscreteFourierTransformSpectrum.fromParams({ length: 1 })).toThrowError(/Length should be >= 2/);
    expect(() => DiscreteFourierTransformSpectrum.fromParams({ minPeriod: 1 })).toThrowError(/MinPeriod should be >= 2/);
    expect(() => DiscreteFourierTransformSpectrum.fromParams({ minPeriod: 10, maxPeriod: 10 })).toThrowError(/MaxPeriod should be > MinPeriod/);
    expect(() => DiscreteFourierTransformSpectrum.fromParams({ length: 10, maxPeriod: 48 })).toThrowError(/MaxPeriod should be <= 2 \* Length/);
    expect(() => DiscreteFourierTransformSpectrum.fromParams({ automaticGainControlDecayFactor: -0.1 })).toThrowError(/AutomaticGainControlDecayFactor/);
    expect(() => DiscreteFourierTransformSpectrum.fromParams({ automaticGainControlDecayFactor: 1.0 })).toThrowError(/AutomaticGainControlDecayFactor/);
  });

  it('should return empty heatmap for NaN input and not prime', () => {
    const x = DiscreteFourierTransformSpectrum.default();
    const h = x.update(Number.NaN, time);
    expect(h instanceof Heatmap).toBe(true);
    expect(h.isEmpty()).toBe(true);
    expect(h.parameterFirst).toBe(10);
    expect(h.parameterLast).toBe(48);
    expect(h.parameterResolution).toBe(1);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 47 (length=48)', () => {
    const x = DiscreteFourierTransformSpectrum.default();
    expect(x.isPrimed()).toBe(false);
    let primedAt = -1;
    for (let i = 0; i < testInput.length; i++) {
      x.update(testInput[i], new Date(time.getTime() + i * 60_000));
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(47);
  });

  it('should match reference snapshots', () => {
    const x = DiscreteFourierTransformSpectrum.default();
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
    // 3 integer cycles in default length=48 window (no DFT leakage).
    const period = 16;
    const bars = 200;

    // Disable AGC/SDC/FloatingNormalization so the peak reflects the raw DFT magnitude.
    const x = DiscreteFourierTransformSpectrum.fromParams({
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
    const x = DiscreteFourierTransformSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 1-element output via updateBar', () => {
    const x = DiscreteFourierTransformSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 1-element output via updateQuote', () => {
    const x = DiscreteFourierTransformSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 1-element output via updateTrade', () => {
    const x = DiscreteFourierTransformSpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
