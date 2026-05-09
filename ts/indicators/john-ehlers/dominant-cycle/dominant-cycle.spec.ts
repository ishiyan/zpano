import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent } from '../../../entities/trade-component';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { DominantCycle } from './dominant-cycle';
import { DominantCycleOutput } from './output';
import { input, expectedPeriod, expectedPhase } from './testdata';

/* eslint-disable max-len */
// Input data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA.xsl, Price, D5…D256, 252 entries.

// Expected period data taken from TA-Lib test_MAMA.xsl, Period Adjustment X5…X256 smoothed as AI18=0.33*X18+0.67*AI17. 252 entries.
// Expected phase data generated using Excel implementation from TA-Lib test_HT.xsl. 252 entries.
// Tolerance used to compare floating-point values. See Go test for rationale.
const tolerance = 1e-4;

// phaseDiff returns the shortest signed angular difference between two angles, in (-180, 180].
// Needed because the Excel reference (test_HT.xsl) accumulates phase values outside the
// (-90°, 360°] range produced by the MBST/port implementation.
function phaseDiff(a: number, b: number): number {
  let d = (a - b) % 360;
  if (d > 180) {
    d -= 360;
  } else if (d <= -180) {
    d += 360;
  }
  return d;
}

describe('DominantCycle', () => {
  const time = new Date(2021, 3, 1);
  // TradeStation convention: skip first 9 bars.
  const skip = 9;
  // Samples required for the EMA to converge past the structural mismatch between the MBST
  // port (seeded from first primed htce.period at index 99) and the Excel reference (smooths
  // from index 0 through zeros). Empirically 1e-4 agreement is reached by i=177.
  const settleSkip = 177;

  it('should have correct output enum values', () => {
    expect(DominantCycleOutput.RawPeriod).toBe(0);
    expect(DominantCycleOutput.Period).toBe(1);
    expect(DominantCycleOutput.Phase).toBe(2);
  });

  it('should return expected mnemonic for default params', () => {
    const dc = DominantCycle.default();
    expect(dc.metadata().mnemonic).toBe('dcp(0.330)');
  });

  it('should return expected mnemonic for explicit estimator types', () => {
    let dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.500)');

    dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 3, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.500, hd(3, 0.200, 0.200))');

    dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminatorUnrolled,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.500, hdu(4, 0.200, 0.200))');

    dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.PhaseAccumulator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.500, pa(4, 0.200, 0.200))');

    dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.DualDifferentiator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.500, dd(4, 0.200, 0.200))');
  });

  it('should return expected mnemonic with non-default components', () => {
    let dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.33,
      barComponent: BarComponent.Median,
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.330, hl/2)');

    dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.33,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.330, b)');

    dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: 0.33,
      tradeComponent: TradeComponent.Volume,
    });
    expect(dc.metadata().mnemonic).toBe('dcp(0.330, v)');
  });

  it('should return expected metadata', () => {
    const dc = DominantCycle.default();
    const meta = dc.metadata();
    const mn = 'dcp(0.330)';
    const mnRaw = 'dcp-raw(0.330)';
    const mnPha = 'dcph(0.330)';

    expect(meta.identifier).toBe(IndicatorIdentifier.DominantCycle);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Dominant cycle period ' + mn);
    expect(meta.outputs.length).toBe(3);

    expect(meta.outputs[0].kind).toBe(DominantCycleOutput.RawPeriod);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mnRaw);
    expect(meta.outputs[0].description).toBe('Dominant cycle raw period ' + mnRaw);

    expect(meta.outputs[1].kind).toBe(DominantCycleOutput.Period);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mn);
    expect(meta.outputs[1].description).toBe('Dominant cycle period ' + mn);

    expect(meta.outputs[2].kind).toBe(DominantCycleOutput.Phase);
    expect(meta.outputs[2].shape).toBe(Shape.Scalar);
    expect(meta.outputs[2].mnemonic).toBe(mnPha);
    expect(meta.outputs[2].description).toBe('Dominant cycle phase ' + mnPha);
  });

  it('should throw if α is out of range (0, 1]', () => {
    expect(() => DominantCycle.fromParams({ alphaEmaPeriodAdditional: 0 })).toThrow();
    expect(() => DominantCycle.fromParams({ alphaEmaPeriodAdditional: -0.1 })).toThrow();
    expect(() => DominantCycle.fromParams({ alphaEmaPeriodAdditional: 1.0001 })).toThrow();
    expect(() => DominantCycle.fromParams({ alphaEmaPeriodAdditional: 0.5 })).not.toThrow();
    expect(() => DominantCycle.fromParams({ alphaEmaPeriodAdditional: 1.0 })).not.toThrow();
  });

  it('should return NaN triple for NaN input', () => {
    const dc = DominantCycle.default();
    const [r, p, h] = dc.update(Number.NaN);
    expect(Number.isNaN(r)).toBe(true);
    expect(Number.isNaN(p)).toBe(true);
    expect(Number.isNaN(h)).toBe(true);
  });

  it('should match reference period (test_MAMA.xsl) past settle window', () => {
    const dc = DominantCycle.default();
    for (let i = skip; i < input.length; i++) {
      const [, period] = dc.update(input[i]);
      if (Number.isNaN(period) || i < settleSkip) continue;
      expect(Math.abs(expectedPeriod[i] - period))
        .withContext(`period[${i}]: expected ${expectedPeriod[i]}, actual ${period}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference phase (test_HT.xsl) past settle window (modulo 360)', () => {
    const dc = DominantCycle.default();
    for (let i = skip; i < input.length; i++) {
      const [, , phase] = dc.update(input[i]);
      if (Number.isNaN(phase) || i < settleSkip) continue;
      if (Number.isNaN(expectedPhase[i])) continue;
      expect(Math.abs(phaseDiff(expectedPhase[i], phase)))
        .withContext(`phase[${i}]: expected ${expectedPhase[i]}, actual ${phase}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should become primed within the input sequence', () => {
    const dc = DominantCycle.default();
    expect(dc.isPrimed()).toBe(false);

    let primedAt = -1;
    for (let i = 0; i < input.length; i++) {
      dc.update(input[i]);
      if (dc.isPrimed() && primedAt < 0) {
        primedAt = i;
      }
    }
    expect(primedAt).toBeGreaterThanOrEqual(0);
    expect(dc.isPrimed()).toBe(true);
  });

  const primeCount = 120;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(3);
    for (let i = 0; i < 3; i++) {
      const s = out[i] as Scalar;
      expect(s.time).toEqual(time);
    }
  }

  it('should produce 3-scalar output via updateScalar', () => {
    const dc = DominantCycle.default();
    for (let i = 0; i < primeCount; i++) dc.update(value);
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(dc.updateScalar(s) as any[]);
  });

  it('should produce 3-scalar output via updateBar', () => {
    const dc = DominantCycle.default();
    for (let i = 0; i < primeCount; i++) dc.update(value);
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(dc.updateBar(bar) as any[]);
  });

  it('should produce 3-scalar output via updateQuote', () => {
    const dc = DominantCycle.default();
    for (let i = 0; i < primeCount; i++) dc.update(value);
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(dc.updateQuote(q) as any[]);
  });

  it('should produce 3-scalar output via updateTrade', () => {
    const dc = DominantCycle.default();
    for (let i = 0; i < primeCount; i++) dc.update(value);
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(dc.updateTrade(t) as any[]);
  });

  it('should return NaN smoothedPrice before primed and finite after', () => {
    const dc = DominantCycle.default();
    expect(Number.isNaN(dc.smoothedPrice)).toBe(true);

    for (let i = 0; i < input.length; i++) {
      dc.update(input[i]);
      if (dc.isPrimed()) {
        expect(Number.isNaN(dc.smoothedPrice)).toBe(false);
        return;
      }
      expect(Number.isNaN(dc.smoothedPrice)).toBe(true);
    }
    fail('indicator never primed');
  });

  it('should expose maxPeriod matching the HTCE max period', () => {
    const dc = DominantCycle.default();
    // Default HomodyneDiscriminator with smoothingLength=4 => maxPeriod=50.
    expect(dc.maxPeriod).toBe(50);
  });
});
