//nolint:testpackage
package onbalancevolume

import (
	"math"
	"time"
)

// C# test data: 12 entries.
// Prices: [1,2,8,4,9,6,7,13,9,10,3,12]
// Volumes: [100,90,200,150,500,100,300,150,100,300,200,100]
// Expected: [100,190,390,240,740,640,940,1090,990,1290,1090,1190]

func testPrices() []float64 {
	return []float64{1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12}
}

func testVolumes() []float64 {
	return []float64{100, 90, 200, 150, 500, 100, 300, 150, 100, 300, 200, 100}
}

func testExpected() []float64 {
	return []float64{100, 190, 390, 240, 740, 640, 940, 1090, 990, 1290, 1090, 1190}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func roundTo(v float64, digits int) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(v*p) / p
}
