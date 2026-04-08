package symbology

import (
	"errors"
	"fmt"
)

// ISIN is an ISO6166 International Securities Identifying Number.
// See https://en.wikipedia.org/wiki/International_Securities_Identification_Number.
//
// ISINs consist of three parts: a two letter country code, a nine character alpha-numeric national
// security identifier, and a single check digit. The country code is the ISO 3166-1 alpha-2 code for the
// country of issue, which is not necessarily the country in which the issuing company is domiciled.
// International securities cleared through Clearstream or Euroclear, which are Europe-wide, use "XS" as
// the country code.
//
// The nine-digit security identifier is the National Securities Identifying Number, or NSIN, assigned by
// governing bodies in each country, known as the national numbering agency (NNA). In North America the
// NNA is the CUSIP organization, meaning that CUSIPs can easily be converted into ISINs by adding the US
// or CA country code to the beginning of the existing CUSIP code and adding an additional check digit at
// the end. In the United Kingdom and Ireland the NNA is the London Stock Exchange and the NSIN is the
// SEDOL, converted in a similar fashion after padding the SEDOL number out with leading zeros. Most other
// countries use similar conversions, but if no country NNA exists then regional NNAs are used instead.
//
// ISIN check digits are based on the Luhn algorithm, also known as the "modulus 10" or "mod 10" algorithm,
// the same "Modulus 10 Double Add Double" technique used in CUSIPs.
//
// Luhn algorithm works as follows. Consider an example of “US0378331005“ (Apple).
// Step 1: Convert any letters to numbers by adding their ordinal position in the alphabet to 9, such that A = 10.
// For example, 'U' becomes 'U'-'A'+10 = 85-65+10 = 30, 'S' becomes 'S'-'A'+10 = 83-65+10 = 28.
// Step 2: Starting from the rightmost digit (not counting the check digit), double the value of every second digit
// (note numbers greater 9 are treated as two separate digits).
// Step 3: If doubling of a number results in a two digit number, then add the digits of the product to get a single
// digit number.
// Step 4: Take the sum of all the digits.
// Step 5: Subtract this sum from the smallest number ending with zero that is greater than or equal to the sum.
//
//	step 0: U   S   0   3   7   8   3   3   1   0   0
//	step 1: 30  28  0   3   7   8   3   3   1   0   0
//	step 2: 60  48  0   3  14   8   6   3   2   0   0
//	step 3: 60  48  0   3   5   8   6   3   2   0   0
//	step 4: 6 + 0 + 4 + 8 + 0 + 3 + 5 + 8 + 6 + 3 + 2 + 0 + 0 = 45
//	step 5: 50 - 45 = 5
//
// This gives the check digit which is also known as the ten's complement of the sum modulo 10.
type ISIN string

const (
	isinLength        = 12
	isinCheckSumIndex = isinLength - 1
	isinCountryLength = 2
	ten               = 10
)

var (
	errInvalidISIN            = errors.New("invalid ISIN")
	errInvalidISINCountryCode = fmt.Errorf("unknown country code: %w", errInvalidISIN)
	errInvalidISINLength12    = fmt.Errorf("length should be 12 symbols: %w", errInvalidISIN)
	errInvalidISINLength11    = fmt.Errorf("length should be at least 11 symbols: %w", errInvalidISIN)
	errInvalidISINLastSymbol  = fmt.Errorf("last symbol should be a digit 0-9: %w", errInvalidISIN)
	errInvalidISINCheckDigit  = fmt.Errorf("invalid check digit (last symbol): %w", errInvalidISIN)
)

// Validate validates the country code and the check digit of the ISIN.
func (isin ISIN) Validate() error {
	if !isin.ValidateCountry() {
		return errInvalidISINCountryCode
	}

	return isin.ValidateCheckDigit()
}

// ValidateCheckDigit validates the check digit of the ISIN.
func (isin ISIN) ValidateCheckDigit() error {
	if len(isin) != isinLength {
		return errInvalidISINLength12
	}

	n := isin[isinCheckSumIndex]
	if n < '0' || n > '9' {
		return errInvalidISINLastSymbol
	}

	n -= '0'

	d, err := isin.CalculateCheckDigit()
	if err != nil {
		return err
	}

	if n != d {
		return errInvalidISINCheckDigit
	}

	return nil
}

