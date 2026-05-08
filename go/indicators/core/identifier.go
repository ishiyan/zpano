package core

import (
	"bytes"
	"fmt"
)

// Identifier identifies an indicator by enumerating all implemented indicators.
type Identifier int

const (
	// SimpleMovingAverage identifies the Simple Moving Average (SMA) indicator.
	SimpleMovingAverage Identifier = iota + 1

	// WeightedMovingAverage identifies the Weighted Moving Average (WMA) indicator.
	WeightedMovingAverage

	// TriangularMovingAverage identifies the Triangular Moving Average (TRIMA) indicator.
	TriangularMovingAverage

	// ExponentialMovingAverage identifies the Exponential Moving Average (EMA) indicator.
	ExponentialMovingAverage

	// DoubleExponentialMovingAverage identifies the Double Exponential Moving Average (DEMA) indicator.
	DoubleExponentialMovingAverage

	// TripleExponentialMovingAverage identifies the Triple Exponential Moving Average (TEMA) indicator.
	TripleExponentialMovingAverage

	// T2ExponentialMovingAverage identifies the T2 Exponential Moving Average (T2) indicator.
	T2ExponentialMovingAverage

	// T3ExponentialMovingAverage identifies the T3 Exponential Moving Average (T3) indicator.
	T3ExponentialMovingAverage

	// KaufmanAdaptiveMovingAverage identifies the Kaufman Adaptive Moving Average (KAMA) indicator.
	KaufmanAdaptiveMovingAverage

	// JurikMovingAverage identifies the Jurik Moving Average (JMA) indicator.
	JurikMovingAverage

	// MesaAdaptiveMovingAverage identifies the Ehlers MESA Adaptive Moving Average (MAMA) indicator.
	MesaAdaptiveMovingAverage

	// FractalAdaptiveMovingAverage identifies the Ehlers Fractal Adaptive Moving Average (FRAMA) indicator.
	FractalAdaptiveMovingAverage

	// DominantCycle identifies the Ehlers Dominant Cycle (DC) indicator, exposing raw period, smoothed period and phase.
	DominantCycle

	// Momentum identifies the momentum (MOM) indicator.
	Momentum

	// RateOfChange identifies the Rate of Change (ROC) indicator.
	RateOfChange

	// RateOfChangePercent identifies the Rate of Change Percent (ROCP) indicator.
	RateOfChangePercent

	// RelativeStrengthIndex identifies the Relative Strength Index (RSI) indicator.
	RelativeStrengthIndex

	// ChandeMomentumOscillator identifies the Chande Momentum Oscillator (CMO) indicator.
	ChandeMomentumOscillator

	// BollingerBands identifies the Bollinger Bands (BB) indicator.
	BollingerBands

	// Variance identifies the Variance (VAR) indicator.
	Variance

	// StandardDeviation identifies the Standard Deviation (STDEV) indicator.
	StandardDeviation

	// GoertzelSpectrum identifies the Goertzel power spectrum (GOERTZEL) indicator.
	GoertzelSpectrum

	// CenterOfGravityOscillator identifies the Ehlers Center of Gravity (COG) oscillator indicator.
	CenterOfGravityOscillator

	// CyberCycle identifies the Ehlers Cyber Cycle (CC) indicator.
	CyberCycle

	// InstantaneousTrendLine identifies the Ehlers Instantaneous Trend Line (iTrend) indicator.
	InstantaneousTrendLine

	// SuperSmoother identifies the Ehlers Super Smoother (SS) indicator.
	SuperSmoother

	// ZeroLagExponentialMovingAverage identifies the Ehlers Zero-lag Exponential Moving Average (ZEMA) indicator.
	ZeroLagExponentialMovingAverage

	// ZeroLagErrorCorrectingExponentialMovingAverage identifies the Ehlers Zero-lag Error-Correcting Exponential Moving Average (ZECEMA) indicator.
	ZeroLagErrorCorrectingExponentialMovingAverage

	// RoofingFilter identifies the Ehlers Roofing Filter indicator.
	RoofingFilter

	// TrueRange identifies the Welles Wilder True Range (TR) indicator.
	TrueRange

	// AverageTrueRange identifies the Welles Wilder Average True Range (ATR) indicator.
	AverageTrueRange

	// NormalizedAverageTrueRange identifies the Welles Wilder Normalized Average True Range (NATR) indicator.
	NormalizedAverageTrueRange

	// DirectionalMovementMinus identifies the Welles Wilder Directional Movement Minus (-DM) indicator.
	DirectionalMovementMinus

	// DirectionalMovementPlus identifies the Welles Wilder Directional Movement Plus (+DM) indicator.
	DirectionalMovementPlus

	// DirectionalIndicatorMinus identifies the Welles Wilder Directional Indicator Minus (-DI) indicator.
	DirectionalIndicatorMinus

	// DirectionalIndicatorPlus identifies the Welles Wilder Directional Indicator Plus (+DI) indicator.
	DirectionalIndicatorPlus

	// DirectionalMovementIndex identifies the Welles Wilder Directional Movement Index (DX) indicator.
	DirectionalMovementIndex

	// AverageDirectionalMovementIndex identifies the Welles Wilder Average Directional Movement Index (ADX) indicator.
	AverageDirectionalMovementIndex

	// AverageDirectionalMovementIndexRating identifies the Welles Wilder Average Directional Movement Index Rating (ADXR) indicator.
	AverageDirectionalMovementIndexRating

	// WilliamsPercentR identifies the Larry Williams Williams %R (WILL%R) indicator.
	WilliamsPercentR

	// PercentagePriceOscillator identifies the Gerald Appel Percentage Price Oscillator (PPO) indicator.
	PercentagePriceOscillator

	// AbsolutePriceOscillator identifies the Absolute Price Oscillator (APO) indicator.
	AbsolutePriceOscillator

	// CommodityChannelIndex identifies the Donald Lambert Commodity Channel Index (CCI) indicator.
	CommodityChannelIndex

	// MoneyFlowIndex identifies the Gene Quong Money Flow Index (MFI) indicator.
	MoneyFlowIndex

	// OnBalanceVolume identifies the Joseph Granville On-Balance Volume (OBV) indicator.
	OnBalanceVolume

	// BalanceOfPower identifies the Igor Livshin Balance of Power (BOP) indicator.
	BalanceOfPower

	// RateOfChangeRatio identifies the Rate of Change Ratio (ROCR / ROCR100) indicator.
	RateOfChangeRatio

	// PearsonsCorrelationCoefficient identifies the Pearson's Correlation Coefficient (CORREL) indicator.
	PearsonsCorrelationCoefficient

	// LinearRegression identifies the Linear Regression (LINEARREG) indicator.
	LinearRegression

	// UltimateOscillator identifies the Larry Williams Ultimate Oscillator (ULTOSC) indicator.
	UltimateOscillator

	// StochasticRelativeStrengthIndex identifies the Tushar Chande Stochastic RSI (STOCHRSI) indicator.
	StochasticRelativeStrengthIndex

	// Stochastic identifies the George Lane Stochastic Oscillator (STOCH) indicator.
	Stochastic

	// Aroon identifies the Tushar Chande Aroon (AROON) indicator.
	Aroon

	// AdvanceDecline identifies the Marc Chaikin Advance-Decline (AD) indicator.
	AdvanceDecline

	// AdvanceDeclineOscillator identifies the Marc Chaikin Advance-Decline Oscillator (ADOSC) indicator.
	AdvanceDeclineOscillator

	// ParabolicStopAndReverse identifies the Welles Wilder Parabolic Stop And Reverse (SAR) indicator.
	ParabolicStopAndReverse

	// TripleExponentialMovingAverageOscillator identifies Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX) indicator.
	TripleExponentialMovingAverageOscillator

	// BollingerBandsTrend identifies John Bollinger's Bollinger Bands Trend (BBTrend) indicator.
	BollingerBandsTrend

	// MovingAverageConvergenceDivergence identifies Gerald Appel's Moving Average Convergence Divergence (MACD) indicator.
	MovingAverageConvergenceDivergence

	// SineWave identifies the Ehlers Sine Wave (SW) indicator, exposing sine value, lead sine, band, dominant cycle period and phase.
	SineWave

	// HilbertTransformerInstantaneousTrendLine identifies the Ehlers Hilbert Transformer Instantaneous Trend Line (HTITL) indicator,
	// exposing the trend line value and the smoothed dominant cycle period.
	HilbertTransformerInstantaneousTrendLine

	// TrendCycleMode identifies the Ehlers Trend / Cycle Mode (TCM) indicator, exposing the trend/cycle value
	// (+1 in trend, −1 in cycle), trend/cycle mode flags, instantaneous trend line, sine wave, lead sine wave,
	// dominant cycle period and phase.
	TrendCycleMode

	// CoronaSpectrum identifies the Ehlers Corona Spectrum (CSPECT) indicator, a heat-map of cyclic activity
	// over a cycle-period range together with the dominant cycle period and its 5-sample median.
	CoronaSpectrum

	// CoronaSignalToNoiseRatio identifies the Ehlers Corona Signal-to-Noise Ratio (CSNR) indicator, a heat-map
	// of SNR plus a smoothed SNR scalar line.
	CoronaSignalToNoiseRatio

	// CoronaSwingPosition identifies the Ehlers Corona Swing Position (CSWING) indicator, a heat-map of swing
	// position with a scalar swing-position line.
	CoronaSwingPosition

	// CoronaTrendVigor identifies the Ehlers Corona Trend Vigor (CTV) indicator, a heat-map of trend vigor
	// with a scalar trend-vigor line.
	CoronaTrendVigor

	// AdaptiveTrendAndCycleFilter identifies the Vladimir Kravchuk Adaptive Trend & Cycle Filter (ATCF)
	// suite, exposing FATL, SATL, RFTL, RSTL, RBCI FIR-filter outputs together with the derived
	// FTLM, STLM, and PCCI composites.
	AdaptiveTrendAndCycleFilter

	// MaximumEntropySpectrum identifies the Maximum Entropy Spectrum (MESPECT) indicator, a
	// heat-map of cyclic activity estimated via Burg's maximum-entropy auto-regressive method.
	MaximumEntropySpectrum

	// DiscreteFourierTransformSpectrum identifies the Discrete Fourier Transform Spectrum
	// (psDft) indicator, a heat-map of cyclic activity estimated via a discrete Fourier
	// transform over a sliding window.
	DiscreteFourierTransformSpectrum

	// CombBandPassSpectrum identifies the Comb Band-Pass Spectrum (cbps) indicator,
	// a heat-map of cyclic activity estimated via a bank of 2-pole band-pass filters,
	// one per integer cycle period, following Ehlers' EasyLanguage listing 10-1.
	CombBandPassSpectrum

	// AutoCorrelationIndicator identifies the Autocorrelation Indicator (aci)
	// heat-map of Pearson correlation coefficients between the current filtered
	// series and a lagged copy of itself, following Ehlers' EasyLanguage listing 8-2.
	AutoCorrelationIndicator

	// AutoCorrelationPeriodogram identifies the Autocorrelation Periodogram (acp)
	// heat-map of cyclic activity estimated via a discrete Fourier transform of the
	// autocorrelation function, following Ehlers' EasyLanguage listing 8-3.
	AutoCorrelationPeriodogram

	// JurikRelativeTrendStrengthIndex identifies the Jurik Relative Trend Strength Index (RSX) indicator.
	JurikRelativeTrendStrengthIndex

	// JurikCompositeFractalBehaviorIndex identifies the Jurik Composite Fractal Behavior Index (CFB) indicator.
	JurikCompositeFractalBehaviorIndex

	// JurikZeroLagVelocity identifies the Jurik Zero Lag Velocity (VEL) indicator.
	JurikZeroLagVelocity

	// JurikDirectionalMovementIndex identifies the Jurik Directional Movement Index (DMX) indicator.
	JurikDirectionalMovementIndex

	// JurikCommodityChannelIndex identifies the Jurik Commodity Channel Index (JCCX) indicator.
	JurikCommodityChannelIndex

	// JurikWaveletSampler identifies the Jurik Wavelet Sampler (WAV) indicator.
	JurikWaveletSampler

	// JurikAdaptiveZeroLagVelocity identifies the Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.
	JurikAdaptiveZeroLagVelocity

	// JurikFractalAdaptiveZeroLagVelocity identifies the Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) indicator.
	JurikFractalAdaptiveZeroLagVelocity

	// JurikAdaptiveRelativeTrendStrengthIndex identifies the Jurik Adaptive Relative Trend Strength Index (JARSX) indicator.
	JurikAdaptiveRelativeTrendStrengthIndex

	// JurikTurningPointOscillator identifies the Jurik Turning Point Oscillator (JTPO) indicator.
	JurikTurningPointOscillator

	last
)

