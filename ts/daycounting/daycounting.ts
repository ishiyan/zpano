// Wikipedia
// https://en.wikipedia.org/wiki/Day_count_convention

// ISDA 2006 Definitions, Section 4.16 page 11
// https://web.archive.org/web/20140913145444/http://www.hsbcnet.com/gbm/attachments/standalone/2006-isda-definitions.pdf

// For Excel YEARFRAC function see
// https://support.microsoft.com/en-us/office/yearfrac-function-3844141e-c76d-4143-82b6-208454ddc6a8
//
// Excel YEARFRAC function:
// Basis Optional: The type of day count basis to use.
// 0: US (NASD) 30/360 (default is not set)
// 1: Actual/actual
// 2: Actual/360
// 3: Actual/365
// 4: European 30/360

// Day counting methods are listed in the ISO 20022, see
// https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm

// Source code
// https://github.com/devind-team/devind_yearfrac
// https://github.com/hcnn/d30360s
// https://github.com/hcnn/d30360e2
// https://github.com/hcnn/d30360e3
// https://github.com/hcnn/d30360p
// https://github.com/hcnn/d30360u
// https://github.com/hcnn/d30360m
// https://github.com/hcnn/d30360n
// https://github.com/hcnn/d30365
// https://github.com/hcnn/act365n
// https://github.com/hcnn/act365f
// https://github.com/hcnn/act360
// https://github.com/hcnn/act_isda
// https://github.com/hcnn/act_afb
// https://github.com/AnatolyBuga/yearfrac

export function isLeapYear(y: number): boolean {
    return !(y % 4) && (Boolean(y % 100) || !(y % 400));
}

export function dateToJd(year: number, month: number, day: number): number {
    const a = Math.floor((14 - month) / 12);
    const y = Math.floor(year + 4800 - a);
    const m = Math.floor(month + (12 * a) - 3);

    let jd = day + Math.floor(((153 * m) + 2) / 5.0) + (y * 365);
    jd += Math.floor(y / 4) - Math.floor(y / 100) + Math.floor(y / 400) - 32045;
    return jd;
}

export function jdToDate(jd: number): [number, number, number] {
    const a = jd + 32044;
    const b = Math.floor(((4 * a) + 3) / 146097);
    const c = a - Math.floor((b * 146097) / 4);

    const d = Math.floor(((4 * c) + 3) / 1461);
    const e = c - Math.floor((d * 1461) / 4);
    const m = Math.floor(((5 * e) + 2) / 153);
    const m2 = Math.floor(m / 10);

    const day = e + 1 - Math.floor(((153 * m) + 2) / 5);
    const month = (m + 3 - (12 * m2));
    const year = ((b * 100) + d - 4800 + m2);

    return [year, month, day];
}

/**
 * Source:
 *     https://github.com/hcnn/d30360s
 * Synonyms:
 *     - 30/360 ICMA
 *     - 30/360 Eurobond Basis
 *     - ISDA-2006
 *     - 30S/360 Special German
 * 
 * ISO 20022:
 *     A011
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 *
 * Method whereby interest is calculated based on a 30-day month
 * and a 360-day year.
 * 
 * Accrued interest to a value date on the last day of a month
 * shall be the same as to the 30th calendar day of the same month,
 * except for February.
 * 
 * This means that a 31st is assumed to be a 30th and the 28 Feb
 * (or 29 Feb for a leap year) is assumed to be a 28th (or 29th).
 * 
 * It is the most commonly used 30/360 method for non-US straight
 * and convertible bonds issued before 01/01/1999.
 */
export function eur30360(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    diffDays += d2 > 30 ? 30 : d2;
    diffDays -= d1 > 30 ? 30 : d1;
    return fracDays ? diffDays : diffDays / 360;
}

/**
 * Source:
 *     https://github.com/hcnn/d30360e2
 * Synonyms:
 *     - 30E2/360
 *     - Eurobond basis model 2
 * 
 * ISO 20022:
 *     A012
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function eur30360Model2(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    const leap1 = isLeapYear(y1);
    if (leap1 && (m2 === 2) && (d2 === 28)) {
        diffDays += d1 === 29 ? 29 : (d1 >= 30 ? 30 : d2);
    } else if (leap1 && (m2 === 2) && (d2 === 29)) {
        diffDays += d1 >= 30 ? 30 : d2;
    } else {
        diffDays += d2 > 30 ? 30 : d2;
    }
    diffDays -= d1 > 30 ? 30 : d1;
    return fracDays ? diffDays : diffDays / 360;
}

/**
 * Source:
 *     https://github.com/hcnn/d30360e3
 * Synonyms:
 *     - 30E3/360
 *     - Eurobond basis model 3
 * 
 * ISO 20022:
 *     A013
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function eur30360Model3(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    if ((m2 === 2) && (d2 >= 28)) {
        diffDays += 30;
    } else {
        diffDays += d2 > 30 ? 30 : d2;
    }
    if ((m1 === 2) && (d1 >= 28)) {
        diffDays -= 30;
    } else {
        diffDays -= d1 > 30 ? 30 : d1;
    }
    return fracDays ? diffDays : diffDays / 360;
}

/**
 * Source:
 *     https://github.com/hcnn/d30360p
 * Synonyms:
 *     - 30E+/360
 */
