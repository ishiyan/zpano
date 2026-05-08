//nolint:testpackage
package variance

import "math"

// testVarianceInput is variance input test data.
func testVarianceInput() []float64 { return []float64{1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12} }

// testVarianceExpectedLength3Population is the Excel (VAR.P) output of population variance of length 3.
func testVarianceExpectedLength3Population() []float64 {
	return []float64{
		math.NaN(), math.NaN(),
		9.55555555555556000, 6.22222222222222000, 4.66666666666667000, 4.22222222222222000, 1.55555555555556000,
		9.55555555555556000, 6.22222222222222000, 2.88888888888889000, 9.55555555555556000, 14.88888888888890000,
	}
}

// testVarianceExpectedLength5Population is the Excel (VAR.P) output of population variance of length 5.
func testVarianceExpectedLength5Population() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		10.16000, 6.56000, 2.96000, 9.36000, 5.76000, 6.00000, 11.04000, 12.24000,
	}
}

// testVarianceExpectedLength3Sample is the Excel (VAR.S) output of sample variance of length 3.
func testVarianceExpectedLength3Sample() []float64 {
	return []float64{
		math.NaN(), math.NaN(),
		14.3333333333333000, 9.3333333333333400, 7.0000000000000000, 6.3333333333333400, 2.3333333333333300,
		14.3333333333333000, 9.3333333333333400, 4.3333333333333400, 14.3333333333333000, 22.3333333333333000,
	}
}

// testVarianceExpectedLength5Sample is the Excel (VAR.S) output of sample variance of length 5.
func testVarianceExpectedLength5Sample() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		12.7000, 8.2000, 3.7000, 11.7000, 7.2000, 7.5000, 13.8000, 15.3000,
	}
}
