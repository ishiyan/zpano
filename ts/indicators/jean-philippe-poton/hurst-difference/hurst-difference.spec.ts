import { } from 'jasmine';

import { HurstDifference } from './hurst-difference';
import {
    testInput,
    expectedFGDIP5, expectedHDIFFP5,
    expectedFGDIP10, expectedHDIFFP10,
    expectedFGDIP15, expectedHDIFFP15,
    expectedFGDIP20, expectedHDIFFP20,
    expectedFGDIP30, expectedHDIFFP30,
    expectedFGDIP50, expectedHDIFFP50,
    expectedFGDIP80, expectedHDIFFP80,
    expectedFGDIP120, expectedHDIFFP120,
} from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('HurstDifference', () => {

    function runTest(period: number, expFgdi: number[], expHdiff: number[]) {
        const ind = new HurstDifference({ period });

        for (let i = 0; i < testInput.length; i++) {
            const [hdiff, fgdi] = ind.updateAll(testInput[i]);
            if (Number.isNaN(expFgdi[i])) {
                expect(fgdi).withContext(`[${i}] fgdi`).toBeNaN();
            } else {
                expect(fgdi).withContext(`[${i}] fgdi`).toBeCloseTo(expFgdi[i], 13);
            }
            if (Number.isNaN(expHdiff[i])) {
                expect(hdiff).withContext(`[${i}] hdiff`).toBeNaN();
            } else {
                expect(hdiff).withContext(`[${i}] hdiff`).toBeCloseTo(expHdiff[i], 13);
            }
        }
    }

    it('should calculate expected output for period 5', () => {
        runTest(5, expectedFGDIP5, expectedHDIFFP5);
    });

    it('should calculate expected output for period 10', () => {
        runTest(10, expectedFGDIP10, expectedHDIFFP10);
    });

    it('should calculate expected output for period 15', () => {
        runTest(15, expectedFGDIP15, expectedHDIFFP15);
    });

    it('should calculate expected output for period 20', () => {
        runTest(20, expectedFGDIP20, expectedHDIFFP20);
    });

    it('should calculate expected output for period 30', () => {
        runTest(30, expectedFGDIP30, expectedHDIFFP30);
    });

    it('should calculate expected output for period 50', () => {
        runTest(50, expectedFGDIP50, expectedHDIFFP50);
    });

    it('should calculate expected output for period 80', () => {
        runTest(80, expectedFGDIP80, expectedHDIFFP80);
    });

    it('should calculate expected output for period 120', () => {
        runTest(120, expectedFGDIP120, expectedHDIFFP120);
    });

    it('should track primed state correctly', () => {
        const ind = new HurstDifference({ period: 30 });
        for (let i = 0; i < 30; i++) {
            ind.update(testInput[i]);
            expect(ind.isPrimed()).toBeFalse();
        }
        ind.update(testInput[30]);
        expect(ind.isPrimed()).toBeTrue();
    });

    it('should pass through NaN', () => {
        const ind = new HurstDifference({ period: 5 });
        const [hdiff, fgdi] = ind.updateAll(Number.NaN);
        expect(hdiff).toBeNaN();
        expect(fgdi).toBeNaN();
    });

    it('should throw for invalid period', () => {
        expect(() => new HurstDifference({ period: 1 })).toThrowError();
    });

    it('should return correct metadata', () => {
        const ind = new HurstDifference({ period: 30 });
        const meta = ind.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.HurstDifference);
        expect(meta.mnemonic).toContain('hurdif(30)');
    });
});
