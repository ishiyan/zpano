import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { TrendCycleMode } from './trend-cycle-mode';
import { TrendCycleModeOutput } from './output';
import {
    input,
    expectedPeriod,
    expectedPhase,
    expectedSine,
    expectedSineLead,
    expectedITL,
    N,
    expectedValue,
    skip,
    settleSkip,
} from './testdata';

/* eslint-disable max-len */
// Input from MBST TrendCycleModeTest.cs (TA-Lib test_MAMA.xsl, Price). 252 entries.
// Expected tcm value array from MBST. First 63 entries NaN, then 138 ±1 entries (201 total).
const tolerance = 1e-4;
// MBST wraps phase into (-180, 180]; zpano into [0, 360). Compare modulo 360.
function phaseDelta(expected: number, actual: number): number {
  let d = (expected - actual) % 360;
  if (d > 180) d -= 360;
  else if (d < -180) d += 360;
  return d;
}

describe('TrendCycleMode', () => {
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(TrendCycleModeOutput.Value).toBe(0);
    expect(TrendCycleModeOutput.IsTrendMode).toBe(1);
    expect(TrendCycleModeOutput.IsCycleMode).toBe(2);
    expect(TrendCycleModeOutput.InstantaneousTrendLine).toBe(3);
    expect(TrendCycleModeOutput.SineWave).toBe(4);
    expect(TrendCycleModeOutput.SineWaveLead).toBe(5);
    expect(TrendCycleModeOutput.DominantCyclePeriod).toBe(6);
    expect(TrendCycleModeOutput.DominantCyclePhase).toBe(7);
  });

  it('should return expected mnemonic for default params', () => {
    const x = TrendCycleMode.default();
    expect(x.metadata().mnemonic).toBe('tcm(0.330, 4, 1.000, 1.500%, hl/2)');
  });

  it('should return expected mnemonic for phase accumulator, tlsl=3', () => {
    const x = TrendCycleMode.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.PhaseAccumulator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
      trendLineSmoothingLength: 3,
      cyclePartMultiplier: 0.5,
      separationPercentage: 2.0,
    });
    expect(x.metadata().mnemonic).toBe('tcm(0.500, 3, 0.500, 2.000%, pa(4, 0.200, 0.200), hl/2)');
  });

  it('should return expected metadata', () => {
    const x = TrendCycleMode.default();
    const meta = x.metadata();
    const mn = 'tcm(0.330, 4, 1.000, 1.500%, hl/2)';
    const mnTrend = 'tcm-trend(0.330, 4, 1.000, 1.500%, hl/2)';
    const mnCycle = 'tcm-cycle(0.330, 4, 1.000, 1.500%, hl/2)';
    const mnITL = 'tcm-itl(0.330, 4, 1.000, 1.500%, hl/2)';
    const mnSine = 'tcm-sine(0.330, 4, 1.000, 1.500%, hl/2)';
    const mnSineLead = 'tcm-sineLead(0.330, 4, 1.000, 1.500%, hl/2)';
    const mnDCP = 'dcp(0.330, hl/2)';
    const mnDCPha = 'dcph(0.330, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.TrendCycleMode);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Trend versus cycle mode ' + mn);
    expect(meta.outputs.length).toBe(8);

    expect(meta.outputs[0].kind).toBe(TrendCycleModeOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Trend versus cycle mode ' + mn);

    expect(meta.outputs[1].kind).toBe(TrendCycleModeOutput.IsTrendMode);
    expect(meta.outputs[1].mnemonic).toBe(mnTrend);
    expect(meta.outputs[1].description).toBe('Trend versus cycle mode, is-trend flag ' + mnTrend);

    expect(meta.outputs[2].kind).toBe(TrendCycleModeOutput.IsCycleMode);
    expect(meta.outputs[2].mnemonic).toBe(mnCycle);
    expect(meta.outputs[2].description).toBe('Trend versus cycle mode, is-cycle flag ' + mnCycle);

    expect(meta.outputs[3].kind).toBe(TrendCycleModeOutput.InstantaneousTrendLine);
    expect(meta.outputs[3].mnemonic).toBe(mnITL);

    expect(meta.outputs[4].kind).toBe(TrendCycleModeOutput.SineWave);
    expect(meta.outputs[4].mnemonic).toBe(mnSine);

    expect(meta.outputs[5].kind).toBe(TrendCycleModeOutput.SineWaveLead);
    expect(meta.outputs[5].mnemonic).toBe(mnSineLead);

    expect(meta.outputs[6].kind).toBe(TrendCycleModeOutput.DominantCyclePeriod);
    expect(meta.outputs[6].mnemonic).toBe(mnDCP);
    expect(meta.outputs[6].description).toBe('Dominant cycle period ' + mnDCP);

    expect(meta.outputs[7].kind).toBe(TrendCycleModeOutput.DominantCyclePhase);
    expect(meta.outputs[7].mnemonic).toBe(mnDCPha);
    expect(meta.outputs[7].description).toBe('Dominant cycle phase ' + mnDCPha);
  });

  it('should throw if α is out of range (0, 1]', () => {
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0 })).toThrowError(/α for additional smoothing/);
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: -0.1 })).toThrowError(/α for additional smoothing/);
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 1.0001 })).toThrowError(/α for additional smoothing/);
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.5 })).not.toThrow();
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 1.0 })).not.toThrow();
  });

  it('should throw if trend line smoothing length is not 2, 3, or 4', () => {
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 1 }))
      .toThrowError(/trend line smoothing length/);
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 5 }))
      .toThrowError(/trend line smoothing length/);
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 2 })).not.toThrow();
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 3 })).not.toThrow();
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 4 })).not.toThrow();
  });

  it('should throw if cycle part multiplier is out of range (0, 10]', () => {
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, cyclePartMultiplier: 0 }))
      .toThrowError(/cycle part multiplier/);
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, cyclePartMultiplier: 10.00001 }))
      .toThrowError(/cycle part multiplier/);
  });

  it('should throw if separation percentage is out of range (0, 100]', () => {
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, separationPercentage: 0 }))
      .toThrowError(/separation percentage/);
    expect(() => TrendCycleMode.fromParams({ alphaEmaPeriodAdditional: 0.33, separationPercentage: 100.00001 }))
      .toThrowError(/separation percentage/);
  });

  it('should return NaN 8-tuple for NaN input', () => {
    const x = TrendCycleMode.default();
    const out = x.update(Number.NaN);
    expect(out.length).toBe(8);
    for (const v of out) {
      expect(Number.isNaN(v)).toBe(true);
    }
  });

  it('should match reference period past settle window', () => {
    const x = TrendCycleMode.default();
    for (let i = skip; i < input.length; i++) {
      const [, , , , , , period] = x.update(input[i]);
      if (Number.isNaN(period) || i < settleSkip) continue;
      expect(Math.abs(expectedPeriod[i] - period))
        .withContext(`period[${i}]: expected ${expectedPeriod[i]}, actual ${period}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference phase (mod 360) past settle window', () => {
    const x = TrendCycleMode.default();
    for (let i = skip; i < input.length; i++) {
      const [, , , , , , , phase] = x.update(input[i]);
      if (Number.isNaN(phase) || Number.isNaN(expectedPhase[i]) || i < settleSkip) continue;
      const d = phaseDelta(expectedPhase[i], phase);
      expect(Math.abs(d))
        .withContext(`phase[${i}]: expected ${expectedPhase[i]}, actual ${phase}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference sine wave past settle window', () => {
    const x = TrendCycleMode.default();
    for (let i = skip; i < input.length; i++) {
      const [, , , , sine] = x.update(input[i]);
      if (Number.isNaN(sine) || Number.isNaN(expectedSine[i]) || i < settleSkip) continue;
      expect(Math.abs(expectedSine[i] - sine))
        .withContext(`sine[${i}]: expected ${expectedSine[i]}, actual ${sine}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference sine wave lead past settle window', () => {
    const x = TrendCycleMode.default();
    for (let i = skip; i < input.length; i++) {
      const [, , , , , sineLead] = x.update(input[i]);
      if (Number.isNaN(sineLead) || Number.isNaN(expectedSineLead[i]) || i < settleSkip) continue;
      expect(Math.abs(expectedSineLead[i] - sineLead))
        .withContext(`sineLead[${i}]: expected ${expectedSineLead[i]}, actual ${sineLead}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference instantaneous trend line past settle window', () => {
    const x = TrendCycleMode.default();
    for (let i = skip; i < input.length; i++) {
      const [, , , itl] = x.update(input[i]);
      if (Number.isNaN(itl) || Number.isNaN(expectedITL[i]) || i < settleSkip) continue;
      expect(Math.abs(expectedITL[i] - itl))
        .withContext(`itl[${i}]: expected ${expectedITL[i]}, actual ${itl}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference tcm value (except MBST known mismatches)', () => {
    const x = TrendCycleMode.default();
    const limit = expectedValue.length;
    for (let i = skip; i < input.length; i++) {
      const [value] = x.update(input[i]);
      if (i >= limit) continue;
      // MBST known mismatches.
      if (i === 70 || i === 71) continue;
      if (Number.isNaN(value) || Number.isNaN(expectedValue[i])) continue;
      expect(Math.abs(expectedValue[i] - value))
        .withContext(`value[${i}]: expected ${expectedValue[i]}, actual ${value}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should have complementary is-trend / is-cycle flags (0/1) aligned with value sign', () => {
    const x = TrendCycleMode.default();
    for (let i = skip; i < input.length; i++) {
      const [value, trend, cycle] = x.update(input[i]);
      if (Number.isNaN(value)) continue;
      expect(trend + cycle).toBe(1);
      if (value > 0) expect(trend).toBe(1);
      if (value < 0) expect(trend).toBe(0);
    }
  });

  it('should become primed within the input sequence', () => {
    const x = TrendCycleMode.default();
    expect(x.isPrimed()).toBe(false);

    let primedAt = -1;
    for (let i = 0; i < input.length; i++) {
      x.update(input[i]);
      if (x.isPrimed() && primedAt < 0) primedAt = i;
    }
    expect(primedAt).toBeGreaterThanOrEqual(0);
    expect(x.isPrimed()).toBe(true);
  });

  const primeCount = 200;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(8);
    for (let i = 0; i < 8; i++) {
      const s = out[i] as Scalar;
      expect(s.time).toEqual(time);
    }
  }

  it('should produce 8-element scalar output via updateScalar', () => {
    const x = TrendCycleMode.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 8-element scalar output via updateBar', () => {
    const x = TrendCycleMode.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 8-element scalar output via updateQuote', () => {
    const x = TrendCycleMode.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 8-element scalar output via updateTrade', () => {
    const x = TrendCycleMode.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
