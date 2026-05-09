import { } from 'jasmine';

import { FractalAdaptiveMovingAverage } from './fractal-adaptive-moving-average';
import { FractalAdaptiveMovingAverageOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { Scalar } from '../../../entities/scalar';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Trade } from '../../../entities/trade';
import {
    inputMid,
    inputHigh,
    inputLow,
    expectedFrama,
    expectedFdim,
} from './testdata';

/* eslint-disable max-len */
// Input data taken from test_FRAMA.xsl reference implementation:
// Mid-Price, D5...D256, 252 entries,
// High, B5...B256, 252 entries,
// Low, C5...C256, 252 entries.
//
// Expected data taken from test_FRAMA.xsl reference implementation:
// FRAMA, R5...R256, 252 entries,
// FDIM, O5...O256, 252 entries.
//
// All parameters have default values.

describe('FractalAdaptiveMovingAverage', () => {
  const time = new Date(2021, 3, 1); // April 1, 2021

  it('should have correct output enum values', () => {
    expect(FractalAdaptiveMovingAverageOutput.Value).toBe(0);
    expect(FractalAdaptiveMovingAverageOutput.Fdim).toBe(1);
  });

  it('should return expected mnemonic', () => {
    let frama = new FractalAdaptiveMovingAverage(
      { length: 16, slowestSmoothingFactor: 0.01 });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010)');

    frama = new FractalAdaptiveMovingAverage(
      { length: 18, slowestSmoothingFactor: 0.005 });
    expect(frama.metadata().mnemonic).toBe('frama(18, 0.005)');

    frama = new FractalAdaptiveMovingAverage(
      { length: 17, slowestSmoothingFactor: 0.01 });
    expect(frama.metadata().mnemonic).toBe('frama(18, 0.010)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const frama = new FractalAdaptiveMovingAverage({
      length: 16, slowestSmoothingFactor: 0.01,
      barComponent: BarComponent.Median,
    });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010, hl/2)');
    expect(frama.metadata().description).toBe('Fractal adaptive moving average frama(16, 0.010, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const frama = new FractalAdaptiveMovingAverage({
      length: 16, slowestSmoothingFactor: 0.01,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const frama = new FractalAdaptiveMovingAverage({
      length: 16, slowestSmoothingFactor: 0.01,
      tradeComponent: TradeComponent.Volume,
    });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010, v)');
  });

  it('should return expected metadata', () => {
    const frama = new FractalAdaptiveMovingAverage(
      { length: 16, slowestSmoothingFactor: 0.01 });
    const meta = frama.metadata();

    const mn = 'frama(16, 0.010)';
    const mnFdim = 'framaDim(16, 0.010)';
    const descr = 'Fractal adaptive moving average ';

    expect(meta.identifier).toBe(IndicatorIdentifier.FractalAdaptiveMovingAverage);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe(descr + mn);
    expect(meta.outputs.length).toBe(2);

    expect(meta.outputs[0].kind).toBe(FractalAdaptiveMovingAverageOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe(descr + mn);

    expect(meta.outputs[1].kind).toBe(FractalAdaptiveMovingAverageOutput.Fdim);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnFdim);
    expect(meta.outputs[1].description).toBe(descr + mnFdim);
  });

  it('should throw if the length is less than 2', () => {
    expect(() => { new FractalAdaptiveMovingAverage({ length: 1, slowestSmoothingFactor: 0.01 }); }).toThrow();
    expect(() => { new FractalAdaptiveMovingAverage({ length: 0, slowestSmoothingFactor: 0.01 }); }).toThrow();
    expect(() => { new FractalAdaptiveMovingAverage({ length: -1, slowestSmoothingFactor: 0.01 }); }).toThrow();
  });

  it('should throw if the slowest smoothing factor is less than 0', () => {
    expect(() => {
      new FractalAdaptiveMovingAverage(
        { length: 16, slowestSmoothingFactor: -0.01 });
    }).toThrow();
  });

  it('should throw if the slowest smoothing factor is greater than 1', () => {
    expect(() => {
      new FractalAdaptiveMovingAverage(
        { length: 16, slowestSmoothingFactor: 1.01 });
    }).toThrow();
  });

  it('should calculate expected FRAMA values from reference implementation', () => {
    const lprimed = 15;
    const eps = 1e-9;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      expect(frama.update(inputMid[i], inputHigh[i], inputLow[i])).toBeNaN();
      expect(frama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < inputMid.length; i++) {
      const act = frama.update(inputMid[i], inputHigh[i], inputLow[i]);
      expect(frama.isPrimed()).toBe(true);
      expect(Math.abs(act - expectedFrama[i])).withContext(`FRAMA [${i}]: expected ${expectedFrama[i]}, actual ${act}`)
        .toBeLessThan(eps);
    }

    expect(frama.update(Number.NaN, Number.NaN, Number.NaN)).toBeNaN();
    expect(frama.update(Number.NaN, 1, 1)).toBeNaN();
    expect(frama.update(1, Number.NaN, 1)).toBeNaN();
    expect(frama.update(1, 1, Number.NaN)).toBeNaN();
  });

  it('should calculate expected Fdim values from reference implementation', () => {
    const lprimed = 15;
    const eps = 1e-9;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    // Access fractal dimension via updateEntity output.
    for (let i = 0; i < lprimed; i++) {
      const output = frama.updateScalar(new Scalar({ time, value: inputMid[i] }));
      // Before primed, scalar uses sample for high/low too, which is different from the reference.
      // Instead, use update directly and check output entity.
    }

    // Re-create to test Fdim properly through the entity path with high/low.
    const frama2 = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama2.update(inputMid[i], inputHigh[i], inputLow[i]);
    }

    for (let i = lprimed; i < inputMid.length; i++) {
      const bar = new Bar({ time, open: inputMid[i], high: inputHigh[i], low: inputLow[i], close: inputMid[i], volume: 0 });
      const output = frama2.updateBar(bar);
      expect(output.length).toBe(2);

      const framaScalar = output[0] as Scalar;
      const fdimScalar = output[1] as Scalar;

      expect(Math.abs(framaScalar.value - expectedFrama[i]))
        .withContext(`FRAMA [${i}]: expected ${expectedFrama[i]}, actual ${framaScalar.value}`)
        .toBeLessThan(eps);

      expect(Math.abs(fdimScalar.value - expectedFdim[i]))
        .withContext(`FDIM [${i}]: expected ${expectedFdim[i]}, actual ${fdimScalar.value}`)
        .toBeLessThan(eps);
    }
  });

  it('should transition primed state correctly', () => {
    const lprimed = 15;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    expect(frama.isPrimed()).toBe(false);

    for (let i = 0; i < lprimed; i++) {
      frama.update(inputMid[i], inputHigh[i], inputLow[i]);
      expect(frama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < inputMid.length; i++) {
      frama.update(inputMid[i], inputHigh[i], inputLow[i]);
      expect(frama.isPrimed()).toBe(true);
    }
  });

  it('should produce correct updateEntity output for scalar', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const s = new Scalar({ time, value: inp });
    const output = frama.updateScalar(s);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    // After priming with zeros, update with 3 should produce a specific value.
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should produce correct updateEntity output for bar', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const b = new Bar({ time, open: inp, high: inp, low: inp, close: inp, volume: 0 });
    const output = frama.updateBar(b);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should produce correct updateEntity output for quote', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const q = new Quote({ time, bidPrice: inp, askPrice: inp, bidSize: 0, askSize: 0 });
    const output = frama.updateQuote(q);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should produce correct updateEntity output for trade', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const r = new Trade({ time, price: inp, volume: 0 });
    const output = frama.updateTrade(r);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should return NaN for fdim when frama is NaN (not primed)', () => {
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    const s = new Scalar({ time, value: 100 });
    const output = frama.updateScalar(s);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(Number.isNaN(s0.value)).toBe(true);
    expect(Number.isNaN(s1.value)).toBe(true);
  });

  it('should match Go updateEntity values exactly', () => {
    const lprimed = 15;
    const inp = 3;
    const expectedFramaValue = 2.999999999999997;
    const expectedFdimValue = 1.0000000000000002;
    const eps = 1e-13;

    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const s = new Scalar({ time, value: inp });
    const output = frama.updateScalar(s);
    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;

    expect(Math.abs(s0.value - expectedFramaValue))
      .withContext(`FRAMA: expected ${expectedFramaValue}, actual ${s0.value}`)
      .toBeLessThan(eps);
    expect(Math.abs(s1.value - expectedFdimValue))
      .withContext(`FDIM: expected ${expectedFdimValue}, actual ${s1.value}`)
      .toBeLessThan(eps);
  });
});
