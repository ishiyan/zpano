import { } from 'jasmine';

import { JurikCommodityChannelIndex } from './jurik-commodity-channel-index';
import {
    testInput,
    expectedLen10, expectedLen14, expectedLen20, expectedLen30,
    expectedLen40, expectedLen50, expectedLen60, expectedLen80, expectedLen100,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikCommodityChannelIndex', () => {
    const cases: { length: number; expected: number[] }[] = [
        { length: 10, expected: expectedLen10 },
        { length: 14, expected: expectedLen14 },
        { length: 20, expected: expectedLen20 },
        { length: 30, expected: expectedLen30 },
        { length: 40, expected: expectedLen40 },
        { length: 50, expected: expectedLen50 },
        { length: 60, expected: expectedLen60 },
        { length: 80, expected: expectedLen80 },
        { length: 100, expected: expectedLen100 },
    ];

    for (const tc of cases) {
        it(`should compute JCCX with length=${tc.length}`, () => {
            const jccx = new JurikCommodityChannelIndex({ length: tc.length });
            for (let i = 0; i < testInput.length; i++) {
                const result = jccx.update(testInput[i]);
                expect(almostEqual(result, tc.expected[i], epsilon))
                    .withContext(`bar ${i}: got ${result}, want ${tc.expected[i]}`)
                    .toBe(true);
            }
        });
    }
});
