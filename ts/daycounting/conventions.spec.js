"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const conventions_1 = require("./conventions");
// ng test mb  --code-coverage --include='**/daycounting/conventions.spec.ts'
// ng test mb  --code-coverage --include='**/daycounting/*.spec.ts'
describe('DayCountConvention', () => {
    describe('fromString', () => {
        it('should convert valid string to convention - raw', () => {
            expect((0, conventions_1.fromString)('raw')).toBe(conventions_1.DayCountConvention.RAW);
        });
        it('should convert valid string to convention - 30/360 us', () => {
            expect((0, conventions_1.fromString)('30/360 us')).toBe(conventions_1.DayCountConvention.THIRTY_360_US);
        });
        it('should convert valid string to convention - 30/360 us eom', () => {
            expect((0, conventions_1.fromString)('30/360 us eom')).toBe(conventions_1.DayCountConvention.THIRTY_360_US_EOM);
        });
        it('should convert valid string to convention - 30/360 us nasd', () => {
            expect((0, conventions_1.fromString)('30/360 us nasd')).toBe(conventions_1.DayCountConvention.THIRTY_360_US_NASD);
        });
        it('should convert valid string to convention - 30/360 eu', () => {
            expect((0, conventions_1.fromString)('30/360 eu')).toBe(conventions_1.DayCountConvention.THIRTY_360_EU);
        });
        it('should convert valid string to convention - 30/360 eu2', () => {
            expect((0, conventions_1.fromString)('30/360 eu2')).toBe(conventions_1.DayCountConvention.THIRTY_360_EU_M2);
        });
        it('should convert valid string to convention - 30/360 eu3', () => {
            expect((0, conventions_1.fromString)('30/360 eu3')).toBe(conventions_1.DayCountConvention.THIRTY_360_EU_M3);
        });
        it('should convert valid string to convention - 30/360 eu+', () => {
            expect((0, conventions_1.fromString)('30/360 eu+')).toBe(conventions_1.DayCountConvention.THIRTY_360_EU_PLUS);
        });
        it('should convert valid string to convention - 30/365', () => {
            expect((0, conventions_1.fromString)('30/365')).toBe(conventions_1.DayCountConvention.THIRTY_365);
        });
        it('should convert valid string to convention - act/360', () => {
            expect((0, conventions_1.fromString)('act/360')).toBe(conventions_1.DayCountConvention.ACT_360);
        });
        it('should convert valid string to convention - act/365 fixed', () => {
            expect((0, conventions_1.fromString)('act/365 fixed')).toBe(conventions_1.DayCountConvention.ACT_365_FIXED);
        });
        it('should convert valid string to convention - act/365 nonleap', () => {
            expect((0, conventions_1.fromString)('act/365 nonleap')).toBe(conventions_1.DayCountConvention.ACT_365_NONLEAP);
        });
        it('should convert valid string to convention - act/act excel', () => {
            expect((0, conventions_1.fromString)('act/act excel')).toBe(conventions_1.DayCountConvention.ACT_ACT_EXCEL);
        });
        it('should convert valid string to convention - act/act isda', () => {
            expect((0, conventions_1.fromString)('act/act isda')).toBe(conventions_1.DayCountConvention.ACT_ACT_ISDA);
        });
        it('should convert valid string to convention - act/act afb', () => {
            expect((0, conventions_1.fromString)('act/act afb')).toBe(conventions_1.DayCountConvention.ACT_ACT_AFB);
        });
        it('should be case insensitive - Act/Act Excel', () => {
            expect((0, conventions_1.fromString)('Act/Act Excel')).toBe(conventions_1.DayCountConvention.ACT_ACT_EXCEL);
        });
        it('should be case insensitive - ACT/ACT AFB', () => {
            expect((0, conventions_1.fromString)('ACT/ACT AFB')).toBe(conventions_1.DayCountConvention.ACT_ACT_AFB);
        });
        it('should be case insensitive - act/act ISDA', () => {
            expect((0, conventions_1.fromString)('act/act ISDA')).toBe(conventions_1.DayCountConvention.ACT_ACT_ISDA);
        });
        it('should throw error for invalid convention', () => {
            expect(() => (0, conventions_1.fromString)('invalid convention')).toThrowError(/Day count convention 'invalid convention' must be one of:/);
        });
    });
});
//# sourceMappingURL=conventions.spec.js.map