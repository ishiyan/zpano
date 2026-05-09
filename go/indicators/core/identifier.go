package core

import (
	"bytes"
	"fmt"
)

// Identifier identifies an indicator by enumerating all implemented indicators.
type Identifier int

const (
	// ── common ────────────────────────────────────────────────────────────

	// AbsolutePriceOscillator identifies the Absolute Price Oscillator (APO) indicator.
	AbsolutePriceOscillator Identifier = iota + 1

	// ExponentialMovingAverage identifies the Exponential Moving Average (EMA) indicator.
	ExponentialMovingAverage

	// LinearRegression identifies the Linear Regression (LINEARREG) indicator.
	LinearRegression

	// Momentum identifies the momentum (MOM) indicator.
	Momentum

	// PearsonsCorrelationCoefficient identifies the Pearson's Correlation Coefficient (CORREL) indicator.
	PearsonsCorrelationCoefficient

	// RateOfChange identifies the Rate of Change (ROC) indicator.
	RateOfChange

	// RateOfChangePercent identifies the Rate of Change Percent (ROCP) indicator.
	RateOfChangePercent

	// RateOfChangeRatio identifies the Rate of Change Ratio (ROCR / ROCR100) indicator.
	RateOfChangeRatio

	// SimpleMovingAverage identifies the Simple Moving Average (SMA) indicator.
	SimpleMovingAverage

	// StandardDeviation identifies the Standard Deviation (STDEV) indicator.
	StandardDeviation

	// TriangularMovingAverage identifies the Triangular Moving Average (TRIMA) indicator.
	TriangularMovingAverage

	// Variance identifies the Variance (VAR) indicator.
	Variance

	// WeightedMovingAverage identifies the Weighted Moving Average (WMA) indicator.
	WeightedMovingAverage

	// ── arnaud legoux ──────────────────────────────────────────────────────

	// ArnaudLegouxMovingAverage identifies the Arnaud Legoux Moving Average (ALMA) indicator.
	ArnaudLegouxMovingAverage

	// ── donald lambert ─────────────────────────────────────────────────────

	// CommodityChannelIndex identifies the Donald Lambert Commodity Channel Index (CCI) indicator.
	CommodityChannelIndex

	// ── gene quong ─────────────────────────────────────────────────────────

	// MoneyFlowIndex identifies the Gene Quong Money Flow Index (MFI) indicator.
	MoneyFlowIndex

	// ── george lane ────────────────────────────────────────────────────────

	// Stochastic identifies the George Lane Stochastic Oscillator (STOCH) indicator.
	Stochastic

	// ── gerald appel ───────────────────────────────────────────────────────

	// MovingAverageConvergenceDivergence identifies Gerald Appel's Moving Average Convergence Divergence (MACD) indicator.
	MovingAverageConvergenceDivergence

	// PercentagePriceOscillator identifies the Gerald Appel Percentage Price Oscillator (PPO) indicator.
	PercentagePriceOscillator

	// ── igor livshin ───────────────────────────────────────────────────────

	// BalanceOfPower identifies the Igor Livshin Balance of Power (BOP) indicator.
	BalanceOfPower

	// ── jack hutson ────────────────────────────────────────────────────────

	// TripleExponentialMovingAverageOscillator identifies Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX) indicator.
	TripleExponentialMovingAverageOscillator

	// ── john bollinger ─────────────────────────────────────────────────────

	// BollingerBands identifies the Bollinger Bands (BB) indicator.
	BollingerBands

	// BollingerBandsTrend identifies John Bollinger's Bollinger Bands Trend (BBTrend) indicator.
	BollingerBandsTrend

	// ── john ehlers ────────────────────────────────────────────────────────

	// AutoCorrelationIndicator identifies the Autocorrelation Indicator (aci)
	// heat-map of Pearson correlation coefficients between the current filtered
	// series and a lagged copy of itself, following Ehlers' EasyLanguage listing 8-2.
	AutoCorrelationIndicator

	// AutoCorrelationPeriodogram identifies the Autocorrelation Periodogram (acp)
	// heat-map of cyclic activity estimated via a discrete Fourier transform of the
	// autocorrelation function, following Ehlers' EasyLanguage listing 8-3.
	AutoCorrelationPeriodogram

	// CenterOfGravityOscillator identifies the Ehlers Center of Gravity (COG) oscillator indicator.
	CenterOfGravityOscillator

	// CombBandPassSpectrum identifies the Comb Band-Pass Spectrum (cbps) indicator,
	// a heat-map of cyclic activity estimated via a bank of 2-pole band-pass filters,
	// one per integer cycle period, following Ehlers' EasyLanguage listing 10-1.
	CombBandPassSpectrum

	// CoronaSignalToNoiseRatio identifies the Ehlers Corona Signal-to-Noise Ratio (CSNR) indicator, a heat-map
	// of SNR plus a smoothed SNR scalar line.
	CoronaSignalToNoiseRatio

	// CoronaSpectrum identifies the Ehlers Corona Spectrum (CSPECT) indicator, a heat-map of cyclic activity
	// over a cycle-period range together with the dominant cycle period and its 5-sample median.
	CoronaSpectrum

	// CoronaSwingPosition identifies the Ehlers Corona Swing Position (CSWING) indicator, a heat-map of swing
	// position with a scalar swing-position line.
	CoronaSwingPosition

	// CoronaTrendVigor identifies the Ehlers Corona Trend Vigor (CTV) indicator, a heat-map of trend vigor
	// with a scalar trend-vigor line.
	CoronaTrendVigor

	// CyberCycle identifies the Ehlers Cyber Cycle (CC) indicator.
	CyberCycle

	// DiscreteFourierTransformSpectrum identifies the Discrete Fourier Transform Spectrum
	// (psDft) indicator, a heat-map of cyclic activity estimated via a discrete Fourier
	// transform over a sliding window.
	DiscreteFourierTransformSpectrum

	// DominantCycle identifies the Ehlers Dominant Cycle (DC) indicator, exposing raw period, smoothed period and phase.
	DominantCycle

	// FractalAdaptiveMovingAverage identifies the Ehlers Fractal Adaptive Moving Average (FRAMA) indicator.
	FractalAdaptiveMovingAverage

	// HilbertTransformerInstantaneousTrendLine identifies the Ehlers Hilbert Transformer Instantaneous Trend Line (HTITL) indicator,
	// exposing the trend line value and the smoothed dominant cycle period.
	HilbertTransformerInstantaneousTrendLine

	// InstantaneousTrendLine identifies the Ehlers Instantaneous Trend Line (iTrend) indicator.
	InstantaneousTrendLine

	// MesaAdaptiveMovingAverage identifies the Ehlers MESA Adaptive Moving Average (MAMA) indicator.
	MesaAdaptiveMovingAverage

	// RoofingFilter identifies the Ehlers Roofing Filter indicator.
	RoofingFilter

	// SineWave identifies the Ehlers Sine Wave (SW) indicator, exposing sine value, lead sine, band, dominant cycle period and phase.
	SineWave

	// SuperSmoother identifies the Ehlers Super Smoother (SS) indicator.
	SuperSmoother

	// TrendCycleMode identifies the Ehlers Trend / Cycle Mode (TCM) indicator, exposing the trend/cycle value
	// (+1 in trend, −1 in cycle), trend/cycle mode flags, instantaneous trend line, sine wave, lead sine wave,
	// dominant cycle period and phase.
	TrendCycleMode

	// ZeroLagErrorCorrectingExponentialMovingAverage identifies the Ehlers Zero-lag Error-Correcting Exponential Moving Average (ZECEMA) indicator.
	ZeroLagErrorCorrectingExponentialMovingAverage

	// ZeroLagExponentialMovingAverage identifies the Ehlers Zero-lag Exponential Moving Average (ZEMA) indicator.
	ZeroLagExponentialMovingAverage

	// ── joseph granville ───────────────────────────────────────────────────

	// OnBalanceVolume identifies the Joseph Granville On-Balance Volume (OBV) indicator.
	OnBalanceVolume

	// ── larry williams ─────────────────────────────────────────────────────

	// UltimateOscillator identifies the Larry Williams Ultimate Oscillator (ULTOSC) indicator.
	UltimateOscillator

	// WilliamsPercentR identifies the Larry Williams Williams %R (WILL%R) indicator.
	WilliamsPercentR

	// ── manfred durschner ──────────────────────────────────────────────────

	// NewMovingAverage identifies the New Moving Average (NMA) indicator by Dürschner.
	NewMovingAverage

	// ── marc chaikin ───────────────────────────────────────────────────────

	// AdvanceDecline identifies the Marc Chaikin Advance-Decline (AD) indicator.
	AdvanceDecline

	// AdvanceDeclineOscillator identifies the Marc Chaikin Advance-Decline Oscillator (ADOSC) indicator.
	AdvanceDeclineOscillator

	// ── mark jurik ─────────────────────────────────────────────────────────

	// JurikAdaptiveRelativeTrendStrengthIndex identifies the Jurik Adaptive Relative Trend Strength Index (JARSX) indicator.
	JurikAdaptiveRelativeTrendStrengthIndex

	// JurikAdaptiveZeroLagVelocity identifies the Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.
	JurikAdaptiveZeroLagVelocity

	// JurikCommodityChannelIndex identifies the Jurik Commodity Channel Index (JCCX) indicator.
	JurikCommodityChannelIndex

	// JurikCompositeFractalBehaviorIndex identifies the Jurik Composite Fractal Behavior Index (CFB) indicator.
	JurikCompositeFractalBehaviorIndex

	// JurikDirectionalMovementIndex identifies the Jurik Directional Movement Index (DMX) indicator.
	JurikDirectionalMovementIndex

	// JurikFractalAdaptiveZeroLagVelocity identifies the Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) indicator.
	JurikFractalAdaptiveZeroLagVelocity

	// JurikMovingAverage identifies the Jurik Moving Average (JMA) indicator.
	JurikMovingAverage

	// JurikRelativeTrendStrengthIndex identifies the Jurik Relative Trend Strength Index (RSX) indicator.
	JurikRelativeTrendStrengthIndex

	// JurikTurningPointOscillator identifies the Jurik Turning Point Oscillator (JTPO) indicator.
	JurikTurningPointOscillator

	// JurikWaveletSampler identifies the Jurik Wavelet Sampler (WAV) indicator.
	JurikWaveletSampler

	// JurikZeroLagVelocity identifies the Jurik Zero Lag Velocity (VEL) indicator.
	JurikZeroLagVelocity

	// ── patrick mulloy ─────────────────────────────────────────────────────

	// DoubleExponentialMovingAverage identifies the Double Exponential Moving Average (DEMA) indicator.
	DoubleExponentialMovingAverage

	// TripleExponentialMovingAverage identifies the Triple Exponential Moving Average (TEMA) indicator.
	TripleExponentialMovingAverage

	// ── perry kaufman ──────────────────────────────────────────────────────

	// KaufmanAdaptiveMovingAverage identifies the Kaufman Adaptive Moving Average (KAMA) indicator.
	KaufmanAdaptiveMovingAverage

	// ── tim tillson ────────────────────────────────────────────────────────

	// T2ExponentialMovingAverage identifies the T2 Exponential Moving Average (T2) indicator.
	T2ExponentialMovingAverage

	// T3ExponentialMovingAverage identifies the T3 Exponential Moving Average (T3) indicator.
	T3ExponentialMovingAverage

	// ── tushar chande ──────────────────────────────────────────────────────

	// Aroon identifies the Tushar Chande Aroon (AROON) indicator.
	Aroon

	// ChandeMomentumOscillator identifies the Chande Momentum Oscillator (CMO) indicator.
	ChandeMomentumOscillator

	// StochasticRelativeStrengthIndex identifies the Tushar Chande Stochastic RSI (STOCHRSI) indicator.
	StochasticRelativeStrengthIndex

	// ── vladimir kravchuk ──────────────────────────────────────────────────

	// AdaptiveTrendAndCycleFilter identifies the Vladimir Kravchuk Adaptive Trend & Cycle Filter (ATCF)
	// suite, exposing FATL, SATL, RFTL, RSTL, RBCI FIR-filter outputs together with the derived
	// FTLM, STLM, and PCCI composites.
	AdaptiveTrendAndCycleFilter

	// ── welles wilder ──────────────────────────────────────────────────────

	// AverageDirectionalMovementIndex identifies the Welles Wilder Average Directional Movement Index (ADX) indicator.
	AverageDirectionalMovementIndex

	// AverageDirectionalMovementIndexRating identifies the Welles Wilder Average Directional Movement Index Rating (ADXR) indicator.
	AverageDirectionalMovementIndexRating

	// AverageTrueRange identifies the Welles Wilder Average True Range (ATR) indicator.
	AverageTrueRange

	// DirectionalIndicatorMinus identifies the Welles Wilder Directional Indicator Minus (-DI) indicator.
	DirectionalIndicatorMinus

	// DirectionalIndicatorPlus identifies the Welles Wilder Directional Indicator Plus (+DI) indicator.
	DirectionalIndicatorPlus

	// DirectionalMovementIndex identifies the Welles Wilder Directional Movement Index (DX) indicator.
	DirectionalMovementIndex

	// DirectionalMovementMinus identifies the Welles Wilder Directional Movement Minus (-DM) indicator.
	DirectionalMovementMinus

	// DirectionalMovementPlus identifies the Welles Wilder Directional Movement Plus (+DM) indicator.
	DirectionalMovementPlus

	// NormalizedAverageTrueRange identifies the Welles Wilder Normalized Average True Range (NATR) indicator.
	NormalizedAverageTrueRange

	// ParabolicStopAndReverse identifies the Welles Wilder Parabolic Stop And Reverse (SAR) indicator.
	ParabolicStopAndReverse

	// RelativeStrengthIndex identifies the Relative Strength Index (RSI) indicator.
	RelativeStrengthIndex

	// TrueRange identifies the Welles Wilder True Range (TR) indicator.
	TrueRange

	// ── custom ────────────────────────────────────────────────────────────

	// GoertzelSpectrum identifies the Goertzel power spectrum (GOERTZEL) indicator.
	GoertzelSpectrum

	// MaximumEntropySpectrum identifies the Maximum Entropy Spectrum (MESPECT) indicator, a
	// heat-map of cyclic activity estimated via Burg's maximum-entropy auto-regressive method.
	MaximumEntropySpectrum

	last
)

