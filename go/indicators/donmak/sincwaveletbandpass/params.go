package sincwaveletbandpass

import "zpano/entities"

// Band selects the frequency band of the Sinc Wavelet Band-Pass filter.
type Band int

const (
	// BandHigh extracts periods 8-16 bars.
	BandHigh Band = iota

	// BandMid extracts periods 16-32 bars.
	BandMid

	// BandLow extracts periods 32-64 bars.
	BandLow

	// BandFull extracts periods 8-64 bars (sum of HIGH + MID + LOW).
	BandFull
)

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Band selects the frequency band (BandHigh, BandMid, BandLow, BandFull).
	//
	// The default value is BandMid.
	Band Band

	// Velocity controls whether a cubic velocity kernel is applied to the band-pass output.
	//
	// The default value is false.
	Velocity bool

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Band: BandMid,
	}
}
