// Package factory provides a generic constructor that maps a core.Identifier
// and a JSON parameter string to a core.Indicator instance. This avoids the
// need for callers to import individual indicator packages directly.
package factory

import (
	"encoding/json"
	"fmt"

	"zpano/indicators/common/absolutepriceoscillator"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/linearregression"
	"zpano/indicators/common/momentum"
	"zpano/indicators/common/pearsonscorrelationcoefficient"
	"zpano/indicators/common/rateofchange"
	"zpano/indicators/common/rateofchangepercent"
	"zpano/indicators/common/rateofchangeratio"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/common/standarddeviation"
	"zpano/indicators/common/triangularmovingaverage"
	"zpano/indicators/common/variance"
	"zpano/indicators/common/weightedmovingaverage"
	"zpano/indicators/core"
	"zpano/indicators/custom/goertzelspectrum"
	"zpano/indicators/custom/maximumentropyspectrum"
	"zpano/indicators/donaldlambert/commoditychannelindex"
	"zpano/indicators/genequong/moneyflowindex"
	"zpano/indicators/georgelane/stochastic"
	"zpano/indicators/geraldappel/movingaverageconvergencedivergence"
	"zpano/indicators/geraldappel/percentagepriceoscillator"
	"zpano/indicators/igorlivshin/balanceofpower"
	"zpano/indicators/jackhutson/tripleexponentialmovingaverageoscillator"
	"zpano/indicators/johnbollinger/bollingerbands"
	"zpano/indicators/johnbollinger/bollingerbandstrend"
	"zpano/indicators/johnehlers/autocorrelationindicator"
	"zpano/indicators/johnehlers/autocorrelationperiodogram"
	"zpano/indicators/johnehlers/centerofgravityoscillator"
	"zpano/indicators/johnehlers/combbandpassspectrum"
	"zpano/indicators/johnehlers/coronasignaltonoiseratio"
	"zpano/indicators/johnehlers/coronaspectrum"
	"zpano/indicators/johnehlers/coronaswingposition"
	"zpano/indicators/johnehlers/coronatrendvigor"
	"zpano/indicators/johnehlers/cybercycle"
	"zpano/indicators/johnehlers/discretefouriertransformspectrum"
	"zpano/indicators/johnehlers/dominantcycle"
	"zpano/indicators/johnehlers/fractaladaptivemovingaverage"
	"zpano/indicators/johnehlers/hilberttransformerinstantaneoustrendline"
	"zpano/indicators/johnehlers/instantaneoustrendline"
	"zpano/indicators/johnehlers/mesaadaptivemovingaverage"
	"zpano/indicators/johnehlers/roofingfilter"
	"zpano/indicators/johnehlers/sinewave"
	"zpano/indicators/johnehlers/supersmoother"
	"zpano/indicators/johnehlers/trendcyclemode"
	"zpano/indicators/johnehlers/zerolagerrorcorrectingexponentialmovingaverage"
	"zpano/indicators/johnehlers/zerolagexponentialmovingaverage"
	"zpano/indicators/josephgranville/onbalancevolume"
	"zpano/indicators/larrywilliams/ultimateoscillator"
	"zpano/indicators/larrywilliams/williamspercentr"
	"zpano/indicators/marcchaikin/advancedecline"
	"zpano/indicators/marcchaikin/advancedeclineoscillator"
	"zpano/indicators/markjurik/jurikcompositefractalbehaviorindex"
	"zpano/indicators/markjurik/jurikdirectionalmovementindex"
	"zpano/indicators/markjurik/jurikmovingaverage"
	"zpano/indicators/markjurik/jurikrelativetrendstrengthindex"
	"zpano/indicators/markjurik/jurikzerolagvelocity"
	"zpano/indicators/patrickmulloy/doubleexponentialmovingaverage"
	"zpano/indicators/patrickmulloy/tripleexponentialmovingaverage"
	"zpano/indicators/perrykaufman/kaufmanadaptivemovingaverage"
	"zpano/indicators/timtillson/t2exponentialmovingaverage"
	"zpano/indicators/timtillson/t3exponentialmovingaverage"
	"zpano/indicators/tusharchande/aroon"
	"zpano/indicators/tusharchande/chandemomentumoscillator"
	"zpano/indicators/tusharchande/stochasticrelativestrengthindex"
	"zpano/indicators/vladimirkravchuk/adaptivetrendandcyclefilter"
	"zpano/indicators/welleswilder/averagedirectionalmovementindex"
	"zpano/indicators/welleswilder/averagedirectionalmovementindexrating"
	"zpano/indicators/welleswilder/averagetruerange"
	"zpano/indicators/welleswilder/directionalindicatorminus"
	"zpano/indicators/welleswilder/directionalindicatorplus"
	"zpano/indicators/welleswilder/directionalmovementindex"
	"zpano/indicators/welleswilder/directionalmovementminus"
	"zpano/indicators/welleswilder/directionalmovementplus"
	"zpano/indicators/welleswilder/normalizedaveragetruerange"
	"zpano/indicators/welleswilder/parabolicstopandreverse"
	"zpano/indicators/welleswilder/relativestrengthindex"
	"zpano/indicators/welleswilder/truerange"
)

