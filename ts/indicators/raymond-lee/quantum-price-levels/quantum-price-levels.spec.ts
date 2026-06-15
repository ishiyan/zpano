import { QuantumPriceLevels } from './quantum-price-levels';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import { Levels } from '../../core/outputs/levels';
import * as td from './testdata';

const TOLERANCE = 1e-9;

interface Result {
  lambda: number;
  sigma: number;
  nqpr: number[];
  resistances: number[];
  supports: number[];
  valid: boolean;
}

function runLast(inputs: number[], lookback: number, numLevels: number, numBins: number, scaleFactor: number): Result {
  const lb = lookback === 0 ? inputs.length - 1 : lookback;
  const ind = new QuantumPriceLevels({ lookback: lb, numLevels, numBins, scaleFactor });
  let last: Result = { lambda: NaN, sigma: NaN, nqpr: [], resistances: [], supports: [], valid: false };
  for (const p of inputs) {
    const r = ind.update(p);
    if (ind.isPrimed()) {
      last = r;
    }
  }
  return last;
}

function expectSeries(actual: number[], expected: number[], label: string): void {
  expect(actual.length).withContext(`${label}: length`).toBe(expected.length);
  for (let i = 0; i < expected.length; i++) {
    const delta = TOLERANCE * Math.max(1, Math.abs(expected[i]));
    expect(Math.abs(actual[i] - expected[i]))
      .withContext(`${label}[${i}] expected ${expected[i]} got ${actual[i]}`)
      .toBeLessThanOrEqual(delta);
  }
}

function check(label: string, r: Result, expNqpr: number[], expUpper: number[], expLower: number[]): void {
  expect(r.valid).withContext(`${label}: valid`).toBe(true);
  expectSeries(r.nqpr, expNqpr, `${label} NQPR`);
  expectSeries(r.resistances, expUpper, `${label} UPPER`);
  expectSeries(r.supports, expLower, `${label} LOWER`);
}

