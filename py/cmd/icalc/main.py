#!/usr/bin/env python3
"""icalc — command-line indicator calculator.

Reads a JSON settings file containing indicator definitions,
creates indicator instances, prints their metadata, then iterates
through embedded bar data printing bar values and all indicator
outputs on each iteration.

Usage: python -m py.cmd.icalc <settings.json>
"""

from __future__ import annotations

import json
import math
import re
import sys
from datetime import datetime, timezone

from py.entities.bar import Bar
from py.entities.scalar import Scalar
from py.indicators.core.identifier import Identifier
from py.indicators.core.indicator import Indicator
from py.indicators.core.outputs.band import Band
from py.indicators.factory.factory import create_indicator


# ---------------------------------------------------------------------------
# camelCase identifier string → Identifier enum mapping
# ---------------------------------------------------------------------------

_IDENTIFIER_MAP: dict[str, Identifier] = {
    'simpleMovingAverage': Identifier.SIMPLE_MOVING_AVERAGE,
    'weightedMovingAverage': Identifier.WEIGHTED_MOVING_AVERAGE,
    'triangularMovingAverage': Identifier.TRIANGULAR_MOVING_AVERAGE,
    'exponentialMovingAverage': Identifier.EXPONENTIAL_MOVING_AVERAGE,
    'doubleExponentialMovingAverage': Identifier.DOUBLE_EXPONENTIAL_MOVING_AVERAGE,
    'tripleExponentialMovingAverage': Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE,
    't2ExponentialMovingAverage': Identifier.T2_EXPONENTIAL_MOVING_AVERAGE,
    't3ExponentialMovingAverage': Identifier.T3_EXPONENTIAL_MOVING_AVERAGE,
    'kaufmanAdaptiveMovingAverage': Identifier.KAUFMAN_ADAPTIVE_MOVING_AVERAGE,
    'jurikMovingAverage': Identifier.JURIK_MOVING_AVERAGE,
    'mesaAdaptiveMovingAverage': Identifier.MESA_ADAPTIVE_MOVING_AVERAGE,
    'fractalAdaptiveMovingAverage': Identifier.FRACTAL_ADAPTIVE_MOVING_AVERAGE,
    'dominantCycle': Identifier.DOMINANT_CYCLE,
    'momentum': Identifier.MOMENTUM,
    'rateOfChange': Identifier.RATE_OF_CHANGE,
    'rateOfChangePercent': Identifier.RATE_OF_CHANGE_PERCENT,
    'rateOfChangeRatio': Identifier.RATE_OF_CHANGE_RATIO,
    'relativeStrengthIndex': Identifier.RELATIVE_STRENGTH_INDEX,
    'chandeMomentumOscillator': Identifier.CHANDE_MOMENTUM_OSCILLATOR,
    'bollingerBands': Identifier.BOLLINGER_BANDS,
    'bollingerBandsTrend': Identifier.BOLLINGER_BANDS_TREND,
    'variance': Identifier.VARIANCE,
    'standardDeviation': Identifier.STANDARD_DEVIATION,
    'goertzelSpectrum': Identifier.GOERTZEL_SPECTRUM,
    'maximumEntropySpectrum': Identifier.MAXIMUM_ENTROPY_SPECTRUM,
    'centerOfGravityOscillator': Identifier.CENTER_OF_GRAVITY_OSCILLATOR,
    'cyberCycle': Identifier.CYBER_CYCLE,
    'instantaneousTrendLine': Identifier.INSTANTANEOUS_TREND_LINE,
    'superSmoother': Identifier.SUPER_SMOOTHER,
    'zeroLagExponentialMovingAverage': Identifier.ZERO_LAG_EXPONENTIAL_MOVING_AVERAGE,
    'zeroLagErrorCorrectingExponentialMovingAverage': Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE,
    'roofingFilter': Identifier.ROOFING_FILTER,
    'trueRange': Identifier.TRUE_RANGE,
    'averageTrueRange': Identifier.AVERAGE_TRUE_RANGE,
    'normalizedAverageTrueRange': Identifier.NORMALIZED_AVERAGE_TRUE_RANGE,
    'directionalMovementMinus': Identifier.DIRECTIONAL_MOVEMENT_MINUS,
    'directionalMovementPlus': Identifier.DIRECTIONAL_MOVEMENT_PLUS,
    'directionalIndicatorMinus': Identifier.DIRECTIONAL_INDICATOR_MINUS,
    'directionalIndicatorPlus': Identifier.DIRECTIONAL_INDICATOR_PLUS,
    'directionalMovementIndex': Identifier.DIRECTIONAL_MOVEMENT_INDEX,
    'averageDirectionalMovementIndex': Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX,
    'averageDirectionalMovementIndexRating': Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX_RATING,
    'williamsPercentR': Identifier.WILLIAMS_PERCENT_R,
    'percentagePriceOscillator': Identifier.PERCENTAGE_PRICE_OSCILLATOR,
    'absolutePriceOscillator': Identifier.ABSOLUTE_PRICE_OSCILLATOR,
    'commodityChannelIndex': Identifier.COMMODITY_CHANNEL_INDEX,
    'moneyFlowIndex': Identifier.MONEY_FLOW_INDEX,
    'onBalanceVolume': Identifier.ON_BALANCE_VOLUME,
    'balanceOfPower': Identifier.BALANCE_OF_POWER,
    'pearsonsCorrelationCoefficient': Identifier.PEARSONS_CORRELATION_COEFFICIENT,
    'linearRegression': Identifier.LINEAR_REGRESSION,
    'ultimateOscillator': Identifier.ULTIMATE_OSCILLATOR,
    'stochasticRelativeStrengthIndex': Identifier.STOCHASTIC_RELATIVE_STRENGTH_INDEX,
    'stochastic': Identifier.STOCHASTIC,
    'aroon': Identifier.AROON,
    'advanceDecline': Identifier.ADVANCE_DECLINE,
    'advanceDeclineOscillator': Identifier.ADVANCE_DECLINE_OSCILLATOR,
    'parabolicStopAndReverse': Identifier.PARABOLIC_STOP_AND_REVERSE,
    'tripleExponentialMovingAverageOscillator': Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE_OSCILLATOR,
    'movingAverageConvergenceDivergence': Identifier.MOVING_AVERAGE_CONVERGENCE_DIVERGENCE,
    'sineWave': Identifier.SINE_WAVE,
    'hilbertTransformerInstantaneousTrendLine': Identifier.HILBERT_TRANSFORMER_INSTANTANEOUS_TREND_LINE,
    'trendCycleMode': Identifier.TREND_CYCLE_MODE,
    'coronaSpectrum': Identifier.CORONA_SPECTRUM,
    'coronaSignalToNoiseRatio': Identifier.CORONA_SIGNAL_TO_NOISE_RATIO,
    'coronaSwingPosition': Identifier.CORONA_SWING_POSITION,
    'coronaTrendVigor': Identifier.CORONA_TREND_VIGOR,
    'adaptiveTrendAndCycleFilter': Identifier.ADAPTIVE_TREND_AND_CYCLE_FILTER,
    'discreteFourierTransformSpectrum': Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
    'combBandPassSpectrum': Identifier.COMB_BAND_PASS_SPECTRUM,
    'autoCorrelationIndicator': Identifier.AUTO_CORRELATION_INDICATOR,
    'autoCorrelationPeriodogram': Identifier.AUTO_CORRELATION_PERIODOGRAM,
    'jurikRelativeTrendStrengthIndex': Identifier.JURIK_RELATIVE_TREND_STRENGTH_INDEX,
    'jurikCompositeFractalBehaviorIndex': Identifier.JURIK_COMPOSITE_FRACTAL_BEHAVIOR_INDEX,
    'jurikZeroLagVelocity': Identifier.JURIK_ZERO_LAG_VELOCITY,
    'jurikDirectionalMovementIndex': Identifier.JURIK_DIRECTIONAL_MOVEMENT_INDEX,
}


