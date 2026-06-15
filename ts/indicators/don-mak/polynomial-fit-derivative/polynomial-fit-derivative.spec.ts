import { PolynomialFitDerivative } from './polynomial-fit-derivative';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import * as td from './testdata';

const TOLERANCE = 9;

// [degree, order, smoothing]
const combos: Array<[number, number, number]> = [
  [2, 1, 0], [2, 1, 3], [2, 1, 6], [2, 2, 0], [2, 2, 3], [2, 2, 6],
  [3, 1, 0], [3, 1, 3], [3, 1, 6], [3, 2, 0], [3, 2, 3], [3, 2, 6],
  [4, 1, 0], [4, 1, 3], [4, 1, 6], [4, 2, 0], [4, 2, 3], [4, 2, 6],
  [5, 1, 0], [5, 1, 3], [5, 1, 6], [5, 2, 0], [5, 2, 3], [5, 2, 6],
  [6, 1, 0], [6, 1, 3], [6, 1, 6], [6, 2, 0], [6, 2, 3], [6, 2, 6],
  [4, 3, 6], [5, 3, 6], [6, 3, 6], [6, 5, 6],
];

const testData = td as unknown as Record<string, number[]>;

describe('PolynomialFitDerivative', () => {
  describe('reference data', () => {
    combos.forEach(([degree, order, smoothing]) => {
      const name = `expectedD${degree}_O${order}_S${smoothing}`;
      it(`matches the reference for ${name}`, () => {
        const expected = testData[name];
        const ind = new PolynomialFitDerivative({ degree, order, smoothing });
        for (let i = 0; i < td.testInput.length; i++) {
          const value = ind.update(td.testInput[i]);
          if (isNaN(expected[i])) {
            expect(value).toBeNaN();
          } else {
            expect(value).toBeCloseTo(expected[i], TOLERANCE);
          }
        }
      });
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new PolynomialFitDerivative({});
      expect(ind.metadata().mnemonic).toBe('pfd(3,1,6)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new PolynomialFitDerivative({ degree: 4, order: 2, smoothing: 3 });
      expect(ind.metadata().mnemonic).toBe('pfd(4,2,3)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const ind = new PolynomialFitDerivative({});
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.PolynomialFitDerivative);
      expect(meta.outputs.length).toBe(1);
    });
  });

  describe('updateScalar', () => {
    it('returns a single output', () => {
      const ind = new PolynomialFitDerivative({});
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of td.testInput) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(1);
      const last = td.testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(td.expectedD3_O1_S6[last], TOLERANCE);
    });
  });

  describe('invalid parameters', () => {
    it('throws when degree < 2', () => {
      expect(() => new PolynomialFitDerivative({ degree: 1 })).toThrowError();
    });
    it('throws when order < 1', () => {
      expect(() => new PolynomialFitDerivative({ order: 0 })).toThrowError();
    });
    it('throws when order > degree', () => {
      expect(() => new PolynomialFitDerivative({ degree: 3, order: 4 })).toThrowError();
    });
    it('throws when smoothing < 0', () => {
      expect(() => new PolynomialFitDerivative({ smoothing: -1 })).toThrowError();
    });
  });
});
