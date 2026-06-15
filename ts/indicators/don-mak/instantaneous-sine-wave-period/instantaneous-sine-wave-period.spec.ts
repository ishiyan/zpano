import { InstantaneousSineWavePeriod } from './instantaneous-sine-wave-period';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedS0_PERIOD, expectedS0_OMEGA, expectedS0_VELOCITY, expectedS0_ACCELERATION,
  expectedS3_PERIOD, expectedS3_OMEGA, expectedS3_VELOCITY, expectedS3_ACCELERATION,
  expectedS6_PERIOD, expectedS6_OMEGA, expectedS6_VELOCITY, expectedS6_ACCELERATION,
  expectedS12_PERIOD, expectedS12_OMEGA, expectedS12_VELOCITY, expectedS12_ACCELERATION,
} from './testdata';

interface Combo {
  name: string;
  smoothing: number;
  period: number[];
  omega: number[];
  velocity: number[];
  acceleration: number[];
}

const combos: Combo[] = [
  { name: 'S0', smoothing: 0, period: expectedS0_PERIOD, omega: expectedS0_OMEGA, velocity: expectedS0_VELOCITY, acceleration: expectedS0_ACCELERATION },
  { name: 'S3', smoothing: 3, period: expectedS3_PERIOD, omega: expectedS3_OMEGA, velocity: expectedS3_VELOCITY, acceleration: expectedS3_ACCELERATION },
  { name: 'S6', smoothing: 6, period: expectedS6_PERIOD, omega: expectedS6_OMEGA, velocity: expectedS6_VELOCITY, acceleration: expectedS6_ACCELERATION },
  { name: 'S12', smoothing: 12, period: expectedS12_PERIOD, omega: expectedS12_OMEGA, velocity: expectedS12_VELOCITY, acceleration: expectedS12_ACCELERATION },
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

describe('InstantaneousSineWavePeriod', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new InstantaneousSineWavePeriod({ smoothing: combo.smoothing });
        const periods: number[] = [];
        const omegas: number[] = [];
        const velocities: number[] = [];
        const accelerations: number[] = [];
        for (const c of testInput) {
          const [period, omega, velocity, acceleration] = ind.update(c);
          periods.push(period);
          omegas.push(omega);
          velocities.push(velocity);
          accelerations.push(acceleration);
        }
        expectSeries(periods, combo.period, `period(${combo.name})`);
        expectSeries(omegas, combo.omega, `omega(${combo.name})`);
        expectSeries(velocities, combo.velocity, `velocity(${combo.name})`);
        expectSeries(accelerations, combo.acceleration, `acceleration(${combo.name})`);
      });
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new InstantaneousSineWavePeriod();
      expect(ind.metadata().mnemonic).toBe('iswp(0,4.00,50.00,20.00,0.01)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new InstantaneousSineWavePeriod({ smoothing: 6 });
      expect(ind.metadata().mnemonic).toBe('iswp(6,4.00,50.00,20.00,0.01)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and seven outputs', () => {
      const ind = new InstantaneousSineWavePeriod();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.InstantaneousSineWavePeriod);
      expect(meta.outputs.length).toBe(7);
    });
  });

  describe('updateScalar', () => {
    it('returns seven outputs in order (period first)', () => {
      const ind = new InstantaneousSineWavePeriod();
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of testInput) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(7);
      const last = testInput.length - 1;
      if (isNaN(expectedS0_PERIOD[last])) {
        expect((out[0] as Scalar).value).toBeNaN();
      } else {
        expect((out[0] as Scalar).value).toBeCloseTo(expectedS0_PERIOD[last], 9);
      }
    });
  });

  describe('invalid parameters', () => {
    it('throws when smoothing < 0', () => {
      expect(() => new InstantaneousSineWavePeriod({ smoothing: -1 })).toThrowError();
    });
    it('throws when minPeriod <= 0', () => {
      expect(() => new InstantaneousSineWavePeriod({ minPeriod: 0 })).toThrowError();
    });
    it('throws when maxPeriod <= minPeriod', () => {
      expect(() => new InstantaneousSineWavePeriod({ minPeriod: 50, maxPeriod: 50 })).toThrowError();
    });
    it('throws when errorThreshold <= 0', () => {
      expect(() => new InstantaneousSineWavePeriod({ errorThreshold: 0 })).toThrowError();
    });
    it('throws when dx <= 0', () => {
      expect(() => new InstantaneousSineWavePeriod({ dx: 0 })).toThrowError();
    });
  });
});