describe('QuantumPriceLevels', () => {
  describe('batch combos (252-bar input)', () => {
    it('default', () => check('default', runLast(td.testInput, 0, 21, 100, 0.21), td.expectedNQPR, td.expectedUPPER, td.expectedLOWER));
    it('F0_10', () => check('F0_10', runLast(td.testInput, 0, 21, 100, 0.10), td.expectedNQPR_F0_10, td.expectedUPPER_F0_10, td.expectedLOWER_F0_10));
    it('F0_42', () => check('F0_42', runLast(td.testInput, 0, 21, 100, 0.42), td.expectedNQPR_F0_42, td.expectedUPPER_F0_42, td.expectedLOWER_F0_42));
    it('B50', () => check('B50', runLast(td.testInput, 0, 21, 50, 0.21), td.expectedNQPR_B50, td.expectedUPPER_B50, td.expectedLOWER_B50));
    it('B50_F0_10', () => check('B50_F0_10', runLast(td.testInput, 0, 21, 50, 0.10), td.expectedNQPR_B50_F0_10, td.expectedUPPER_B50_F0_10, td.expectedLOWER_B50_F0_10));
    it('B50_F0_42', () => check('B50_F0_42', runLast(td.testInput, 0, 21, 50, 0.42), td.expectedNQPR_B50_F0_42, td.expectedUPPER_B50_F0_42, td.expectedLOWER_B50_F0_42));
    it('L5', () => check('L5', runLast(td.testInput, 0, 5, 100, 0.21), td.expectedNQPR_L5, td.expectedUPPER_L5, td.expectedLOWER_L5));
    it('L10', () => check('L10', runLast(td.testInput, 0, 10, 100, 0.21), td.expectedNQPR_L10, td.expectedUPPER_L10, td.expectedLOWER_L10));
    it('L10_B50_F0_42', () => check('L10_B50_F0_42', runLast(td.testInput, 0, 10, 50, 0.42), td.expectedNQPR_L10_B50_F0_42, td.expectedUPPER_L10_B50_F0_42, td.expectedLOWER_L10_B50_F0_42));
    it('2K', () => check('2K', runLast(td.testInput2K, 0, 21, 100, 0.21), td.expectedNQPR_2K, td.expectedUPPER_2K, td.expectedLOWER_2K));
  });

  describe('reference-price combos (re-projected)', () => {
    const r = runLast(td.testInput, 0, 21, 100, 0.21);
    const project = (ref: number): { up: number[]; lo: number[] } => ({
      up: r.nqpr.map((m) => ref * m),
      lo: r.nqpr.map((m) => ref / m),
    });

    it('R50_0', () => {
      expectSeries(r.nqpr, td.expectedNQPR_R50_0, 'R50_0 NQPR');
      const { up, lo } = project(50.0);
      expectSeries(up, td.expectedUPPER_R50_0, 'R50_0 UPPER');
      expectSeries(lo, td.expectedLOWER_R50_0, 'R50_0 LOWER');
    });

    it('R1000_0', () => {
      const { up, lo } = project(1000.0);
      expectSeries(r.nqpr, td.expectedNQPR_R1000_0, 'R1000_0 NQPR');
      expectSeries(up, td.expectedUPPER_R1000_0, 'R1000_0 UPPER');
      expectSeries(lo, td.expectedLOWER_R1000_0, 'R1000_0 LOWER');
    });

    it('R1_2345', () => {
      const { up, lo } = project(1.2345);
      expectSeries(r.nqpr, td.expectedNQPR_R1_2345, 'R1_2345 NQPR');
      expectSeries(up, td.expectedUPPER_R1_2345, 'R1_2345 UPPER');
      expectSeries(lo, td.expectedLOWER_R1_2345, 'R1_2345 LOWER');
    });
  });

  describe('streaming combos', () => {
    it('S100', () => check('S100', runLast(td.testInput, 100, 21, 100, 0.21), td.expectedNQPR_S100, td.expectedUPPER_S100, td.expectedLOWER_S100));
    it('S150_B50', () => check('S150_B50', runLast(td.testInput, 150, 21, 50, 0.21), td.expectedNQPR_S150_B50, td.expectedUPPER_S150_B50, td.expectedLOWER_S150_B50));
    it('S200_F0_42', () => check('S200_F0_42', runLast(td.testInput, 200, 21, 100, 0.42), td.expectedNQPR_S200_F0_42, td.expectedUPPER_S200_F0_42, td.expectedLOWER_S200_F0_42));
  });

  describe('scalars', () => {
    it('lambda and sigma match documented values', () => {
      const r = runLast(td.testInput, 0, 21, 100, 0.21);
      expect(Math.abs(r.lambda - 9.739608012591481e-01)).toBeLessThanOrEqual(1e-9);
      expect(Math.abs(r.sigma - 2.662021797593086e-02)).toBeLessThanOrEqual(1e-9);
    });
  });

  describe('mnemonic & metadata', () => {
    it('default mnemonic', () => {
      expect(new QuantumPriceLevels({}).metadata().mnemonic).toBe('qpl(2048,21,100,0.21)');
    });
    it('5 outputs and identifier', () => {
      const meta = new QuantumPriceLevels({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.QuantumPriceLevels);
      expect(meta.outputs.length).toBe(5);
    });
  });

  describe('updateScalar', () => {
    it('returns 5 outputs with Levels', () => {
      const ind = new QuantumPriceLevels({ lookback: 100 });
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const p of td.testInput) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = p;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(5);
      expect((out[3] as Levels).levels.length).toBe(21);
    });
  });

  describe('invalid parameters', () => {
    it('throws when lookback < 2', () => expect(() => new QuantumPriceLevels({ lookback: 1 })).toThrowError());
    it('throws when numLevels < 1', () => expect(() => new QuantumPriceLevels({ numLevels: 0 })).toThrowError());
    it('throws when numBins < 2', () => expect(() => new QuantumPriceLevels({ numBins: 1 })).toThrowError());
    it('throws when scaleFactor <= 0', () => expect(() => new QuantumPriceLevels({ scaleFactor: 0 })).toThrowError());
  });
});
