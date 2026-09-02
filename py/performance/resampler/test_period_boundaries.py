```python
import unittest
from datetime import datetime
from zoneinfo import ZoneInfo

from your_package.resampling import (
    DayBoundary,
    HourBoundary,
    MinuteBoundary,
)


class TestMinuteBoundary(unittest.TestCase):

    def test_default_is_one_minute(self):
        boundary = MinuteBoundary()

        timestamp = datetime(2026, 8, 31, 10, 37, 42, 123456)

        self.assertEqual(
            boundary.period_start(timestamp),
            datetime(2026, 8, 31, 10, 37),
        )

    def test_one_minute_period_start(self):
        boundary = MinuteBoundary(1)

        timestamp = datetime(2026, 8, 31, 10, 37, 42, 123456)

        self.assertEqual(
            boundary.period_start(timestamp),
            datetime(2026, 8, 31, 10, 37, 0),
        )

    def test_one_minute_next_period(self):
        boundary = MinuteBoundary(1)

        start = datetime(2026, 8, 31, 10, 37)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2026, 8, 31, 10, 38),
        )

    def test_five_minute_period_start(self):
        boundary = MinuteBoundary(5)

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 31, 10, 37, 42)
            ),
            datetime(2026, 8, 31, 10, 35),
        )

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 31, 10, 40, 1)
            ),
            datetime(2026, 8, 31, 10, 40),
        )

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 31, 10, 44, 59)
            ),
            datetime(2026, 8, 31, 10, 40),
        )

    def test_five_minute_next_period(self):
        boundary = MinuteBoundary(5)

        start = datetime(2026, 8, 31, 10, 35)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2026, 8, 31, 10, 40),
        )

    def test_valid_minute_sizes(self):
        for minutes in (1, 2, 3, 4, 5, 6, 10, 12, 15, 20, 30, 60):
            with self.subTest(minutes=minutes):
                boundary = MinuteBoundary(minutes)
                self.assertEqual(boundary.minutes, minutes)

    def test_invalid_minute_sizes(self):
        for minutes in (0, -1, -5, 7, 11, 13, 17, 25, 31):
            with self.subTest(minutes=minutes):
                with self.assertRaisesRegex(ValueError, "minutes"):
                    MinuteBoundary(minutes)

    def test_period_start_preserves_timezone(self):
        timezone = ZoneInfo("Europe/Amsterdam")

        timestamp = datetime(
            2026,
            8,
            31,
            10,
            37,
            42,
            tzinfo=timezone,
        )

        result = MinuteBoundary().period_start(timestamp)

        self.assertEqual(
            result,
            datetime(
                2026,
                8,
                31,
                10,
                37,
                tzinfo=timezone,
            ),
        )


class TestHourBoundary(unittest.TestCase):

    def test_default_is_one_hour(self):
        boundary = HourBoundary()

        timestamp = datetime(2026, 8, 31, 14, 37, 42, 123456)

        self.assertEqual(
            boundary.period_start(timestamp),
            datetime(2026, 8, 31, 14),
        )

    def test_one_hour_next_period(self):
        boundary = HourBoundary(1)

        start = datetime(2026, 8, 31, 14)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2026, 8, 31, 15),
        )

    def test_four_hour_period_start(self):
        boundary = HourBoundary(4)

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 31, 9, 15)
            ),
            datetime(2026, 8, 31, 8),
        )

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 31, 11, 59, 59)
            ),
            datetime(2026, 8, 31, 8),
        )

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 31, 12)
            ),
            datetime(2026, 8, 31, 12),
        )

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 31, 15, 59)
            ),
            datetime(2026, 8, 31, 12),
        )

    def test_four_hour_next_period(self):
        boundary = HourBoundary(4)

        start = datetime(2026, 8, 31, 12)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2026, 8, 31, 16),
        )

    def test_valid_hour_sizes(self):
        for hours in (1, 2, 3, 4, 6, 8, 12, 24):
            with self.subTest(hours=hours):
                boundary = HourBoundary(hours)
                self.assertEqual(boundary.hours, hours)

    def test_invalid_hour_sizes(self):
        for hours in (0, -1, -4, 5, 7, 10, 11, 13):
            with self.subTest(hours=hours):
                with self.assertRaisesRegex(ValueError, "hours"):
                    HourBoundary(hours)

    def test_period_start_preserves_timezone(self):
        timezone = ZoneInfo("Europe/Amsterdam")

        timestamp = datetime(
            2026,
            8,
            31,
            14,
            37,
            42,
            tzinfo=timezone,
        )

        result = HourBoundary().period_start(timestamp)

        self.assertEqual(
            result,
            datetime(
                2026,
                8,
                31,
                14,
                tzinfo=timezone,
            ),
        )


class TestDayBoundary(unittest.TestCase):

    def test_default_is_one_day(self):
        boundary = DayBoundary()

        timestamp = datetime(
            2026,
            8,
            31,
            14,
            37,
            42,
            123456,
        )

        self.assertEqual(
            boundary.period_start(timestamp),
            datetime(2026, 8, 31),
        )

    def test_one_day_next_period(self):
        boundary = DayBoundary()

        start = datetime(2026, 8, 31)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2026, 9, 1),
        )

    def test_day_boundary_at_midnight(self):
        boundary = DayBoundary()

        timestamp = datetime(2026, 8, 31, 0, 0, 0)

        self.assertEqual(
            boundary.period_start(timestamp),
            timestamp,
        )

    def test_day_boundary_before_midnight(self):
        boundary = DayBoundary()

        timestamp = datetime(
            2026,
            8,
            31,
            23,
            59,
            59,
            999999,
        )

        self.assertEqual(
            boundary.period_start(timestamp),
            datetime(2026, 8, 31),
        )

    def test_day_boundary_crosses_month(self):
        boundary = DayBoundary()

        start = datetime(2026, 8, 31)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2026, 9, 1),
        )

    def test_day_boundary_crosses_year(self):
        boundary = DayBoundary()

        start = datetime(2026, 12, 31)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2027, 1, 1),
        )

    def test_multiple_day_period_start(self):
        boundary = DayBoundary(3)

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 1, 12)
            ),
            datetime(2026, 8, 1),
        )

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 3, 12)
            ),
            datetime(2026, 8, 3),
        )

        self.assertEqual(
            boundary.period_start(
                datetime(2026, 8, 5, 12)
            ),
            datetime(2026, 8, 4),
        )

    def test_multiple_day_next_period(self):
        boundary = DayBoundary(3)

        start = datetime(2026, 8, 4)

        self.assertEqual(
            boundary.next_period_start(start),
            datetime(2026, 8, 7),
        )

    def test_valid_day_sizes(self):
        for days in (1, 2, 3, 7, 30):
            with self.subTest(days=days):
                boundary = DayBoundary(days)
                self.assertEqual(boundary.days, days)

    def test_invalid_day_sizes(self):
        for days in (0, -1, -7):
            with self.subTest(days=days):
                with self.assertRaisesRegex(ValueError, "days"):
                    DayBoundary(days)

    def test_period_start_preserves_timezone(self):
        timezone = ZoneInfo("Europe/Amsterdam")

        timestamp = datetime(
            2026,
            8,
            31,
            14,
            37,
            42,
            tzinfo=timezone,
        )

        result = DayBoundary().period_start(timestamp)

        self.assertEqual(
            result,
            datetime(
                2026,
                8,
                31,
                tzinfo=timezone,
            ),
        )


# ============================================================================
# Common PeriodBoundary contract
# ============================================================================

class TestPeriodBoundaryContract(unittest.TestCase):

    def test_period_start_is_idempotent(self):
        boundaries = (
            MinuteBoundary(1),
            MinuteBoundary(5),
            HourBoundary(1),
            HourBoundary(4),
            DayBoundary(1),
            DayBoundary(3),
        )

        timestamps = (
            datetime(2026, 8, 31, 10, 37, 42, 123456),
            datetime(2026, 8, 31, 23, 59, 59),
            datetime(2026, 12, 31, 23, 59, 59),
        )

        for boundary in boundaries:
            for timestamp in timestamps:
                with self.subTest(
                    boundary=boundary,
                    timestamp=timestamp,
                ):
                    start = boundary.period_start(timestamp)

                    self.assertEqual(
                        boundary.period_start(start),
                        start,
                    )

    def test_timestamp_belongs_to_returned_period(self):
        boundaries = (
            MinuteBoundary(1),
            MinuteBoundary(5),
            HourBoundary(1),
            HourBoundary(4),
            DayBoundary(1),
            DayBoundary(3),
        )

        timestamps = (
            datetime(2026, 8, 31, 10, 37, 42, 123456),
            datetime(2026, 8, 31, 23, 59, 59),
            datetime(2026, 12, 31, 23, 59, 59),
        )

        for boundary in boundaries:
            for timestamp in timestamps:
                with self.subTest(
                    boundary=boundary,
                    timestamp=timestamp,
                ):
                    start = boundary.period_start(timestamp)
                    next_start = boundary.next_period_start(start)

                    self.assertGreaterEqual(timestamp, start)
                    self.assertLess(timestamp, next_start)

    def test_next_period_is_after_period_start(self):
        boundaries = (
            MinuteBoundary(1),
            MinuteBoundary(5),
            HourBoundary(1),
            HourBoundary(4),
            DayBoundary(1),
            DayBoundary(3),
        )

        timestamp = datetime(2026, 8, 31, 10, 37, 42)

        for boundary in boundaries:
            with self.subTest(boundary=boundary):
                start = boundary.period_start(timestamp)
                next_start = boundary.next_period_start(start)

                self.assertGreater(next_start, start)


if __name__ == "__main__":
    unittest.main()
```
