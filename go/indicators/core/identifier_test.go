//nolint:testpackage
package core

import (
	"testing"
)

func TestIdentifierString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i    Identifier
		text string
	}{
		{SimpleMovingAverage, simpleMovingAverage},
		{WeightedMovingAverage, weightedMovingAverage},
		{TriangularMovingAverage, triangularMovingAverage},
		{ExponentialMovingAverage, exponentialMovingAverage},
		{DoubleExponentialMovingAverage, doubleExponentialMovingAverage},
		{TripleExponentialMovingAverage, tripleExponentialMovingAverage},
		{T3ExponentialMovingAverage, t3ExponentialMovingAverage},
		{KaufmanAdaptiveMovingAverage, kaufmanAdaptiveMovingAverage},
		{JurikMovingAverage, jurikMovingAverage},
		{MesaAdaptiveMovingAverage, mesaAdaptiveMovingAverage},
		{FractalAdaptiveMovingAverage, fractalAdaptiveMovingAverage},
		{Momentum, momentum},
		{RateOfChange, rateOfChange},
		{RateOfChangePercent, rateOfChangePercent},
		{RelativeStrengthIndex, relativeStrengthIndex},
		{ChandeMomentumOscillator, chandeMomentumOscillator},
		{BollingerBands, bollingerBands},
		{Variance, variance},
		{StandardDeviation, standardDeviation},
		{GoertzelSpectrum, goertzelSpectrum},
		{CenterOfGravityOscillator, centerOfGravityOscillator},
		{CyberCycle, cyberCycle},
		{InstantaneousTrendLine, instantaneousTrendLine},
		{SuperSmoother, superSmoother},
		{ZeroLagExponentialMovingAverage, zeroLagExponentialMovingAverage},
		{ZeroLagErrorCorrectingExponentialMovingAverage, zeroLagErrorCorrectingExponentialMovingAverage},
		{RoofingFilter, roofingFilter},
		{TrueRange, trueRange},
		{AverageTrueRange, averageTrueRange},
		{NormalizedAverageTrueRange, normalizedAverageTrueRange},
		{DirectionalMovementMinus, directionalMovementMinus},
		{DirectionalMovementPlus, directionalMovementPlus},
		{DirectionalIndicatorMinus, directionalIndicatorMinus},
		{DirectionalIndicatorPlus, directionalIndicatorPlus},
		{DirectionalMovementIndex, directionalMovementIndex},
		{AverageDirectionalMovementIndex, averageDirectionalMovementIndex},
		{AverageDirectionalMovementIndexRating, averageDirectionalMovementIndexRating},
		{WilliamsPercentR, williamsPercentR},
		{PercentagePriceOscillator, percentagePriceOscillator},
		{AbsolutePriceOscillator, absolutePriceOscillator},
		{CommodityChannelIndex, commodityChannelIndex},
		{MoneyFlowIndex, moneyFlowIndex},
		{OnBalanceVolume, onBalanceVolume},
		{BalanceOfPower, balanceOfPower},
		{RateOfChangeRatio, rateOfChangeRatio},
		{PearsonsCorrelationCoefficient, pearsonsCorrelationCoefficient},
		{LinearRegression, linearRegression},
		{UltimateOscillator, ultimateOscillator},
		{StochasticRelativeStrengthIndex, stochasticRelativeStrengthIndex},
		{Stochastic, stochastic},
		{Aroon, aroon},
		{AdvanceDecline, advanceDecline},
		{AdvanceDeclineOscillator, advanceDeclineOscillator},
		{ParabolicStopAndReverse, parabolicStopAndReverse},
		{TripleExponentialMovingAverageOscillator, tripleExponentialMovingAverageOscillator},
		{BollingerBandsTrend, bollingerBandsTrend},
		{MovingAverageConvergenceDivergence, movingAverageConvergenceDivergence},
		{SineWave, sineWave},
		{HilbertTransformerInstantaneousTrendLine, hilbertTransformerInstantaneousTrendLine},
		{TrendCycleMode, trendCycleMode},
		{CoronaSpectrum, coronaSpectrum},
		{CoronaSignalToNoiseRatio, coronaSignalToNoiseRatio},
		{CoronaSwingPosition, coronaSwingPosition},
		{CoronaTrendVigor, coronaTrendVigor},
		{AdaptiveTrendAndCycleFilter, adaptiveTrendAndCycleFilter},
		{MaximumEntropySpectrum, maximumEntropySpectrum},
		{DiscreteFourierTransformSpectrum, discreteFourierTransformSpectrum},
		{CombBandPassSpectrum, combBandPassSpectrum},
		{AutoCorrelationIndicator, autoCorrelationIndicator},
		{AutoCorrelationPeriodogram, autoCorrelationPeriodogram},
		{DominantCycle, dominantCycle},
		{last, unknown},
		{Identifier(0), unknown},
		{Identifier(9999), unknown},
		{Identifier(-9999), unknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.i.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.i, exp, act)
		}
	}
}

func TestIdentifierIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i       Identifier
		boolean bool
	}{
		{SimpleMovingAverage, true},
		{WeightedMovingAverage, true},
		{TriangularMovingAverage, true},
		{ExponentialMovingAverage, true},
		{DoubleExponentialMovingAverage, true},
		{TripleExponentialMovingAverage, true},
		{T3ExponentialMovingAverage, true},
		{KaufmanAdaptiveMovingAverage, true},
		{JurikMovingAverage, true},
		{MesaAdaptiveMovingAverage, true},
		{FractalAdaptiveMovingAverage, true},
		{Momentum, true},
		{RateOfChange, true},
		{RateOfChangePercent, true},
		{RelativeStrengthIndex, true},
		{ChandeMomentumOscillator, true},
		{BollingerBands, true},
		{Variance, true},
		{StandardDeviation, true},
		{GoertzelSpectrum, true},
		{CenterOfGravityOscillator, true},
		{CyberCycle, true},
		{InstantaneousTrendLine, true},
		{SuperSmoother, true},
		{ZeroLagExponentialMovingAverage, true},
		{ZeroLagErrorCorrectingExponentialMovingAverage, true},
		{RoofingFilter, true},
		{TrueRange, true},
		{AverageTrueRange, true},
		{NormalizedAverageTrueRange, true},
		{DirectionalMovementMinus, true},
		{DirectionalMovementPlus, true},
		{DirectionalIndicatorMinus, true},
		{DirectionalIndicatorPlus, true},
		{DirectionalMovementIndex, true},
		{AverageDirectionalMovementIndex, true},
		{AverageDirectionalMovementIndexRating, true},
		{WilliamsPercentR, true},
		{PercentagePriceOscillator, true},
		{AbsolutePriceOscillator, true},
		{CommodityChannelIndex, true},
		{MoneyFlowIndex, true},
		{OnBalanceVolume, true},
		{BalanceOfPower, true},
		{RateOfChangeRatio, true},
		{PearsonsCorrelationCoefficient, true},
		{LinearRegression, true},
		{UltimateOscillator, true},
		{StochasticRelativeStrengthIndex, true},
		{Stochastic, true},
		{Aroon, true},
		{AdvanceDecline, true},
		{AdvanceDeclineOscillator, true},
		{ParabolicStopAndReverse, true},
		{TripleExponentialMovingAverageOscillator, true},
		{BollingerBandsTrend, true},
		{MovingAverageConvergenceDivergence, true},
		{SineWave, true},
		{HilbertTransformerInstantaneousTrendLine, true},
		{TrendCycleMode, true},
		{CoronaSpectrum, true},
		{CoronaSignalToNoiseRatio, true},
		{CoronaSwingPosition, true},
		{CoronaTrendVigor, true},
		{AdaptiveTrendAndCycleFilter, true},
		{MaximumEntropySpectrum, true},
		{DiscreteFourierTransformSpectrum, true},
		{CombBandPassSpectrum, true},
		{AutoCorrelationIndicator, true},
		{AutoCorrelationPeriodogram, true},
		{DominantCycle, true},
		{last, false},
		{Identifier(0), false},
		{Identifier(9999), false},
		{Identifier(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.i.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.i, exp, act)
		}
	}
}

func TestIdentifierMarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var nilstr string
	tests := []struct {
		i         Identifier
		json      string
		succeeded bool
	}{
		{SimpleMovingAverage, dqs + simpleMovingAverage + dqs, true},
		{WeightedMovingAverage, dqs + weightedMovingAverage + dqs, true},
		{TriangularMovingAverage, dqs + triangularMovingAverage + dqs, true},
		{ExponentialMovingAverage, dqs + exponentialMovingAverage + dqs, true},
		{DoubleExponentialMovingAverage, dqs + doubleExponentialMovingAverage + dqs, true},
		{TripleExponentialMovingAverage, dqs + tripleExponentialMovingAverage + dqs, true},
		{T3ExponentialMovingAverage, dqs + t3ExponentialMovingAverage + dqs, true},
		{KaufmanAdaptiveMovingAverage, dqs + kaufmanAdaptiveMovingAverage + dqs, true},
		{JurikMovingAverage, dqs + jurikMovingAverage + dqs, true},
		{MesaAdaptiveMovingAverage, dqs + mesaAdaptiveMovingAverage + dqs, true},
		{FractalAdaptiveMovingAverage, dqs + fractalAdaptiveMovingAverage + dqs, true},
		{Momentum, dqs + momentum + dqs, true},
		{RateOfChange, dqs + rateOfChange + dqs, true},
		{RateOfChangePercent, dqs + rateOfChangePercent + dqs, true},
		{RelativeStrengthIndex, dqs + relativeStrengthIndex + dqs, true},
		{ChandeMomentumOscillator, dqs + chandeMomentumOscillator + dqs, true},
		{BollingerBands, dqs + bollingerBands + dqs, true},
		{Variance, dqs + variance + dqs, true},
		{StandardDeviation, dqs + standardDeviation + dqs, true},
		{GoertzelSpectrum, dqs + goertzelSpectrum + dqs, true},
		{CenterOfGravityOscillator, dqs + centerOfGravityOscillator + dqs, true},
		{CyberCycle, dqs + cyberCycle + dqs, true},
		{InstantaneousTrendLine, dqs + instantaneousTrendLine + dqs, true},
		{SuperSmoother, dqs + superSmoother + dqs, true},
		{ZeroLagExponentialMovingAverage, dqs + zeroLagExponentialMovingAverage + dqs, true},
		{ZeroLagErrorCorrectingExponentialMovingAverage, dqs + zeroLagErrorCorrectingExponentialMovingAverage + dqs, true},
		{RoofingFilter, dqs + roofingFilter + dqs, true},
		{TrueRange, dqs + trueRange + dqs, true},
		{AverageTrueRange, dqs + averageTrueRange + dqs, true},
		{NormalizedAverageTrueRange, dqs + normalizedAverageTrueRange + dqs, true},
		{DirectionalMovementMinus, dqs + directionalMovementMinus + dqs, true},
		{DirectionalMovementPlus, dqs + directionalMovementPlus + dqs, true},
		{DirectionalIndicatorMinus, dqs + directionalIndicatorMinus + dqs, true},
		{DirectionalIndicatorPlus, dqs + directionalIndicatorPlus + dqs, true},
		{DirectionalMovementIndex, dqs + directionalMovementIndex + dqs, true},
		{AverageDirectionalMovementIndex, dqs + averageDirectionalMovementIndex + dqs, true},
		{AverageDirectionalMovementIndexRating, dqs + averageDirectionalMovementIndexRating + dqs, true},
		{WilliamsPercentR, dqs + williamsPercentR + dqs, true},
		{PercentagePriceOscillator, dqs + percentagePriceOscillator + dqs, true},
		{AbsolutePriceOscillator, dqs + absolutePriceOscillator + dqs, true},
		{CommodityChannelIndex, dqs + commodityChannelIndex + dqs, true},
		{MoneyFlowIndex, dqs + moneyFlowIndex + dqs, true},
		{OnBalanceVolume, dqs + onBalanceVolume + dqs, true},
		{BalanceOfPower, dqs + balanceOfPower + dqs, true},
		{RateOfChangeRatio, dqs + rateOfChangeRatio + dqs, true},
		{PearsonsCorrelationCoefficient, dqs + pearsonsCorrelationCoefficient + dqs, true},
		{LinearRegression, dqs + linearRegression + dqs, true},
		{UltimateOscillator, dqs + ultimateOscillator + dqs, true},
		{StochasticRelativeStrengthIndex, dqs + stochasticRelativeStrengthIndex + dqs, true},
		{Stochastic, dqs + stochastic + dqs, true},
		{Aroon, dqs + aroon + dqs, true},
		{AdvanceDecline, dqs + advanceDecline + dqs, true},
		{AdvanceDeclineOscillator, dqs + advanceDeclineOscillator + dqs, true},
		{ParabolicStopAndReverse, dqs + parabolicStopAndReverse + dqs, true},
		{TripleExponentialMovingAverageOscillator, dqs + tripleExponentialMovingAverageOscillator + dqs, true},
		{BollingerBandsTrend, dqs + bollingerBandsTrend + dqs, true},
		{MovingAverageConvergenceDivergence, dqs + movingAverageConvergenceDivergence + dqs, true},
		{SineWave, dqs + sineWave + dqs, true},
		{HilbertTransformerInstantaneousTrendLine, dqs + hilbertTransformerInstantaneousTrendLine + dqs, true},
		{TrendCycleMode, dqs + trendCycleMode + dqs, true},
		{CoronaSpectrum, dqs + coronaSpectrum + dqs, true},
		{CoronaSignalToNoiseRatio, dqs + coronaSignalToNoiseRatio + dqs, true},
		{CoronaSwingPosition, dqs + coronaSwingPosition + dqs, true},
		{CoronaTrendVigor, dqs + coronaTrendVigor + dqs, true},
		{AdaptiveTrendAndCycleFilter, dqs + adaptiveTrendAndCycleFilter + dqs, true},
		{MaximumEntropySpectrum, dqs + maximumEntropySpectrum + dqs, true},
		{DiscreteFourierTransformSpectrum, dqs + discreteFourierTransformSpectrum + dqs, true},
		{CombBandPassSpectrum, dqs + combBandPassSpectrum + dqs, true},
		{AutoCorrelationIndicator, dqs + autoCorrelationIndicator + dqs, true},
		{AutoCorrelationPeriodogram, dqs + autoCorrelationPeriodogram + dqs, true},
		{DominantCycle, dqs + dominantCycle + dqs, true},
		{last, nilstr, false},
		{Identifier(9999), nilstr, false},
		{Identifier(-9999), nilstr, false},
		{Identifier(0), nilstr, false},
	}

	for _, tt := range tests {
		exp := tt.json
		bs, err := tt.i.MarshalJSON()

		if err != nil && tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success '%v', got error %v", tt.i, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tt.i)

			continue
		}

		act := string(bs)
		if exp != act {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tt.i, exp, act)
		}
	}
}

