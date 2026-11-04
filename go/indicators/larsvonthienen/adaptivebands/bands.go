// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package adaptivebands

import ()

// GridBands performs the exact 101-level band search rather than an interpolated quantile.
func GridBands(values []float64, leveling float64) (float64, float64) {
	low, high := values[0], values[0]
	for _, value := range values {
		low = min(low, value)
		high = max(high, value)
	}
	step := (high - low) / 100
	fraction := leveling / 100
	lower, upper := low, high
	for k := range 101 {
		test := low + step*float64(k)
		below := 0
		for _, value := range values {
			if value < test {
				below++
			}
		}
		if float64(below)/float64(len(values)) >= fraction {
			lower = test
			break
		}
	}
	for k := range 101 {
		test := high - step*float64(k)
		above := 0
		for _, value := range values {
			if value >= test {
				above++
			}
		}
		if float64(above)/float64(len(values)) >= fraction {
			upper = test
			break
		}
	}
	return lower, upper
}
