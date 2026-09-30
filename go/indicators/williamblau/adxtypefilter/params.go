package adxtypefilter

import "zpano/entities"

// Source specifies the bipolar momentum the ADX-Type Filter is applied to.
type Source int

const (
	// TsiMomentum is the TSI numerator C - C[q-1] (named instance TSI_ATF).
	TsiMomentum Source = iota

	// SmiMomentum is the SMI raw stochastic momentum C - 0.5*(HH(q) + LL(q))
	// (named instance SMI_ATF).
	SmiMomentum

	// DtiMomentum is the DTI numerator max(H - H[q-1], 0) - max(L[q-1] - L, 0).
	DtiMomentum

	// TviBalance is the TVI tick balance upticks - downticks (2C - H - L for a bar).
	TviBalance

	// TsiNormalized is the single-smoothed normalized TSI(q, r, 1, 1), which
	// replaces the inner EMA.
	TsiNormalized
)

// Params describes parameters to create an instance of the indicator.
//
// The field names Q, R and S are the canonical symbols from William Blau's
// Momentum, Direction, and Divergence (Wiley, 1995), Appendix B, Figure B-24.
// They are kept verbatim for fidelity with the book and the test-data naming.
type Params struct {
	// Source specifies the bipolar momentum the filter is applied to.
	//
	// If zero (TsiMomentum), the TSI numerator is used.
	Source Source

	// Q is the momentum look-back period.
	//
	// If zero, the source default is used: 2 for TsiMomentum, DtiMomentum and
	// TsiNormalized, 32 for SmiMomentum. The look-back is not used by
	// TviBalance. A non-zero value should be greater than 0.
	Q int

	// R is the period of the inner EMA, applied to the signed momentum.
	//
	// For TsiNormalized this is the smoothing period of the normalized TSI,
	// which replaces the inner EMA. The value should be greater than 0. The
	// default value is 32.
	R int

	// S is the period of the outer EMA, applied to the rectified momentum.
	//
	// The value should be greater than 0. The default value is 32.
	S int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// Used only by the TsiMomentum and TsiNormalized sources. If zero, the default (BarClosePrice)
	// is used and the component is not shown in the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// Used only by the TsiMomentum and TsiNormalized sources. If zero, the default (QuoteMidPrice)
	// is used and the component is not shown in the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// Used only by the TsiMomentum and TsiNormalized sources. If zero, the default (TradePrice)
	// is used and the component is not shown in the indicator mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a Params value populated with conventional defaults.
//
// The look-back Q is left at zero so that it resolves to the default of
// whichever source is selected.
func DefaultParams() *Params {
	return &Params{
		Source: TsiMomentum,
		R:      32,
		S:      32,
	}
}
