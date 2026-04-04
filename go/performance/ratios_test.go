package performance

import (
	"fmt"
	"math"
	"testing"
	"time"

	"portf_py/daycounting/conventions"
)

// epsilon is the tolerance for floating-point comparisons (13+ decimal places).
const epsilon = 1e-13

// Test data from 'Portfolio bacon' dataset from PerformanceAnalytics R package.
// See: https://www.rdocumentation.org/packages/PerformanceAnalytics/versions/2.0.4/topics/portfolio_bacon
var (
	baconDatesPrevious = []time.Time{
		d(2024, 6, 30), d(2024, 7, 1), d(2024, 7, 2), d(2024, 7, 3),
		d(2024, 7, 4), d(2024, 7, 5), d(2024, 7, 6), d(2024, 7, 7),
		d(2024, 7, 8), d(2024, 7, 9), d(2024, 7, 10), d(2024, 7, 11),
		d(2024, 7, 12), d(2024, 7, 13), d(2024, 7, 14), d(2024, 7, 15),
		d(2024, 7, 16), d(2024, 7, 17), d(2024, 7, 18), d(2024, 7, 19),
		d(2024, 7, 20), d(2024, 7, 21), d(2024, 7, 22), d(2024, 7, 23),
	}
	baconDates = []time.Time{
		d(2024, 7, 1), d(2024, 7, 2), d(2024, 7, 3), d(2024, 7, 4),
		d(2024, 7, 5), d(2024, 7, 6), d(2024, 7, 7), d(2024, 7, 8),
		d(2024, 7, 9), d(2024, 7, 10), d(2024, 7, 11), d(2024, 7, 12),
		d(2024, 7, 13), d(2024, 7, 14), d(2024, 7, 15), d(2024, 7, 16),
		d(2024, 7, 17), d(2024, 7, 18), d(2024, 7, 19), d(2024, 7, 20),
		d(2024, 7, 21), d(2024, 7, 22), d(2024, 7, 23), d(2024, 7, 24),
	}
	baconPortfolioReturns = []float64{
		0.003, 0.026, 0.011, -0.010,
		0.015, 0.025, 0.016, 0.067,
		-0.014, 0.040, -0.005, 0.081,
		0.040, -0.037, -0.061, 0.017,
		-0.049, -0.022, 0.070, 0.058,
		-0.065, 0.024, -0.005, -0.009,
	}
	baconBenchmarkReturns = []float64{
		0.002, 0.025, 0.018, -0.011,
		0.014, 0.018, 0.014, 0.065,
		-0.015, 0.042, -0.006, 0.083,
		0.039, -0.038, -0.062, 0.015,
		-0.048, 0.021, 0.060, 0.056,
		-0.067, 0.019, -0.003, 0.000,
	}
	baconLen = len(baconPortfolioReturns)
)

func d(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// useNil is a sentinel for expected nil values.
func isNil(v *float64) bool {
	return v == nil
}

// newRatiosWithRF creates a Ratios instance with de-annualized risk-free rate.
func newRatiosWithRF(rf float64) *Ratios {
	rfAnnual := math.Pow(1+rf, 252) - 1
	r := New(Daily, rfAnnual, 0, conventions.RAW)
	r.Reset()
	return r
}

// newRatiosWithMAR creates a Ratios instance with de-annualized target return.
func newRatiosWithMAR(mar float64) *Ratios {
	marAnnual := math.Pow(1+mar, 252) - 1
	r := New(Daily, 0, marAnnual, conventions.RAW)
	r.Reset()
	return r
}

// newRatiosWithRFandMAR creates a Ratios instance with de-annualized rf and target return.
func newRatiosWithRFandMAR(rf, mar float64) *Ratios {
	rfAnnual := math.Pow(1+rf, 252) - 1
	marAnnual := math.Pow(1+mar, 252) - 1
	r := New(Daily, rfAnnual, marAnnual, conventions.RAW)
	r.Reset()
	return r
}

// addBaconReturn adds a single bacon return at index i.
func addBaconReturn(r *Ratios, i int) {
	r.AddReturn(
		baconPortfolioReturns[i],
		baconBenchmarkReturns[i],
		1,
		baconDatesPrevious[i],
		baconDates[i],
	)
}

// ---------- Kurtosis ----------

func TestKurtosis(t *testing.T) {
	expectedKurtosisExcess := []*float64{
		nil, fp(-2.00000000000000000), fp(-1.50000000000000000),
		fp(-1.17592035552795000), fp(-0.94669079980875600), fp(-0.96028723389787100),
		fp(-0.57793300076120100), fp(0.78641242115027200), fp(0.59954237086621500),
		fp(-0.01187577489273160), fp(0.07517391430462480), fp(-0.27406990671095100),
		fp(-0.38022416153835900), fp(-0.31560370425738600), fp(-0.16235155227201600),
		fp(0.02528905226985100), fp(-0.33285099821964000), fp(-0.37425348407483000),
		fp(-0.58502674157514900), fp(-0.69334606360953100), fp(-0.77381631285861200),
		fp(-0.68208349704651200), fp(-0.61779722177118000), fp(-0.56754620589212500),
	}

	t.Run("conformance to R PerformanceAnalytics", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedKurtosisExcess[i]
			actual := ratios.Kurtosis()
			if expected == nil {
				if actual != nil {
					t.Errorf("step %d: expected nil, got %v", i, *actual)
				}
			} else {
				if actual == nil {
					t.Errorf("step %d: expected %v, got nil", i, *expected)
				} else if !almostEqual(*actual, *expected, epsilon) {
					t.Errorf("step %d: expected %.16f, got %.16f, diff=%.2e", i, *expected, *actual, math.Abs(*actual-*expected))
				}
			}
		}
	})
}

