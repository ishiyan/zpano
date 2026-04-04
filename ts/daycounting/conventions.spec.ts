import { } from 'jasmine';

import { DayCountConvention, fromString } from './conventions';

// ng test mb  --code-coverage --include='**/daycounting/conventions.spec.ts'
// ng test mb  --code-coverage --include='**/daycounting/*.spec.ts'

describe('DayCountConvention', () => {
    describe('fromString', () => {
        it('should convert valid string to convention - raw', () => {
            expect(fromString('raw')).toBe(DayCountConvention.RAW);
        });

        it('should convert valid string to convention - 30/360 us', () => {
            expect(fromString('30/360 us')).toBe(DayCountConvention.THIRTY_360_US);
        });

        it('should convert valid string to convention - 30/360 us eom', () => {
            expect(fromString('30/360 us eom')).toBe(DayCountConvention.THIRTY_360_US_EOM);
        });

        it('should convert valid string to convention - 30/360 us nasd', () => {
            expect(fromString('30/360 us nasd')).toBe(DayCountConvention.THIRTY_360_US_NASD);
        });

        it('should convert valid string to convention - 30/360 eu', () => {
            expect(fromString('30/360 eu')).toBe(DayCountConvention.THIRTY_360_EU);
        });

        it('should convert valid string to convention - 30/360 eu2', () => {
            expect(fromString('30/360 eu2')).toBe(DayCountConvention.THIRTY_360_EU_M2);
        });

        it('should convert valid string to convention - 30/360 eu3', () => {
            expect(fromString('30/360 eu3')).toBe(DayCountConvention.THIRTY_360_EU_M3);
        });

        it('should convert valid string to convention - 30/360 eu+', () => {
            expect(fromString('30/360 eu+')).toBe(DayCountConvention.THIRTY_360_EU_PLUS);
        });

        it('should convert valid string to convention - 30/365', () => {
            expect(fromString('30/365')).toBe(DayCountConvention.THIRTY_365);
        });

        it('should convert valid string to convention - act/360', () => {
            expect(fromString('act/360')).toBe(DayCountConvention.ACT_360);
        });

        it('should convert valid string to convention - act/365 fixed', () => {
            expect(fromString('act/365 fixed')).toBe(DayCountConvention.ACT_365_FIXED);
        });

        it('should convert valid string to convention - act/365 nonleap', () => {
            expect(fromString('act/365 nonleap')).toBe(DayCountConvention.ACT_365_NONLEAP);
        });

        it('should convert valid string to convention - act/act excel', () => {
            expect(fromString('act/act excel')).toBe(DayCountConvention.ACT_ACT_EXCEL);
        });

        it('should convert valid string to convention - act/act isda', () => {
            expect(fromString('act/act isda')).toBe(DayCountConvention.ACT_ACT_ISDA);
        });

        it('should convert valid string to convention - act/act afb', () => {
            expect(fromString('act/act afb')).toBe(DayCountConvention.ACT_ACT_AFB);
        });

        it('should be case insensitive - Act/Act Excel', () => {
            expect(fromString('Act/Act Excel')).toBe(DayCountConvention.ACT_ACT_EXCEL);
        });

        it('should be case insensitive - ACT/ACT AFB', () => {
            expect(fromString('ACT/ACT AFB')).toBe(DayCountConvention.ACT_ACT_AFB);
        });

        it('should be case insensitive - act/act ISDA', () => {
            expect(fromString('act/act ISDA')).toBe(DayCountConvention.ACT_ACT_ISDA);
        });

        it('should throw error for invalid convention', () => {
            expect(() => fromString('invalid convention')).toThrowError(
                /Day count convention 'invalid convention' must be one of:/
            );
        });
    });
});
