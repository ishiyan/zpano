import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import {
  testInputFourSinusoids,
  testInputTest1,
  testInputTest2,
  testInputTest3,
} from './data.spec';
import { MaximumEntropySpectrum } from './maximum-entropy-spectrum';
import { MaximumEntropySpectrumEstimator } from './estimator';
import { MaximumEntropySpectrumOutput } from './output';

const tolerance = 1e-12;
const minMaxTolerance = 1e-10;

// 252-entry TA-Lib MAMA reference series (Price D5..D256). Mirrors the Go test input.
const testInput: number[] = [
  92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550, 91.6875,
  94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175, 90.6100, 91.0000,
  88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450, 86.7350, 86.8600, 87.5475,
  85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050, 85.8125, 84.5950, 83.6575, 84.4550,
  83.5000, 86.7825, 88.1725, 89.2650, 90.8600, 90.7825, 91.8600, 90.3600, 89.8600, 90.9225,
  89.5000, 87.6725, 86.5000, 84.2825, 82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325,
  89.5000, 88.0950, 90.6250, 92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775,
  88.2500, 86.9075, 84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
  103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
  120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
  114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
  114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
  124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
  137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
  123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
  122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
  123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
  130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
  127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
  121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
  106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250, 97.5900,
  95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850, 94.6250, 95.1200,
  94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
  103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
  106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
  109.5300, 108.0600,
];

type Spot = [number, number]; // [index, value]

interface Snapshot {
  i: number;
  valueMin: number;
  valueMax: number;
  spots: Spot[];
}

// Snapshots captured from the Go implementation after verifying the Burg AR coefficients
// against Paul Bourke's reference values (via MBST's MaximumEntropySpectrumEstimatorTest).
const snapshots: Snapshot[] = [
  {
    i: 59, valueMin: 0, valueMax: 1,
    spots: [[0, 0.000000000000000], [14, 0.124709393535801], [28, 0.021259483287733],
            [42, 0.726759100473496], [57, 0.260829244402141]],
  },
  {
    i: 60, valueMin: 0, valueMax: 0.3803558166,
    spots: [[0, 0.000000000000000], [14, 0.047532484316402], [28, 0.156007210177695],
            [42, 0.204392941920655], [57, 0.099988829337396]],
  },
  {
    i: 100, valueMin: 0, valueMax: 0.7767627734,
    spots: [[0, 0.000000000000000], [14, 0.005541589459818], [28, 0.019544065000896],
            [42, 0.045342308770863], [57, 0.776762773404885]],
  },
  {
    i: 150, valueMin: 0, valueMax: 0.0126783313,
    spots: [[0, 0.000347619185321], [14, 0.001211800388686], [28, 0.001749939543675],
            [42, 0.010949450171300], [57, 0.001418701588812]],
  },
  {
    i: 200, valueMin: 0, valueMax: 0.5729940203,
    spots: [[0, 0.000000000000000], [14, 0.047607367831419], [28, 0.013304430092822],
            [42, 0.137193402225458], [57, 0.506646287515276]],
  },
];

// Expected AR coefficients from MBST's MaximumEntropySpectrumEstimatorTest.cs (originally
// from http://paulbourke.net/miscellaneous/ar/). Compared at MBST's tolerance: rounded to
// `dec` decimals.
interface CoefCase {
  name: string;
  input: readonly number[];
  degree: number;
  dec: number;
  want: number[];
}

const coefCases: CoefCase[] = [
  { name: 'sinusoids/1', input: testInputFourSinusoids, degree: 1, dec: 1, want: [0.941872] },
  { name: 'sinusoids/2', input: testInputFourSinusoids, degree: 2, dec: 1, want: [1.826156, -0.938849] },
  { name: 'sinusoids/3', input: testInputFourSinusoids, degree: 3, dec: 1, want: [2.753231, -2.740306, 0.985501] },
  { name: 'sinusoids/4', input: testInputFourSinusoids, degree: 4, dec: 1, want: [3.736794, -5.474295, 3.731127, -0.996783] },
  { name: 'test1/5', input: testInputTest1, degree: 5, dec: 1, want: [1.4, -0.7, 0.04, 0.7, -0.5] },
  { name: 'test2/7', input: testInputTest2, degree: 7, dec: 0, want: [0.677, 0.175, 0.297, 0.006, -0.114, -0.083, -0.025] },
  { name: 'test3/2', input: testInputTest3, degree: 2, dec: 1, want: [1.02, -0.53] },
];

function roundDec(v: number, dec: number): number {
  const p = Math.pow(10, dec);
  return Math.round(v * p) / p;
}

