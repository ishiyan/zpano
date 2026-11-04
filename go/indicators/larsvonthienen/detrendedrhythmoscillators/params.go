// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
// Numerical source: Lars von Thienen / WhenToTrade (2017), MPL-2.0.
// Research adaptations are documented in description.md.
package detrendedrhythmoscillators

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// FastLength configures fast length; the reference default is 3.
	FastLength int
	// SlowLength configures slow length; the reference default is 50.
	SlowLength int
	// ZigPeriod configures zig period; the reference default is 30.
	ZigPeriod int
	// CycleMemory configures cycle memory; the reference default is 15.
	CycleMemory int
	// Detrend configures detrend; the reference default is True.
	Detrend bool
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params {
	return &Params{FastLength: 3, SlowLength: 50, ZigPeriod: 30, CycleMemory: 15, Detrend: true}
}
