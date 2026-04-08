/**
 * CUSIP (Committee on Uniform Security Identification Procedures) validation.
 *
 * The CUSIP number consists of a base number of six characters known as the
 * issuer number and a two character suffix known as the issue number. The 9th
 * character is a check digit.
 *
 * See https://www.cusip.com/static/html/cusipaccess/CUSIPIntro_%207.26.2007.pdf.
 */

const TEN = 10;

const CUSIP_LENGTH = 9;
const CUSIP_CHECK_SUM_INDEX = CUSIP_LENGTH - 1;
const CUSIP_ISSUE_SECOND_INDEX = CUSIP_LENGTH - 2;
const CUSIP_ISSUE_FIRST_INDEX = CUSIP_LENGTH - 3;

function toOrdinalNumberCusip(ch: string, i: number): number {
    if (ch >= '0' && ch <= '9') {
        return ch.charCodeAt(0) - '0'.charCodeAt(0);
    }
    if (ch >= 'A' && ch <= 'Z') {
        return ch.charCodeAt(0) - 'A'.charCodeAt(0) + TEN;
    }
    if (ch === '*') {
        return 36;
    }
    if (ch === '@') {
        return 37;
    }
    if (ch === '#') {
        return 38;
    }

    throw new Error(
        `symbol at position ${i} should be either a digit 0-9, ` +
        `a letter A-Z or special symbols @*#: invalid CUSIP`
    );
}

/** CUSIP validator and check digit calculator. */
export class CUSIP {
    private readonly value: string;

    constructor(value: string) {
        this.value = value;
    }

    /** Validate the CUSIP. */
    validate(): void {
        if (this.value.length < CUSIP_LENGTH) {
            throw new Error('length should be 9 symbols: invalid CUSIP');
        }

        const n = this.value[CUSIP_CHECK_SUM_INDEX];
        if (n < '0' || n > '9') {
            throw new Error('last symbol should be a digit 0-9: invalid CUSIP');
        }

        const nVal = n.charCodeAt(0) - '0'.charCodeAt(0);
        const d = this.calculateCheckDigit();

        if (nVal !== d) {
            // A fix for incorrect CUSIPs in SEC 13F Security List.
            // See https://quant.stackexchange.com/questions/16392/sec-13f-security-list-has-incorrect-cusip-numbers.
            if (this.value[CUSIP_ISSUE_FIRST_INDEX] === '9') {
                if (this.value[CUSIP_ISSUE_SECOND_INDEX] === '0' ||
                    this.value[CUSIP_ISSUE_SECOND_INDEX] === '5') {
                    return;
                }
            }

            throw new Error('invalid check digit (last symbol): invalid CUSIP');
        }
    }

    /** Calculate a check digit of the CUSIP according to the Luhn algorithm. */
    calculateCheckDigit(): number {
        if (this.value.length < CUSIP_CHECK_SUM_INDEX) {
            throw new Error('length should be at least 8 symbols: invalid CUSIP');
        }

        let total = 0;

        for (let i = 0; i < CUSIP_CHECK_SUM_INDEX; i++) {
            let n = toOrdinalNumberCusip(this.value[i], i);

            if (i % 2 === 1) {
                n *= 2;
            }

            total += Math.floor(n / TEN) + n % TEN;
        }

        total = (TEN - total % TEN) % TEN;

        return total;
    }
}
