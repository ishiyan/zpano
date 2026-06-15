import { MovingMiniMax } from './moving-mini-max';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import { Levels } from '../../core/outputs/levels';
import { Polyline } from '../../core/outputs/polyline';
import * as td from './testdata';

const TOLERANCE = 1e-9;

interface MiniMaxLevel {
  price: number;
  offset: number;
  strength: number;
}

interface Result {
  up: number;
  down: number;
  resistances: MiniMaxLevel[];
  supports: MiniMaxLevel[];
  upDist: number[];
  downDist: number[];
  valid: boolean;
}

function runLast(inputs: number[], m: number, n: number, numExtrema: number): Result {
  const ind = new MovingMiniMax({ m, n, numExtrema });
  let last: Result = {
    up: NaN, down: NaN, resistances: [], supports: [], upDist: [], downDist: [], valid: false,
  };
  for (const p of inputs) {
    const r = ind.update(p);
    if (ind.isPrimed()) {
      last = r;
    }
  }
  return last;
}

function expectSeries(actual: number[], expected: number[], label: string): void {
  expect(actual.length).withContext(`${label}: length`).toBe(expected.length);
  for (let i = 0; i < expected.length; i++) {
    const delta = TOLERANCE * Math.max(1, Math.abs(expected[i]));
    expect(Math.abs(actual[i] - expected[i]))
      .withContext(`${label}[${i}] expected ${expected[i]} got ${actual[i]}`)
      .toBeLessThanOrEqual(delta);
  }
}

function expectLevels(actual: MiniMaxLevel[], expected: td.Extremum[], label: string): void {
  expect(actual.length).withContext(`${label}: length`).toBe(expected.length);
  for (let i = 0; i < expected.length; i++) {
    expect(Math.abs(actual[i].price - expected[i].price))
      .withContext(`${label}[${i}].price`).toBeLessThanOrEqual(TOLERANCE * Math.max(1, Math.abs(expected[i].price)));
    expect(actual[i].offset).withContext(`${label}[${i}].offset`).toBe(expected[i].offset);
    expect(Math.abs(actual[i].strength - expected[i].strength))
      .withContext(`${label}[${i}].strength`).toBeLessThanOrEqual(TOLERANCE * Math.max(1, Math.abs(expected[i].strength)));
  }
}

function check(
  label: string, r: Result, expUp: number[], expDown: number[], expRes: td.Extremum[], expSup: td.Extremum[],
): void {
  expect(r.valid).withContext(`${label}: valid`).toBe(true);
  expectSeries(r.upDist, expUp, `${label} UP`);
  expectSeries(r.downDist, expDown, `${label} DOWN`);
  expectLevels(r.resistances, expRes, `${label} RES`);
  expectLevels(r.supports, expSup, `${label} SUP`);
}

interface Combo {
  name: string;
  m: number;
  n: number;
  e: number;
  up: number[];
  down: number[];
  res: td.Extremum[];
  sup: td.Extremum[];
}

