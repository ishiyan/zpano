import { ParabolicVertex } from './parabolic-vertex';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { BarComponent } from '../../../entities/bar-component';
import { Scalar } from '../../../entities/scalar';
import * as td from './testdata';

const TOLERANCE = 1e-9;

function expectSeries(actual: number[], expected: number[], label: string): void {
  expect(actual.length).toBe(expected.length);
  for (let i = 0; i < expected.length; i++) {
    if (isNaN(expected[i])) {
      expect(actual[i]).withContext(`${label}[${i}]`).toBeNaN();
    } else {
      // Combined absolute + relative tolerance. Near collinear points the vertex
      // location is ill-conditioned (denom -> 0); a relative tolerance preserves
      // 13+ significant-digit agreement.
      const delta = TOLERANCE * Math.max(1, Math.abs(expected[i]));
      expect(Math.abs(actual[i] - expected[i]))
        .withContext(`${label}[${i}] expected ${expected[i]} got ${actual[i]}`)
        .toBeLessThanOrEqual(delta);
    }
  }
}

function run(inputs: number[]): number[] {
  const ind = new ParabolicVertex({});
  return inputs.map((c) => ind.update(c));
}

describe('ParabolicVertex', () => {
  describe('reference data', () => {
    it('matches the raw market data series', () => {
      expectSeries(run(td.iNPUT_CLOSE), td.expectedRAW, 'RAW');
    });

    it('matches the EMA(6) smoothed series', () => {
      expectSeries(run(td.iNPUT_EMA6), td.expectedEMA6, 'EMA6');
    });

    it('matches the EMA(20) smoothed series', () => {
      expectSeries(run(td.iNPUT_EMA20), td.expectedEMA20, 'EMA20');
    });

    it('matches the synthetic parabola series (TEST1)', () => {
      expectSeries(run(td.tEST1_INPUT_PARABOLA), td.tEST1_EXPECTED, 'TEST1');
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      expect(new ParabolicVertex({}).metadata().mnemonic).toBe('pvtx');
    });

    it('formats a mnemonic with a non-default component', () => {
      expect(new ParabolicVertex({ barComponent: BarComponent.Median }).metadata().mnemonic).toBe('pvtx(hl/2)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const meta = new ParabolicVertex({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.ParabolicVertex);
      expect(meta.outputs.length).toBe(1);
    });
  });

  describe('priming', () => {
    it('returns NaN during priming and is primed after 3 bars', () => {
      const ind = new ParabolicVertex({});
      expect(ind.update(1)).toBeNaN();
      expect(ind.isPrimed()).toBe(false);
      expect(ind.update(2)).toBeNaN();
      expect(ind.isPrimed()).toBe(false);
      // Three collinear points -> zero curvature -> NaN, but primed.
      expect(ind.update(3)).toBeNaN();
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('updateScalar', () => {
    it('returns a single output', () => {
      const ind = new ParabolicVertex({});
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of td.iNPUT_CLOSE) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(1);
    });
  });
});
