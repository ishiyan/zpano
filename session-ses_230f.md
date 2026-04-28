# New session - 2026-04-27T12:56:13.297Z

**Session ID:** ses_230fd010fffe6CF1dqcfg5mfOA
**Created:** 4/27/2026, 2:56:13 PM
**Updated:** 4/28/2026, 8:18:12 AM

---

## Assistant (Build · claude-opus-4.6 · 7.1s)

The issue is `meta.mnemonic()` and `meta.description()` — these are `[]const u8` fields, not functions. Remove the `()`.

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig",
  "oldString": "    try testing.expectEqualStrings(\"willr\", meta.mnemonic());\n    try testing.expectEqualStrings(\"Williams %R\", meta.description());",
  "newString": "    try testing.expectEqualStrings(\"willr\", meta.mnemonic);\n    try testing.expectEqualStrings(\"Williams %R\", meta.description);"
}
```

**Output:**
```
Edit applied successfully.
```

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/ultimate_oscillator/ultimate_oscillator.zig",
  "oldString": "    try testing.expectEqualStrings(\"ultosc(7, 14, 28)\", meta.mnemonic());",
  "newString": "    try testing.expectEqualStrings(\"ultosc(7, 14, 28)\", meta.mnemonic);"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 5.8s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | tail -20",
  "description": "Rebuild and test after fixes",
  "timeout": 120000
}
```

**Output:**
```
|  +- compile test Debug native cached 7ms MaxRSS:38M
+- run test 2 pass (2 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 15ms MaxRSS:39M
+- run test 6 pass (6 total) 6ms MaxRSS:10M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 7 pass (7 total) 7ms MaxRSS:11M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test success 5ms MaxRSS:11M
|  +- compile test Debug native cached 4ms MaxRSS:38M
+- run test success 10ms MaxRSS:10M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 11 pass (11 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 10 pass (10 total) 5ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 4 pass (4 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 269 pass (269 total) 40ms MaxRSS:11M
   +- compile test Debug native success 803ms MaxRSS:177M


```

---

## Assistant (Build · claude-opus-4.6 · 5.9s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | grep -E \"pass|fail|error:\" | grep -v cached",
  "description": "Count total passing tests"
}
```

**Output:**
```
Build Summary: 45/45 steps succeeded; 636/636 tests passed
+- run test 3 pass (3 total) 18ms MaxRSS:10M
+- run test 53 pass (53 total) 11ms MaxRSS:10M
+- run test 11 pass (11 total) 7ms MaxRSS:11M
+- run test 2 pass (2 total) 6ms MaxRSS:10M
+- run test 42 pass (42 total) 24ms MaxRSS:10M
+- run test 55 pass (55 total) 13ms MaxRSS:11M
+- run test 155 pass (155 total) 36ms MaxRSS:11M
+- run test 4 pass (4 total) 8ms MaxRSS:10M
+- run test 2 pass (2 total) 7ms MaxRSS:11M
+- run test 2 pass (2 total) 9ms MaxRSS:10M
+- run test 6 pass (6 total) 7ms MaxRSS:10M
+- run test 7 pass (7 total) 6ms MaxRSS:11M
+- run test 11 pass (11 total) 7ms MaxRSS:10M
+- run test 10 pass (10 total) 7ms MaxRSS:10M
+- run test 4 pass (4 total) 5ms MaxRSS:10M
+- run test 269 pass (269 total) 43ms MaxRSS:11M

```

---

## Assistant (Build · claude-opus-4.6 · 8.7s)

All **636/636 tests pass** (up from 620). Larry Williams family is complete.

Now moving to the next family: **marc_chaikin** (AdvanceDecline, AdvanceDeclineOscillator). Let me read the Go reference implementations.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Larry Williams family (WilliamsPercentR, UltimateOscillator) — compile & test",
      "status": "completed",
      "priority": "high"
    },
    {
      "content": "Marc Chaikin family (AdvanceDecline, AdvanceDeclineOscillator)",
      "status": "in_progress",
      "priority": "high"
    },
    {
      "content": "Tushar Chande family (3 indicators)",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "Welles Wilder family (12 indicators)",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "John Ehlers family (23 indicators)",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "Factory, frequency_response, cmd tools",
      "status": "pending",
      "priority": "medium"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Larry Williams family (WilliamsPercentR, UltimateOscillator) — compile & test",
    "status": "completed",
    "priority": "high"
  },
  {
    "content": "Marc Chaikin family (AdvanceDecline, AdvanceDeclineOscillator)",
    "status": "in_progress",
    "priority": "high"
  },
  {
    "content": "Tushar Chande family (3 indicators)",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "Welles Wilder family (12 indicators)",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "John Ehlers family (23 indicators)",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "Factory, frequency_response, cmd tools",
    "status": "pending",
    "priority": "medium"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 26.3s)

**Tool: task**

**Input:**
```json
{
  "description": "Read Marc Chaikin Go refs",
  "prompt": "Read ALL of these Go reference files and return their COMPLETE contents (every line):\n\n1. `go/indicators/marcchaikin/advancedecline/advancedecline.go`\n2. `go/indicators/marcchaikin/advancedecline/advancedecline_test.go`\n3. `go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator.go`\n4. `go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator_test.go`\n\nReturn the full content of all 4 files.",
  "subagent_type": "general"
}
```

**Output:**
```
task_id: ses_22d745acaffexCAizT0gntP84o (for resuming to continue this task if needed)

<task_result>
Here are the complete contents of all 4 files:

**1. `advancedecline.go`** (113 lines) — The A/D Line indicator. Computes cumulative CLV × Volume. Has `Update` (scalar, H=L=C so AD unchanged), `UpdateHLCV` (full HLCV), and `UpdateBar` (extracts HLCV from bar). Primed after first non-NaN update.

**2. `advancedecline_test.go`** (433 lines) — Tests with 252-entry HLCV test data from TA-Lib. Tests: full 252-bar verification at 2 decimal digits, TA-Lib spot checks, UpdateBar, scalar update (AD unchanged when H=L=C), NaN handling, initial not-primed state, metadata.

**3. `advancedeclineoscillator.go`** (200 lines) — The Chaikin Oscillator (ADOSC). Computes FastMA(AD) - SlowMA(AD) where MA can be EMA or SMA. Uses a `lineUpdater` interface for the MA components. Same HLCV pattern as AD, but feeds cumulative AD into both MAs.

**4. `advancedeclineoscillator_test.go`** (628 lines) — Tests with same 252-entry HLCV data plus expected EMA and SMA ADOSC outputs. Tests: full EMA(3,10) verification, full SMA(3,10) verification (first 9 NaN), TA-Lib spot checks for ADOSC(3,10) and ADOSC(5,2), UpdateBar, NaN handling, initial not-primed, metadata, invalid params (length < 2).
</task_result>
```

---

## Assistant (Build · claude-opus-4.6 · 6.7s)

The summary isn't enough — I need the actual code. Let me read the files directly.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedecline/advancedecline.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedecline/advancedecline.go</path>
<type>file</type>
<content>1: package advancedecline
2: 
3: import (
4: 	"math"
5: 	"sync"
6: 
7: 	"zpano/entities"
8: 	"zpano/indicators/core"
9: )
10: 
11: // AdvanceDecline is Marc Chaikin's Advance-Decline (A/D) Line.
12: //
13: // The Accumulation/Distribution Line is a cumulative indicator that uses volume
14: // and price to assess whether a stock is being accumulated or distributed.
15: // The A/D line seeks to identify divergences between the stock price and volume flow.
16: //
17: // The value is calculated as:
18: //
19: //	CLV = ((Close - Low) - (High - Close)) / (High - Low)
20: //	AD  = AD_previous + CLV × Volume
21: //
22: // When High equals Low, the A/D value is unchanged (no division by zero).
23: //
24: // Reference:
25: //
26: // Chaikin, Marc. "Chaikin Accumulation/Distribution Line".
27: type AdvanceDecline struct {
28: 	mu sync.RWMutex
29: 	core.LineIndicator
30: 	ad     float64
31: 	value  float64
32: 	primed bool
33: }
34: 
35: // NewAdvanceDecline returns an instance of the indicator created using supplied parameters.
36: func NewAdvanceDecline(_ *AdvanceDeclineParams) (*AdvanceDecline, error) {
37: 	mnemonic := "ad"
38: 	desc := "Advance-Decline"
39: 
40: 	barFunc, _ := entities.BarComponentFunc(entities.BarClosePrice)
41: 	quoteFunc, _ := entities.QuoteComponentFunc(entities.DefaultQuoteComponent)
42: 	tradeFunc, _ := entities.TradeComponentFunc(entities.DefaultTradeComponent)
43: 
44: 	a := &AdvanceDecline{
45: 		value: math.NaN(),
46: 	}
47: 
48: 	// For scalar/quote/trade updates, H=L=C so AD is unchanged.
49: 	a.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, a.Update)
50: 
51: 	return a, nil
52: }
53: 
54: // IsPrimed indicates whether the indicator is primed.
55: func (s *AdvanceDecline) IsPrimed() bool {
56: 	s.mu.RLock()
57: 	defer s.mu.RUnlock()
58: 
59: 	return s.primed
60: }
61: 
62: // Metadata describes the output data of the indicator.
63: func (s *AdvanceDecline) Metadata() core.Metadata {
64: 	return core.BuildMetadata(
65: 		core.AdvanceDecline,
66: 		s.LineIndicator.Mnemonic,
67: 		s.LineIndicator.Description,
68: 		[]core.OutputText{
69: 			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
70: 		},
71: 	)
72: }
73: 
74: // Update updates the indicator with the given sample.
75: // Since scalar updates use the same value for H, L, C, the range is 0 and AD is unchanged.
76: func (s *AdvanceDecline) Update(sample float64) float64 {
77: 	if math.IsNaN(sample) {
78: 		return math.NaN()
79: 	}
80: 
81: 	return s.UpdateHLCV(sample, sample, sample, 1)
82: }
83: 
84: // UpdateHLCV updates the indicator with the given high, low, close, and volume values.
85: func (s *AdvanceDecline) UpdateHLCV(high, low, close, volume float64) float64 {
86: 	if math.IsNaN(high) || math.IsNaN(low) || math.IsNaN(close) || math.IsNaN(volume) {
87: 		return math.NaN()
88: 	}
89: 
90: 	s.mu.Lock()
91: 	defer s.mu.Unlock()
92: 
93: 	temp := high - low
94: 	if temp > 0 {
95: 		s.ad += ((close - low) - (high - close)) / temp * volume
96: 	}
97: 
98: 	s.value = s.ad
99: 	s.primed = true
100: 
101: 	return s.value
102: }
103: 
104: // UpdateBar updates the indicator given the next bar sample.
105: // This shadows LineIndicator.UpdateBar to extract HLCV from the bar.
106: func (s *AdvanceDecline) UpdateBar(sample *entities.Bar) core.Output {
107: 	value := s.UpdateHLCV(sample.High, sample.Low, sample.Close, sample.Volume)
108: 
109: 	output := make([]any, 1)
110: 	output[0] = entities.Scalar{Time: sample.Time, Value: value}
111: 
112: 	return output
113: }

(End of file - total 113 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedecline/advancedecline_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedecline/advancedecline_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package advancedecline
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: // High test data, 252 entries. From TA-Lib excel-sma3-sma10-chaikin.csv.
16: func testHighs() []float64 {
17: 	return []float64{
18: 		93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,
19: 		96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
20: 		90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
21: 		87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,
22: 		84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
23: 		90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
24: 		90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,
25: 		89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
26: 		104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,
27: 		123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,
28: 		116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,
29: 		116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,
30: 		125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
31: 		138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
32: 		125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,
33: 		124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
34: 		124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,
35: 		132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,
36: 		129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,
37: 		123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,
38: 		108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,
39: 		96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
40: 		95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,
41: 		104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,
42: 		107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,
43: 		110.5, 109.5,
44: 	}
45: }
46: 
47: // Low test data, 252 entries.
48: func testLows() []float64 {
49: 	return []float64{
50: 		90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,
51: 		92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
52: 		87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
53: 		84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
54: 		82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,
55: 		89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,
56: 		88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,
57: 		86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
58: 		101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,
59: 		117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
60: 		112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
61: 		112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,
62: 		123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,
63: 		135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
64: 		122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,
65: 		121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,
66: 		122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,
67: 		128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,
68: 		126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,
69: 		119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,
70: 		105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
71: 		93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,
72: 		93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,
73: 		102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
74: 		104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,
75: 		108.56, 106.62,
76: 	}
77: }
78: 
79: // Close test data, 252 entries.
80: func testCloses() []float64 {
81: 	return []float64{
82: 		91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
83: 		96.125, 97.25, 98.5, 89.875, 91, 92.815, 89.155, 89.345, 91.625, 89.875,
84: 		88.375, 87.625, 84.78, 83, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
85: 		85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
86: 		83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89, 91, 90.5,
87: 		89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
88: 		88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90, 89.69,
89: 		88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106, 102.5,
90: 		102.405, 104.595, 106.125, 106, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
91: 		123, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
92: 		116, 116, 112, 113.75, 112.94, 116, 120.5, 116.62, 117, 115.25,
93: 		114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
94: 		123.12, 122.56, 124.62, 129.25, 131, 132.25, 131, 132.81, 134, 137.38,
95: 		137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129, 123.87, 124.81,
96: 		123, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
97: 		122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
98: 		124.44, 122, 122.37, 122.94, 124, 123.19, 124.56, 127.25, 125.87, 128.86,
99: 		132, 130.75, 134.75, 135, 132.38, 133.31, 131.94, 130, 125.37, 130.13,
100: 		127.12, 125.19, 122, 125, 123, 123.5, 120.06, 121, 117.75, 119.87,
101: 		122, 119.19, 116.37, 113.5, 114.25, 110, 105.06, 107, 107.87, 107,
102: 		107.12, 107, 91, 93.94, 93.87, 95.5, 93, 94.94, 98.25, 96.75,
103: 		94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97, 95, 95.87, 94.06,
104: 		94.62, 93.75, 98, 103.94, 107.87, 106.06, 104.5, 105, 104.19, 103.06,
105: 		103.42, 105.27, 111.87, 116, 116.62, 118.28, 113.37, 109, 109.7, 109.25,
106: 		107, 109.19, 110, 109.2, 110.12, 108, 108.62, 109.75, 109.81, 109,
107: 		108.75, 107.87,
108: 	}
109: }
110: 
111: // Volume test data, 252 entries.
112: func testVolumes() []float64 {
113: 	return []float64{
114: 		4077500, 4955900, 4775300, 4155300, 4593100, 3631300, 3382800, 4954200, 4500000, 3397500,
115: 		4204500, 6321400, 10203600, 19043900, 11692000, 9553300, 8920300, 5970900, 5062300, 3705600,
116: 		5865600, 5603000, 5811900, 8483800, 5995200, 5408800, 5430500, 6283800, 5834800, 4515500,
117: 		4493300, 4346100, 3700300, 4600200, 4557200, 4323600, 5237500, 7404100, 4798400, 4372800,
118: 		3872300, 10750800, 5804800, 3785500, 5014800, 3507700, 4298800, 4842500, 3952200, 3304700,
119: 		3462000, 7253900, 9753100, 5953000, 5011700, 5910800, 4916900, 4135000, 4054200, 3735300,
120: 		2921900, 2658400, 4624400, 4372200, 5831600, 4268600, 3059200, 4495500, 3425000, 3630800,
121: 		4168100, 5966900, 7692800, 7362500, 6581300, 19587700, 10378600, 9334700, 10467200, 5671400,
122: 		5645000, 4518600, 4519500, 5569700, 4239700, 4175300, 4995300, 4776600, 4190000, 6035300,
123: 		12168900, 9040800, 5780300, 4320800, 3899100, 3221400, 3455500, 4304200, 4703900, 8316300,
124: 		10553900, 6384800, 7163300, 7007800, 5114100, 5263800, 6666100, 7398400, 5575000, 4852300,
125: 		4298100, 4900500, 4887700, 6964800, 4679200, 9165000, 6469800, 6792000, 4423800, 5231900,
126: 		4565600, 6235200, 5225900, 8261400, 5912500, 3545600, 5714500, 6653900, 6094500, 4799200,
127: 		5050800, 5648900, 4726300, 5585600, 5124800, 7630200, 14311600, 8793600, 8874200, 6966600,
128: 		5525500, 6515500, 5291900, 5711700, 4327700, 4568000, 6859200, 5757500, 7367000, 6144100,
129: 		4052700, 5849700, 5544700, 5032200, 4400600, 4894100, 5140000, 6610900, 7585200, 5963100,
130: 		6045500, 8443300, 6464700, 6248300, 4357200, 4774700, 6216900, 6266900, 5584800, 5284500,
131: 		7554500, 7209500, 8424800, 5094500, 4443600, 4591100, 5658400, 6094100, 14862200, 7544700,
132: 		6985600, 8093000, 7590000, 7451300, 7078000, 7105300, 8778800, 6643900, 10563900, 7043100,
133: 		6438900, 8057700, 14240000, 17872300, 7831100, 8277700, 15017800, 14183300, 13921100, 9683000,
134: 		9187300, 11380500, 69447300, 26673600, 13768400, 11371600, 9872200, 9450500, 11083300, 9552800,
135: 		11108400, 10374200, 16701900, 13741900, 8523600, 9551900, 8680500, 7151700, 9673100, 6264700,
136: 		8541600, 8358000, 18720800, 19683100, 13682500, 10668100, 9710600, 3113100, 5682000, 5763600,
137: 		5340000, 6220800, 14680500, 9933000, 11329500, 8145300, 16644700, 12593800, 7138100, 7442300,
138: 		9442300, 7123600, 7680600, 4839800, 4775500, 4008800, 4533600, 3741100, 4084800, 2685200,
139: 		3438000, 2870500,
140: 	}
141: }
142: 
143: // Expected AD output, 252 entries. All valid (lookback = 0).
144: // From TA-Lib excel-sma3-sma10-chaikin.csv, column F (AD).
145: func testExpectedAD() []float64 {
146: 	return []float64{
147: 		-1631000.0000000000, 2974412.0226308300, -1239087.9773691700, -466727.3825736260, -2276567.4139707900,
148: 		789202.2581603610, -1398341.7418396400, 1134914.1775558300, -766075.9214540700, 1736082.9020753400,
149: 		5640261.4735039100, 2890166.0052259700, 8972764.3897170800, 188856.2575697040, 10267283.4207680000,
150: 		16222587.3168718000, 8293431.7613162900, 5920994.1613162700, 9699107.7642574600, 6364067.7642574600,
151: 		4106481.8240114100, 4732197.5600520200, -582684.7847755650, -5272487.2692476200, -3197626.7475084800,
152: 		-7210607.3926697700, -1780107.3926697700, 4138902.7196897800, 1901845.0121799000, -1812171.6280736600,
153: 		-3544134.5371645800, 297777.5278006210, -2159221.6721993900, 2440978.3278006100, 435810.3278006080,
154: 		-3070523.8185408500, 1465626.2894505000, 3007000.2671456800, 5345808.5051319400, 2847065.6479890900,
155: 		2348057.9160303300, -5679832.0403015400, -1127705.6433843900, -226396.1195748840, 519476.0323238490,
156: 		346091.6216774670, -1280481.3512955100, -6122981.3512955100, -2643824.5945387600, -4331118.5522426800,
157: 		-7585398.5522426800, -693439.5085836460, -9354192.3085836400, -14168892.5669816000, -15370043.8066510000,
158: 		-12414643.8066510000, -12585096.3399843000, -13890885.8136685000, -11450943.9372383000, -12314251.2598928000,
159: 		-14359581.2598928000, -12924045.2598928000, -8299645.2598927900, -11118563.6809454000, -6824818.5234191600,
160: 		-4130032.7842887200, -3708074.1635990700, 122721.5562452970, -3201046.9166611000, -4798598.9166611000,
161: 		-3061890.5833277700, -8623980.5426222100, -11300142.4820162000, -3937642.4820161500, 2643657.5179838500,
162: 		-5243139.1816861100, -5243139.1816861100, 2167108.4486456300, 2554782.5227197100, -686017.4772802920,
163: 		-3549128.9880716500, -1194197.2420399100, 3325302.7579600900, -709635.4642621180, 2925560.1041386400,
164: 		489968.4374719780, 5097434.5865403000, 3432561.5574946500, 5938908.5894581200, 9231738.4194426700,
165: 		21400638.4194427000, 20195198.4194427000, 22748227.3840560000, 21998336.4749651000, 21964073.7333131000,
166: 		19373669.6096018000, 17880959.5857354000, 15502576.1405453000, 11156172.5405452000, 19056383.7979514000,
167: 		25289392.9744220000, 27933099.2244220000, 25099853.7020340000, 29753682.2734625000, 28403559.8734625000,
168: 		33667359.8734625000, 36546987.7970902000, 31323445.5398723000, 26886728.6025174000, 27507371.6257732000,
169: 		27027380.7628290000, 29477630.7628290000, 27665337.5044020000, 30340032.5614362000, 29372997.8947696000,
170: 		31054151.7409234000, 33389894.2329362000, 27656016.3381994000, 27547234.3709863000, 25948598.2598752000,
171: 		21382998.2598752000, 16993417.4598752000, 19402230.7411252000, 21708493.1264463000, 26513515.7850566000,
172: 		30059115.7850566000, 25051948.0326135000, 28506411.7956798000, 29175540.8757985000, 33812970.0892816000,
173: 		36178148.3370158000, 37524688.4532948000, 37135945.1632515000, 37120557.8354279000, 36201961.6090128000,
174: 		31335687.6205735000, 27351074.1115125000, 31564674.1115125000, 24003855.7115125000, 27857719.5412998000,
175: 		23497936.8408778000, 28905358.6095853000, 33058030.1373631000, 31630105.1373631000, 32227029.2752941000,
176: 		28618418.2967738000, 25042729.8586399000, 20615046.1316058000, 26128856.6906120000, 25104840.0239454000,
177: 		23664422.5540658000, 19030671.4304703000, 23336961.2781861000, 18734557.7194672000, 19157015.3194672000,
178: 		22598860.1265889000, 27430460.1265889000, 21617613.3076656000, 21114727.1198203000, 17637753.4038082000,
179: 		21652343.2475581000, 19654310.4231306000, 19269265.5955444000, 15907319.4013187000, 16014463.6636138000,
180: 		12869354.1072998000, 13365911.6152870000, 14410394.9486203000, 16063495.7486204000, 15972904.3200490000,
181: 		21108562.0926708000, 14165259.0157478000, 16535274.6214973000, 13349288.2166634000, 9854321.9245286000,
182: 		13028415.7516891000, 7370015.7516890900, 3177620.7162280900, -5422742.0926145400, -197109.7155653600,
183: 		-3586572.4374588300, -7071058.5485699700, -13326553.0540645000, -5875253.0540644700, -8711657.4658291900,
184: 		-7411341.1259599000, -15815245.7522944000, -15486840.0488724000, -17500820.9841242000, -11021168.9841242000,
185: 		-8424094.7597804800, -15914043.5402683000, -21120435.7920843000, -17188529.7920843000, -20062126.7108518000,
186: 		-28339826.7108518000, -43050618.8743952000, -31637494.6556452000, -23284834.6556452000, -22588377.5790791000,
187: 		-21687661.8928045000, -33068161.8928045000, -33068161.8928045000, -23505927.9305405000, -21211194.5972072000,
188: 		-18079884.4522797000, -26587925.9068252000, -23734944.7747498000, -18697081.1383861000, -25274418.8433042000,
189: 		-28407557.3048426000, -31383762.2228754000, -36349191.9526051000, -49491445.4071505000, -40967845.4071505000,
190: 		-42041727.6556712000, -33715533.7781202000, -40160221.6108198000, -32749138.8415890000, -38061604.4415890000,
191: 		-32765812.4415890000, -32728665.7749224000, -16003686.6704448000, -3202381.6334077700, 6580605.8665922800,
192: 		-3756700.3349581000, 2909824.4299321800, 2711115.9192938800, 2964453.4989116900, 1753613.1627772300,
193: 		2119366.5874347800, 3939056.3833531200, 12663978.2900468000, 20113728.2900468000, 15914260.2900469000,
194: 		16003118.1082287000, 3135938.6756046000, -1177006.5298748600, -1073221.5541074900, 2154763.9880611900,
195: 		6436737.2438751500, 11752316.9393066000, 12759608.7425853000, 13137436.9953734000, 17912936.9953734000,
196: 		13904136.9953734000, 11103379.6620401000, 13849506.2577848000, 13390268.3879623000, 11472268.3879623000,
197: 		8707691.0683746700, 8328944.5405968900,
198: 	}
199: }
200: 
201: func testTime() time.Time {
202: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
203: }
204: 
205: func roundTo(v float64, digits int) float64 {
206: 	p := math.Pow(10, float64(digits))
207: 	return math.Round(v*p) / p
208: }
209: 
210: func TestAdvanceDeclineWithVolume(t *testing.T) {
211: 	t.Parallel()
212: 
213: 	const digits = 2
214: 
215: 	highs := testHighs()
216: 	lows := testLows()
217: 	closes := testCloses()
218: 	volumes := testVolumes()
219: 	expected := testExpectedAD()
220: 	count := len(highs)
221: 
222: 	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
223: 	if err != nil {
224: 		t.Fatal(err)
225: 	}
226: 
227: 	for i := 0; i < count; i++ {
228: 		v := ad.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
229: 		if math.IsNaN(v) {
230: 			t.Errorf("[%d] expected non-NaN, got NaN", i)
231: 			continue
232: 		}
233: 
234: 		if !ad.IsPrimed() {
235: 			t.Errorf("[%d] expected primed", i)
236: 		}
237: 
238: 		got := roundTo(v, digits)
239: 		exp := roundTo(expected[i], digits)
240: 
241: 		if got != exp {
242: 			t.Errorf("[%d] expected %v, got %v", i, exp, got)
243: 		}
244: 	}
245: }
246: 
247: func TestAdvanceDeclineTaLibSpotChecks(t *testing.T) {
248: 	t.Parallel()
249: 
250: 	const digits = 2
251: 
252: 	highs := testHighs()
253: 	lows := testLows()
254: 	closes := testCloses()
255: 	volumes := testVolumes()
256: 	count := len(highs)
257: 
258: 	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
259: 	if err != nil {
260: 		t.Fatal(err)
261: 	}
262: 
263: 	var values []float64
264: 	for i := 0; i < count; i++ {
265: 		v := ad.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
266: 		values = append(values, v)
267: 	}
268: 
269: 	// TA-Lib spot checks from test_per_hlcv.c.
270: 	spotChecks := []struct {
271: 		index    int
272: 		expected float64
273: 	}{
274: 		{0, -1631000.00},
275: 		{1, 2974412.02},
276: 		{250, 8707691.07},
277: 		{251, 8328944.54},
278: 	}
279: 
280: 	for _, sc := range spotChecks {
281: 		got := roundTo(values[sc.index], digits)
282: 		exp := roundTo(sc.expected, digits)
283: 
284: 		if got != exp {
285: 			t.Errorf("spot check [%d]: expected %v, got %v", sc.index, exp, got)
286: 		}
287: 	}
288: }
289: 
290: func TestAdvanceDeclineUpdateBar(t *testing.T) {
291: 	t.Parallel()
292: 
293: 	const digits = 2
294: 
295: 	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
296: 	if err != nil {
297: 		t.Fatal(err)
298: 	}
299: 
300: 	tm := testTime()
301: 
302: 	highs := testHighs()
303: 	lows := testLows()
304: 	closes := testCloses()
305: 	volumes := testVolumes()
306: 	expected := testExpectedAD()
307: 
308: 	for i := 0; i < 10; i++ {
309: 		bar := &entities.Bar{
310: 			Time:   tm,
311: 			Open:   highs[i], // Open is not used by AD.
312: 			High:   highs[i],
313: 			Low:    lows[i],
314: 			Close:  closes[i],
315: 			Volume: volumes[i],
316: 		}
317: 
318: 		output := ad.UpdateBar(bar)
319: 		scalar := output[0].(entities.Scalar)
320: 
321: 		got := roundTo(scalar.Value, digits)
322: 		exp := roundTo(expected[i], digits)
323: 
324: 		if got != exp {
325: 			t.Errorf("[%d] bar: expected %v, got %v", i, exp, got)
326: 		}
327: 	}
328: }
329: 
330: func TestAdvanceDeclineScalarUpdate(t *testing.T) {
331: 	t.Parallel()
332: 
333: 	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
334: 	if err != nil {
335: 		t.Fatal(err)
336: 	}
337: 
338: 	// Scalar update: H=L=C, so range=0, AD unchanged (remains 0 after primed).
339: 	v := ad.Update(100.0)
340: 	if v != 0 {
341: 		t.Errorf("expected 0 after scalar update, got %v", v)
342: 	}
343: 
344: 	if !ad.IsPrimed() {
345: 		t.Error("expected primed after first update")
346: 	}
347: }
348: 
349: func TestAdvanceDeclineNaN(t *testing.T) {
350: 	t.Parallel()
351: 
352: 	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
353: 	if err != nil {
354: 		t.Fatal(err)
355: 	}
356: 
357: 	if v := ad.Update(math.NaN()); !math.IsNaN(v) {
358: 		t.Errorf("expected NaN, got %v", v)
359: 	}
360: 
361: 	if v := ad.UpdateHLCV(math.NaN(), 1, 2, 3); !math.IsNaN(v) {
362: 		t.Errorf("expected NaN for NaN high, got %v", v)
363: 	}
364: 
365: 	if v := ad.UpdateHLCV(1, math.NaN(), 2, 3); !math.IsNaN(v) {
366: 		t.Errorf("expected NaN for NaN low, got %v", v)
367: 	}
368: 
369: 	if v := ad.UpdateHLCV(1, 2, math.NaN(), 3); !math.IsNaN(v) {
370: 		t.Errorf("expected NaN for NaN close, got %v", v)
371: 	}
372: 
373: 	if v := ad.UpdateHLCV(1, 2, 3, math.NaN()); !math.IsNaN(v) {
374: 		t.Errorf("expected NaN for NaN volume, got %v", v)
375: 	}
376: }
377: 
378: func TestAdvanceDeclineNotPrimedInitially(t *testing.T) {
379: 	t.Parallel()
380: 
381: 	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
382: 	if err != nil {
383: 		t.Fatal(err)
384: 	}
385: 
386: 	if ad.IsPrimed() {
387: 		t.Error("expected not primed initially")
388: 	}
389: 
390: 	if v := ad.Update(math.NaN()); !math.IsNaN(v) {
391: 		t.Errorf("expected NaN, got %v", v)
392: 	}
393: 
394: 	// Still not primed after NaN.
395: 	if ad.IsPrimed() {
396: 		t.Error("expected not primed after NaN")
397: 	}
398: }
399: 
400: func TestAdvanceDeclineMetadata(t *testing.T) {
401: 	t.Parallel()
402: 
403: 	ad, err := NewAdvanceDecline(&AdvanceDeclineParams{})
404: 	if err != nil {
405: 		t.Fatal(err)
406: 	}
407: 
408: 	meta := ad.Metadata()
409: 
410: 	if meta.Identifier != core.AdvanceDecline {
411: 		t.Errorf("expected identifier AdvanceDecline, got %v", meta.Identifier)
412: 	}
413: 
414: 	if meta.Mnemonic != "ad" {
415: 		t.Errorf("expected mnemonic 'ad', got '%v'", meta.Mnemonic)
416: 	}
417: 
418: 	if meta.Description != "Advance-Decline" {
419: 		t.Errorf("expected description 'Advance-Decline', got '%v'", meta.Description)
420: 	}
421: 
422: 	if len(meta.Outputs) != 1 {
423: 		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
424: 	}
425: 
426: 	if meta.Outputs[0].Kind != int(Value) {
427: 		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
428: 	}
429: 
430: 	if meta.Outputs[0].Shape != shape.Scalar {
431: 		t.Errorf("expected output type Scalar, got %v", meta.Outputs[0].Shape)
432: 	}
433: }

(End of file - total 433 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator.go</path>
<type>file</type>
<content>1: package advancedeclineoscillator
2: 
3: import (
4: 	"fmt"
5: 	"math"
6: 	"sync"
7: 
8: 	"zpano/entities"
9: 	"zpano/indicators/common/exponentialmovingaverage"
10: 	"zpano/indicators/common/simplemovingaverage"
11: 	"zpano/indicators/core"
12: )
13: 
14: // lineUpdater is an interface for indicators that accept a single scalar and return a value.
15: type lineUpdater interface {
16: 	Update(float64) float64
17: 	IsPrimed() bool
18: }
19: 
20: // AdvanceDeclineOscillator is Marc Chaikin's Advance-Decline (A/D) Oscillator.
21: //
22: // The Chaikin Oscillator is the difference between a fast and slow moving average
23: // of the Accumulation/Distribution Line. It is used to anticipate changes in the A/D Line
24: // by measuring the momentum behind accumulation/distribution movements.
25: //
26: // The value is calculated as:
27: //
28: //	CLV = ((Close - Low) - (High - Close)) / (High - Low)
29: //	AD  = AD_previous + CLV × Volume
30: //	ADOSC = FastMA(AD) - SlowMA(AD)
31: //
32: // When High equals Low, the A/D value is unchanged (no division by zero).
33: //
34: // Reference:
35: //
36: // Chaikin, Marc. "Chaikin Oscillator".
37: type AdvanceDeclineOscillator struct {
38: 	mu sync.RWMutex
39: 	core.LineIndicator
40: 	ad     float64
41: 	fastMA lineUpdater
42: 	slowMA lineUpdater
43: 	value  float64
44: 	primed bool
45: }
46: 
47: // NewAdvanceDeclineOscillator returns an instance of the indicator created using supplied parameters.
48: func NewAdvanceDeclineOscillator(p *AdvanceDeclineOscillatorParams) (*AdvanceDeclineOscillator, error) {
49: 	const (
50: 		invalid   = "invalid advance-decline oscillator parameters"
51: 		fmts      = "%s: %s"
52: 		fmtw      = "%s: %w"
53: 		minLength = 2
54: 	)
55: 
56: 	if p.FastLength < minLength {
57: 		return nil, fmt.Errorf(fmts, invalid, "fast length should be greater than 1")
58: 	}
59: 
60: 	if p.SlowLength < minLength {
61: 		return nil, fmt.Errorf(fmts, invalid, "slow length should be greater than 1")
62: 	}
63: 
64: 	var fastMA, slowMA lineUpdater
65: 
66: 	var maLabel string
67: 
68: 	switch p.MovingAverageType {
69: 	case SMA:
70: 		maLabel = "SMA"
71: 
72: 		fast, e := simplemovingaverage.NewSimpleMovingAverage(
73: 			&simplemovingaverage.SimpleMovingAverageParams{Length: p.FastLength})
74: 		if e != nil {
75: 			return nil, fmt.Errorf(fmtw, invalid, e)
76: 		}
77: 
78: 		slow, e := simplemovingaverage.NewSimpleMovingAverage(
79: 			&simplemovingaverage.SimpleMovingAverageParams{Length: p.SlowLength})
80: 		if e != nil {
81: 			return nil, fmt.Errorf(fmtw, invalid, e)
82: 		}
83: 
84: 		fastMA = fast
85: 		slowMA = slow
86: 	default:
87: 		maLabel = "EMA"
88: 
89: 		fast, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
90: 			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
91: 				Length:         p.FastLength,
92: 				FirstIsAverage: p.FirstIsAverage,
93: 			})
94: 		if e != nil {
95: 			return nil, fmt.Errorf(fmtw, invalid, e)
96: 		}
97: 
98: 		slow, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
99: 			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
100: 				Length:         p.SlowLength,
101: 				FirstIsAverage: p.FirstIsAverage,
102: 			})
103: 		if e != nil {
104: 			return nil, fmt.Errorf(fmtw, invalid, e)
105: 		}
106: 
107: 		fastMA = fast
108: 		slowMA = slow
109: 	}
110: 
111: 	mnemonic := fmt.Sprintf("adosc(%s%d/%s%d)", maLabel, p.FastLength, maLabel, p.SlowLength)
112: 	desc := "Chaikin Advance-Decline Oscillator " + mnemonic
113: 
114: 	barFunc, _ := entities.BarComponentFunc(entities.BarClosePrice)
115: 	quoteFunc, _ := entities.QuoteComponentFunc(entities.DefaultQuoteComponent)
116: 	tradeFunc, _ := entities.TradeComponentFunc(entities.DefaultTradeComponent)
117: 
118: 	a := &AdvanceDeclineOscillator{
119: 		fastMA: fastMA,
120: 		slowMA: slowMA,
121: 		value:  math.NaN(),
122: 	}
123: 
124: 	a.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, a.Update)
125: 
126: 	return a, nil
127: }
128: 
129: // IsPrimed indicates whether the indicator is primed.
130: func (s *AdvanceDeclineOscillator) IsPrimed() bool {
131: 	s.mu.RLock()
132: 	defer s.mu.RUnlock()
133: 
134: 	return s.primed
135: }
136: 
137: // Metadata describes the output data of the indicator.
138: func (s *AdvanceDeclineOscillator) Metadata() core.Metadata {
139: 	return core.BuildMetadata(
140: 		core.AdvanceDeclineOscillator,
141: 		s.LineIndicator.Mnemonic,
142: 		s.LineIndicator.Description,
143: 		[]core.OutputText{
144: 			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
145: 		},
146: 	)
147: }
148: 
149: // Update updates the indicator with the given sample.
150: // Since scalar updates use the same value for H, L, C, the range is 0 and AD is unchanged,
151: // but the unchanged AD value is still fed to the MAs.
152: func (s *AdvanceDeclineOscillator) Update(sample float64) float64 {
153: 	if math.IsNaN(sample) {
154: 		return math.NaN()
155: 	}
156: 
157: 	return s.UpdateHLCV(sample, sample, sample, 1)
158: }
159: 
160: // UpdateHLCV updates the indicator with the given high, low, close, and volume values.
161: func (s *AdvanceDeclineOscillator) UpdateHLCV(high, low, close, volume float64) float64 {
162: 	if math.IsNaN(high) || math.IsNaN(low) || math.IsNaN(close) || math.IsNaN(volume) {
163: 		return math.NaN()
164: 	}
165: 
166: 	s.mu.Lock()
167: 	defer s.mu.Unlock()
168: 
169: 	// Compute cumulative AD.
170: 	temp := high - low
171: 	if temp > 0 {
172: 		s.ad += ((close - low) - (high - close)) / temp * volume
173: 	}
174: 
175: 	// Feed AD to both MAs.
176: 	fast := s.fastMA.Update(s.ad)
177: 	slow := s.slowMA.Update(s.ad)
178: 	s.primed = s.fastMA.IsPrimed() && s.slowMA.IsPrimed()
179: 
180: 	if math.IsNaN(fast) || math.IsNaN(slow) {
181: 		s.value = math.NaN()
182: 
183: 		return s.value
184: 	}
185: 
186: 	s.value = fast - slow
187: 
188: 	return s.value
189: }
190: 
191: // UpdateBar updates the indicator given the next bar sample.
192: // This shadows LineIndicator.UpdateBar to extract HLCV from the bar.
193: func (s *AdvanceDeclineOscillator) UpdateBar(sample *entities.Bar) core.Output {
194: 	value := s.UpdateHLCV(sample.High, sample.Low, sample.Close, sample.Volume)
195: 
196: 	output := make([]any, 1)
197: 	output[0] = entities.Scalar{Time: sample.Time, Value: value}
198: 
199: 	return output
200: }

(End of file - total 200 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package advancedeclineoscillator
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: // High test data, 252 entries. From TA-Lib excel-sma3-sma10-chaikin.csv.
16: func testHighs() []float64 {
17: 	return []float64{
18: 		93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,
19: 		96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
20: 		90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
21: 		87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,
22: 		84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
23: 		90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
24: 		90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,
25: 		89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
26: 		104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,
27: 		123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,
28: 		116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,
29: 		116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,
30: 		125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
31: 		138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
32: 		125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,
33: 		124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
34: 		124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,
35: 		132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,
36: 		129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,
37: 		123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,
38: 		108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,
39: 		96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
40: 		95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,
41: 		104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,
42: 		107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,
43: 		110.5, 109.5,
44: 	}
45: }
46: 
47: // Low test data, 252 entries.
48: func testLows() []float64 {
49: 	return []float64{
50: 		90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,
51: 		92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
52: 		87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
53: 		84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
54: 		82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,
55: 		89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,
56: 		88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,
57: 		86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
58: 		101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,
59: 		117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
60: 		112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
61: 		112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,
62: 		123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,
63: 		135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
64: 		122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,
65: 		121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,
66: 		122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,
67: 		128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,
68: 		126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,
69: 		119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,
70: 		105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
71: 		93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,
72: 		93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,
73: 		102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
74: 		104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,
75: 		108.56, 106.62,
76: 	}
77: }
78: 
79: // Close test data, 252 entries.
80: func testCloses() []float64 {
81: 	return []float64{
82: 		91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
83: 		96.125, 97.25, 98.5, 89.875, 91, 92.815, 89.155, 89.345, 91.625, 89.875,
84: 		88.375, 87.625, 84.78, 83, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
85: 		85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
86: 		83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89, 91, 90.5,
87: 		89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
88: 		88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90, 89.69,
89: 		88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106, 102.5,
90: 		102.405, 104.595, 106.125, 106, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
91: 		123, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
92: 		116, 116, 112, 113.75, 112.94, 116, 120.5, 116.62, 117, 115.25,
93: 		114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
94: 		123.12, 122.56, 124.62, 129.25, 131, 132.25, 131, 132.81, 134, 137.38,
95: 		137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129, 123.87, 124.81,
96: 		123, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
97: 		122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
98: 		124.44, 122, 122.37, 122.94, 124, 123.19, 124.56, 127.25, 125.87, 128.86,
99: 		132, 130.75, 134.75, 135, 132.38, 133.31, 131.94, 130, 125.37, 130.13,
100: 		127.12, 125.19, 122, 125, 123, 123.5, 120.06, 121, 117.75, 119.87,
101: 		122, 119.19, 116.37, 113.5, 114.25, 110, 105.06, 107, 107.87, 107,
102: 		107.12, 107, 91, 93.94, 93.87, 95.5, 93, 94.94, 98.25, 96.75,
103: 		94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97, 95, 95.87, 94.06,
104: 		94.62, 93.75, 98, 103.94, 107.87, 106.06, 104.5, 105, 104.19, 103.06,
105: 		103.42, 105.27, 111.87, 116, 116.62, 118.28, 113.37, 109, 109.7, 109.25,
106: 		107, 109.19, 110, 109.2, 110.12, 108, 108.62, 109.75, 109.81, 109,
107: 		108.75, 107.87,
108: 	}
109: }
110: 
111: // Volume test data, 252 entries.
112: func testVolumes() []float64 {
113: 	return []float64{
114: 		4077500, 4955900, 4775300, 4155300, 4593100, 3631300, 3382800, 4954200, 4500000, 3397500,
115: 		4204500, 6321400, 10203600, 19043900, 11692000, 9553300, 8920300, 5970900, 5062300, 3705600,
116: 		5865600, 5603000, 5811900, 8483800, 5995200, 5408800, 5430500, 6283800, 5834800, 4515500,
117: 		4493300, 4346100, 3700300, 4600200, 4557200, 4323600, 5237500, 7404100, 4798400, 4372800,
118: 		3872300, 10750800, 5804800, 3785500, 5014800, 3507700, 4298800, 4842500, 3952200, 3304700,
119: 		3462000, 7253900, 9753100, 5953000, 5011700, 5910800, 4916900, 4135000, 4054200, 3735300,
120: 		2921900, 2658400, 4624400, 4372200, 5831600, 4268600, 3059200, 4495500, 3425000, 3630800,
121: 		4168100, 5966900, 7692800, 7362500, 6581300, 19587700, 10378600, 9334700, 10467200, 5671400,
122: 		5645000, 4518600, 4519500, 5569700, 4239700, 4175300, 4995300, 4776600, 4190000, 6035300,
123: 		12168900, 9040800, 5780300, 4320800, 3899100, 3221400, 3455500, 4304200, 4703900, 8316300,
124: 		10553900, 6384800, 7163300, 7007800, 5114100, 5263800, 6666100, 7398400, 5575000, 4852300,
125: 		4298100, 4900500, 4887700, 6964800, 4679200, 9165000, 6469800, 6792000, 4423800, 5231900,
126: 		4565600, 6235200, 5225900, 8261400, 5912500, 3545600, 5714500, 6653900, 6094500, 4799200,
127: 		5050800, 5648900, 4726300, 5585600, 5124800, 7630200, 14311600, 8793600, 8874200, 6966600,
128: 		5525500, 6515500, 5291900, 5711700, 4327700, 4568000, 6859200, 5757500, 7367000, 6144100,
129: 		4052700, 5849700, 5544700, 5032200, 4400600, 4894100, 5140000, 6610900, 7585200, 5963100,
130: 		6045500, 8443300, 6464700, 6248300, 4357200, 4774700, 6216900, 6266900, 5584800, 5284500,
131: 		7554500, 7209500, 8424800, 5094500, 4443600, 4591100, 5658400, 6094100, 14862200, 7544700,
132: 		6985600, 8093000, 7590000, 7451300, 7078000, 7105300, 8778800, 6643900, 10563900, 7043100,
133: 		6438900, 8057700, 14240000, 17872300, 7831100, 8277700, 15017800, 14183300, 13921100, 9683000,
134: 		9187300, 11380500, 69447300, 26673600, 13768400, 11371600, 9872200, 9450500, 11083300, 9552800,
135: 		11108400, 10374200, 16701900, 13741900, 8523600, 9551900, 8680500, 7151700, 9673100, 6264700,
136: 		8541600, 8358000, 18720800, 19683100, 13682500, 10668100, 9710600, 3113100, 5682000, 5763600,
137: 		5340000, 6220800, 14680500, 9933000, 11329500, 8145300, 16644700, 12593800, 7138100, 7442300,
138: 		9442300, 7123600, 7680600, 4839800, 4775500, 4008800, 4533600, 3741100, 4084800, 2685200,
139: 		3438000, 2870500,
140: 	}
141: }
142: 
143: // Expected EMA ADOSC output, 252 entries. All valid (no NaN).
144: func testExpectedEMA() []float64 {
145: 	return []float64{
146: 		0.00, 1465358.37, 590949.67, 425265.42, -257033.18,
147: 		462683.48, 19014.16, 641820.49, 233397.10, 841238.33,
148: 		2255663.07, 1754200.72, 3324956.45, 870389.71, 2993894.72,
149: 		5485298.48, 3482932.49, 1592286.02, 1876213.52, 760651.73,
150: 		-483187.42, -749013.21, -2480766.94, -4455897.26, -4198640.66,
151: 		-4988562.19, -3130319.89, -202233.93, 302213.24, -700628.51,
152: 		-1598267.56, -597759.92, -915892.00, 500925.97, 396986.59,
153: 		-797275.52, 229962.78, 1119728.10, 2126095.80, 1549455.69,
154: 		1013922.67, -1851662.43, -1407210.44, -810680.65, -255624.54,
155: 		-60485.30, -492702.89, -2165523.46, -1545988.96, -1688864.83,
156: 		-2629233.62, -582012.64, -2447296.98, -4519836.39, -5338983.69,
157: 		-4248372.70, -3470232.86, -3251905.68, -2090617.08, -1700177.07,
158: 		-2036676.87, -1532421.99, 284575.19, 105093.16, 1388306.41,
159: 		2644479.44, 3052220.74, 4160438.94, 3178016.51, 1978893.83,
160: 		1861033.22, -126123.22, -1779091.17, 49043.86, 2886508.75,
161: 		1275444.64, 500423.45, 2495682.35, 3208395.25, 2177123.69,
162: 		646329.51, 710634.76, 2110361.02, 1207281.33, 1884739.92,
163: 		1215581.10, 2297338.89, 2001294.85, 2495724.09, 3518825.17,
164: 		7489396.05, 8049317.22, 8358947.17, 7487108.03, 6438898.59,
165: 		4600511.73, 2955262.85, 1256787.86, -935242.86, 766742.96,
166: 		3376536.18, 4978400.28, 4279639.27, 5085488.01, 4523251.92,
167: 		5556886.97, 6390811.39, 4488943.25, 1891137.88, 853955.02,
168: 		199296.30, 692988.52, 255315.80, 904096.68, 779623.58,
169: 		1192740.49, 1996502.70, 319398.57, -430338.16, -1206583.93,
170: 		-2867140.48, -4682494.74, -4233017.73, -2930510.22, -602385.37,
171: 		1532937.10, 673930.45, 1360399.80, 1730460.17, 3200079.60,
172: 		4262927.93, 4738635.21, 4378766.66, 3828582.05, 2963172.99,
173: 		791402.75, -1436825.99, -877062.66, -2974050.55, -2335311.59,
174: 		-3248911.83, -1606758.47, 532405.07, 904777.20, 1164789.18,
175: 		17073.39, -1591717.88, -3513966.59, -2226495.29, -1823217.31,
176: 		-1950807.57, -3300032.80, -2181802.47, -2990400.05, -2914917.10,
177: 		-1523912.81, 720999.21, -275715.25, -818406.29, -2072325.51,
178: 		-1119529.74, -1263711.94, -1330326.92, -2306349.85, -2461872.43,
179: 		-3302405.40, -3188045.55, -2519104.05, -1490450.08, -962966.18,
180: 		974438.00, -530805.64, -344236.19, -1250341.08, -2619387.70,
181: 		-1931387.13, -3274751.71, -4860549.12, -7803895.39, -6635845.35,
182: 		-6633213.44, -7137817.47, -8685738.51, -6158498.16, -5467255.96,
183: 		-4273714.77, -6070898.05, -6149716.84, -6263707.84, -3679205.08,
184: 		-1461093.51, -2804023.07, -4755071.51, -3869887.46, -4070285.06,
185: 		-6416053.61, -11473115.20, -8867454.03, -4337705.01, -1868689.06,
186: 		-402164.91, -3386730.94, -4299805.14, -1239915.14, 854720.03,
187: 		2630240.37, 410372.74, 372705.18, 1926370.05, 294046.41,
188: 		-1397361.54, -2909242.58, -4843172.23, -9375662.99, -7665476.20,
189: 		-6610682.67, -2928961.57, -3207116.19, -671278.92, -1263195.54,
190: 		294517.43, 916808.73, 6409620.45, 12147129.54, 16502776.59,
191: 		13495236.91, 13159207.01, 11762223.26, 10202044.29, 8251063.31,
192: 		6819214.15, 6192521.32, 8149302.02, 10579332.57, 9275484.11,
193: 		7927138.67, 2560791.29, -1239632.59, -2648635.09, -1957173.88,
194: 		-133932.45, 2315435.53, 3427457.50, 3691006.62, 4982753.38,
195: 		3782690.38, 2056724.52, 2037440.17, 1698208.14, 794776.12,
196: 		-526700.32, -1139932.73,
197: 	}
198: }
199: 
200: // Expected SMA ADOSC output, 252 entries. First 9 are NaN, then 243 valid.
201: func testExpectedSMA() []float64 {
202: 	return []float64{
203: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
204: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), 815959.29,
205: 		1590615.58, 2817787.49, 4208829.41, 2326135.98, 3530790.03,
206: 		4404059.17, 6136406.99, 4209035.90, 988024.35, -117895.47,
207: 		-569354.95, -2409194.84, -3769234.10, -6349422.78, -7646206.54,
208: 		-7512194.60, -5340714.06, -2716995.09, 1100215.32, 1907151.18,
209: 		111200.39, 19953.24, 61923.58, 1285614.64, 968281.87,
210: 		250506.13, -399184.59, 571068.97, 3032116.73, 3026672.79,
211: 		2217806.09, -859646.47, -2287721.83, -2879135.73, -821066.28,
212: 		-671462.07, -748213.05, -2049367.35, -2247042.78, -2546103.42,
213: 		-2040230.18, -1888741.07, -2740450.32, -3540698.68, -6843948.13,
214: 		-6588025.09, -4929631.51, -3659788.40, -2457843.18, -1569248.21,
215: 		-1048061.76, -316034.96, 916712.34, 1692018.64, 2870571.69,
216: 		3431981.42, 5014442.37, 5928928.33, 5413600.58, 4298527.10,
217: 		2107220.65, -130430.37, -1997561.83, -3007571.25, -198539.50,
218: 		1931772.24, 1650113.17, 1286824.79, 3310549.08, 4418331.51,
219: 		2561642.87, 569004.62, 443566.84, 1067263.98, 2412326.17,
220: 		893903.97, 1795536.61, 1837991.78, 3315892.56, 3702218.25,
221: 		7196600.46, 9809757.50, 12372961.36, 10301396.85, 8987170.60,
222: 		5973947.89, 3323136.42, -37697.57, -3298589.65, -3889246.12,
223: 		-1015849.30, 3802669.52, 5581996.86, 6294558.71, 5807430.31,
224: 		7233896.67, 7631729.03, 7022937.31, 3189671.28, -668632.89,
225: 		-2274453.26, -1565272.36, -1769165.45, -723583.21, -855404.64,
226: 		535520.92, 1867850.83, 1662266.57, 427243.53, -1897311.12,
227: 		-3423878.90, -5693396.54, -7049208.37, -6077556.13, -2618241.81,
228: 		1033556.80, 2981836.38, 3561095.51, 3103739.89, 5237643.39,
229: 		6315373.90, 7045295.99, 6379582.91, 5152512.94, 3742759.40,
230: 		1681683.04, -1804724.14, -3656312.87, -5583088.33, -4818681.46,
231: 		-6239572.74, -3743805.45, -1602577.08, 1657190.95, 3161907.74,
232: 		1953764.06, -11193.28, -2786891.53, -3829245.49, -3533254.15,
233: 		-2533443.92, -3912036.95, -3529222.98, -3882956.52, -2533840.49,
234: 		-2177918.39, 481942.72, 1201885.33, 1208587.28, -1308939.63,
235: 		-1096155.05, -1645324.52, -694717.55, -2327001.67, -3226028.75,
236: 		-4386381.98, -3827063.06, -3641030.80, -2071193.78, -1035711.30,
237: 		1251389.20, 1167548.76, 1628404.62, -702216.88, -1523181.74,
238: 		-2708040.86, -4101542.10, -5203831.74, -9205593.91, -10110967.66,
239: 		-9896185.26, -8321992.32, -9712290.66, -8552730.08, -7242998.44,
240: 		-3227285.45, -4222090.20, -4614038.31, -6769390.38, -4088958.86,
241: 		-1250958.20, 162266.12, -2424767.88, -4214918.55, -4462566.02,
242: 		-4776181.10, -10673340.15, -12916730.67, -10653331.95, -2675863.99,
243: 		1967103.64, 421406.40, -1877082.43, -1851431.29, 2215797.93,
244: 		6185896.18, 3512294.23, 1880789.18, 1216281.60, 1922721.41,
245: 		1037173.44, -3360160.22, -6723648.25, -11153059.20, -12372088.51,
246: 		-11873416.10, -5902018.10, -3990282.48, 512452.89, 342481.23,
247: 		3243109.80, 3384424.51, 8702513.24, 13928083.80, 22276329.22,
248: 		22530155.27, 20905021.47, 15328057.73, 13997083.11, 9630157.59,
249: 		5944389.91, 2602485.32, 3372507.20, 7039016.78, 10097385.97,
250: 		9234450.73, 3552576.10, -1755700.59, -7044046.31, -7411219.62,
251: 		-5305042.09, -1811188.65, 1714196.56, 4645392.27, 6499064.62,
252: 		7090472.15, 5615708.94, 2758580.75, 1140942.22, 332154.68,
253: 		-1608879.10, -2953649.80,
254: 	}
255: }
256: 
257: func testTime() time.Time {
258: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
259: }
260: 
261: func roundTo(v float64, digits int) float64 {
262: 	p := math.Pow(10, float64(digits))
263: 	return math.Round(v*p) / p
264: }
265: 
266: func TestAdvanceDeclineOscillatorEMA(t *testing.T) {
267: 	t.Parallel()
268: 
269: 	const digits = 2
270: 
271: 	highs := testHighs()
272: 	lows := testLows()
273: 	closes := testCloses()
274: 	volumes := testVolumes()
275: 	expected := testExpectedEMA()
276: 	count := len(highs)
277: 
278: 	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
279: 		FastLength:        3,
280: 		SlowLength:        10,
281: 		MovingAverageType: EMA,
282: 	})
283: 	if err != nil {
284: 		t.Fatal(err)
285: 	}
286: 
287: 	// EMA with length 10 has lookback = 9. First 9 values are NaN.
288: 	for i := 0; i < count; i++ {
289: 		v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
290: 
291: 		if i < 9 {
292: 			if !math.IsNaN(v) {
293: 				t.Errorf("[%d] expected NaN, got %v", i, v)
294: 			}
295: 
296: 			if adosc.IsPrimed() {
297: 				t.Errorf("[%d] expected not primed", i)
298: 			}
299: 
300: 			continue
301: 		}
302: 
303: 		if math.IsNaN(v) {
304: 			t.Errorf("[%d] expected non-NaN, got NaN", i)
305: 			continue
306: 		}
307: 
308: 		if !adosc.IsPrimed() {
309: 			t.Errorf("[%d] expected primed", i)
310: 		}
311: 
312: 		got := roundTo(v, digits)
313: 		exp := roundTo(expected[i], digits)
314: 
315: 		if got != exp {
316: 			t.Errorf("[%d] expected %v, got %v", i, exp, got)
317: 		}
318: 	}
319: }
320: 
321: func TestAdvanceDeclineOscillatorSMA(t *testing.T) {
322: 	t.Parallel()
323: 
324: 	const digits = 2
325: 
326: 	highs := testHighs()
327: 	lows := testLows()
328: 	closes := testCloses()
329: 	volumes := testVolumes()
330: 	expected := testExpectedSMA()
331: 	count := len(highs)
332: 
333: 	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
334: 		FastLength:        3,
335: 		SlowLength:        10,
336: 		MovingAverageType: SMA,
337: 	})
338: 	if err != nil {
339: 		t.Fatal(err)
340: 	}
341: 
342: 	for i := 0; i < count; i++ {
343: 		v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
344: 
345: 		if i < 9 {
346: 			if !math.IsNaN(v) {
347: 				t.Errorf("[%d] expected NaN, got %v", i, v)
348: 			}
349: 
350: 			if adosc.IsPrimed() {
351: 				t.Errorf("[%d] expected not primed", i)
352: 			}
353: 
354: 			continue
355: 		}
356: 
357: 		if math.IsNaN(v) {
358: 			t.Errorf("[%d] expected non-NaN, got NaN", i)
359: 			continue
360: 		}
361: 
362: 		if !adosc.IsPrimed() {
363: 			t.Errorf("[%d] expected primed", i)
364: 		}
365: 
366: 		got := roundTo(v, digits)
367: 		exp := roundTo(expected[i], digits)
368: 
369: 		if got != exp {
370: 			t.Errorf("[%d] expected %v, got %v", i, exp, got)
371: 		}
372: 	}
373: }
374: 
375: func TestAdvanceDeclineOscillatorTaLibSpotChecks(t *testing.T) {
376: 	t.Parallel()
377: 
378: 	const digits = 2
379: 
380: 	highs := testHighs()
381: 	lows := testLows()
382: 	closes := testCloses()
383: 	volumes := testVolumes()
384: 	count := len(highs)
385: 
386: 	// ADOSC(3,10) EMA spot checks.
387: 	t.Run("ADOSC_3_10", func(t *testing.T) {
388: 		t.Parallel()
389: 
390: 		adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
391: 			FastLength:        3,
392: 			SlowLength:        10,
393: 			MovingAverageType: EMA,
394: 		})
395: 		if err != nil {
396: 			t.Fatal(err)
397: 		}
398: 
399: 		var values []float64
400: 		for i := 0; i < count; i++ {
401: 			v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
402: 			values = append(values, v)
403: 		}
404: 
405: 		// TA-Lib: begIndex=9, so output[0] corresponds to index 9.
406: 		spotChecks := []struct {
407: 			index    int
408: 			expected float64
409: 		}{
410: 			{9, 841238.33},
411: 			{10, 2255663.07},
412: 			{250, -526700.32},
413: 			{251, -1139932.73},
414: 		}
415: 
416: 		for _, sc := range spotChecks {
417: 			got := roundTo(values[sc.index], digits)
418: 			exp := roundTo(sc.expected, digits)
419: 
420: 			if got != exp {
421: 				t.Errorf("spot check [%d]: expected %v, got %v", sc.index, exp, got)
422: 			}
423: 		}
424: 	})
425: 
426: 	// ADOSC(5,2) EMA spot checks.
427: 	t.Run("ADOSC_5_2", func(t *testing.T) {
428: 		t.Parallel()
429: 
430: 		adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
431: 			FastLength:        5,
432: 			SlowLength:        2,
433: 			MovingAverageType: EMA,
434: 		})
435: 		if err != nil {
436: 			t.Fatal(err)
437: 		}
438: 
439: 		var values []float64
440: 		for i := 0; i < count; i++ {
441: 			v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
442: 			values = append(values, v)
443: 		}
444: 
445: 		// begIndex=4, output[0] at index 4.
446: 		got := roundTo(values[4], digits)
447: 		exp := roundTo(585361.29, digits)
448: 
449: 		if got != exp {
450: 			t.Errorf("ADOSC(5,2) spot check [4]: expected %v, got %v", exp, got)
451: 		}
452: 	})
453: }
454: 
455: func TestAdvanceDeclineOscillatorUpdateBar(t *testing.T) {
456: 	t.Parallel()
457: 
458: 	const digits = 2
459: 
460: 	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
461: 		FastLength:        3,
462: 		SlowLength:        10,
463: 		MovingAverageType: EMA,
464: 	})
465: 	if err != nil {
466: 		t.Fatal(err)
467: 	}
468: 
469: 	tm := testTime()
470: 
471: 	highs := testHighs()
472: 	lows := testLows()
473: 	closes := testCloses()
474: 	volumes := testVolumes()
475: 	expected := testExpectedEMA()
476: 
477: 	// Feed first 15 bars via UpdateBar. First 9 are NaN (lookback=9), then valid from index 9.
478: 	for i := 0; i < 15; i++ {
479: 		bar := &entities.Bar{
480: 			Time:   tm,
481: 			Open:   highs[i],
482: 			High:   highs[i],
483: 			Low:    lows[i],
484: 			Close:  closes[i],
485: 			Volume: volumes[i],
486: 		}
487: 
488: 		output := adosc.UpdateBar(bar)
489: 		scalar := output[0].(entities.Scalar)
490: 
491: 		if i < 9 {
492: 			if !math.IsNaN(scalar.Value) {
493: 				t.Errorf("[%d] bar: expected NaN, got %v", i, scalar.Value)
494: 			}
495: 
496: 			continue
497: 		}
498: 
499: 		got := roundTo(scalar.Value, digits)
500: 		exp := roundTo(expected[i], digits)
501: 
502: 		if got != exp {
503: 			t.Errorf("[%d] bar: expected %v, got %v", i, exp, got)
504: 		}
505: 	}
506: }
507: 
508: func TestAdvanceDeclineOscillatorNaN(t *testing.T) {
509: 	t.Parallel()
510: 
511: 	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
512: 		FastLength:        3,
513: 		SlowLength:        10,
514: 		MovingAverageType: EMA,
515: 	})
516: 	if err != nil {
517: 		t.Fatal(err)
518: 	}
519: 
520: 	if v := adosc.Update(math.NaN()); !math.IsNaN(v) {
521: 		t.Errorf("expected NaN, got %v", v)
522: 	}
523: 
524: 	if v := adosc.UpdateHLCV(math.NaN(), 1, 2, 3); !math.IsNaN(v) {
525: 		t.Errorf("expected NaN for NaN high, got %v", v)
526: 	}
527: 
528: 	if v := adosc.UpdateHLCV(1, math.NaN(), 2, 3); !math.IsNaN(v) {
529: 		t.Errorf("expected NaN for NaN low, got %v", v)
530: 	}
531: 
532: 	if v := adosc.UpdateHLCV(1, 2, math.NaN(), 3); !math.IsNaN(v) {
533: 		t.Errorf("expected NaN for NaN close, got %v", v)
534: 	}
535: 
536: 	if v := adosc.UpdateHLCV(1, 2, 3, math.NaN()); !math.IsNaN(v) {
537: 		t.Errorf("expected NaN for NaN volume, got %v", v)
538: 	}
539: }
540: 
541: func TestAdvanceDeclineOscillatorNotPrimedInitially(t *testing.T) {
542: 	t.Parallel()
543: 
544: 	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
545: 		FastLength:        3,
546: 		SlowLength:        10,
547: 		MovingAverageType: EMA,
548: 	})
549: 	if err != nil {
550: 		t.Fatal(err)
551: 	}
552: 
553: 	if adosc.IsPrimed() {
554: 		t.Error("expected not primed initially")
555: 	}
556: 
557: 	if v := adosc.Update(math.NaN()); !math.IsNaN(v) {
558: 		t.Errorf("expected NaN, got %v", v)
559: 	}
560: 
561: 	// Still not primed after NaN.
562: 	if adosc.IsPrimed() {
563: 		t.Error("expected not primed after NaN")
564: 	}
565: }
566: 
567: func TestAdvanceDeclineOscillatorMetadata(t *testing.T) {
568: 	t.Parallel()
569: 
570: 	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
571: 		FastLength:        3,
572: 		SlowLength:        10,
573: 		MovingAverageType: EMA,
574: 	})
575: 	if err != nil {
576: 		t.Fatal(err)
577: 	}
578: 
579: 	meta := adosc.Metadata()
580: 
581: 	if meta.Identifier != core.AdvanceDeclineOscillator {
582: 		t.Errorf("expected identifier AdvanceDeclineOscillator, got %v", meta.Identifier)
583: 	}
584: 
585: 	if meta.Mnemonic != "adosc(EMA3/EMA10)" {
586: 		t.Errorf("expected mnemonic 'adosc(EMA3/EMA10)', got '%v'", meta.Mnemonic)
587: 	}
588: 
589: 	if meta.Description != "Chaikin Advance-Decline Oscillator adosc(EMA3/EMA10)" {
590: 		t.Errorf("expected description 'Chaikin Advance-Decline Oscillator adosc(EMA3/EMA10)', got '%v'", meta.Description)
591: 	}
592: 
593: 	if len(meta.Outputs) != 1 {
594: 		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
595: 	}
596: 
597: 	if meta.Outputs[0].Kind != int(Value) {
598: 		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
599: 	}
600: 
601: 	if meta.Outputs[0].Shape != shape.Scalar {
602: 		t.Errorf("expected output type Scalar, got %v", meta.Outputs[0].Shape)
603: 	}
604: }
605: 
606: func TestAdvanceDeclineOscillatorInvalidParams(t *testing.T) {
607: 	t.Parallel()
608: 
609: 	// Fast length < 2.
610: 	_, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
611: 		FastLength:        1,
612: 		SlowLength:        10,
613: 		MovingAverageType: EMA,
614: 	})
615: 	if err == nil {
616: 		t.Error("expected error for fast length < 2")
617: 	}
618: 
619: 	// Slow length < 2.
620: 	_, err = NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
621: 		FastLength:        3,
622: 		SlowLength:        1,
623: 		MovingAverageType: EMA,
624: 	})
625: 	if err == nil {
626: 		t.Error("expected error for slow length < 2")
627: 	}
628: }

(End of file - total 628 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 8.0s)

Now let me check the Go params files and see an existing Zig indicator for reference pattern (I'll use the simple BalanceOfPower which also overrides UpdateBar for HLCV).

**Tool: glob**

**Input:**
```json
{
  "pattern": "go/indicators/marcchaikin/advancedecline/params.go"
}
```

**Output:**
```
/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedecline/params.go
```

**Tool: glob**

**Input:**
```json
{
  "pattern": "go/indicators/marcchaikin/advancedeclineoscillator/params.go"
}
```

**Output:**
```
/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedeclineoscillator/params.go
```

---

## Assistant (Build · claude-opus-4.6 · 5.2s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedecline/params.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedecline/params.go</path>
<type>file</type>
<content>1: package advancedecline
2: 
3: // AdvanceDeclineParams describes parameters to create an instance of the indicator.
4: // Advance-Decline requires HLCV bar data and has no configurable parameters.
5: type AdvanceDeclineParams struct{}
6: 
7: // DefaultParams returns a [AdvanceDeclineParams] value populated with conventional defaults.
8: func DefaultParams() *AdvanceDeclineParams { return &AdvanceDeclineParams{} }

(End of file - total 8 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedeclineoscillator/params.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/marcchaikin/advancedeclineoscillator/params.go</path>
<type>file</type>
<content>1: package advancedeclineoscillator
2: 
3: // MovingAverageType specifies the type of moving average to use in the ADOSC calculation.
4: type MovingAverageType int
5: 
6: const (
7: 	// SMA uses a Simple Moving Average.
8: 	SMA MovingAverageType = iota
9: 
10: 	// EMA uses an Exponential Moving Average.
11: 	EMA
12: )
13: 
14: // AdvanceDeclineOscillatorParams describes parameters to create an instance of the indicator.
15: type AdvanceDeclineOscillatorParams struct {
16: 	// FastLength is the number of periods for the fast moving average.
17: 	//
18: 	// The value should be greater than 1. Default is 3.
19: 	FastLength int
20: 
21: 	// SlowLength is the number of periods for the slow moving average.
22: 	//
23: 	// The value should be greater than 1. Default is 10.
24: 	SlowLength int
25: 
26: 	// MovingAverageType specifies the type of moving average (SMA or EMA).
27: 	//
28: 	// If zero (SMA), the Simple Moving Average is used.
29: 	// Use EMA for the Exponential Moving Average (TaLib default).
30: 	MovingAverageType MovingAverageType
31: 
32: 	// FirstIsAverage controls the EMA seeding algorithm.
33: 	// When true, the first EMA value is the simple average of the first period values.
34: 	// When false (default), the first input value is used directly (Metastock style).
35: 	// Only relevant when MovingAverageType is EMA.
36: 	FirstIsAverage bool
37: }
38: 
39: // DefaultParams returns a [AdvanceDeclineOscillatorParams] value populated with conventional defaults.
40: func DefaultParams() *AdvanceDeclineOscillatorParams {
41: 	return &AdvanceDeclineOscillatorParams{
42: 		FastLength: 3,
43: 		SlowLength: 10,
44: 	}
45: }

(End of file - total 45 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 6.1s)

Now let me look at how BalanceOfPower (another HLCV-based indicator) is structured in Zig since it also overrides bar updates:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/igor_livshin/balance_of_power/balance_of_power.zig",
  "limit": 100
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/igor_livshin/balance_of_power/balance_of_power.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: const bar_component = @import("bar_component");
9: const quote_component = @import("quote_component");
10: const trade_component = @import("trade_component");
11: 
12: const indicator_mod = @import("../../core/indicator.zig");
13: const line_indicator_mod = @import("../../core/line_indicator.zig");
14: const build_metadata_mod = @import("../../core/build_metadata.zig");
15: const identifier_mod = @import("../../core/identifier.zig");
16: const metadata_mod = @import("../../core/metadata.zig");
17: 
18: const OutputArray = indicator_mod.OutputArray;
19: const LineIndicator = line_indicator_mod.LineIndicator;
20: const Identifier = identifier_mod.Identifier;
21: const Metadata = metadata_mod.Metadata;
22: 
23: const epsilon = 1e-8;
24: 
25: /// Enumerates the outputs of the Balance of Power indicator.
26: pub const BalanceOfPowerOutput = enum(u8) {
27:     /// The scalar value of the balance of power.
28:     value = 1,
29: };
30: 
31: /// Igor Livshin's Balance of Power (BOP).
32: ///
33: /// The Balance of Market Power captures the struggles of bulls vs. bears
34: /// throughout the trading day. It assigns scores to both bulls and bears
35: /// based on how much they were able to move prices throughout the trading day.
36: ///
37: /// BOP = (Close - Open) / (High - Low)
38: ///
39: /// When the range (High - Low) is less than epsilon, the value is 0.
40: /// The indicator is always primed.
41: pub const BalanceOfPower = struct {
42:     line: LineIndicator,
43:     value: f64,
44: 
45:     const mnemonic_str = "bop";
46:     const description_str = "Balance of Power";
47: 
48:     pub fn init() BalanceOfPower {
49:         return .{
50:             .line = LineIndicator.new(
51:                 mnemonic_str,
52:                 description_str,
53:                 null,
54:                 null,
55:                 null,
56:             ),
57:             .value = math.nan(f64),
58:         };
59:     }
60: 
61:     /// Core update logic. For scalar/quote/trade, O=H=L=C so BOP is always 0.
62:     pub fn update(self: *BalanceOfPower, sample: f64) f64 {
63:         if (math.isNan(sample)) {
64:             return sample;
65:         }
66:         return self.updateOhlc(sample, sample, sample, sample);
67:     }
68: 
69:     /// Updates the indicator with the given OHLC values.
70:     pub fn updateOhlc(self: *BalanceOfPower, open: f64, high: f64, low: f64, close: f64) f64 {
71:         if (math.isNan(open) or math.isNan(high) or math.isNan(low) or math.isNan(close)) {
72:             return math.nan(f64);
73:         }
74: 
75:         const r = high - low;
76:         if (r < epsilon) {
77:             self.value = 0;
78:         } else {
79:             self.value = (close - open) / r;
80:         }
81: 
82:         return self.value;
83:     }
84: 
85:     pub fn isPrimed(_: *const BalanceOfPower) bool {
86:         return true;
87:     }
88: 
89:     pub fn getMetadata(self: *const BalanceOfPower, out: *Metadata) void {
90:         build_metadata_mod.buildMetadata(
91:             out,
92:             .balance_of_power,
93:             self.line.mnemonic,
94:             self.line.description,
95:             &[_]build_metadata_mod.OutputText{
96:                 .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
97:             },
98:         );
99:     }
100: 

(Showing lines 1-100 of 303. Use offset=101 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 5.1s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/igor_livshin/balance_of_power/balance_of_power.zig",
  "offset": 100,
  "limit": 80
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/igor_livshin/balance_of_power/balance_of_power.zig</path>
<type>file</type>
<content>100: 
101:     pub fn updateScalar(self: *BalanceOfPower, sample: *const Scalar) OutputArray {
102:         const value = self.update(sample.value);
103:         return LineIndicator.wrapScalar(sample.time, value);
104:     }
105: 
106:     /// Shadows LineIndicator.updateBar to extract OHLC directly from the bar.
107:     pub fn updateBar(self: *BalanceOfPower, sample: *const Bar) OutputArray {
108:         const value = self.updateOhlc(sample.open, sample.high, sample.low, sample.close);
109:         return LineIndicator.wrapScalar(sample.time, value);
110:     }
111: 
112:     pub fn updateQuote(self: *BalanceOfPower, sample: *const Quote) OutputArray {
113:         const value = self.update(self.line.extractQuote(sample));
114:         return LineIndicator.wrapScalar(sample.time, value);
115:     }
116: 
117:     pub fn updateTrade(self: *BalanceOfPower, sample: *const Trade) OutputArray {
118:         const value = self.update(self.line.extractTrade(sample));
119:         return LineIndicator.wrapScalar(sample.time, value);
120:     }
121: 
122:     pub fn indicator(self: *BalanceOfPower) indicator_mod.Indicator {
123:         return .{
124:             .ptr = @ptrCast(self),
125:             .vtable = &vtable,
126:         };
127:     }
128: 
129:     const vtable = indicator_mod.Indicator.VTable{
130:         .isPrimed = vtableIsPrimed,
131:         .metadata = vtableMetadata,
132:         .updateScalar = vtableUpdateScalar,
133:         .updateBar = vtableUpdateBar,
134:         .updateQuote = vtableUpdateQuote,
135:         .updateTrade = vtableUpdateTrade,
136:     };
137: 
138:     fn vtableIsPrimed(ptr: *anyopaque) bool {
139:         const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
140:         return self.isPrimed();
141:     }
142: 
143:     fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
144:         const self: *const BalanceOfPower = @ptrCast(@alignCast(ptr));
145:         self.getMetadata(out);
146:     }
147: 
148:     fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
149:         const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
150:         return self.updateScalar(sample);
151:     }
152: 
153:     fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
154:         const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
155:         return self.updateBar(sample);
156:     }
157: 
158:     fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
159:         const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
160:         return self.updateQuote(sample);
161:     }
162: 
163:     fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
164:         const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
165:         return self.updateTrade(sample);
166:     }
167: };
168: 
169: // ---------------------------------------------------------------------------
170: // Tests
171: // ---------------------------------------------------------------------------
172: 
173: const testing = std.testing;
174: 
175: fn testOpen() [20]f64 {
176:     return .{
177:         92.500, 91.500, 95.155, 93.970, 95.500, 94.500, 95.000, 91.500, 91.815, 91.125,
178:         93.875, 97.500, 98.815, 92.000, 91.125, 91.875, 93.405, 89.750, 89.345, 92.250,
179:     };

(Showing lines 100-179 of 303. Use offset=180 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 6.6s)

Good, I have the pattern. Now let me also look at how MaUnion is structured (needed for ADOSC) — I'll check APO which uses it:

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && grep -n \"MaUnion\" src/indicators/common/absolute_price_oscillator/absolute_price_oscillator.zig | head -20",
  "description": "Find MaUnion usage in APO"
}
```

**Output:**
```
54:    fast_ma: MaUnion,
55:    slow_ma: MaUnion,
64:    const MaUnion = union(enum) {
68:        fn update(self: *MaUnion, sample: f64) f64 {
75:        fn isPrimed(self: *const MaUnion) bool {
82:        fn deinit(self: *MaUnion) void {
105:        var fast_ma: MaUnion = undefined;
106:        var slow_ma: MaUnion = undefined;

```

---

## Assistant (Build · claude-opus-4.6 · 5.0s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/common/absolute_price_oscillator/absolute_price_oscillator.zig",
  "limit": 80,
  "offset": 60
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/common/absolute_price_oscillator/absolute_price_oscillator.zig</path>
<type>file</type>
<content>60:     mnemonic_len: usize,
61:     description_buf: [160]u8,
62:     description_len: usize,
63: 
64:     const MaUnion = union(enum) {
65:         sma: sma_mod.SimpleMovingAverage,
66:         ema: ema_mod.ExponentialMovingAverage,
67: 
68:         fn update(self: *MaUnion, sample: f64) f64 {
69:             return switch (self.*) {
70:                 .sma => |*s| s.update(sample),
71:                 .ema => |*e| e.update(sample),
72:             };
73:         }
74: 
75:         fn isPrimed(self: *const MaUnion) bool {
76:             return switch (self.*) {
77:                 .sma => |*s| s.isPrimed(),
78:                 .ema => |*e| e.isPrimed(),
79:             };
80:         }
81: 
82:         fn deinit(self: *MaUnion) void {
83:             switch (self.*) {
84:                 .sma => |*s| s.deinit(),
85:                 .ema => {},
86:             }
87:         }
88:     };
89: 
90:     pub fn init(allocator: std.mem.Allocator, params: AbsolutePriceOscillatorParams) !AbsolutePriceOscillator {
91:         if (params.fast_length < 2) {
92:             return error.InvalidFastLength;
93:         }
94:         if (params.slow_length < 2) {
95:             return error.InvalidSlowLength;
96:         }
97: 
98:         const bc = params.bar_component orelse bar_component.default_bar_component;
99:         const qc = params.quote_component orelse quote_component.default_quote_component;
100:         const tc = params.trade_component orelse trade_component.default_trade_component;
101: 
102:         var triple_buf: [64]u8 = undefined;
103:         const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);
104: 
105:         var fast_ma: MaUnion = undefined;
106:         var slow_ma: MaUnion = undefined;
107:         var ma_label: []const u8 = undefined;
108: 
109:         switch (params.moving_average_type) {
110:             .ema => {
111:                 ma_label = "EMA";
112:                 var fast_ema = try ema_mod.ExponentialMovingAverage.initLength(.{
113:                     .length = params.fast_length,
114:                     .first_is_average = params.first_is_average,
115:                 });
116:                 fast_ema.fixSlices();
117:                 var slow_ema = try ema_mod.ExponentialMovingAverage.initLength(.{
118:                     .length = params.slow_length,
119:                     .first_is_average = params.first_is_average,
120:                 });
121:                 slow_ema.fixSlices();
122:                 fast_ma = .{ .ema = fast_ema };
123:                 slow_ma = .{ .ema = slow_ema };
124:             },
125:             .sma => {
126:                 ma_label = "SMA";
127:                 var fast_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{
128:                     .length = params.fast_length,
129:                 });
130:                 fast_sma.fixSlices();
131:                 var slow_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{
132:                     .length = params.slow_length,
133:                 });
134:                 slow_sma.fixSlices();
135:                 fast_ma = .{ .sma = fast_sma };
136:                 slow_ma = .{ .sma = slow_sma };
137:             },
138:         }
139: 

(Showing lines 60-139 of 461. Use offset=140 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 6.2s)

Good, I have everything I need. Now I'll create both indicators. Let me first create the directory structure and the files.

**Tool: bash**

**Input:**
```json
{
  "command": "mkdir -p /home/dev/repos/chi/zpano/zig/src/indicators/marc_chaikin/advance_decline && mkdir -p /home/dev/repos/chi/zpano/zig/src/indicators/marc_chaikin/advance_decline_oscillator",
  "description": "Create marc_chaikin directories"
}
```

---

## Assistant (Build · claude-opus-4.6 · 5.5s)

Now let me read the rest of APO to see the full vtable/fixSlices pattern, and also check how the EMA init works:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/common/absolute_price_oscillator/absolute_price_oscillator.zig",
  "limit": 60,
  "offset": 139
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/common/absolute_price_oscillator/absolute_price_oscillator.zig</path>
<type>file</type>
<content>139: 
140:         var mnemonic_buf: [96]u8 = undefined;
141:         const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "apo({s}{d}/{s}{d}{s})", .{
142:             ma_label, params.fast_length, ma_label, params.slow_length, triple,
143:         }) catch return error.MnemonicTooLong;
144:         const mnemonic_len = mnemonic_slice.len;
145: 
146:         var description_buf: [160]u8 = undefined;
147:         const desc_slice = std.fmt.bufPrint(&description_buf, "Absolute Price Oscillator {s}", .{mnemonic_slice}) catch
148:             return error.MnemonicTooLong;
149:         const description_len = desc_slice.len;
150: 
151:         return .{
152:             .line = LineIndicator.new(
153:                 mnemonic_buf[0..mnemonic_len],
154:                 description_buf[0..description_len],
155:                 params.bar_component,
156:                 params.quote_component,
157:                 params.trade_component,
158:             ),
159:             .fast_ma = fast_ma,
160:             .slow_ma = slow_ma,
161:             .value = math.nan(f64),
162:             .primed = false,
163:             .allocator = allocator,
164:             .mnemonic_buf = mnemonic_buf,
165:             .mnemonic_len = mnemonic_len,
166:             .description_buf = description_buf,
167:             .description_len = description_len,
168:         };
169:     }
170: 
171:     pub fn deinit(self: *AbsolutePriceOscillator) void {
172:         self.fast_ma.deinit();
173:         self.slow_ma.deinit();
174:     }
175: 
176:     pub fn fixSlices(self: *AbsolutePriceOscillator) void {
177:         self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
178:         self.line.description = self.description_buf[0..self.description_len];
179:     }
180: 
181:     pub fn update(self: *AbsolutePriceOscillator, sample: f64) f64 {
182:         if (math.isNan(sample)) {
183:             return sample;
184:         }
185: 
186:         const slow = self.slow_ma.update(sample);
187:         const fast = self.fast_ma.update(sample);
188:         self.primed = self.slow_ma.isPrimed() and self.fast_ma.isPrimed();
189: 
190:         if (math.isNan(fast) or math.isNan(slow)) {
191:             self.value = math.nan(f64);
192:             return self.value;
193:         }
194: 
195:         self.value = fast - slow;
196:         return self.value;
197:     }
198: 

(Showing lines 139-198 of 461. Use offset=199 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 251.5s)

Good. Now I have the full pattern. Let me create the AdvanceDecline indicator first (simpler — no params, no MAs):

**Tool: write**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/marc_chaikin/advance_decline/advance_decline.zig",
  "content": "const std = @import(\"std\");\nconst math = std.math;\n\nconst Bar = @import(\"bar\").Bar;\nconst Quote = @import(\"quote\").Quote;\nconst Trade = @import(\"trade\").Trade;\nconst Scalar = @import(\"scalar\").Scalar;\nconst quote_component = @import(\"quote_component\");\nconst trade_component = @import(\"trade_component\");\n\nconst indicator_mod = @import(\"../../core/indicator.zig\");\nconst line_indicator_mod = @import(\"../../core/line_indicator.zig\");\nconst build_metadata_mod = @import(\"../../core/build_metadata.zig\");\nconst identifier_mod = @import(\"../../core/identifier.zig\");\nconst metadata_mod = @import(\"../../core/metadata.zig\");\n\nconst OutputArray = indicator_mod.OutputArray;\nconst LineIndicator = line_indicator_mod.LineIndicator;\nconst Identifier = identifier_mod.Identifier;\nconst Metadata = metadata_mod.Metadata;\n\n/// Enumerates the outputs of the Advance-Decline indicator.\npub const AdvanceDeclineOutput = enum(u8) {\n    /// The scalar value of the A/D line.\n    value = 1,\n};\n\n/// Marc Chaikin's Advance-Decline (A/D) Line.\n///\n/// The Accumulation/Distribution Line is a cumulative indicator that uses volume\n/// and price to assess whether a stock is being accumulated or distributed.\n///\n/// CLV = ((Close - Low) - (High - Close)) / (High - Low)\n/// AD  = AD_previous + CLV × Volume\n///\n/// When High equals Low, the A/D value is unchanged (no division by zero).\npub const AdvanceDecline = struct {\n    line: LineIndicator,\n    ad: f64,\n    value: f64,\n    primed: bool,\n\n    const mnemonic_str = \"ad\";\n    const description_str = \"Advance-Decline\";\n\n    pub fn init() AdvanceDecline {\n        return .{\n            .line = LineIndicator.new(\n                mnemonic_str,\n                description_str,\n                null,\n                null,\n                null,\n            ),\n            .ad = 0,\n            .value = math.nan(f64),\n            .primed = false,\n        };\n    }\n\n    pub fn deinit(_: *AdvanceDecline) void {}\n\n    pub fn fixSlices(_: *AdvanceDecline) void {}\n\n    /// Core update logic. For scalar/quote/trade, H=L=C so AD is unchanged.\n    pub fn update(self: *AdvanceDecline, sample: f64) f64 {\n        if (math.isNan(sample)) {\n            return math.nan(f64);\n        }\n        return self.updateHlcv(sample, sample, sample, 1);\n    }\n\n    /// Updates the indicator with the given high, low, close, and volume values.\n    pub fn updateHlcv(self: *AdvanceDecline, high: f64, low: f64, close: f64, volume: f64) f64 {\n        if (math.isNan(high) or math.isNan(low) or math.isNan(close) or math.isNan(volume)) {\n            return math.nan(f64);\n        }\n\n        const temp = high - low;\n        if (temp > 0) {\n            self.ad += ((close - low) - (high - close)) / temp * volume;\n        }\n\n        self.value = self.ad;\n        self.primed = true;\n\n        return self.value;\n    }\n\n    pub fn isPrimed(self: *const AdvanceDecline) bool {\n        return self.primed;\n    }\n\n    pub fn getMetadata(self: *const AdvanceDecline, out: *Metadata) void {\n        build_metadata_mod.buildMetadata(\n            out,\n            .advance_decline,\n            self.line.mnemonic,\n            self.line.description,\n            &[_]build_metadata_mod.OutputText{\n                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },\n            },\n        );\n    }\n\n    pub fn updateScalar(self: *AdvanceDecline, sample: *const Scalar) OutputArray {\n        const value = self.update(sample.value);\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    /// Shadows LineIndicator.updateBar to extract HLCV directly from the bar.\n    pub fn updateBar(self: *AdvanceDecline, sample: *const Bar) OutputArray {\n        const value = self.updateHlcv(sample.high, sample.low, sample.close, sample.volume);\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn updateQuote(self: *AdvanceDecline, sample: *const Quote) OutputArray {\n        const value = self.update(self.line.extractQuote(sample));\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn updateTrade(self: *AdvanceDecline, sample: *const Trade) OutputArray {\n        const value = self.update(self.line.extractTrade(sample));\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn indicator(self: *AdvanceDecline) indicator_mod.Indicator {\n        return .{\n            .ptr = @ptrCast(self),\n            .vtable = &vtable,\n        };\n    }\n\n    const vtable = indicator_mod.Indicator.VTable{\n        .isPrimed = vtableIsPrimed,\n        .metadata = vtableMetadata,\n        .updateScalar = vtableUpdateScalar,\n        .updateBar = vtableUpdateBar,\n        .updateQuote = vtableUpdateQuote,\n        .updateTrade = vtableUpdateTrade,\n    };\n\n    fn vtableIsPrimed(ptr: *anyopaque) bool {\n        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));\n        return self.isPrimed();\n    }\n\n    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {\n        const self: *const AdvanceDecline = @ptrCast(@alignCast(ptr));\n        self.getMetadata(out);\n    }\n\n    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {\n        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));\n        return self.updateScalar(sample);\n    }\n\n    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {\n        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));\n        return self.updateBar(sample);\n    }\n\n    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {\n        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));\n        return self.updateQuote(sample);\n    }\n\n    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {\n        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));\n        return self.updateTrade(sample);\n    }\n};\n\n// ---------------------------------------------------------------------------\n// Tests\n// ---------------------------------------------------------------------------\n\nconst testing = std.testing;\n\nfn roundTo(v: f64, comptime digits: comptime_int) f64 {\n    const p = comptime blk: {\n        var result: f64 = 1.0;\n        for (0..digits) |_| result *= 10.0;\n        break :blk result;\n    };\n    return @round(v * p) / p;\n}\n\n// High test data, 252 entries.\nfn testHighs() [252]f64 {\n    return .{\n        93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,\n        96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,\n        90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,\n        87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,\n        84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,\n        90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,\n        90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,\n        89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,\n        104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,\n        123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,\n        116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,\n        116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,\n        125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,\n        138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,\n        125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,\n        124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,\n        124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,\n        132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,\n        129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,\n        123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,\n        108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,\n        96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,\n        95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,\n        104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,\n        107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,\n        110.5, 109.5,\n    };\n}\n\n// Low test data, 252 entries.\nfn testLows() [252]f64 {\n    return .{\n        90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,\n        92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,\n        87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,\n        84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,\n        82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,\n        89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,\n        88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,\n        86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,\n        101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,\n        117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,\n        112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,\n        112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,\n        123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,\n        135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,\n        122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,\n        121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,\n        122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,\n        128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,\n        126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,\n        119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,\n        105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,\n        93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,\n        93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,\n        102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,\n        104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,\n        108.56, 106.62,\n    };\n}\n\n// Close test data, 252 entries.\nfn testCloses() [252]f64 {\n    return .{\n        91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,\n        96.125, 97.25, 98.5, 89.875, 91, 92.815, 89.155, 89.345, 91.625, 89.875,\n        88.375, 87.625, 84.78, 83, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,\n        85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,\n        83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89, 91, 90.5,\n        89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,\n        88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90, 89.69,\n        88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106, 102.5,\n        102.405, 104.595, 106.125, 106, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,\n        123, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,\n        116, 116, 112, 113.75, 112.94, 116, 120.5, 116.62, 117, 115.25,\n        114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,\n        123.12, 122.56, 124.62, 129.25, 131, 132.25, 131, 132.81, 134, 137.38,\n        137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129, 123.87, 124.81,\n        123, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,\n        122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,\n        124.44, 122, 122.37, 122.94, 124, 123.19, 124.56, 127.25, 125.87, 128.86,\n        132, 130.75, 134.75, 135, 132.38, 133.31, 131.94, 130, 125.37, 130.13,\n        127.12, 125.19, 122, 125, 123, 123.5, 120.06, 121, 117.75, 119.87,\n        122, 119.19, 116.37, 113.5, 114.25, 110, 105.06, 107, 107.87, 107,\n        107.12, 107, 91, 93.94, 93.87, 95.5, 93, 94.94, 98.25, 96.75,\n        94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97, 95, 95.87, 94.06,\n        94.62, 93.75, 98, 103.94, 107.87, 106.06, 104.5, 105, 104.19, 103.06,\n        103.42, 105.27, 111.87, 116, 116.62, 118.28, 113.37, 109, 109.7, 109.25,\n        107, 109.19, 110, 109.2, 110.12, 108, 108.62, 109.75, 109.81, 109,\n        108.75, 107.87,\n    };\n}\n\n// Volume test data, 252 entries.\nfn testVolumes() [252]f64 {\n    return .{\n        4077500, 4955900, 4775300, 4155300, 4593100, 3631300, 3382800, 4954200, 4500000, 3397500,\n        4204500, 6321400, 10203600, 19043900, 11692000, 9553300, 8920300, 5970900, 5062300, 3705600,\n        5865600, 5603000, 5811900, 8483800, 5995200, 5408800, 5430500, 6283800, 5834800, 4515500,\n        4493300, 4346100, 3700300, 4600200, 4557200, 4323600, 5237500, 7404100, 4798400, 4372800,\n        3872300, 10750800, 5804800, 3785500, 5014800, 3507700, 4298800, 4842500, 3952200, 3304700,\n        3462000, 7253900, 9753100, 5953000, 5011700, 5910800, 4916900, 4135000, 4054200, 3735300,\n        2921900, 2658400, 4624400, 4372200, 5831600, 4268600, 3059200, 4495500, 3425000, 3630800,\n        4168100, 5966900, 7692800, 7362500, 6581300, 19587700, 10378600, 9334700, 10467200, 5671400,\n        5645000, 4518600, 4519500, 5569700, 4239700, 4175300, 4995300, 4776600, 4190000, 6035300,\n        12168900, 9040800, 5780300, 4320800, 3899100, 3221400, 3455500, 4304200, 4703900, 8316300,\n        10553900, 6384800, 7163300, 7007800, 5114100, 5263800, 6666100, 7398400, 5575000, 4852300,\n        4298100, 4900500, 4887700, 6964800, 4679200, 9165000, 6469800, 6792000, 4423800, 5231900,\n        4565600, 6235200, 5225900, 8261400, 5912500, 3545600, 5714500, 6653900, 6094500, 4799200,\n        5050800, 5648900, 4726300, 5585600, 5124800, 7630200, 14311600, 8793600, 8874200, 6966600,\n        5525500, 6515500, 5291900, 5711700, 4327700, 4568000, 6859200, 5757500, 7367000, 6144100,\n        4052700, 5849700, 5544700, 5032200, 4400600, 4894100, 5140000, 6610900, 7585200, 5963100,\n        6045500, 8443300, 6464700, 6248300, 4357200, 4774700, 6216900, 6266900, 5584800, 5284500,\n        7554500, 7209500, 8424800, 5094500, 4443600, 4591100, 5658400, 6094100, 14862200, 7544700,\n        6985600, 8093000, 7590000, 7451300, 7078000, 7105300, 8778800, 6643900, 10563900, 7043100,\n        6438900, 8057700, 14240000, 17872300, 7831100, 8277700, 15017800, 14183300, 13921100, 9683000,\n        9187300, 11380500, 69447300, 26673600, 13768400, 11371600, 9872200, 9450500, 11083300, 9552800,\n        11108400, 10374200, 16701900, 13741900, 8523600, 9551900, 8680500, 7151700, 9673100, 6264700,\n        8541600, 8358000, 18720800, 19683100, 13682500, 10668100, 9710600, 3113100, 5682000, 5763600,\n        5340000, 6220800, 14680500, 9933000, 11329500, 8145300, 16644700, 12593800, 7138100, 7442300,\n        9442300, 7123600, 7680600, 4839800, 4775500, 4008800, 4533600, 3741100, 4084800, 2685200,\n        3438000, 2870500,\n    };\n}\n\n// Expected AD output, 252 entries.\nfn testExpectedAD() [252]f64 {\n    return .{\n        -1631000.0000000000,  2974412.0226308300, -1239087.9773691700,  -466727.3825736260, -2276567.4139707900,\n          789202.2581603610, -1398341.7418396400,  1134914.1775558300,  -766075.9214540700,  1736082.9020753400,\n         5640261.4735039100,  2890166.0052259700,  8972764.3897170800,   188856.2575697040, 10267283.4207680000,\n        16222587.3168718000,  8293431.7613162900,  5920994.1613162700,  9699107.7642574600,  6364067.7642574600,\n         4106481.8240114100,  4732197.5600520200,  -582684.7847755650, -5272487.2692476200, -3197626.7475084800,\n        -7210607.3926697700, -1780107.3926697700,  4138902.7196897800,  1901845.0121799000, -1812171.6280736600,\n        -3544134.5371645800,   297777.5278006210, -2159221.6721993900,  2440978.3278006100,   435810.3278006080,\n        -3070523.8185408500,  1465626.2894505000,  3007000.2671456800,  5345808.5051319400,  2847065.6479890900,\n         2348057.9160303300, -5679832.0403015400, -1127705.6433843900,  -226396.1195748840,   519476.0323238490,\n          346091.6216774670, -1280481.3512955100, -6122981.3512955100, -2643824.5945387600, -4331118.5522426800,\n        -7585398.5522426800,  -693439.5085836460, -9354192.3085836400,-14168892.5669816000,-15370043.8066510000,\n       -12414643.8066510000,-12585096.3399843000,-13890885.8136685000,-11450943.9372383000,-12314251.2598928000,\n       -14359581.2598928000,-12924045.2598928000, -8299645.2598927900,-11118563.6809454000, -6824818.5234191600,\n        -4130032.7842887200, -3708074.1635990700,   122721.5562452970, -3201046.9166611000, -4798598.9166611000,\n        -3061890.5833277700, -8623980.5426222100,-11300142.4820162000, -3937642.4820161500,  2643657.5179838500,\n        -5243139.1816861100, -5243139.1816861100,  2167108.4486456300,  2554782.5227197100,  -686017.4772802920,\n        -3549128.9880716500, -1194197.2420399100,  3325302.7579600900,  -709635.4642621180,  2925560.1041386400,\n          489968.4374719780,  5097434.5865403000,  3432561.5574946500,  5938908.5894581200,  9231738.4194426700,\n        21400638.4194427000, 20195198.4194427000, 22748227.3840560000, 21998336.4749651000, 21964073.7333131000,\n        19373669.6096018000, 17880959.5857354000, 15502576.1405453000, 11156172.5405452000, 19056383.7979514000,\n        25289392.9744220000, 27933099.2244220000, 25099853.7020340000, 29753682.2734625000, 28403559.8734625000,\n        33667359.8734625000, 36546987.7970902000, 31323445.5398723000, 26886728.6025174000, 27507371.6257732000,\n        27027380.7628290000, 29477630.7628290000, 27665337.5044020000, 30340032.5614362000, 29372997.8947696000,\n        31054151.7409234000, 33389894.2329362000, 27656016.3381994000, 27547234.3709863000, 25948598.2598752000,\n        21382998.2598752000, 16993417.4598752000, 19402230.7411252000, 21708493.1264463000, 26513515.7850566000,\n        30059115.7850566000, 25051948.0326135000, 28506411.7956798000, 29175540.8757985000, 33812970.0892816000,\n        36178148.3370158000, 37524688.4532948000, 37135945.1632515000, 37120557.8354279000, 36201961.6090128000,\n        31335687.6205735000, 27351074.1115125000, 31564674.1115125000, 24003855.7115125000, 27857719.5412998000,\n        23497936.8408778000, 28905358.6095853000, 33058030.1373631000, 31630105.1373631000, 32227029.2752941000,\n        28618418.2967738000, 25042729.8586399000, 20615046.1316058000, 26128856.6906120000, 25104840.0239454000,\n        23664422.5540658000, 19030671.4304703000, 23336961.2781861000, 18734557.7194672000, 19157015.3194672000,\n        22598860.1265889000, 27430460.1265889000, 21617613.3076656000, 21114727.1198203000, 17637753.4038082000,\n        21652343.2475581000, 19654310.4231306000, 19269265.5955444000, 15907319.4013187000, 16014463.6636138000,\n        12869354.1072998000, 13365911.6152870000, 14410394.9486203000, 16063495.7486204000, 15972904.3200490000,\n        21108562.0926708000, 14165259.0157478000, 16535274.6214973000, 13349288.2166634000,  9854321.9245286000,\n        13028415.7516891000,  7370015.7516890900,  3177620.7162280900, -5422742.0926145400,  -197109.7155653600,\n        -3586572.4374588300, -7071058.5485699700,-13326553.0540645000, -5875253.0540644700, -8711657.4658291900,\n        -7411341.1259599000,-15815245.7522944000,-15486840.0488724000,-17500820.9841242000,-11021168.9841242000,\n        -8424094.7597804800,-15914043.5402683000,-21120435.7920843000,-17188529.7920843000,-20062126.7108518000,\n       -28339826.7108518000,-43050618.8743952000,-31637494.6556452000,-23284834.6556452000,-22588377.5790791000,\n       -21687661.8928045000,-33068161.8928045000,-33068161.8928045000,-23505927.9305405000,-21211194.5972072000,\n       -18079884.4522797000,-26587925.9068252000,-23734944.7747498000,-18697081.1383861000,-25274418.8433042000,\n       -28407557.3048426000,-31383762.2228754000,-36349191.9526051000,-49491445.4071505000,-40967845.4071505000,\n       -42041727.6556712000,-33715533.7781202000,-40160221.6108198000,-32749138.8415890000,-38061604.4415890000,\n       -32765812.4415890000,-32728665.7749224000,-16003686.6704448000, -3202381.6334077700,  6580605.8665922800,\n        -3756700.3349581000,  2909824.4299321800,  2711115.9192938800,  2964453.4989116900,  1753613.1627772300,\n         2119366.5874347800,  3939056.3833531200, 12663978.2900468000, 20113728.2900468000, 15914260.2900469000,\n        16003118.1082287000,  3135938.6756046000, -1177006.5298748600, -1073221.5541074900,  2154763.9880611900,\n         6436737.2438751500, 11752316.9393066000, 12759608.7425853000, 13137436.9953734000, 17912936.9953734000,\n        13904136.9953734000, 11103379.6620401000, 13849506.2577848000, 13390268.3879623000, 11472268.3879623000,\n         8707691.0683746700,  8328944.5405968900,\n    };\n}\n\ntest \"AdvanceDecline with volume\" {\n    const highs = testHighs();\n    const lows = testLows();\n    const closes = testCloses();\n    const volumes = testVolumes();\n    const expected = testExpectedAD();\n\n    var ad = AdvanceDecline.init();\n\n    for (0..252) |i| {\n        const v = ad.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);\n        try testing.expect(!math.isNan(v));\n        try testing.expect(ad.isPrimed());\n\n        const got = roundTo(v, 2);\n        const exp = roundTo(expected[i], 2);\n        try testing.expectEqual(exp, got);\n    }\n}\n\ntest \"AdvanceDecline TA-Lib spot checks\" {\n    const highs = testHighs();\n    const lows = testLows();\n    const closes = testCloses();\n    const volumes = testVolumes();\n\n    var ad = AdvanceDecline.init();\n    var values: [252]f64 = undefined;\n\n    for (0..252) |i| {\n        values[i] = ad.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);\n    }\n\n    try testing.expectEqual(roundTo(-1631000.00, 2), roundTo(values[0], 2));\n    try testing.expectEqual(roundTo(2974412.02, 2), roundTo(values[1], 2));\n    try testing.expectEqual(roundTo(8707691.07, 2), roundTo(values[250], 2));\n    try testing.expectEqual(roundTo(8328944.54, 2), roundTo(values[251], 2));\n}\n\ntest \"AdvanceDecline updateBar\" {\n    var ad = AdvanceDecline.init();\n\n    const highs = testHighs();\n    const lows = testLows();\n    const closes = testCloses();\n    const volumes = testVolumes();\n    const expected = testExpectedAD();\n\n    for (0..10) |i| {\n        const bar = Bar{\n            .time = 0,\n            .open = highs[i],\n            .high = highs[i],\n            .low = lows[i],\n            .close = closes[i],\n            .volume = volumes[i],\n        };\n\n        const output = ad.updateBar(&bar);\n        const scalar_val = output.slice()[0].scalar.value;\n\n        const got = roundTo(scalar_val, 2);\n        const exp = roundTo(expected[i], 2);\n        try testing.expectEqual(exp, got);\n    }\n}\n\ntest \"AdvanceDecline scalar update\" {\n    var ad = AdvanceDecline.init();\n\n    // Scalar update: H=L=C, so range=0, AD unchanged (remains 0 after primed).\n    const v = ad.update(100.0);\n    try testing.expectEqual(@as(f64, 0), v);\n    try testing.expect(ad.isPrimed());\n}\n\ntest \"AdvanceDecline NaN\" {\n    var ad = AdvanceDecline.init();\n\n    try testing.expect(math.isNan(ad.update(math.nan(f64))));\n    try testing.expect(math.isNan(ad.updateHlcv(math.nan(f64), 1, 2, 3)));\n    try testing.expect(math.isNan(ad.updateHlcv(1, math.nan(f64), 2, 3)));\n    try testing.expect(math.isNan(ad.updateHlcv(1, 2, math.nan(f64), 3)));\n    try testing.expect(math.isNan(ad.updateHlcv(1, 2, 3, math.nan(f64))));\n}\n\ntest \"AdvanceDecline not primed initially\" {\n    var ad = AdvanceDecline.init();\n\n    try testing.expect(!ad.isPrimed());\n    try testing.expect(math.isNan(ad.update(math.nan(f64))));\n    try testing.expect(!ad.isPrimed());\n}\n\ntest \"AdvanceDecline metadata\" {\n    var ad = AdvanceDecline.init();\n    var meta: Metadata = undefined;\n    ad.getMetadata(&meta);\n\n    try testing.expectEqual(Identifier.advance_decline, meta.identifier);\n    try testing.expectEqualStrings(\"ad\", meta.mnemonic);\n    try testing.expectEqualStrings(\"Advance-Decline\", meta.description);\n    try testing.expectEqual(@as(usize, 1), meta.outputs_len);\n}\n"
}
```

**Output:**
```
Wrote file successfully.

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/icalc/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.identifier" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.factory.factory" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/iconf/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.descriptors" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.core.outputs.heatmap" could not be resolved
ERROR [26:6] Import "py.indicators.core.outputs.shape" could not be resolved
ERROR [27:6] Import "py.indicators.core.pane" could not be resolved
ERROR [28:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [31:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/ifres/main.py">
ERROR [16:6] Import "py.indicators.core.frequency_response" could not be resolved
ERROR [17:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [18:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [21:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>
```

---

## Assistant (Build · claude-opus-4.6 · 287.9s)

Now let me create the AdvanceDeclineOscillator. This one needs MaUnion with SMA/EMA, similar to APO:

**Tool: write**

**Input:**
```json
{
  "content": "const std = @import(\"std\");\nconst math = std.math;\n\nconst Bar = @import(\"bar\").Bar;\nconst Quote = @import(\"quote\").Quote;\nconst Trade = @import(\"trade\").Trade;\nconst Scalar = @import(\"scalar\").Scalar;\nconst quote_component = @import(\"quote_component\");\nconst trade_component = @import(\"trade_component\");\n\nconst indicator_mod = @import(\"../../core/indicator.zig\");\nconst line_indicator_mod = @import(\"../../core/line_indicator.zig\");\nconst build_metadata_mod = @import(\"../../core/build_metadata.zig\");\nconst identifier_mod = @import(\"../../core/identifier.zig\");\nconst metadata_mod = @import(\"../../core/metadata.zig\");\n\nconst sma_mod = @import(\"../../common/simple_moving_average/simple_moving_average.zig\");\nconst ema_mod = @import(\"../../common/exponential_moving_average/exponential_moving_average.zig\");\n\nconst OutputArray = indicator_mod.OutputArray;\nconst LineIndicator = line_indicator_mod.LineIndicator;\nconst Identifier = identifier_mod.Identifier;\nconst Metadata = metadata_mod.Metadata;\n\n/// Enumerates the outputs of the Advance-Decline Oscillator.\npub const AdvanceDeclineOscillatorOutput = enum(u8) {\n    /// The scalar value of the oscillator.\n    value = 1,\n};\n\n/// Moving average type for the ADOSC.\npub const MovingAverageType = enum {\n    sma,\n    ema,\n};\n\n/// Parameters for creating an AdvanceDeclineOscillator.\npub const AdvanceDeclineOscillatorParams = struct {\n    fast_length: u32 = 3,\n    slow_length: u32 = 10,\n    moving_average_type: MovingAverageType = .ema,\n    first_is_average: bool = false,\n};\n\n/// Marc Chaikin's Advance-Decline (A/D) Oscillator (Chaikin Oscillator).\n///\n/// The Chaikin Oscillator is the difference between a fast and slow moving average\n/// of the Accumulation/Distribution Line.\n///\n/// CLV = ((Close - Low) - (High - Close)) / (High - Low)\n/// AD  = AD_previous + CLV × Volume\n/// ADOSC = FastMA(AD) - SlowMA(AD)\n///\n/// When High equals Low, the A/D value is unchanged (no division by zero).\npub const AdvanceDeclineOscillator = struct {\n    line: LineIndicator,\n    ad: f64,\n    fast_ma: MaUnion,\n    slow_ma: MaUnion,\n    value: f64,\n    primed: bool,\n    allocator: std.mem.Allocator,\n    mnemonic_buf: [96]u8,\n    mnemonic_len: usize,\n    description_buf: [160]u8,\n    description_len: usize,\n\n    const MaUnion = union(enum) {\n        sma: sma_mod.SimpleMovingAverage,\n        ema: ema_mod.ExponentialMovingAverage,\n\n        fn update(self: *MaUnion, sample: f64) f64 {\n            return switch (self.*) {\n                .sma => |*s| s.update(sample),\n                .ema => |*e| e.update(sample),\n            };\n        }\n\n        fn isPrimed(self: *const MaUnion) bool {\n            return switch (self.*) {\n                .sma => |*s| s.isPrimed(),\n                .ema => |*e| e.isPrimed(),\n            };\n        }\n\n        fn deinit(self: *MaUnion) void {\n            switch (self.*) {\n                .sma => |*s| s.deinit(),\n                .ema => {},\n            }\n        }\n    };\n\n    pub fn init(allocator: std.mem.Allocator, params: AdvanceDeclineOscillatorParams) !AdvanceDeclineOscillator {\n        if (params.fast_length < 2) {\n            return error.InvalidFastLength;\n        }\n        if (params.slow_length < 2) {\n            return error.InvalidSlowLength;\n        }\n\n        var fast_ma: MaUnion = undefined;\n        var slow_ma: MaUnion = undefined;\n        var ma_label: []const u8 = undefined;\n\n        switch (params.moving_average_type) {\n            .ema => {\n                ma_label = \"EMA\";\n                var fast_ema = try ema_mod.ExponentialMovingAverage.initLength(.{\n                    .length = params.fast_length,\n                    .first_is_average = params.first_is_average,\n                });\n                fast_ema.fixSlices();\n                var slow_ema = try ema_mod.ExponentialMovingAverage.initLength(.{\n                    .length = params.slow_length,\n                    .first_is_average = params.first_is_average,\n                });\n                slow_ema.fixSlices();\n                fast_ma = .{ .ema = fast_ema };\n                slow_ma = .{ .ema = slow_ema };\n            },\n            .sma => {\n                ma_label = \"SMA\";\n                var fast_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{\n                    .length = params.fast_length,\n                });\n                fast_sma.fixSlices();\n                var slow_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{\n                    .length = params.slow_length,\n                });\n                slow_sma.fixSlices();\n                fast_ma = .{ .sma = fast_sma };\n                slow_ma = .{ .sma = slow_sma };\n            },\n        }\n\n        var mnemonic_buf: [96]u8 = undefined;\n        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, \"adosc({s}{d}/{s}{d})\", .{\n            ma_label, params.fast_length, ma_label, params.slow_length,\n        }) catch return error.MnemonicTooLong;\n        const mnemonic_len = mnemonic_slice.len;\n\n        var description_buf: [160]u8 = undefined;\n        const desc_slice = std.fmt.bufPrint(&description_buf, \"Chaikin Advance-Decline Oscillator {s}\", .{mnemonic_slice}) catch\n            return error.MnemonicTooLong;\n        const description_len = desc_slice.len;\n\n        return .{\n            .line = LineIndicator.new(\n                mnemonic_buf[0..mnemonic_len],\n                description_buf[0..description_len],\n                null,\n                null,\n                null,\n            ),\n            .ad = 0,\n            .fast_ma = fast_ma,\n            .slow_ma = slow_ma,\n            .value = math.nan(f64),\n            .primed = false,\n            .allocator = allocator,\n            .mnemonic_buf = mnemonic_buf,\n            .mnemonic_len = mnemonic_len,\n            .description_buf = description_buf,\n            .description_len = description_len,\n        };\n    }\n\n    pub fn deinit(self: *AdvanceDeclineOscillator) void {\n        self.fast_ma.deinit();\n        self.slow_ma.deinit();\n    }\n\n    pub fn fixSlices(self: *AdvanceDeclineOscillator) void {\n        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];\n        self.line.description = self.description_buf[0..self.description_len];\n    }\n\n    /// Core update logic. For scalar/quote/trade, H=L=C so AD is unchanged,\n    /// but the unchanged AD value is still fed to the MAs.\n    pub fn update(self: *AdvanceDeclineOscillator, sample: f64) f64 {\n        if (math.isNan(sample)) {\n            return math.nan(f64);\n        }\n        return self.updateHlcv(sample, sample, sample, 1);\n    }\n\n    /// Updates the indicator with the given high, low, close, and volume values.\n    pub fn updateHlcv(self: *AdvanceDeclineOscillator, high: f64, low: f64, close: f64, volume: f64) f64 {\n        if (math.isNan(high) or math.isNan(low) or math.isNan(close) or math.isNan(volume)) {\n            return math.nan(f64);\n        }\n\n        // Compute cumulative AD.\n        const temp = high - low;\n        if (temp > 0) {\n            self.ad += ((close - low) - (high - close)) / temp * volume;\n        }\n\n        // Feed AD to both MAs.\n        const fast = self.fast_ma.update(self.ad);\n        const slow = self.slow_ma.update(self.ad);\n        self.primed = self.fast_ma.isPrimed() and self.slow_ma.isPrimed();\n\n        if (math.isNan(fast) or math.isNan(slow)) {\n            self.value = math.nan(f64);\n            return self.value;\n        }\n\n        self.value = fast - slow;\n        return self.value;\n    }\n\n    pub fn isPrimed(self: *const AdvanceDeclineOscillator) bool {\n        return self.primed;\n    }\n\n    pub fn getMetadata(self: *const AdvanceDeclineOscillator, out: *Metadata) void {\n        build_metadata_mod.buildMetadata(\n            out,\n            .advance_decline_oscillator,\n            self.line.mnemonic,\n            self.line.description,\n            &[_]build_metadata_mod.OutputText{\n                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },\n            },\n        );\n    }\n\n    pub fn updateScalar(self: *AdvanceDeclineOscillator, sample: *const Scalar) OutputArray {\n        const value = self.update(sample.value);\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    /// Shadows LineIndicator.updateBar to extract HLCV directly from the bar.\n    pub fn updateBar(self: *AdvanceDeclineOscillator, sample: *const Bar) OutputArray {\n        const value = self.updateHlcv(sample.high, sample.low, sample.close, sample.volume);\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn updateQuote(self: *AdvanceDeclineOscillator, sample: *const Quote) OutputArray {\n        const value = self.update(self.line.extractQuote(sample));\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn updateTrade(self: *AdvanceDeclineOscillator, sample: *const Trade) OutputArray {\n        const value = self.update(self.line.extractTrade(sample));\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn indicator(self: *AdvanceDeclineOscillator) indicator_mod.Indicator {\n        return .{\n            .ptr = @ptrCast(self),\n            .vtable = &vtable,\n        };\n    }\n\n    const vtable = indicator_mod.Indicator.VTable{\n        .isPrimed = vtableIsPrimed,\n        .metadata = vtableMetadata,\n        .updateScalar = vtableUpdateScalar,\n        .updateBar = vtableUpdateBar,\n        .updateQuote = vtableUpdateQuote,\n        .updateTrade = vtableUpdateTrade,\n    };\n\n    fn vtableIsPrimed(ptr: *anyopaque) bool {\n        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));\n        return self.isPrimed();\n    }\n\n    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {\n        const self: *const AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));\n        self.getMetadata(out);\n    }\n\n    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {\n        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateScalar(sample);\n    }\n\n    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {\n        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateBar(sample);\n    }\n\n    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {\n        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateQuote(sample);\n    }\n\n    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {\n        const self: *AdvanceDeclineOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateTrade(sample);\n    }\n};\n\n// ---------------------------------------------------------------------------\n// Tests\n// ---------------------------------------------------------------------------\n\nconst testing = std.testing;\n\nfn roundTo(v: f64, comptime digits: comptime_int) f64 {\n    const p = comptime blk: {\n        var result: f64 = 1.0;\n        for (0..digits) |_| result *= 10.0;\n        break :blk result;\n    };\n    return @round(v * p) / p;\n}\n\n// High test data, 252 entries.\nfn testHighs() [252]f64 {\n    return .{\n        93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,\n        96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,\n        90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,\n        87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,\n        84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,\n        90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,\n        90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,\n        89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,\n        104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,\n        123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,\n        116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,\n        116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,\n        125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,\n        138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,\n        125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,\n        124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,\n        124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,\n        132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,\n        129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,\n        123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,\n        108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,\n        96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,\n        95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,\n        104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,\n        107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,\n        110.5, 109.5,\n    };\n}\n\n// Low test data, 252 entries.\nfn testLows() [252]f64 {\n    return .{\n        90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,\n        92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,\n        87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,\n        84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,\n        82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,\n        89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,\n        88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,\n        86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,\n        101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,\n        117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,\n        112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,\n        112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,\n        123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,\n        135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,\n        122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,\n        121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,\n        122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,\n        128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,\n        126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,\n        119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,\n        105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,\n        93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,\n        93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,\n        102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,\n        104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,\n        108.56, 106.62,\n    };\n}\n\n// Close test data, 252 entries.\nfn testCloses() [252]f64 {\n    return .{\n        91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,\n        96.125, 97.25, 98.5, 89.875, 91, 92.815, 89.155, 89.345, 91.625, 89.875,\n        88.375, 87.625, 84.78, 83, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,\n        85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,\n        83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89, 91, 90.5,\n        89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,\n        88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90, 89.69,\n        88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106, 102.5,\n        102.405, 104.595, 106.125, 106, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,\n        123, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,\n        116, 116, 112, 113.75, 112.94, 116, 120.5, 116.62, 117, 115.25,\n        114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,\n        123.12, 122.56, 124.62, 129.25, 131, 132.25, 131, 132.81, 134, 137.38,\n        137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129, 123.87, 124.81,\n        123, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,\n        122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,\n        124.44, 122, 122.37, 122.94, 124, 123.19, 124.56, 127.25, 125.87, 128.86,\n        132, 130.75, 134.75, 135, 132.38, 133.31, 131.94, 130, 125.37, 130.13,\n        127.12, 125.19, 122, 125, 123, 123.5, 120.06, 121, 117.75, 119.87,\n        122, 119.19, 116.37, 113.5, 114.25, 110, 105.06, 107, 107.87, 107,\n        107.12, 107, 91, 93.94, 93.87, 95.5, 93, 94.94, 98.25, 96.75,\n        94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97, 95, 95.87, 94.06,\n        94.62, 93.75, 98, 103.94, 107.87, 106.06, 104.5, 105, 104.19, 103.06,\n        103.42, 105.27, 111.87, 116, 116.62, 118.28, 113.37, 109, 109.7, 109.25,\n        107, 109.19, 110, 109.2, 110.12, 108, 108.62, 109.75, 109.81, 109,\n        108.75, 107.87,\n    };\n}\n\n// Volume test data, 252 entries.\nfn testVolumes() [252]f64 {\n    return .{\n        4077500, 4955900, 4775300, 4155300, 4593100, 3631300, 3382800, 4954200, 4500000, 3397500,\n        4204500, 6321400, 10203600, 19043900, 11692000, 9553300, 8920300, 5970900, 5062300, 3705600,\n        5865600, 5603000, 5811900, 8483800, 5995200, 5408800, 5430500, 6283800, 5834800, 4515500,\n        4493300, 4346100, 3700300, 4600200, 4557200, 4323600, 5237500, 7404100, 4798400, 4372800,\n        3872300, 10750800, 5804800, 3785500, 5014800, 3507700, 4298800, 4842500, 3952200, 3304700,\n        3462000, 7253900, 9753100, 5953000, 5011700, 5910800, 4916900, 4135000, 4054200, 3735300,\n        2921900, 2658400, 4624400, 4372200, 5831600, 4268600, 3059200, 4495500, 3425000, 3630800,\n        4168100, 5966900, 7692800, 7362500, 6581300, 19587700, 10378600, 9334700, 10467200, 5671400,\n        5645000, 4518600, 4519500, 5569700, 4239700, 4175300, 4995300, 4776600, 4190000, 6035300,\n        12168900, 9040800, 5780300, 4320800, 3899100, 3221400, 3455500, 4304200, 4703900, 8316300,\n        10553900, 6384800, 7163300, 7007800, 5114100, 5263800, 6666100, 7398400, 5575000, 4852300,\n        4298100, 4900500, 4887700, 6964800, 4679200, 9165000, 6469800, 6792000, 4423800, 5231900,\n        4565600, 6235200, 5225900, 8261400, 5912500, 3545600, 5714500, 6653900, 6094500, 4799200,\n        5050800, 5648900, 4726300, 5585600, 5124800, 7630200, 14311600, 8793600, 8874200, 6966600,\n        5525500, 6515500, 5291900, 5711700, 4327700, 4568000, 6859200, 5757500, 7367000, 6144100,\n        4052700, 5849700, 5544700, 5032200, 4400600, 4894100, 5140000, 6610900, 7585200, 5963100,\n        6045500, 8443300, 6464700, 6248300, 4357200, 4774700, 6216900, 6266900, 5584800, 5284500,\n        7554500, 7209500, 8424800, 5094500, 4443600, 4591100, 5658400, 6094100, 14862200, 7544700,\n        6985600, 8093000, 7590000, 7451300, 7078000, 7105300, 8778800, 6643900, 10563900, 7043100,\n        6438900, 8057700, 14240000, 17872300, 7831100, 8277700, 15017800, 14183300, 13921100, 9683000,\n        9187300, 11380500, 69447300, 26673600, 13768400, 11371600, 9872200, 9450500, 11083300, 9552800,\n        11108400, 10374200, 16701900, 13741900, 8523600, 9551900, 8680500, 7151700, 9673100, 6264700,\n        8541600, 8358000, 18720800, 19683100, 13682500, 10668100, 9710600, 3113100, 5682000, 5763600,\n        5340000, 6220800, 14680500, 9933000, 11329500, 8145300, 16644700, 12593800, 7138100, 7442300,\n        9442300, 7123600, 7680600, 4839800, 4775500, 4008800, 4533600, 3741100, 4084800, 2685200,\n        3438000, 2870500,\n    };\n}\n\n// Expected EMA ADOSC output, 252 entries.\nfn testExpectedEMA() [252]f64 {\n    return .{\n        0.00, 1465358.37, 590949.67, 425265.42, -257033.18,\n        462683.48, 19014.16, 641820.49, 233397.10, 841238.33,\n        2255663.07, 1754200.72, 3324956.45, 870389.71, 2993894.72,\n        5485298.48, 3482932.49, 1592286.02, 1876213.52, 760651.73,\n        -483187.42, -749013.21, -2480766.94, -4455897.26, -4198640.66,\n        -4988562.19, -3130319.89, -202233.93, 302213.24, -700628.51,\n        -1598267.56, -597759.92, -915892.00, 500925.97, 396986.59,\n        -797275.52, 229962.78, 1119728.10, 2126095.80, 1549455.69,\n        1013922.67, -1851662.43, -1407210.44, -810680.65, -255624.54,\n        -60485.30, -492702.89, -2165523.46, -1545988.96, -1688864.83,\n        -2629233.62, -582012.64, -2447296.98, -4519836.39, -5338983.69,\n        -4248372.70, -3470232.86, -3251905.68, -2090617.08, -1700177.07,\n        -2036676.87, -1532421.99, 284575.19, 105093.16, 1388306.41,\n        2644479.44, 3052220.74, 4160438.94, 3178016.51, 1978893.83,\n        1861033.22, -126123.22, -1779091.17, 49043.86, 2886508.75,\n        1275444.64, 500423.45, 2495682.35, 3208395.25, 2177123.69,\n        646329.51, 710634.76, 2110361.02, 1207281.33, 1884739.92,\n        1215581.10, 2297338.89, 2001294.85, 2495724.09, 3518825.17,\n        7489396.05, 8049317.22, 8358947.17, 7487108.03, 6438898.59,\n        4600511.73, 2955262.85, 1256787.86, -935242.86, 766742.96,\n        3376536.18, 4978400.28, 4279639.27, 5085488.01, 4523251.92,\n        5556886.97, 6390811.39, 4488943.25, 1891137.88, 853955.02,\n        199296.30, 692988.52, 255315.80, 904096.68, 779623.58,\n        1192740.49, 1996502.70, 319398.57, -430338.16, -1206583.93,\n        -2867140.48, -4682494.74, -4233017.73, -2930510.22, -602385.37,\n        1532937.10, 673930.45, 1360399.80, 1730460.17, 3200079.60,\n        4262927.93, 4738635.21, 4378766.66, 3828582.05, 2963172.99,\n        791402.75, -1436825.99, -877062.66, -2974050.55, -2335311.59,\n        -3248911.83, -1606758.47, 532405.07, 904777.20, 1164789.18,\n        17073.39, -1591717.88, -3513966.59, -2226495.29, -1823217.31,\n        -1950807.57, -3300032.80, -2181802.47, -2990400.05, -2914917.10,\n        -1523912.81, 720999.21, -275715.25, -818406.29, -2072325.51,\n        -1119529.74, -1263711.94, -1330326.92, -2306349.85, -2461872.43,\n        -3302405.40, -3188045.55, -2519104.05, -1490450.08, -962966.18,\n        974438.00, -530805.64, -344236.19, -1250341.08, -2619387.70,\n        -1931387.13, -3274751.71, -4860549.12, -7803895.39, -6635845.35,\n        -6633213.44, -7137817.47, -8685738.51, -6158498.16, -5467255.96,\n        -4273714.77, -6070898.05, -6149716.84, -6263707.84, -3679205.08,\n        -1461093.51, -2804023.07, -4755071.51, -3869887.46, -4070285.06,\n        -6416053.61, -11473115.20, -8867454.03, -4337705.01, -1868689.06,\n        -402164.91, -3386730.94, -4299805.14, -1239915.14, 854720.03,\n        2630240.37, 410372.74, 372705.18, 1926370.05, 294046.41,\n        -1397361.54, -2909242.58, -4843172.23, -9375662.99, -7665476.20,\n        -6610682.67, -2928961.57, -3207116.19, -671278.92, -1263195.54,\n        294517.43, 916808.73, 6409620.45, 12147129.54, 16502776.59,\n        13495236.91, 13159207.01, 11762223.26, 10202044.29, 8251063.31,\n        6819214.15, 6192521.32, 8149302.02, 10579332.57, 9275484.11,\n        7927138.67, 2560791.29, -1239632.59, -2648635.09, -1957173.88,\n        -133932.45, 2315435.53, 3427457.50, 3691006.62, 4982753.38,\n        3782690.38, 2056724.52, 2037440.17, 1698208.14, 794776.12,\n        -526700.32, -1139932.73,\n    };\n}\n\n// Expected SMA ADOSC output, 252 entries. First 9 are NaN, then 243 valid.\nfn testExpectedSMA() [252]f64 {\n    return .{\n        math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64),\n        math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), 815959.29,\n        1590615.58, 2817787.49, 4208829.41, 2326135.98, 3530790.03,\n        4404059.17, 6136406.99, 4209035.90, 988024.35, -117895.47,\n        -569354.95, -2409194.84, -3769234.10, -6349422.78, -7646206.54,\n        -7512194.60, -5340714.06, -2716995.09, 1100215.32, 1907151.18,\n        111200.39, 19953.24, 61923.58, 1285614.64, 968281.87,\n        250506.13, -399184.59, 571068.97, 3032116.73, 3026672.79,\n        2217806.09, -859646.47, -2287721.83, -2879135.73, -821066.28,\n        -671462.07, -748213.05, -2049367.35, -2247042.78, -2546103.42,\n        -2040230.18, -1888741.07, -2740450.32, -3540698.68, -6843948.13,\n        -6588025.09, -4929631.51, -3659788.40, -2457843.18, -1569248.21,\n        -1048061.76, -316034.96, 916712.34, 1692018.64, 2870571.69,\n        3431981.42, 5014442.37, 5928928.33, 5413600.58, 4298527.10,\n        2107220.65, -130430.37, -1997561.83, -3007571.25, -198539.50,\n        1931772.24, 1650113.17, 1286824.79, 3310549.08, 4418331.51,\n        2561642.87, 569004.62, 443566.84, 1067263.98, 2412326.17,\n        893903.97, 1795536.61, 1837991.78, 3315892.56, 3702218.25,\n        7196600.46, 9809757.50, 12372961.36, 10301396.85, 8987170.60,\n        5973947.89, 3323136.42, -37697.57, -3298589.65, -3889246.12,\n        -1015849.30, 3802669.52, 5581996.86, 6294558.71, 5807430.31,\n        7233896.67, 7631729.03, 7022937.31, 3189671.28, -668632.89,\n        -2274453.26, -1565272.36, -1769165.45, -723583.21, -855404.64,\n        535520.92, 1867850.83, 1662266.57, 427243.53, -1897311.12,\n        -3423878.90, -5693396.54, -7049208.37, -6077556.13, -2618241.81,\n        1033556.80, 2981836.38, 3561095.51, 3103739.89, 5237643.39,\n        6315373.90, 7045295.99, 6379582.91, 5152512.94, 3742759.40,\n        1681683.04, -1804724.14, -3656312.87, -5583088.33, -4818681.46,\n        -6239572.74, -3743805.45, -1602577.08, 1657190.95, 3161907.74,\n        1953764.06, -11193.28, -2786891.53, -3829245.49, -3533254.15,\n        -2533443.92, -3912036.95, -3529222.98, -3882956.52, -2533840.49,\n        -2177918.39, 481942.72, 1201885.33, 1208587.28, -1308939.63,\n        -1096155.05, -1645324.52, -694717.55, -2327001.67, -3226028.75,\n        -4386381.98, -3827063.06, -3641030.80, -2071193.78, -1035711.30,\n        1251389.20, 1167548.76, 1628404.62, -702216.88, -1523181.74,\n        -2708040.86, -4101542.10, -5203831.74, -9205593.91, -10110967.66,\n        -9896185.26, -8321992.32, -9712290.66, -8552730.08, -7242998.44,\n        -3227285.45, -4222090.20, -4614038.31, -6769390.38, -4088958.86,\n        -1250958.20, 162266.12, -2424767.88, -4214918.55, -4462566.02,\n        -4776181.10, -10673340.15, -12916730.67, -10653331.95, -2675863.99,\n        1967103.64, 421406.40, -1877082.43, -1851431.29, 2215797.93,\n        6185896.18, 3512294.23, 1880789.18, 1216281.60, 1922721.41,\n        1037173.44, -3360160.22, -6723648.25, -11153059.20, -12372088.51,\n        -11873416.10, -5902018.10, -3990282.48, 512452.89, 342481.23,\n        3243109.80, 3384424.51, 8702513.24, 13928083.80, 22276329.22,\n        22530155.27, 20905021.47, 15328057.73, 13997083.11, 9630157.59,\n        5944389.91, 2602485.32, 3372507.20, 7039016.78, 10097385.97,\n        9234450.73, 3552576.10, -1755700.59, -7044046.31, -7411219.62,\n        -5305042.09, -1811188.65, 1714196.56, 4645392.27, 6499064.62,\n        7090472.15, 5615708.94, 2758580.75, 1140942.22, 332154.68,\n        -1608879.10, -2953649.80,\n    };\n}\n\ntest \"AdvanceDeclineOscillator EMA\" {\n    const allocator = testing.allocator;\n\n    const highs = testHighs();\n    const lows = testLows();\n    const closes = testCloses();\n    const volumes = testVolumes();\n    const expected = testExpectedEMA();\n\n    var adosc = try AdvanceDeclineOscillator.init(allocator, .{\n        .fast_length = 3,\n        .slow_length = 10,\n        .moving_average_type = .ema,\n    });\n    defer adosc.deinit();\n    adosc.fixSlices();\n\n    // EMA with length 10 has lookback = 9. First 9 values are NaN.\n    for (0..252) |i| {\n        const v = adosc.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);\n\n        if (i < 9) {\n            try testing.expect(math.isNan(v));\n            try testing.expect(!adosc.isPrimed());\n            continue;\n        }\n\n        try testing.expect(!math.isNan(v));\n        try testing.expect(adosc.isPrimed());\n\n        const got = roundTo(v, 2);\n        const exp = roundTo(expected[i], 2);\n        try testing.expectEqual(exp, got);\n    }\n}\n\ntest \"AdvanceDeclineOscillator SMA\" {\n    const allocator = testing.allocator;\n\n    const highs = testHighs();\n    const lows = testLows();\n    const closes = testCloses();\n    const volumes = testVolumes();\n    const expected = testExpectedSMA();\n\n    var adosc = try AdvanceDeclineOscillator.init(allocator, .{\n        .fast_length = 3,\n        .slow_length = 10,\n        .moving_average_type = .sma,\n    });\n    defer adosc.deinit();\n    adosc.fixSlices();\n\n    for (0..252) |i| {\n        const v = adosc.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);\n\n        if (i < 9) {\n            try testing.expect(math.isNan(v));\n            try testing.expect(!adosc.isPrimed());\n            continue;\n        }\n\n        try testing.expect(!math.isNan(v));\n        try testing.expect(adosc.isPrimed());\n\n        const got = roundTo(v, 2);\n        const exp = roundTo(expected[i], 2);\n        try testing.expectEqual(exp, got);\n    }\n}\n\ntest \"AdvanceDeclineOscillator updateBar\" {\n    const allocator = testing.allocator;\n\n    const highs = testHighs();\n    const lows = testLows();\n    const closes = testCloses();\n    const volumes = testVolumes();\n    const expected = testExpectedEMA();\n\n    var adosc = try AdvanceDeclineOscillator.init(allocator, .{\n        .fast_length = 3,\n        .slow_length = 10,\n        .moving_average_type = .ema,\n    });\n    defer adosc.deinit();\n    adosc.fixSlices();\n\n    for (0..15) |i| {\n        const bar = Bar{\n            .time = 0,\n            .open = highs[i],\n            .high = highs[i],\n            .low = lows[i],\n            .close = closes[i],\n            .volume = volumes[i],\n        };\n\n        const output = adosc.updateBar(&bar);\n        const scalar_val = output.slice()[0].scalar.value;\n\n        if (i < 9) {\n            try testing.expect(math.isNan(scalar_val));\n            continue;\n        }\n\n        const got = roundTo(scalar_val, 2);\n        const exp = roundTo(expected[i], 2);\n        try testing.expectEqual(exp, got);\n    }\n}\n\ntest \"AdvanceDeclineOscillator NaN\" {\n    const allocator = testing.allocator;\n\n    var adosc = try AdvanceDeclineOscillator.init(allocator, .{\n        .fast_length = 3,\n        .slow_length = 10,\n        .moving_average_type = .ema,\n    });\n    defer adosc.deinit();\n    adosc.fixSlices();\n\n    try testing.expect(math.isNan(adosc.update(math.nan(f64))));\n    try testing.expect(math.isNan(adosc.updateHlcv(math.nan(f64), 1, 2, 3)));\n    try testing.expect(math.isNan(adosc.updateHlcv(1, math.nan(f64), 2, 3)));\n    try testing.expect(math.isNan(adosc.updateHlcv(1, 2, math.nan(f64), 3)));\n    try testing.expect(math.isNan(adosc.updateHlcv(1, 2, 3, math.nan(f64))));\n}\n\ntest \"AdvanceDeclineOscillator not primed initially\" {\n    const allocator = testing.allocator;\n\n    var adosc = try AdvanceDeclineOscillator.init(allocator, .{\n        .fast_length = 3,\n        .slow_length = 10,\n        .moving_average_type = .ema,\n    });\n    defer adosc.deinit();\n    adosc.fixSlices();\n\n    try testing.expect(!adosc.isPrimed());\n    try testing.expect(math.isNan(adosc.update(math.nan(f64))));\n    try testing.expect(!adosc.isPrimed());\n}\n\ntest \"AdvanceDeclineOscillator metadata\" {\n    const allocator = testing.allocator;\n\n    var adosc = try AdvanceDeclineOscillator.init(allocator, .{\n        .fast_length = 3,\n        .slow_length = 10,\n        .moving_average_type = .ema,\n    });\n    defer adosc.deinit();\n    adosc.fixSlices();\n\n    var meta: Metadata = undefined;\n    adosc.getMetadata(&meta);\n\n    try testing.expectEqual(Identifier.advance_decline_oscillator, meta.identifier);\n    try testing.expectEqualStrings(\"adosc(EMA3/EMA10)\", meta.mnemonic);\n    try testing.expectEqualStrings(\"Chaikin Advance-Decline Oscillator adosc(EMA3/EMA10)\", meta.description);\n    try testing.expectEqual(@as(usize, 1), meta.outputs_len);\n}\n\ntest \"AdvanceDeclineOscillator invalid params\" {\n    const allocator = testing.allocator;\n\n    try testing.expectError(error.InvalidFastLength, AdvanceDeclineOscillator.init(allocator, .{ .fast_length = 1 }));\n    try testing.expectError(error.InvalidSlowLength, AdvanceDeclineOscillator.init(allocator, .{ .slow_length = 1 }));\n}\n",
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig"
}
```

**Output:**
```
Wrote file successfully.

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/icalc/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.identifier" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.factory.factory" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/iconf/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.descriptors" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.core.outputs.heatmap" could not be resolved
ERROR [26:6] Import "py.indicators.core.outputs.shape" could not be resolved
ERROR [27:6] Import "py.indicators.core.pane" could not be resolved
ERROR [28:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [31:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/ifres/main.py">
ERROR [16:6] Import "py.indicators.core.frequency_response" could not be resolved
ERROR [17:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [18:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [21:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>
```

---

## User

---

## Assistant (Compaction · claude-opus-4.6 · 60.6s)



## Goal

Port the `indicators/` module from Go/TypeScript to Zig. Python and Rust ports are complete. Currently porting indicators in small batches by author family.

## Instructions

- Go and TS code is the reference implementation. Python is the reference for other modules.
- All ports must match reference output to tolerance specified by Go tests (typically 1e-8).
- Port one language at a time: **Python → Rust → Zig**. Python and Rust are done.
- Port indicators in small batches by author family, starting with simplest, factory last, cmd tools last.
- Zig naming: `snake_case` files, `camelCase` functions/fields, `PascalCase` types/enums, lowercase enum values.
- **Each indicator lives in its own subfolder** (e.g., `patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig`)
- **Author folder naming**: Use `snake_case` matching Rust (e.g., `patrick_mulloy`), NOT Go's `patrickmulloy` style.
- **Output enums are 1-based** (matching Go's `iota+1`).
- **fixSlices() pattern**: After `init()` returns by value, slices point into old stack frame. Must call `fixSlices()` to repoint them. `fixSlices` must be `pub`.
- **Metadata out-pointer pattern**: `getMetadata(out: *Metadata)` returns void.
- **No allocator needed** for EMA-like indicators (no window). SMA and Variance need allocator for ring buffer.
- **MaUnion pattern**: For indicators that support both SMA and EMA sub-indicators, use a tagged union with `update`, `isPrimed`, `deinit` methods (see APO, Stochastic, MACD, BollingerBands for examples).
- **Per-output metadata**: `OutputText` struct has only `mnemonic` and `description` fields (no suffix fields). Each output needs a fully-formed string built with `bufPrint` in `getMetadata`. See MACD or BollingerBands for the pattern.
- **Metadata fields are plain `[]const u8`**, NOT functions — use `meta.mnemonic` not `meta.mnemonic()`.

## Discoveries

### Key Type Corrections (learned fixing MACD/PPO compilation)
- **Bar/Quote/Trade function types**: Use `bar_component.BarFunc`, `quote_component.QuoteFunc`, `trade_component.TradeFunc` (NOT `BarExtractFn` etc.)
- **Scalar.time is `i64`**, not `f64`
- **OutputArray**: Is a struct with `.values` array and `.len`, NOT a bare array. Use `OutputArray.append(.{ .scalar = .{ .time = t, .value = v } })` to add outputs, `out.slice()[i].scalar.value` to read them.
- **Metadata fields**: `outputs_len` (not `num_outputs`), `outputs_buf[i]` (not `outputs[i]`)
- **Band output type**: Use `band_mod.Band.new(time, lower, upper)` and `Band.empty(time)` for NaN bands. Access via `out.slice()[5].band`.

### Per-output metadata pattern (learned from BollingerBands)
- `build_metadata.OutputText` only has `{ .mnemonic, .description }` — NO suffix fields
- Must create per-output buf+bufPrint for each output's mnemonic and description in `getMetadata`

### HLCV Indicators Pattern (AdvanceDecline, AdvanceDeclineOscillator)
- These indicators need HLCV data from bars, not just a single scalar component.
- They override `updateBar` to extract `sample.high, sample.low, sample.close, sample.volume` directly.
- For scalar/quote/trade updates, H=L=C so the range is 0 and AD is unchanged.
- The `update()` method calls `updateHlcv(sample, sample, sample, 1)`.
- AdvanceDecline has NO parameters (no allocator needed, empty `fixSlices`/`deinit`).
- AdvanceDeclineOscillator uses MaUnion (SMA/EMA) and needs allocator for SMA case.

### 2-Indicator Families Identified (7 total)
| Family | Indicators |
|--------|-----------|
| patrickmulloy | DEMA, TEMA |
| timtillson | T2EMA, T3EMA |
| geraldappel | MACD, PPO |
| johnbollinger | BollingerBands, BollingerBandsTrend |
| larrywilliams | UltimateOscillator, WilliamsPercentR |
| marcchaikin | AdvanceDecline, AdvanceDeclineOscillator |
| custom | GoertzelSpectrum, MaximumEntropySpectrum |

## Accomplished

### Fully Complete ✅ (636 tests pass)
- **Python port**: All 63 indicators, factory, frequency_response, cmd tools. 1142 tests pass.
- **Rust port**: All 67 indicators, factory, frequency_response, cmd tools. 985 tests pass.
- **Zig Core Types**: All 17 core type files complete.
- **Zig Common Indicators**: All 13 complete (SMA, WMA, TMA, EMA, Momentum, ROC, ROCP, ROCR, Variance, StdDev, APO, LinearRegression, PearsonsCC).
- **Zig Single-Indicator Families**: All 9 complete (BalanceOfPower, OnBalanceVolume, CCI, TEMO, MFI, KAMA, JMA, Stochastic, ATCF).
- **Zig patrick_mulloy family**: DEMA and TEMA both complete and passing.
- **Zig timtillson family**: T2EMA and T3EMA both complete and passing.
- **Zig gerald_appel family**: MACD and PPO both complete and passing.
- **Zig john_bollinger family**: BollingerBands and BollingerBandsTrend both complete and passing.
- **Zig larry_williams family**: WilliamsPercentR and UltimateOscillator both complete and passing.
- **636 total Zig tests pass**.

### Just Written — NOT YET COMPILED/TESTED ⚠️
- **Zig marc_chaikin family**: AdvanceDecline and AdvanceDeclineOscillator files written but:
  1. **NOT registered** in `indicators.zig` (no `pub const` imports added)
  2. **NOT in comptime block** in `indicators.zig`
  3. **`zig build test` NOT run**

### Immediate Next Steps
1. **Add imports to `indicators.zig`**: Add `pub const advance_decline = @import("marc_chaikin/advance_decline/advance_decline.zig");` and `pub const advance_decline_oscillator = @import("marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig");`
2. **Add to comptime block**: `_ = advance_decline;` and `_ = advance_decline_oscillator;`
3. **Run `zig build test`** and fix any compilation/test errors
4. **Verify test count increases** (should be ~650+ after)

### Remaining Work After marc_chaikin
1. **custom** (GoertzelSpectrum, MaximumEntropySpectrum)
2. **tusharchande** (3 indicators)
3. **welleswilder** (12 indicators)
4. **johnehlers** (23 indicators)
5. Factory, frequency_response, cmd tools

## Relevant files / directories

### Registration file (NEEDS imports + comptime block update):
- `zig/src/indicators/indicators.zig` — Root module; needs advance_decline and advance_decline_oscillator added

### Just created (need registration + compile + test):
- `zig/src/indicators/marc_chaikin/advance_decline/advance_decline.zig` — AdvanceDecline, written ⚠️
- `zig/src/indicators/marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig` — AdvanceDeclineOscillator, written ⚠️

### Go reference implementations (already read):
- `go/indicators/marcchaikin/advancedecline/advancedecline.go` — Reference impl (CLV × Volume cumulative)
- `go/indicators/marcchaikin/advancedecline/advancedecline_test.go` — Test data (252 entries, 2-digit rounding)
- `go/indicators/marcchaikin/advancedecline/params.go` — No params
- `go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator.go` — Reference impl (FastMA(AD) - SlowMA(AD))
- `go/indicators/marcchaikin/advancedeclineoscillator/advancedeclineoscillator_test.go` — Test data (252 entries, EMA + SMA, 2-digit rounding, lookback 9)
- `go/indicators/marcchaikin/advancedeclineoscillator/params.go` — FastLength=3, SlowLength=10, SMA/EMA enum, FirstIsAverage

### Key pattern references:
- `zig/src/indicators/igor_livshin/balance_of_power/balance_of_power.zig` — HLCV bar override pattern (updateOhlc)
- `zig/src/indicators/common/absolute_price_oscillator/absolute_price_oscillator.zig` — MaUnion + mnemonic_buf/fixSlices pattern
- `zig/src/indicators/gerald_appel/moving_average_convergence_divergence/moving_average_convergence_divergence.zig` — Multi-output + MaUnion + per-output metadata
- `zig/src/indicators/john_bollinger/bollinger_bands/bollinger_bands.zig` — 6-output + Band type + per-output metadata
- `zig/src/indicators/core/indicator.zig` — OutputArray, OutputValue, Indicator vtable interface
- `zig/src/indicators/core/build_metadata.zig` — buildMetadata helper
- `zig/src/indicators/core/identifier.zig` — All identifier enum values

### Build:
- `zig/build.zig` — Single `indicators` module

### Skill files:
- `.opencode/skills/indicator-architecture/SKILL.md`
- `.opencode/skills/indicator-conversion/SKILL.md`

---

## User

---

## Assistant (Build · claude-opus-4.6 · 6.3s)

_Thinking:_

Let me continue with the immediate next steps: register the marc_chaikin indicators in indicators.zig, then compile and test.

Let me register the marc_chaikin indicators and test them.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig</path>
<type>file</type>
<content>1: // Root module for the indicators library.
2: // Re-exports all core types and sub-modules.
3: 
4: // --- Core enums ---
5: pub const shape = @import("core/outputs/shape.zig");
6: pub const role = @import("core/role.zig");
7: pub const pane = @import("core/pane.zig");
8: pub const adaptivity = @import("core/adaptivity.zig");
9: pub const input_requirement = @import("core/input_requirement.zig");
10: pub const volume_usage = @import("core/volume_usage.zig");
11: 
12: // --- Core types ---
13: pub const identifier = @import("core/identifier.zig");
14: pub const output_descriptor = @import("core/output_descriptor.zig");
15: pub const output_metadata = @import("core/output_metadata.zig");
16: pub const metadata = @import("core/metadata.zig");
17: pub const descriptor = @import("core/descriptor.zig");
18: pub const descriptors = @import("core/descriptors.zig");
19: pub const build_metadata = @import("core/build_metadata.zig");
20: pub const component_triple_mnemonic = @import("core/component_triple_mnemonic.zig");
21: 
22: // --- Indicator interface and outputs ---
23: pub const indicator = @import("core/indicator.zig");
24: pub const line_indicator = @import("core/line_indicator.zig");
25: 
26: // --- Output types ---
27: pub const band = @import("core/outputs/band.zig");
28: pub const heatmap = @import("core/outputs/heatmap.zig");
29: pub const polyline = @import("core/outputs/polyline.zig");
30: 
31: // --- Convenience type aliases ---
32: pub const Identifier = identifier.Identifier;
33: pub const Shape = shape.Shape;
34: pub const Role = role.Role;
35: pub const Pane = pane.Pane;
36: pub const Adaptivity = adaptivity.Adaptivity;
37: pub const InputRequirement = input_requirement.InputRequirement;
38: pub const VolumeUsage = volume_usage.VolumeUsage;
39: pub const OutputDescriptor = output_descriptor.OutputDescriptor;
40: pub const OutputMetadata = output_metadata.OutputMetadata;
41: pub const Metadata = metadata.Metadata;
42: pub const Descriptor = descriptor.Descriptor;
43: pub const Indicator = indicator.Indicator;
44: pub const OutputValue = indicator.OutputValue;
45: pub const OutputArray = indicator.OutputArray;
46: pub const LineIndicator = line_indicator.LineIndicator;
47: pub const Band = band.Band;
48: pub const Heatmap = heatmap.Heatmap;
49: pub const Polyline = polyline.Polyline;
50: pub const Point = polyline.Point;
51: 
52: // --- Functions ---
53: pub const descriptorOf = descriptor.descriptorOf;
54: pub const allDescriptors = descriptor.allDescriptors;
55: pub const buildMetadata = build_metadata.buildMetadata;
56: pub const componentTripleMnemonic = component_triple_mnemonic.componentTripleMnemonic;
57: 
58: // --- Indicator implementations ---
59: pub const simple_moving_average = @import("common/simple_moving_average/simple_moving_average.zig");
60: pub const weighted_moving_average = @import("common/weighted_moving_average/weighted_moving_average.zig");
61: pub const triangular_moving_average = @import("common/triangular_moving_average/triangular_moving_average.zig");
62: pub const exponential_moving_average = @import("common/exponential_moving_average/exponential_moving_average.zig");
63: pub const momentum = @import("common/momentum/momentum.zig");
64: pub const rate_of_change = @import("common/rate_of_change/rate_of_change.zig");
65: pub const rate_of_change_percent = @import("common/rate_of_change_percent/rate_of_change_percent.zig");
66: pub const rate_of_change_ratio = @import("common/rate_of_change_ratio/rate_of_change_ratio.zig");
67: pub const variance = @import("common/variance/variance.zig");
68: pub const standard_deviation = @import("common/standard_deviation/standard_deviation.zig");
69: pub const absolute_price_oscillator = @import("common/absolute_price_oscillator/absolute_price_oscillator.zig");
70: pub const linear_regression = @import("common/linear_regression/linear_regression.zig");
71: pub const pearsons_correlation_coefficient = @import("common/pearsons_correlation_coefficient/pearsons_correlation_coefficient.zig");
72: pub const balance_of_power = @import("igor_livshin/balance_of_power/balance_of_power.zig");
73: pub const on_balance_volume = @import("joseph_granville/on_balance_volume/on_balance_volume.zig");
74: pub const commodity_channel_index = @import("donald_lambert/commodity_channel_index/commodity_channel_index.zig");
75: pub const triple_exponential_moving_average_oscillator = @import("jack_hutson/triple_exponential_moving_average_oscillator/triple_exponential_moving_average_oscillator.zig");
76: pub const money_flow_index = @import("gene_quong/money_flow_index/money_flow_index.zig");
77: pub const kaufman_adaptive_moving_average = @import("perry_kaufman/kaufman_adaptive_moving_average/kaufman_adaptive_moving_average.zig");
78: pub const jurik_moving_average = @import("mark_jurik/jurik_moving_average/jurik_moving_average.zig");
79: pub const stochastic = @import("george_lane/stochastic/stochastic.zig");
80: pub const adaptive_trend_and_cycle_filter = @import("vladimir_kravchuk/adaptive_trend_and_cycle_filter/adaptive_trend_and_cycle_filter.zig");
81: pub const double_exponential_moving_average = @import("patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig");
82: pub const triple_exponential_moving_average = @import("patrick_mulloy/triple_exponential_moving_average/triple_exponential_moving_average.zig");
83: pub const t2_exponential_moving_average = @import("tim_tillson/t2_exponential_moving_average/t2_exponential_moving_average.zig");
84: pub const t3_exponential_moving_average = @import("tim_tillson/t3_exponential_moving_average/t3_exponential_moving_average.zig");
85: pub const percentage_price_oscillator = @import("gerald_appel/percentage_price_oscillator/percentage_price_oscillator.zig");
86: pub const moving_average_convergence_divergence = @import("gerald_appel/moving_average_convergence_divergence/moving_average_convergence_divergence.zig");
87: pub const bollinger_bands = @import("john_bollinger/bollinger_bands/bollinger_bands.zig");
88: pub const bollinger_bands_trend = @import("john_bollinger/bollinger_bands_trend/bollinger_bands_trend.zig");
89: pub const williams_percent_r = @import("larry_williams/williams_percent_r/williams_percent_r.zig");
90: pub const ultimate_oscillator = @import("larry_williams/ultimate_oscillator/ultimate_oscillator.zig");
91: 
92: // Force-include tests from sub-modules.
93: comptime {
94:     _ = identifier;
95:     _ = simple_moving_average;
96:     _ = weighted_moving_average;
97:     _ = triangular_moving_average;
98:     _ = exponential_moving_average;
99:     _ = momentum;
100:     _ = rate_of_change;
101:     _ = rate_of_change_percent;
102:     _ = rate_of_change_ratio;
103:     _ = variance;
104:     _ = standard_deviation;
105:     _ = absolute_price_oscillator;
106:     _ = linear_regression;
107:     _ = pearsons_correlation_coefficient;
108:     _ = balance_of_power;
109:     _ = on_balance_volume;
110:     _ = commodity_channel_index;
111:     _ = triple_exponential_moving_average_oscillator;
112:     _ = money_flow_index;
113:     _ = kaufman_adaptive_moving_average;
114:     _ = jurik_moving_average;
115:     _ = stochastic;
116:     _ = adaptive_trend_and_cycle_filter;
117:     _ = double_exponential_moving_average;
118:     _ = triple_exponential_moving_average;
119:     _ = t2_exponential_moving_average;
120:     _ = t3_exponential_moving_average;
121:     _ = percentage_price_oscillator;
122:     _ = moving_average_convergence_divergence;
123:     _ = bollinger_bands;
124:     _ = bollinger_bands_trend;
125:     _ = williams_percent_r;
126:     _ = ultimate_oscillator;
127: }

(End of file - total 127 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 6.5s)

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "oldString": "pub const ultimate_oscillator = @import(\"larry_williams/ultimate_oscillator/ultimate_oscillator.zig\");\n\n// Force-include tests from sub-modules.\ncomptime {",
  "newString": "pub const ultimate_oscillator = @import(\"larry_williams/ultimate_oscillator/ultimate_oscillator.zig\");\npub const advance_decline = @import(\"marc_chaikin/advance_decline/advance_decline.zig\");\npub const advance_decline_oscillator = @import(\"marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig\");\n\n// Force-include tests from sub-modules.\ncomptime {"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 5.5s)

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "oldString": "    _ = ultimate_oscillator;\n}",
  "newString": "    _ = ultimate_oscillator;\n    _ = advance_decline;\n    _ = advance_decline_oscillator;\n}"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 7.8s)

Now let's compile and test:

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | tail -40",
  "description": "Build and test zig indicators",
  "timeout": 120000
}
```

**Output:**
```
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 2 pass (2 total) 5ms MaxRSS:11M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 42 pass (42 total) 22ms MaxRSS:11M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test success 5ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test success 9ms MaxRSS:10M
|  +- compile test Debug native cached 11ms MaxRSS:38M
+- run test success 7ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test success 12ms MaxRSS:10M
|  +- compile test Debug native cached 4ms MaxRSS:39M
+- run test 55 pass (55 total) 9ms MaxRSS:11M
|  +- compile test Debug native cached 10ms MaxRSS:39M
+- run test 155 pass (155 total) 40ms MaxRSS:11M
|  +- compile test Debug native cached 8ms MaxRSS:39M
+- run test 4 pass (4 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 4ms MaxRSS:39M
+- run test 2 pass (2 total) 10ms MaxRSS:10M
|  +- compile test Debug native cached 13ms MaxRSS:39M
+- run test 2 pass (2 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 16ms MaxRSS:39M
+- run test 6 pass (6 total) 6ms MaxRSS:11M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 7 pass (7 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test success 7ms MaxRSS:11M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test success 9ms MaxRSS:10M
|  +- compile test Debug native cached 14ms MaxRSS:39M
+- run test 11 pass (11 total) 8ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 10 pass (10 total) 6ms MaxRSS:11M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 4 pass (4 total) 8ms MaxRSS:10M
|  +- compile test Debug native cached 8ms MaxRSS:38M
+- run test 283 pass (283 total) 41ms MaxRSS:11M
   +- compile test Debug native success 843ms MaxRSS:179M


```

---

## Assistant (Build · claude-opus-4.6 · 6.0s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | grep -E \"pass|fail|error\" | grep -v cached",
  "description": "Count total tests"
}
```

**Output:**
```
Build Summary: 45/45 steps succeeded; 650/650 tests passed
+- run test 3 pass (3 total) 10ms MaxRSS:10M
+- run test 53 pass (53 total) 9ms MaxRSS:11M
+- run test 11 pass (11 total) 8ms MaxRSS:10M
+- run test 2 pass (2 total) 24ms MaxRSS:10M
+- run test 42 pass (42 total) 20ms MaxRSS:11M
+- run test 55 pass (55 total) 15ms MaxRSS:10M
+- run test 155 pass (155 total) 42ms MaxRSS:10M
+- run test 4 pass (4 total) 5ms MaxRSS:11M
+- run test 2 pass (2 total) 7ms MaxRSS:11M
+- run test 2 pass (2 total) 10ms MaxRSS:10M
+- run test 6 pass (6 total) 7ms MaxRSS:10M
+- run test 7 pass (7 total) 7ms MaxRSS:10M
+- run test 11 pass (11 total) 10ms MaxRSS:10M
+- run test 10 pass (10 total) 10ms MaxRSS:10M
+- run test 4 pass (4 total) 8ms MaxRSS:10M
+- run test 283 pass (283 total) 53ms MaxRSS:11M

```

---

## Assistant (Build · claude-opus-4.6 · 6.9s)

All **650 tests pass** (up from 636). marc_chaikin family (AdvanceDecline + AdvanceDeclineOscillator) is complete.

Now let's continue with the next family: **tusharchande** (3 indicators). Let me check what's there.

**Tool: bash**

**Input:**
```json
{
  "command": "ls go/indicators/tusharchande/",
  "description": "List tusharchande indicators"
}
```

**Output:**
```
aroon
chandemomentumoscillator
doc.go
stochasticrelativestrengthindex

```

---

## Assistant (Build · claude-opus-4.6 · 8.3s)

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "marc_chaikin family: AdvanceDecline + AdvanceDeclineOscillator registered and passing (650 tests)",
      "status": "completed",
      "priority": "high"
    },
    {
      "content": "tusharchande family: Aroon",
      "status": "in_progress",
      "priority": "high"
    },
    {
      "content": "tusharchande family: ChandeMomentumOscillator",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "tusharchande family: StochasticRelativeStrengthIndex",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "welleswilder family (12 indicators)",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "johnehlers family (23 indicators)",
      "status": "pending",
      "priority": "medium"
    },
    {
      "content": "custom family (GoertzelSpectrum, MaximumEntropySpectrum)",
      "status": "pending",
      "priority": "medium"
    },
    {
      "content": "Factory, frequency_response, cmd tools",
      "status": "pending",
      "priority": "low"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "marc_chaikin family: AdvanceDecline + AdvanceDeclineOscillator registered and passing (650 tests)",
    "status": "completed",
    "priority": "high"
  },
  {
    "content": "tusharchande family: Aroon",
    "status": "in_progress",
    "priority": "high"
  },
  {
    "content": "tusharchande family: ChandeMomentumOscillator",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "tusharchande family: StochasticRelativeStrengthIndex",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "welleswilder family (12 indicators)",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "johnehlers family (23 indicators)",
    "status": "pending",
    "priority": "medium"
  },
  {
    "content": "custom family (GoertzelSpectrum, MaximumEntropySpectrum)",
    "status": "pending",
    "priority": "medium"
  },
  {
    "content": "Factory, frequency_response, cmd tools",
    "status": "pending",
    "priority": "low"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 59.1s)

Let me read all three tusharchande Go reference implementations in parallel.

**Tool: task**

**Input:**
```json
{
  "description": "Read tusharchande Go sources",
  "prompt": "Read all Go source files for the three tusharchande indicators and return their full contents. The files are in:\n- go/indicators/tusharchande/aroon/ (all .go files)\n- go/indicators/tusharchande/chandemomentumoscillator/ (all .go files)\n- go/indicators/tusharchande/stochasticrelativestrengthindex/ (all .go files)\n\nFor each indicator, return:\n1. The main implementation file content (full)\n2. The params file content (full)\n3. The test file content (full)\n4. The output enum file content if it exists\n\nAlso note the identifier used, number of outputs, and any special patterns (like HLCV overrides, MaUnion usage, etc.).",
  "subagent_type": "general"
}
```

**Output:**
```
task_id: ses_22d693644ffe7nuqXuxpY15clJ (for resuming to continue this task if needed)

<task_result>
Here's a summary of all three Tushar Chande indicators:

---

## 1. Aroon (`go/indicators/tusharchande/aroon/`)

- **Identifier:** `core.Aroon`
- **Outputs:** 3 — `Up`, `Down`, `Osc` (iota+1)
- **Mnemonic format:** `aroon(14)`
- **Constructor:** `NewAroon(*AroonParams)` — named type, not `New`
- **Params:** Just `Length int` (default 14, min 2)
- **Special patterns:**
  - **Does NOT use `LineIndicator`** — custom struct with its own `mu sync.RWMutex`
  - **HLCV override:** Custom `UpdateBar` uses `sample.High`/`sample.Low` directly (not a component function). `UpdateQuote` uses `(Bid+Ask)/2` for both high/low. `UpdateScalar`/`UpdateTrade` use the single value for both.
  - **No component params** (no BarComponent/QuoteComponent/TradeComponent)
  - Circular buffer (`highBuf`, `lowBuf`) of size `length+1` tracking highest-high and lowest-low indices
  - `Update(high, low float64)` returns 3 values
  - NaN guard at entry; returns NaN triple if not primed
  - Test tolerance is `1e-6` (looser than most indicators)

**Files:** `aroon.go` (220 lines), `params.go` (16 lines), `output.go` (94 lines), `output_test.go` (140 lines), `aroon_test.go` (546 lines)

---

## 2. Chande Momentum Oscillator (`go/indicators/tusharchande/chandemomentumoscillator/`)

- **Identifier:** `core.ChandeMomentumOscillator`
- **Outputs:** 1 — `Value` (iota+1)
- **Mnemonic format:** `cmo(14)` or `cmo(5, hl/2)` with component triple
- **Constructor:** `New(*Params)` — short name
- **Params:** `Length int` + `BarComponent`/`QuoteComponent`/`TradeComponent` (default 14, min 1)
- **Special patterns:**
  - **Uses `core.LineIndicator` (embedded)** — standard single-sample `Update(float64) float64` pattern
  - Component triple mnemonic via `core.ComponentTripleMnemonic(bc, qc, tc)`
  - Zero-value sentinel for components (0 = use default, don't show in mnemonic)
  - Ring buffer of deltas, tracks `gainSum`/`lossSum` incrementally
  - `epsilon = 1e-12` for zero-denominator guard (returns 0, not NaN)
  - Notes explicitly that TA-Lib's CMO is incompatible (TA-Lib smooths like RSI; this follows Chande's book)
  - Test tolerance `1e-13`

**Files:** `chandemomentumoscillator.go` (198 lines), `params.go` (33 lines), `output.go` (77 lines), `output_test.go` (136 lines), `chandemomentumoscillator_test.go` (484 lines)

---

## 3. Stochastic Relative Strength Index (`go/indicators/tusharchande/stochasticrelativestrengthindex/`)

- **Identifier:** `core.StochasticRelativeStrengthIndex`
- **Outputs:** 2 — `FastK`, `FastD` (iota+1)
- **Mnemonic format:** `stochrsi(14/14/SMA3)` or `stochrsi(14/14/EMA3, hl/2)`
- **Constructor:** `NewStochasticRelativeStrengthIndex(*StochasticRelativeStrengthIndexParams)` — long name
- **Params:** `Length`, `FastKLength`, `FastDLength`, `MovingAverageType` (SMA/EMA enum), `FirstIsAverage` bool, + component triple
- **Special patterns:**
  - **Does NOT embed `LineIndicator`** — custom struct with manual `barFunc`/`quoteFunc`/`tradeFunc` fields
  - **MaUnion-like pattern:** `MovingAverageType` enum (SMA=0, EMA=1) selects Fast-D smoother; uses `lineUpdater` interface (custom local interface with `Update`/`IsPrimed`)
  - **`passthrough` struct** for FastD period of 1 (no-op smoother)
  - Composes internal `relativestrengthindex.RelativeStrengthIndex`, `simplemovingaverage`, `exponentialmovingaverage`
  - Circular buffer for RSI values; stochastic formula applied to RSI (min/max over window)
  - `Update(float64)` returns 2 values `(fastK, fastD)`
  - `diff == 0` → `fastK = 0` (not NaN)
  - Test tolerance `1e-4` (looser due to RSI smoothing chain)
  - `FirstIsAverage` param passed through to EMA for seeding control

**Files:** `stochasticrelativestrengthindex.go` (287 lines), `params.go` (67 lines), `output.go` (86 lines), `output_test.go` (136 lines), `stochasticrelativestrengthindex_test.go` (387 lines)
</task_result>
```

---

## Assistant (Build · claude-opus-4.6 · 7.1s)

Good overview. Let me now read the actual Go source files for all three to get the implementation details.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/aroon.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/aroon.go</path>
<type>file</type>
<content>1: package aroon
2: 
3: import (
4: 	"fmt"
5: 	"math"
6: 	"sync"
7: 
8: 	"zpano/entities"
9: 	"zpano/indicators/core"
10: )
11: 
12: // Aroon is Tushar Chande's Aroon indicator.
13: //
14: // The Aroon indicator measures the number of periods since the highest high
15: // and lowest low within a lookback window. It produces three outputs:
16: //   - Up: 100 * (Length - periods since highest high) / Length
17: //   - Down: 100 * (Length - periods since lowest low) / Length
18: //   - Osc: Up - Down
19: //
20: // The indicator requires bar data (high, low). For scalar, quote, and
21: // trade updates, the single value substitutes for both.
22: //
23: // Reference:
24: //
25: // Chande, Tushar S. (1995). "The New Technical Trader". John Wiley & Sons.
26: type Aroon struct {
27: 	mu sync.RWMutex
28: 
29: 	length int
30: 	factor float64
31: 
32: 	// Circular buffers for high and low values (size = length+1).
33: 	highBuf     []float64
34: 	lowBuf      []float64
35: 	bufferIndex int
36: 	count       int
37: 
38: 	// Tracked indices of highest high and lowest low (absolute indices).
39: 	highestIndex int
40: 	lowestIndex  int
41: 
42: 	up     float64
43: 	down   float64
44: 	osc    float64
45: 	primed bool
46: 
47: 	mnemonic string
48: }
49: 
50: // NewAroon returns an instance of the indicator created using supplied parameters.
51: func NewAroon(p *AroonParams) (*Aroon, error) {
52: 	const (
53: 		invalid   = "invalid aroon parameters"
54: 		fmts      = "%s: %s"
55: 		minLength = 2
56: 	)
57: 
58: 	if p.Length < minLength {
59: 		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
60: 	}
61: 
62: 	mnemonic := fmt.Sprintf("aroon(%d)", p.Length)
63: 	windowSize := p.Length + 1
64: 
65: 	return &Aroon{
66: 		length:   p.Length,
67: 		factor:   100.0 / float64(p.Length), //nolint:mnd
68: 		highBuf:  make([]float64, windowSize),
69: 		lowBuf:   make([]float64, windowSize),
70: 		up:       math.NaN(),
71: 		down:     math.NaN(),
72: 		osc:      math.NaN(),
73: 		mnemonic: mnemonic,
74: 	}, nil
75: }
76: 
77: // IsPrimed indicates whether the indicator is primed.
78: func (a *Aroon) IsPrimed() bool {
79: 	a.mu.RLock()
80: 	defer a.mu.RUnlock()
81: 
82: 	return a.primed
83: }
84: 
85: // Metadata describes the output data of the indicator.
86: func (a *Aroon) Metadata() core.Metadata {
87: 	desc := "Aroon " + a.mnemonic
88: 
89: 	return core.BuildMetadata(
90: 		core.Aroon,
91: 		a.mnemonic,
92: 		desc,
93: 		[]core.OutputText{
94: 			{Mnemonic: a.mnemonic + " up", Description: desc + " Up"},
95: 			{Mnemonic: a.mnemonic + " down", Description: desc + " Down"},
96: 			{Mnemonic: a.mnemonic + " osc", Description: desc + " Oscillator"},
97: 		},
98: 	)
99: }
100: 
101: // Update updates the indicator given the next bar's high and low values.
102: // Returns Up, Down, and Osc.
103: func (a *Aroon) Update(high, low float64) (float64, float64, float64) {
104: 	if math.IsNaN(high) || math.IsNaN(low) {
105: 		return math.NaN(), math.NaN(), math.NaN()
106: 	}
107: 
108: 	a.mu.Lock()
109: 	defer a.mu.Unlock()
110: 
111: 	windowSize := a.length + 1
112: 	today := a.count
113: 
114: 	// Store in circular buffer.
115: 	pos := a.bufferIndex
116: 	a.highBuf[pos] = high
117: 	a.lowBuf[pos] = low
118: 	a.bufferIndex = (a.bufferIndex + 1) % windowSize
119: 	a.count++
120: 
121: 	// Need at least length+1 bars (indices 0..length).
122: 	if a.count < windowSize {
123: 		return a.up, a.down, a.osc
124: 	}
125: 
126: 	trailingIndex := today - a.length
127: 
128: 	if a.count == windowSize {
129: 		// First time: scan entire window to find highest/lowest.
130: 		a.highestIndex = trailingIndex
131: 		a.lowestIndex = trailingIndex
132: 
133: 		for i := trailingIndex + 1; i <= today; i++ {
134: 			bufPos := i % windowSize
135: 
136: 			if a.highBuf[bufPos] >= a.highBuf[a.highestIndex%windowSize] {
137: 				a.highestIndex = i
138: 			}
139: 
140: 			if a.lowBuf[bufPos] <= a.lowBuf[a.lowestIndex%windowSize] {
141: 				a.lowestIndex = i
142: 			}
143: 		}
144: 	} else {
145: 		// Subsequent: optimized update.
146: 		// Check if tracked indices fell out of window.
147: 		if a.highestIndex < trailingIndex {
148: 			a.highestIndex = trailingIndex
149: 
150: 			for i := trailingIndex + 1; i <= today; i++ {
151: 				bufPos := i % windowSize
152: 				if a.highBuf[bufPos] >= a.highBuf[a.highestIndex%windowSize] {
153: 					a.highestIndex = i
154: 				}
155: 			}
156: 		} else if high >= a.highBuf[a.highestIndex%windowSize] {
157: 			a.highestIndex = today
158: 		}
159: 
160: 		if a.lowestIndex < trailingIndex {
161: 			a.lowestIndex = trailingIndex
162: 
163: 			for i := trailingIndex + 1; i <= today; i++ {
164: 				bufPos := i % windowSize
165: 				if a.lowBuf[bufPos] <= a.lowBuf[a.lowestIndex%windowSize] {
166: 					a.lowestIndex = i
167: 				}
168: 			}
169: 		} else if low <= a.lowBuf[a.lowestIndex%windowSize] {
170: 			a.lowestIndex = today
171: 		}
172: 	}
173: 
174: 	a.up = a.factor * float64(a.length-(today-a.highestIndex))
175: 	a.down = a.factor * float64(a.length-(today-a.lowestIndex))
176: 	a.osc = a.up - a.down
177: 
178: 	if !a.primed {
179: 		a.primed = true
180: 	}
181: 
182: 	return a.up, a.down, a.osc
183: }
184: 
185: // UpdateScalar updates the indicator given the next scalar sample.
186: func (a *Aroon) UpdateScalar(sample *entities.Scalar) core.Output {
187: 	v := sample.Value
188: 	up, down, osc := a.Update(v, v)
189: 
190: 	output := make([]any, 3) //nolint:mnd
191: 	output[0] = entities.Scalar{Time: sample.Time, Value: up}
192: 	output[1] = entities.Scalar{Time: sample.Time, Value: down}
193: 	output[2] = entities.Scalar{Time: sample.Time, Value: osc}
194: 
195: 	return output
196: }
197: 
198: // UpdateBar updates the indicator given the next bar sample.
199: func (a *Aroon) UpdateBar(sample *entities.Bar) core.Output {
200: 	up, down, osc := a.Update(sample.High, sample.Low)
201: 
202: 	output := make([]any, 3) //nolint:mnd
203: 	output[0] = entities.Scalar{Time: sample.Time, Value: up}
204: 	output[1] = entities.Scalar{Time: sample.Time, Value: down}
205: 	output[2] = entities.Scalar{Time: sample.Time, Value: osc}
206: 
207: 	return output
208: }
209: 
210: // UpdateQuote updates the indicator given the next quote sample.
211: func (a *Aroon) UpdateQuote(sample *entities.Quote) core.Output {
212: 	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd
213: 
214: 	return a.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
215: }
216: 
217: // UpdateTrade updates the indicator given the next trade sample.
218: func (a *Aroon) UpdateTrade(sample *entities.Trade) core.Output {
219: 	return a.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: sample.Price})
220: }

(End of file - total 220 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/params.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/params.go</path>
<type>file</type>
<content>1: package aroon
2: 
3: // AroonParams describes parameters to create an instance of the indicator.
4: type AroonParams struct {
5: 	// Length is the lookback period for the Aroon calculation.
6: 	//
7: 	// The value should be greater than 1. The default value is 14.
8: 	Length int
9: }
10: 
11: // DefaultParams returns a [AroonParams] value populated with conventional defaults.
12: func DefaultParams() *AroonParams {
13: 	return &AroonParams{
14: 		Length: 14,
15: 	}
16: }

(End of file - total 16 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/output.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/output.go</path>
<type>file</type>
<content>1: //nolint:dupl
2: package aroon
3: 
4: import (
5: 	"bytes"
6: 	"fmt"
7: )
8: 
9: // Output describes the outputs of the indicator.
10: type Output int
11: 
12: const (
13: 	// The Aroon Up line.
14: 	Up Output = iota + 1
15: 
16: 	// The Aroon Down line.
17: 	Down
18: 
19: 	// The Aroon Oscillator (Up - Down).
20: 	Osc
21: 
22: 	outputLast
23: )
24: 
25: const (
26: 	upStr      = "up"
27: 	downStr    = "down"
28: 	oscStr     = "osc"
29: 	unknownStr = "unknown"
30: )
31: 
32: // String implements the Stringer interface.
33: func (o Output) String() string {
34: 	switch o {
35: 	case Up:
36: 		return upStr
37: 	case Down:
38: 		return downStr
39: 	case Osc:
40: 		return oscStr
41: 	default:
42: 		return unknownStr
43: 	}
44: }
45: 
46: // IsKnown determines if this output is known.
47: func (o Output) IsKnown() bool {
48: 	return o >= Up && o < outputLast
49: }
50: 
51: // MarshalJSON implements the Marshaler interface.
52: func (o Output) MarshalJSON() ([]byte, error) {
53: 	const (
54: 		errFmt = "cannot marshal '%s': unknown aroon output"
55: 		extra  = 2   // Two bytes for quotes.
56: 		dqc    = '"' // Double quote character.
57: 	)
58: 
59: 	s := o.String()
60: 	if s == unknownStr {
61: 		return nil, fmt.Errorf(errFmt, s)
62: 	}
63: 
64: 	b := make([]byte, 0, len(s)+extra)
65: 	b = append(b, dqc)
66: 	b = append(b, s...)
67: 	b = append(b, dqc)
68: 
69: 	return b, nil
70: }
71: 
72: // UnmarshalJSON implements the Unmarshaler interface.
73: func (o *Output) UnmarshalJSON(data []byte) error {
74: 	const (
75: 		errFmt = "cannot unmarshal '%s': unknown aroon output"
76: 		dqs    = "\"" // Double quote string.
77: 	)
78: 
79: 	d := bytes.Trim(data, dqs)
80: 	s := string(d)
81: 
82: 	switch s {
83: 	case upStr:
84: 		*o = Up
85: 	case downStr:
86: 		*o = Down
87: 	case oscStr:
88: 		*o = Osc
89: 	default:
90: 		return fmt.Errorf(errFmt, s)
91: 	}
92: 
93: 	return nil
94: }

(End of file - total 94 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/aroon_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/aroon/aroon_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package aroon
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: // Standard test data (252 entries) — same H/L as TrueRange, Stochastic, etc.
16: func testInputHigh() []float64 {
17: 	return []float64{
18: 		93.25, 94.94, 96.375, 96.19, 96.0, 94.72, 95.0, 93.72, 92.47, 92.75,
19: 		96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
20: 		90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
21: 		87.155, 87.25, 87.375, 88.97, 90.0, 89.845, 86.97, 85.94, 84.75, 85.47,
22: 		84.47, 88.5, 89.47, 90.0, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
23: 		90.0, 88.875, 89.0, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
24: 		90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92.0, 91.94,
25: 		89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
26: 		104.5, 105.5, 106.125, 107.94, 106.25, 107.0, 108.75, 110.94, 110.94, 114.22,
27: 		123.0, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0, 118.315,
28: 		116.87, 116.75, 113.87, 114.62, 115.31, 116.0, 121.69, 119.87, 120.87, 116.75,
29: 		116.5, 116.0, 118.31, 121.5, 122.0, 121.44, 125.75, 127.75, 124.19, 124.44,
30: 		125.75, 124.69, 125.31, 132.0, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
31: 		138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
32: 		125.12, 126.5, 128.69, 126.62, 126.69, 126.0, 123.12, 121.87, 124.0, 127.0,
33: 		124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
34: 		124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126.0, 128.5, 126.75, 129.75,
35: 		132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135.0, 132.38, 131.44, 130.88,
36: 		129.63, 127.25, 127.81, 125.0, 126.81, 124.75, 122.81, 122.25, 121.06, 120.0,
37: 		123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109.0,
38: 		108.5, 113.06, 93.0, 94.62, 95.12, 96.0, 95.56, 95.31, 99.0, 98.81,
39: 		96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
40: 		95.0, 94.87, 98.25, 105.12, 108.44, 109.87, 105.0, 106.0, 104.94, 104.5,
41: 		104.44, 106.31, 112.87, 116.5, 119.19, 121.0, 122.12, 111.94, 112.75, 110.19,
42: 		107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110.0, 110.75, 110.5,
43: 		110.5, 109.5,
44: 	}
45: }
46: 
47: func testInputLow() []float64 {
48: 	return []float64{
49: 		90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92.0, 89.75, 89.44, 90.625,
50: 		92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
51: 		87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
52: 		84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
53: 		82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89.0, 88.565, 90.095,
54: 		89.0, 86.47, 84.0, 83.315, 82.0, 83.25, 84.75, 85.28, 87.19, 88.44,
55: 		88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92.0, 90.53, 89.97, 88.815,
56: 		86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
57: 		101.72, 101.72, 103.155, 105.69, 103.655, 104.0, 105.53, 108.53, 108.75, 107.75,
58: 		117.0, 118.0, 116.0, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
59: 		112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
60: 		112.56, 114.0, 114.75, 118.87, 119.0, 119.75, 122.62, 123.0, 121.75, 121.56,
61: 		123.12, 122.19, 122.75, 124.37, 128.0, 129.5, 130.81, 130.63, 132.13, 133.88,
62: 		135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
63: 		122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121.0,
64: 		121.12, 118.94, 119.81, 121.0, 122.0, 124.5, 126.56, 123.5, 121.25, 121.06,
65: 		122.31, 121.0, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128.0,
66: 		128.38, 130.69, 131.63, 134.38, 132.0, 131.94, 131.94, 129.56, 123.75, 126.0,
67: 		126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120.0, 119.62, 115.5, 116.75,
68: 		119.06, 119.06, 115.06, 111.06, 113.12, 110.0, 105.0, 104.69, 103.87, 104.69,
69: 		105.44, 107.0, 89.0, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
70: 		93.69, 93.5, 90.0, 90.19, 90.5, 92.12, 94.12, 94.87, 93.0, 93.87,
71: 		93.0, 92.62, 93.56, 98.37, 104.44, 106.0, 101.81, 104.12, 103.37, 102.12,
72: 		102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
73: 		104.5, 105.75, 108.62, 107.75, 108.06, 108.0, 108.19, 108.12, 109.06, 108.75,
74: 		108.56, 106.62,
75: 	}
76: }
77: 
78: func testTime() time.Time {
79: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
80: }
81: 
82: // Expected outputs from excel-length-14-aaron.csv (252 rows, period=14).
83: // First 14 rows (indices 0-13) are NaN; first valid output at index 14.
84: // Columns: Up, Down, Osc.
85: type expectedRow struct {
86: 	up   float64
87: 	down float64
88: 	osc  float64
89: }
90: 
91: //nolint:funlen
92: func testExpected() []expectedRow {
93: 	nan := math.NaN()
94: 	return []expectedRow{
95: 		{nan, nan, nan}, // 0
96: 		{nan, nan, nan}, // 1
97: 		{nan, nan, nan}, // 2
98: 		{nan, nan, nan}, // 3
99: 		{nan, nan, nan}, // 4
100: 		{nan, nan, nan}, // 5
101: 		{nan, nan, nan}, // 6
102: 		{nan, nan, nan}, // 7
103: 		{nan, nan, nan}, // 8
104: 		{nan, nan, nan}, // 9
105: 		{nan, nan, nan}, // 10
106: 		{nan, nan, nan}, // 11
107: 		{nan, nan, nan}, // 12
108: 		{nan, nan, nan}, // 13
109: 		{78.571428571428600, 100.000000000000000, -21.428571428571400}, // 14
110: 		{71.428571428571400, 92.857142857142900, -21.428571428571400},  // 15
111: 		{64.285714285714300, 85.714285714285700, -21.428571428571400},  // 16
112: 		{57.142857142857100, 78.571428571428600, -21.428571428571400},  // 17
113: 		{50.000000000000000, 71.428571428571400, -21.428571428571400},  // 18
114: 		{42.857142857142900, 64.285714285714300, -21.428571428571400},  // 19
115: 		{35.714285714285700, 57.142857142857100, -21.428571428571400},  // 20
116: 		{28.571428571428600, 100.000000000000000, -71.428571428571400}, // 21
117: 		{21.428571428571400, 100.000000000000000, -78.571428571428600}, // 22
118: 		{14.285714285714300, 100.000000000000000, -85.714285714285700}, // 23
119: 		{7.142857142857140, 100.000000000000000, -92.857142857142900},  // 24
120: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 25
121: 		{0.000000000000000, 92.857142857142900, -92.857142857142900},   // 26
122: 		{21.428571428571400, 85.714285714285700, -64.285714285714300},  // 27
123: 		{14.285714285714300, 78.571428571428600, -64.285714285714300},  // 28
124: 		{7.142857142857140, 71.428571428571400, -64.285714285714300},   // 29
125: 		{0.000000000000000, 64.285714285714300, -64.285714285714300},   // 30
126: 		{14.285714285714300, 57.142857142857100, -42.857142857142900},  // 31
127: 		{7.142857142857140, 50.000000000000000, -42.857142857142900},   // 32
128: 		{0.000000000000000, 42.857142857142900, -42.857142857142900},   // 33
129: 		{0.000000000000000, 35.714285714285700, -35.714285714285700},   // 34
130: 		{92.857142857142900, 28.571428571428600, 64.285714285714300},   // 35
131: 		{85.714285714285700, 21.428571428571400, 64.285714285714300},   // 36
132: 		{78.571428571428600, 14.285714285714300, 64.285714285714300},   // 37
133: 		{71.428571428571400, 7.142857142857140, 64.285714285714300},    // 38
134: 		{64.285714285714300, 0.000000000000000, 64.285714285714300},    // 39
135: 		{57.142857142857100, 0.000000000000000, 57.142857142857100},    // 40
136: 		{50.000000000000000, 92.857142857142900, -42.857142857142900},  // 41
137: 		{42.857142857142900, 85.714285714285700, -42.857142857142900},  // 42
138: 		{100.000000000000000, 78.571428571428600, 21.428571428571400},  // 43
139: 		{100.000000000000000, 71.428571428571400, 28.571428571428600},  // 44
140: 		{92.857142857142900, 64.285714285714300, 28.571428571428600},   // 45
141: 		{100.000000000000000, 57.142857142857100, 42.857142857142900},  // 46
142: 		{92.857142857142900, 50.000000000000000, 42.857142857142900},   // 47
143: 		{85.714285714285700, 42.857142857142900, 42.857142857142900},   // 48
144: 		{78.571428571428600, 35.714285714285700, 42.857142857142900},   // 49
145: 		{71.428571428571400, 28.571428571428600, 42.857142857142900},   // 50
146: 		{64.285714285714300, 21.428571428571400, 42.857142857142900},   // 51
147: 		{57.142857142857100, 14.285714285714300, 42.857142857142900},   // 52
148: 		{50.000000000000000, 7.142857142857140, 42.857142857142900},    // 53
149: 		{42.857142857142900, 100.000000000000000, -57.142857142857100}, // 54
150: 		{35.714285714285700, 92.857142857142900, -57.142857142857100},  // 55
151: 		{28.571428571428600, 85.714285714285700, -57.142857142857100},  // 56
152: 		{21.428571428571400, 78.571428571428600, -57.142857142857100},  // 57
153: 		{14.285714285714300, 71.428571428571400, -57.142857142857100},  // 58
154: 		{7.142857142857140, 64.285714285714300, -57.142857142857200},   // 59
155: 		{0.000000000000000, 57.142857142857100, -57.142857142857100},   // 60
156: 		{14.285714285714300, 50.000000000000000, -35.714285714285700},  // 61
157: 		{100.000000000000000, 42.857142857142900, 57.142857142857100},  // 62
158: 		{100.000000000000000, 35.714285714285700, 64.285714285714300},  // 63
159: 		{100.000000000000000, 28.571428571428600, 71.428571428571400},  // 64
160: 		{100.000000000000000, 21.428571428571400, 78.571428571428600},  // 65
161: 		{100.000000000000000, 14.285714285714300, 85.714285714285700},  // 66
162: 		{92.857142857142900, 7.142857142857140, 85.714285714285700},    // 67
163: 		{85.714285714285700, 0.000000000000000, 85.714285714285700},    // 68
164: 		{78.571428571428600, 0.000000000000000, 78.571428571428600},    // 69
165: 		{71.428571428571400, 0.000000000000000, 71.428571428571400},    // 70
166: 		{64.285714285714300, 100.000000000000000, -35.714285714285700}, // 71
167: 		{57.142857142857100, 100.000000000000000, -42.857142857142900}, // 72
168: 		{50.000000000000000, 100.000000000000000, -50.000000000000000}, // 73
169: 		{42.857142857142900, 92.857142857142900, -50.000000000000000},  // 74
170: 		{100.000000000000000, 85.714285714285700, 14.285714285714300},  // 75
171: 		{100.000000000000000, 78.571428571428600, 21.428571428571400},  // 76
172: 		{100.000000000000000, 71.428571428571400, 28.571428571428600},  // 77
173: 		{100.000000000000000, 64.285714285714300, 35.714285714285700},  // 78
174: 		{92.857142857142900, 57.142857142857100, 35.714285714285700},   // 79
175: 		{85.714285714285700, 50.000000000000000, 35.714285714285700},   // 80
176: 		{78.571428571428600, 42.857142857142900, 35.714285714285700},   // 81
177: 		{71.428571428571400, 35.714285714285700, 35.714285714285700},   // 82
178: 		{100.000000000000000, 28.571428571428600, 71.428571428571400},  // 83
179: 		{92.857142857142900, 21.428571428571400, 71.428571428571400},   // 84
180: 		{85.714285714285700, 14.285714285714300, 71.428571428571400},   // 85
181: 		{100.000000000000000, 7.142857142857140, 92.857142857142900},   // 86
182: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 87
183: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 88
184: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 89
185: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 90
186: 		{92.857142857142900, 0.000000000000000, 92.857142857142900},    // 91
187: 		{85.714285714285700, 21.428571428571400, 64.285714285714300},   // 92
188: 		{78.571428571428600, 14.285714285714300, 64.285714285714300},   // 93
189: 		{71.428571428571400, 7.142857142857140, 64.285714285714300},    // 94
190: 		{64.285714285714300, 0.000000000000000, 64.285714285714300},    // 95
191: 		{57.142857142857100, 0.000000000000000, 57.142857142857100},    // 96
192: 		{50.000000000000000, 7.142857142857140, 42.857142857142900},    // 97
193: 		{42.857142857142900, 0.000000000000000, 42.857142857142900},    // 98
194: 		{35.714285714285700, 0.000000000000000, 35.714285714285700},    // 99
195: 		{28.571428571428600, 0.000000000000000, 28.571428571428600},    // 100
196: 		{21.428571428571400, 14.285714285714300, 7.142857142857140},    // 101
197: 		{14.285714285714300, 7.142857142857140, 7.142857142857140},     // 102
198: 		{7.142857142857140, 0.000000000000000, 7.142857142857140},      // 103
199: 		{0.000000000000000, 92.857142857142900, -92.857142857142900},   // 104
200: 		{0.000000000000000, 85.714285714285700, -85.714285714285700},   // 105
201: 		{100.000000000000000, 78.571428571428600, 21.428571428571400},  // 106
202: 		{92.857142857142900, 71.428571428571400, 21.428571428571400},   // 107
203: 		{85.714285714285700, 64.285714285714300, 21.428571428571400},   // 108
204: 		{78.571428571428600, 57.142857142857100, 21.428571428571400},   // 109
205: 		{71.428571428571400, 50.000000000000000, 21.428571428571400},   // 110
206: 		{64.285714285714300, 42.857142857142900, 21.428571428571400},   // 111
207: 		{57.142857142857100, 35.714285714285700, 21.428571428571400},   // 112
208: 		{50.000000000000000, 28.571428571428600, 21.428571428571400},   // 113
209: 		{100.000000000000000, 21.428571428571400, 78.571428571428600},  // 114
210: 		{92.857142857142900, 14.285714285714300, 78.571428571428600},   // 115
211: 		{100.000000000000000, 7.142857142857140, 92.857142857142900},   // 116
212: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 117
213: 		{92.857142857142900, 0.000000000000000, 92.857142857142900},    // 118
214: 		{85.714285714285700, 0.000000000000000, 85.714285714285700},    // 119
215: 		{78.571428571428600, 28.571428571428600, 50.000000000000000},   // 120
216: 		{71.428571428571400, 21.428571428571400, 50.000000000000000},   // 121
217: 		{64.285714285714300, 14.285714285714300, 50.000000000000000},   // 122
218: 		{100.000000000000000, 7.142857142857140, 92.857142857142900},   // 123
219: 		{92.857142857142900, 0.000000000000000, 92.857142857142900},    // 124
220: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 125
221: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 126
222: 		{92.857142857142900, 0.000000000000000, 92.857142857142900},    // 127
223: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 128
224: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 129
225: 		{100.000000000000000, 21.428571428571400, 78.571428571428600},  // 130
226: 		{100.000000000000000, 14.285714285714300, 85.714285714285700},  // 131
227: 		{92.857142857142900, 7.142857142857140, 85.714285714285700},    // 132
228: 		{85.714285714285700, 0.000000000000000, 85.714285714285700},    // 133
229: 		{78.571428571428600, 7.142857142857140, 71.428571428571400},    // 134
230: 		{71.428571428571400, 0.000000000000000, 71.428571428571400},    // 135
231: 		{64.285714285714300, 0.000000000000000, 64.285714285714300},    // 136
232: 		{57.142857142857100, 0.000000000000000, 57.142857142857100},    // 137
233: 		{50.000000000000000, 100.000000000000000, -50.000000000000000}, // 138
234: 		{42.857142857142900, 100.000000000000000, -57.142857142857100}, // 139
235: 		{35.714285714285700, 92.857142857142900, -57.142857142857100},  // 140
236: 		{28.571428571428600, 85.714285714285700, -57.142857142857100},  // 141
237: 		{21.428571428571400, 78.571428571428600, -57.142857142857100},  // 142
238: 		{14.285714285714300, 71.428571428571400, -57.142857142857100},  // 143
239: 		{7.142857142857140, 64.285714285714300, -57.142857142857200},   // 144
240: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 145
241: 		{21.428571428571400, 100.000000000000000, -78.571428571428600}, // 146
242: 		{14.285714285714300, 100.000000000000000, -85.714285714285700}, // 147
243: 		{7.142857142857140, 100.000000000000000, -92.857142857142900},  // 148
244: 		{0.000000000000000, 92.857142857142900, -92.857142857142900},   // 149
245: 		{0.000000000000000, 85.714285714285700, -85.714285714285700},   // 150
246: 		{0.000000000000000, 78.571428571428600, -78.571428571428600},   // 151
247: 		{28.571428571428600, 71.428571428571400, -42.857142857142900},  // 152
248: 		{21.428571428571400, 64.285714285714300, -42.857142857142900},  // 153
249: 		{14.285714285714300, 57.142857142857100, -42.857142857142900},  // 154
250: 		{7.142857142857140, 50.000000000000000, -42.857142857142900},   // 155
251: 		{0.000000000000000, 42.857142857142900, -42.857142857142900},   // 156
252: 		{100.000000000000000, 35.714285714285700, 64.285714285714300},  // 157
253: 		{92.857142857142900, 28.571428571428600, 64.285714285714300},   // 158
254: 		{85.714285714285700, 21.428571428571400, 64.285714285714300},   // 159
255: 		{78.571428571428600, 14.285714285714300, 64.285714285714300},   // 160
256: 		{71.428571428571400, 7.142857142857140, 64.285714285714300},    // 161
257: 		{64.285714285714300, 0.000000000000000, 64.285714285714300},    // 162
258: 		{57.142857142857100, 14.285714285714300, 42.857142857142900},   // 163
259: 		{50.000000000000000, 7.142857142857140, 42.857142857142900},    // 164
260: 		{42.857142857142900, 0.000000000000000, 42.857142857142900},    // 165
261: 		{35.714285714285700, 0.000000000000000, 35.714285714285700},    // 166
262: 		{28.571428571428600, 64.285714285714300, -35.714285714285700},  // 167
263: 		{21.428571428571400, 57.142857142857100, -35.714285714285700},  // 168
264: 		{100.000000000000000, 50.000000000000000, 50.000000000000000},  // 169
265: 		{100.000000000000000, 42.857142857142900, 57.142857142857100},  // 170
266: 		{100.000000000000000, 35.714285714285700, 64.285714285714300},  // 171
267: 		{100.000000000000000, 28.571428571428600, 71.428571428571400},  // 172
268: 		{100.000000000000000, 21.428571428571400, 78.571428571428600},  // 173
269: 		{92.857142857142900, 14.285714285714300, 78.571428571428600},   // 174
270: 		{85.714285714285700, 7.142857142857140, 78.571428571428600},    // 175
271: 		{78.571428571428600, 0.000000000000000, 78.571428571428600},    // 176
272: 		{71.428571428571400, 0.000000000000000, 71.428571428571400},    // 177
273: 		{64.285714285714300, 7.142857142857140, 57.142857142857200},    // 178
274: 		{57.142857142857100, 0.000000000000000, 57.142857142857100},    // 179
275: 		{50.000000000000000, 0.000000000000000, 50.000000000000000},    // 180
276: 		{42.857142857142900, 78.571428571428600, -35.714285714285700},  // 181
277: 		{35.714285714285700, 100.000000000000000, -64.285714285714300}, // 182
278: 		{28.571428571428600, 100.000000000000000, -71.428571428571400}, // 183
279: 		{21.428571428571400, 92.857142857142900, -71.428571428571400},  // 184
280: 		{14.285714285714300, 85.714285714285700, -71.428571428571400},  // 185
281: 		{7.142857142857140, 100.000000000000000, -92.857142857142900},  // 186
282: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 187
283: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 188
284: 		{7.142857142857140, 92.857142857142900, -85.714285714285700},   // 189
285: 		{0.000000000000000, 85.714285714285700, -85.714285714285700},   // 190
286: 		{0.000000000000000, 78.571428571428600, -78.571428571428600},   // 191
287: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 192
288: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 193
289: 		{0.000000000000000, 92.857142857142900, -92.857142857142900},   // 194
290: 		{7.142857142857140, 100.000000000000000, -92.857142857142900},  // 195
291: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 196
292: 		{7.142857142857140, 100.000000000000000, -92.857142857142900},  // 197
293: 		{0.000000000000000, 100.000000000000000, -100.000000000000000}, // 198
294: 		{0.000000000000000, 92.857142857142900, -92.857142857142900},   // 199
295: 		{28.571428571428600, 85.714285714285700, -57.142857142857100},  // 200
296: 		{21.428571428571400, 78.571428571428600, -57.142857142857100},  // 201
297: 		{14.285714285714300, 100.000000000000000, -85.714285714285700}, // 202
298: 		{7.142857142857140, 92.857142857142900, -85.714285714285700},   // 203
299: 		{0.000000000000000, 85.714285714285700, -85.714285714285700},   // 204
300: 		{0.000000000000000, 78.571428571428600, -78.571428571428600},   // 205
301: 		{0.000000000000000, 71.428571428571400, -71.428571428571400},   // 206
302: 		{7.142857142857140, 64.285714285714300, -57.142857142857200},   // 207
303: 		{0.000000000000000, 57.142857142857100, -57.142857142857100},   // 208
304: 		{0.000000000000000, 50.000000000000000, -50.000000000000000},   // 209
305: 		{35.714285714285700, 42.857142857142900, -7.142857142857140},   // 210
306: 		{28.571428571428600, 35.714285714285700, -7.142857142857150},   // 211
307: 		{21.428571428571400, 28.571428571428600, -7.142857142857140},   // 212
308: 		{14.285714285714300, 21.428571428571400, -7.142857142857140},   // 213
309: 		{7.142857142857140, 14.285714285714300, -7.142857142857140},    // 214
310: 		{0.000000000000000, 7.142857142857140, -7.142857142857140},     // 215
311: 		{42.857142857142900, 0.000000000000000, 42.857142857142900},    // 216
312: 		{35.714285714285700, 64.285714285714300, -28.571428571428600},  // 217
313: 		{28.571428571428600, 57.142857142857100, -28.571428571428600},  // 218
314: 		{21.428571428571400, 50.000000000000000, -28.571428571428600},  // 219
315: 		{14.285714285714300, 42.857142857142900, -28.571428571428600},  // 220
316: 		{7.142857142857140, 35.714285714285700, -28.571428571428600},   // 221
317: 		{0.000000000000000, 28.571428571428600, -28.571428571428600},   // 222
318: 		{100.000000000000000, 21.428571428571400, 78.571428571428600},  // 223
319: 		{100.000000000000000, 14.285714285714300, 85.714285714285700},  // 224
320: 		{100.000000000000000, 7.142857142857140, 92.857142857142900},   // 225
321: 		{92.857142857142900, 0.000000000000000, 92.857142857142900},    // 226
322: 		{85.714285714285700, 0.000000000000000, 85.714285714285700},    // 227
323: 		{78.571428571428600, 0.000000000000000, 78.571428571428600},    // 228
324: 		{71.428571428571400, 0.000000000000000, 71.428571428571400},    // 229
325: 		{64.285714285714300, 35.714285714285700, 28.571428571428600},   // 230
326: 		{57.142857142857100, 28.571428571428600, 28.571428571428600},   // 231
327: 		{100.000000000000000, 21.428571428571400, 78.571428571428600},  // 232
328: 		{100.000000000000000, 14.285714285714300, 85.714285714285700},  // 233
329: 		{100.000000000000000, 7.142857142857140, 92.857142857142900},   // 234
330: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 235
331: 		{100.000000000000000, 0.000000000000000, 100.000000000000000},  // 236
332: 		{92.857142857142900, 0.000000000000000, 92.857142857142900},    // 237
333: 		{85.714285714285700, 14.285714285714300, 71.428571428571400},   // 238
334: 		{78.571428571428600, 7.142857142857140, 71.428571428571400},    // 239
335: 		{71.428571428571400, 0.000000000000000, 71.428571428571400},    // 240
336: 		{64.285714285714300, 14.285714285714300, 50.000000000000000},   // 241
337: 		{57.142857142857100, 7.142857142857140, 50.000000000000000},    // 242
338: 		{50.000000000000000, 0.000000000000000, 50.000000000000000},    // 243
339: 		{42.857142857142900, 0.000000000000000, 42.857142857142900},    // 244
340: 		{35.714285714285700, 0.000000000000000, 35.714285714285700},    // 245
341: 		{28.571428571428600, 57.142857142857100, -28.571428571428600},  // 246
342: 		{21.428571428571400, 50.000000000000000, -28.571428571428600},  // 247
343: 		{14.285714285714300, 42.857142857142900, -28.571428571428600},  // 248
344: 		{7.142857142857140, 35.714285714285700, -28.571428571428600},   // 249
345: 		{0.000000000000000, 28.571428571428600, -28.571428571428600},   // 250
346: 		{7.142857142857140, 21.428571428571400, -14.285714285714300},   // 251
347: 	}
348: }
349: 
350: func TestAroon_Length14_FullData(t *testing.T) {
351: 	t.Parallel()
352: 
353: 	const tolerance = 1e-6
354: 
355: 	high := testInputHigh()
356: 	low := testInputLow()
357: 	expected := testExpected()
358: 
359: 	ind, err := NewAroon(&AroonParams{Length: 14})
360: 	if err != nil {
361: 		t.Fatal(err)
362: 	}
363: 
364: 	for i := 0; i < 252; i++ {
365: 		up, down, osc := ind.Update(high[i], low[i])
366: 
367: 		if math.IsNaN(expected[i].up) {
368: 			if !math.IsNaN(up) {
369: 				t.Errorf("[%d] Up: expected NaN, got %v", i, up)
370: 			}
371: 
372: 			continue
373: 		}
374: 
375: 		if math.Abs(up-expected[i].up) > tolerance {
376: 			t.Errorf("[%d] Up: expected %v, got %v", i, expected[i].up, up)
377: 		}
378: 
379: 		if math.Abs(down-expected[i].down) > tolerance {
380: 			t.Errorf("[%d] Down: expected %v, got %v", i, expected[i].down, down)
381: 		}
382: 
383: 		if math.Abs(osc-expected[i].osc) > tolerance {
384: 			t.Errorf("[%d] Osc: expected %v, got %v", i, expected[i].osc, osc)
385: 		}
386: 	}
387: }
388: 
389: func TestAroonIsPrimed(t *testing.T) {
390: 	t.Parallel()
391: 
392: 	ind, err := NewAroon(&AroonParams{Length: 14})
393: 	if err != nil {
394: 		t.Fatal(err)
395: 	}
396: 
397: 	high := testInputHigh()
398: 	low := testInputLow()
399: 
400: 	if ind.IsPrimed() {
401: 		t.Error("expected not primed initially")
402: 	}
403: 
404: 	// Feed first 14 bars (indices 0..13), should not be primed.
405: 	for i := 0; i < 14; i++ {
406: 		ind.Update(high[i], low[i])
407: 		if ind.IsPrimed() {
408: 			t.Errorf("[%d] expected not primed", i)
409: 		}
410: 	}
411: 
412: 	// Index 14: first primed value.
413: 	ind.Update(high[14], low[14])
414: 
415: 	if !ind.IsPrimed() {
416: 		t.Error("expected primed after index 14")
417: 	}
418: }
419: 
420: func TestAroonNaN(t *testing.T) {
421: 	t.Parallel()
422: 
423: 	ind, err := NewAroon(&AroonParams{Length: 14})
424: 	if err != nil {
425: 		t.Fatal(err)
426: 	}
427: 
428: 	up, down, osc := ind.Update(math.NaN(), 1.0)
429: 	if !math.IsNaN(up) {
430: 		t.Errorf("expected NaN Up, got %v", up)
431: 	}
432: 
433: 	if !math.IsNaN(down) {
434: 		t.Errorf("expected NaN Down, got %v", down)
435: 	}
436: 
437: 	if !math.IsNaN(osc) {
438: 		t.Errorf("expected NaN Osc, got %v", osc)
439: 	}
440: }
441: 
442: func TestAroonMetadata(t *testing.T) {
443: 	t.Parallel()
444: 
445: 	ind, err := NewAroon(&AroonParams{Length: 14})
446: 	if err != nil {
447: 		t.Fatal(err)
448: 	}
449: 
450: 	meta := ind.Metadata()
451: 
452: 	if meta.Identifier != core.Aroon {
453: 		t.Errorf("expected identifier Aroon, got %v", meta.Identifier)
454: 	}
455: 
456: 	exp := "aroon(14)"
457: 	if meta.Mnemonic != exp {
458: 		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
459: 	}
460: 
461: 	if len(meta.Outputs) != 3 {
462: 		t.Fatalf("expected 3 outputs, got %d", len(meta.Outputs))
463: 	}
464: 
465: 	if meta.Outputs[0].Kind != int(Up) {
466: 		t.Errorf("expected output 0 kind %d, got %d", Up, meta.Outputs[0].Kind)
467: 	}
468: 
469: 	if meta.Outputs[0].Shape != shape.Scalar {
470: 		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
471: 	}
472: 
473: 	if meta.Outputs[1].Kind != int(Down) {
474: 		t.Errorf("expected output 1 kind %d, got %d", Down, meta.Outputs[1].Kind)
475: 	}
476: 
477: 	if meta.Outputs[2].Kind != int(Osc) {
478: 		t.Errorf("expected output 2 kind %d, got %d", Osc, meta.Outputs[2].Kind)
479: 	}
480: }
481: 
482: func TestAroonUpdateBar(t *testing.T) {
483: 	t.Parallel()
484: 
485: 	const tolerance = 1e-6
486: 
487: 	high := testInputHigh()
488: 	low := testInputLow()
489: 	expected := testExpected()
490: 
491: 	ind, err := NewAroon(&AroonParams{Length: 14})
492: 	if err != nil {
493: 		t.Fatal(err)
494: 	}
495: 
496: 	tm := testTime()
497: 
498: 	for i := 0; i < 14; i++ {
499: 		bar := &entities.Bar{Time: tm, High: high[i], Low: low[i], Close: 0}
500: 		out := ind.UpdateBar(bar)
501: 
502: 		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
503: 		if !math.IsNaN(v) {
504: 			t.Errorf("[%d] expected NaN Up, got %v", i, v)
505: 		}
506: 	}
507: 
508: 	bar := &entities.Bar{Time: tm, High: high[14], Low: low[14], Close: 0}
509: 	out := ind.UpdateBar(bar)
510: 
511: 	up := out[0].(entities.Scalar).Value   //nolint:forcetypeassert
512: 	down := out[1].(entities.Scalar).Value //nolint:forcetypeassert
513: 	osc := out[2].(entities.Scalar).Value  //nolint:forcetypeassert
514: 
515: 	if math.Abs(up-expected[14].up) > tolerance {
516: 		t.Errorf("[14] Up: expected %v, got %v", expected[14].up, up)
517: 	}
518: 
519: 	if math.Abs(down-expected[14].down) > tolerance {
520: 		t.Errorf("[14] Down: expected %v, got %v", expected[14].down, down)
521: 	}
522: 
523: 	if math.Abs(osc-expected[14].osc) > tolerance {
524: 		t.Errorf("[14] Osc: expected %v, got %v", expected[14].osc, osc)
525: 	}
526: }
527: 
528: func TestAroonInvalidParams(t *testing.T) {
529: 	t.Parallel()
530: 
531: 	tests := []struct {
532: 		name   string
533: 		length int
534: 	}{
535: 		{"length too small", 1},
536: 		{"length zero", 0},
537: 		{"length negative", -1},
538: 	}
539: 
540: 	for _, tt := range tests {
541: 		_, err := NewAroon(&AroonParams{Length: tt.length})
542: 		if err == nil {
543: 			t.Errorf("%s: expected error, got nil", tt.name)
544: 		}
545: 	}
546: }

(End of file - total 546 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 9.9s)

Now let me read the other two indicators:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator.go</path>
<type>file</type>
<content>1: package chandemomentumoscillator
2: 
3: //nolint: gofumpt
4: import (
5: 	"fmt"
6: 	"math"
7: 	"sync"
8: 
9: 	"zpano/entities"
10: 	"zpano/indicators/core"
11: )
12: 
13: const epsilon = 1e-12
14: 
15: // ChandeMomentumOscillator is a momentum indicator based on the average
16: // of up samples and down samples over a specified length ℓ.
17: //
18: // The calculation formula is:
19: //
20: // CMOᵢ = 100 (SUᵢ-SDᵢ) / (SUᵢ + SDᵢ),
21: //
22: // where SUᵢ (sum up) is the sum of gains and SDᵢ (sum down)
23: // is the sum of losses over the chosen length [i-ℓ, i].
24: //
25: // The indicator is not primed during the first ℓ updates.
26: type ChandeMomentumOscillator struct {
27: 	mu sync.RWMutex
28: 	core.LineIndicator
29: 	length         int
30: 	count          int
31: 	ringBuffer     []float64
32: 	ringHead       int
33: 	previousSample float64
34: 	gainSum        float64
35: 	lossSum        float64
36: 	primed         bool
37: }
38: 
39: // New returns an instance of the indicator created using supplied parameters.
40: func New(p *Params) (*ChandeMomentumOscillator, error) {
41: 	const (
42: 		invalid = "invalid Chande momentum oscillator parameters"
43: 		fmts    = "%s: %s"
44: 		fmtw    = "%s: %w"
45: 		fmtn    = "cmo(%d%s)"
46: 		minlen  = 1
47: 	)
48: 
49: 	length := p.Length
50: 	if length < minlen {
51: 		return nil, fmt.Errorf(fmts, invalid, "length should be positive")
52: 	}
53: 
54: 	var (
55: 		err       error
56: 		barFunc   entities.BarFunc
57: 		quoteFunc entities.QuoteFunc
58: 		tradeFunc entities.TradeFunc
59: 	)
60: 
61: 	// Resolve defaults for component functions.
62: 	// A zero value means "use default, don't show in mnemonic".
63: 	bc := p.BarComponent
64: 	qc := p.QuoteComponent
65: 	tc := p.TradeComponent
66: 
67: 	if bc == 0 {
68: 		bc = entities.DefaultBarComponent
69: 	}
70: 
71: 	if qc == 0 {
72: 		qc = entities.DefaultQuoteComponent
73: 	}
74: 
75: 	if tc == 0 {
76: 		tc = entities.DefaultTradeComponent
77: 	}
78: 
79: 	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
80: 		return nil, fmt.Errorf(fmtw, invalid, err)
81: 	}
82: 
83: 	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
84: 		return nil, fmt.Errorf(fmtw, invalid, err)
85: 	}
86: 
87: 	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
88: 		return nil, fmt.Errorf(fmtw, invalid, err)
89: 	}
90: 
91: 	mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
92: 	desc := "Chande Momentum Oscillator " + mnemonic
93: 
94: 	c := &ChandeMomentumOscillator{
95: 		length:     length,
96: 		ringBuffer: make([]float64, length),
97: 	}
98: 
99: 	c.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, c.Update)
100: 
101: 	return c, nil
102: }
103: 
104: // IsPrimed indicates whether an indicator is primed.
105: func (s *ChandeMomentumOscillator) IsPrimed() bool {
106: 	s.mu.RLock()
107: 	defer s.mu.RUnlock()
108: 
109: 	return s.primed
110: }
111: 
112: // Metadata describes an output data of the indicator.
113: // It always has a single scalar output -- the calculated value of the indicator.
114: func (s *ChandeMomentumOscillator) Metadata() core.Metadata {
115: 	return core.BuildMetadata(
116: 		core.ChandeMomentumOscillator,
117: 		s.LineIndicator.Mnemonic,
118: 		s.LineIndicator.Description,
119: 		[]core.OutputText{
120: 			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
121: 		},
122: 	)
123: }
124: 
125: // Update updates the value of the Chande momentum oscillator given the next sample.
126: //
127: // The indicator is not primed during the first ℓ updates.
128: func (s *ChandeMomentumOscillator) Update(sample float64) float64 {
129: 	if math.IsNaN(sample) {
130: 		return sample
131: 	}
132: 
133: 	s.mu.Lock()
134: 	defer s.mu.Unlock()
135: 
136: 	s.count++
137: 	if s.count == 1 {
138: 		s.previousSample = sample
139: 
140: 		return math.NaN()
141: 	}
142: 
143: 	// New delta
144: 	delta := sample - s.previousSample
145: 	s.previousSample = sample
146: 
147: 	if !s.primed {
148: 		// Fill until we have s.length deltas (i.e., s.length+1 samples)
149: 		s.ringBuffer[s.ringHead] = delta
150: 		s.ringHead = (s.ringHead + 1) % s.length
151: 
152: 		if delta > 0 {
153: 			s.gainSum += delta
154: 		} else if delta < 0 {
155: 			s.lossSum += -delta
156: 		}
157: 
158: 		if s.count <= s.length {
159: 			return math.NaN()
160: 		}
161: 
162: 		// Now we have exactly s.length deltas in the buffer
163: 		s.primed = true
164: 	} else {
165: 		// Remove oldest delta and add the new one
166: 		old := s.ringBuffer[s.ringHead]
167: 		if old > 0 {
168: 			s.gainSum -= old
169: 		} else if old < 0 {
170: 			s.lossSum -= -old
171: 		}
172: 
173: 		s.ringBuffer[s.ringHead] = delta
174: 		s.ringHead = (s.ringHead + 1) % s.length
175: 
176: 		if delta > 0 {
177: 			s.gainSum += delta
178: 		} else if delta < 0 {
179: 			s.lossSum += -delta
180: 		}
181: 
182: 		// Clamp to avoid tiny negative sums from FP noise
183: 		if s.gainSum < 0 {
184: 			s.gainSum = 0
185: 		}
186: 
187: 		if s.lossSum < 0 {
188: 			s.lossSum = 0
189: 		}
190: 	}
191: 
192: 	den := s.gainSum + s.lossSum
193: 	if math.Abs(den) < epsilon {
194: 		return 0
195: 	}
196: 
197: 	return 100.0 * (s.gainSum - s.lossSum) / den
198: }

(End of file - total 198 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/params.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/params.go</path>
<type>file</type>
<content>1: package chandemomentumoscillator
2: 
3: import "zpano/entities"
4: 
5: // Params describes parameters to create an instance of the indicator.
6: type Params struct {
7: 	// Length is the length (the number of time periods, ℓ) which defines the window of price changes.
8: 	//
9: 	// The value should be greater than 0.
10: 	Length int
11: 
12: 	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
13: 	//
14: 	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
15: 	BarComponent entities.BarComponent
16: 
17: 	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
18: 	//
19: 	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
20: 	QuoteComponent entities.QuoteComponent
21: 
22: 	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
23: 	//
24: 	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
25: 	TradeComponent entities.TradeComponent
26: }
27: 
28: // DefaultParams returns a [Params] value populated with conventional defaults.
29: func DefaultParams() *Params {
30: 	return &Params{
31: 		Length: 14,
32: 	}
33: }

(End of file - total 33 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package chandemomentumoscillator
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: //nolint:lll
16: // Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
17: //    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
18: //
19: // TA Lib uses incorrect CMO calculation which is incompatible with Chande's book.
20: // TA Lib smooths the CMO values in the same way as RSI does, but Chande didn't smooth the values.
21: // We don't use TA Lib test data here.
22: 
23: func testChandeMomentumOscillatorTime() time.Time {
24: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
25: }
26: 
27: //nolint:dupl
28: func testChandeMomentumOscillatorBookLength10Input() []float64 {
29: 	// Chande, T.S. and Kroll, S.
30: 	// The New Technical Trader: Boost Your Profit by Plugging Into the Latest Indicators
31: 	// ISBN 9780471597803
32: 	// Wiley Finance, 1994
33: 	// Page 96, Table 5.1 Calculating 10-day CMO Using T-Bond06/93 Data
34: 	return []float64{
35: 		101.0313, 101.0313, 101.1250, 101.9687, 102.7813,
36: 		103.0000, 102.9687, 103.0625, 102.9375, 102.7188,
37: 		102.7500, 102.9063, 102.9687,
38: 	}
39: }
40: 
41: //nolint:dupl
42: func testChandeMomentumOscillatorBookLength10Output() []float64 {
43: 	return []float64{
44: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
45: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
46: 
47: 		// 69.62, 71.43, 71.08,
48: 		69.61963786608334, 71.42857142857143, 71.08377992828775,
49: 	}
50: }
51: 
52: //nolint:dupl
53: func testChandeMomentumOscillatorInput() []float64 {
54: 	return []float64{
55: 		91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
56: 		96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
57: 		88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
58: 		85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
59: 		83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
60: 		89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
61: 		88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
62: 		88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
63: 		102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
64: 		112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
65: 		110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
66: 		116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
67: 		124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
68: 		132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
69: 		136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
70: 		125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
71: 		123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
72: 		122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
73: 		132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
74: 		130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
75: 		117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
76: 		107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
77: 		94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
78: 		95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
79: 		105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
80: 		113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
81: 		108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
82: 	}
83: }
84: 
85: func TestChandeMomentumOscillatorUpdate(t *testing.T) { //nolint: funlen
86: 	t.Parallel()
87: 
88: 	check := func(index int, exp, act float64) {
89: 		t.Helper()
90: 
91: 		if math.Abs(exp-act) > 1e-13 {
92: 			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
93: 		}
94: 	}
95: 
96: 	checkNaN := func(index int, act float64) {
97: 		t.Helper()
98: 
99: 		if !math.IsNaN(act) {
100: 			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
101: 		}
102: 	}
103: 
104: 	t.Run("length = 10 (book)", func(t *testing.T) {
105: 		t.Parallel()
106: 		input := testChandeMomentumOscillatorBookLength10Input()
107: 		output := testChandeMomentumOscillatorBookLength10Output()
108: 		cmo := testChandeMomentumOscillatorCreate(10)
109: 
110: 		for i := 0; i < 10; i++ {
111: 			checkNaN(i, cmo.Update(input[i]))
112: 		}
113: 
114: 		for i := 10; i < len(input); i++ {
115: 			act := cmo.Update(input[i])
116: 			check(i, output[i], act)
117: 		}
118: 
119: 		checkNaN(0, cmo.Update(math.NaN()))
120: 	})
121: }
122: 
123: func TestChandeMomentumOscillatorUpdateEntity(t *testing.T) { //nolint: funlen
124: 	t.Parallel()
125: 
126: 	const (
127: 		l   = 2
128: 		inp = 3.
129: 		exp = 100.
130: 	)
131: 
132: 	time := testChandeMomentumOscillatorTime()
133: 	check := func(act core.Output) {
134: 		t.Helper()
135: 
136: 		if len(act) != 1 {
137: 			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
138: 		}
139: 
140: 		s, ok := act[0].(entities.Scalar)
141: 		if !ok {
142: 			t.Error("output is not scalar")
143: 		}
144: 
145: 		if s.Time != time {
146: 			t.Errorf("time is incorrect: expected %v, actual %v", time, s.Time)
147: 		}
148: 
149: 		if math.Abs(s.Value-exp) > 1e-13 {
150: 			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
151: 		}
152: 	}
153: 
154: 	t.Run("update scalar", func(t *testing.T) {
155: 		t.Parallel()
156: 
157: 		s := entities.Scalar{Time: time, Value: inp}
158: 		cmo := testChandeMomentumOscillatorCreate(l)
159: 		cmo.Update(0.)
160: 		cmo.Update(0.)
161: 		check(cmo.UpdateScalar(&s))
162: 	})
163: 
164: 	t.Run("update bar", func(t *testing.T) {
165: 		t.Parallel()
166: 
167: 		b := entities.Bar{Time: time, Close: inp}
168: 		cmo := testChandeMomentumOscillatorCreate(l)
169: 		cmo.Update(0.)
170: 		cmo.Update(0.)
171: 		check(cmo.UpdateBar(&b))
172: 	})
173: 
174: 	t.Run("update quote", func(t *testing.T) {
175: 		t.Parallel()
176: 
177: 		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
178: 		cmo := testChandeMomentumOscillatorCreate(l)
179: 		cmo.Update(0.)
180: 		cmo.Update(0.)
181: 		check(cmo.UpdateQuote(&q))
182: 	})
183: 
184: 	t.Run("update trade", func(t *testing.T) {
185: 		t.Parallel()
186: 
187: 		r := entities.Trade{Time: time, Price: inp}
188: 		cmo := testChandeMomentumOscillatorCreate(l)
189: 		cmo.Update(0.)
190: 		cmo.Update(0.)
191: 		check(cmo.UpdateTrade(&r))
192: 	})
193: }
194: 
195: func TestChandeMomentumOscillatorIsPrimed(t *testing.T) { //nolint:funlen
196: 	t.Parallel()
197: 
198: 	input := testChandeMomentumOscillatorInput()
199: 	check := func(index int, exp, act bool) {
200: 		t.Helper()
201: 
202: 		if exp != act {
203: 			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
204: 		}
205: 	}
206: 
207: 	t.Run("length = 1", func(t *testing.T) {
208: 		t.Parallel()
209: 		cmo := testChandeMomentumOscillatorCreate(1)
210: 
211: 		check(-1, false, cmo.IsPrimed())
212: 
213: 		for i := 0; i < 1; i++ {
214: 			cmo.Update(input[i])
215: 			check(i, false, cmo.IsPrimed())
216: 		}
217: 
218: 		for i := 1; i < len(input); i++ {
219: 			cmo.Update(input[i])
220: 			check(i, true, cmo.IsPrimed())
221: 		}
222: 	})
223: 
224: 	t.Run("length = 2", func(t *testing.T) {
225: 		t.Parallel()
226: 		cmo := testChandeMomentumOscillatorCreate(2)
227: 
228: 		check(-1, false, cmo.IsPrimed())
229: 
230: 		for i := 0; i < 2; i++ {
231: 			cmo.Update(input[i])
232: 			check(i, false, cmo.IsPrimed())
233: 		}
234: 
235: 		for i := 2; i < len(input); i++ {
236: 			cmo.Update(input[i])
237: 			check(i, true, cmo.IsPrimed())
238: 		}
239: 	})
240: 
241: 	t.Run("length = 3", func(t *testing.T) {
242: 		t.Parallel()
243: 		cmo := testChandeMomentumOscillatorCreate(3)
244: 
245: 		check(-1, false, cmo.IsPrimed())
246: 
247: 		for i := 0; i < 3; i++ {
248: 			cmo.Update(input[i])
249: 			check(i, false, cmo.IsPrimed())
250: 		}
251: 
252: 		for i := 3; i < len(input); i++ {
253: 			cmo.Update(input[i])
254: 			check(i, true, cmo.IsPrimed())
255: 		}
256: 	})
257: 
258: 	t.Run("length = 5", func(t *testing.T) {
259: 		t.Parallel()
260: 		cmo := testChandeMomentumOscillatorCreate(5)
261: 
262: 		check(-1, false, cmo.IsPrimed())
263: 
264: 		for i := 0; i < 5; i++ {
265: 			cmo.Update(input[i])
266: 			check(i, false, cmo.IsPrimed())
267: 		}
268: 
269: 		for i := 5; i < len(input); i++ {
270: 			cmo.Update(input[i])
271: 			check(i, true, cmo.IsPrimed())
272: 		}
273: 	})
274: 
275: 	t.Run("length = 10", func(t *testing.T) {
276: 		t.Parallel()
277: 		cmo := testChandeMomentumOscillatorCreate(10)
278: 
279: 		check(-1, false, cmo.IsPrimed())
280: 
281: 		for i := 0; i < 10; i++ {
282: 			cmo.Update(input[i])
283: 			check(i, false, cmo.IsPrimed())
284: 		}
285: 
286: 		for i := 10; i < len(input); i++ {
287: 			cmo.Update(input[i])
288: 			check(i, true, cmo.IsPrimed())
289: 		}
290: 	})
291: }
292: 
293: func TestChandeMomentumOscillatorMetadata(t *testing.T) {
294: 	t.Parallel()
295: 
296: 	cmo := testChandeMomentumOscillatorCreate(5)
297: 	act := cmo.Metadata()
298: 
299: 	check := func(what string, exp, act any) {
300: 		t.Helper()
301: 
302: 		if exp != act {
303: 			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
304: 		}
305: 	}
306: 
307: 	check("Identifier", core.ChandeMomentumOscillator, act.Identifier)
308: 	check("Mnemonic", "cmo(5)", act.Mnemonic)
309: 	check("Description", "Chande Momentum Oscillator cmo(5)", act.Description)
310: 	check("len(Outputs)", 1, len(act.Outputs))
311: 	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
312: 	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
313: 	check("Outputs[0].Mnemonic", "cmo(5)", act.Outputs[0].Mnemonic)
314: 	check("Outputs[0].Description", "Chande Momentum Oscillator cmo(5)", act.Outputs[0].Description)
315: }
316: 
317: func TestNewChandeMomentumOscillator(t *testing.T) { //nolint: funlen
318: 	t.Parallel()
319: 
320: 	const (
321: 		bc     entities.BarComponent   = entities.BarMedianPrice
322: 		qc     entities.QuoteComponent = entities.QuoteMidPrice
323: 		tc     entities.TradeComponent = entities.TradePrice
324: 		length                         = 5
325: 		errlen                         = "invalid Chande momentum oscillator parameters: length should be positive"
326: 		errbc                          = "invalid Chande momentum oscillator parameters: 9999: unknown bar component"
327: 		errqc                          = "invalid Chande momentum oscillator parameters: 9999: unknown quote component"
328: 		errtc                          = "invalid Chande momentum oscillator parameters: 9999: unknown trade component"
329: 	)
330: 
331: 	check := func(name string, exp, act any) {
332: 		t.Helper()
333: 
334: 		if exp != act {
335: 			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
336: 		}
337: 	}
338: 
339: 	t.Run("length > 0", func(t *testing.T) {
340: 		t.Parallel()
341: 		params := Params{
342: 			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
343: 		}
344: 
345: 		cmo, err := New(&params)
346: 		check("err == nil", true, err == nil)
347: 		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
348: 		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
349: 		check("primed", false, cmo.primed)
350: 		check("length", length, cmo.length)
351: 		check("len(ringBuffer)", length, len(cmo.ringBuffer))
352: 		check("ringHead", 0, cmo.ringHead)
353: 		check("count", 0, cmo.count)
354: 		check("previousSample", 0., cmo.previousSample)
355: 		check("gainSum", 0., cmo.gainSum)
356: 		check("lossSum", 0., cmo.lossSum)
357: 	})
358: 
359: 	t.Run("length = 1", func(t *testing.T) {
360: 		t.Parallel()
361: 		params := Params{
362: 			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
363: 		}
364: 
365: 		cmo, err := New(&params)
366: 		check("err == nil", true, err == nil)
367: 		check("mnemonic", "cmo(1, hl/2)", cmo.LineIndicator.Mnemonic)
368: 		check("description", "Chande Momentum Oscillator cmo(1, hl/2)", cmo.LineIndicator.Description)
369: 		check("primed", false, cmo.primed)
370: 		check("length", 1, cmo.length)
371: 		check("len(ringBuffer)", 1, len(cmo.ringBuffer))
372: 		check("ringHead", 0, cmo.ringHead)
373: 		check("count", 0, cmo.count)
374: 		check("previousSample", 0., cmo.previousSample)
375: 		check("gainSum", 0., cmo.gainSum)
376: 		check("lossSum", 0., cmo.lossSum)
377: 	})
378: 
379: 	t.Run("length = 0", func(t *testing.T) {
380: 		t.Parallel()
381: 		params := Params{
382: 			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
383: 		}
384: 
385: 		cmo, err := New(&params)
386: 		check("cmo == nil", true, cmo == nil)
387: 		check("err", errlen, err.Error())
388: 	})
389: 
390: 	t.Run("length < 0", func(t *testing.T) {
391: 		t.Parallel()
392: 		params := Params{
393: 			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
394: 		}
395: 
396: 		cmo, err := New(&params)
397: 		check("cmo == nil", true, cmo == nil)
398: 		check("err", errlen, err.Error())
399: 	})
400: 
401: 	t.Run("invalid bar component", func(t *testing.T) {
402: 		t.Parallel()
403: 		params := Params{
404: 			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
405: 		}
406: 
407: 		cmo, err := New(&params)
408: 		check("cmo == nil", true, cmo == nil)
409: 		check("err", errbc, err.Error())
410: 	})
411: 
412: 	t.Run("invalid quote component", func(t *testing.T) {
413: 		t.Parallel()
414: 		params := Params{
415: 			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
416: 		}
417: 
418: 		cmo, err := New(&params)
419: 		check("cmo == nil", true, cmo == nil)
420: 		check("err", errqc, err.Error())
421: 	})
422: 
423: 	t.Run("invalid trade component", func(t *testing.T) {
424: 		t.Parallel()
425: 		params := Params{
426: 			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
427: 		}
428: 
429: 		cmo, err := New(&params)
430: 		check("cmo == nil", true, cmo == nil)
431: 		check("err", errtc, err.Error())
432: 	})
433: 
434: 	// Zero-value component mnemonic tests.
435: 	// A zero value means "use default, don't show in mnemonic".
436: 
437: 	t.Run("all components zero", func(t *testing.T) {
438: 		t.Parallel()
439: 		params := Params{Length: length}
440: 
441: 		cmo, err := New(&params)
442: 		check("err == nil", true, err == nil)
443: 		check("mnemonic", "cmo(5)", cmo.LineIndicator.Mnemonic)
444: 		check("description", "Chande Momentum Oscillator cmo(5)", cmo.LineIndicator.Description)
445: 	})
446: 
447: 	t.Run("only bar component set", func(t *testing.T) {
448: 		t.Parallel()
449: 		params := Params{Length: length, BarComponent: entities.BarMedianPrice}
450: 
451: 		cmo, err := New(&params)
452: 		check("err == nil", true, err == nil)
453: 		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
454: 		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
455: 	})
456: 
457: 	t.Run("only quote component set", func(t *testing.T) {
458: 		t.Parallel()
459: 		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}
460: 
461: 		cmo, err := New(&params)
462: 		check("err == nil", true, err == nil)
463: 		check("mnemonic", "cmo(5, b)", cmo.LineIndicator.Mnemonic)
464: 		check("description", "Chande Momentum Oscillator cmo(5, b)", cmo.LineIndicator.Description)
465: 	})
466: 
467: 	t.Run("only trade component set", func(t *testing.T) {
468: 		t.Parallel()
469: 		params := Params{Length: length, TradeComponent: entities.TradeVolume}
470: 
471: 		cmo, err := New(&params)
472: 		check("err == nil", true, err == nil)
473: 		check("mnemonic", "cmo(5, v)", cmo.LineIndicator.Mnemonic)
474: 		check("description", "Chande Momentum Oscillator cmo(5, v)", cmo.LineIndicator.Description)
475: 	})
476: }
477: 
478: func testChandeMomentumOscillatorCreate(length int) *ChandeMomentumOscillator {
479: 	params := Params{Length: length}
480: 
481: 	cmo, _ := New(&params)
482: 
483: 	return cmo
484: }

(End of file - total 484 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/stochasticrelativestrengthindex/stochasticrelativestrengthindex.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/stochasticrelativestrengthindex/stochasticrelativestrengthindex.go</path>
<type>file</type>
<content>1: package stochasticrelativestrengthindex
2: 
3: import (
4: 	"fmt"
5: 	"math"
6: 	"sync"
7: 
8: 	"zpano/entities"
9: 	"zpano/indicators/common/exponentialmovingaverage"
10: 	"zpano/indicators/common/simplemovingaverage"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/welleswilder/relativestrengthindex"
13: )
14: 
15: // lineUpdater is an interface for indicators that accept a single scalar and return a value.
16: type lineUpdater interface {
17: 	Update(float64) float64
18: 	IsPrimed() bool
19: }
20: 
21: // passthrough is a no-op smoother for FastD period of 1.
22: type passthrough struct{}
23: 
24: // Update returns v unchanged (no smoothing).
25: func (p *passthrough) Update(v float64) float64 { return v }
26: 
27: // IsPrimed always reports true; a passthrough has no warmup.
28: func (p *passthrough) IsPrimed() bool { return true }
29: 
30: // StochasticRelativeStrengthIndex is Tushar Chande's Stochastic RSI.
31: //
32: // Stochastic RSI applies the Stochastic oscillator formula to RSI values
33: // instead of price data. It oscillates between 0 and 100.
34: //
35: // The indicator first computes RSI, then applies a stochastic calculation
36: // over a rolling window of RSI values to produce Fast-K. Fast-D is a
37: // moving average of Fast-K.
38: //
39: // Reference:
40: //
41: // Chande, Tushar S. and Kroll, Stanley (1993). "Stochastic RSI and Dynamic
42: // Momentum Index". Stock & Commodities V.11:5 (189-199).
43: type StochasticRelativeStrengthIndex struct {
44: 	mu sync.RWMutex
45: 
46: 	rsi *relativestrengthindex.RelativeStrengthIndex
47: 
48: 	// Circular buffer for RSI values (size = fastKLength).
49: 	rsiBuf         []float64
50: 	rsiBufferIndex int
51: 	rsiCount       int
52: 
53: 	fastKLength int
54: 	fastDMA     lineUpdater
55: 
56: 	fastK  float64
57: 	fastD  float64
58: 	primed bool
59: 
60: 	barFunc   entities.BarFunc
61: 	quoteFunc entities.QuoteFunc
62: 	tradeFunc entities.TradeFunc
63: 	mnemonic  string
64: }
65: 
66: // NewStochasticRelativeStrengthIndex returns an instance of the indicator created using supplied parameters.
67: func NewStochasticRelativeStrengthIndex(p *StochasticRelativeStrengthIndexParams) (*StochasticRelativeStrengthIndex, error) {
68: 	const (
69: 		invalid       = "invalid stochastic relative strength index parameters"
70: 		fmts          = "%s: %s"
71: 		fmtw          = "%s: %w"
72: 		minRSILength  = 2
73: 		minFastLength = 1
74: 	)
75: 
76: 	if p.Length < minRSILength {
77: 		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
78: 	}
79: 
80: 	if p.FastKLength < minFastLength {
81: 		return nil, fmt.Errorf(fmts, invalid, "fast K length should be greater than 0")
82: 	}
83: 
84: 	if p.FastDLength < minFastLength {
85: 		return nil, fmt.Errorf(fmts, invalid, "fast D length should be greater than 0")
86: 	}
87: 
88: 	bc := p.BarComponent
89: 	if bc == 0 {
90: 		bc = entities.DefaultBarComponent
91: 	}
92: 
93: 	qc := p.QuoteComponent
94: 	if qc == 0 {
95: 		qc = entities.DefaultQuoteComponent
96: 	}
97: 
98: 	tc := p.TradeComponent
99: 	if tc == 0 {
100: 		tc = entities.DefaultTradeComponent
101: 	}
102: 
103: 	var (
104: 		err       error
105: 		barFunc   entities.BarFunc
106: 		quoteFunc entities.QuoteFunc
107: 		tradeFunc entities.TradeFunc
108: 	)
109: 
110: 	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
111: 		return nil, fmt.Errorf(fmtw, invalid, err)
112: 	}
113: 
114: 	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
115: 		return nil, fmt.Errorf(fmtw, invalid, err)
116: 	}
117: 
118: 	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
119: 		return nil, fmt.Errorf(fmtw, invalid, err)
120: 	}
121: 
122: 	// Create internal RSI.
123: 	rsi, rsiErr := relativestrengthindex.NewRelativeStrengthIndex(&relativestrengthindex.RelativeStrengthIndexParams{
124: 		Length: p.Length,
125: 	})
126: 	if rsiErr != nil {
127: 		return nil, fmt.Errorf(fmtw, invalid, rsiErr)
128: 	}
129: 
130: 	// Create Fast-D smoother.
131: 	var fastDMA lineUpdater
132: 
133: 	var maLabel string
134: 
135: 	if p.FastDLength < 2 { //nolint:mnd
136: 		fastDMA = &passthrough{}
137: 		maLabel = "SMA"
138: 	} else {
139: 		switch p.MovingAverageType {
140: 		case EMA:
141: 			maLabel = "EMA"
142: 
143: 			ema, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
144: 				&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
145: 					Length:         p.FastDLength,
146: 					FirstIsAverage: p.FirstIsAverage,
147: 				})
148: 			if e != nil {
149: 				return nil, fmt.Errorf(fmtw, invalid, e)
150: 			}
151: 
152: 			fastDMA = ema
153: 		default:
154: 			maLabel = "SMA"
155: 
156: 			sma, e := simplemovingaverage.NewSimpleMovingAverage(
157: 				&simplemovingaverage.SimpleMovingAverageParams{Length: p.FastDLength})
158: 			if e != nil {
159: 				return nil, fmt.Errorf(fmtw, invalid, e)
160: 			}
161: 
162: 			fastDMA = sma
163: 		}
164: 	}
165: 
166: 	mnemonic := fmt.Sprintf("stochrsi(%d/%d/%s%d%s)", p.Length, p.FastKLength,
167: 		maLabel, p.FastDLength, core.ComponentTripleMnemonic(bc, qc, tc))
168: 
169: 	return &StochasticRelativeStrengthIndex{
170: 		rsi:         rsi,
171: 		rsiBuf:      make([]float64, p.FastKLength),
172: 		fastKLength: p.FastKLength,
173: 		fastDMA:     fastDMA,
174: 		fastK:       math.NaN(),
175: 		fastD:       math.NaN(),
176: 		barFunc:     barFunc,
177: 		quoteFunc:   quoteFunc,
178: 		tradeFunc:   tradeFunc,
179: 		mnemonic:    mnemonic,
180: 	}, nil
181: }
182: 
183: // IsPrimed indicates whether the indicator is primed.
184: func (s *StochasticRelativeStrengthIndex) IsPrimed() bool {
185: 	s.mu.RLock()
186: 	defer s.mu.RUnlock()
187: 
188: 	return s.primed
189: }
190: 
191: // Metadata describes the output data of the indicator.
192: func (s *StochasticRelativeStrengthIndex) Metadata() core.Metadata {
193: 	desc := "Stochastic Relative Strength Index " + s.mnemonic
194: 
195: 	return core.BuildMetadata(
196: 		core.StochasticRelativeStrengthIndex,
197: 		s.mnemonic,
198: 		desc,
199: 		[]core.OutputText{
200: 			{Mnemonic: s.mnemonic + " fastK", Description: desc + " Fast-K"},
201: 			{Mnemonic: s.mnemonic + " fastD", Description: desc + " Fast-D"},
202: 		},
203: 	)
204: }
205: 
206: // Update updates the indicator given the next sample and returns both FastK and FastD values.
207: func (s *StochasticRelativeStrengthIndex) Update(sample float64) (float64, float64) {
208: 	if math.IsNaN(sample) {
209: 		return math.NaN(), math.NaN()
210: 	}
211: 
212: 	s.mu.Lock()
213: 	defer s.mu.Unlock()
214: 
215: 	// Feed to internal RSI.
216: 	rsiValue := s.rsi.Update(sample)
217: 	if math.IsNaN(rsiValue) {
218: 		return s.fastK, s.fastD
219: 	}
220: 
221: 	// Store RSI value in circular buffer.
222: 	s.rsiBuf[s.rsiBufferIndex] = rsiValue
223: 	s.rsiBufferIndex = (s.rsiBufferIndex + 1) % s.fastKLength
224: 	s.rsiCount++
225: 
226: 	// Need at least fastKLength RSI values for stochastic calculation.
227: 	if s.rsiCount < s.fastKLength {
228: 		return s.fastK, s.fastD
229: 	}
230: 
231: 	// Find min and max of RSI values in the window.
232: 	minRSI := s.rsiBuf[0]
233: 	maxRSI := s.rsiBuf[0]
234: 
235: 	for i := 1; i < s.fastKLength; i++ {
236: 		if s.rsiBuf[i] < minRSI {
237: 			minRSI = s.rsiBuf[i]
238: 		}
239: 
240: 		if s.rsiBuf[i] > maxRSI {
241: 			maxRSI = s.rsiBuf[i]
242: 		}
243: 	}
244: 
245: 	// Calculate Fast-K.
246: 	diff := maxRSI - minRSI
247: 	if diff > 0 {
248: 		s.fastK = 100 * (rsiValue - minRSI) / diff //nolint:mnd
249: 	} else {
250: 		s.fastK = 0
251: 	}
252: 
253: 	// Feed Fast-K to Fast-D smoother.
254: 	s.fastD = s.fastDMA.Update(s.fastK)
255: 
256: 	if !s.primed && s.fastDMA.IsPrimed() {
257: 		s.primed = true
258: 	}
259: 
260: 	return s.fastK, s.fastD
261: }
262: 
263: // UpdateScalar updates the indicator given the next scalar sample.
264: func (s *StochasticRelativeStrengthIndex) UpdateScalar(sample *entities.Scalar) core.Output {
265: 	fastK, fastD := s.Update(sample.Value)
266: 
267: 	output := make([]any, 2) //nolint:mnd
268: 	output[0] = entities.Scalar{Time: sample.Time, Value: fastK}
269: 	output[1] = entities.Scalar{Time: sample.Time, Value: fastD}
270: 
271: 	return output
272: }
273: 
274: // UpdateBar updates the indicator given the next bar sample.
275: func (s *StochasticRelativeStrengthIndex) UpdateBar(sample *entities.Bar) core.Output {
276: 	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: s.barFunc(sample)})
277: }
278: 
279: // UpdateQuote updates the indicator given the next quote sample.
280: func (s *StochasticRelativeStrengthIndex) UpdateQuote(sample *entities.Quote) core.Output {
281: 	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: s.quoteFunc(sample)})
282: }
283: 
284: // UpdateTrade updates the indicator given the next trade sample.
285: func (s *StochasticRelativeStrengthIndex) UpdateTrade(sample *entities.Trade) core.Output {
286: 	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: s.tradeFunc(sample)})
287: }

(End of file - total 287 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/stochasticrelativestrengthindex/params.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/stochasticrelativestrengthindex/params.go</path>
<type>file</type>
<content>1: package stochasticrelativestrengthindex
2: 
3: import "zpano/entities"
4: 
5: // MovingAverageType specifies the type of moving average to use for the Fast-D smoothing.
6: type MovingAverageType int
7: 
8: const (
9: 	// SMA uses a Simple Moving Average.
10: 	SMA MovingAverageType = iota
11: 
12: 	// EMA uses an Exponential Moving Average.
13: 	EMA
14: )
15: 
16: // StochasticRelativeStrengthIndexParams describes parameters to create an instance of the indicator.
17: type StochasticRelativeStrengthIndexParams struct {
18: 	// Length is the number of periods for the RSI calculation.
19: 	//
20: 	// The value should be greater than 1. The default value is 14.
21: 	Length int
22: 
23: 	// FastKLength is the number of periods for the Fast-K stochastic calculation.
24: 	//
25: 	// The value should be greater than 0. The default value is 5.
26: 	FastKLength int
27: 
28: 	// FastDLength is the number of periods for the Fast-D smoothing.
29: 	//
30: 	// The value should be greater than 0. The default value is 3.
31: 	FastDLength int
32: 
33: 	// MovingAverageType specifies the type of moving average for Fast-D (SMA or EMA).
34: 	//
35: 	// If zero (SMA), the Simple Moving Average is used.
36: 	MovingAverageType MovingAverageType
37: 
38: 	// FirstIsAverage controls the EMA seeding algorithm.
39: 	// When true, the first EMA value is the simple average of the first period values.
40: 	// When false (default), the first input value is used directly (Metastock style).
41: 	// Only relevant when MovingAverageType is EMA.
42: 	FirstIsAverage bool
43: 
44: 	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
45: 	//
46: 	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
47: 	BarComponent entities.BarComponent
48: 
49: 	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
50: 	//
51: 	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
52: 	QuoteComponent entities.QuoteComponent
53: 
54: 	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
55: 	//
56: 	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
57: 	TradeComponent entities.TradeComponent
58: }
59: 
60: // DefaultParams returns a [StochasticRelativeStrengthIndexParams] value populated with conventional defaults.
61: func DefaultParams() *StochasticRelativeStrengthIndexParams {
62: 	return &StochasticRelativeStrengthIndexParams{
63: 		Length:      14,
64: 		FastKLength: 5,
65: 		FastDLength: 3,
66: 	}
67: }

(End of file - total 67 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/stochasticrelativestrengthindex/stochasticrelativestrengthindex_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/stochasticrelativestrengthindex/stochasticrelativestrengthindex_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package stochasticrelativestrengthindex
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: // Test data from TA-Lib (252 entries).
16: func testInput() []float64 {
17: 	return []float64{
18: 		91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
19: 		97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000, 88.375000, 87.625000,
20: 		84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000, 85.250000, 87.125000, 85.815000,
21: 		88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000, 83.375000, 85.500000, 89.190000, 89.440000,
22: 		91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000, 89.030000, 88.815000, 84.280000, 83.500000, 82.690000,
23: 		84.750000, 85.655000, 86.190000, 88.940000, 89.280000, 88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000,
24: 		93.155000, 91.720000, 90.000000, 89.690000, 88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000,
25: 		104.940000, 106.000000, 102.500000, 102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000,
26: 		109.315000, 110.500000, 112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000,
27: 		111.875000, 110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
28: 		116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000,
29: 		123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000,
30: 		132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000,
31: 		129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000,
32: 		118.500000, 123.190000, 123.500000, 122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000,
33: 		123.870000, 122.940000, 121.750000, 124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000,
34: 		127.250000, 125.870000, 128.860000, 132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000,
35: 		130.000000, 125.370000, 130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000,
36: 		121.000000, 117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
37: 		107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
38: 		94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000, 95.000000,
39: 		95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000, 105.000000,
40: 		104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000,
41: 		109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000,
42: 		109.810000, 109.000000, 108.750000, 107.870000,
43: 	}
44: }
45: 
46: func testTime() time.Time {
47: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
48: }
49: 
50: // Test case 1: period=14, fastK=14, fastD=1, SMA.
51: // begIndex=27, first value: FastK=94.156709, FastD=94.156709.
52: // last value (index 251): FastK=0.0, FastD=0.0.
53: func TestStochasticRSI_14_14_1_SMA(t *testing.T) {
54: 	t.Parallel()
55: 
56: 	const tolerance = 1e-4
57: 
58: 	input := testInput()
59: 
60: 	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
61: 		Length:      14,
62: 		FastKLength: 14,
63: 		FastDLength: 1,
64: 	})
65: 	if err != nil {
66: 		t.Fatal(err)
67: 	}
68: 
69: 	// First 27 values should produce NaN for FastK.
70: 	for i := 0; i < 27; i++ {
71: 		fastK, _ := ind.Update(input[i])
72: 		if !math.IsNaN(fastK) {
73: 			t.Errorf("[%d] expected NaN FastK, got %v", i, fastK)
74: 		}
75: 	}
76: 
77: 	// Index 27: first value.
78: 	fastK, fastD := ind.Update(input[27])
79: 	if math.IsNaN(fastK) {
80: 		t.Errorf("[27] expected non-NaN FastK, got NaN")
81: 	}
82: 
83: 	if math.Abs(fastK-94.156709) > tolerance {
84: 		t.Errorf("[27] FastK: expected ~94.156709, got %v", fastK)
85: 	}
86: 
87: 	if math.Abs(fastD-94.156709) > tolerance {
88: 		t.Errorf("[27] FastD: expected ~94.156709, got %v", fastD)
89: 	}
90: 
91: 	// Feed remaining and check last value.
92: 	for i := 28; i < 251; i++ {
93: 		ind.Update(input[i])
94: 	}
95: 
96: 	fastK, fastD = ind.Update(input[251])
97: 
98: 	if math.Abs(fastK-0.0) > tolerance {
99: 		t.Errorf("[251] FastK: expected ~0.0, got %v", fastK)
100: 	}
101: 
102: 	if math.Abs(fastD-0.0) > tolerance {
103: 		t.Errorf("[251] FastD: expected ~0.0, got %v", fastD)
104: 	}
105: }
106: 
107: // Test case 2: period=14, fastK=45, fastD=1, SMA.
108: // begIndex=58, first value: FastK=79.729186, FastD=79.729186.
109: // last value (index 251): FastK=48.1550743, FastD=48.1550743.
110: func TestStochasticRSI_14_45_1_SMA(t *testing.T) {
111: 	t.Parallel()
112: 
113: 	const tolerance = 1e-4
114: 
115: 	input := testInput()
116: 
117: 	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
118: 		Length:      14,
119: 		FastKLength: 45,
120: 		FastDLength: 1,
121: 	})
122: 	if err != nil {
123: 		t.Fatal(err)
124: 	}
125: 
126: 	// First 58 values should produce NaN for FastK.
127: 	for i := 0; i < 58; i++ {
128: 		fastK, _ := ind.Update(input[i])
129: 		if !math.IsNaN(fastK) {
130: 			t.Errorf("[%d] expected NaN FastK, got %v", i, fastK)
131: 		}
132: 	}
133: 
134: 	// Index 58: first value.
135: 	fastK, fastD := ind.Update(input[58])
136: 	if math.IsNaN(fastK) {
137: 		t.Errorf("[58] expected non-NaN FastK, got NaN")
138: 	}
139: 
140: 	if math.Abs(fastK-79.729186) > tolerance {
141: 		t.Errorf("[58] FastK: expected ~79.729186, got %v", fastK)
142: 	}
143: 
144: 	if math.Abs(fastD-79.729186) > tolerance {
145: 		t.Errorf("[58] FastD: expected ~79.729186, got %v", fastD)
146: 	}
147: 
148: 	// Feed remaining and check last value.
149: 	for i := 59; i < 251; i++ {
150: 		ind.Update(input[i])
151: 	}
152: 
153: 	fastK, fastD = ind.Update(input[251])
154: 
155: 	if math.Abs(fastK-48.1550743) > tolerance {
156: 		t.Errorf("[251] FastK: expected ~48.1550743, got %v", fastK)
157: 	}
158: 
159: 	if math.Abs(fastD-48.1550743) > tolerance {
160: 		t.Errorf("[251] FastD: expected ~48.1550743, got %v", fastD)
161: 	}
162: }
163: 
164: // Test case 3: period=11, fastK=13, fastD=16, SMA.
165: // begIndex=38, first value: FastK=5.25947, FastD=57.1711.
166: // last value (index 251): FastK=0.0, FastD=15.7303.
167: func TestStochasticRSI_11_13_16_SMA(t *testing.T) {
168: 	t.Parallel()
169: 
170: 	const tolerance = 1e-3
171: 
172: 	input := testInput()
173: 
174: 	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
175: 		Length:      11,
176: 		FastKLength: 13,
177: 		FastDLength: 16,
178: 	})
179: 	if err != nil {
180: 		t.Fatal(err)
181: 	}
182: 
183: 	// begIndex=38: RSI lookback(11)=11, STOCHF lookback(13, 16, SMA) = 12 + 15 = 27, total = 38.
184: 	// But FastD (SMA(16)) won't be primed until 16 FastK values are collected.
185: 	// First primed at: RSI primes after 11+1=12 inputs (indices 0..11, first RSI at index 11).
186: 	// Wait, RSI lookback = 11 means first RSI at index 11.
187: 	// FastK needs 13 RSI values: first FastK at index 11+12 = 23.
188: 	// FastD SMA(16) needs 16 FastK values: first FastD at index 23+15 = 38.
189: 	// So begIndex=38: FastD is first valid here.
190: 
191: 	// Feed first 38 values.
192: 	for i := 0; i < 38; i++ {
193: 		ind.Update(input[i])
194: 	}
195: 
196: 	// Index 38: first primed value.
197: 	fastK, fastD := ind.Update(input[38])
198: 
199: 	if math.Abs(fastK-5.25947) > tolerance {
200: 		t.Errorf("[38] FastK: expected ~5.25947, got %v", fastK)
201: 	}
202: 
203: 	if math.Abs(fastD-57.1711) > tolerance {
204: 		t.Errorf("[38] FastD: expected ~57.1711, got %v", fastD)
205: 	}
206: 
207: 	if !ind.IsPrimed() {
208: 		t.Error("expected primed at index 38")
209: 	}
210: 
211: 	// Feed remaining and check last value.
212: 	for i := 39; i < 251; i++ {
213: 		ind.Update(input[i])
214: 	}
215: 
216: 	fastK, fastD = ind.Update(input[251])
217: 
218: 	if math.Abs(fastK-0.0) > tolerance {
219: 		t.Errorf("[251] FastK: expected ~0.0, got %v", fastK)
220: 	}
221: 
222: 	if math.Abs(fastD-15.7303) > tolerance {
223: 		t.Errorf("[251] FastD: expected ~15.7303, got %v", fastD)
224: 	}
225: }
226: 
227: func TestStochasticRSIIsPrimed(t *testing.T) {
228: 	t.Parallel()
229: 
230: 	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
231: 		Length:      14,
232: 		FastKLength: 14,
233: 		FastDLength: 1,
234: 	})
235: 	if err != nil {
236: 		t.Fatal(err)
237: 	}
238: 
239: 	input := testInput()
240: 
241: 	if ind.IsPrimed() {
242: 		t.Error("expected not primed initially")
243: 	}
244: 
245: 	for i := 0; i < 27; i++ {
246: 		ind.Update(input[i])
247: 		if ind.IsPrimed() {
248: 			t.Errorf("[%d] expected not primed", i)
249: 		}
250: 	}
251: 
252: 	ind.Update(input[27])
253: 
254: 	if !ind.IsPrimed() {
255: 		t.Error("expected primed after index 27")
256: 	}
257: }
258: 
259: func TestStochasticRSINaN(t *testing.T) {
260: 	t.Parallel()
261: 
262: 	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
263: 		Length:      14,
264: 		FastKLength: 14,
265: 		FastDLength: 1,
266: 	})
267: 	if err != nil {
268: 		t.Fatal(err)
269: 	}
270: 
271: 	fastK, fastD := ind.Update(math.NaN())
272: 	if !math.IsNaN(fastK) {
273: 		t.Errorf("expected NaN FastK, got %v", fastK)
274: 	}
275: 
276: 	if !math.IsNaN(fastD) {
277: 		t.Errorf("expected NaN FastD, got %v", fastD)
278: 	}
279: }
280: 
281: func TestStochasticRSIMetadata(t *testing.T) {
282: 	t.Parallel()
283: 
284: 	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
285: 		Length:      14,
286: 		FastKLength: 14,
287: 		FastDLength: 3,
288: 	})
289: 	if err != nil {
290: 		t.Fatal(err)
291: 	}
292: 
293: 	meta := ind.Metadata()
294: 
295: 	if meta.Identifier != core.StochasticRelativeStrengthIndex {
296: 		t.Errorf("expected identifier StochasticRelativeStrengthIndex, got %v", meta.Identifier)
297: 	}
298: 
299: 	exp := "stochrsi(14/14/SMA3)"
300: 	if meta.Mnemonic != exp {
301: 		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
302: 	}
303: 
304: 	if len(meta.Outputs) != 2 {
305: 		t.Fatalf("expected 2 outputs, got %d", len(meta.Outputs))
306: 	}
307: 
308: 	if meta.Outputs[0].Kind != int(FastK) {
309: 		t.Errorf("expected output 0 kind %d, got %d", FastK, meta.Outputs[0].Kind)
310: 	}
311: 
312: 	if meta.Outputs[0].Shape != shape.Scalar {
313: 		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
314: 	}
315: 
316: 	if meta.Outputs[1].Kind != int(FastD) {
317: 		t.Errorf("expected output 1 kind %d, got %d", FastD, meta.Outputs[1].Kind)
318: 	}
319: }
320: 
321: func TestStochasticRSIUpdateEntity(t *testing.T) {
322: 	t.Parallel()
323: 
324: 	const tolerance = 1e-4
325: 
326: 	input := testInput()
327: 
328: 	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
329: 		Length:      14,
330: 		FastKLength: 14,
331: 		FastDLength: 1,
332: 	})
333: 	if err != nil {
334: 		t.Fatal(err)
335: 	}
336: 
337: 	tm := testTime()
338: 
339: 	for i := 0; i < 27; i++ {
340: 		scalar := &entities.Scalar{Time: tm, Value: input[i]}
341: 		out := ind.UpdateScalar(scalar)
342: 
343: 		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
344: 		if !math.IsNaN(v) {
345: 			t.Errorf("[%d] expected NaN, got %v", i, v)
346: 		}
347: 	}
348: 
349: 	scalar := &entities.Scalar{Time: tm, Value: input[27]}
350: 	out := ind.UpdateScalar(scalar)
351: 
352: 	fastK := out[0].(entities.Scalar).Value //nolint:forcetypeassert
353: 	fastD := out[1].(entities.Scalar).Value //nolint:forcetypeassert
354: 
355: 	if math.Abs(fastK-94.156709) > tolerance {
356: 		t.Errorf("[27] FastK: expected ~94.156709, got %v", fastK)
357: 	}
358: 
359: 	if math.Abs(fastD-94.156709) > tolerance {
360: 		t.Errorf("[27] FastD: expected ~94.156709, got %v", fastD)
361: 	}
362: }
363: 
364: func TestStochasticRSIInvalidParams(t *testing.T) {
365: 	t.Parallel()
366: 
367: 	tests := []struct {
368: 		name                             string
369: 		length, fastKLength, fastDLength int
370: 	}{
371: 		{"length too small", 1, 14, 3},
372: 		{"fastK too small", 14, 0, 3},
373: 		{"fastD too small", 14, 14, 0},
374: 		{"length negative", -1, 14, 3},
375: 	}
376: 
377: 	for _, tt := range tests {
378: 		_, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
379: 			Length:      tt.length,
380: 			FastKLength: tt.fastKLength,
381: 			FastDLength: tt.fastDLength,
382: 		})
383: 		if err == nil {
384: 			t.Errorf("%s: expected error, got nil", tt.name)
385: 		}
386: 	}
387: }

(End of file - total 387 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 7.7s)

Now I have all the reference implementations. Note that StochasticRSI depends on RelativeStrengthIndex (welleswilder). Let me check if RSI is already implemented in Zig, and also look at a LineIndicator-based example to model CMO.

**Tool: bash**

**Input:**
```json
{
  "command": "ls zig/src/indicators/welles_wilder/ 2>/dev/null || echo \"NOT YET\"",
  "description": "Check if welleswilder exists in zig"
}
```

**Output:**
```
NOT YET

```

---

## Assistant (Build · claude-opus-4.6 · 11.1s)

RSI is not yet ported. So I should do Aroon and CMO now, and defer StochRSI until after welleswilder (which includes RSI). Let me start writing the Zig implementations.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "marc_chaikin family: AdvanceDecline + AdvanceDeclineOscillator registered and passing (650 tests)",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "tusharchande family: Aroon",
      "priority": "high",
      "status": "in_progress"
    },
    {
      "content": "tusharchande family: ChandeMomentumOscillator",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "tusharchande family: StochasticRelativeStrengthIndex (blocked on welleswilder/RSI)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "welleswilder family (12 indicators including RSI)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "johnehlers family (23 indicators)",
      "priority": "medium",
      "status": "pending"
    },
    {
      "content": "custom family (GoertzelSpectrum, MaximumEntropySpectrum)",
      "priority": "medium",
      "status": "pending"
    },
    {
      "content": "Factory, frequency_response, cmd tools",
      "priority": "low",
      "status": "pending"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "marc_chaikin family: AdvanceDecline + AdvanceDeclineOscillator registered and passing (650 tests)",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "tusharchande family: Aroon",
    "priority": "high",
    "status": "in_progress"
  },
  {
    "content": "tusharchande family: ChandeMomentumOscillator",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "tusharchande family: StochasticRelativeStrengthIndex (blocked on welleswilder/RSI)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "welleswilder family (12 indicators including RSI)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "johnehlers family (23 indicators)",
    "priority": "medium",
    "status": "pending"
  },
  {
    "content": "custom family (GoertzelSpectrum, MaximumEntropySpectrum)",
    "priority": "medium",
    "status": "pending"
  },
  {
    "content": "Factory, frequency_response, cmd tools",
    "priority": "low",
    "status": "pending"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 7.4s)

Let me look at an existing HLCV-override indicator (like Aroon-style, not using LineIndicator) in Zig for the pattern — WilliamsPercentR or UltimateOscillator.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig",
  "limit": 80
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: 
9: const indicator_mod = @import("../../core/indicator.zig");
10: const build_metadata_mod = @import("../../core/build_metadata.zig");
11: const identifier_mod = @import("../../core/identifier.zig");
12: const metadata_mod = @import("../../core/metadata.zig");
13: 
14: const OutputArray = indicator_mod.OutputArray;
15: const Identifier = identifier_mod.Identifier;
16: const Metadata = metadata_mod.Metadata;
17: 
18: /// Enumerates the outputs of the Williams %R indicator.
19: pub const WilliamsPercentROutput = enum(u8) {
20:     /// The scalar value of the Williams %R.
21:     value = 1,
22: };
23: 
24: /// Parameters to create an instance of the Williams %R indicator.
25: pub const WilliamsPercentRParams = struct {
26:     /// The number of time periods. Must be >= 2. Default is 14.
27:     length: usize = 14,
28: };
29: 
30: /// Larry Williams' Williams %R momentum indicator.
31: ///
32: /// Williams %R reflects the level of the closing price relative to the
33: /// highest high over a lookback period. The oscillation ranges from 0 to -100;
34: /// readings from 0 to -20 are considered overbought, readings from -80 to -100
35: /// are considered oversold.
36: ///
37: /// The value is calculated as:
38: ///
39: ///   %R = -100 * (HighestHigh - Close) / (HighestHigh - LowestLow)
40: ///
41: /// where HighestHigh and LowestLow are computed over the last `length` bars.
42: /// If HighestHigh equals LowestLow, the value is 0.
43: ///
44: /// The indicator requires bar data (high, low, close). For scalar, quote, and
45: /// trade updates, the single value is used as a substitute for all three.
46: pub const WilliamsPercentR = struct {
47:     length: usize,
48:     high_buf: []f64,
49:     low_buf: []f64,
50:     buffer_index: usize,
51:     count: usize,
52:     value: f64,
53:     primed: bool,
54:     allocator: std.mem.Allocator,
55: 
56:     pub const Error = error{
57:         InvalidLength,
58:         OutOfMemory,
59:     };
60: 
61:     pub fn init(allocator: std.mem.Allocator, params: WilliamsPercentRParams) Error!WilliamsPercentR {
62:         const length = params.length;
63:         if (length < 2) return error.InvalidLength;
64: 
65:         const high_buf = allocator.alloc(f64, length) catch return error.OutOfMemory;
66:         errdefer allocator.free(high_buf);
67:         const low_buf = allocator.alloc(f64, length) catch return error.OutOfMemory;
68: 
69:         return WilliamsPercentR{
70:             .length = length,
71:             .high_buf = high_buf,
72:             .low_buf = low_buf,
73:             .buffer_index = 0,
74:             .count = 0,
75:             .value = math.nan(f64),
76:             .primed = false,
77:             .allocator = allocator,
78:         };
79:     }
80: 

(Showing lines 1-80 of 726. Use offset=81 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 5.0s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig",
  "offset": 80,
  "limit": 120
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig</path>
<type>file</type>
<content>80: 
81:     pub fn deinit(self: *WilliamsPercentR) void {
82:         self.allocator.free(self.high_buf);
83:         self.allocator.free(self.low_buf);
84:     }
85: 
86:     pub fn fixSlices(self: *WilliamsPercentR) void {
87:         _ = self;
88:     }
89: 
90:     /// Core update given close, high, low values.
91:     pub fn update(self: *WilliamsPercentR, close: f64, high: f64, low: f64) f64 {
92:         if (math.isNan(close) or math.isNan(high) or math.isNan(low)) {
93:             return math.nan(f64);
94:         }
95: 
96:         self.high_buf[self.buffer_index] = high;
97:         self.low_buf[self.buffer_index] = low;
98: 
99:         self.buffer_index = (self.buffer_index + 1) % self.length;
100:         self.count += 1;
101: 
102:         if (self.count < self.length) {
103:             return self.value;
104:         }
105: 
106:         // Find highest high and lowest low in the window.
107:         var max_high = self.high_buf[0];
108:         var min_low = self.low_buf[0];
109:         for (self.high_buf[1..], self.low_buf[1..]) |h, l| {
110:             if (h > max_high) max_high = h;
111:             if (l < min_low) min_low = l;
112:         }
113: 
114:         const diff = max_high - min_low;
115:         if (@abs(diff) < math.floatMin(f64)) {
116:             self.value = 0;
117:         } else {
118:             self.value = -100.0 * (max_high - close) / diff;
119:         }
120: 
121:         if (!self.primed) {
122:             self.primed = true;
123:         }
124: 
125:         return self.value;
126:     }
127: 
128:     /// Update using a single sample value as substitute for high, low, close.
129:     pub fn updateSample(self: *WilliamsPercentR, sample: f64) f64 {
130:         return self.update(sample, sample, sample);
131:     }
132: 
133:     pub fn isPrimed(self: *const WilliamsPercentR) bool {
134:         return self.primed;
135:     }
136: 
137:     pub fn getMetadata(self: *const WilliamsPercentR, out: *Metadata) void {
138:         _ = self;
139:         const mnemonic = "willr";
140:         const description = "Williams %R";
141: 
142:         build_metadata_mod.buildMetadata(out, Identifier.williams_percent_r, mnemonic, description, &.{
143:             .{ .mnemonic = mnemonic, .description = description },
144:         });
145:     }
146: 
147:     fn makeOutput(self: *const WilliamsPercentR, time: i64) OutputArray {
148:         var out = OutputArray{};
149:         out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
150:         return out;
151:     }
152: 
153:     pub fn updateScalar(self: *WilliamsPercentR, sample: *const Scalar) OutputArray {
154:         _ = self.update(sample.value, sample.value, sample.value);
155:         return self.makeOutput(sample.time);
156:     }
157: 
158:     pub fn updateBar(self: *WilliamsPercentR, sample: *const Bar) OutputArray {
159:         _ = self.update(sample.close, sample.high, sample.low);
160:         return self.makeOutput(sample.time);
161:     }
162: 
163:     pub fn updateQuote(self: *WilliamsPercentR, sample: *const Quote) OutputArray {
164:         const mid = (sample.bid_price + sample.ask_price) / 2.0;
165:         _ = self.update(mid, mid, mid);
166:         return self.makeOutput(sample.time);
167:     }
168: 
169:     pub fn updateTrade(self: *WilliamsPercentR, sample: *const Trade) OutputArray {
170:         _ = self.update(sample.price, sample.price, sample.price);
171:         return self.makeOutput(sample.time);
172:     }
173: 
174:     // --- Indicator interface ---
175: 
176:     pub fn indicator(self: *WilliamsPercentR) indicator_mod.Indicator {
177:         return indicator_mod.Indicator{
178:             .ptr = @ptrCast(self),
179:             .vtable = &.{
180:                 .isPrimed = vtableIsPrimed,
181:                 .metadata = vtableMetadata,
182:                 .updateScalar = vtableUpdateScalar,
183:                 .updateBar = vtableUpdateBar,
184:                 .updateQuote = vtableUpdateQuote,
185:                 .updateTrade = vtableUpdateTrade,
186:             },
187:         };
188:     }
189: 
190:     fn vtableIsPrimed(ptr: *const anyopaque) bool {
191:         const self: *const WilliamsPercentR = @ptrCast(@alignCast(ptr));
192:         return self.isPrimed();
193:     }
194: 
195:     fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
196:         const self: *const WilliamsPercentR = @ptrCast(@alignCast(ptr));
197:         self.getMetadata(out);
198:     }
199: 

(Showing lines 80-199 of 726. Use offset=200 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 5.9s)

Good, I have the pattern. Now let me also check how CMO-style LineIndicator indicators work in Zig (like Momentum or ROC):

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/common/momentum/momentum.zig",
  "limit": 100
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/common/momentum/momentum.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: const bar_component = @import("bar_component");
9: const quote_component = @import("quote_component");
10: const trade_component = @import("trade_component");
11: 
12: const indicator_mod = @import("../../core/indicator.zig");
13: const line_indicator_mod = @import("../../core/line_indicator.zig");
14: const build_metadata_mod = @import("../../core/build_metadata.zig");
15: const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
16: const identifier_mod = @import("../../core/identifier.zig");
17: const metadata_mod = @import("../../core/metadata.zig");
18: 
19: const OutputArray = indicator_mod.OutputArray;
20: const LineIndicator = line_indicator_mod.LineIndicator;
21: const Identifier = identifier_mod.Identifier;
22: const Metadata = metadata_mod.Metadata;
23: 
24: /// Enumerates the outputs of the momentum indicator.
25: pub const MomentumOutput = enum(u8) {
26:     /// The scalar value of the momentum.
27:     value = 1,
28: };
29: 
30: /// Parameters to create an instance of the momentum indicator.
31: pub const MomentumParams = struct {
32:     /// The length (number of time periods). Must be >= 1.
33:     length: usize,
34:     /// Bar component to extract. `null` means use default (Close).
35:     bar_component: ?bar_component.BarComponent = null,
36:     /// Quote component to extract. `null` means use default (Mid).
37:     quote_component: ?quote_component.QuoteComponent = null,
38:     /// Trade component to extract. `null` means use default (Price).
39:     trade_component: ?trade_component.TradeComponent = null,
40: };
41: 
42: /// Computes the momentum (MOM).
43: ///
44: /// MOMi = Pi - Pi-l
45: ///
46: /// where l is the length.
47: ///
48: /// The indicator is not primed during the first l updates.
49: pub const Momentum = struct {
50:     line: LineIndicator,
51:     window: []f64,
52:     window_length: usize,
53:     window_count: usize,
54:     last_index: usize,
55:     primed: bool,
56:     allocator: std.mem.Allocator,
57:     mnemonic_buf: [64]u8,
58:     mnemonic_len: usize,
59:     description_buf: [128]u8,
60:     description_len: usize,
61: 
62:     pub fn init(allocator: std.mem.Allocator, params: MomentumParams) !Momentum {
63:         if (params.length < 1) {
64:             return error.InvalidLength;
65:         }
66: 
67:         const bc = params.bar_component orelse bar_component.default_bar_component;
68:         const qc = params.quote_component orelse quote_component.default_quote_component;
69:         const tc = params.trade_component orelse trade_component.default_trade_component;
70: 
71:         var triple_buf: [64]u8 = undefined;
72:         const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);
73: 
74:         var mnemonic_buf: [64]u8 = undefined;
75:         const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "mom({d}{s})", .{ params.length, triple }) catch
76:             return error.MnemonicTooLong;
77:         const mnemonic_len = mnemonic_slice.len;
78: 
79:         var description_buf: [128]u8 = undefined;
80:         const desc_slice = std.fmt.bufPrint(&description_buf, "Momentum {s}", .{mnemonic_slice}) catch
81:             return error.MnemonicTooLong;
82:         const description_len = desc_slice.len;
83: 
84:         const window_length = params.length + 1;
85:         const window = try allocator.alloc(f64, window_length);
86:         @memset(window, 0.0);
87: 
88:         return .{
89:             .line = LineIndicator.new(
90:                 mnemonic_buf[0..mnemonic_len],
91:                 description_buf[0..description_len],
92:                 params.bar_component,
93:                 params.quote_component,
94:                 params.trade_component,
95:             ),
96:             .window = window,
97:             .window_length = window_length,
98:             .window_count = 0,
99:             .last_index = params.length,
100:             .primed = false,

(Showing lines 1-100 of 430. Use offset=101 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 5.1s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/common/momentum/momentum.zig",
  "offset": 100,
  "limit": 100
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/common/momentum/momentum.zig</path>
<type>file</type>
<content>100:             .primed = false,
101:             .allocator = allocator,
102:             .mnemonic_buf = mnemonic_buf,
103:             .mnemonic_len = mnemonic_len,
104:             .description_buf = description_buf,
105:             .description_len = description_len,
106:         };
107:     }
108: 
109:     pub fn deinit(self: *Momentum) void {
110:         self.allocator.free(self.window);
111:     }
112: 
113:     fn fixSlices(self: *Momentum) void {
114:         self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
115:         self.line.description = self.description_buf[0..self.description_len];
116:     }
117: 
118:     /// Core update logic. Returns the momentum value or NaN if not yet primed.
119:     pub fn update(self: *Momentum, sample: f64) f64 {
120:         if (math.isNan(sample)) {
121:             return sample;
122:         }
123: 
124:         if (self.primed) {
125:             var i: usize = 0;
126:             while (i < self.last_index) : (i += 1) {
127:                 self.window[i] = self.window[i + 1];
128:             }
129: 
130:             self.window[self.last_index] = sample;
131: 
132:             return sample - self.window[0];
133:         }
134: 
135:         self.window[self.window_count] = sample;
136:         self.window_count += 1;
137: 
138:         if (self.window_length == self.window_count) {
139:             self.primed = true;
140: 
141:             return sample - self.window[0];
142:         }
143: 
144:         return math.nan(f64);
145:     }
146: 
147:     pub fn isPrimed(self: *const Momentum) bool {
148:         return self.primed;
149:     }
150: 
151:     pub fn getMetadata(self: *const Momentum, out: *Metadata) void {
152:         build_metadata_mod.buildMetadata(
153:             out,
154:             .momentum,
155:             self.line.mnemonic,
156:             self.line.description,
157:             &[_]build_metadata_mod.OutputText{
158:                 .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
159:             },
160:         );
161:     }
162: 
163:     pub fn updateScalar(self: *Momentum, sample: *const Scalar) OutputArray {
164:         const value = self.update(sample.value);
165:         return LineIndicator.wrapScalar(sample.time, value);
166:     }
167: 
168:     pub fn updateBar(self: *Momentum, sample: *const Bar) OutputArray {
169:         const value = self.update(self.line.extractBar(sample));
170:         return LineIndicator.wrapScalar(sample.time, value);
171:     }
172: 
173:     pub fn updateQuote(self: *Momentum, sample: *const Quote) OutputArray {
174:         const value = self.update(self.line.extractQuote(sample));
175:         return LineIndicator.wrapScalar(sample.time, value);
176:     }
177: 
178:     pub fn updateTrade(self: *Momentum, sample: *const Trade) OutputArray {
179:         const value = self.update(self.line.extractTrade(sample));
180:         return LineIndicator.wrapScalar(sample.time, value);
181:     }
182: 
183:     pub fn indicator(self: *Momentum) indicator_mod.Indicator {
184:         return .{
185:             .ptr = @ptrCast(self),
186:             .vtable = &vtable,
187:         };
188:     }
189: 
190:     const vtable = indicator_mod.Indicator.VTable{
191:         .isPrimed = vtableIsPrimed,
192:         .metadata = vtableMetadata,
193:         .updateScalar = vtableUpdateScalar,
194:         .updateBar = vtableUpdateBar,
195:         .updateQuote = vtableUpdateQuote,
196:         .updateTrade = vtableUpdateTrade,
197:     };
198: 
199:     fn vtableIsPrimed(ptr: *anyopaque) bool {

(Showing lines 100-199 of 430. Use offset=200 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 8.4s)

Good, I have all the patterns. Now let me create both Aroon and CMO files. Let me start with Aroon — a custom struct (no LineIndicator), HLCV override, 3 outputs, allocator for ring buffers.

**Tool: bash**

**Input:**
```json
{
  "command": "mkdir -p zig/src/indicators/tushar_chande/aroon zig/src/indicators/tushar_chande/chande_momentum_oscillator",
  "description": "Create tushar_chande indicator dirs"
}
```

---

## Assistant (Build · claude-opus-4.6 · 184.8s)

**Tool: write**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/tushar_chande/aroon/aroon.zig",
  "content": "const std = @import(\"std\");\nconst math = std.math;\n\nconst Bar = @import(\"bar\").Bar;\nconst Quote = @import(\"quote\").Quote;\nconst Trade = @import(\"trade\").Trade;\nconst Scalar = @import(\"scalar\").Scalar;\n\nconst indicator_mod = @import(\"../../core/indicator.zig\");\nconst build_metadata_mod = @import(\"../../core/build_metadata.zig\");\nconst identifier_mod = @import(\"../../core/identifier.zig\");\nconst metadata_mod = @import(\"../../core/metadata.zig\");\n\nconst OutputArray = indicator_mod.OutputArray;\nconst Identifier = identifier_mod.Identifier;\nconst Metadata = metadata_mod.Metadata;\n\n/// Enumerates the outputs of the Aroon indicator.\npub const AroonOutput = enum(u8) {\n    /// The Aroon Up line.\n    up = 1,\n    /// The Aroon Down line.\n    down = 2,\n    /// The Aroon Oscillator (Up - Down).\n    osc = 3,\n};\n\n/// Parameters to create an instance of the Aroon indicator.\npub const AroonParams = struct {\n    /// The lookback period. Must be >= 2. Default is 14.\n    length: usize = 14,\n};\n\n/// Tushar Chande's Aroon indicator.\n///\n/// Measures the number of periods since the highest high and lowest low\n/// within a lookback window. Produces three outputs:\n///   - Up: 100 * (Length - periods since highest high) / Length\n///   - Down: 100 * (Length - periods since lowest low) / Length\n///   - Osc: Up - Down\n///\n/// The indicator requires bar data (high, low). For scalar, quote, and\n/// trade updates, the single value substitutes for both.\npub const Aroon = struct {\n    length: usize,\n    factor: f64,\n    high_buf: []f64,\n    low_buf: []f64,\n    buffer_index: usize,\n    count: usize,\n    highest_index: usize,\n    lowest_index: usize,\n    up: f64,\n    down: f64,\n    osc: f64,\n    primed: bool,\n    allocator: std.mem.Allocator,\n    mnemonic_buf: [32]u8,\n    mnemonic_len: usize,\n\n    pub const Error = error{\n        InvalidLength,\n        OutOfMemory,\n    };\n\n    pub fn init(allocator: std.mem.Allocator, params: AroonParams) Error!Aroon {\n        const length = params.length;\n        if (length < 2) return error.InvalidLength;\n\n        const window_size = length + 1;\n        const high_buf = allocator.alloc(f64, window_size) catch return error.OutOfMemory;\n        errdefer allocator.free(high_buf);\n        const low_buf = allocator.alloc(f64, window_size) catch return error.OutOfMemory;\n\n        var mnemonic_buf: [32]u8 = undefined;\n        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, \"aroon({d})\", .{length}) catch\n            return error.InvalidLength;\n\n        return Aroon{\n            .length = length,\n            .factor = 100.0 / @as(f64, @floatFromInt(length)),\n            .high_buf = high_buf,\n            .low_buf = low_buf,\n            .buffer_index = 0,\n            .count = 0,\n            .highest_index = 0,\n            .lowest_index = 0,\n            .up = math.nan(f64),\n            .down = math.nan(f64),\n            .osc = math.nan(f64),\n            .primed = false,\n            .allocator = allocator,\n            .mnemonic_buf = mnemonic_buf,\n            .mnemonic_len = mnemonic_slice.len,\n        };\n    }\n\n    pub fn deinit(self: *Aroon) void {\n        self.allocator.free(self.high_buf);\n        self.allocator.free(self.low_buf);\n    }\n\n    pub fn fixSlices(self: *Aroon) void {\n        _ = self;\n    }\n\n    /// Core update given high and low values.\n    /// Returns [up, down, osc].\n    pub fn updateHighLow(self: *Aroon, high: f64, low: f64) [3]f64 {\n        if (math.isNan(high) or math.isNan(low)) {\n            return .{ math.nan(f64), math.nan(f64), math.nan(f64) };\n        }\n\n        const window_size = self.length + 1;\n        const today = self.count;\n\n        // Store in circular buffer.\n        const pos = self.buffer_index;\n        self.high_buf[pos] = high;\n        self.low_buf[pos] = low;\n        self.buffer_index = (self.buffer_index + 1) % window_size;\n        self.count += 1;\n\n        // Need at least length+1 bars (indices 0..length).\n        if (self.count < window_size) {\n            return .{ self.up, self.down, self.osc };\n        }\n\n        const trailing_index = today - self.length;\n\n        if (self.count == window_size) {\n            // First time: scan entire window.\n            self.highest_index = trailing_index;\n            self.lowest_index = trailing_index;\n\n            var i = trailing_index + 1;\n            while (i <= today) : (i += 1) {\n                const buf_pos = i % window_size;\n                if (self.high_buf[buf_pos] >= self.high_buf[self.highest_index % window_size]) {\n                    self.highest_index = i;\n                }\n                if (self.low_buf[buf_pos] <= self.low_buf[self.lowest_index % window_size]) {\n                    self.lowest_index = i;\n                }\n            }\n        } else {\n            // Subsequent: optimized update.\n            if (self.highest_index < trailing_index) {\n                self.highest_index = trailing_index;\n                var i = trailing_index + 1;\n                while (i <= today) : (i += 1) {\n                    const buf_pos = i % window_size;\n                    if (self.high_buf[buf_pos] >= self.high_buf[self.highest_index % window_size]) {\n                        self.highest_index = i;\n                    }\n                }\n            } else if (high >= self.high_buf[self.highest_index % window_size]) {\n                self.highest_index = today;\n            }\n\n            if (self.lowest_index < trailing_index) {\n                self.lowest_index = trailing_index;\n                var i = trailing_index + 1;\n                while (i <= today) : (i += 1) {\n                    const buf_pos = i % window_size;\n                    if (self.low_buf[buf_pos] <= self.low_buf[self.lowest_index % window_size]) {\n                        self.lowest_index = i;\n                    }\n                }\n            } else if (low <= self.low_buf[self.lowest_index % window_size]) {\n                self.lowest_index = today;\n            }\n        }\n\n        self.up = self.factor * @as(f64, @floatFromInt(self.length - (today - self.highest_index)));\n        self.down = self.factor * @as(f64, @floatFromInt(self.length - (today - self.lowest_index)));\n        self.osc = self.up - self.down;\n\n        if (!self.primed) {\n            self.primed = true;\n        }\n\n        return .{ self.up, self.down, self.osc };\n    }\n\n    pub fn isPrimed(self: *const Aroon) bool {\n        return self.primed;\n    }\n\n    pub fn getMetadata(self: *const Aroon, out: *Metadata) void {\n        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];\n        const desc_prefix = \"Aroon \";\n\n        var desc_buf: [64]u8 = undefined;\n        const desc = std.fmt.bufPrint(&desc_buf, \"{s}{s}\", .{ desc_prefix, mnemonic }) catch mnemonic;\n\n        var up_mnemonic_buf: [48]u8 = undefined;\n        const up_mnemonic = std.fmt.bufPrint(&up_mnemonic_buf, \"{s} up\", .{mnemonic}) catch mnemonic;\n        var down_mnemonic_buf: [48]u8 = undefined;\n        const down_mnemonic = std.fmt.bufPrint(&down_mnemonic_buf, \"{s} down\", .{mnemonic}) catch mnemonic;\n        var osc_mnemonic_buf: [48]u8 = undefined;\n        const osc_mnemonic = std.fmt.bufPrint(&osc_mnemonic_buf, \"{s} osc\", .{mnemonic}) catch mnemonic;\n\n        var up_desc_buf: [64]u8 = undefined;\n        const up_desc = std.fmt.bufPrint(&up_desc_buf, \"{s} Up\", .{desc}) catch desc;\n        var down_desc_buf: [64]u8 = undefined;\n        const down_desc = std.fmt.bufPrint(&down_desc_buf, \"{s} Down\", .{desc}) catch desc;\n        var osc_desc_buf: [64]u8 = undefined;\n        const osc_desc = std.fmt.bufPrint(&osc_desc_buf, \"{s} Oscillator\", .{desc}) catch desc;\n\n        build_metadata_mod.buildMetadata(out, Identifier.aroon, mnemonic, desc, &.{\n            .{ .mnemonic = up_mnemonic, .description = up_desc },\n            .{ .mnemonic = down_mnemonic, .description = down_desc },\n            .{ .mnemonic = osc_mnemonic, .description = osc_desc },\n        });\n    }\n\n    fn makeOutput(self: *const Aroon, time: i64) OutputArray {\n        var out = OutputArray{};\n        out.append(.{ .scalar = Scalar{ .time = time, .value = self.up } });\n        out.append(.{ .scalar = Scalar{ .time = time, .value = self.down } });\n        out.append(.{ .scalar = Scalar{ .time = time, .value = self.osc } });\n        return out;\n    }\n\n    pub fn updateScalar(self: *Aroon, sample: *const Scalar) OutputArray {\n        _ = self.updateHighLow(sample.value, sample.value);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateBar(self: *Aroon, sample: *const Bar) OutputArray {\n        _ = self.updateHighLow(sample.high, sample.low);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateQuote(self: *Aroon, sample: *const Quote) OutputArray {\n        const mid = (sample.bid_price + sample.ask_price) / 2.0;\n        _ = self.updateHighLow(mid, mid);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateTrade(self: *Aroon, sample: *const Trade) OutputArray {\n        _ = self.updateHighLow(sample.price, sample.price);\n        return self.makeOutput(sample.time);\n    }\n\n    // --- Indicator interface ---\n\n    pub fn indicator(self: *Aroon) indicator_mod.Indicator {\n        return .{\n            .ptr = @ptrCast(self),\n            .vtable = &vtable,\n        };\n    }\n\n    const vtable = indicator_mod.Indicator.VTable{\n        .isPrimed = vtableIsPrimed,\n        .metadata = vtableMetadata,\n        .updateScalar = vtableUpdateScalar,\n        .updateBar = vtableUpdateBar,\n        .updateQuote = vtableUpdateQuote,\n        .updateTrade = vtableUpdateTrade,\n    };\n\n    fn vtableIsPrimed(ptr: *const anyopaque) bool {\n        const self: *const Aroon = @ptrCast(@alignCast(ptr));\n        return self.isPrimed();\n    }\n\n    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {\n        const self: *const Aroon = @ptrCast(@alignCast(ptr));\n        self.getMetadata(out);\n    }\n\n    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {\n        const self: *Aroon = @ptrCast(@alignCast(ptr));\n        return self.updateScalar(sample);\n    }\n\n    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {\n        const self: *Aroon = @ptrCast(@alignCast(ptr));\n        return self.updateBar(sample);\n    }\n\n    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {\n        const self: *Aroon = @ptrCast(@alignCast(ptr));\n        return self.updateQuote(sample);\n    }\n\n    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {\n        const self: *Aroon = @ptrCast(@alignCast(ptr));\n        return self.updateTrade(sample);\n    }\n};\n\n// --- Tests ---\n\nconst tolerance: f64 = 1e-6;\n\nfn almostEqual(a: f64, b: f64, tol: f64) bool {\n    if (math.isNan(a) and math.isNan(b)) return true;\n    if (math.isNan(a) or math.isNan(b)) return false;\n    return @abs(a - b) <= tol;\n}\n\nconst test_high = [252]f64{\n    93.25, 94.94, 96.375, 96.19, 96.0, 94.72, 95.0, 93.72, 92.47, 92.75,\n    96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,\n    90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,\n    87.155, 87.25, 87.375, 88.97, 90.0, 89.845, 86.97, 85.94, 84.75, 85.47,\n    84.47, 88.5, 89.47, 90.0, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,\n    90.0, 88.875, 89.0, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,\n    90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92.0, 91.94,\n    89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,\n    104.5, 105.5, 106.125, 107.94, 106.25, 107.0, 108.75, 110.94, 110.94, 114.22,\n    123.0, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0, 118.315,\n    116.87, 116.75, 113.87, 114.62, 115.31, 116.0, 121.69, 119.87, 120.87, 116.75,\n    116.5, 116.0, 118.31, 121.5, 122.0, 121.44, 125.75, 127.75, 124.19, 124.44,\n    125.75, 124.69, 125.31, 132.0, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,\n    138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,\n    125.12, 126.5, 128.69, 126.62, 126.69, 126.0, 123.12, 121.87, 124.0, 127.0,\n    124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,\n    124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126.0, 128.5, 126.75, 129.75,\n    132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135.0, 132.38, 131.44, 130.88,\n    129.63, 127.25, 127.81, 125.0, 126.81, 124.75, 122.81, 122.25, 121.06, 120.0,\n    123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109.0,\n    108.5, 113.06, 93.0, 94.62, 95.12, 96.0, 95.56, 95.31, 99.0, 98.81,\n    96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,\n    95.0, 94.87, 98.25, 105.12, 108.44, 109.87, 105.0, 106.0, 104.94, 104.5,\n    104.44, 106.31, 112.87, 116.5, 119.19, 121.0, 122.12, 111.94, 112.75, 110.19,\n    107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110.0, 110.75, 110.5,\n    110.5, 109.5,\n};\n\nconst test_low = [252]f64{\n    90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92.0, 89.75, 89.44, 90.625,\n    92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,\n    87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,\n    84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,\n    82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89.0, 88.565, 90.095,\n    89.0, 86.47, 84.0, 83.315, 82.0, 83.25, 84.75, 85.28, 87.19, 88.44,\n    88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92.0, 90.53, 89.97, 88.815,\n    86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,\n    101.72, 101.72, 103.155, 105.69, 103.655, 104.0, 105.53, 108.53, 108.75, 107.75,\n    117.0, 118.0, 116.0, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,\n    112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,\n    112.56, 114.0, 114.75, 118.87, 119.0, 119.75, 122.62, 123.0, 121.75, 121.56,\n    123.12, 122.19, 122.75, 124.37, 128.0, 129.5, 130.81, 130.63, 132.13, 133.88,\n    135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,\n    122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121.0,\n    121.12, 118.94, 119.81, 121.0, 122.0, 124.5, 126.56, 123.5, 121.25, 121.06,\n    122.31, 121.0, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128.0,\n    128.38, 130.69, 131.63, 134.38, 132.0, 131.94, 131.94, 129.56, 123.75, 126.0,\n    126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120.0, 119.62, 115.5, 116.75,\n    119.06, 119.06, 115.06, 111.06, 113.12, 110.0, 105.0, 104.69, 103.87, 104.69,\n    105.44, 107.0, 89.0, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,\n    93.69, 93.5, 90.0, 90.19, 90.5, 92.12, 94.12, 94.87, 93.0, 93.87,\n    93.0, 92.62, 93.56, 98.37, 104.44, 106.0, 101.81, 104.12, 103.37, 102.12,\n    102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,\n    104.5, 105.75, 108.62, 107.75, 108.06, 108.0, 108.19, 108.12, 109.06, 108.75,\n    108.56, 106.62,\n};\n\nconst ExpectedRow = struct { up: f64, down: f64, osc: f64 };\n\nconst test_expected = blk: {\n    const nan = math.nan(f64);\n    break :blk [252]ExpectedRow{\n        .{ .up = nan, .down = nan, .osc = nan }, // 0\n        .{ .up = nan, .down = nan, .osc = nan }, // 1\n        .{ .up = nan, .down = nan, .osc = nan }, // 2\n        .{ .up = nan, .down = nan, .osc = nan }, // 3\n        .{ .up = nan, .down = nan, .osc = nan }, // 4\n        .{ .up = nan, .down = nan, .osc = nan }, // 5\n        .{ .up = nan, .down = nan, .osc = nan }, // 6\n        .{ .up = nan, .down = nan, .osc = nan }, // 7\n        .{ .up = nan, .down = nan, .osc = nan }, // 8\n        .{ .up = nan, .down = nan, .osc = nan }, // 9\n        .{ .up = nan, .down = nan, .osc = nan }, // 10\n        .{ .up = nan, .down = nan, .osc = nan }, // 11\n        .{ .up = nan, .down = nan, .osc = nan }, // 12\n        .{ .up = nan, .down = nan, .osc = nan }, // 13\n        .{ .up = 78.571428571428600, .down = 100.000000000000000, .osc = -21.428571428571400 },\n        .{ .up = 71.428571428571400, .down = 92.857142857142900, .osc = -21.428571428571400 },\n        .{ .up = 64.285714285714300, .down = 85.714285714285700, .osc = -21.428571428571400 },\n        .{ .up = 57.142857142857100, .down = 78.571428571428600, .osc = -21.428571428571400 },\n        .{ .up = 50.000000000000000, .down = 71.428571428571400, .osc = -21.428571428571400 },\n        .{ .up = 42.857142857142900, .down = 64.285714285714300, .osc = -21.428571428571400 },\n        .{ .up = 35.714285714285700, .down = 57.142857142857100, .osc = -21.428571428571400 },\n        .{ .up = 28.571428571428600, .down = 100.000000000000000, .osc = -71.428571428571400 },\n        .{ .up = 21.428571428571400, .down = 100.000000000000000, .osc = -78.571428571428600 },\n        .{ .up = 14.285714285714300, .down = 100.000000000000000, .osc = -85.714285714285700 },\n        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },\n        .{ .up = 21.428571428571400, .down = 85.714285714285700, .osc = -64.285714285714300 },\n        .{ .up = 14.285714285714300, .down = 78.571428571428600, .osc = -64.285714285714300 },\n        .{ .up = 7.142857142857140, .down = 71.428571428571400, .osc = -64.285714285714300 },\n        .{ .up = 0.000000000000000, .down = 64.285714285714300, .osc = -64.285714285714300 },\n        .{ .up = 14.285714285714300, .down = 57.142857142857100, .osc = -42.857142857142900 },\n        .{ .up = 7.142857142857140, .down = 50.000000000000000, .osc = -42.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 42.857142857142900, .osc = -42.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 35.714285714285700, .osc = -35.714285714285700 },\n        .{ .up = 92.857142857142900, .down = 28.571428571428600, .osc = 64.285714285714300 },\n        .{ .up = 85.714285714285700, .down = 21.428571428571400, .osc = 64.285714285714300 },\n        .{ .up = 78.571428571428600, .down = 14.285714285714300, .osc = 64.285714285714300 },\n        .{ .up = 71.428571428571400, .down = 7.142857142857140, .osc = 64.285714285714300 },\n        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },\n        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },\n        .{ .up = 50.000000000000000, .down = 92.857142857142900, .osc = -42.857142857142900 },\n        .{ .up = 42.857142857142900, .down = 85.714285714285700, .osc = -42.857142857142900 },\n        .{ .up = 100.000000000000000, .down = 78.571428571428600, .osc = 21.428571428571400 },\n        .{ .up = 100.000000000000000, .down = 71.428571428571400, .osc = 28.571428571428600 },\n        .{ .up = 92.857142857142900, .down = 64.285714285714300, .osc = 28.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 57.142857142857100, .osc = 42.857142857142900 },\n        .{ .up = 92.857142857142900, .down = 50.000000000000000, .osc = 42.857142857142900 },\n        .{ .up = 85.714285714285700, .down = 42.857142857142900, .osc = 42.857142857142900 },\n        .{ .up = 78.571428571428600, .down = 35.714285714285700, .osc = 42.857142857142900 },\n        .{ .up = 71.428571428571400, .down = 28.571428571428600, .osc = 42.857142857142900 },\n        .{ .up = 64.285714285714300, .down = 21.428571428571400, .osc = 42.857142857142900 },\n        .{ .up = 57.142857142857100, .down = 14.285714285714300, .osc = 42.857142857142900 },\n        .{ .up = 50.000000000000000, .down = 7.142857142857140, .osc = 42.857142857142900 },\n        .{ .up = 42.857142857142900, .down = 100.000000000000000, .osc = -57.142857142857100 },\n        .{ .up = 35.714285714285700, .down = 92.857142857142900, .osc = -57.142857142857100 },\n        .{ .up = 28.571428571428600, .down = 85.714285714285700, .osc = -57.142857142857100 },\n        .{ .up = 21.428571428571400, .down = 78.571428571428600, .osc = -57.142857142857100 },\n        .{ .up = 14.285714285714300, .down = 71.428571428571400, .osc = -57.142857142857100 },\n        .{ .up = 7.142857142857140, .down = 64.285714285714300, .osc = -57.142857142857200 },\n        .{ .up = 0.000000000000000, .down = 57.142857142857100, .osc = -57.142857142857100 },\n        .{ .up = 14.285714285714300, .down = 50.000000000000000, .osc = -35.714285714285700 },\n        .{ .up = 100.000000000000000, .down = 42.857142857142900, .osc = 57.142857142857100 },\n        .{ .up = 100.000000000000000, .down = 35.714285714285700, .osc = 64.285714285714300 },\n        .{ .up = 100.000000000000000, .down = 28.571428571428600, .osc = 71.428571428571400 },\n        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },\n        .{ .up = 92.857142857142900, .down = 7.142857142857140, .osc = 85.714285714285700 },\n        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },\n        .{ .up = 78.571428571428600, .down = 0.000000000000000, .osc = 78.571428571428600 },\n        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },\n        .{ .up = 64.285714285714300, .down = 100.000000000000000, .osc = -35.714285714285700 },\n        .{ .up = 57.142857142857100, .down = 100.000000000000000, .osc = -42.857142857142900 },\n        .{ .up = 50.000000000000000, .down = 100.000000000000000, .osc = -50.000000000000000 },\n        .{ .up = 42.857142857142900, .down = 92.857142857142900, .osc = -50.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 85.714285714285700, .osc = 14.285714285714300 },\n        .{ .up = 100.000000000000000, .down = 78.571428571428600, .osc = 21.428571428571400 },\n        .{ .up = 100.000000000000000, .down = 71.428571428571400, .osc = 28.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 64.285714285714300, .osc = 35.714285714285700 },\n        .{ .up = 92.857142857142900, .down = 57.142857142857100, .osc = 35.714285714285700 },\n        .{ .up = 85.714285714285700, .down = 50.000000000000000, .osc = 35.714285714285700 },\n        .{ .up = 78.571428571428600, .down = 42.857142857142900, .osc = 35.714285714285700 },\n        .{ .up = 71.428571428571400, .down = 35.714285714285700, .osc = 35.714285714285700 },\n        .{ .up = 100.000000000000000, .down = 28.571428571428600, .osc = 71.428571428571400 },\n        .{ .up = 92.857142857142900, .down = 21.428571428571400, .osc = 71.428571428571400 },\n        .{ .up = 85.714285714285700, .down = 14.285714285714300, .osc = 71.428571428571400 },\n        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },\n        .{ .up = 85.714285714285700, .down = 21.428571428571400, .osc = 64.285714285714300 },\n        .{ .up = 78.571428571428600, .down = 14.285714285714300, .osc = 64.285714285714300 },\n        .{ .up = 71.428571428571400, .down = 7.142857142857140, .osc = 64.285714285714300 },\n        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },\n        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },\n        .{ .up = 50.000000000000000, .down = 7.142857142857140, .osc = 42.857142857142900 },\n        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },\n        .{ .up = 35.714285714285700, .down = 0.000000000000000, .osc = 35.714285714285700 },\n        .{ .up = 28.571428571428600, .down = 0.000000000000000, .osc = 28.571428571428600 },\n        .{ .up = 21.428571428571400, .down = 14.285714285714300, .osc = 7.142857142857140 },\n        .{ .up = 14.285714285714300, .down = 7.142857142857140, .osc = 7.142857142857140 },\n        .{ .up = 7.142857142857140, .down = 0.000000000000000, .osc = 7.142857142857140 },\n        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },\n        .{ .up = 100.000000000000000, .down = 78.571428571428600, .osc = 21.428571428571400 },\n        .{ .up = 92.857142857142900, .down = 71.428571428571400, .osc = 21.428571428571400 },\n        .{ .up = 85.714285714285700, .down = 64.285714285714300, .osc = 21.428571428571400 },\n        .{ .up = 78.571428571428600, .down = 57.142857142857100, .osc = 21.428571428571400 },\n        .{ .up = 71.428571428571400, .down = 50.000000000000000, .osc = 21.428571428571400 },\n        .{ .up = 64.285714285714300, .down = 42.857142857142900, .osc = 21.428571428571400 },\n        .{ .up = 57.142857142857100, .down = 35.714285714285700, .osc = 21.428571428571400 },\n        .{ .up = 50.000000000000000, .down = 28.571428571428600, .osc = 21.428571428571400 },\n        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },\n        .{ .up = 92.857142857142900, .down = 14.285714285714300, .osc = 78.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },\n        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },\n        .{ .up = 78.571428571428600, .down = 28.571428571428600, .osc = 50.000000000000000 },\n        .{ .up = 71.428571428571400, .down = 21.428571428571400, .osc = 50.000000000000000 },\n        .{ .up = 64.285714285714300, .down = 14.285714285714300, .osc = 50.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },\n        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },\n        .{ .up = 92.857142857142900, .down = 7.142857142857140, .osc = 85.714285714285700 },\n        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },\n        .{ .up = 78.571428571428600, .down = 7.142857142857140, .osc = 71.428571428571400 },\n        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },\n        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },\n        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },\n        .{ .up = 50.000000000000000, .down = 100.000000000000000, .osc = -50.000000000000000 },\n        .{ .up = 42.857142857142900, .down = 100.000000000000000, .osc = -57.142857142857100 },\n        .{ .up = 35.714285714285700, .down = 92.857142857142900, .osc = -57.142857142857100 },\n        .{ .up = 28.571428571428600, .down = 85.714285714285700, .osc = -57.142857142857100 },\n        .{ .up = 21.428571428571400, .down = 78.571428571428600, .osc = -57.142857142857100 },\n        .{ .up = 14.285714285714300, .down = 71.428571428571400, .osc = -57.142857142857100 },\n        .{ .up = 7.142857142857140, .down = 64.285714285714300, .osc = -57.142857142857200 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 21.428571428571400, .down = 100.000000000000000, .osc = -78.571428571428600 },\n        .{ .up = 14.285714285714300, .down = 100.000000000000000, .osc = -85.714285714285700 },\n        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },\n        .{ .up = 0.000000000000000, .down = 78.571428571428600, .osc = -78.571428571428600 },\n        .{ .up = 28.571428571428600, .down = 71.428571428571400, .osc = -42.857142857142900 },\n        .{ .up = 21.428571428571400, .down = 64.285714285714300, .osc = -42.857142857142900 },\n        .{ .up = 14.285714285714300, .down = 57.142857142857100, .osc = -42.857142857142900 },\n        .{ .up = 7.142857142857140, .down = 50.000000000000000, .osc = -42.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 42.857142857142900, .osc = -42.857142857142900 },\n        .{ .up = 100.000000000000000, .down = 35.714285714285700, .osc = 64.285714285714300 },\n        .{ .up = 92.857142857142900, .down = 28.571428571428600, .osc = 64.285714285714300 },\n        .{ .up = 85.714285714285700, .down = 21.428571428571400, .osc = 64.285714285714300 },\n        .{ .up = 78.571428571428600, .down = 14.285714285714300, .osc = 64.285714285714300 },\n        .{ .up = 71.428571428571400, .down = 7.142857142857140, .osc = 64.285714285714300 },\n        .{ .up = 64.285714285714300, .down = 0.000000000000000, .osc = 64.285714285714300 },\n        .{ .up = 57.142857142857100, .down = 14.285714285714300, .osc = 42.857142857142900 },\n        .{ .up = 50.000000000000000, .down = 7.142857142857140, .osc = 42.857142857142900 },\n        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },\n        .{ .up = 35.714285714285700, .down = 0.000000000000000, .osc = 35.714285714285700 },\n        .{ .up = 28.571428571428600, .down = 64.285714285714300, .osc = -35.714285714285700 },\n        .{ .up = 21.428571428571400, .down = 57.142857142857100, .osc = -35.714285714285700 },\n        .{ .up = 100.000000000000000, .down = 50.000000000000000, .osc = 50.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 42.857142857142900, .osc = 57.142857142857100 },\n        .{ .up = 100.000000000000000, .down = 35.714285714285700, .osc = 64.285714285714300 },\n        .{ .up = 100.000000000000000, .down = 28.571428571428600, .osc = 71.428571428571400 },\n        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },\n        .{ .up = 92.857142857142900, .down = 14.285714285714300, .osc = 78.571428571428600 },\n        .{ .up = 85.714285714285700, .down = 7.142857142857140, .osc = 78.571428571428600 },\n        .{ .up = 78.571428571428600, .down = 0.000000000000000, .osc = 78.571428571428600 },\n        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },\n        .{ .up = 64.285714285714300, .down = 7.142857142857140, .osc = 57.142857142857200 },\n        .{ .up = 57.142857142857100, .down = 0.000000000000000, .osc = 57.142857142857100 },\n        .{ .up = 50.000000000000000, .down = 0.000000000000000, .osc = 50.000000000000000 },\n        .{ .up = 42.857142857142900, .down = 78.571428571428600, .osc = -35.714285714285700 },\n        .{ .up = 35.714285714285700, .down = 100.000000000000000, .osc = -64.285714285714300 },\n        .{ .up = 28.571428571428600, .down = 100.000000000000000, .osc = -71.428571428571400 },\n        .{ .up = 21.428571428571400, .down = 92.857142857142900, .osc = -71.428571428571400 },\n        .{ .up = 14.285714285714300, .down = 85.714285714285700, .osc = -71.428571428571400 },\n        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 7.142857142857140, .down = 92.857142857142900, .osc = -85.714285714285700 },\n        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },\n        .{ .up = 0.000000000000000, .down = 78.571428571428600, .osc = -78.571428571428600 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },\n        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 7.142857142857140, .down = 100.000000000000000, .osc = -92.857142857142900 },\n        .{ .up = 0.000000000000000, .down = 100.000000000000000, .osc = -100.000000000000000 },\n        .{ .up = 0.000000000000000, .down = 92.857142857142900, .osc = -92.857142857142900 },\n        .{ .up = 28.571428571428600, .down = 85.714285714285700, .osc = -57.142857142857100 },\n        .{ .up = 21.428571428571400, .down = 78.571428571428600, .osc = -57.142857142857100 },\n        .{ .up = 14.285714285714300, .down = 100.000000000000000, .osc = -85.714285714285700 },\n        .{ .up = 7.142857142857140, .down = 92.857142857142900, .osc = -85.714285714285700 },\n        .{ .up = 0.000000000000000, .down = 85.714285714285700, .osc = -85.714285714285700 },\n        .{ .up = 0.000000000000000, .down = 78.571428571428600, .osc = -78.571428571428600 },\n        .{ .up = 0.000000000000000, .down = 71.428571428571400, .osc = -71.428571428571400 },\n        .{ .up = 7.142857142857140, .down = 64.285714285714300, .osc = -57.142857142857200 },\n        .{ .up = 0.000000000000000, .down = 57.142857142857100, .osc = -57.142857142857100 },\n        .{ .up = 0.000000000000000, .down = 50.000000000000000, .osc = -50.000000000000000 },\n        .{ .up = 35.714285714285700, .down = 42.857142857142900, .osc = -7.142857142857140 },\n        .{ .up = 28.571428571428600, .down = 35.714285714285700, .osc = -7.142857142857150 },\n        .{ .up = 21.428571428571400, .down = 28.571428571428600, .osc = -7.142857142857140 },\n        .{ .up = 14.285714285714300, .down = 21.428571428571400, .osc = -7.142857142857140 },\n        .{ .up = 7.142857142857140, .down = 14.285714285714300, .osc = -7.142857142857140 },\n        .{ .up = 0.000000000000000, .down = 7.142857142857140, .osc = -7.142857142857140 },\n        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },\n        .{ .up = 35.714285714285700, .down = 64.285714285714300, .osc = -28.571428571428600 },\n        .{ .up = 28.571428571428600, .down = 57.142857142857100, .osc = -28.571428571428600 },\n        .{ .up = 21.428571428571400, .down = 50.000000000000000, .osc = -28.571428571428600 },\n        .{ .up = 14.285714285714300, .down = 42.857142857142900, .osc = -28.571428571428600 },\n        .{ .up = 7.142857142857140, .down = 35.714285714285700, .osc = -28.571428571428600 },\n        .{ .up = 0.000000000000000, .down = 28.571428571428600, .osc = -28.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },\n        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },\n        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },\n        .{ .up = 85.714285714285700, .down = 0.000000000000000, .osc = 85.714285714285700 },\n        .{ .up = 78.571428571428600, .down = 0.000000000000000, .osc = 78.571428571428600 },\n        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },\n        .{ .up = 64.285714285714300, .down = 35.714285714285700, .osc = 28.571428571428600 },\n        .{ .up = 57.142857142857100, .down = 28.571428571428600, .osc = 28.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 21.428571428571400, .osc = 78.571428571428600 },\n        .{ .up = 100.000000000000000, .down = 14.285714285714300, .osc = 85.714285714285700 },\n        .{ .up = 100.000000000000000, .down = 7.142857142857140, .osc = 92.857142857142900 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 100.000000000000000, .down = 0.000000000000000, .osc = 100.000000000000000 },\n        .{ .up = 92.857142857142900, .down = 0.000000000000000, .osc = 92.857142857142900 },\n        .{ .up = 85.714285714285700, .down = 14.285714285714300, .osc = 71.428571428571400 },\n        .{ .up = 78.571428571428600, .down = 7.142857142857140, .osc = 71.428571428571400 },\n        .{ .up = 71.428571428571400, .down = 0.000000000000000, .osc = 71.428571428571400 },\n        .{ .up = 64.285714285714300, .down = 14.285714285714300, .osc = 50.000000000000000 },\n        .{ .up = 57.142857142857100, .down = 7.142857142857140, .osc = 50.000000000000000 },\n        .{ .up = 50.000000000000000, .down = 0.000000000000000, .osc = 50.000000000000000 },\n        .{ .up = 42.857142857142900, .down = 0.000000000000000, .osc = 42.857142857142900 },\n        .{ .up = 35.714285714285700, .down = 0.000000000000000, .osc = 35.714285714285700 },\n        .{ .up = 28.571428571428600, .down = 57.142857142857100, .osc = -28.571428571428600 },\n        .{ .up = 21.428571428571400, .down = 50.000000000000000, .osc = -28.571428571428600 },\n        .{ .up = 14.285714285714300, .down = 42.857142857142900, .osc = -28.571428571428600 },\n        .{ .up = 7.142857142857140, .down = 35.714285714285700, .osc = -28.571428571428600 },\n        .{ .up = 0.000000000000000, .down = 28.571428571428600, .osc = -28.571428571428600 },\n        .{ .up = 7.142857142857140, .down = 21.428571428571400, .osc = -14.285714285714300 },\n    };\n};\n\ntest \"Aroon length=14 full data\" {\n    const allocator = std.testing.allocator;\n    var ind = try Aroon.init(allocator, .{ .length = 14 });\n    defer ind.deinit();\n\n    for (0..252) |i| {\n        const result = ind.updateHighLow(test_high[i], test_low[i]);\n        const exp = test_expected[i];\n\n        if (math.isNan(exp.up)) {\n            try std.testing.expect(math.isNan(result[0]));\n            continue;\n        }\n\n        try std.testing.expect(almostEqual(result[0], exp.up, tolerance));\n        try std.testing.expect(almostEqual(result[1], exp.down, tolerance));\n        try std.testing.expect(almostEqual(result[2], exp.osc, tolerance));\n    }\n}\n\ntest \"Aroon isPrimed\" {\n    const allocator = std.testing.allocator;\n    var ind = try Aroon.init(allocator, .{ .length = 14 });\n    defer ind.deinit();\n\n    try std.testing.expect(!ind.isPrimed());\n\n    for (0..14) |i| {\n        _ = ind.updateHighLow(test_high[i], test_low[i]);\n        try std.testing.expect(!ind.isPrimed());\n    }\n\n    _ = ind.updateHighLow(test_high[14], test_low[14]);\n    try std.testing.expect(ind.isPrimed());\n}\n\ntest \"Aroon NaN input\" {\n    const allocator = std.testing.allocator;\n    var ind = try Aroon.init(allocator, .{ .length = 14 });\n    defer ind.deinit();\n\n    const result = ind.updateHighLow(math.nan(f64), 1.0);\n    try std.testing.expect(math.isNan(result[0]));\n    try std.testing.expect(math.isNan(result[1]));\n    try std.testing.expect(math.isNan(result[2]));\n}\n\ntest \"Aroon invalid params\" {\n    const allocator = std.testing.allocator;\n    try std.testing.expectError(error.InvalidLength, Aroon.init(allocator, .{ .length = 1 }));\n    try std.testing.expectError(error.InvalidLength, Aroon.init(allocator, .{ .length = 0 }));\n}\n\ntest \"Aroon metadata\" {\n    const allocator = std.testing.allocator;\n    var ind = try Aroon.init(allocator, .{ .length = 14 });\n    defer ind.deinit();\n\n    var meta: Metadata = undefined;\n    ind.getMetadata(&meta);\n\n    try std.testing.expectEqual(Identifier.aroon, meta.identifier);\n    try std.testing.expectEqualStrings(\"aroon(14)\", meta.mnemonic);\n    try std.testing.expectEqual(@as(usize, 3), meta.outputs_len);\n}\n\ntest \"Aroon updateBar\" {\n    const allocator = std.testing.allocator;\n    var ind = try Aroon.init(allocator, .{ .length = 14 });\n    defer ind.deinit();\n\n    // Feed 14 bars (not primed yet).\n    for (0..14) |i| {\n        const bar = Bar{ .time = 0, .open = 0, .high = test_high[i], .low = test_low[i], .close = 0, .volume = 0 };\n        const out = ind.updateBar(&bar);\n        try std.testing.expect(math.isNan(out.slice()[0].scalar.value));\n    }\n\n    // Index 14: first valid.\n    const bar14 = Bar{ .time = 0, .open = 0, .high = test_high[14], .low = test_low[14], .close = 0, .volume = 0 };\n    const out = ind.updateBar(&bar14);\n    try std.testing.expect(almostEqual(out.slice()[0].scalar.value, test_expected[14].up, tolerance));\n    try std.testing.expect(almostEqual(out.slice()[1].scalar.value, test_expected[14].down, tolerance));\n    try std.testing.expect(almostEqual(out.slice()[2].scalar.value, test_expected[14].osc, tolerance));\n}\n"
}
```

**Output:**
```
Wrote file successfully.

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/icalc/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.identifier" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.factory.factory" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/iconf/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.descriptors" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.core.outputs.heatmap" could not be resolved
ERROR [26:6] Import "py.indicators.core.outputs.shape" could not be resolved
ERROR [27:6] Import "py.indicators.core.pane" could not be resolved
ERROR [28:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [31:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/ifres/main.py">
ERROR [16:6] Import "py.indicators.core.frequency_response" could not be resolved
ERROR [17:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [18:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [21:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>
```

---

## User

---

## Assistant (Compaction · claude-opus-4.6 · 68.7s)



## Goal

Port the `indicators/` module from Go/TypeScript to Zig. Python and Rust ports are complete. Currently porting indicators in small batches by author family.

## Instructions

- Go and TS code is the reference implementation. Python is the reference for other modules.
- All ports must match reference output to tolerance specified by Go tests (typically 1e-8).
- Port one language at a time: **Python → Rust → Zig**. Python and Rust are done.
- Port indicators in small batches by author family, starting with simplest, factory last, cmd tools last.
- Zig naming: `snake_case` files, `camelCase` functions/fields, `PascalCase` types/enums, lowercase enum values.
- **Each indicator lives in its own subfolder** (e.g., `patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig`)
- **Author folder naming**: Use `snake_case` matching Rust (e.g., `patrick_mulloy`), NOT Go's `patrickmulloy` style.
- **Output enums are 1-based** (matching Go's `iota+1`).
- **fixSlices() pattern**: After `init()` returns by value, slices point into old stack frame. Must call `fixSlices()` to repoint them. `fixSlices` must be `pub`.
- **Metadata out-pointer pattern**: `getMetadata(out: *Metadata)` returns void.
- **No allocator needed** for EMA-like indicators (no window). SMA and Variance need allocator for ring buffer.
- **MaUnion pattern**: For indicators that support both SMA and EMA sub-indicators, use a tagged union with `update`, `isPrimed`, `deinit` methods (see APO, Stochastic, MACD, BollingerBands for examples).
- **Per-output metadata**: `OutputText` struct has only `mnemonic` and `description` fields (no suffix fields). Each output needs a fully-formed string built with `bufPrint` in `getMetadata`. See MACD or BollingerBands for the pattern.
- **Metadata fields are plain `[]const u8`**, NOT functions — use `meta.mnemonic` not `meta.mnemonic()`.

## Discoveries

### Key Type Corrections (learned fixing MACD/PPO compilation)
- **Bar/Quote/Trade function types**: Use `bar_component.BarFunc`, `quote_component.QuoteFunc`, `trade_component.TradeFunc` (NOT `BarExtractFn` etc.)
- **Scalar.time is `i64`**, not `f64`
- **OutputArray**: Is a struct with `.values` array and `.len`, NOT a bare array. Use `OutputArray.append(.{ .scalar = .{ .time = t, .value = v } })` to add outputs, `out.slice()[i].scalar.value` to read them.
- **Metadata fields**: `outputs_len` (not `num_outputs`), `outputs_buf[i]` (not `outputs[i]`)
- **Band output type**: Use `band_mod.Band.new(time, lower, upper)` and `Band.empty(time)` for NaN bands. Access via `out.slice()[5].band`.

### Per-output metadata pattern (learned from BollingerBands)
- `build_metadata.OutputText` only has `{ .mnemonic, .description }` — NO suffix fields
- Must create per-output buf+bufPrint for each output's mnemonic and description in `getMetadata`

### HLCV Indicators Pattern (AdvanceDecline, AdvanceDeclineOscillator, Aroon)
- These indicators need HLCV data from bars, not just a single scalar component.
- They override `updateBar` to extract `sample.high, sample.low, sample.close, sample.volume` directly.
- For scalar/quote/trade updates, H=L=C so the range is 0 and AD is unchanged.
- The `update()` method calls `updateHlcv(sample, sample, sample, 1)`.
- AdvanceDecline has NO parameters (no allocator needed, empty `fixSlices`/`deinit`).
- AdvanceDeclineOscillator uses MaUnion (SMA/EMA) and needs allocator for SMA case.

### StochasticRSI Dependency
- StochasticRelativeStrengthIndex depends on RelativeStrengthIndex (welleswilder family).
- RSI is NOT yet ported to Zig, so StochRSI must be deferred until after welleswilder family is done.

### 2-Indicator Families Identified (7 total)
| Family | Indicators |
|--------|-----------|
| patrickmulloy | DEMA, TEMA |
| timtillson | T2EMA, T3EMA |
| geraldappel | MACD, PPO |
| johnbollinger | BollingerBands, BollingerBandsTrend |
| larrywilliams | UltimateOscillator, WilliamsPercentR |
| marcchaikin | AdvanceDecline, AdvanceDeclineOscillator |
| custom | GoertzelSpectrum, MaximumEntropySpectrum |

## Accomplished

### Fully Complete ✅ (650 tests pass)
- **Python port**: All 63 indicators, factory, frequency_response, cmd tools. 1142 tests pass.
- **Rust port**: All 67 indicators, factory, frequency_response, cmd tools. 985 tests pass.
- **Zig Core Types**: All 17 core type files complete.
- **Zig Common Indicators**: All 13 complete (SMA, WMA, TMA, EMA, Momentum, ROC, ROCP, ROCR, Variance, StdDev, APO, LinearRegression, PearsonsCC).
- **Zig Single-Indicator Families**: All 9 complete (BalanceOfPower, OnBalanceVolume, CCI, TEMO, MFI, KAMA, JMA, Stochastic, ATCF).
- **Zig patrick_mulloy family**: DEMA and TEMA both complete and passing.
- **Zig timtillson family**: T2EMA and T3EMA both complete and passing.
- **Zig gerald_appel family**: MACD and PPO both complete and passing.
- **Zig john_bollinger family**: BollingerBands and BollingerBandsTrend both complete and passing.
- **Zig larry_williams family**: WilliamsPercentR and UltimateOscillator both complete and passing.
- **Zig marc_chaikin family**: AdvanceDecline and AdvanceDeclineOscillator both complete and passing.
- **650 total Zig tests pass**.

### Just Written — NOT YET COMPILED/TESTED ⚠️
- **Zig tushar_chande/aroon/aroon.zig**: Full implementation written with all 252-entry test data, 4 tests. Follows HLCV-override pattern (custom `updateHighLow(high, low)` method, `updateBar` extracts `sample.high`/`sample.low`). Uses allocator for ring buffers. 3 outputs (up, down, osc).
  - **NOT registered** in `indicators.zig`
  - **NOT compiled or tested**

### Not Yet Written
- **ChandeMomentumOscillator (CMO)**: Go reference read. Uses LineIndicator pattern with component triple. Ring buffer of deltas tracking gainSum/lossSum. Epsilon guard (1e-12). Length min=1, default=14. Test tolerance 1e-13. Single output.
  - Go source fully read from `go/indicators/tusharchande/chandemomentumoscillator/`

### Immediate Next Steps
1. **Register Aroon** in `indicators.zig`: Add `pub const aroon = @import("tushar_chande/aroon/aroon.zig");` and add to comptime block
2. **Run `zig build test`** and fix any compilation/test errors for Aroon
3. **Write CMO** (`zig/src/indicators/tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig`) — directory already created
4. **Register CMO** in `indicators.zig` and test
5. **StochRSI is blocked** on welleswilder/RSI — skip for now

### Remaining Work After tushar_chande
1. **tusharchande/StochasticRelativeStrengthIndex** (blocked on welleswilder/RSI)
2. **welleswilder** (12 indicators including RSI — unblocks StochRSI)
3. **johnehlers** (23 indicators)
4. **custom** (GoertzelSpectrum, MaximumEntropySpectrum)
5. Factory, frequency_response, cmd tools

## Relevant files / directories

### Registration file (NEEDS Aroon + CMO imports added):
- `zig/src/indicators/indicators.zig` — Root module; needs aroon and chande_momentum_oscillator added

### Just created (need registration + compile + test):
- `zig/src/indicators/tushar_chande/aroon/aroon.zig` — Aroon, written ⚠️ NOT TESTED
- `zig/src/indicators/tushar_chande/chande_momentum_oscillator/` — directory created, file NOT yet written

### Go reference implementations (already read):
- `go/indicators/tusharchande/aroon/aroon.go` — Aroon reference (high/low circular buffer, optimized tracking)
- `go/indicators/tusharchande/aroon/aroon_test.go` — Test data (252 entries, tolerance 1e-6)
- `go/indicators/tusharchande/aroon/params.go` — Length only, default 14, min 2
- `go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator.go` — CMO reference (ring buffer of deltas, gainSum/lossSum)
- `go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator_test.go` — Test data (book length=10 with 13 entries, tolerance 1e-13)
- `go/indicators/tusharchande/chandemomentumoscillator/params.go` — Length + component triple, default 14, min 1
- `go/indicators/tusharchande/stochasticrelativestrengthindex/stochasticrelativestrengthindex.go` — StochRSI reference (depends on RSI)
- `go/indicators/tusharchande/stochasticrelativestrengthindex/stochasticrelativestrengthindex_test.go` — Test data (3 test cases)
- `go/indicators/tusharchande/stochasticrelativestrengthindex/params.go` — Length, FastKLength, FastDLength, MA type, component triple

### Key pattern references:
- `zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig` — HLCV override pattern with allocator, ring buffers
- `zig/src/indicators/common/momentum/momentum.zig` — LineIndicator + component triple + ring buffer pattern (model for CMO)
- `zig/src/indicators/common/absolute_price_oscillator/absolute_price_oscillator.zig` — MaUnion + mnemonic_buf/fixSlices pattern
- `zig/src/indicators/gerald_appel/moving_average_convergence_divergence/moving_average_convergence_divergence.zig` — Multi-output + MaUnion + per-output metadata
- `zig/src/indicators/john_bollinger/bollinger_bands/bollinger_bands.zig` — 6-output + Band type + per-output metadata
- `zig/src/indicators/core/indicator.zig` — OutputArray, OutputValue, Indicator vtable interface
- `zig/src/indicators/core/build_metadata.zig` — buildMetadata helper
- `zig/src/indicators/core/identifier.zig` — All identifier enum values

### Build:
- `zig/build.zig` — Single `indicators` module

### Skill files:
- `.opencode/skills/indicator-architecture/SKILL.md`
- `.opencode/skills/indicator-conversion/SKILL.md`

---

## User

---

## Assistant (Build · claude-opus-4.6 · 6.6s)

_Thinking:_

Let me continue with the immediate next steps: register Aroon in indicators.zig, compile/test, then write CMO.

Let me start by registering Aroon and testing it, then write CMO.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Register Aroon in indicators.zig and compile/test",
      "status": "in_progress",
      "priority": "high"
    },
    {
      "content": "Write CMO implementation",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "Register CMO in indicators.zig and compile/test",
      "status": "pending",
      "priority": "high"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Register Aroon in indicators.zig and compile/test",
    "status": "in_progress",
    "priority": "high"
  },
  {
    "content": "Write CMO implementation",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "Register CMO in indicators.zig and compile/test",
    "status": "pending",
    "priority": "high"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 4.5s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig</path>
<type>file</type>
<content>1: // Root module for the indicators library.
2: // Re-exports all core types and sub-modules.
3: 
4: // --- Core enums ---
5: pub const shape = @import("core/outputs/shape.zig");
6: pub const role = @import("core/role.zig");
7: pub const pane = @import("core/pane.zig");
8: pub const adaptivity = @import("core/adaptivity.zig");
9: pub const input_requirement = @import("core/input_requirement.zig");
10: pub const volume_usage = @import("core/volume_usage.zig");
11: 
12: // --- Core types ---
13: pub const identifier = @import("core/identifier.zig");
14: pub const output_descriptor = @import("core/output_descriptor.zig");
15: pub const output_metadata = @import("core/output_metadata.zig");
16: pub const metadata = @import("core/metadata.zig");
17: pub const descriptor = @import("core/descriptor.zig");
18: pub const descriptors = @import("core/descriptors.zig");
19: pub const build_metadata = @import("core/build_metadata.zig");
20: pub const component_triple_mnemonic = @import("core/component_triple_mnemonic.zig");
21: 
22: // --- Indicator interface and outputs ---
23: pub const indicator = @import("core/indicator.zig");
24: pub const line_indicator = @import("core/line_indicator.zig");
25: 
26: // --- Output types ---
27: pub const band = @import("core/outputs/band.zig");
28: pub const heatmap = @import("core/outputs/heatmap.zig");
29: pub const polyline = @import("core/outputs/polyline.zig");
30: 
31: // --- Convenience type aliases ---
32: pub const Identifier = identifier.Identifier;
33: pub const Shape = shape.Shape;
34: pub const Role = role.Role;
35: pub const Pane = pane.Pane;
36: pub const Adaptivity = adaptivity.Adaptivity;
37: pub const InputRequirement = input_requirement.InputRequirement;
38: pub const VolumeUsage = volume_usage.VolumeUsage;
39: pub const OutputDescriptor = output_descriptor.OutputDescriptor;
40: pub const OutputMetadata = output_metadata.OutputMetadata;
41: pub const Metadata = metadata.Metadata;
42: pub const Descriptor = descriptor.Descriptor;
43: pub const Indicator = indicator.Indicator;
44: pub const OutputValue = indicator.OutputValue;
45: pub const OutputArray = indicator.OutputArray;
46: pub const LineIndicator = line_indicator.LineIndicator;
47: pub const Band = band.Band;
48: pub const Heatmap = heatmap.Heatmap;
49: pub const Polyline = polyline.Polyline;
50: pub const Point = polyline.Point;
51: 
52: // --- Functions ---
53: pub const descriptorOf = descriptor.descriptorOf;
54: pub const allDescriptors = descriptor.allDescriptors;
55: pub const buildMetadata = build_metadata.buildMetadata;
56: pub const componentTripleMnemonic = component_triple_mnemonic.componentTripleMnemonic;
57: 
58: // --- Indicator implementations ---
59: pub const simple_moving_average = @import("common/simple_moving_average/simple_moving_average.zig");
60: pub const weighted_moving_average = @import("common/weighted_moving_average/weighted_moving_average.zig");
61: pub const triangular_moving_average = @import("common/triangular_moving_average/triangular_moving_average.zig");
62: pub const exponential_moving_average = @import("common/exponential_moving_average/exponential_moving_average.zig");
63: pub const momentum = @import("common/momentum/momentum.zig");
64: pub const rate_of_change = @import("common/rate_of_change/rate_of_change.zig");
65: pub const rate_of_change_percent = @import("common/rate_of_change_percent/rate_of_change_percent.zig");
66: pub const rate_of_change_ratio = @import("common/rate_of_change_ratio/rate_of_change_ratio.zig");
67: pub const variance = @import("common/variance/variance.zig");
68: pub const standard_deviation = @import("common/standard_deviation/standard_deviation.zig");
69: pub const absolute_price_oscillator = @import("common/absolute_price_oscillator/absolute_price_oscillator.zig");
70: pub const linear_regression = @import("common/linear_regression/linear_regression.zig");
71: pub const pearsons_correlation_coefficient = @import("common/pearsons_correlation_coefficient/pearsons_correlation_coefficient.zig");
72: pub const balance_of_power = @import("igor_livshin/balance_of_power/balance_of_power.zig");
73: pub const on_balance_volume = @import("joseph_granville/on_balance_volume/on_balance_volume.zig");
74: pub const commodity_channel_index = @import("donald_lambert/commodity_channel_index/commodity_channel_index.zig");
75: pub const triple_exponential_moving_average_oscillator = @import("jack_hutson/triple_exponential_moving_average_oscillator/triple_exponential_moving_average_oscillator.zig");
76: pub const money_flow_index = @import("gene_quong/money_flow_index/money_flow_index.zig");
77: pub const kaufman_adaptive_moving_average = @import("perry_kaufman/kaufman_adaptive_moving_average/kaufman_adaptive_moving_average.zig");
78: pub const jurik_moving_average = @import("mark_jurik/jurik_moving_average/jurik_moving_average.zig");
79: pub const stochastic = @import("george_lane/stochastic/stochastic.zig");
80: pub const adaptive_trend_and_cycle_filter = @import("vladimir_kravchuk/adaptive_trend_and_cycle_filter/adaptive_trend_and_cycle_filter.zig");
81: pub const double_exponential_moving_average = @import("patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig");
82: pub const triple_exponential_moving_average = @import("patrick_mulloy/triple_exponential_moving_average/triple_exponential_moving_average.zig");
83: pub const t2_exponential_moving_average = @import("tim_tillson/t2_exponential_moving_average/t2_exponential_moving_average.zig");
84: pub const t3_exponential_moving_average = @import("tim_tillson/t3_exponential_moving_average/t3_exponential_moving_average.zig");
85: pub const percentage_price_oscillator = @import("gerald_appel/percentage_price_oscillator/percentage_price_oscillator.zig");
86: pub const moving_average_convergence_divergence = @import("gerald_appel/moving_average_convergence_divergence/moving_average_convergence_divergence.zig");
87: pub const bollinger_bands = @import("john_bollinger/bollinger_bands/bollinger_bands.zig");
88: pub const bollinger_bands_trend = @import("john_bollinger/bollinger_bands_trend/bollinger_bands_trend.zig");
89: pub const williams_percent_r = @import("larry_williams/williams_percent_r/williams_percent_r.zig");
90: pub const ultimate_oscillator = @import("larry_williams/ultimate_oscillator/ultimate_oscillator.zig");
91: pub const advance_decline = @import("marc_chaikin/advance_decline/advance_decline.zig");
92: pub const advance_decline_oscillator = @import("marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig");
93: 
94: // Force-include tests from sub-modules.
95: comptime {
96:     _ = identifier;
97:     _ = simple_moving_average;
98:     _ = weighted_moving_average;
99:     _ = triangular_moving_average;
100:     _ = exponential_moving_average;
101:     _ = momentum;
102:     _ = rate_of_change;
103:     _ = rate_of_change_percent;
104:     _ = rate_of_change_ratio;
105:     _ = variance;
106:     _ = standard_deviation;
107:     _ = absolute_price_oscillator;
108:     _ = linear_regression;
109:     _ = pearsons_correlation_coefficient;
110:     _ = balance_of_power;
111:     _ = on_balance_volume;
112:     _ = commodity_channel_index;
113:     _ = triple_exponential_moving_average_oscillator;
114:     _ = money_flow_index;
115:     _ = kaufman_adaptive_moving_average;
116:     _ = jurik_moving_average;
117:     _ = stochastic;
118:     _ = adaptive_trend_and_cycle_filter;
119:     _ = double_exponential_moving_average;
120:     _ = triple_exponential_moving_average;
121:     _ = t2_exponential_moving_average;
122:     _ = t3_exponential_moving_average;
123:     _ = percentage_price_oscillator;
124:     _ = moving_average_convergence_divergence;
125:     _ = bollinger_bands;
126:     _ = bollinger_bands_trend;
127:     _ = williams_percent_r;
128:     _ = ultimate_oscillator;
129:     _ = advance_decline;
130:     _ = advance_decline_oscillator;
131: }

(End of file - total 131 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 6.2s)

Now let me add Aroon to the registration file:

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "oldString": "pub const advance_decline_oscillator = @import(\"marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig\");\n\n// Force-include tests from sub-modules.\ncomptime {",
  "newString": "pub const advance_decline_oscillator = @import(\"marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig\");\npub const aroon = @import(\"tushar_chande/aroon/aroon.zig\");\n\n// Force-include tests from sub-modules.\ncomptime {"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 5.1s)

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "newString": "    _ = advance_decline_oscillator;\n    _ = aroon;\n}",
  "oldString": "    _ = advance_decline_oscillator;\n}"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 6.5s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | tail -40",
  "description": "Build and test Zig indicators",
  "timeout": 120000
}
```

**Output:**
```
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 2 pass (2 total) 6ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 42 pass (42 total) 25ms MaxRSS:11M
|  +- compile test Debug native cached 8ms MaxRSS:39M
+- run test success 5ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test success 5ms MaxRSS:10M
|  +- compile test Debug native cached 8ms MaxRSS:39M
+- run test success 10ms MaxRSS:10M
|  +- compile test Debug native cached 8ms MaxRSS:39M
+- run test success 21ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 55 pass (55 total) 10ms MaxRSS:11M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 155 pass (155 total) 47ms MaxRSS:10M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 4 pass (4 total) 10ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:37M
+- run test 2 pass (2 total) 9ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 2 pass (2 total) 6ms MaxRSS:10M
|  +- compile test Debug native cached 4ms MaxRSS:39M
+- run test 6 pass (6 total) 7ms MaxRSS:11M
|  +- compile test Debug native cached 9ms MaxRSS:39M
+- run test 7 pass (7 total) 6ms MaxRSS:11M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test success 5ms MaxRSS:11M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test success 6ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 11 pass (11 total) 10ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 10 pass (10 total) 8ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:37M
+- run test 4 pass (4 total) 6ms MaxRSS:10M
|  +- compile test Debug native cached 10ms MaxRSS:39M
+- run test 289 pass (289 total) 44ms MaxRSS:11M
   +- compile test Debug native success 837ms MaxRSS:177M


```

---

## Assistant (Build · claude-opus-4.6 · 5.3s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | grep -E \"^\\d+ pass|FAIL|error:|total\"",
  "description": "Count total tests"
}
```

**Output:**
```
+- run test 3 pass (3 total) 6ms MaxRSS:10M
+- run test 53 pass (53 total) 10ms MaxRSS:11M
+- run test 11 pass (11 total) 9ms MaxRSS:10M
+- run test 2 pass (2 total) 5ms MaxRSS:11M
+- run test 42 pass (42 total) 27ms MaxRSS:10M
+- run test 55 pass (55 total) 9ms MaxRSS:10M
+- run test 155 pass (155 total) 39ms MaxRSS:10M
+- run test 4 pass (4 total) 6ms MaxRSS:11M
+- run test 2 pass (2 total) 7ms MaxRSS:10M
+- run test 2 pass (2 total) 18ms MaxRSS:10M
+- run test 6 pass (6 total) 6ms MaxRSS:10M
+- run test 7 pass (7 total) 15ms MaxRSS:10M
+- run test 11 pass (11 total) 6ms MaxRSS:11M
+- run test 10 pass (10 total) 6ms MaxRSS:11M
+- run test 4 pass (4 total) 6ms MaxRSS:10M
+- run test 289 pass (289 total) 49ms MaxRSS:11M

```

---

## Assistant (Build · claude-opus-4.6 · 5.3s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | grep -oP '\\d+ pass' | awk -F' ' '{s+=$1} END{print s}'",
  "description": "Sum total tests"
}
```

**Output:**
```
656

```

---

## Assistant (Build · claude-opus-4.6 · 7.7s)

Aroon compiles and all 656 tests pass (up from 650 — 4 new Aroon tests + 2 others). Now let me write CMO.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Register Aroon in indicators.zig and compile/test",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Write CMO implementation",
      "priority": "high",
      "status": "in_progress"
    },
    {
      "content": "Register CMO in indicators.zig and compile/test",
      "priority": "high",
      "status": "pending"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Register Aroon in indicators.zig and compile/test",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Write CMO implementation",
    "priority": "high",
    "status": "in_progress"
  },
  {
    "content": "Register CMO in indicators.zig and compile/test",
    "priority": "high",
    "status": "pending"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 7.5s)

Let me read the Go reference and a similar Zig indicator (momentum with ring buffer) for patterns:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator.go</path>
<type>file</type>
<content>1: package chandemomentumoscillator
2: 
3: //nolint: gofumpt
4: import (
5: 	"fmt"
6: 	"math"
7: 	"sync"
8: 
9: 	"zpano/entities"
10: 	"zpano/indicators/core"
11: )
12: 
13: const epsilon = 1e-12
14: 
15: // ChandeMomentumOscillator is a momentum indicator based on the average
16: // of up samples and down samples over a specified length ℓ.
17: //
18: // The calculation formula is:
19: //
20: // CMOᵢ = 100 (SUᵢ-SDᵢ) / (SUᵢ + SDᵢ),
21: //
22: // where SUᵢ (sum up) is the sum of gains and SDᵢ (sum down)
23: // is the sum of losses over the chosen length [i-ℓ, i].
24: //
25: // The indicator is not primed during the first ℓ updates.
26: type ChandeMomentumOscillator struct {
27: 	mu sync.RWMutex
28: 	core.LineIndicator
29: 	length         int
30: 	count          int
31: 	ringBuffer     []float64
32: 	ringHead       int
33: 	previousSample float64
34: 	gainSum        float64
35: 	lossSum        float64
36: 	primed         bool
37: }
38: 
39: // New returns an instance of the indicator created using supplied parameters.
40: func New(p *Params) (*ChandeMomentumOscillator, error) {
41: 	const (
42: 		invalid = "invalid Chande momentum oscillator parameters"
43: 		fmts    = "%s: %s"
44: 		fmtw    = "%s: %w"
45: 		fmtn    = "cmo(%d%s)"
46: 		minlen  = 1
47: 	)
48: 
49: 	length := p.Length
50: 	if length < minlen {
51: 		return nil, fmt.Errorf(fmts, invalid, "length should be positive")
52: 	}
53: 
54: 	var (
55: 		err       error
56: 		barFunc   entities.BarFunc
57: 		quoteFunc entities.QuoteFunc
58: 		tradeFunc entities.TradeFunc
59: 	)
60: 
61: 	// Resolve defaults for component functions.
62: 	// A zero value means "use default, don't show in mnemonic".
63: 	bc := p.BarComponent
64: 	qc := p.QuoteComponent
65: 	tc := p.TradeComponent
66: 
67: 	if bc == 0 {
68: 		bc = entities.DefaultBarComponent
69: 	}
70: 
71: 	if qc == 0 {
72: 		qc = entities.DefaultQuoteComponent
73: 	}
74: 
75: 	if tc == 0 {
76: 		tc = entities.DefaultTradeComponent
77: 	}
78: 
79: 	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
80: 		return nil, fmt.Errorf(fmtw, invalid, err)
81: 	}
82: 
83: 	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
84: 		return nil, fmt.Errorf(fmtw, invalid, err)
85: 	}
86: 
87: 	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
88: 		return nil, fmt.Errorf(fmtw, invalid, err)
89: 	}
90: 
91: 	mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
92: 	desc := "Chande Momentum Oscillator " + mnemonic
93: 
94: 	c := &ChandeMomentumOscillator{
95: 		length:     length,
96: 		ringBuffer: make([]float64, length),
97: 	}
98: 
99: 	c.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, c.Update)
100: 
101: 	return c, nil
102: }
103: 
104: // IsPrimed indicates whether an indicator is primed.
105: func (s *ChandeMomentumOscillator) IsPrimed() bool {
106: 	s.mu.RLock()
107: 	defer s.mu.RUnlock()
108: 
109: 	return s.primed
110: }
111: 
112: // Metadata describes an output data of the indicator.
113: // It always has a single scalar output -- the calculated value of the indicator.
114: func (s *ChandeMomentumOscillator) Metadata() core.Metadata {
115: 	return core.BuildMetadata(
116: 		core.ChandeMomentumOscillator,
117: 		s.LineIndicator.Mnemonic,
118: 		s.LineIndicator.Description,
119: 		[]core.OutputText{
120: 			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
121: 		},
122: 	)
123: }
124: 
125: // Update updates the value of the Chande momentum oscillator given the next sample.
126: //
127: // The indicator is not primed during the first ℓ updates.
128: func (s *ChandeMomentumOscillator) Update(sample float64) float64 {
129: 	if math.IsNaN(sample) {
130: 		return sample
131: 	}
132: 
133: 	s.mu.Lock()
134: 	defer s.mu.Unlock()
135: 
136: 	s.count++
137: 	if s.count == 1 {
138: 		s.previousSample = sample
139: 
140: 		return math.NaN()
141: 	}
142: 
143: 	// New delta
144: 	delta := sample - s.previousSample
145: 	s.previousSample = sample
146: 
147: 	if !s.primed {
148: 		// Fill until we have s.length deltas (i.e., s.length+1 samples)
149: 		s.ringBuffer[s.ringHead] = delta
150: 		s.ringHead = (s.ringHead + 1) % s.length
151: 
152: 		if delta > 0 {
153: 			s.gainSum += delta
154: 		} else if delta < 0 {
155: 			s.lossSum += -delta
156: 		}
157: 
158: 		if s.count <= s.length {
159: 			return math.NaN()
160: 		}
161: 
162: 		// Now we have exactly s.length deltas in the buffer
163: 		s.primed = true
164: 	} else {
165: 		// Remove oldest delta and add the new one
166: 		old := s.ringBuffer[s.ringHead]
167: 		if old > 0 {
168: 			s.gainSum -= old
169: 		} else if old < 0 {
170: 			s.lossSum -= -old
171: 		}
172: 
173: 		s.ringBuffer[s.ringHead] = delta
174: 		s.ringHead = (s.ringHead + 1) % s.length
175: 
176: 		if delta > 0 {
177: 			s.gainSum += delta
178: 		} else if delta < 0 {
179: 			s.lossSum += -delta
180: 		}
181: 
182: 		// Clamp to avoid tiny negative sums from FP noise
183: 		if s.gainSum < 0 {
184: 			s.gainSum = 0
185: 		}
186: 
187: 		if s.lossSum < 0 {
188: 			s.lossSum = 0
189: 		}
190: 	}
191: 
192: 	den := s.gainSum + s.lossSum
193: 	if math.Abs(den) < epsilon {
194: 		return 0
195: 	}
196: 
197: 	return 100.0 * (s.gainSum - s.lossSum) / den
198: }

(End of file - total 198 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/chandemomentumoscillator_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package chandemomentumoscillator
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: //nolint:lll
16: // Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
17: //    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
18: //
19: // TA Lib uses incorrect CMO calculation which is incompatible with Chande's book.
20: // TA Lib smooths the CMO values in the same way as RSI does, but Chande didn't smooth the values.
21: // We don't use TA Lib test data here.
22: 
23: func testChandeMomentumOscillatorTime() time.Time {
24: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
25: }
26: 
27: //nolint:dupl
28: func testChandeMomentumOscillatorBookLength10Input() []float64 {
29: 	// Chande, T.S. and Kroll, S.
30: 	// The New Technical Trader: Boost Your Profit by Plugging Into the Latest Indicators
31: 	// ISBN 9780471597803
32: 	// Wiley Finance, 1994
33: 	// Page 96, Table 5.1 Calculating 10-day CMO Using T-Bond06/93 Data
34: 	return []float64{
35: 		101.0313, 101.0313, 101.1250, 101.9687, 102.7813,
36: 		103.0000, 102.9687, 103.0625, 102.9375, 102.7188,
37: 		102.7500, 102.9063, 102.9687,
38: 	}
39: }
40: 
41: //nolint:dupl
42: func testChandeMomentumOscillatorBookLength10Output() []float64 {
43: 	return []float64{
44: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
45: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
46: 
47: 		// 69.62, 71.43, 71.08,
48: 		69.61963786608334, 71.42857142857143, 71.08377992828775,
49: 	}
50: }
51: 
52: //nolint:dupl
53: func testChandeMomentumOscillatorInput() []float64 {
54: 	return []float64{
55: 		91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
56: 		96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
57: 		88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
58: 		85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
59: 		83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
60: 		89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
61: 		88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
62: 		88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
63: 		102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
64: 		112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
65: 		110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
66: 		116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
67: 		124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
68: 		132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
69: 		136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
70: 		125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
71: 		123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
72: 		122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
73: 		132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
74: 		130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
75: 		117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
76: 		107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
77: 		94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
78: 		95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
79: 		105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
80: 		113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
81: 		108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
82: 	}
83: }
84: 
85: func TestChandeMomentumOscillatorUpdate(t *testing.T) { //nolint: funlen
86: 	t.Parallel()
87: 
88: 	check := func(index int, exp, act float64) {
89: 		t.Helper()
90: 
91: 		if math.Abs(exp-act) > 1e-13 {
92: 			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
93: 		}
94: 	}
95: 
96: 	checkNaN := func(index int, act float64) {
97: 		t.Helper()
98: 
99: 		if !math.IsNaN(act) {
100: 			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
101: 		}
102: 	}
103: 
104: 	t.Run("length = 10 (book)", func(t *testing.T) {
105: 		t.Parallel()
106: 		input := testChandeMomentumOscillatorBookLength10Input()
107: 		output := testChandeMomentumOscillatorBookLength10Output()
108: 		cmo := testChandeMomentumOscillatorCreate(10)
109: 
110: 		for i := 0; i < 10; i++ {
111: 			checkNaN(i, cmo.Update(input[i]))
112: 		}
113: 
114: 		for i := 10; i < len(input); i++ {
115: 			act := cmo.Update(input[i])
116: 			check(i, output[i], act)
117: 		}
118: 
119: 		checkNaN(0, cmo.Update(math.NaN()))
120: 	})
121: }
122: 
123: func TestChandeMomentumOscillatorUpdateEntity(t *testing.T) { //nolint: funlen
124: 	t.Parallel()
125: 
126: 	const (
127: 		l   = 2
128: 		inp = 3.
129: 		exp = 100.
130: 	)
131: 
132: 	time := testChandeMomentumOscillatorTime()
133: 	check := func(act core.Output) {
134: 		t.Helper()
135: 
136: 		if len(act) != 1 {
137: 			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
138: 		}
139: 
140: 		s, ok := act[0].(entities.Scalar)
141: 		if !ok {
142: 			t.Error("output is not scalar")
143: 		}
144: 
145: 		if s.Time != time {
146: 			t.Errorf("time is incorrect: expected %v, actual %v", time, s.Time)
147: 		}
148: 
149: 		if math.Abs(s.Value-exp) > 1e-13 {
150: 			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
151: 		}
152: 	}
153: 
154: 	t.Run("update scalar", func(t *testing.T) {
155: 		t.Parallel()
156: 
157: 		s := entities.Scalar{Time: time, Value: inp}
158: 		cmo := testChandeMomentumOscillatorCreate(l)
159: 		cmo.Update(0.)
160: 		cmo.Update(0.)
161: 		check(cmo.UpdateScalar(&s))
162: 	})
163: 
164: 	t.Run("update bar", func(t *testing.T) {
165: 		t.Parallel()
166: 
167: 		b := entities.Bar{Time: time, Close: inp}
168: 		cmo := testChandeMomentumOscillatorCreate(l)
169: 		cmo.Update(0.)
170: 		cmo.Update(0.)
171: 		check(cmo.UpdateBar(&b))
172: 	})
173: 
174: 	t.Run("update quote", func(t *testing.T) {
175: 		t.Parallel()
176: 
177: 		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
178: 		cmo := testChandeMomentumOscillatorCreate(l)
179: 		cmo.Update(0.)
180: 		cmo.Update(0.)
181: 		check(cmo.UpdateQuote(&q))
182: 	})
183: 
184: 	t.Run("update trade", func(t *testing.T) {
185: 		t.Parallel()
186: 
187: 		r := entities.Trade{Time: time, Price: inp}
188: 		cmo := testChandeMomentumOscillatorCreate(l)
189: 		cmo.Update(0.)
190: 		cmo.Update(0.)
191: 		check(cmo.UpdateTrade(&r))
192: 	})
193: }
194: 
195: func TestChandeMomentumOscillatorIsPrimed(t *testing.T) { //nolint:funlen
196: 	t.Parallel()
197: 
198: 	input := testChandeMomentumOscillatorInput()
199: 	check := func(index int, exp, act bool) {
200: 		t.Helper()
201: 
202: 		if exp != act {
203: 			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
204: 		}
205: 	}
206: 
207: 	t.Run("length = 1", func(t *testing.T) {
208: 		t.Parallel()
209: 		cmo := testChandeMomentumOscillatorCreate(1)
210: 
211: 		check(-1, false, cmo.IsPrimed())
212: 
213: 		for i := 0; i < 1; i++ {
214: 			cmo.Update(input[i])
215: 			check(i, false, cmo.IsPrimed())
216: 		}
217: 
218: 		for i := 1; i < len(input); i++ {
219: 			cmo.Update(input[i])
220: 			check(i, true, cmo.IsPrimed())
221: 		}
222: 	})
223: 
224: 	t.Run("length = 2", func(t *testing.T) {
225: 		t.Parallel()
226: 		cmo := testChandeMomentumOscillatorCreate(2)
227: 
228: 		check(-1, false, cmo.IsPrimed())
229: 
230: 		for i := 0; i < 2; i++ {
231: 			cmo.Update(input[i])
232: 			check(i, false, cmo.IsPrimed())
233: 		}
234: 
235: 		for i := 2; i < len(input); i++ {
236: 			cmo.Update(input[i])
237: 			check(i, true, cmo.IsPrimed())
238: 		}
239: 	})
240: 
241: 	t.Run("length = 3", func(t *testing.T) {
242: 		t.Parallel()
243: 		cmo := testChandeMomentumOscillatorCreate(3)
244: 
245: 		check(-1, false, cmo.IsPrimed())
246: 
247: 		for i := 0; i < 3; i++ {
248: 			cmo.Update(input[i])
249: 			check(i, false, cmo.IsPrimed())
250: 		}
251: 
252: 		for i := 3; i < len(input); i++ {
253: 			cmo.Update(input[i])
254: 			check(i, true, cmo.IsPrimed())
255: 		}
256: 	})
257: 
258: 	t.Run("length = 5", func(t *testing.T) {
259: 		t.Parallel()
260: 		cmo := testChandeMomentumOscillatorCreate(5)
261: 
262: 		check(-1, false, cmo.IsPrimed())
263: 
264: 		for i := 0; i < 5; i++ {
265: 			cmo.Update(input[i])
266: 			check(i, false, cmo.IsPrimed())
267: 		}
268: 
269: 		for i := 5; i < len(input); i++ {
270: 			cmo.Update(input[i])
271: 			check(i, true, cmo.IsPrimed())
272: 		}
273: 	})
274: 
275: 	t.Run("length = 10", func(t *testing.T) {
276: 		t.Parallel()
277: 		cmo := testChandeMomentumOscillatorCreate(10)
278: 
279: 		check(-1, false, cmo.IsPrimed())
280: 
281: 		for i := 0; i < 10; i++ {
282: 			cmo.Update(input[i])
283: 			check(i, false, cmo.IsPrimed())
284: 		}
285: 
286: 		for i := 10; i < len(input); i++ {
287: 			cmo.Update(input[i])
288: 			check(i, true, cmo.IsPrimed())
289: 		}
290: 	})
291: }
292: 
293: func TestChandeMomentumOscillatorMetadata(t *testing.T) {
294: 	t.Parallel()
295: 
296: 	cmo := testChandeMomentumOscillatorCreate(5)
297: 	act := cmo.Metadata()
298: 
299: 	check := func(what string, exp, act any) {
300: 		t.Helper()
301: 
302: 		if exp != act {
303: 			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
304: 		}
305: 	}
306: 
307: 	check("Identifier", core.ChandeMomentumOscillator, act.Identifier)
308: 	check("Mnemonic", "cmo(5)", act.Mnemonic)
309: 	check("Description", "Chande Momentum Oscillator cmo(5)", act.Description)
310: 	check("len(Outputs)", 1, len(act.Outputs))
311: 	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
312: 	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
313: 	check("Outputs[0].Mnemonic", "cmo(5)", act.Outputs[0].Mnemonic)
314: 	check("Outputs[0].Description", "Chande Momentum Oscillator cmo(5)", act.Outputs[0].Description)
315: }
316: 
317: func TestNewChandeMomentumOscillator(t *testing.T) { //nolint: funlen
318: 	t.Parallel()
319: 
320: 	const (
321: 		bc     entities.BarComponent   = entities.BarMedianPrice
322: 		qc     entities.QuoteComponent = entities.QuoteMidPrice
323: 		tc     entities.TradeComponent = entities.TradePrice
324: 		length                         = 5
325: 		errlen                         = "invalid Chande momentum oscillator parameters: length should be positive"
326: 		errbc                          = "invalid Chande momentum oscillator parameters: 9999: unknown bar component"
327: 		errqc                          = "invalid Chande momentum oscillator parameters: 9999: unknown quote component"
328: 		errtc                          = "invalid Chande momentum oscillator parameters: 9999: unknown trade component"
329: 	)
330: 
331: 	check := func(name string, exp, act any) {
332: 		t.Helper()
333: 
334: 		if exp != act {
335: 			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
336: 		}
337: 	}
338: 
339: 	t.Run("length > 0", func(t *testing.T) {
340: 		t.Parallel()
341: 		params := Params{
342: 			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
343: 		}
344: 
345: 		cmo, err := New(&params)
346: 		check("err == nil", true, err == nil)
347: 		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
348: 		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
349: 		check("primed", false, cmo.primed)
350: 		check("length", length, cmo.length)
351: 		check("len(ringBuffer)", length, len(cmo.ringBuffer))
352: 		check("ringHead", 0, cmo.ringHead)
353: 		check("count", 0, cmo.count)
354: 		check("previousSample", 0., cmo.previousSample)
355: 		check("gainSum", 0., cmo.gainSum)
356: 		check("lossSum", 0., cmo.lossSum)
357: 	})
358: 
359: 	t.Run("length = 1", func(t *testing.T) {
360: 		t.Parallel()
361: 		params := Params{
362: 			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
363: 		}
364: 
365: 		cmo, err := New(&params)
366: 		check("err == nil", true, err == nil)
367: 		check("mnemonic", "cmo(1, hl/2)", cmo.LineIndicator.Mnemonic)
368: 		check("description", "Chande Momentum Oscillator cmo(1, hl/2)", cmo.LineIndicator.Description)
369: 		check("primed", false, cmo.primed)
370: 		check("length", 1, cmo.length)
371: 		check("len(ringBuffer)", 1, len(cmo.ringBuffer))
372: 		check("ringHead", 0, cmo.ringHead)
373: 		check("count", 0, cmo.count)
374: 		check("previousSample", 0., cmo.previousSample)
375: 		check("gainSum", 0., cmo.gainSum)
376: 		check("lossSum", 0., cmo.lossSum)
377: 	})
378: 
379: 	t.Run("length = 0", func(t *testing.T) {
380: 		t.Parallel()
381: 		params := Params{
382: 			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
383: 		}
384: 
385: 		cmo, err := New(&params)
386: 		check("cmo == nil", true, cmo == nil)
387: 		check("err", errlen, err.Error())
388: 	})
389: 
390: 	t.Run("length < 0", func(t *testing.T) {
391: 		t.Parallel()
392: 		params := Params{
393: 			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
394: 		}
395: 
396: 		cmo, err := New(&params)
397: 		check("cmo == nil", true, cmo == nil)
398: 		check("err", errlen, err.Error())
399: 	})
400: 
401: 	t.Run("invalid bar component", func(t *testing.T) {
402: 		t.Parallel()
403: 		params := Params{
404: 			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
405: 		}
406: 
407: 		cmo, err := New(&params)
408: 		check("cmo == nil", true, cmo == nil)
409: 		check("err", errbc, err.Error())
410: 	})
411: 
412: 	t.Run("invalid quote component", func(t *testing.T) {
413: 		t.Parallel()
414: 		params := Params{
415: 			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
416: 		}
417: 
418: 		cmo, err := New(&params)
419: 		check("cmo == nil", true, cmo == nil)
420: 		check("err", errqc, err.Error())
421: 	})
422: 
423: 	t.Run("invalid trade component", func(t *testing.T) {
424: 		t.Parallel()
425: 		params := Params{
426: 			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
427: 		}
428: 
429: 		cmo, err := New(&params)
430: 		check("cmo == nil", true, cmo == nil)
431: 		check("err", errtc, err.Error())
432: 	})
433: 
434: 	// Zero-value component mnemonic tests.
435: 	// A zero value means "use default, don't show in mnemonic".
436: 
437: 	t.Run("all components zero", func(t *testing.T) {
438: 		t.Parallel()
439: 		params := Params{Length: length}
440: 
441: 		cmo, err := New(&params)
442: 		check("err == nil", true, err == nil)
443: 		check("mnemonic", "cmo(5)", cmo.LineIndicator.Mnemonic)
444: 		check("description", "Chande Momentum Oscillator cmo(5)", cmo.LineIndicator.Description)
445: 	})
446: 
447: 	t.Run("only bar component set", func(t *testing.T) {
448: 		t.Parallel()
449: 		params := Params{Length: length, BarComponent: entities.BarMedianPrice}
450: 
451: 		cmo, err := New(&params)
452: 		check("err == nil", true, err == nil)
453: 		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
454: 		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
455: 	})
456: 
457: 	t.Run("only quote component set", func(t *testing.T) {
458: 		t.Parallel()
459: 		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}
460: 
461: 		cmo, err := New(&params)
462: 		check("err == nil", true, err == nil)
463: 		check("mnemonic", "cmo(5, b)", cmo.LineIndicator.Mnemonic)
464: 		check("description", "Chande Momentum Oscillator cmo(5, b)", cmo.LineIndicator.Description)
465: 	})
466: 
467: 	t.Run("only trade component set", func(t *testing.T) {
468: 		t.Parallel()
469: 		params := Params{Length: length, TradeComponent: entities.TradeVolume}
470: 
471: 		cmo, err := New(&params)
472: 		check("err == nil", true, err == nil)
473: 		check("mnemonic", "cmo(5, v)", cmo.LineIndicator.Mnemonic)
474: 		check("description", "Chande Momentum Oscillator cmo(5, v)", cmo.LineIndicator.Description)
475: 	})
476: }
477: 
478: func testChandeMomentumOscillatorCreate(length int) *ChandeMomentumOscillator {
479: 	params := Params{Length: length}
480: 
481: 	cmo, _ := New(&params)
482: 
483: 	return cmo
484: }

(End of file - total 484 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/params.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/params.go</path>
<type>file</type>
<content>1: package chandemomentumoscillator
2: 
3: import "zpano/entities"
4: 
5: // Params describes parameters to create an instance of the indicator.
6: type Params struct {
7: 	// Length is the length (the number of time periods, ℓ) which defines the window of price changes.
8: 	//
9: 	// The value should be greater than 0.
10: 	Length int
11: 
12: 	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
13: 	//
14: 	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
15: 	BarComponent entities.BarComponent
16: 
17: 	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
18: 	//
19: 	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
20: 	QuoteComponent entities.QuoteComponent
21: 
22: 	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
23: 	//
24: 	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
25: 	TradeComponent entities.TradeComponent
26: }
27: 
28: // DefaultParams returns a [Params] value populated with conventional defaults.
29: func DefaultParams() *Params {
30: 	return &Params{
31: 		Length: 14,
32: 	}
33: }

(End of file - total 33 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/common/momentum/momentum.zig"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/common/momentum/momentum.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: const bar_component = @import("bar_component");
9: const quote_component = @import("quote_component");
10: const trade_component = @import("trade_component");
11: 
12: const indicator_mod = @import("../../core/indicator.zig");
13: const line_indicator_mod = @import("../../core/line_indicator.zig");
14: const build_metadata_mod = @import("../../core/build_metadata.zig");
15: const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
16: const identifier_mod = @import("../../core/identifier.zig");
17: const metadata_mod = @import("../../core/metadata.zig");
18: 
19: const OutputArray = indicator_mod.OutputArray;
20: const LineIndicator = line_indicator_mod.LineIndicator;
21: const Identifier = identifier_mod.Identifier;
22: const Metadata = metadata_mod.Metadata;
23: 
24: /// Enumerates the outputs of the momentum indicator.
25: pub const MomentumOutput = enum(u8) {
26:     /// The scalar value of the momentum.
27:     value = 1,
28: };
29: 
30: /// Parameters to create an instance of the momentum indicator.
31: pub const MomentumParams = struct {
32:     /// The length (number of time periods). Must be >= 1.
33:     length: usize,
34:     /// Bar component to extract. `null` means use default (Close).
35:     bar_component: ?bar_component.BarComponent = null,
36:     /// Quote component to extract. `null` means use default (Mid).
37:     quote_component: ?quote_component.QuoteComponent = null,
38:     /// Trade component to extract. `null` means use default (Price).
39:     trade_component: ?trade_component.TradeComponent = null,
40: };
41: 
42: /// Computes the momentum (MOM).
43: ///
44: /// MOMi = Pi - Pi-l
45: ///
46: /// where l is the length.
47: ///
48: /// The indicator is not primed during the first l updates.
49: pub const Momentum = struct {
50:     line: LineIndicator,
51:     window: []f64,
52:     window_length: usize,
53:     window_count: usize,
54:     last_index: usize,
55:     primed: bool,
56:     allocator: std.mem.Allocator,
57:     mnemonic_buf: [64]u8,
58:     mnemonic_len: usize,
59:     description_buf: [128]u8,
60:     description_len: usize,
61: 
62:     pub fn init(allocator: std.mem.Allocator, params: MomentumParams) !Momentum {
63:         if (params.length < 1) {
64:             return error.InvalidLength;
65:         }
66: 
67:         const bc = params.bar_component orelse bar_component.default_bar_component;
68:         const qc = params.quote_component orelse quote_component.default_quote_component;
69:         const tc = params.trade_component orelse trade_component.default_trade_component;
70: 
71:         var triple_buf: [64]u8 = undefined;
72:         const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);
73: 
74:         var mnemonic_buf: [64]u8 = undefined;
75:         const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "mom({d}{s})", .{ params.length, triple }) catch
76:             return error.MnemonicTooLong;
77:         const mnemonic_len = mnemonic_slice.len;
78: 
79:         var description_buf: [128]u8 = undefined;
80:         const desc_slice = std.fmt.bufPrint(&description_buf, "Momentum {s}", .{mnemonic_slice}) catch
81:             return error.MnemonicTooLong;
82:         const description_len = desc_slice.len;
83: 
84:         const window_length = params.length + 1;
85:         const window = try allocator.alloc(f64, window_length);
86:         @memset(window, 0.0);
87: 
88:         return .{
89:             .line = LineIndicator.new(
90:                 mnemonic_buf[0..mnemonic_len],
91:                 description_buf[0..description_len],
92:                 params.bar_component,
93:                 params.quote_component,
94:                 params.trade_component,
95:             ),
96:             .window = window,
97:             .window_length = window_length,
98:             .window_count = 0,
99:             .last_index = params.length,
100:             .primed = false,
101:             .allocator = allocator,
102:             .mnemonic_buf = mnemonic_buf,
103:             .mnemonic_len = mnemonic_len,
104:             .description_buf = description_buf,
105:             .description_len = description_len,
106:         };
107:     }
108: 
109:     pub fn deinit(self: *Momentum) void {
110:         self.allocator.free(self.window);
111:     }
112: 
113:     fn fixSlices(self: *Momentum) void {
114:         self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
115:         self.line.description = self.description_buf[0..self.description_len];
116:     }
117: 
118:     /// Core update logic. Returns the momentum value or NaN if not yet primed.
119:     pub fn update(self: *Momentum, sample: f64) f64 {
120:         if (math.isNan(sample)) {
121:             return sample;
122:         }
123: 
124:         if (self.primed) {
125:             var i: usize = 0;
126:             while (i < self.last_index) : (i += 1) {
127:                 self.window[i] = self.window[i + 1];
128:             }
129: 
130:             self.window[self.last_index] = sample;
131: 
132:             return sample - self.window[0];
133:         }
134: 
135:         self.window[self.window_count] = sample;
136:         self.window_count += 1;
137: 
138:         if (self.window_length == self.window_count) {
139:             self.primed = true;
140: 
141:             return sample - self.window[0];
142:         }
143: 
144:         return math.nan(f64);
145:     }
146: 
147:     pub fn isPrimed(self: *const Momentum) bool {
148:         return self.primed;
149:     }
150: 
151:     pub fn getMetadata(self: *const Momentum, out: *Metadata) void {
152:         build_metadata_mod.buildMetadata(
153:             out,
154:             .momentum,
155:             self.line.mnemonic,
156:             self.line.description,
157:             &[_]build_metadata_mod.OutputText{
158:                 .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
159:             },
160:         );
161:     }
162: 
163:     pub fn updateScalar(self: *Momentum, sample: *const Scalar) OutputArray {
164:         const value = self.update(sample.value);
165:         return LineIndicator.wrapScalar(sample.time, value);
166:     }
167: 
168:     pub fn updateBar(self: *Momentum, sample: *const Bar) OutputArray {
169:         const value = self.update(self.line.extractBar(sample));
170:         return LineIndicator.wrapScalar(sample.time, value);
171:     }
172: 
173:     pub fn updateQuote(self: *Momentum, sample: *const Quote) OutputArray {
174:         const value = self.update(self.line.extractQuote(sample));
175:         return LineIndicator.wrapScalar(sample.time, value);
176:     }
177: 
178:     pub fn updateTrade(self: *Momentum, sample: *const Trade) OutputArray {
179:         const value = self.update(self.line.extractTrade(sample));
180:         return LineIndicator.wrapScalar(sample.time, value);
181:     }
182: 
183:     pub fn indicator(self: *Momentum) indicator_mod.Indicator {
184:         return .{
185:             .ptr = @ptrCast(self),
186:             .vtable = &vtable,
187:         };
188:     }
189: 
190:     const vtable = indicator_mod.Indicator.VTable{
191:         .isPrimed = vtableIsPrimed,
192:         .metadata = vtableMetadata,
193:         .updateScalar = vtableUpdateScalar,
194:         .updateBar = vtableUpdateBar,
195:         .updateQuote = vtableUpdateQuote,
196:         .updateTrade = vtableUpdateTrade,
197:     };
198: 
199:     fn vtableIsPrimed(ptr: *anyopaque) bool {
200:         const self: *Momentum = @ptrCast(@alignCast(ptr));
201:         return self.isPrimed();
202:     }
203: 
204:     fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
205:         const self: *const Momentum = @ptrCast(@alignCast(ptr));
206:         self.getMetadata(out);
207:     }
208: 
209:     fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
210:         const self: *Momentum = @ptrCast(@alignCast(ptr));
211:         return self.updateScalar(sample);
212:     }
213: 
214:     fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
215:         const self: *Momentum = @ptrCast(@alignCast(ptr));
216:         return self.updateBar(sample);
217:     }
218: 
219:     fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
220:         const self: *Momentum = @ptrCast(@alignCast(ptr));
221:         return self.updateQuote(sample);
222:     }
223: 
224:     fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
225:         const self: *Momentum = @ptrCast(@alignCast(ptr));
226:         return self.updateTrade(sample);
227:     }
228: 
229:     pub const InitError = error{
230:         InvalidLength,
231:         MnemonicTooLong,
232:         OutOfMemory,
233:     };
234: };
235: 
236: // ---------------------------------------------------------------------------
237: // Tests
238: // ---------------------------------------------------------------------------
239: 
240: const testing = std.testing;
241: 
242: fn testInput() [252]f64 {
243:     return .{
244:         91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,
245:         96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,
246:         88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,
247:         85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,
248:         83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,
249:         89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,
250:         88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,
251:         88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,
252:         102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,
253:         123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,
254:         116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
255:         114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,
256:         123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,
257:         137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,
258:         123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,
259:         122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,
260:         124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
261:         132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,
262:         127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,
263:         122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,
264:         107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,
265:         94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
266:         94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,
267:         103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,
268:         107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,
269:         108.750000, 107.870000,
270:     };
271: }
272: 
273: fn createMomentum(allocator: std.mem.Allocator, length: usize) !Momentum {
274:     var mom = try Momentum.init(allocator, .{ .length = length });
275:     mom.fixSlices();
276:     return mom;
277: }
278: 
279: test "momentum update length 14" {
280:     const input = testInput();
281:     var mom = try createMomentum(testing.allocator, 14);
282:     defer mom.deinit();
283: 
284:     // First 13 updates (index 0..12) produce NaN.
285:     for (0..13) |i| {
286:         const act = mom.update(input[i]);
287:         try testing.expect(math.isNan(act));
288:     }
289: 
290:     // From index 13 onward, primed.
291:     for (13..252) |i| {
292:         const act = mom.update(input[i]);
293: 
294:         if (i == 14) try testing.expect(@abs(act - (-0.50)) < 1e-13);
295:         if (i == 15) try testing.expect(@abs(act - (-2.00)) < 1e-13);
296:         if (i == 16) try testing.expect(@abs(act - (-5.22)) < 1e-13);
297:         if (i == 251) try testing.expect(@abs(act - (-1.13)) < 1e-13);
298:     }
299: 
300:     // NaN passthrough
301:     try testing.expect(math.isNan(mom.update(math.nan(f64))));
302: }
303: 
304: test "momentum is primed" {
305:     const input = testInput();
306: 
307:     inline for ([_]usize{ 1, 2, 3, 5, 10 }) |length| {
308:         var mom = try createMomentum(testing.allocator, length);
309:         defer mom.deinit();
310: 
311:         try testing.expect(!mom.isPrimed());
312: 
313:         for (0..length) |i| {
314:             _ = mom.update(input[i]);
315:             try testing.expect(!mom.isPrimed());
316:         }
317: 
318:         for (length..252) |i| {
319:             _ = mom.update(input[i]);
320:             try testing.expect(mom.isPrimed());
321:         }
322:     }
323: }
324: 
325: test "momentum metadata" {
326:     var mom = try createMomentum(testing.allocator, 5);
327:     defer mom.deinit();
328:     var m: Metadata = undefined;
329:     mom.getMetadata(&m);
330: 
331:     try testing.expectEqual(Identifier.momentum, m.identifier);
332:     try testing.expectEqual(@as(usize, 1), m.outputs_len);
333:     try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
334:     try testing.expectEqualStrings("mom(5)", m.outputs_buf[0].mnemonic);
335:     try testing.expectEqualStrings("Momentum mom(5)", m.outputs_buf[0].description);
336: }
337: 
338: test "momentum update entity" {
339:     const length: usize = 2;
340:     const inp: f64 = 3.0;
341:     const exp: f64 = 3.0; // mom = 3.0 - 0.0 = 3.0
342:     const time: i64 = 1617235200;
343: 
344:     // scalar
345:     {
346:         var mom = try createMomentum(testing.allocator, length);
347:         defer mom.deinit();
348:         _ = mom.update(0.0);
349:         _ = mom.update(0.0);
350:         const out = mom.updateScalar(&.{ .time = time, .value = inp });
351:         try testing.expectEqual(@as(usize, 1), out.len);
352:         const s = out.slice()[0].scalar;
353:         try testing.expectEqual(time, s.time);
354:         try testing.expectEqual(exp, s.value);
355:     }
356: 
357:     // bar
358:     {
359:         var mom = try createMomentum(testing.allocator, length);
360:         defer mom.deinit();
361:         _ = mom.update(0.0);
362:         _ = mom.update(0.0);
363:         const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
364:         const out = mom.updateBar(&bar);
365:         const s = out.slice()[0].scalar;
366:         try testing.expectEqual(exp, s.value);
367:     }
368: 
369:     // quote
370:     {
371:         var mom = try createMomentum(testing.allocator, length);
372:         defer mom.deinit();
373:         _ = mom.update(0.0);
374:         _ = mom.update(0.0);
375:         const quote = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
376:         const out = mom.updateQuote(&quote);
377:         const s = out.slice()[0].scalar;
378:         try testing.expectEqual(exp, s.value);
379:     }
380: 
381:     // trade
382:     {
383:         var mom = try createMomentum(testing.allocator, length);
384:         defer mom.deinit();
385:         _ = mom.update(0.0);
386:         _ = mom.update(0.0);
387:         const trade = Trade{ .time = time, .price = inp, .volume = 0 };
388:         const out = mom.updateTrade(&trade);
389:         const s = out.slice()[0].scalar;
390:         try testing.expectEqual(exp, s.value);
391:     }
392: }
393: 
394: test "momentum init invalid length" {
395:     const result = Momentum.init(testing.allocator, .{ .length = 0 });
396:     try testing.expectError(error.InvalidLength, result);
397: }
398: 
399: test "momentum mnemonic components" {
400:     // all defaults -> no component suffix
401:     {
402:         var mom = try createMomentum(testing.allocator, 5);
403:         defer mom.deinit();
404:         try testing.expectEqualStrings("mom(5)", mom.line.mnemonic);
405:     }
406: 
407:     // bar component set to Median
408:     {
409:         var mom = try Momentum.init(testing.allocator, .{
410:             .length = 5,
411:             .bar_component = .median,
412:         });
413:         defer mom.deinit();
414:         mom.fixSlices();
415:         try testing.expectEqualStrings("mom(5, hl/2)", mom.line.mnemonic);
416:         try testing.expectEqualStrings("Momentum mom(5, hl/2)", mom.line.description);
417:     }
418: 
419:     // bar=high, trade=volume
420:     {
421:         var mom = try Momentum.init(testing.allocator, .{
422:             .length = 5,
423:             .bar_component = .high,
424:             .trade_component = .volume,
425:         });
426:         defer mom.deinit();
427:         mom.fixSlices();
428:         try testing.expectEqualStrings("mom(5, h, v)", mom.line.mnemonic);
429:     }
430: }

(End of file - total 430 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 7.4s)

Now I have everything I need. CMO is very similar to Momentum but uses a ring buffer of deltas with gainSum/lossSum tracking. Let me also check the Go test's output enum:

**Tool: grep**

**Input:**
```json
{
  "pattern": "type.*Output.*iota",
  "path": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator"
}
```

**Output:**
```
No files found
```

---

## Assistant (Build · claude-opus-4.6 · 4.2s)

**Tool: glob**

**Input:**
```json
{
  "pattern": "**/chandemomentumoscillator/output*.go"
}
```

**Output:**
```
/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/output_test.go
/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/output.go
```

---

## Assistant (Build · claude-opus-4.6 · 4.4s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/output.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/tusharchande/chandemomentumoscillator/output.go</path>
<type>file</type>
<content>1: //nolint:dupl
2: package chandemomentumoscillator
3: 
4: import (
5: 	"bytes"
6: 	"fmt"
7: )
8: 
9: // Output describes the outputs of the indicator.
10: type Output int
11: 
12: const (
13: 	// The scalar value of the Chande momentum oscillator.
14: 	Value Output = iota + 1
15: 	last
16: )
17: 
18: const (
19: 	valueStr   = "value"
20: 	unknownStr = "unknown"
21: )
22: 
23: // String implements the Stringer interface.
24: func (o Output) String() string {
25: 	switch o {
26: 	case Value:
27: 		return valueStr
28: 	default:
29: 		return unknownStr
30: 	}
31: }
32: 
33: // IsKnown determines if this output is known.
34: func (o Output) IsKnown() bool {
35: 	return o >= Value && o < last
36: }
37: 
38: // MarshalJSON implements the Marshaler interface.
39: func (o Output) MarshalJSON() ([]byte, error) {
40: 	const (
41: 		errFmt = "cannot marshal '%s': unknown Chande momentum oscillator output"
42: 		extra  = 2   // Two bytes for quotes.
43: 		dqc    = '"' // Double quote character.
44: 	)
45: 
46: 	s := o.String()
47: 	if s == unknownStr {
48: 		return nil, fmt.Errorf(errFmt, s)
49: 	}
50: 
51: 	b := make([]byte, 0, len(s)+extra)
52: 	b = append(b, dqc)
53: 	b = append(b, s...)
54: 	b = append(b, dqc)
55: 
56: 	return b, nil
57: }
58: 
59: // UnmarshalJSON implements the Unmarshaler interface.
60: func (o *Output) UnmarshalJSON(data []byte) error {
61: 	const (
62: 		errFmt = "cannot unmarshal '%s': unknown Chande momentum oscillator output"
63: 		dqs    = "\"" // Double quote string.
64: 	)
65: 
66: 	d := bytes.Trim(data, dqs)
67: 	s := string(d)
68: 
69: 	switch s {
70: 	case valueStr:
71: 		*o = Value
72: 	default:
73: 		return fmt.Errorf(errFmt, s)
74: 	}
75: 
76: 	return nil
77: }

(End of file - total 77 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 86.6s)

Good, single output `Value = iota + 1` (so 1-based). Now let me write the CMO:

**Tool: write**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig",
  "content": "const std = @import(\"std\");\nconst math = std.math;\n\nconst Bar = @import(\"bar\").Bar;\nconst Quote = @import(\"quote\").Quote;\nconst Trade = @import(\"trade\").Trade;\nconst Scalar = @import(\"scalar\").Scalar;\nconst bar_component = @import(\"bar_component\");\nconst quote_component = @import(\"quote_component\");\nconst trade_component = @import(\"trade_component\");\n\nconst indicator_mod = @import(\"../../core/indicator.zig\");\nconst line_indicator_mod = @import(\"../../core/line_indicator.zig\");\nconst build_metadata_mod = @import(\"../../core/build_metadata.zig\");\nconst component_triple_mnemonic_mod = @import(\"../../core/component_triple_mnemonic.zig\");\nconst identifier_mod = @import(\"../../core/identifier.zig\");\nconst metadata_mod = @import(\"../../core/metadata.zig\");\n\nconst OutputArray = indicator_mod.OutputArray;\nconst LineIndicator = line_indicator_mod.LineIndicator;\nconst Identifier = identifier_mod.Identifier;\nconst Metadata = metadata_mod.Metadata;\n\nconst epsilon = 1e-12;\n\n/// Enumerates the outputs of the Chande Momentum Oscillator.\npub const ChandeMomentumOscillatorOutput = enum(u8) {\n    /// The scalar value of the CMO.\n    value = 1,\n};\n\n/// Parameters to create an instance of the Chande Momentum Oscillator.\npub const ChandeMomentumOscillatorParams = struct {\n    /// The length (number of time periods). Must be >= 1.\n    length: usize,\n    /// Bar component to extract. `null` means use default (Close).\n    bar_component: ?bar_component.BarComponent = null,\n    /// Quote component to extract. `null` means use default (Mid).\n    quote_component: ?quote_component.QuoteComponent = null,\n    /// Trade component to extract. `null` means use default (Price).\n    trade_component: ?trade_component.TradeComponent = null,\n};\n\n/// Chande Momentum Oscillator (CMO).\n///\n/// CMOi = 100 * (SUi - SDi) / (SUi + SDi)\n///\n/// where SUi is the sum of gains and SDi is the sum of losses\n/// over the chosen length.\n///\n/// The indicator is not primed during the first length updates.\npub const ChandeMomentumOscillator = struct {\n    line: LineIndicator,\n    ring_buffer: []f64,\n    length: usize,\n    ring_head: usize,\n    count: usize,\n    previous_sample: f64,\n    gain_sum: f64,\n    loss_sum: f64,\n    primed: bool,\n    allocator: std.mem.Allocator,\n    mnemonic_buf: [64]u8,\n    mnemonic_len: usize,\n    description_buf: [128]u8,\n    description_len: usize,\n\n    pub fn init(allocator: std.mem.Allocator, params: ChandeMomentumOscillatorParams) !ChandeMomentumOscillator {\n        if (params.length < 1) {\n            return error.InvalidLength;\n        }\n\n        const bc = params.bar_component orelse bar_component.default_bar_component;\n        const qc = params.quote_component orelse quote_component.default_quote_component;\n        const tc = params.trade_component orelse trade_component.default_trade_component;\n\n        var triple_buf: [64]u8 = undefined;\n        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);\n\n        var mnemonic_buf: [64]u8 = undefined;\n        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, \"cmo({d}{s})\", .{ params.length, triple }) catch\n            return error.MnemonicTooLong;\n        const mnemonic_len = mnemonic_slice.len;\n\n        var description_buf: [128]u8 = undefined;\n        const desc_slice = std.fmt.bufPrint(&description_buf, \"Chande Momentum Oscillator {s}\", .{mnemonic_slice}) catch\n            return error.MnemonicTooLong;\n        const description_len = desc_slice.len;\n\n        const ring_buffer = try allocator.alloc(f64, params.length);\n        @memset(ring_buffer, 0.0);\n\n        return .{\n            .line = LineIndicator.new(\n                mnemonic_buf[0..mnemonic_len],\n                description_buf[0..description_len],\n                params.bar_component,\n                params.quote_component,\n                params.trade_component,\n            ),\n            .ring_buffer = ring_buffer,\n            .length = params.length,\n            .ring_head = 0,\n            .count = 0,\n            .previous_sample = 0.0,\n            .gain_sum = 0.0,\n            .loss_sum = 0.0,\n            .primed = false,\n            .allocator = allocator,\n            .mnemonic_buf = mnemonic_buf,\n            .mnemonic_len = mnemonic_len,\n            .description_buf = description_buf,\n            .description_len = description_len,\n        };\n    }\n\n    pub fn deinit(self: *ChandeMomentumOscillator) void {\n        self.allocator.free(self.ring_buffer);\n    }\n\n    pub fn fixSlices(self: *ChandeMomentumOscillator) void {\n        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];\n        self.line.description = self.description_buf[0..self.description_len];\n    }\n\n    /// Core update logic.\n    pub fn update(self: *ChandeMomentumOscillator, sample: f64) f64 {\n        if (math.isNan(sample)) {\n            return sample;\n        }\n\n        self.count += 1;\n        if (self.count == 1) {\n            self.previous_sample = sample;\n            return math.nan(f64);\n        }\n\n        // New delta\n        const delta = sample - self.previous_sample;\n        self.previous_sample = sample;\n\n        if (!self.primed) {\n            // Fill until we have self.length deltas (i.e., self.length+1 samples)\n            self.ring_buffer[self.ring_head] = delta;\n            self.ring_head = (self.ring_head + 1) % self.length;\n\n            if (delta > 0) {\n                self.gain_sum += delta;\n            } else if (delta < 0) {\n                self.loss_sum += -delta;\n            }\n\n            if (self.count <= self.length) {\n                return math.nan(f64);\n            }\n\n            // Now we have exactly self.length deltas in the buffer\n            self.primed = true;\n        } else {\n            // Remove oldest delta and add the new one\n            const old = self.ring_buffer[self.ring_head];\n            if (old > 0) {\n                self.gain_sum -= old;\n            } else if (old < 0) {\n                self.loss_sum -= -old;\n            }\n\n            self.ring_buffer[self.ring_head] = delta;\n            self.ring_head = (self.ring_head + 1) % self.length;\n\n            if (delta > 0) {\n                self.gain_sum += delta;\n            } else if (delta < 0) {\n                self.loss_sum += -delta;\n            }\n\n            // Clamp to avoid tiny negative sums from FP noise\n            if (self.gain_sum < 0) {\n                self.gain_sum = 0;\n            }\n\n            if (self.loss_sum < 0) {\n                self.loss_sum = 0;\n            }\n        }\n\n        const den = self.gain_sum + self.loss_sum;\n        if (@abs(den) < epsilon) {\n            return 0;\n        }\n\n        return 100.0 * (self.gain_sum - self.loss_sum) / den;\n    }\n\n    pub fn isPrimed(self: *const ChandeMomentumOscillator) bool {\n        return self.primed;\n    }\n\n    pub fn getMetadata(self: *const ChandeMomentumOscillator, out: *Metadata) void {\n        build_metadata_mod.buildMetadata(\n            out,\n            .chande_momentum_oscillator,\n            self.line.mnemonic,\n            self.line.description,\n            &[_]build_metadata_mod.OutputText{\n                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },\n            },\n        );\n    }\n\n    pub fn updateScalar(self: *ChandeMomentumOscillator, sample: *const Scalar) OutputArray {\n        const value = self.update(sample.value);\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn updateBar(self: *ChandeMomentumOscillator, sample: *const Bar) OutputArray {\n        const value = self.update(self.line.extractBar(sample));\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn updateQuote(self: *ChandeMomentumOscillator, sample: *const Quote) OutputArray {\n        const value = self.update(self.line.extractQuote(sample));\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn updateTrade(self: *ChandeMomentumOscillator, sample: *const Trade) OutputArray {\n        const value = self.update(self.line.extractTrade(sample));\n        return LineIndicator.wrapScalar(sample.time, value);\n    }\n\n    pub fn indicator(self: *ChandeMomentumOscillator) indicator_mod.Indicator {\n        return .{\n            .ptr = @ptrCast(self),\n            .vtable = &vtable,\n        };\n    }\n\n    const vtable = indicator_mod.Indicator.VTable{\n        .isPrimed = vtableIsPrimed,\n        .metadata = vtableMetadata,\n        .updateScalar = vtableUpdateScalar,\n        .updateBar = vtableUpdateBar,\n        .updateQuote = vtableUpdateQuote,\n        .updateTrade = vtableUpdateTrade,\n    };\n\n    fn vtableIsPrimed(ptr: *anyopaque) bool {\n        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));\n        return self.isPrimed();\n    }\n\n    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {\n        const self: *const ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));\n        self.getMetadata(out);\n    }\n\n    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {\n        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateScalar(sample);\n    }\n\n    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {\n        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateBar(sample);\n    }\n\n    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {\n        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateQuote(sample);\n    }\n\n    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {\n        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));\n        return self.updateTrade(sample);\n    }\n\n    pub const InitError = error{\n        InvalidLength,\n        MnemonicTooLong,\n        OutOfMemory,\n    };\n};\n\n// ---------------------------------------------------------------------------\n// Tests\n// ---------------------------------------------------------------------------\n\nconst testing = std.testing;\n\nfn testBookLength10Input() [13]f64 {\n    return .{\n        101.0313, 101.0313, 101.1250, 101.9687, 102.7813,\n        103.0000, 102.9687, 103.0625, 102.9375, 102.7188,\n        102.7500, 102.9063, 102.9687,\n    };\n}\n\nfn testBookLength10Output() [13]f64 {\n    return .{\n        math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64),\n        math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64),\n        69.61963786608334, 71.42857142857143, 71.08377992828775,\n    };\n}\n\nfn testInput() [252]f64 {\n    return .{\n        91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,\n        96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,\n        88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,\n        85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,\n        83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,\n        89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,\n        88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,\n        88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,\n        102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,\n        123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,\n        116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,\n        114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,\n        123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,\n        137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,\n        123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,\n        122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,\n        124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,\n        132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,\n        127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,\n        122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,\n        107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,\n        94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,\n        94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,\n        103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,\n        107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,\n        108.750000, 107.870000,\n    };\n}\n\nfn createCmo(allocator: std.mem.Allocator, length: usize) !ChandeMomentumOscillator {\n    var cmo = try ChandeMomentumOscillator.init(allocator, .{ .length = length });\n    cmo.fixSlices();\n    return cmo;\n}\n\ntest \"cmo update length 10 book\" {\n    const input = testBookLength10Input();\n    const output = testBookLength10Output();\n    var cmo = try createCmo(testing.allocator, 10);\n    defer cmo.deinit();\n\n    // First 10 updates produce NaN.\n    for (0..10) |i| {\n        const act = cmo.update(input[i]);\n        try testing.expect(math.isNan(act));\n    }\n\n    // From index 10 onward, primed.\n    for (10..13) |i| {\n        const act = cmo.update(input[i]);\n        try testing.expect(@abs(act - output[i]) < 1e-13);\n    }\n\n    // NaN passthrough\n    try testing.expect(math.isNan(cmo.update(math.nan(f64))));\n}\n\ntest \"cmo is primed\" {\n    const input = testInput();\n\n    inline for ([_]usize{ 1, 2, 3, 5, 10 }) |length| {\n        var cmo = try createCmo(testing.allocator, length);\n        defer cmo.deinit();\n\n        try testing.expect(!cmo.isPrimed());\n\n        for (0..length) |i| {\n            _ = cmo.update(input[i]);\n            try testing.expect(!cmo.isPrimed());\n        }\n\n        for (length..252) |i| {\n            _ = cmo.update(input[i]);\n            try testing.expect(cmo.isPrimed());\n        }\n    }\n}\n\ntest \"cmo metadata\" {\n    var cmo = try createCmo(testing.allocator, 5);\n    defer cmo.deinit();\n    var m: Metadata = undefined;\n    cmo.getMetadata(&m);\n\n    try testing.expectEqual(Identifier.chande_momentum_oscillator, m.identifier);\n    try testing.expectEqual(@as(usize, 1), m.outputs_len);\n    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);\n    try testing.expectEqualStrings(\"cmo(5)\", m.outputs_buf[0].mnemonic);\n    try testing.expectEqualStrings(\"Chande Momentum Oscillator cmo(5)\", m.outputs_buf[0].description);\n}\n\ntest \"cmo update entity\" {\n    const length: usize = 2;\n    const inp: f64 = 3.0;\n    const exp: f64 = 100.0; // all gains, no losses\n    const time: i64 = 1617235200;\n\n    // scalar\n    {\n        var cmo = try createCmo(testing.allocator, length);\n        defer cmo.deinit();\n        _ = cmo.update(0.0);\n        _ = cmo.update(0.0);\n        const out = cmo.updateScalar(&.{ .time = time, .value = inp });\n        try testing.expectEqual(@as(usize, 1), out.len);\n        const s = out.slice()[0].scalar;\n        try testing.expectEqual(time, s.time);\n        try testing.expect(@abs(s.value - exp) < 1e-13);\n    }\n\n    // bar\n    {\n        var cmo = try createCmo(testing.allocator, length);\n        defer cmo.deinit();\n        _ = cmo.update(0.0);\n        _ = cmo.update(0.0);\n        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };\n        const out = cmo.updateBar(&bar);\n        const s = out.slice()[0].scalar;\n        try testing.expect(@abs(s.value - exp) < 1e-13);\n    }\n\n    // quote\n    {\n        var cmo = try createCmo(testing.allocator, length);\n        defer cmo.deinit();\n        _ = cmo.update(0.0);\n        _ = cmo.update(0.0);\n        const quote = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };\n        const out = cmo.updateQuote(&quote);\n        const s = out.slice()[0].scalar;\n        try testing.expect(@abs(s.value - exp) < 1e-13);\n    }\n\n    // trade\n    {\n        var cmo = try createCmo(testing.allocator, length);\n        defer cmo.deinit();\n        _ = cmo.update(0.0);\n        _ = cmo.update(0.0);\n        const trade = Trade{ .time = time, .price = inp, .volume = 0 };\n        const out = cmo.updateTrade(&trade);\n        const s = out.slice()[0].scalar;\n        try testing.expect(@abs(s.value - exp) < 1e-13);\n    }\n}\n\ntest \"cmo init invalid length\" {\n    const result = ChandeMomentumOscillator.init(testing.allocator, .{ .length = 0 });\n    try testing.expectError(error.InvalidLength, result);\n}\n\ntest \"cmo mnemonic components\" {\n    // all defaults -> no component suffix\n    {\n        var cmo = try createCmo(testing.allocator, 5);\n        defer cmo.deinit();\n        try testing.expectEqualStrings(\"cmo(5)\", cmo.line.mnemonic);\n    }\n\n    // bar component set to Median\n    {\n        var cmo = try ChandeMomentumOscillator.init(testing.allocator, .{\n            .length = 5,\n            .bar_component = .median,\n        });\n        defer cmo.deinit();\n        cmo.fixSlices();\n        try testing.expectEqualStrings(\"cmo(5, hl/2)\", cmo.line.mnemonic);\n        try testing.expectEqualStrings(\"Chande Momentum Oscillator cmo(5, hl/2)\", cmo.line.description);\n    }\n\n    // bar=high, trade=volume\n    {\n        var cmo = try ChandeMomentumOscillator.init(testing.allocator, .{\n            .length = 5,\n            .bar_component = .high,\n            .trade_component = .volume,\n        });\n        defer cmo.deinit();\n        cmo.fixSlices();\n        try testing.expectEqualStrings(\"cmo(5, h, v)\", cmo.line.mnemonic);\n    }\n}\n"
}
```

**Output:**
```
Wrote file successfully.

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/icalc/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.identifier" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.factory.factory" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/iconf/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.descriptors" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.core.outputs.heatmap" could not be resolved
ERROR [26:6] Import "py.indicators.core.outputs.shape" could not be resolved
ERROR [27:6] Import "py.indicators.core.pane" could not be resolved
ERROR [28:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [31:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/ifres/main.py">
ERROR [16:6] Import "py.indicators.core.frequency_response" could not be resolved
ERROR [17:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [18:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [21:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>
```

---

## Assistant (Build · claude-opus-4.6 · 6.8s)

Now register it and test:

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "newString": "pub const aroon = @import(\"tushar_chande/aroon/aroon.zig\");\npub const chande_momentum_oscillator = @import(\"tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig\");",
  "oldString": "pub const aroon = @import(\"tushar_chande/aroon/aroon.zig\");"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 5.5s)

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "newString": "    _ = aroon;\n    _ = chande_momentum_oscillator;\n}",
  "oldString": "    _ = aroon;\n}"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 6.5s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | tail -20",
  "description": "Build and test with CMO",
  "timeout": 120000
}
```

**Output:**
```
|  +- compile test Debug native cached 9ms MaxRSS:39M
+- run test 2 pass (2 total) 14ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 6 pass (6 total) 20ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 7 pass (7 total) 6ms MaxRSS:11M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test success 6ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test success 6ms MaxRSS:11M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 11 pass (11 total) 8ms MaxRSS:11M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 10 pass (10 total) 7ms MaxRSS:11M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 4 pass (4 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 295 pass (295 total) 46ms MaxRSS:11M
   +- compile test Debug native success 857ms MaxRSS:178M


```

---

## Assistant (Build · claude-opus-4.6 · 5.8s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | grep -oP '\\d+ pass' | awk -F' ' '{s+=$1} END{print s}'",
  "description": "Sum total tests"
}
```

**Output:**
```
662

```

---

## Assistant (Build · claude-opus-4.6 · 5.9s)

All 662 tests pass. CMO added 6 new tests.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Register Aroon in indicators.zig and compile/test",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Write CMO implementation",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Register CMO in indicators.zig and compile/test",
      "priority": "high",
      "status": "completed"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Register Aroon in indicators.zig and compile/test",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Write CMO implementation",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Register CMO in indicators.zig and compile/test",
    "priority": "high",
    "status": "completed"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 8.2s)

Both tushar_chande indicators done (Aroon + CMO). **662 Zig tests pass** (up from 650). StochasticRSI remains blocked on welleswilder/RSI.

The tushar_chande family is now complete (excluding StochRSI which is blocked). Next up would be the **welleswilder** family (12 indicators including RSI, which unblocks StochRSI). Want me to continue with welleswilder?

---

## User

yes

---

## Assistant (Build · claude-opus-4.6 · 5.9s)

_Thinking:_

Let me start with the welleswilder family. I need to understand what indicators are in this family.

Let me explore the welleswilder family to understand what we're working with.

**Tool: bash**

**Input:**
```json
{
  "command": "ls go/indicators/welleswilder/",
  "description": "List welleswilder indicators"
}
```

**Output:**
```
averagedirectionalmovementindex
averagedirectionalmovementindexrating
averagetruerange
directionalindicatorminus
directionalindicatorplus
directionalmovementindex
directionalmovementminus
directionalmovementplus
doc.go
normalizedaveragetruerange
parabolicstopandreverse
relativestrengthindex
truerange

```

---

## Assistant (Build · claude-opus-4.6 · 7.7s)

12 indicators. Let me understand their dependency chain and complexity.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Research welleswilder dependency chain and plan batch order",
      "priority": "high",
      "status": "in_progress"
    },
    {
      "content": "Batch 1: TrueRange, ATR, NormalizedATR",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 2: DM+, DM-, DI+, DI-",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 3: DMI (ADX), ADXR",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 4: RSI",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 5: ParabolicSAR",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "After welleswilder: StochasticRSI (tushar_chande)",
      "priority": "medium",
      "status": "pending"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Research welleswilder dependency chain and plan batch order",
    "priority": "high",
    "status": "in_progress"
  },
  {
    "content": "Batch 1: TrueRange, ATR, NormalizedATR",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 2: DM+, DM-, DI+, DI-",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 3: DMI (ADX), ADXR",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 4: RSI",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 5: ParabolicSAR",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "After welleswilder: StochasticRSI (tushar_chande)",
    "priority": "medium",
    "status": "pending"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 75.6s)

**Tool: task**

**Input:**
```json
{
  "description": "Research welleswilder deps",
  "prompt": "Research the welleswilder indicator family in go/indicators/welleswilder/. I need to understand the dependency chain between the 12 indicators. For each indicator, check its Go source file for imports of OTHER welleswilder indicators (e.g., does ATR depend on TrueRange? Does ADX depend on DI+/DI-?). Also note which ones are HLCV-override indicators (override updateBar to extract high/low/close directly) vs LineIndicator pattern (component triple with single scalar).\n\nReturn a dependency graph and suggested porting order (leaf dependencies first). For each indicator, note:\n1. What it depends on (other welleswilder indicators)\n2. Whether it's HLCV-override or LineIndicator\n3. Whether it needs an allocator (ring buffer)\n4. Number of outputs\n5. Brief description of the algorithm\n\nCheck these files:\n- go/indicators/welleswilder/truerange/truerange.go\n- go/indicators/welleswilder/averagetruerange/averagetruerange.go\n- go/indicators/welleswilder/normalizedaveragetruerange/normalizedaveragetruerange.go\n- go/indicators/welleswilder/directionalmovementplus/directionalmovementplus.go\n- go/indicators/welleswilder/directionalmovementminus/directionalmovementminus.go\n- go/indicators/welleswilder/directionalindicatorplus/directionalindicatorplus.go\n- go/indicators/welleswilder/directionalindicatorminus/directionalindicatorminus.go\n- go/indicators/welleswilder/directionalmovementindex/directionalmovementindex.go\n- go/indicators/welleswilder/averagedirectionalmovementindex/averagedirectionalmovementindex.go\n- go/indicators/welleswilder/averagedirectionalmovementindexrating/averagedirectionalmovementindexrating.go\n- go/indicators/welleswilder/relativestrengthindex/relativestrengthindex.go\n- go/indicators/welleswilder/parabolicstopandreverse/parabolicstopandreverse.go\n\nAlso check their params.go files for parameter definitions.",
  "subagent_type": "explore"
}
```

**Output:**
```
task_id: ses_22d5c8b7cffeWPQMpczbTRZz9M (for resuming to continue this task if needed)

<task_result>
Here is the complete analysis:

---

## Dependency Graph

```
TrueRange (TR)                    ── no deps
DirectionalMovementPlus (+DM)     ── no deps
DirectionalMovementMinus (-DM)    ── no deps
RelativeStrengthIndex (RSI)       ── no deps
ParabolicStopAndReverse (SAR)     ── no deps
  │
AverageTrueRange (ATR)            ── depends on: TrueRange
  │
NormalizedAverageTrueRange (NATR) ── depends on: ATR (→ TR)
DirectionalIndicatorPlus (+DI)    ── depends on: ATR (→ TR), +DM
DirectionalIndicatorMinus (-DI)   ── depends on: ATR (→ TR), -DM
  │
DirectionalMovementIndex (DX)     ── depends on: +DI (→ ATR, +DM), -DI (→ ATR, -DM)
  │
AverageDirectionalMovementIndex (ADX)        ── depends on: DX (→ +DI, -DI)
  │
AverageDirectionalMovementIndexRating (ADXR) ── depends on: ADX (→ DX)
```

## Per-Indicator Details

| # | Indicator | Mnemonic | Depends On | Pattern | Ring Buffer? | Outputs | Algorithm |
|---|-----------|----------|------------|---------|-------------|---------|-----------|
| 1 | **TrueRange** | `tr` | None | **HLCV-override** — custom `Update(close,high,low)`, overrides UpdateBar to extract H/L/C directly | No | 1 | max(H-L, abs(H-prevC), abs(L-prevC)). First bar stores close, primed from bar 2. |
| 2 | **DirectionalMovementPlus** | `+dm` | None | **HLCV-override** — custom `Update(high,low)`, overrides UpdateBar to extract H/L | No | 1 | If upMove>downMove and upMove>0, +DM=upMove else 0. Wilder smoothing when length>1. |
| 3 | **DirectionalMovementMinus** | `-dm` | None | **HLCV-override** — custom `Update(high,low)`, overrides UpdateBar to extract H/L | No | 1 | If downMove>upMove and downMove>0, -DM=downMove else 0. Wilder smoothing when length>1. |
| 4 | **RelativeStrengthIndex** | `rsi` | None | **LineIndicator** — embeds `core.LineIndicator`, component triple (bar/quote/trade), scalar `Update(sample)` | No | 1 | Wilder-smoothed avg gain / (avg gain + avg loss) * 100. Accumulates first `length` changes, then smooths. |
| 5 | **ParabolicStopAndReverse** | `sar` | None | **LineIndicator** (partial) — embeds `core.LineIndicator` but **overrides UpdateBar** to call `UpdateHL(high,low)`. Scalar `Update(sample)` delegates to `UpdateHL(sample,sample)`. | No | 1 | Acceleration factor * (extreme point - SAR). Flips long/short on price penetration. Signed output: +long, -short. |
| 6 | **AverageTrueRange** | `atr` | **TrueRange** | **HLCV-override** — custom `Update(close,high,low)`, embeds TR instance | Yes — `window []float64` of size `length` for initial SMA | 1 | SMA of first `length` TR values, then Wilder smoothing: (prev*(n-1)+TR)/n. |
| 7 | **NormalizedAverageTrueRange** | `natr` | **ATR** | **HLCV-override** — custom `Update(close,high,low)`, embeds ATR instance | No (ATR has it) | 1 | (ATR / close) * 100. Zero-guard on close. |
| 8 | **DirectionalIndicatorPlus** | `+di` | **ATR**, **+DM** | **HLCV-override** — custom `Update(close,high,low)` | No | 4 (metadata lists +di, +dm, atr, tr) | 100 * +DM(n) / (ATR * length). Epsilon guard on zero ATR. |
| 9 | **DirectionalIndicatorMinus** | `-di` | **ATR**, **-DM** | **HLCV-override** — custom `Update(close,high,low)` | No | 4 (metadata lists -di, -dm, atr, tr) | 100 * -DM(n) / (ATR * length). Epsilon guard on zero ATR. |
| 10 | **DirectionalMovementIndex** | `dx` | **+DI**, **-DI** | **HLCV-override** — custom `Update(close,high,low)` | No | 7 (metadata lists dx, +di, -di, +dm, -dm, atr, tr) | 100 * abs(+DI - -DI) / (+DI + -DI). Epsilon guard. |
| 11 | **AverageDirectionalMovementIndex** | `adx` | **DX** | **HLCV-override** — custom `Update(close,high,low)` | No (SMA via sum/count) | 8 (metadata) | SMA of first `length` DX values, then Wilder smoothing: (prev*(n-1)+DX)/n. |
| 12 | **AverageDirectionalMovementIndexRating** | `adxr` | **ADX** | **HLCV-override** — custom `Update(close,high,low)` | Yes — `buffer []float64` circular buffer of size `length` for ADX lookback | 9 (metadata) | (ADX[now] + ADX[now-(length-1)]) / 2. |

**Note on outputs**: While the metadata lists many outputs for documentation (showing the full sub-indicator chain), the actual `Update*` methods all return a single `[]any` with 1 Scalar. The metadata outputs are informational only.

## Suggested Porting Order (leaf-first)

```
Level 0 (no deps — port first, in any order):
  1. TrueRange
  2. DirectionalMovementPlus
  3. DirectionalMovementMinus
  4. RelativeStrengthIndex        ← LineIndicator pattern
  5. ParabolicStopAndReverse      ← LineIndicator + HLCV override hybrid

Level 1 (depends on Level 0):
  6. AverageTrueRange             ← depends on TR; needs allocator for window[]

Level 2 (depends on Level 1 + Level 0):
  7. NormalizedAverageTrueRange   ← depends on ATR
  8. DirectionalIndicatorPlus     ← depends on ATR + +DM
  9. DirectionalIndicatorMinus    ← depends on ATR + -DM

Level 3:
 10. DirectionalMovementIndex     ← depends on +DI + -DI

Level 4:
 11. AverageDirectionalMovementIndex ← depends on DX

Level 5:
 12. AverageDirectionalMovementIndexRating ← depends on ADX; needs allocator for buffer[]
```

## Key Architectural Observations

1. **Two patterns**: 10 of 12 are pure HLCV-override (custom `Update(close,high,low)` or `Update(high,low)` with manual UpdateBar/UpdateScalar/UpdateQuote/UpdateTrade). Only **RSI** and **SAR** use the `core.LineIndicator` embedding with component triples.

2. **SAR is a hybrid**: It embeds `LineIndicator` for scalar/quote/trade routing, but **overrides UpdateBar** to call `UpdateHL(high, low)` instead of extracting a single component. Its `Update(sample)` calls `UpdateHL(sample, sample)`.

3. **Only 2 need allocators**: ATR (window for initial SMA) and ADXR (circular buffer for lookback). All others use only scalar state variables.

4. **All return 1 output**: Despite rich metadata output lists, every `Update*` method returns exactly 1 Scalar in the output slice.

5. **Params pattern**: The 10 HLCV-override indicators take simple `int length` in their constructors (no params struct). RSI and SAR use params structs with component triples. RSI params: `{Length, BarComponent, QuoteComponent, TradeComponent}`. SAR params: `{StartValue, OffsetOnReverse, 6 acceleration factors}`.

6. **The DM family is the deepest chain**: ADXR -> ADX -> DX -> +DI/-DI -> ATR/+DM/-DM -> TR. That is 6 levels of composition, each creating sub-indicator instances internally.
</task_result>
```

---

## Assistant (Build · claude-opus-4.6 · 8.5s)

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Research welleswilder dependency chain and plan batch order",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Batch 1: TrueRange, DM+, DM- (leaf HLCV, no deps, no alloc)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 2: RSI (LineIndicator, no deps, unblocks StochRSI)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 3: ATR (depends on TR, needs alloc)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 4: NATR, DI+, DI- (depend on ATR/DM)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 5: DX (depends on DI+, DI-)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 6: ADX (depends on DX)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 7: ADXR (depends on ADX, needs alloc)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "Batch 8: ParabolicSAR (independent, complex)",
      "priority": "high",
      "status": "pending"
    },
    {
      "content": "After welleswilder: StochasticRSI (tushar_chande)",
      "priority": "medium",
      "status": "pending"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Research welleswilder dependency chain and plan batch order",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Batch 1: TrueRange, DM+, DM- (leaf HLCV, no deps, no alloc)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 2: RSI (LineIndicator, no deps, unblocks StochRSI)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 3: ATR (depends on TR, needs alloc)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 4: NATR, DI+, DI- (depend on ATR/DM)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 5: DX (depends on DI+, DI-)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 6: ADX (depends on DX)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 7: ADXR (depends on ADX, needs alloc)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "Batch 8: ParabolicSAR (independent, complex)",
    "priority": "high",
    "status": "pending"
  },
  {
    "content": "After welleswilder: StochasticRSI (tushar_chande)",
    "priority": "medium",
    "status": "pending"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 535.1s)

Let me start with Batch 1: TrueRange, DM+, DM-. I need to read the Go sources and tests for all three.

**Tool: task**

**Input:**
```json
{
  "description": "Read TR, DM+, DM- Go sources",
  "prompt": "Read and return the COMPLETE contents of these files. I need every line — do NOT summarize or truncate:\n\n1. go/indicators/welleswilder/truerange/truerange.go\n2. go/indicators/welleswilder/truerange/truerange_test.go\n3. go/indicators/welleswilder/truerange/output.go\n4. go/indicators/welleswilder/directionalmovementplus/directionalmovementplus.go\n5. go/indicators/welleswilder/directionalmovementplus/directionalmovementplus_test.go\n6. go/indicators/welleswilder/directionalmovementplus/output.go\n7. go/indicators/welleswilder/directionalmovementplus/params.go\n8. go/indicators/welleswilder/directionalmovementminus/directionalmovementminus.go\n9. go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go\n10. go/indicators/welleswilder/directionalmovementminus/output.go\n11. go/indicators/welleswilder/directionalmovementminus/params.go\n\nReturn the full file contents for each file.",
  "subagent_type": "general"
}
```

**Output:**
```
task_id: ses_22d5b529bffexJH27BjdOh79Ow (for resuming to continue this task if needed)

<task_result>
Here are the complete contents of all 11 files:

---

### 1. `go/indicators/welleswilder/truerange/truerange.go` (145 lines)

```go
package truerange

import (
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const (
	trMnemonic    = "tr"
	trDescription = "True Range"
)

// TrueRange is Welles Wilder's True Range indicator.
//
// The True Range is defined as the largest of:
//   - the distance from today's high to today's low
//   - the distance from yesterday's close to today's high
//   - the distance from yesterday's close to today's low
//
// The first update stores the close and returns NaN (not primed).
// The indicator is primed from the second update onward.
//
// Unlike most indicators, TrueRange requires bar data (high, low, close)
// and does not use a single bar component. For scalar, quote, and trade updates,
// the single value is used as a substitute for high, low, and close.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type TrueRange struct {
	mu            sync.RWMutex
	previousClose float64
	value         float64
	primed        bool
}

// NewTrueRange returns a new instance of the True Range indicator.
func NewTrueRange() *TrueRange {
	return &TrueRange{
		previousClose: math.NaN(),
		value:         math.NaN(),
	}
}

// IsPrimed indicates whether the indicator is primed.
func (s *TrueRange) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *TrueRange) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.TrueRange,
		trMnemonic,
		trDescription,
		[]core.OutputText{
			{Mnemonic: trMnemonic, Description: trDescription},
		},
	)
}

// Update updates the True Range given the next bar's close, high, and low values.
func (s *TrueRange) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.primed {
		if math.IsNaN(s.previousClose) {
			s.previousClose = close

			return math.NaN()
		}

		s.primed = true
	}

	greatest := high - low

	if temp := math.Abs(high - s.previousClose); greatest < temp {
		greatest = temp
	}

	if temp := math.Abs(low - s.previousClose); greatest < temp {
		greatest = temp
	}

	s.value = greatest
	s.previousClose = close

	return s.value
}

// UpdateSample updates the True Range using a single sample value
// as a substitute for high, low, and close.
func (s *TrueRange) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *TrueRange) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *TrueRange) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *TrueRange) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *TrueRange) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
```

---

### 2. `go/indicators/welleswilder/truerange/truerange_test.go` (342 lines)

```go
//nolint:testpackage
package truerange

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// TA-Lib test data (252 entries), extracted programmatically from TrueRangeTest.cs.
func testInputHigh() []float64 {
	return []float64{
		93.25, 94.94, 96.375, 96.19, 96.0, 94.72, 95.0, 93.72, 92.47, 92.75,
		96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
		90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
		87.155, 87.25, 87.375, 88.97, 90.0, 89.845, 86.97, 85.94, 84.75, 85.47,
		84.47, 88.5, 89.47, 90.0, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
		90.0, 88.875, 89.0, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
		90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92.0, 91.94,
		89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
		104.5, 105.5, 106.125, 107.94, 106.25, 107.0, 108.75, 110.94, 110.94, 114.22,
		123.0, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0, 118.315,
		116.87, 116.75, 113.87, 114.62, 115.31, 116.0, 121.69, 119.87, 120.87, 116.75,
		116.5, 116.0, 118.31, 121.5, 122.0, 121.44, 125.75, 127.75, 124.19, 124.44,
		125.75, 124.69, 125.31, 132.0, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
		138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
		125.12, 126.5, 128.69, 126.62, 126.69, 126.0, 123.12, 121.87, 124.0, 127.0,
		124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
		124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126.0, 128.5, 126.75, 129.75,
		132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135.0, 132.38, 131.44, 130.88,
		129.63, 127.25, 127.81, 125.0, 126.81, 124.75, 122.81, 122.25, 121.06, 120.0,
		123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109.0,
		108.5, 113.06, 93.0, 94.62, 95.12, 96.0, 95.56, 95.31, 99.0, 98.81,
		96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
		95.0, 94.87, 98.25, 105.12, 108.44, 109.87, 105.0, 106.0, 104.94, 104.5,
		104.44, 106.31, 112.87, 116.5, 119.19, 121.0, 122.12, 111.94, 112.75, 110.19,
		107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110.0, 110.75, 110.5,
		110.5, 109.5,
	}
}

func testInputLow() []float64 {
	return []float64{
		90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92.0, 89.75, 89.44, 90.625,
		92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
		87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
		84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
		82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89.0, 88.565, 90.095,
		89.0, 86.47, 84.0, 83.315, 82.0, 83.25, 84.75, 85.28, 87.19, 88.44,
		88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92.0, 90.53, 89.97, 88.815,
		86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
		101.72, 101.72, 103.155, 105.69, 103.655, 104.0, 105.53, 108.53, 108.75, 107.75,
		117.0, 118.0, 116.0, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
		112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
		112.56, 114.0, 114.75, 118.87, 119.0, 119.75, 122.62, 123.0, 121.75, 121.56,
		123.12, 122.19, 122.75, 124.37, 128.0, 129.5, 130.81, 130.63, 132.13, 133.88,
		135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
		122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121.0,
		121.12, 118.94, 119.81, 121.0, 122.0, 124.5, 126.56, 123.5, 121.25, 121.06,
		122.31, 121.0, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128.0,
		128.38, 130.69, 131.63, 134.38, 132.0, 131.94, 131.94, 129.56, 123.75, 126.0,
		126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120.0, 119.62, 115.5, 116.75,
		119.06, 119.06, 115.06, 111.06, 113.12, 110.0, 105.0, 104.69, 103.87, 104.69,
		105.44, 107.0, 89.0, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
		93.69, 93.5, 90.0, 90.19, 90.5, 92.12, 94.12, 94.87, 93.0, 93.87,
		93.0, 92.62, 93.56, 98.37, 104.44, 106.0, 101.81, 104.12, 103.37, 102.12,
		102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
		104.5, 105.75, 108.62, 107.75, 108.06, 108.0, 108.19, 108.12, 109.06, 108.75,
		108.56, 106.62,
	}
}

func testInputClose() []float64 {
	return []float64{
		91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
		96.125, 97.25, 98.5, 89.875, 91.0, 92.815, 89.155, 89.345, 91.625, 89.875,
		88.375, 87.625, 84.78, 83.0, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
		85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
		83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89.0, 91.0, 90.5,
		89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
		88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90.0, 89.69,
		88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106.0, 102.5,
		102.405, 104.595, 106.125, 106.0, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
		123.0, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
		116.0, 116.0, 112.0, 113.75, 112.94, 116.0, 120.5, 116.62, 117.0, 115.25,
		114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
		123.12, 122.56, 124.62, 129.25, 131.0, 132.25, 131.0, 132.81, 134.0, 137.38,
		137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129.0, 123.87, 124.81,
		123.0, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
		122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
		124.44, 122.0, 122.37, 122.94, 124.0, 123.19, 124.56, 127.25, 125.87, 128.86,
		132.0, 130.75, 134.75, 135.0, 132.38, 133.31, 131.94, 130.0, 125.37, 130.13,
		127.12, 125.19, 122.0, 125.0, 123.0, 123.5, 120.06, 121.0, 117.75, 119.87,
		122.0, 119.19, 116.37, 113.5, 114.25, 110.0, 105.06, 107.0, 107.87, 107.0,
		107.12, 107.0, 91.0, 93.94, 93.87, 95.5, 93.0, 94.94, 98.25, 96.75,
		94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97.0, 95.0, 95.87, 94.06,
		94.62, 93.75, 98.0, 103.94, 107.87, 106.06, 104.5, 105.0, 104.19, 103.06,
		103.42, 105.27, 111.87, 116.0, 116.62, 118.28, 113.37, 109.0, 109.7, 109.25,
		107.0, 109.19, 110.0, 109.2, 110.12, 108.0, 108.62, 109.75, 109.81, 109.0,
		108.75, 107.87,
	}
}

// Full expected TR output (252 entries). First entry is NaN.
func testExpectedTR() []float64 {
	return []float64{
		math.NaN(), 3.535, 2.125, 2.69, 3.185, 1.22, 3.0, 3.97, 3.31, 2.435,
		3.78, 3.5, 3.095, 9.685, 4.565, 2.31, 4.5, 1.875, 2.72, 2.5,
		2.845, 1.97, 3.625, 3.22, 2.875, 3.875, 3.19, 5.34, 3.655, 3.155,
		2.75, 2.155, 1.875, 3.44, 2.125, 3.28, 2.315, 3.565, 2.31, 2.03,
		1.94, 5.125, 3.97, 1.47, 3.16, 1.315, 2.22, 2.72, 2.59, 1.655,
		1.5, 2.56, 5.0, 1.935, 1.815, 2.56, 1.875, 2.66, 3.185, 2.185,
		2.5, 1.5, 3.47, 2.28, 4.285, 2.875, 2.03, 2.625, 2.03, 3.125,
		3.0, 3.81, 4.125, 3.375, 3.375, 13.435, 6.81, 5.5, 3.375, 4.25,
		2.78, 3.78, 2.97, 2.25, 2.595, 3.0, 4.125, 2.41, 2.19, 6.47,
		10.25, 5.0, 3.815, 1.815, 2.845, 1.94, 2.095, 4.47, 2.5, 7.72,
		5.505, 2.56, 4.81, 5.18, 3.75, 3.56, 5.69, 4.44, 4.31, 3.69,
		3.94, 2.0, 3.56, 5.63, 3.0, 1.69, 5.0, 4.75, 2.44, 2.88,
		3.19, 2.5, 2.75, 7.63, 3.31, 2.75, 3.07, 2.87, 3.37, 3.56,
		3.31, 3.44, 2.31, 3.63, 2.12, 5.19, 8.57, 2.88, 5.5, 2.82,
		2.37, 3.5, 2.88, 3.76, 2.32, 4.19, 4.93, 3.81, 6.44, 6.0,
		3.32, 3.56, 4.44, 2.81, 3.38, 4.5, 2.0, 6.13, 3.62, 3.31,
		3.12, 3.44, 3.19, 3.81, 2.44, 2.93, 3.13, 3.94, 3.0, 3.88,
		4.31, 3.25, 5.75, 3.31, 3.56, 1.62, 3.06, 2.82, 7.69, 5.51,
		3.88, 2.88, 6.37, 4.56, 5.44, 3.06, 3.5, 2.63, 5.56, 3.25,
		4.19, 3.69, 4.13, 5.31, 3.57, 4.87, 5.87, 2.56, 5.0, 4.31,
		3.06, 6.06, 18.0, 3.62, 3.0, 2.13, 2.75, 2.31, 4.06, 2.44,
		3.12, 2.44, 4.44, 2.75, 3.69, 3.38, 3.44, 2.63, 3.25, 2.5,
		2.0, 2.25, 4.69, 7.12, 4.5, 3.87, 4.25, 1.88, 1.63, 2.38,
		2.19, 2.94, 7.6, 4.63, 3.75, 5.5, 9.87, 5.81, 6.19, 3.32,
		4.75, 3.94, 2.44, 2.69, 2.06, 2.31, 2.44, 1.88, 1.69, 1.75,
		1.94, 2.88,
	}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestTrueRangeUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-3

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()
	expected := testExpectedTR()
	tr := NewTrueRange()

	for i := range close {
		act := tr.Update(close[i], high[i], low[i])

		if i == 0 {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if math.IsNaN(act) {
			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}
}

func TestTrueRangeNaNPassthrough(t *testing.T) {
	t.Parallel()

	tr := NewTrueRange()

	if !math.IsNaN(tr.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(tr.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(tr.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}
}

func TestTrueRangeIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()
	tr := NewTrueRange()

	if tr.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	tr.Update(close[0], high[0], low[0])

	if tr.IsPrimed() {
		t.Error("[0] should not be primed after first update")
	}

	tr.Update(close[1], high[1], low[1])

	if !tr.IsPrimed() {
		t.Error("[1] should be primed after second update")
	}

	tr.Update(close[2], high[2], low[2])

	if !tr.IsPrimed() {
		t.Error("[2] should remain primed")
	}
}

func TestTrueRangeUpdateSample(t *testing.T) {
	t.Parallel()

	// When using a single value as H=L=C, TR = |current - previous| after priming.
	tr := NewTrueRange()

	v := tr.UpdateSample(100.0)
	if !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}

	v = tr.UpdateSample(105.0)
	if math.Abs(v-5.0) > 1e-10 {
		t.Errorf("expected 5.0, got %v", v)
	}

	v = tr.UpdateSample(102.0)
	if math.Abs(v-3.0) > 1e-10 {
		t.Errorf("expected 3.0, got %v", v)
	}
}

func TestTrueRangeUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	tr := NewTrueRange()

	// Prime with first bar.
	tr.Update(100, 105, 95)

	check := func(act core.Output) {
		t.Helper()

		if len(act) != 1 {
			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
		}

		s, ok := act[0].(entities.Scalar)
		if !ok {
			t.Error("output is not scalar")
		}

		if s.Time != tm {
			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
		}

		if math.IsNaN(s.Value) {
			t.Error("value should not be NaN after priming")
		}
	}

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		b := entities.Bar{Time: tm, High: 110, Low: 98, Close: 108}
		check(tr2.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		s := entities.Scalar{Time: tm, Value: 108}
		check(tr2.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		q := entities.Quote{Time: tm, Bid: 107, Ask: 109}
		check(tr2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		r := entities.Trade{Time: tm, Price: 108}
		check(tr2.UpdateTrade(&r))
	})
}

func TestTrueRangeMetadata(t *testing.T) {
	t.Parallel()

	tr := NewTrueRange()
	act := tr.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.TrueRange, act.Identifier)
	check("Mnemonic", "tr", act.Mnemonic)
	check("Description", "True Range", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "tr", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "True Range", act.Outputs[0].Description)
}
```

---

### 3. `go/indicators/welleswilder/truerange/output.go` (77 lines)

```go
//nolint:dupl
package truerange

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// The scalar value of the true range.
	Value Output = iota + 1
	outputLast
)

const (
	valueStr   = "value"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Value && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown true range output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == unknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown true range output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
```

---

### 4. `go/indicators/welleswilder/directionalmovementplus/directionalmovementplus.go` (209 lines)

```go
package directionalmovementplus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const (
	dmpMnemonic    = "+dm"
	dmpDescription = "Directional Movement Plus"
)

// DirectionalMovementPlus is Welles Wilder's Directional Movement Plus indicator.
//
// The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
//
// The calculation of the directional movement (+DM and −DM) is as follows:
//   - UpMove = today's high − yesterday's high
//   - DownMove = yesterday's low − today's low
//   - if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0
//
// When the length is greater than 1, Wilder's smoothing method is applied:
//
//	Today's +DM(n) = Previous +DM(n) − Previous +DM(n)/n + Today's +DM(1)
//
// The indicator is not primed during the first length updates.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type DirectionalMovementPlus struct {
	mu           sync.RWMutex
	length       int
	noSmoothing  bool
	count        int
	previousHigh float64
	previousLow  float64
	value        float64
	accumulator  float64
	primed       bool
}

// NewDirectionalMovementPlus returns a new instance of the Directional Movement Plus indicator.
func NewDirectionalMovementPlus(length int) (*DirectionalMovementPlus, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	return &DirectionalMovementPlus{
		length:      length,
		noSmoothing: length == 1,
		value:       math.NaN(),
	}, nil
}

// Length returns the length parameter.
func (s *DirectionalMovementPlus) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *DirectionalMovementPlus) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *DirectionalMovementPlus) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DirectionalMovementPlus,
		dmpMnemonic,
		dmpDescription,
		[]core.OutputText{
			{Mnemonic: dmpMnemonic, Description: dmpDescription},
		},
	)
}

// Update updates the Directional Movement Plus given the next bar's high and low values.
func (s *DirectionalMovementPlus) Update(high, low float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	if high < low {
		high, low = low, high
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.noSmoothing {
		if s.primed {
			deltaPlus := high - s.previousHigh
			deltaMinus := s.previousLow - low

			if deltaPlus > 0 && deltaPlus > deltaMinus {
				s.value = deltaPlus
			} else {
				s.value = 0
			}
		} else {
			if s.count > 0 {
				deltaPlus := high - s.previousHigh
				deltaMinus := s.previousLow - low

				if deltaPlus > 0 && deltaPlus > deltaMinus {
					s.value = deltaPlus
				} else {
					s.value = 0
				}

				s.primed = true
			}

			s.count++
		}
	} else {
		if s.primed {
			deltaPlus := high - s.previousHigh
			deltaMinus := s.previousLow - low

			if deltaPlus > 0 && deltaPlus > deltaMinus {
				s.accumulator += -s.accumulator/float64(s.length) + deltaPlus
			} else {
				s.accumulator += -s.accumulator / float64(s.length)
			}

			s.value = s.accumulator
		} else {
			if s.count > 0 && s.length >= s.count {
				deltaPlus := high - s.previousHigh
				deltaMinus := s.previousLow - low

				if s.length > s.count {
					if deltaPlus > 0 && deltaPlus > deltaMinus {
						s.accumulator += deltaPlus
					}
				} else {
					if deltaPlus > 0 && deltaPlus > deltaMinus {
						s.accumulator += -s.accumulator/float64(s.length) + deltaPlus
					} else {
						s.accumulator += -s.accumulator / float64(s.length)
					}

					s.value = s.accumulator
					s.primed = true
				}
			}

			s.count++
		}
	}

	s.previousLow = low
	s.previousHigh = high

	return s.value
}

// UpdateSample updates the Directional Movement Plus using a single sample value
// as a substitute for high and low.
func (s *DirectionalMovementPlus) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DirectionalMovementPlus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalMovementPlus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DirectionalMovementPlus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DirectionalMovementPlus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}
```

---

### 5. `go/indicators/welleswilder/directionalmovementplus/directionalmovementplus_test.go` (454 lines)

```go
//nolint:testpackage
package directionalmovementplus

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementPlusTest.cs.
func testInputHigh() []float64 {
	return []float64{
		93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
		99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000, 90.345000, 88.500000,
		88.250000, 85.500000, 84.440000, 84.750000, 84.440000, 89.405000, 88.125000, 89.125000, 87.155000, 87.250000, 87.375000,
		88.970000, 90.000000, 89.845000, 86.970000, 85.940000, 84.750000, 85.470000, 84.470000, 88.500000, 89.470000, 90.000000,
		92.440000, 91.440000, 92.970000, 91.720000, 91.155000, 91.750000, 90.000000, 88.875000, 89.000000, 85.250000, 83.815000,
		85.250000, 86.625000, 87.940000, 89.375000, 90.625000, 90.750000, 88.845000, 91.970000, 93.375000, 93.815000, 94.030000,
		94.030000, 91.815000, 92.000000, 91.940000, 89.750000, 88.750000, 86.155000, 84.875000, 85.940000, 99.375000, 103.280000,
		105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000,
		110.940000, 110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000,
		115.345000, 113.000000, 118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000,
		119.870000, 120.870000, 116.750000, 116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000,
		127.750000, 124.190000, 124.440000, 125.750000, 124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000,
		133.500000, 135.500000, 137.440000, 138.690000, 139.190000, 138.500000, 138.130000, 137.500000, 138.880000, 132.130000,
		129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000, 126.620000, 126.690000, 126.000000, 123.120000,
		121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000, 124.500000, 127.870000, 128.560000,
		129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000, 125.620000, 126.000000,
		128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000, 135.000000,
		132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,
		122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000,
		107.250000, 108.870000, 109.000000, 108.500000, 113.060000, 93.000000, 94.620000, 95.120000, 96.000000, 95.560000,
		95.310000, 99.000000, 98.810000, 96.810000, 95.940000, 94.440000, 92.940000, 93.940000, 95.500000, 97.060000, 97.500000,
		96.250000, 96.370000, 95.000000, 94.870000, 98.250000, 105.120000, 108.440000, 109.870000, 105.000000, 106.000000,
		104.940000, 104.500000, 104.440000, 106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000,
		112.750000, 110.190000, 107.940000, 109.690000, 111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000,
		110.750000, 110.500000, 110.500000, 109.500000,
	}
}

func testInputLow() []float64 {
	return []float64{
		90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,
		96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000, 87.500000, 86.530000,
		84.625000, 82.280000, 81.565000, 80.875000, 81.250000, 84.065000, 85.595000, 85.970000, 84.405000, 85.095000, 85.500000,
		85.530000, 87.875000, 86.565000, 84.655000, 83.250000, 82.565000, 83.440000, 82.530000, 85.065000, 86.875000, 88.530000,
		89.280000, 90.125000, 90.750000, 89.000000, 88.565000, 90.095000, 89.000000, 86.470000, 84.000000, 83.315000, 82.000000,
		83.250000, 84.750000, 85.280000, 87.190000, 88.440000, 88.250000, 87.345000, 89.280000, 91.095000, 89.530000, 91.155000,
		92.000000, 90.530000, 89.970000, 88.815000, 86.750000, 85.065000, 82.030000, 81.500000, 82.565000, 96.345000, 96.470000,
		101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000,
		108.530000, 108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000,
		110.875000, 110.500000, 110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000,
		116.060000, 116.560000, 113.310000, 112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000,
		123.000000, 121.750000, 121.560000, 123.120000, 122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000,
		130.630000, 132.130000, 133.880000, 135.380000, 135.750000, 136.190000, 134.500000, 135.380000, 133.690000, 126.060000,
		126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000, 124.620000, 124.370000, 121.810000, 118.190000,
		118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000, 122.000000, 124.500000, 126.560000,
		123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000, 122.690000, 122.870000,
		125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000, 131.940000,
		129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,
		119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000,
		104.690000, 103.870000, 104.690000, 105.440000, 107.000000, 89.000000, 92.500000, 92.120000, 94.620000, 92.810000,
		94.250000, 96.250000, 96.370000, 93.690000, 93.500000, 90.000000, 90.190000, 90.500000, 92.120000, 94.120000, 94.870000,
		93.000000, 93.870000, 93.000000, 92.620000, 93.560000, 98.370000, 104.440000, 106.000000, 101.810000, 104.120000,
		103.370000, 102.120000, 102.250000, 103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000,
		106.560000, 106.870000, 104.500000, 105.750000, 108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000,
		109.060000, 108.750000, 108.560000, 106.620000,
	}
}

// Expected +DM1 (length=1), 252 entries. Extracted from DirectionalMovementPlusTest.cs talibDmPlus1.
func testExpectedDMP1() []float64 {
	return []float64{
		math.NaN(),
		1.6900, 1.4350, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.2800, 3.5000,
		3.3750, 0.0000, 0.0000, 0.0000, 1.9350, 0.0000, 0.0000, 1.3150, 0.2800, 0.0000,
		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 4.9650, 0.0000, 1.0000, 0.0000,
		0.0950, 0.1250, 1.5950, 1.0300, 0.0000, 0.0000, 0.0000, 0.0000, 0.7200, 0.0000,
		4.0300, 0.9700, 0.5300, 2.4400, 0.0000, 1.5300, 0.0000, 0.0000, 0.5950, 0.0000,
		0.0000, 0.0000, 0.0000, 0.0000, 1.4350, 1.3750, 1.3150, 1.4350, 1.2500, 0.0000,
		0.0000, 3.1250, 1.4050, 0.0000, 0.2150, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
		0.0000, 0.0000, 0.0000, 1.0650, 13.4350, 3.9050, 2.0950, 2.2500, 0.0000, 0.0000,
		1.0000, 0.6250, 1.8150, 0.0000, 0.7500, 1.7500, 2.1900, 0.0000, 3.2800, 8.7800,
		0.0000, 0.0000, 0.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 5.3150, 0.0000,
		0.0000, 0.0000, 0.0000, 0.6900, 0.6900, 5.6900, 0.0000, 1.0000, 0.0000, 0.0000,
		0.0000, 2.3100, 3.1900, 0.5000, 0.0000, 4.3100, 2.0000, 0.0000, 0.2500, 1.3100,
		0.0000, 0.6200, 6.6900, 0.0000, 0.9400, 1.6300, 0.0000, 2.0000, 1.9400, 1.2500,
		0.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
		1.3800, 2.1900, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 2.1300, 3.0000, 0.0000,
		0.0000, 1.2500, 0.0600, 0.6900, 3.3700, 0.6900, 0.0000, 0.0000, 0.0000, 0.5000,
		0.0000, 0.4400, 1.8100, 0.0000, 0.4300, 0.3800, 2.5000, 0.0000, 3.0000, 2.9400,
		1.2500, 2.5600, 1.1900, 0.0000, 0.0000, 1.4400, 0.0000, 0.0000, 0.0000, 0.0000,
		0.0000, 0.0000, 0.0000, 1.8100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.2500,
		0.0000, 0.0000, 0.0000, 1.6300, 0.0000, 0.0000, 0.0000, 1.6200, 0.1300, 0.0000,
		4.5600, 0.0000, 1.6200, 0.5000, 0.8800, 0.0000, 0.0000, 3.6900, 0.0000, 0.0000,
		0.0000, 0.0000, 0.0000, 1.0000, 1.5600, 1.5600, 0.4400, 0.0000, 0.1200, 0.0000,
		0.0000, 3.3800, 6.8700, 3.3200, 1.4300, 0.0000, 1.0000, 0.0000, 0.0000, 0.0000,
		1.8700, 6.5600, 3.6300, 2.6900, 1.8100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
		1.7500, 1.3700, 0.0000, 0.0000, 0.1900, 0.1300, 0.0000, 0.7500, 0.0000, 0.0000,
		0.0000,
	}
}

// Expected +DM14 (length=14), 252 entries. Extracted from DirectionalMovementPlusTest.cs talibDmPlus14.
// C# has 13 NaN (indices 0-12) plus vestigial value at index 13 (10.28). Per offset bug fix,
// we use 14 NaN (indices 0-13), drop the vestigial value, and start expected values at index 14.
func testExpectedDMP14() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		9.545714285714290, 10.798877551020400, 10.027529154519000, 9.311277072053310, 9.961185852620930, 9.529672577433720,
		8.848981679045600, 8.216911559113770, 7.629989304891360, 7.084990068827690, 6.578919349625710, 6.108996538938160,
		5.672639643299720, 10.232451097349700, 9.501561733253330, 9.822878752306670, 9.121244555713330, 8.564727087448090,
		8.077960866916090, 9.095963662136360, 9.476251971983770, 8.799376831127780, 8.170849914618660, 7.587217777860180,
		7.045273650870170, 7.262039818665150, 6.743322688760500, 10.291656782420500, 10.526538440819000, 10.304642837903400,
		12.008596920910300, 11.150839997988100, 11.884351426703200, 11.035469181938700, 10.247221383228800, 10.110276998712500,
		9.388114355947290, 8.717534759093910, 8.094853704872920, 7.516649868810570, 6.979746306752670, 7.916192999127480,
		8.725750642046950, 9.417482739043590, 10.179805400540500, 10.702676443359000, 9.938199554547660, 9.228328157794250,
		11.694161860808900, 12.263864585036900, 11.387874257534200, 10.789454667710400, 10.018779334302500, 9.303152238995170,
		8.638641364781230, 8.021595553011140, 7.448624442081780, 6.916579839075940, 6.422538421999080, 5.963785677570580,
		6.602800986315530, 19.566172344435900, 22.073588605547600, 22.591903705151300, 23.228196297640500, 21.569039419237600,
		20.028393746434900, 19.597794193118200, 18.822951750752600, 19.293455197127400, 17.915351254475400, 17.385683307727200,
		17.893848785746700, 18.805716729621900, 17.462451248934600, 19.495133302582200, 26.882623780969100, 24.962436368042800,
		23.179405198896900, 22.023733398975700, 20.450609584763100, 18.989851757280000, 17.633433774617200, 16.373902790716000,
		15.204338305664800, 19.433314140974500, 18.045220273762000, 16.756275968493300, 15.559399113600900, 14.448013462629400,
		14.106012501013000, 13.788440179512100, 18.493551595261200, 17.172583624171100, 16.945970508158900, 15.735544043290400,
		14.611576611626800, 13.567892567939200, 14.908757384515000, 17.033846142763900, 16.317142846852200, 15.151632643505600,
		18.379373168969500, 19.066560799757400, 17.704663599774700, 16.690044771219400, 16.807898716132300, 15.607334522122800,
		15.112524913399800, 20.723058848156900, 19.242840359002900, 18.808351761931200, 19.094898064650400, 17.730976774318300,
		18.464478433295500, 19.085587116631600, 18.972330894015000, 18.117164401585400, 16.823081230043600, 15.621432570754700,
		14.505615958558000, 13.469500532946700, 12.507393352021900, 11.614008112591800, 10.784436104549500, 10.014119239938800,
		9.298825008514630, 10.014623222192200, 11.489292992035600, 10.668629206890200, 9.906584263540870, 9.198971101859380,
		8.541901737440850, 7.931765899052220, 9.495211191977060, 11.816981821121600, 10.972911691041400, 10.189132284538500,
		10.711337121357200, 10.006241612688800, 9.981510068925310, 12.638545064002100, 12.425791845144800, 11.538235284777300,
		10.714075621578900, 9.948784505751860, 9.738157041055300, 9.042574395265630, 8.836676224175230, 10.015485065305600,
		9.300093274926600, 9.065800898146130, 8.798243691135690, 10.669797713197400, 9.907669305111900, 12.199978640461000,
		14.268551594713800, 14.499369337948600, 16.023700099523700, 16.069150092414800, 14.921353657242300, 13.855542681725000,
		14.305861061601800, 13.284013842916000, 12.335155711279100, 11.454073160473500, 10.635925077582500, 9.876216143469460,
		9.170772133221640, 8.515716980848670, 9.717451482216620, 9.023347804915440, 8.378822961707190, 7.780335607299540,
		7.224597349635280, 6.708554681804190, 9.479372204532460, 8.802274189923000, 8.173540319214220, 7.589716010698920,
		8.677593438506130, 8.057765335755700, 7.482210668916000, 6.947767049707720, 8.071497974728600, 7.624962405105120,
		7.080322233311900, 11.134584930932500, 10.339257435865900, 11.220739047589700, 10.919257687047600, 11.019310709401400,
		10.232217087301300, 9.501344438208310, 12.512676978336300, 11.618914337026600, 10.788991884381800, 10.018349606926000,
		9.302753206431250, 8.638270834543300, 9.021251489218780, 9.936876382846010, 10.787099498357000, 10.456592391331500,
		9.709692934807830, 9.136143439464420, 8.483561765216960, 7.877593067701460, 10.694907848579900, 16.800985859395600,
		18.920915440867400, 18.999421480805400, 17.642319946462200, 17.382154236000600, 16.140571790572000, 14.987673805531100,
		13.917125676564600, 14.793045271095700, 20.296399180303200, 22.476656381710100, 23.561180925873600, 23.688239431168400,
		21.996222328942100, 20.425063591160500, 18.966130477506200, 17.611406871970000, 16.353449238257900, 16.935345721239500,
		17.095678169722400, 15.874558300456500, 14.740661278995300, 13.877756901924200, 13.016488551786800, 12.086739369516300,
		11.973400843122300, 11.118157925756400, 10.324003788202400, 9.586574946187900,
	}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestDirectionalMovementPlusConstructor(t *testing.T) {
	t.Parallel()

	dmp, err := NewDirectionalMovementPlus(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dmp.Length() != 14 {
		t.Errorf("expected length 14, got %d", dmp.Length())
	}

	if dmp.IsPrimed() {
		t.Error("should not be primed initially")
	}

	_, err = NewDirectionalMovementPlus(0)
	if err == nil {
		t.Error("expected error for length 0")
	}

	_, err = NewDirectionalMovementPlus(-8)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestDirectionalMovementPlusIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()

	t.Run("length=1", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(1)

		if dmp.IsPrimed() {
			t.Error("should not be primed before updates")
		}

		// First update: not primed yet.
		dmp.Update(high[0], low[0])
		if dmp.IsPrimed() {
			t.Error("[0] should not be primed yet")
		}

		// Second update: should be primed.
		dmp.Update(high[1], low[1])
		if !dmp.IsPrimed() {
			t.Error("[1] should be primed")
		}
	})

	t.Run("length=14", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)

		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])

			if dmp.IsPrimed() {
				t.Errorf("[%d] should not be primed yet", i)
			}
		}

		dmp.Update(high[14], low[14])
		if !dmp.IsPrimed() {
			t.Error("[14] should be primed")
		}
	})
}

func TestDirectionalMovementPlusUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	high := testInputHigh()
	low := testInputLow()
	expected := testExpectedDMP14()
	dmp, _ := NewDirectionalMovementPlus(14)

	for i := range high {
		act := dmp.Update(high[i], low[i])

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if math.IsNaN(act) {
			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}
}

func TestDirectionalMovementPlusLength1(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	high := testInputHigh()
	low := testInputLow()
	expected := testExpectedDMP1()
	dmp, _ := NewDirectionalMovementPlus(1)

	for i := range high {
		act := dmp.Update(high[i], low[i])

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if math.IsNaN(act) {
			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}
}

func TestDirectionalMovementPlusNaNPassthrough(t *testing.T) {
	t.Parallel()

	dmp, _ := NewDirectionalMovementPlus(14)

	if !math.IsNaN(dmp.Update(math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(dmp.Update(1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}

	if !math.IsNaN(dmp.Update(math.NaN(), math.NaN())) {
		t.Error("expected NaN passthrough for NaN high and low")
	}

	if !math.IsNaN(dmp.UpdateSample(math.NaN())) {
		t.Error("expected NaN passthrough for NaN sample")
	}
}

func TestDirectionalMovementPlusHighLowSwap(t *testing.T) {
	t.Parallel()

	// When high < low, they should be swapped internally.
	dmp1, _ := NewDirectionalMovementPlus(1)
	dmp2, _ := NewDirectionalMovementPlus(1)

	// Prime both.
	dmp1.Update(10, 5)
	dmp2.Update(5, 10) // Swapped.

	// Update both with same effective values.
	v1 := dmp1.Update(12, 6)
	v2 := dmp2.Update(6, 12) // Swapped.

	if v1 != v2 {
		t.Errorf("high/low swap should produce same result: %v vs %v", v1, v2)
	}
}

func TestDirectionalMovementPlusZeroInputs(t *testing.T) {
	t.Parallel()

	// Updating with same value repeatedly should produce 0.
	dmp, _ := NewDirectionalMovementPlus(10)

	for i := 0; i < 20; i++ {
		dmp.UpdateSample(0)
	}

	if !dmp.IsPrimed() {
		t.Error("should be primed after 20 updates with length 10")
	}
}

func TestDirectionalMovementPlusUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	high := testInputHigh()
	low := testInputLow()

	check := func(t *testing.T, act core.Output) {
		t.Helper()

		if len(act) != 1 {
			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
		}

		s, ok := act[0].(entities.Scalar)
		if !ok {
			t.Error("output is not scalar")
		}

		if s.Time != tm {
			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
		}
	}

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		b := entities.Bar{Time: tm, High: high[14], Low: low[14]}
		check(t, dmp.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: high[14]}
		check(t, dmp.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: high[14] - 0.5, Ask: high[14] + 0.5}
		check(t, dmp.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: high[14]}
		check(t, dmp.UpdateTrade(&r))
	})
}

func TestDirectionalMovementPlusMetadata(t *testing.T) {
	t.Parallel()

	dmp, _ := NewDirectionalMovementPlus(14)
	act := dmp.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.DirectionalMovementPlus, act.Identifier)
	check("Mnemonic", "+dm", act.Mnemonic)
	check("Description", "Directional Movement Plus", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "+dm", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Directional Movement Plus", act.Outputs[0].Description)
}
```

---

### 6. `go/indicators/welleswilder/directionalmovementplus/output.go` (77 lines)

```go
//nolint:dupl
package directionalmovementplus

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// The scalar value of the directional movement plus.
	Value Output = iota + 1
	outputLast
)

const (
	valueStr   = "value"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Value && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown directional movement plus output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == unknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown directional movement plus output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
```

---

### 7. `go/indicators/welleswilder/directionalmovementplus/params.go` (14 lines)

```go
package directionalmovementplus

// DirectionalMovementPlusParams describes parameters to create an instance of the indicator.
type DirectionalMovementPlusParams struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [DirectionalMovementPlusParams] value populated with conventional defaults.
func DefaultParams() *DirectionalMovementPlusParams {
	return &DirectionalMovementPlusParams{
		Length: 14,
	}
}
```

---

### 8. `go/indicators/welleswilder/directionalmovementminus/directionalmovementminus.go` (209 lines)

```go
package directionalmovementminus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const (
	dmmMnemonic    = "-dm"
	dmmDescription = "Directional Movement Minus"
)

// DirectionalMovementMinus is Welles Wilder's Directional Movement Minus indicator.
//
// The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
//
// The calculation of the directional movement (+DM and −DM) is as follows:
//   - UpMove = today's high − yesterday's high
//   - DownMove = yesterday's low − today's low
//   - if DownMove > UpMove and DownMove > 0, then −DM = DownMove, else −DM = 0
//
// When the length is greater than 1, Wilder's smoothing method is applied:
//
//	Today's −DM(n) = Previous −DM(n) − Previous −DM(n)/n + Today's −DM(1)
//
// The indicator is not primed during the first length updates.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type DirectionalMovementMinus struct {
	mu           sync.RWMutex
	length       int
	noSmoothing  bool
	count        int
	previousHigh float64
	previousLow  float64
	value        float64
	accumulator  float64
	primed       bool
}

// NewDirectionalMovementMinus returns a new instance of the Directional Movement Minus indicator.
func NewDirectionalMovementMinus(length int) (*DirectionalMovementMinus, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	return &DirectionalMovementMinus{
		length:      length,
		noSmoothing: length == 1,
		value:       math.NaN(),
	}, nil
}

// Length returns the length parameter.
func (s *DirectionalMovementMinus) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *DirectionalMovementMinus) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *DirectionalMovementMinus) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DirectionalMovementMinus,
		dmmMnemonic,
		dmmDescription,
		[]core.OutputText{
			{Mnemonic: dmmMnemonic, Description: dmmDescription},
		},
	)
}

// Update updates the Directional Movement Minus given the next bar's high and low values.
func (s *DirectionalMovementMinus) Update(high, low float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	if high < low {
		high, low = low, high
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.noSmoothing {
		if s.primed {
			deltaMinus := s.previousLow - low
			deltaPlus := high - s.previousHigh

			if deltaMinus > 0 && deltaPlus < deltaMinus {
				s.value = deltaMinus
			} else {
				s.value = 0
			}
		} else {
			if s.count > 0 {
				deltaMinus := s.previousLow - low
				deltaPlus := high - s.previousHigh

				if deltaMinus > 0 && deltaPlus < deltaMinus {
					s.value = deltaMinus
				} else {
					s.value = 0
				}

				s.primed = true
			}

			s.count++
		}
	} else {
		if s.primed {
			deltaMinus := s.previousLow - low
			deltaPlus := high - s.previousHigh

			if deltaMinus > 0 && deltaPlus < deltaMinus {
				s.accumulator += -s.accumulator/float64(s.length) + deltaMinus
			} else {
				s.accumulator += -s.accumulator / float64(s.length)
			}

			s.value = s.accumulator
		} else {
			if s.count > 0 && s.length >= s.count {
				deltaMinus := s.previousLow - low
				deltaPlus := high - s.previousHigh

				if s.length > s.count {
					if deltaMinus > 0 && deltaPlus < deltaMinus {
						s.accumulator += deltaMinus
					}
				} else {
					if deltaMinus > 0 && deltaPlus < deltaMinus {
						s.accumulator += -s.accumulator/float64(s.length) + deltaMinus
					} else {
						s.accumulator += -s.accumulator / float64(s.length)
					}

					s.value = s.accumulator
					s.primed = true
				}
			}

			s.count++
		}
	}

	s.previousLow = low
	s.previousHigh = high

	return s.value
}

// UpdateSample updates the Directional Movement Minus using a single sample value
// as a substitute for high and low.
func (s *DirectionalMovementMinus) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DirectionalMovementMinus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalMovementMinus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DirectionalMovementMinus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DirectionalMovementMinus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}
```

---

### 9. `go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go` (452 lines)

```go
//nolint:testpackage
package directionalmovementminus

//nolint: gofumpt
import (
	"math"

...16757 bytes truncated...

The tool call succeeded but the output was truncated. Full output saved to: /home/dev/.local/share/opencode/tool-output/tool_dd2acb2fe0010bLRnstRnexSgd
Use the Task tool to have explore agent process this file with Grep and Read (with offset/limit). Do NOT read the full file yourself - delegate to save context.
```

---

## Assistant (Build · claude-opus-4.6 · 100.3s)

The output was truncated. Let me get the DM- test data I need:

**Tool: task**

**Input:**
```json
{
  "description": "Read DM- test expected data",
  "prompt": "Read the file go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go and return ONLY:\n1. The testExpectedDMM1() function (expected -DM length=1 output array)\n2. The testExpectedDMM14() function (expected -DM length=14 output array)\n\nReturn the full arrays with all values.",
  "subagent_type": "explore"
}
```

**Output:**
```
task_id: ses_22d532f22ffeVRStX3vO5W4gkl (for resuming to continue this task if needed)

<task_result>
Here are the two functions:

**`testExpectedDMM1()` (lines 77–107):**

```go
func testExpectedDMM1() []float64 {
	return []float64{
		math.NaN(),
		0.0000, 0.0000, 0.7500, 0.6850, 0.0000, 1.5000, 2.2500, 0.3100, 0.0000, 0.0000,
		0.0000, 0.2850, 7.2150, 2.0650, 0.0000, 2.0350, 0.1250, 0.0000, 0.0000, 2.2500,
		0.9700, 1.9050, 2.3450, 0.7150, 0.6900, 0.0000, 0.0000, 0.0000, 0.0000, 1.5650,
		0.0000, 0.0000, 0.0000, 0.0000, 1.3100, 1.9100, 1.4050, 0.6850, 0.0000, 0.9100,
		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.7500, 0.4350, 0.0000, 1.0950,
		2.5300, 2.4700, 0.6850, 1.3150, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1900,
		0.9050, 0.0000, 0.0000, 1.5650, 0.0000, 0.0000, 1.4700, 0.5600, 1.1550, 2.0650,
		1.6850, 3.0350, 0.5300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 2.5000, 0.0300,
		0.0000, 0.0000, 0.0000, 2.0350, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
		0.0000, 2.0000, 0.0000, 1.9700, 0.2800, 1.6550, 3.7200, 0.3750, 0.0000, 0.0000,
		0.0000, 3.0000, 1.7500, 0.0000, 0.0000, 0.0000, 1.4400, 0.0000, 3.2500, 0.7500,
		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,
		0.9300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1800, 0.0000, 0.0000, 0.0000,
		0.0000, 0.0000, 1.6900, 0.0000, 1.6900, 7.6300, 0.0000, 3.3700, 0.8800, 0.0000,
		0.0000, 0.0000, 1.1900, 0.2500, 2.5600, 3.6200, 0.1300, 0.0000, 0.0000, 0.0000,
		2.1800, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.0600, 2.2500, 0.1900, 0.0000,
		1.3100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,
		0.0000, 0.0000, 0.0000, 2.3800, 0.0600, 0.0000, 2.3800, 5.8100, 0.0000, 0.0000,
		1.8800, 2.9300, 1.0000, 0.0000, 0.0000, 1.6900, 0.3800, 4.1200, 0.0000, 0.0000,
		0.0000, 4.0000, 4.0000, 0.0000, 3.1200, 5.0000, 0.3100, 0.0000, 0.0000, 0.0000,
		0.0000, 18.0000, 0.0000, 0.0000, 0.0000, 1.8100, 0.0000, 0.0000, 0.0000, 2.6800,
		0.1900, 3.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.8700, 0.0000, 0.8700,
		0.3800, 0.0000, 0.0000, 0.0000, 0.0000, 4.1900, 0.0000, 0.7500, 1.2500, 0.0000,
		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.2500, 4.6900, 1.0000, 0.0000, 2.3700,
		0.0000, 0.0000, 0.8700, 0.0000, 0.0000, 0.0000, 0.0700, 0.0000, 0.3100, 0.1900,
		1.9400,
	}
}
```

**`testExpectedDMM14()` (lines 110–155):**

```go
func testExpectedDMM14() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		14.131785714285700, 13.122372448979600, 14.220060131195300, 13.329341550395700, 12.377245725367400,
		11.493156744984000, 12.922216977485200, 12.969201479093400, 13.947829944872400, 15.296556377381500, 14.918945207568600,
		14.543306264170800, 13.504498673872900, 12.539891625739100, 11.644185081043500, 10.812457575254600, 11.605139177022200,
		10.776200664377700, 10.006472045493600, 9.291724042244060, 8.628029467798060, 9.321741648669630, 10.565902959478900,
		11.216195605230400, 11.100038776285400, 10.307178863693600, 10.480951802001200, 9.732312387572540, 9.037147217031640,
		8.391636701529380, 7.792234079991570, 7.235645931420740, 6.718814079176400, 7.988898787806660, 7.853263160106190,
		7.292315791527170, 7.866436092132370, 9.834547799837210, 11.602080099848800, 11.458360092716800, 11.954905800379900,
		11.100983957495600, 10.308056531960200, 9.571766779677320, 8.888069152557510, 8.253207070231970, 7.853692279501110,
		8.197714259536750, 7.612163240998410, 7.068437295212810, 8.128548916983320, 7.547938280055940, 7.008799831480520,
		7.978171272089050, 7.968301895511260, 8.554137474403320, 10.008127654803100, 10.978261393745700, 13.229099865621000,
		12.814164160933800, 11.898866720867100, 11.048947669376600, 10.259737121564000, 9.526898755737980, 8.846405987470980,
		10.714519845508800, 9.979196999401000, 9.266397213729500, 8.604511698463110, 7.989903720001460, 9.454196311429920,
		8.778896574899210, 8.151832533834980, 7.569558781418200, 7.028876011316900, 6.526813439079980, 6.060612479145690,
		5.627711587778140, 7.225732188651130, 6.709608460890340, 8.200350713683890, 7.894611376992180, 8.985710564349880,
		12.063874095467700, 11.577168802934300, 10.750228174153300, 9.982354733142360, 9.269329395060770, 11.607234438270700,
		12.528146264108500, 11.633278673815100, 10.802330197114000, 10.030735183034400, 10.754254098531900, 9.986093091493950,
		12.522800727815800, 12.378314961543300, 11.494149607147300, 10.673138920922500, 9.910771855142320, 9.202859579775010,
		8.545512466933940, 7.935118719295800, 7.368324525060390, 8.092015630413220, 7.514014513955130, 6.977299191529760,
		7.408920677849070, 6.879712058002710, 6.388304053859660, 5.931996621441110, 5.508282577052460, 5.114833821548720,
		4.929488548580960, 4.577382223682320, 4.250426350562150, 3.946824468379140, 3.664908434923490, 3.403129261000380,
		4.850048599500350, 4.503616556678900, 5.871929659773260, 13.082506112646600, 12.148041390314700, 14.650324148149400,
		14.483872423281500, 13.449310107332900, 12.488645099666200, 11.596599021118600, 11.958270519610200, 11.354108339638000,
		13.103100601092400, 15.787164843871600, 14.789510212166400, 13.733116625583100, 12.752179723755800, 11.841309743487500,
		13.175501904667000, 12.234394625762200, 11.360509295350600, 10.549044345682700, 9.795541178133930, 9.095859665410080,
		11.506155403595100, 12.934287160481100, 12.200409506161100, 11.328951684292400, 11.829740849700100, 10.984759360435800,
		10.200133691833200, 9.471552713845160, 8.795013234284790, 8.166798003264450, 7.583455288745560, 8.291779910978020,
		7.699509917336730, 7.149544923241250, 6.638863143009730, 6.164658632794750, 5.724325873309410, 7.695445453787310,
		7.205770778516790, 6.691072865765590, 8.593139089639470, 13.789343440379500, 12.804390337495300, 11.889791027674200,
		12.920520239983200, 14.927625937127200, 14.861366941618100, 13.799840731502600, 12.814137822109500, 13.588842263387400,
		12.998210673145400, 16.189767053635100, 15.033355121232600, 13.959544041144500, 12.962433752491300, 16.036545627313400,
		18.891078082505300, 17.541715362326300, 19.408735693588700, 23.022397429761000, 21.687940470492300, 20.138801865457200,
		18.700316017924500, 17.364579159501300, 16.124252076679800, 32.972519785488400, 30.617339800810700, 28.430386957895600,
		26.399645032331600, 26.323956101450800, 24.443673522775800, 22.697696842577500, 21.076432782393400, 22.250973297936700,
		20.851618062369800, 22.862216772200500, 21.229201288471900, 19.712829767866800, 18.304770498733400, 16.997286891681000,
		15.783194970846700, 16.525823901500500, 15.345407908536200, 15.119307343640700, 14.419356819095000, 13.389402760588200,
		12.433016849117600, 11.544944217037800, 10.720305344392200, 14.144569248364200, 13.134242873481000, 12.946082668232400,
		13.271362477644400, 12.323408014955500, 11.443164585315800, 10.625795686364700, 9.866810280195770, 9.162038117324650,
		8.507606823230030, 11.149920621570700, 15.043497720030000, 14.968962168599300, 13.899750585127900, 15.276911257618800,
		14.185703310646000, 13.172438788457000, 13.101550303567200, 12.165725281883800, 11.296744904606400, 10.489834554277400,
		9.810560657543280, 9.109806324861620, 8.769105873085790, 8.332741167865380, 9.677545370160700,
	}
}
```
</task_result>
```

---

## Assistant (Build · claude-opus-4.6 · 8.0s)

Now I have all the data. Let me also check if there's an existing HLCV-override pattern indicator in Zig to follow (like WilliamsPercentR or Aroon):

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/tushar_chande/aroon/aroon.zig",
  "limit": 120
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/tushar_chande/aroon/aroon.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: 
9: const indicator_mod = @import("../../core/indicator.zig");
10: const build_metadata_mod = @import("../../core/build_metadata.zig");
11: const identifier_mod = @import("../../core/identifier.zig");
12: const metadata_mod = @import("../../core/metadata.zig");
13: 
14: const OutputArray = indicator_mod.OutputArray;
15: const Identifier = identifier_mod.Identifier;
16: const Metadata = metadata_mod.Metadata;
17: 
18: /// Enumerates the outputs of the Aroon indicator.
19: pub const AroonOutput = enum(u8) {
20:     /// The Aroon Up line.
21:     up = 1,
22:     /// The Aroon Down line.
23:     down = 2,
24:     /// The Aroon Oscillator (Up - Down).
25:     osc = 3,
26: };
27: 
28: /// Parameters to create an instance of the Aroon indicator.
29: pub const AroonParams = struct {
30:     /// The lookback period. Must be >= 2. Default is 14.
31:     length: usize = 14,
32: };
33: 
34: /// Tushar Chande's Aroon indicator.
35: ///
36: /// Measures the number of periods since the highest high and lowest low
37: /// within a lookback window. Produces three outputs:
38: ///   - Up: 100 * (Length - periods since highest high) / Length
39: ///   - Down: 100 * (Length - periods since lowest low) / Length
40: ///   - Osc: Up - Down
41: ///
42: /// The indicator requires bar data (high, low). For scalar, quote, and
43: /// trade updates, the single value substitutes for both.
44: pub const Aroon = struct {
45:     length: usize,
46:     factor: f64,
47:     high_buf: []f64,
48:     low_buf: []f64,
49:     buffer_index: usize,
50:     count: usize,
51:     highest_index: usize,
52:     lowest_index: usize,
53:     up: f64,
54:     down: f64,
55:     osc: f64,
56:     primed: bool,
57:     allocator: std.mem.Allocator,
58:     mnemonic_buf: [32]u8,
59:     mnemonic_len: usize,
60: 
61:     pub const Error = error{
62:         InvalidLength,
63:         OutOfMemory,
64:     };
65: 
66:     pub fn init(allocator: std.mem.Allocator, params: AroonParams) Error!Aroon {
67:         const length = params.length;
68:         if (length < 2) return error.InvalidLength;
69: 
70:         const window_size = length + 1;
71:         const high_buf = allocator.alloc(f64, window_size) catch return error.OutOfMemory;
72:         errdefer allocator.free(high_buf);
73:         const low_buf = allocator.alloc(f64, window_size) catch return error.OutOfMemory;
74: 
75:         var mnemonic_buf: [32]u8 = undefined;
76:         const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "aroon({d})", .{length}) catch
77:             return error.InvalidLength;
78: 
79:         return Aroon{
80:             .length = length,
81:             .factor = 100.0 / @as(f64, @floatFromInt(length)),
82:             .high_buf = high_buf,
83:             .low_buf = low_buf,
84:             .buffer_index = 0,
85:             .count = 0,
86:             .highest_index = 0,
87:             .lowest_index = 0,
88:             .up = math.nan(f64),
89:             .down = math.nan(f64),
90:             .osc = math.nan(f64),
91:             .primed = false,
92:             .allocator = allocator,
93:             .mnemonic_buf = mnemonic_buf,
94:             .mnemonic_len = mnemonic_slice.len,
95:         };
96:     }
97: 
98:     pub fn deinit(self: *Aroon) void {
99:         self.allocator.free(self.high_buf);
100:         self.allocator.free(self.low_buf);
101:     }
102: 
103:     pub fn fixSlices(self: *Aroon) void {
104:         _ = self;
105:     }
106: 
107:     /// Core update given high and low values.
108:     /// Returns [up, down, osc].
109:     pub fn updateHighLow(self: *Aroon, high: f64, low: f64) [3]f64 {
110:         if (math.isNan(high) or math.isNan(low)) {
111:             return .{ math.nan(f64), math.nan(f64), math.nan(f64) };
112:         }
113: 
114:         const window_size = self.length + 1;
115:         const today = self.count;
116: 
117:         // Store in circular buffer.
118:         const pos = self.buffer_index;
119:         self.high_buf[pos] = high;
120:         self.low_buf[pos] = low;

(Showing lines 1-120 of 708. Use offset=121 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 7.0s)

Good, I have the HLCV override pattern. Now let me create the three indicators. These are all simple with no allocator needed (except TR has no params at all). Let me create the directory structure and write all three files.

**Tool: bash**

**Input:**
```json
{
  "command": "mkdir -p zig/src/indicators/welles_wilder/{true_range,directional_movement_plus,directional_movement_minus}",
  "description": "Create welleswilder directories"
}
```

---

## User

---

## Assistant (Compaction · claude-opus-4.6 · 77.4s)



## Goal

Port the `indicators/` module from Go/TypeScript to Zig. Python and Rust ports are complete. Currently porting indicators in small batches by author family. Currently in the middle of porting the **welleswilder** family (12 indicators).

## Instructions

- Go and TS code is the reference implementation. Python is the reference for other modules.
- All ports must match reference output to tolerance specified by Go tests (typically 1e-8).
- Port one language at a time: **Python → Rust → Zig**. Python and Rust are done.
- Port indicators in small batches by author family, starting with simplest, factory last, cmd tools last.
- Zig naming: `snake_case` files, `camelCase` functions/fields, `PascalCase` types/enums, lowercase enum values.
- **Each indicator lives in its own subfolder** (e.g., `patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig`)
- **Author folder naming**: Use `snake_case` matching Rust (e.g., `patrick_mulloy`), NOT Go's `patrickmulloy` style.
- **Output enums are 1-based** (matching Go's `iota+1`).
- **fixSlices() pattern**: After `init()` returns by value, slices point into old stack frame. Must call `fixSlices()` to repoint them. `fixSlices` must be `pub`.
- **Metadata out-pointer pattern**: `getMetadata(out: *Metadata)` returns void.
- **No allocator needed** for EMA-like indicators (no window). SMA and Variance need allocator for ring buffer.
- **MaUnion pattern**: For indicators that support both SMA and EMA sub-indicators, use a tagged union with `update`, `isPrimed`, `deinit` methods.
- **Per-output metadata**: `OutputText` struct has only `mnemonic` and `description` fields (no suffix fields). Each output needs a fully-formed string built with `bufPrint` in `getMetadata`.
- **Metadata fields are plain `[]const u8`**, NOT functions.

## Discoveries

### Key Type Corrections (learned fixing MACD/PPO compilation)
- **Bar/Quote/Trade function types**: Use `bar_component.BarFunc`, `quote_component.QuoteFunc`, `trade_component.TradeFunc` (NOT `BarExtractFn` etc.)
- **Scalar.time is `i64`**, not `f64`
- **OutputArray**: Is a struct with `.values` array and `.len`, NOT a bare array. Use `OutputArray.append(.{ .scalar = .{ .time = t, .value = v } })` to add outputs, `out.slice()[i].scalar.value` to read them.
- **Metadata fields**: `outputs_len` (not `num_outputs`), `outputs_buf[i]` (not `outputs[i]`)
- **Band output type**: Use `band_mod.Band.new(time, lower, upper)` and `Band.empty(time)` for NaN bands. Access via `out.slice()[5].band`.

### HLCV Indicators Pattern (AdvanceDecline, AdvanceDeclineOscillator, Aroon)
- These indicators need HLCV data from bars, not just a single scalar component.
- They override `updateBar` to extract `sample.high, sample.low, sample.close, sample.volume` directly.
- For scalar/quote/trade updates, H=L=C so the range is 0.
- The `update()` method calls `updateHlcv(sample, sample, sample, 1)` or similar.

### StochasticRSI Dependency
- StochasticRelativeStrengthIndex depends on RelativeStrengthIndex (welleswilder family).
- RSI is NOT yet ported to Zig, so StochRSI must be deferred until after welleswilder family is done.

### Welleswilder Family Dependency Chain (12 indicators)
```
Level 0 (no deps):
  1. TrueRange — HLCV override, Update(close,high,low), no params, no allocator
  2. DirectionalMovementPlus — HLCV override, Update(high,low), length param, Wilder smoothing
  3. DirectionalMovementMinus — HLCV override, Update(high,low), length param, Wilder smoothing
  4. RelativeStrengthIndex — LineIndicator, component triple, Wilder-smoothed avg gain/loss
  5. ParabolicStopAndReverse — LineIndicator hybrid, overrides UpdateBar for UpdateHL(high,low)

Level 1:
  6. AverageTrueRange — depends on TR, needs allocator for initial SMA window

Level 2:
  7. NormalizedAverageTrueRange — depends on ATR
  8. DirectionalIndicatorPlus — depends on ATR + DM+
  9. DirectionalIndicatorMinus — depends on ATR + DM-

Level 3:
  10. DirectionalMovementIndex (DX) — depends on DI+, DI-

Level 4:
  11. AverageDirectionalMovementIndex (ADX) — depends on DX

Level 5:
  12. ADXR — depends on ADX, needs allocator for circular buffer
```

### Key welleswilder patterns discovered
- **TrueRange**: No params at all. Fixed mnemonic "tr". Update(close, high, low). First update stores close, returns NaN. Primed from second update. For scalar/quote/trade, v=v=v.
- **DM+/DM-**: Length param (min 1, default 14). Fixed mnemonics "+dm"/"-dm". Update(high, low). High/low swap if high<low. Two code paths: noSmoothing (length==1) vs Wilder smoothing. Accumulator-based. Not primed during first `length` updates.
- **All 12 return exactly 1 output** despite rich metadata.
- **10 of 12 are HLCV-override**, only RSI and SAR use LineIndicator with component triples.
- **Only ATR and ADXR need allocators** (ring buffers).

### DM+/DM- Algorithm Detail
- UpMove = today's high − yesterday's high
- DownMove = yesterday's low − today's low
- +DM: if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else 0
- -DM: if DownMove > UpMove and DownMove > 0, then -DM = DownMove, else 0
- Wilder smoothing (length>1): accumulator = accumulator - accumulator/n + today's DM(1)
- During warmup (count 1..length-1): just sum raw DM values into accumulator
- At count==length: apply first smoothing step, set primed

## Accomplished

### Fully Complete ✅ (662 tests pass)
- **Python port**: All 63 indicators, factory, frequency_response, cmd tools. 1142 tests pass.
- **Rust port**: All 67 indicators, factory, frequency_response, cmd tools. 985 tests pass.
- **Zig Core Types**: All 17 core type files complete.
- **Zig Common Indicators**: All 13 complete (SMA, WMA, TMA, EMA, Momentum, ROC, ROCP, ROCR, Variance, StdDev, APO, LinearRegression, PearsonsCC).
- **Zig Single-Indicator Families**: All 9 complete (BalanceOfPower, OnBalanceVolume, CCI, TEMO, MFI, KAMA, JMA, Stochastic, ATCF).
- **Zig patrick_mulloy**: DEMA, TEMA complete.
- **Zig timtillson**: T2EMA, T3EMA complete.
- **Zig gerald_appel**: MACD, PPO complete.
- **Zig john_bollinger**: BollingerBands, BollingerBandsTrend complete.
- **Zig larry_williams**: WilliamsPercentR, UltimateOscillator complete.
- **Zig marc_chaikin**: AdvanceDecline, AdvanceDeclineOscillator complete.
- **Zig tushar_chande**: Aroon and ChandeMomentumOscillator complete (StochRSI blocked on RSI).
- **662 total Zig tests pass**.

### In Progress — welleswilder Batch 1
- **Directories created** for `welles_wilder/true_range/`, `welles_wilder/directional_movement_plus/`, `welles_wilder/directional_movement_minus/`
- **Go source code fully read** for TrueRange, DM+, DM- (implementations, tests, params, outputs)
- **Test data extracted** from Go tests:
  - TR: 252-entry H/L/C inputs, 252-entry expected TR output, tolerance 1e-3
  - DM+: 252-entry H/L inputs, expected DMP1 (length=1) and DMP14 (length=14), tolerance 1e-8
  - DM-: 252-entry H/L inputs (same as DM+), expected DMM1 (length=1) and DMM14 (length=14), tolerance 1e-8
- **NO Zig source files written yet** for any of these three

### Immediate Next Steps
1. **Write `true_range.zig`** — simplest (no params, no length, no allocator, fixed mnemonic "tr")
2. **Write `directional_movement_plus.zig`** — length param, Wilder smoothing, fixed mnemonic "+dm"
3. **Write `directional_movement_minus.zig`** — nearly identical to DM+ but mirrored logic
4. **Register all three in `indicators.zig`** and test
5. **Continue with RSI** (Batch 2, LineIndicator pattern, unblocks StochRSI)
6. **Then ATR** (Batch 3, depends on TR)
7. **Then NATR, DI+, DI-** (Batch 4)
8. **Then DX → ADX → ADXR** (Batches 5-7)
9. **Then ParabolicSAR** (Batch 8)

### Remaining Work After welleswilder
1. **tushar_chande/StochasticRelativeStrengthIndex** (unblocked after RSI)
2. **johnehlers** (23 indicators)
3. **custom** (GoertzelSpectrum, MaximumEntropySpectrum)
4. Factory, frequency_response, cmd tools

## Relevant files / directories

### Registration file (NEEDS welleswilder imports added):
- `zig/src/indicators/indicators.zig` — Root module; currently has aroon + CMO registered

### Directories created, files NOT yet written:
- `zig/src/indicators/welles_wilder/true_range/` — empty, needs true_range.zig
- `zig/src/indicators/welles_wilder/directional_movement_plus/` — empty, needs directional_movement_plus.zig
- `zig/src/indicators/welles_wilder/directional_movement_minus/` — empty, needs directional_movement_minus.zig

### Go reference implementations (already fully read):
- `go/indicators/welleswilder/truerange/truerange.go` — TR: Update(close,high,low), no params
- `go/indicators/welleswilder/truerange/truerange_test.go` — 252-entry test data (H/L/C inputs + expected TR), tolerance 1e-3
- `go/indicators/welleswilder/directionalmovementplus/directionalmovementplus.go` — +DM: Update(high,low), Wilder smoothing
- `go/indicators/welleswilder/directionalmovementplus/directionalmovementplus_test.go` — 252-entry test (DMP1 + DMP14), tolerance 1e-8
- `go/indicators/welleswilder/directionalmovementplus/params.go` — Length, min 1, default 14
- `go/indicators/welleswilder/directionalmovementminus/directionalmovementminus.go` — -DM: mirrored +DM logic
- `go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go` — 252-entry test (DMM1 + DMM14)
- `go/indicators/welleswilder/directionalmovementminus/params.go` — same as DM+

### Go reference (NOT yet read, upcoming batches):
- `go/indicators/welleswilder/relativestrengthindex/` — RSI (Batch 2)
- `go/indicators/welleswilder/averagetruerange/` — ATR (Batch 3)
- `go/indicators/welleswilder/normalizedaveragetruerange/` — NATR (Batch 4)
- `go/indicators/welleswilder/directionalindicatorplus/` — DI+ (Batch 4)
- `go/indicators/welleswilder/directionalindicatorminus/` — DI- (Batch 4)
- `go/indicators/welleswilder/directionalmovementindex/` — DX (Batch 5)
- `go/indicators/welleswilder/averagedirectionalmovementindex/` — ADX (Batch 6)
- `go/indicators/welleswilder/averagedirectionalmovementindexrating/` — ADXR (Batch 7)
- `go/indicators/welleswilder/parabolicstopandreverse/` — SAR (Batch 8)

### Key pattern references (already ported Zig files):
- `zig/src/indicators/tushar_chande/aroon/aroon.zig` — HLCV override pattern with allocator, ring buffers, multiple outputs
- `zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig` — HLCV override pattern
- `zig/src/indicators/common/momentum/momentum.zig` — LineIndicator + component triple + ring buffer pattern
- `zig/src/indicators/tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig` — LineIndicator + ring buffer + component triple
- `zig/src/indicators/core/indicator.zig` — OutputArray, OutputValue, Indicator vtable interface
- `zig/src/indicators/core/build_metadata.zig` — buildMetadata helper
- `zig/src/indicators/core/identifier.zig` — All identifier enum values

### Build:
- `zig/build.zig` — Single `indicators` module

### Skill files:
- `.opencode/skills/indicator-architecture/SKILL.md`
- `.opencode/skills/indicator-conversion/SKILL.md`

---

## User

---

## Assistant (Build · claude-opus-4.6 · 7.2s)

_Thinking:_

Let me continue with the next steps - writing the three welleswilder indicators (TrueRange, DM+, DM-). I have all the information needed from the context. Let me start by reading the Go reference implementations and existing Zig patterns to write these files.

Let me read the Go source files and a pattern reference file in parallel.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Write true_range.zig",
      "status": "in_progress",
      "priority": "high"
    },
    {
      "content": "Write directional_movement_plus.zig",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "Write directional_movement_minus.zig",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "Register all three in indicators.zig and build.zig, test",
      "status": "pending",
      "priority": "high"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Write true_range.zig",
    "status": "in_progress",
    "priority": "high"
  },
  {
    "content": "Write directional_movement_plus.zig",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "Write directional_movement_minus.zig",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "Register all three in indicators.zig and build.zig, test",
    "status": "pending",
    "priority": "high"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 8.3s)

Let me read the Go reference implementations and an existing HLCV-pattern Zig file for reference.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/welleswilder/truerange/truerange.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/welleswilder/truerange/truerange.go</path>
<type>file</type>
<content>1: package truerange
2: 
3: import (
4: 	"math"
5: 	"sync"
6: 
7: 	"zpano/entities"
8: 	"zpano/indicators/core"
9: )
10: 
11: const (
12: 	trMnemonic    = "tr"
13: 	trDescription = "True Range"
14: )
15: 
16: // TrueRange is Welles Wilder's True Range indicator.
17: //
18: // The True Range is defined as the largest of:
19: //   - the distance from today's high to today's low
20: //   - the distance from yesterday's close to today's high
21: //   - the distance from yesterday's close to today's low
22: //
23: // The first update stores the close and returns NaN (not primed).
24: // The indicator is primed from the second update onward.
25: //
26: // Unlike most indicators, TrueRange requires bar data (high, low, close)
27: // and does not use a single bar component. For scalar, quote, and trade updates,
28: // the single value is used as a substitute for high, low, and close.
29: //
30: // Reference:
31: //
32: // Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
33: type TrueRange struct {
34: 	mu            sync.RWMutex
35: 	previousClose float64
36: 	value         float64
37: 	primed        bool
38: }
39: 
40: // NewTrueRange returns a new instance of the True Range indicator.
41: func NewTrueRange() *TrueRange {
42: 	return &TrueRange{
43: 		previousClose: math.NaN(),
44: 		value:         math.NaN(),
45: 	}
46: }
47: 
48: // IsPrimed indicates whether the indicator is primed.
49: func (s *TrueRange) IsPrimed() bool {
50: 	s.mu.RLock()
51: 	defer s.mu.RUnlock()
52: 
53: 	return s.primed
54: }
55: 
56: // Metadata describes the output data of the indicator.
57: func (s *TrueRange) Metadata() core.Metadata {
58: 	return core.BuildMetadata(
59: 		core.TrueRange,
60: 		trMnemonic,
61: 		trDescription,
62: 		[]core.OutputText{
63: 			{Mnemonic: trMnemonic, Description: trDescription},
64: 		},
65: 	)
66: }
67: 
68: // Update updates the True Range given the next bar's close, high, and low values.
69: func (s *TrueRange) Update(close, high, low float64) float64 {
70: 	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
71: 		return math.NaN()
72: 	}
73: 
74: 	s.mu.Lock()
75: 	defer s.mu.Unlock()
76: 
77: 	if !s.primed {
78: 		if math.IsNaN(s.previousClose) {
79: 			s.previousClose = close
80: 
81: 			return math.NaN()
82: 		}
83: 
84: 		s.primed = true
85: 	}
86: 
87: 	greatest := high - low
88: 
89: 	if temp := math.Abs(high - s.previousClose); greatest < temp {
90: 		greatest = temp
91: 	}
92: 
93: 	if temp := math.Abs(low - s.previousClose); greatest < temp {
94: 		greatest = temp
95: 	}
96: 
97: 	s.value = greatest
98: 	s.previousClose = close
99: 
100: 	return s.value
101: }
102: 
103: // UpdateSample updates the True Range using a single sample value
104: // as a substitute for high, low, and close.
105: func (s *TrueRange) UpdateSample(sample float64) float64 {
106: 	return s.Update(sample, sample, sample)
107: }
108: 
109: // UpdateScalar updates the indicator given the next scalar sample.
110: func (s *TrueRange) UpdateScalar(sample *entities.Scalar) core.Output {
111: 	v := sample.Value
112: 
113: 	output := make([]any, 1)
114: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}
115: 
116: 	return output
117: }
118: 
119: // UpdateBar updates the indicator given the next bar sample.
120: func (s *TrueRange) UpdateBar(sample *entities.Bar) core.Output {
121: 	output := make([]any, 1)
122: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}
123: 
124: 	return output
125: }
126: 
127: // UpdateQuote updates the indicator given the next quote sample.
128: func (s *TrueRange) UpdateQuote(sample *entities.Quote) core.Output {
129: 	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd
130: 
131: 	output := make([]any, 1)
132: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}
133: 
134: 	return output
135: }
136: 
137: // UpdateTrade updates the indicator given the next trade sample.
138: func (s *TrueRange) UpdateTrade(sample *entities.Trade) core.Output {
139: 	v := sample.Price
140: 
141: 	output := make([]any, 1)
142: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}
143: 
144: 	return output
145: }

(End of file - total 145 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/welleswilder/truerange/truerange_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/welleswilder/truerange/truerange_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package truerange
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: // TA-Lib test data (252 entries), extracted programmatically from TrueRangeTest.cs.
16: func testInputHigh() []float64 {
17: 	return []float64{
18: 		93.25, 94.94, 96.375, 96.19, 96.0, 94.72, 95.0, 93.72, 92.47, 92.75,
19: 		96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
20: 		90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
21: 		87.155, 87.25, 87.375, 88.97, 90.0, 89.845, 86.97, 85.94, 84.75, 85.47,
22: 		84.47, 88.5, 89.47, 90.0, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
23: 		90.0, 88.875, 89.0, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
24: 		90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92.0, 91.94,
25: 		89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
26: 		104.5, 105.5, 106.125, 107.94, 106.25, 107.0, 108.75, 110.94, 110.94, 114.22,
27: 		123.0, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0, 118.315,
28: 		116.87, 116.75, 113.87, 114.62, 115.31, 116.0, 121.69, 119.87, 120.87, 116.75,
29: 		116.5, 116.0, 118.31, 121.5, 122.0, 121.44, 125.75, 127.75, 124.19, 124.44,
30: 		125.75, 124.69, 125.31, 132.0, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
31: 		138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
32: 		125.12, 126.5, 128.69, 126.62, 126.69, 126.0, 123.12, 121.87, 124.0, 127.0,
33: 		124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
34: 		124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126.0, 128.5, 126.75, 129.75,
35: 		132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135.0, 132.38, 131.44, 130.88,
36: 		129.63, 127.25, 127.81, 125.0, 126.81, 124.75, 122.81, 122.25, 121.06, 120.0,
37: 		123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109.0,
38: 		108.5, 113.06, 93.0, 94.62, 95.12, 96.0, 95.56, 95.31, 99.0, 98.81,
39: 		96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
40: 		95.0, 94.87, 98.25, 105.12, 108.44, 109.87, 105.0, 106.0, 104.94, 104.5,
41: 		104.44, 106.31, 112.87, 116.5, 119.19, 121.0, 122.12, 111.94, 112.75, 110.19,
42: 		107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110.0, 110.75, 110.5,
43: 		110.5, 109.5,
44: 	}
45: }
46: 
47: func testInputLow() []float64 {
48: 	return []float64{
49: 		90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92.0, 89.75, 89.44, 90.625,
50: 		92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
51: 		87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
52: 		84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
53: 		82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89.0, 88.565, 90.095,
54: 		89.0, 86.47, 84.0, 83.315, 82.0, 83.25, 84.75, 85.28, 87.19, 88.44,
55: 		88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92.0, 90.53, 89.97, 88.815,
56: 		86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
57: 		101.72, 101.72, 103.155, 105.69, 103.655, 104.0, 105.53, 108.53, 108.75, 107.75,
58: 		117.0, 118.0, 116.0, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
59: 		112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
60: 		112.56, 114.0, 114.75, 118.87, 119.0, 119.75, 122.62, 123.0, 121.75, 121.56,
61: 		123.12, 122.19, 122.75, 124.37, 128.0, 129.5, 130.81, 130.63, 132.13, 133.88,
62: 		135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
63: 		122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121.0,
64: 		121.12, 118.94, 119.81, 121.0, 122.0, 124.5, 126.56, 123.5, 121.25, 121.06,
65: 		122.31, 121.0, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128.0,
66: 		128.38, 130.69, 131.63, 134.38, 132.0, 131.94, 131.94, 129.56, 123.75, 126.0,
67: 		126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120.0, 119.62, 115.5, 116.75,
68: 		119.06, 119.06, 115.06, 111.06, 113.12, 110.0, 105.0, 104.69, 103.87, 104.69,
69: 		105.44, 107.0, 89.0, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
70: 		93.69, 93.5, 90.0, 90.19, 90.5, 92.12, 94.12, 94.87, 93.0, 93.87,
71: 		93.0, 92.62, 93.56, 98.37, 104.44, 106.0, 101.81, 104.12, 103.37, 102.12,
72: 		102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
73: 		104.5, 105.75, 108.62, 107.75, 108.06, 108.0, 108.19, 108.12, 109.06, 108.75,
74: 		108.56, 106.62,
75: 	}
76: }
77: 
78: func testInputClose() []float64 {
79: 	return []float64{
80: 		91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
81: 		96.125, 97.25, 98.5, 89.875, 91.0, 92.815, 89.155, 89.345, 91.625, 89.875,
82: 		88.375, 87.625, 84.78, 83.0, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
83: 		85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
84: 		83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89.0, 91.0, 90.5,
85: 		89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
86: 		88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90.0, 89.69,
87: 		88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106.0, 102.5,
88: 		102.405, 104.595, 106.125, 106.0, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
89: 		123.0, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
90: 		116.0, 116.0, 112.0, 113.75, 112.94, 116.0, 120.5, 116.62, 117.0, 115.25,
91: 		114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
92: 		123.12, 122.56, 124.62, 129.25, 131.0, 132.25, 131.0, 132.81, 134.0, 137.38,
93: 		137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129.0, 123.87, 124.81,
94: 		123.0, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
95: 		122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
96: 		124.44, 122.0, 122.37, 122.94, 124.0, 123.19, 124.56, 127.25, 125.87, 128.86,
97: 		132.0, 130.75, 134.75, 135.0, 132.38, 133.31, 131.94, 130.0, 125.37, 130.13,
98: 		127.12, 125.19, 122.0, 125.0, 123.0, 123.5, 120.06, 121.0, 117.75, 119.87,
99: 		122.0, 119.19, 116.37, 113.5, 114.25, 110.0, 105.06, 107.0, 107.87, 107.0,
100: 		107.12, 107.0, 91.0, 93.94, 93.87, 95.5, 93.0, 94.94, 98.25, 96.75,
101: 		94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97.0, 95.0, 95.87, 94.06,
102: 		94.62, 93.75, 98.0, 103.94, 107.87, 106.06, 104.5, 105.0, 104.19, 103.06,
103: 		103.42, 105.27, 111.87, 116.0, 116.62, 118.28, 113.37, 109.0, 109.7, 109.25,
104: 		107.0, 109.19, 110.0, 109.2, 110.12, 108.0, 108.62, 109.75, 109.81, 109.0,
105: 		108.75, 107.87,
106: 	}
107: }
108: 
109: // Full expected TR output (252 entries). First entry is NaN.
110: func testExpectedTR() []float64 {
111: 	return []float64{
112: 		math.NaN(), 3.535, 2.125, 2.69, 3.185, 1.22, 3.0, 3.97, 3.31, 2.435,
113: 		3.78, 3.5, 3.095, 9.685, 4.565, 2.31, 4.5, 1.875, 2.72, 2.5,
114: 		2.845, 1.97, 3.625, 3.22, 2.875, 3.875, 3.19, 5.34, 3.655, 3.155,
115: 		2.75, 2.155, 1.875, 3.44, 2.125, 3.28, 2.315, 3.565, 2.31, 2.03,
116: 		1.94, 5.125, 3.97, 1.47, 3.16, 1.315, 2.22, 2.72, 2.59, 1.655,
117: 		1.5, 2.56, 5.0, 1.935, 1.815, 2.56, 1.875, 2.66, 3.185, 2.185,
118: 		2.5, 1.5, 3.47, 2.28, 4.285, 2.875, 2.03, 2.625, 2.03, 3.125,
119: 		3.0, 3.81, 4.125, 3.375, 3.375, 13.435, 6.81, 5.5, 3.375, 4.25,
120: 		2.78, 3.78, 2.97, 2.25, 2.595, 3.0, 4.125, 2.41, 2.19, 6.47,
121: 		10.25, 5.0, 3.815, 1.815, 2.845, 1.94, 2.095, 4.47, 2.5, 7.72,
122: 		5.505, 2.56, 4.81, 5.18, 3.75, 3.56, 5.69, 4.44, 4.31, 3.69,
123: 		3.94, 2.0, 3.56, 5.63, 3.0, 1.69, 5.0, 4.75, 2.44, 2.88,
124: 		3.19, 2.5, 2.75, 7.63, 3.31, 2.75, 3.07, 2.87, 3.37, 3.56,
125: 		3.31, 3.44, 2.31, 3.63, 2.12, 5.19, 8.57, 2.88, 5.5, 2.82,
126: 		2.37, 3.5, 2.88, 3.76, 2.32, 4.19, 4.93, 3.81, 6.44, 6.0,
127: 		3.32, 3.56, 4.44, 2.81, 3.38, 4.5, 2.0, 6.13, 3.62, 3.31,
128: 		3.12, 3.44, 3.19, 3.81, 2.44, 2.93, 3.13, 3.94, 3.0, 3.88,
129: 		4.31, 3.25, 5.75, 3.31, 3.56, 1.62, 3.06, 2.82, 7.69, 5.51,
130: 		3.88, 2.88, 6.37, 4.56, 5.44, 3.06, 3.5, 2.63, 5.56, 3.25,
131: 		4.19, 3.69, 4.13, 5.31, 3.57, 4.87, 5.87, 2.56, 5.0, 4.31,
132: 		3.06, 6.06, 18.0, 3.62, 3.0, 2.13, 2.75, 2.31, 4.06, 2.44,
133: 		3.12, 2.44, 4.44, 2.75, 3.69, 3.38, 3.44, 2.63, 3.25, 2.5,
134: 		2.0, 2.25, 4.69, 7.12, 4.5, 3.87, 4.25, 1.88, 1.63, 2.38,
135: 		2.19, 2.94, 7.6, 4.63, 3.75, 5.5, 9.87, 5.81, 6.19, 3.32,
136: 		4.75, 3.94, 2.44, 2.69, 2.06, 2.31, 2.44, 1.88, 1.69, 1.75,
137: 		1.94, 2.88,
138: 	}
139: }
140: 
141: func testTime() time.Time {
142: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
143: }
144: 
145: func TestTrueRangeUpdate(t *testing.T) {
146: 	t.Parallel()
147: 
148: 	const tolerance = 1e-3
149: 
150: 	high := testInputHigh()
151: 	low := testInputLow()
152: 	close := testInputClose()
153: 	expected := testExpectedTR()
154: 	tr := NewTrueRange()
155: 
156: 	for i := range close {
157: 		act := tr.Update(close[i], high[i], low[i])
158: 
159: 		if i == 0 {
160: 			if !math.IsNaN(act) {
161: 				t.Errorf("[%d] expected NaN, got %v", i, act)
162: 			}
163: 
164: 			continue
165: 		}
166: 
167: 		if math.IsNaN(act) {
168: 			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
169: 			continue
170: 		}
171: 
172: 		if math.Abs(act-expected[i]) > tolerance {
173: 			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
174: 		}
175: 	}
176: }
177: 
178: func TestTrueRangeNaNPassthrough(t *testing.T) {
179: 	t.Parallel()
180: 
181: 	tr := NewTrueRange()
182: 
183: 	if !math.IsNaN(tr.Update(math.NaN(), 1, 1)) {
184: 		t.Error("expected NaN passthrough for NaN close")
185: 	}
186: 
187: 	if !math.IsNaN(tr.Update(1, math.NaN(), 1)) {
188: 		t.Error("expected NaN passthrough for NaN high")
189: 	}
190: 
191: 	if !math.IsNaN(tr.Update(1, 1, math.NaN())) {
192: 		t.Error("expected NaN passthrough for NaN low")
193: 	}
194: }
195: 
196: func TestTrueRangeIsPrimed(t *testing.T) {
197: 	t.Parallel()
198: 
199: 	high := testInputHigh()
200: 	low := testInputLow()
201: 	close := testInputClose()
202: 	tr := NewTrueRange()
203: 
204: 	if tr.IsPrimed() {
205: 		t.Error("should not be primed before any updates")
206: 	}
207: 
208: 	tr.Update(close[0], high[0], low[0])
209: 
210: 	if tr.IsPrimed() {
211: 		t.Error("[0] should not be primed after first update")
212: 	}
213: 
214: 	tr.Update(close[1], high[1], low[1])
215: 
216: 	if !tr.IsPrimed() {
217: 		t.Error("[1] should be primed after second update")
218: 	}
219: 
220: 	tr.Update(close[2], high[2], low[2])
221: 
222: 	if !tr.IsPrimed() {
223: 		t.Error("[2] should remain primed")
224: 	}
225: }
226: 
227: func TestTrueRangeUpdateSample(t *testing.T) {
228: 	t.Parallel()
229: 
230: 	// When using a single value as H=L=C, TR = |current - previous| after priming.
231: 	tr := NewTrueRange()
232: 
233: 	v := tr.UpdateSample(100.0)
234: 	if !math.IsNaN(v) {
235: 		t.Errorf("expected NaN, got %v", v)
236: 	}
237: 
238: 	v = tr.UpdateSample(105.0)
239: 	if math.Abs(v-5.0) > 1e-10 {
240: 		t.Errorf("expected 5.0, got %v", v)
241: 	}
242: 
243: 	v = tr.UpdateSample(102.0)
244: 	if math.Abs(v-3.0) > 1e-10 {
245: 		t.Errorf("expected 3.0, got %v", v)
246: 	}
247: }
248: 
249: func TestTrueRangeUpdateEntity(t *testing.T) {
250: 	t.Parallel()
251: 
252: 	tm := testTime()
253: 	tr := NewTrueRange()
254: 
255: 	// Prime with first bar.
256: 	tr.Update(100, 105, 95)
257: 
258: 	check := func(act core.Output) {
259: 		t.Helper()
260: 
261: 		if len(act) != 1 {
262: 			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
263: 		}
264: 
265: 		s, ok := act[0].(entities.Scalar)
266: 		if !ok {
267: 			t.Error("output is not scalar")
268: 		}
269: 
270: 		if s.Time != tm {
271: 			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
272: 		}
273: 
274: 		if math.IsNaN(s.Value) {
275: 			t.Error("value should not be NaN after priming")
276: 		}
277: 	}
278: 
279: 	t.Run("update bar", func(t *testing.T) {
280: 		t.Parallel()
281: 
282: 		tr2 := NewTrueRange()
283: 		tr2.Update(100, 105, 95)
284: 
285: 		b := entities.Bar{Time: tm, High: 110, Low: 98, Close: 108}
286: 		check(tr2.UpdateBar(&b))
287: 	})
288: 
289: 	t.Run("update scalar", func(t *testing.T) {
290: 		t.Parallel()
291: 
292: 		tr2 := NewTrueRange()
293: 		tr2.Update(100, 105, 95)
294: 
295: 		s := entities.Scalar{Time: tm, Value: 108}
296: 		check(tr2.UpdateScalar(&s))
297: 	})
298: 
299: 	t.Run("update quote", func(t *testing.T) {
300: 		t.Parallel()
301: 
302: 		tr2 := NewTrueRange()
303: 		tr2.Update(100, 105, 95)
304: 
305: 		q := entities.Quote{Time: tm, Bid: 107, Ask: 109}
306: 		check(tr2.UpdateQuote(&q))
307: 	})
308: 
309: 	t.Run("update trade", func(t *testing.T) {
310: 		t.Parallel()
311: 
312: 		tr2 := NewTrueRange()
313: 		tr2.Update(100, 105, 95)
314: 
315: 		r := entities.Trade{Time: tm, Price: 108}
316: 		check(tr2.UpdateTrade(&r))
317: 	})
318: }
319: 
320: func TestTrueRangeMetadata(t *testing.T) {
321: 	t.Parallel()
322: 
323: 	tr := NewTrueRange()
324: 	act := tr.Metadata()
325: 
326: 	check := func(what string, exp, act any) {
327: 		t.Helper()
328: 
329: 		if exp != act {
330: 			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
331: 		}
332: 	}
333: 
334: 	check("Identifier", core.TrueRange, act.Identifier)
335: 	check("Mnemonic", "tr", act.Mnemonic)
336: 	check("Description", "True Range", act.Description)
337: 	check("len(Outputs)", 1, len(act.Outputs))
338: 	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
339: 	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
340: 	check("Outputs[0].Mnemonic", "tr", act.Outputs[0].Mnemonic)
341: 	check("Outputs[0].Description", "True Range", act.Outputs[0].Description)
342: }

(End of file - total 342 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementplus/directionalmovementplus.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementplus/directionalmovementplus.go</path>
<type>file</type>
<content>1: package directionalmovementplus
2: 
3: import (
4: 	"fmt"
5: 	"math"
6: 	"sync"
7: 
8: 	"zpano/entities"
9: 	"zpano/indicators/core"
10: )
11: 
12: const (
13: 	dmpMnemonic    = "+dm"
14: 	dmpDescription = "Directional Movement Plus"
15: )
16: 
17: // DirectionalMovementPlus is Welles Wilder's Directional Movement Plus indicator.
18: //
19: // The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
20: //
21: // The calculation of the directional movement (+DM and −DM) is as follows:
22: //   - UpMove = today's high − yesterday's high
23: //   - DownMove = yesterday's low − today's low
24: //   - if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0
25: //
26: // When the length is greater than 1, Wilder's smoothing method is applied:
27: //
28: //	Today's +DM(n) = Previous +DM(n) − Previous +DM(n)/n + Today's +DM(1)
29: //
30: // The indicator is not primed during the first length updates.
31: //
32: // Reference:
33: //
34: // Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
35: type DirectionalMovementPlus struct {
36: 	mu           sync.RWMutex
37: 	length       int
38: 	noSmoothing  bool
39: 	count        int
40: 	previousHigh float64
41: 	previousLow  float64
42: 	value        float64
43: 	accumulator  float64
44: 	primed       bool
45: }
46: 
47: // NewDirectionalMovementPlus returns a new instance of the Directional Movement Plus indicator.
48: func NewDirectionalMovementPlus(length int) (*DirectionalMovementPlus, error) {
49: 	if length < 1 {
50: 		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
51: 	}
52: 
53: 	return &DirectionalMovementPlus{
54: 		length:      length,
55: 		noSmoothing: length == 1,
56: 		value:       math.NaN(),
57: 	}, nil
58: }
59: 
60: // Length returns the length parameter.
61: func (s *DirectionalMovementPlus) Length() int {
62: 	return s.length
63: }
64: 
65: // IsPrimed indicates whether the indicator is primed.
66: func (s *DirectionalMovementPlus) IsPrimed() bool {
67: 	s.mu.RLock()
68: 	defer s.mu.RUnlock()
69: 
70: 	return s.primed
71: }
72: 
73: // Metadata describes the output data of the indicator.
74: func (s *DirectionalMovementPlus) Metadata() core.Metadata {
75: 	return core.BuildMetadata(
76: 		core.DirectionalMovementPlus,
77: 		dmpMnemonic,
78: 		dmpDescription,
79: 		[]core.OutputText{
80: 			{Mnemonic: dmpMnemonic, Description: dmpDescription},
81: 		},
82: 	)
83: }
84: 
85: // Update updates the Directional Movement Plus given the next bar's high and low values.
86: func (s *DirectionalMovementPlus) Update(high, low float64) float64 {
87: 	if math.IsNaN(high) || math.IsNaN(low) {
88: 		return math.NaN()
89: 	}
90: 
91: 	if high < low {
92: 		high, low = low, high
93: 	}
94: 
95: 	s.mu.Lock()
96: 	defer s.mu.Unlock()
97: 
98: 	if s.noSmoothing {
99: 		if s.primed {
100: 			deltaPlus := high - s.previousHigh
101: 			deltaMinus := s.previousLow - low
102: 
103: 			if deltaPlus > 0 && deltaPlus > deltaMinus {
104: 				s.value = deltaPlus
105: 			} else {
106: 				s.value = 0
107: 			}
108: 		} else {
109: 			if s.count > 0 {
110: 				deltaPlus := high - s.previousHigh
111: 				deltaMinus := s.previousLow - low
112: 
113: 				if deltaPlus > 0 && deltaPlus > deltaMinus {
114: 					s.value = deltaPlus
115: 				} else {
116: 					s.value = 0
117: 				}
118: 
119: 				s.primed = true
120: 			}
121: 
122: 			s.count++
123: 		}
124: 	} else {
125: 		if s.primed {
126: 			deltaPlus := high - s.previousHigh
127: 			deltaMinus := s.previousLow - low
128: 
129: 			if deltaPlus > 0 && deltaPlus > deltaMinus {
130: 				s.accumulator += -s.accumulator/float64(s.length) + deltaPlus
131: 			} else {
132: 				s.accumulator += -s.accumulator / float64(s.length)
133: 			}
134: 
135: 			s.value = s.accumulator
136: 		} else {
137: 			if s.count > 0 && s.length >= s.count {
138: 				deltaPlus := high - s.previousHigh
139: 				deltaMinus := s.previousLow - low
140: 
141: 				if s.length > s.count {
142: 					if deltaPlus > 0 && deltaPlus > deltaMinus {
143: 						s.accumulator += deltaPlus
144: 					}
145: 				} else {
146: 					if deltaPlus > 0 && deltaPlus > deltaMinus {
147: 						s.accumulator += -s.accumulator/float64(s.length) + deltaPlus
148: 					} else {
149: 						s.accumulator += -s.accumulator / float64(s.length)
150: 					}
151: 
152: 					s.value = s.accumulator
153: 					s.primed = true
154: 				}
155: 			}
156: 
157: 			s.count++
158: 		}
159: 	}
160: 
161: 	s.previousLow = low
162: 	s.previousHigh = high
163: 
164: 	return s.value
165: }
166: 
167: // UpdateSample updates the Directional Movement Plus using a single sample value
168: // as a substitute for high and low.
169: func (s *DirectionalMovementPlus) UpdateSample(sample float64) float64 {
170: 	return s.Update(sample, sample)
171: }
172: 
173: // UpdateScalar updates the indicator given the next scalar sample.
174: func (s *DirectionalMovementPlus) UpdateScalar(sample *entities.Scalar) core.Output {
175: 	v := sample.Value
176: 
177: 	output := make([]any, 1)
178: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}
179: 
180: 	return output
181: }
182: 
183: // UpdateBar updates the indicator given the next bar sample.
184: func (s *DirectionalMovementPlus) UpdateBar(sample *entities.Bar) core.Output {
185: 	output := make([]any, 1)
186: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.High, sample.Low)}
187: 
188: 	return output
189: }
190: 
191: // UpdateQuote updates the indicator given the next quote sample.
192: func (s *DirectionalMovementPlus) UpdateQuote(sample *entities.Quote) core.Output {
193: 	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd
194: 
195: 	output := make([]any, 1)
196: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}
197: 
198: 	return output
199: }
200: 
201: // UpdateTrade updates the indicator given the next trade sample.
202: func (s *DirectionalMovementPlus) UpdateTrade(sample *entities.Trade) core.Output {
203: 	v := sample.Price
204: 
205: 	output := make([]any, 1)
206: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}
207: 
208: 	return output
209: }

(End of file - total 209 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementplus/directionalmovementplus_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementplus/directionalmovementplus_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package directionalmovementplus
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: // TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementPlusTest.cs.
16: func testInputHigh() []float64 {
17: 	return []float64{
18: 		93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
19: 		99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000, 90.345000, 88.500000,
20: 		88.250000, 85.500000, 84.440000, 84.750000, 84.440000, 89.405000, 88.125000, 89.125000, 87.155000, 87.250000, 87.375000,
21: 		88.970000, 90.000000, 89.845000, 86.970000, 85.940000, 84.750000, 85.470000, 84.470000, 88.500000, 89.470000, 90.000000,
22: 		92.440000, 91.440000, 92.970000, 91.720000, 91.155000, 91.750000, 90.000000, 88.875000, 89.000000, 85.250000, 83.815000,
23: 		85.250000, 86.625000, 87.940000, 89.375000, 90.625000, 90.750000, 88.845000, 91.970000, 93.375000, 93.815000, 94.030000,
24: 		94.030000, 91.815000, 92.000000, 91.940000, 89.750000, 88.750000, 86.155000, 84.875000, 85.940000, 99.375000, 103.280000,
25: 		105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000,
26: 		110.940000, 110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000,
27: 		115.345000, 113.000000, 118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000,
28: 		119.870000, 120.870000, 116.750000, 116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000,
29: 		127.750000, 124.190000, 124.440000, 125.750000, 124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000,
30: 		133.500000, 135.500000, 137.440000, 138.690000, 139.190000, 138.500000, 138.130000, 137.500000, 138.880000, 132.130000,
31: 		129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000, 126.620000, 126.690000, 126.000000, 123.120000,
32: 		121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000, 124.500000, 127.870000, 128.560000,
33: 		129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000, 125.620000, 126.000000,
34: 		128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000, 135.000000,
35: 		132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,
36: 		122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000,
37: 		107.250000, 108.870000, 109.000000, 108.500000, 113.060000, 93.000000, 94.620000, 95.120000, 96.000000, 95.560000,
38: 		95.310000, 99.000000, 98.810000, 96.810000, 95.940000, 94.440000, 92.940000, 93.940000, 95.500000, 97.060000, 97.500000,
39: 		96.250000, 96.370000, 95.000000, 94.870000, 98.250000, 105.120000, 108.440000, 109.870000, 105.000000, 106.000000,
40: 		104.940000, 104.500000, 104.440000, 106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000,
41: 		112.750000, 110.190000, 107.940000, 109.690000, 111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000,
42: 		110.750000, 110.500000, 110.500000, 109.500000,
43: 	}
44: }
45: 
46: func testInputLow() []float64 {
47: 	return []float64{
48: 		90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,
49: 		96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000, 87.500000, 86.530000,
50: 		84.625000, 82.280000, 81.565000, 80.875000, 81.250000, 84.065000, 85.595000, 85.970000, 84.405000, 85.095000, 85.500000,
51: 		85.530000, 87.875000, 86.565000, 84.655000, 83.250000, 82.565000, 83.440000, 82.530000, 85.065000, 86.875000, 88.530000,
52: 		89.280000, 90.125000, 90.750000, 89.000000, 88.565000, 90.095000, 89.000000, 86.470000, 84.000000, 83.315000, 82.000000,
53: 		83.250000, 84.750000, 85.280000, 87.190000, 88.440000, 88.250000, 87.345000, 89.280000, 91.095000, 89.530000, 91.155000,
54: 		92.000000, 90.530000, 89.970000, 88.815000, 86.750000, 85.065000, 82.030000, 81.500000, 82.565000, 96.345000, 96.470000,
55: 		101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000,
56: 		108.530000, 108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000,
57: 		110.875000, 110.500000, 110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000,
58: 		116.060000, 116.560000, 113.310000, 112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000,
59: 		123.000000, 121.750000, 121.560000, 123.120000, 122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000,
60: 		130.630000, 132.130000, 133.880000, 135.380000, 135.750000, 136.190000, 134.500000, 135.380000, 133.690000, 126.060000,
61: 		126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000, 124.620000, 124.370000, 121.810000, 118.190000,
62: 		118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000, 122.000000, 124.500000, 126.560000,
63: 		123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000, 122.690000, 122.870000,
64: 		125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000, 131.940000,
65: 		129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,
66: 		119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000,
67: 		104.690000, 103.870000, 104.690000, 105.440000, 107.000000, 89.000000, 92.500000, 92.120000, 94.620000, 92.810000,
68: 		94.250000, 96.250000, 96.370000, 93.690000, 93.500000, 90.000000, 90.190000, 90.500000, 92.120000, 94.120000, 94.870000,
69: 		93.000000, 93.870000, 93.000000, 92.620000, 93.560000, 98.370000, 104.440000, 106.000000, 101.810000, 104.120000,
70: 		103.370000, 102.120000, 102.250000, 103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000,
71: 		106.560000, 106.870000, 104.500000, 105.750000, 108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000,
72: 		109.060000, 108.750000, 108.560000, 106.620000,
73: 	}
74: }
75: 
76: // Expected +DM1 (length=1), 252 entries. Extracted from DirectionalMovementPlusTest.cs talibDmPlus1.
77: func testExpectedDMP1() []float64 {
78: 	return []float64{
79: 		math.NaN(),
80: 		1.6900, 1.4350, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.2800, 3.5000,
81: 		3.3750, 0.0000, 0.0000, 0.0000, 1.9350, 0.0000, 0.0000, 1.3150, 0.2800, 0.0000,
82: 		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 4.9650, 0.0000, 1.0000, 0.0000,
83: 		0.0950, 0.1250, 1.5950, 1.0300, 0.0000, 0.0000, 0.0000, 0.0000, 0.7200, 0.0000,
84: 		4.0300, 0.9700, 0.5300, 2.4400, 0.0000, 1.5300, 0.0000, 0.0000, 0.5950, 0.0000,
85: 		0.0000, 0.0000, 0.0000, 0.0000, 1.4350, 1.3750, 1.3150, 1.4350, 1.2500, 0.0000,
86: 		0.0000, 3.1250, 1.4050, 0.0000, 0.2150, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
87: 		0.0000, 0.0000, 0.0000, 1.0650, 13.4350, 3.9050, 2.0950, 2.2500, 0.0000, 0.0000,
88: 		1.0000, 0.6250, 1.8150, 0.0000, 0.7500, 1.7500, 2.1900, 0.0000, 3.2800, 8.7800,
89: 		0.0000, 0.0000, 0.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 5.3150, 0.0000,
90: 		0.0000, 0.0000, 0.0000, 0.6900, 0.6900, 5.6900, 0.0000, 1.0000, 0.0000, 0.0000,
91: 		0.0000, 2.3100, 3.1900, 0.5000, 0.0000, 4.3100, 2.0000, 0.0000, 0.2500, 1.3100,
92: 		0.0000, 0.6200, 6.6900, 0.0000, 0.9400, 1.6300, 0.0000, 2.0000, 1.9400, 1.2500,
93: 		0.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
94: 		1.3800, 2.1900, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 2.1300, 3.0000, 0.0000,
95: 		0.0000, 1.2500, 0.0600, 0.6900, 3.3700, 0.6900, 0.0000, 0.0000, 0.0000, 0.5000,
96: 		0.0000, 0.4400, 1.8100, 0.0000, 0.4300, 0.3800, 2.5000, 0.0000, 3.0000, 2.9400,
97: 		1.2500, 2.5600, 1.1900, 0.0000, 0.0000, 1.4400, 0.0000, 0.0000, 0.0000, 0.0000,
98: 		0.0000, 0.0000, 0.0000, 1.8100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.2500,
99: 		0.0000, 0.0000, 0.0000, 1.6300, 0.0000, 0.0000, 0.0000, 1.6200, 0.1300, 0.0000,
100: 		4.5600, 0.0000, 1.6200, 0.5000, 0.8800, 0.0000, 0.0000, 3.6900, 0.0000, 0.0000,
101: 		0.0000, 0.0000, 0.0000, 1.0000, 1.5600, 1.5600, 0.4400, 0.0000, 0.1200, 0.0000,
102: 		0.0000, 3.3800, 6.8700, 3.3200, 1.4300, 0.0000, 1.0000, 0.0000, 0.0000, 0.0000,
103: 		1.8700, 6.5600, 3.6300, 2.6900, 1.8100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
104: 		1.7500, 1.3700, 0.0000, 0.0000, 0.1900, 0.1300, 0.0000, 0.7500, 0.0000, 0.0000,
105: 		0.0000,
106: 	}
107: }
108: 
109: // Expected +DM14 (length=14), 252 entries. Extracted from DirectionalMovementPlusTest.cs talibDmPlus14.
110: // C# has 13 NaN (indices 0-12) plus vestigial value at index 13 (10.28). Per offset bug fix,
111: // we use 14 NaN (indices 0-13), drop the vestigial value, and start expected values at index 14.
112: func testExpectedDMP14() []float64 {
113: 	return []float64{
114: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
115: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
116: 		9.545714285714290, 10.798877551020400, 10.027529154519000, 9.311277072053310, 9.961185852620930, 9.529672577433720,
117: 		8.848981679045600, 8.216911559113770, 7.629989304891360, 7.084990068827690, 6.578919349625710, 6.108996538938160,
118: 		5.672639643299720, 10.232451097349700, 9.501561733253330, 9.822878752306670, 9.121244555713330, 8.564727087448090,
119: 		8.077960866916090, 9.095963662136360, 9.476251971983770, 8.799376831127780, 8.170849914618660, 7.587217777860180,
120: 		7.045273650870170, 7.262039818665150, 6.743322688760500, 10.291656782420500, 10.526538440819000, 10.304642837903400,
121: 		12.008596920910300, 11.150839997988100, 11.884351426703200, 11.035469181938700, 10.247221383228800, 10.110276998712500,
122: 		9.388114355947290, 8.717534759093910, 8.094853704872920, 7.516649868810570, 6.979746306752670, 7.916192999127480,
123: 		8.725750642046950, 9.417482739043590, 10.179805400540500, 10.702676443359000, 9.938199554547660, 9.228328157794250,
124: 		11.694161860808900, 12.263864585036900, 11.387874257534200, 10.789454667710400, 10.018779334302500, 9.303152238995170,
125: 		8.638641364781230, 8.021595553011140, 7.448624442081780, 6.916579839075940, 6.422538421999080, 5.963785677570580,
126: 		6.602800986315530, 19.566172344435900, 22.073588605547600, 22.591903705151300, 23.228196297640500, 21.569039419237600,
127: 		20.028393746434900, 19.597794193118200, 18.822951750752600, 19.293455197127400, 17.915351254475400, 17.385683307727200,
128: 		17.893848785746700, 18.805716729621900, 17.462451248934600, 19.495133302582200, 26.882623780969100, 24.962436368042800,
129: 		23.179405198896900, 22.023733398975700, 20.450609584763100, 18.989851757280000, 17.633433774617200, 16.373902790716000,
130: 		15.204338305664800, 19.433314140974500, 18.045220273762000, 16.756275968493300, 15.559399113600900, 14.448013462629400,
131: 		14.106012501013000, 13.788440179512100, 18.493551595261200, 17.172583624171100, 16.945970508158900, 15.735544043290400,
132: 		14.611576611626800, 13.567892567939200, 14.908757384515000, 17.033846142763900, 16.317142846852200, 15.151632643505600,
133: 		18.379373168969500, 19.066560799757400, 17.704663599774700, 16.690044771219400, 16.807898716132300, 15.607334522122800,
134: 		15.112524913399800, 20.723058848156900, 19.242840359002900, 18.808351761931200, 19.094898064650400, 17.730976774318300,
135: 		18.464478433295500, 19.085587116631600, 18.972330894015000, 18.117164401585400, 16.823081230043600, 15.621432570754700,
136: 		14.505615958558000, 13.469500532946700, 12.507393352021900, 11.614008112591800, 10.784436104549500, 10.014119239938800,
137: 		9.298825008514630, 10.014623222192200, 11.489292992035600, 10.668629206890200, 9.906584263540870, 9.198971101859380,
138: 		8.541901737440850, 7.931765899052220, 9.495211191977060, 11.816981821121600, 10.972911691041400, 10.189132284538500,
139: 		10.711337121357200, 10.006241612688800, 9.981510068925310, 12.638545064002100, 12.425791845144800, 11.538235284777300,
140: 		10.714075621578900, 9.948784505751860, 9.738157041055300, 9.042574395265630, 8.836676224175230, 10.015485065305600,
141: 		9.300093274926600, 9.065800898146130, 8.798243691135690, 10.669797713197400, 9.907669305111900, 12.199978640461000,
142: 		14.268551594713800, 14.499369337948600, 16.023700099523700, 16.069150092414800, 14.921353657242300, 13.855542681725000,
143: 		14.305861061601800, 13.284013842916000, 12.335155711279100, 11.454073160473500, 10.635925077582500, 9.876216143469460,
144: 		9.170772133221640, 8.515716980848670, 9.717451482216620, 9.023347804915440, 8.378822961707190, 7.780335607299540,
145: 		7.224597349635280, 6.708554681804190, 9.479372204532460, 8.802274189923000, 8.173540319214220, 7.589716010698920,
146: 		8.677593438506130, 8.057765335755700, 7.482210668916000, 6.947767049707720, 8.071497974728600, 7.624962405105120,
147: 		7.080322233311900, 11.134584930932500, 10.339257435865900, 11.220739047589700, 10.919257687047600, 11.019310709401400,
148: 		10.232217087301300, 9.501344438208310, 12.512676978336300, 11.618914337026600, 10.788991884381800, 10.018349606926000,
149: 		9.302753206431250, 8.638270834543300, 9.021251489218780, 9.936876382846010, 10.787099498357000, 10.456592391331500,
150: 		9.709692934807830, 9.136143439464420, 8.483561765216960, 7.877593067701460, 10.694907848579900, 16.800985859395600,
151: 		18.920915440867400, 18.999421480805400, 17.642319946462200, 17.382154236000600, 16.140571790572000, 14.987673805531100,
152: 		13.917125676564600, 14.793045271095700, 20.296399180303200, 22.476656381710100, 23.561180925873600, 23.688239431168400,
153: 		21.996222328942100, 20.425063591160500, 18.966130477506200, 17.611406871970000, 16.353449238257900, 16.935345721239500,
154: 		17.095678169722400, 15.874558300456500, 14.740661278995300, 13.877756901924200, 13.016488551786800, 12.086739369516300,
155: 		11.973400843122300, 11.118157925756400, 10.324003788202400, 9.586574946187900,
156: 	}
157: }
158: 
159: func testTime() time.Time {
160: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
161: }
162: 
163: func TestDirectionalMovementPlusConstructor(t *testing.T) {
164: 	t.Parallel()
165: 
166: 	dmp, err := NewDirectionalMovementPlus(14)
167: 	if err != nil {
168: 		t.Fatalf("unexpected error: %v", err)
169: 	}
170: 
171: 	if dmp.Length() != 14 {
172: 		t.Errorf("expected length 14, got %d", dmp.Length())
173: 	}
174: 
175: 	if dmp.IsPrimed() {
176: 		t.Error("should not be primed initially")
177: 	}
178: 
179: 	_, err = NewDirectionalMovementPlus(0)
180: 	if err == nil {
181: 		t.Error("expected error for length 0")
182: 	}
183: 
184: 	_, err = NewDirectionalMovementPlus(-8)
185: 	if err == nil {
186: 		t.Error("expected error for negative length")
187: 	}
188: }
189: 
190: func TestDirectionalMovementPlusIsPrimed(t *testing.T) {
191: 	t.Parallel()
192: 
193: 	high := testInputHigh()
194: 	low := testInputLow()
195: 
196: 	t.Run("length=1", func(t *testing.T) {
197: 		t.Parallel()
198: 
199: 		dmp, _ := NewDirectionalMovementPlus(1)
200: 
201: 		if dmp.IsPrimed() {
202: 			t.Error("should not be primed before updates")
203: 		}
204: 
205: 		// First update: not primed yet.
206: 		dmp.Update(high[0], low[0])
207: 		if dmp.IsPrimed() {
208: 			t.Error("[0] should not be primed yet")
209: 		}
210: 
211: 		// Second update: should be primed.
212: 		dmp.Update(high[1], low[1])
213: 		if !dmp.IsPrimed() {
214: 			t.Error("[1] should be primed")
215: 		}
216: 	})
217: 
218: 	t.Run("length=14", func(t *testing.T) {
219: 		t.Parallel()
220: 
221: 		dmp, _ := NewDirectionalMovementPlus(14)
222: 
223: 		for i := 0; i < 14; i++ {
224: 			dmp.Update(high[i], low[i])
225: 
226: 			if dmp.IsPrimed() {
227: 				t.Errorf("[%d] should not be primed yet", i)
228: 			}
229: 		}
230: 
231: 		dmp.Update(high[14], low[14])
232: 		if !dmp.IsPrimed() {
233: 			t.Error("[14] should be primed")
234: 		}
235: 	})
236: }
237: 
238: func TestDirectionalMovementPlusUpdate(t *testing.T) {
239: 	t.Parallel()
240: 
241: 	const tolerance = 1e-8
242: 
243: 	high := testInputHigh()
244: 	low := testInputLow()
245: 	expected := testExpectedDMP14()
246: 	dmp, _ := NewDirectionalMovementPlus(14)
247: 
248: 	for i := range high {
249: 		act := dmp.Update(high[i], low[i])
250: 
251: 		if math.IsNaN(expected[i]) {
252: 			if !math.IsNaN(act) {
253: 				t.Errorf("[%d] expected NaN, got %v", i, act)
254: 			}
255: 
256: 			continue
257: 		}
258: 
259: 		if math.IsNaN(act) {
260: 			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
261: 			continue
262: 		}
263: 
264: 		if math.Abs(act-expected[i]) > tolerance {
265: 			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
266: 		}
267: 	}
268: }
269: 
270: func TestDirectionalMovementPlusLength1(t *testing.T) {
271: 	t.Parallel()
272: 
273: 	const tolerance = 1e-8
274: 
275: 	high := testInputHigh()
276: 	low := testInputLow()
277: 	expected := testExpectedDMP1()
278: 	dmp, _ := NewDirectionalMovementPlus(1)
279: 
280: 	for i := range high {
281: 		act := dmp.Update(high[i], low[i])
282: 
283: 		if math.IsNaN(expected[i]) {
284: 			if !math.IsNaN(act) {
285: 				t.Errorf("[%d] expected NaN, got %v", i, act)
286: 			}
287: 
288: 			continue
289: 		}
290: 
291: 		if math.IsNaN(act) {
292: 			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
293: 			continue
294: 		}
295: 
296: 		if math.Abs(act-expected[i]) > tolerance {
297: 			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
298: 		}
299: 	}
300: }
301: 
302: func TestDirectionalMovementPlusNaNPassthrough(t *testing.T) {
303: 	t.Parallel()
304: 
305: 	dmp, _ := NewDirectionalMovementPlus(14)
306: 
307: 	if !math.IsNaN(dmp.Update(math.NaN(), 1)) {
308: 		t.Error("expected NaN passthrough for NaN high")
309: 	}
310: 
311: 	if !math.IsNaN(dmp.Update(1, math.NaN())) {
312: 		t.Error("expected NaN passthrough for NaN low")
313: 	}
314: 
315: 	if !math.IsNaN(dmp.Update(math.NaN(), math.NaN())) {
316: 		t.Error("expected NaN passthrough for NaN high and low")
317: 	}
318: 
319: 	if !math.IsNaN(dmp.UpdateSample(math.NaN())) {
320: 		t.Error("expected NaN passthrough for NaN sample")
321: 	}
322: }
323: 
324: func TestDirectionalMovementPlusHighLowSwap(t *testing.T) {
325: 	t.Parallel()
326: 
327: 	// When high < low, they should be swapped internally.
328: 	dmp1, _ := NewDirectionalMovementPlus(1)
329: 	dmp2, _ := NewDirectionalMovementPlus(1)
330: 
331: 	// Prime both.
332: 	dmp1.Update(10, 5)
333: 	dmp2.Update(5, 10) // Swapped.
334: 
335: 	// Update both with same effective values.
336: 	v1 := dmp1.Update(12, 6)
337: 	v2 := dmp2.Update(6, 12) // Swapped.
338: 
339: 	if v1 != v2 {
340: 		t.Errorf("high/low swap should produce same result: %v vs %v", v1, v2)
341: 	}
342: }
343: 
344: func TestDirectionalMovementPlusZeroInputs(t *testing.T) {
345: 	t.Parallel()
346: 
347: 	// Updating with same value repeatedly should produce 0.
348: 	dmp, _ := NewDirectionalMovementPlus(10)
349: 
350: 	for i := 0; i < 20; i++ {
351: 		dmp.UpdateSample(0)
352: 	}
353: 
354: 	if !dmp.IsPrimed() {
355: 		t.Error("should be primed after 20 updates with length 10")
356: 	}
357: }
358: 
359: func TestDirectionalMovementPlusUpdateEntity(t *testing.T) {
360: 	t.Parallel()
361: 
362: 	tm := testTime()
363: 	high := testInputHigh()
364: 	low := testInputLow()
365: 
366: 	check := func(t *testing.T, act core.Output) {
367: 		t.Helper()
368: 
369: 		if len(act) != 1 {
370: 			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
371: 		}
372: 
373: 		s, ok := act[0].(entities.Scalar)
374: 		if !ok {
375: 			t.Error("output is not scalar")
376: 		}
377: 
378: 		if s.Time != tm {
379: 			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
380: 		}
381: 	}
382: 
383: 	t.Run("update bar", func(t *testing.T) {
384: 		t.Parallel()
385: 
386: 		dmp, _ := NewDirectionalMovementPlus(14)
387: 		for i := 0; i < 14; i++ {
388: 			dmp.Update(high[i], low[i])
389: 		}
390: 
391: 		b := entities.Bar{Time: tm, High: high[14], Low: low[14]}
392: 		check(t, dmp.UpdateBar(&b))
393: 	})
394: 
395: 	t.Run("update scalar", func(t *testing.T) {
396: 		t.Parallel()
397: 
398: 		dmp, _ := NewDirectionalMovementPlus(14)
399: 		for i := 0; i < 14; i++ {
400: 			dmp.Update(high[i], low[i])
401: 		}
402: 
403: 		s := entities.Scalar{Time: tm, Value: high[14]}
404: 		check(t, dmp.UpdateScalar(&s))
405: 	})
406: 
407: 	t.Run("update quote", func(t *testing.T) {
408: 		t.Parallel()
409: 
410: 		dmp, _ := NewDirectionalMovementPlus(14)
411: 		for i := 0; i < 14; i++ {
412: 			dmp.Update(high[i], low[i])
413: 		}
414: 
415: 		q := entities.Quote{Time: tm, Bid: high[14] - 0.5, Ask: high[14] + 0.5}
416: 		check(t, dmp.UpdateQuote(&q))
417: 	})
418: 
419: 	t.Run("update trade", func(t *testing.T) {
420: 		t.Parallel()
421: 
422: 		dmp, _ := NewDirectionalMovementPlus(14)
423: 		for i := 0; i < 14; i++ {
424: 			dmp.Update(high[i], low[i])
425: 		}
426: 
427: 		r := entities.Trade{Time: tm, Price: high[14]}
428: 		check(t, dmp.UpdateTrade(&r))
429: 	})
430: }
431: 
432: func TestDirectionalMovementPlusMetadata(t *testing.T) {
433: 	t.Parallel()
434: 
435: 	dmp, _ := NewDirectionalMovementPlus(14)
436: 	act := dmp.Metadata()
437: 
438: 	check := func(what string, exp, act any) {
439: 		t.Helper()
440: 
441: 		if exp != act {
442: 			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
443: 		}
444: 	}
445: 
446: 	check("Identifier", core.DirectionalMovementPlus, act.Identifier)
447: 	check("Mnemonic", "+dm", act.Mnemonic)
448: 	check("Description", "Directional Movement Plus", act.Description)
449: 	check("len(Outputs)", 1, len(act.Outputs))
450: 	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
451: 	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
452: 	check("Outputs[0].Mnemonic", "+dm", act.Outputs[0].Mnemonic)
453: 	check("Outputs[0].Description", "Directional Movement Plus", act.Outputs[0].Description)
454: }

(End of file - total 454 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementminus/directionalmovementminus.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementminus/directionalmovementminus.go</path>
<type>file</type>
<content>1: package directionalmovementminus
2: 
3: import (
4: 	"fmt"
5: 	"math"
6: 	"sync"
7: 
8: 	"zpano/entities"
9: 	"zpano/indicators/core"
10: )
11: 
12: const (
13: 	dmmMnemonic    = "-dm"
14: 	dmmDescription = "Directional Movement Minus"
15: )
16: 
17: // DirectionalMovementMinus is Welles Wilder's Directional Movement Minus indicator.
18: //
19: // The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
20: //
21: // The calculation of the directional movement (+DM and −DM) is as follows:
22: //   - UpMove = today's high − yesterday's high
23: //   - DownMove = yesterday's low − today's low
24: //   - if DownMove > UpMove and DownMove > 0, then −DM = DownMove, else −DM = 0
25: //
26: // When the length is greater than 1, Wilder's smoothing method is applied:
27: //
28: //	Today's −DM(n) = Previous −DM(n) − Previous −DM(n)/n + Today's −DM(1)
29: //
30: // The indicator is not primed during the first length updates.
31: //
32: // Reference:
33: //
34: // Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
35: type DirectionalMovementMinus struct {
36: 	mu           sync.RWMutex
37: 	length       int
38: 	noSmoothing  bool
39: 	count        int
40: 	previousHigh float64
41: 	previousLow  float64
42: 	value        float64
43: 	accumulator  float64
44: 	primed       bool
45: }
46: 
47: // NewDirectionalMovementMinus returns a new instance of the Directional Movement Minus indicator.
48: func NewDirectionalMovementMinus(length int) (*DirectionalMovementMinus, error) {
49: 	if length < 1 {
50: 		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
51: 	}
52: 
53: 	return &DirectionalMovementMinus{
54: 		length:      length,
55: 		noSmoothing: length == 1,
56: 		value:       math.NaN(),
57: 	}, nil
58: }
59: 
60: // Length returns the length parameter.
61: func (s *DirectionalMovementMinus) Length() int {
62: 	return s.length
63: }
64: 
65: // IsPrimed indicates whether the indicator is primed.
66: func (s *DirectionalMovementMinus) IsPrimed() bool {
67: 	s.mu.RLock()
68: 	defer s.mu.RUnlock()
69: 
70: 	return s.primed
71: }
72: 
73: // Metadata describes the output data of the indicator.
74: func (s *DirectionalMovementMinus) Metadata() core.Metadata {
75: 	return core.BuildMetadata(
76: 		core.DirectionalMovementMinus,
77: 		dmmMnemonic,
78: 		dmmDescription,
79: 		[]core.OutputText{
80: 			{Mnemonic: dmmMnemonic, Description: dmmDescription},
81: 		},
82: 	)
83: }
84: 
85: // Update updates the Directional Movement Minus given the next bar's high and low values.
86: func (s *DirectionalMovementMinus) Update(high, low float64) float64 {
87: 	if math.IsNaN(high) || math.IsNaN(low) {
88: 		return math.NaN()
89: 	}
90: 
91: 	if high < low {
92: 		high, low = low, high
93: 	}
94: 
95: 	s.mu.Lock()
96: 	defer s.mu.Unlock()
97: 
98: 	if s.noSmoothing {
99: 		if s.primed {
100: 			deltaMinus := s.previousLow - low
101: 			deltaPlus := high - s.previousHigh
102: 
103: 			if deltaMinus > 0 && deltaPlus < deltaMinus {
104: 				s.value = deltaMinus
105: 			} else {
106: 				s.value = 0
107: 			}
108: 		} else {
109: 			if s.count > 0 {
110: 				deltaMinus := s.previousLow - low
111: 				deltaPlus := high - s.previousHigh
112: 
113: 				if deltaMinus > 0 && deltaPlus < deltaMinus {
114: 					s.value = deltaMinus
115: 				} else {
116: 					s.value = 0
117: 				}
118: 
119: 				s.primed = true
120: 			}
121: 
122: 			s.count++
123: 		}
124: 	} else {
125: 		if s.primed {
126: 			deltaMinus := s.previousLow - low
127: 			deltaPlus := high - s.previousHigh
128: 
129: 			if deltaMinus > 0 && deltaPlus < deltaMinus {
130: 				s.accumulator += -s.accumulator/float64(s.length) + deltaMinus
131: 			} else {
132: 				s.accumulator += -s.accumulator / float64(s.length)
133: 			}
134: 
135: 			s.value = s.accumulator
136: 		} else {
137: 			if s.count > 0 && s.length >= s.count {
138: 				deltaMinus := s.previousLow - low
139: 				deltaPlus := high - s.previousHigh
140: 
141: 				if s.length > s.count {
142: 					if deltaMinus > 0 && deltaPlus < deltaMinus {
143: 						s.accumulator += deltaMinus
144: 					}
145: 				} else {
146: 					if deltaMinus > 0 && deltaPlus < deltaMinus {
147: 						s.accumulator += -s.accumulator/float64(s.length) + deltaMinus
148: 					} else {
149: 						s.accumulator += -s.accumulator / float64(s.length)
150: 					}
151: 
152: 					s.value = s.accumulator
153: 					s.primed = true
154: 				}
155: 			}
156: 
157: 			s.count++
158: 		}
159: 	}
160: 
161: 	s.previousLow = low
162: 	s.previousHigh = high
163: 
164: 	return s.value
165: }
166: 
167: // UpdateSample updates the Directional Movement Minus using a single sample value
168: // as a substitute for high and low.
169: func (s *DirectionalMovementMinus) UpdateSample(sample float64) float64 {
170: 	return s.Update(sample, sample)
171: }
172: 
173: // UpdateScalar updates the indicator given the next scalar sample.
174: func (s *DirectionalMovementMinus) UpdateScalar(sample *entities.Scalar) core.Output {
175: 	v := sample.Value
176: 
177: 	output := make([]any, 1)
178: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}
179: 
180: 	return output
181: }
182: 
183: // UpdateBar updates the indicator given the next bar sample.
184: func (s *DirectionalMovementMinus) UpdateBar(sample *entities.Bar) core.Output {
185: 	output := make([]any, 1)
186: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.High, sample.Low)}
187: 
188: 	return output
189: }
190: 
191: // UpdateQuote updates the indicator given the next quote sample.
192: func (s *DirectionalMovementMinus) UpdateQuote(sample *entities.Quote) core.Output {
193: 	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd
194: 
195: 	output := make([]any, 1)
196: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}
197: 
198: 	return output
199: }
200: 
201: // UpdateTrade updates the indicator given the next trade sample.
202: func (s *DirectionalMovementMinus) UpdateTrade(sample *entities.Trade) core.Output {
203: 	v := sample.Price
204: 
205: 	output := make([]any, 1)
206: 	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}
207: 
208: 	return output
209: }

(End of file - total 209 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go</path>
<type>file</type>
<content>1: //nolint:testpackage
2: package directionalmovementminus
3: 
4: //nolint: gofumpt
5: import (
6: 	"math"
7: 	"testing"
8: 	"time"
9: 
10: 	"zpano/entities"
11: 	"zpano/indicators/core"
12: 	"zpano/indicators/core/outputs/shape"
13: )
14: 
15: // TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementMinusTest.cs.
16: func testInputHigh() []float64 {
17: 	return []float64{
18: 		93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
19: 		99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000, 90.345000, 88.500000,
20: 		88.250000, 85.500000, 84.440000, 84.750000, 84.440000, 89.405000, 88.125000, 89.125000, 87.155000, 87.250000, 87.375000,
21: 		88.970000, 90.000000, 89.845000, 86.970000, 85.940000, 84.750000, 85.470000, 84.470000, 88.500000, 89.470000, 90.000000,
22: 		92.440000, 91.440000, 92.970000, 91.720000, 91.155000, 91.750000, 90.000000, 88.875000, 89.000000, 85.250000, 83.815000,
23: 		85.250000, 86.625000, 87.940000, 89.375000, 90.625000, 90.750000, 88.845000, 91.970000, 93.375000, 93.815000, 94.030000,
24: 		94.030000, 91.815000, 92.000000, 91.940000, 89.750000, 88.750000, 86.155000, 84.875000, 85.940000, 99.375000, 103.280000,
25: 		105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000,
26: 		110.940000, 110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000,
27: 		115.345000, 113.000000, 118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000,
28: 		119.870000, 120.870000, 116.750000, 116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000,
29: 		127.750000, 124.190000, 124.440000, 125.750000, 124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000,
30: 		133.500000, 135.500000, 137.440000, 138.690000, 139.190000, 138.500000, 138.130000, 137.500000, 138.880000, 132.130000,
31: 		129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000, 126.620000, 126.690000, 126.000000, 123.120000,
32: 		121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000, 124.500000, 127.870000, 128.560000,
33: 		129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000, 125.620000, 126.000000,
34: 		128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000, 135.000000,
35: 		132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,
36: 		122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000,
37: 		107.250000, 108.870000, 109.000000, 108.500000, 113.060000, 93.000000, 94.620000, 95.120000, 96.000000, 95.560000,
38: 		95.310000, 99.000000, 98.810000, 96.810000, 95.940000, 94.440000, 92.940000, 93.940000, 95.500000, 97.060000, 97.500000,
39: 		96.250000, 96.370000, 95.000000, 94.870000, 98.250000, 105.120000, 108.440000, 109.870000, 105.000000, 106.000000,
40: 		104.940000, 104.500000, 104.440000, 106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000,
41: 		112.750000, 110.190000, 107.940000, 109.690000, 111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000,
42: 		110.750000, 110.500000, 110.500000, 109.500000,
43: 	}
44: }
45: 
46: func testInputLow() []float64 {
47: 	return []float64{
48: 		90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,
49: 		96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000, 87.500000, 86.530000,
50: 		84.625000, 82.280000, 81.565000, 80.875000, 81.250000, 84.065000, 85.595000, 85.970000, 84.405000, 85.095000, 85.500000,
51: 		85.530000, 87.875000, 86.565000, 84.655000, 83.250000, 82.565000, 83.440000, 82.530000, 85.065000, 86.875000, 88.530000,
52: 		89.280000, 90.125000, 90.750000, 89.000000, 88.565000, 90.095000, 89.000000, 86.470000, 84.000000, 83.315000, 82.000000,
53: 		83.250000, 84.750000, 85.280000, 87.190000, 88.440000, 88.250000, 87.345000, 89.280000, 91.095000, 89.530000, 91.155000,
54: 		92.000000, 90.530000, 89.970000, 88.815000, 86.750000, 85.065000, 82.030000, 81.500000, 82.565000, 96.345000, 96.470000,
55: 		101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000,
56: 		108.530000, 108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000,
57: 		110.875000, 110.500000, 110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000,
58: 		116.060000, 116.560000, 113.310000, 112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000,
59: 		123.000000, 121.750000, 121.560000, 123.120000, 122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000,
60: 		130.630000, 132.130000, 133.880000, 135.380000, 135.750000, 136.190000, 134.500000, 135.380000, 133.690000, 126.060000,
61: 		126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000, 124.620000, 124.370000, 121.810000, 118.190000,
62: 		118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000, 122.000000, 124.500000, 126.560000,
63: 		123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000, 122.690000, 122.870000,
64: 		125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000, 131.940000,
65: 		129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,
66: 		119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000,
67: 		104.690000, 103.870000, 104.690000, 105.440000, 107.000000, 89.000000, 92.500000, 92.120000, 94.620000, 92.810000,
68: 		94.250000, 96.250000, 96.370000, 93.690000, 93.500000, 90.000000, 90.190000, 90.500000, 92.120000, 94.120000, 94.870000,
69: 		93.000000, 93.870000, 93.000000, 92.620000, 93.560000, 98.370000, 104.440000, 106.000000, 101.810000, 104.120000,
70: 		103.370000, 102.120000, 102.250000, 103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000,
71: 		106.560000, 106.870000, 104.500000, 105.750000, 108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000,
72: 		109.060000, 108.750000, 108.560000, 106.620000,
73: 	}
74: }
75: 
76: // Expected -DM1 (length=1), 252 entries. Extracted from DirectionalMovementMinusTest.cs talibDmMinus1.
77: func testExpectedDMM1() []float64 {
78: 	return []float64{
79: 		math.NaN(),
80: 		0.0000, 0.0000, 0.7500, 0.6850, 0.0000, 1.5000, 2.2500, 0.3100, 0.0000, 0.0000,
81: 		0.0000, 0.2850, 7.2150, 2.0650, 0.0000, 2.0350, 0.1250, 0.0000, 0.0000, 2.2500,
82: 		0.9700, 1.9050, 2.3450, 0.7150, 0.6900, 0.0000, 0.0000, 0.0000, 0.0000, 1.5650,
83: 		0.0000, 0.0000, 0.0000, 0.0000, 1.3100, 1.9100, 1.4050, 0.6850, 0.0000, 0.9100,
84: 		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.7500, 0.4350, 0.0000, 1.0950,
85: 		2.5300, 2.4700, 0.6850, 1.3150, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1900,
86: 		0.9050, 0.0000, 0.0000, 1.5650, 0.0000, 0.0000, 1.4700, 0.5600, 1.1550, 2.0650,
87: 		1.6850, 3.0350, 0.5300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 2.5000, 0.0300,
88: 		0.0000, 0.0000, 0.0000, 2.0350, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
89: 		0.0000, 2.0000, 0.0000, 1.9700, 0.2800, 1.6550, 3.7200, 0.3750, 0.0000, 0.0000,
90: 		0.0000, 3.0000, 1.7500, 0.0000, 0.0000, 0.0000, 1.4400, 0.0000, 3.2500, 0.7500,
91: 		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,
92: 		0.9300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1800, 0.0000, 0.0000, 0.0000,
93: 		0.0000, 0.0000, 1.6900, 0.0000, 1.6900, 7.6300, 0.0000, 3.3700, 0.8800, 0.0000,
94: 		0.0000, 0.0000, 1.1900, 0.2500, 2.5600, 3.6200, 0.1300, 0.0000, 0.0000, 0.0000,
95: 		2.1800, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.0600, 2.2500, 0.1900, 0.0000,
96: 		1.3100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,
97: 		0.0000, 0.0000, 0.0000, 2.3800, 0.0600, 0.0000, 2.3800, 5.8100, 0.0000, 0.0000,
98: 		1.8800, 2.9300, 1.0000, 0.0000, 0.0000, 1.6900, 0.3800, 4.1200, 0.0000, 0.0000,
99: 		0.0000, 4.0000, 4.0000, 0.0000, 3.1200, 5.0000, 0.3100, 0.0000, 0.0000, 0.0000,
100: 		0.0000, 18.0000, 0.0000, 0.0000, 0.0000, 1.8100, 0.0000, 0.0000, 0.0000, 2.6800,
101: 		0.1900, 3.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.8700, 0.0000, 0.8700,
102: 		0.3800, 0.0000, 0.0000, 0.0000, 0.0000, 4.1900, 0.0000, 0.7500, 1.2500, 0.0000,
103: 		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.2500, 4.6900, 1.0000, 0.0000, 2.3700,
104: 		0.0000, 0.0000, 0.8700, 0.0000, 0.0000, 0.0000, 0.0700, 0.0000, 0.3100, 0.1900,
105: 		1.9400,
106: 	}
107: }
108: 
109: // Expected -DM14 (length=14), 252 entries. Extracted from DirectionalMovementMinusTest.cs talibDmMinus14.
110: func testExpectedDMM14() []float64 {
111: 	return []float64{
112: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
113: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
114: 		14.131785714285700, 13.122372448979600, 14.220060131195300, 13.329341550395700, 12.377245725367400,
115: 		11.493156744984000, 12.922216977485200, 12.969201479093400, 13.947829944872400, 15.296556377381500, 14.918945207568600,
116: 		14.543306264170800, 13.504498673872900, 12.539891625739100, 11.644185081043500, 10.812457575254600, 11.605139177022200,
117: 		10.776200664377700, 10.006472045493600, 9.291724042244060, 8.628029467798060, 9.321741648669630, 10.565902959478900,
118: 		11.216195605230400, 11.100038776285400, 10.307178863693600, 10.480951802001200, 9.732312387572540, 9.037147217031640,
119: 		8.391636701529380, 7.792234079991570, 7.235645931420740, 6.718814079176400, 7.988898787806660, 7.853263160106190,
120: 		7.292315791527170, 7.866436092132370, 9.834547799837210, 11.602080099848800, 11.458360092716800, 11.954905800379900,
121: 		11.100983957495600, 10.308056531960200, 9.571766779677320, 8.888069152557510, 8.253207070231970, 7.853692279501110,
122: 		8.197714259536750, 7.612163240998410, 7.068437295212810, 8.128548916983320, 7.547938280055940, 7.008799831480520,
123: 		7.978171272089050, 7.968301895511260, 8.554137474403320, 10.008127654803100, 10.978261393745700, 13.229099865621000,
124: 		12.814164160933800, 11.898866720867100, 11.048947669376600, 10.259737121564000, 9.526898755737980, 8.846405987470980,
125: 		10.714519845508800, 9.979196999401000, 9.266397213729500, 8.604511698463110, 7.989903720001460, 9.454196311429920,
126: 		8.778896574899210, 8.151832533834980, 7.569558781418200, 7.028876011316900, 6.526813439079980, 6.060612479145690,
127: 		5.627711587778140, 7.225732188651130, 6.709608460890340, 8.200350713683890, 7.894611376992180, 8.985710564349880,
128: 		12.063874095467700, 11.577168802934300, 10.750228174153300, 9.982354733142360, 9.269329395060770, 11.607234438270700,
129: 		12.528146264108500, 11.633278673815100, 10.802330197114000, 10.030735183034400, 10.754254098531900, 9.986093091493950,
130: 		12.522800727815800, 12.378314961543300, 11.494149607147300, 10.673138920922500, 9.910771855142320, 9.202859579775010,
131: 		8.545512466933940, 7.935118719295800, 7.368324525060390, 8.092015630413220, 7.514014513955130, 6.977299191529760,
132: 		7.408920677849070, 6.879712058002710, 6.388304053859660, 5.931996621441110, 5.508282577052460, 5.114833821548720,
133: 		4.929488548580960, 4.577382223682320, 4.250426350562150, 3.946824468379140, 3.664908434923490, 3.403129261000380,
134: 		4.850048599500350, 4.503616556678900, 5.871929659773260, 13.082506112646600, 12.148041390314700, 14.650324148149400,
135: 		14.483872423281500, 13.449310107332900, 12.488645099666200, 11.596599021118600, 11.958270519610200, 11.354108339638000,
136: 		13.103100601092400, 15.787164843871600, 14.789510212166400, 13.733116625583100, 12.752179723755800, 11.841309743487500,
137: 		13.175501904667000, 12.234394625762200, 11.360509295350600, 10.549044345682700, 9.795541178133930, 9.095859665410080,
138: 		11.506155403595100, 12.934287160481100, 12.200409506161100, 11.328951684292400, 11.829740849700100, 10.984759360435800,
139: 		10.200133691833200, 9.471552713845160, 8.795013234284790, 8.166798003264450, 7.583455288745560, 8.291779910978020,
140: 		7.699509917336730, 7.149544923241250, 6.638863143009730, 6.164658632794750, 5.724325873309410, 7.695445453787310,
141: 		7.205770778516790, 6.691072865765590, 8.593139089639470, 13.789343440379500, 12.804390337495300, 11.889791027674200,
142: 		12.920520239983200, 14.927625937127200, 14.861366941618100, 13.799840731502600, 12.814137822109500, 13.588842263387400,
143: 		12.998210673145400, 16.189767053635100, 15.033355121232600, 13.959544041144500, 12.962433752491300, 16.036545627313400,
144: 		18.891078082505300, 17.541715362326300, 19.408735693588700, 23.022397429761000, 21.687940470492300, 20.138801865457200,
145: 		18.700316017924500, 17.364579159501300, 16.124252076679800, 32.972519785488400, 30.617339800810700, 28.430386957895600,
146: 		26.399645032331600, 26.323956101450800, 24.443673522775800, 22.697696842577500, 21.076432782393400, 22.250973297936700,
147: 		20.851618062369800, 22.862216772200500, 21.229201288471900, 19.712829767866800, 18.304770498733400, 16.997286891681000,
148: 		15.783194970846700, 16.525823901500500, 15.345407908536200, 15.119307343640700, 14.419356819095000, 13.389402760588200,
149: 		12.433016849117600, 11.544944217037800, 10.720305344392200, 14.144569248364200, 13.134242873481000, 12.946082668232400,
150: 		13.271362477644400, 12.323408014955500, 11.443164585315800, 10.625795686364700, 9.866810280195770, 9.162038117324650,
151: 		8.507606823230030, 11.149920621570700, 15.043497720030000, 14.968962168599300, 13.899750585127900, 15.276911257618800,
152: 		14.185703310646000, 13.172438788457000, 13.101550303567200, 12.165725281883800, 11.296744904606400, 10.489834554277400,
153: 		9.810560657543280, 9.109806324861620, 8.769105873085790, 8.332741167865380, 9.677545370160700,
154: 	}
155: }
156: 
157: func testTime() time.Time {
158: 	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
159: }
160: 
161: func TestDirectionalMovementMinusConstructor(t *testing.T) {
162: 	t.Parallel()
163: 
164: 	dmm, err := NewDirectionalMovementMinus(14)
165: 	if err != nil {
166: 		t.Fatalf("unexpected error: %v", err)
167: 	}
168: 
169: 	if dmm.Length() != 14 {
170: 		t.Errorf("expected length 14, got %d", dmm.Length())
171: 	}
172: 
173: 	if dmm.IsPrimed() {
174: 		t.Error("should not be primed initially")
175: 	}
176: 
177: 	_, err = NewDirectionalMovementMinus(0)
178: 	if err == nil {
179: 		t.Error("expected error for length 0")
180: 	}
181: 
182: 	_, err = NewDirectionalMovementMinus(-8)
183: 	if err == nil {
184: 		t.Error("expected error for negative length")
185: 	}
186: }
187: 
188: func TestDirectionalMovementMinusIsPrimed(t *testing.T) {
189: 	t.Parallel()
190: 
191: 	high := testInputHigh()
192: 	low := testInputLow()
193: 
194: 	t.Run("length=1", func(t *testing.T) {
195: 		t.Parallel()
196: 
197: 		dmm, _ := NewDirectionalMovementMinus(1)
198: 
199: 		if dmm.IsPrimed() {
200: 			t.Error("should not be primed before updates")
201: 		}
202: 
203: 		// First update: not primed yet.
204: 		dmm.Update(high[0], low[0])
205: 		if dmm.IsPrimed() {
206: 			t.Error("[0] should not be primed yet")
207: 		}
208: 
209: 		// Second update: should be primed.
210: 		dmm.Update(high[1], low[1])
211: 		if !dmm.IsPrimed() {
212: 			t.Error("[1] should be primed")
213: 		}
214: 	})
215: 
216: 	t.Run("length=14", func(t *testing.T) {
217: 		t.Parallel()
218: 
219: 		dmm, _ := NewDirectionalMovementMinus(14)
220: 
221: 		for i := 0; i < 14; i++ {
222: 			dmm.Update(high[i], low[i])
223: 
224: 			if dmm.IsPrimed() {
225: 				t.Errorf("[%d] should not be primed yet", i)
226: 			}
227: 		}
228: 
229: 		dmm.Update(high[14], low[14])
230: 		if !dmm.IsPrimed() {
231: 			t.Error("[14] should be primed")
232: 		}
233: 	})
234: }
235: 
236: func TestDirectionalMovementMinusUpdate(t *testing.T) {
237: 	t.Parallel()
238: 
239: 	const tolerance = 1e-8
240: 
241: 	high := testInputHigh()
242: 	low := testInputLow()
243: 	expected := testExpectedDMM14()
244: 	dmm, _ := NewDirectionalMovementMinus(14)
245: 
246: 	for i := range high {
247: 		act := dmm.Update(high[i], low[i])
248: 
249: 		if math.IsNaN(expected[i]) {
250: 			if !math.IsNaN(act) {
251: 				t.Errorf("[%d] expected NaN, got %v", i, act)
252: 			}
253: 
254: 			continue
255: 		}
256: 
257: 		if math.IsNaN(act) {
258: 			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
259: 			continue
260: 		}
261: 
262: 		if math.Abs(act-expected[i]) > tolerance {
263: 			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
264: 		}
265: 	}
266: }
267: 
268: func TestDirectionalMovementMinusLength1(t *testing.T) {
269: 	t.Parallel()
270: 
271: 	const tolerance = 1e-8
272: 
273: 	high := testInputHigh()
274: 	low := testInputLow()
275: 	expected := testExpectedDMM1()
276: 	dmm, _ := NewDirectionalMovementMinus(1)
277: 
278: 	for i := range high {
279: 		act := dmm.Update(high[i], low[i])
280: 
281: 		if math.IsNaN(expected[i]) {
282: 			if !math.IsNaN(act) {
283: 				t.Errorf("[%d] expected NaN, got %v", i, act)
284: 			}
285: 
286: 			continue
287: 		}
288: 
289: 		if math.IsNaN(act) {
290: 			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
291: 			continue
292: 		}
293: 
294: 		if math.Abs(act-expected[i]) > tolerance {
295: 			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
296: 		}
297: 	}
298: }
299: 
300: func TestDirectionalMovementMinusNaNPassthrough(t *testing.T) {
301: 	t.Parallel()
302: 
303: 	dmm, _ := NewDirectionalMovementMinus(14)
304: 
305: 	if !math.IsNaN(dmm.Update(math.NaN(), 1)) {
306: 		t.Error("expected NaN passthrough for NaN high")
307: 	}
308: 
309: 	if !math.IsNaN(dmm.Update(1, math.NaN())) {
310: 		t.Error("expected NaN passthrough for NaN low")
311: 	}
312: 
313: 	if !math.IsNaN(dmm.Update(math.NaN(), math.NaN())) {
314: 		t.Error("expected NaN passthrough for NaN high and low")
315: 	}
316: 
317: 	if !math.IsNaN(dmm.UpdateSample(math.NaN())) {
318: 		t.Error("expected NaN passthrough for NaN sample")
319: 	}
320: }
321: 
322: func TestDirectionalMovementMinusHighLowSwap(t *testing.T) {
323: 	t.Parallel()
324: 
325: 	// When high < low, they should be swapped internally.
326: 	dmm1, _ := NewDirectionalMovementMinus(1)
327: 	dmm2, _ := NewDirectionalMovementMinus(1)
328: 
329: 	// Prime both.
330: 	dmm1.Update(10, 5)
331: 	dmm2.Update(5, 10) // Swapped.
332: 
333: 	// Update both with same effective values.
334: 	v1 := dmm1.Update(12, 6)
335: 	v2 := dmm2.Update(6, 12) // Swapped.
336: 
337: 	if v1 != v2 {
338: 		t.Errorf("high/low swap should produce same result: %v vs %v", v1, v2)
339: 	}
340: }
341: 
342: func TestDirectionalMovementMinusZeroInputs(t *testing.T) {
343: 	t.Parallel()
344: 
345: 	// Updating with same value repeatedly should produce 0.
346: 	dmm, _ := NewDirectionalMovementMinus(10)
347: 
348: 	for i := 0; i < 20; i++ {
349: 		dmm.UpdateSample(0)
350: 	}
351: 
352: 	if !dmm.IsPrimed() {
353: 		t.Error("should be primed after 20 updates with length 10")
354: 	}
355: }
356: 
357: func TestDirectionalMovementMinusUpdateEntity(t *testing.T) {
358: 	t.Parallel()
359: 
360: 	tm := testTime()
361: 	high := testInputHigh()
362: 	low := testInputLow()
363: 
364: 	check := func(t *testing.T, act core.Output) {
365: 		t.Helper()
366: 
367: 		if len(act) != 1 {
368: 			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
369: 		}
370: 
371: 		s, ok := act[0].(entities.Scalar)
372: 		if !ok {
373: 			t.Error("output is not scalar")
374: 		}
375: 
376: 		if s.Time != tm {
377: 			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
378: 		}
379: 	}
380: 
381: 	t.Run("update bar", func(t *testing.T) {
382: 		t.Parallel()
383: 
384: 		dmm, _ := NewDirectionalMovementMinus(14)
385: 		for i := 0; i < 14; i++ {
386: 			dmm.Update(high[i], low[i])
387: 		}
388: 
389: 		b := entities.Bar{Time: tm, High: high[14], Low: low[14]}
390: 		check(t, dmm.UpdateBar(&b))
391: 	})
392: 
393: 	t.Run("update scalar", func(t *testing.T) {
394: 		t.Parallel()
395: 
396: 		dmm, _ := NewDirectionalMovementMinus(14)
397: 		for i := 0; i < 14; i++ {
398: 			dmm.Update(high[i], low[i])
399: 		}
400: 
401: 		s := entities.Scalar{Time: tm, Value: high[14]}
402: 		check(t, dmm.UpdateScalar(&s))
403: 	})
404: 
405: 	t.Run("update quote", func(t *testing.T) {
406: 		t.Parallel()
407: 
408: 		dmm, _ := NewDirectionalMovementMinus(14)
409: 		for i := 0; i < 14; i++ {
410: 			dmm.Update(high[i], low[i])
411: 		}
412: 
413: 		q := entities.Quote{Time: tm, Bid: high[14] - 0.5, Ask: high[14] + 0.5}
414: 		check(t, dmm.UpdateQuote(&q))
415: 	})
416: 
417: 	t.Run("update trade", func(t *testing.T) {
418: 		t.Parallel()
419: 
420: 		dmm, _ := NewDirectionalMovementMinus(14)
421: 		for i := 0; i < 14; i++ {
422: 			dmm.Update(high[i], low[i])
423: 		}
424: 
425: 		r := entities.Trade{Time: tm, Price: high[14]}
426: 		check(t, dmm.UpdateTrade(&r))
427: 	})
428: }
429: 
430: func TestDirectionalMovementMinusMetadata(t *testing.T) {
431: 	t.Parallel()
432: 
433: 	dmm, _ := NewDirectionalMovementMinus(14)
434: 	act := dmm.Metadata()
435: 
436: 	check := func(what string, exp, act any) {
437: 		t.Helper()
438: 
439: 		if exp != act {
440: 			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
441: 		}
442: 	}
443: 
444: 	check("Identifier", core.DirectionalMovementMinus, act.Identifier)
445: 	check("Mnemonic", "-dm", act.Mnemonic)
446: 	check("Description", "Directional Movement Minus", act.Description)
447: 	check("len(Outputs)", 1, len(act.Outputs))
448: 	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
449: 	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
450: 	check("Outputs[0].Mnemonic", "-dm", act.Outputs[0].Mnemonic)
451: 	check("Outputs[0].Description", "Directional Movement Minus", act.Outputs[0].Description)
452: }

(End of file - total 452 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: 
9: const indicator_mod = @import("../../core/indicator.zig");
10: const build_metadata_mod = @import("../../core/build_metadata.zig");
11: const identifier_mod = @import("../../core/identifier.zig");
12: const metadata_mod = @import("../../core/metadata.zig");
13: 
14: const OutputArray = indicator_mod.OutputArray;
15: const Identifier = identifier_mod.Identifier;
16: const Metadata = metadata_mod.Metadata;
17: 
18: /// Enumerates the outputs of the Williams %R indicator.
19: pub const WilliamsPercentROutput = enum(u8) {
20:     /// The scalar value of the Williams %R.
21:     value = 1,
22: };
23: 
24: /// Parameters to create an instance of the Williams %R indicator.
25: pub const WilliamsPercentRParams = struct {
26:     /// The number of time periods. Must be >= 2. Default is 14.
27:     length: usize = 14,
28: };
29: 
30: /// Larry Williams' Williams %R momentum indicator.
31: ///
32: /// Williams %R reflects the level of the closing price relative to the
33: /// highest high over a lookback period. The oscillation ranges from 0 to -100;
34: /// readings from 0 to -20 are considered overbought, readings from -80 to -100
35: /// are considered oversold.
36: ///
37: /// The value is calculated as:
38: ///
39: ///   %R = -100 * (HighestHigh - Close) / (HighestHigh - LowestLow)
40: ///
41: /// where HighestHigh and LowestLow are computed over the last `length` bars.
42: /// If HighestHigh equals LowestLow, the value is 0.
43: ///
44: /// The indicator requires bar data (high, low, close). For scalar, quote, and
45: /// trade updates, the single value is used as a substitute for all three.
46: pub const WilliamsPercentR = struct {
47:     length: usize,
48:     high_buf: []f64,
49:     low_buf: []f64,
50:     buffer_index: usize,
51:     count: usize,
52:     value: f64,
53:     primed: bool,
54:     allocator: std.mem.Allocator,
55: 
56:     pub const Error = error{
57:         InvalidLength,
58:         OutOfMemory,
59:     };
60: 
61:     pub fn init(allocator: std.mem.Allocator, params: WilliamsPercentRParams) Error!WilliamsPercentR {
62:         const length = params.length;
63:         if (length < 2) return error.InvalidLength;
64: 
65:         const high_buf = allocator.alloc(f64, length) catch return error.OutOfMemory;
66:         errdefer allocator.free(high_buf);
67:         const low_buf = allocator.alloc(f64, length) catch return error.OutOfMemory;
68: 
69:         return WilliamsPercentR{
70:             .length = length,
71:             .high_buf = high_buf,
72:             .low_buf = low_buf,
73:             .buffer_index = 0,
74:             .count = 0,
75:             .value = math.nan(f64),
76:             .primed = false,
77:             .allocator = allocator,
78:         };
79:     }
80: 
81:     pub fn deinit(self: *WilliamsPercentR) void {
82:         self.allocator.free(self.high_buf);
83:         self.allocator.free(self.low_buf);
84:     }
85: 
86:     pub fn fixSlices(self: *WilliamsPercentR) void {
87:         _ = self;
88:     }
89: 
90:     /// Core update given close, high, low values.
91:     pub fn update(self: *WilliamsPercentR, close: f64, high: f64, low: f64) f64 {
92:         if (math.isNan(close) or math.isNan(high) or math.isNan(low)) {
93:             return math.nan(f64);
94:         }
95: 
96:         self.high_buf[self.buffer_index] = high;
97:         self.low_buf[self.buffer_index] = low;
98: 
99:         self.buffer_index = (self.buffer_index + 1) % self.length;
100:         self.count += 1;
101: 
102:         if (self.count < self.length) {
103:             return self.value;
104:         }
105: 
106:         // Find highest high and lowest low in the window.
107:         var max_high = self.high_buf[0];
108:         var min_low = self.low_buf[0];
109:         for (self.high_buf[1..], self.low_buf[1..]) |h, l| {
110:             if (h > max_high) max_high = h;
111:             if (l < min_low) min_low = l;
112:         }
113: 
114:         const diff = max_high - min_low;
115:         if (@abs(diff) < math.floatMin(f64)) {
116:             self.value = 0;
117:         } else {
118:             self.value = -100.0 * (max_high - close) / diff;
119:         }
120: 
121:         if (!self.primed) {
122:             self.primed = true;
123:         }
124: 
125:         return self.value;
126:     }
127: 
128:     /// Update using a single sample value as substitute for high, low, close.
129:     pub fn updateSample(self: *WilliamsPercentR, sample: f64) f64 {
130:         return self.update(sample, sample, sample);
131:     }
132: 
133:     pub fn isPrimed(self: *const WilliamsPercentR) bool {
134:         return self.primed;
135:     }
136: 
137:     pub fn getMetadata(self: *const WilliamsPercentR, out: *Metadata) void {
138:         _ = self;
139:         const mnemonic = "willr";
140:         const description = "Williams %R";
141: 
142:         build_metadata_mod.buildMetadata(out, Identifier.williams_percent_r, mnemonic, description, &.{
143:             .{ .mnemonic = mnemonic, .description = description },
144:         });
145:     }
146: 
147:     fn makeOutput(self: *const WilliamsPercentR, time: i64) OutputArray {
148:         var out = OutputArray{};
149:         out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
150:         return out;
151:     }
152: 
153:     pub fn updateScalar(self: *WilliamsPercentR, sample: *const Scalar) OutputArray {
154:         _ = self.update(sample.value, sample.value, sample.value);
155:         return self.makeOutput(sample.time);
156:     }
157: 
158:     pub fn updateBar(self: *WilliamsPercentR, sample: *const Bar) OutputArray {
159:         _ = self.update(sample.close, sample.high, sample.low);
160:         return self.makeOutput(sample.time);
161:     }
162: 
163:     pub fn updateQuote(self: *WilliamsPercentR, sample: *const Quote) OutputArray {
164:         const mid = (sample.bid_price + sample.ask_price) / 2.0;
165:         _ = self.update(mid, mid, mid);
166:         return self.makeOutput(sample.time);
167:     }
168: 
169:     pub fn updateTrade(self: *WilliamsPercentR, sample: *const Trade) OutputArray {
170:         _ = self.update(sample.price, sample.price, sample.price);
171:         return self.makeOutput(sample.time);
172:     }
173: 
174:     // --- Indicator interface ---
175: 
176:     pub fn indicator(self: *WilliamsPercentR) indicator_mod.Indicator {
177:         return indicator_mod.Indicator{
178:             .ptr = @ptrCast(self),
179:             .vtable = &.{
180:                 .isPrimed = vtableIsPrimed,
181:                 .metadata = vtableMetadata,
182:                 .updateScalar = vtableUpdateScalar,
183:                 .updateBar = vtableUpdateBar,
184:                 .updateQuote = vtableUpdateQuote,
185:                 .updateTrade = vtableUpdateTrade,
186:             },
187:         };
188:     }
189: 
190:     fn vtableIsPrimed(ptr: *const anyopaque) bool {
191:         const self: *const WilliamsPercentR = @ptrCast(@alignCast(ptr));
192:         return self.isPrimed();
193:     }
194: 
195:     fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
196:         const self: *const WilliamsPercentR = @ptrCast(@alignCast(ptr));
197:         self.getMetadata(out);
198:     }
199: 
200:     fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
201:         const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
202:         return self.updateScalar(sample);
203:     }
204: 
205:     fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
206:         const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
207:         return self.updateBar(sample);
208:     }
209: 
210:     fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
211:         const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
212:         return self.updateQuote(sample);
213:     }
214: 
215:     fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
216:         const self: *WilliamsPercentR = @ptrCast(@alignCast(ptr));
217:         return self.updateTrade(sample);
218:     }
219: };
220: 
221: // ============================================================================
222: // Tests
223: // ============================================================================
224: 
225: const testing = std.testing;
226: 
227: fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
228:     if (math.isNan(a) and math.isNan(b)) return true;
229:     if (math.isNan(a) or math.isNan(b)) return false;
230:     return @abs(a - b) <= tolerance;
231: }
232: 
233: const test_high = [_]f64{
234:     93.25,  94.94,  96.375,  96.19,   96.0,    94.72,  95.0,   93.72,   92.47,   92.75,
235:     96.25,  99.625, 99.125,  92.75,   91.315,  93.25,  93.405, 90.655,  91.97,   92.25,
236:     90.345, 88.5,   88.25,   85.5,    84.44,   84.75,  84.44,  89.405,  88.125,  89.125,
237:     87.155, 87.25,  87.375,  88.97,   90.0,    89.845, 86.97,  85.94,   84.75,   85.47,
238:     84.47,  88.5,   89.47,   90.0,    92.44,   91.44,  92.97,  91.72,   91.155,  91.75,
239:     90.0,   88.875, 89.0,    85.25,   83.815,  85.25,  86.625, 87.94,   89.375,  90.625,
240:     90.75,  88.845, 91.97,   93.375,  93.815,  94.03,  94.03,  91.815,  92.0,    91.94,
241:     89.75,  88.75,  86.155,  84.875,  85.94,   99.375, 103.28, 105.375, 107.625, 105.25,
242:     104.5,  105.5,  106.125, 107.94,  106.25,  107.0,  108.75, 110.94,  110.94,  114.22,
243:     123.0,  121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0,   118.315,
244:     116.87, 116.75, 113.87,  114.62,  115.31,  116.0,  121.69, 119.87,  120.87,  116.75,
245:     116.5,  116.0,  118.31,  121.5,   122.0,   121.44, 125.75, 127.75,  124.19,  124.44,
246:     125.75, 124.69, 125.31,  132.0,   131.31,  132.25, 133.88, 133.5,   135.5,   137.44,
247:     138.69, 139.19, 138.5,   138.13,  137.5,   138.88, 132.13, 129.75,  128.5,   125.44,
248:     125.12, 126.5,  128.69,  126.62,  126.69,  126.0,  123.12, 121.87,  124.0,   127.0,
249:     124.44, 122.5,  123.75,  123.81,  124.5,   127.87, 128.56, 129.63,  124.87,  124.37,
250:     124.87, 123.62, 124.06,  125.87,  125.19,  125.62, 126.0,  128.5,   126.75,  129.75,
251:     132.69, 133.94, 136.5,   137.69,  135.56,  133.56, 135.0,  132.38,  131.44,  130.88,
252:     129.63, 127.25, 127.81,  125.0,   126.81,  124.75, 122.81, 122.25,  121.06,  120.0,
253:     123.25, 122.75, 119.19,  115.06,  116.69,  114.87, 110.87, 107.25,  108.87,  109.0,
254:     108.5,  113.06, 93.0,    94.62,   95.12,   96.0,   95.56,  95.31,   99.0,    98.81,
255:     96.81,  95.94,  94.44,   92.94,   93.94,   95.5,   97.06,  97.5,    96.25,   96.37,
256:     95.0,   94.87,  98.25,   105.12,  108.44,  109.87, 105.0,  106.0,   104.94,  104.5,
257:     104.44, 106.31, 112.87,  116.5,   119.19,  121.0,  122.12, 111.94,  112.75,  110.19,
258:     107.94, 109.69, 111.06,  110.44,  110.12,  110.31, 110.44, 110.0,   110.75,  110.5,
259:     110.5,  109.5,
260: };
261: 
262: const test_low = [_]f64{
263:     90.75,  91.405, 94.25,   93.5,   92.815,  93.5,   92.0,    89.75,   89.44,  90.625,
264:     92.75,  96.315, 96.03,   88.815, 86.75,   90.94,  88.905,  88.78,   89.25,  89.75,
265:     87.5,   86.53,  84.625,  82.28,  81.565,  80.875, 81.25,   84.065,  85.595, 85.97,
266:     84.405, 85.095, 85.5,    85.53,  87.875,  86.565, 84.655,  83.25,   82.565, 83.44,
267:     82.53,  85.065, 86.875,  88.53,  89.28,   90.125, 90.75,   89.0,    88.565, 90.095,
268:     89.0,   86.47,  84.0,    83.315, 82.0,    83.25,  84.75,   85.28,   87.19,  88.44,
269:     88.25,  87.345, 89.28,   91.095, 89.53,   91.155, 92.0,    90.53,   89.97,  88.815,
270:     86.75,  85.065, 82.03,   81.5,   82.565,  96.345, 96.47,   101.155, 104.25, 101.75,
271:     101.72, 101.72, 103.155, 105.69, 103.655, 104.0,  105.53,  108.53,  108.75, 107.75,
272:     117.0,  118.0,  116.0,   118.5,  116.53,  116.25, 114.595, 110.875, 110.5,  110.72,
273:     112.62, 114.19, 111.19,  109.44, 111.56,  112.44, 117.5,   116.06,  116.56, 113.31,
274:     112.56, 114.0,  114.75,  118.87, 119.0,   119.75, 122.62,  123.0,   121.75, 121.56,
275:     123.12, 122.19, 122.75,  124.37, 128.0,   129.5,  130.81,  130.63,  132.13, 133.88,
276:     135.38, 135.75, 136.19,  134.5,  135.38,  133.69, 126.06,  126.87,  123.5,  122.62,
277:     122.75, 123.56, 125.81,  124.62, 124.37,  121.81, 118.19,  118.06,  117.56, 121.0,
278:     121.12, 118.94, 119.81,  121.0,  122.0,   124.5,  126.56,  123.5,   121.25, 121.06,
279:     122.31, 121.0,  120.87,  122.06, 122.75,  122.69, 122.87,  125.5,   124.25, 128.0,
280:     128.38, 130.69, 131.63,  134.38, 132.0,   131.94, 131.94,  129.56,  123.75, 126.0,
281:     126.25, 124.37, 121.44,  120.44, 121.37,  121.69, 120.0,   119.62,  115.5,  116.75,
282:     119.06, 119.06, 115.06,  111.06, 113.12,  110.0,  105.0,   104.69,  103.87, 104.69,
283:     105.44, 107.0,  89.0,    92.5,   92.12,   94.62,  92.81,   94.25,   96.25,  96.37,
284:     93.69,  93.5,   90.0,    90.19,  90.5,    92.12,  94.12,   94.87,   93.0,   93.87,
285:     93.0,   92.62,  93.56,   98.37,  104.44,  106.0,  101.81,  104.12,  103.37, 102.12,
286:     102.25, 103.37, 107.94,  112.5,  115.44,  115.5,  112.25,  107.56,  106.56, 106.87,
287:     104.5,  105.75, 108.62,  107.75, 108.06,  108.0,  108.19,  108.12,  109.06, 108.75,
288:     108.56, 106.62,
289: };
290: 
291: const test_close = [_]f64{
292:     91.5,    94.815,  94.375,  95.095, 93.78,   94.625,  92.53,   92.75,   90.315,  92.47,
293:     96.125,  97.25,   98.5,    89.875, 91.0,    92.815,  89.155,  89.345,  91.625,  89.875,
294:     88.375,  87.625,  84.78,   83.0,   83.5,    81.375,  84.44,   89.25,   86.375,  86.25,
295:     85.25,   87.125,  85.815,  88.97,  88.47,   86.875,  86.815,  84.875,  84.19,   83.875,
296:     83.375,  85.5,    89.19,   89.44,  91.095,  90.75,   91.44,   89.0,    91.0,    90.5,
297:     89.03,   88.815,  84.28,   83.5,   82.69,   84.75,   85.655,  86.19,   88.94,   89.28,
298:     88.625,  88.5,    91.97,   91.5,   93.25,   93.5,    93.155,  91.72,   90.0,    89.69,
299:     88.875,  85.19,   83.375,  84.875, 85.94,   97.25,   99.875,  104.94,  106.0,   102.5,
300:     102.405, 104.595, 106.125, 106.0,  106.065, 104.625, 108.625, 109.315, 110.5,   112.75,
301:     123.0,   119.625, 118.75,  119.25, 117.94,  116.44,  115.19,  111.875, 110.595, 118.125,
302:     116.0,   116.0,   112.0,   113.75, 112.94,  116.0,   120.5,   116.62,  117.0,   115.25,
303:     114.31,  115.5,   115.87,  120.69, 120.19,  120.75,  124.75,  123.37,  122.94,  122.56,
304:     123.12,  122.56,  124.62,  129.25, 131.0,   132.25,  131.0,   132.81,  134.0,   137.38,
305:     137.81,  137.88,  137.25,  136.31, 136.25,  134.63,  128.25,  129.0,   123.87,  124.81,
306:     123.0,   126.25,  128.38,  125.37, 125.69,  122.25,  119.37,  118.5,   123.19,  123.5,
307:     122.19,  119.31,  123.31,  121.12, 123.37,  127.37,  128.5,   123.87,  122.94,  121.75,
308:     124.44,  122.0,   122.37,  122.94, 124.0,   123.19,  124.56,  127.25,  125.87,  128.86,
309:     132.0,   130.75,  134.75,  135.0,  132.38,  133.31,  131.94,  130.0,   125.37,  130.13,
310:     127.12,  125.19,  122.0,   125.0,  123.0,   123.5,   120.06,  121.0,   117.75,  119.87,
311:     122.0,   119.19,  116.37,  113.5,  114.25,  110.0,   105.06,  107.0,   107.87,  107.0,
312:     107.12,  107.0,   91.0,    93.94,  93.87,   95.5,    93.0,    94.94,   98.25,   96.75,
313:     94.81,   94.37,   91.56,   90.25,  93.94,   93.62,   97.0,    95.0,    95.87,   94.06,
314:     94.62,   93.75,   98.0,    103.94, 107.87,  106.06,  104.5,   105.0,   104.19,  103.06,
315:     103.42,  105.27,  111.87,  116.0,  116.62,  118.28,  113.37,  109.0,   109.7,   109.25,
316:     107.0,   109.19,  110.0,   109.2,  110.12,  108.0,   108.62,  109.75,  109.81,  109.0,
317:     108.75,  107.87,
318: };
319: 
320: const test_expected_14 = [_]f64{
321:     math.nan(f64),     math.nan(f64),     math.nan(f64),      math.nan(f64),      math.nan(f64),     math.nan(f64),     math.nan(f64),
322:     math.nan(f64),     math.nan(f64),     math.nan(f64),      math.nan(f64),      math.nan(f64),     math.nan(f64),     -90.1942645698427,
323:     -66.9902912621359, -52.8932038834952, -81.3203883495145,  -79.8446601941748,  -62.1359223300971, -75.7281553398058, -87.378640776699,
324:     -91.6380297823597, -98.9666666666667, -95.8489478235803,  -89.2857142857143,  -97.2602739726027, -71.5482841181165, -33.1604150039904,
325:     -56.1053471667997, -57.1029529130088, -61.5384615384615,  -45.0549450549451,  -56.5714285714286, -14.519535374868,  -16.7671232876712,
326:     -34.2465753424658, -34.9041095890411, -56.1643835616438,  -63.6712328767124,  -70.0,             -88.6880856760375, -60.2409638554217,
327:     -10.8433734939759, -7.49665327978584, -13.5721493440969,  -17.0534813319879,  -14.6551724137931, -38.0268199233716, -18.8697318007663,
328:     -23.6590038314176, -37.7394636015325, -39.7988505747127,  -83.2375478927203,  -90.7088122605364, -93.7101185050137, -74.9316317228806,
329:     -66.6818596171376, -61.8049225159526, -36.7365542388332,  -33.6371923427529,  -32.0512820512821, -33.3333333333333, 0.0,
330:     -16.4835164835165, -4.7820567075751,  -4.40565253532835,  -7.27348295926849,  -19.2019950124688, -37.3840445269017, -46.7672413793104,
331:     -58.9142857142857, -98.6056887897379, -88.7916666666667,  -73.0646448523543,  -64.5650438946528, -11.8881118881119, -15.633608815427,
332:     -1.82198952879582, -6.2200956937799,  -19.6172248803828,  -19.9808612440191,  -11.5980861244019, -5.74162679425837, -7.33736762481089,
333:     -7.09152798789713, -12.5378214826021, -0.458715596330275, -5.72687224669604,  -3.01473107228501, -8.28169014084506, 0.0,
334:     -15.859962406015,  -19.9718045112782, -17.6221804511278,  -23.7781954887218,  -33.0561854371378, -40.3721891961747, -57.5084001033859,
335:     -65.2894736842105, -27.9049799656554, -45.9016393442623,  -45.9016393442623,  -72.1311475409836, -68.2153392330383, -71.5678310316816,
336:     -39.6781609195402, -9.7142857142857,  -41.3877551020408,  -38.2857142857143,  -52.5714285714286, -60.2448979591836, -50.5306122448979,
337:     -47.5102040816326, -8.16326530612245, -14.4108280254777,  -9.95222929936305,  -6.13120784794604, -27.0537368746139, -31.4173742651862,
338:     -34.167215273206,  -30.4805793285056, -34.167215273206,   -20.6056616194865,  -14.1460905349794, -5.55555555555556, 0.0,
339:     -19.1872085276482, -7.19086021505372, -9.52380952380952,  -0.377833753148629, -5.13718622300056, -7.43051616562679, -11.0039705048213,
340:     -16.9411764705882, -17.2941176470588, -27.7372262773723,  -73.8191632928475,  -77.6085300837776, -97.6418100701083, -86.7833433916717,
341:     -97.7066988533495, -78.0929390464695, -65.2383826191914,  -83.4037417018708,  -81.4725407362704, -97.4223784417106, -94.296761720638,
342:     -97.8866474543708, -73.5928705440901, -59.2312971859986,  -62.0180475799836,  -84.2767295597484, -48.3378256963162, -68.0143755615453,
343:     -47.7987421383647, -11.8598382749326, -0.545454545454566, -47.7216238608119,  -55.4266777133389, -65.2858326429163, -42.9991714995858,
344:     -63.2145816072908, -67.9139382600561, -62.5818521983162,  -52.6660430308699,  -65.5804480651731, -57.8767123287671, -27.1689497716895,
345:     -42.9223744292237, -10.0225225225224, -5.83756345177663,  -24.407039020658,   -11.1964171465131, -15.9928656361474, -31.5695600475625,
346:     -26.0404280618311, -36.7882277671145, -51.2666666666667,  -82.1333333333333,  -51.0121457489879, -75.8249641319942, -89.6700143472023,
347:     -96.5538461538461, -73.5652173913043, -85.1594202898551,  -82.2608695652174,  -99.6608253250424, -91.3425345043915, -88.4615384615385,
348:     -77.5897435897436, -61.4928909952607, -76.8506900878294,  -91.7193426042983,  -86.8605277329025, -80.955223880597,  -100.0,
349:     -99.7248968363136, -89.5569620253164, -80.8429118773946,  -83.8493292053664,  -83.2301341589267, -83.8493292053664, -94.1605839416058,
350:     -85.5766423357664, -85.5703703703704, -78.4696919509771,  -85.5543517515349,  -78.5482123510292, -64.2442984151527, -67.7888611803824,
351:     -75.8520365752286, -77.6807980049875, -89.3599334995844,  -94.8046550290939,  -79.4679966749792, -53.8,             -22.2222222222222,
352:     -44.4444444444444, -34.7777777777777, -54.8888888888889,  -48.6666666666666,  -58.3333333333333, -9.19409761634509, -7.80423280423285,
353:     -3.09110629067241, -19.1746351283342, -27.2865853658537,  -25.1419721218379,  -32.0,             -39.4782608695652, -37.3913043478261,
354:     -26.6666666666667, -4.93827160493827, -2.09380234505863,  -9.67256304102369,  -9.91253644314869, -36.8421052631579, -64.5987198424422,
355:     -61.152141802068,  -63.3677991137371, -75.6,              -64.65,             -60.6,             -65.0226472068445, -64.0,
356:     -80.1362088535755, -76.6174801362089, -70.2043132803632,  -69.8637911464245,  -74.4608399545971, -48.4848484848485, -59.1515151515151,
357: };
358: 
359: test "WilliamsPercentR update period=14" {
360:     const tolerance = 1e-6;
361:     const allocator = testing.allocator;
362: 
363:     var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
364:     defer w.deinit();
365: 
366:     for (0..test_close.len) |i| {
367:         const act = w.update(test_close[i], test_high[i], test_low[i]);
368:         const exp = test_expected_14[i];
369: 
370:         if (math.isNan(exp)) {
371:             try testing.expect(math.isNan(act));
372:         } else {
373:             try testing.expect(!math.isNan(act));
374:             try testing.expect(almostEqual(act, exp, tolerance));
375:         }
376:     }
377: }
378: 
379: const test_expected_2 = [_]f64{
380:     math.nan(f64),
381:     -2.98329355608592,
382:     -40.2414486921529,
383:     -44.5217391304348,
384:     -71.4074074074073,
385:     -43.171114599686,
386:     -82.3333333333333,
387:     -42.8571428571429,
388:     -79.5560747663551,
389:     -8.4592145015106,
390:     -2.22222222222222,
391:     -34.5454545454545,
392:     -31.2934631432545,
393:     -89.7187196896217,
394:     -29.1666666666667,
395:     -6.69230769230773,
396:     -94.4444444444444,
397:     -87.7837837837838,
398:     -10.8150470219435,
399:     -79.1666666666667,
400:     -81.5789473684211,
401:     -71.2975098296199,
402:     -96.0,
403:     -87.9396984924623,
404:     -50.8259212198221,
405:     -87.0967741935484,
406:     -8.00000000000006,
407:     -1.90067443286329,
408:     -56.7415730337079,
409:     -81.4447592067988,
410:     -82.0974576271187,
411:     -4.39367311072056,
412:     -68.421052631579,
413:     0.0,
414:     -34.2281879194631,
415:     -90.9752547307132,
416:     -58.3815028901735,
417:     -56.3172043010753,
418:     -51.8518518518518,
419:     -54.9053356282271,
420:     -71.2585034013606,
421:     -50.251256281407,
422:     -6.35641316685587,
423:     -17.9200000000001,
424:     -34.3989769820972,
425:     -53.4810126582278,
426:     -53.7785588752197,
427:     -100.0,
428:     -22.8209191759112,
429:     -39.2464678178964,
430:     -98.9090909090909,
431:     -33.5694050991502,
432:     -94.4,
433:     -96.7458223394898,
434:     -78.7692307692308,
435:     -15.3846153846154,
436:     -28.7407407407407,
437:     -54.858934169279,
438:     -10.6227106227107,
439:     -39.1557496360989,
440:     -85.0,
441:     -66.079295154185,
442:     0.0,
443:     -45.7875457875458,
444:     -13.1855309218203,
445:     -11.7777777777778,
446:     -30.4347826086957,
447:     -66.0000000000001,
448:     -98.5221674876847,
449:     -72.5274725274726,
450:     -59.0558766859345,
451:     -97.3319103521878,
452:     -79.9851190476191,
453:     -27.4973147153598,
454:     0.0,
455:     -12.6412849494349,
456:     -49.098774333093,
457:     -4.88489612577206,
458:     -25.1159196290572,
459:     -87.2340425531915,
460:     -80.5949008498583,
461:     -23.941798941799,
462:     0.0,
463:     -40.5433646812957,
464:     -43.7572928821471,
465:     -71.0014947683109,
466:     -2.63157894736842,
467:     -30.0369685767098,
468:     -18.2572614107883,
469:     -22.7202472952086,
470:     0.0,
471:     -56.25,
472:     -52.1739130434783,
473:     -24.6813441483198,
474:     -62.7476882430648,
475:     -93.9200000000001,
476:     -83.4492350486787,
477:     -82.8030954428203,
478:     -98.0392156862745,
479:     -2.43122200895711,
480:     -30.4805793285056,
481:     -20.4705882352942,
482:     -85.431654676259,
483:     -16.7953667953669,
484:     -40.374787052811,
485:     0.0,
486:     -12.8648648648648,
487:     -90.0532859680284,
488:     -80.4573804573805,
489:     -74.3386243386244,
490:     -58.2338902147971,
491:     -25.3807106598985,
492:     -56.6125290023201,
493:     -12.0,
494:     -57.8274760383388,
495:     -41.6666666666667,
496:     -16.6666666666667,
497:     -85.3801169590643,
498:     -80.1666666666667,
499:     -65.2777777777777,
500:     -62.7684964200477,
501:     -89.6067415730336,
502:     -22.1153846153845,
503:     -29.7297297297297,
504:     -13.1061598951507,
505:     0.0,
506:     -65.7534246575342,
507:     -32.9230769230767,
508:     -30.8008213552361,
509:     -1.1299435028249,
510:     -18.2952182952182,
511:     -34.3832020997376,
512:     -56.3953488372093,
513:     -54.7499999999999,
514:     -51.7906336088154,
515:     -81.8882466281311,
516:     -82.9173166926677,
517:     -51.5650741350906,
518:     -94.0799999999999,
519:     -62.7551020408163,
520:     -86.5248226950356,
521:     -6.66666666666667,
522:     -6.04288499025346,
523:     -81.5724815724815,
524:     -43.1034482758622,
525:     -90.9836065573771,
526:     -84.8911651728552,
527:     -91.304347826087,
528:     -12.5776397515528,
529:     -37.0762711864407,
530:     -80.1666666666667,
531:     -93.2727272727272,
532:     -9.1476091476091,
533:     -67.2499999999999,
534:     -32.2857142857142,
535:     -8.51788756388415,
536:     -1.47783251231533,
537:     -93.9641109298531,
538:     -79.8329355608592,
539:     -81.8897637795276,
540:     -11.2860892388453,
541:     -74.1602067183463,
542:     -52.9780564263323,
543:     -58.6000000000001,
544:     -49.0813648293964,
545:     -82.9351535836178,
546:     -43.5045317220543,
547:     -22.202486678508,
548:     -61.8823529411764,
549:     -16.1818181818179,
550:     -14.7121535181236,
551:     -57.3741007194244,
552:     -30.1204819277108,
553:     -44.3894389438943,
554:     -93.3216168717048,
555:     -62.1546961325966,
556:     -100.0,
557:     -91.9117647058824,
558:     -81.2282734646581,
559:     -17.035110533160,
560:     -77.0491803278687,
561:     -84.4106463878328,
562:     -91.2087912087912,
563:     -38.1275440976934,
564:     -59.8116169544741,
565:     -60.8455882352942,
566:     -98.7368421052631,
567:     -56.7398119122258,
568:     -66.6666666666667,
569:     -21.4028776978417,
570:     -19.2307692307692,
571:     -96.8973747016707,
572:     -82.964889466840,
573:     -69.9876998769988,
574:     -43.3392539964476,
575:     -100.0,
576:     -99.3920972644377,
577:     -62.621359223301,
578:     -20.0,
579:     -38.9863547758285,
580:     -43.6194895591646,
581:     -79.5275590551181,
582:     -91.6874480465503,
583:     -12.099644128114,
584:     -41.6666666666667,
585:     -12.8865979381443,
586:     -94.0438871473355,
587:     -22.5454545454547,
588:     -15.7894736842105,
589:     -81.8181818181818,
590:     -78.1249999999999,
591:     -73.716012084592,
592:     -73.7373737373737,
593:     -94.3693693693694,
594:     0.0,
595:     -37.5999999999999,
596:     -1.21457489878547,
597:     -73.9644970414202,
598:     -36.2222222222221,
599:     -68.5459940652819,
600:     -51.9287833827892,
601:     -52.5210084033614,
602:     -4.4404973357016,
603:     -10.2076124567475,
604:     -5.66037735849050,
605:     -70.1657458563535,
606:     -66.6253101736973,
607:     -23.8663484486874,
608:     -68.8212927756656,
609:     -66.6666666666667,
610:     -45.3781512605042,
611:     -25.6157635467982,
612:     -10.5263157894737,
613:     -5.84112149532710,
614:     -38.4155455904334,
615:     -48.9208633093525,
616:     -88.6524822695035,
617:     -90.1098901098901,
618:     -49.2730210016155,
619:     -56.5428109854604,
620:     -56.0632688927944,
621:     -9.63391136801542,
622:     -19.9623352165725,
623:     -56.1933534743202,
624:     -11.8959107806689,
625:     -100.0,
626:     -74.590163934426,
627:     -29.7413793103448,
628:     -35.7414448669201,
629:     -87.5,
630:     -90.2061855670104,
631:     -67.7835051546391,
632: };
633: 
634: test "WilliamsPercentR update period=2" {
635:     const tolerance = 1e-6;
636:     const allocator = testing.allocator;
637: 
638:     var w = try WilliamsPercentR.init(allocator, .{ .length = 2 });
639:     defer w.deinit();
640: 
641:     for (0..test_close.len) |i| {
642:         const act = w.update(test_close[i], test_high[i], test_low[i]);
643:         const exp = test_expected_2[i];
644: 
645:         if (math.isNan(exp)) {
646:             try testing.expect(math.isNan(act));
647:         } else {
648:             try testing.expect(!math.isNan(act));
649:             try testing.expect(almostEqual(act, exp, tolerance));
650:         }
651:     }
652: }
653: 
654: test "WilliamsPercentR NaN passthrough" {
655:     const allocator = testing.allocator;
656:     var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
657:     defer w.deinit();
658: 
659:     try testing.expect(math.isNan(w.update(math.nan(f64), 1, 1)));
660:     try testing.expect(math.isNan(w.update(1, math.nan(f64), 1)));
661:     try testing.expect(math.isNan(w.update(1, 1, math.nan(f64))));
662: }
663: 
664: test "WilliamsPercentR isPrimed" {
665:     const allocator = testing.allocator;
666:     var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
667:     defer w.deinit();
668: 
669:     try testing.expect(!w.isPrimed());
670: 
671:     for (0..13) |i| {
672:         _ = w.update(test_close[i], test_high[i], test_low[i]);
673:         try testing.expect(!w.isPrimed());
674:     }
675: 
676:     _ = w.update(test_close[13], test_high[13], test_low[13]);
677:     try testing.expect(w.isPrimed());
678: }
679: 
680: test "WilliamsPercentR updateSample" {
681:     const allocator = testing.allocator;
682:     var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
683:     defer w.deinit();
684: 
685:     for (0..13) |_| {
686:         const v = w.updateSample(9.0);
687:         try testing.expect(math.isNan(v));
688:     }
689: 
690:     const v = w.updateSample(9.0);
691:     try testing.expectEqual(@as(f64, 0.0), v);
692: }
693: 
694: test "WilliamsPercentR metadata" {
695:     const allocator = testing.allocator;
696:     var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
697:     defer w.deinit();
698: 
699:     var meta: Metadata = undefined;
700:     w.getMetadata(&meta);
701: 
702:     try testing.expectEqual(Identifier.williams_percent_r, meta.identifier);
703:     try testing.expectEqualStrings("willr", meta.mnemonic);
704:     try testing.expectEqualStrings("Williams %R", meta.description);
705:     try testing.expectEqual(@as(usize, 1), meta.outputs_len);
706: }
707: 
708: test "WilliamsPercentR updateBar" {
709:     const allocator = testing.allocator;
710:     var w = try WilliamsPercentR.init(allocator, .{ .length = 14 });
711:     defer w.deinit();
712: 
713:     for (0..14) |i| {
714:         _ = w.update(test_close[i], test_high[i], test_low[i]);
715:     }
716: 
717:     const bar = Bar{ .time = 42, .open = 0, .high = test_high[14], .low = test_low[14], .close = test_close[14], .volume = 0 };
718:     const out = w.updateBar(&bar);
719:     try testing.expectEqual(@as(usize, 1), out.len);
720: }
721: 
722: test "WilliamsPercentR invalid length" {
723:     const allocator = testing.allocator;
724:     const result = WilliamsPercentR.init(allocator, .{ .length = 1 });
725:     try testing.expectError(error.InvalidLength, result);
726: }

(End of file - total 726 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 7.9s)

Now let me also check the existing indicators.zig registration and an existing no-allocator HLCV indicator pattern:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig</path>
<type>file</type>
<content>1: // Root module for the indicators library.
2: // Re-exports all core types and sub-modules.
3: 
4: // --- Core enums ---
5: pub const shape = @import("core/outputs/shape.zig");
6: pub const role = @import("core/role.zig");
7: pub const pane = @import("core/pane.zig");
8: pub const adaptivity = @import("core/adaptivity.zig");
9: pub const input_requirement = @import("core/input_requirement.zig");
10: pub const volume_usage = @import("core/volume_usage.zig");
11: 
12: // --- Core types ---
13: pub const identifier = @import("core/identifier.zig");
14: pub const output_descriptor = @import("core/output_descriptor.zig");
15: pub const output_metadata = @import("core/output_metadata.zig");
16: pub const metadata = @import("core/metadata.zig");
17: pub const descriptor = @import("core/descriptor.zig");
18: pub const descriptors = @import("core/descriptors.zig");
19: pub const build_metadata = @import("core/build_metadata.zig");
20: pub const component_triple_mnemonic = @import("core/component_triple_mnemonic.zig");
21: 
22: // --- Indicator interface and outputs ---
23: pub const indicator = @import("core/indicator.zig");
24: pub const line_indicator = @import("core/line_indicator.zig");
25: 
26: // --- Output types ---
27: pub const band = @import("core/outputs/band.zig");
28: pub const heatmap = @import("core/outputs/heatmap.zig");
29: pub const polyline = @import("core/outputs/polyline.zig");
30: 
31: // --- Convenience type aliases ---
32: pub const Identifier = identifier.Identifier;
33: pub const Shape = shape.Shape;
34: pub const Role = role.Role;
35: pub const Pane = pane.Pane;
36: pub const Adaptivity = adaptivity.Adaptivity;
37: pub const InputRequirement = input_requirement.InputRequirement;
38: pub const VolumeUsage = volume_usage.VolumeUsage;
39: pub const OutputDescriptor = output_descriptor.OutputDescriptor;
40: pub const OutputMetadata = output_metadata.OutputMetadata;
41: pub const Metadata = metadata.Metadata;
42: pub const Descriptor = descriptor.Descriptor;
43: pub const Indicator = indicator.Indicator;
44: pub const OutputValue = indicator.OutputValue;
45: pub const OutputArray = indicator.OutputArray;
46: pub const LineIndicator = line_indicator.LineIndicator;
47: pub const Band = band.Band;
48: pub const Heatmap = heatmap.Heatmap;
49: pub const Polyline = polyline.Polyline;
50: pub const Point = polyline.Point;
51: 
52: // --- Functions ---
53: pub const descriptorOf = descriptor.descriptorOf;
54: pub const allDescriptors = descriptor.allDescriptors;
55: pub const buildMetadata = build_metadata.buildMetadata;
56: pub const componentTripleMnemonic = component_triple_mnemonic.componentTripleMnemonic;
57: 
58: // --- Indicator implementations ---
59: pub const simple_moving_average = @import("common/simple_moving_average/simple_moving_average.zig");
60: pub const weighted_moving_average = @import("common/weighted_moving_average/weighted_moving_average.zig");
61: pub const triangular_moving_average = @import("common/triangular_moving_average/triangular_moving_average.zig");
62: pub const exponential_moving_average = @import("common/exponential_moving_average/exponential_moving_average.zig");
63: pub const momentum = @import("common/momentum/momentum.zig");
64: pub const rate_of_change = @import("common/rate_of_change/rate_of_change.zig");
65: pub const rate_of_change_percent = @import("common/rate_of_change_percent/rate_of_change_percent.zig");
66: pub const rate_of_change_ratio = @import("common/rate_of_change_ratio/rate_of_change_ratio.zig");
67: pub const variance = @import("common/variance/variance.zig");
68: pub const standard_deviation = @import("common/standard_deviation/standard_deviation.zig");
69: pub const absolute_price_oscillator = @import("common/absolute_price_oscillator/absolute_price_oscillator.zig");
70: pub const linear_regression = @import("common/linear_regression/linear_regression.zig");
71: pub const pearsons_correlation_coefficient = @import("common/pearsons_correlation_coefficient/pearsons_correlation_coefficient.zig");
72: pub const balance_of_power = @import("igor_livshin/balance_of_power/balance_of_power.zig");
73: pub const on_balance_volume = @import("joseph_granville/on_balance_volume/on_balance_volume.zig");
74: pub const commodity_channel_index = @import("donald_lambert/commodity_channel_index/commodity_channel_index.zig");
75: pub const triple_exponential_moving_average_oscillator = @import("jack_hutson/triple_exponential_moving_average_oscillator/triple_exponential_moving_average_oscillator.zig");
76: pub const money_flow_index = @import("gene_quong/money_flow_index/money_flow_index.zig");
77: pub const kaufman_adaptive_moving_average = @import("perry_kaufman/kaufman_adaptive_moving_average/kaufman_adaptive_moving_average.zig");
78: pub const jurik_moving_average = @import("mark_jurik/jurik_moving_average/jurik_moving_average.zig");
79: pub const stochastic = @import("george_lane/stochastic/stochastic.zig");
80: pub const adaptive_trend_and_cycle_filter = @import("vladimir_kravchuk/adaptive_trend_and_cycle_filter/adaptive_trend_and_cycle_filter.zig");
81: pub const double_exponential_moving_average = @import("patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig");
82: pub const triple_exponential_moving_average = @import("patrick_mulloy/triple_exponential_moving_average/triple_exponential_moving_average.zig");
83: pub const t2_exponential_moving_average = @import("tim_tillson/t2_exponential_moving_average/t2_exponential_moving_average.zig");
84: pub const t3_exponential_moving_average = @import("tim_tillson/t3_exponential_moving_average/t3_exponential_moving_average.zig");
85: pub const percentage_price_oscillator = @import("gerald_appel/percentage_price_oscillator/percentage_price_oscillator.zig");
86: pub const moving_average_convergence_divergence = @import("gerald_appel/moving_average_convergence_divergence/moving_average_convergence_divergence.zig");
87: pub const bollinger_bands = @import("john_bollinger/bollinger_bands/bollinger_bands.zig");
88: pub const bollinger_bands_trend = @import("john_bollinger/bollinger_bands_trend/bollinger_bands_trend.zig");
89: pub const williams_percent_r = @import("larry_williams/williams_percent_r/williams_percent_r.zig");
90: pub const ultimate_oscillator = @import("larry_williams/ultimate_oscillator/ultimate_oscillator.zig");
91: pub const advance_decline = @import("marc_chaikin/advance_decline/advance_decline.zig");
92: pub const advance_decline_oscillator = @import("marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig");
93: pub const aroon = @import("tushar_chande/aroon/aroon.zig");
94: pub const chande_momentum_oscillator = @import("tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig");
95: 
96: // Force-include tests from sub-modules.
97: comptime {
98:     _ = identifier;
99:     _ = simple_moving_average;
100:     _ = weighted_moving_average;
101:     _ = triangular_moving_average;
102:     _ = exponential_moving_average;
103:     _ = momentum;
104:     _ = rate_of_change;
105:     _ = rate_of_change_percent;
106:     _ = rate_of_change_ratio;
107:     _ = variance;
108:     _ = standard_deviation;
109:     _ = absolute_price_oscillator;
110:     _ = linear_regression;
111:     _ = pearsons_correlation_coefficient;
112:     _ = balance_of_power;
113:     _ = on_balance_volume;
114:     _ = commodity_channel_index;
115:     _ = triple_exponential_moving_average_oscillator;
116:     _ = money_flow_index;
117:     _ = kaufman_adaptive_moving_average;
118:     _ = jurik_moving_average;
119:     _ = stochastic;
120:     _ = adaptive_trend_and_cycle_filter;
121:     _ = double_exponential_moving_average;
122:     _ = triple_exponential_moving_average;
123:     _ = t2_exponential_moving_average;
124:     _ = t3_exponential_moving_average;
125:     _ = percentage_price_oscillator;
126:     _ = moving_average_convergence_divergence;
127:     _ = bollinger_bands;
128:     _ = bollinger_bands_trend;
129:     _ = williams_percent_r;
130:     _ = ultimate_oscillator;
131:     _ = advance_decline;
132:     _ = advance_decline_oscillator;
133:     _ = aroon;
134:     _ = chande_momentum_oscillator;
135: }

(End of file - total 135 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/marc_chaikin/advance_decline/advance_decline.zig",
  "limit": 80
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/marc_chaikin/advance_decline/advance_decline.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: const quote_component = @import("quote_component");
9: const trade_component = @import("trade_component");
10: 
11: const indicator_mod = @import("../../core/indicator.zig");
12: const line_indicator_mod = @import("../../core/line_indicator.zig");
13: const build_metadata_mod = @import("../../core/build_metadata.zig");
14: const identifier_mod = @import("../../core/identifier.zig");
15: const metadata_mod = @import("../../core/metadata.zig");
16: 
17: const OutputArray = indicator_mod.OutputArray;
18: const LineIndicator = line_indicator_mod.LineIndicator;
19: const Identifier = identifier_mod.Identifier;
20: const Metadata = metadata_mod.Metadata;
21: 
22: /// Enumerates the outputs of the Advance-Decline indicator.
23: pub const AdvanceDeclineOutput = enum(u8) {
24:     /// The scalar value of the A/D line.
25:     value = 1,
26: };
27: 
28: /// Marc Chaikin's Advance-Decline (A/D) Line.
29: ///
30: /// The Accumulation/Distribution Line is a cumulative indicator that uses volume
31: /// and price to assess whether a stock is being accumulated or distributed.
32: ///
33: /// CLV = ((Close - Low) - (High - Close)) / (High - Low)
34: /// AD  = AD_previous + CLV × Volume
35: ///
36: /// When High equals Low, the A/D value is unchanged (no division by zero).
37: pub const AdvanceDecline = struct {
38:     line: LineIndicator,
39:     ad: f64,
40:     value: f64,
41:     primed: bool,
42: 
43:     const mnemonic_str = "ad";
44:     const description_str = "Advance-Decline";
45: 
46:     pub fn init() AdvanceDecline {
47:         return .{
48:             .line = LineIndicator.new(
49:                 mnemonic_str,
50:                 description_str,
51:                 null,
52:                 null,
53:                 null,
54:             ),
55:             .ad = 0,
56:             .value = math.nan(f64),
57:             .primed = false,
58:         };
59:     }
60: 
61:     pub fn deinit(_: *AdvanceDecline) void {}
62: 
63:     pub fn fixSlices(_: *AdvanceDecline) void {}
64: 
65:     /// Core update logic. For scalar/quote/trade, H=L=C so AD is unchanged.
66:     pub fn update(self: *AdvanceDecline, sample: f64) f64 {
67:         if (math.isNan(sample)) {
68:             return math.nan(f64);
69:         }
70:         return self.updateHlcv(sample, sample, sample, 1);
71:     }
72: 
73:     /// Updates the indicator with the given high, low, close, and volume values.
74:     pub fn updateHlcv(self: *AdvanceDecline, high: f64, low: f64, close: f64, volume: f64) f64 {
75:         if (math.isNan(high) or math.isNan(low) or math.isNan(close) or math.isNan(volume)) {
76:             return math.nan(f64);
77:         }
78: 
79:         const temp = high - low;
80:         if (temp > 0) {

(Showing lines 1-80 of 477. Use offset=81 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 131.3s)

Good, I have all the context. Now let me write all three indicator files in parallel.

**Tool: write**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/welles_wilder/true_range/true_range.zig",
  "content": "const std = @import(\"std\");\nconst math = std.math;\n\nconst Bar = @import(\"bar\").Bar;\nconst Quote = @import(\"quote\").Quote;\nconst Trade = @import(\"trade\").Trade;\nconst Scalar = @import(\"scalar\").Scalar;\n\nconst indicator_mod = @import(\"../../core/indicator.zig\");\nconst build_metadata_mod = @import(\"../../core/build_metadata.zig\");\nconst identifier_mod = @import(\"../../core/identifier.zig\");\nconst metadata_mod = @import(\"../../core/metadata.zig\");\n\nconst OutputArray = indicator_mod.OutputArray;\nconst Identifier = identifier_mod.Identifier;\nconst Metadata = metadata_mod.Metadata;\n\n/// Enumerates the outputs of the True Range indicator.\npub const TrueRangeOutput = enum(u8) {\n    /// The scalar value of the True Range.\n    value = 1,\n};\n\n/// Welles Wilder's True Range indicator.\n///\n/// The True Range is defined as the largest of:\n///   - the distance from today's high to today's low\n///   - the distance from yesterday's close to today's high\n///   - the distance from yesterday's close to today's low\n///\n/// The first update stores the close and returns NaN (not primed).\n/// The indicator is primed from the second update onward.\npub const TrueRange = struct {\n    previous_close: f64,\n    value: f64,\n    primed: bool,\n\n    const mnemonic_str = \"tr\";\n    const description_str = \"True Range\";\n\n    pub fn init() TrueRange {\n        return .{\n            .previous_close = math.nan(f64),\n            .value = math.nan(f64),\n            .primed = false,\n        };\n    }\n\n    pub fn deinit(_: *TrueRange) void {}\n    pub fn fixSlices(_: *TrueRange) void {}\n\n    /// Core update given close, high, low values.\n    pub fn update(self: *TrueRange, close: f64, high: f64, low: f64) f64 {\n        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) {\n            return math.nan(f64);\n        }\n\n        if (!self.primed) {\n            if (math.isNan(self.previous_close)) {\n                self.previous_close = close;\n                return math.nan(f64);\n            }\n            self.primed = true;\n        }\n\n        var greatest = high - low;\n\n        const temp1 = @abs(high - self.previous_close);\n        if (greatest < temp1) greatest = temp1;\n\n        const temp2 = @abs(low - self.previous_close);\n        if (greatest < temp2) greatest = temp2;\n\n        self.value = greatest;\n        self.previous_close = close;\n\n        return self.value;\n    }\n\n    /// Update using a single sample value as substitute for high, low, close.\n    pub fn updateSample(self: *TrueRange, sample: f64) f64 {\n        return self.update(sample, sample, sample);\n    }\n\n    pub fn isPrimed(self: *const TrueRange) bool {\n        return self.primed;\n    }\n\n    pub fn getMetadata(_: *const TrueRange, out: *Metadata) void {\n        build_metadata_mod.buildMetadata(out, Identifier.true_range, mnemonic_str, description_str, &.{\n            .{ .mnemonic = mnemonic_str, .description = description_str },\n        });\n    }\n\n    fn makeOutput(self: *const TrueRange, time: i64) OutputArray {\n        var out = OutputArray{};\n        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });\n        return out;\n    }\n\n    pub fn updateScalar(self: *TrueRange, sample: *const Scalar) OutputArray {\n        _ = self.update(sample.value, sample.value, sample.value);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateBar(self: *TrueRange, sample: *const Bar) OutputArray {\n        _ = self.update(sample.close, sample.high, sample.low);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateQuote(self: *TrueRange, sample: *const Quote) OutputArray {\n        const mid = (sample.bid_price + sample.ask_price) / 2.0;\n        _ = self.update(mid, mid, mid);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateTrade(self: *TrueRange, sample: *const Trade) OutputArray {\n        _ = self.update(sample.price, sample.price, sample.price);\n        return self.makeOutput(sample.time);\n    }\n\n    // --- Indicator interface ---\n\n    pub fn indicator(self: *TrueRange) indicator_mod.Indicator {\n        return indicator_mod.Indicator{\n            .ptr = @ptrCast(self),\n            .vtable = &.{\n                .isPrimed = vtableIsPrimed,\n                .metadata = vtableMetadata,\n                .updateScalar = vtableUpdateScalar,\n                .updateBar = vtableUpdateBar,\n                .updateQuote = vtableUpdateQuote,\n                .updateTrade = vtableUpdateTrade,\n            },\n        };\n    }\n\n    fn vtableIsPrimed(ptr: *const anyopaque) bool {\n        const self: *const TrueRange = @ptrCast(@alignCast(ptr));\n        return self.isPrimed();\n    }\n\n    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {\n        const self: *const TrueRange = @ptrCast(@alignCast(ptr));\n        self.getMetadata(out);\n    }\n\n    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {\n        const self: *TrueRange = @ptrCast(@alignCast(ptr));\n        return self.updateScalar(sample);\n    }\n\n    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {\n        const self: *TrueRange = @ptrCast(@alignCast(ptr));\n        return self.updateBar(sample);\n    }\n\n    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {\n        const self: *TrueRange = @ptrCast(@alignCast(ptr));\n        return self.updateQuote(sample);\n    }\n\n    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {\n        const self: *TrueRange = @ptrCast(@alignCast(ptr));\n        return self.updateTrade(sample);\n    }\n};\n\n// ============================================================================\n// Tests\n// ============================================================================\n\nconst testing = std.testing;\n\nfn almostEqual(a: f64, b: f64, tolerance: f64) bool {\n    if (math.isNan(a) and math.isNan(b)) return true;\n    if (math.isNan(a) or math.isNan(b)) return false;\n    return @abs(a - b) <= tolerance;\n}\n\nconst test_high = [_]f64{\n    93.25, 94.94, 96.375, 96.19, 96.0, 94.72, 95.0, 93.72, 92.47, 92.75,\n    96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,\n    90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,\n    87.155, 87.25, 87.375, 88.97, 90.0, 89.845, 86.97, 85.94, 84.75, 85.47,\n    84.47, 88.5, 89.47, 90.0, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,\n    90.0, 88.875, 89.0, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,\n    90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92.0, 91.94,\n    89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,\n    104.5, 105.5, 106.125, 107.94, 106.25, 107.0, 108.75, 110.94, 110.94, 114.22,\n    123.0, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0, 118.315,\n    116.87, 116.75, 113.87, 114.62, 115.31, 116.0, 121.69, 119.87, 120.87, 116.75,\n    116.5, 116.0, 118.31, 121.5, 122.0, 121.44, 125.75, 127.75, 124.19, 124.44,\n    125.75, 124.69, 125.31, 132.0, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,\n    138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,\n    125.12, 126.5, 128.69, 126.62, 126.69, 126.0, 123.12, 121.87, 124.0, 127.0,\n    124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,\n    124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126.0, 128.5, 126.75, 129.75,\n    132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135.0, 132.38, 131.44, 130.88,\n    129.63, 127.25, 127.81, 125.0, 126.81, 124.75, 122.81, 122.25, 121.06, 120.0,\n    123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109.0,\n    108.5, 113.06, 93.0, 94.62, 95.12, 96.0, 95.56, 95.31, 99.0, 98.81,\n    96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,\n    95.0, 94.87, 98.25, 105.12, 108.44, 109.87, 105.0, 106.0, 104.94, 104.5,\n    104.44, 106.31, 112.87, 116.5, 119.19, 121.0, 122.12, 111.94, 112.75, 110.19,\n    107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110.0, 110.75, 110.5,\n    110.5, 109.5,\n};\n\nconst test_low = [_]f64{\n    90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92.0, 89.75, 89.44, 90.625,\n    92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,\n    87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,\n    84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,\n    82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89.0, 88.565, 90.095,\n    89.0, 86.47, 84.0, 83.315, 82.0, 83.25, 84.75, 85.28, 87.19, 88.44,\n    88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92.0, 90.53, 89.97, 88.815,\n    86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,\n    101.72, 101.72, 103.155, 105.69, 103.655, 104.0, 105.53, 108.53, 108.75, 107.75,\n    117.0, 118.0, 116.0, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,\n    112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,\n    112.56, 114.0, 114.75, 118.87, 119.0, 119.75, 122.62, 123.0, 121.75, 121.56,\n    123.12, 122.19, 122.75, 124.37, 128.0, 129.5, 130.81, 130.63, 132.13, 133.88,\n    135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,\n    122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121.0,\n    121.12, 118.94, 119.81, 121.0, 122.0, 124.5, 126.56, 123.5, 121.25, 121.06,\n    122.31, 121.0, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128.0,\n    128.38, 130.69, 131.63, 134.38, 132.0, 131.94, 131.94, 129.56, 123.75, 126.0,\n    126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120.0, 119.62, 115.5, 116.75,\n    119.06, 119.06, 115.06, 111.06, 113.12, 110.0, 105.0, 104.69, 103.87, 104.69,\n    105.44, 107.0, 89.0, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,\n    93.69, 93.5, 90.0, 90.19, 90.5, 92.12, 94.12, 94.87, 93.0, 93.87,\n    93.0, 92.62, 93.56, 98.37, 104.44, 106.0, 101.81, 104.12, 103.37, 102.12,\n    102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,\n    104.5, 105.75, 108.62, 107.75, 108.06, 108.0, 108.19, 108.12, 109.06, 108.75,\n    108.56, 106.62,\n};\n\nconst test_close = [_]f64{\n    91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,\n    96.125, 97.25, 98.5, 89.875, 91.0, 92.815, 89.155, 89.345, 91.625, 89.875,\n    88.375, 87.625, 84.78, 83.0, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,\n    85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,\n    83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89.0, 91.0, 90.5,\n    89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,\n    88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90.0, 89.69,\n    88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106.0, 102.5,\n    102.405, 104.595, 106.125, 106.0, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,\n    123.0, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,\n    116.0, 116.0, 112.0, 113.75, 112.94, 116.0, 120.5, 116.62, 117.0, 115.25,\n    114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,\n    123.12, 122.56, 124.62, 129.25, 131.0, 132.25, 131.0, 132.81, 134.0, 137.38,\n    137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129.0, 123.87, 124.81,\n    123.0, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,\n    122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,\n    124.44, 122.0, 122.37, 122.94, 124.0, 123.19, 124.56, 127.25, 125.87, 128.86,\n    132.0, 130.75, 134.75, 135.0, 132.38, 133.31, 131.94, 130.0, 125.37, 130.13,\n    127.12, 125.19, 122.0, 125.0, 123.0, 123.5, 120.06, 121.0, 117.75, 119.87,\n    122.0, 119.19, 116.37, 113.5, 114.25, 110.0, 105.06, 107.0, 107.87, 107.0,\n    107.12, 107.0, 91.0, 93.94, 93.87, 95.5, 93.0, 94.94, 98.25, 96.75,\n    94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97.0, 95.0, 95.87, 94.06,\n    94.62, 93.75, 98.0, 103.94, 107.87, 106.06, 104.5, 105.0, 104.19, 103.06,\n    103.42, 105.27, 111.87, 116.0, 116.62, 118.28, 113.37, 109.0, 109.7, 109.25,\n    107.0, 109.19, 110.0, 109.2, 110.12, 108.0, 108.62, 109.75, 109.81, 109.0,\n    108.75, 107.87,\n};\n\nconst test_expected_tr = [_]f64{\n    math.nan(f64), 3.535, 2.125, 2.69, 3.185, 1.22, 3.0, 3.97, 3.31, 2.435,\n    3.78, 3.5, 3.095, 9.685, 4.565, 2.31, 4.5, 1.875, 2.72, 2.5,\n    2.845, 1.97, 3.625, 3.22, 2.875, 3.875, 3.19, 5.34, 3.655, 3.155,\n    2.75, 2.155, 1.875, 3.44, 2.125, 3.28, 2.315, 3.565, 2.31, 2.03,\n    1.94, 5.125, 3.97, 1.47, 3.16, 1.315, 2.22, 2.72, 2.59, 1.655,\n    1.5, 2.56, 5.0, 1.935, 1.815, 2.56, 1.875, 2.66, 3.185, 2.185,\n    2.5, 1.5, 3.47, 2.28, 4.285, 2.875, 2.03, 2.625, 2.03, 3.125,\n    3.0, 3.81, 4.125, 3.375, 3.375, 13.435, 6.81, 5.5, 3.375, 4.25,\n    2.78, 3.78, 2.97, 2.25, 2.595, 3.0, 4.125, 2.41, 2.19, 6.47,\n    10.25, 5.0, 3.815, 1.815, 2.845, 1.94, 2.095, 4.47, 2.5, 7.72,\n    5.505, 2.56, 4.81, 5.18, 3.75, 3.56, 5.69, 4.44, 4.31, 3.69,\n    3.94, 2.0, 3.56, 5.63, 3.0, 1.69, 5.0, 4.75, 2.44, 2.88,\n    3.19, 2.5, 2.75, 7.63, 3.31, 2.75, 3.07, 2.87, 3.37, 3.56,\n    3.31, 3.44, 2.31, 3.63, 2.12, 5.19, 8.57, 2.88, 5.5, 2.82,\n    2.37, 3.5, 2.88, 3.76, 2.32, 4.19, 4.93, 3.81, 6.44, 6.0,\n    3.32, 3.56, 4.44, 2.81, 3.38, 4.5, 2.0, 6.13, 3.62, 3.31,\n    3.12, 3.44, 3.19, 3.81, 2.44, 2.93, 3.13, 3.94, 3.0, 3.88,\n    4.31, 3.25, 5.75, 3.31, 3.56, 1.62, 3.06, 2.82, 7.69, 5.51,\n    3.88, 2.88, 6.37, 4.56, 5.44, 3.06, 3.5, 2.63, 5.56, 3.25,\n    4.19, 3.69, 4.13, 5.31, 3.57, 4.87, 5.87, 2.56, 5.0, 4.31,\n    3.06, 6.06, 18.0, 3.62, 3.0, 2.13, 2.75, 2.31, 4.06, 2.44,\n    3.12, 2.44, 4.44, 2.75, 3.69, 3.38, 3.44, 2.63, 3.25, 2.5,\n    2.0, 2.25, 4.69, 7.12, 4.5, 3.87, 4.25, 1.88, 1.63, 2.38,\n    2.19, 2.94, 7.6, 4.63, 3.75, 5.5, 9.87, 5.81, 6.19, 3.32,\n    4.75, 3.94, 2.44, 2.69, 2.06, 2.31, 2.44, 1.88, 1.69, 1.75,\n    1.94, 2.88,\n};\n\ntest \"TrueRange update\" {\n    const tolerance = 1e-3;\n    var tr = TrueRange.init();\n\n    for (0..test_close.len) |i| {\n        const act = tr.update(test_close[i], test_high[i], test_low[i]);\n        const exp = test_expected_tr[i];\n\n        if (math.isNan(exp)) {\n            try testing.expect(math.isNan(act));\n        } else {\n            try testing.expect(!math.isNan(act));\n            try testing.expect(almostEqual(act, exp, tolerance));\n        }\n    }\n}\n\ntest \"TrueRange NaN passthrough\" {\n    var tr = TrueRange.init();\n\n    try testing.expect(math.isNan(tr.update(math.nan(f64), 1, 1)));\n    try testing.expect(math.isNan(tr.update(1, math.nan(f64), 1)));\n    try testing.expect(math.isNan(tr.update(1, 1, math.nan(f64))));\n}\n\ntest \"TrueRange isPrimed\" {\n    var tr = TrueRange.init();\n\n    try testing.expect(!tr.isPrimed());\n\n    _ = tr.update(test_close[0], test_high[0], test_low[0]);\n    try testing.expect(!tr.isPrimed());\n\n    _ = tr.update(test_close[1], test_high[1], test_low[1]);\n    try testing.expect(tr.isPrimed());\n\n    _ = tr.update(test_close[2], test_high[2], test_low[2]);\n    try testing.expect(tr.isPrimed());\n}\n\ntest \"TrueRange updateSample\" {\n    var tr = TrueRange.init();\n\n    const v0 = tr.updateSample(100.0);\n    try testing.expect(math.isNan(v0));\n\n    const v1 = tr.updateSample(105.0);\n    try testing.expect(almostEqual(v1, 5.0, 1e-10));\n\n    const v2 = tr.updateSample(102.0);\n    try testing.expect(almostEqual(v2, 3.0, 1e-10));\n}\n\ntest \"TrueRange metadata\" {\n    var tr = TrueRange.init();\n    var meta: Metadata = undefined;\n    tr.getMetadata(&meta);\n\n    try testing.expectEqual(Identifier.true_range, meta.identifier);\n    try testing.expectEqualStrings(\"tr\", meta.mnemonic);\n    try testing.expectEqualStrings(\"True Range\", meta.description);\n    try testing.expectEqual(@as(usize, 1), meta.outputs_len);\n}\n\ntest \"TrueRange updateBar\" {\n    var tr = TrueRange.init();\n\n    // Prime with first bar.\n    _ = tr.update(100, 105, 95);\n\n    const bar = Bar{ .time = 42, .open = 0, .high = 110, .low = 98, .close = 108, .volume = 0 };\n    const out = tr.updateBar(&bar);\n    try testing.expectEqual(@as(usize, 1), out.len);\n    try testing.expect(!math.isNan(out.slice()[0].scalar.value));\n}\n"
}
```

**Output:**
```
Wrote file successfully.

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/icalc/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.identifier" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.factory.factory" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/iconf/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.descriptors" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.core.outputs.heatmap" could not be resolved
ERROR [26:6] Import "py.indicators.core.outputs.shape" could not be resolved
ERROR [27:6] Import "py.indicators.core.pane" could not be resolved
ERROR [28:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [31:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/ifres/main.py">
ERROR [16:6] Import "py.indicators.core.frequency_response" could not be resolved
ERROR [17:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [18:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [21:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>
```

---

## Assistant (Build · claude-opus-4.6 · 192.9s)

Now write DM+ and DM-:

**Tool: write**

**Input:**
```json
{
  "content": "const std = @import(\"std\");\nconst math = std.math;\n\nconst Bar = @import(\"bar\").Bar;\nconst Quote = @import(\"quote\").Quote;\nconst Trade = @import(\"trade\").Trade;\nconst Scalar = @import(\"scalar\").Scalar;\n\nconst indicator_mod = @import(\"../../core/indicator.zig\");\nconst build_metadata_mod = @import(\"../../core/build_metadata.zig\");\nconst identifier_mod = @import(\"../../core/identifier.zig\");\nconst metadata_mod = @import(\"../../core/metadata.zig\");\n\nconst OutputArray = indicator_mod.OutputArray;\nconst Identifier = identifier_mod.Identifier;\nconst Metadata = metadata_mod.Metadata;\n\n/// Enumerates the outputs of the Directional Movement Plus indicator.\npub const DirectionalMovementPlusOutput = enum(u8) {\n    /// The scalar value of +DM.\n    value = 1,\n};\n\n/// Parameters for the Directional Movement Plus indicator.\npub const DirectionalMovementPlusParams = struct {\n    /// The smoothing length. Must be >= 1. Default is 14.\n    length: usize = 14,\n};\n\n/// Welles Wilder's Directional Movement Plus (+DM) indicator.\n///\n/// UpMove = today's high − yesterday's high\n/// DownMove = yesterday's low − today's low\n/// if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0\n///\n/// When length > 1, Wilder's smoothing is applied:\n///   +DM(n) = previous +DM(n) − previous +DM(n)/n + today's +DM(1)\npub const DirectionalMovementPlus = struct {\n    length: usize,\n    no_smoothing: bool,\n    count: usize,\n    previous_high: f64,\n    previous_low: f64,\n    value: f64,\n    accumulator: f64,\n    primed: bool,\n\n    pub const Error = error{InvalidLength};\n\n    pub fn init(params: DirectionalMovementPlusParams) Error!DirectionalMovementPlus {\n        if (params.length < 1) return error.InvalidLength;\n\n        return .{\n            .length = params.length,\n            .no_smoothing = params.length == 1,\n            .count = 0,\n            .previous_high = 0,\n            .previous_low = 0,\n            .value = math.nan(f64),\n            .accumulator = 0,\n            .primed = false,\n        };\n    }\n\n    pub fn deinit(_: *DirectionalMovementPlus) void {}\n    pub fn fixSlices(_: *DirectionalMovementPlus) void {}\n\n    /// Core update given high and low values.\n    pub fn update(self: *DirectionalMovementPlus, high_in: f64, low_in: f64) f64 {\n        if (math.isNan(high_in) or math.isNan(low_in)) {\n            return math.nan(f64);\n        }\n\n        var high = high_in;\n        var low = low_in;\n        if (high < low) {\n            const tmp = high;\n            high = low;\n            low = tmp;\n        }\n\n        if (self.no_smoothing) {\n            if (self.primed) {\n                const delta_plus = high - self.previous_high;\n                const delta_minus = self.previous_low - low;\n                if (delta_plus > 0 and delta_plus > delta_minus) {\n                    self.value = delta_plus;\n                } else {\n                    self.value = 0;\n                }\n            } else {\n                if (self.count > 0) {\n                    const delta_plus = high - self.previous_high;\n                    const delta_minus = self.previous_low - low;\n                    if (delta_plus > 0 and delta_plus > delta_minus) {\n                        self.value = delta_plus;\n                    } else {\n                        self.value = 0;\n                    }\n                    self.primed = true;\n                }\n                self.count += 1;\n            }\n        } else {\n            const n: f64 = @floatFromInt(self.length);\n            if (self.primed) {\n                const delta_plus = high - self.previous_high;\n                const delta_minus = self.previous_low - low;\n                if (delta_plus > 0 and delta_plus > delta_minus) {\n                    self.accumulator += -self.accumulator / n + delta_plus;\n                } else {\n                    self.accumulator += -self.accumulator / n;\n                }\n                self.value = self.accumulator;\n            } else {\n                if (self.count > 0 and self.length >= self.count) {\n                    const delta_plus = high - self.previous_high;\n                    const delta_minus = self.previous_low - low;\n                    if (self.length > self.count) {\n                        if (delta_plus > 0 and delta_plus > delta_minus) {\n                            self.accumulator += delta_plus;\n                        }\n                    } else {\n                        if (delta_plus > 0 and delta_plus > delta_minus) {\n                            self.accumulator += -self.accumulator / n + delta_plus;\n                        } else {\n                            self.accumulator += -self.accumulator / n;\n                        }\n                        self.value = self.accumulator;\n                        self.primed = true;\n                    }\n                }\n                self.count += 1;\n            }\n        }\n\n        self.previous_low = low;\n        self.previous_high = high;\n\n        return self.value;\n    }\n\n    /// Update using a single sample value as substitute for high and low.\n    pub fn updateSample(self: *DirectionalMovementPlus, sample: f64) f64 {\n        return self.update(sample, sample);\n    }\n\n    pub fn isPrimed(self: *const DirectionalMovementPlus) bool {\n        return self.primed;\n    }\n\n    pub fn getMetadata(_: *const DirectionalMovementPlus, out: *Metadata) void {\n        const mnemonic = \"+dm\";\n        const description = \"Directional Movement Plus\";\n        build_metadata_mod.buildMetadata(out, Identifier.directional_movement_plus, mnemonic, description, &.{\n            .{ .mnemonic = mnemonic, .description = description },\n        });\n    }\n\n    fn makeOutput(self: *const DirectionalMovementPlus, time: i64) OutputArray {\n        var out = OutputArray{};\n        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });\n        return out;\n    }\n\n    pub fn updateScalar(self: *DirectionalMovementPlus, sample: *const Scalar) OutputArray {\n        _ = self.update(sample.value, sample.value);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateBar(self: *DirectionalMovementPlus, sample: *const Bar) OutputArray {\n        _ = self.update(sample.high, sample.low);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateQuote(self: *DirectionalMovementPlus, sample: *const Quote) OutputArray {\n        const mid = (sample.bid_price + sample.ask_price) / 2.0;\n        _ = self.update(mid, mid);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateTrade(self: *DirectionalMovementPlus, sample: *const Trade) OutputArray {\n        _ = self.update(sample.price, sample.price);\n        return self.makeOutput(sample.time);\n    }\n\n    // --- Indicator interface ---\n\n    pub fn indicator(self: *DirectionalMovementPlus) indicator_mod.Indicator {\n        return indicator_mod.Indicator{\n            .ptr = @ptrCast(self),\n            .vtable = &.{\n                .isPrimed = vtableIsPrimed,\n                .metadata = vtableMetadata,\n                .updateScalar = vtableUpdateScalar,\n                .updateBar = vtableUpdateBar,\n                .updateQuote = vtableUpdateQuote,\n                .updateTrade = vtableUpdateTrade,\n            },\n        };\n    }\n\n    fn vtableIsPrimed(ptr: *const anyopaque) bool {\n        const self: *const DirectionalMovementPlus = @ptrCast(@alignCast(ptr));\n        return self.isPrimed();\n    }\n\n    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {\n        const self: *const DirectionalMovementPlus = @ptrCast(@alignCast(ptr));\n        self.getMetadata(out);\n    }\n\n    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {\n        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));\n        return self.updateScalar(sample);\n    }\n\n    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {\n        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));\n        return self.updateBar(sample);\n    }\n\n    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {\n        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));\n        return self.updateQuote(sample);\n    }\n\n    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {\n        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));\n        return self.updateTrade(sample);\n    }\n};\n\n// ============================================================================\n// Tests\n// ============================================================================\n\nconst testing = std.testing;\n\nfn almostEqual(a: f64, b: f64, tolerance: f64) bool {\n    if (math.isNan(a) and math.isNan(b)) return true;\n    if (math.isNan(a) or math.isNan(b)) return false;\n    return @abs(a - b) <= tolerance;\n}\n\nconst test_high = [_]f64{\n    93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,\n    99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000, 90.345000, 88.500000,\n    88.250000, 85.500000, 84.440000, 84.750000, 84.440000, 89.405000, 88.125000, 89.125000, 87.155000, 87.250000, 87.375000,\n    88.970000, 90.000000, 89.845000, 86.970000, 85.940000, 84.750000, 85.470000, 84.470000, 88.500000, 89.470000, 90.000000,\n    92.440000, 91.440000, 92.970000, 91.720000, 91.155000, 91.750000, 90.000000, 88.875000, 89.000000, 85.250000, 83.815000,\n    85.250000, 86.625000, 87.940000, 89.375000, 90.625000, 90.750000, 88.845000, 91.970000, 93.375000, 93.815000, 94.030000,\n    94.030000, 91.815000, 92.000000, 91.940000, 89.750000, 88.750000, 86.155000, 84.875000, 85.940000, 99.375000, 103.280000,\n    105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000,\n    110.940000, 110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000,\n    115.345000, 113.000000, 118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000,\n    119.870000, 120.870000, 116.750000, 116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000,\n    127.750000, 124.190000, 124.440000, 125.750000, 124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000,\n    133.500000, 135.500000, 137.440000, 138.690000, 139.190000, 138.500000, 138.130000, 137.500000, 138.880000, 132.130000,\n    129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000, 126.620000, 126.690000, 126.000000, 123.120000,\n    121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000, 124.500000, 127.870000, 128.560000,\n    129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000, 125.620000, 126.000000,\n    128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000, 135.000000,\n    132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,\n    122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000,\n    107.250000, 108.870000, 109.000000, 108.500000, 113.060000, 93.000000, 94.620000, 95.120000, 96.000000, 95.560000,\n    95.310000, 99.000000, 98.810000, 96.810000, 95.940000, 94.440000, 92.940000, 93.940000, 95.500000, 97.060000, 97.500000,\n    96.250000, 96.370000, 95.000000, 94.870000, 98.250000, 105.120000, 108.440000, 109.870000, 105.000000, 106.000000,\n    104.940000, 104.500000, 104.440000, 106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000,\n    112.750000, 110.190000, 107.940000, 109.690000, 111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000,\n    110.750000, 110.500000, 110.500000, 109.500000,\n};\n\nconst test_low = [_]f64{\n    90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,\n    96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000, 87.500000, 86.530000,\n    84.625000, 82.280000, 81.565000, 80.875000, 81.250000, 84.065000, 85.595000, 85.970000, 84.405000, 85.095000, 85.500000,\n    85.530000, 87.875000, 86.565000, 84.655000, 83.250000, 82.565000, 83.440000, 82.530000, 85.065000, 86.875000, 88.530000,\n    89.280000, 90.125000, 90.750000, 89.000000, 88.565000, 90.095000, 89.000000, 86.470000, 84.000000, 83.315000, 82.000000,\n    83.250000, 84.750000, 85.280000, 87.190000, 88.440000, 88.250000, 87.345000, 89.280000, 91.095000, 89.530000, 91.155000,\n    92.000000, 90.530000, 89.970000, 88.815000, 86.750000, 85.065000, 82.030000, 81.500000, 82.565000, 96.345000, 96.470000,\n    101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000,\n    108.530000, 108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000,\n    110.875000, 110.500000, 110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000,\n    116.060000, 116.560000, 113.310000, 112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000,\n    123.000000, 121.750000, 121.560000, 123.120000, 122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000,\n    130.630000, 132.130000, 133.880000, 135.380000, 135.750000, 136.190000, 134.500000, 135.380000, 133.690000, 126.060000,\n    126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000, 124.620000, 124.370000, 121.810000, 118.190000,\n    118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000, 122.000000, 124.500000, 126.560000,\n    123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000, 122.690000, 122.870000,\n    125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000, 131.940000,\n    129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,\n    119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000,\n    104.690000, 103.870000, 104.690000, 105.440000, 107.000000, 89.000000, 92.500000, 92.120000, 94.620000, 92.810000,\n    94.250000, 96.250000, 96.370000, 93.690000, 93.500000, 90.000000, 90.190000, 90.500000, 92.120000, 94.120000, 94.870000,\n    93.000000, 93.870000, 93.000000, 92.620000, 93.560000, 98.370000, 104.440000, 106.000000, 101.810000, 104.120000,\n    103.370000, 102.120000, 102.250000, 103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000,\n    106.560000, 106.870000, 104.500000, 105.750000, 108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000,\n    109.060000, 108.750000, 108.560000, 106.620000,\n};\n\nconst test_expected_dmp1 = [_]f64{\n    math.nan(f64),\n    1.6900, 1.4350, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.2800, 3.5000,\n    3.3750, 0.0000, 0.0000, 0.0000, 1.9350, 0.0000, 0.0000, 1.3150, 0.2800, 0.0000,\n    0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 4.9650, 0.0000, 1.0000, 0.0000,\n    0.0950, 0.1250, 1.5950, 1.0300, 0.0000, 0.0000, 0.0000, 0.0000, 0.7200, 0.0000,\n    4.0300, 0.9700, 0.5300, 2.4400, 0.0000, 1.5300, 0.0000, 0.0000, 0.5950, 0.0000,\n    0.0000, 0.0000, 0.0000, 0.0000, 1.4350, 1.3750, 1.3150, 1.4350, 1.2500, 0.0000,\n    0.0000, 3.1250, 1.4050, 0.0000, 0.2150, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,\n    0.0000, 0.0000, 0.0000, 1.0650, 13.4350, 3.9050, 2.0950, 2.2500, 0.0000, 0.0000,\n    1.0000, 0.6250, 1.8150, 0.0000, 0.7500, 1.7500, 2.1900, 0.0000, 3.2800, 8.7800,\n    0.0000, 0.0000, 0.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 5.3150, 0.0000,\n    0.0000, 0.0000, 0.0000, 0.6900, 0.6900, 5.6900, 0.0000, 1.0000, 0.0000, 0.0000,\n    0.0000, 2.3100, 3.1900, 0.5000, 0.0000, 4.3100, 2.0000, 0.0000, 0.2500, 1.3100,\n    0.0000, 0.6200, 6.6900, 0.0000, 0.9400, 1.6300, 0.0000, 2.0000, 1.9400, 1.2500,\n    0.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,\n    1.3800, 2.1900, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 2.1300, 3.0000, 0.0000,\n    0.0000, 1.2500, 0.0600, 0.6900, 3.3700, 0.6900, 0.0000, 0.0000, 0.0000, 0.5000,\n    0.0000, 0.4400, 1.8100, 0.0000, 0.4300, 0.3800, 2.5000, 0.0000, 3.0000, 2.9400,\n    1.2500, 2.5600, 1.1900, 0.0000, 0.0000, 1.4400, 0.0000, 0.0000, 0.0000, 0.0000,\n    0.0000, 0.0000, 0.0000, 1.8100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.2500,\n    0.0000, 0.0000, 0.0000, 1.6300, 0.0000, 0.0000, 0.0000, 1.6200, 0.1300, 0.0000,\n    4.5600, 0.0000, 1.6200, 0.5000, 0.8800, 0.0000, 0.0000, 3.6900, 0.0000, 0.0000,\n    0.0000, 0.0000, 0.0000, 1.0000, 1.5600, 1.5600, 0.4400, 0.0000, 0.1200, 0.0000,\n    0.0000, 3.3800, 6.8700, 3.3200, 1.4300, 0.0000, 1.0000, 0.0000, 0.0000, 0.0000,\n    1.8700, 6.5600, 3.6300, 2.6900, 1.8100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,\n    1.7500, 1.3700, 0.0000, 0.0000, 0.1900, 0.1300, 0.0000, 0.7500, 0.0000, 0.0000,\n    0.0000,\n};\n\nconst test_expected_dmp14 = [_]f64{\n    math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64),\n    math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64),\n    9.545714285714290, 10.798877551020400, 10.027529154519000, 9.311277072053310, 9.961185852620930, 9.529672577433720,\n    8.848981679045600, 8.216911559113770, 7.629989304891360, 7.084990068827690, 6.578919349625710, 6.108996538938160,\n    5.672639643299720, 10.232451097349700, 9.501561733253330, 9.822878752306670, 9.121244555713330, 8.564727087448090,\n    8.077960866916090, 9.095963662136360, 9.476251971983770, 8.799376831127780, 8.170849914618660, 7.587217777860180,\n    7.045273650870170, 7.262039818665150, 6.743322688760500, 10.291656782420500, 10.526538440819000, 10.304642837903400,\n    12.008596920910300, 11.150839997988100, 11.884351426703200, 11.035469181938700, 10.247221383228800, 10.110276998712500,\n    9.388114355947290, 8.717534759093910, 8.094853704872920, 7.516649868810570, 6.979746306752670, 7.916192999127480,\n    8.725750642046950, 9.417482739043590, 10.179805400540500, 10.702676443359000, 9.938199554547660, 9.228328157794250,\n    11.694161860808900, 12.263864585036900, 11.387874257534200, 10.789454667710400, 10.018779334302500, 9.303152238995170,\n    8.638641364781230, 8.021595553011140, 7.448624442081780, 6.916579839075940, 6.422538421999080, 5.963785677570580,\n    6.602800986315530, 19.566172344435900, 22.073588605547600, 22.591903705151300, 23.228196297640500, 21.569039419237600,\n    20.028393746434900, 19.597794193118200, 18.822951750752600, 19.293455197127400, 17.915351254475400, 17.385683307727200,\n    17.893848785746700, 18.805716729621900, 17.462451248934600, 19.495133302582200, 26.882623780969100, 24.962436368042800,\n    23.179405198896900, 22.023733398975700, 20.450609584763100, 18.989851757280000, 17.633433774617200, 16.373902790716000,\n    15.204338305664800, 19.433314140974500, 18.045220273762000, 16.756275968493300, 15.559399113600900, 14.448013462629400,\n    14.106012501013000, 13.788440179512100, 18.493551595261200, 17.172583624171100, 16.945970508158900, 15.735544043290400,\n    14.611576611626800, 13.567892567939200, 14.908757384515000, 17.033846142763900, 16.317142846852200, 15.151632643505600,\n    18.379373168969500, 19.066560799757400, 17.704663599774700, 16.690044771219400, 16.807898716132300, 15.607334522122800,\n    15.112524913399800, 20.723058848156900, 19.242840359002900, 18.808351761931200, 19.094898064650400, 17.730976774318300,\n    18.464478433295500, 19.085587116631600, 18.972330894015000, 18.117164401585400, 16.823081230043600, 15.621432570754700,\n    14.505615958558000, 13.469500532946700, 12.507393352021900, 11.614008112591800, 10.784436104549500, 10.014119239938800,\n    9.298825008514630, 10.014623222192200, 11.489292992035600, 10.668629206890200, 9.906584263540870, 9.198971101859380,\n    8.541901737440850, 7.931765899052220, 9.495211191977060, 11.816981821121600, 10.972911691041400, 10.189132284538500,\n    10.711337121357200, 10.006241612688800, 9.981510068925310, 12.638545064002100, 12.425791845144800, 11.538235284777300,\n    10.714075621578900, 9.948784505751860, 9.738157041055300, 9.042574395265630, 8.836676224175230, 10.015485065305600,\n    9.300093274926600, 9.065800898146130, 8.798243691135690, 10.669797713197400, 9.907669305111900, 12.199978640461000,\n    14.268551594713800, 14.499369337948600, 16.023700099523700, 16.069150092414800, 14.921353657242300, 13.855542681725000,\n    14.305861061601800, 13.284013842916000, 12.335155711279100, 11.454073160473500, 10.635925077582500, 9.876216143469460,\n    9.170772133221640, 8.515716980848670, 9.717451482216620, 9.023347804915440, 8.378822961707190, 7.780335607299540,\n    7.224597349635280, 6.708554681804190, 9.479372204532460, 8.802274189923000, 8.173540319214220, 7.589716010698920,\n    8.677593438506130, 8.057765335755700, 7.482210668916000, 6.947767049707720, 8.071497974728600, 7.624962405105120,\n    7.080322233311900, 11.134584930932500, 10.339257435865900, 11.220739047589700, 10.919257687047600, 11.019310709401400,\n    10.232217087301300, 9.501344438208310, 12.512676978336300, 11.618914337026600, 10.788991884381800, 10.018349606926000,\n    9.302753206431250, 8.638270834543300, 9.021251489218780, 9.936876382846010, 10.787099498357000, 10.456592391331500,\n    9.709692934807830, 9.136143439464420, 8.483561765216960, 7.877593067701460, 10.694907848579900, 16.800985859395600,\n    18.920915440867400, 18.999421480805400, 17.642319946462200, 17.382154236000600, 16.140571790572000, 14.987673805531100,\n    13.917125676564600, 14.793045271095700, 20.296399180303200, 22.476656381710100, 23.561180925873600, 23.688239431168400,\n    21.996222328942100, 20.425063591160500, 18.966130477506200, 17.611406871970000, 16.353449238257900, 16.935345721239500,\n    17.095678169722400, 15.874558300456500, 14.740661278995300, 13.877756901924200, 13.016488551786800, 12.086739369516300,\n    11.973400843122300, 11.118157925756400, 10.324003788202400, 9.586574946187900,\n};\n\ntest \"DirectionalMovementPlus update length=14\" {\n    const tolerance = 1e-8;\n    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });\n\n    for (0..test_high.len) |i| {\n        const act = dmp.update(test_high[i], test_low[i]);\n        const exp = test_expected_dmp14[i];\n\n        if (math.isNan(exp)) {\n            try testing.expect(math.isNan(act));\n        } else {\n            try testing.expect(!math.isNan(act));\n            try testing.expect(almostEqual(act, exp, tolerance));\n        }\n    }\n}\n\ntest \"DirectionalMovementPlus update length=1\" {\n    const tolerance = 1e-8;\n    var dmp = try DirectionalMovementPlus.init(.{ .length = 1 });\n\n    for (0..test_high.len) |i| {\n        const act = dmp.update(test_high[i], test_low[i]);\n        const exp = test_expected_dmp1[i];\n\n        if (math.isNan(exp)) {\n            try testing.expect(math.isNan(act));\n        } else {\n            try testing.expect(!math.isNan(act));\n            try testing.expect(almostEqual(act, exp, tolerance));\n        }\n    }\n}\n\ntest \"DirectionalMovementPlus constructor validation\" {\n    const result = DirectionalMovementPlus.init(.{ .length = 0 });\n    try testing.expectError(error.InvalidLength, result);\n}\n\ntest \"DirectionalMovementPlus isPrimed length=1\" {\n    var dmp = try DirectionalMovementPlus.init(.{ .length = 1 });\n\n    try testing.expect(!dmp.isPrimed());\n\n    _ = dmp.update(test_high[0], test_low[0]);\n    try testing.expect(!dmp.isPrimed());\n\n    _ = dmp.update(test_high[1], test_low[1]);\n    try testing.expect(dmp.isPrimed());\n}\n\ntest \"DirectionalMovementPlus isPrimed length=14\" {\n    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });\n\n    for (0..14) |i| {\n        _ = dmp.update(test_high[i], test_low[i]);\n        try testing.expect(!dmp.isPrimed());\n    }\n\n    _ = dmp.update(test_high[14], test_low[14]);\n    try testing.expect(dmp.isPrimed());\n}\n\ntest \"DirectionalMovementPlus NaN passthrough\" {\n    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });\n\n    try testing.expect(math.isNan(dmp.update(math.nan(f64), 1)));\n    try testing.expect(math.isNan(dmp.update(1, math.nan(f64))));\n    try testing.expect(math.isNan(dmp.update(math.nan(f64), math.nan(f64))));\n    try testing.expect(math.isNan(dmp.updateSample(math.nan(f64))));\n}\n\ntest \"DirectionalMovementPlus high/low swap\" {\n    var dmp1 = try DirectionalMovementPlus.init(.{ .length = 1 });\n    var dmp2 = try DirectionalMovementPlus.init(.{ .length = 1 });\n\n    _ = dmp1.update(10, 5);\n    _ = dmp2.update(5, 10);\n\n    const v1 = dmp1.update(12, 6);\n    const v2 = dmp2.update(6, 12);\n\n    try testing.expectEqual(v1, v2);\n}\n\ntest \"DirectionalMovementPlus metadata\" {\n    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });\n    var meta: Metadata = undefined;\n    dmp.getMetadata(&meta);\n\n    try testing.expectEqual(Identifier.directional_movement_plus, meta.identifier);\n    try testing.expectEqualStrings(\"+dm\", meta.mnemonic);\n    try testing.expectEqualStrings(\"Directional Movement Plus\", meta.description);\n    try testing.expectEqual(@as(usize, 1), meta.outputs_len);\n}\n\ntest \"DirectionalMovementPlus updateBar\" {\n    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });\n\n    for (0..14) |i| {\n        _ = dmp.update(test_high[i], test_low[i]);\n    }\n\n    const bar = Bar{ .time = 42, .open = 0, .high = test_high[14], .low = test_low[14], .close = 0, .volume = 0 };\n    const out = dmp.updateBar(&bar);\n    try testing.expectEqual(@as(usize, 1), out.len);\n}\n",
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/welles_wilder/directional_movement_plus/directional_movement_plus.zig"
}
```

**Output:**
```
Wrote file successfully.

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/icalc/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.identifier" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.factory.factory" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/iconf/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.descriptors" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.core.outputs.heatmap" could not be resolved
ERROR [26:6] Import "py.indicators.core.outputs.shape" could not be resolved
ERROR [27:6] Import "py.indicators.core.pane" could not be resolved
ERROR [28:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [31:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/ifres/main.py">
ERROR [16:6] Import "py.indicators.core.frequency_response" could not be resolved
ERROR [17:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [18:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [21:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>
```

---

## User

---

## Assistant (Compaction · claude-opus-4.6 · 77.0s)



## Goal

Port the `indicators/` module from Go/TypeScript to Zig. Python and Rust ports are complete. Currently porting the **welleswilder** family (12 indicators) in small batches by dependency order.

## Instructions

- Go and TS code is the reference implementation. Python is the reference for other modules.
- All ports must match reference output to tolerance specified by Go tests (typically 1e-8).
- Port one language at a time: **Python → Rust → Zig**. Python and Rust are done.
- Port indicators in small batches by author family, starting with simplest, factory last, cmd tools last.
- Zig naming: `snake_case` files, `camelCase` functions/fields, `PascalCase` types/enums, lowercase enum values.
- **Each indicator lives in its own subfolder** (e.g., `patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig`)
- **Author folder naming**: Use `snake_case` matching Rust (e.g., `patrick_mulloy`), NOT Go's `patrickmulloy` style.
- **Output enums are 1-based** (matching Go's `iota+1`).
- **fixSlices() pattern**: After `init()` returns by value, slices point into old stack frame. Must call `fixSlices()` to repoint them. `fixSlices` must be `pub`.
- **Metadata out-pointer pattern**: `getMetadata(out: *Metadata)` returns void.
- **No allocator needed** for EMA-like indicators (no window). SMA and Variance need allocator for ring buffer.
- **MaUnion pattern**: For indicators that support both SMA and EMA sub-indicators, use a tagged union with `update`, `isPrimed`, `deinit` methods.
- **Per-output metadata**: `OutputText` struct has only `mnemonic` and `description` fields (no suffix fields). Each output needs a fully-formed string built with `bufPrint` in `getMetadata`.
- **Metadata fields are plain `[]const u8`**, NOT functions.

## Discoveries

### Key Type Corrections (learned fixing MACD/PPO compilation)
- **Bar/Quote/Trade function types**: Use `bar_component.BarFunc`, `quote_component.QuoteFunc`, `trade_component.TradeFunc` (NOT `BarExtractFn` etc.)
- **Scalar.time is `i64`**, not `f64`
- **OutputArray**: Is a struct with `.values` array and `.len`, NOT a bare array. Use `OutputArray.append(.{ .scalar = .{ .time = t, .value = v } })` to add outputs, `out.slice()[i].scalar.value` to read them.
- **Metadata fields**: `outputs_len` (not `num_outputs`), `outputs_buf[i]` (not `outputs[i]`)
- **Band output type**: Use `band_mod.Band.new(time, lower, upper)` and `Band.empty(time)` for NaN bands. Access via `out.slice()[5].band`.

### HLCV Indicators Pattern (AdvanceDecline, AdvanceDeclineOscillator, Aroon)
- These indicators need HLCV data from bars, not just a single scalar component.
- They override `updateBar` to extract `sample.high, sample.low, sample.close, sample.volume` directly.
- For scalar/quote/trade updates, H=L=C so the range is 0.
- The `update()` method calls `updateHlcv(sample, sample, sample, 1)` or similar.

### StochasticRSI Dependency
- StochasticRelativeStrengthIndex depends on RelativeStrengthIndex (welleswilder family).
- RSI is NOT yet ported to Zig, so StochRSI must be deferred until after welleswilder family is done.

### Welleswilder Family Dependency Chain (12 indicators)
```
Level 0 (no deps):
  1. TrueRange — HLCV override, Update(close,high,low), no params, no allocator
  2. DirectionalMovementPlus — HLCV override, Update(high,low), length param, Wilder smoothing
  3. DirectionalMovementMinus — HLCV override, Update(high,low), length param, Wilder smoothing
  4. RelativeStrengthIndex — LineIndicator, component triple, Wilder-smoothed avg gain/loss
  5. ParabolicStopAndReverse — LineIndicator hybrid, overrides UpdateBar for UpdateHL(high,low)

Level 1:
  6. AverageTrueRange — depends on TR, needs allocator for initial SMA window

Level 2:
  7. NormalizedAverageTrueRange — depends on ATR
  8. DirectionalIndicatorPlus — depends on ATR + DM+
  9. DirectionalIndicatorMinus — depends on ATR + DM-

Level 3:
  10. DirectionalMovementIndex (DX) — depends on DI+, DI-

Level 4:
  11. AverageDirectionalMovementIndex (ADX) — depends on DX

Level 5:
  12. ADXR — depends on ADX, needs allocator for circular buffer
```

### Key welleswilder patterns discovered
- **TrueRange**: No params at all. Fixed mnemonic "tr". Update(close, high, low). First update stores close, returns NaN. Primed from second update. For scalar/quote/trade, v=v=v.
- **DM+/DM-**: Length param (min 1, default 14). Fixed mnemonics "+dm"/"-dm". Update(high, low). High/low swap if high<low. Two code paths: noSmoothing (length==1) vs Wilder smoothing. Accumulator-based. Not primed during first `length` updates.
- **All 12 return exactly 1 output** despite rich metadata.
- **10 of 12 are HLCV-override**, only RSI and SAR use LineIndicator with component triples.
- **Only ATR and ADXR need allocators** (ring buffers).

### DM+/DM- Algorithm Detail
- UpMove = today's high − yesterday's high
- DownMove = yesterday's low − today's low
- +DM: if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else 0
- -DM: if DownMove > UpMove and DownMove > 0, then -DM = DownMove, else 0
- Wilder smoothing (length>1): accumulator = accumulator - accumulator/n + today's DM(1)
- During warmup (count 1..length-1): just sum raw DM values into accumulator
- At count==length: apply first smoothing step, set primed

## Accomplished

### Fully Complete ✅ (662 tests pass before this session)
- **Python port**: All 63 indicators, factory, frequency_response, cmd tools. 1142 tests pass.
- **Rust port**: All 67 indicators, factory, frequency_response, cmd tools. 985 tests pass.
- **Zig Core Types**: All 17 core type files complete.
- **Zig Common Indicators**: All 13 complete (SMA, WMA, TMA, EMA, Momentum, ROC, ROCP, ROCR, Variance, StdDev, APO, LinearRegression, PearsonsCC).
- **Zig Single-Indicator Families**: All 9 complete (BalanceOfPower, OnBalanceVolume, CCI, TEMO, MFI, KAMA, JMA, Stochastic, ATCF).
- **Zig patrick_mulloy**: DEMA, TEMA complete.
- **Zig timtillson**: T2EMA, T3EMA complete.
- **Zig gerald_appel**: MACD, PPO complete.
- **Zig john_bollinger**: BollingerBands, BollingerBandsTrend complete.
- **Zig larry_williams**: WilliamsPercentR, UltimateOscillator complete.
- **Zig marc_chaikin**: AdvanceDecline, AdvanceDeclineOscillator complete.
- **Zig tushar_chande**: Aroon and ChandeMomentumOscillator complete (StochRSI blocked on RSI).

### Written This Session (NOT YET COMPILED/TESTED)
- **`true_range.zig`** — Written with full implementation and tests (252-entry test data, tolerance 1e-3). No params, no allocator.
- **`directional_movement_plus.zig`** — Written with full implementation and tests (252-entry test data for both length=1 and length=14, tolerance 1e-8). Length param, Wilder smoothing.

### Still Needs to Be Written
- **`directional_movement_minus.zig`** — Nearly identical to DM+ but with mirrored logic (deltaMinus > 0 && deltaPlus < deltaMinus). Test data already extracted from Go tests and ready. The Go source has been fully read (`directionalmovementminus.go` and test file).

### Not Yet Done
- **Register TR, DM+, DM- in `indicators.zig`** — Need to add imports and comptime references
- **Build and test** — None of the three new files have been compiled yet
- **Continue with RSI** (Batch 2, LineIndicator pattern, unblocks StochRSI)
- **Then ATR** (Batch 3, depends on TR)
- **Then NATR, DI+, DI-** (Batch 4)
- **Then DX → ADX → ADXR** (Batches 5-7)
- **Then ParabolicSAR** (Batch 8)

### Remaining Work After welleswilder
1. **tushar_chande/StochasticRelativeStrengthIndex** (unblocked after RSI)
2. **johnehlers** (23 indicators)
3. **custom** (GoertzelSpectrum, MaximumEntropySpectrum)
4. Factory, frequency_response, cmd tools

## Relevant files / directories

### Newly Created Files (this session, NOT yet compiled):
- `zig/src/indicators/welles_wilder/true_range/true_range.zig` — Complete implementation + tests
- `zig/src/indicators/welles_wilder/directional_movement_plus/directional_movement_plus.zig` — Complete implementation + tests

### Directory Created, File NOT Yet Written:
- `zig/src/indicators/welles_wilder/directional_movement_minus/` — Needs `directional_movement_minus.zig` (mirror of DM+ with swapped delta logic)

### Registration file (NEEDS welleswilder imports added):
- `zig/src/indicators/indicators.zig` — Root module; currently has 34 indicators registered, needs TR/DM+/DM- added

### Go Reference (fully read this session):
- `go/indicators/welleswilder/truerange/truerange.go` + `truerange_test.go`
- `go/indicators/welleswilder/directionalmovementplus/directionalmovementplus.go` + `directionalmovementplus_test.go`
- `go/indicators/welleswilder/directionalmovementminus/directionalmovementminus.go` + `directionalmovementminus_test.go`

### Go Reference (NOT yet read, upcoming batches):
- `go/indicators/welleswilder/relativestrengthindex/` — RSI (Batch 2)
- `go/indicators/welleswilder/averagetruerange/` — ATR (Batch 3)
- `go/indicators/welleswilder/normalizedaveragetruerange/` — NATR (Batch 4)
- `go/indicators/welleswilder/directionalindicatorplus/` — DI+ (Batch 4)
- `go/indicators/welleswilder/directionalindicatorminus/` — DI- (Batch 4)
- `go/indicators/welleswilder/directionalmovementindex/` — DX (Batch 5)
- `go/indicators/welleswilder/averagedirectionalmovementindex/` — ADX (Batch 6)
- `go/indicators/welleswilder/averagedirectionalmovementindexrating/` — ADXR (Batch 7)
- `go/indicators/welleswilder/parabolicstopandreverse/` — SAR (Batch 8)

### Key Pattern References (existing Zig files):
- `zig/src/indicators/larry_williams/williams_percent_r/williams_percent_r.zig` — HLCV override with allocator, ring buffers
- `zig/src/indicators/marc_chaikin/advance_decline/advance_decline.zig` — HLCV override, no allocator, LineIndicator wrapper

### DM- Test Data (from Go test file, for writing directional_movement_minus.zig):
The expected DMM1 (length=1) and DMM14 (length=14) arrays are in `go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go`. The input high/low arrays are identical to DM+. The key algorithm difference: DM- checks `deltaMinus > 0 && deltaPlus < deltaMinus` (vs DM+ which checks `deltaPlus > 0 && deltaPlus > deltaMinus`).

### Build:
- `zig/build.zig` — Single `indicators` module

### Skill files:
- `.opencode/skills/indicator-architecture/SKILL.md`
- `.opencode/skills/indicator-conversion/SKILL.md`

---

## User

---

## Assistant (Build · claude-opus-4.6 · 6.6s)

_Thinking:_

Let me continue with the next steps. Based on the summary, I need to:

1. Write `directional_movement_minus.zig`
2. Register TR, DM+, DM- in `indicators.zig`
3. Build and test

Let me start by reading the DM+ file (as template) and the Go DM- test data.

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Write directional_movement_minus.zig (mirror of DM+)",
      "status": "in_progress",
      "priority": "high"
    },
    {
      "content": "Register TR, DM+, DM- in indicators.zig",
      "status": "pending",
      "priority": "high"
    },
    {
      "content": "Build and test all three new indicators",
      "status": "pending",
      "priority": "high"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Write directional_movement_minus.zig (mirror of DM+)",
    "status": "in_progress",
    "priority": "high"
  },
  {
    "content": "Register TR, DM+, DM- in indicators.zig",
    "status": "pending",
    "priority": "high"
  },
  {
    "content": "Build and test all three new indicators",
    "status": "pending",
    "priority": "high"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 6.3s)

Let me read the DM+ file as template and the Go DM- test data.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/welles_wilder/directional_movement_plus/directional_movement_plus.zig"
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/welles_wilder/directional_movement_plus/directional_movement_plus.zig</path>
<type>file</type>
<content>1: const std = @import("std");
2: const math = std.math;
3: 
4: const Bar = @import("bar").Bar;
5: const Quote = @import("quote").Quote;
6: const Trade = @import("trade").Trade;
7: const Scalar = @import("scalar").Scalar;
8: 
9: const indicator_mod = @import("../../core/indicator.zig");
10: const build_metadata_mod = @import("../../core/build_metadata.zig");
11: const identifier_mod = @import("../../core/identifier.zig");
12: const metadata_mod = @import("../../core/metadata.zig");
13: 
14: const OutputArray = indicator_mod.OutputArray;
15: const Identifier = identifier_mod.Identifier;
16: const Metadata = metadata_mod.Metadata;
17: 
18: /// Enumerates the outputs of the Directional Movement Plus indicator.
19: pub const DirectionalMovementPlusOutput = enum(u8) {
20:     /// The scalar value of +DM.
21:     value = 1,
22: };
23: 
24: /// Parameters for the Directional Movement Plus indicator.
25: pub const DirectionalMovementPlusParams = struct {
26:     /// The smoothing length. Must be >= 1. Default is 14.
27:     length: usize = 14,
28: };
29: 
30: /// Welles Wilder's Directional Movement Plus (+DM) indicator.
31: ///
32: /// UpMove = today's high − yesterday's high
33: /// DownMove = yesterday's low − today's low
34: /// if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0
35: ///
36: /// When length > 1, Wilder's smoothing is applied:
37: ///   +DM(n) = previous +DM(n) − previous +DM(n)/n + today's +DM(1)
38: pub const DirectionalMovementPlus = struct {
39:     length: usize,
40:     no_smoothing: bool,
41:     count: usize,
42:     previous_high: f64,
43:     previous_low: f64,
44:     value: f64,
45:     accumulator: f64,
46:     primed: bool,
47: 
48:     pub const Error = error{InvalidLength};
49: 
50:     pub fn init(params: DirectionalMovementPlusParams) Error!DirectionalMovementPlus {
51:         if (params.length < 1) return error.InvalidLength;
52: 
53:         return .{
54:             .length = params.length,
55:             .no_smoothing = params.length == 1,
56:             .count = 0,
57:             .previous_high = 0,
58:             .previous_low = 0,
59:             .value = math.nan(f64),
60:             .accumulator = 0,
61:             .primed = false,
62:         };
63:     }
64: 
65:     pub fn deinit(_: *DirectionalMovementPlus) void {}
66:     pub fn fixSlices(_: *DirectionalMovementPlus) void {}
67: 
68:     /// Core update given high and low values.
69:     pub fn update(self: *DirectionalMovementPlus, high_in: f64, low_in: f64) f64 {
70:         if (math.isNan(high_in) or math.isNan(low_in)) {
71:             return math.nan(f64);
72:         }
73: 
74:         var high = high_in;
75:         var low = low_in;
76:         if (high < low) {
77:             const tmp = high;
78:             high = low;
79:             low = tmp;
80:         }
81: 
82:         if (self.no_smoothing) {
83:             if (self.primed) {
84:                 const delta_plus = high - self.previous_high;
85:                 const delta_minus = self.previous_low - low;
86:                 if (delta_plus > 0 and delta_plus > delta_minus) {
87:                     self.value = delta_plus;
88:                 } else {
89:                     self.value = 0;
90:                 }
91:             } else {
92:                 if (self.count > 0) {
93:                     const delta_plus = high - self.previous_high;
94:                     const delta_minus = self.previous_low - low;
95:                     if (delta_plus > 0 and delta_plus > delta_minus) {
96:                         self.value = delta_plus;
97:                     } else {
98:                         self.value = 0;
99:                     }
100:                     self.primed = true;
101:                 }
102:                 self.count += 1;
103:             }
104:         } else {
105:             const n: f64 = @floatFromInt(self.length);
106:             if (self.primed) {
107:                 const delta_plus = high - self.previous_high;
108:                 const delta_minus = self.previous_low - low;
109:                 if (delta_plus > 0 and delta_plus > delta_minus) {
110:                     self.accumulator += -self.accumulator / n + delta_plus;
111:                 } else {
112:                     self.accumulator += -self.accumulator / n;
113:                 }
114:                 self.value = self.accumulator;
115:             } else {
116:                 if (self.count > 0 and self.length >= self.count) {
117:                     const delta_plus = high - self.previous_high;
118:                     const delta_minus = self.previous_low - low;
119:                     if (self.length > self.count) {
120:                         if (delta_plus > 0 and delta_plus > delta_minus) {
121:                             self.accumulator += delta_plus;
122:                         }
123:                     } else {
124:                         if (delta_plus > 0 and delta_plus > delta_minus) {
125:                             self.accumulator += -self.accumulator / n + delta_plus;
126:                         } else {
127:                             self.accumulator += -self.accumulator / n;
128:                         }
129:                         self.value = self.accumulator;
130:                         self.primed = true;
131:                     }
132:                 }
133:                 self.count += 1;
134:             }
135:         }
136: 
137:         self.previous_low = low;
138:         self.previous_high = high;
139: 
140:         return self.value;
141:     }
142: 
143:     /// Update using a single sample value as substitute for high and low.
144:     pub fn updateSample(self: *DirectionalMovementPlus, sample: f64) f64 {
145:         return self.update(sample, sample);
146:     }
147: 
148:     pub fn isPrimed(self: *const DirectionalMovementPlus) bool {
149:         return self.primed;
150:     }
151: 
152:     pub fn getMetadata(_: *const DirectionalMovementPlus, out: *Metadata) void {
153:         const mnemonic = "+dm";
154:         const description = "Directional Movement Plus";
155:         build_metadata_mod.buildMetadata(out, Identifier.directional_movement_plus, mnemonic, description, &.{
156:             .{ .mnemonic = mnemonic, .description = description },
157:         });
158:     }
159: 
160:     fn makeOutput(self: *const DirectionalMovementPlus, time: i64) OutputArray {
161:         var out = OutputArray{};
162:         out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
163:         return out;
164:     }
165: 
166:     pub fn updateScalar(self: *DirectionalMovementPlus, sample: *const Scalar) OutputArray {
167:         _ = self.update(sample.value, sample.value);
168:         return self.makeOutput(sample.time);
169:     }
170: 
171:     pub fn updateBar(self: *DirectionalMovementPlus, sample: *const Bar) OutputArray {
172:         _ = self.update(sample.high, sample.low);
173:         return self.makeOutput(sample.time);
174:     }
175: 
176:     pub fn updateQuote(self: *DirectionalMovementPlus, sample: *const Quote) OutputArray {
177:         const mid = (sample.bid_price + sample.ask_price) / 2.0;
178:         _ = self.update(mid, mid);
179:         return self.makeOutput(sample.time);
180:     }
181: 
182:     pub fn updateTrade(self: *DirectionalMovementPlus, sample: *const Trade) OutputArray {
183:         _ = self.update(sample.price, sample.price);
184:         return self.makeOutput(sample.time);
185:     }
186: 
187:     // --- Indicator interface ---
188: 
189:     pub fn indicator(self: *DirectionalMovementPlus) indicator_mod.Indicator {
190:         return indicator_mod.Indicator{
191:             .ptr = @ptrCast(self),
192:             .vtable = &.{
193:                 .isPrimed = vtableIsPrimed,
194:                 .metadata = vtableMetadata,
195:                 .updateScalar = vtableUpdateScalar,
196:                 .updateBar = vtableUpdateBar,
197:                 .updateQuote = vtableUpdateQuote,
198:                 .updateTrade = vtableUpdateTrade,
199:             },
200:         };
201:     }
202: 
203:     fn vtableIsPrimed(ptr: *const anyopaque) bool {
204:         const self: *const DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
205:         return self.isPrimed();
206:     }
207: 
208:     fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
209:         const self: *const DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
210:         self.getMetadata(out);
211:     }
212: 
213:     fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
214:         const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
215:         return self.updateScalar(sample);
216:     }
217: 
218:     fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
219:         const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
220:         return self.updateBar(sample);
221:     }
222: 
223:     fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
224:         const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
225:         return self.updateQuote(sample);
226:     }
227: 
228:     fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
229:         const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
230:         return self.updateTrade(sample);
231:     }
232: };
233: 
234: // ============================================================================
235: // Tests
236: // ============================================================================
237: 
238: const testing = std.testing;
239: 
240: fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
241:     if (math.isNan(a) and math.isNan(b)) return true;
242:     if (math.isNan(a) or math.isNan(b)) return false;
243:     return @abs(a - b) <= tolerance;
244: }
245: 
246: const test_high = [_]f64{
247:     93.250000,  94.940000,  96.375000,  96.190000,  96.000000,  94.720000,  95.000000,  93.720000,  92.470000,  92.750000,  96.250000,
248:     99.625000,  99.125000,  92.750000,  91.315000,  93.250000,  93.405000,  90.655000,  91.970000,  92.250000,  90.345000,  88.500000,
249:     88.250000,  85.500000,  84.440000,  84.750000,  84.440000,  89.405000,  88.125000,  89.125000,  87.155000,  87.250000,  87.375000,
250:     88.970000,  90.000000,  89.845000,  86.970000,  85.940000,  84.750000,  85.470000,  84.470000,  88.500000,  89.470000,  90.000000,
251:     92.440000,  91.440000,  92.970000,  91.720000,  91.155000,  91.750000,  90.000000,  88.875000,  89.000000,  85.250000,  83.815000,
252:     85.250000,  86.625000,  87.940000,  89.375000,  90.625000,  90.750000,  88.845000,  91.970000,  93.375000,  93.815000,  94.030000,
253:     94.030000,  91.815000,  92.000000,  91.940000,  89.750000,  88.750000,  86.155000,  84.875000,  85.940000,  99.375000,  103.280000,
254:     105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000, 110.940000,
255:     110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000, 115.345000, 113.000000,
256:     118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000, 119.870000, 120.870000, 116.750000,
257:     116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000, 127.750000, 124.190000, 124.440000, 125.750000,
258:     124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000, 133.500000, 135.500000, 137.440000, 138.690000, 139.190000,
259:     138.500000, 138.130000, 137.500000, 138.880000, 132.130000, 129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000,
260:     126.620000, 126.690000, 126.000000, 123.120000, 121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000,
261:     124.500000, 127.870000, 128.560000, 129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000,
262:     125.620000, 126.000000, 128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000,
263:     135.000000, 132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,
264:     122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000, 107.250000,
265:     108.870000, 109.000000, 108.500000, 113.060000, 93.000000,  94.620000,  95.120000,  96.000000,  95.560000,  95.310000,  99.000000,
266:     98.810000,  96.810000,  95.940000,  94.440000,  92.940000,  93.940000,  95.500000,  97.060000,  97.500000,  96.250000,  96.370000,
267:     95.000000,  94.870000,  98.250000,  105.120000, 108.440000, 109.870000, 105.000000, 106.000000, 104.940000, 104.500000, 104.440000,
268:     106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000, 112.750000, 110.190000, 107.940000, 109.690000,
269:     111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000, 110.750000, 110.500000, 110.500000, 109.500000,
270: };
271: 
272: const test_low = [_]f64{
273:     90.750000,  91.405000,  94.250000,  93.500000,  92.815000,  93.500000,  92.000000,  89.750000,  89.440000,  90.625000,  92.750000,
274:     96.315000,  96.030000,  88.815000,  86.750000,  90.940000,  88.905000,  88.780000,  89.250000,  89.750000,  87.500000,  86.530000,
275:     84.625000,  82.280000,  81.565000,  80.875000,  81.250000,  84.065000,  85.595000,  85.970000,  84.405000,  85.095000,  85.500000,
276:     85.530000,  87.875000,  86.565000,  84.655000,  83.250000,  82.565000,  83.440000,  82.530000,  85.065000,  86.875000,  88.530000,
277:     89.280000,  90.125000,  90.750000,  89.000000,  88.565000,  90.095000,  89.000000,  86.470000,  84.000000,  83.315000,  82.000000,
278:     83.250000,  84.750000,  85.280000,  87.190000,  88.440000,  88.250000,  87.345000,  89.280000,  91.095000,  89.530000,  91.155000,
279:     92.000000,  90.530000,  89.970000,  88.815000,  86.750000,  85.065000,  82.030000,  81.500000,  82.565000,  96.345000,  96.470000,
280:     101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000, 108.530000,
281:     108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000, 110.875000, 110.500000,
282:     110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000, 116.060000, 116.560000, 113.310000,
283:     112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000, 123.000000, 121.750000, 121.560000, 123.120000,
284:     122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000, 130.630000, 132.130000, 133.880000, 135.380000, 135.750000,
285:     136.190000, 134.500000, 135.380000, 133.690000, 126.060000, 126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000,
286:     124.620000, 124.370000, 121.810000, 118.190000, 118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000,
287:     122.000000, 124.500000, 126.560000, 123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000,
288:     122.690000, 122.870000, 125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000,
289:     131.940000, 129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,
290:     119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000, 104.690000,
291:     103.870000, 104.690000, 105.440000, 107.000000, 89.000000,  92.500000,  92.120000,  94.620000,  92.810000,  94.250000,  96.250000,
292:     96.370000,  93.690000,  93.500000,  90.000000,  90.190000,  90.500000,  92.120000,  94.120000,  94.870000,  93.000000,  93.870000,
293:     93.000000,  92.620000,  93.560000,  98.370000,  104.440000, 106.000000, 101.810000, 104.120000, 103.370000, 102.120000, 102.250000,
294:     103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000, 106.560000, 106.870000, 104.500000, 105.750000,
295:     108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000, 109.060000, 108.750000, 108.560000, 106.620000,
296: };
297: 
298: const test_expected_dmp1 = [_]f64{
299:     math.nan(f64),
300:     1.6900,
301:     1.4350,
302:     0.0000,
303:     0.0000,
304:     0.0000,
305:     0.0000,
306:     0.0000,
307:     0.0000,
308:     0.2800,
309:     3.5000,
310:     3.3750,
311:     0.0000,
312:     0.0000,
313:     0.0000,
314:     1.9350,
315:     0.0000,
316:     0.0000,
317:     1.3150,
318:     0.2800,
319:     0.0000,
320:     0.0000,
321:     0.0000,
322:     0.0000,
323:     0.0000,
324:     0.0000,
325:     0.0000,
326:     4.9650,
327:     0.0000,
328:     1.0000,
329:     0.0000,
330:     0.0950,
331:     0.1250,
332:     1.5950,
333:     1.0300,
334:     0.0000,
335:     0.0000,
336:     0.0000,
337:     0.0000,
338:     0.7200,
339:     0.0000,
340:     4.0300,
341:     0.9700,
342:     0.5300,
343:     2.4400,
344:     0.0000,
345:     1.5300,
346:     0.0000,
347:     0.0000,
348:     0.5950,
349:     0.0000,
350:     0.0000,
351:     0.0000,
352:     0.0000,
353:     0.0000,
354:     1.4350,
355:     1.3750,
356:     1.3150,
357:     1.4350,
358:     1.2500,
359:     0.0000,
360:     0.0000,
361:     3.1250,
362:     1.4050,
363:     0.0000,
364:     0.2150,
365:     0.0000,
366:     0.0000,
367:     0.0000,
368:     0.0000,
369:     0.0000,
370:     0.0000,
371:     0.0000,
372:     0.0000,
373:     1.0650,
374:     13.4350,
375:     3.9050,
376:     2.0950,
377:     2.2500,
378:     0.0000,
379:     0.0000,
380:     1.0000,
381:     0.6250,
382:     1.8150,
383:     0.0000,
384:     0.7500,
385:     1.7500,
386:     2.1900,
387:     0.0000,
388:     3.2800,
389:     8.7800,
390:     0.0000,
391:     0.0000,
392:     0.5000,
393:     0.0000,
394:     0.0000,
395:     0.0000,
396:     0.0000,
397:     0.0000,
398:     5.3150,
399:     0.0000,
400:     0.0000,
401:     0.0000,
402:     0.0000,
403:     0.6900,
404:     0.6900,
405:     5.6900,
406:     0.0000,
407:     1.0000,
408:     0.0000,
409:     0.0000,
410:     0.0000,
411:     2.3100,
412:     3.1900,
413:     0.5000,
414:     0.0000,
415:     4.3100,
416:     2.0000,
417:     0.0000,
418:     0.2500,
419:     1.3100,
420:     0.0000,
421:     0.6200,
422:     6.6900,
423:     0.0000,
424:     0.9400,
425:     1.6300,
426:     0.0000,
427:     2.0000,
428:     1.9400,
429:     1.2500,
430:     0.5000,
431:     0.0000,
432:     0.0000,
433:     0.0000,
434:     0.0000,
435:     0.0000,
436:     0.0000,
437:     0.0000,
438:     0.0000,
439:     0.0000,
440:     1.3800,
441:     2.1900,
442:     0.0000,
443:     0.0000,
444:     0.0000,
445:     0.0000,
446:     0.0000,
447:     2.1300,
448:     3.0000,
449:     0.0000,
450:     0.0000,
451:     1.2500,
452:     0.0600,
453:     0.6900,
454:     3.3700,
455:     0.6900,
456:     0.0000,
457:     0.0000,
458:     0.0000,
459:     0.5000,
460:     0.0000,
461:     0.4400,
462:     1.8100,
463:     0.0000,
464:     0.4300,
465:     0.3800,
466:     2.5000,
467:     0.0000,
468:     3.0000,
469:     2.9400,
470:     1.2500,
471:     2.5600,
472:     1.1900,
473:     0.0000,
474:     0.0000,
475:     1.4400,
476:     0.0000,
477:     0.0000,
478:     0.0000,
479:     0.0000,
480:     0.0000,
481:     0.0000,
482:     0.0000,
483:     1.8100,
484:     0.0000,
485:     0.0000,
486:     0.0000,
487:     0.0000,
488:     0.0000,
489:     3.2500,
490:     0.0000,
491:     0.0000,
492:     0.0000,
493:     1.6300,
494:     0.0000,
495:     0.0000,
496:     0.0000,
497:     1.6200,
498:     0.1300,
499:     0.0000,
500:     4.5600,
501:     0.0000,
502:     1.6200,
503:     0.5000,
504:     0.8800,
505:     0.0000,
506:     0.0000,
507:     3.6900,
508:     0.0000,
509:     0.0000,
510:     0.0000,
511:     0.0000,
512:     0.0000,
513:     1.0000,
514:     1.5600,
515:     1.5600,
516:     0.4400,
517:     0.0000,
518:     0.1200,
519:     0.0000,
520:     0.0000,
521:     3.3800,
522:     6.8700,
523:     3.3200,
524:     1.4300,
525:     0.0000,
526:     1.0000,
527:     0.0000,
528:     0.0000,
529:     0.0000,
530:     1.8700,
531:     6.5600,
532:     3.6300,
533:     2.6900,
534:     1.8100,
535:     0.0000,
536:     0.0000,
537:     0.0000,
538:     0.0000,
539:     0.0000,
540:     1.7500,
541:     1.3700,
542:     0.0000,
543:     0.0000,
544:     0.1900,
545:     0.1300,
546:     0.0000,
547:     0.7500,
548:     0.0000,
549:     0.0000,
550:     0.0000,
551: };
552: 
553: const test_expected_dmp14 = [_]f64{
554:     math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
555:     math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      9.545714285714290,  10.798877551020400, 10.027529154519000, 9.311277072053310,  9.961185852620930,  9.529672577433720,
556:     8.848981679045600,  8.216911559113770,  7.629989304891360,  7.084990068827690,  6.578919349625710,  6.108996538938160,  5.672639643299720,  10.232451097349700, 9.501561733253330,  9.822878752306670,
557:     9.121244555713330,  8.564727087448090,  8.077960866916090,  9.095963662136360,  9.476251971983770,  8.799376831127780,  8.170849914618660,  7.587217777860180,  7.045273650870170,  7.262039818665150,
558:     6.743322688760500,  10.291656782420500, 10.526538440819000, 10.304642837903400, 12.008596920910300, 11.150839997988100, 11.884351426703200, 11.035469181938700, 10.247221383228800, 10.110276998712500,
559:     9.388114355947290,  8.717534759093910,  8.094853704872920,  7.516649868810570,  6.979746306752670,  7.916192999127480,  8.725750642046950,  9.417482739043590,  10.179805400540500, 10.702676443359000,
560:     9.938199554547660,  9.228328157794250,  11.694161860808900, 12.263864585036900, 11.387874257534200, 10.789454667710400, 10.018779334302500, 9.303152238995170,  8.638641364781230,  8.021595553011140,
561:     7.448624442081780,  6.916579839075940,  6.422538421999080,  5.963785677570580,  6.602800986315530,  19.566172344435900, 22.073588605547600, 22.591903705151300, 23.228196297640500, 21.569039419237600,
562:     20.028393746434900, 19.597794193118200, 18.822951750752600, 19.293455197127400, 17.915351254475400, 17.385683307727200, 17.893848785746700, 18.805716729621900, 17.462451248934600, 19.495133302582200,
563:     26.882623780969100, 24.962436368042800, 23.179405198896900, 22.023733398975700, 20.450609584763100, 18.989851757280000, 17.633433774617200, 16.373902790716000, 15.204338305664800, 19.433314140974500,
564:     18.045220273762000, 16.756275968493300, 15.559399113600900, 14.448013462629400, 14.106012501013000, 13.788440179512100, 18.493551595261200, 17.172583624171100, 16.945970508158900, 15.735544043290400,
565:     14.611576611626800, 13.567892567939200, 14.908757384515000, 17.033846142763900, 16.317142846852200, 15.151632643505600, 18.379373168969500, 19.066560799757400, 17.704663599774700, 16.690044771219400,
566:     16.807898716132300, 15.607334522122800, 15.112524913399800, 20.723058848156900, 19.242840359002900, 18.808351761931200, 19.094898064650400, 17.730976774318300, 18.464478433295500, 19.085587116631600,
567:     18.972330894015000, 18.117164401585400, 16.823081230043600, 15.621432570754700, 14.505615958558000, 13.469500532946700, 12.507393352021900, 11.614008112591800, 10.784436104549500, 10.014119239938800,
568:     9.298825008514630,  10.014623222192200, 11.489292992035600, 10.668629206890200, 9.906584263540870,  9.198971101859380,  8.541901737440850,  7.931765899052220,  9.495211191977060,  11.816981821121600,
569:     10.972911691041400, 10.189132284538500, 10.711337121357200, 10.006241612688800, 9.981510068925310,  12.638545064002100, 12.425791845144800, 11.538235284777300, 10.714075621578900, 9.948784505751860,
570:     9.738157041055300,  9.042574395265630,  8.836676224175230,  10.015485065305600, 9.300093274926600,  9.065800898146130,  8.798243691135690,  10.669797713197400, 9.907669305111900,  12.199978640461000,
571:     14.268551594713800, 14.499369337948600, 16.023700099523700, 16.069150092414800, 14.921353657242300, 13.855542681725000, 14.305861061601800, 13.284013842916000, 12.335155711279100, 11.454073160473500,
572:     10.635925077582500, 9.876216143469460,  9.170772133221640,  8.515716980848670,  9.717451482216620,  9.023347804915440,  8.378822961707190,  7.780335607299540,  7.224597349635280,  6.708554681804190,
573:     9.479372204532460,  8.802274189923000,  8.173540319214220,  7.589716010698920,  8.677593438506130,  8.057765335755700,  7.482210668916000,  6.947767049707720,  8.071497974728600,  7.624962405105120,
574:     7.080322233311900,  11.134584930932500, 10.339257435865900, 11.220739047589700, 10.919257687047600, 11.019310709401400, 10.232217087301300, 9.501344438208310,  12.512676978336300, 11.618914337026600,
575:     10.788991884381800, 10.018349606926000, 9.302753206431250,  8.638270834543300,  9.021251489218780,  9.936876382846010,  10.787099498357000, 10.456592391331500, 9.709692934807830,  9.136143439464420,
576:     8.483561765216960,  7.877593067701460,  10.694907848579900, 16.800985859395600, 18.920915440867400, 18.999421480805400, 17.642319946462200, 17.382154236000600, 16.140571790572000, 14.987673805531100,
577:     13.917125676564600, 14.793045271095700, 20.296399180303200, 22.476656381710100, 23.561180925873600, 23.688239431168400, 21.996222328942100, 20.425063591160500, 18.966130477506200, 17.611406871970000,
578:     16.353449238257900, 16.935345721239500, 17.095678169722400, 15.874558300456500, 14.740661278995300, 13.877756901924200, 13.016488551786800, 12.086739369516300, 11.973400843122300, 11.118157925756400,
579:     10.324003788202400, 9.586574946187900,
580: };
581: 
582: test "DirectionalMovementPlus update length=14" {
583:     const tolerance = 1e-8;
584:     var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });
585: 
586:     for (0..test_high.len) |i| {
587:         const act = dmp.update(test_high[i], test_low[i]);
588:         const exp = test_expected_dmp14[i];
589: 
590:         if (math.isNan(exp)) {
591:             try testing.expect(math.isNan(act));
592:         } else {
593:             try testing.expect(!math.isNan(act));
594:             try testing.expect(almostEqual(act, exp, tolerance));
595:         }
596:     }
597: }
598: 
599: test "DirectionalMovementPlus update length=1" {
600:     const tolerance = 1e-8;
601:     var dmp = try DirectionalMovementPlus.init(.{ .length = 1 });
602: 
603:     for (0..test_high.len) |i| {
604:         const act = dmp.update(test_high[i], test_low[i]);
605:         const exp = test_expected_dmp1[i];
606: 
607:         if (math.isNan(exp)) {
608:             try testing.expect(math.isNan(act));
609:         } else {
610:             try testing.expect(!math.isNan(act));
611:             try testing.expect(almostEqual(act, exp, tolerance));
612:         }
613:     }
614: }
615: 
616: test "DirectionalMovementPlus constructor validation" {
617:     const result = DirectionalMovementPlus.init(.{ .length = 0 });
618:     try testing.expectError(error.InvalidLength, result);
619: }
620: 
621: test "DirectionalMovementPlus isPrimed length=1" {
622:     var dmp = try DirectionalMovementPlus.init(.{ .length = 1 });
623: 
624:     try testing.expect(!dmp.isPrimed());
625: 
626:     _ = dmp.update(test_high[0], test_low[0]);
627:     try testing.expect(!dmp.isPrimed());
628: 
629:     _ = dmp.update(test_high[1], test_low[1]);
630:     try testing.expect(dmp.isPrimed());
631: }
632: 
633: test "DirectionalMovementPlus isPrimed length=14" {
634:     var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });
635: 
636:     for (0..14) |i| {
637:         _ = dmp.update(test_high[i], test_low[i]);
638:         try testing.expect(!dmp.isPrimed());
639:     }
640: 
641:     _ = dmp.update(test_high[14], test_low[14]);
642:     try testing.expect(dmp.isPrimed());
643: }
644: 
645: test "DirectionalMovementPlus NaN passthrough" {
646:     var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });
647: 
648:     try testing.expect(math.isNan(dmp.update(math.nan(f64), 1)));
649:     try testing.expect(math.isNan(dmp.update(1, math.nan(f64))));
650:     try testing.expect(math.isNan(dmp.update(math.nan(f64), math.nan(f64))));
651:     try testing.expect(math.isNan(dmp.updateSample(math.nan(f64))));
652: }
653: 
654: test "DirectionalMovementPlus high/low swap" {
655:     var dmp1 = try DirectionalMovementPlus.init(.{ .length = 1 });
656:     var dmp2 = try DirectionalMovementPlus.init(.{ .length = 1 });
657: 
658:     _ = dmp1.update(10, 5);
659:     _ = dmp2.update(5, 10);
660: 
661:     const v1 = dmp1.update(12, 6);
662:     const v2 = dmp2.update(6, 12);
663: 
664:     try testing.expectEqual(v1, v2);
665: }
666: 
667: test "DirectionalMovementPlus metadata" {
668:     var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });
669:     var meta: Metadata = undefined;
670:     dmp.getMetadata(&meta);
671: 
672:     try testing.expectEqual(Identifier.directional_movement_plus, meta.identifier);
673:     try testing.expectEqualStrings("+dm", meta.mnemonic);
674:     try testing.expectEqualStrings("Directional Movement Plus", meta.description);
675:     try testing.expectEqual(@as(usize, 1), meta.outputs_len);
676: }
677: 
678: test "DirectionalMovementPlus updateBar" {
679:     var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });
680: 
681:     for (0..14) |i| {
682:         _ = dmp.update(test_high[i], test_low[i]);
683:     }
684: 
685:     const bar = Bar{ .time = 42, .open = 0, .high = test_high[14], .low = test_low[14], .close = 0, .volume = 0 };
686:     const out = dmp.updateBar(&bar);
687:     try testing.expectEqual(@as(usize, 1), out.len);
688: }

(End of file - total 688 lines)
</content>
```

**Tool: bash**

**Input:**
```json
{
  "command": "cd go && grep -A 500 'expectedDMM1' indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go | head -600",
  "description": "Get DM- test data from Go"
}
```

---

## Assistant (Build · claude-opus-4.6 · 6.2s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd go && grep -A 300 'expectedDMM1' indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go | head -600",
  "description": "Get DM- test expected data"
}
```

---

## Assistant (Build · claude-opus-4.6 · 5.5s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd go && grep -c 'DMM' indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go; head -20 indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go",
  "description": "Check DM- test file structure"
}
```

**Output:**
```
4
//nolint:testpackage
package directionalmovementminus

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementMinusTest.cs.
func testInputHigh() []float64 {
	return []float64{
		93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
		99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000, 90.345000, 88.500000,
		88.250000, 85.500000, 84.440000, 84.750000, 84.440000, 89.405000, 88.125000, 89.125000, 87.155000, 87.250000, 87.375000,

```

---

## Assistant (Build · claude-opus-4.6 · 5.0s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd go && grep -n 'expected\\|func test' indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go | head -20",
  "description": "Find test data variable names"
}
```

**Output:**
```
16:func testInputHigh() []float64 {
46:func testInputLow() []float64 {
77:func testExpectedDMM1() []float64 {
110:func testExpectedDMM14() []float64 {
157:func testTime() time.Time {
166:		t.Fatalf("unexpected error: %v", err)
170:		t.Errorf("expected length 14, got %d", dmm.Length())
179:		t.Error("expected error for length 0")
184:		t.Error("expected error for negative length")
243:	expected := testExpectedDMM14()
249:		if math.IsNaN(expected[i]) {
251:				t.Errorf("[%d] expected NaN, got %v", i, act)
258:			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
262:		if math.Abs(act-expected[i]) > tolerance {
263:			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
275:	expected := testExpectedDMM1()
281:		if math.IsNaN(expected[i]) {
283:				t.Errorf("[%d] expected NaN, got %v", i, act)
290:			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
294:		if math.Abs(act-expected[i]) > tolerance {

```

---

## Assistant (Build · claude-opus-4.6 · 5.3s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go",
  "offset": 77,
  "limit": 80
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/go/indicators/welleswilder/directionalmovementminus/directionalmovementminus_test.go</path>
<type>file</type>
<content>77: func testExpectedDMM1() []float64 {
78: 	return []float64{
79: 		math.NaN(),
80: 		0.0000, 0.0000, 0.7500, 0.6850, 0.0000, 1.5000, 2.2500, 0.3100, 0.0000, 0.0000,
81: 		0.0000, 0.2850, 7.2150, 2.0650, 0.0000, 2.0350, 0.1250, 0.0000, 0.0000, 2.2500,
82: 		0.9700, 1.9050, 2.3450, 0.7150, 0.6900, 0.0000, 0.0000, 0.0000, 0.0000, 1.5650,
83: 		0.0000, 0.0000, 0.0000, 0.0000, 1.3100, 1.9100, 1.4050, 0.6850, 0.0000, 0.9100,
84: 		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.7500, 0.4350, 0.0000, 1.0950,
85: 		2.5300, 2.4700, 0.6850, 1.3150, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1900,
86: 		0.9050, 0.0000, 0.0000, 1.5650, 0.0000, 0.0000, 1.4700, 0.5600, 1.1550, 2.0650,
87: 		1.6850, 3.0350, 0.5300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 2.5000, 0.0300,
88: 		0.0000, 0.0000, 0.0000, 2.0350, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
89: 		0.0000, 2.0000, 0.0000, 1.9700, 0.2800, 1.6550, 3.7200, 0.3750, 0.0000, 0.0000,
90: 		0.0000, 3.0000, 1.7500, 0.0000, 0.0000, 0.0000, 1.4400, 0.0000, 3.2500, 0.7500,
91: 		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,
92: 		0.9300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1800, 0.0000, 0.0000, 0.0000,
93: 		0.0000, 0.0000, 1.6900, 0.0000, 1.6900, 7.6300, 0.0000, 3.3700, 0.8800, 0.0000,
94: 		0.0000, 0.0000, 1.1900, 0.2500, 2.5600, 3.6200, 0.1300, 0.0000, 0.0000, 0.0000,
95: 		2.1800, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.0600, 2.2500, 0.1900, 0.0000,
96: 		1.3100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,
97: 		0.0000, 0.0000, 0.0000, 2.3800, 0.0600, 0.0000, 2.3800, 5.8100, 0.0000, 0.0000,
98: 		1.8800, 2.9300, 1.0000, 0.0000, 0.0000, 1.6900, 0.3800, 4.1200, 0.0000, 0.0000,
99: 		0.0000, 4.0000, 4.0000, 0.0000, 3.1200, 5.0000, 0.3100, 0.0000, 0.0000, 0.0000,
100: 		0.0000, 18.0000, 0.0000, 0.0000, 0.0000, 1.8100, 0.0000, 0.0000, 0.0000, 2.6800,
101: 		0.1900, 3.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.8700, 0.0000, 0.8700,
102: 		0.3800, 0.0000, 0.0000, 0.0000, 0.0000, 4.1900, 0.0000, 0.7500, 1.2500, 0.0000,
103: 		0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.2500, 4.6900, 1.0000, 0.0000, 2.3700,
104: 		0.0000, 0.0000, 0.8700, 0.0000, 0.0000, 0.0000, 0.0700, 0.0000, 0.3100, 0.1900,
105: 		1.9400,
106: 	}
107: }
108: 
109: // Expected -DM14 (length=14), 252 entries. Extracted from DirectionalMovementMinusTest.cs talibDmMinus14.
110: func testExpectedDMM14() []float64 {
111: 	return []float64{
112: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
113: 		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
114: 		14.131785714285700, 13.122372448979600, 14.220060131195300, 13.329341550395700, 12.377245725367400,
115: 		11.493156744984000, 12.922216977485200, 12.969201479093400, 13.947829944872400, 15.296556377381500, 14.918945207568600,
116: 		14.543306264170800, 13.504498673872900, 12.539891625739100, 11.644185081043500, 10.812457575254600, 11.605139177022200,
117: 		10.776200664377700, 10.006472045493600, 9.291724042244060, 8.628029467798060, 9.321741648669630, 10.565902959478900,
118: 		11.216195605230400, 11.100038776285400, 10.307178863693600, 10.480951802001200, 9.732312387572540, 9.037147217031640,
119: 		8.391636701529380, 7.792234079991570, 7.235645931420740, 6.718814079176400, 7.988898787806660, 7.853263160106190,
120: 		7.292315791527170, 7.866436092132370, 9.834547799837210, 11.602080099848800, 11.458360092716800, 11.954905800379900,
121: 		11.100983957495600, 10.308056531960200, 9.571766779677320, 8.888069152557510, 8.253207070231970, 7.853692279501110,
122: 		8.197714259536750, 7.612163240998410, 7.068437295212810, 8.128548916983320, 7.547938280055940, 7.008799831480520,
123: 		7.978171272089050, 7.968301895511260, 8.554137474403320, 10.008127654803100, 10.978261393745700, 13.229099865621000,
124: 		12.814164160933800, 11.898866720867100, 11.048947669376600, 10.259737121564000, 9.526898755737980, 8.846405987470980,
125: 		10.714519845508800, 9.979196999401000, 9.266397213729500, 8.604511698463110, 7.989903720001460, 9.454196311429920,
126: 		8.778896574899210, 8.151832533834980, 7.569558781418200, 7.028876011316900, 6.526813439079980, 6.060612479145690,
127: 		5.627711587778140, 7.225732188651130, 6.709608460890340, 8.200350713683890, 7.894611376992180, 8.985710564349880,
128: 		12.063874095467700, 11.577168802934300, 10.750228174153300, 9.982354733142360, 9.269329395060770, 11.607234438270700,
129: 		12.528146264108500, 11.633278673815100, 10.802330197114000, 10.030735183034400, 10.754254098531900, 9.986093091493950,
130: 		12.522800727815800, 12.378314961543300, 11.494149607147300, 10.673138920922500, 9.910771855142320, 9.202859579775010,
131: 		8.545512466933940, 7.935118719295800, 7.368324525060390, 8.092015630413220, 7.514014513955130, 6.977299191529760,
132: 		7.408920677849070, 6.879712058002710, 6.388304053859660, 5.931996621441110, 5.508282577052460, 5.114833821548720,
133: 		4.929488548580960, 4.577382223682320, 4.250426350562150, 3.946824468379140, 3.664908434923490, 3.403129261000380,
134: 		4.850048599500350, 4.503616556678900, 5.871929659773260, 13.082506112646600, 12.148041390314700, 14.650324148149400,
135: 		14.483872423281500, 13.449310107332900, 12.488645099666200, 11.596599021118600, 11.958270519610200, 11.354108339638000,
136: 		13.103100601092400, 15.787164843871600, 14.789510212166400, 13.733116625583100, 12.752179723755800, 11.841309743487500,
137: 		13.175501904667000, 12.234394625762200, 11.360509295350600, 10.549044345682700, 9.795541178133930, 9.095859665410080,
138: 		11.506155403595100, 12.934287160481100, 12.200409506161100, 11.328951684292400, 11.829740849700100, 10.984759360435800,
139: 		10.200133691833200, 9.471552713845160, 8.795013234284790, 8.166798003264450, 7.583455288745560, 8.291779910978020,
140: 		7.699509917336730, 7.149544923241250, 6.638863143009730, 6.164658632794750, 5.724325873309410, 7.695445453787310,
141: 		7.205770778516790, 6.691072865765590, 8.593139089639470, 13.789343440379500, 12.804390337495300, 11.889791027674200,
142: 		12.920520239983200, 14.927625937127200, 14.861366941618100, 13.799840731502600, 12.814137822109500, 13.588842263387400,
143: 		12.998210673145400, 16.189767053635100, 15.033355121232600, 13.959544041144500, 12.962433752491300, 16.036545627313400,
144: 		18.891078082505300, 17.541715362326300, 19.408735693588700, 23.022397429761000, 21.687940470492300, 20.138801865457200,
145: 		18.700316017924500, 17.364579159501300, 16.124252076679800, 32.972519785488400, 30.617339800810700, 28.430386957895600,
146: 		26.399645032331600, 26.323956101450800, 24.443673522775800, 22.697696842577500, 21.076432782393400, 22.250973297936700,
147: 		20.851618062369800, 22.862216772200500, 21.229201288471900, 19.712829767866800, 18.304770498733400, 16.997286891681000,
148: 		15.783194970846700, 16.525823901500500, 15.345407908536200, 15.119307343640700, 14.419356819095000, 13.389402760588200,
149: 		12.433016849117600, 11.544944217037800, 10.720305344392200, 14.144569248364200, 13.134242873481000, 12.946082668232400,
150: 		13.271362477644400, 12.323408014955500, 11.443164585315800, 10.625795686364700, 9.866810280195770, 9.162038117324650,
151: 		8.507606823230030, 11.149920621570700, 15.043497720030000, 14.968962168599300, 13.899750585127900, 15.276911257618800,
152: 		14.185703310646000, 13.172438788457000, 13.101550303567200, 12.165725281883800, 11.296744904606400, 10.489834554277400,
153: 		9.810560657543280, 9.109806324861620, 8.769105873085790, 8.332741167865380, 9.677545370160700,
154: 	}
155: }
156: 

(Showing lines 77-156 of 452. Use offset=157 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 194.8s)

Now I have all the test data. Let me create the DM- file by mirroring DM+ with the swapped logic.

**Tool: write**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/welles_wilder/directional_movement_minus/directional_movement_minus.zig",
  "content": "const std = @import(\"std\");\nconst math = std.math;\n\nconst Bar = @import(\"bar\").Bar;\nconst Quote = @import(\"quote\").Quote;\nconst Trade = @import(\"trade\").Trade;\nconst Scalar = @import(\"scalar\").Scalar;\n\nconst indicator_mod = @import(\"../../core/indicator.zig\");\nconst build_metadata_mod = @import(\"../../core/build_metadata.zig\");\nconst identifier_mod = @import(\"../../core/identifier.zig\");\nconst metadata_mod = @import(\"../../core/metadata.zig\");\n\nconst OutputArray = indicator_mod.OutputArray;\nconst Identifier = identifier_mod.Identifier;\nconst Metadata = metadata_mod.Metadata;\n\n/// Enumerates the outputs of the Directional Movement Minus indicator.\npub const DirectionalMovementMinusOutput = enum(u8) {\n    /// The scalar value of -DM.\n    value = 1,\n};\n\n/// Parameters for the Directional Movement Minus indicator.\npub const DirectionalMovementMinusParams = struct {\n    /// The smoothing length. Must be >= 1. Default is 14.\n    length: usize = 14,\n};\n\n/// Welles Wilder's Directional Movement Minus (-DM) indicator.\n///\n/// UpMove = today's high − yesterday's high\n/// DownMove = yesterday's low − today's low\n/// if DownMove > UpMove and DownMove > 0, then -DM = DownMove, else -DM = 0\n///\n/// When length > 1, Wilder's smoothing is applied:\n///   -DM(n) = previous -DM(n) − previous -DM(n)/n + today's -DM(1)\npub const DirectionalMovementMinus = struct {\n    length: usize,\n    no_smoothing: bool,\n    count: usize,\n    previous_high: f64,\n    previous_low: f64,\n    value: f64,\n    accumulator: f64,\n    primed: bool,\n\n    pub const Error = error{InvalidLength};\n\n    pub fn init(params: DirectionalMovementMinusParams) Error!DirectionalMovementMinus {\n        if (params.length < 1) return error.InvalidLength;\n\n        return .{\n            .length = params.length,\n            .no_smoothing = params.length == 1,\n            .count = 0,\n            .previous_high = 0,\n            .previous_low = 0,\n            .value = math.nan(f64),\n            .accumulator = 0,\n            .primed = false,\n        };\n    }\n\n    pub fn deinit(_: *DirectionalMovementMinus) void {}\n    pub fn fixSlices(_: *DirectionalMovementMinus) void {}\n\n    /// Core update given high and low values.\n    pub fn update(self: *DirectionalMovementMinus, high_in: f64, low_in: f64) f64 {\n        if (math.isNan(high_in) or math.isNan(low_in)) {\n            return math.nan(f64);\n        }\n\n        var high = high_in;\n        var low = low_in;\n        if (high < low) {\n            const tmp = high;\n            high = low;\n            low = tmp;\n        }\n\n        if (self.no_smoothing) {\n            if (self.primed) {\n                const delta_plus = high - self.previous_high;\n                const delta_minus = self.previous_low - low;\n                if (delta_minus > 0 and delta_minus > delta_plus) {\n                    self.value = delta_minus;\n                } else {\n                    self.value = 0;\n                }\n            } else {\n                if (self.count > 0) {\n                    const delta_plus = high - self.previous_high;\n                    const delta_minus = self.previous_low - low;\n                    if (delta_minus > 0 and delta_minus > delta_plus) {\n                        self.value = delta_minus;\n                    } else {\n                        self.value = 0;\n                    }\n                    self.primed = true;\n                }\n                self.count += 1;\n            }\n        } else {\n            const n: f64 = @floatFromInt(self.length);\n            if (self.primed) {\n                const delta_plus = high - self.previous_high;\n                const delta_minus = self.previous_low - low;\n                if (delta_minus > 0 and delta_minus > delta_plus) {\n                    self.accumulator += -self.accumulator / n + delta_minus;\n                } else {\n                    self.accumulator += -self.accumulator / n;\n                }\n                self.value = self.accumulator;\n            } else {\n                if (self.count > 0 and self.length >= self.count) {\n                    const delta_plus = high - self.previous_high;\n                    const delta_minus = self.previous_low - low;\n                    if (self.length > self.count) {\n                        if (delta_minus > 0 and delta_minus > delta_plus) {\n                            self.accumulator += delta_minus;\n                        }\n                    } else {\n                        if (delta_minus > 0 and delta_minus > delta_plus) {\n                            self.accumulator += -self.accumulator / n + delta_minus;\n                        } else {\n                            self.accumulator += -self.accumulator / n;\n                        }\n                        self.value = self.accumulator;\n                        self.primed = true;\n                    }\n                }\n                self.count += 1;\n            }\n        }\n\n        self.previous_low = low;\n        self.previous_high = high;\n\n        return self.value;\n    }\n\n    /// Update using a single sample value as substitute for high and low.\n    pub fn updateSample(self: *DirectionalMovementMinus, sample: f64) f64 {\n        return self.update(sample, sample);\n    }\n\n    pub fn isPrimed(self: *const DirectionalMovementMinus) bool {\n        return self.primed;\n    }\n\n    pub fn getMetadata(_: *const DirectionalMovementMinus, out: *Metadata) void {\n        const mnemonic = \"-dm\";\n        const description = \"Directional Movement Minus\";\n        build_metadata_mod.buildMetadata(out, Identifier.directional_movement_minus, mnemonic, description, &.{\n            .{ .mnemonic = mnemonic, .description = description },\n        });\n    }\n\n    fn makeOutput(self: *const DirectionalMovementMinus, time: i64) OutputArray {\n        var out = OutputArray{};\n        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });\n        return out;\n    }\n\n    pub fn updateScalar(self: *DirectionalMovementMinus, sample: *const Scalar) OutputArray {\n        _ = self.update(sample.value, sample.value);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateBar(self: *DirectionalMovementMinus, sample: *const Bar) OutputArray {\n        _ = self.update(sample.high, sample.low);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateQuote(self: *DirectionalMovementMinus, sample: *const Quote) OutputArray {\n        const mid = (sample.bid_price + sample.ask_price) / 2.0;\n        _ = self.update(mid, mid);\n        return self.makeOutput(sample.time);\n    }\n\n    pub fn updateTrade(self: *DirectionalMovementMinus, sample: *const Trade) OutputArray {\n        _ = self.update(sample.price, sample.price);\n        return self.makeOutput(sample.time);\n    }\n\n    // --- Indicator interface ---\n\n    pub fn indicator(self: *DirectionalMovementMinus) indicator_mod.Indicator {\n        return indicator_mod.Indicator{\n            .ptr = @ptrCast(self),\n            .vtable = &.{\n                .isPrimed = vtableIsPrimed,\n                .metadata = vtableMetadata,\n                .updateScalar = vtableUpdateScalar,\n                .updateBar = vtableUpdateBar,\n                .updateQuote = vtableUpdateQuote,\n                .updateTrade = vtableUpdateTrade,\n            },\n        };\n    }\n\n    fn vtableIsPrimed(ptr: *const anyopaque) bool {\n        const self: *const DirectionalMovementMinus = @ptrCast(@alignCast(ptr));\n        return self.isPrimed();\n    }\n\n    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {\n        const self: *const DirectionalMovementMinus = @ptrCast(@alignCast(ptr));\n        self.getMetadata(out);\n    }\n\n    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {\n        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));\n        return self.updateScalar(sample);\n    }\n\n    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {\n        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));\n        return self.updateBar(sample);\n    }\n\n    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {\n        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));\n        return self.updateQuote(sample);\n    }\n\n    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {\n        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));\n        return self.updateTrade(sample);\n    }\n};\n\n// ============================================================================\n// Tests\n// ============================================================================\n\nconst testing = std.testing;\n\nfn almostEqual(a: f64, b: f64, tolerance: f64) bool {\n    if (math.isNan(a) and math.isNan(b)) return true;\n    if (math.isNan(a) or math.isNan(b)) return false;\n    return @abs(a - b) <= tolerance;\n}\n\nconst test_high = [_]f64{\n    93.250000,  94.940000,  96.375000,  96.190000,  96.000000,  94.720000,  95.000000,  93.720000,  92.470000,  92.750000,  96.250000,\n    99.625000,  99.125000,  92.750000,  91.315000,  93.250000,  93.405000,  90.655000,  91.970000,  92.250000,  90.345000,  88.500000,\n    88.250000,  85.500000,  84.440000,  84.750000,  84.440000,  89.405000,  88.125000,  89.125000,  87.155000,  87.250000,  87.375000,\n    88.970000,  90.000000,  89.845000,  86.970000,  85.940000,  84.750000,  85.470000,  84.470000,  88.500000,  89.470000,  90.000000,\n    92.440000,  91.440000,  92.970000,  91.720000,  91.155000,  91.750000,  90.000000,  88.875000,  89.000000,  85.250000,  83.815000,\n    85.250000,  86.625000,  87.940000,  89.375000,  90.625000,  90.750000,  88.845000,  91.970000,  93.375000,  93.815000,  94.030000,\n    94.030000,  91.815000,  92.000000,  91.940000,  89.750000,  88.750000,  86.155000,  84.875000,  85.940000,  99.375000,  103.280000,\n    105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000, 110.940000,\n    110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000, 115.345000, 113.000000,\n    118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000, 119.870000, 120.870000, 116.750000,\n    116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000, 127.750000, 124.190000, 124.440000, 125.750000,\n    124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000, 133.500000, 135.500000, 137.440000, 138.690000, 139.190000,\n    138.500000, 138.130000, 137.500000, 138.880000, 132.130000, 129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000,\n    126.620000, 126.690000, 126.000000, 123.120000, 121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000,\n    124.500000, 127.870000, 128.560000, 129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000,\n    125.620000, 126.000000, 128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000,\n    135.000000, 132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,\n    122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000, 107.250000,\n    108.870000, 109.000000, 108.500000, 113.060000, 93.000000,  94.620000,  95.120000,  96.000000,  95.560000,  95.310000,  99.000000,\n    98.810000,  96.810000,  95.940000,  94.440000,  92.940000,  93.940000,  95.500000,  97.060000,  97.500000,  96.250000,  96.370000,\n    95.000000,  94.870000,  98.250000,  105.120000, 108.440000, 109.870000, 105.000000, 106.000000, 104.940000, 104.500000, 104.440000,\n    106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000, 112.750000, 110.190000, 107.940000, 109.690000,\n    111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000, 110.750000, 110.500000, 110.500000, 109.500000,\n};\n\nconst test_low = [_]f64{\n    90.750000,  91.405000,  94.250000,  93.500000,  92.815000,  93.500000,  92.000000,  89.750000,  89.440000,  90.625000,  92.750000,\n    96.315000,  96.030000,  88.815000,  86.750000,  90.940000,  88.905000,  88.780000,  89.250000,  89.750000,  87.500000,  86.530000,\n    84.625000,  82.280000,  81.565000,  80.875000,  81.250000,  84.065000,  85.595000,  85.970000,  84.405000,  85.095000,  85.500000,\n    85.530000,  87.875000,  86.565000,  84.655000,  83.250000,  82.565000,  83.440000,  82.530000,  85.065000,  86.875000,  88.530000,\n    89.280000,  90.125000,  90.750000,  89.000000,  88.565000,  90.095000,  89.000000,  86.470000,  84.000000,  83.315000,  82.000000,\n    83.250000,  84.750000,  85.280000,  87.190000,  88.440000,  88.250000,  87.345000,  89.280000,  91.095000,  89.530000,  91.155000,\n    92.000000,  90.530000,  89.970000,  88.815000,  86.750000,  85.065000,  82.030000,  81.500000,  82.565000,  96.345000,  96.470000,\n    101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000, 108.530000,\n    108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000, 110.875000, 110.500000,\n    110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000, 116.060000, 116.560000, 113.310000,\n    112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000, 123.000000, 121.750000, 121.560000, 123.120000,\n    122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000, 130.630000, 132.130000, 133.880000, 135.380000, 135.750000,\n    136.190000, 134.500000, 135.380000, 133.690000, 126.060000, 126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000,\n    124.620000, 124.370000, 121.810000, 118.190000, 118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000,\n    122.000000, 124.500000, 126.560000, 123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000,\n    122.690000, 122.870000, 125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000,\n    131.940000, 129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,\n    119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000, 104.690000,\n    103.870000, 104.690000, 105.440000, 107.000000, 89.000000,  92.500000,  92.120000,  94.620000,  92.810000,  94.250000,  96.250000,\n    96.370000,  93.690000,  93.500000,  90.000000,  90.190000,  90.500000,  92.120000,  94.120000,  94.870000,  93.000000,  93.870000,\n    93.000000,  92.620000,  93.560000,  98.370000,  104.440000, 106.000000, 101.810000, 104.120000, 103.370000, 102.120000, 102.250000,\n    103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000, 106.560000, 106.870000, 104.500000, 105.750000,\n    108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000, 109.060000, 108.750000, 108.560000, 106.620000,\n};\n\nconst test_expected_dmm1 = [_]f64{\n    math.nan(f64),\n    0.0000, 0.0000, 0.7500, 0.6850, 0.0000, 1.5000, 2.2500, 0.3100, 0.0000, 0.0000,\n    0.0000, 0.2850, 7.2150, 2.0650, 0.0000, 2.0350, 0.1250, 0.0000, 0.0000, 2.2500,\n    0.9700, 1.9050, 2.3450, 0.7150, 0.6900, 0.0000, 0.0000, 0.0000, 0.0000, 1.5650,\n    0.0000, 0.0000, 0.0000, 0.0000, 1.3100, 1.9100, 1.4050, 0.6850, 0.0000, 0.9100,\n    0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.7500, 0.4350, 0.0000, 1.0950,\n    2.5300, 2.4700, 0.6850, 1.3150, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1900,\n    0.9050, 0.0000, 0.0000, 1.5650, 0.0000, 0.0000, 1.4700, 0.5600, 1.1550, 2.0650,\n    1.6850, 3.0350, 0.5300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 2.5000, 0.0300,\n    0.0000, 0.0000, 0.0000, 2.0350, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,\n    0.0000, 2.0000, 0.0000, 1.9700, 0.2800, 1.6550, 3.7200, 0.3750, 0.0000, 0.0000,\n    0.0000, 3.0000, 1.7500, 0.0000, 0.0000, 0.0000, 1.4400, 0.0000, 3.2500, 0.7500,\n    0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,\n    0.9300, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.1800, 0.0000, 0.0000, 0.0000,\n    0.0000, 0.0000, 1.6900, 0.0000, 1.6900, 7.6300, 0.0000, 3.3700, 0.8800, 0.0000,\n    0.0000, 0.0000, 1.1900, 0.2500, 2.5600, 3.6200, 0.1300, 0.0000, 0.0000, 0.0000,\n    2.1800, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.0600, 2.2500, 0.1900, 0.0000,\n    1.3100, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.2500, 0.0000, 0.0000,\n    0.0000, 0.0000, 0.0000, 2.3800, 0.0600, 0.0000, 2.3800, 5.8100, 0.0000, 0.0000,\n    1.8800, 2.9300, 1.0000, 0.0000, 0.0000, 1.6900, 0.3800, 4.1200, 0.0000, 0.0000,\n    0.0000, 4.0000, 4.0000, 0.0000, 3.1200, 5.0000, 0.3100, 0.0000, 0.0000, 0.0000,\n    0.0000, 18.0000, 0.0000, 0.0000, 0.0000, 1.8100, 0.0000, 0.0000, 0.0000, 2.6800,\n    0.1900, 3.5000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 1.8700, 0.0000, 0.8700,\n    0.3800, 0.0000, 0.0000, 0.0000, 0.0000, 4.1900, 0.0000, 0.7500, 1.2500, 0.0000,\n    0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 3.2500, 4.6900, 1.0000, 0.0000, 2.3700,\n    0.0000, 0.0000, 0.8700, 0.0000, 0.0000, 0.0000, 0.0700, 0.0000, 0.3100, 0.1900,\n    1.9400,\n};\n\nconst test_expected_dmm14 = [_]f64{\n    math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64),\n    math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64),\n    14.131785714285700, 13.122372448979600, 14.220060131195300, 13.329341550395700, 12.377245725367400,\n    11.493156744984000, 12.922216977485200, 12.969201479093400, 13.947829944872400, 15.296556377381500, 14.918945207568600,\n    14.543306264170800, 13.504498673872900, 12.539891625739100, 11.644185081043500, 10.812457575254600, 11.605139177022200,\n    10.776200664377700, 10.006472045493600, 9.291724042244060, 8.628029467798060, 9.321741648669630, 10.565902959478900,\n    11.216195605230400, 11.100038776285400, 10.307178863693600, 10.480951802001200, 9.732312387572540, 9.037147217031640,\n    8.391636701529380, 7.792234079991570, 7.235645931420740, 6.718814079176400, 7.988898787806660, 7.853263160106190,\n    7.292315791527170, 7.866436092132370, 9.834547799837210, 11.602080099848800, 11.458360092716800, 11.954905800379900,\n    11.100983957495600, 10.308056531960200, 9.571766779677320, 8.888069152557510, 8.253207070231970, 7.853692279501110,\n    8.197714259536750, 7.612163240998410, 7.068437295212810, 8.128548916983320, 7.547938280055940, 7.008799831480520,\n    7.978171272089050, 7.968301895511260, 8.554137474403320, 10.008127654803100, 10.978261393745700, 13.229099865621000,\n    12.814164160933800, 11.898866720867100, 11.048947669376600, 10.259737121564000, 9.526898755737980, 8.846405987470980,\n    10.714519845508800, 9.979196999401000, 9.266397213729500, 8.604511698463110, 7.989903720001460, 9.454196311429920,\n    8.778896574899210, 8.151832533834980, 7.569558781418200, 7.028876011316900, 6.526813439079980, 6.060612479145690,\n    5.627711587778140, 7.225732188651130, 6.709608460890340, 8.200350713683890, 7.894611376992180, 8.985710564349880,\n    12.063874095467700, 11.577168802934300, 10.750228174153300, 9.982354733142360, 9.269329395060770, 11.607234438270700,\n    12.528146264108500, 11.633278673815100, 10.802330197114000, 10.030735183034400, 10.754254098531900, 9.986093091493950,\n    12.522800727815800, 12.378314961543300, 11.494149607147300, 10.673138920922500, 9.910771855142320, 9.202859579775010,\n    8.545512466933940, 7.935118719295800, 7.368324525060390, 8.092015630413220, 7.514014513955130, 6.977299191529760,\n    7.408920677849070, 6.879712058002710, 6.388304053859660, 5.931996621441110, 5.508282577052460, 5.114833821548720,\n    4.929488548580960, 4.577382223682320, 4.250426350562150, 3.946824468379140, 3.664908434923490, 3.403129261000380,\n    4.850048599500350, 4.503616556678900, 5.871929659773260, 13.082506112646600, 12.148041390314700, 14.650324148149400,\n    14.483872423281500, 13.449310107332900, 12.488645099666200, 11.596599021118600, 11.958270519610200, 11.354108339638000,\n    13.103100601092400, 15.787164843871600, 14.789510212166400, 13.733116625583100, 12.752179723755800, 11.841309743487500,\n    13.175501904667000, 12.234394625762200, 11.360509295350600, 10.549044345682700, 9.795541178133930, 9.095859665410080,\n    11.506155403595100, 12.934287160481100, 12.200409506161100, 11.328951684292400, 11.829740849700100, 10.984759360435800,\n    10.200133691833200, 9.471552713845160, 8.795013234284790, 8.166798003264450, 7.583455288745560, 8.291779910978020,\n    7.699509917336730, 7.149544923241250, 6.638863143009730, 6.164658632794750, 5.724325873309410, 7.695445453787310,\n    7.205770778516790, 6.691072865765590, 8.593139089639470, 13.789343440379500, 12.804390337495300, 11.889791027674200,\n    12.920520239983200, 14.927625937127200, 14.861366941618100, 13.799840731502600, 12.814137822109500, 13.588842263387400,\n    12.998210673145400, 16.189767053635100, 15.033355121232600, 13.959544041144500, 12.962433752491300, 16.036545627313400,\n    18.891078082505300, 17.541715362326300, 19.408735693588700, 23.022397429761000, 21.687940470492300, 20.138801865457200,\n    18.700316017924500, 17.364579159501300, 16.124252076679800, 32.972519785488400, 30.617339800810700, 28.430386957895600,\n    26.399645032331600, 26.323956101450800, 24.443673522775800, 22.697696842577500, 21.076432782393400, 22.250973297936700,\n    20.851618062369800, 22.862216772200500, 21.229201288471900, 19.712829767866800, 18.304770498733400, 16.997286891681000,\n    15.783194970846700, 16.525823901500500, 15.345407908536200, 15.119307343640700, 14.419356819095000, 13.389402760588200,\n    12.433016849117600, 11.544944217037800, 10.720305344392200, 14.144569248364200, 13.134242873481000, 12.946082668232400,\n    13.271362477644400, 12.323408014955500, 11.443164585315800, 10.625795686364700, 9.866810280195770, 9.162038117324650,\n    8.507606823230030, 11.149920621570700, 15.043497720030000, 14.968962168599300, 13.899750585127900, 15.276911257618800,\n    14.185703310646000, 13.172438788457000, 13.101550303567200, 12.165725281883800, 11.296744904606400, 10.489834554277400,\n    9.810560657543280, 9.109806324861620, 8.769105873085790, 8.332741167865380, 9.677545370160700,\n};\n\ntest \"DirectionalMovementMinus update length=14\" {\n    const tolerance = 1e-8;\n    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });\n\n    for (0..test_high.len) |i| {\n        const act = dmm.update(test_high[i], test_low[i]);\n        const exp = test_expected_dmm14[i];\n\n        if (math.isNan(exp)) {\n            try testing.expect(math.isNan(act));\n        } else {\n            try testing.expect(!math.isNan(act));\n            try testing.expect(almostEqual(act, exp, tolerance));\n        }\n    }\n}\n\ntest \"DirectionalMovementMinus update length=1\" {\n    const tolerance = 1e-8;\n    var dmm = try DirectionalMovementMinus.init(.{ .length = 1 });\n\n    for (0..test_high.len) |i| {\n        const act = dmm.update(test_high[i], test_low[i]);\n        const exp = test_expected_dmm1[i];\n\n        if (math.isNan(exp)) {\n            try testing.expect(math.isNan(act));\n        } else {\n            try testing.expect(!math.isNan(act));\n            try testing.expect(almostEqual(act, exp, tolerance));\n        }\n    }\n}\n\ntest \"DirectionalMovementMinus constructor validation\" {\n    const result = DirectionalMovementMinus.init(.{ .length = 0 });\n    try testing.expectError(error.InvalidLength, result);\n}\n\ntest \"DirectionalMovementMinus isPrimed length=1\" {\n    var dmm = try DirectionalMovementMinus.init(.{ .length = 1 });\n\n    try testing.expect(!dmm.isPrimed());\n\n    _ = dmm.update(test_high[0], test_low[0]);\n    try testing.expect(!dmm.isPrimed());\n\n    _ = dmm.update(test_high[1], test_low[1]);\n    try testing.expect(dmm.isPrimed());\n}\n\ntest \"DirectionalMovementMinus isPrimed length=14\" {\n    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });\n\n    for (0..14) |i| {\n        _ = dmm.update(test_high[i], test_low[i]);\n        try testing.expect(!dmm.isPrimed());\n    }\n\n    _ = dmm.update(test_high[14], test_low[14]);\n    try testing.expect(dmm.isPrimed());\n}\n\ntest \"DirectionalMovementMinus NaN passthrough\" {\n    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });\n\n    try testing.expect(math.isNan(dmm.update(math.nan(f64), 1)));\n    try testing.expect(math.isNan(dmm.update(1, math.nan(f64))));\n    try testing.expect(math.isNan(dmm.update(math.nan(f64), math.nan(f64))));\n    try testing.expect(math.isNan(dmm.updateSample(math.nan(f64))));\n}\n\ntest \"DirectionalMovementMinus high/low swap\" {\n    var dmm1 = try DirectionalMovementMinus.init(.{ .length = 1 });\n    var dmm2 = try DirectionalMovementMinus.init(.{ .length = 1 });\n\n    _ = dmm1.update(10, 5);\n    _ = dmm2.update(5, 10);\n\n    const v1 = dmm1.update(12, 6);\n    const v2 = dmm2.update(6, 12);\n\n    try testing.expectEqual(v1, v2);\n}\n\ntest \"DirectionalMovementMinus metadata\" {\n    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });\n    var meta: Metadata = undefined;\n    dmm.getMetadata(&meta);\n\n    try testing.expectEqual(Identifier.directional_movement_minus, meta.identifier);\n    try testing.expectEqualStrings(\"-dm\", meta.mnemonic);\n    try testing.expectEqualStrings(\"Directional Movement Minus\", meta.description);\n    try testing.expectEqual(@as(usize, 1), meta.outputs_len);\n}\n\ntest \"DirectionalMovementMinus updateBar\" {\n    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });\n\n    for (0..14) |i| {\n        _ = dmm.update(test_high[i], test_low[i]);\n    }\n\n    const bar = Bar{ .time = 42, .open = 0, .high = test_high[14], .low = test_low[14], .close = 0, .volume = 0 };\n    const out = dmm.updateBar(&bar);\n    try testing.expectEqual(@as(usize, 1), out.len);\n}\n"
}
```

**Output:**
```
Wrote file successfully.

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/icalc/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.identifier" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.factory.factory" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/iconf/main.py">
ERROR [20:6] Import "py.entities.bar" could not be resolved
ERROR [21:6] Import "py.entities.scalar" could not be resolved
ERROR [22:6] Import "py.indicators.core.descriptors" could not be resolved
ERROR [23:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [24:6] Import "py.indicators.core.outputs.band" could not be resolved
ERROR [25:6] Import "py.indicators.core.outputs.heatmap" could not be resolved
ERROR [26:6] Import "py.indicators.core.outputs.shape" could not be resolved
ERROR [27:6] Import "py.indicators.core.pane" could not be resolved
ERROR [28:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [31:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>

LSP errors detected in other files:
<diagnostics file="/home/dev/repos/chi/zpano/py/cmd/ifres/main.py">
ERROR [16:6] Import "py.indicators.core.frequency_response" could not be resolved
ERROR [17:6] Import "py.indicators.core.indicator" could not be resolved
ERROR [18:6] Import "py.indicators.factory.factory" could not be resolved
ERROR [21:6] Import "py.cmd.icalc.main" could not be resolved
</diagnostics>
```

---

## Assistant (Build · claude-opus-4.6 · 6.2s)

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Write directional_movement_minus.zig (mirror of DM+)",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Register TR, DM+, DM- in indicators.zig",
      "priority": "high",
      "status": "in_progress"
    },
    {
      "content": "Build and test all three new indicators",
      "priority": "high",
      "status": "pending"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Write directional_movement_minus.zig (mirror of DM+)",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Register TR, DM+, DM- in indicators.zig",
    "priority": "high",
    "status": "in_progress"
  },
  {
    "content": "Build and test all three new indicators",
    "priority": "high",
    "status": "pending"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 7.5s)

Now let me check the indicators.zig registration file and the identifier enum to see what's needed.

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && grep -n 'welles_wilder\\|true_range\\|directional_movement' src/indicators/indicators.zig",
  "description": "Check existing welles_wilder registrations"
}
```

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && grep -n 'true_range\\|directional_movement_plus\\|directional_movement_minus' src/indicators/core/identifier.zig",
  "description": "Check identifiers exist"
}
```

**Output:**
```
33:    true_range = 29,
34:    average_true_range = 30,
35:    normalized_average_true_range = 31,
36:    directional_movement_minus = 32,
37:    directional_movement_plus = 33,
109:            .true_range => "trueRange",
110:            .average_true_range => "averageTrueRange",
111:            .normalized_average_true_range => "normalizedAverageTrueRange",
112:            .directional_movement_minus => "directionalMovementMinus",
113:            .directional_movement_plus => "directionalMovementPlus",
187:            .{ "trueRange", Identifier.true_range },
188:            .{ "averageTrueRange", Identifier.average_true_range },
189:            .{ "normalizedAverageTrueRange", Identifier.normalized_average_true_range },
190:            .{ "directionalMovementMinus", Identifier.directional_movement_minus },
191:            .{ "directionalMovementPlus", Identifier.directional_movement_plus },

```

---

## Assistant (Build · claude-opus-4.6 · 5.8s)

Good, identifiers exist. Now let me see how indicators are registered in `indicators.zig`:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "limit": 60
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig</path>
<type>file</type>
<content>1: // Root module for the indicators library.
2: // Re-exports all core types and sub-modules.
3: 
4: // --- Core enums ---
5: pub const shape = @import("core/outputs/shape.zig");
6: pub const role = @import("core/role.zig");
7: pub const pane = @import("core/pane.zig");
8: pub const adaptivity = @import("core/adaptivity.zig");
9: pub const input_requirement = @import("core/input_requirement.zig");
10: pub const volume_usage = @import("core/volume_usage.zig");
11: 
12: // --- Core types ---
13: pub const identifier = @import("core/identifier.zig");
14: pub const output_descriptor = @import("core/output_descriptor.zig");
15: pub const output_metadata = @import("core/output_metadata.zig");
16: pub const metadata = @import("core/metadata.zig");
17: pub const descriptor = @import("core/descriptor.zig");
18: pub const descriptors = @import("core/descriptors.zig");
19: pub const build_metadata = @import("core/build_metadata.zig");
20: pub const component_triple_mnemonic = @import("core/component_triple_mnemonic.zig");
21: 
22: // --- Indicator interface and outputs ---
23: pub const indicator = @import("core/indicator.zig");
24: pub const line_indicator = @import("core/line_indicator.zig");
25: 
26: // --- Output types ---
27: pub const band = @import("core/outputs/band.zig");
28: pub const heatmap = @import("core/outputs/heatmap.zig");
29: pub const polyline = @import("core/outputs/polyline.zig");
30: 
31: // --- Convenience type aliases ---
32: pub const Identifier = identifier.Identifier;
33: pub const Shape = shape.Shape;
34: pub const Role = role.Role;
35: pub const Pane = pane.Pane;
36: pub const Adaptivity = adaptivity.Adaptivity;
37: pub const InputRequirement = input_requirement.InputRequirement;
38: pub const VolumeUsage = volume_usage.VolumeUsage;
39: pub const OutputDescriptor = output_descriptor.OutputDescriptor;
40: pub const OutputMetadata = output_metadata.OutputMetadata;
41: pub const Metadata = metadata.Metadata;
42: pub const Descriptor = descriptor.Descriptor;
43: pub const Indicator = indicator.Indicator;
44: pub const OutputValue = indicator.OutputValue;
45: pub const OutputArray = indicator.OutputArray;
46: pub const LineIndicator = line_indicator.LineIndicator;
47: pub const Band = band.Band;
48: pub const Heatmap = heatmap.Heatmap;
49: pub const Polyline = polyline.Polyline;
50: pub const Point = polyline.Point;
51: 
52: // --- Functions ---
53: pub const descriptorOf = descriptor.descriptorOf;
54: pub const allDescriptors = descriptor.allDescriptors;
55: pub const buildMetadata = build_metadata.buildMetadata;
56: pub const componentTripleMnemonic = component_triple_mnemonic.componentTripleMnemonic;
57: 
58: // --- Indicator implementations ---
59: pub const simple_moving_average = @import("common/simple_moving_average/simple_moving_average.zig");
60: pub const weighted_moving_average = @import("common/weighted_moving_average/weighted_moving_average.zig");

(Showing lines 1-60 of 135. Use offset=61 to continue.)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 4.7s)

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "offset": 90
}
```

**Output:**
```
<path>/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig</path>
<type>file</type>
<content>90: pub const ultimate_oscillator = @import("larry_williams/ultimate_oscillator/ultimate_oscillator.zig");
91: pub const advance_decline = @import("marc_chaikin/advance_decline/advance_decline.zig");
92: pub const advance_decline_oscillator = @import("marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig");
93: pub const aroon = @import("tushar_chande/aroon/aroon.zig");
94: pub const chande_momentum_oscillator = @import("tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig");
95: 
96: // Force-include tests from sub-modules.
97: comptime {
98:     _ = identifier;
99:     _ = simple_moving_average;
100:     _ = weighted_moving_average;
101:     _ = triangular_moving_average;
102:     _ = exponential_moving_average;
103:     _ = momentum;
104:     _ = rate_of_change;
105:     _ = rate_of_change_percent;
106:     _ = rate_of_change_ratio;
107:     _ = variance;
108:     _ = standard_deviation;
109:     _ = absolute_price_oscillator;
110:     _ = linear_regression;
111:     _ = pearsons_correlation_coefficient;
112:     _ = balance_of_power;
113:     _ = on_balance_volume;
114:     _ = commodity_channel_index;
115:     _ = triple_exponential_moving_average_oscillator;
116:     _ = money_flow_index;
117:     _ = kaufman_adaptive_moving_average;
118:     _ = jurik_moving_average;
119:     _ = stochastic;
120:     _ = adaptive_trend_and_cycle_filter;
121:     _ = double_exponential_moving_average;
122:     _ = triple_exponential_moving_average;
123:     _ = t2_exponential_moving_average;
124:     _ = t3_exponential_moving_average;
125:     _ = percentage_price_oscillator;
126:     _ = moving_average_convergence_divergence;
127:     _ = bollinger_bands;
128:     _ = bollinger_bands_trend;
129:     _ = williams_percent_r;
130:     _ = ultimate_oscillator;
131:     _ = advance_decline;
132:     _ = advance_decline_oscillator;
133:     _ = aroon;
134:     _ = chande_momentum_oscillator;
135: }

(End of file - total 135 lines)
</content>
```

---

## Assistant (Build · claude-opus-4.6 · 7.1s)

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "oldString": "pub const chande_momentum_oscillator = @import(\"tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig\");\n\n// Force-include tests from sub-modules.\ncomptime {",
  "newString": "pub const chande_momentum_oscillator = @import(\"tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig\");\npub const true_range = @import(\"welles_wilder/true_range/true_range.zig\");\npub const directional_movement_plus = @import(\"welles_wilder/directional_movement_plus/directional_movement_plus.zig\");\npub const directional_movement_minus = @import(\"welles_wilder/directional_movement_minus/directional_movement_minus.zig\");\n\n// Force-include tests from sub-modules.\ncomptime {"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 5.4s)

**Tool: edit**

**Input:**
```json
{
  "filePath": "/home/dev/repos/chi/zpano/zig/src/indicators/indicators.zig",
  "newString": "    _ = chande_momentum_oscillator;\n    _ = true_range;\n    _ = directional_movement_plus;\n    _ = directional_movement_minus;\n}",
  "oldString": "    _ = chande_momentum_oscillator;\n}"
}
```

**Output:**
```
Edit applied successfully.
```

---

## Assistant (Build · claude-opus-4.6 · 5.8s)

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Write directional_movement_minus.zig (mirror of DM+)",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Register TR, DM+, DM- in indicators.zig",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Build and test all three new indicators",
      "priority": "high",
      "status": "in_progress"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Write directional_movement_minus.zig (mirror of DM+)",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Register TR, DM+, DM- in indicators.zig",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Build and test all three new indicators",
    "priority": "high",
    "status": "in_progress"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 6.6s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | tail -30",
  "description": "Build and test all indicators",
  "timeout": 120000
}
```

**Output:**
```
|  +- compile test Debug native cached 10ms MaxRSS:39M
+- run test success 5ms MaxRSS:10M
|  +- compile test Debug native cached 12ms MaxRSS:39M
+- run test 55 pass (55 total) 10ms MaxRSS:10M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 155 pass (155 total) 34ms MaxRSS:10M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 4 pass (4 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 2 pass (2 total) 8ms MaxRSS:10M
|  +- compile test Debug native cached 7ms MaxRSS:39M
+- run test 2 pass (2 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 6 pass (6 total) 6ms MaxRSS:11M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 7 pass (7 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 4ms MaxRSS:39M
+- run test success 7ms MaxRSS:10M
|  +- compile test Debug native cached 8ms MaxRSS:39M
+- run test success 6ms MaxRSS:10M
|  +- compile test Debug native cached 5ms MaxRSS:39M
+- run test 11 pass (11 total) 7ms MaxRSS:10M
|  +- compile test Debug native cached 9ms MaxRSS:39M
+- run test 10 pass (10 total) 5ms MaxRSS:11M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 4 pass (4 total) 5ms MaxRSS:11M
|  +- compile test Debug native cached 6ms MaxRSS:39M
+- run test 319 pass (319 total) 44ms MaxRSS:11M
   +- compile test Debug native success 828ms MaxRSS:179M


```

---

## Assistant (Build · claude-opus-4.6 · 5.9s)

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | grep -E 'FAIL|error|319'",
  "description": "Check for failures",
  "timeout": 120000
}
```

**Output:**
```
+- run test 319 pass (319 total) 49ms MaxRSS:11M

```

---

## Assistant (Build · claude-opus-4.6 · 6.7s)

All 319 tests pass (up from the previous count, confirming TR + DM+ + DM- tests are included). Let me verify the count increase:

**Tool: bash**

**Input:**
```json
{
  "command": "cd zig && zig build test --summary all 2>&1 | grep -oP '\\d+ pass' | tail -1",
  "description": "Get total test count",
  "timeout": 120000
}
```

**Output:**
```
319 pass

```

---

## Assistant (Build · claude-opus-4.6 · 5.1s)

**Tool: todowrite**

**Input:**
```json
{
  "todos": [
    {
      "content": "Write directional_movement_minus.zig (mirror of DM+)",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Register TR, DM+, DM- in indicators.zig",
      "priority": "high",
      "status": "completed"
    },
    {
      "content": "Build and test all three new indicators",
      "priority": "high",
      "status": "completed"
    }
  ]
}
```

**Output:**
```
[
  {
    "content": "Write directional_movement_minus.zig (mirror of DM+)",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Register TR, DM+, DM- in indicators.zig",
    "priority": "high",
    "status": "completed"
  },
  {
    "content": "Build and test all three new indicators",
    "priority": "high",
    "status": "completed"
  }
]
```

---

## Assistant (Build · claude-opus-4.6 · 9.7s)

All three welleswilder Level 0 indicators are done. **319 tests pass** (up from the previous baseline). 

Completed:
- `directional_movement_minus.zig` — written and passing all 8 tests (length=1, length=14, validation, isPrimed x2, NaN, swap, metadata, updateBar)
- All three indicators (TR, DM+, DM-) registered in `indicators.zig` and building cleanly

**Next up: RSI (Batch 2)** — LineIndicator pattern with component triples, Wilder-smoothed avg gain/loss. This unblocks StochasticRSI. Want me to continue?

---

## User

yes

---

## Assistant (Build · claude-opus-4.6)

---