const combos: Combo[] = [
  { name: 'm3_n50_e1', m: 3, n: 50, e: 1, up: td.expected_M3_N50_E1_Up, down: td.expected_M3_N50_E1_Down, res: td.expected_M3_N50_E1_Resistances, sup: td.expected_M3_N50_E1_Supports },
  { name: 'm3_n50_e3', m: 3, n: 50, e: 3, up: td.expected_M3_N50_E3_Up, down: td.expected_M3_N50_E3_Down, res: td.expected_M3_N50_E3_Resistances, sup: td.expected_M3_N50_E3_Supports },
  { name: 'm3_n100_e1', m: 3, n: 100, e: 1, up: td.expected_M3_N100_E1_Up, down: td.expected_M3_N100_E1_Down, res: td.expected_M3_N100_E1_Resistances, sup: td.expected_M3_N100_E1_Supports },
  { name: 'm3_n100_e3', m: 3, n: 100, e: 3, up: td.expected_M3_N100_E3_Up, down: td.expected_M3_N100_E3_Down, res: td.expected_M3_N100_E3_Resistances, sup: td.expected_M3_N100_E3_Supports },
  { name: 'm3_n252_e1', m: 3, n: 252, e: 1, up: td.expected_M3_N252_E1_Up, down: td.expected_M3_N252_E1_Down, res: td.expected_M3_N252_E1_Resistances, sup: td.expected_M3_N252_E1_Supports },
  { name: 'm3_n252_e3', m: 3, n: 252, e: 3, up: td.expected_M3_N252_E3_Up, down: td.expected_M3_N252_E3_Down, res: td.expected_M3_N252_E3_Resistances, sup: td.expected_M3_N252_E3_Supports },
  { name: 'm5_n50_e1', m: 5, n: 50, e: 1, up: td.expected_M5_N50_E1_Up, down: td.expected_M5_N50_E1_Down, res: td.expected_M5_N50_E1_Resistances, sup: td.expected_M5_N50_E1_Supports },
  { name: 'm5_n50_e3', m: 5, n: 50, e: 3, up: td.expected_M5_N50_E3_Up, down: td.expected_M5_N50_E3_Down, res: td.expected_M5_N50_E3_Resistances, sup: td.expected_M5_N50_E3_Supports },
  { name: 'm5_n100_e1', m: 5, n: 100, e: 1, up: td.expected_M5_N100_E1_Up, down: td.expected_M5_N100_E1_Down, res: td.expected_M5_N100_E1_Resistances, sup: td.expected_M5_N100_E1_Supports },
  { name: 'm5_n100_e3', m: 5, n: 100, e: 3, up: td.expected_M5_N100_E3_Up, down: td.expected_M5_N100_E3_Down, res: td.expected_M5_N100_E3_Resistances, sup: td.expected_M5_N100_E3_Supports },
  { name: 'm5_n252_e1', m: 5, n: 252, e: 1, up: td.expected_M5_N252_E1_Up, down: td.expected_M5_N252_E1_Down, res: td.expected_M5_N252_E1_Resistances, sup: td.expected_M5_N252_E1_Supports },
  { name: 'm5_n252_e3', m: 5, n: 252, e: 3, up: td.expected_M5_N252_E3_Up, down: td.expected_M5_N252_E3_Down, res: td.expected_M5_N252_E3_Resistances, sup: td.expected_M5_N252_E3_Supports },
  { name: 'm10_n50_e1', m: 10, n: 50, e: 1, up: td.expected_M10_N50_E1_Up, down: td.expected_M10_N50_E1_Down, res: td.expected_M10_N50_E1_Resistances, sup: td.expected_M10_N50_E1_Supports },
  { name: 'm10_n50_e3', m: 10, n: 50, e: 3, up: td.expected_M10_N50_E3_Up, down: td.expected_M10_N50_E3_Down, res: td.expected_M10_N50_E3_Resistances, sup: td.expected_M10_N50_E3_Supports },
  { name: 'm10_n100_e1', m: 10, n: 100, e: 1, up: td.expected_M10_N100_E1_Up, down: td.expected_M10_N100_E1_Down, res: td.expected_M10_N100_E1_Resistances, sup: td.expected_M10_N100_E1_Supports },
  { name: 'm10_n100_e3', m: 10, n: 100, e: 3, up: td.expected_M10_N100_E3_Up, down: td.expected_M10_N100_E3_Down, res: td.expected_M10_N100_E3_Resistances, sup: td.expected_M10_N100_E3_Supports },
  { name: 'm10_n252_e1', m: 10, n: 252, e: 1, up: td.expected_M10_N252_E1_Up, down: td.expected_M10_N252_E1_Down, res: td.expected_M10_N252_E1_Resistances, sup: td.expected_M10_N252_E1_Supports },
  { name: 'm10_n252_e3', m: 10, n: 252, e: 3, up: td.expected_M10_N252_E3_Up, down: td.expected_M10_N252_E3_Down, res: td.expected_M10_N252_E3_Resistances, sup: td.expected_M10_N252_E3_Supports },
  { name: 'm20_n50_e1', m: 20, n: 50, e: 1, up: td.expected_M20_N50_E1_Up, down: td.expected_M20_N50_E1_Down, res: td.expected_M20_N50_E1_Resistances, sup: td.expected_M20_N50_E1_Supports },
  { name: 'm20_n50_e3', m: 20, n: 50, e: 3, up: td.expected_M20_N50_E3_Up, down: td.expected_M20_N50_E3_Down, res: td.expected_M20_N50_E3_Resistances, sup: td.expected_M20_N50_E3_Supports },
  { name: 'm20_n100_e1', m: 20, n: 100, e: 1, up: td.expected_M20_N100_E1_Up, down: td.expected_M20_N100_E1_Down, res: td.expected_M20_N100_E1_Resistances, sup: td.expected_M20_N100_E1_Supports },
  { name: 'm20_n100_e3', m: 20, n: 100, e: 3, up: td.expected_M20_N100_E3_Up, down: td.expected_M20_N100_E3_Down, res: td.expected_M20_N100_E3_Resistances, sup: td.expected_M20_N100_E3_Supports },
  { name: 'm20_n252_e1', m: 20, n: 252, e: 1, up: td.expected_M20_N252_E1_Up, down: td.expected_M20_N252_E1_Down, res: td.expected_M20_N252_E1_Resistances, sup: td.expected_M20_N252_E1_Supports },
  { name: 'm20_n252_e3', m: 20, n: 252, e: 3, up: td.expected_M20_N252_E3_Up, down: td.expected_M20_N252_E3_Down, res: td.expected_M20_N252_E3_Resistances, sup: td.expected_M20_N252_E3_Supports },
];

