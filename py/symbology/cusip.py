"""CUSIP (Committee on Uniform Security Identification Procedures) validation.

The CUSIP number consists of a base number of six characters known as the
issuer number and a two character suffix known as the issue number. The 9th
character is a check digit.

See https://www.cusip.com/static/html/cusipaccess/CUSIPIntro_%207.26.2007.pdf.
"""

_TEN = 10

_CUSIP_LENGTH = 9
_CUSIP_CHECK_SUM_INDEX = _CUSIP_LENGTH - 1
_CUSIP_ISSUE_SECOND_INDEX = _CUSIP_LENGTH - 2
_CUSIP_ISSUE_FIRST_INDEX = _CUSIP_LENGTH - 3


class CUSIP:
    """CUSIP validator and check digit calculator."""

    def __init__(self, value: str) -> None:
        self._value = value

    def validate(self) -> None:
        """Validate the CUSIP.

        Raises:
            ValueError: If the CUSIP is invalid.
        """
        if len(self._value) < _CUSIP_LENGTH:
            raise ValueError("length should be 9 symbols: invalid CUSIP")

        n = self._value[_CUSIP_CHECK_SUM_INDEX]
        if n < '0' or n > '9':
            raise ValueError("last symbol should be a digit 0-9: invalid CUSIP")

        n_val = ord(n) - ord('0')

        d = self.calculate_check_digit()

        if n_val != d:
            # A fix for incorrect CUSIPs in SEC 13F Security List.
            # See https://quant.stackexchange.com/questions/16392/sec-13f-security-list-has-incorrect-cusip-numbers.
            if self._value[_CUSIP_ISSUE_FIRST_INDEX] == '9':
                if self._value[_CUSIP_ISSUE_SECOND_INDEX] in ('0', '5'):
                    return

            raise ValueError("invalid check digit (last symbol): invalid CUSIP")

    def calculate_check_digit(self) -> int:
        """Calculate a check digit of the CUSIP according to the Luhn algorithm.

        Returns:
            The check digit as an integer 0-9.

        Raises:
            ValueError: If the CUSIP is too short or contains invalid characters.
        """
        if len(self._value) < _CUSIP_CHECK_SUM_INDEX:
            raise ValueError("length should be at least 8 symbols: invalid CUSIP")

        total = 0

        for i in range(_CUSIP_CHECK_SUM_INDEX):
            n = _to_ordinal_number_cusip(self._value[i], i)

            if i % 2 == 1:
                n *= 2

            total += n // _TEN + n % _TEN

        total = (_TEN - total % _TEN) % _TEN

        return total


def _to_ordinal_number_cusip(ch: str, i: int) -> int:
    """Convert a character to its ordinal number for CUSIP calculation.

    Raises:
        ValueError: If the character is invalid for the given position.
    """
    if '0' <= ch <= '9':
        return ord(ch) - ord('0')
    if 'A' <= ch <= 'Z':
        return ord(ch) - ord('A') + _TEN
    if ch == '*':
        return 36
    if ch == '@':
        return 37
    if ch == '#':
        return 38

    raise ValueError(
        f"symbol at position {i} should be either a digit 0-9, "
        f"a letter A-Z or special symbols @*#: invalid CUSIP")
