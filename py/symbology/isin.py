"""ISIN (ISO 6166 International Securities Identifying Number) validation.

ISINs consist of three parts: a two letter country code, a nine character
alpha-numeric national security identifier, and a single check digit.

See https://en.wikipedia.org/wiki/International_Securities_Identification_Number.
"""

_TEN = 10

_ISIN_LENGTH = 12
_ISIN_CHECK_SUM_INDEX = _ISIN_LENGTH - 1
_ISIN_COUNTRY_LENGTH = 2

# Valid country codes grouped by first letter.
_VALID_COUNTRIES: dict[str, set[str]] = {
    'A': {'D', 'E', 'F', 'G', 'I', 'L', 'M', 'N', 'O', 'Q', 'R', 'S', 'T', 'U', 'W', 'Z'},
    'B': {'A', 'B', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'M', 'N', 'O', 'R', 'S', 'T', 'V', 'W', 'Y', 'Z'},
    'C': {'A', 'C', 'D', 'F', 'G', 'H', 'I', 'K', 'L', 'M', 'N', 'O', 'R', 'U', 'V', 'X', 'Y', 'Z'},
    'D': {'E', 'J', 'K', 'M', 'O', 'Z'},
    'E': {'C', 'E', 'G', 'R', 'S', 'T', 'U'},
    'F': {'I', 'J', 'K', 'M', 'O', 'R'},
    'G': {'A', 'B', 'D', 'E', 'G', 'H', 'I', 'L', 'M', 'N', 'Q', 'R', 'S', 'T', 'U', 'W', 'Y'},
    'H': {'K', 'M', 'N', 'R', 'T', 'U'},
    'I': {'D', 'E', 'L', 'M', 'N', 'O', 'Q', 'R', 'S', 'T'},
    'J': {'E', 'M', 'O', 'P'},
    'K': {'E', 'G', 'H', 'I', 'M', 'N', 'P', 'R', 'W', 'Y', 'Z'},
    'L': {'A', 'B', 'C', 'I', 'K', 'R', 'S', 'T', 'U', 'V', 'Y'},
    'M': {'A', 'C', 'D', 'E', 'G', 'H', 'K', 'L', 'M', 'N', 'O', 'P', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'},
    'N': {'A', 'C', 'E', 'F', 'G', 'I', 'L', 'O', 'P', 'R', 'S', 'U', 'Z'},
    'O': {'M'},
    'P': {'A', 'E', 'F', 'G', 'H', 'K', 'L', 'M', 'N', 'S', 'T', 'W', 'Y'},
    'Q': {'A', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'},
    'R': {'O', 'U', 'W'},
    'S': {'A', 'B', 'C', 'D', 'E', 'G', 'H', 'I', 'K', 'L', 'M', 'N', 'O', 'R', 'S', 'T', 'V', 'Y', 'Z'},
    'T': {'C', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'O', 'R', 'T', 'V', 'W', 'Z'},
    'U': {'A', 'G', 'M', 'S', 'Y', 'Z'},
    'V': {'A', 'C', 'E', 'G', 'I', 'N', 'U'},
    'W': {'F', 'S'},
    'X': {'A', 'B', 'C', 'D', 'F', 'K', 'L', 'S'},
    'Y': {'E', 'T'},
    'Z': {'A', 'M', 'W'},
}


class ISIN:
    """ISIN validator and check digit calculator."""

    def __init__(self, value: str) -> None:
        self._value = value

    def validate(self) -> None:
        """Validate the country code and the check digit of the ISIN.

        Raises:
            ValueError: If the ISIN is invalid.
        """
        if not self.validate_country():
            raise ValueError(f"unknown country code: invalid ISIN")

        self.validate_check_digit()

    def validate_check_digit(self) -> None:
        """Validate the check digit of the ISIN.

        Raises:
            ValueError: If the ISIN has an invalid check digit.
        """
        if len(self._value) != _ISIN_LENGTH:
            raise ValueError("length should be 12 symbols: invalid ISIN")

        n = self._value[_ISIN_CHECK_SUM_INDEX]
        if n < '0' or n > '9':
            raise ValueError("last symbol should be a digit 0-9: invalid ISIN")

        n_val = ord(n) - ord('0')

        d = self.calculate_check_digit()

        if n_val != d:
            raise ValueError("invalid check digit (last symbol): invalid ISIN")

    def calculate_check_digit(self) -> int:
        """Calculate a check digit of the ISIN according to the Luhn algorithm.

        Returns:
            The check digit as an integer 0-9.

        Raises:
            ValueError: If the ISIN is too short or contains invalid characters.
        """
        if len(self._value) < _ISIN_CHECK_SUM_INDEX:
            raise ValueError("length should be at least 11 symbols: invalid ISIN")

        total = 0
        multiply = True

        for i in range(_ISIN_CHECK_SUM_INDEX - 1, -1, -1):
            n = _to_ordinal_number_isin(self._value[i], i)

            if n < _TEN:
                if multiply:
                    n *= 2
                    total += n % _TEN + n // _TEN
                else:
                    total += n

                multiply = not multiply
            else:
                if multiply:
                    total += n // _TEN
                    n %= _TEN
                else:
                    total += n % _TEN
                    n //= _TEN

                n *= 2
                total += n % _TEN + n // _TEN

        total = (_TEN - total % _TEN) % _TEN

        return total

    def validate_country(self) -> bool:
        """Validate if two first letters of the ISIN represent a valid country code.

        Returns:
            True if the country code is valid, False otherwise.
        """
        if len(self._value) < _ISIN_COUNTRY_LENGTH:
            return False

        first = self._value[0]
        second = self._value[1]

        seconds = _VALID_COUNTRIES.get(first)
        if seconds is None:
            return False

        return second in seconds


def _to_ordinal_number_isin(ch: str, i: int) -> int:
    """Convert a character to its ordinal number for ISIN calculation.

    Raises:
        ValueError: If the character is invalid for the given position.
    """
    if i < _ISIN_COUNTRY_LENGTH:
        if 'A' <= ch <= 'Z':
            return ord(ch) - ord('A') + _TEN
        raise ValueError(
            f"symbol at position {i} should be a letter A-Z: invalid ISIN")

    if '0' <= ch <= '9':
        return ord(ch) - ord('0')
    if 'A' <= ch <= 'Z':
        return ord(ch) - ord('A') + _TEN

    raise ValueError(
        f"symbol at position {i} should be either a digit 0-9 or a letter A-Z: invalid ISIN")