def _camel_to_snake(name: str) -> str:
    """Convert camelCase to snake_case."""
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()


def _convert_params(params: dict) -> dict:
    """Convert camelCase param keys to snake_case."""
    return {_camel_to_snake(k): v for k, v in params.items()}


# ---------------------------------------------------------------------------
# Metadata printing
# ---------------------------------------------------------------------------

def _print_metadata(indicators: list[Indicator]) -> None:
    print("=== Indicator Metadata ===")
    print()

    for i, ind in enumerate(indicators):
        meta = ind.metadata()
        print(f"[{i}] {meta.mnemonic}")
        print(f"  Identifier:  {meta.identifier.name}")
        print(f"  Description: {meta.description}")
        print(f"  Outputs ({len(meta.outputs)}):")

        for j, out in enumerate(meta.outputs):
            print(f"    [{j}] kind={out.kind} shape={out.shape} "
                  f"mnemonic={out.mnemonic!r} description={out.description!r}")

        meta_dict = {
            'identifier': int(meta.identifier),
            'mnemonic': meta.mnemonic,
            'description': meta.description,
            'outputs': [
                {
                    'kind': int(out.kind),
                    'shape': int(out.shape),
                    'mnemonic': out.mnemonic,
                    'description': out.description,
                }
                for out in meta.outputs
            ],
        }
        meta_json = json.dumps(meta_dict, indent=2)
        # Indent each line by 2 spaces
        indented = '\n'.join('  ' + line for line in meta_json.split('\n'))
        print(f"  Full metadata JSON:\n{indented}")
        print()


