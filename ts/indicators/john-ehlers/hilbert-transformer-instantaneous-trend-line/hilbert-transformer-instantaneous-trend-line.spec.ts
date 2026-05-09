import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { HilbertTransformerInstantaneousTrendLine } from './hilbert-transformer-instantaneous-trend-line';
import { HilbertTransformerInstantaneousTrendLineOutput } from './output';
import { input, expectedPeriod, expectedValue } from './testdata';

/* eslint-disable max-len */
// Input data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA.xsl, Price, D5…D256, 252 entries.
// Expected period data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA.xsl, Period Adjustment X5…X256.
// 252 entries, smoothed as AI18=0.33*X18+0.67*AI17.
// Expected instantaneous trend line values taken from MBST InstantaneousTrendLineTest.cs,
// generated using Excel implementation from TA-Lib (http://ta-lib.org/) tests, test_HT.xsl.
// 252 entries.
const tolerance = 1e-4;

describe('HilbertTransformerInstantaneousTrendLine', () => {
  const time = new Date(2021, 3, 1);
  const skip = 9;
  const settleSkip = 177;

  it('should have correct output enum values', () => {
    expect(HilbertTransformerInstantaneousTrendLineOutput.Value).toBe(0);
    expect(HilbertTransformerInstantaneousTrendLineOutput.DominantCyclePeriod).toBe(1);
  });

  it('should return expected mnemonic for default params (BarComponent.Median)', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    expect(x.metadata().mnemonic).toBe('htitl(0.330, 4, 1.000, hl/2)');
  });

  it('should return expected mnemonic for tlsl=2 with Close component', () => {
    const x = HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33,
      trendLineSmoothingLength: 2,
      cyclePartMultiplier: 1.0,
      barComponent: BarComponent.Close,
    });
    expect(x.metadata().mnemonic).toBe('htitl(0.330, 2, 1.000)');
  });

  it('should return expected mnemonic for phase accumulator', () => {
    const x = HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      trendLineSmoothingLength: 3,
      cyclePartMultiplier: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.PhaseAccumulator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
      barComponent: BarComponent.Close,
    });
    expect(x.metadata().mnemonic).toBe('htitl(0.500, 3, 0.500, pa(4, 0.200, 0.200))');
  });

  it('should return expected metadata', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    const meta = x.metadata();
    const mn = 'htitl(0.330, 4, 1.000, hl/2)';
    const mnDCP = 'dcp(0.330, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.HilbertTransformerInstantaneousTrendLine);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Hilbert transformer instantaneous trend line ' + mn);
    expect(meta.outputs.length).toBe(2);

    expect(meta.outputs[0].kind).toBe(HilbertTransformerInstantaneousTrendLineOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Hilbert transformer instantaneous trend line ' + mn);

    expect(meta.outputs[1].kind).toBe(HilbertTransformerInstantaneousTrendLineOutput.DominantCyclePeriod);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnDCP);
    expect(meta.outputs[1].description).toBe('Dominant cycle period ' + mnDCP);
  });

  it('should throw if α is out of range (0, 1]', () => {
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({ alphaEmaPeriodAdditional: 0 })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({ alphaEmaPeriodAdditional: -0.1 })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({ alphaEmaPeriodAdditional: 1.0001 })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({ alphaEmaPeriodAdditional: 0.5 })).not.toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({ alphaEmaPeriodAdditional: 1.0 })).not.toThrow();
  });

  it('should throw if trendLineSmoothingLength is out of range [2, 4]', () => {
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 1,
    })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 5,
    })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 2,
    })).not.toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 3,
    })).not.toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 4,
    })).not.toThrow();
  });

  it('should throw if cyclePartMultiplier is out of range (0, 10]', () => {
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, cyclePartMultiplier: 0,
    })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, cyclePartMultiplier: -1,
    })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, cyclePartMultiplier: 10.0001,
    })).toThrow();
    expect(() => HilbertTransformerInstantaneousTrendLine.fromParams({
      alphaEmaPeriodAdditional: 0.33, cyclePartMultiplier: 10.0,
    })).not.toThrow();
  });

  it('should return NaN pair for NaN input', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    const [v, p] = x.update(Number.NaN);
    expect(Number.isNaN(v)).toBe(true);
    expect(Number.isNaN(p)).toBe(true);
  });

  it('should match reference value (MBST InstantaneousTrendLineTest) past settle window', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    for (let i = skip; i < input.length; i++) {
      const [value] = x.update(input[i]);
      if (Number.isNaN(value) || i < settleSkip) continue;
      if (Number.isNaN(expectedValue[i])) continue;
      expect(Math.abs(expectedValue[i] - value))
        .withContext(`value[${i}]: expected ${expectedValue[i]}, actual ${value}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference period (test_MAMA.xsl) past settle window', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    for (let i = skip; i < input.length; i++) {
      const [, period] = x.update(input[i]);
      if (Number.isNaN(period) || i < settleSkip) continue;
      expect(Math.abs(expectedPeriod[i] - period))
        .withContext(`period[${i}]: expected ${expectedPeriod[i]}, actual ${period}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should become primed within the input sequence', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    expect(x.isPrimed()).toBe(false);

    let primedAt = -1;
    for (let i = 0; i < input.length; i++) {
      x.update(input[i]);
      if (x.isPrimed() && primedAt < 0) {
        primedAt = i;
      }
    }
    expect(primedAt).toBeGreaterThanOrEqual(0);
    expect(x.isPrimed()).toBe(true);
  });

  const primeCount = 200;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(2);
    for (let i = 0; i < 2; i++) {
      const s = out[i] as Scalar;
      expect(s.time).toEqual(time);
    }
  }

  it('should produce 2-element output via updateScalar', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(x.updateScalar(s) as any[]);
  });

  it('should produce 2-element output via updateBar', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(x.updateBar(bar) as any[]);
  });

  it('should produce 2-element output via updateQuote', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(x.updateQuote(q) as any[]);
  });

  it('should produce 2-element output via updateTrade', () => {
    const x = HilbertTransformerInstantaneousTrendLine.default();
    for (let i = 0; i < primeCount; i++) x.update(input[i % input.length]);
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(x.updateTrade(t) as any[]);
  });
});
