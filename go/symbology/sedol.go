package symbology

import (
	"errors"
	"fmt"
)

// SEDOL (Stock Exchange Daily Official List) is list of security identifiers used in
// the United Kingdom and Ireland for clearing purposes. The numbers are assigned by
// the London Stock Exchange, on request by the security issuer. SEDOL codes serve as
// the NSIN for all securities issued in the United Kingdom and are therefore part of
// the security's ISIN as well.
//
// SEDOL codes are seven characters in length, consisting of two parts: a six-place
// alphanumeric code and a trailing check digit. There are three types of SEDOL codes.
//
// ➊ Old style SEDOL codes issued prior to 2004 were composed only of digits.
// They cannot begin with the leading digit 9.
//
// ➊ New style SEDOL codes issued after 2004, were changed to be alpha-numeric and
// are issued sequentially, beginning with B000009. They begin with a leading letter
// followed by five alphanumeric characters and the trailing check digit.
// Vowels 'AEUIO' are never used.
//
// ➊ User defined SEDOL codes begin with a leading digit 9 followed by five
// alphanumeric characters and the trailing check digit. The alphanumeric characters
// may be vowels. There will be no codes issued with 9 as the lead character.
// This allows the 9-series to be reserved for end user allocation.
//
// The check digit for SEDOL codes is chosen to make the total weighted sum of all
// seven characters a multiple of 10. The check digit is computed using a weighted
// sum of the first six characters. Letters are converted to numbers by adding their
// ordinal position in the alphabet to 9, such that B = 11 and Z = 35. The resulting
// string of 7 numbers is then multiplied by the weighting factors [1, 3, 1, 7, 3, 9, 1].
// The check digit is chosen to make the total sum, including the check digit, a multiple
// of 10, which can be calculated from the weighted sum of the first six characters as
//
//	(10 - (weighted sum modulo 10)) modulo 10.
//
// For British and Irish securities, SEDOL codes are converted to ISINs by padding
// the front with two zeros, then adding the country code on the front and the ISIN
// check digit at the end.
//
// See:
// http://www.londonstockexchange.com/products-and-services/reference-data/sedol-master-file/sedol-master-file.htm,
// http://www.londonstockexchange.com/products-and-services/reference-data/sedol-master-file/documentation/sedol-technical-specification.pdf,
//
//nolint:lll
type SEDOL string

const (
	sedolLength               = 7
	sedolCheckSumIndex        = sedolLength - 1
	sedolUserDefinedCharacter = '9'
	sedolUserDefined          = 1
	sedolOldStyle             = 2
	sedolNewStyle             = 3
)

var (
	errInvalidSEDOL           = errors.New("invalid SEDOL")
	errInvalidSEDOLLength7    = fmt.Errorf("length should be 7 symbols: %w", errInvalidSEDOL)
	errInvalidSEDOLLength6    = fmt.Errorf("length should be at least 6 symbols: %w", errInvalidSEDOL)
	errInvalidSEDOLLastSymbol = fmt.Errorf("last symbol should be a digit 0-9: %w", errInvalidSEDOL)
	errInvalidSEDOLCheckDigit = fmt.Errorf("invalid check digit (last symbol): %w", errInvalidSEDOL)
)

// Validate validates the SEDOL.
func (sedol SEDOL) Validate() error {
	if len(sedol) < sedolLength {
		return errInvalidSEDOLLength7
	}

	n := sedol[sedolCheckSumIndex]
	if n < '0' || n > '9' {
		return errInvalidSEDOLLastSymbol
	}

	n -= '0'

	d, err := sedol.CalculateCheckDigit()
	if err != nil {
		return err
	}

	if n != d {
		return errInvalidSEDOLCheckDigit
	}

	return nil
}

// CalculateCheckDigit calculates a check digit of the SEDOL.
//
//nolint:cyclop
func (sedol SEDOL) CalculateCheckDigit() (byte, error) {
	if len(sedol) < sedolCheckSumIndex {
		return 0, errInvalidSEDOLLength6
	}

	style := sedolNewStyle
	sum := 0

	for i := range sedolCheckSumIndex {
		var n int

		b := sedol[i]

		switch {
		case b >= '0' && b <= '9':
			n = int(b - '0')

			if i == 0 {
				if b == sedolUserDefinedCharacter {
					style = sedolUserDefined
				} else {
					style = sedolOldStyle
				}
			}
		case b >= 'A' && b <= 'Z':
			if style == sedolOldStyle {
				return 0, fmt.Errorf(
					"symbol at position %v should be a digit 0-9 in old style SEDOL: %w ", i, errInvalidSEDOL)
			}

			if style == sedolNewStyle {
				switch b {
				case 'A', 'E', 'U', 'I', 'O':
					return 0, fmt.Errorf(
						"symbol at position %v should not be a vowel AEUIO in user defined SEDOL: %w ", i, errInvalidSEDOL)
				}
			}

			n = int(b - 'A' + ten)
		default:
			return 0, fmt.Errorf(
				"symbol at position %v should be either a digit 0-9 or a letter A-Z: %w ", i, errInvalidSEDOL)
		}

		sum += applyWeight(i, n)
	}

	sum = (ten - sum%ten) % ten

	return byte(sum), nil
}

// sedolWeights are the positional weights used in the SEDOL check digit calculation.
var sedolWeights = [sedolCheckSumIndex]int{1, 3, 1, 7, 3, 9}

func applyWeight(i, n int) int {
	return n * sedolWeights[i]
}
