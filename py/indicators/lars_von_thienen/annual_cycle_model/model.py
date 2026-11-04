# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from dataclasses import asdict
from itertools import combinations
from ..cycle_scanner.scanner import ScanConfig, scan_cycles, modeled_value
from ..hodrick_prescott_smoothing_and_detrending.solver import _cycle_data
def pearson(left, right):
    """Population centered Pearson correlation; constant inputs are undefined."""
    if len(left) != len(right) or not left:
        raise ValueError('correlation requires nonempty equal-length vectors')
    a, b = math.fsum(left)/len(left), math.fsum(right)/len(right)
    x, y = [v-a for v in left], [v-b for v in right]
    xx, yy = math.fsum(v*v for v in x), math.fsum(v*v for v in y)
    if xx == 0 or yy == 0:
        return math.nan
    return max(-1.0, min(1.0, math.fsum(u*v for u,v in zip(x,y))/math.sqrt(xx*yy)))

def choose_subset(target, waves, periods, max_components=7):
    """Exhaustive subsets of a bounded pool using covariance/Gram sums.

    The score equals correlation(target, sum(waves_in_subset)); no amplitude
    refit, sign inversion or greedy replacement is added. Cache inner products
    so each candidate combination costs O(k^2) rather than O(N*k).
    """
    _integer(max_components, 'max_components')
    if not waves or len(waves) != len(periods) or len(waves) > 16:
        raise ValueError('require 1..16 candidate waves and matching periods')
    n = len(target)
    if n < 6 or any(len(wave) != n for wave in waves):
        raise ValueError('target and waves must share >= 6 samples')
    target_mean = math.fsum(target)/n
    tc = [value-target_mean for value in target]
    centered = []
    for wave in waves:
        mean = math.fsum(wave)/n
        centered.append([value-mean for value in wave])
    energy = math.fsum(value*value for value in tc)
    if energy == 0:
        raise ValueError('constant target has no defined correlation')
    cov = [math.fsum(a*b for a,b in zip(tc,wave)) for wave in centered]
    gram = [[math.fsum(a*b for a,b in zip(left,right)) for right in centered]
            for left in centered]
    best, best_score, best_key, checked = None, -math.inf, None, 0
    for count in range(1, min(max_components,len(waves))+1):
        for subset in combinations(range(len(waves)), count):
            checked += 1
            numerator = math.fsum(cov[i] for i in subset)
            denominator = math.fsum(gram[i][j] for i in subset for j in subset)
            if denominator <= 0:
                continue
            score = max(-1.0, min(1.0, numerator/math.sqrt(energy*denominator)))
            key = (count, tuple(sorted(periods[i] for i in subset)))
            if score > best_score+1e-12 or (abs(score-best_score) <= 1e-12 and key < best_key):
                best, best_score, best_key = subset, score, key
    if best is None or best_score <= 0:
        raise ValueError('no positive-correlation annual subset')
    return best, best_score, checked

def fit_annual_values(prices, year, min_period=5, max_period=125,
                      smoothing=1600.0, genuine_threshold=49.0,
                      candidate_pool=10, max_components=7, peak_distance=2):
    """Numerical annual fit. Dated completeness is enforced by decode_annual.

    phase_cycles=0 is essential: reconstruct the COMPLETE historical year,
    rather than projecting a phase fit that describes only its final subset.
    """
    _integer(year, 'year')
    _integer(candidate_pool, 'candidate_pool')
    _integer(max_components, 'max_components')
    if candidate_pool > 16 or max_components > 7:
        raise ValueError('candidate_pool <= 16 and max_components <= 7')
    config = ScanConfig(len(prices), min_period, max_period, smoothing,
                        genuine_threshold, 0.0, candidate_pool, peak_distance)
    scan = scan_cycles(prices, config)
    cycles = [cycle for cycle in scan.cycles if math.isfinite(cycle.period)]
    if not cycles:
        raise ValueError('no accepted annual cycles; inspect scanner diagnostics')
    n = len(prices)
    waves = [[modeled_value(cycle, i-(n-1)) for i in range(n)] for cycle in cycles]
    target = _cycle_data(prices, smoothing)
    subset, correlation, checked = choose_subset(target, waves,
                                                [cycle.period for cycle in cycles], max_components)
    selected = [cycles[i] for i in subset]
    selected.sort(key=lambda cycle: cycle.period)
    # Saved phases use the FIRST trading observation as origin, with zero-based
    # ordinal indices. Year boundaries are dates, never approximate 252-day lags.
    components = [dict(period=c.period, amplitude=c.fit_amplitude,
                       phase=(c.phase-2.0*math.pi*(n-1)/c.period) % (2.0*math.pi),
                       strength=c.strength, genuine_percent=c.genuine_percent)
                  for c in selected]
    return dict(schema='lvt-annual-cycle-model', version=1, research_reconstruction=True,
                year=year, sample_count=n, origin='first-observed-trading-day, zero-based ordinal',
                phase_convention='sine radians at ordinal zero', components=components,
                correlation=correlation, subsets_checked=checked, candidate_count=len(cycles),
                config=dict(asdict(config), candidate_pool=candidate_pool, max_components=max_components))

