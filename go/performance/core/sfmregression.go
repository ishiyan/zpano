package core

import (
	"math"

	"zpano/streamingkbn"
)

// SFMRegression is a streaming single factor model (CAPM) regression of
// the excess portfolio returns on the excess benchmark returns,
//
//	r - rf = α + β·(b - rf)
//
// with separate slopes fitted over bull (b - rf > 0) and bear
// (b - rf < 0) periods.
//
// Use [NewSFMRegression] to create an instance.
type SFMRegression struct {
	riskFreeRate float64
	full         *streamingkbn.LinearRegressionKleinKBN
	bull         *streamingkbn.LinearRegressionKleinKBN
	bear         *streamingkbn.LinearRegressionKleinKBN
}

// NewSFMRegression returns a new empty SFMRegression using the given
// risk-free rate (in the same periodicity as the returns).
func NewSFMRegression(riskFreeRate float64) *SFMRegression {
	return &SFMRegression{
		riskFreeRate: riskFreeRate,
		full:         streamingkbn.NewLinearRegressionKleinKBN(),
		bull:         streamingkbn.NewLinearRegressionKleinKBN(),
		bear:         streamingkbn.NewLinearRegressionKleinKBN(),
	}
}

// RiskFreeRate returns the risk-free rate.
func (s *SFMRegression) RiskFreeRate() float64 { return s.riskFreeRate }

// Reset clears all accumulated state (the risk-free rate is kept).
func (s *SFMRegression) Reset() {
	s.full.Reset()
	s.bull.Reset()
	s.bear.Reset()
}

// Revert removes a previously added (portfolio, benchmark) return pair.
func (s *SFMRegression) Revert(ret, benchmark float64) {
	x := benchmark - s.riskFreeRate
	y := ret - s.riskFreeRate

	s.full.Revert(x, y)

	if x > 0 {
		s.bull.Revert(x, y)
	} else if x < 0 {
		s.bear.Revert(x, y)
	}
}

// Update adds a (portfolio, benchmark) return pair.
func (s *SFMRegression) Update(ret, benchmark float64) {
	x := benchmark - s.riskFreeRate
	y := ret - s.riskFreeRate

	s.full.Update(x, y)

	if x > 0 {
		s.bull.Update(x, y)
	} else if x < 0 {
		s.bear.Update(x, y)
	}
}

// Alpha returns the regression intercept α (NaN when undefined).
func (s *SFMRegression) Alpha() float64 { return s.full.Intercept() }

// Beta returns the regression slope β (NaN when undefined).
func (s *SFMRegression) Beta() float64 { return s.full.Slope() }

// BetaBull returns the slope fitted over bull periods (NaN when undefined).
func (s *SFMRegression) BetaBull() float64 { return s.bull.Slope() }

// BetaBear returns the slope fitted over bear periods (NaN when undefined).
func (s *SFMRegression) BetaBear() float64 { return s.bear.Slope() }

// R2 returns the coefficient of determination, the squared correlation
// (NaN when undefined).
func (s *SFMRegression) R2() float64 {
	corr := s.full.Correlation()
	if math.IsNaN(corr) {
		return math.NaN()
	}
	return corr * corr
}