// ---------- Sharpe Ratio ----------

func TestSharpeRatio(t *testing.T) {
	expectedStdDev := map[float64][]*float64{
		0: {
			nil, fp(0.8915694197569510), fp(1.1419253390798400),
			fp(0.4977924836999790), fp(0.6680426571226850), fp(0.8511810078441020),
			fp(0.9735918376312110), fp(0.8462916062735410), fp(0.6475912629068400),
			fp(0.7524743687246650), fp(0.6702597534059590), fp(0.7244562693337180),
			fp(0.7945207458232130), fp(0.5805910371128360), fp(0.3566360956461000),
			fp(0.3758075293232440), fp(0.2578994439571370), fp(0.2131725662300710),
			fp(0.2880753096781920), fp(0.3448210835211740), fp(0.2337747541463060),
			fp(0.2546053055676570), fp(0.2430648040410730), fp(0.2275684556623890),
		},
		0.05: {
			nil, fp(-2.1828078897497800), fp(-3.1402946824695500),
			fp(-2.8208240742998800), fp(-3.0433054380033400), fp(-2.7967375972020500),
			fp(-2.9887005248213900), fp(-1.3662354689514000), fp(-1.4489272141296900),
			fp(-1.3494093427967500), fp(-1.4483773981646000), fp(-0.9801467173338540),
			fp(-0.9561181856516630), fp(-0.9946559628057110), fp(-1.0011155375243300),
			fp(-1.0290804307636500), fp(-1.0706734491553900), fp(-1.1284729554976500),
			fp(-0.9967676208113950), fp(-0.9275814386971810), fp(-0.9577955946576800),
			fp(-0.9630722427994000), fp(-0.9992664166133000), fp(-1.0367007424619900),
		},
		0.10: {
			nil, fp(-5.257185199256510), fp(-7.422514704018940),
			fp(-6.139440632299740), fp(-6.754653533129370), fp(-6.444656202248210),
			fp(-6.950992887274000), fp(-3.578762544176350), fp(-3.545445691166220),
			fp(-3.451293054318160), fp(-3.567014549735160), fp(-2.684749704001430),
			fp(-2.706757117126540), fp(-2.569902962724260), fp(-2.358867170694770),
			fp(-2.433968390850540), fp(-2.399246342267910), fp(-2.470118477225370),
			fp(-2.281610551300980), fp(-2.199983960915530), fp(-2.149365943461670),
			fp(-2.180749791166460), fp(-2.241597637267670), fp(-2.300969940586380),
		},
	}

	for _, rf := range []float64{0, 0.05, 0.10} {
		rf := rf
		t.Run(fmt.Sprintf("rf=%.2f", rf), func(t *testing.T) {
			ratios := newRatiosWithRF(rf)
			for i := 0; i < baconLen; i++ {
				addBaconReturn(ratios, i)
				expected := expectedStdDev[rf][i]
				actual := ratios.SharpeRatio(false, false)
				assertNullableFloat(t, i, expected, actual)
			}
		})
	}

	t.Run("ignore risk-free rate", func(t *testing.T) {
		for _, rf := range []float64{0, 0.05, 0.10} {
			ratios := newRatiosWithRF(rf)
			for i := 0; i < baconLen; i++ {
				addBaconReturn(ratios, i)
				expected := expectedStdDev[0][i]
				actual := ratios.SharpeRatio(true, false)
				assertNullableFloat(t, i, expected, actual)
			}
		}
	})
}

// ---------- Sortino Ratio ----------

