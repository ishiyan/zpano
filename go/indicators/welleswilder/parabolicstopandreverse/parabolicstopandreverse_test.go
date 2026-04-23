//nolint:testpackage
package parabolicstopandreverse

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// Wilder's original SAR test data (38 bars).
func wilderHighs() []float64 {
	return []float64{
		51.12,
		52.35, 52.1, 51.8, 52.1, 52.5, 52.8, 52.5, 53.5, 53.5, 53.8, 54.2, 53.4, 53.5,
		54.4, 55.2, 55.7, 57, 57.5, 58, 57.7, 58, 57.5, 57, 56.7, 57.5,
		56.70, 56.00, 56.20, 54.80, 55.50, 54.70, 54.00, 52.50, 51.00, 51.50, 51.70, 53.00,
	}
}

func wilderLows() []float64 {
	return []float64{
		50.0,
		51.5, 51, 50.5, 51.25, 51.7, 51.85, 51.5, 52.3, 52.5, 53, 53.5, 52.5, 52.1, 53,
		54, 55, 56, 56.5, 57, 56.5, 57.3, 56.7, 56.3, 56.2, 56,
		55.50, 55.00, 54.90, 54.00, 54.50, 53.80, 53.00, 51.50, 50.00, 50.50, 50.20, 51.50,
	}
}

// High test data, 252 entries. Standard TA-Lib test dataset.
func testHighs() []float64 {
	return []float64{
		93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,
		96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
		90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
		87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,
		84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
		90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
		90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,
		89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
		104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,
		123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,
		116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,
		116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,
		125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
		138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
		125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,
		124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
		124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,
		132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,
		129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,
		123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,
		108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,
		96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
		95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,
		104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,
		107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,
		110.5, 109.5,
	}
}

// Low test data, 252 entries.
func testLows() []float64 {
	return []float64{
		90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,
		92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
		87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
		84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
		82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,
		89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,
		88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,
		86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
		101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,
		117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
		112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
		112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,
		123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,
		135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
		122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,
		121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,
		122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,
		128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,
		126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,
		119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,
		105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
		93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,
		93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,
		102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
		104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,
		108.56, 106.62,
	}
}

