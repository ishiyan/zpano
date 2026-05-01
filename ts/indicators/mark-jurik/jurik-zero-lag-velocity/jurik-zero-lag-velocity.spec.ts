import { } from 'jasmine';

import { JurikZeroLagVelocity } from './jurik-zero-lag-velocity';
import {
    testInput,
    expectedDepth2, expectedDepth3, expectedDepth4, expectedDepth5,
    expectedDepth6, expectedDepth7, expectedDepth8, expectedDepth9,
    expectedDepth10, expectedDepth11, expectedDepth12, expectedDepth13,
    expectedDepth14, expectedDepth15,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikZeroLagVelocity', () => {
    const cases: { depth: number; expected: number[] }[] = [
        { depth: 2, expected: expectedDepth2 },
        { depth: 3, expected: expectedDepth3 },
        { depth: 4, expected: expectedDepth4 },
        { depth: 5, expected: expectedDepth5 },
        { depth: 6, expected: expectedDepth6 },
        { depth: 7, expected: expectedDepth7 },
        { depth: 8, expected: expectedDepth8 },
        { depth: 9, expected: expectedDepth9 },
        { depth: 10, expected: expectedDepth10 },
        { depth: 11, expected: expectedDepth11 },
        { depth: 12, expected: expectedDepth12 },
        { depth: 13, expected: expectedDepth13 },
        { depth: 14, expected: expectedDepth14 },
        { depth: 15, expected: expectedDepth15 },
    ];

    for (const tc of cases) {
        it(`should compute VEL with depth=${tc.depth}`, () => {
            const vel = new JurikZeroLagVelocity({ depth: tc.depth });
            for (let i = 0; i < testInput.length; i++) {
                const result = vel.update(testInput[i]);
                expect(almostEqual(result, tc.expected[i], epsilon))
                    .withContext(`bar ${i}: got ${result}, want ${tc.expected[i]}`)
                    .toBe(true);
            }
        });
    }
});
