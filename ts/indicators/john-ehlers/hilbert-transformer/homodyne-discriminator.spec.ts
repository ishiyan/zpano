import { } from 'jasmine';

import { HilbertTransformerHomodyneDiscriminator } from './homodyne-discriminator';
import { input, expectedSmoothed, expectedDetrended, expectedQuadrature, expectedInPhase, expectedPeriod } from './testdata-homodyne-discriminator';

// ng test mb  --code-coverage --include='**/indicators/**/*.spec.ts'
// ng test mb  --code-coverage --include='**/indicators/john-ehlers/hilbert-transformer/*.spec.ts'

/* eslint-disable max-len */

describe('HilbertTransformerHomodyneDiscriminator', () => {
  const epsilon = 1e-8;

  it('should throw if the smoothing length is less than 2', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 1, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 }
      );
    }).toThrow();
  });

  it('should throw if the smoothing length is greater than 4', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 5, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is 0', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0, alphaEmaPeriod: 0.2 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is negative', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: -0.001, alphaEmaPeriod: 0.2 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is 1', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 1, alphaEmaPeriod: 0.2 }
      );
    }).toThrow();
  });

  it('should throw if alpha quad is greater than 1', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 1.001, alphaEmaPeriod: 0.2 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is 0', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is negative', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: -0.001 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is 1', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 1 }
      );
    }).toThrow();
  });

  it('should throw if alpha period is greater than 1', () => {
    expect(() => {
      new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 1.001 }
      );
    }).toThrow();
  });

  it('should calculate expected output and prime state', () => {
    const lenPrimed = 4 + 7 * 3;
    const hthd = new HilbertTransformerHomodyneDiscriminator(
      { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 }
    );

    for (let i = 0; i < input.length; i++) {
      hthd.update(input[i]);

      if (i < lenPrimed) {
        expect(hthd.primed).withContext(`primed ${i}: expected false, actual true`)
          .toBe(false);
      } else {
        expect(hthd.primed).withContext(`primed ${i}: expected true, actual false`)
          .toBe(true);
      }

      let exp = expectedSmoothed[i];
      if (Number.isNaN(exp)) {
        expect(hthd.primed).withContext(`smoothed ${i} primed: expected false, actual true`)
          .toBe(false);
      } else {
        const act = hthd.smoothed;
        expect(act).withContext(`smoothed ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      exp = expectedDetrended[i];
      if (Number.isNaN(exp)) {
        expect(hthd.primed).withContext(`detrended ${i} primed: expected false, actual true`)
          .toBe(false);
      } else {
        const act = hthd.detrended;
        expect(act).withContext(`detrended ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      exp = expectedQuadrature[i];
      if (Number.isNaN(exp)) {
        expect(hthd.primed).withContext(`quadrature ${i} primed: expected false, actual true`)
          .toBe(false);
      } else {
        const act = hthd.quadrature;
        expect(act).withContext(`quadrature ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      exp = expectedInPhase[i];
      if (Number.isNaN(exp)) {
        expect(hthd.primed).withContext(`in-phase ${i} primed: expected false, actual true`)
          .toBe(false);
      } else {
        const act = hthd.inPhase;
        expect(act).withContext(`in-phase ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }

      exp = expectedPeriod[i];
      if (Number.isNaN(exp)) {
        expect(hthd.primed).withContext(`period ${i} primed: expected false, actual true`)
          .toBe(false);
      } else {
        const act = hthd.period;
        expect(act).withContext(`period ${i}: expected ${exp}, actual ${act}`)
          .toBeCloseTo(exp, epsilon);
      }
    }

    const previousValue = hthd.period;
    hthd.update(Number.NaN);
    const newValue = hthd.period;
    expect(previousValue === newValue).withContext('updating with NaN should not change period')
      .toBeTrue();
  });

  it('should respect custom warmUpPeriod', () => {
    const lprimed = 50;
    const hthd = new HilbertTransformerHomodyneDiscriminator(
      { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2, warmUpPeriod: lprimed }
    );

    expect(hthd.primed).withContext('before any update: expected false').toBe(false);

    for (let i = 0; i < lprimed; i++) {
      hthd.update(input[i]);
      expect(hthd.primed).withContext(`primed ${i + 1}: expected false, actual ${hthd.primed}`)
        .toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      hthd.update(input[i]);
      expect(hthd.primed).withContext(`primed ${i + 1}: expected true, actual ${hthd.primed}`)
        .toBe(true);
    }
  });

  const update = function (omega: number): HilbertTransformerHomodyneDiscriminator {
    const updates = 512;
    const hthd = new HilbertTransformerHomodyneDiscriminator(
      { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 }
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
    expect(act).withContext(
      `period ${period} (omega ${omega}) inside (min,max) -> period expected ${exp} actual ${act}`)
      .toBeCloseTo(exp, 1e-2);

    period = 3;
    omega = 2 * Math.PI / period;
    exp = 6;
    act = update(omega).period;
    expect(act).withContext(
      `period ${period} (omega ${omega}) < min -> period expected ${exp} actual ${act}`)
      .toBeCloseTo(exp, 1e-14);

    period = 60;
    omega = 2 * Math.PI / period;
    exp = 50;
    act = update(omega).period;
    expect(act).withContext(
      `period ${period} (omega ${omega}) < min -> period expected ${exp} actual ${act}`)
      .toBeCloseTo(exp, 1e-14);  
  });
});
