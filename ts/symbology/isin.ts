/**
 * ISIN (ISO 6166 International Securities Identifying Number) validation.
 *
 * ISINs consist of three parts: a two letter country code, a nine character
 * alpha-numeric national security identifier, and a single check digit.
 *
 * See https://en.wikipedia.org/wiki/International_Securities_Identification_Number.
 */

const TEN = 10;

const ISIN_LENGTH = 12;
const ISIN_CHECK_SUM_INDEX = ISIN_LENGTH - 1;
const ISIN_COUNTRY_LENGTH = 2;

/** Valid country codes grouped by first letter. */
const VALID_COUNTRIES: Record<string, Set<string>> = {
    'A': new Set(['D', 'E', 'F', 'G', 'I', 'L', 'M', 'N', 'O', 'Q', 'R', 'S', 'T', 'U', 'W', 'Z']),
    'B': new Set(['A', 'B', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'M', 'N', 'O', 'R', 'S', 'T', 'V', 'W', 'Y', 'Z']),
    'C': new Set(['A', 'C', 'D', 'F', 'G', 'H', 'I', 'K', 'L', 'M', 'N', 'O', 'R', 'U', 'V', 'X', 'Y', 'Z']),
    'D': new Set(['E', 'J', 'K', 'M', 'O', 'Z']),
    'E': new Set(['C', 'E', 'G', 'R', 'S', 'T', 'U']),
    'F': new Set(['I', 'J', 'K', 'M', 'O', 'R']),
    'G': new Set(['A', 'B', 'D', 'E', 'G', 'H', 'I', 'L', 'M', 'N', 'Q', 'R', 'S', 'T', 'U', 'W', 'Y']),
    'H': new Set(['K', 'M', 'N', 'R', 'T', 'U']),
    'I': new Set(['D', 'E', 'L', 'M', 'N', 'O', 'Q', 'R', 'S', 'T']),
    'J': new Set(['E', 'M', 'O', 'P']),
    'K': new Set(['E', 'G', 'H', 'I', 'M', 'N', 'P', 'R', 'W', 'Y', 'Z']),
    'L': new Set(['A', 'B', 'C', 'I', 'K', 'R', 'S', 'T', 'U', 'V', 'Y']),
    'M': new Set(['A', 'C', 'D', 'E', 'G', 'H', 'K', 'L', 'M', 'N', 'O', 'P', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z']),
    'N': new Set(['A', 'C', 'E', 'F', 'G', 'I', 'L', 'O', 'P', 'R', 'S', 'U', 'Z']),
    'O': new Set(['M']),
    'P': new Set(['A', 'E', 'F', 'G', 'H', 'K', 'L', 'M', 'N', 'S', 'T', 'W', 'Y']),
    'Q': new Set(['A', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z']),
    'R': new Set(['O', 'U', 'W']),
    'S': new Set(['A', 'B', 'C', 'D', 'E', 'G', 'H', 'I', 'K', 'L', 'M', 'N', 'O', 'R', 'S', 'T', 'V', 'Y', 'Z']),
    'T': new Set(['C', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'O', 'R', 'T', 'V', 'W', 'Z']),
    'U': new Set(['A', 'G', 'M', 'S', 'Y', 'Z']),
    'V': new Set(['A', 'C', 'E', 'G', 'I', 'N', 'U']),
    'W': new Set(['F', 'S']),
    'X': new Set(['A', 'B', 'C', 'D', 'F', 'K', 'L', 'S']),
    'Y': new Set(['E', 'T']),
    'Z': new Set(['A', 'M', 'W']),
};

function toOrdinalNumberIsin(ch: string, i: number): number {
    if (i < ISIN_COUNTRY_LENGTH) {
        if (ch >= 'A' && ch <= 'Z') {
            return ch.charCodeAt(0) - 'A'.charCodeAt(0) + TEN;
        }
        throw new Error(`symbol at position ${i} should be a letter A-Z: invalid ISIN`);
    }

    if (ch >= '0' && ch <= '9') {
        return ch.charCodeAt(0) - '0'.charCodeAt(0);
    }
    if (ch >= 'A' && ch <= 'Z') {
        return ch.charCodeAt(0) - 'A'.charCodeAt(0) + TEN;
    }

    throw new Error(
        `symbol at position ${i} should be either a digit 0-9 or a letter A-Z: invalid ISIN`
    );
}

/** ISIN validator and check digit calculator. */
export class ISIN {
    private readonly value: string;

    constructor(value: string) {
        this.value = value;
    }

    /** Validate the country code and the check digit of the ISIN. */
    validate(): void {
        if (!this.validateCountry()) {
            throw new Error('unknown country code: invalid ISIN');
        }

        this.validateCheckDigit();
    }

    /** Validate the check digit of the ISIN. */
    validateCheckDigit(): void {
        if (this.value.length !== ISIN_LENGTH) {
            throw new Error('length should be 12 symbols: invalid ISIN');
        }

        const n = this.value[ISIN_CHECK_SUM_INDEX];
        if (n < '0' || n > '9') {
            throw new Error('last symbol should be a digit 0-9: invalid ISIN');
        }

        const nVal = n.charCodeAt(0) - '0'.charCodeAt(0);
        const d = this.calculateCheckDigit();

        if (nVal !== d) {
            throw new Error('invalid check digit (last symbol): invalid ISIN');
        }
    }

    /** Calculate a check digit of the ISIN according to the Luhn algorithm. */
    calculateCheckDigit(): number {
        if (this.value.length < ISIN_CHECK_SUM_INDEX) {
            throw new Error('length should be at least 11 symbols: invalid ISIN');
        }

        let total = 0;
        let multiply = true;

        for (let i = ISIN_CHECK_SUM_INDEX - 1; i >= 0; i--) {
            let n = toOrdinalNumberIsin(this.value[i], i);

            if (n < TEN) {
                if (multiply) {
                    n *= 2;
                    total += n % TEN + Math.floor(n / TEN);
                } else {
                    total += n;
                }

                multiply = !multiply;
            } else {
                if (multiply) {
                    total += Math.floor(n / TEN);
                    n %= TEN;
                } else {
                    total += n % TEN;
                    n = Math.floor(n / TEN);
                }

                n *= 2;
                total += n % TEN + Math.floor(n / TEN);
            }
        }

        total = (TEN - total % TEN) % TEN;

        return total;
    }

    /** Validate if the first two letters represent a valid country code. */
    validateCountry(): boolean {
        if (this.value.length < ISIN_COUNTRY_LENGTH) {
            return false;
        }

        const first = this.value[0];
        const second = this.value[1];

        const seconds = VALID_COUNTRIES[first];
        if (seconds === undefined) {
            return false;
        }

        return seconds.has(second);
    }
}
