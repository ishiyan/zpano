import { } from 'jasmine';

import { JurikFractalAdaptiveZeroLagVelocity } from './jurik-fractal-adaptive-zero-lag-velocity';
import {
    testInput,
    expectedLo2Hi15, expectedLo2Hi30, expectedLo2Hi60,
    expectedLo5Hi15, expectedLo5Hi30, expectedLo5Hi60,
    expectedLo10Hi15, expectedLo10Hi30, expectedLo10Hi60,
    expectedFtype2, expectedFtype3, expectedFtype4,
    expectedSmooth5, expectedSmooth20, expectedSmooth40,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikFractalAdaptiveZeroLagVelocity', () => {
    const cases: { loDepth: number; hiDepth: number; fractalType: number; smooth: number; expected: number[] }[] = [
        { loDepth: 2, hiDepth: 15, fractalType: 1, smooth: 10, expected: expectedLo2Hi15 },
        { loDepth: 2, hiDepth: 30, fractalType: 1, smooth: 10, expected: expectedLo2Hi30 },
        { loDepth: 2, hiDepth: 60, fractalType: 1, smooth: 10, expected: expectedLo2Hi60 },
        { loDepth: 5, hiDepth: 15, fractalType: 1, smooth: 10, expected: expectedLo5Hi15 },
        { loDepth: 5, hiDepth: 30, fractalType: 1, smooth: 10, expected: expectedLo5Hi30 },
        { loDepth: 5, hiDepth: 60, fractalType: 1, smooth: 10, expected: expectedLo5Hi60 },
        { loDepth: 10, hiDepth: 15, fractalType: 1, smooth: 10, expected: expectedLo10Hi15 },
        { loDepth: 10, hiDepth: 30, fractalType: 1, smooth: 10, expected: expectedLo10Hi30 },
        { loDepth: 10, hiDepth: 60, fractalType: 1, smooth: 10, expected: expectedLo10Hi60 },
        { loDepth: 5, hiDepth: 30, fractalType: 2, smooth: 10, expected: expectedFtype2 },
        { loDepth: 5, hiDepth: 30, fractalType: 3, smooth: 10, expected: expectedFtype3 },
        { loDepth: 5, hiDepth: 30, fractalType: 4, smooth: 10, expected: expectedFtype4 },
        { loDepth: 5, hiDepth: 30, fractalType: 1, smooth: 5, expected: expectedSmooth5 },
        { loDepth: 5, hiDepth: 30, fractalType: 1, smooth: 20, expected: expectedSmooth20 },
        { loDepth: 5, hiDepth: 30, fractalType: 1, smooth: 40, expected: expectedSmooth40 },
    ];

    for (const tc of cases) {
        it(`should compute JVELCFB with lo=${tc.loDepth} hi=${tc.hiDepth} ftype=${tc.fractalType} smooth=${tc.smooth}`, () => {
            const ind = new JurikFractalAdaptiveZeroLagVelocity({
                loDepth: tc.loDepth, hiDepth: tc.hiDepth,
                fractalType: tc.fractalType, smooth: tc.smooth,
            });
            for (let i = 0; i < testInput.length; i++) {
                const result = ind.update(testInput[i]);
                expect(almostEqual(result, tc.expected[i], epsilon))
                    .withContext(`bar ${i}: got ${result}, want ${tc.expected[i]}`)
                    .toBe(true);
            }
        });
    }
});
