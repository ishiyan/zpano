import { DayCountConvention } from './conventions';
import {
    us30360, us30360Nasd, eur30360Plus, eur30360,
    us30360Eom, eur30360Model2, eur30360Model3,
    act365Fixed, act360, actActIsda, actActAfb,
    actActExcel, thirty365, act365Nonleap
} from './daycounting';

const SECONDS_IN_GREGORIAN_YEAR = 31_556_952;
const SECONDS_IN_DAY = 60 * 60 * 24;

/**
 * Calculates the fraction between two dates using a specified day count convention.
 * 
 * @param dateTime1 - The first date
 * @param dateTime2 - The second date
 * @param method - The day count convention to use
 * @param dayFrac - If true, returns fraction in days; if false, returns fraction in years
 * @returns The calculated fraction between the two dates
 */
export function frac(
    dateTime1: Date,
    dateTime2: Date,
    method: DayCountConvention,
    dayFrac: boolean
): number {
    let dt1 = dateTime1;
    let dt2 = dateTime2;
    
    if (dateTime1 > dateTime2) {
        [dt1, dt2] = [dt2, dt1];
    }

    if (method === DayCountConvention.RAW) {
        const diffSeconds = (dt2.getTime() - dt1.getTime()) / 1000;
        return diffSeconds / (dayFrac ? SECONDS_IN_DAY : SECONDS_IN_GREGORIAN_YEAR);
    }

    const y1 = dt1.getFullYear();
    const m1 = dt1.getMonth() + 1; // JS months are 0-indexed
    const d1 = dt1.getDate();

    const y2 = dt2.getFullYear();
    const m2 = dt2.getMonth() + 1;
    const d2 = dt2.getDate();

    // Time as a fraction of the day
    const tm1 = (dt1.getHours() * 3600 + dt1.getMinutes() * 60 + dt1.getSeconds()) / 86400;
    const tm2 = (dt2.getHours() * 3600 + dt2.getMinutes() * 60 + dt2.getSeconds()) / 86400;

    switch (method) {
        case DayCountConvention.THIRTY_360_US:
            return us30360(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.THIRTY_360_US_EOM:
            return us30360Eom(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.THIRTY_360_US_NASD:
            return us30360Nasd(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.THIRTY_360_EU:
            return eur30360(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.THIRTY_360_EU_M2:
            return eur30360Model2(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.THIRTY_360_EU_M3:
            return eur30360Model3(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.THIRTY_360_EU_PLUS:
            return eur30360Plus(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.THIRTY_365:
            return thirty365(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.ACT_360:
            return act360(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.ACT_365_FIXED:
            return act365Fixed(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.ACT_365_NONLEAP:
            return act365Nonleap(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.ACT_ACT_EXCEL:
            return actActExcel(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.ACT_ACT_ISDA:
            return actActIsda(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        case DayCountConvention.ACT_ACT_AFB:
            return actActAfb(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac);
        default:
            throw new Error(
                `Unknown day count convention: ${DayCountConvention[method]}`
            );
    }

}

/**
 * Calculates the year fraction between two dates using a specified day count convention.
 * 
 * @param dateTime1 - The first date
 * @param dateTime2 - The second date
 * @param method - The day count convention to use (defaults to RAW)
 * @returns The year fraction between the two dates
 */
export function yearFrac(
    dateTime1: Date,
    dateTime2: Date,
    method: DayCountConvention = DayCountConvention.RAW
): number {
    return frac(dateTime1, dateTime2, method, false);
}

/**
 * Calculates the day fraction between two dates using a specified day count convention.
 * 
 * @param dateTime1 - The first date
 * @param dateTime2 - The second date
 * @param method - The day count convention to use (defaults to RAW)
 * @returns The day fraction between the two dates
 */
export function dayFrac(
    dateTime1: Date,
    dateTime2: Date,
    method: DayCountConvention = DayCountConvention.RAW
): number {
    return frac(dateTime1, dateTime2, method, true);
}
