package mexicanhatwavelet

import "zpano/entities"

// Band selects the frequency band of the Mexican Hat Wavelet filter.
type Band int

const (
	// BandHigh is the high-frequency band (a_f = 1.483, period ~ 4.6 bars).
	BandHigh Band = iota

	// BandMid is the mid-frequency band (a_f = 4.048, period ~ 13.5 bars).
	BandMid

	// BandLow is the low-frequency band (a_f = 15.97, period ~ 54 bars).
	BandLow

	// BandCustom uses a user-specified dilation or period.
	BandCustom
)

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Band selects the frequency band (BandHigh, BandMid, BandLow, BandCustom).
	//
	// The default value is BandMid.
	Band Band

	// Dilation is the custom dilation parameter a_f, used only when Band is BandCustom.
	//
	// The value should be > 0. Zero means unset.
	Dilation float64

	// Period is the custom center period in bars, used only when Band is BandCustom.
	//
	// The value should be > 2. Zero means unset. Mutually exclusive with Dilation.
	Period float64

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
