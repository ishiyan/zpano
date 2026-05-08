import { } from 'jasmine';

import { ArnaudLegouxMovingAverage } from './arnaud-legoux-moving-average';
import { ArnaudLegouxMovingAverageParams } from './params';
import {
    testInput,
    expectedW9_S6_O0_85,
    expectedW9_S6_O0_5,
    expectedW10_S6_O0_85,
    expectedW5_S6_O0_9,
    expectedW1_S6_O0_85,
    expectedW3_S6_O0_85,
    expectedW21_S6_O0_85,
    expectedW50_S6_O0_85,
    expectedW9_S6_O0,
    expectedW9_S6_O1,
    expectedW9_S2_O0_85,
    expectedW9_S20_O0_85,
    expectedW9_S0_5_O0_85,
    expectedW15_S4_O0_7,
} from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('ArnaudLegouxMovingAverage', () => {

    function runTest(name: string, params: ArnaudLegouxMovingAverageParams, expected: number[]) {
        it(name, () => {
            const alma = new ArnaudLegouxMovingAverage(params);
            const warmup = params.window === 1 ? 0 : params.window - 1;

            for (let i = 0; i < warmup; i++) {
                expect(alma.update(testInput[i])).toBeNaN();
                expect(alma.isPrimed()).toBe(false);
            }

            for (let i = warmup; i < testInput.length; i++) {
                const act = alma.update(testInput[i]);
                const exp = expected[i];
                if (Number.isNaN(exp)) {
                    expect(act).toBeNaN();
                } else {
                    expect(act).toBeCloseTo(exp, 13);
                }
                expect(alma.isPrimed()).toBe(true);
            }

            expect(alma.update(Number.NaN)).toBeNaN();
        });
    }

    runTest('w9 s6 o0.85 (default)', { window: 9, sigma: 6.0, offset: 0.85 }, expectedW9_S6_O0_85);
    runTest('w9 s6 o0.5', { window: 9, sigma: 6.0, offset: 0.5 }, expectedW9_S6_O0_5);
    runTest('w10 s6 o0.85', { window: 10, sigma: 6.0, offset: 0.85 }, expectedW10_S6_O0_85);
    runTest('w5 s6 o0.9', { window: 5, sigma: 6.0, offset: 0.9 }, expectedW5_S6_O0_9);
    runTest('w1 s6 o0.85', { window: 1, sigma: 6.0, offset: 0.85 }, expectedW1_S6_O0_85);
    runTest('w3 s6 o0.85', { window: 3, sigma: 6.0, offset: 0.85 }, expectedW3_S6_O0_85);
    runTest('w21 s6 o0.85', { window: 21, sigma: 6.0, offset: 0.85 }, expectedW21_S6_O0_85);
    runTest('w50 s6 o0.85', { window: 50, sigma: 6.0, offset: 0.85 }, expectedW50_S6_O0_85);
    runTest('w9 s6 o0', { window: 9, sigma: 6.0, offset: 0.0 }, expectedW9_S6_O0);
    runTest('w9 s6 o1', { window: 9, sigma: 6.0, offset: 1.0 }, expectedW9_S6_O1);
    runTest('w9 s2 o0.85', { window: 9, sigma: 2.0, offset: 0.85 }, expectedW9_S2_O0_85);
    runTest('w9 s20 o0.85', { window: 9, sigma: 20.0, offset: 0.85 }, expectedW9_S20_O0_85);
    runTest('w9 s0.5 o0.85', { window: 9, sigma: 0.5, offset: 0.85 }, expectedW9_S0_5_O0_85);
    runTest('w15 s4 o0.7', { window: 15, sigma: 4.0, offset: 0.7 }, expectedW15_S4_O0_7);

    it('should return expected metadata', () => {
        const alma = new ArnaudLegouxMovingAverage({ window: 9, sigma: 6, offset: 0.85 });
        const meta = alma.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.ArnaudLegouxMovingAverage);
        expect(meta.mnemonic).toBe('alma(9, 6, 0.85)');
        expect(meta.description).toBe('Arnaud Legoux moving average alma(9, 6, 0.85)');
        expect(meta.outputs.length).toBe(1);
        expect(meta.outputs[0].mnemonic).toBe('alma(9, 6, 0.85)');
    });

    it('should throw if window is less than 1', () => {
        expect(() => { new ArnaudLegouxMovingAverage({ window: 0, sigma: 6, offset: 0.85 }); }).toThrow();
    });

    it('should throw if sigma is not positive', () => {
        expect(() => { new ArnaudLegouxMovingAverage({ window: 9, sigma: 0, offset: 0.85 }); }).toThrow();
    });

    it('should throw if offset is out of range', () => {
        expect(() => { new ArnaudLegouxMovingAverage({ window: 9, sigma: 6, offset: -0.1 }); }).toThrow();
        expect(() => { new ArnaudLegouxMovingAverage({ window: 9, sigma: 6, offset: 1.1 }); }).toThrow();
    });
});