// Expected SAREXT output for 252-bar dataset with default parameters.
// Positive = long position, negative = short position.
// Index 0 = NaN (lookback = 1), indices 1-251 = valid.
func testExpected() []float64 {
	return []float64{
		math.NaN(), 90.7500000000, 90.8338000000, 91.0554480000, 91.2682300800,
		91.4725008768, 91.6686008417, -96.3750000000, -96.2425000000, -95.9704000000,
		89.4400000000, 89.5762000000, 89.9781520000, -99.6250000000, -99.4088000000,
		-98.9024480000, -98.4163500800, -97.9496960768, -97.5017082337, -97.0716399044,
		-96.6587743082, -96.2624233359, -95.6784779357, -94.7941997009, -93.5427797308,
		-92.1054461631, -90.5331837003, 80.8750000000, 81.0456000000, 81.2127880000,
		81.3766322400, 81.5371995952, 81.6945556033, 81.8487644912, 81.9998892014,
		82.3198936333, 82.6270978880, 82.9220139725, -90.0000000000, -89.8513000000,
		-89.7055740000, -89.4185510400, 82.5300000000, 82.6688000000, 82.9620480000,
		83.5307251200, 84.0652816128, 84.7776590838, 85.4330463571, 86.0360026485,
		86.5907224366, -92.9700000000, -92.8400000000, -92.4864000000, -91.9361160000,
		-91.1412267200, -90.4099285824, -89.7371342958, 82.0000000000, 82.1475000000,
		82.4866000000, 82.9824040000, 83.4484597600, 84.1301829792, 85.0546646813,
		86.1059049195, 87.2152782308, 88.1693392785, 88.9898317795, -94.0300000000,
		-93.9257000000, -93.6386720000, -93.1242516800, -92.2367115456, -91.1630403910,
		81.5000000000, 81.8575000000, 82.7144000000, 84.0740360000, 85.9581131200,
		87.6914640704, 89.2861469448, 90.7532551892, 92.1029947741, 93.6866952966,
		95.1120257670, 96.3948231903, 97.8774444074, 99.7062021904, 101.2789338837,
		103.3495044623, 106.8865936591, 109.7870068005, 112.1653455764, 114.1155833726,
		115.7147783656, -123.0000000000, -122.8319000000, -122.3536240000, -121.6424065600,
		-120.9738621664, -120.3454304364, -119.7547046102, -119.1994223336, -118.4186685469,
		-117.7003750632, 109.4400000000, 109.6850000000, 109.9251000000, 110.1603980000,
		110.3909900400, 110.6169702392, 110.8384308344, 111.0554622177, 111.2681529734,
		111.6974268544, 112.1095297803, 112.9279579934, 114.1137213540, 115.2046236457,
		116.2082537540, 117.1315934537, 117.9810659774, 118.7625806992, 120.0863226293,
		121.2776903663, 122.5943675224, 124.1743560693, 125.5331462196, 127.1278428244,
		128.9840311160, 130.9252248928, 132.5781799143, 133.9005439314, 134.5000000000,
		-139.1900000000, -139.0800000000, -138.8800000000, -138.3672000000, -137.4751680000,
		-136.2867545600, -135.1934141952, -134.1875410596, -133.2621377748, -132.4107667528,
		-131.6275054126, -130.6457548713, -129.1510642868, -127.5983152866, 117.5600000000,
		117.5600000000, 117.7488000000, 117.9338240000, 118.1151475200, 118.2928445696,
		118.4669876782, 118.8431081711, 119.4261216808, 120.2424319463, 120.9934373906,
		121.0600000000, -129.6300000000, -129.4574000000, -129.1139040000, -128.7841478400,
		-128.4675819264, -128.1636786493, 120.8700000000, 121.0226000000, 121.1721480000,
		121.5152620800, 122.1857463552, 123.1260866468, 124.4634779821, 126.0506606243,
		127.4473813493, 128.6764955874, -137.6900000000, -137.5274000000, -136.9763040000,
		-136.4472518400, -135.9393617664, -135.4517872957, -134.6110800580, -133.4773936534,
		-132.4344021611, -131.4748499882, -130.3273649894, -129.0424811907, -127.1465338240,
		-125.5160190886, -124.1137764162, -123.2500000000, -122.7500000000, -120.6458000000,
		-118.9203560000, -117.1362848000, -114.8700000000, -112.8340000000, -111.0412000000,
		-109.6069600000, 103.8700000000, -113.0600000000, -113.0600000000, -112.5788000000,
		-112.1072240000, -111.6450795200, -111.1921779296, -110.7483343710, -110.3133676836,
		-109.8871003299, -109.4693583233, -109.0599711569, -108.6587717337, -108.2655962990,
		-107.8802843731, -107.5026786856, -107.1326251119, -106.7699726096, -106.4145731575,
		-106.0662816943, -105.7249560604, -105.3904569392, 89.0000000000, 89.3224000000,
		90.0871040000, 91.2740777600, 92.3898330944, 93.4386431087, 94.4245245222,
		95.3512530509, 96.2223778678, 97.0412351958, 98.3075363801, 100.1267827421,
		102.4143688130, 105.0163571792, -122.1200000000, -122.1200000000, -121.4976000000,
		-120.9000960000, -119.9160902400, -118.9911248256, -118.1216573361, -117.3043578959,
		-116.5360964221, -115.8139306368, -115.1350947986, -114.4969891107, -113.8971697641,
		-113.3333395782, -112.8033392035,
	}
}

func TestParabolicStopAndReverse252Bar(t *testing.T) {
	t.Parallel()

	const tol = 1e-6

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := testHighs()
	lows := testLows()
	expected := testExpected()

	for i := range highs {
		result := sar.UpdateHL(highs[i], lows[i])

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result) {
				t.Errorf("[%d] expected NaN, got %v", i, result)
			}

			continue
		}

		diff := math.Abs(result - expected[i])
		if diff > tol {
			t.Errorf("[%d] expected %.10f, got %.10f, diff %.10f", i, expected[i], result, diff)
		}
	}
}

func TestParabolicStopAndReverseWilder(t *testing.T) {
	t.Parallel()

	const tol = 1e-3

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := wilderHighs()
	lows := wilderLows()
	results := make([]float64, len(highs))

	for i := range highs {
		results[i] = sar.UpdateHL(highs[i], lows[i])
	}

	// Wilder spot checks from test_sar.c (TA_SAR, absolute values).
	// expectedBegIndex = 1, so output[0] corresponds to results[1].
	// TA_SAR always returns positive values, SAREXT returns signed.
	spotChecks := []struct {
		outIndex int     // index into output array (begIndex-relative)
		expected float64 // TA_SAR absolute expected value
	}{
		{0, 50.00},
		{1, 50.047},
		{4, 50.182},
		{35, 52.93},
		{36, 50.00},
	}

	for _, sc := range spotChecks {
		actual := math.Abs(results[sc.outIndex+1]) // +1 because results[0] = NaN
		diff := math.Abs(actual - sc.expected)

		if diff > tol {
			t.Errorf("Wilder spot check output[%d]: expected %.4f, got %.4f, diff %.6f",
				sc.outIndex, sc.expected, actual, diff)
		}
	}
}

