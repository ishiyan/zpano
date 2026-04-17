package core

import (
	"bytes"
	"fmt"
)

// Type Identifies an indicator by enumerating all implemented indicators.
type Type int

const (
	// SimpleMovingAverage identifies the Simple Moving Average (SMA) indicator.
	SimpleMovingAverage Type = iota + 1

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
	kaufmanAdaptiveMovingAverage                   = "kaufmanAdaptiveMovingAverageMovingAverage"
	jurikMovingAverage                             = "jurikMovingAverage"
	mesaAdaptiveMovingAverage                      = "mesaAdaptiveMovingAverage"
	fractalAdaptiveMovingAverage                   = "fractalAdaptiveMovingAverage"
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
)

// String implements the Stringer interface.
//
//nolint:exhaustive,cyclop,funlen
func (t Type) String() string {
	switch t {
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
	default:
		return unknown
	}
}

// IsKnown determines if this indicator type is known.
func (t Type) IsKnown() bool {
	return t >= SimpleMovingAverage && t < last
}

// MarshalJSON implements the Marshaler interface.
func (t Type) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown indicator type"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := t.String()
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
func (t *Type) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown indicator type"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case simpleMovingAverage:
		*t = SimpleMovingAverage
	case weightedMovingAverage:
		*t = WeightedMovingAverage
	case triangularMovingAverage:
		*t = TriangularMovingAverage
	case exponentialMovingAverage:
		*t = ExponentialMovingAverage
	case doubleExponentialMovingAverage:
		*t = DoubleExponentialMovingAverage
	case tripleExponentialMovingAverage:
		*t = TripleExponentialMovingAverage
	case t2ExponentialMovingAverage:
		*t = T2ExponentialMovingAverage
	case t3ExponentialMovingAverage:
		*t = T3ExponentialMovingAverage
	case kaufmanAdaptiveMovingAverage:
		*t = KaufmanAdaptiveMovingAverage
	case jurikMovingAverage:
		*t = JurikMovingAverage
	case mesaAdaptiveMovingAverage:
		*t = MesaAdaptiveMovingAverage
	case fractalAdaptiveMovingAverage:
		*t = FractalAdaptiveMovingAverage
	case momentum:
		*t = Momentum
	case rateOfChange:
		*t = RateOfChange
	case rateOfChangePercent:
		*t = RateOfChangePercent
	case relativeStrengthIndex:
		*t = RelativeStrengthIndex
	case chandeMomentumOscillator:
		*t = ChandeMomentumOscillator
	case bollingerBands:
		*t = BollingerBands
	case variance:
		*t = Variance
	case standardDeviation:
		*t = StandardDeviation
	case goertzelSpectrum:
		*t = GoertzelSpectrum
	case centerOfGravityOscillator:
		*t = CenterOfGravityOscillator
	case cyberCycle:
		*t = CyberCycle
	case instantaneousTrendLine:
		*t = InstantaneousTrendLine
	case superSmoother:
		*t = SuperSmoother
	case zeroLagExponentialMovingAverage:
		*t = ZeroLagExponentialMovingAverage
	case zeroLagErrorCorrectingExponentialMovingAverage:
		*t = ZeroLagErrorCorrectingExponentialMovingAverage
	case roofingFilter:
		*t = RoofingFilter
	case trueRange:
		*t = TrueRange
	case averageTrueRange:
		*t = AverageTrueRange
	case normalizedAverageTrueRange:
		*t = NormalizedAverageTrueRange
	case directionalMovementMinus:
		*t = DirectionalMovementMinus
	case directionalMovementPlus:
		*t = DirectionalMovementPlus
	case directionalIndicatorMinus:
		*t = DirectionalIndicatorMinus
	case directionalIndicatorPlus:
		*t = DirectionalIndicatorPlus
	case directionalMovementIndex:
		*t = DirectionalMovementIndex
	case averageDirectionalMovementIndex:
		*t = AverageDirectionalMovementIndex
	case averageDirectionalMovementIndexRating:
		*t = AverageDirectionalMovementIndexRating
	case williamsPercentR:
		*t = WilliamsPercentR
	case percentagePriceOscillator:
		*t = PercentagePriceOscillator
	case absolutePriceOscillator:
		*t = AbsolutePriceOscillator
	case commodityChannelIndex:
		*t = CommodityChannelIndex
	case moneyFlowIndex:
		*t = MoneyFlowIndex
	case onBalanceVolume:
		*t = OnBalanceVolume
	case balanceOfPower:
		*t = BalanceOfPower
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
