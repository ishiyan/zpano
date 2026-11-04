// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package hodrickprescottsmoothinganddetrending

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// HPTrend solves the complete two-sided HP system with natural endpoints.
// Batch estimates can revise history; the streaming indicator emits only the endpoint.
func HPTrend(values []float64, smoothing float64) []float64 {
	n := len(values)
	if n < 6 {
		panic("HP requires at least six observations")
	}
	if err := harmoniccycle.Real(smoothing, "smoothing", 0); err != nil {
		panic(err)
	}
	a := make([]float64, n)
	b := make([]float64, n)
	c := make([]float64, n)
	for i, value := range values {
		harmoniccycle.Sample(value)
		if math.IsNaN(value) {
			panic("HP batch data must be finite")
		}
		a[i] = 6*smoothing + 1
		b[i] = -4 * smoothing
		c[i] = smoothing
	}
	a[0] = 1 + smoothing
	a[n-1] = 1 + smoothing
	a[1] = 1 + 5*smoothing
	a[n-2] = 1 + 5*smoothing
	b[0] = -2 * smoothing
	b[n-2] = -2 * smoothing
	b[n-1] = 0
	c[n-1] = 0
	c[n-2] = 0
	var h1, h2, h3, h4, h5, hh2, hh3, hh5 float64
	for i := range n {
		z := harmoniccycle.Finite(a[i] - h4*h1 - hh5*hh2)
		if z <= 0 {
			panic("HP elimination lost positive definiteness")
		}
		hb := b[i]
		hh1 := h1
		h1 = (hb - h4*h2) / z
		b[i] = h1
		hc := c[i]
		hh2 = h2
		h2 = hc / z
		c[i] = h2
		a[i] = harmoniccycle.Finite((values[i] - hh3*hh5 - h3*h4) / z)
		hh3 = h3
		h3 = a[i]
		h4 = hb - h5*hh1
		hh5 = h5
		h5 = hc
	}
	output := make([]float64, n)
	h2 = 0
	h1 = a[n-1]
	output[n-1] = h1
	for i := n - 1; i > 0; i-- {
		output[i-1] = harmoniccycle.Finite(a[i-1] - b[i-1]*h1 - c[i-1]*h2)
		h2, h1 = h1, output[i-1]
	}
	return output
}

// CycleData disables detrending at zero smoothing; HPTrend at zero instead returns identity.
func CycleData(values []float64, smoothing float64) []float64 {
	if smoothing == 0 {
		return append([]float64{}, values...)
	}
	trend := HPTrend(values, smoothing)
	for i := range trend {
		trend[i] = harmoniccycle.Finite(values[i] - trend[i])
	}
	return trend
}
