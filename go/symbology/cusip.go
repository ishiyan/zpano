package symbology

import (
	"errors"
	"fmt"
)

// CUSIP (Committee on Uniform Security Identification Procedures) acronym refers
// to the 9-character alphanumeric security identifiers that they distribute for
// all North American securities for the purposes of facilitating clearing and
// settlement of trades.
//
// The CUSIP distribution system is owned by the American Bankers Association and is
// operated by Standard & Poor's. The CUSIP Services Bureau acts as the
// National Numbering Association (NNA) for North America, and the CUSIP serves
// as the National Securities Identification Number (NSIN) for products issued from
// both the United States and Canada.
//
// The CUSIP number consists of a base number of six characters known as the issuer
// number (the 4th, 5th and/or 6th position of which may be alpha or numeric) and a
// two character suffix (either numeric or alphabetic or both) known as the issue
// number. The 9th character is a check digit.
//
// The first six characters ("CUSIP-6") uniquely identify the issuer. A single alphabetical
// file of corporate, municipal, and government issuers has been developed, and an
// issuer number of six characters has been assigned to each in alphabetical sequence.
// One number will be assigned to an issuer, except few cases where additional issuer
// numbers will be assigned. The numbers from 990000 up are reserved, as are xxx990
// and up within each group of 1000 (i.e., 100990 to 1009ZZ). See the "Official List of
// Section 13(f) Securities" https://www.sec.gov/divisions/investment/13flists.htm.
//
// The issue number (the 7th and 8th digit) uniquely identifies each individual issue
// of an issuer, the format is dependent on the type of security. Each individual rate
// and maturity is considered a separate issue for numbering purposes. In general,
// two numbers are used for equity and two letters (or one numeric and one letter) are
// used for fixed income.
//
// The first issue number for an issuer's equity securities is 10. The unit position
// of the equity number is used to identify rights, warrants and so on and is assigned
// on an as-available basis. When there are insufficient tens positions available for
// all individual issues, the necessary additional numbers are found through the use
// of the first open two-position digit in reverse sequence starting with 88 and assigned
// in descending order. Issue numbers 00-09 are reserved for future use.
//
// Issue number 01 has been designated to identify options for an issuer. Issue number
// 89 will be reserved for overflow linkage and will not be assigned to a specific issue.
//
// The issue number assigned to an issuer's fixed income securities may consist of two
// alphabetic characters (AA etc.), one alphabetic character followed by one digit
// (A2 etc.), or one digit followed by one alphabetic character (2A etc.), assigned in
// that order. A separate issue number is assigned to each rate and/or maturity for each
// issue of bonds thus a serial bond having 40 different maturities is assigned 40
// separate issue numbers but general obligations of a municipality having the same
// issue date, rate and maturity are normally assigned the same number regardless of
// purpose. The alphabetic letter 'I' and numeric '1' as well as the alphabetic 'O' and
// numeric zero are not used in the assignment of issue numbers to fixed income securities.
// Issue Number 9Z will be reserved for overflow linkage and will not be assigned to a
// specific use.
//
// Issue Numbers 90 through 99 in the equity group, and 9A through 9Y in the fixed income
// group, are reserved for the user specifically for assignment to those issues of an
// eligible issuer where no CUSIP issue number has been assigned.
//
// The 9th digit is an automatically generated check digit using the "Modulus 10 Double Add
// Double" technique. To calculate the check digit every second digit is multiplied by two.
// Letters are converted to numbers by adding their ordinal position in the alphabet to 9,
// such that A = 10 and M = 22. The resulting string of digits (numbers greater than 10
// becoming two separate digits) are added up. The ten's-complement of the last number is
// the check digit. In other words, the sum of the digits, including the check-digit, is
// a multiple of 10. Some clearing bodies ignore or truncate the last digit.
//
// For securities and other financial instruments actively traded on an international
// basis, which are either underwritten (debt issues) or domiciled (equities) outside the
// United States and Canada, the security will be identified by a CINS (CUSIP International
// Numbering System) number.
//
// CINS numbers employ the same Issuer (6 characters)/Issue (2 character) & check digit
// concept espoused by the CUSIP Numbering System. The first position of a CINS code is
// always represented by an alpha character, signifying the Issuer's country code (domicile)
// or geographic region: A = Austria, B = Belgium, C = Canada, D = Germany, E = Spain,
// F = France, G = United Kingdom, H = Switzerland, J = Japan, K = Denmark, L = Luxembourg,
// M = Mid-East, N = Netherlands, P = South America, Q = Australia, R = Norway,
// S = South Africa, T = Italy, U = United States, V = Africa Other, W = Sweden,
// X = Europe Other, Y = Asia.
//
// See https://www.cusip.com/static/html/cusipaccess/CUSIPIntro_%207.26.2007.pdf.
type CUSIP string

