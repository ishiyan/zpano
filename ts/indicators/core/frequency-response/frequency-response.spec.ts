import { } from 'jasmine';

import { FrequencyResponse, FrequencyResponseFilter } from './frequency-response';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';

describe('FrequencyResponse', () => {
  it('validates signal length', () => {
    const isValid = (len: number): boolean => {
      switch (len) {
        case 4:
        case 8:
        case 16:
        case 32:
        case 64:
        case 128:
        case 256:
        case 512:
        case 1024:
        case 2048:
        case 4096:
        case 8192:
          return true;
        default:
          return false;
      }
    };

    const maxLength = 8199;
    for (let i = -1; i < maxLength; i++) {
      const expected = isValid(i);
      // eslint-disable-next-line @typescript-eslint/dot-notation
      const actual = FrequencyResponse['isValidSignalLength'](i);
      expect(actual).toBe(expected);
    }
  });

  it('prepares frequency domain points', () => {
    const l = 7;
    const expected = [1 / l, 2 / l, 3 / l, 4 / l, 5 / l, 6 / l, 7 / l];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['prepareFrequencyDomain'](l);

    for (let i = 0; i < expected.length - 1; i++) {
      expect(actual[i]).toBe(expected[i]);
    }
  });

  it('prepares filtered signal', () => {
    const len = 7;
    const warmup = 5;
    const expected = [1000, 0, 0, 0, 0, 0, 0];

    class IdentityFilter implements FrequencyResponseFilter {
      metadata(): { mnemonic: string } {
        return { mnemonic: 'identity' };
      }
      update(sample: number): number {
        return sample;
      }
    }

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['prepareFilteredSignal'](len, new IdentityFilter(), warmup);

    for (let i = 0; i < len - 1; i++) {
      expect(actual[i]).toBe(expected[i]);
    }
  });

  it('calculates FFT', () => {
    const expected = [16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
    const actual = [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['directRealFastFourierTransform'](actual);

    for (let i = 0; i < expected.length - 1; i++) {
      expect(actual[i]).toBe(expected[i]);
    }
  });

  it('converts array to percentages capping top to 200%', () => {
    const len = 5;
    const expected = [600 / 6, 200, 500 / 6, 400 / 6, 300 / 6];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['createFrequencyResponseComponent'](len);
    actual.data = [6, 70, 5, 4, 3];
    actual.min = 3;
    actual.max = 7;

    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['toPercents'](len, actual, actual);

    expect(actual.min).toBe(0);
    expect(actual.max).toBe(200);
    for (let i = 0; i < len; i++) {
      expect(actual.data[i]).toBe(expected[i]);
    }
  });

  it('converts array to percentages capping top to one of ranges', () => {
    const len = 5;
    const expected = [600 / 6, 700 / 6, 500 / 6, 400 / 6, 300 / 6];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['createFrequencyResponseComponent'](len);
    actual.data = [6, 7, 5, 4, 3];
    actual.min = 3;
    actual.max = 7;

    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['toPercents'](len, actual, actual);

    expect(actual.min).toBe(0);
    expect(actual.max).toBe(120);
    for (let i = 0; i < len; i++) {
      expect(actual.data[i]).toBe(expected[i]);
    }
  });

  it('converts array to percentages when max is zero', () => {
    const len = 5;
    const expected = [0, 700 / 7, 500 / 7, 400 / 7, 300 / 7];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['createFrequencyResponseComponent'](len);
    actual.data = [0, 7, 5, 4, 3];
    actual.min = 0;
    actual.max = 7;

    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['toPercents'](len, actual, actual);

    expect(actual.min).toBe(0);
    expect(actual.max).toBe(110);
    for (let i = 0; i < len; i++) {
      expect(actual.data[i]).toBe(expected[i]);
    }
  });

  it('converts array to decibels capping top and bottom to one of ranges', () => {
    const len = 5;
    const expected = [0, 20 * Math.log10(7 / 6), 20 * Math.log10(5 / 6), 20 * Math.log10(4 / 6), 20 * Math.log10(3 / 6)];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['createFrequencyResponseComponent'](len);
    actual.data = [6, 7, 5, 4, 3];
    actual.min = 3;
    actual.max = 7;

    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['toDecibels'](len, actual, actual);

    expect(actual.min).toBe(-10);
    expect(actual.max).toBe(5);
    for (let i = 0; i < len; i++) {
      expect(actual.data[i]).toBe(expected[i]);
    }
  });

  it('converts array to decibels capping top to 10 dB and bottom to -100 dB', () => {
    const len = 5;
    const expected = [0, 10, 20 * Math.log10(5 / 6), 20 * Math.log10(4 / 6), -100];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['createFrequencyResponseComponent'](len);
    actual.data = [6, 700, 5, 4, 0.00003];
    actual.min = 0.00003;
    actual.max = 700;

    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['toDecibels'](len, actual, actual);

    expect(actual.min).toBe(-100);
    expect(actual.max).toBe(10);
    for (let i = 0; i < len; i++) {
      expect(actual.data[i]).toBe(expected[i]);
    }
  });

  it('converts array to decibels when max is zero', () => {
    const len = 5;
    const expected = [-100, 0, 20 * Math.log10(5 / 7), 20 * Math.log10(4 / 7), 20 * Math.log10(3 / 7)];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['createFrequencyResponseComponent'](len);
    actual.data = [0, 7, 5, 4, 3];
    actual.min = 0;
    actual.max = 7;

    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['toDecibels'](len, actual, actual);

    expect(actual.min).toBe(-100);
    expect(actual.max).toBe(5);
    for (let i = 0; i < len; i++) {
      expect(actual.data[i]).toBe(expected[i]);
    }
  });

  it('unwraps phase degrees from [-180, 180] range', () => {
    const wrapped = [-80, -180, -90, -180, -90, 0, 90, 180, 90, 180, 80];
    const expected = [-80, -280, -280, -460, -460, -460, -460, -460, -640, -640, -840];

    // eslint-disable-next-line @typescript-eslint/dot-notation
    const actual = FrequencyResponse['createFrequencyResponseComponent'](wrapped.length);
    // eslint-disable-next-line @typescript-eslint/dot-notation
    FrequencyResponse['unwrapPhaseDegrees'](wrapped.length, wrapped, actual, 89);

    for (let i = 0; i < wrapped.length; i++) {
      expect(actual.data[i]).toBe(expected[i]);
    }
  });

  it('calculation throws if signal length is invalid', () => {
    expect(() => FrequencyResponse.calculate(129, new SimpleMovingAverage({length: 7}), 7)).toThrow();
  });

  it('calculates the response', () => {
    const len = 5;
    const actual = FrequencyResponse.calculate(16, new SimpleMovingAverage({length: len}), len);

    expect(actual.label).toEqual('sma(5)');

    const l = actual.frequencies.length;
    expect(actual.phaseDegrees.data.length).toEqual(l);
    expect(actual.powerDecibel.data.length).toEqual(l);
    expect(actual.powerPercent.data.length).toEqual(l);
    expect(actual.amplitudeDecibel.data.length).toEqual(l);
    expect(actual.amplitudePercent.data.length).toEqual(l);

    expect(actual.phaseDegrees.min).toEqual(-180);
    expect(actual.phaseDegrees.max).toEqual(180);
    for (let i = 0; i < l; i++) {
      expect(actual.phaseDegrees.data[i]).toBeGreaterThanOrEqual(-180);
      expect(actual.phaseDegrees.data[i]).toBeLessThanOrEqual(180);
    }

    expect(actual.powerDecibel.min).toEqual(-50);
    expect(actual.powerDecibel.max).toEqual(5);
    for (let i = 0; i < l; i++) {
      expect(actual.powerDecibel.data[i]).toBeGreaterThan(-50);
      expect(actual.powerDecibel.data[i]).toBeLessThanOrEqual(0);
    }

    expect(actual.powerPercent.min).toEqual(0);
    expect(actual.powerPercent.max).toEqual(110);
    for (let i = 0; i < l; i++) {
      expect(actual.powerPercent.data[i]).toBeGreaterThan(0);
      expect(actual.powerPercent.data[i]).toBeLessThanOrEqual(100);
    }

    expect(actual.amplitudeDecibel.min).toEqual(-30);
    expect(actual.amplitudeDecibel.max).toEqual(5);
    for (let i = 0; i < l; i++) {
      expect(actual.amplitudeDecibel.data[i]).toBeGreaterThan(-30);
      expect(actual.amplitudeDecibel.data[i]).toBeLessThanOrEqual(0);
    }

    expect(actual.amplitudePercent.min).toEqual(0);
    expect(actual.amplitudePercent.max).toEqual(110);
    for (let i = 0; i < l; i++) {
      expect(actual.amplitudePercent.data[i]).toBeGreaterThan(0);
      expect(actual.amplitudePercent.data[i]).toBeLessThanOrEqual(100);
    }
  });
});