def annual_wave(model, index):
    return math.fsum(c['amplitude']*math.sin(2.0*math.pi*index/c['period']+c['phase'])
                     for c in model['components'])

class AnnualWaveReader:
    """Sequential model playback, no price fitting or price-input pretense."""
    def __init__(self, model):
        validate_annual_model(model)
        self.model = model
        self.reset()

    def reset(self):
        self.index = 0

    def update(self):
        value = annual_wave(self.model, self.index)
        self.index += 1
        return value

def validate_dates(dates, year=None):
    if not dates or any(type(value) is not date for value in dates):
        raise ValueError('require nonempty datetime.date sequence')
    if any(a >= b for a,b in zip(dates,dates[1:])):
        raise ValueError('dates must be unique and strictly increasing')
    if year is not None and any(value.year != year for value in dates):
        raise ValueError('every date must belong to the requested year')

def decode_annual(records, year, as_of=None, expected_dates=None,
                  confirm_complete=False, **parameters):
    _integer(year, 'year')
    as_of = date.today() if as_of is None else as_of
    if type(as_of) is not date or as_of < date(year+1,1,1):
        raise ValueError('annual decoding requires a completed year as of the analysis date')
    validate_dates([day for day,_ in records])
    rows = [(day,value) for day,value in records if day.year == year]
    dates = [day for day,_ in rows]
    validate_dates(dates, year)
    if expected_dates is not None:
        validate_dates(expected_dates, year)
        if dates != expected_dates:
            raise ValueError('observed dates must exactly match the supplied full-year trading calendar')
        coverage = 'exact match to supplied full-year calendar'
    else:
        # Month endpoints cannot establish absence of missing sessions. The
        # fallback requires an explicit caller assertion, recorded in the model.
        if confirm_complete is not True or dates[0].month != 1 or dates[-1].month != 12:
            raise ValueError('provide full-year calendar or confirm_complete=True with Jan..Dec coverage')
        coverage = 'caller asserted complete year; session completeness unverified'
    if dates[0].month != 1 or dates[-1].month != 12:
        raise ValueError('calendar must span January through December')
    model = fit_annual_values([value for _,value in rows], year, **parameters)
    model.update(start_date=dates[0].isoformat(), end_date=dates[-1].isoformat(),
                 dates=[day.isoformat() for day in dates], coverage=coverage,
                 fitted_as_of=as_of.isoformat())
    return model

def validate_annual_model(model):
    """Reject incompatible/NaN JSON models before any forecast computation."""
    if not isinstance(model, dict):
        raise ValueError('annual model must be a JSON object')
    if model.get('schema') != 'lvt-annual-cycle-model' or model.get('version') != 1:
        raise ValueError('unsupported annual model schema/version')
    _integer(model.get('year'), 'year')
    _integer(model.get('sample_count'), 'sample_count', 6)
    components = model.get('components')
    if not isinstance(components, list) or not 1 <= len(components) <= 7:
        raise ValueError('annual model requires 1..7 components')
    periods = []
    for component in components:
        if not isinstance(component, dict) or any(
                type(component.get(name)) not in (int, float)
                for name in ('period','amplitude','phase')):
            raise ValueError('model components require numeric period/amplitude/phase')
        period = _real(component.get('period'), 'period', 2.0000000000000004)
        _real(component.get('amplitude'), 'amplitude')
        phase = _sample(component.get('phase'))
        if not math.isfinite(phase):
            raise ValueError('model phase must be finite')
        periods.append(period)
    if len(set(periods)) != len(periods):
        raise ValueError('model contains duplicate wavelengths')
    if model.get('phase_convention') != 'sine radians at ordinal zero':
        raise ValueError('unsupported phase convention')
    if 'dates' in model:
        dates = [date.fromisoformat(day) for day in model['dates']]
        validate_dates(dates, model['year'])
        if len(dates) != model['sample_count']:
            raise ValueError('model dates must match sample_count')
    if 'fitted_as_of' in model and date.fromisoformat(model['fitted_as_of']) < date(model['year']+1,1,1):
        raise ValueError('saved annual fit predates completion of its source year')
