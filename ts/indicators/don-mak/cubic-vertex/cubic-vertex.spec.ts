import { CubicVertex } from './cubic-vertex';
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
      // Combined absolute + relative tolerance (ill-conditioned near degenerate points).
      const delta = TOLERANCE * Math.max(1, Math.abs(expected[i]));
      expect(Math.abs(actual[i] - expected[i]))
        .withContext(`${label}[${i}] expected ${expected[i]} got ${actual[i]}`)
        .toBeLessThanOrEqual(delta);
    }
  }
}

function run(inputs: number[]): { near: number[]; far: number[] } {
  const ind = new CubicVertex({});
  const near: number[] = [];
  const far: number[] = [];
  for (const c of inputs) {
    const [n, f] = ind.update(c);
    near.push(n);
    far.push(f);
  }
  return { near, far };
}

describe('CubicVertex', () => {
  describe('reference data', () => {
    it('matches the raw market data series', () => {
      const { near, far } = run(td.iNPUT_CLOSE);
      expectSeries(near, td.expectedRAW_NEAR, 'RAW_NEAR');
      expectSeries(far, td.expectedRAW_FAR, 'RAW_FAR');
    });

    it('matches the EMA(6) smoothed series', () => {
      const { near, far } = run(td.iNPUT_EMA6);
      expectSeries(near, td.expectedEMA6_NEAR, 'EMA6_NEAR');
      expectSeries(far, td.expectedEMA6_FAR, 'EMA6_FAR');
    });

    it('matches the EMA(20) smoothed series', () => {
      const { near, far } = run(td.iNPUT_EMA20);
      expectSeries(near, td.expectedEMA20_NEAR, 'EMA20_NEAR');
      expectSeries(far, td.expectedEMA20_FAR, 'EMA20_FAR');
    });

    it('matches the synthetic cubic series (TEST1)', () => {
      const { near, far } = run(td.tEST1_INPUT_CUBIC);
      expectSeries(near, td.tEST1_EXPECTED_NEAR, 'TEST1_NEAR');
      expectSeries(far, td.tEST1_EXPECTED_FAR, 'TEST1_FAR');
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      expect(new CubicVertex({}).metadata().mnemonic).toBe('cvtx');
    });

    it('formats a mnemonic with a non-default component', () => {
      expect(new CubicVertex({ barComponent: BarComponent.Median }).metadata().mnemonic).toBe('cvtx(hl/2)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const meta = new CubicVertex({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.CubicVertex);
      expect(meta.outputs.length).toBe(2);
    });
  });

  describe('priming', () => {
    it('returns NaN during priming and is primed after 4 bars', () => {
      const ind = new CubicVertex({});
      for (let i = 0; i < 3; i++) {
        const [n, f] = ind.update(1);
        expect(n).toBeNaN();
        expect(f).toBeNaN();
        expect(ind.isPrimed()).toBe(false);
      }
      // Four collinear points -> c == 0 and d == 0 -> both NaN, but primed.
      const [n, f] = ind.update(1);
      expect(n).toBeNaN();
      expect(f).toBeNaN();
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('updateScalar', () => {
    it('returns two outputs', () => {
      const ind = new CubicVertex({});
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of td.iNPUT_CLOSE) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(2);
    });
  });
});
