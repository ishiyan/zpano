package core

import (
	"fmt"
	"math"
	"testing"

	"zpano/streamingkbn"
)

func TestProbabilisticSharpeRatioFourMomentAssumptions(t *testing.T) {
	t.Parallel()
	values := []float64{-0.03, -0.01, 0.02, 0.04, 0.07}
	moments := streamingkbn.NewRawMomentsKleinKBN(1, true, true)
	for _, v := range values {
		moments.Update(v)
	}
	n := float64(len(values))
	mean := pySum(values) / n
	mu2 := pySum(mapf(values, func(x float64) float64 { return math.Pow(x-mean, 2) })) / n
	skew := pySum(mapf(values, func(x float64) float64 { return math.Pow(x-mean, 3) })) / n / math.Pow(mu2, 1.5)
	kurt := pySum(mapf(values, func(x float64) float64 { return math.Pow(x-mean, 4) })) / n / math.Pow(mu2, 2)
	sr := 0.75
	referenceSR := 0.1

	for _, zeroSkewness := range []bool{false, true} {
		for _, normalKurtosis := range []bool{false, true} {
			s := skew
			if zeroSkewness {
				s = 0.0
			}
			k := kurt
			if normalKurtosis {
				k = 3.0
			}
			denominator := math.Sqrt(1 - sr*s + sr*sr*(k-1)/4)
			z := (sr - referenceSR) * math.Sqrt(n-1) / denominator
			expected := 0.5 * (1 + math.Erf(z/math.Sqrt(2)))
			actual := ProbabilisticSharpeRatio(moments, sr, referenceSR, zeroSkewness, normalKurtosis)
			assertAlmostEqual(t, actual, expected, places(13),
				fmt.Sprintf("zeroSkewness=%v normalKurtosis=%v", zeroSkewness, normalKurtosis))
		}
	}
}

func TestProbabilisticSharpeRatioUnavailableSharpeOrMoments(t *testing.T) {
	t.Parallel()
	moments := streamingkbn.NewRawMomentsKleinKBN(1, true, true)
	moments.Update(0.01)
	assertNaN(t, ProbabilisticSharpeRatio(moments, math.NaN(), 0.0, false, true), "NaN sr")
	assertNaN(t, ProbabilisticSharpeRatio(moments, 0.5, 0.0, false, true), "skewness unavailable")
	assertNaN(t, ProbabilisticSharpeRatio(moments, 0.5, 0.0, true, false), "kurtosis unavailable")
}
