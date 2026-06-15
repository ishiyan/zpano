import { AdaptiveExponentialMovingAverage } from './adaptive-exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedDEFAULT,
  expectedA0_8_A0_02,
  expectedW0_5,
  expectedW1_5,
  expectedS0,
  expectedS6,
  expectedDEFAULT_OMEGA,
  expectedDEFAULT_ALPHA,
  test1InputSine,
  test1Expected,
  test1ExpectedOmega,
  test1ExpectedAlpha,
} from './testdata';

interface Combo {
  name: string;
  alphaMax: number;
  alphaMin: number;
  omega0: number;
  smoothing: number;
  expected: number[];
}

const valueCombos: Combo[] = [
  { name: 'DEFAULT', alphaMax: 0.5, alphaMin: 0.05, omega0: 1.0, smoothing: 3, expected: expectedDEFAULT },
  { name: 'A0_8_A0_02', alphaMax: 0.8, alphaMin: 0.02, omega0: 1.0, smoothing: 3, expected: expectedA0_8_A0_02 },
  { name: 'W0_5', alphaMax: 0.5, alphaMin: 0.05, omega0: 0.5, smoothing: 3, expected: expectedW0_5 },
  { name: 'W1_5', alphaMax: 0.5, alphaMin: 0.05, omega0: 1.5, smoothing: 3, expected: expectedW1_5 },
  { name: 'S0', alphaMax: 0.5, alphaMin: 0.05, omega0: 1.0, smoothing: 0, expected: expectedS0 },
  { name: 'S6', alphaMax: 0.5, alphaMin: 0.05, omega0: 1.0, smoothing: 6, expected: expectedS6 },
];

function expectSeries(actual: number[], expected: number[], label: string): void {
  expect(actual.length).toBe(expected.length);
  for (let i = 0; i < expected.length; i++) {
    if (isNaN(expected[i])) {
      expect(actual[i]).toBeNaN();
    } else {
      expect(actual[i]).toBeCloseTo(expected[i], 9);
    }
  }
}

describe('AdaptiveExponentialMovingAverage', () => {
  describe('value output', () => {
    valueCombos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new AdaptiveExponentialMovingAverage({
          alphaMax: combo.alphaMax, alphaMin: combo.alphaMin, omega0: combo.omega0, smoothing: combo.smoothing,
        });
        const values = testInput.map((c) => ind.update(c)[0]);
        expectSeries(values, combo.expected, `value(${combo.name})`);
      });
    });
  });

  describe('omega and alpha (default params)', () => {
    it('matches the reference omega and alpha series', () => {
      const ind = new AdaptiveExponentialMovingAverage();
      const omegas: number[] = [];
      const alphas: number[] = [];
      for (const c of testInput) {
        const [, omega, alpha] = ind.update(c);
        omegas.push(omega);
        alphas.push(alpha);
      }
      expectSeries(omegas, expectedDEFAULT_OMEGA, 'omega');
      expectSeries(alphas, expectedDEFAULT_ALPHA, 'alpha');
    });
  });

  describe('sine wave (default params)', () => {
    it('matches the reference value, omega and alpha', () => {
      const ind = new AdaptiveExponentialMovingAverage();
      const values: number[] = [];
      const omegas: number[] = [];
      const alphas: number[] = [];
      for (const c of test1InputSine) {
        const [v, o, a] = ind.update(c);
        values.push(v);
        omegas.push(o);
        alphas.push(a);
      }
      expectSeries(values, test1Expected, 'value');
      expectSeries(omegas, test1ExpectedOmega, 'omega');
      expectSeries(alphas, test1ExpectedAlpha, 'alpha');
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new AdaptiveExponentialMovingAverage();
      expect(ind.metadata().mnemonic).toBe('aema(0.50,0.05,1.00,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new AdaptiveExponentialMovingAverage({ alphaMax: 0.8, alphaMin: 0.02, omega0: 1.5, smoothing: 6 });
      expect(ind.metadata().mnemonic).toBe('aema(0.80,0.02,1.50,6)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and three outputs', () => {
      const ind = new AdaptiveExponentialMovingAverage();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.AdaptiveExponentialMovingAverage);
      expect(meta.outputs.length).toBe(3);
    });
  });

  describe('updateScalar', () => {
    it('returns [value, omega, alpha] in order', () => {
      const ind = new AdaptiveExponentialMovingAverage();
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of testInput) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      const last = testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(expectedDEFAULT[last], 9);
      expect((out[2] as Scalar).value).toBeCloseTo(expectedDEFAULT_ALPHA[last], 9);
    });
  });

  describe('invalid parameters', () => {
    it('throws when alpha_min >= alpha_max', () => {
      expect(() => new AdaptiveExponentialMovingAverage({ alphaMax: 0.05, alphaMin: 0.5 })).toThrowError();
    });
    it('throws when alpha_max > 1', () => {
      expect(() => new AdaptiveExponentialMovingAverage({ alphaMax: 1.5 })).toThrowError();
    });
    it('throws when omega0 >= pi', () => {
      expect(() => new AdaptiveExponentialMovingAverage({ omega0: 4.0 })).toThrowError();
    });
    it('throws when smoothing < 0', () => {
      expect(() => new AdaptiveExponentialMovingAverage({ smoothing: -1 })).toThrowError();
    });
  });
});