func TestSortinoRatio(t *testing.T) {
	expectedMAR := map[float64][]*float64{
		0: {
			nil, nil, nil,
			fp(1.5), fp(2.01246117974981), fp(2.85773803324704),
			fp(3.25049446787935), fp(5.40936687607709), fp(2.69307029756515),
			fp(3.29008543386979), fp(2.92819766175444), fp(4.10863007844407),
			fp(4.56665101160337), fp(1.67730613630736), fp(0.691483512929973),
			fp(0.727302390567925), fp(0.452770753672167), fp(0.370054264368203),
			fp(0.536498400203865), fp(0.665303673385798), fp(0.401733515514418),
			fp(0.438224836666163), fp(0.418857174247308), fp(0.392372028795065),
		},
		0.05: {
			fp(-1), fp(-0.951329033501053), fp(-0.967821008377905),
			fp(-0.955961761235827), fp(-0.959422032420532), fp(-0.950640505399932),
			fp(-0.95521850710367), fp(-0.835987494907806), fp(-0.84620916319764),
			fp(-0.825850705880606), fp(-0.841892559996059), fp(-0.739594446201381),
			fp(-0.729168016460068), fp(-0.735413445987151), fp(-0.731283824494091),
			fp(-0.739823509257131), fp(-0.750430484501361), fp(-0.766429130761335),
			fp(-0.726278292165206), fp(-0.700204514919608), fp(-0.709303305303401),
			fp(-0.71078905810419), fp(-0.723223919287678), fp(-0.735374254070636),
		},
		0.10: {
			fp(-1), fp(-0.991075392350217), fp(-0.994004065367307),
			fp(-0.990197182430257), fp(-0.991346575643354), fp(-0.990116442284817),
			fp(-0.991246179848335), fp(-0.967496714088971), fp(-0.966414074414246),
			fp(-0.964235550350565), fp(-0.966082469415414), fp(-0.94189872360901),
			fp(-0.942394085487388), fp(-0.936339897881547), fp(-0.925395562084343),
			fp(-0.929178181525619), fp(-0.927078590338157), fp(-0.930569106181964),
			fp(-0.919801251226696), fp(-0.914287704738154), fp(-0.910539461748751),
			fp(-0.912598194566486), fp(-0.916559222172636), fp(-0.920182172840589),
		},
	}

	for _, mar := range []float64{0, 0.05, 0.10} {
		mar := mar
		t.Run(fmt.Sprintf("mar=%.2f", mar), func(t *testing.T) {
			ratios := newRatiosWithMAR(mar)
			for i := 0; i < baconLen; i++ {
				addBaconReturn(ratios, i)
				expected := expectedMAR[mar][i]
				actual := ratios.SortinoRatio(false, false)
				assertNullableFloat(t, i, expected, actual)
			}
		})
	}

	t.Run("Jack Schwager sqrt(2) version", func(t *testing.T) {
		for _, mar := range []float64{0, 0.05, 0.10} {
			ratios := newRatiosWithMAR(mar)
			for i := 0; i < baconLen; i++ {
				addBaconReturn(ratios, i)
				expected := expectedMAR[mar][i]
				actual := ratios.SortinoRatio(false, true)
				if expected == nil {
					if actual != nil {
						t.Errorf("mar=%.2f step %d: expected nil, got %v", mar, i, *actual)
					}
				} else {
					expectedDivided := *expected / sqrt2
					if actual == nil {
						t.Errorf("mar=%.2f step %d: expected %v, got nil", mar, i, expectedDivided)
					} else if !almostEqual(*actual, expectedDivided, epsilon) {
						t.Errorf("mar=%.2f step %d: expected %.16f, got %.16f", mar, i, expectedDivided, *actual)
					}
				}
			}
		}
	})
}

// ---------- Omega Ratio ----------

func TestOmegaRatio(t *testing.T) {
	expectedLossThreshold := map[float64][]*float64{
		0: {
			nil, nil, nil,
			fp(4.000000000000000), fp(5.500000000000000), fp(8.000000000000000),
			fp(9.600000000000000), fp(16.300000000000000), fp(6.791666666666670),
			fp(8.458333333333330), fp(7.000000000000000), fp(9.793103448275860),
			fp(11.172413793103400), fp(4.909090909090910), fp(2.551181102362210),
			fp(2.685039370078740), fp(1.937500000000000), fp(1.722222222222220),
			fp(2.075757575757580), fp(2.368686868686870), fp(1.783269961977190),
			fp(1.874524714828900), fp(1.839552238805970), fp(1.779783393501810),
		},
		0.02: {
			fp(0.00000000000000000), fp(0.35294117647058800), fp(0.23076923076923100),
			fp(0.10714285714285700), fp(0.09836065573770490), fp(0.18032786885245900),
			fp(0.16923076923076900), fp(0.89230769230769200), fp(0.58585858585858600),
			fp(0.78787878787878800), fp(0.62903225806451600), fp(1.12096774193548000),
			fp(1.28225806451613000), fp(0.87845303867403300), fp(0.60687022900763400),
			fp(0.60000000000000000), fp(0.47604790419161700), fp(0.42287234042553200),
			fp(0.55585106382978700), fp(0.65691489361702100), fp(0.53579175704989200),
			fp(0.54446854663774400), fp(0.51646090534979400), fp(0.48737864077669900),
		},
		0.04: {
			fp(0.00000000000000000), fp(0.00000000000000000), fp(0.00000000000000000),
			fp(0.00000000000000000), fp(0.00000000000000000), fp(0.00000000000000000),
			fp(0.00000000000000000), fp(0.13917525773195900), fp(0.10887096774193500),
			fp(0.10887096774193500), fp(0.09215017064846420), fp(0.23208191126279900),
			fp(0.23208191126279900), fp(0.18378378378378400), fp(0.14437367303609300),
			fp(0.13765182186234800), fp(0.11663807890223000), fp(0.10542635658914700),
			fp(0.15193798449612400), fp(0.17984496124031000), fp(0.15466666666666700),
			fp(0.15143603133159300), fp(0.14303329223181300), fp(0.13488372093023300),
		},
		0.06: {
			fp(0.00000000000000000), fp(0.00000000000000000), fp(0.00000000000000000),
			fp(0.00000000000000000), fp(0.00000000000000000), fp(0.00000000000000000),
			fp(0.00000000000000000), fp(0.02095808383233530), fp(0.01715686274509810),
			fp(0.01635514018691590), fp(0.01419878296146050), fp(0.05679513184584180),
			fp(0.05458089668615990), fp(0.04590163934426230), fp(0.03830369357045140),
			fp(0.03617571059431530), fp(0.03171007927519820), fp(0.02901554404145080),
			fp(0.03937823834196890), fp(0.03929679420889350), fp(0.03479853479853480),
			fp(0.03368794326241140), fp(0.03185247275775360), fp(0.03011093502377180),
		},
	}

	for _, l := range []float64{0, 0.02, 0.04, 0.06} {
		l := l
		t.Run(fmt.Sprintf("threshold=%.2f", l), func(t *testing.T) {
			ratios := newRatiosWithMAR(l)
			for i := 0; i < baconLen; i++ {
				addBaconReturn(ratios, i)
				expected := expectedLossThreshold[l][i]
				actual := ratios.OmegaRatio()
				assertNullableFloat(t, i, expected, actual)
			}
		})
	}
}