const (
	unknown = "unknown"

	// ── common ────────────────────────────────────────────────────────────
	absolutePriceOscillator        = "absolutePriceOscillator"
	exponentialMovingAverage       = "exponentialMovingAverage"
	linearRegression               = "linearRegression"
	momentum                       = "momentum"
	pearsonsCorrelationCoefficient = "pearsonsCorrelationCoefficient"
	rateOfChange                   = "rateOfChange"
	rateOfChangePercent            = "rateOfChangePercent"
	rateOfChangeRatio              = "rateOfChangeRatio"
	simpleMovingAverage            = "simpleMovingAverage"
	standardDeviation              = "standardDeviation"
	triangularMovingAverage        = "triangularMovingAverage"
	variance                       = "variance"
	weightedMovingAverage          = "weightedMovingAverage"

	// ── arnaud legoux ──────────────────────────────────────────────────────
	arnaudLegouxMovingAverage = "arnaudLegouxMovingAverage"

	// ── donald lambert ─────────────────────────────────────────────────────
	commodityChannelIndex = "commodityChannelIndex"

	// ── gene quong ─────────────────────────────────────────────────────────
	moneyFlowIndex = "moneyFlowIndex"

	// ── george lane ────────────────────────────────────────────────────────
	stochastic = "stochastic"

	// ── gerald appel ───────────────────────────────────────────────────────
	movingAverageConvergenceDivergence = "movingAverageConvergenceDivergence"
	percentagePriceOscillator          = "percentagePriceOscillator"

	// ── igor livshin ───────────────────────────────────────────────────────
	balanceOfPower = "balanceOfPower"

	// ── jack hutson ────────────────────────────────────────────────────────
	tripleExponentialMovingAverageOscillator = "tripleExponentialMovingAverageOscillator"

	// ── john bollinger ─────────────────────────────────────────────────────
	bollingerBands      = "bollingerBands"
	bollingerBandsTrend = "bollingerBandsTrend"

	// ── john ehlers ────────────────────────────────────────────────────────
	autoCorrelationIndicator                       = "autoCorrelationIndicator"
	autoCorrelationPeriodogram                     = "autoCorrelationPeriodogram"
	centerOfGravityOscillator                      = "centerOfGravityOscillator"
	combBandPassSpectrum                           = "combBandPassSpectrum"
	coronaSignalToNoiseRatio                       = "coronaSignalToNoiseRatio"
	coronaSpectrum                                 = "coronaSpectrum"
	coronaSwingPosition                            = "coronaSwingPosition"
	coronaTrendVigor                               = "coronaTrendVigor"
	cyberCycle                                     = "cyberCycle"
	discreteFourierTransformSpectrum               = "discreteFourierTransformSpectrum"
	dominantCycle                                  = "dominantCycle"
	fractalAdaptiveMovingAverage                   = "fractalAdaptiveMovingAverage"
	hilbertTransformerInstantaneousTrendLine       = "hilbertTransformerInstantaneousTrendLine"
	instantaneousTrendLine                         = "instantaneousTrendLine"
	mesaAdaptiveMovingAverage                      = "mesaAdaptiveMovingAverage"
	roofingFilter                                  = "roofingFilter"
	sineWave                                       = "sineWave"
	superSmoother                                  = "superSmoother"
	trendCycleMode                                 = "trendCycleMode"
	zeroLagErrorCorrectingExponentialMovingAverage = "zeroLagErrorCorrectingExponentialMovingAverage"
	zeroLagExponentialMovingAverage                = "zeroLagExponentialMovingAverage"

	// ── joseph granville ───────────────────────────────────────────────────
	onBalanceVolume = "onBalanceVolume"

	// ── larry williams ─────────────────────────────────────────────────────
	ultimateOscillator = "ultimateOscillator"
	williamsPercentR   = "williamsPercentR"

	// ── manfred durschner ──────────────────────────────────────────────────
	newMovingAverage = "newMovingAverage"

	// ── marc chaikin ───────────────────────────────────────────────────────
	advanceDecline           = "advanceDecline"
	advanceDeclineOscillator = "advanceDeclineOscillator"

	// ── mark jurik ─────────────────────────────────────────────────────────
	jurikAdaptiveRelativeTrendStrengthIndex = "jurikAdaptiveRelativeTrendStrengthIndex"
	jurikAdaptiveZeroLagVelocity            = "jurikAdaptiveZeroLagVelocity"
	jurikCommodityChannelIndex              = "jurikCommodityChannelIndex"
	jurikCompositeFractalBehaviorIndex      = "jurikCompositeFractalBehaviorIndex"
	jurikDirectionalMovementIndex           = "jurikDirectionalMovementIndex"
	jurikFractalAdaptiveZeroLagVelocity     = "jurikFractalAdaptiveZeroLagVelocity"
	jurikMovingAverage                      = "jurikMovingAverage"
	jurikRelativeTrendStrengthIndex         = "jurikRelativeTrendStrengthIndex"
	jurikTurningPointOscillator             = "jurikTurningPointOscillator"
	jurikWaveletSampler                     = "jurikWaveletSampler"
	jurikZeroLagVelocity                    = "jurikZeroLagVelocity"

	// ── patrick mulloy ─────────────────────────────────────────────────────
	doubleExponentialMovingAverage = "doubleExponentialMovingAverage"
	tripleExponentialMovingAverage = "tripleExponentialMovingAverage"

	// ── perry kaufman ──────────────────────────────────────────────────────
	kaufmanAdaptiveMovingAverage = "kaufmanAdaptiveMovingAverage"

	// ── tim tillson ────────────────────────────────────────────────────────
	t2ExponentialMovingAverage = "t2ExponentialMovingAverage"
	t3ExponentialMovingAverage = "t3ExponentialMovingAverage"

	// ── tushar chande ──────────────────────────────────────────────────────
	aroon                           = "aroon"
	chandeMomentumOscillator        = "chandeMomentumOscillator"
	stochasticRelativeStrengthIndex = "stochasticRelativeStrengthIndex"

	// ── vladimir kravchuk ──────────────────────────────────────────────────
	adaptiveTrendAndCycleFilter = "adaptiveTrendAndCycleFilter"

	// ── welles wilder ──────────────────────────────────────────────────────
	averageDirectionalMovementIndex       = "averageDirectionalMovementIndex"
	averageDirectionalMovementIndexRating = "averageDirectionalMovementIndexRating"
	averageTrueRange                      = "averageTrueRange"
	directionalIndicatorMinus             = "directionalIndicatorMinus"
	directionalIndicatorPlus              = "directionalIndicatorPlus"
	directionalMovementIndex              = "directionalMovementIndex"
	directionalMovementMinus              = "directionalMovementMinus"
	directionalMovementPlus               = "directionalMovementPlus"
	normalizedAverageTrueRange            = "normalizedAverageTrueRange"
	parabolicStopAndReverse               = "parabolicStopAndReverse"
	relativeStrengthIndex                 = "relativeStrengthIndex"
	trueRange                             = "trueRange"

	// ── custom ────────────────────────────────────────────────────────────
	goertzelSpectrum       = "goertzelSpectrum"
	maximumEntropySpectrum = "maximumEntropySpectrum"
)

