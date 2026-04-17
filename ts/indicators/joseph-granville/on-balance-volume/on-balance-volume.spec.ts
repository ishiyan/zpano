import { } from 'jasmine';

import { OnBalanceVolume } from './on-balance-volume';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { OnBalanceVolumeOutput } from './on-balance-volume-output';

// C# test data: 12 entries.
const prices = [1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12];
const volumes = [100, 90, 200, 150, 500, 100, 300, 150, 100, 300, 200, 100];
const expected = [100, 190, 390, 240, 740, 640, 940, 1090, 990, 1290, 1090, 1190];

function roundTo(v: number, digits: number): number {
  const p = Math.pow(10, digits);
  return Math.round(v * p) / p;
}

describe('OnBalanceVolume', () => {

  it('should calculate OBV with real volume correctly', () => {
    const digits = 1;
    const obv = new OnBalanceVolume();

    for (let i = 0; i < prices.length; i++) {
      const v = obv.updateWithVolume(prices[i], volumes[i]);
      expect(v).not.toBeNaN();
      expect(obv.isPrimed()).toBe(true);
      expect(roundTo(v, digits)).toBe(roundTo(expected[i], digits));
    }
  });

  it('should be primed after first update', () => {
    const obv = new OnBalanceVolume();
    expect(obv.isPrimed()).toBe(false);

    obv.updateWithVolume(1.0, 100.0);
    expect(obv.isPrimed()).toBe(true);

    obv.updateWithVolume(2.0, 50.0);
    expect(obv.isPrimed()).toBe(true);
  });

  it('should pass NaN through', () => {
    const obv = new OnBalanceVolume();
    expect(obv.update(Number.NaN)).toBeNaN();
    expect(obv.updateWithVolume(1.0, Number.NaN)).toBeNaN();
    expect(obv.updateWithVolume(Number.NaN, Number.NaN)).toBeNaN();
  });

  it('should return correct metadata', () => {
    const obv = new OnBalanceVolume();
    const meta = obv.metadata();

    expect(meta.type).toBe(IndicatorType.OnBalanceVolume);
    expect(meta.mnemonic).toBe('obv');
    expect(meta.description).toBe('On-Balance Volume OBV');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(OnBalanceVolumeOutput.OnBalanceVolumeValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
  });

  it('should not change value when prices are equal', () => {
    const obv = new OnBalanceVolume();

    const v1 = obv.updateWithVolume(10.0, 100.0);
    expect(v1).toBe(100.0);

    const v2 = obv.updateWithVolume(10.0, 200.0);
    expect(v2).toBe(100.0);
  });

  it('should work with update (volume=1)', () => {
    const obv = new OnBalanceVolume();

    const v1 = obv.update(10.0);
    expect(v1).toBe(1); // first call: value = volume = 1

    const v2 = obv.update(15.0);
    expect(v2).toBe(2); // price up: value += 1

    const v3 = obv.update(12.0);
    expect(v3).toBe(1); // price down: value -= 1

    const v4 = obv.update(12.0);
    expect(v4).toBe(1); // price equal: unchanged
  });
});
