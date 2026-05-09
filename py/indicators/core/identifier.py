"""Enumerates all implemented indicators."""

from enum import IntEnum


class Identifier(IntEnum):
    """Identifies an indicator by enumerating all implemented indicators."""

    # ── common ────────────────────────────────────────────────────────────

    # Identifies the Absolute Price Oscillator (APO) indicator.
    ABSOLUTE_PRICE_OSCILLATOR = 0

    # Identifies the Exponential Moving Average (EMA) indicator.
    EXPONENTIAL_MOVING_AVERAGE = 1

    # Identifies the Linear Regression (LINEARREG) indicator.
    LINEAR_REGRESSION = 2

    # Identifies the Momentum (MOM) indicator.
    MOMENTUM = 3

    # Identifies the Pearson's Correlation Coefficient (CORREL) indicator.
    PEARSONS_CORRELATION_COEFFICIENT = 4

    # Identifies the Rate of Change (ROC) indicator.
    RATE_OF_CHANGE = 5

    # Identifies the Rate of Change Percent (ROCP) indicator.
    RATE_OF_CHANGE_PERCENT = 6

    # Identifies the Rate of Change Ratio (ROCR / ROCR100) indicator.
    RATE_OF_CHANGE_RATIO = 7

    # Identifies the Simple Moving Average (SMA) indicator.
    SIMPLE_MOVING_AVERAGE = 8

    # Identifies the Standard Deviation (STDEV) indicator.
    STANDARD_DEVIATION = 9

    # Identifies the Triangular Moving Average (TRIMA) indicator.
    TRIANGULAR_MOVING_AVERAGE = 10

    # Identifies the Variance (VAR) indicator.
    VARIANCE = 11

    # Identifies the Weighted Moving Average (WMA) indicator.
    WEIGHTED_MOVING_AVERAGE = 12

    # ── arnaud legoux ──────────────────────────────────────────────────────

    # Identifies the Arnaud Legoux Moving Average (ALMA) indicator.
    ARNAUD_LEGOUX_MOVING_AVERAGE = 13

    # ── donald lambert ─────────────────────────────────────────────────────

    # Identifies the Donald Lambert Commodity Channel Index (CCI) indicator.
    COMMODITY_CHANNEL_INDEX = 14

    # ── gene quong ─────────────────────────────────────────────────────────

    # Identifies the Gene Quong Money Flow Index (MFI) indicator.
    MONEY_FLOW_INDEX = 15

    # ── george lane ────────────────────────────────────────────────────────

    # Identifies the George Lane Stochastic Oscillator (STOCH) indicator.
    STOCHASTIC = 16

    # ── gerald appel ───────────────────────────────────────────────────────

    # Identifies Gerald Appel's Moving Average Convergence Divergence (MACD) indicator.
    MOVING_AVERAGE_CONVERGENCE_DIVERGENCE = 17

    # Identifies the Gerald Appel Percentage Price Oscillator (PPO) indicator.
    PERCENTAGE_PRICE_OSCILLATOR = 18

    # ── igor livshin ───────────────────────────────────────────────────────

    # Identifies the Igor Livshin Balance of Power (BOP) indicator.
    BALANCE_OF_POWER = 19

    # ── jack hutson ────────────────────────────────────────────────────────

    # Identifies Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX) indicator.
    TRIPLE_EXPONENTIAL_MOVING_AVERAGE_OSCILLATOR = 20

    # ── john bollinger ─────────────────────────────────────────────────────

    # Identifies the Bollinger Bands (BB) indicator.
    BOLLINGER_BANDS = 21

    # Identifies John Bollinger's Bollinger Bands Trend (BBTrend) indicator.
    BOLLINGER_BANDS_TREND = 22

    # ── john ehlers ────────────────────────────────────────────────────────

    # Identifies the Ehlers Autocorrelation Indicator (ACI) heatmap, a bank of Pearson
    # correlation coefficients between the current filtered series and a lagged copy
    # of itself, following EasyLanguage listing 8-2.
    AUTO_CORRELATION_INDICATOR = 23

    # Identifies the Ehlers Autocorrelation Periodogram (ACP) heatmap, a discrete Fourier
    # transform of the autocorrelation function over a configurable cycle-period range,
    # following EasyLanguage listing 8-3.
    AUTO_CORRELATION_PERIODOGRAM = 24

    # Identifies the Ehlers Center of Gravity Oscillator (COG) indicator.
    CENTER_OF_GRAVITY_OSCILLATOR = 25

    # Identifies the Ehlers Comb Band-Pass Spectrum (CBPS) heatmap indicator, a bank of
    # 2-pole band-pass filters (one per cycle period) fed by a Butterworth highpass +
    # Super Smoother pre-filter cascade, following EasyLanguage listing 10-1.
    COMB_BAND_PASS_SPECTRUM = 26

    # Identifies the Ehlers Corona Signal To Noise Ratio (CSNR) heatmap indicator, exposing
    # the intensity raster heatmap column and the current SNR mapped into the parameter range.
    CORONA_SIGNAL_TO_NOISE_RATIO = 27

    # Identifies the Ehlers Corona Spectrum (CSPECT) heatmap indicator, exposing the dB
    # heatmap column, the weighted dominant cycle estimate and its 5-sample median.
    CORONA_SPECTRUM = 28

    # Identifies the Ehlers Corona Swing Position (CSWING) heatmap indicator, exposing the
    # intensity raster heatmap column and the current swing position mapped into the parameter range.
    CORONA_SWING_POSITION = 29

    # Identifies the Ehlers Corona Trend Vigor (CTV) heatmap indicator, exposing the intensity
    # raster heatmap column and the current trend vigor scaled into the parameter range.
    CORONA_TREND_VIGOR = 30

    # Identifies the Ehlers Cyber Cycle (CC) indicator.
    CYBER_CYCLE = 31

    # Identifies the Ehlers Discrete Fourier Transform Spectrum (DFTPS) heatmap indicator,
    # a mean-subtracted DFT power spectrum over a configurable cycle-period range.
    DISCRETE_FOURIER_TRANSFORM_SPECTRUM = 32

    # Identifies the Ehlers Dominant Cycle (DC) indicator, exposing raw period, smoothed period and phase.
    DOMINANT_CYCLE = 33

    # Identifies the Ehlers Fractal Adaptive Moving Average (FRAMA) indicator.
    FRACTAL_ADAPTIVE_MOVING_AVERAGE = 34

    # Identifies the Ehlers Hilbert Transformer Instantaneous Trend Line (HTITL) indicator,
    # exposing trend value and dominant cycle period.
    HILBERT_TRANSFORMER_INSTANTANEOUS_TREND_LINE = 35

    # Identifies the Ehlers Instantaneous Trend Line (iTrend) indicator.
    INSTANTANEOUS_TREND_LINE = 36

    # Identifies the Ehlers MESA Adaptive Moving Average (MAMA) indicator.
    MESA_ADAPTIVE_MOVING_AVERAGE = 37

    # Identifies the Ehlers Roofing Filter indicator.
    ROOFING_FILTER = 38

    # Identifies the Ehlers Sine Wave (SW) indicator, exposing sine value, lead sine,
    # band, dominant cycle period and phase.
    SINE_WAVE = 39

    # Identifies the Ehlers Super Smoother (SS) indicator.
    SUPER_SMOOTHER = 40

    # Identifies the Ehlers Trend / Cycle Mode (TCM) indicator, exposing the trend/cycle
    # value (+1 in trend, -1 in cycle), trend/cycle mode flags, instantaneous trend line,
    # sine wave, lead sine wave, dominant cycle period and phase.
    TREND_CYCLE_MODE = 41

    # Identifies the Ehlers Zero-lag Error-Correcting Exponential Moving Average (ZECEMA) indicator.
    ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE = 42

    # Identifies the Ehlers Zero-lag Exponential Moving Average (ZEMA) indicator.
    ZERO_LAG_EXPONENTIAL_MOVING_AVERAGE = 43

    # ── joseph granville ───────────────────────────────────────────────────

    # Identifies the Joseph Granville On-Balance Volume (OBV) indicator.
    ON_BALANCE_VOLUME = 44

    # ── larry williams ─────────────────────────────────────────────────────

    # Identifies the Larry Williams Ultimate Oscillator (ULTOSC) indicator.
    ULTIMATE_OSCILLATOR = 45

    # Identifies the Larry Williams Williams %R (WILL%R) indicator.
    WILLIAMS_PERCENT_R = 46

    # ── manfred durschner ──────────────────────────────────────────────────

    # Identifies the New Moving Average (NMA) indicator by Durschner.
    NEW_MOVING_AVERAGE = 47

    # ── marc chaikin ───────────────────────────────────────────────────────

    # Identifies the Marc Chaikin Advance-Decline (A/D) indicator.
    ADVANCE_DECLINE = 48

    # Identifies the Marc Chaikin Advance-Decline Oscillator (ADOSC) indicator.
    ADVANCE_DECLINE_OSCILLATOR = 49

    # ── mark jurik ─────────────────────────────────────────────────────────

    # Identifies the Jurik Adaptive Relative Trend Strength Index (JARSX) indicator.
    JURIK_ADAPTIVE_RELATIVE_TREND_STRENGTH_INDEX = 50

    # Identifies the Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.
    JURIK_ADAPTIVE_ZERO_LAG_VELOCITY = 51

    # Identifies the Jurik Commodity Channel Index (JCCX) indicator.
    JURIK_COMMODITY_CHANNEL_INDEX = 52

    # Identifies the Jurik Composite Fractal Behavior Index (CFB) indicator.
    JURIK_COMPOSITE_FRACTAL_BEHAVIOR_INDEX = 53

    # Identifies the Jurik Directional Movement Index (DMX) indicator.
    JURIK_DIRECTIONAL_MOVEMENT_INDEX = 54

    # Identifies the Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) indicator.
    JURIK_FRACTAL_ADAPTIVE_ZERO_LAG_VELOCITY = 55

    # Identifies the Jurik Moving Average (JMA) indicator.
    JURIK_MOVING_AVERAGE = 56

    # Identifies the Jurik Relative Trend Strength Index (RSX) indicator.
    JURIK_RELATIVE_TREND_STRENGTH_INDEX = 57

    # Identifies the Jurik Turning Point Oscillator (JTPO) indicator.
    JURIK_TURNING_POINT_OSCILLATOR = 58

    # Identifies the Jurik Wavelet Sampler (WAV) indicator.
    JURIK_WAVELET_SAMPLER = 59

    # Identifies the Jurik Zero Lag Velocity (VEL) indicator.
    JURIK_ZERO_LAG_VELOCITY = 60

    # ── patrick mulloy ─────────────────────────────────────────────────────

    # Identifies the Double Exponential Moving Average (DEMA) indicator.
    DOUBLE_EXPONENTIAL_MOVING_AVERAGE = 61

    # Identifies the Triple Exponential Moving Average (TEMA) indicator.
    TRIPLE_EXPONENTIAL_MOVING_AVERAGE = 62

    # ── perry kaufman ──────────────────────────────────────────────────────

    # Identifies the Kaufman Adaptive Moving Average (KAMA) indicator.
    KAUFMAN_ADAPTIVE_MOVING_AVERAGE = 63

    # ── tim tillson ────────────────────────────────────────────────────────

    # Identifies the T2 Exponential Moving Average (T2EMA) indicator.
    T2_EXPONENTIAL_MOVING_AVERAGE = 64

    # Identifies the T3 Exponential Moving Average (T3EMA) indicator.
    T3_EXPONENTIAL_MOVING_AVERAGE = 65

    # ── tushar chande ──────────────────────────────────────────────────────

    # Identifies the Tushar Chande Aroon (AROON) indicator.
    AROON = 66

    # Identifies the Chande Momentum Oscillator (CMO) indicator.
    CHANDE_MOMENTUM_OSCILLATOR = 67

    # Identifies the Tushar Chande Stochastic RSI (STOCHRSI) indicator.
    STOCHASTIC_RELATIVE_STRENGTH_INDEX = 68

    # ── vladimir kravchuk ──────────────────────────────────────────────────

    # Identifies Vladimir Kravchuk's Adaptive Trend and Cycle Filter (ATCF) suite: a bank
    # of five FIR filters (FATL, SATL, RFTL, RSTL, RBCI) plus three composites (FTLM, STLM, PCCI).
    ADAPTIVE_TREND_AND_CYCLE_FILTER = 69

    # ── welles wilder ──────────────────────────────────────────────────────

    # Identifies the Welles Wilder Average Directional Movement Index (ADX) indicator.
    AVERAGE_DIRECTIONAL_MOVEMENT_INDEX = 70

    # Identifies the Welles Wilder Average Directional Movement Index Rating (ADXR) indicator.
    AVERAGE_DIRECTIONAL_MOVEMENT_INDEX_RATING = 71

    # Identifies the Welles Wilder Average True Range (ATR) indicator.
    AVERAGE_TRUE_RANGE = 72

    # Identifies the Welles Wilder Directional Indicator Minus (-DI) indicator.
    DIRECTIONAL_INDICATOR_MINUS = 73

    # Identifies the Welles Wilder Directional Indicator Plus (+DI) indicator.
    DIRECTIONAL_INDICATOR_PLUS = 74

    # Identifies the Welles Wilder Directional Movement Index (DX) indicator.
    DIRECTIONAL_MOVEMENT_INDEX = 75

    # Identifies the Welles Wilder Directional Movement Minus (-DM) indicator.
    DIRECTIONAL_MOVEMENT_MINUS = 76

    # Identifies the Welles Wilder Directional Movement Plus (+DM) indicator.
    DIRECTIONAL_MOVEMENT_PLUS = 77

    # Identifies the Welles Wilder Normalized Average True Range (NATR) indicator.
    NORMALIZED_AVERAGE_TRUE_RANGE = 78

    # Identifies the Welles Wilder Parabolic Stop And Reverse (SAR) indicator.
    PARABOLIC_STOP_AND_REVERSE = 79

    # Identifies the Relative Strength Index (RSI) indicator.
    RELATIVE_STRENGTH_INDEX = 80

    # Identifies the Welles Wilder True Range (TR) indicator.
    TRUE_RANGE = 81

    # ── custom ────────────────────────────────────────────────────────────

    # Identifies the Goertzel power spectrum (GOERTZEL) indicator.
    GOERTZEL_SPECTRUM = 82

    # Identifies the Maximum Entropy Spectrum (MESPECT) heatmap indicator, a Burg
    # maximum-entropy auto-regressive power spectrum over a configurable cycle-period range.
    MAXIMUM_ENTROPY_SPECTRUM = 83
