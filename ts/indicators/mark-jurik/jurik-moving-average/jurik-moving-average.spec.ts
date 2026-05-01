import { } from 'jasmine';

import { JurikMovingAverage } from './jurik-moving-average';
import {
    input,
    expectedLength20PhaseMin100, expectedLength20PhaseMin30, expectedLength20Phase0,
    expectedLength20Phase30, expectedLength20Phase100,
    expectedLength2Phase1, expectedLength5Phase1, expectedLength10Phase1,
    expectedLength25Phase1, expectedLength50Phase1, expectedLength75Phase1, expectedLength100Phase1,
} from './testdata';

describe('JurikMovingAverage', () => {

    it('should return expected mnemonic', () => {
        const jma = new JurikMovingAverage({length: 7, phase: -1});
        expect(jma.metadata().mnemonic).toBe('jma(7, -1)');
    });

    it('should throw if length is less than 1', () => {
        expect(() => { new JurikMovingAverage({length: 0, phase: 0}); }).toThrow();
    });

    it('should throw if phase is less than -100', () => {
        expect(() => { new JurikMovingAverage({length: 10, phase: -100.01}); }).toThrow();
    });

    it('should throw if phase is greater than 100', () => {
        expect(() => { new JurikMovingAverage({length: 10, phase: 100.01}); }).toThrow();
    });

    const verify = (length: number, phase: number, expected: Array<number>, outputName: string) => {
        const lenPrimed = 30;
        const epsilon = 1e-13;
        const jma = new JurikMovingAverage({length: length, phase: phase});

        for (let i = 0; i < lenPrimed; i++) {
            expect(jma.update(input[i])).toBeNaN();
            expect(jma.isPrimed()).toBe(false);
        }

        for (let i = lenPrimed; i < input.length; i++) {
            expect(jma.update(input[i])).withContext(`i = ${i}: ${outputName}`).toBeCloseTo(expected[i], epsilon);
            expect(jma.isPrimed()).toBe(true);
        }

        expect(jma.update(Number.NaN)).toBeNaN();
    };

    it('should calculate expected output and prime state is output data generated using original Jurik DLL', () => {
        verify(20, -100, expectedLength20PhaseMin100, 'jma(20,-100)');
        verify(20, -30, expectedLength20PhaseMin30, 'jma(20,-30)');
        verify(20, 0, expectedLength20Phase0, 'jma(20,0)');
        verify(20, 30, expectedLength20Phase30, 'jma(20,30)');
        verify(20, 100, expectedLength20Phase100, 'jma(20,100)');
        verify(2, 1, expectedLength2Phase1, 'jma(2,1)');
        verify(5, 1, expectedLength5Phase1, 'jma(5,1)');
        verify(10, 1, expectedLength10Phase1, 'jma(10,1)');
        verify(25, 1, expectedLength25Phase1, 'jma(25,1)');
        verify(50, 1, expectedLength50Phase1, 'jma(50,1)');
        verify(75, 1, expectedLength75Phase1, 'jma(75,1)');
        verify(100, 1, expectedLength100Phase1, 'jma(100,1)');
    });
});
