"""CoronaTrendVigor indicator package."""

from .corona_trend_vigor import (
    CoronaTrendVigor,
    CoronaTrendVigorSnapshot,
    Snapshot,
    TrendOutput,
)
from .output import CoronaTrendVigorOutput
from .params import (
    CoronaTrendVigorParams,
    Params,
    corona_trend_vigor_mnemonic,
    default_params,
)

__all__ = [
    "CoronaTrendVigor",
    "CoronaTrendVigorOutput",
    "CoronaTrendVigorParams",
    "CoronaTrendVigorSnapshot",
    "Params",
    "Snapshot",
    "TrendOutput",
    "corona_trend_vigor_mnemonic",
    "default_params",
]
