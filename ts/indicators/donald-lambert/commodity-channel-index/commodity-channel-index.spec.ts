import { } from 'jasmine';

import { CommodityChannelIndex } from './commodity-channel-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { CommodityChannelIndexOutput } from './output';
import { input } from './testdata';

// Test data from TA-Lib (252 entries), used by MBST C# tests.
// Typical price input: test_CCI.xsl, TYPPRICE, F4…F255.
describe('CommodityChannelIndex', () => {

  it('should throw if length is less than 2', () => {
    expect(() => { new CommodityChannelIndex({ length: 1 }); }).toThrow();
  });

  it('should throw if length is negative', () => {
    expect(() => { new CommodityChannelIndex({ length: -8 }); }).toThrow();
  });

  it('should calculate CCI(11) correctly', () => {
    const tolerance = 5e-8;
    const cci = new CommodityChannelIndex({ length: 11 });

    // First 10 values NaN.
    for (let i = 0; i < 10; i++) {
      expect(cci.update(input[i])).toBeNaN();
    }

    // Index 10: first value.
    let v = cci.update(input[10]);
    expect(Math.abs(v - 87.92686612269590)).toBeLessThan(tolerance);

    // Index 11.
    v = cci.update(input[11]);
    expect(Math.abs(v - 180.00543014506300)).toBeLessThan(tolerance);

    // Feed remaining.
    for (let i = 12; i < 251; i++) {
      cci.update(input[i]);
    }

    // Last value.
    v = cci.update(input[251]);
    expect(Math.abs(v - (-169.65514382823800))).toBeLessThan(tolerance);
    expect(cci.isPrimed()).toBe(true);
  });

  it('should calculate CCI(2) correctly', () => {
    const tolerance = 5e-7;
    const cci = new CommodityChannelIndex({ length: 2 });

    // First value NaN.
    expect(cci.update(input[0])).toBeNaN();

    // Index 1: first value.
    let v = cci.update(input[1]);
    expect(Math.abs(v - 66.66666666666670)).toBeLessThan(tolerance);

    // Feed remaining.
    for (let i = 2; i < 251; i++) {
      cci.update(input[i]);
    }

    // Last value.
    v = cci.update(input[251]);
    expect(Math.abs(v - (-66.66666666666590))).toBeLessThan(tolerance);
  });

  it('should report correct primed state', () => {
    const cci = new CommodityChannelIndex({ length: 5 });

    expect(cci.isPrimed()).toBe(false);

    for (let i = 1; i <= 4; i++) {
      cci.update(i);
      expect(cci.isPrimed()).toBe(false);
    }

    cci.update(5);
    expect(cci.isPrimed()).toBe(true);

    cci.update(6);
    expect(cci.isPrimed()).toBe(true);
  });

  it('should pass NaN through', () => {
    const cci = new CommodityChannelIndex({ length: 5 });
    expect(cci.update(Number.NaN)).toBeNaN();
  });

  it('should return correct metadata', () => {
    const cci = new CommodityChannelIndex({ length: 20 });
    const meta = cci.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.CommodityChannelIndex);
    expect(meta.mnemonic).toBe('cci(20)');
    expect(meta.description).toBe('Commodity Channel Index cci(20)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(CommodityChannelIndexOutput.CommodityChannelIndexValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should accept custom inverse scaling factor', () => {
    const cci = new CommodityChannelIndex({ length: 5, inverseScalingFactor: 0.03 });

    for (let i = 1; i <= 5; i++) {
      cci.update(i);
    }

    expect(cci.isPrimed()).toBe(true);
  });
});
