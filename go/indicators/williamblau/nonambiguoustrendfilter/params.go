package nonambiguoustrendfilter

import "zpano/entities"

// Base specifies the base oscillator the Nonambiguous Trend Filter is applied to.
type Base int

const (
	// Tsi is the True Strength Index (named instance TSI_Trade); defaults q=2, r=32, s=13, u=3.
	Tsi Base = iota

	// Smi is the Stochastic Momentum Index (named instance SMI_Trade); defaults q=32, r=64, s=7, u=1.
	Smi

	// Dti is the Directional Trend Index (named instance DTI_Trade); defaults q=2, r=28, s=28, u=5.
	Dti

	// Tvi is the Tick Volume Indicator (named instance TVI_Trade); defaults r=32, s=32, u=5.
	Tvi

	// Mdi is the Mean Deviation Index (MDI_Trade); defaults r=20, s=5, u=3.
	Mdi

	// Cmi is the Candlestick Momentum Index (CMI_Trade); defaults r=20, s=5, u=3.
	Cmi

	// Csi is the Candlestick Strength Index (CSI_Trade); defaults r=32, s=32, u=1.
	Csi
)

// Params describes parameters to create an instance of the indicator.
//
// The filter itself is parameterless; Q, R, S and U are the canonical symbols
// of the base oscillator from William Blau's Momentum, Direction, and
// Divergence (Wiley, 1995). A zero value selects the book default of the
// selected base (see [Base]).
type Params struct {
	// Base specifies the base oscillator the filter is applied to.
	//
	// If zero (Tsi), the True Strength Index is used.
	Base Base

	// Q is the momentum look-back period of the base (Tsi, Smi and Dti only).
	//
	// If zero, the base default is used. A non-zero value should be greater than 0.
	Q int

	// R is the period of the 1st EMA of the base smoothing cascade.
	//
	// If zero, the base default is used. A non-zero value should be greater than 0.
	R int

	// S is the period of the 2nd EMA of the base smoothing cascade.
	//
	// If zero, the base default is used. A non-zero value should be greater than 0.
	S int

	// U is the period of the 3rd EMA of the base smoothing cascade.
	//
	// If zero, the base default is used. A non-zero value should be greater than 0.
	U int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// Used only by the Tsi and Mdi bases. If zero, the default (BarClosePrice) is used and the
	// component is not shown in the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// Used only by the Tsi and Mdi bases. If zero, the default (QuoteMidPrice) is used and the
	// component is not shown in the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// Used only by the Tsi and Mdi bases. If zero, the default (TradePrice) is used and the
	// component is not shown in the indicator mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a Params value populated with conventional defaults.
//
// The base periods are left at zero so that they resolve to the book defaults
// of whichever base is selected.
func DefaultParams() *Params {
	return &Params{
		Base: Tsi,
	}
}
