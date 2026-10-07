package streamingkbn

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func TestMomentsKleinKBN_LargeCount(t *testing.T) {
	t.Parallel()
	cases := []struct {
		n                                            int64
		skewFisher, skewSample                       float64
		kurtExcess, kurtSample, kurtCorrected        float64
		updatedMean, updatedM2, updatedM3, updatedM4 float64
	}{
		{65536, 1.154726968150136, 1.1547533982570972,
			-0.666625973768376, 2.333374026231624, 2.333511362318447,
			1.0000457756687062, 196616.999862673, 393215.99917605054, 1376264.9986268054},
		{4000000000, 1.1547005388122642, 1.154700539245277,
			-0.666666666, 2.333333334, 2.33333333625,
			1.00000000075, 12000000009, 24000000000, 84000000009},
	}
	for _, c := range cases {
		if strconv.IntSize == 32 && c.n > math.MaxInt32 {
			continue
		}
		t.Run(fmt.Sprint(c.n), func(t *testing.T) {
			// Exact state for n samples: three quarters zero, one quarter four.
			// Synthetic state reaches the count boundaries without billions of updates.
			// Expected values were computed with rational arithmetic and 60-digit roots.
			n := float64(c.n)
			r := NewRawMomentsKleinKBN(0, false, true)
			r.n = int(c.n)
			r.x1.Set(n)
			r.x2.Set(4 * n)
			r.x3.Set(16 * n)
			r.x4.Set(64 * n)
			r.mean.Set(1)
			r.s.Set(3 * n)
			assertAlmostEqual(t, "raw Fisher skewness", r.SkewnessFisher(), c.skewFisher, 13)
			assertAlmostEqual(t, "raw sample skewness", r.SkewnessSample(), c.skewSample, 13)
			assertAlmostEqual(t, "raw skewness dispatch", r.Skewness(), c.skewFisher, 13)
			assertAlmostEqual(t, "raw excess kurtosis", r.KurtosisSampleExcess(), c.kurtExcess, 13)
			assertAlmostEqual(t, "raw sample kurtosis", r.KurtosisSample(), c.kurtSample, 13)
			assertAlmostEqual(t, "raw corrected kurtosis", r.KurtosisSampleCorrected(), c.kurtCorrected, 13)
			assertAlmostEqual(t, "raw kurtosis dispatch", r.Kurtosis(), c.kurtExcess, 13)

			m := NewCentralMomentsKleinKBN(0, false, true)
			m.n = int(c.n)
			m.m1.Set(1)
			m.m2.Set(3 * n)
			m.m3.Set(6 * n)
			m.m4.Set(21 * n)
			assertAlmostEqual(t, "central skewness", m.Skewness(), c.skewFisher, 13)
			assertAlmostEqual(t, "central excess kurtosis", m.Kurtosis(), c.kurtExcess, 13)
			m.SetFisher(false)
			assertAlmostEqual(t, "central Pearson kurtosis", m.Kurtosis(), c.kurtSample, 13)
			m.Update(4)
			assertEqualInt(t, "updated count", m.N(), int(c.n)+1)
			for _, v := range []struct {
				name      string
				got, want float64
			}{
				{"updated mean", m.Mean(), c.updatedMean},
				{"updated M2", m.m2.Value(), c.updatedM2},
				{"updated M3", m.m3.Value(), c.updatedM3},
				{"updated M4", m.m4.Value(), c.updatedM4},
			} {
				if !almostEqual(v.got, v.want, 1e-14*math.Max(1, math.Abs(v.want))) {
					t.Errorf("%s = %.17g, want %.17g", v.name, v.got, v.want)
				}
			}
			m.Revert(4)
			assertEqualInt(t, "restored count", m.N(), int(c.n))
			assertAlmostEqual(t, "restored mean", m.Mean(), 1, 13)
			assertAlmostEqual(t, "restored variance", m.Variance(), 3, 13)
			assertAlmostEqual(t, "restored skewness", m.Skewness(), c.skewFisher, 13)
			assertAlmostEqual(t, "restored kurtosis", m.Kurtosis(), c.kurtSample, 13)
		})
	}
}

func TestMomentsKleinKBN_65536Samples(t *testing.T) {
	t.Parallel()
	r := NewRawMomentsKleinKBN(0, false, true)
	c := NewCentralMomentsKleinKBN(0, false, true)
	for i := 0; i < 65536; i++ {
		x := float64(i % 3)
		r.Update(x)
		c.Update(x)
	}
	// Python reference for the same input sequence.
	for _, v := range []struct {
		name      string
		got, want float64
	}{
		{"raw skewness", r.Skewness(), 2.8032934033032035e-05},
		{"central skewness", c.Skewness(), 2.8032934033030243e-05},
		{"raw kurtosis", r.Kurtosis(), -1.5000343327119314},
		{"central kurtosis", c.Kurtosis(), -1.5000343327119314},
	} {
		assertAlmostEqual(t, v.name, v.got, v.want, 13)
	}
}
