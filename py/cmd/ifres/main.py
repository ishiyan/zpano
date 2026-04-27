#!/usr/bin/env python3
"""ifres — command-line indicator frequency response calculator.

Reads a JSON settings file containing indicator definitions,
creates indicator instances, determines each indicator's warmup period,
and calculates the frequency response with signal length 1024.

Usage: python -m py.cmd.ifres <settings.json>
"""

from __future__ import annotations

import json
import sys

from py.indicators.core.frequency_response import FrequencyResponse, Component, calculate
from py.indicators.core.indicator import Indicator
from py.indicators.factory.factory import create_indicator

# Reuse the identifier map and param conversion from icalc.
from py.cmd.icalc.main import _IDENTIFIER_MAP, _convert_params

SIGNAL_LENGTH = 1024
MAX_WARMUP = 10000
PHASE_DEGREES_UNWRAPPING_LIMIT = 179.0


class _IndicatorUpdater:
    """Adapts an Indicator to the frequency_response Updater protocol."""

    def __init__(self, ind: Indicator) -> None:
        self._ind = ind

    def metadata(self):
        return self._ind.metadata()

    def update(self, sample: float) -> float:
        return self._ind.update(sample)  # type: ignore[attr-defined]


def _detect_warmup(updater, ind: Indicator) -> int:
    """Feed zeros into the indicator until it is primed."""
    for i in range(MAX_WARMUP):
        if ind.is_primed():
            return i
        updater.update(0.0)
    return MAX_WARMUP


def _print_component(name: str, c: Component) -> None:
    n = len(c.data)
    print(f"  {name:<25s} min={c.min:10.4f}  max={c.max:10.4f}", end='')

    if n == 0:
        print()
        return

    preview = 3
    if n <= preview * 2:
        print(f"  data={c.data}", end='')
    else:
        print(f"  data=[{c.data[0]:.4f} {c.data[1]:.4f} {c.data[2]:.4f} "
              f"... {c.data[n-3]:.4f} {c.data[n-2]:.4f} {c.data[n-1]:.4f}]", end='')

    print()


def _print_frequency_response(fr: FrequencyResponse, warmup: int) -> None:
    print(f"=== {fr.label} (warmup={warmup}) ===")
    print(f"  Spectrum length: {len(fr.normalized_frequency)}")

    _print_component("PowerPercent", fr.power_percent)
    _print_component("PowerDecibel", fr.power_decibel)
    _print_component("AmplitudePercent", fr.amplitude_percent)
    _print_component("AmplitudeDecibel", fr.amplitude_decibel)
    _print_component("PhaseDegrees", fr.phase_degrees)
    _print_component("PhaseDegreesUnwrapped", fr.phase_degrees_unwrapped)

    print()


def main() -> None:
    if len(sys.argv) < 2:
        print("usage: python -m py.cmd.ifres <settings.json>", file=sys.stderr)
        sys.exit(1)

    with open(sys.argv[1]) as f:
        entries = json.load(f)

    for e in entries:
        id_str = e['identifier']
        ident = _IDENTIFIER_MAP.get(id_str)
        if ident is None:
            print(f"error: unknown indicator identifier: {id_str}", file=sys.stderr)
            sys.exit(1)

        params = _convert_params(e.get('params', {}))

        # Create a probe instance to determine warmup period.
        probe = create_indicator(ident, params if params else None)
        if not hasattr(probe, 'update'):
            print(f"indicator {id_str} does not support frequency response (no update method)",
                  file=sys.stderr)
            continue

        probe_updater = _IndicatorUpdater(probe)
        warmup = _detect_warmup(probe_updater, probe)

        # Create a fresh instance for actual calculation.
        ind = create_indicator(ident, params if params else None)
        updater = _IndicatorUpdater(ind)

        fr = calculate(SIGNAL_LENGTH, updater, warmup, PHASE_DEGREES_UNWRAPPING_LIMIT)
        _print_frequency_response(fr, warmup)


if __name__ == '__main__':
    main()
