import { } from 'jasmine';

import { BalanceOfPower } from './balance-of-power';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { BalanceOfPowerOutput } from './balance-of-power-output';

// First 20 entries from the C# TA-Lib test data.
const inputOpen = [
  92.500, 91.500, 95.155, 93.970, 95.500, 94.500, 95.000, 91.500, 91.815, 91.125,
  93.875, 97.500, 98.815, 92.000, 91.125, 91.875, 93.405, 89.750, 89.345, 92.250,
];

const inputHigh = [
  93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000,
  96.250000, 99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000,
];

const inputLow = [
  90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000,
  92.750000, 96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000,
];

const inputClose = [
  91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
  96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
];

const expectedBop = [
  -0.400000000000000, 0.937765205091938, -0.367058823529412, 0.418215613382900, -0.540031397174254,
  0.102459016393443, -0.823333333333333, 0.314861460957179, -0.495049504950495, 0.632941176470588,
  0.642857142857143, -0.075528700906344, -0.101777059773828, -0.540025412960610, -0.027382256297919,
  0.406926406926406, -0.944444444444444, -0.216000000000001, 0.838235294117648, -0.950000000000000,
];

function roundTo(v: number, digits: number): number {
  const p = Math.pow(10, digits);
  return Math.round(v * p) / p;
}

describe('BalanceOfPower', () => {

  it('should calculate BOP from OHLC correctly', () => {
    const digits = 9;
    const bop = new BalanceOfPower();

    for (let i = 0; i < inputOpen.length; i++) {
      const v = bop.updateOHLC(inputOpen[i], inputHigh[i], inputLow[i], inputClose[i]);
      expect(v).not.toBeNaN();
      expect(bop.isPrimed()).toBe(true);
      expect(roundTo(v, digits)).toBe(roundTo(expectedBop[i], digits));
    }
  });

  it('should always be primed', () => {
    const bop = new BalanceOfPower();
    expect(bop.isPrimed()).toBe(true);

    bop.updateOHLC(92.5, 93.25, 90.75, 91.5);
    expect(bop.isPrimed()).toBe(true);
  });

  it('should pass NaN through', () => {
    const bop = new BalanceOfPower();
    expect(bop.update(Number.NaN)).toBeNaN();
    expect(bop.updateOHLC(Number.NaN, 1.0, 2.0, 3.0)).toBeNaN();
    expect(bop.updateOHLC(1.0, Number.NaN, 2.0, 3.0)).toBeNaN();
    expect(bop.updateOHLC(1.0, 2.0, Number.NaN, 3.0)).toBeNaN();
    expect(bop.updateOHLC(1.0, 2.0, 3.0, Number.NaN)).toBeNaN();
  });

  it('should return correct metadata', () => {
    const bop = new BalanceOfPower();
    const meta = bop.metadata();

    expect(meta.type).toBe(IndicatorType.BalanceOfPower);
    expect(meta.mnemonic).toBe('bop');
    expect(meta.description).toBe('Balance of Power');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(BalanceOfPowerOutput.BalanceOfPowerValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
  });

  it('should return 0 for zero range', () => {
    const bop = new BalanceOfPower();
    expect(bop.updateOHLC(0.001, 0.001, 0.001, 0.001)).toBe(0);
  });

  it('should return 0 for scalar update', () => {
    const bop = new BalanceOfPower();
    expect(bop.update(50.0)).toBe(0);
    expect(bop.update(100.0)).toBe(0);
  });
});
