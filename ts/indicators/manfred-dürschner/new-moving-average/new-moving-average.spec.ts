import { } from 'jasmine';

import { NewMovingAverage } from './new-moving-average';
import { NewMovingAverageParams, MAType } from './params';
import {
    testInput,
    expectedSec4PriAutoLWMA,
    expectedSec8PriAutoLWMA,
    expectedSec16PriAutoLWMA,
    expectedPri16Sec8LWMA,
    expectedPri32Sec8LWMA,
    expectedPri64Sec8LWMA,
    expectedPri8Sec4LWMA,
    expectedPri16Sec4LWMA,
    expectedPri32Sec4LWMA,
    expectedSec8SMA,
    expectedSec8EMA,
    expectedSec8SMMA,
} from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('NewMovingAverage', () => {

    function runTest(name: string, params: NewMovingAverageParams, expected: number[]) {
        it(name, () => {
            const nma = new NewMovingAverage(params);

            for (let i = 0; i < testInput.length; i++) {
                const act = nma.update(testInput[i]);
                const exp = expected[i];
                if (Number.isNaN(exp)) {
                    expect(act).toBeNaN();
                } else {
                    expect(act).toBeCloseTo(exp, 13);
                }
            }

            expect(nma.isPrimed()).toBe(true);
            expect(nma.update(NaN)).toBeNaN();
        });
    }

    runTest('sec=4 pri=auto LWMA', { primary_period: 0, secondary_period: 4, ma_type: MAType.LWMA }, expectedSec4PriAutoLWMA);
    runTest('sec=8 pri=auto LWMA (default)', { primary_period: 0, secondary_period: 8, ma_type: MAType.LWMA }, expectedSec8PriAutoLWMA);
    runTest('sec=16 pri=auto LWMA', { primary_period: 0, secondary_period: 16, ma_type: MAType.LWMA }, expectedSec16PriAutoLWMA);
    runTest('pri=16 sec=8 LWMA', { primary_period: 16, secondary_period: 8, ma_type: MAType.LWMA }, expectedPri16Sec8LWMA);
    runTest('pri=32 sec=8 LWMA', { primary_period: 32, secondary_period: 8, ma_type: MAType.LWMA }, expectedPri32Sec8LWMA);
    runTest('pri=64 sec=8 LWMA', { primary_period: 64, secondary_period: 8, ma_type: MAType.LWMA }, expectedPri64Sec8LWMA);
    runTest('pri=8 sec=4 LWMA', { primary_period: 8, secondary_period: 4, ma_type: MAType.LWMA }, expectedPri8Sec4LWMA);
    runTest('pri=16 sec=4 LWMA', { primary_period: 16, secondary_period: 4, ma_type: MAType.LWMA }, expectedPri16Sec4LWMA);
    runTest('pri=32 sec=4 LWMA', { primary_period: 32, secondary_period: 4, ma_type: MAType.LWMA }, expectedPri32Sec4LWMA);
    runTest('sec=8 SMA', { primary_period: 0, secondary_period: 8, ma_type: MAType.SMA }, expectedSec8SMA);
    runTest('sec=8 EMA', { primary_period: 0, secondary_period: 8, ma_type: MAType.EMA }, expectedSec8EMA);
    runTest('sec=8 SMMA', { primary_period: 0, secondary_period: 8, ma_type: MAType.SMMA }, expectedSec8SMMA);

    it('should return expected metadata', () => {
        const nma = new NewMovingAverage({ primary_period: 0, secondary_period: 8, ma_type: MAType.LWMA });
        const meta = nma.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.NewMovingAverage);
        expect(meta.mnemonic).toBe('nma(32, 8, 3)');
        expect(meta.description).toBe('New moving average nma(32, 8, 3)');
        expect(meta.outputs.length).toBe(1);
    });

    it('should not be primed before warmup', () => {
        const nma = new NewMovingAverage({ primary_period: 0, secondary_period: 8, ma_type: MAType.LWMA });
        expect(nma.isPrimed()).toBe(false);
        nma.update(100);
        expect(nma.isPrimed()).toBe(false);
    });

    it('should pass through NaN', () => {
        const nma = new NewMovingAverage({ primary_period: 0, secondary_period: 8, ma_type: MAType.LWMA });
        expect(nma.update(NaN)).toBeNaN();
    });
});
