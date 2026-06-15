import { MexicanHatWavelet } from './mexican-hat-wavelet';
import { Band } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import * as td from './testdata';

const TOLERANCE = 9;

function expectSeries(ind: MexicanHatWavelet, inputs: number[], expected: number[], label: string): void {
  expect(inputs.length).toBe(expected.length);
  for (let i = 0; i < inputs.length; i++) {
    const value = ind.update(inputs[i]);
    if (isNaN(expected[i])) {
      expect(value).withContext(`${label}[${i}]`).toBeNaN();
    } else {
      expect(value).withContext(`${label}[${i}]`).toBeCloseTo(expected[i], TOLERANCE);
    }
  }
}

describe('MexicanHatWavelet', () => {
  describe('reference data', () => {
    it('matches the preset bands on INPUT_CLOSE', () => {
      expectSeries(new MexicanHatWavelet({ band: Band.High }), td.testInput, td.expectedHIGH, 'HIGH');
      expectSeries(new MexicanHatWavelet({ band: Band.Mid }), td.testInput, td.expectedMID, 'MID');
      expectSeries(new MexicanHatWavelet({ band: Band.Low }), td.testInput, td.expectedLOW, 'LOW');
    });

    it('matches the custom period bands on INPUT_CLOSE', () => {
      expectSeries(new MexicanHatWavelet({ band: Band.Custom, period: 8 }), td.testInput, td.expectedP8, 'P8');
      expectSeries(new MexicanHatWavelet({ band: Band.Custom, period: 20 }), td.testInput, td.expectedP20, 'P20');
      expectSeries(new MexicanHatWavelet({ band: Band.Custom, period: 32 }), td.testInput, td.expectedP32, 'P32');
    });

    it('matches the custom dilation bands on INPUT_CLOSE', () => {
      expectSeries(new MexicanHatWavelet({ band: Band.Custom, dilation: 2 }), td.testInput, td.expectedD2_0, 'D2_0');
      expectSeries(new MexicanHatWavelet({ band: Band.Custom, dilation: 8 }), td.testInput, td.expectedD8_0, 'D8_0');
    });

    it('matches the sine-wave MID series', () => {
      expectSeries(new MexicanHatWavelet({ band: Band.Mid }), td.test1InputSine, td.test1ExpectedMID, 'TEST1_MID');
    });

    it('matches the mixed-input bands', () => {
      expectSeries(new MexicanHatWavelet({ band: Band.High }), td.test2InputMixed, td.test2ExpectedHIGH, 'TEST2_HIGH');
      expectSeries(new MexicanHatWavelet({ band: Band.Mid }), td.test2InputMixed, td.test2ExpectedMID, 'TEST2_MID');
      expectSeries(new MexicanHatWavelet({ band: Band.Low }), td.test2InputMixed, td.test2ExpectedLOW, 'TEST2_LOW');
    });
  });

  describe('mnemonic', () => {
    it('formats the default (mid) mnemonic', () => {
      expect(new MexicanHatWavelet({}).metadata().mnemonic).toBe('mhw(mid)');
    });

    it('formats the preset mnemonics', () => {
      expect(new MexicanHatWavelet({ band: Band.High }).metadata().mnemonic).toBe('mhw(high)');
      expect(new MexicanHatWavelet({ band: Band.Low }).metadata().mnemonic).toBe('mhw(low)');
    });

    it('formats the custom mnemonics', () => {
      expect(new MexicanHatWavelet({ band: Band.Custom, dilation: 2 }).metadata().mnemonic).toBe('mhw(d2.00)');
      expect(new MexicanHatWavelet({ band: Band.Custom, period: 20 }).metadata().mnemonic).toBe('mhw(p20.00)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const meta = new MexicanHatWavelet({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.MexicanHatWavelet);
      expect(meta.outputs.length).toBe(1);
    });
  });

  describe('updateScalar', () => {
    it('returns a single output', () => {
      const ind = new MexicanHatWavelet({ band: Band.High });
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of td.testInput) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(1);
      const last = td.testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(td.expectedHIGH[last], TOLERANCE);
    });
  });

  describe('invalid parameters', () => {
    it('throws when custom has no dilation or period', () => {
      expect(() => new MexicanHatWavelet({ band: Band.Custom })).toThrowError();
    });
    it('throws when custom has both dilation and period', () => {
      expect(() => new MexicanHatWavelet({ band: Band.Custom, dilation: 2, period: 20 })).toThrowError();
    });
    it('throws when custom period <= 2', () => {
      expect(() => new MexicanHatWavelet({ band: Band.Custom, period: 2 })).toThrowError();
    });
    it('throws when custom dilation <= 0', () => {
      expect(() => new MexicanHatWavelet({ band: Band.Custom, dilation: -1 })).toThrowError();
    });
  });
});
