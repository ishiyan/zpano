import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Band } from '../../core/outputs/band';
import { Shape } from '../../core/outputs/shape/shape';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { SineWave } from './sine-wave';
import { SineWaveOutput } from './output';
import { input, expectedSine, expectedSineLead } from './testdata';

/* eslint-disable max-len */
// Input data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA.xsl, Price, D5…D256, 252 entries.
// Expected sine wave values taken from MBST SineWaveTest.cs. 252 entries.
// Expected sine wave lead values taken from MBST SineWaveTest.cs. 252 entries.
const tolerance = 1e-4;

describe('SineWave', () => {
  const time = new Date(2021, 3, 1);
  const skip = 9;
  const settleSkip = 177;

  it('should have correct output enum values', () => {
    expect(SineWaveOutput.Value).toBe(0);
    expect(SineWaveOutput.Lead).toBe(1);
    expect(SineWaveOutput.Band).toBe(2);
    expect(SineWaveOutput.DominantCyclePeriod).toBe(3);
    expect(SineWaveOutput.DominantCyclePhase).toBe(4);
  });

  it('should return expected mnemonic for default params (BarComponent.Median)', () => {
    const sw = SineWave.default();
    expect(sw.metadata().mnemonic).toBe('sw(0.330, hl/2)');
  });

  it('should return expected mnemonic for explicit components matching Close default', () => {
    const sw = SineWave.fromParams({ alphaEmaPeriodAdditional: 0.5, barComponent: BarComponent.Close });
    expect(sw.metadata().mnemonic).toBe('sw(0.500)');
  });

  it('should return expected mnemonic for phase accumulator', () => {
    const sw = SineWave.fromParams({
      alphaEmaPeriodAdditional: 0.5,
      estimatorType: HilbertTransformerCycleEstimatorType.PhaseAccumulator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
      barComponent: BarComponent.Close,
    });
    expect(sw.metadata().mnemonic).toBe('sw(0.500, pa(4, 0.200, 0.200))');
  });

  it('should return expected metadata', () => {
    const sw = SineWave.default();
    const meta = sw.metadata();
    const mn = 'sw(0.330, hl/2)';
    const mnLead = 'sw-lead(0.330, hl/2)';
    const mnBand = 'sw-band(0.330, hl/2)';
    const mnDCP = 'dcp(0.330, hl/2)';
    const mnDCPha = 'dcph(0.330, hl/2)';

    expect(meta.identifier).toBe(IndicatorIdentifier.SineWave);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe('Sine wave ' + mn);
    expect(meta.outputs.length).toBe(5);

    expect(meta.outputs[0].kind).toBe(SineWaveOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe('Sine wave ' + mn);

    expect(meta.outputs[1].kind).toBe(SineWaveOutput.Lead);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnLead);
    expect(meta.outputs[1].description).toBe('Sine wave lead ' + mnLead);

    expect(meta.outputs[2].kind).toBe(SineWaveOutput.Band);
    expect(meta.outputs[2].shape).toBe(Shape.Band);
    expect(meta.outputs[2].mnemonic).toBe(mnBand);
    expect(meta.outputs[2].description).toBe('Sine wave band ' + mnBand);

    expect(meta.outputs[3].kind).toBe(SineWaveOutput.DominantCyclePeriod);
    expect(meta.outputs[3].shape).toBe(Shape.Scalar);
    expect(meta.outputs[3].mnemonic).toBe(mnDCP);
    expect(meta.outputs[3].description).toBe('Dominant cycle period ' + mnDCP);

    expect(meta.outputs[4].kind).toBe(SineWaveOutput.DominantCyclePhase);
    expect(meta.outputs[4].shape).toBe(Shape.Scalar);
    expect(meta.outputs[4].mnemonic).toBe(mnDCPha);
    expect(meta.outputs[4].description).toBe('Dominant cycle phase ' + mnDCPha);
  });

  it('should throw if α is out of range (0, 1]', () => {
    expect(() => SineWave.fromParams({ alphaEmaPeriodAdditional: 0 })).toThrow();
    expect(() => SineWave.fromParams({ alphaEmaPeriodAdditional: -0.1 })).toThrow();
    expect(() => SineWave.fromParams({ alphaEmaPeriodAdditional: 1.0001 })).toThrow();
    expect(() => SineWave.fromParams({ alphaEmaPeriodAdditional: 0.5 })).not.toThrow();
    expect(() => SineWave.fromParams({ alphaEmaPeriodAdditional: 1.0 })).not.toThrow();
  });

  it('should return NaN quadruple for NaN input', () => {
    const sw = SineWave.default();
    const [v, l, p, h] = sw.update(Number.NaN);
    expect(Number.isNaN(v)).toBe(true);
    expect(Number.isNaN(l)).toBe(true);
    expect(Number.isNaN(p)).toBe(true);
    expect(Number.isNaN(h)).toBe(true);
  });

  it('should match reference sine (MBST SineWaveTest) past settle window', () => {
    const sw = SineWave.default();
    for (let i = skip; i < input.length; i++) {
      const [value] = sw.update(input[i]);
      if (Number.isNaN(value) || i < settleSkip) continue;
      if (Number.isNaN(expectedSine[i])) continue;
      expect(Math.abs(expectedSine[i] - value))
        .withContext(`sine[${i}]: expected ${expectedSine[i]}, actual ${value}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should match reference sine lead (MBST SineWaveLeadTest) past settle window', () => {
    const sw = SineWave.default();
    for (let i = skip; i < input.length; i++) {
      const [, lead] = sw.update(input[i]);
      if (Number.isNaN(lead) || i < settleSkip) continue;
      if (Number.isNaN(expectedSineLead[i])) continue;
      expect(Math.abs(expectedSineLead[i] - lead))
        .withContext(`sineLead[${i}]: expected ${expectedSineLead[i]}, actual ${lead}`)
        .toBeLessThan(tolerance);
    }
  });

  it('should become primed within the input sequence', () => {
    const sw = SineWave.default();
    expect(sw.isPrimed()).toBe(false);

    let primedAt = -1;
    for (let i = 0; i < input.length; i++) {
      sw.update(input[i]);
      if (sw.isPrimed() && primedAt < 0) {
        primedAt = i;
      }
    }
    expect(primedAt).toBeGreaterThanOrEqual(0);
    expect(sw.isPrimed()).toBe(true);
  });

  const primeCount = 200;
  const value = 100.0;

  function checkOutput(out: any[]): void {
    expect(out.length).toBe(5);
    // indices 0, 1, 3, 4 are scalars; 2 is a band.
    for (const i of [0, 1, 3, 4]) {
      const s = out[i] as Scalar;
      expect(s.time).toEqual(time);
    }
    const b = out[2] as Band;
    expect(b.time).toEqual(time);
  }

  it('should produce 5-element output via updateScalar', () => {
    const sw = SineWave.default();
    for (let i = 0; i < primeCount; i++) sw.update(input[i % input.length]);
    const s = new Scalar();
    s.time = time;
    s.value = value;
    checkOutput(sw.updateScalar(s) as any[]);
  });

  it('should produce 5-element output via updateBar', () => {
    const sw = SineWave.default();
    for (let i = 0; i < primeCount; i++) sw.update(input[i % input.length]);
    const bar = new Bar({ time, open: value, high: value, low: value, close: value, volume: 0 });
    checkOutput(sw.updateBar(bar) as any[]);
  });

  it('should produce 5-element output via updateQuote', () => {
    const sw = SineWave.default();
    for (let i = 0; i < primeCount; i++) sw.update(input[i % input.length]);
    const q = new Quote({ time, bid: value, ask: value, bidSize: 0, askSize: 0 });
    checkOutput(sw.updateQuote(q) as any[]);
  });

  it('should produce 5-element output via updateTrade', () => {
    const sw = SineWave.default();
    for (let i = 0; i < primeCount; i++) sw.update(input[i % input.length]);
    const t = new Trade({ time, price: value, volume: 0 });
    checkOutput(sw.updateTrade(t) as any[]);
  });

  it('should order band as {upper: value, lower: lead}', () => {
    const sw = SineWave.default();
    for (let i = 0; i < primeCount; i++) sw.update(input[i % input.length]);
    const s = new Scalar();
    s.time = time;
    s.value = input[0];
    const out = sw.updateScalar(s) as any[];
    const sv = out[SineWaveOutput.Value] as Scalar;
    const sl = out[SineWaveOutput.Lead] as Scalar;
    const band = out[SineWaveOutput.Band] as Band;
    expect(band.upper).toBe(sv.value);
    expect(band.lower).toBe(sl.value);
  });
});