// lengthParam is a helper struct for indicators that take a bare int length.
type lengthParam struct {
	Length int `json:"length"`
}

// hasKey checks if a JSON object contains a specific key.
func hasKey(data []byte, key string) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return false
	}

	_, ok := m[key]

	return ok
}

// unmarshal is a helper that returns a formatted error on failure.
func unmarshal(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse params: %w", err)
	}

	return nil
}

// New creates an indicator from its identifier and a JSON-encoded parameter
// string. If params is empty or "null", default parameters are used.
//
// For indicators with Length and SmoothingFactor constructor variants, the
// factory auto-detects which to use: if the JSON contains a "smoothingFactor"
// key the SmoothingFactor variant is used, otherwise the Length variant.
//
//nolint:cyclop,funlen,gocyclo // large switch is intentional — one case per identifier.
func New(identifier core.Identifier, params string) (core.Indicator, error) {
	b := []byte(params)
	if len(b) == 0 {
		b = []byte("{}")
	}

	switch identifier {
	// ── common ────────────────────────────────────────────────────────────

	case core.SimpleMovingAverage:
		p := simplemovingaverage.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return simplemovingaverage.NewSimpleMovingAverage(p)

	case core.WeightedMovingAverage:
		p := weightedmovingaverage.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return weightedmovingaverage.NewWeightedMovingAverage(p)

	case core.TriangularMovingAverage:
		p := triangularmovingaverage.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return triangularmovingaverage.NewTriangularMovingAverage(p)

	case core.ExponentialMovingAverage:
		if hasKey(b, "smoothingFactor") {
			p := &exponentialmovingaverage.ExponentialMovingAverageSmoothingFactorParams{}
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return exponentialmovingaverage.NewExponentialMovingAverageSmoothingFactor(p)
		}

		p := exponentialmovingaverage.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return exponentialmovingaverage.NewExponentialMovingAverageLength(p)

	case core.Variance:
		p := variance.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return variance.NewVariance(p)

	case core.StandardDeviation:
		p := standarddeviation.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return standarddeviation.NewStandardDeviation(p)

	case core.Momentum:
		p := momentum.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return momentum.New(p)

	case core.RateOfChange:
		p := rateofchange.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return rateofchange.New(p)

	case core.RateOfChangePercent:
		p := rateofchangepercent.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return rateofchangepercent.New(p)

	case core.RateOfChangeRatio:
		p := rateofchangeratio.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return rateofchangeratio.New(p)

	case core.AbsolutePriceOscillator:
		p := absolutepriceoscillator.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return absolutepriceoscillator.NewAbsolutePriceOscillator(p)

	case core.PearsonsCorrelationCoefficient:
		p := pearsonscorrelationcoefficient.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return pearsonscorrelationcoefficient.New(p)

	case core.LinearRegression:
		p := linearregression.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return linearregression.New(p)

	// ── custom ───────────────────────────────────────────────────────────

	case core.GoertzelSpectrum:
		if isEmptyObject(b) {
			return goertzelspectrum.NewGoertzelSpectrumDefault()
		}

		p := &goertzelspectrum.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return goertzelspectrum.NewGoertzelSpectrumParams(p)

	case core.MaximumEntropySpectrum:
		if isEmptyObject(b) {
			return maximumentropyspectrum.NewMaximumEntropySpectrumDefault()
		}

		p := &maximumentropyspectrum.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return maximumentropyspectrum.NewMaximumEntropySpectrumParams(p)

	// ── donaldlambert ────────────────────────────────────────────────────

	case core.CommodityChannelIndex:
		p := commoditychannelindex.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return commoditychannelindex.NewCommodityChannelIndex(p)

	// ── genequong ────────────────────────────────────────────────────────

	case core.MoneyFlowIndex:
		p := moneyflowindex.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return moneyflowindex.NewMoneyFlowIndex(p)

	// ── georgelane ───────────────────────────────────────────────────────

	case core.Stochastic:
		p := stochastic.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return stochastic.NewStochastic(p)

	// ── geraldappel ──────────────────────────────────────────────────────

	case core.PercentagePriceOscillator:
		p := percentagepriceoscillator.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return percentagepriceoscillator.NewPercentagePriceOscillator(p)

	case core.MovingAverageConvergenceDivergence:
		p := movingaverageconvergencedivergence.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return movingaverageconvergencedivergence.NewMovingAverageConvergenceDivergence(p)

	// ── igorlivshin ──────────────────────────────────────────────────────

	case core.BalanceOfPower:
		return balanceofpower.NewBalanceOfPower(&balanceofpower.BalanceOfPowerParams{})

	// ── jackhutson ───────────────────────────────────────────────────────

	case core.TripleExponentialMovingAverageOscillator:
		p := tripleexponentialmovingaverageoscillator.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return tripleexponentialmovingaverageoscillator.NewTripleExponentialMovingAverageOscillator(p)

	// ── johnbollinger ────────────────────────────────────────────────────

	case core.BollingerBands:
		p := bollingerbands.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return bollingerbands.NewBollingerBands(p)

	case core.BollingerBandsTrend:
		p := bollingerbandstrend.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return bollingerbandstrend.NewBollingerBandsTrend(p)

	// ── johnehlers ───────────────────────────────────────────────────────

	case core.SuperSmoother:
		p := supersmoother.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return supersmoother.NewSuperSmoother(p)

	case core.CenterOfGravityOscillator:
		p := centerofgravityoscillator.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return centerofgravityoscillator.NewCenterOfGravityOscillator(p)

	case core.CyberCycle:
		if hasKey(b, "smoothingFactor") {
			p := cybercycle.DefaultSmoothingFactorParams()
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return cybercycle.NewCyberCycleSmoothingFactor(p)
		}

		p := cybercycle.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return cybercycle.NewCyberCycleLength(p)

	case core.InstantaneousTrendLine:
		if hasKey(b, "smoothingFactor") {
			p := instantaneoustrendline.DefaultSmoothingFactorParams()
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return instantaneoustrendline.NewInstantaneousTrendLineSmoothingFactor(p)
		}

		p := instantaneoustrendline.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return instantaneoustrendline.NewInstantaneousTrendLineLength(p)

	case core.ZeroLagExponentialMovingAverage:
		p := zerolagexponentialmovingaverage.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return zerolagexponentialmovingaverage.NewZeroLagExponentialMovingAverage(p)

	case core.ZeroLagErrorCorrectingExponentialMovingAverage:
		p := zerolagerrorcorrectingexponentialmovingaverage.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return zerolagerrorcorrectingexponentialmovingaverage.NewZeroLagErrorCorrectingExponentialMovingAverage(p)

	case core.RoofingFilter:
		p := roofingfilter.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return roofingfilter.NewRoofingFilter(p)

	case core.MesaAdaptiveMovingAverage:
		if hasKey(b, "fastLimitSmoothingFactor") || hasKey(b, "slowLimitSmoothingFactor") {
			p := mesaadaptivemovingaverage.DefaultSmoothingFactorParams()
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return mesaadaptivemovingaverage.NewMesaAdaptiveMovingAverageSmoothingFactor(p)
		}

		if isEmptyObject(b) {
			return mesaadaptivemovingaverage.NewMesaAdaptiveMovingAverageDefault()
		}

		p := mesaadaptivemovingaverage.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return mesaadaptivemovingaverage.NewMesaAdaptiveMovingAverageLength(p)

	case core.FractalAdaptiveMovingAverage:
		p := fractaladaptivemovingaverage.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return fractaladaptivemovingaverage.NewFractalAdaptiveMovingAverage(p)

	case core.DominantCycle:
		if isEmptyObject(b) {
			return dominantcycle.NewDominantCycleDefault()
		}

		p := &dominantcycle.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return dominantcycle.NewDominantCycleParams(p)

	case core.SineWave:
		if isEmptyObject(b) {
			return sinewave.NewSineWaveDefault()
		}

		p := &sinewave.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return sinewave.NewSineWaveParams(p)

	case core.HilbertTransformerInstantaneousTrendLine:
		if isEmptyObject(b) {
			return hilberttransformerinstantaneoustrendline.NewHilbertTransformerInstantaneousTrendLineDefault()
		}

		p := &hilberttransformerinstantaneoustrendline.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return hilberttransformerinstantaneoustrendline.NewHilbertTransformerInstantaneousTrendLineParams(p)

	case core.TrendCycleMode:
		if isEmptyObject(b) {
			return trendcyclemode.NewTrendCycleModeDefault()
		}

		p := &trendcyclemode.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return trendcyclemode.NewTrendCycleModeParams(p)

	case core.CoronaSpectrum:
		if isEmptyObject(b) {
			return coronaspectrum.NewCoronaSpectrumDefault()
		}

		p := &coronaspectrum.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return coronaspectrum.NewCoronaSpectrumParams(p)

	case core.CoronaSignalToNoiseRatio:
		if isEmptyObject(b) {
			return coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()
		}

		p := &coronasignaltonoiseratio.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioParams(p)

	case core.CoronaSwingPosition:
		if isEmptyObject(b) {
			return coronaswingposition.NewCoronaSwingPositionDefault()
		}

		p := &coronaswingposition.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return coronaswingposition.NewCoronaSwingPositionParams(p)

	case core.CoronaTrendVigor:
		if isEmptyObject(b) {
			return coronatrendvigor.NewCoronaTrendVigorDefault()
		}

		p := &coronatrendvigor.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return coronatrendvigor.NewCoronaTrendVigorParams(p)

	case core.AutoCorrelationIndicator:
		if isEmptyObject(b) {
			return autocorrelationindicator.NewAutoCorrelationIndicatorDefault()
		}

		p := &autocorrelationindicator.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return autocorrelationindicator.NewAutoCorrelationIndicatorParams(p)

	case core.AutoCorrelationPeriodogram:
		if isEmptyObject(b) {
			return autocorrelationperiodogram.NewAutoCorrelationPeriodogramDefault()
		}

		p := &autocorrelationperiodogram.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return autocorrelationperiodogram.NewAutoCorrelationPeriodogramParams(p)

	case core.CombBandPassSpectrum:
		if isEmptyObject(b) {
			return combbandpassspectrum.NewCombBandPassSpectrumDefault()
		}

		p := &combbandpassspectrum.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return combbandpassspectrum.NewCombBandPassSpectrumParams(p)

	case core.DiscreteFourierTransformSpectrum:
		if isEmptyObject(b) {
			return discretefouriertransformspectrum.NewDiscreteFourierTransformSpectrumDefault()
		}

		p := &discretefouriertransformspectrum.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return discretefouriertransformspectrum.NewDiscreteFourierTransformSpectrumParams(p)

	// ── josephgranville ──────────────────────────────────────────────────

	case core.OnBalanceVolume:
		p := &onbalancevolume.OnBalanceVolumeParams{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return onbalancevolume.NewOnBalanceVolume(p)

	// ── larrywilliams ────────────────────────────────────────────────────

	case core.WilliamsPercentR:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return williamspercentr.NewWilliamsPercentR(p.Length), nil

	case core.UltimateOscillator:
		p := ultimateoscillator.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return ultimateoscillator.NewUltimateOscillator(p)

	// ── marcchaikin ──────────────────────────────────────────────────────

	case core.AdvanceDecline:
		return advancedecline.NewAdvanceDecline(&advancedecline.AdvanceDeclineParams{})

	case core.AdvanceDeclineOscillator:
		p := advancedeclineoscillator.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return advancedeclineoscillator.NewAdvanceDeclineOscillator(p)

	// ── markjurik ────────────────────────────────────────────────────────

	case core.JurikMovingAverage:
		p := jurikmovingaverage.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return jurikmovingaverage.NewJurikMovingAverage(p)

	case core.JurikRelativeTrendStrengthIndex:
		p := jurikrelativetrendstrengthindex.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return jurikrelativetrendstrengthindex.NewJurikRelativeTrendStrengthIndex(p)

	case core.JurikCompositeFractalBehaviorIndex:
		p := jurikcompositefractalbehaviorindex.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return jurikcompositefractalbehaviorindex.NewJurikCompositeFractalBehaviorIndex(p)

	case core.JurikZeroLagVelocity:
		p := jurikzerolagvelocity.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return jurikzerolagvelocity.NewJurikZeroLagVelocity(p)

	case core.JurikDirectionalMovementIndex:
		p := jurikdirectionalmovementindex.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return jurikdirectionalmovementindex.NewJurikDirectionalMovementIndex(p)

	// ── patrickmulloy ────────────────────────────────────────────────────

	case core.DoubleExponentialMovingAverage:
		if hasKey(b, "smoothingFactor") {
			p := &doubleexponentialmovingaverage.DoubleExponentialMovingAverageSmoothingFactorParams{}
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return doubleexponentialmovingaverage.NewDoubleExponentialMovingAverageSmoothingFactor(p)
		}

		p := doubleexponentialmovingaverage.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return doubleexponentialmovingaverage.NewDoubleExponentialMovingAverageLength(p)

	case core.TripleExponentialMovingAverage:
		if hasKey(b, "smoothingFactor") {
			p := &tripleexponentialmovingaverage.TripleExponentialMovingAverageSmoothingFactorParams{}
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return tripleexponentialmovingaverage.NewTripleExponentialMovingAverageSmoothingFactor(p)
		}

		p := tripleexponentialmovingaverage.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return tripleexponentialmovingaverage.NewTripleExponentialMovingAverageLength(p)

	// ── perrykaufman ─────────────────────────────────────────────────────

	case core.KaufmanAdaptiveMovingAverage:
		if hasKey(b, "fastestSmoothingFactor") || hasKey(b, "slowestSmoothingFactor") {
			p := kaufmanadaptivemovingaverage.DefaultSmoothingFactorParams()
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return kaufmanadaptivemovingaverage.NewKaufmanAdaptiveMovingAverageSmoothingFactor(p)
		}

		p := kaufmanadaptivemovingaverage.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return kaufmanadaptivemovingaverage.NewKaufmanAdaptiveMovingAverageLength(p)

	// ── timtillson ───────────────────────────────────────────────────────

	case core.T2ExponentialMovingAverage:
		if hasKey(b, "smoothingFactor") {
			p := t2exponentialmovingaverage.DefaultSmoothingFactorParams()
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return t2exponentialmovingaverage.NewT2ExponentialMovingAverageSmoothingFactor(p)
		}

		p := t2exponentialmovingaverage.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return t2exponentialmovingaverage.NewT2ExponentialMovingAverageLength(p)

	case core.T3ExponentialMovingAverage:
		if hasKey(b, "smoothingFactor") {
			p := t3exponentialmovingaverage.DefaultSmoothingFactorParams()
			if err := unmarshal(b, p); err != nil {
				return nil, err
			}

			return t3exponentialmovingaverage.NewT3ExponentialMovingAverageSmoothingFactor(p)
		}

		p := t3exponentialmovingaverage.DefaultLengthParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return t3exponentialmovingaverage.NewT3ExponentialMovingAverageLength(p)

	// ── tusharchande ─────────────────────────────────────────────────────

	case core.ChandeMomentumOscillator:
		p := chandemomentumoscillator.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return chandemomentumoscillator.New(p)

	case core.StochasticRelativeStrengthIndex:
		p := stochasticrelativestrengthindex.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return stochasticrelativestrengthindex.NewStochasticRelativeStrengthIndex(p)

	case core.Aroon:
		p := aroon.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return aroon.NewAroon(p)

	// ── vladimirkravchuk ─────────────────────────────────────────────────

	case core.AdaptiveTrendAndCycleFilter:
		if isEmptyObject(b) {
			return adaptivetrendandcyclefilter.NewAdaptiveTrendAndCycleFilterDefault()
		}

		p := &adaptivetrendandcyclefilter.Params{}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return adaptivetrendandcyclefilter.NewAdaptiveTrendAndCycleFilterParams(p)

	// ── welleswilder ─────────────────────────────────────────────────────

	case core.TrueRange:
		return truerange.NewTrueRange(), nil

	case core.AverageTrueRange:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return averagetruerange.NewAverageTrueRange(p.Length)

	case core.NormalizedAverageTrueRange:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return normalizedaveragetruerange.NewNormalizedAverageTrueRange(p.Length)

	case core.DirectionalMovementMinus:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return directionalmovementminus.NewDirectionalMovementMinus(p.Length)

	case core.DirectionalMovementPlus:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return directionalmovementplus.NewDirectionalMovementPlus(p.Length)

	case core.DirectionalIndicatorMinus:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return directionalindicatorminus.NewDirectionalIndicatorMinus(p.Length)

	case core.DirectionalIndicatorPlus:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return directionalindicatorplus.NewDirectionalIndicatorPlus(p.Length)

	case core.DirectionalMovementIndex:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return directionalmovementindex.NewDirectionalMovementIndex(p.Length)

	case core.AverageDirectionalMovementIndex:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return averagedirectionalmovementindex.NewAverageDirectionalMovementIndex(p.Length)

	case core.AverageDirectionalMovementIndexRating:
		p := &lengthParam{Length: 14}
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return averagedirectionalmovementindexrating.NewAverageDirectionalMovementIndexRating(p.Length)

	case core.RelativeStrengthIndex:
		p := relativestrengthindex.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return relativestrengthindex.NewRelativeStrengthIndex(p)

	case core.ParabolicStopAndReverse:
		p := parabolicstopandreverse.DefaultParams()
		if err := unmarshal(b, p); err != nil {
			return nil, err
		}

		return parabolicstopandreverse.NewParabolicStopAndReverse(p)

	default:
		return nil, fmt.Errorf("unsupported indicator: %s", identifier)
	}
}

// isEmptyObject returns true if data is "{}" (with optional whitespace).
func isEmptyObject(data []byte) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return false
	}

	return len(m) == 0
}
