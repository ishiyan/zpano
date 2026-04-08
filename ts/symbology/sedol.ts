/**
 * SEDOL (Stock Exchange Daily Official List) validation.
 *
 * SEDOL codes are seven characters in length, consisting of two parts:
 * a six-place alphanumeric code and a trailing check digit. There are
 * three types of SEDOL codes: old style, new style, and user defined.
 *
 * See http://www.londonstockexchange.com/products-and-services/reference-data/sedol-master-file/sedol-master-file.htm.
 */

const TEN = 10;

const SEDOL_LENGTH = 7;
const SEDOL_CHECK_SUM_INDEX = SEDOL_LENGTH - 1;
const SEDOL_USER_DEFINED_CHARACTER = '9';
const SEDOL_USER_DEFINED = 1;
const SEDOL_OLD_STYLE = 2;
const SEDOL_NEW_STYLE = 3;

const SEDOL_WEIGHTS = [1, 3, 1, 7, 3, 9];

/** SEDOL validator and check digit calculator. */
export class SEDOL {
    private readonly value: string;

    constructor(value: string) {
        this.value = value;
    }

    /** Validate the SEDOL. */
    validate(): void {
        if (this.value.length < SEDOL_LENGTH) {
            throw new Error('length should be 7 symbols: invalid SEDOL');
        }

        const n = this.value[SEDOL_CHECK_SUM_INDEX];
        if (n < '0' || n > '9') {
            throw new Error('last symbol should be a digit 0-9: invalid SEDOL');
        }

        const nVal = n.charCodeAt(0) - '0'.charCodeAt(0);
        const d = this.calculateCheckDigit();

        if (nVal !== d) {
            throw new Error('invalid check digit (last symbol): invalid SEDOL');
        }
    }

    /** Calculate a check digit of the SEDOL. */
    calculateCheckDigit(): number {
        if (this.value.length < SEDOL_CHECK_SUM_INDEX) {
            throw new Error('length should be at least 6 symbols: invalid SEDOL');
        }

        let style = SEDOL_NEW_STYLE;
        let total = 0;

        for (let i = 0; i < SEDOL_CHECK_SUM_INDEX; i++) {
            const b = this.value[i];
            let n: number;

            if (b >= '0' && b <= '9') {
                n = b.charCodeAt(0) - '0'.charCodeAt(0);

                if (i === 0) {
                    if (b === SEDOL_USER_DEFINED_CHARACTER) {
                        style = SEDOL_USER_DEFINED;
                    } else {
                        style = SEDOL_OLD_STYLE;
                    }
                }
            } else if (b >= 'A' && b <= 'Z') {
                if (style === SEDOL_OLD_STYLE) {
                    throw new Error(
                        `symbol at position ${i} should be a digit 0-9 ` +
                        `in old style SEDOL: invalid SEDOL`
                    );
                }

                if (style === SEDOL_NEW_STYLE) {
                    if (b === 'A' || b === 'E' || b === 'U' || b === 'I' || b === 'O') {
                        throw new Error(
                            `symbol at position ${i} should not be a vowel ` +
                            `AEUIO in user defined SEDOL: invalid SEDOL`
                        );
                    }
                }

                n = b.charCodeAt(0) - 'A'.charCodeAt(0) + TEN;
            } else {
                throw new Error(
                    `symbol at position ${i} should be either a digit 0-9 ` +
                    `or a letter A-Z: invalid SEDOL`
                );
            }

            total += n * SEDOL_WEIGHTS[i];
        }

        total = (TEN - total % TEN) % TEN;

        return total;
    }
}
