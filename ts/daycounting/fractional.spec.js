"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fractional_1 = require("./fractional");
const conventions_1 = require("./conventions");
const daycounting_1 = require("./daycounting");
// ng test mb  --code-coverage --include='**/daycounting/fractional.spec.ts'
// ng test mb  --code-coverage --include='**/daycounting/*.spec.ts'
const SECONDS_IN_GREGORIAN_YEAR = 31556952;
const SECONDS_IN_LEAP_YEAR = 31622400;
const SECONDS_IN_NONLEAP_YEAR = 31536000;
describe('yearFrac', () => {
    describe('RAW method', () => {
        it('should calculate year fraction for leap year correctly', () => {
            const dateTime1 = new Date(2020, 0, 1, 0, 0, 0);
            const dateTime2 = new Date(2021, 0, 1, 0, 0, 0);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.RAW))
                .toBeCloseTo(SECONDS_IN_LEAP_YEAR / SECONDS_IN_GREGORIAN_YEAR, 16);
        });
        it('should calculate year fraction for non-leap year correctly', () => {
            const dateTime1 = new Date(2021, 0, 1, 0, 0, 0);
            const dateTime2 = new Date(2022, 0, 1, 0, 0, 0);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.RAW))
                .toBeCloseTo(SECONDS_IN_NONLEAP_YEAR / SECONDS_IN_GREGORIAN_YEAR, 16);
        });
    });
    describe('invalid method', () => {
        it('should throw error for invalid method', () => {
            const dateTime1 = new Date(2020, 0, 1, 0, 0, 0);
            const dateTime2 = new Date(2021, 0, 1, 0, 0, 0);
            // TypeScript will prevent passing invalid enum values,
            // but we can test with an invalid number cast
            expect(() => (0, fractional_1.yearFrac)(dateTime1, dateTime2, 999))
                .toThrowError(/Unknown day count convention: undefined/);
        });
    });
    describe('valid methods', () => {
        const y1 = 2020;
        ;
        const m1 = 1;
        const d1 = 1;
        const y2 = 2021;
        const m2 = 1;
        const d2 = 1;
        it('should calculate year fraction for THIRTY_360_US method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.us30360)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_360_US))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for THIRTY_360_US_EOM method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.us30360Eom)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_360_US_EOM))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for THIRTY_360_US_NASD method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.us30360Nasd)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_360_US_NASD))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for THIRTY_360_EU method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.eur30360)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_360_EU))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for THIRTY_360_EU_M2 method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.eur30360Model2)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_360_EU_M2))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for THIRTY_360_EU_M3 method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.eur30360Model3)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_360_EU_M3))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for THIRTY_360_EU_PLUS method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.eur30360Plus)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_360_EU_PLUS))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for THIRTY_365 method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.thirty365)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.THIRTY_365))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for ACT_360 method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.act360)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.ACT_360))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for ACT_365_FIXED method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.act365Fixed)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.ACT_365_FIXED))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for ACT_365_NONLEAP method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.act365Nonleap)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.ACT_365_NONLEAP))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for ACT_ACT_EXCEL method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.actActExcel)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.ACT_ACT_EXCEL))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for ACT_ACT_ISDA method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.actActIsda)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.ACT_ACT_ISDA))
                .toBeCloseTo(result, 16);
        });
        it('should calculate year fraction for ACT_ACT_AFB method', () => {
            const dateTime1 = new Date(y1, m1 - 1, d1, 0, 0, 0);
            const dateTime2 = new Date(y2, m2 - 1, d2, 0, 0, 0);
            const result = (0, daycounting_1.actActAfb)(y1, m1, d1, y2, m2, d2);
            expect((0, fractional_1.yearFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.ACT_ACT_AFB))
                .toBeCloseTo(result, 16);
        });
    });
});
describe('dayFrac', () => {
    describe('RAW method', () => {
        it('should calculate day fraction for leap year correctly', () => {
            const dateTime1 = new Date(2020, 0, 1, 0, 0, 0);
            const dateTime2 = new Date(2021, 0, 1, 0, 0, 0);
            expect((0, fractional_1.dayFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.RAW))
                .toBeCloseTo(366, 16);
        });
        it('should calculate day fraction for non-leap year correctly', () => {
            const dateTime1 = new Date(2021, 0, 1, 0, 0, 0);
            const dateTime2 = new Date(2022, 0, 1, 0, 0, 0);
            expect((0, fractional_1.dayFrac)(dateTime1, dateTime2, conventions_1.DayCountConvention.RAW))
                .toBeCloseTo(365, 16);
        });
    });
    describe('invalid method', () => {
        it('should throw error for invalid method', () => {
            const dateTime1 = new Date(2020, 0, 1, 0, 0, 0);
            const dateTime2 = new Date(2021, 0, 1, 0, 0, 0);
            // TypeScript will prevent passing invalid enum values,
            // but we can test with an invalid number cast
            expect(() => (0, fractional_1.dayFrac)(dateTime1, dateTime2, 999))
                .toThrowError(/Unknown day count convention: undefined/);
        });
    });
});
//# sourceMappingURL=fractional.spec.js.map