package core

import (
	"fmt"
	"slices"
	"testing"

	"zpano/streamingkbn"
)

func TestRiskHelpersHistoricalQuantileTailAndRiskFreeRate(t *testing.T) {
	t.Parallel()
	returns := []float64{-0.2, -0.1, 0.0, 0.1}
	assertAlmostEqual(t, VarHistorical(returns, 0.0, 0.75), 0.125, places(15), "var")
	assertAlmostEqual(t, EsHistorical(returns, 0.0, 0.75), 0.2, places(15), "es")
	assertAlmostEqual(t, VarHistorical(returns, 0.01, 0.75), 0.135, places(15), "var rf")
	assertAlmostEqual(t, EsHistorical(returns, 0.01, 0.75), 0.21, places(15), "es rf")
	if !slices.Equal(returns, []float64{-0.2, -0.1, 0.0, 0.1}) {
		t.Errorf("returns were modified: %v", returns)
	}
}

func TestRiskHelpersEmptyHistoricalInputs(t *testing.T) {
	t.Parallel()
	for _, returns := range [][]float64{nil, {}} {
		msg := fmt.Sprintf("returns=%#v", returns)
		assertNaN(t, VarHistorical(returns, 0.0, 0.95), msg+" var")
		assertNaN(t, EsHistorical(returns, 0.0, 0.95), msg+" es")
	}
}

func TestRiskHelpersCornishFisherFallsBackForOneSample(t *testing.T) {
	t.Parallel()
	moments := streamingkbn.NewRawMomentsKleinKBN(1, true, true)
	moments.Update(0.02)
	assertAlmostEqual(t, VarGaussian(moments, 0.95), -0.02, places(15), "var gaussian")
	assertAlmostEqual(t, EsGaussian(moments, 0.95), -0.02, places(15), "es gaussian")
	assertExact(t, VarCornishFisher(moments, 0.95), VarGaussian(moments, 0.95), "var cornish-fisher")
	assertExact(t, EsCornishFisher(moments, 0.95), EsGaussian(moments, 0.95), "es cornish-fisher")
}
