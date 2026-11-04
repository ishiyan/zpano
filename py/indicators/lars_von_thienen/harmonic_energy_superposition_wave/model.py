# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..annual_cycle_model.model import annual_wave, validate_annual_model
METONIC_LAGS = (8, 11, 19, 27, 30, 38)
def superposition_values(models, target_year, count, normalization='rms',
                         calendar_policy='extend', invert=False):
    """Align historical frozen waves at ordinal zero, then average six.

    RMS scaling uses each SOURCE year's full wave only. Extending to a longer
    target calendar evaluates the stored sinusoids; no time stretching occurs.
    """
    _integer(target_year, 'target_year', 39)
    _integer(count, 'count')
    if normalization not in ('none','rms') or calendar_policy not in ('strict','extend'):
        raise ValueError('normalization=none/rms; calendar_policy=strict/extend')
    if type(invert) is not bool:
        raise ValueError('invert must be a boolean')
    by_year = {}
    for model in models:
        validate_annual_model(model)
        if model['year'] in by_year:
            raise ValueError('duplicate source model year')
        by_year[model['year']] = model
    years = [target_year-lag for lag in METONIC_LAGS]
    if set(by_year) != set(years):
        raise ValueError('require exactly target years minus 8,11,19,27,30,38')
    source_waves, scales = [], []
    for year in years:
        model = by_year[year]
        n = model['sample_count']
        if calendar_policy == 'strict' and count != n:
            raise ValueError('strict ordinal alignment requires equal annual sample counts')
        base = [annual_wave(model,i) for i in range(n)]
        mean = math.fsum(base)/n if normalization == 'rms' else 0.0
        scale = math.sqrt(math.fsum((value-mean)**2 for value in base)/n) if normalization == 'rms' else 1.0
        if scale == 0:
            raise ValueError('zero-energy annual wave cannot be RMS normalized')
        source_waves.append([(annual_wave(model,i)-mean)/scale for i in range(count)])
        scales.append(dict(year=year, mean=mean, scale=scale, sample_count=n))
    sign = -1.0 if invert else 1.0
    values = [sign*math.fsum(wave[i] for wave in source_waves)/6.0 for i in range(count)]
    return values, dict(schema='lvt-superposition-model', version=1,
                        research_reconstruction=True, target_year=target_year,
                        source_years=years, lags=list(METONIC_LAGS), normalization=normalization,
                        calendar_policy=calendar_policy, invert=invert,
                        sample_count=count, scales=scales, annual_models=[by_year[year] for year in years])

class SuperpositionWaveReader:
    def __init__(self, models, target_year, count, normalization='rms', calendar_policy='extend', invert=False):
        self.values, self.model = superposition_values(models, target_year, count,
                                                       normalization, calendar_policy, invert)
        self.reset()

    def reset(self):
        self.index = 0

    def update(self):
        if self.index >= len(self.values):
            raise StopIteration('end of declared target calendar')
        value = self.values[self.index]
        self.index += 1
        return value


def validate_superposition_model(model: dict) -> None:
    """Validate the portable schema and its declared ordinal source mapping."""
    if model.get('schema') != 'lvt-superposition-model' or model.get('version') != 1:
        raise ValueError('unsupported superposition model')
    year = model['target_year']
    if model.get('lags') != list(METONIC_LAGS) or model.get('source_years') != [year-lag for lag in METONIC_LAGS]:
        raise ValueError('superposition source years and lags disagree')
    if type(model.get('invert')) is not bool:
        raise ValueError('inversion must be boolean')
    superposition_values(model['annual_models'], year, model['sample_count'], model['normalization'], model['calendar_policy'], model['invert'])
    if model.get('dates') is not None:
        from datetime import date
        from ..annual_cycle_model.model import validate_dates
        dates = [date.fromisoformat(day) for day in model['dates']]
        validate_dates(dates, year)
        if len(dates) != model['sample_count']:
            raise ValueError('dates must match sample_count')
