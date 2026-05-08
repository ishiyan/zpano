import { } from 'jasmine';

import { JurikTurningPointOscillator } from './jurik-turning-point-oscillator';
import {
    testInput,
    expectedLen5, expectedLen7, expectedLen10, expectedLen14,
    expectedLen20, expectedLen28, expectedLen40, expectedLen60, expectedLen80,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikTurningPointOscillator', () => {
    const cases: { length: number; expected: number[] }[] = [
        { length: 5, expected: expectedLen5 },
        { length: 7, expected: expectedLen7 },
        { length: 10, expected: expectedLen10 },
        { length: 14, expected: expectedLen14 },
        { length: 20, expected: expectedLen20 },
        { length: 28, expected: expectedLen28 },
        { length: 40, expected: expectedLen40 },
        { length: 60, expected: expectedLen60 },
        { length: 80, expected: expectedLen80 },
    ];

    for (const tc of cases) {
        it(`should compute JTPO with length=${tc.length}`, () => {
            const ind = new JurikTurningPointOscillator({ length: tc.length });
            for (let i = 0; i < testInput.length; i++) {
                const result = ind.update(testInput[i]);
                expect(almostEqual(result, tc.expected[i], epsilon))
                    .withContext(`bar ${i}: got ${result}, want ${tc.expected[i]}`)
                    .toBe(true);
            }
        });
    }
});