# ---------------------------------------------------------------------------
# Output printing
# ---------------------------------------------------------------------------

def _print_output(outputs_meta: list, output: list) -> None:
    parts: list[str] = []
    for i, val in enumerate(output):
        name = f"out[{i}]"
        if i < len(outputs_meta):
            name = outputs_meta[i].mnemonic

        if isinstance(val, Scalar):
            if math.isnan(val.value):
                parts.append(f"{name}=NaN")
            else:
                parts.append(f"{name}={val.value:.4f}")
        elif isinstance(val, Band):
            if val.is_empty():
                parts.append(f"{name}=Band(NaN)")
            else:
                parts.append(f"{name}=Band({val.lower:.4f},{val.upper:.4f})")
        else:
            parts.append(f"{name}={val}")

    print(' '.join(parts), end='')


# ---------------------------------------------------------------------------
# Embedded test data (252 bars, identical to Go/TS icalc)
# ---------------------------------------------------------------------------

def _test_highs() -> list[float]:
    return [
        93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,
        96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
        90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
        87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,
        84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
        90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
        90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,
        89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
        104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,
        123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,
        116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,
        116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,
        125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
        138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
        125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,
        124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
        124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,
        132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,
        129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,
        123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,
        108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,
        96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
        95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,
        104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,
        107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,
        110.5, 109.5,
    ]


def _test_lows() -> list[float]:
    return [
        90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,
        92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
        87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
        84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
        82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,
        89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,
        88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,
        86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
        101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,
        117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
        112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
        112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,
        123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,
        135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
        122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,
        121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,
        122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,
        128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,
        126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,
        119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,
        105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
        93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,
        93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,
        102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
        104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,
        108.56, 106.62,
    ]


def _test_closes() -> list[float]:
    return [
        91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
        96.125, 97.25, 98.5, 89.875, 91, 92.815, 89.155, 89.345, 91.625, 89.875,
        88.375, 87.625, 84.78, 83, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
        85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
        83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89, 91, 90.5,
        89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
        88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90, 89.69,
        88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106, 102.5,
        102.405, 104.595, 106.125, 106, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
        123, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
        116, 116, 112, 113.75, 112.94, 116, 120.5, 116.62, 117, 115.25,
        114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
        123.12, 122.56, 124.62, 129.25, 131, 132.25, 131, 132.81, 134, 137.38,
        137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129, 123.87, 124.81,
        123, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
        122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
        124.44, 122, 122.37, 122.94, 124, 123.19, 124.56, 127.25, 125.87, 128.86,
        132, 130.75, 134.75, 135, 132.38, 133.31, 131.94, 130, 125.37, 130.13,
        127.12, 125.19, 122, 125, 123, 123.5, 120.06, 121, 117.75, 119.87,
        122, 119.19, 116.37, 113.5, 114.25, 110, 105.06, 107, 107.87, 107,
        107.12, 107, 91, 93.94, 93.87, 95.5, 93, 94.94, 98.25, 96.75,
        94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97, 95, 95.87, 94.06,
        94.62, 93.75, 98, 103.94, 107.87, 106.06, 104.5, 105, 104.19, 103.06,
        103.42, 105.27, 111.87, 116, 116.62, 118.28, 113.37, 109, 109.7, 109.25,
        107, 109.19, 110, 109.2, 110.12, 108, 108.62, 109.75, 109.81, 109,
        108.75, 107.87,
    ]


