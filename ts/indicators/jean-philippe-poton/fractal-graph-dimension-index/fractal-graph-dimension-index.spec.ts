import { } from 'jasmine';

import { FractalGraphDimensionIndex } from './fractal-graph-dimension-index';
import {
    testInput,
    expectedFgdiP5, expectedUpperP5, expectedLowerP5, expectedStddevP5,
    expectedFgdiP10, expectedUpperP10, expectedLowerP10, expectedStddevP10,
    expectedFgdiP15, expectedUpperP15, expectedLowerP15, expectedStddevP15,
    expectedFgdiP20, expectedUpperP20, expectedLowerP20, expectedStddevP20,
    expectedFgdiP30, expectedUpperP30, expectedLowerP30, expectedStddevP30,
    expectedFgdiP50, expectedUpperP50, expectedLowerP50, expectedStddevP50,
    expectedFgdiP80, expectedUpperP80, expectedLowerP80, expectedStddevP80,
    expectedFgdiP120, expectedUpperP120, expectedLowerP120, expectedStddevP120,
} from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('FractalGraphDimensionIndex', () => {

    function runTest(period: number, expFgdi: number[], expUpper: number[], expLower: number[], expStddev: number[]) {
        const ind = new FractalGraphDimensionIndex({ period });

        for (let i = 0; i < testInput.length; i++) {
            const [fgdi, upper, lower, stddev] = ind.updateAll(testInput[i]);
            if (Number.isNaN(expFgdi[i])) {
                expect(fgdi).toBeNaN();
            } else {
                expect(fgdi).toBeCloseTo(expFgdi[i], 13);
            }
            if (Number.isNaN(expUpper[i])) {
                expect(upper).toBeNaN();
            } else {
                expect(upper).toBeCloseTo(expUpper[i], 13);
            }
            if (Number.isNaN(expLower[i])) {
                expect(lower).toBeNaN();
            } else {
                expect(lower).toBeCloseTo(expLower[i], 13);
            }
            if (Number.isNaN(expStddev[i])) {
                expect(stddev).toBeNaN();
            } else {
                expect(stddev).toBeCloseTo(expStddev[i], 13);
            }
        }
    }

    it('should calculate expected output for period 5', () => {
        runTest(5, expectedFgdiP5, expectedUpperP5, expectedLowerP5, expectedStddevP5);
    });

    it('should calculate expected output for period 10', () => {
        runTest(10, expectedFgdiP10, expectedUpperP10, expectedLowerP10, expectedStddevP10);
    });

    it('should calculate expected output for period 15', () => {
        runTest(15, expectedFgdiP15, expectedUpperP15, expectedLowerP15, expectedStddevP15);
    });

    it('should calculate expected output for period 20', () => {
        runTest(20, expectedFgdiP20, expectedUpperP20, expectedLowerP20, expectedStddevP20);
    });

    it('should calculate expected output for period 30', () => {
        runTest(30, expectedFgdiP30, expectedUpperP30, expectedLowerP30, expectedStddevP30);
    });

    it('should calculate expected output for period 50', () => {
        runTest(50, expectedFgdiP50, expectedUpperP50, expectedLowerP50, expectedStddevP50);
    });

    it('should calculate expected output for period 80', () => {
        runTest(80, expectedFgdiP80, expectedUpperP80, expectedLowerP80, expectedStddevP80);
    });

    it('should calculate expected output for period 120', () => {
        runTest(120, expectedFgdiP120, expectedUpperP120, expectedLowerP120, expectedStddevP120);
    });

    it('should report primed state correctly', () => {
        const ind = new FractalGraphDimensionIndex({ period: 30 });
        for (let i = 0; i < 29; i++) {
            ind.update(testInput[i]);
            expect(ind.isPrimed()).toBe(false);
        }
        ind.update(testInput[29]);
        expect(ind.isPrimed()).toBe(true);
    });

    it('should pass NaN through', () => {
        const ind = new FractalGraphDimensionIndex({ period: 5 });
        const [fgdi, upper, lower, stddev] = ind.updateAll(Number.NaN);
        expect(fgdi).toBeNaN();
        expect(upper).toBeNaN();
        expect(lower).toBeNaN();
        expect(stddev).toBeNaN();
    });

    it('should throw if period is less than 2', () => {
        expect(() => { new FractalGraphDimensionIndex({ period: 1 }); }).toThrow();
    });

    it('should have correct identifier in metadata', () => {
        const ind = new FractalGraphDimensionIndex({ period: 30 });
        expect(ind.metadata().identifier).toBe(IndicatorIdentifier.FractalGraphDimensionIndex);
    });

    it('should return expected mnemonic', () => {
        const ind = new FractalGraphDimensionIndex({ period: 30 });
        expect(ind.metadata().mnemonic).toBe('fgdi(30)');
    });
});