describe('MovingMiniMax', () => {
  describe('reference data combos', () => {
    for (const c of combos) {
      it(c.name, () => check(c.name, runLast(td.testInput, c.m, c.n, c.e), c.up, c.down, c.res, c.sup));
    }
  });

  it('latest up/down equal distribution tails', () => {
    const r = runLast(td.testInput, 3, 50, 1);
    expect(Math.abs(r.up - r.upDist[r.upDist.length - 1])).toBeLessThanOrEqual(1e-12);
    expect(Math.abs(r.down - r.downDist[r.downDist.length - 1])).toBeLessThanOrEqual(1e-12);
  });

  it('default mnemonic', () => {
    const ind = new MovingMiniMax();
    expect(ind.metadata().mnemonic).toBe('mmm(5,50,3)');
  });

  it('metadata identifier and outputs', () => {
    const ind = new MovingMiniMax();
    const meta = ind.metadata();
    expect(meta.identifier).toBe(IndicatorIdentifier.MovingMiniMax);
    expect(meta.outputs.length).toBe(6);
  });

  it('updateScalar produces six outputs with polylines', () => {
    const ind = new MovingMiniMax({ m: 5, n: 50, numExtrema: 3 });
    let out = ind.updateScalar(Object.assign(new Scalar(), { time: new Date(0), value: 0 }));
    for (const p of td.testInput) {
      out = ind.updateScalar(Object.assign(new Scalar(), { time: new Date(0), value: p }));
    }
    expect(out.length).toBe(6);
    expect(out[2] instanceof Levels).toBe(true);
    expect(out[4] instanceof Polyline).toBe(true);
    expect((out[4] as Polyline).points.length).toBe(50);
  });

  it('throws on invalid params', () => {
    expect(() => new MovingMiniMax({ m: 0 })).toThrowError();
    expect(() => new MovingMiniMax({ m: 5, n: 10 })).toThrowError();
    expect(() => new MovingMiniMax({ numExtrema: 0 })).toThrowError();
  });
});