// CalculateCheckDigit calculates a check digit of the ISIN according to the Luhn algorithm.
func (isin ISIN) CalculateCheckDigit() (byte, error) {
	if len(isin) < isinCheckSumIndex {
		return 0, errInvalidISINLength11
	}

	sum := 0
	multiply := true

	for i := isinCheckSumIndex - 1; i >= 0; i-- {
		n, err := toOrdinalNumberISIN(isin[i], i)
		if err != nil {
			return 0, err
		}

		if n < ten { //nolint:nestif
			if multiply {
				n *= 2
				sum += n%ten + n/ten
			} else {
				sum += n
			}

			multiply = !multiply
		} else {
			if multiply {
				sum += n / ten
				n %= ten
			} else {
				sum += n % ten
				n /= ten
			}

			n *= 2
			sum += n%ten + n/ten
		}
	}

	sum = (ten - sum%ten) % ten

	return byte(sum), nil
}

// ValidateCountry validates if two first letters of the ISIN represent a valid country code.
//
//gocyclo:ignore
//nolint:funlen,gocognit,cyclop,maintidx
func (isin ISIN) ValidateCountry() bool {
	if len(isin) < isinCountryLength {
		return false
	}

	first := isin[0]
	second := isin[1]

	switch first {
	case 'A':
		switch second {
		case 'D', // AD, Andorra
			'E', // AE, United Arab Emirates
			'F', // AF, Afghanistan
			'G', // AG, Antigua and Barbuda
			'I', // AI, Anguilla
			'L', // AL, Albania
			'M', // AM, Armenia
			'N', // AN, Netherlands Antilles
			'O', // AO, Angola
			'Q', // AQ, Antarctica
			'R', // AR, Argentina
			'S', // AS, American Samoa
			'T', // AT, Austria
			'U', // AU, Australia
			'W', // AW, Aruba
			'Z': // AZ, Azerbaijan
			return true
		}
	case 'B':
		switch second {
		case 'A', // BA, Bosnia and Herzegovina
			'B', // BB, Barbados
			'D', // BD, Bangladesh
			'E', // BE, Belgium
			'F', // BF, Burkina Faso
			'G', // BG, Bulgaria
			'H', // BH, Bahrain
			'I', // BI, Burundi
			'J', // BJ, Benin
			'M', // BM, Bermuda
			'N', // BN, Brunei
			'O', // BO, Bolivia
			'R', // BR, Brazil
			'S', // BS, Bahamas
			'T', // BT, Bhutan
			'V', // BV, Bouvet Island
			'W', // BW, Botswana
			'Y', // BY, Belarus
			'Z': // BZ, Belize
			return true
		}
	case 'C':
		switch second {
		case 'A', // CA, Canada
			'C', // CC, Cocos Islands (or Keeling Islands)
			'D', // CD, Congo Democratic Republic of
			'F', // CF, Central African Republic
			'G', // CG, Congo (Republic of)
			'H', // CH, Switzerland
			'I', // CI, Côte d'Ivoire
			'K', // CK, Cook Islands
			'L', // CL, Chile
			'M', // CM, Cameroon
			'N', // CN, China
			'O', // CO, Colombia
			'R', // CR, Costa Rica
			'U', // CU, Cuba
			'V', // CV, Cape Verde
			'X', // CX, Christmas Island
			'Y', // CY, Cyprus
			'Z': // CZ, Czech Republic
			return true
		}
	case 'D':
		switch second {
		case 'E', // DE, Federal Republic of Germany
			'J', // DJ, Djibouti
			'K', // DK, Denmark
			'M', // DM, Dominica
			'O', // DO, Dominican Republic
			'Z': // DZ, Algeria
			return true
		}
	case 'E':
		switch second {
		case 'C', // EC, Ecuador
			'E', // EE, Estonia
			'G', // EG, Egypt
			'R', // ER, Eritrea
			'S', // ES, Spain (excluding XB, XC)
			'T', // ET, Ethiopia
			'U': // EU, European Union, also used as pseudo country code
			return true
		}
	case 'F':
		switch second {
		case 'I', // FI, Finland
			'J', // FJ, Fiji
			'K', // FK, Falkland Islands
			'M', // FM, Federated States of Micronesia
			'O', // FO, Faroe Islands
			'R': // FR, France
			return true
		}
	case 'G':
		switch second {
		case 'A', // GA, Gabon
			'B', // GB, United Kingdom
			'D', // GD, Grenada
			'E', // GE, Georgia
			'G', // GG, Guernsey
			'H', // GH, Ghana
			'I', // GI, Gibraltar
			'L', // GL, Greenland
			'M', // GM, Gambia
			'N', // GN, Guinea
			'Q', // GQ, Equatorial Guinea
			'R', // GR, Greece
			'S', // GS, South Georgia and South Sandwich Islands
			'T', // GT, Guatemala
			'U', // GU, Guam
			'W', // GW, Guinea Bissau
			'Y': // GY, Guyana
			return true
		}
	case 'H':
		switch second {
		case 'K', // HK, Hong Kong
			'M', // HM, Heard Island and McDonald Islands
			'N', // HN, Honduras
			'R', // HR, Croatia
			'T', // HT, Haiti
			'U': // HU, Hungary
			return true
		}
	case 'I':
		switch second {
		case 'D', // ID, Indonesia
			'E', // IE, Ireland
			'L', // IL, Israel
			'M', // IM, Isle of Man
			'N', // IN, India
			'O', // IO, British Indian Ocean Territory
			'Q', // IQ, Iraq
			'R', // IR, Iran
			'S', // IS, Iceland
			'T': // IT, Italy
			return true
		}
	case 'J':
		switch second {
		case 'E', // JE, Jersey
			'M', // JM, Jamaica
			'O', // JO, Jordan
			'P': // JP, Japan
			return true
		}
	case 'K':
		switch second {
		case 'E', // KE, Kenya
			'G', // KG, Kyrgyzstan
			'H', // KH, Cambodia (Kampuchea)
			'I', // KI, Kiribati
			'M', // KM, Comoros (excluding Mayotte)
			'N', // KN, St. Christopher and Nevis
			'P', // KP, North Korea
			'R', // KR, South Korea
			'W', // KW, Kuwait
			'Y', // KY, Cayman Islands
			'Z': // KZ, Kazakhstan
			return true
		}
	case 'L':
		switch second {
		case 'A', // LA, Laos
			'B', // LB, Lebanon
			'C', // LC, St. Lucia
			'I', // LI, Liechtenstein
			'K', // LK, Sri Lanka
			'R', // LR, Liberia
			'S', // LS, Lesotho
			'T', // LT, Lithuania
			'U', // LU, Luxembourg
			'V', // LV, Latvia
			'Y': // LY, Libya
			return true
		}
	case 'M':
		switch second {
		case 'A', // MA, Morocco
			'C', // MC, Monaco
			'D', // MD, Moldova
			'E', // ME, Montenegro
			'G', // MG, Madagascar
			'H', // MH, Republic of the Marshall Islands
			'K', // MK, Former Republic of Macedonia (FYROM)
			'L', // ML, Mali
			'M', // MM, Myanmar
			'N', // MN, Mongolia
			'O', // MO, Macao
			'P', // MP, Northern Mariana Islands
			'R', // MR, Mauritania
			'S', // MS, Montserrat
			'T', // MT, Malta
			'U', // MU, Mauritius
			'V', // MV, Maldives
			'W', // MW, Malawi
			'X', // MX, Mexico
			'Y', // MY, Malaysia
			'Z': // MZ, Mozambique
			return true
		}
	case 'N':
		switch second {
		case 'A', // NA, Namibia
			'C', // NC, New Caledonia and dependencies
			'E', // NE, Niger
			'F', // NF, Norfolk Island
			'G', // NG, Nigeria
			'I', // NI, Nicaragua
			'L', // NL, Netherlands
			'O', // NO, Norway
			'P', // NP, Nepal
			'R', // NR, Nauru
			'S', // NS, Unknown, NSCEX0000018
			'U', // NU, Niue Island
			'Z': // NZ, New Zealand
			return true
		}
	case 'O':
		if second == 'M' { // OM, Oman
			return true
		}
	case 'P':
		switch second {
		case 'A', // PA, Panama
			'E', // PE, Peru
			'F', // PF, French Polynesia
			'G', // PG, Papua New Guinea
			'H', // PH, Philippines
			'K', // PK, Pakistan
			'L', // PL, Poland
			'M', // PM, St Pierre and Miquelon
			'N', // PN, Pitcairn
			'S', // PS, Occupied palestinian Territory
			'T', // PT, Portugal
			'W', // PW, Palau
			'Y': // PY, Paraguay
			return true
		}
	case 'Q':
		switch second {
		case 'A', // QA, Qatar
			'Q', // QQ, Stores and provisions
			'R', // QR, Stores and provisions within the framework of intra-Community trade
			'S', // QS, Stores and provisions within the framework of trade with Third Countries
			'T', // QT, used as pseudo country code
			'U', // QU, Countries and territories not specified
			'V', // QV, Countries and territories not specified in the framework of intra-Community trade
			'W', // QW, Countries and territories not specified within the framework of trade with the third countries
			'X', // QX, Countries and territories not specified for commercial or military reasons
			'Y', // QY, QX in the framework of intra-Community trade
			'Z': // QZ, QX in the framework of trade with third countries
			return true
		}
	case 'R':
		switch second {
		case 'O', // RO, Romania
			'U', // RU, Russian Federation
			'W': // RW, Rwanda
			return true
		}
	case 'S':
		switch second {
		case 'A', // SA, Saudi Arabia
			'B', // SB, Solomon Islands
			'C', // SC, Seychelles and dependencies
			'D', // SD, Sudan
			'E', // SE, Sweden
			'G', // SG, Singapore
			'H', // SH, St Helena and dependencies
			'I', // SI, Slovenia
			'K', // SK, Slovakia
			'L', // SL, Sierra Leone
			'M', // SM, San Marino
			'N', // SN, Senegal
			'O', // SO, Somalia
			'R', // SR, Surinam
			'S', // SS, South Sudan
			'T', // ST, São Tomé and Principe
			'V', // SV, El Salvador
			'Y', // SY, Syria
			'Z': // SZ, Swaziland
			return true
		}
	case 'T':
		switch second {
		case 'C', // TC, Turks and Caicos Islands
			'D', // TD, Chad
			'F', // TF, French Southern Territories
			'G', // TG, Togo
			'H', // TH, Thailand
			'J', // TJ, Tajikistan
			'K', // TK, Tokelau
			'L', // TL, East Timor
			'M', // TM, Turkmenistan
			'N', // TN, Tunisia
			'O', // TO, Tonga
			'R', // TR, Turkey
			'T', // TT, Trinidad and Tobago
			'V', // TV, Tuvalu
			'W', // TW, Taiwan
			'Z': // TZ, Tanzania
			return true
		}
	case 'U':
		switch second {
		case 'A', // UA, Ukraine
			'G', // UG, Uganda
			'M', // UM, United States Minor outlying islands
			'S', // US, United States of America
			'Y', // UY, Uruguay
			'Z': // UZ, Uzbekistan
			return true
		}
	case 'V':
		switch second {
		case 'A', // VA, Vatican City State
			'C', // VC, St Vincent
			'E', // VE, Venezuela
			'G', // VG, British Virgin Islands
			'I', // VI, Virgin Islands of U.S
			'N', // VN, Vietnam
			'U': // VU, Vanuatu
			return true
		}
	case 'W':
		switch second {
		case 'F', // WF, Wallis and Futuna Islands
			'S': // WS, Samoa
			return true
		}
	case 'X':
		switch second {
		case 'A', // XA, used as pseudo country code
			'B', // XB, used as pseudo country code
			'C', // XC, Ceuta, also used as pseudo country code
			'D', // XD, used as pseudo country code
			'F', // XF, used as pseudo country code
			'K', // XK, Kosovo
			'L', // XL, Melilla
			'S': // XS, Serbia, also used as pseudo country code
			return true
		}
	case 'Y':
		switch second {
		case 'E', // YE, Yemen
			'T': // YT, Mayotte
			return true
		}
	case 'Z':
		switch second {
		case 'A', // ZA, South Africa
			'M', // ZM, Zambia
			'W': // ZW, Zimbabwe
			return true
		}
	}

	return false
}

func toOrdinalNumberISIN(b byte, i int) (int, error) {
	if i < isinCountryLength {
		switch {
		case b >= 'A' && b <= 'Z':
			return int(b - 'A' + ten), nil
		default:
			return 0, fmt.Errorf("symbol at position %v should be a letter A-Z: %w", i, errInvalidISIN)
		}
	}

	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), nil
	case b >= 'A' && b <= 'Z':
		return int(b - 'A' + ten), nil
	default:
		return 0, fmt.Errorf("symbol at position %v should be either a digit 0-9 or a letter A-Z: %w", i, errInvalidISIN)
	}
}
