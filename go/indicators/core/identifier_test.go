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
		// ── common ────────────────────────────────────────────────────────────
		{AbsolutePriceOscillator, absolutePriceOscillator},
		{ExponentialMovingAverage, exponentialMovingAverage},
		{LinearRegression, linearRegression},
		{Momentum, momentum},
		{PearsonsCorrelationCoefficient, pearsonsCorrelationCoefficient},
		{RateOfChange, rateOfChange},
		{RateOfChangePercent, rateOfChangePercent},
		{RateOfChangeRatio, rateOfChangeRatio},
		{SimpleMovingAverage, simpleMovingAverage},
		{StandardDeviation, standardDeviation},
		{TriangularMovingAverage, triangularMovingAverage},
		{Variance, variance},
		{WeightedMovingAverage, weightedMovingAverage},
		// ── arnaud legoux ──────────────────────────────────────────────────────
		{ArnaudLegouxMovingAverage, arnaudLegouxMovingAverage},
		// ── donald lambert ─────────────────────────────────────────────────────
		{CommodityChannelIndex, commodityChannelIndex},
		// ── gene quong ─────────────────────────────────────────────────────────
		{MoneyFlowIndex, moneyFlowIndex},
		// ── george lane ────────────────────────────────────────────────────────
		{Stochastic, stochastic},
		// ── gerald appel ───────────────────────────────────────────────────────
		{MovingAverageConvergenceDivergence, movingAverageConvergenceDivergence},
		{PercentagePriceOscillator, percentagePriceOscillator},
		// ── igor livshin ───────────────────────────────────────────────────────
		{BalanceOfPower, balanceOfPower},
		// ── jack hutson ────────────────────────────────────────────────────────
		{TripleExponentialMovingAverageOscillator, tripleExponentialMovingAverageOscillator},
		// ── john bollinger ─────────────────────────────────────────────────────
		{BollingerBands, bollingerBands},
		{BollingerBandsTrend, bollingerBandsTrend},
		// ── john ehlers ────────────────────────────────────────────────────────
		{AutoCorrelationIndicator, autoCorrelationIndicator},
		{AutoCorrelationPeriodogram, autoCorrelationPeriodogram},
		{CenterOfGravityOscillator, centerOfGravityOscillator},
		{CombBandPassSpectrum, combBandPassSpectrum},
		{CoronaSignalToNoiseRatio, coronaSignalToNoiseRatio},
		{CoronaSpectrum, coronaSpectrum},
		{CoronaSwingPosition, coronaSwingPosition},
		{CoronaTrendVigor, coronaTrendVigor},
		{CyberCycle, cyberCycle},
		{DiscreteFourierTransformSpectrum, discreteFourierTransformSpectrum},
		{DominantCycle, dominantCycle},
		{FractalAdaptiveMovingAverage, fractalAdaptiveMovingAverage},
		{HilbertTransformerInstantaneousTrendLine, hilbertTransformerInstantaneousTrendLine},
		{InstantaneousTrendLine, instantaneousTrendLine},
		{MesaAdaptiveMovingAverage, mesaAdaptiveMovingAverage},
		{RoofingFilter, roofingFilter},
		{SineWave, sineWave},
		{SuperSmoother, superSmoother},
		{TrendCycleMode, trendCycleMode},
		{ZeroLagErrorCorrectingExponentialMovingAverage, zeroLagErrorCorrectingExponentialMovingAverage},
		{ZeroLagExponentialMovingAverage, zeroLagExponentialMovingAverage},
		// ── joseph granville ───────────────────────────────────────────────────
		{OnBalanceVolume, onBalanceVolume},
		// ── larry williams ─────────────────────────────────────────────────────
		{UltimateOscillator, ultimateOscillator},
		{WilliamsPercentR, williamsPercentR},
		// ── manfred durschner ──────────────────────────────────────────────────
		{NewMovingAverage, newMovingAverage},
		// ── marc chaikin ───────────────────────────────────────────────────────
		{AdvanceDecline, advanceDecline},
		{AdvanceDeclineOscillator, advanceDeclineOscillator},
		// ── mark jurik ─────────────────────────────────────────────────────────
		{JurikAdaptiveRelativeTrendStrengthIndex, jurikAdaptiveRelativeTrendStrengthIndex},
		{JurikAdaptiveZeroLagVelocity, jurikAdaptiveZeroLagVelocity},
		{JurikCommodityChannelIndex, jurikCommodityChannelIndex},
		{JurikCompositeFractalBehaviorIndex, jurikCompositeFractalBehaviorIndex},
		{JurikDirectionalMovementIndex, jurikDirectionalMovementIndex},
		{JurikFractalAdaptiveZeroLagVelocity, jurikFractalAdaptiveZeroLagVelocity},
		{JurikMovingAverage, jurikMovingAverage},
		{JurikRelativeTrendStrengthIndex, jurikRelativeTrendStrengthIndex},
		{JurikTurningPointOscillator, jurikTurningPointOscillator},
		{JurikWaveletSampler, jurikWaveletSampler},
		{JurikZeroLagVelocity, jurikZeroLagVelocity},
		// ── patrick mulloy ─────────────────────────────────────────────────────
		{DoubleExponentialMovingAverage, doubleExponentialMovingAverage},
		{TripleExponentialMovingAverage, tripleExponentialMovingAverage},
		// ── perry kaufman ──────────────────────────────────────────────────────
		{KaufmanAdaptiveMovingAverage, kaufmanAdaptiveMovingAverage},
		// ── tim tillson ────────────────────────────────────────────────────────
		{T2ExponentialMovingAverage, t2ExponentialMovingAverage},
		{T3ExponentialMovingAverage, t3ExponentialMovingAverage},
		// ── tushar chande ──────────────────────────────────────────────────────
		{Aroon, aroon},
		{ChandeMomentumOscillator, chandeMomentumOscillator},
		{StochasticRelativeStrengthIndex, stochasticRelativeStrengthIndex},
		// ── vladimir kravchuk ──────────────────────────────────────────────────
		{AdaptiveTrendAndCycleFilter, adaptiveTrendAndCycleFilter},
		// ── welles wilder ──────────────────────────────────────────────────────
		{AverageDirectionalMovementIndex, averageDirectionalMovementIndex},
		{AverageDirectionalMovementIndexRating, averageDirectionalMovementIndexRating},
		{AverageTrueRange, averageTrueRange},
		{DirectionalIndicatorMinus, directionalIndicatorMinus},
		{DirectionalIndicatorPlus, directionalIndicatorPlus},
		{DirectionalMovementIndex, directionalMovementIndex},
		{DirectionalMovementMinus, directionalMovementMinus},
		{DirectionalMovementPlus, directionalMovementPlus},
		{NormalizedAverageTrueRange, normalizedAverageTrueRange},
		{ParabolicStopAndReverse, parabolicStopAndReverse},
		{RelativeStrengthIndex, relativeStrengthIndex},
		{TrueRange, trueRange},
		// ── custom ────────────────────────────────────────────────────────────
		{GoertzelSpectrum, goertzelSpectrum},
		{MaximumEntropySpectrum, maximumEntropySpectrum},
		// ── boundary ──────────────────────────────────────────────────────────
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
		// ── common ────────────────────────────────────────────────────────────
		{AbsolutePriceOscillator, false},
		{ExponentialMovingAverage, true},
		{LinearRegression, true},
		{Momentum, true},
		{PearsonsCorrelationCoefficient, true},
		{RateOfChange, true},
		{RateOfChangePercent, true},
		{RateOfChangeRatio, true},
		{SimpleMovingAverage, true},
		{StandardDeviation, true},
		{TriangularMovingAverage, true},
		{Variance, true},
		{WeightedMovingAverage, true},
		// ── arnaud legoux ──────────────────────────────────────────────────────
		{ArnaudLegouxMovingAverage, true},
		// ── donald lambert ─────────────────────────────────────────────────────
		{CommodityChannelIndex, true},
		// ── gene quong ─────────────────────────────────────────────────────────
		{MoneyFlowIndex, true},
		// ── george lane ────────────────────────────────────────────────────────
		{Stochastic, true},
		// ── gerald appel ───────────────────────────────────────────────────────
		{MovingAverageConvergenceDivergence, true},
		{PercentagePriceOscillator, true},
		// ── igor livshin ───────────────────────────────────────────────────────
		{BalanceOfPower, true},
		// ── jack hutson ────────────────────────────────────────────────────────
		{TripleExponentialMovingAverageOscillator, true},
		// ── john bollinger ─────────────────────────────────────────────────────
		{BollingerBands, true},
		{BollingerBandsTrend, true},
		// ── john ehlers ────────────────────────────────────────────────────────
		{AutoCorrelationIndicator, true},
		{AutoCorrelationPeriodogram, true},
		{CenterOfGravityOscillator, true},
		{CombBandPassSpectrum, true},
		{CoronaSignalToNoiseRatio, true},
		{CoronaSpectrum, true},
		{CoronaSwingPosition, true},
		{CoronaTrendVigor, true},
		{CyberCycle, true},
		{DiscreteFourierTransformSpectrum, true},
		{DominantCycle, true},
		{FractalAdaptiveMovingAverage, true},
		{HilbertTransformerInstantaneousTrendLine, true},
		{InstantaneousTrendLine, true},
		{MesaAdaptiveMovingAverage, true},
		{RoofingFilter, true},
		{SineWave, true},
		{SuperSmoother, true},
		{TrendCycleMode, true},
		{ZeroLagErrorCorrectingExponentialMovingAverage, true},
		{ZeroLagExponentialMovingAverage, true},
		// ── joseph granville ───────────────────────────────────────────────────
		{OnBalanceVolume, true},
		// ── larry williams ─────────────────────────────────────────────────────
		{UltimateOscillator, true},
		{WilliamsPercentR, true},
		// ── manfred durschner ──────────────────────────────────────────────────
		{NewMovingAverage, true},
		// ── marc chaikin ───────────────────────────────────────────────────────
		{AdvanceDecline, true},
		{AdvanceDeclineOscillator, true},
		// ── mark jurik ─────────────────────────────────────────────────────────
		{JurikAdaptiveRelativeTrendStrengthIndex, true},
		{JurikAdaptiveZeroLagVelocity, true},
		{JurikCommodityChannelIndex, true},
		{JurikCompositeFractalBehaviorIndex, true},
		{JurikDirectionalMovementIndex, true},
		{JurikFractalAdaptiveZeroLagVelocity, true},
		{JurikMovingAverage, true},
		{JurikRelativeTrendStrengthIndex, true},
		{JurikTurningPointOscillator, true},
		{JurikWaveletSampler, true},
		{JurikZeroLagVelocity, true},
		// ── patrick mulloy ─────────────────────────────────────────────────────
		{DoubleExponentialMovingAverage, true},
		{TripleExponentialMovingAverage, true},
		// ── perry kaufman ──────────────────────────────────────────────────────
		{KaufmanAdaptiveMovingAverage, true},
		// ── tim tillson ────────────────────────────────────────────────────────
		{T2ExponentialMovingAverage, true},
		{T3ExponentialMovingAverage, true},
		// ── tushar chande ──────────────────────────────────────────────────────
		{Aroon, true},
		{ChandeMomentumOscillator, true},
		{StochasticRelativeStrengthIndex, true},
		// ── vladimir kravchuk ──────────────────────────────────────────────────
		{AdaptiveTrendAndCycleFilter, true},
		// ── welles wilder ──────────────────────────────────────────────────────
		{AverageDirectionalMovementIndex, true},
		{AverageDirectionalMovementIndexRating, true},
		{AverageTrueRange, true},
		{DirectionalIndicatorMinus, true},
		{DirectionalIndicatorPlus, true},
		{DirectionalMovementIndex, true},
		{DirectionalMovementMinus, true},
		{DirectionalMovementPlus, true},
		{NormalizedAverageTrueRange, true},
		{ParabolicStopAndReverse, true},
		{RelativeStrengthIndex, true},
		{TrueRange, true},
		// ── custom ────────────────────────────────────────────────────────────
		{GoertzelSpectrum, true},
		{MaximumEntropySpectrum, true},
		// ── boundary ──────────────────────────────────────────────────────────
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
		// ── common ────────────────────────────────────────────────────────────
		{AbsolutePriceOscillator, dqs + absolutePriceOscillator + dqs, true},
		{ExponentialMovingAverage, dqs + exponentialMovingAverage + dqs, true},
		{LinearRegression, dqs + linearRegression + dqs, true},
		{Momentum, dqs + momentum + dqs, true},
		{PearsonsCorrelationCoefficient, dqs + pearsonsCorrelationCoefficient + dqs, true},
		{RateOfChange, dqs + rateOfChange + dqs, true},
		{RateOfChangePercent, dqs + rateOfChangePercent + dqs, true},
		{RateOfChangeRatio, dqs + rateOfChangeRatio + dqs, true},
		{SimpleMovingAverage, dqs + simpleMovingAverage + dqs, true},
		{StandardDeviation, dqs + standardDeviation + dqs, true},
		{TriangularMovingAverage, dqs + triangularMovingAverage + dqs, true},
		{Variance, dqs + variance + dqs, true},
		{WeightedMovingAverage, dqs + weightedMovingAverage + dqs, true},
		// ── arnaud legoux ──────────────────────────────────────────────────────
		{ArnaudLegouxMovingAverage, dqs + arnaudLegouxMovingAverage + dqs, true},
		// ── donald lambert ─────────────────────────────────────────────────────
		{CommodityChannelIndex, dqs + commodityChannelIndex + dqs, true},
		// ── gene quong ─────────────────────────────────────────────────────────
		{MoneyFlowIndex, dqs + moneyFlowIndex + dqs, true},
		// ── george lane ────────────────────────────────────────────────────────
		{Stochastic, dqs + stochastic + dqs, true},
		// ── gerald appel ───────────────────────────────────────────────────────
		{MovingAverageConvergenceDivergence, dqs + movingAverageConvergenceDivergence + dqs, true},
		{PercentagePriceOscillator, dqs + percentagePriceOscillator + dqs, true},
		// ── igor livshin ───────────────────────────────────────────────────────
		{BalanceOfPower, dqs + balanceOfPower + dqs, true},
		// ── jack hutson ────────────────────────────────────────────────────────
		{TripleExponentialMovingAverageOscillator, dqs + tripleExponentialMovingAverageOscillator + dqs, true},
		// ── john bollinger ─────────────────────────────────────────────────────
		{BollingerBands, dqs + bollingerBands + dqs, true},
		{BollingerBandsTrend, dqs + bollingerBandsTrend + dqs, true},
		// ── john ehlers ────────────────────────────────────────────────────────
		{AutoCorrelationIndicator, dqs + autoCorrelationIndicator + dqs, true},
		{AutoCorrelationPeriodogram, dqs + autoCorrelationPeriodogram + dqs, true},
		{CenterOfGravityOscillator, dqs + centerOfGravityOscillator + dqs, true},
		{CombBandPassSpectrum, dqs + combBandPassSpectrum + dqs, true},
		{CoronaSignalToNoiseRatio, dqs + coronaSignalToNoiseRatio + dqs, true},
		{CoronaSpectrum, dqs + coronaSpectrum + dqs, true},
		{CoronaSwingPosition, dqs + coronaSwingPosition + dqs, true},
		{CoronaTrendVigor, dqs + coronaTrendVigor + dqs, true},
		{CyberCycle, dqs + cyberCycle + dqs, true},
		{DiscreteFourierTransformSpectrum, dqs + discreteFourierTransformSpectrum + dqs, true},
		{DominantCycle, dqs + dominantCycle + dqs, true},
		{FractalAdaptiveMovingAverage, dqs + fractalAdaptiveMovingAverage + dqs, true},
		{HilbertTransformerInstantaneousTrendLine, dqs + hilbertTransformerInstantaneousTrendLine + dqs, true},
		{InstantaneousTrendLine, dqs + instantaneousTrendLine + dqs, true},
		{MesaAdaptiveMovingAverage, dqs + mesaAdaptiveMovingAverage + dqs, true},
		{RoofingFilter, dqs + roofingFilter + dqs, true},
		{SineWave, dqs + sineWave + dqs, true},
		{SuperSmoother, dqs + superSmoother + dqs, true},
		{TrendCycleMode, dqs + trendCycleMode + dqs, true},
		{ZeroLagErrorCorrectingExponentialMovingAverage, dqs + zeroLagErrorCorrectingExponentialMovingAverage + dqs, true},
		{ZeroLagExponentialMovingAverage, dqs + zeroLagExponentialMovingAverage + dqs, true},
		// ── joseph granville ───────────────────────────────────────────────────
		{OnBalanceVolume, dqs + onBalanceVolume + dqs, true},
		// ── larry williams ─────────────────────────────────────────────────────
		{UltimateOscillator, dqs + ultimateOscillator + dqs, true},
		{WilliamsPercentR, dqs + williamsPercentR + dqs, true},
		// ── manfred durschner ──────────────────────────────────────────────────
		{NewMovingAverage, dqs + newMovingAverage + dqs, true},
		// ── marc chaikin ───────────────────────────────────────────────────────
		{AdvanceDecline, dqs + advanceDecline + dqs, true},
		{AdvanceDeclineOscillator, dqs + advanceDeclineOscillator + dqs, true},
		// ── mark jurik ─────────────────────────────────────────────────────────
		{JurikAdaptiveRelativeTrendStrengthIndex, dqs + jurikAdaptiveRelativeTrendStrengthIndex + dqs, true},
		{JurikAdaptiveZeroLagVelocity, dqs + jurikAdaptiveZeroLagVelocity + dqs, true},
		{JurikCommodityChannelIndex, dqs + jurikCommodityChannelIndex + dqs, true},
		{JurikCompositeFractalBehaviorIndex, dqs + jurikCompositeFractalBehaviorIndex + dqs, true},
		{JurikDirectionalMovementIndex, dqs + jurikDirectionalMovementIndex + dqs, true},
		{JurikFractalAdaptiveZeroLagVelocity, dqs + jurikFractalAdaptiveZeroLagVelocity + dqs, true},
		{JurikMovingAverage, dqs + jurikMovingAverage + dqs, true},
		{JurikRelativeTrendStrengthIndex, dqs + jurikRelativeTrendStrengthIndex + dqs, true},
		{JurikTurningPointOscillator, dqs + jurikTurningPointOscillator + dqs, true},
		{JurikWaveletSampler, dqs + jurikWaveletSampler + dqs, true},
		{JurikZeroLagVelocity, dqs + jurikZeroLagVelocity + dqs, true},
		// ── patrick mulloy ─────────────────────────────────────────────────────
		{DoubleExponentialMovingAverage, dqs + doubleExponentialMovingAverage + dqs, true},
		{TripleExponentialMovingAverage, dqs + tripleExponentialMovingAverage + dqs, true},
		// ── perry kaufman ──────────────────────────────────────────────────────
		{KaufmanAdaptiveMovingAverage, dqs + kaufmanAdaptiveMovingAverage + dqs, true},
		// ── tim tillson ────────────────────────────────────────────────────────
		{T2ExponentialMovingAverage, dqs + t2ExponentialMovingAverage + dqs, true},
		{T3ExponentialMovingAverage, dqs + t3ExponentialMovingAverage + dqs, true},
		// ── tushar chande ──────────────────────────────────────────────────────
		{Aroon, dqs + aroon + dqs, true},
		{ChandeMomentumOscillator, dqs + chandeMomentumOscillator + dqs, true},
		{StochasticRelativeStrengthIndex, dqs + stochasticRelativeStrengthIndex + dqs, true},
		// ── vladimir kravchuk ──────────────────────────────────────────────────
		{AdaptiveTrendAndCycleFilter, dqs + adaptiveTrendAndCycleFilter + dqs, true},
		// ── welles wilder ──────────────────────────────────────────────────────
		{AverageDirectionalMovementIndex, dqs + averageDirectionalMovementIndex + dqs, true},
		{AverageDirectionalMovementIndexRating, dqs + averageDirectionalMovementIndexRating + dqs, true},
		{AverageTrueRange, dqs + averageTrueRange + dqs, true},
		{DirectionalIndicatorMinus, dqs + directionalIndicatorMinus + dqs, true},
		{DirectionalIndicatorPlus, dqs + directionalIndicatorPlus + dqs, true},
		{DirectionalMovementIndex, dqs + directionalMovementIndex + dqs, true},
		{DirectionalMovementMinus, dqs + directionalMovementMinus + dqs, true},
		{DirectionalMovementPlus, dqs + directionalMovementPlus + dqs, true},
		{NormalizedAverageTrueRange, dqs + normalizedAverageTrueRange + dqs, true},
		{ParabolicStopAndReverse, dqs + parabolicStopAndReverse + dqs, true},
		{RelativeStrengthIndex, dqs + relativeStrengthIndex + dqs, true},
		{TrueRange, dqs + trueRange + dqs, true},
		// ── custom ────────────────────────────────────────────────────────────
		{GoertzelSpectrum, dqs + goertzelSpectrum + dqs, true},
		{MaximumEntropySpectrum, dqs + maximumEntropySpectrum + dqs, true},
		// ── boundary ──────────────────────────────────────────────────────────
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
		// ── common ────────────────────────────────────────────────────────────
		{AbsolutePriceOscillator, dqs + absolutePriceOscillator + dqs, true},
		{ExponentialMovingAverage, dqs + exponentialMovingAverage + dqs, true},
		{LinearRegression, dqs + linearRegression + dqs, true},
		{Momentum, dqs + momentum + dqs, true},
		{PearsonsCorrelationCoefficient, dqs + pearsonsCorrelationCoefficient + dqs, true},
		{RateOfChange, dqs + rateOfChange + dqs, true},
		{RateOfChangePercent, dqs + rateOfChangePercent + dqs, true},
		{RateOfChangeRatio, dqs + rateOfChangeRatio + dqs, true},
		{SimpleMovingAverage, dqs + simpleMovingAverage + dqs, true},
		{StandardDeviation, dqs + standardDeviation + dqs, true},
		{TriangularMovingAverage, dqs + triangularMovingAverage + dqs, true},
		{Variance, dqs + variance + dqs, true},
		{WeightedMovingAverage, dqs + weightedMovingAverage + dqs, true},
		// ── arnaud legoux ──────────────────────────────────────────────────────
		{ArnaudLegouxMovingAverage, dqs + arnaudLegouxMovingAverage + dqs, true},
		// ── donald lambert ─────────────────────────────────────────────────────
		{CommodityChannelIndex, dqs + commodityChannelIndex + dqs, true},
		// ── gene quong ─────────────────────────────────────────────────────────
		{MoneyFlowIndex, dqs + moneyFlowIndex + dqs, true},
		// ── george lane ────────────────────────────────────────────────────────
		{Stochastic, dqs + stochastic + dqs, true},
		// ── gerald appel ───────────────────────────────────────────────────────
		{MovingAverageConvergenceDivergence, dqs + movingAverageConvergenceDivergence + dqs, true},
		{PercentagePriceOscillator, dqs + percentagePriceOscillator + dqs, true},
		// ── igor livshin ───────────────────────────────────────────────────────
		{BalanceOfPower, dqs + balanceOfPower + dqs, true},
		// ── jack hutson ────────────────────────────────────────────────────────
		{TripleExponentialMovingAverageOscillator, dqs + tripleExponentialMovingAverageOscillator + dqs, true},
		// ── john bollinger ─────────────────────────────────────────────────────
		{BollingerBands, dqs + bollingerBands + dqs, true},
		{BollingerBandsTrend, dqs + bollingerBandsTrend + dqs, true},
		// ── john ehlers ────────────────────────────────────────────────────────
		{AutoCorrelationIndicator, dqs + autoCorrelationIndicator + dqs, true},
		{AutoCorrelationPeriodogram, dqs + autoCorrelationPeriodogram + dqs, true},
		{CenterOfGravityOscillator, dqs + centerOfGravityOscillator + dqs, true},
		{CombBandPassSpectrum, dqs + combBandPassSpectrum + dqs, true},
		{CoronaSignalToNoiseRatio, dqs + coronaSignalToNoiseRatio + dqs, true},
		{CoronaSpectrum, dqs + coronaSpectrum + dqs, true},
		{CoronaSwingPosition, dqs + coronaSwingPosition + dqs, true},
		{CoronaTrendVigor, dqs + coronaTrendVigor + dqs, true},
		{CyberCycle, dqs + cyberCycle + dqs, true},
		{DiscreteFourierTransformSpectrum, dqs + discreteFourierTransformSpectrum + dqs, true},
		{DominantCycle, dqs + dominantCycle + dqs, true},
		{FractalAdaptiveMovingAverage, dqs + fractalAdaptiveMovingAverage + dqs, true},
		{HilbertTransformerInstantaneousTrendLine, dqs + hilbertTransformerInstantaneousTrendLine + dqs, true},
		{InstantaneousTrendLine, dqs + instantaneousTrendLine + dqs, true},
		{MesaAdaptiveMovingAverage, dqs + mesaAdaptiveMovingAverage + dqs, true},
		{RoofingFilter, dqs + roofingFilter + dqs, true},
		{SineWave, dqs + sineWave + dqs, true},
		{SuperSmoother, dqs + superSmoother + dqs, true},
		{TrendCycleMode, dqs + trendCycleMode + dqs, true},
		{ZeroLagErrorCorrectingExponentialMovingAverage, dqs + zeroLagErrorCorrectingExponentialMovingAverage + dqs, true},
		{ZeroLagExponentialMovingAverage, dqs + zeroLagExponentialMovingAverage + dqs, true},
		// ── joseph granville ───────────────────────────────────────────────────
		{OnBalanceVolume, dqs + onBalanceVolume + dqs, true},
		// ── larry williams ─────────────────────────────────────────────────────
		{UltimateOscillator, dqs + ultimateOscillator + dqs, true},
		{WilliamsPercentR, dqs + williamsPercentR + dqs, true},
		// ── manfred durschner ──────────────────────────────────────────────────
		{NewMovingAverage, dqs + newMovingAverage + dqs, true},
		// ── marc chaikin ───────────────────────────────────────────────────────
		{AdvanceDecline, dqs + advanceDecline + dqs, true},
		{AdvanceDeclineOscillator, dqs + advanceDeclineOscillator + dqs, true},
		// ── mark jurik ─────────────────────────────────────────────────────────
		{JurikAdaptiveRelativeTrendStrengthIndex, dqs + jurikAdaptiveRelativeTrendStrengthIndex + dqs, true},
		{JurikAdaptiveZeroLagVelocity, dqs + jurikAdaptiveZeroLagVelocity + dqs, true},
		{JurikCommodityChannelIndex, dqs + jurikCommodityChannelIndex + dqs, true},
		{JurikCompositeFractalBehaviorIndex, dqs + jurikCompositeFractalBehaviorIndex + dqs, true},
		{JurikDirectionalMovementIndex, dqs + jurikDirectionalMovementIndex + dqs, true},
		{JurikFractalAdaptiveZeroLagVelocity, dqs + jurikFractalAdaptiveZeroLagVelocity + dqs, true},
		{JurikMovingAverage, dqs + jurikMovingAverage + dqs, true},
		{JurikRelativeTrendStrengthIndex, dqs + jurikRelativeTrendStrengthIndex + dqs, true},
		{JurikTurningPointOscillator, dqs + jurikTurningPointOscillator + dqs, true},
		{JurikWaveletSampler, dqs + jurikWaveletSampler + dqs, true},
		{JurikZeroLagVelocity, dqs + jurikZeroLagVelocity + dqs, true},
		// ── patrick mulloy ─────────────────────────────────────────────────────
		{DoubleExponentialMovingAverage, dqs + doubleExponentialMovingAverage + dqs, true},
		{TripleExponentialMovingAverage, dqs + tripleExponentialMovingAverage + dqs, true},
		// ── perry kaufman ──────────────────────────────────────────────────────
		{KaufmanAdaptiveMovingAverage, dqs + kaufmanAdaptiveMovingAverage + dqs, true},
		// ── tim tillson ────────────────────────────────────────────────────────
		{T2ExponentialMovingAverage, dqs + t2ExponentialMovingAverage + dqs, true},
		{T3ExponentialMovingAverage, dqs + t3ExponentialMovingAverage + dqs, true},
		// ── tushar chande ──────────────────────────────────────────────────────
		{Aroon, dqs + aroon + dqs, true},
		{ChandeMomentumOscillator, dqs + chandeMomentumOscillator + dqs, true},
		{StochasticRelativeStrengthIndex, dqs + stochasticRelativeStrengthIndex + dqs, true},
		// ── vladimir kravchuk ──────────────────────────────────────────────────
		{AdaptiveTrendAndCycleFilter, dqs + adaptiveTrendAndCycleFilter + dqs, true},
		// ── welles wilder ──────────────────────────────────────────────────────
		{AverageDirectionalMovementIndex, dqs + averageDirectionalMovementIndex + dqs, true},
		{AverageDirectionalMovementIndexRating, dqs + averageDirectionalMovementIndexRating + dqs, true},
		{AverageTrueRange, dqs + averageTrueRange + dqs, true},
		{DirectionalIndicatorMinus, dqs + directionalIndicatorMinus + dqs, true},
		{DirectionalIndicatorPlus, dqs + directionalIndicatorPlus + dqs, true},
		{DirectionalMovementIndex, dqs + directionalMovementIndex + dqs, true},
		{DirectionalMovementMinus, dqs + directionalMovementMinus + dqs, true},
		{DirectionalMovementPlus, dqs + directionalMovementPlus + dqs, true},
		{NormalizedAverageTrueRange, dqs + normalizedAverageTrueRange + dqs, true},
		{ParabolicStopAndReverse, dqs + parabolicStopAndReverse + dqs, true},
		{RelativeStrengthIndex, dqs + relativeStrengthIndex + dqs, true},
		{TrueRange, dqs + trueRange + dqs, true},
		// ── custom ────────────────────────────────────────────────────────────
		{GoertzelSpectrum, dqs + goertzelSpectrum + dqs, true},
		{MaximumEntropySpectrum, dqs + maximumEntropySpectrum + dqs, true},
		// ── boundary ──────────────────────────────────────────────────────────
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