def _test_volumes() -> list[float]:
    return [
        4077500, 4955900, 4775300, 4155300, 4593100, 3631300, 3382800, 4954200, 4500000, 3397500,
        4204500, 6321400, 10203600, 19043900, 11692000, 9553300, 8920300, 5970900, 5062300, 3705600,
        5865600, 5603000, 5811900, 8483800, 5995200, 5408800, 5430500, 6283800, 5834800, 4515500,
        4493300, 4346100, 3700300, 4600200, 4557200, 4323600, 5237500, 7404100, 4798400, 4372800,
        3872300, 10750800, 5804800, 3785500, 5014800, 3507700, 4298800, 4842500, 3952200, 3304700,
        3462000, 7253900, 9753100, 5953000, 5011700, 5910800, 4916900, 4135000, 4054200, 3735300,
        2921900, 2658400, 4624400, 4372200, 5831600, 4268600, 3059200, 4495500, 3425000, 3630800,
        4168100, 5966900, 7692800, 7362500, 6581300, 19587700, 10378600, 9334700, 10467200, 5671400,
        5645000, 4518600, 4519500, 5569700, 4239700, 4175300, 4995300, 4776600, 4190000, 6035300,
        12168900, 9040800, 5780300, 4320800, 3899100, 3221400, 3455500, 4304200, 4703900, 8316300,
        10553900, 6384800, 7163300, 7007800, 5114100, 5263800, 6666100, 7398400, 5575000, 4852300,
        4298100, 4900500, 4887700, 6964800, 4679200, 9165000, 6469800, 6792000, 4423800, 5231900,
        4565600, 6235200, 5225900, 8261400, 5912500, 3545600, 5714500, 6653900, 6094500, 4799200,
        5050800, 5648900, 4726300, 5585600, 5124800, 7630200, 14311600, 8793600, 8874200, 6966600,
        5525500, 6515500, 5291900, 5711700, 4327700, 4568000, 6859200, 5757500, 7367000, 6144100,
        4052700, 5849700, 5544700, 5032200, 4400600, 4894100, 5140000, 6610900, 7585200, 5963100,
        6045500, 8443300, 6464700, 6248300, 4357200, 4774700, 6216900, 6266900, 5584800, 5284500,
        7554500, 7209500, 8424800, 5094500, 4443600, 4591100, 5658400, 6094100, 14862200, 7544700,
        6985600, 8093000, 7590000, 7451300, 7078000, 7105300, 8778800, 6643900, 10563900, 7043100,
        6438900, 8057700, 14240000, 17872300, 7831100, 8277700, 15017800, 14183300, 13921100, 9683000,
        9187300, 11380500, 69447300, 26673600, 13768400, 11371600, 9872200, 9450500, 11083300, 9552800,
        11108400, 10374200, 16701900, 13741900, 8523600, 9551900, 8680500, 7151700, 9673100, 6264700,
        8541600, 8358000, 18720800, 19683100, 13682500, 10668100, 9710600, 3113100, 5682000, 5763600,
        5340000, 6220800, 14680500, 9933000, 11329500, 8145300, 16644700, 12593800, 7138100, 7442300,
        9442300, 7123600, 7680600, 4839800, 4775500, 4008800, 4533600, 3741100, 4084800, 2685200,
        3438000, 2870500,
    ]


def _test_bars() -> list[Bar]:
    """Build 252 test bars from embedded TA-Lib reference data."""
    highs = _test_highs()
    lows = _test_lows()
    closes = _test_closes()
    volumes = _test_volumes()

    bars: list[Bar] = []
    base_time = datetime(2020, 1, 2, tzinfo=timezone.utc)

    for i in range(len(closes)):
        open_price = closes[0] if i == 0 else closes[i - 1]
        from datetime import timedelta
        t = base_time + timedelta(days=i)

        bars.append(Bar(
            time=t,
            open=open_price,
            high=highs[i],
            low=lows[i],
            close=closes[i],
            volume=volumes[i],
        ))

    return bars


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    if len(sys.argv) < 2:
        print("usage: python -m py.cmd.icalc <settings.json>", file=sys.stderr)
        sys.exit(1)

    with open(sys.argv[1]) as f:
        entries = json.load(f)

    indicators: list[Indicator] = []
    for e in entries:
        id_str = e['identifier']
        ident = _IDENTIFIER_MAP.get(id_str)
        if ident is None:
            print(f"error: unknown indicator identifier: {id_str}", file=sys.stderr)
            sys.exit(1)

        params = _convert_params(e.get('params', {}))
        ind = create_indicator(ident, params if params else None)
        indicators.append(ind)

    _print_metadata(indicators)

    bars = _test_bars()

    print()
    print("=== Bar Data & Indicator Outputs ===")
    print()

    for i, bar in enumerate(bars):
        print(f"Bar[{i:3d}] {bar.time.strftime('%Y-%m-%d')}  "
              f"O={bar.open:.4f} H={bar.high:.4f} L={bar.low:.4f} "
              f"C={bar.close:.4f} V={bar.volume:.0f}")

        for ind in indicators:
            meta = ind.metadata()
            output = ind.update_bar(bar)

            print(f"  {meta.mnemonic:<45s} primed={str(ind.is_primed()):<5s} ", end='')
            _print_output(meta.outputs, output)
            print()

        print()


if __name__ == '__main__':
    main()