const (
	unknown                                        = "unknown"
	simpleMovingAverage                            = "simpleMovingAverage"
	weightedMovingAverage                          = "weightedMovingAverage"
	triangularMovingAverage                        = "triangularMovingAverage"
	exponentialMovingAverage                       = "exponentialMovingAverage"
	doubleExponentialMovingAverage                 = "doubleExponentialMovingAverage"
	tripleExponentialMovingAverage                 = "tripleExponentialMovingAverage"
	t2ExponentialMovingAverage                     = "t2ExponentialMovingAverage"
	t3ExponentialMovingAverage                     = "t3ExponentialMovingAverage"
	kaufmanAdaptiveMovingAverage                   = "kaufmanAdaptiveMovingAverage"
	jurikMovingAverage                             = "jurikMovingAverage"
	mesaAdaptiveMovingAverage                      = "mesaAdaptiveMovingAverage"
	fractalAdaptiveMovingAverage                   = "fractalAdaptiveMovingAverage"
	dominantCycle                                  = "dominantCycle"
	momentum                                       = "momentum"
	rateOfChange                                   = "rateOfChange"
	rateOfChangePercent                            = "rateOfChangePercent"
	relativeStrengthIndex                          = "relativeStrengthIndex"
	chandeMomentumOscillator                       = "chandeMomentumOscillator"
	bollingerBands                                 = "bollingerBands"
	variance                                       = "variance"
	standardDeviation                              = "standardDeviation"
	goertzelSpectrum                               = "goertzelSpectrum"
	centerOfGravityOscillator                      = "centerOfGravityOscillator"
	cyberCycle                                     = "cyberCycle"
	instantaneousTrendLine                         = "instantaneousTrendLine"
	superSmoother                                  = "superSmoother"
	zeroLagExponentialMovingAverage                = "zeroLagExponentialMovingAverage"
	zeroLagErrorCorrectingExponentialMovingAverage = "zeroLagErrorCorrectingExponentialMovingAverage"
	roofingFilter                                  = "roofingFilter"
	trueRange                                      = "trueRange"
	averageTrueRange                               = "averageTrueRange"
	normalizedAverageTrueRange                     = "normalizedAverageTrueRange"
	directionalMovementMinus                       = "directionalMovementMinus"
	directionalMovementPlus                        = "directionalMovementPlus"
	directionalIndicatorMinus                      = "directionalIndicatorMinus"
	directionalIndicatorPlus                       = "directionalIndicatorPlus"
	directionalMovementIndex                       = "directionalMovementIndex"
	averageDirectionalMovementIndex                = "averageDirectionalMovementIndex"
	averageDirectionalMovementIndexRating          = "averageDirectionalMovementIndexRating"
	williamsPercentR                               = "williamsPercentR"
	percentagePriceOscillator                      = "percentagePriceOscillator"
	absolutePriceOscillator                        = "absolutePriceOscillator"
	commodityChannelIndex                          = "commodityChannelIndex"
	moneyFlowIndex                                 = "moneyFlowIndex"
	onBalanceVolume                                = "onBalanceVolume"
	balanceOfPower                                 = "balanceOfPower"
	rateOfChangeRatio                              = "rateOfChangeRatio"
	pearsonsCorrelationCoefficient                 = "pearsonsCorrelationCoefficient"
	linearRegression                               = "linearRegression"
	ultimateOscillator                             = "ultimateOscillator"
	stochasticRelativeStrengthIndex                = "stochasticRelativeStrengthIndex"
	stochastic                                     = "stochastic"
	aroon                                          = "aroon"
	advanceDecline                                 = "advanceDecline"
	advanceDeclineOscillator                       = "advanceDeclineOscillator"
	parabolicStopAndReverse                        = "parabolicStopAndReverse"
	tripleExponentialMovingAverageOscillator       = "tripleExponentialMovingAverageOscillator"
	bollingerBandsTrend                            = "bollingerBandsTrend"
	movingAverageConvergenceDivergence             = "movingAverageConvergenceDivergence"
	sineWave                                       = "sineWave"
	hilbertTransformerInstantaneousTrendLine       = "hilbertTransformerInstantaneousTrendLine"
	trendCycleMode                                 = "trendCycleMode"
	coronaSpectrum                                 = "coronaSpectrum"
	coronaSignalToNoiseRatio                       = "coronaSignalToNoiseRatio"
	coronaSwingPosition                            = "coronaSwingPosition"
	coronaTrendVigor                               = "coronaTrendVigor"
	adaptiveTrendAndCycleFilter                    = "adaptiveTrendAndCycleFilter"
	maximumEntropySpectrum                         = "maximumEntropySpectrum"
	discreteFourierTransformSpectrum               = "discreteFourierTransformSpectrum"
	combBandPassSpectrum                           = "combBandPassSpectrum"
	autoCorrelationIndicator                       = "autoCorrelationIndicator"
	autoCorrelationPeriodogram                     = "autoCorrelationPeriodogram"
	jurikRelativeTrendStrengthIndex                = "jurikRelativeTrendStrengthIndex"
	jurikCompositeFractalBehaviorIndex             = "jurikCompositeFractalBehaviorIndex"
	jurikZeroLagVelocity                           = "jurikZeroLagVelocity"
	jurikDirectionalMovementIndex                  = "jurikDirectionalMovementIndex"
	jurikCommodityChannelIndex                     = "jurikCommodityChannelIndex"
	jurikWaveletSampler                            = "jurikWaveletSampler"
	jurikAdaptiveZeroLagVelocity                   = "jurikAdaptiveZeroLagVelocity"
	jurikFractalAdaptiveZeroLagVelocity            = "jurikFractalAdaptiveZeroLagVelocity"
	jurikAdaptiveRelativeTrendStrengthIndex         = "jurikAdaptiveRelativeTrendStrengthIndex"
	jurikTurningPointOscillator                    = "jurikTurningPointOscillator"
)