export function eur30360Plus(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    diffDays += d2 === 31 ? 32 : d2;
    diffDays -= d1 > 30 ? 30 : d1;
    return fracDays ? diffDays : diffDays / 360;
}

/**
 * Source:
 *     https://github.com/hcnn/d30360u
 * Synonyms:
 *     - 30/360 ISDA
 *     - 30U/360
 *     - 30/360 US
 *     - 30/360 Bond Basis
 *     - 30/360 U.S. Municipal
 *     - American Basic Rule
 * 
 * ISO 20022:
 *     A001
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function us30360(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    if ((d2 === 31) && (d1 >= 30)) {
        diffDays += 30;
    } else {
        diffDays += d2;
    }
    diffDays -= d1 > 30 ? 30 : d1;
    return fracDays ? diffDays : diffDays / 360;
}

/**
 * Source:
 *     https://github.com/hcnn/d30360m
 * Synonyms:
 *     - 30/360 US EOM
 */
export function us30360Eom(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    const rule2 = (m1 === 2) && (d1 >= 28);
    const rule3 = rule2 && (m2 === 2) && (d2 >= 28);
    const rule4 = (d2 === 31) && (d1 >= 30);
    if (rule2) {
        diffDays -= 30;
    } else {
        diffDays -= d1 > 30 ? 30 : d1;
    }
    if (rule4 || rule3) {
        diffDays += 30;
    } else {
        diffDays += d2;
    }
    return fracDays ? diffDays : diffDays / 360;
}

/**
 * Source:
 *     https://github.com/hcnn/d30360n
 * Synonyms:
 *     - 30/360 NASD
 */
export function us30360Nasd(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    if (d2 === 31) {
        diffDays += d1 < 30 ? 32 : 30;
    } else {
        diffDays += d2;
    }
    diffDays -= d1 > 30 ? 30 : d1;
    return fracDays ? diffDays : diffDays / 360;
}

/**
 * Source:
 *     https://github.com/hcnn/d30365
 * Synonyms:
 *     - 30/365
 * 
 * ISO 20022:
 *     A002
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function thirty365(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = 360 * (y2 - y1) + 30 * (m2 - m1) + df2 - df1;
    if (d2 === 31 && d1 >= 30) {
        diffDays += 30;
    } else {
        diffDays += d2;
    }
    diffDays -= d1 > 30 ? 30 : d1;
    return fracDays ? diffDays : diffDays / 365;
}

/**
 * Source:
 *     https://github.com/hcnn/act365n
 * Synonyms:
 *     - Actual/365NL
 *     - Actual/365 Non-Leap
 * 
 * ISO 20022:
 *     A014
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function act365Nonleap(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = dateToJd(y2, m2, d2) - dateToJd(y1, m1, d1) + df2 - df1;
    let leapYears = 0;
    if (isLeapYear(y1) && (m1 <= 2)) {
        leapYears += 1;
    }
    if ((y1 !== y2) && isLeapYear(y2) && (m2 >= 3)) {
        leapYears += 1;
    }
    if ((y1 + 1) < y2) {
        let now = y1 + 1;
        while (now < y2) {
            if (isLeapYear(now)) {
                leapYears += 1;
            }
            now += 1;
        }
    }
    diffDays -= leapYears;
    return fracDays ? diffDays : diffDays / 365;
}

/**
 * Source:
 *     https://github.com/hcnn/act365f
 * Synonyms:
 *     - Actual/365 Fixed
 *     - Act/365 Fixed
 *     - A/365 Fixed
 *     - A/365F
 *     - English
 * 
 * ISO 20022:
 *     A005
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function act365Fixed(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = dateToJd(y2, m2, d2) - dateToJd(y1, m1, d1);
    diffDays += df2 - df1;
    return fracDays ? diffDays : diffDays / 365;
}

/**
 * Source:
 *     https://github.com/hcnn/act360
 * Synonyms:
 *     - Actual/360
 *     - Act/360
 *     - A/360
 *     - French
 * 
 * ISO 20022:
 *     A004
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function act360(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    let diffDays = dateToJd(y2, m2, d2) - dateToJd(y1, m1, d1);
    diffDays += df2 - df1;
    return fracDays ? diffDays : diffDays / 360;
}

function feb29Between(date1: Date, y1: number, date2: Date, y2: number): boolean {
    // Check each year in the range
    for (let y = y1; y <= y2; y++) {
        if (isLeapYear(y)) {
            const leapDay = new Date(y, 1, 29); // Month is 0-indexed in JS
            if (date1 <= leapDay && leapDay <= date2) {
                return true;
            }
        }
    }
    return false;
}

function appearsLeYear(y1: number, m1: number, d1: number, y2: number, m2: number, d2: number): boolean {
    // Returns true if date1 and date2 "appear" to be 1 year or less apart.
    // This compares the values of year, month, and day directly to each other.
    // Requires date1 <= date2; returns boolean. Used by basis 1.
    if (y1 === y2) {
        return true;
    }
    if ((y1 + 1) === y2 && (m1 > m2 || (m1 === m2 && d1 >= d2))) {
        return true;
    }
    return false;
}

/**
 * Excel-compatible Actual/Actual (basis 1) method.
 *
 * Cannot find it in ISO 20022.
 *
 * Found it on github (https://github.com/AnatolyBuga/yearfrac)
 * and verified it with Excel.
 *
 * Other actual/actual methods from ISO 20022 produce
 * different figures compared to Excel.
 */
