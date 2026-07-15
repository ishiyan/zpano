package streamingkbn

import (
	"math"
	"math/rand/v2"
	"testing"
)

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// NaiveSum is a simple naive accumulator for comparison.
type NaiveSum struct {
	value float64
}

func (n *NaiveSum) Reset()           { n.value = 0 }
func (n *NaiveSum) Set(x float64)    { n.value = x }
func (n *NaiveSum) Update(x float64) { n.value += x }
func (n *NaiveSum) Value() float64   { return n.value }

func TestKleinKBNAccumulator_Peters(t *testing.T) {
	t.Parallel()

	data := []float64{1.0, 1e100, 1.0, -1e100}
	naive := &NaiveSum{}
	kbn := &KleinKBNAccumulator{}
	for _, x := range data {
		naive.Update(x)
		kbn.Update(x)
	}
	got := kbn.Value()
	if !almostEqual(got, 2.0, 1e-15) {
		t.Errorf("KBN sum = %v, want 2.0", got)
	}

	if math.Abs(kbn.Value()) <= math.Abs(naive.Value()) {
		t.Errorf("KBN sum %v is not more accurate than naive sum %v", kbn.Value(), naive.Value())
	}
}

func TestKleinKBNAccumulator_Numpy(t *testing.T) {
	t.Parallel()

	data := []float64{
		-0.41253261766461263,
		41287272281118.43,
		-1.4727977348624173e-14,
		5670.3302557520055,
		2.119245229045646e-11,
		-0.003679264134906428,
		-6.892634568678797e-14,
		-0.0006984744181630712,
		-4054136.048352595,
		-1003.101760720037,
		-1.4436349910427172e-17,
		-41287268231649.57,
	}
	expected := -0.377392919181026
	kbn := &KleinKBNAccumulator{}
	for _, x := range data {
		kbn.Update(x)
	}
	got := kbn.Value()
	if !almostEqual(got, expected, 1e-16) {
		t.Errorf("KBN sum = %v, want %v", got, expected)
	}
}

func TestKleinKBNAccumulator_BetterAccuracyThanNaive(t *testing.T) {
	t.Parallel()

	spread := 1e7
	naive := &NaiveSum{}
	kbn := &KleinKBNAccumulator{}

	rng := rand.New(rand.NewPCG(42, 0))
	for range 1000000 {
		x := rng.Float64() * spread
		naive.Update(x)
		kbn.Update(x)
	}

	rng2 := rand.New(rand.NewPCG(42, 0))
	for range 1000000 {
		x := rng2.Float64() * spread
		naive.Update(-x)
		kbn.Update(-x)
	}

	if math.Abs(kbn.Value()) > math.Abs(naive.Value()) {
		t.Errorf("KBN sum %v is not more accurate than naive sum %v", kbn.Value(), naive.Value())
	}
}

func TestKleinKBNAccumulator_Revert(t *testing.T) {
	t.Parallel()

	kbn := &KleinKBNAccumulator{}
	if !almostEqual(kbn.Value(), 0.0, 1e-15) {
		t.Fatalf("initial value = %v, want 0.0", kbn.Value())
	}

	kbn.Update(1.5)
	kbn.Update(2.5)
	expectedBefore := kbn.Value()

	kbn.Revert(2.5)
	if !almostEqual(kbn.Value(), 1.5, 1e-15) {
		t.Errorf("after revert 2.5: %v, want 1.5", kbn.Value())
	}

	kbn.Revert(1.5)
	if !almostEqual(kbn.Value(), 0.0, 1e-15) {
		t.Errorf("after revert 1.5: %v, want 0.0", kbn.Value())
	}

	_ = expectedBefore
}

func TestKleinKBNAccumulator_Reset(t *testing.T) {
	t.Parallel()

	kbn := &KleinKBNAccumulator{}
	kbn.Update(1.5)
	kbn.Reset()
	if !almostEqual(kbn.Value(), 0.0, 1e-15) {
		t.Errorf("after reset: %v, want 0.0", kbn.Value())
	}

	kbn.Update(1.5)
	if !almostEqual(kbn.Value(), 1.5, 1e-15) {
		t.Errorf("after update 1.5: %v, want 1.5", kbn.Value())
	}
}