// String implements the Stringer interface.
//
//nolint:exhaustive,cyclop,funlen
func (i Identifier) String() string {
	switch i {
	// ── common ────────────────────────────────────────────────────────────
	case AbsolutePriceOscillator:
		return absolutePriceOscillator
	case ExponentialMovingAverage:
		return exponentialMovingAverage
	case LinearRegression:
		return linearRegression
	case Momentum:
		return momentum
	case PearsonsCorrelationCoefficient:
		return pearsonsCorrelationCoefficient
	case RateOfChange:
		return rateOfChange
	case RateOfChangePercent:
		return rateOfChangePercent
	case RateOfChangeRatio:
		return rateOfChangeRatio
	case SimpleMovingAverage:
		return simpleMovingAverage
	case StandardDeviation:
		return standardDeviation
	case TriangularMovingAverage:
		return triangularMovingAverage
	case Variance:
		return variance
	case WeightedMovingAverage:
		return weightedMovingAverage
	// ── arnaud legoux ──────────────────────────────────────────────────────
	case ArnaudLegouxMovingAverage:
		return arnaudLegouxMovingAverage
	// ── donald lambert ─────────────────────────────────────────────────────
	case CommodityChannelIndex:
		return commodityChannelIndex
	// ── gene quong ─────────────────────────────────────────────────────────
	case MoneyFlowIndex:
		return moneyFlowIndex
	// ── george lane ────────────────────────────────────────────────────────
	case Stochastic:
		return stochastic
	// ── gerald appel ───────────────────────────────────────────────────────
	case MovingAverageConvergenceDivergence:
		return movingAverageConvergenceDivergence
	case PercentagePriceOscillator:
		return percentagePriceOscillator
	// ── igor livshin ───────────────────────────────────────────────────────
	case BalanceOfPower:
		return balanceOfPower
	// ── jack hutson ────────────────────────────────────────────────────────
	case TripleExponentialMovingAverageOscillator:
		return tripleExponentialMovingAverageOscillator
	// ── john bollinger ─────────────────────────────────────────────────────
	case BollingerBands:
		return bollingerBands
	case BollingerBandsTrend:
		return bollingerBandsTrend
	// ── john ehlers ────────────────────────────────────────────────────────
	case AutoCorrelationIndicator:
		return autoCorrelationIndicator
	case AutoCorrelationPeriodogram:
		return autoCorrelationPeriodogram
	case CenterOfGravityOscillator:
		return centerOfGravityOscillator
	case CombBandPassSpectrum:
		return combBandPassSpectrum
	case CoronaSignalToNoiseRatio:
		return coronaSignalToNoiseRatio
	case CoronaSpectrum:
		return coronaSpectrum
	case CoronaSwingPosition:
		return coronaSwingPosition
	case CoronaTrendVigor:
		return coronaTrendVigor
	case CyberCycle:
		return cyberCycle
	case DiscreteFourierTransformSpectrum:
		return discreteFourierTransformSpectrum
	case DominantCycle:
		return dominantCycle
	case FractalAdaptiveMovingAverage:
		return fractalAdaptiveMovingAverage
	case HilbertTransformerInstantaneousTrendLine:
		return hilbertTransformerInstantaneousTrendLine
	case InstantaneousTrendLine:
		return instantaneousTrendLine
	case MesaAdaptiveMovingAverage:
		return mesaAdaptiveMovingAverage
	case RoofingFilter:
		return roofingFilter
	case SineWave:
		return sineWave
	case SuperSmoother:
		return superSmoother
	case TrendCycleMode:
		return trendCycleMode
	case ZeroLagErrorCorrectingExponentialMovingAverage:
		return zeroLagErrorCorrectingExponentialMovingAverage
	case ZeroLagExponentialMovingAverage:
		return zeroLagExponentialMovingAverage
	// ── joseph granville ───────────────────────────────────────────────────
	case OnBalanceVolume:
		return onBalanceVolume
	// ── larry williams ─────────────────────────────────────────────────────
	case UltimateOscillator:
		return ultimateOscillator
	case WilliamsPercentR:
		return williamsPercentR
	// ── manfred durschner ──────────────────────────────────────────────────
	case NewMovingAverage:
		return newMovingAverage
	// ── marc chaikin ───────────────────────────────────────────────────────
	case AdvanceDecline:
		return advanceDecline
	case AdvanceDeclineOscillator:
		return advanceDeclineOscillator
	// ── mark jurik ─────────────────────────────────────────────────────────
	case JurikAdaptiveRelativeTrendStrengthIndex:
		return jurikAdaptiveRelativeTrendStrengthIndex
	case JurikAdaptiveZeroLagVelocity:
		return jurikAdaptiveZeroLagVelocity
	case JurikCommodityChannelIndex:
		return jurikCommodityChannelIndex
	case JurikCompositeFractalBehaviorIndex:
		return jurikCompositeFractalBehaviorIndex
	case JurikDirectionalMovementIndex:
		return jurikDirectionalMovementIndex
	case JurikFractalAdaptiveZeroLagVelocity:
		return jurikFractalAdaptiveZeroLagVelocity
	case JurikMovingAverage:
		return jurikMovingAverage
	case JurikRelativeTrendStrengthIndex:
		return jurikRelativeTrendStrengthIndex
	case JurikTurningPointOscillator:
		return jurikTurningPointOscillator
	case JurikWaveletSampler:
		return jurikWaveletSampler
	case JurikZeroLagVelocity:
		return jurikZeroLagVelocity
	// ── patrick mulloy ─────────────────────────────────────────────────────
	case DoubleExponentialMovingAverage:
		return doubleExponentialMovingAverage
	case TripleExponentialMovingAverage:
		return tripleExponentialMovingAverage
	// ── perry kaufman ──────────────────────────────────────────────────────
	case KaufmanAdaptiveMovingAverage:
		return kaufmanAdaptiveMovingAverage
	// ── tim tillson ────────────────────────────────────────────────────────
	case T2ExponentialMovingAverage:
		return t2ExponentialMovingAverage
	case T3ExponentialMovingAverage:
		return t3ExponentialMovingAverage
	// ── tushar chande ──────────────────────────────────────────────────────
	case Aroon:
		return aroon
	case ChandeMomentumOscillator:
		return chandeMomentumOscillator
	case StochasticRelativeStrengthIndex:
		return stochasticRelativeStrengthIndex
	// ── vladimir kravchuk ──────────────────────────────────────────────────
	case AdaptiveTrendAndCycleFilter:
		return adaptiveTrendAndCycleFilter
	// ── welles wilder ──────────────────────────────────────────────────────
	case AverageDirectionalMovementIndex:
		return averageDirectionalMovementIndex
	case AverageDirectionalMovementIndexRating:
		return averageDirectionalMovementIndexRating
	case AverageTrueRange:
		return averageTrueRange
	case DirectionalIndicatorMinus:
		return directionalIndicatorMinus
	case DirectionalIndicatorPlus:
		return directionalIndicatorPlus
	case DirectionalMovementIndex:
		return directionalMovementIndex
	case DirectionalMovementMinus:
		return directionalMovementMinus
	case DirectionalMovementPlus:
		return directionalMovementPlus
	case NormalizedAverageTrueRange:
		return normalizedAverageTrueRange
	case ParabolicStopAndReverse:
		return parabolicStopAndReverse
	case RelativeStrengthIndex:
		return relativeStrengthIndex
	case TrueRange:
		return trueRange
	// ── custom ────────────────────────────────────────────────────────────
	case GoertzelSpectrum:
		return goertzelSpectrum
	case MaximumEntropySpectrum:
		return maximumEntropySpectrum
	default:
		return unknown
	}
}