export function actActExcel(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    const date1 = new Date(y1, m1 - 1, d1);
    const date2 = new Date(y2, m2 - 1, d2);
    
    if (appearsLeYear(y1, m1, d1, y2, m2, d2)) {
        let yearDays: number;
        if (y1 === y2 && isLeapYear(y1)) {
            yearDays = 366; // leap year
        } else if (feb29Between(date1, y1, date2, y2) || (m2 === 2 && d2 === 29)) {
            yearDays = 366; // leap year feb29
        } else {
            yearDays = 365; // leap year else
        }
        const df = (date2.getTime() - date1.getTime()) / (1000 * 86400);
        return fracDays ? (df + df2 - df1) : (df + df2 - df1) / yearDays;
    } else {
        const yearDays = (new Date(y2 + 1, 0, 1).getTime() - new Date(y1, 0, 1).getTime()) / (1000 * 86400);
        const avgYearDays = yearDays / (y2 - y1 + 1);
        const df = (date2.getTime() - date1.getTime()) / (1000 * 86400);
        return fracDays ? (df + df2 - df1) : (df + df2 - df1) / avgYearDays;
    }
}

/**
 * Source:
 *     https://github.com/hcnn/act_isda
 * Synonyms:
 *     - Actual/Actual ISDA
 *     - Act/Act ISDA
 *     - Actual/365 ISDA
 *     - Act/365 ISDA
 * 
 * ISO 20022:
 *     A008
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function actActIsda(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    if (y1 === y2) {
        const denom = isLeapYear(y2) ? 366 : 365;
        let diffDays = dateToJd(y2, m2, d2) - dateToJd(y1, m1, d1);
        diffDays += df2 - df1;
        return fracDays ? diffDays : (diffDays / denom);
    } else {
        const denomA = isLeapYear(y1) ? 366 : 365;
        const diffA = dateToJd(y1, 12, 31) - dateToJd(y1, m1, d1);

        const denomB = isLeapYear(y2) ? 366 : 365;
        const diffB = dateToJd(y2, m2, d2) - dateToJd(y2, 1, 1);

        if (fracDays) {
            let diff = diffA - df1 + diffB + df2;
            let year = y1 + 1;
            while (year < y2) {
                if (isLeapYear(year)) {
                    diff += 366;
                } else {
                    diff += 365;
                }
                year += 1;
            }
            return diff;
        } else {
            return (diffA - df1) / denomA + (diffB + df2) / denomB + y2 - y1 - 1;
        }
    }
}

/**
 * Source:
 *     https://github.com/hcnn/act_afb
 * Synonyms:
 *     - Actual/Actual AFB
 *     - Actual/Actual FBF
 * 
 * ISO 20022:
 *     A010
 *     https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
 */
export function actActAfb(
    y1: number, m1: number, d1: number,
    y2: number, m2: number, d2: number,
    df1: number = 0, df2: number = 0,
    fracDays: boolean = false
): number {
    if (y1 === y2) {
        const denom = (m1 < 3 && isLeapYear(y1)) ? 366 : 365;
        let diffDays = dateToJd(y2, m2, d2) - dateToJd(y1, m1, d1);
        diffDays += df2 - df1;
        return fracDays ? diffDays : (diffDays / denom);
    } else {
        const denomA = (m1 < 3 && isLeapYear(y1)) ? 366 : 365;
        let diffA = dateToJd(y1, 12, 31);
        diffA -= dateToJd(y1, m1, d1);

        const denomB = (m2 >= 3 && isLeapYear(y2)) ? 366 : 365;
        let diffB = dateToJd(y2, m2, d2);
        diffB -= dateToJd(y2, 1, 1);

        if (fracDays) {
            let diff = diffA - df1 + diffB + df2;
            let year = y1 + 1;
            while (year < y2) {
                if (isLeapYear(year)) {
                    diff += 366;
                } else {
                    diff += 365;
                }
                year += 1;
            }
            return diff;
        } else {
            return (diffA - df1) / denomA + (diffB + df2) / denomB + y2 - y1 - 1;
        }
    }
}