// String implements the Stringer interface.
//
//nolint:exhaustive,cyclop,funlen
func (i Identifier) String() string {
	switch i {
	case SimpleMovingAverage:
		return simpleMovingAverage
	case WeightedMovingAverage:
		return weightedMovingAverage
	case TriangularMovingAverage:
		return triangularMovingAverage
	case ExponentialMovingAverage:
		return exponentialMovingAverage
	case DoubleExponentialMovingAverage:
		return doubleExponentialMovingAverage
	case TripleExponentialMovingAverage:
		return tripleExponentialMovingAverage
	case T2ExponentialMovingAverage:
		return t2ExponentialMovingAverage
	case T3ExponentialMovingAverage:
		return t3ExponentialMovingAverage
	case KaufmanAdaptiveMovingAverage:
		return kaufmanAdaptiveMovingAverage
	case JurikMovingAverage:
		return jurikMovingAverage
	case MesaAdaptiveMovingAverage:
		return mesaAdaptiveMovingAverage
	case FractalAdaptiveMovingAverage:
		return fractalAdaptiveMovingAverage
	case DominantCycle:
		return dominantCycle
	case Momentum:
		return momentum
	case RateOfChange:
		return rateOfChange
	case RateOfChangePercent:
		return rateOfChangePercent
	case RelativeStrengthIndex:
		return relativeStrengthIndex
	case ChandeMomentumOscillator:
		return chandeMomentumOscillator
	case BollingerBands:
		return bollingerBands
	case Variance:
		return variance
	case StandardDeviation:
		return standardDeviation
	case GoertzelSpectrum:
		return goertzelSpectrum
	case CenterOfGravityOscillator:
		return centerOfGravityOscillator
	case CyberCycle:
		return cyberCycle
	case InstantaneousTrendLine:
		return instantaneousTrendLine
	case SuperSmoother:
		return superSmoother
	case ZeroLagExponentialMovingAverage:
		return zeroLagExponentialMovingAverage
	case ZeroLagErrorCorrectingExponentialMovingAverage:
		return zeroLagErrorCorrectingExponentialMovingAverage
	case RoofingFilter:
		return roofingFilter
	case TrueRange:
		return trueRange
	case AverageTrueRange:
		return averageTrueRange
	case NormalizedAverageTrueRange:
		return normalizedAverageTrueRange
	case DirectionalMovementMinus:
		return directionalMovementMinus
	case DirectionalMovementPlus:
		return directionalMovementPlus
	case DirectionalIndicatorMinus:
		return directionalIndicatorMinus
	case DirectionalIndicatorPlus:
		return directionalIndicatorPlus
	case DirectionalMovementIndex:
		return directionalMovementIndex
	case AverageDirectionalMovementIndex:
		return averageDirectionalMovementIndex
	case AverageDirectionalMovementIndexRating:
		return averageDirectionalMovementIndexRating
	case WilliamsPercentR:
		return williamsPercentR
	case PercentagePriceOscillator:
		return percentagePriceOscillator
	case AbsolutePriceOscillator:
		return absolutePriceOscillator
	case CommodityChannelIndex:
		return commodityChannelIndex
	case MoneyFlowIndex:
		return moneyFlowIndex
	case OnBalanceVolume:
		return onBalanceVolume
	case BalanceOfPower:
		return balanceOfPower
	case RateOfChangeRatio:
		return rateOfChangeRatio
	case PearsonsCorrelationCoefficient:
		return pearsonsCorrelationCoefficient
	case LinearRegression:
		return linearRegression
	case UltimateOscillator:
		return ultimateOscillator
	case StochasticRelativeStrengthIndex:
		return stochasticRelativeStrengthIndex
	case Stochastic:
		return stochastic
	case Aroon:
		return aroon
	case AdvanceDecline:
		return advanceDecline
	case AdvanceDeclineOscillator:
		return advanceDeclineOscillator
	case ParabolicStopAndReverse:
		return parabolicStopAndReverse
	case TripleExponentialMovingAverageOscillator:
		return tripleExponentialMovingAverageOscillator
	case BollingerBandsTrend:
		return bollingerBandsTrend
	case MovingAverageConvergenceDivergence:
		return movingAverageConvergenceDivergence
	case SineWave:
		return sineWave
	case HilbertTransformerInstantaneousTrendLine:
		return hilbertTransformerInstantaneousTrendLine
	case TrendCycleMode:
		return trendCycleMode
	case CoronaSpectrum:
		return coronaSpectrum
	case CoronaSignalToNoiseRatio:
		return coronaSignalToNoiseRatio
	case CoronaSwingPosition:
		return coronaSwingPosition
	case CoronaTrendVigor:
		return coronaTrendVigor
	case AdaptiveTrendAndCycleFilter:
		return adaptiveTrendAndCycleFilter
	case MaximumEntropySpectrum:
		return maximumEntropySpectrum
	case DiscreteFourierTransformSpectrum:
		return discreteFourierTransformSpectrum
	case CombBandPassSpectrum:
		return combBandPassSpectrum
	case AutoCorrelationIndicator:
		return autoCorrelationIndicator
	case AutoCorrelationPeriodogram:
		return autoCorrelationPeriodogram
	case JurikRelativeTrendStrengthIndex:
		return jurikRelativeTrendStrengthIndex
	case JurikCompositeFractalBehaviorIndex:
		return jurikCompositeFractalBehaviorIndex
	case JurikZeroLagVelocity:
		return jurikZeroLagVelocity
	case JurikDirectionalMovementIndex:
		return jurikDirectionalMovementIndex
	case JurikCommodityChannelIndex:
		return jurikCommodityChannelIndex
	case JurikWaveletSampler:
		return jurikWaveletSampler
	case JurikAdaptiveZeroLagVelocity:
		return jurikAdaptiveZeroLagVelocity
	case JurikFractalAdaptiveZeroLagVelocity:
		return jurikFractalAdaptiveZeroLagVelocity
	case JurikAdaptiveRelativeTrendStrengthIndex:
		return jurikAdaptiveRelativeTrendStrengthIndex
	case JurikTurningPointOscillator:
		return jurikTurningPointOscillator
	default:
		return unknown
	}
}