func TestIdentifierUnmarshalJSON(t *testing.T) {
	t.Parallel()

	const dqs = "\""

	var zero Identifier
	tests := []struct {
		i         Identifier
		json      string
		succeeded bool
	}{
		{SimpleMovingAverage, dqs + simpleMovingAverage + dqs, true},
		{WeightedMovingAverage, dqs + weightedMovingAverage + dqs, true},
		{TriangularMovingAverage, dqs + triangularMovingAverage + dqs, true},
		{ExponentialMovingAverage, dqs + exponentialMovingAverage + dqs, true},
		{DoubleExponentialMovingAverage, dqs + doubleExponentialMovingAverage + dqs, true},
		{TripleExponentialMovingAverage, dqs + tripleExponentialMovingAverage + dqs, true},
		{T3ExponentialMovingAverage, dqs + t3ExponentialMovingAverage + dqs, true},
		{KaufmanAdaptiveMovingAverage, dqs + kaufmanAdaptiveMovingAverage + dqs, true},
		{JurikMovingAverage, dqs + jurikMovingAverage + dqs, true},
		{MesaAdaptiveMovingAverage, dqs + mesaAdaptiveMovingAverage + dqs, true},
		{FractalAdaptiveMovingAverage, dqs + fractalAdaptiveMovingAverage + dqs, true},
		{Momentum, dqs + momentum + dqs, true},
		{RateOfChange, dqs + rateOfChange + dqs, true},
		{RateOfChangePercent, dqs + rateOfChangePercent + dqs, true},
		{RelativeStrengthIndex, dqs + relativeStrengthIndex + dqs, true},
		{ChandeMomentumOscillator, dqs + chandeMomentumOscillator + dqs, true},
		{BollingerBands, dqs + bollingerBands + dqs, true},
		{Variance, dqs + variance + dqs, true},
		{StandardDeviation, dqs + standardDeviation + dqs, true},
		{GoertzelSpectrum, dqs + goertzelSpectrum + dqs, true},
		{CenterOfGravityOscillator, dqs + centerOfGravityOscillator + dqs, true},
		{CyberCycle, dqs + cyberCycle + dqs, true},
		{InstantaneousTrendLine, dqs + instantaneousTrendLine + dqs, true},
		{SuperSmoother, dqs + superSmoother + dqs, true},
		{ZeroLagExponentialMovingAverage, dqs + zeroLagExponentialMovingAverage + dqs, true},
		{ZeroLagErrorCorrectingExponentialMovingAverage, dqs + zeroLagErrorCorrectingExponentialMovingAverage + dqs, true},
		{RoofingFilter, dqs + roofingFilter + dqs, true},
		{TrueRange, dqs + trueRange + dqs, true},
		{AverageTrueRange, dqs + averageTrueRange + dqs, true},
		{NormalizedAverageTrueRange, dqs + normalizedAverageTrueRange + dqs, true},
		{DirectionalMovementMinus, dqs + directionalMovementMinus + dqs, true},
		{DirectionalMovementPlus, dqs + directionalMovementPlus + dqs, true},
		{DirectionalIndicatorMinus, dqs + directionalIndicatorMinus + dqs, true},
		{DirectionalIndicatorPlus, dqs + directionalIndicatorPlus + dqs, true},
		{DirectionalMovementIndex, dqs + directionalMovementIndex + dqs, true},
		{AverageDirectionalMovementIndex, dqs + averageDirectionalMovementIndex + dqs, true},
		{AverageDirectionalMovementIndexRating, dqs + averageDirectionalMovementIndexRating + dqs, true},
		{WilliamsPercentR, dqs + williamsPercentR + dqs, true},
		{PercentagePriceOscillator, dqs + percentagePriceOscillator + dqs, true},
		{AbsolutePriceOscillator, dqs + absolutePriceOscillator + dqs, true},
		{CommodityChannelIndex, dqs + commodityChannelIndex + dqs, true},
		{MoneyFlowIndex, dqs + moneyFlowIndex + dqs, true},
		{OnBalanceVolume, dqs + onBalanceVolume + dqs, true},
		{BalanceOfPower, dqs + balanceOfPower + dqs, true},
		{RateOfChangeRatio, dqs + rateOfChangeRatio + dqs, true},
		{PearsonsCorrelationCoefficient, dqs + pearsonsCorrelationCoefficient + dqs, true},
		{LinearRegression, dqs + linearRegression + dqs, true},
		{UltimateOscillator, dqs + ultimateOscillator + dqs, true},
		{StochasticRelativeStrengthIndex, dqs + stochasticRelativeStrengthIndex + dqs, true},
		{Stochastic, dqs + stochastic + dqs, true},
		{Aroon, dqs + aroon + dqs, true},
		{AdvanceDecline, dqs + advanceDecline + dqs, true},
		{AdvanceDeclineOscillator, dqs + advanceDeclineOscillator + dqs, true},
		{ParabolicStopAndReverse, dqs + parabolicStopAndReverse + dqs, true},
		{TripleExponentialMovingAverageOscillator, dqs + tripleExponentialMovingAverageOscillator + dqs, true},
		{BollingerBandsTrend, dqs + bollingerBandsTrend + dqs, true},
		{MovingAverageConvergenceDivergence, dqs + movingAverageConvergenceDivergence + dqs, true},
		{SineWave, dqs + sineWave + dqs, true},
		{HilbertTransformerInstantaneousTrendLine, dqs + hilbertTransformerInstantaneousTrendLine + dqs, true},
		{TrendCycleMode, dqs + trendCycleMode + dqs, true},
		{CoronaSpectrum, dqs + coronaSpectrum + dqs, true},
		{CoronaSignalToNoiseRatio, dqs + coronaSignalToNoiseRatio + dqs, true},
		{CoronaSwingPosition, dqs + coronaSwingPosition + dqs, true},
		{CoronaTrendVigor, dqs + coronaTrendVigor + dqs, true},
		{AdaptiveTrendAndCycleFilter, dqs + adaptiveTrendAndCycleFilter + dqs, true},
		{MaximumEntropySpectrum, dqs + maximumEntropySpectrum + dqs, true},
		{DiscreteFourierTransformSpectrum, dqs + discreteFourierTransformSpectrum + dqs, true},
		{CombBandPassSpectrum, dqs + combBandPassSpectrum + dqs, true},
		{AutoCorrelationIndicator, dqs + autoCorrelationIndicator + dqs, true},
		{AutoCorrelationPeriodogram, dqs + autoCorrelationPeriodogram + dqs, true},
		{DominantCycle, dqs + dominantCycle + dqs, true},
		{zero, "\"unknown\"", false},
		{zero, "\"foobar\"", false},
	}

	for _, tt := range tests {
		exp := tt.i
		bs := []byte(tt.json)

		var act Identifier

		err := act.UnmarshalJSON(bs)
		if err != nil && tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success '%v', got error %v", tt.json, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("MarshalJSON('%v'): expected error, got success", tt.json)

			continue
		}

		if exp != act {
			t.Errorf("MarshalJSON('%v'): expected '%v', actual '%v'", tt.json, exp, act)
		}
	}
}