// ---------- Kappa Ratio ----------

func TestKappaRatio(t *testing.T) {
	// Order 1, MAR=0
	expectedOrder1Mar0 := []*float64{
		nil, nil, nil,
		fp(3.0000000000000000), fp(4.5000000000000000), fp(7.0000000000000000),
		fp(8.6000000000000000), fp(15.300000000000000), fp(5.7916666666666700),
		fp(7.4583333333333300), fp(6.0000000000000000), fp(8.7931034482758600),
		fp(10.172413793103400), fp(3.9090909090909090), fp(1.5511811023622000),
		fp(1.6850393700787400), fp(0.9375000000000000), fp(0.7222222222222220),
		fp(1.0757575757575800), fp(1.3686868686868700), fp(0.7832699619771860),
		fp(0.8745247148288970), fp(0.8395522388059700), fp(0.7797833935018050),
	}

	t.Run("order=1, mar=0", func(t *testing.T) {
		ratios := newRatiosWithMAR(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedOrder1Mar0[i]
			actual := ratios.KappaRatio(1)
			assertNullableFloat(t, i, expected, actual)
		}
	})

	// Order 2, MAR=0 (should equal Sortino)
	expectedOrder2Mar0 := []*float64{
		nil, nil, nil,
		fp(1.5000000000000000), fp(2.0124611797498100), fp(2.8577380332470400),
		fp(3.2504944678793500), fp(5.4093668760770900), fp(2.6930702975651500),
		fp(3.2900854338697900), fp(2.9281976617544400), fp(4.1086300784440700),
		fp(4.5666510116033700), fp(1.6773061363073600), fp(0.6914835129299730),
		fp(0.7273023905679250), fp(0.4527707536721670), fp(0.3700542643682030),
		fp(0.5364984002038650), fp(0.6653036733857980), fp(0.4017335155144180),
		fp(0.4382248366661630), fp(0.4188571742473080), fp(0.3923720287950650),
	}

	t.Run("order=2, mar=0", func(t *testing.T) {
		ratios := newRatiosWithMAR(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedOrder2Mar0[i]
			actual := ratios.KappaRatio(2)
			assertNullableFloat(t, i, expected, actual)
		}
	})

	// Order 3, MAR=0
	expectedOrder3Mar0 := []*float64{
		nil, nil, nil,
		fp(1.1905507889761500), fp(1.5389783520090300), fp(2.1199740249708300),
		fp(2.3501725959775100), fp(3.8250000000000000), fp(2.0689080079822300),
		fp(2.4835586338430600), fp(2.2408934899599800), fp(3.0989871337864400),
		fp(3.3988098734763700), fp(1.1713241279859900), fp(0.4942094486331960),
		fp(0.5142481946830330), fp(0.3389522803724070), fp(0.2803047018509310),
		fp(0.4027354737116720), fp(0.4951749226471700), fp(0.3070994714658920),
		fp(0.3324074590706010), fp(0.3156667962042520), fp(0.2944582876612480),
	}

	t.Run("order=3, mar=0", func(t *testing.T) {
		ratios := newRatiosWithMAR(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedOrder3Mar0[i]
			actual := ratios.KappaRatio(3)
			assertNullableFloat(t, i, expected, actual)
		}
	})

	// Kappa3Ratio should match KappaRatio(3)
	t.Run("Kappa3Ratio matches KappaRatio(3)", func(t *testing.T) {
		ratios := newRatiosWithMAR(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedOrder3Mar0[i]
			actual := ratios.Kappa3Ratio()
			assertNullableFloat(t, i, expected, actual)
		}
	})

	// Order 4, MAR=0
	expectedOrder4Mar0 := []*float64{
		nil, nil, nil,
		fp(1.0606601717798200), fp(1.3458139030991000), fp(1.8259320100855000),
		fp(1.9983654900858500), fp(3.2164287883454600), fp(1.8033735333115700),
		fp(2.1458818396425000), fp(1.9358196995813000), fp(2.6577517731212100),
		fp(2.8955073548113600), fp(0.9572325404178820), fp(0.4101535241803780),
		fp(0.4244948756831930), fp(0.2893122764499430), fp(0.2395667425326850),
		fp(0.3426567344926290), fp(0.4195093677307130), fp(0.2646839611869220),
		fp(0.2853879958557920), fp(0.2700286247384100), fp(0.2510732850424660),
	}

	t.Run("order=4, mar=0", func(t *testing.T) {
		ratios := newRatiosWithMAR(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedOrder4Mar0[i]
			actual := ratios.KappaRatio(4)
			assertNullableFloat(t, i, expected, actual)
		}
	})

	// Order 1 with various MARs
	expectedOrder1MARs := map[float64][]*float64{
		0.05: {
			fp(-1.0000000000000000), fp(-1.0000000000000000), fp(-1.0000000000000000),
			fp(-1.0000000000000000), fp(-1.0000000000000000), fp(-1.0000000000000000),
			fp(-1.0000000000000000), fp(-0.9356060606060610), fp(-0.9481707317073170),
			fp(-0.9497041420118340), fp(-0.9567430025445290), fp(-0.8778625954198470),
			fp(-0.8808933002481390), fp(-0.9020408163265300), fp(-0.9201331114808650),
			fp(-0.9242902208201890), fp(-0.9345156889495230), fp(-0.9403726708074530),
			fp(-0.9155279503105590), fp(-0.9055900621118010), fp(-0.9173913043478260),
			fp(-0.9196617336152220), fp(-0.9240759240759240), fp(-0.9283018867924530),
		},
	}

	for mar, expected := range expectedOrder1MARs {
		mar := mar
		expected := expected
		t.Run(fmt.Sprintf("order=1, mar=%.2f", mar), func(t *testing.T) {
			ratios := newRatiosWithMAR(mar)
			for i := 0; i < baconLen; i++ {
				addBaconReturn(ratios, i)
				actual := ratios.KappaRatio(1)
				assertNullableFloat(t, i, expected[i], actual)
			}
		})
	}
}

