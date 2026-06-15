import { VelocityCorrectedExponentialMovingAverage } from './velocity-corrected-exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import * as td from './testdata';

const TOLERANCE = 9;

// [period, degree]
const combos: Array<[number, number]> = [
  [3, 2], [3, 3], [3, 4], [3, 5],
  [6, 2], [6, 3], [6, 4], [6, 5],
  [12, 2], [12, 3], [12, 4], [12, 5],
];

const testData = td as unknown as Record<string, number[]>;

function expectSeries(actual: number[], expected: number[], label: string): void {
  expect(actual.length).toBe(expected.length);
  for (let i = 0; i < expected.length; i++) {
    if (isNaN(expected[i])) {
      expect(actual[i]).withContext(`${label}[${i}]`).toBeNaN();
    } else {
      expect(actual[i]).withContext(`${label}[${i}]`).toBeCloseTo(expected[i], TOLERANCE);
    }
  }
}

describe('VelocityCorrectedExponentialMovingAverage', () => {
  describe('reference data', () => {
    combos.forEach(([period, degree]) => {
      const name = `EXPECTED_P${period}_D${degree}`;
      it(`matches the reference for ${name}`, () => {
        const expected = testData[name];
        const ind = new VelocityCorrectedExponentialMovingAverage({ period, degree });
        const values = td.INPUT_CLOSE.map((c) => ind.update(c));
        expectSeries(values, expected, name);
      });
    });

    it('matches the linear-input series (TEST1)', () => {
      const ind = new VelocityCorrectedExponentialMovingAverage({ period: 6, degree: 3 });
      const values = td.TEST1_INPUT_LINEAR.map((c) => ind.update(c));
      expectSeries(values, td.TEST1_EXPECTED_P6_D3, 'TEST1');
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      expect(new VelocityCorrectedExponentialMovingAverage({}).metadata().mnemonic).toBe('vcema(6,3)');
    });

    it('formats a custom mnemonic', () => {
      expect(new VelocityCorrectedExponentialMovingAverage({ period: 12, degree: 5 }).metadata().mnemonic).toBe('vcema(12,5)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const meta = new VelocityCorrectedExponentialMovingAverage({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.VelocityCorrectedExponentialMovingAverage);
      expect(meta.outputs.length).toBe(1);
    });
  });

  describe('updateScalar', () => {
    it('returns a single output', () => {
      const ind = new VelocityCorrectedExponentialMovingAverage({});
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of td.INPUT_CLOSE) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(1);
      const last = td.INPUT_CLOSE.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(td.EXPECTED_P6_D3[last], TOLERANCE);
    });
  });

  describe('invalid parameters', () => {
    it('throws when period < 2', () => {
      expect(() => new VelocityCorrectedExponentialMovingAverage({ period: 1 })).toThrowError();
    });
    it('throws when degree < 2', () => {
      expect(() => new VelocityCorrectedExponentialMovingAverage({ degree: 1 })).toThrowError();
    });
  });
});