func TestParabolicStopAndReverseIsPrimed(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sar.IsPrimed() {
		t.Error("expected not primed before any data")
	}

	// First bar — still not primed.
	sar.UpdateHL(93.25, 90.75)

	if sar.IsPrimed() {
		t.Error("expected not primed after 1 bar")
	}

	// Second bar — should be primed.
	sar.UpdateHL(94.94, 91.405)

	if !sar.IsPrimed() {
		t.Error("expected primed after 2 bars")
	}
}

func TestParabolicStopAndReverseMetadata(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := sar.Metadata()

	if meta.Identifier != core.ParabolicStopAndReverse {
		t.Errorf("expected identifier ParabolicStopAndReverse, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "sar()" {
		t.Errorf("expected mnemonic 'sar()', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected output type ScalarType, got %v", meta.Outputs[0].Shape)
	}
}

func TestParabolicStopAndReverseConstructorValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params ParabolicStopAndReverseParams
		valid  bool
	}{
		{"defaults", ParabolicStopAndReverseParams{}, true},
		{"negative long init", ParabolicStopAndReverseParams{AccelerationInitLong: -0.01}, false},
		{"negative short step", ParabolicStopAndReverseParams{AccelerationShort: -0.01}, false},
		{"negative offset", ParabolicStopAndReverseParams{OffsetOnReverse: -0.01}, false},
		{"custom valid", ParabolicStopAndReverseParams{
			AccelerationInitLong:  0.01,
			AccelerationLong:      0.01,
			AccelerationMaxLong:   0.10,
			AccelerationInitShort: 0.03,
			AccelerationShort:     0.03,
			AccelerationMaxShort:  0.30,
		}, true},
		{"start value positive", ParabolicStopAndReverseParams{StartValue: 100.0}, true},
		{"start value negative", ParabolicStopAndReverseParams{StartValue: -100.0}, true},
	}

	for _, tt := range tests {
		_, err := NewParabolicStopAndReverse(&tt.params)
		if tt.valid && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		}

		if !tt.valid && err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}

func TestParabolicStopAndReverseUpdateBar(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()

	bar1 := &entities.Bar{Time: now, Open: 91, High: 93.25, Low: 90.75, Close: 91.5, Volume: 1000}
	out1 := sar.UpdateBar(bar1)
	scalar1 := out1[0].(entities.Scalar)

	if !math.IsNaN(scalar1.Value) {
		t.Errorf("expected NaN for first bar, got %v", scalar1.Value)
	}

	bar2 := &entities.Bar{Time: now.Add(time.Minute), Open: 92, High: 94.94, Low: 91.405, Close: 94.815, Volume: 1000}
	out2 := sar.UpdateBar(bar2)
	scalar2 := out2[0].(entities.Scalar)

	if math.IsNaN(scalar2.Value) {
		t.Error("expected valid value for second bar, got NaN")
	}
}

func TestParabolicStopAndReverseNaN(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Feed two valid bars to prime.
	sar.UpdateHL(93.25, 90.75)
	sar.UpdateHL(94.94, 91.405)

	// Feed NaN — should not corrupt state.
	result := sar.UpdateHL(math.NaN(), 92.0)
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for NaN input, got %v", result)
	}

	// Feed valid data — should still work.
	result = sar.UpdateHL(96.375, 94.25)
	if math.IsNaN(result) {
		t.Error("expected valid output after NaN, got NaN")
	}
}

func TestParabolicStopAndReverseForcedStartLong(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{
		StartValue: 85.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := testHighs()
	lows := testLows()

	// First bar: NaN.
	result := sar.UpdateHL(highs[0], lows[0])
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for first bar, got %v", result)
	}

	// Second bar: should be positive (long).
	result = sar.UpdateHL(highs[1], lows[1])
	if result <= 0 {
		t.Errorf("expected positive (long) SAR with forced long start, got %v", result)
	}
}

func TestParabolicStopAndReverseForcedStartShort(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{
		StartValue: -100.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := testHighs()
	lows := testLows()

	// First bar: NaN.
	result := sar.UpdateHL(highs[0], lows[0])
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for first bar, got %v", result)
	}

	// Second bar: should be negative (short).
	result = sar.UpdateHL(highs[1], lows[1])
	if result >= 0 {
		t.Errorf("expected negative (short) SAR with forced short start, got %v", result)
	}
}
