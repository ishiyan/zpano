import { } from 'jasmine';

import { HilbertTransformerPhaseAccumulator } from './phase-accumulator';
import { input, expectedSmoothed, expectedDetrended, expectedQuadrature, expectedInPhase, expectedPeriod } from './testdata-phase-accumulator';

// ng test mb  --code-coverage --include='**/indicators/**/*.spec.ts'
// ng test mb  --code-coverage --include='**/indicators/john-ehlers/hilbert-transformer/*.spec.ts'

/* eslint-disable max-len */

describe('HilbertTransformerPhaseAccumulator', () => {
  const epsilon = 1e-8;

  it('should throw if the smoothing length is less than 2', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 1, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 0.25 }
      );
    }).toThrow();
  });

  it('should throw if the smoothing length is greater than 4', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 5, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 0.25 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is 0', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0, alphaEmaPeriod: 0.25 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is negative', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: -0.001, alphaEmaPeriod: 0.25 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is 1', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 1, alphaEmaPeriod: 0.25 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is greater than 1', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 1.001, alphaEmaPeriod: 0.25 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is 0', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 0 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is negative', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: -0.001 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is 1', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 1 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is greater than 1', () => {
    expect(() => {
      new HilbertTransformerPhaseAccumulator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 1.001 }
      );
    }).toThrow();
  });

  it('should calculate expected output and prime state', () => {
    const lenPrimed = 18;
    const htpa = new HilbertTransformerPhaseAccumulator(
      { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 0.25 }
    );

    for (let i = 0; i < input.length; i++) {
      htpa.update(input[i]);

      if (i < lenPrimed) {
        expect(htpa.primed).withContext(`primed ${i}: expected false, actual ${htpa.primed}`)
          .toBe(false);
      } else {
        expect(htpa.primed).withContext(`primed ${i}: expected true, actual ${htpa.primed}`)
          .toBe(true);
      }

      let exp = expectedSmoothed[i];
      if (Number.isNaN(exp)) {
        expect(htpa.primed).withContext(`smoothed ${i} primed: expected false, actual true`)
          .toBe(false);
      } else {
        const act = htpa.smoothed;
        expect(act).withContext(`smoothed ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      // This should have been len(input), but after 24, the calculated
      // period is different from the expected data produced by homodyne
      // discriminator. This makes the detrended, quadrature, in-phase
      // and period data also different.
      const last = 24;
      exp = expectedDetrended[i];
      if (Number.isNaN(exp)) {
        expect(htpa.primed).withContext(`detrended ${i} primed: expected false, actual true`)
          .toBe(false);
      } else if (i < last) {
        const act = htpa.detrended;
        expect(act).withContext(`detrended ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      exp = expectedQuadrature[i];
      if (Number.isNaN(exp)) {
        expect(htpa.primed).withContext(`quadrature ${i} primed: expected false, actual true`)
          .toBe(false);
      } else if (i < last) {
        const act = htpa.quadrature;
        expect(act).withContext(`quadrature ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      exp = expectedInPhase[i];
      if (Number.isNaN(exp)) {
        expect(htpa.primed).withContext(`in-phase ${i} primed: expected false, actual true`)
          .toBe(false);
      } else if (i < last) {
        const act = htpa.inPhase;
        expect(act).withContext(`in-phase ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      exp = expectedPeriod[i];
      if (Number.isNaN(exp)) {
        if (i < lenPrimed) {
          expect(htpa.primed).withContext(`period ${i} primed: expected false, actual true`)
            .toBe(false);
        }
      } else if (i < 23) {
        const act = htpa.period;
        expect(act).withContext(`period ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }
    }

    const previousValue = htpa.period;
    htpa.update(Number.NaN);
    const newValue = htpa.period;
    expect(previousValue === newValue).withContext('updating with NaN should not change period')
      .toBeTrue();
  });

  it('should respect custom warmUpPeriod', () => {
    const lprimed = 50;
    const htpa = new HilbertTransformerPhaseAccumulator(
      { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 0.25, warmUpPeriod: lprimed }
    );

    expect(htpa.primed).withContext('before any update: expected false').toBe(false);

    for (let i = 0; i < lprimed; i++) {
      htpa.update(input[i]);
      expect(htpa.primed).withContext(`primed ${i + 1}: expected false, actual ${htpa.primed}`)
        .toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      htpa.update(input[i]);
      expect(htpa.primed).withContext(`primed ${i + 1}: expected true, actual ${htpa.primed}`)
        .toBe(true);
    }
  });

  const update = function (omega: number): HilbertTransformerPhaseAccumulator {
    const updates = 512;
    const hthd = new HilbertTransformerPhaseAccumulator(
      { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.15, alphaEmaPeriod: 0.25 }
    );

    for (let i = 0; i < updates; ++i) {
      hthd.update(Math.sin(omega * i));
    }

    return hthd;
  }

  it('should calculate correct period of sinusoid', () => {
    let period = 30;
    let omega = 2 * Math.PI / period;
    let exp = period;
    let act = update(omega).period;
    expect(Math.abs(act-exp)).withContext(
      `period ${period} (omega ${omega}) inside (min,max) -> period expected ${exp} actual ${act}`)
      .toBeLessThan(0.6);

    period = 3;
    omega = 2 * Math.PI / period;
    exp = 6;
    act = update(omega).period;
    expect(Math.abs(act-exp)).withContext(
      `period ${period} (omega ${omega}) < min -> period expected ${exp} actual ${act}`)
      .toBeLessThan(1.1);

    period = 60;
    omega = 2 * Math.PI / period;
    exp = 50;
    act = update(omega).period;
    expect(Math.abs(act-exp)).withContext(
      `period ${period} (omega ${omega}) < min -> period expected ${exp} actual ${act}`)
      .toBeLessThan(12.5);
  });
});
