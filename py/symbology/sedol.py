"""SEDOL (Stock Exchange Daily Official List) validation.

SEDOL codes are seven characters in length, consisting of two parts:
a six-place alphanumeric code and a trailing check digit. There are
three types of SEDOL codes: old style, new style, and user defined.

See http://www.londonstockexchange.com/products-and-services/reference-data/sedol-master-file/sedol-master-file.htm.
"""

_TEN = 10

_SEDOL_LENGTH = 7
_SEDOL_CHECK_SUM_INDEX = _SEDOL_LENGTH - 1
_SEDOL_USER_DEFINED_CHARACTER = '9'
_SEDOL_USER_DEFINED = 1
_SEDOL_OLD_STYLE = 2
_SEDOL_NEW_STYLE = 3

_SEDOL_WEIGHTS = [1, 3, 1, 7, 3, 9]


class SEDOL:
    """SEDOL validator and check digit calculator."""

    def __init__(self, value: str) -> None:
        self._value = value

    def validate(self) -> None:
        """Validate the SEDOL.

        Raises:
            ValueError: If the SEDOL is invalid.
        """
        if len(self._value) < _SEDOL_LENGTH:
            raise ValueError("length should be 7 symbols: invalid SEDOL")

        n = self._value[_SEDOL_CHECK_SUM_INDEX]
        if n < '0' or n > '9':
            raise ValueError("last symbol should be a digit 0-9: invalid SEDOL")

        n_val = ord(n) - ord('0')

        d = self.calculate_check_digit()

        if n_val != d:
            raise ValueError("invalid check digit (last symbol): invalid SEDOL")

    def calculate_check_digit(self) -> int:
        """Calculate a check digit of the SEDOL.

        Returns:
            The check digit as an integer 0-9.

        Raises:
            ValueError: If the SEDOL is too short or contains invalid characters.
        """
        if len(self._value) < _SEDOL_CHECK_SUM_INDEX:
            raise ValueError("length should be at least 6 symbols: invalid SEDOL")

        style = _SEDOL_NEW_STYLE
        total = 0

        for i in range(_SEDOL_CHECK_SUM_INDEX):
            b = self._value[i]

            if '0' <= b <= '9':
                n = ord(b) - ord('0')

                if i == 0:
                    if b == _SEDOL_USER_DEFINED_CHARACTER:
                        style = _SEDOL_USER_DEFINED
                    else:
                        style = _SEDOL_OLD_STYLE
            elif 'A' <= b <= 'Z':
                if style == _SEDOL_OLD_STYLE:
                    raise ValueError(
                        f"symbol at position {i} should be a digit 0-9 "
                        f"in old style SEDOL: invalid SEDOL")

                if style == _SEDOL_NEW_STYLE:
                    if b in ('A', 'E', 'U', 'I', 'O'):
                        raise ValueError(
                            f"symbol at position {i} should not be a vowel "
                            f"AEUIO in user defined SEDOL: invalid SEDOL")

                n = ord(b) - ord('A') + _TEN
            else:
                raise ValueError(
                    f"symbol at position {i} should be either a digit 0-9 "
                    f"or a letter A-Z: invalid SEDOL")

            total += n * _SEDOL_WEIGHTS[i]

        total = (_TEN - total % _TEN) % _TEN

        return total