describe('MaximumEntropySpectrum', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(MaximumEntropySpectrumOutput.Value).toBe(0);
  });

  it('should return expected mnemonic for default params', () => {
    const x = MaximumEntropySpectrum.default();
    expect(x.metadata().mnemonic).toBe('mespect(60, 30, 2, 59, 1, hl/2)');
  });

  it('should return expected metadata', () => {
    const x = MaximumEntropySpectrum.default();
    const meta = x.metadata();
    const mn = 'mespect(60, 30, 2, 59, 1, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.MaximumEntropySpectrum);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Maximum entropy spectrum ' + mn);
    expect(meta.outputs.length).toBe(1);

    expect(meta.outputs[0].kind).toBe(MaximumEntropySpectrumOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Maximum entropy spectrum ' + mn);
  });

  it('should return expected mnemonics for flag overrides', () => {
    const cases: Array<[any, string]> = [
      [{}, 'mespect(60, 30, 2, 59, 1, hl/2)'],
      [{ disableAutomaticGainControl: true }, 'mespect(60, 30, 2, 59, 1, no-agc, hl/2)'],
      [{ automaticGainControlDecayFactor: 0.8 }, 'mespect(60, 30, 2, 59, 1, agc=0.8, hl/2)'],
      [{ fixedNormalization: true }, 'mespect(60, 30, 2, 59, 1, no-fn, hl/2)'],
      [
        { disableAutomaticGainControl: true, fixedNormalization: true },
        'mespect(60, 30, 2, 59, 1, no-agc, no-fn, hl/2)',
      ],
    ];
    for (const [p, mn] of cases) {
      const x = MaximumEntropySpectrum.fromParams(p);
      expect(x.metadata().mnemonic).withContext(`params=${JSON.stringify(p)}`).toBe(mn);
    }
  });

  it('should throw for invalid params', () => {
    expect(() => MaximumEntropySpectrum.fromParams({ length: 1 })).toThrowError(/Length should be >= 2/);
    expect(() => MaximumEntropySpectrum.fromParams({ length: 4, degree: 4 })).toThrowError(/Degree should be > 0 and < Length/);
    expect(() => MaximumEntropySpectrum.fromParams({ minPeriod: 1 })).toThrowError(/MinPeriod should be >= 2/);
    expect(() => MaximumEntropySpectrum.fromParams({ minPeriod: 10, maxPeriod: 10 })).toThrowError(/MaxPeriod should be > MinPeriod/);
    expect(() => MaximumEntropySpectrum.fromParams({ length: 10, degree: 5, maxPeriod: 59 })).toThrowError(/MaxPeriod should be <= 2 \* Length/);
    expect(() => MaximumEntropySpectrum.fromParams({ automaticGainControlDecayFactor: -0.1 })).toThrowError(/AutomaticGainControlDecayFactor/);
    expect(() => MaximumEntropySpectrum.fromParams({ automaticGainControlDecayFactor: 1.0 })).toThrowError(/AutomaticGainControlDecayFactor/);
  });

  it('should return empty heatmap for NaN input and not prime', () => {
    const x = MaximumEntropySpectrum.default();
    const h = x.update(Number.NaN, time);
    expect(h instanceof Heatmap).toBe(true);
    expect(h.isEmpty()).toBe(true);
    expect(h.parameterFirst).toBe(2);
    expect(h.parameterLast).toBe(59);
    expect(h.parameterResolution).toBe(1);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 59 (length=60)', () => {
    const x = MaximumEntropySpectrum.default();
    expect(x.isPrimed()).toBe(false);
    let primedAt = -1;
    for (let i = 0; i < testInput.length; i++) {
      x.update(testInput[i], new Date(time.getTime() + i * 60_000));
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(59);
  });

  it('should match MBST Burg AR coefficients', () => {
    for (const tc of coefCases) {
      const length = tc.input.length;
      const e = new MaximumEntropySpectrumEstimator(length, tc.degree, 2, length * 2, 1, false, 0.995);
      for (let i = 0; i < length; i++) e.inputSeries[i] = tc.input[i];
      e.calculate();

      expect(e.coefficients.length).withContext(tc.name).toBe(tc.degree);

      for (let i = 0; i < tc.want.length; i++) {
        const got = roundDec(e.coefficients[i], tc.dec);
        const exp = roundDec(tc.want[i], tc.dec);
        expect(got).withContext(`${tc.name} coef[${i}] (raw ${e.coefficients[i]})`).toBe(exp);
      }
    }
  });

  it('should match reference snapshots', () => {
    const x = MaximumEntropySpectrum.default();
    let si = 0;
    for (let i = 0; i < testInput.length; i++) {
      const t = new Date(time.getTime() + i * 60_000);
      const h = x.update(testInput[i], t);

      expect(h.parameterFirst).toBe(2);
      expect(h.parameterLast).toBe(59);
      expect(h.parameterResolution).toBe(1);

      if (!x.isPrimed()) {
        expect(h.isEmpty()).withContext(`bar ${i}`).toBe(true);
        continue;
      }

      expect(h.values.length).toBe(58);

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
    const x = MaximumEntropySpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 1-element output via updateBar', () => {
    const x = MaximumEntropySpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 1-element output via updateQuote', () => {
    const x = MaximumEntropySpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 1-element output via updateTrade', () => {
    const x = MaximumEntropySpectrum.default();
    for (let i = 0; i < primeCount; i++) {
      x.update(testInput[i % testInput.length], time);
    }
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