// IsKnown determines if this indicator identifier is known.
func (i Identifier) IsKnown() bool {
	return i > AbsolutePriceOscillator && i < last
}

// MarshalJSON implements the Marshaler interface.
func (i Identifier) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown indicator identifier"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := i.String()
	if s == unknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
//
//nolint:cyclop,funlen
func (i *Identifier) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown indicator identifier"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	// ── common ────────────────────────────────────────────────────────────
	case absolutePriceOscillator:
		*i = AbsolutePriceOscillator
	case exponentialMovingAverage:
		*i = ExponentialMovingAverage
	case linearRegression:
		*i = LinearRegression
	case momentum:
		*i = Momentum
	case pearsonsCorrelationCoefficient:
		*i = PearsonsCorrelationCoefficient
	case rateOfChange:
		*i = RateOfChange
	case rateOfChangePercent:
		*i = RateOfChangePercent
	case rateOfChangeRatio:
		*i = RateOfChangeRatio
	case simpleMovingAverage:
		*i = SimpleMovingAverage
	case standardDeviation:
		*i = StandardDeviation
	case triangularMovingAverage:
		*i = TriangularMovingAverage
	case variance:
		*i = Variance
	case weightedMovingAverage:
		*i = WeightedMovingAverage
	// ── arnaud legoux ──────────────────────────────────────────────────────
	case arnaudLegouxMovingAverage:
		*i = ArnaudLegouxMovingAverage
	// ── donald lambert ─────────────────────────────────────────────────────
	case commodityChannelIndex:
		*i = CommodityChannelIndex
	// ── gene quong ─────────────────────────────────────────────────────────
	case moneyFlowIndex:
		*i = MoneyFlowIndex
	// ── george lane ────────────────────────────────────────────────────────
	case stochastic:
		*i = Stochastic
	// ── gerald appel ───────────────────────────────────────────────────────
	case movingAverageConvergenceDivergence:
		*i = MovingAverageConvergenceDivergence
	case percentagePriceOscillator:
		*i = PercentagePriceOscillator
	// ── igor livshin ───────────────────────────────────────────────────────
	case balanceOfPower:
		*i = BalanceOfPower
	// ── jack hutson ────────────────────────────────────────────────────────
	case tripleExponentialMovingAverageOscillator:
		*i = TripleExponentialMovingAverageOscillator
	// ── john bollinger ─────────────────────────────────────────────────────
	case bollingerBands:
		*i = BollingerBands
	case bollingerBandsTrend:
		*i = BollingerBandsTrend
	// ── john ehlers ────────────────────────────────────────────────────────
	case autoCorrelationIndicator:
		*i = AutoCorrelationIndicator
	case autoCorrelationPeriodogram:
		*i = AutoCorrelationPeriodogram
	case centerOfGravityOscillator:
		*i = CenterOfGravityOscillator
	case combBandPassSpectrum:
		*i = CombBandPassSpectrum
	case coronaSignalToNoiseRatio:
		*i = CoronaSignalToNoiseRatio
	case coronaSpectrum:
		*i = CoronaSpectrum
	case coronaSwingPosition:
		*i = CoronaSwingPosition
	case coronaTrendVigor:
		*i = CoronaTrendVigor
	case cyberCycle:
		*i = CyberCycle
	case discreteFourierTransformSpectrum:
		*i = DiscreteFourierTransformSpectrum
	case dominantCycle:
		*i = DominantCycle
	case fractalAdaptiveMovingAverage:
		*i = FractalAdaptiveMovingAverage
	case hilbertTransformerInstantaneousTrendLine:
		*i = HilbertTransformerInstantaneousTrendLine
	case instantaneousTrendLine:
		*i = InstantaneousTrendLine
	case mesaAdaptiveMovingAverage:
		*i = MesaAdaptiveMovingAverage
	case roofingFilter:
		*i = RoofingFilter
	case sineWave:
		*i = SineWave
	case superSmoother:
		*i = SuperSmoother
	case trendCycleMode:
		*i = TrendCycleMode
	case zeroLagErrorCorrectingExponentialMovingAverage:
		*i = ZeroLagErrorCorrectingExponentialMovingAverage
	case zeroLagExponentialMovingAverage:
		*i = ZeroLagExponentialMovingAverage
	// ── joseph granville ───────────────────────────────────────────────────
	case onBalanceVolume:
		*i = OnBalanceVolume
	// ── larry williams ─────────────────────────────────────────────────────
	case ultimateOscillator:
		*i = UltimateOscillator
	case williamsPercentR:
		*i = WilliamsPercentR
	// ── manfred durschner ──────────────────────────────────────────────────
	case newMovingAverage:
		*i = NewMovingAverage
	// ── marc chaikin ───────────────────────────────────────────────────────
	case advanceDecline:
		*i = AdvanceDecline
	case advanceDeclineOscillator:
		*i = AdvanceDeclineOscillator
	// ── mark jurik ─────────────────────────────────────────────────────────
	case jurikAdaptiveRelativeTrendStrengthIndex:
		*i = JurikAdaptiveRelativeTrendStrengthIndex
	case jurikAdaptiveZeroLagVelocity:
		*i = JurikAdaptiveZeroLagVelocity
	case jurikCommodityChannelIndex:
		*i = JurikCommodityChannelIndex
	case jurikCompositeFractalBehaviorIndex:
		*i = JurikCompositeFractalBehaviorIndex
	case jurikDirectionalMovementIndex:
		*i = JurikDirectionalMovementIndex
	case jurikFractalAdaptiveZeroLagVelocity:
		*i = JurikFractalAdaptiveZeroLagVelocity
	case jurikMovingAverage:
		*i = JurikMovingAverage
	case jurikRelativeTrendStrengthIndex:
		*i = JurikRelativeTrendStrengthIndex
	case jurikTurningPointOscillator:
		*i = JurikTurningPointOscillator
	case jurikWaveletSampler:
		*i = JurikWaveletSampler
	case jurikZeroLagVelocity:
		*i = JurikZeroLagVelocity
	// ── patrick mulloy ─────────────────────────────────────────────────────
	case doubleExponentialMovingAverage:
		*i = DoubleExponentialMovingAverage
	case tripleExponentialMovingAverage:
		*i = TripleExponentialMovingAverage
	// ── perry kaufman ──────────────────────────────────────────────────────
	case kaufmanAdaptiveMovingAverage:
		*i = KaufmanAdaptiveMovingAverage
	// ── tim tillson ────────────────────────────────────────────────────────
	case t2ExponentialMovingAverage:
		*i = T2ExponentialMovingAverage
	case t3ExponentialMovingAverage:
		*i = T3ExponentialMovingAverage
	// ── tushar chande ──────────────────────────────────────────────────────
	case aroon:
		*i = Aroon
	case chandeMomentumOscillator:
		*i = ChandeMomentumOscillator
	case stochasticRelativeStrengthIndex:
		*i = StochasticRelativeStrengthIndex
	// ── vladimir kravchuk ──────────────────────────────────────────────────
	case adaptiveTrendAndCycleFilter:
		*i = AdaptiveTrendAndCycleFilter
	// ── welles wilder ──────────────────────────────────────────────────────
	case averageDirectionalMovementIndex:
		*i = AverageDirectionalMovementIndex
	case averageDirectionalMovementIndexRating:
		*i = AverageDirectionalMovementIndexRating
	case averageTrueRange:
		*i = AverageTrueRange
	case directionalIndicatorMinus:
		*i = DirectionalIndicatorMinus
	case directionalIndicatorPlus:
		*i = DirectionalIndicatorPlus
	case directionalMovementIndex:
		*i = DirectionalMovementIndex
	case directionalMovementMinus:
		*i = DirectionalMovementMinus
	case directionalMovementPlus:
		*i = DirectionalMovementPlus
	case normalizedAverageTrueRange:
		*i = NormalizedAverageTrueRange
	case parabolicStopAndReverse:
		*i = ParabolicStopAndReverse
	case relativeStrengthIndex:
		*i = RelativeStrengthIndex
	case trueRange:
		*i = TrueRange
	// ── custom ────────────────────────────────────────────────────────────
	case goertzelSpectrum:
		*i = GoertzelSpectrum
	case maximumEntropySpectrum:
		*i = MaximumEntropySpectrum
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
