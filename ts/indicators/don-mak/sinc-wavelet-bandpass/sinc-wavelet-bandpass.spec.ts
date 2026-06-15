import { SincWaveletBandpass } from './sinc-wavelet-bandpass';
import { Band } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import * as td from './testdata';

const TOLERANCE = 9;

function expectSeries(ind: SincWaveletBandpass, inputs: number[], expected: number[], label: string): void {
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

describe('SincWaveletBandpass', () => {
  describe('reference data', () => {
    it('matches the bands without velocity on INPUT_CLOSE', () => {
      expectSeries(new SincWaveletBandpass({ band: Band.High }), td.testInput, td.expectedHIGH, 'HIGH');
      expectSeries(new SincWaveletBandpass({ band: Band.Mid }), td.testInput, td.expectedMID, 'MID');
      expectSeries(new SincWaveletBandpass({ band: Band.Low }), td.testInput, td.expectedLOW, 'LOW');
      expectSeries(new SincWaveletBandpass({ band: Band.Full }), td.testInput, td.expectedFULL, 'FULL');
    });

    it('matches the bands with velocity on INPUT_CLOSE', () => {
      expectSeries(new SincWaveletBandpass({ band: Band.High, velocity: true }), td.testInput, td.expectedHIGH_V, 'HIGH_V');
      expectSeries(new SincWaveletBandpass({ band: Band.Mid, velocity: true }), td.testInput, td.expectedMID_V, 'MID_V');
      expectSeries(new SincWaveletBandpass({ band: Band.Low, velocity: true }), td.testInput, td.expectedLOW_V, 'LOW_V');
      expectSeries(new SincWaveletBandpass({ band: Band.Full, velocity: true }), td.testInput, td.expectedFULL_V, 'FULL_V');
    });

    it('matches the sine-wave MID series', () => {
      expectSeries(new SincWaveletBandpass({ band: Band.Mid }), td.test1InputSine, td.test1ExpectedMID, 'TEST1_MID');
    });

    it('matches the mixed-input velocity bands', () => {
      expectSeries(new SincWaveletBandpass({ band: Band.High, velocity: true }), td.test2InputMixed, td.test2ExpectedHIGH_V, 'TEST2_HIGH_V');
      expectSeries(new SincWaveletBandpass({ band: Band.Mid, velocity: true }), td.test2InputMixed, td.test2ExpectedMID_V, 'TEST2_MID_V');
      expectSeries(new SincWaveletBandpass({ band: Band.Low, velocity: true }), td.test2InputMixed, td.test2ExpectedLOW_V, 'TEST2_LOW_V');
    });
  });

  describe('mnemonic', () => {
    it('formats the default (mid) mnemonic', () => {
      expect(new SincWaveletBandpass({}).metadata().mnemonic).toBe('swb(mid)');
    });

    it('formats the band mnemonics', () => {
      expect(new SincWaveletBandpass({ band: Band.High }).metadata().mnemonic).toBe('swb(high)');
      expect(new SincWaveletBandpass({ band: Band.Full }).metadata().mnemonic).toBe('swb(full)');
    });

    it('formats the velocity mnemonics', () => {
      expect(new SincWaveletBandpass({ band: Band.Mid, velocity: true }).metadata().mnemonic).toBe('swb(mid,v)');
      expect(new SincWaveletBandpass({ band: Band.Full, velocity: true }).metadata().mnemonic).toBe('swb(full,v)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const meta = new SincWaveletBandpass({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.SincWaveletBandpass);
      expect(meta.outputs.length).toBe(1);
    });
  });

  describe('updateScalar', () => {
    it('returns a single output', () => {
      const ind = new SincWaveletBandpass({ band: Band.High });
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
});
