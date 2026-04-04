package performance

// Periodicity represents the frequency of performance measurement periods.
type Periodicity int

const (
	// Daily periodicity (252 trading days per year).
	Daily Periodicity = iota

	// Weekly periodicity (52 weeks per year).
	Weekly

	// Monthly periodicity (12 months per year).
	Monthly

	// Quarterly periodicity (4 quarters per year).
	Quarterly

	// Annual periodicity (1 period per year).
	Annual
)

// PeriodsPerAnnum returns the number of periods per year for a given periodicity.
func (p Periodicity) PeriodsPerAnnum() int {
	switch p {
	case Daily:
		return 252
	case Weekly:
		return 52
	case Monthly:
		return 12
	case Quarterly:
		return 4
	case Annual:
		return 1
	default:
		return 252
	}
}

// DaysPerPeriod returns the number of trading days per period for a given periodicity.
func (p Periodicity) DaysPerPeriod() float64 {
	switch p {
	case Daily:
		return 1
	case Weekly:
		return 252.0 / 52.0
	case Monthly:
		return 252.0 / 12.0
	case Quarterly:
		return 252.0 / 4.0
	case Annual:
		return 252
	default:
		return 1
	}
}