const (
	cusipLength           = 9
	cusipCheckSumIndex    = cusipLength - 1
	cusipIssueSecondIndex = cusipLength - 2
	cusipIssueFirstIndex  = cusipLength - 3
)

var (
	errInvalidCUSIP           = errors.New("invalid CUSIP")
	errInvalidCUSIPLength9    = fmt.Errorf("length should be 9 symbols: %w", errInvalidCUSIP)
	errInvalidCUSIPLength8    = fmt.Errorf("length should be at least 8 symbols: %w", errInvalidCUSIP)
	errInvalidCUSIPLastSymbol = fmt.Errorf("last symbol should be a digit 0-9: %w", errInvalidCUSIP)
	errInvalidCUSIPCheckDigit = fmt.Errorf("invalid check digit (last symbol): %w", errInvalidCUSIP)
)

// Validate validates the CUSIP.
func (cusip CUSIP) Validate() error {
	if len(cusip) < cusipLength {
		return errInvalidCUSIPLength9
	}

	n := cusip[cusipCheckSumIndex]
	if n < '0' || n > '9' {
		return errInvalidCUSIPLastSymbol
	}

	n -= '0'

	d, err := cusip.CalculateCheckDigit()
	if err != nil {
		return err
	}

	if n != d {
		// A fix for incorrect CUSIPs in SEC 13F Security List.
		// See https://quant.stackexchange.com/questions/16392/sec-13f-security-list-has-incorrect-cusip-numbers.
		// When making CUSIPs for all of the options on the list, it does this by taking the first 6 digits of
		// the underlying equity (which makes sense, as this represents the issuer, which should be the same for
		// the stock and the option), and then for the 7th and 8th digit, it uses 90 for calls and 95 for puts.
		// It then uses the 9th digit from the underlying stock as the 9th digit for the option, which is why
		// the checksum doesn't work.
		// In short, they override 7th and 8th digit for options and DON'T recalculate checksum digit.
		if cusip[cusipIssueFirstIndex] == '9' {
			switch cusip[cusipIssueSecondIndex] {
			case '0', '5':
				return nil
			}
		}

		return errInvalidCUSIPCheckDigit
	}

	return nil
}

// CalculateCheckDigit calculates a check digit of the CUSIP according to the Luhn algorithm.
func (cusip CUSIP) CalculateCheckDigit() (byte, error) {
	if len(cusip) < cusipCheckSumIndex {
		return 0, errInvalidCUSIPLength8
	}

	sum := 0

	for i := range cusipCheckSumIndex {
		n, err := toOrdinalNumberCUSIP(cusip[i], i)
		if err != nil {
			return 0, err
		}

		if i%2 == 1 {
			n *= 2
		}

		sum += n/ten + n%ten
	}

	sum = (ten - sum%ten) % ten

	return byte(sum), nil
}

func toOrdinalNumberCUSIP(b byte, i int) (int, error) {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), nil
	case b >= 'A' && b <= 'Z':
		return int(b - 'A' + ten), nil
	case b == '*':
		return 36, nil //nolint:gomnd
	case b == '@':
		return 37, nil //nolint:gomnd
	case b == '#':
		return 38, nil //nolint:gomnd
	default:
		return 0, fmt.Errorf(
			"symbol at position %v should be either a digit 0-9, a letter A-Z or special symbols @*#: %w", i, errInvalidCUSIP)
	}
}