// IsKnown determines if this indicator identifier is known.
func (i Identifier) IsKnown() bool {
	return i >= SimpleMovingAverage && i < last
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
	case simpleMovingAverage:
		*i = SimpleMovingAverage
	case weightedMovingAverage:
		*i = WeightedMovingAverage
	case triangularMovingAverage:
		*i = TriangularMovingAverage
	case exponentialMovingAverage:
		*i = ExponentialMovingAverage
	case doubleExponentialMovingAverage:
		*i = DoubleExponentialMovingAverage
	case tripleExponentialMovingAverage:
		*i = TripleExponentialMovingAverage
	case t2ExponentialMovingAverage:
		*i = T2ExponentialMovingAverage
	case t3ExponentialMovingAverage:
		*i = T3ExponentialMovingAverage
	case kaufmanAdaptiveMovingAverage:
		*i = KaufmanAdaptiveMovingAverage
	case jurikMovingAverage:
		*i = JurikMovingAverage
	case mesaAdaptiveMovingAverage:
		*i = MesaAdaptiveMovingAverage
	case fractalAdaptiveMovingAverage:
		*i = FractalAdaptiveMovingAverage
	case dominantCycle:
		*i = DominantCycle
	case momentum:
		*i = Momentum
	case rateOfChange:
		*i = RateOfChange
	case rateOfChangePercent:
		*i = RateOfChangePercent
	case relativeStrengthIndex:
		*i = RelativeStrengthIndex
	case chandeMomentumOscillator:
		*i = ChandeMomentumOscillator
	case bollingerBands:
		*i = BollingerBands
	case variance:
		*i = Variance
	case standardDeviation:
		*i = StandardDeviation
	case goertzelSpectrum:
		*i = GoertzelSpectrum
	case centerOfGravityOscillator:
		*i = CenterOfGravityOscillator
	case cyberCycle:
		*i = CyberCycle
	case instantaneousTrendLine:
		*i = InstantaneousTrendLine
	case superSmoother:
		*i = SuperSmoother
	case zeroLagExponentialMovingAverage:
		*i = ZeroLagExponentialMovingAverage
	case zeroLagErrorCorrectingExponentialMovingAverage:
		*i = ZeroLagErrorCorrectingExponentialMovingAverage
	case roofingFilter:
		*i = RoofingFilter
	case trueRange:
		*i = TrueRange
	case averageTrueRange:
		*i = AverageTrueRange
	case normalizedAverageTrueRange:
		*i = NormalizedAverageTrueRange
	case directionalMovementMinus:
		*i = DirectionalMovementMinus
	case directionalMovementPlus:
		*i = DirectionalMovementPlus
	case directionalIndicatorMinus:
		*i = DirectionalIndicatorMinus
	case directionalIndicatorPlus:
		*i = DirectionalIndicatorPlus
	case directionalMovementIndex:
		*i = DirectionalMovementIndex
	case averageDirectionalMovementIndex:
		*i = AverageDirectionalMovementIndex
	case averageDirectionalMovementIndexRating:
		*i = AverageDirectionalMovementIndexRating
	case williamsPercentR:
		*i = WilliamsPercentR
	case percentagePriceOscillator:
		*i = PercentagePriceOscillator
	case absolutePriceOscillator:
		*i = AbsolutePriceOscillator
	case commodityChannelIndex:
		*i = CommodityChannelIndex
	case moneyFlowIndex:
		*i = MoneyFlowIndex
	case onBalanceVolume:
		*i = OnBalanceVolume
	case balanceOfPower:
		*i = BalanceOfPower
	case rateOfChangeRatio:
		*i = RateOfChangeRatio
	case pearsonsCorrelationCoefficient:
		*i = PearsonsCorrelationCoefficient
	case linearRegression:
		*i = LinearRegression
	case ultimateOscillator:
		*i = UltimateOscillator
	case stochasticRelativeStrengthIndex:
		*i = StochasticRelativeStrengthIndex
	case stochastic:
		*i = Stochastic
	case aroon:
		*i = Aroon
	case advanceDecline:
		*i = AdvanceDecline
	case advanceDeclineOscillator:
		*i = AdvanceDeclineOscillator
	case parabolicStopAndReverse:
		*i = ParabolicStopAndReverse
	case tripleExponentialMovingAverageOscillator:
		*i = TripleExponentialMovingAverageOscillator
	case bollingerBandsTrend:
		*i = BollingerBandsTrend
	case movingAverageConvergenceDivergence:
		*i = MovingAverageConvergenceDivergence
	case sineWave:
		*i = SineWave
	case hilbertTransformerInstantaneousTrendLine:
		*i = HilbertTransformerInstantaneousTrendLine
	case trendCycleMode:
		*i = TrendCycleMode
	case coronaSpectrum:
		*i = CoronaSpectrum
	case coronaSignalToNoiseRatio:
		*i = CoronaSignalToNoiseRatio
	case coronaSwingPosition:
		*i = CoronaSwingPosition
	case coronaTrendVigor:
		*i = CoronaTrendVigor
	case adaptiveTrendAndCycleFilter:
		*i = AdaptiveTrendAndCycleFilter
	case maximumEntropySpectrum:
		*i = MaximumEntropySpectrum
	case discreteFourierTransformSpectrum:
		*i = DiscreteFourierTransformSpectrum
	case combBandPassSpectrum:
		*i = CombBandPassSpectrum
	case autoCorrelationIndicator:
		*i = AutoCorrelationIndicator
	case autoCorrelationPeriodogram:
		*i = AutoCorrelationPeriodogram
	case jurikRelativeTrendStrengthIndex:
		*i = JurikRelativeTrendStrengthIndex
	case jurikCompositeFractalBehaviorIndex:
		*i = JurikCompositeFractalBehaviorIndex
	case jurikZeroLagVelocity:
		*i = JurikZeroLagVelocity
	case jurikDirectionalMovementIndex:
		*i = JurikDirectionalMovementIndex
	case jurikCommodityChannelIndex:
		*i = JurikCommodityChannelIndex
	case jurikWaveletSampler:
		*i = JurikWaveletSampler
	case jurikAdaptiveZeroLagVelocity:
		*i = JurikAdaptiveZeroLagVelocity
	case jurikFractalAdaptiveZeroLagVelocity:
		*i = JurikFractalAdaptiveZeroLagVelocity
	case jurikAdaptiveRelativeTrendStrengthIndex:
		*i = JurikAdaptiveRelativeTrendStrengthIndex
	case jurikTurningPointOscillator:
		*i = JurikTurningPointOscillator
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