// ---------- Bernardo-Ledoit Ratio ----------

func TestBernardoLedoitRatio(t *testing.T) {
	expectedValues := []*float64{
		nil, nil, nil,
		fp(4.000000000000000), fp(5.500000000000000), fp(8.000000000000000),
		fp(9.600000000000000), fp(16.30000000000000), fp(6.791666666666670),
		fp(8.458333333333330), fp(7.000000000000000), fp(9.793103448275860),
		fp(11.17241379310340), fp(4.909090909090910), fp(2.551181102362200),
		fp(2.685039370078740), fp(1.937500000000000), fp(1.722222222222220),
		fp(2.075757575757580), fp(2.368686868686870), fp(1.783269961977190),
		fp(1.874524714828900), fp(1.839552238805970), fp(1.779783393501800),
	}

	t.Run("conformance to R", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedValues[i]
			actual := ratios.BernardoLedoitRatio()
			assertNullableFloat(t, i, expected, actual)
		}
	})
}

// ---------- Upside Potential Ratio ----------

func TestUpsidePotentialRatio(t *testing.T) {
	expectedFullMar0 := []*float64{
		nil, nil, nil,
		fp(2.0000000000000000), fp(2.4596747752497700), fp(3.2659863237109000),
		fp(3.6284589408885800), fp(5.7629202666703600), fp(3.1580608525404200),
		fp(3.7312142071260700), fp(3.4162306053801800), fp(4.5758860481494800),
		fp(5.0155760263033600), fp(2.1063844502464600), fp(1.1372622243112200),
		fp(1.1589257718862700), fp(0.9357262242558120), fp(0.8824370919549460),
		fp(1.0352152229285800), fp(1.1513927041252400), fp(0.9146263047391370),
		fp(0.9393254107670360), fp(0.9177626084618780), fp(0.8955528249813290),
	}

	t.Run("full=true, mar=0", func(t *testing.T) {
		ratios := newRatiosWithMAR(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedFullMar0[i]
			actual := ratios.UpsidePotentialRatio(true)
			assertNullableFloat(t, i, expected, actual)
		}
	})
}

