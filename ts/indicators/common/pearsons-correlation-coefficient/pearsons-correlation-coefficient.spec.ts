import { } from 'jasmine';

import { PearsonsCorrelationCoefficient } from './pearsons-correlation-coefficient';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { PearsonsCorrelationCoefficientOutput } from './output';
import { Bar } from '../../../entities/bar';
import { highInput, lowInput, excelExpected } from './testdata';

// High input data from TA-Lib test_data.c (252 entries).
// Low input data from TA-Lib test_data.c (252 entries).
// Excel expected CORREL values for period=20 (252 values).
function roundTo(v: number, digits: number): number {
  const p = Math.pow(10, digits);
  return Math.round(v * p) / p;
}

describe('PearsonsCorrelationCoefficient', () => {

  it('should throw if length is less than 1', () => {
    expect(() => { new PearsonsCorrelationCoefficient({ length: 0 }); }).toThrow();
  });

  it('should throw if length is negative', () => {
    expect(() => { new PearsonsCorrelationCoefficient({ length: -8 }); }).toThrow();
  });

  it('should calculate CORREL(20) with TaLib spot checks', () => {
    const c = new PearsonsCorrelationCoefficient({ length: 20 });

    for (let i = 0; i < 19; i++) {
      const v = c.updatePair(highInput[i], lowInput[i]);
      expect(v).toBeNaN();
      expect(c.isPrimed()).toBe(false);
    }

    const results: number[] = [];
    for (let i = 19; i < highInput.length; i++) {
      const v = c.updatePair(highInput[i], lowInput[i]);
      expect(v).not.toBeNaN();
      expect(c.isPrimed()).toBe(true);
      results.push(v);
    }

    // TaLib spot checks: output index 0, 1, and 232 (last).
    expect(roundTo(results[0], 7)).toBe(roundTo(0.9401569, 7));
    expect(roundTo(results[1], 7)).toBe(roundTo(0.9471812, 7));
    expect(roundTo(results[232], 7)).toBe(roundTo(0.8866901, 7));
  });

  it('should match Excel verification data with high precision', () => {
    const digits = 9;
    const c = new PearsonsCorrelationCoefficient({ length: 20 });

    for (let i = 0; i < 19; i++) {
      const v = c.updatePair(highInput[i], lowInput[i]);
      expect(v).toBeNaN();
    }

    for (let i = 19; i < highInput.length; i++) {
      const v = c.updatePair(highInput[i], lowInput[i]);
      expect(roundTo(v, digits)).toBe(roundTo(excelExpected[i], digits));
    }
  });

  it('should report correct primed state', () => {
    const c = new PearsonsCorrelationCoefficient({ length: 5 });
    expect(c.isPrimed()).toBe(false);

    for (let i = 0; i < 4; i++) {
      c.updatePair(highInput[i], lowInput[i]);
      expect(c.isPrimed()).toBe(false);
    }

    c.updatePair(highInput[4], lowInput[4]);
    expect(c.isPrimed()).toBe(true);

    c.updatePair(highInput[5], lowInput[5]);
    expect(c.isPrimed()).toBe(true);
  });

  it('should pass NaN through', () => {
    const c = new PearsonsCorrelationCoefficient({ length: 5 });
    expect(c.updatePair(Number.NaN, 1.0)).toBeNaN();
    expect(c.updatePair(1.0, Number.NaN)).toBeNaN();
    expect(c.updatePair(Number.NaN, Number.NaN)).toBeNaN();
  });

  it('should return 0 for constant input (zero variance)', () => {
    const c = new PearsonsCorrelationCoefficient({ length: 2 });
    c.update(5);
    const v = c.update(5);
    expect(v).toBe(0);
  });

  it('should return correct metadata', () => {
    const c = new PearsonsCorrelationCoefficient({ length: 20 });
    const meta = c.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.PearsonsCorrelationCoefficient);
    expect(meta.mnemonic).toBe('correl(20)');
    expect(meta.description).toBe('Pearsons Correlation Coefficient correl(20)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(PearsonsCorrelationCoefficientOutput.PearsonsCorrelationCoefficientValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should work with updateBar extracting high and low', () => {
    const c = new PearsonsCorrelationCoefficient({ length: 2 });
    const now = new Date();

    const bar1 = new Bar();
    bar1.time = now;
    bar1.high = 10;
    bar1.low = 5;
    bar1.open = 7;
    bar1.close = 8;
    bar1.volume = 100;
    c.updatePair(bar1.high, bar1.low);

    const bar2 = new Bar();
    bar2.time = now;
    bar2.high = 20;
    bar2.low = 10;
    bar2.open = 15;
    bar2.close = 18;
    bar2.volume = 200;

    const out = c.updateBar(bar2);
    expect(out.length).toBe(1);
    expect(out[0].time).toBe(now);
    expect((out[0] as any).value).not.toBeNaN();
  });
});
