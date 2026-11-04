# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..harmonic_cycle.validation import _integer, _real, _sample, _finite
from ..metonic.calendar import _Calendar, _DelayMean
from ..zero_lag_exponential_moving_average.smoother import _ZLEMA
class SentimentOutput(NamedTuple):
    value: float
    daily_sentiment: float
    smoothed_close: float
    smoothed_open: float

class MetonicSentimentIndex:
    """Original open-close score or documented close-change variant.

    mode=0 is ORIGINAL open/close, mode=1 is chapter 11's close-only variant.
    Feed every calendar day, filling absent days BEFORE ZLEMA smoothing.
    """
    def __init__(self, length=13, delay1=6939, delay2=4016, delay3=2922, mode=0):
        self.length = _integer(length,'length')
        self.mode = _integer(mode,'mode',0)
        if self.mode not in (0,1):
            raise ValueError("mode must be 0 (open-close) or 1 (close-change)")

        self.delay1 = _integer(delay1,'delay1')
        self.delay2 = _integer(delay2,'delay2')
        self.delay3 = _integer(delay3,'delay3')
        self.delays = (self.delay1,self.delay2,self.delay3)
        if len(set(self.delays))!=3:
            raise ValueError("three delays must be distinct")

        self.daily_warmup_bars = (self.length-1)//2 + (1 if self.mode==1 else 0)
        self.warmup_bars = max(self.delays)+self.daily_warmup_bars
        self.reset()

    def reset(self):
        self.samples_seen = 0
        self.close_smoother = _ZLEMA(self.length)
        self.open_smoother = _ZLEMA(self.length)
        self.calendar = _Calendar()
        self.delay = _DelayMean(self.delays)
        self.previous_close = math.nan
        self.last_value = math.nan

    @property
    def is_primed(self):
        return not math.isnan(self.last_value)

    def update(self, close, open_price=None, time=None):
        close = _sample(close)
        if math.isnan(close):
            return SentimentOutput(*(math.nan,)*4)
        open_value = math.nan
        if self.mode==0:
            if open_price is None:
                raise ValueError("original mode requires open_price")
            open_value = _sample(open_price)
            if math.isnan(open_value):
                return SentimentOutput(*(math.nan,)*4)
        self.calendar.validate(time)
        smoothed_close = self.close_smoother.update(close)
        smoothed_open = self.open_smoother.update(open_value) if self.mode==0 else math.nan
        comparison = smoothed_open if self.mode==0 else self.previous_close
        sentiment = math.nan
        if self.samples_seen>=self.daily_warmup_bars and not math.isnan(comparison):
            # Source sends ties to -1; no neutral sentiment is introduced.
            sentiment = 1.0 if smoothed_close>comparison else -1.0
        self.previous_close = smoothed_close
        mean = self.delay.update(sentiment)
        self.last_value = 3.0*mean if not math.isnan(mean) else math.nan
        self.samples_seen += 1
        self.calendar.commit(time)
        return SentimentOutput(self.last_value,sentiment,smoothed_close,smoothed_open)

Engine = MetonicSentimentIndex