// ---------- Cumulative Return ----------

func TestCumulativeReturn(t *testing.T) {
	expectedValues := []float64{
		0.00299999999999989, 0.02907799999999990, 0.04039785799999970,
		0.02999387941999990, 0.04544378761129960, 0.07157988230158210,
		0.08872516041840740, 0.16166974616644100, 0.14540636972011000,
		0.19122262450891500, 0.18526651138637000, 0.28127309880866600,
		0.33252402276101300, 0.28322063391885500, 0.20494417524980500,
		0.22542822622905200, 0.16538224314382800, 0.13974383379466400,
		0.21952590216029100, 0.29025840448558800, 0.20639160819402400,
		0.23534500679068100, 0.22916828175672800, 0.21810576722091700,
	}

	t.Run("conformance to R", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			actual := ratios.CumulativeReturn()
			if !almostEqual(actual, expectedValues[i], epsilon) {
				t.Errorf("step %d: expected %.16f, got %.16f, diff=%.2e", i, expectedValues[i], actual, math.Abs(actual-expectedValues[i]))
			}
		}
	})
}

// ---------- Drawdowns (Cumulative) ----------

func TestDrawdownsCumulative(t *testing.T) {
	expectedDrawdowns := []float64{
		0.000000000000000000, 0.000000000000000000, 0.000000000000000000,
		-0.009999999999999900, 0.000000000000000000, 0.000000000000000000,
		0.000000000000000000, 0.000000000000000000, -0.014000000000000000,
		0.000000000000000000, -0.005000000000000120, 0.000000000000000000,
		0.000000000000000000, -0.037000000000000100, -0.095743000000000000,
		-0.080370631000000200, -0.125432470081000000, -0.144672955739218000,
		-0.084800062640963400, -0.031718466274139200, -0.094656765966320100,
		-0.072928528349511800, -0.077563885707764200, -0.085865810736394400,
	}
	expectedWorst := 0.1446729557392180

	t.Run("conformance to R", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
		}
		actual := ratios.DrawdownsCumulative()
		for i, exp := range expectedDrawdowns {
			if !almostEqual(actual[i], exp, epsilon) {
				t.Errorf("step %d: expected %.18f, got %.18f", i, exp, actual[i])
			}
		}
		if !almostEqual(ratios.WorstDrawdownsCumulative(), expectedWorst, epsilon) {
			t.Errorf("worst drawdown: expected %.16f, got %.16f", expectedWorst, ratios.WorstDrawdownsCumulative())
		}
	})
}

// ---------- Calmar Ratio ----------

func TestCalmarRatio(t *testing.T) {
	expectedValues := []*float64{
		nil, nil, nil,
		fp(0.74155751780918500), fp(0.89279126631479400), fp(1.15889854414036000),
		fp(1.22179559465027000), fp(1.89088510302246000), fp(1.08562360529801000),
		fp(1.26085762243604000), fp(1.11225700971196000), fp(1.49066405029967000),
		fp(1.59487944362032000), fp(0.48572827216522300), fp(0.13062513296618000),
		fp(0.13355239428276700), fp(0.07209886266479390), fp(0.05041253535660620),
		fp(0.07257832783270360), fp(0.08863890501902830), fp(0.06203631318696950),
		fp(0.06672377010548700), fp(0.06228923867560830), fp(0.05705690600200920),
	}

	// Note: Python tests use places=12 for Calmar
	calmarEps := 1e-12

	t.Run("conformance to R", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedValues[i]
			actual := ratios.CalmarRatio()
			assertNullableFloatEps(t, i, expected, actual, calmarEps)
		}
	})
}

// ---------- Sterling Ratio ----------

func TestSterlingRatio(t *testing.T) {
	expectedExcess := map[float64][]*float64{
		0: {
			nil, nil, nil,
			fp(0.74155751780918500), fp(0.89279126631479400), fp(1.15889854414036000),
			fp(1.22179559465027000), fp(1.89088510302246000), fp(1.08562360529801000),
			fp(1.26085762243604000), fp(1.11225700971196000), fp(1.49066405029967000),
			fp(1.59487944362032000), fp(0.48572827216522300), fp(0.13062513296618000),
			fp(0.13355239428276700), fp(0.07209886266479390), fp(0.05041253535660620),
			fp(0.07257832783270360), fp(0.08863890501902830), fp(0.06203631318696950),
			fp(0.06672377010548700), fp(0.06228923867560830), fp(0.05705690600200920),
		},
		0.02: {
			fp(0.14999999999999500), fp(0.72174090072224500), fp(0.66442920035313400),
			fp(0.24718583926972700), fp(0.29759708877159600), fp(0.38629951471345200),
			fp(0.40726519821675300), fp(0.63029503434081700), fp(0.44702148453447300),
			fp(0.51917666806190000), fp(0.45798818046963000), fp(0.61380284424104300),
			fp(0.65671506502013300), fp(0.31529729947567100), fp(0.10805355058691200),
			fp(0.11047499102161600), fp(0.06218376425181400), fp(0.04428978919828240),
			fp(0.06376348297771220), fp(0.07787345727187060), fp(0.05450182606873110),
			fp(0.05861997797933450), fp(0.05472403303561820), fp(0.05012718208397050),
		},
	}

	// Note: Python tests use places=12 for Sterling
	sterlingEps := 1e-12

	for _, excess := range []float64{0, 0.02} {
		excess := excess
		t.Run(fmt.Sprintf("excess=%.2f", excess), func(t *testing.T) {
			excessAnnual := math.Pow(1+excess, 252) - 1
			ratios := New(Daily, 0, 0, conventions.RAW)
			ratios.Reset()
			for i := 0; i < baconLen; i++ {
				addBaconReturn(ratios, i)
				expected := expectedExcess[excess][i]
				actual := ratios.SterlingRatio(excessAnnual)
				assertNullableFloatEps(t, i, expected, actual, sterlingEps)
			}
		})
	}
}

