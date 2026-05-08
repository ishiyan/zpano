//nolint:testpackage
package adaptivetrendandcyclefilter

import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

//nolint:funlen
func TestAdaptiveTrendAndCycleFilterUpdate(t *testing.T) {
	t.Parallel()

	input := testATCFInput()
	snaps := testATCFSnapshots()

	x, err := NewAdaptiveTrendAndCycleFilterDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0

	for i := range input {
		fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci := x.Update(input[i])

		if si < len(snaps) && snaps[si].i == i {
			s := snaps[si]
			if !closeEnough(s.fatl, fatl) {
				t.Errorf("[%d] fatl: expected %v, got %v", i, s.fatl, fatl)
			}

			if !closeEnough(s.satl, satl) {
				t.Errorf("[%d] satl: expected %v, got %v", i, s.satl, satl)
			}

			if !closeEnough(s.rftl, rftl) {
				t.Errorf("[%d] rftl: expected %v, got %v", i, s.rftl, rftl)
			}

			if !closeEnough(s.rstl, rstl) {
				t.Errorf("[%d] rstl: expected %v, got %v", i, s.rstl, rstl)
			}

			if !closeEnough(s.rbci, rbci) {
				t.Errorf("[%d] rbci: expected %v, got %v", i, s.rbci, rbci)
			}

			if !closeEnough(s.ftlm, ftlm) {
				t.Errorf("[%d] ftlm: expected %v, got %v", i, s.ftlm, ftlm)
			}

			if !closeEnough(s.stlm, stlm) {
				t.Errorf("[%d] stlm: expected %v, got %v", i, s.stlm, stlm)
			}

			if !closeEnough(s.pcci, pcci) {
				t.Errorf("[%d] pcci: expected %v, got %v", i, s.pcci, pcci)
			}

			si++
		}
	}

	if si != len(snaps) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(snaps), si)
	}
}

func TestAdaptiveTrendAndCycleFilterPrimesAtBar90(t *testing.T) {
	t.Parallel()

	x, _ := NewAdaptiveTrendAndCycleFilterDefault()

	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testATCFInput()
	primedAt := -1

	for i := range input {
		x.Update(input[i])

		if x.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}

	// IsPrimed mirrors RSTL (91-tap FIR) → first primed at i=90.
	if primedAt != 90 {
		t.Errorf("expected priming at index 90, got %d", primedAt)
	}
}

func TestAdaptiveTrendAndCycleFilterNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewAdaptiveTrendAndCycleFilterDefault()

	fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci := x.Update(math.NaN())

	for _, v := range []float64{fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci} {
		if !math.IsNaN(v) {
			t.Errorf("expected NaN output for NaN input, got %v", v)
		}
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

//nolint:funlen
func TestAdaptiveTrendAndCycleFilterMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewAdaptiveTrendAndCycleFilterDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mn := "atcf()"

	check("Identifier", core.AdaptiveTrendAndCycleFilter, md.Identifier)
	check("Mnemonic", mn, md.Mnemonic)
	check("Description", "Adaptive trend and cycle filter "+mn, md.Description)
	check("len(Outputs)", 8, len(md.Outputs))

	type o struct {
		kind int
		mn   string
	}

	expected := []o{
		{int(Fatl), "fatl()"},
		{int(Satl), "satl()"},
		{int(Rftl), "rftl()"},
		{int(Rstl), "rstl()"},
		{int(Rbci), "rbci()"},
		{int(Ftlm), "ftlm()"},
		{int(Stlm), "stlm()"},
		{int(Pcci), "pcci()"},
	}

	for i, e := range expected {
		check("Outputs[i].Kind", e.kind, md.Outputs[i].Kind)
		check("Outputs[i].Shape", shape.Scalar, md.Outputs[i].Shape)
		check("Outputs[i].Mnemonic", e.mn, md.Outputs[i].Mnemonic)
	}
}

//nolint:funlen
func TestAdaptiveTrendAndCycleFilterUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 100
		inp        = 100.
		outputLen  = 8
	)

	tm := testATCFTime()
	input := testATCFInput()

	check := func(act core.Output) {
		t.Helper()

		if len(act) != outputLen {
			t.Errorf("len(output): expected %v, actual %v", outputLen, len(act))

			return
		}

		for i := 0; i < outputLen; i++ {
			s, ok := act[i].(entities.Scalar)
			if !ok {
				t.Errorf("output[%d] is not a scalar", i)

				continue
			}

			if s.Time != tm {
				t.Errorf("output[%d].Time: expected %v, actual %v", i, tm, s.Time)
			}
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x, _ := NewAdaptiveTrendAndCycleFilterDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i])
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, Open: inp, High: inp, Low: inp, Close: inp}
		x, _ := NewAdaptiveTrendAndCycleFilterDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i])
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewAdaptiveTrendAndCycleFilterDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i])
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewAdaptiveTrendAndCycleFilterDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i])
		}

		check(x.UpdateTrade(&r))
	})
}
