import { } from 'jasmine';

import { JurikWaveletSampler } from './jurik-wavelet-sampler';
import {
    testInput,
    expectedWAVCol0, expectedWAVCol1, expectedWAVCol2, expectedWAVCol3,
    expectedWAVCol4, expectedWAVCol5, expectedWAVCol6, expectedWAVCol7,
    expectedWAVCol8, expectedWAVCol9, expectedWAVCol10, expectedWAVCol11,
    expectedIndex6Col0, expectedIndex6Col1, expectedIndex6Col2,
    expectedIndex6Col3, expectedIndex6Col4, expectedIndex6Col5,
    expectedIndex16Col0, expectedIndex16Col1, expectedIndex16Col2, expectedIndex16Col3,
    expectedIndex16Col4, expectedIndex16Col5, expectedIndex16Col6, expectedIndex16Col7,
    expectedIndex16Col8, expectedIndex16Col9, expectedIndex16Col10, expectedIndex16Col11,
    expectedIndex16Col12, expectedIndex16Col13, expectedIndex16Col14, expectedIndex16Col15,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikWaveletSampler', () => {
    it('should compute WAV with default index=12', () => {
        const ind = new JurikWaveletSampler({ index: 12 });
        const expectedCols = [
            expectedWAVCol0, expectedWAVCol1, expectedWAVCol2, expectedWAVCol3,
            expectedWAVCol4, expectedWAVCol5, expectedWAVCol6, expectedWAVCol7,
            expectedWAVCol8, expectedWAVCol9, expectedWAVCol10, expectedWAVCol11,
        ];

        for (let i = 0; i < testInput.length; i++) {
            const result = ind.update(testInput[i]);
            expect(almostEqual(result, expectedWAVCol0[i], epsilon))
                .withContext(`bar ${i} col0: got ${result}, want ${expectedWAVCol0[i]}`)
                .toBe(true);

            const cols = ind.columns();
            for (let c = 0; c < 12; c++) {
                expect(almostEqual(cols[c], expectedCols[c][i], epsilon))
                    .withContext(`bar ${i} col ${c}: got ${cols[c]}, want ${expectedCols[c][i]}`)
                    .toBe(true);
            }
        }
    });

    it('should compute WAV with index=6', () => {
        const ind = new JurikWaveletSampler({ index: 6 });
        const expectedCols = [
            expectedIndex6Col0, expectedIndex6Col1, expectedIndex6Col2,
            expectedIndex6Col3, expectedIndex6Col4, expectedIndex6Col5,
        ];

        for (let i = 0; i < testInput.length; i++) {
            ind.update(testInput[i]);
            const cols = ind.columns();
            for (let c = 0; c < 6; c++) {
                expect(almostEqual(cols[c], expectedCols[c][i], epsilon))
                    .withContext(`bar ${i} col ${c}: got ${cols[c]}, want ${expectedCols[c][i]}`)
                    .toBe(true);
            }
        }
    });

    it('should compute WAV with index=16', () => {
        const ind = new JurikWaveletSampler({ index: 16 });
        const expectedCols = [
            expectedIndex16Col0, expectedIndex16Col1, expectedIndex16Col2, expectedIndex16Col3,
            expectedIndex16Col4, expectedIndex16Col5, expectedIndex16Col6, expectedIndex16Col7,
            expectedIndex16Col8, expectedIndex16Col9, expectedIndex16Col10, expectedIndex16Col11,
            expectedIndex16Col12, expectedIndex16Col13, expectedIndex16Col14, expectedIndex16Col15,
        ];

        for (let i = 0; i < testInput.length; i++) {
            ind.update(testInput[i]);
            const cols = ind.columns();
            for (let c = 0; c < 16; c++) {
                expect(almostEqual(cols[c], expectedCols[c][i], epsilon))
                    .withContext(`bar ${i} col ${c}: got ${cols[c]}, want ${expectedCols[c][i]}`)
                    .toBe(true);
            }
        }
    });
});