// ---------- Burke Ratio ----------

func TestBurkeRatio(t *testing.T) {
	expectedUnmodifiedRf0 := []*float64{
		nil, nil, nil,
		fp(0.74155751780925900), fp(0.89279126631488400), fp(1.15889854414048000),
		fp(1.22179559465039000), fp(1.89088510302265000), fp(0.88340826476302900),
		fp(1.02600204980225000), fp(0.86912185514484500), fp(1.16481055500805000),
		fp(1.24624485947780000), fp(0.43717141205593600), fp(0.12556405008668100),
		fp(0.12837789439234000), fp(0.08147141226635310), fp(0.05962926105099300),
		fp(0.08584753824355520), fp(0.10484440763127000), fp(0.06479655403731050),
		fp(0.06969257444720940), fp(0.06501838430134020), fp(0.05929350254553110),
	}

	expectedModifiedRf0 := []*float64{
		nil, nil, nil,
		fp(1.4831150356185200), fp(1.9963419611982000), fp(2.8387100967984600),
		fp(3.2325672963992100), fp(5.3482307151677600), fp(2.6502247942890900),
		fp(3.2445033613766200), fp(2.8825510906130700), fp(4.0350221249328900),
		fp(4.4933997426306100), fp(1.6357456432054900), fp(0.4863074748680710),
		fp(0.5135115775693600), fp(0.3359152382424160), fp(0.2529855290778000),
		fp(0.3742007437554000), fp(0.4688784450484330), fp(0.2969351136482720),
		fp(0.3268871495298580), fp(0.3118172170272280), fp(0.2904776525979330),
	}

	// Note: Python tests use places=12 for Burke
	burkeEps := 1e-12

	t.Run("unmodified, rf=0", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedUnmodifiedRf0[i]
			actual := ratios.BurkeRatio(false)
			assertNullableFloatEps(t, i, expected, actual, burkeEps)
		}
	})

	t.Run("modified, rf=0", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedModifiedRf0[i]
			actual := ratios.BurkeRatio(true)
			assertNullableFloatEps(t, i, expected, actual, burkeEps)
		}
	})
}

// ---------- Drawdown Peaks ----------

func TestDrawdownPeaks(t *testing.T) {
	expectedPeaks := []float64{
		0.00000000000000000, 0.00000000000000000, 0.00000000000000000,
		-0.00999999999999890, 0.00000000000000000, 0.00000000000000000,
		0.00000000000000000, 0.00000000000000000, -0.01400000000000290,
		0.00000000000000000, -0.00499999999999945, 0.00000000000000000,
		0.00000000000000000, -0.03699999999999810, -0.09797742999999580,
		-0.08099408616309980, -0.12995439906088300, -0.15192580909309000,
		-0.08203215715946180, -0.02407973581061150, -0.08906408398233760,
		-0.06508545936249050, -0.07008220508951670, -0.07907589769106100,
	}

	t.Run("conformance to R", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
		}
		actualPeaks := ratios.DrawdownsPeaks()
		for i, exp := range expectedPeaks {
			if !almostEqual(actualPeaks[i], exp, epsilon) {
				t.Errorf("step %d: expected %.17f, got %.17f, diff=%.2e", i, exp, actualPeaks[i], math.Abs(actualPeaks[i]-exp))
			}
		}
	})
}

// ---------- Pain Index ----------

func TestPainIndex(t *testing.T) {
	expectedValues := []*float64{
		fp(0.000000000000000000), fp(0.000000000000000000), fp(0.000000000000000000),
		fp(0.002499999999999720), fp(0.001999999999999780), fp(0.001666666666666480),
		fp(0.001428571428571270), fp(0.001249999999999860), fp(0.002666666666666870),
		fp(0.002400000000000180), fp(0.002636363636363750), fp(0.002416666666666770),
		fp(0.002230769230769330), fp(0.004714285714285670), fp(0.010931828666666300),
		fp(0.015310719760193400), fp(0.022054465601410500), fp(0.029269540239837100),
		fp(0.032046520077712100), fp(0.031648180864357100), fp(0.034382271489022800),
		fp(0.035777870937816800), fp(0.037269363727021100), fp(0.039011302642189500),
	}

	t.Run("conformance to R", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedValues[i]
			actual := ratios.PainIndex()
			assertNullableFloat(t, i, expected, actual)
		}
	})
}

