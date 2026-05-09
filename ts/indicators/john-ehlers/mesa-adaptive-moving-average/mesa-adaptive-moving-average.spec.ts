import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent } from '../../../entities/trade-component';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Band } from '../../core/outputs/band';
import { Shape } from '../../core/outputs/shape/shape';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { MesaAdaptiveMovingAverage } from './mesa-adaptive-moving-average';
import { MesaAdaptiveMovingAverageOutput } from './output';
import { input, expectedMama, expectedFama } from './testdata';

/* eslint-disable max-len */
// Input data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA.xsl, Price, D5…D256, 252 entries.
// Expected data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA_new.xsl,
// MAMA, L5…L256, 252 entries,
// FAMA, M5…M256, 252 entries.
// All parameters have default values.

describe('MesaAdaptiveMovingAverage', () => {
  const eps = 1e-10;
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(MesaAdaptiveMovingAverageOutput.Value).toBe(0);
    expect(MesaAdaptiveMovingAverageOutput.Fama).toBe(1);
    expect(MesaAdaptiveMovingAverageOutput.Band).toBe(2);
  });

  it('should return expected mnemonic for length params', () => {
    let mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39)');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40)');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminatorUnrolled,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hdu(4, 0.200, 0.200))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.PhaseAccumulator,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, pa(4, 0.150, 0.250))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.DualDifferentiator,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, dd(4, 0.150, 0.150))');
  });

  it('should return expected mnemonic for smoothing factor params', () => {
    let mama = MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: 0.05,
    });
    expect(mama.metadata().mnemonic).toBe('mama(0.500, 0.050)');

    mama = MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: 0.05,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminatorUnrolled,
    });
    expect(mama.metadata().mnemonic).toBe('mama(0.500, 0.050, hdu(4, 0.200, 0.200))');
  });

  it('should return expected mnemonic for explicit estimator params', () => {
    // Default HomodyneDiscriminator params produce no moniker suffix.
    let mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40)');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 3, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hd(3, 0.200, 0.200))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.3, alphaEmaPeriod: 0.2 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hd(4, 0.300, 0.200))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.3 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hd(4, 0.200, 0.300))');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 3, slowLimitLength: 39,
      barComponent: BarComponent.Median,
    });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 3, slowLimitLength: 39,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 3, slowLimitLength: 39,
      tradeComponent: TradeComponent.Volume,
    });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39, v)');
  });

  it('should return expected metadata', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });
    const meta = mama.metadata();
    const mn = 'mama(3, 39)';
    const mnFama = 'fama(3, 39)';
    const mnBand = 'mama-fama(3, 39)';
    const descr = 'Mesa adaptive moving average ';

    expect(meta.identifier).toBe(IndicatorIdentifier.MesaAdaptiveMovingAverage);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe(descr + mn);
    expect(meta.outputs.length).toBe(3);

    expect(meta.outputs[0].kind).toBe(MesaAdaptiveMovingAverageOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe(descr + mn);

    expect(meta.outputs[1].kind).toBe(MesaAdaptiveMovingAverageOutput.Fama);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnFama);
    expect(meta.outputs[1].description).toBe(descr + mnFama);

    expect(meta.outputs[2].kind).toBe(MesaAdaptiveMovingAverageOutput.Band);
    expect(meta.outputs[2].shape).toBe(Shape.Band);
    expect(meta.outputs[2].mnemonic).toBe(mnBand);
    expect(meta.outputs[2].description).toBe(descr + mnBand);
  });

  it('should throw if the fast limit length is less than 2', () => {
    expect(() => MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 1, slowLimitLength: 39 })).toThrow();
  });

  it('should throw if the slow limit length is less than 2', () => {
    expect(() => MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 1 })).toThrow();
  });

  it('should throw if the fast limit smoothing factor is out of range', () => {
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: -0.01, slowLimitSmoothingFactor: 0.05,
    })).toThrow();
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 1.01, slowLimitSmoothingFactor: 0.05,
    })).toThrow();
  });

  it('should throw if the slow limit smoothing factor is out of range', () => {
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: -0.01,
    })).toThrow();
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: 1.01,
    })).toThrow();
  });

  it('should calculate expected MAMA update values and prime state', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < lprimed; i++) {
      expect(mama.update(input[i])).toBeNaN();
      expect(mama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      const act = mama.update(input[i]);
      expect(mama.isPrimed()).toBe(true);
      expect(Math.abs(act - expectedMama[i]))
        .withContext(`MAMA[${i}]: expected ${expectedMama[i]}, actual ${act}`)
        .toBeLessThan(eps);
    }

    expect(mama.update(Number.NaN)).toBeNaN();
  });

  it('should produce expected MAMA/FAMA/Band via updateScalar', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < lprimed; i++) {
      const out = mama.updateScalar(new Scalar({ time, value: input[i] }));
      expect(out.length).toBe(3);
      expect(Number.isNaN((out[0] as Scalar).value)).toBe(true);
      expect(Number.isNaN((out[1] as Scalar).value)).toBe(true);
      expect(Number.isNaN((out[2] as Band).upper)).toBe(true);
      expect(Number.isNaN((out[2] as Band).lower)).toBe(true);
      expect(mama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      const out = mama.updateScalar(new Scalar({ time, value: input[i] }));
      expect(out.length).toBe(3);
      expect(mama.isPrimed()).toBe(true);

      const s0 = out[0] as Scalar;
      const s1 = out[1] as Scalar;
      const b = out[2] as Band;

      expect(Math.abs(s0.value - expectedMama[i]))
        .withContext(`MAMA[${i}]: expected ${expectedMama[i]}, actual ${s0.value}`)
        .toBeLessThan(eps);
      expect(Math.abs(s1.value - expectedFama[i]))
        .withContext(`FAMA[${i}]: expected ${expectedFama[i]}, actual ${s1.value}`)
        .toBeLessThan(eps);
      expect(b.upper).toBe(s0.value);
      expect(b.lower).toBe(s1.value);
      expect(b.time).toBe(time);
    }
  });

  it('should produce expected output via updateBar', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < input.length; i++) {
      const bar = new Bar({ time, open: input[i], high: input[i], low: input[i], close: input[i], volume: 0 });
      const out = mama.updateBar(bar);
      expect(out.length).toBe(3);
      if (i >= lprimed) {
        expect(Math.abs((out[0] as Scalar).value - expectedMama[i])).toBeLessThan(eps);
        expect(Math.abs((out[1] as Scalar).value - expectedFama[i])).toBeLessThan(eps);
      }
    }
  });

  it('should produce expected output via updateQuote', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < input.length; i++) {
      const q = new Quote({ time, bidPrice: input[i], askPrice: input[i], bidSize: 0, askSize: 0 });
      const out = mama.updateQuote(q);
      expect(out.length).toBe(3);
      if (i >= lprimed) {
        expect(Math.abs((out[0] as Scalar).value - expectedMama[i])).toBeLessThan(eps);
        expect(Math.abs((out[1] as Scalar).value - expectedFama[i])).toBeLessThan(eps);
      }
    }
  });

  it('should produce expected output via updateTrade', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < input.length; i++) {
      const r = new Trade({ time, price: input[i], volume: 0 });
      const out = mama.updateTrade(r);
      expect(out.length).toBe(3);
      if (i >= lprimed) {
        expect(Math.abs((out[0] as Scalar).value - expectedMama[i])).toBeLessThan(eps);
        expect(Math.abs((out[1] as Scalar).value - expectedFama[i])).toBeLessThan(eps);
      }
    }
  });

  it('should construct via default() factory', () => {
    const mama = MesaAdaptiveMovingAverage.default();
    expect(mama.metadata().mnemonic).toBe('mama(3, 39)');
    expect(mama.isPrimed()).toBe(false);
  });
});
