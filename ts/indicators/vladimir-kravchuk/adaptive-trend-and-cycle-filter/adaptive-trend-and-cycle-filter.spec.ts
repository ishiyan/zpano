import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { AdaptiveTrendAndCycleFilter } from './adaptive-trend-and-cycle-filter';
import { AdaptiveTrendAndCycleFilterOutput } from './output';
import { atcfTestInput, N, atcfSnapshots } from './testdata';
import type { AtcfSnap } from './testdata';

const tolerance = 1e-10;

// 252-entry TA-Lib MAMA reference series (Price D5..D256), mirroring the Go
// test `testATCFInput` in adaptivetrendandcyclefilter_test.go.
// Snapshot tuple: [i, fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci].
// NaN positions are represented by Number.NaN and matched via Number.isNaN.
// Values mirror Go `testATCFSnapshots` in adaptivetrendandcyclefilter_test.go.
function closeEnough(expected: number, actual: number): boolean {
  if (Number.isNaN(expected)) return Number.isNaN(actual);
  return Math.abs(expected - actual) <= tolerance;
}

describe('AdaptiveTrendAndCycleFilter', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(AdaptiveTrendAndCycleFilterOutput.Fatl).toBe(0);
    expect(AdaptiveTrendAndCycleFilterOutput.Satl).toBe(1);
    expect(AdaptiveTrendAndCycleFilterOutput.Rftl).toBe(2);
    expect(AdaptiveTrendAndCycleFilterOutput.Rstl).toBe(3);
    expect(AdaptiveTrendAndCycleFilterOutput.Rbci).toBe(4);
    expect(AdaptiveTrendAndCycleFilterOutput.Ftlm).toBe(5);
    expect(AdaptiveTrendAndCycleFilterOutput.Stlm).toBe(6);
    expect(AdaptiveTrendAndCycleFilterOutput.Pcci).toBe(7);
  });

  it('should return expected mnemonic for default params', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    expect(x.metadata().mnemonic).toBe('atcf()');
  });

  it('should return expected metadata', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    const meta = x.metadata();
    const mn = 'atcf()';

    expect(meta.identifier).toBe(IndicatorIdentifier.AdaptiveTrendAndCycleFilter);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Adaptive trend and cycle filter ' + mn);
    expect(meta.outputs.length).toBe(8);

    const expected: Array<[AdaptiveTrendAndCycleFilterOutput, string]> = [
      [AdaptiveTrendAndCycleFilterOutput.Fatl, 'fatl()'],
      [AdaptiveTrendAndCycleFilterOutput.Satl, 'satl()'],
      [AdaptiveTrendAndCycleFilterOutput.Rftl, 'rftl()'],
      [AdaptiveTrendAndCycleFilterOutput.Rstl, 'rstl()'],
      [AdaptiveTrendAndCycleFilterOutput.Rbci, 'rbci()'],
      [AdaptiveTrendAndCycleFilterOutput.Ftlm, 'ftlm()'],
      [AdaptiveTrendAndCycleFilterOutput.Stlm, 'stlm()'],
      [AdaptiveTrendAndCycleFilterOutput.Pcci, 'pcci()'],
    ];

    for (let i = 0; i < expected.length; i++) {
      expect(meta.outputs[i].kind).toBe(expected[i][0]);
      expect(meta.outputs[i].shape).toBe(Shape.Scalar);
      expect(meta.outputs[i].mnemonic).toBe(expected[i][1]);
    }
  });

  it('should return all NaN for NaN input and remain unprimed', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    const out = x.update(Number.NaN);
    for (const v of out) expect(Number.isNaN(v)).toBe(true);
    expect(x.isPrimed()).toBe(false);
  });

  it('should prime at sample index 90 (RSTL is the longest pole, 91 taps)', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    expect(x.isPrimed()).toBe(false);

    let primedAt = -1;
    for (let i = 0; i < atcfTestInput.length; i++) {
      x.update(atcfTestInput[i]);
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBe(90);
  });

  it('should match locked-in snapshots across 252 samples', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    let si = 0;

    for (let i = 0; i < atcfTestInput.length; i++) {
      const out = x.update(atcfTestInput[i]);

      if (si < atcfSnapshots.length && atcfSnapshots[si][0] === i) {
        const s = atcfSnapshots[si];
        const labels = ['fatl', 'satl', 'rftl', 'rstl', 'rbci', 'ftlm', 'stlm', 'pcci'];
        for (let k = 0; k < 8; k++) {
          expect(closeEnough(s[k + 1], out[k]))
            .withContext(`[${i}] ${labels[k]}: expected ${s[k + 1]}, actual ${out[k]}`)
            .toBe(true);
        }
        si++;
      }
    }
    expect(si).toBe(atcfSnapshots.length);
  });

  const primeCount = 100;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(8);
    for (let i = 0; i < 8; i++) {
      const s = out[i] as Scalar;
      expect(s.time).toEqual(time);
    }
  }

  it('should produce 8-element output via updateScalar', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    for (let i = 0; i < primeCount; i++) x.update(atcfTestInput[i]);
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 8-element output via updateBar', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    for (let i = 0; i < primeCount; i++) x.update(atcfTestInput[i]);
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 8-element output via updateQuote', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    for (let i = 0; i < primeCount; i++) x.update(atcfTestInput[i]);
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 8-element output via updateTrade', () => {
    const x = AdaptiveTrendAndCycleFilter.default();
    for (let i = 0; i < primeCount; i++) x.update(atcfTestInput[i]);
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