// ---------- Pain Ratio ----------

func TestPainRatio(t *testing.T) {
	expectedRf0 := []*float64{
		nil, nil, nil,
		fp(2.9662300712370400), fp(4.4639563315744200), fp(6.9533912648428800),
		fp(8.5525691625527300), fp(15.127080824181200), fp(5.6995239278141000),
		fp(7.3550027975430300), fp(5.9064682584701500), fp(8.6355710500115400),
		fp(10.009243404789200), fp(3.8122309845695300), fp(1.1440393448276400),
		fp(0.8351473403007020), fp(0.4100547525167660), fp(0.2491781707736400),
		fp(0.3276524622550170), fp(0.4051939805814540), fp(0.2610350161067210),
		fp(0.2698071401733870), fp(0.2417956028428860), fp(0.2115948629646200),
	}

	// Note: Python tests use places=12 for Pain Ratio
	painRatioEps := 1e-12

	t.Run("rf=0", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedRf0[i]
			actual := ratios.PainRatio()
			assertNullableFloatEps(t, i, expected, actual, painRatioEps)
		}
	})
}

// ---------- Ulcer Index ----------

func TestUlcerIndex(t *testing.T) {
	expectedValues := []*float64{
		fp(0.000000000000000000), fp(0.000000000000000000), fp(0.000000000000000000),
		fp(0.004999999999999450), fp(0.004472135954999090), fp(0.004082482904638180),
		fp(0.003779644730091860), fp(0.003535533905932350), fp(0.005734883511362320),
		fp(0.005440588203494720), fp(0.005402019824271570), fp(0.005172040216394730),
		fp(0.004969135507541710), fp(0.010987005311470400), fp(0.027434256917710300),
		fp(0.033400616370435100), fp(0.045203959104378100), fp(0.056676085534212600),
		fp(0.058286267631993700), fp(0.057065017555245000), fp(0.058983749023907400),
		fp(0.059274727347882400), fp(0.059785256332054100), fp(0.060711532990550200),
	}

	t.Run("conformance to R", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedValues[i]
			actual := ratios.UlcerIndex()
			assertNullableFloat(t, i, expected, actual)
		}
	})
}

// ---------- Martin Ratio ----------

func TestMartinRatio(t *testing.T) {
	expectedRf0 := []*float64{
		nil, nil, nil,
		fp(1.4831150356185200), fp(1.9963419611982000), fp(2.8387100967984600),
		fp(3.2325672963992100), fp(5.3482307151677600), fp(2.6502247942890900),
		fp(3.2445033613766200), fp(2.8825510906130700), fp(4.0350221249328900),
		fp(4.4933997426306100), fp(1.6357456432054900), fp(0.4558695408843890),
		fp(0.3828284707085000), fp(0.2000607604567100), fp(0.1286844429639630),
		fp(0.1801474281465850), fp(0.2247200286966700), fp(0.1521601617470090),
		fp(0.1628539762413560), fp(0.1507322845601700), fp(0.1359641377846650),
	}

	// Note: Python tests use places=12 for Martin Ratio
	martinEps := 1e-12

	t.Run("rf=0", func(t *testing.T) {
		ratios := newRatiosWithRF(0)
		for i := 0; i < baconLen; i++ {
			addBaconReturn(ratios, i)
			expected := expectedRf0[i]
			actual := ratios.MartinRatio()
			assertNullableFloatEps(t, i, expected, actual, martinEps)
		}
	})
}

// ---------- Helper functions ----------

// fp creates a pointer to a float64 value.
func fp(v float64) *float64 {
	return &v
}

// assertNullableFloat asserts that two nullable float64 pointers are equal within epsilon.
func assertNullableFloat(t *testing.T, step int, expected *float64, actual *float64) {
	t.Helper()
	assertNullableFloatEps(t, step, expected, actual, epsilon)
}

// assertNullableFloatEps asserts that two nullable float64 pointers are equal within the given tolerance.
func assertNullableFloatEps(t *testing.T, step int, expected *float64, actual *float64, eps float64) {
	t.Helper()
	if expected == nil {
		if actual != nil {
			t.Errorf("step %d: expected nil, got %v", step, *actual)
		}
		return
	}
	if actual == nil {
		t.Errorf("step %d: expected %.16f, got nil", step, *expected)
		return
	}
	if !almostEqual(*actual, *expected, eps) {
		t.Errorf("step %d: expected %.16f, got %.16f, diff=%.2e", step, *expected, *actual, math.Abs(*actual-*expected))
	}
}
