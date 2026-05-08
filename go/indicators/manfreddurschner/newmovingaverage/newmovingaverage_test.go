//nolint:testpackage
package newmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testNmaTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testNmaCreate(primaryPeriod int, secondaryPeriod int, maType MAType) *NewMovingAverage {
	nma, _ := NewNewMovingAverage(&NewMovingAverageParams{
		PrimaryPeriod:   primaryPeriod,
		SecondaryPeriod: secondaryPeriod,
		MAType:          maType,
	})

	return nma
}

func testNmaRun(t *testing.T, name string, primaryPeriod int, secondaryPeriod int, maType MAType, expected []float64) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		nma := testNmaCreate(primaryPeriod, secondaryPeriod, maType)

		for i := 0; i < len(testInput); i++ {
			act := nma.Update(testInput[i])
			exp := expected[i]

			if math.IsNaN(exp) {
				if !math.IsNaN(act) {
					t.Errorf("[%v] expected NaN, got %v", i, act)
				}
			} else if math.Abs(act-exp) > 1e-13 {
				t.Errorf("[%v] expected %v, got %v", i, exp, act)
			}
		}

		// NaN passthrough.
		act := nma.Update(math.NaN())
		if !math.IsNaN(act) {
			t.Errorf("expected NaN passthrough, got %v", act)
		}
	})
}

func TestNewMovingAverageUpdate(t *testing.T) {
	t.Parallel()

	// LWMA tests with various period combinations.
	testNmaRun(t, "sec4 pri_auto LWMA", 0, 4, LWMA, expectedSec4PriAutoLWMA)
	testNmaRun(t, "sec8 pri_auto LWMA (default)", 0, 8, LWMA, expectedSec8PriAutoLWMA)
	testNmaRun(t, "sec16 pri_auto LWMA", 0, 16, LWMA, expectedSec16PriAutoLWMA)
	testNmaRun(t, "pri16 sec8 LWMA", 16, 8, LWMA, expectedPri16Sec8LWMA)
	testNmaRun(t, "pri32 sec8 LWMA", 32, 8, LWMA, expectedPri32Sec8LWMA)
	testNmaRun(t, "pri64 sec8 LWMA", 64, 8, LWMA, expectedPri64Sec8LWMA)
	testNmaRun(t, "pri8 sec4 LWMA", 8, 4, LWMA, expectedPri8Sec4LWMA)
	testNmaRun(t, "pri16 sec4 LWMA", 16, 4, LWMA, expectedPri16Sec4LWMA)
	testNmaRun(t, "pri32 sec4 LWMA", 32, 4, LWMA, expectedPri32Sec4LWMA)

	// Other MA types with default periods.
	testNmaRun(t, "sec8 pri_auto SMA", 0, 8, SMA, expectedSec8SMA)
	testNmaRun(t, "sec8 pri_auto EMA", 0, 8, EMA, expectedSec8EMA)
	testNmaRun(t, "sec8 pri_auto SMMA", 0, 8, SMMA, expectedSec8SMMA)
}

func TestNewMovingAverageUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testNmaTime()

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		nma := testNmaCreate(8, 4, LWMA)
		for i := 0; i < 10; i++ {
			s := entities.Scalar{Time: tm, Value: testInput[i]}
			nma.UpdateScalar(&s)
		}

		s := entities.Scalar{Time: tm, Value: testInput[10]}
		out := nma.UpdateScalar(&s)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		exp := expectedPri8Sec4LWMA[10]
		if math.Abs(sc.Value-exp) > 1e-13 {
			t.Errorf("expected %v, got %v", exp, sc.Value)
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		nma := testNmaCreate(8, 4, LWMA)
		for i := 0; i < 10; i++ {
			b := entities.Bar{Time: tm, Open: testInput[i], High: testInput[i], Low: testInput[i], Close: testInput[i], Volume: 1.0}
			nma.UpdateBar(&b)
		}

		b := entities.Bar{Time: tm, Open: testInput[10], High: testInput[10], Low: testInput[10], Close: testInput[10], Volume: 1.0}
		out := nma.UpdateBar(&b)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		exp := expectedPri8Sec4LWMA[10]
		if math.Abs(sc.Value-exp) > 1e-13 {
			t.Errorf("expected %v, got %v", exp, sc.Value)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		nma := testNmaCreate(8, 4, LWMA)
		for i := 0; i < 10; i++ {
			q := entities.Quote{Time: tm, Bid: testInput[i], Ask: testInput[i]}
			nma.UpdateQuote(&q)
		}

		q := entities.Quote{Time: tm, Bid: testInput[10], Ask: testInput[10]}
		out := nma.UpdateQuote(&q)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		exp := expectedPri8Sec4LWMA[10]
		if math.Abs(sc.Value-exp) > 1e-13 {
			t.Errorf("expected %v, got %v", exp, sc.Value)
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		nma := testNmaCreate(8, 4, LWMA)
		for i := 0; i < 10; i++ {
			r := entities.Trade{Time: tm, Price: testInput[i]}
			nma.UpdateTrade(&r)
		}

		r := entities.Trade{Time: tm, Price: testInput[10]}
		out := nma.UpdateTrade(&r)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		exp := expectedPri8Sec4LWMA[10]
		if math.Abs(sc.Value-exp) > 1e-13 {
			t.Errorf("expected %v, got %v", exp, sc.Value)
		}
	})
}

func TestNewMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	nma := testNmaCreate(0, 8, LWMA)

	// With default params: pri=32, sec=8. Warmup = 32 + 8 - 2 = 38 bars.
	for i := 0; i < 38; i++ {
		nma.Update(testInput[i])
		if nma.IsPrimed() {
			t.Errorf("expected not primed at bar %d", i)
		}
	}

	nma.Update(testInput[38])
	if !nma.IsPrimed() {
		t.Errorf("expected primed at bar 38")
	}
}

func TestNewMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	nma := testNmaCreate(0, 8, LWMA)
	m := nma.Metadata()

	if m.Identifier != core.NewMovingAverage {
		t.Errorf("expected NewMovingAverage identifier, got %v", m.Identifier)
	}

	if m.Mnemonic != "nma(32, 8, 3)" {
		t.Errorf("unexpected mnemonic: %s", m.Mnemonic)
	}

	if len(m.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(m.Outputs))
	}

	if m.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output, got %v", m.Outputs[0].Shape)
	}
}

func TestNewMovingAverageConstructorErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		_, err := NewNewMovingAverage(&NewMovingAverageParams{
			PrimaryPeriod:   32,
			SecondaryPeriod: 8,
			MAType:          LWMA,
			BarComponent:    entities.BarComponent(99),
		})

		if err == nil {
			t.Error("expected error for invalid bar component")
		}
	})
}

