package core

import (
	"math"

	"zpano/streamingkbn"
)

// ProbabilisticSharpeRatio returns the probabilistic Sharpe ratio (Bailey
// and López de Prado): the probability that the true Sharpe ratio exceeds
// referenceSR given the observed Sharpe ratio sr over returnsKBN.N()
// observations,
//
//	PSR = Φ((sr - referenceSR)·√(n - 1) / √(1 - sr·γ₃ + sr²·(γ₄ - 1)/4))
//
// where γ₃ is the (biased) skewness and γ₄ the (non-excess) kurtosis of the
// returns. If zeroSkewness is true, γ₃ = 0 is assumed; if normalKurtosis is
// true, γ₄ = 3 is assumed (the conventional defaults are referenceSR = 0,
// zeroSkewness = false, normalKurtosis = true).
//
// Returns NaN when sr is NaN, when a required moment is unavailable, or
// when the denominator is zero.
func ProbabilisticSharpeRatio(returnsKBN *streamingkbn.RawMomentsKleinKBN, sr, referenceSR float64,
	zeroSkewness, normalKurtosis bool,
) float64 {
	if math.IsNaN(sr) {
		return math.NaN()
	}
	var skewness float64
	if zeroSkewness {
		skewness = 0
	} else {
		skewness = returnsKBN.SkewnessMoment()
		if math.IsNaN(skewness) {
			return math.NaN()
		}
	}
	var kurtosis float64
	if normalKurtosis {
		kurtosis = 3 // excess kurtosis = 0, so K = 3
	} else {
		kurtosis = returnsKBN.KurtosisExcess()
		if math.IsNaN(kurtosis) {
			return math.NaN()
		}
		kurtosis += 3 // convert to regular kurtosis
	}

	denom := math.Sqrt(1 - sr*skewness + (sr*sr)*(kurtosis-1)/4)
	if denom == 0 {
		return math.NaN()
	}

	z := (sr - referenceSR) * math.Sqrt(float64(returnsKBN.N()-1)) / denom
	return NormCDF(z)
}
