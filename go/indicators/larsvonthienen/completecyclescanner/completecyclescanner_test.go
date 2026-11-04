// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package completecyclescanner

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 49.0
		p.MaxPeriod = 80
		p.MaxSegments = 5
		p.MinPeriod = 5
		p.PeakDistance = 2
		p.PhaseCycles = 2.0
		p.Smoothing = 1600.0
		p.TopK = 3
		p.Window = 200
		indicator, err := NewCompleteCycleScanner(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"ACCEPTED_COUNT": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ACCEPTED_COUNT, "AMPLITUDE_P5": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P5, "GENUINE_P5": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P5, "AMPLITUDE_P6": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P6, "GENUINE_P6": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P6, "AMPLITUDE_P7": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P7, "GENUINE_P7": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P7, "AMPLITUDE_P8": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P8, "GENUINE_P8": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P8, "AMPLITUDE_P9": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P9, "GENUINE_P9": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P9, "AMPLITUDE_P10": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P10, "GENUINE_P10": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P10, "AMPLITUDE_P11": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P11, "GENUINE_P11": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P11, "AMPLITUDE_P12": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P12, "GENUINE_P12": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P12, "AMPLITUDE_P13": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P13, "GENUINE_P13": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P13, "AMPLITUDE_P14": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P14, "GENUINE_P14": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P14, "AMPLITUDE_P15": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P15, "GENUINE_P15": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P15, "AMPLITUDE_P16": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P16, "GENUINE_P16": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P16, "AMPLITUDE_P17": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P17, "GENUINE_P17": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P17, "AMPLITUDE_P18": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P18, "GENUINE_P18": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P18, "AMPLITUDE_P19": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P19, "GENUINE_P19": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P19, "AMPLITUDE_P20": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P20, "GENUINE_P20": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P20, "AMPLITUDE_P21": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P21, "GENUINE_P21": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P21, "AMPLITUDE_P22": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P22, "GENUINE_P22": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P22, "AMPLITUDE_P23": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P23, "GENUINE_P23": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P23, "AMPLITUDE_P24": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P24, "GENUINE_P24": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P24, "AMPLITUDE_P25": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P25, "GENUINE_P25": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P25, "AMPLITUDE_P26": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P26, "GENUINE_P26": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P26, "AMPLITUDE_P27": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P27, "GENUINE_P27": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P27, "AMPLITUDE_P28": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P28, "GENUINE_P28": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P28, "AMPLITUDE_P29": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P29, "GENUINE_P29": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P29, "AMPLITUDE_P30": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P30, "GENUINE_P30": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P30, "AMPLITUDE_P31": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P31, "GENUINE_P31": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P31, "AMPLITUDE_P32": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P32, "GENUINE_P32": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P32, "AMPLITUDE_P33": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P33, "GENUINE_P33": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P33, "AMPLITUDE_P34": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P34, "GENUINE_P34": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P34, "AMPLITUDE_P35": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P35, "GENUINE_P35": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P35, "AMPLITUDE_P36": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P36, "GENUINE_P36": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P36, "AMPLITUDE_P37": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P37, "GENUINE_P37": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P37, "AMPLITUDE_P38": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P38, "GENUINE_P38": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P38, "AMPLITUDE_P39": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P39, "GENUINE_P39": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P39, "AMPLITUDE_P40": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P40, "GENUINE_P40": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P40, "AMPLITUDE_P41": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P41, "GENUINE_P41": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P41, "AMPLITUDE_P42": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P42, "GENUINE_P42": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P42, "AMPLITUDE_P43": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P43, "GENUINE_P43": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P43, "AMPLITUDE_P44": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P44, "GENUINE_P44": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P44, "AMPLITUDE_P45": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P45, "GENUINE_P45": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P45, "AMPLITUDE_P46": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P46, "GENUINE_P46": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P46, "AMPLITUDE_P47": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P47, "GENUINE_P47": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P47, "AMPLITUDE_P48": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P48, "GENUINE_P48": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P48, "AMPLITUDE_P49": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P49, "GENUINE_P49": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P49, "AMPLITUDE_P50": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P50, "GENUINE_P50": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P50, "AMPLITUDE_P51": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P51, "GENUINE_P51": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P51, "AMPLITUDE_P52": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P52, "GENUINE_P52": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P52, "AMPLITUDE_P53": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P53, "GENUINE_P53": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P53, "AMPLITUDE_P54": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P54, "GENUINE_P54": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P54, "AMPLITUDE_P55": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P55, "GENUINE_P55": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P55, "AMPLITUDE_P56": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P56, "GENUINE_P56": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P56, "AMPLITUDE_P57": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P57, "GENUINE_P57": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P57, "AMPLITUDE_P58": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P58, "GENUINE_P58": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P58, "AMPLITUDE_P59": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P59, "GENUINE_P59": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P59, "AMPLITUDE_P60": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P60, "GENUINE_P60": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P60, "AMPLITUDE_P61": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P61, "GENUINE_P61": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P61, "AMPLITUDE_P62": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P62, "GENUINE_P62": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P62, "AMPLITUDE_P63": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P63, "GENUINE_P63": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P63, "AMPLITUDE_P64": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P64, "GENUINE_P64": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P64, "AMPLITUDE_P65": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P65, "GENUINE_P65": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P65, "AMPLITUDE_P66": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P66, "GENUINE_P66": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P66, "AMPLITUDE_P67": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P67, "GENUINE_P67": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P67, "AMPLITUDE_P68": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P68, "GENUINE_P68": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P68, "AMPLITUDE_P69": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P69, "GENUINE_P69": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P69, "AMPLITUDE_P70": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P70, "GENUINE_P70": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P70, "AMPLITUDE_P71": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P71, "GENUINE_P71": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P71, "AMPLITUDE_P72": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P72, "GENUINE_P72": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P72, "AMPLITUDE_P73": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P73, "GENUINE_P73": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P73, "AMPLITUDE_P74": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P74, "GENUINE_P74": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P74, "AMPLITUDE_P75": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P75, "GENUINE_P75": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P75, "AMPLITUDE_P76": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P76, "GENUINE_P76": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P76, "AMPLITUDE_P77": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P77, "GENUINE_P77": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P77, "AMPLITUDE_P78": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P78, "GENUINE_P78": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P78, "AMPLITUDE_P79": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P79, "GENUINE_P79": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P79, "AMPLITUDE_P80": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE_P80, "GENUINE_P80": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_P80, "RANK1_PERIOD": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_PERIOD, "RANK1_AMPLITUDE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_AMPLITUDE, "RANK1_PHASE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_PHASE, "RANK1_STRENGTH": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_STRENGTH, "RANK1_GENUINE_PERCENT": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_GENUINE_PERCENT, "RANK1_SEGMENTS": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_SEGMENTS, "RANK1_FIT_AMPLITUDE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_FIT_AMPLITUDE, "RANK1_FIT_SAMPLES": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_FIT_SAMPLES, "RANK1_VALUE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_VALUE, "RANK1_SLOPE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_SLOPE, "RANK1_ETA_TOP": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_ETA_TOP, "RANK1_ETA_BOTTOM": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK1_ETA_BOTTOM, "RANK2_PERIOD": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_PERIOD, "RANK2_AMPLITUDE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_AMPLITUDE, "RANK2_PHASE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_PHASE, "RANK2_STRENGTH": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_STRENGTH, "RANK2_GENUINE_PERCENT": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_GENUINE_PERCENT, "RANK2_SEGMENTS": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_SEGMENTS, "RANK2_FIT_AMPLITUDE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_FIT_AMPLITUDE, "RANK2_FIT_SAMPLES": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_FIT_SAMPLES, "RANK2_VALUE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_VALUE, "RANK2_SLOPE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_SLOPE, "RANK2_ETA_TOP": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_ETA_TOP, "RANK2_ETA_BOTTOM": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK2_ETA_BOTTOM, "RANK3_PERIOD": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_PERIOD, "RANK3_AMPLITUDE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_AMPLITUDE, "RANK3_PHASE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_PHASE, "RANK3_STRENGTH": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_STRENGTH, "RANK3_GENUINE_PERCENT": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_GENUINE_PERCENT, "RANK3_SEGMENTS": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_SEGMENTS, "RANK3_FIT_AMPLITUDE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_FIT_AMPLITUDE, "RANK3_FIT_SAMPLES": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_FIT_SAMPLES, "RANK3_VALUE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_VALUE, "RANK3_SLOPE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_SLOPE, "RANK3_ETA_TOP": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_ETA_TOP, "RANK3_ETA_BOTTOM": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_RANK3_ETA_BOTTOM}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"ACCEPTED_COUNT": result.AcceptedCount}
			actual["AMPLITUDE_P5"] = math.NaN()
			if len(result.Spectrum) > 0 {
				actual["AMPLITUDE_P5"] = result.Spectrum[0]
			}
			actual["GENUINE_P5"] = math.NaN()
			if len(result.GenuinePercent) > 0 {
				actual["GENUINE_P5"] = result.GenuinePercent[0]
			}
			actual["AMPLITUDE_P6"] = math.NaN()
			if len(result.Spectrum) > 1 {
				actual["AMPLITUDE_P6"] = result.Spectrum[1]
			}
			actual["GENUINE_P6"] = math.NaN()
			if len(result.GenuinePercent) > 1 {
				actual["GENUINE_P6"] = result.GenuinePercent[1]
			}
			actual["AMPLITUDE_P7"] = math.NaN()
			if len(result.Spectrum) > 2 {
				actual["AMPLITUDE_P7"] = result.Spectrum[2]
			}
			actual["GENUINE_P7"] = math.NaN()
			if len(result.GenuinePercent) > 2 {
				actual["GENUINE_P7"] = result.GenuinePercent[2]
			}
			actual["AMPLITUDE_P8"] = math.NaN()
			if len(result.Spectrum) > 3 {
				actual["AMPLITUDE_P8"] = result.Spectrum[3]
			}
			actual["GENUINE_P8"] = math.NaN()
			if len(result.GenuinePercent) > 3 {
				actual["GENUINE_P8"] = result.GenuinePercent[3]
			}
			actual["AMPLITUDE_P9"] = math.NaN()
			if len(result.Spectrum) > 4 {
				actual["AMPLITUDE_P9"] = result.Spectrum[4]
			}
			actual["GENUINE_P9"] = math.NaN()
			if len(result.GenuinePercent) > 4 {
				actual["GENUINE_P9"] = result.GenuinePercent[4]
			}
			actual["AMPLITUDE_P10"] = math.NaN()
			if len(result.Spectrum) > 5 {
				actual["AMPLITUDE_P10"] = result.Spectrum[5]
			}
			actual["GENUINE_P10"] = math.NaN()
			if len(result.GenuinePercent) > 5 {
				actual["GENUINE_P10"] = result.GenuinePercent[5]
			}
			actual["AMPLITUDE_P11"] = math.NaN()
			if len(result.Spectrum) > 6 {
				actual["AMPLITUDE_P11"] = result.Spectrum[6]
			}
			actual["GENUINE_P11"] = math.NaN()
			if len(result.GenuinePercent) > 6 {
				actual["GENUINE_P11"] = result.GenuinePercent[6]
			}
			actual["AMPLITUDE_P12"] = math.NaN()
			if len(result.Spectrum) > 7 {
				actual["AMPLITUDE_P12"] = result.Spectrum[7]
			}
			actual["GENUINE_P12"] = math.NaN()
			if len(result.GenuinePercent) > 7 {
				actual["GENUINE_P12"] = result.GenuinePercent[7]
			}
			actual["AMPLITUDE_P13"] = math.NaN()
			if len(result.Spectrum) > 8 {
				actual["AMPLITUDE_P13"] = result.Spectrum[8]
			}
			actual["GENUINE_P13"] = math.NaN()
			if len(result.GenuinePercent) > 8 {
				actual["GENUINE_P13"] = result.GenuinePercent[8]
			}
			actual["AMPLITUDE_P14"] = math.NaN()
			if len(result.Spectrum) > 9 {
				actual["AMPLITUDE_P14"] = result.Spectrum[9]
			}
			actual["GENUINE_P14"] = math.NaN()
			if len(result.GenuinePercent) > 9 {
				actual["GENUINE_P14"] = result.GenuinePercent[9]
			}
			actual["AMPLITUDE_P15"] = math.NaN()
			if len(result.Spectrum) > 10 {
				actual["AMPLITUDE_P15"] = result.Spectrum[10]
			}
			actual["GENUINE_P15"] = math.NaN()
			if len(result.GenuinePercent) > 10 {
				actual["GENUINE_P15"] = result.GenuinePercent[10]
			}
			actual["AMPLITUDE_P16"] = math.NaN()
			if len(result.Spectrum) > 11 {
				actual["AMPLITUDE_P16"] = result.Spectrum[11]
			}
			actual["GENUINE_P16"] = math.NaN()
			if len(result.GenuinePercent) > 11 {
				actual["GENUINE_P16"] = result.GenuinePercent[11]
			}
			actual["AMPLITUDE_P17"] = math.NaN()
			if len(result.Spectrum) > 12 {
				actual["AMPLITUDE_P17"] = result.Spectrum[12]
			}
			actual["GENUINE_P17"] = math.NaN()
			if len(result.GenuinePercent) > 12 {
				actual["GENUINE_P17"] = result.GenuinePercent[12]
			}
			actual["AMPLITUDE_P18"] = math.NaN()
			if len(result.Spectrum) > 13 {
				actual["AMPLITUDE_P18"] = result.Spectrum[13]
			}
			actual["GENUINE_P18"] = math.NaN()
			if len(result.GenuinePercent) > 13 {
				actual["GENUINE_P18"] = result.GenuinePercent[13]
			}
			actual["AMPLITUDE_P19"] = math.NaN()
			if len(result.Spectrum) > 14 {
				actual["AMPLITUDE_P19"] = result.Spectrum[14]
			}
			actual["GENUINE_P19"] = math.NaN()
			if len(result.GenuinePercent) > 14 {
				actual["GENUINE_P19"] = result.GenuinePercent[14]
			}
			actual["AMPLITUDE_P20"] = math.NaN()
			if len(result.Spectrum) > 15 {
				actual["AMPLITUDE_P20"] = result.Spectrum[15]
			}
			actual["GENUINE_P20"] = math.NaN()
			if len(result.GenuinePercent) > 15 {
				actual["GENUINE_P20"] = result.GenuinePercent[15]
			}
			actual["AMPLITUDE_P21"] = math.NaN()
			if len(result.Spectrum) > 16 {
				actual["AMPLITUDE_P21"] = result.Spectrum[16]
			}
			actual["GENUINE_P21"] = math.NaN()
			if len(result.GenuinePercent) > 16 {
				actual["GENUINE_P21"] = result.GenuinePercent[16]
			}
			actual["AMPLITUDE_P22"] = math.NaN()
			if len(result.Spectrum) > 17 {
				actual["AMPLITUDE_P22"] = result.Spectrum[17]
			}
			actual["GENUINE_P22"] = math.NaN()
			if len(result.GenuinePercent) > 17 {
				actual["GENUINE_P22"] = result.GenuinePercent[17]
			}
			actual["AMPLITUDE_P23"] = math.NaN()
			if len(result.Spectrum) > 18 {
				actual["AMPLITUDE_P23"] = result.Spectrum[18]
			}
			actual["GENUINE_P23"] = math.NaN()
			if len(result.GenuinePercent) > 18 {
				actual["GENUINE_P23"] = result.GenuinePercent[18]
			}
			actual["AMPLITUDE_P24"] = math.NaN()
			if len(result.Spectrum) > 19 {
				actual["AMPLITUDE_P24"] = result.Spectrum[19]
			}
			actual["GENUINE_P24"] = math.NaN()
			if len(result.GenuinePercent) > 19 {
				actual["GENUINE_P24"] = result.GenuinePercent[19]
			}
			actual["AMPLITUDE_P25"] = math.NaN()
			if len(result.Spectrum) > 20 {
				actual["AMPLITUDE_P25"] = result.Spectrum[20]
			}
			actual["GENUINE_P25"] = math.NaN()
			if len(result.GenuinePercent) > 20 {
				actual["GENUINE_P25"] = result.GenuinePercent[20]
			}
			actual["AMPLITUDE_P26"] = math.NaN()
			if len(result.Spectrum) > 21 {
				actual["AMPLITUDE_P26"] = result.Spectrum[21]
			}
			actual["GENUINE_P26"] = math.NaN()
			if len(result.GenuinePercent) > 21 {
				actual["GENUINE_P26"] = result.GenuinePercent[21]
			}
			actual["AMPLITUDE_P27"] = math.NaN()
			if len(result.Spectrum) > 22 {
				actual["AMPLITUDE_P27"] = result.Spectrum[22]
			}
			actual["GENUINE_P27"] = math.NaN()
			if len(result.GenuinePercent) > 22 {
				actual["GENUINE_P27"] = result.GenuinePercent[22]
			}
			actual["AMPLITUDE_P28"] = math.NaN()
			if len(result.Spectrum) > 23 {
				actual["AMPLITUDE_P28"] = result.Spectrum[23]
			}
			actual["GENUINE_P28"] = math.NaN()
			if len(result.GenuinePercent) > 23 {
				actual["GENUINE_P28"] = result.GenuinePercent[23]
			}
			actual["AMPLITUDE_P29"] = math.NaN()
			if len(result.Spectrum) > 24 {
				actual["AMPLITUDE_P29"] = result.Spectrum[24]
			}
			actual["GENUINE_P29"] = math.NaN()
			if len(result.GenuinePercent) > 24 {
				actual["GENUINE_P29"] = result.GenuinePercent[24]
			}
			actual["AMPLITUDE_P30"] = math.NaN()
			if len(result.Spectrum) > 25 {
				actual["AMPLITUDE_P30"] = result.Spectrum[25]
			}
			actual["GENUINE_P30"] = math.NaN()
			if len(result.GenuinePercent) > 25 {
				actual["GENUINE_P30"] = result.GenuinePercent[25]
			}
			actual["AMPLITUDE_P31"] = math.NaN()
			if len(result.Spectrum) > 26 {
				actual["AMPLITUDE_P31"] = result.Spectrum[26]
			}
			actual["GENUINE_P31"] = math.NaN()
			if len(result.GenuinePercent) > 26 {
				actual["GENUINE_P31"] = result.GenuinePercent[26]
			}
			actual["AMPLITUDE_P32"] = math.NaN()
			if len(result.Spectrum) > 27 {
				actual["AMPLITUDE_P32"] = result.Spectrum[27]
			}
			actual["GENUINE_P32"] = math.NaN()
			if len(result.GenuinePercent) > 27 {
				actual["GENUINE_P32"] = result.GenuinePercent[27]
			}
			actual["AMPLITUDE_P33"] = math.NaN()
			if len(result.Spectrum) > 28 {
				actual["AMPLITUDE_P33"] = result.Spectrum[28]
			}
			actual["GENUINE_P33"] = math.NaN()
			if len(result.GenuinePercent) > 28 {
				actual["GENUINE_P33"] = result.GenuinePercent[28]
			}
			actual["AMPLITUDE_P34"] = math.NaN()
			if len(result.Spectrum) > 29 {
				actual["AMPLITUDE_P34"] = result.Spectrum[29]
			}
			actual["GENUINE_P34"] = math.NaN()
			if len(result.GenuinePercent) > 29 {
				actual["GENUINE_P34"] = result.GenuinePercent[29]
			}
			actual["AMPLITUDE_P35"] = math.NaN()
			if len(result.Spectrum) > 30 {
				actual["AMPLITUDE_P35"] = result.Spectrum[30]
			}
			actual["GENUINE_P35"] = math.NaN()
			if len(result.GenuinePercent) > 30 {
				actual["GENUINE_P35"] = result.GenuinePercent[30]
			}
			actual["AMPLITUDE_P36"] = math.NaN()
			if len(result.Spectrum) > 31 {
				actual["AMPLITUDE_P36"] = result.Spectrum[31]
			}
			actual["GENUINE_P36"] = math.NaN()
			if len(result.GenuinePercent) > 31 {
				actual["GENUINE_P36"] = result.GenuinePercent[31]
			}
			actual["AMPLITUDE_P37"] = math.NaN()
			if len(result.Spectrum) > 32 {
				actual["AMPLITUDE_P37"] = result.Spectrum[32]
			}
			actual["GENUINE_P37"] = math.NaN()
			if len(result.GenuinePercent) > 32 {
				actual["GENUINE_P37"] = result.GenuinePercent[32]
			}
			actual["AMPLITUDE_P38"] = math.NaN()
			if len(result.Spectrum) > 33 {
				actual["AMPLITUDE_P38"] = result.Spectrum[33]
			}
			actual["GENUINE_P38"] = math.NaN()
			if len(result.GenuinePercent) > 33 {
				actual["GENUINE_P38"] = result.GenuinePercent[33]
			}
			actual["AMPLITUDE_P39"] = math.NaN()
			if len(result.Spectrum) > 34 {
				actual["AMPLITUDE_P39"] = result.Spectrum[34]
			}
			actual["GENUINE_P39"] = math.NaN()
			if len(result.GenuinePercent) > 34 {
				actual["GENUINE_P39"] = result.GenuinePercent[34]
			}
			actual["AMPLITUDE_P40"] = math.NaN()
			if len(result.Spectrum) > 35 {
				actual["AMPLITUDE_P40"] = result.Spectrum[35]
			}
			actual["GENUINE_P40"] = math.NaN()
			if len(result.GenuinePercent) > 35 {
				actual["GENUINE_P40"] = result.GenuinePercent[35]
			}
			actual["AMPLITUDE_P41"] = math.NaN()
			if len(result.Spectrum) > 36 {
				actual["AMPLITUDE_P41"] = result.Spectrum[36]
			}
			actual["GENUINE_P41"] = math.NaN()
			if len(result.GenuinePercent) > 36 {
				actual["GENUINE_P41"] = result.GenuinePercent[36]
			}
			actual["AMPLITUDE_P42"] = math.NaN()
			if len(result.Spectrum) > 37 {
				actual["AMPLITUDE_P42"] = result.Spectrum[37]
			}
			actual["GENUINE_P42"] = math.NaN()
			if len(result.GenuinePercent) > 37 {
				actual["GENUINE_P42"] = result.GenuinePercent[37]
			}
			actual["AMPLITUDE_P43"] = math.NaN()
			if len(result.Spectrum) > 38 {
				actual["AMPLITUDE_P43"] = result.Spectrum[38]
			}
			actual["GENUINE_P43"] = math.NaN()
			if len(result.GenuinePercent) > 38 {
				actual["GENUINE_P43"] = result.GenuinePercent[38]
			}
			actual["AMPLITUDE_P44"] = math.NaN()
			if len(result.Spectrum) > 39 {
				actual["AMPLITUDE_P44"] = result.Spectrum[39]
			}
			actual["GENUINE_P44"] = math.NaN()
			if len(result.GenuinePercent) > 39 {
				actual["GENUINE_P44"] = result.GenuinePercent[39]
			}
			actual["AMPLITUDE_P45"] = math.NaN()
			if len(result.Spectrum) > 40 {
				actual["AMPLITUDE_P45"] = result.Spectrum[40]
			}
			actual["GENUINE_P45"] = math.NaN()
			if len(result.GenuinePercent) > 40 {
				actual["GENUINE_P45"] = result.GenuinePercent[40]
			}
			actual["AMPLITUDE_P46"] = math.NaN()
			if len(result.Spectrum) > 41 {
				actual["AMPLITUDE_P46"] = result.Spectrum[41]
			}
			actual["GENUINE_P46"] = math.NaN()
			if len(result.GenuinePercent) > 41 {
				actual["GENUINE_P46"] = result.GenuinePercent[41]
			}
			actual["AMPLITUDE_P47"] = math.NaN()
			if len(result.Spectrum) > 42 {
				actual["AMPLITUDE_P47"] = result.Spectrum[42]
			}
			actual["GENUINE_P47"] = math.NaN()
			if len(result.GenuinePercent) > 42 {
				actual["GENUINE_P47"] = result.GenuinePercent[42]
			}
			actual["AMPLITUDE_P48"] = math.NaN()
			if len(result.Spectrum) > 43 {
				actual["AMPLITUDE_P48"] = result.Spectrum[43]
			}
			actual["GENUINE_P48"] = math.NaN()
			if len(result.GenuinePercent) > 43 {
				actual["GENUINE_P48"] = result.GenuinePercent[43]
			}
			actual["AMPLITUDE_P49"] = math.NaN()
			if len(result.Spectrum) > 44 {
				actual["AMPLITUDE_P49"] = result.Spectrum[44]
			}
			actual["GENUINE_P49"] = math.NaN()
			if len(result.GenuinePercent) > 44 {
				actual["GENUINE_P49"] = result.GenuinePercent[44]
			}
			actual["AMPLITUDE_P50"] = math.NaN()
			if len(result.Spectrum) > 45 {
				actual["AMPLITUDE_P50"] = result.Spectrum[45]
			}
			actual["GENUINE_P50"] = math.NaN()
			if len(result.GenuinePercent) > 45 {
				actual["GENUINE_P50"] = result.GenuinePercent[45]
			}
			actual["AMPLITUDE_P51"] = math.NaN()
			if len(result.Spectrum) > 46 {
				actual["AMPLITUDE_P51"] = result.Spectrum[46]
			}
			actual["GENUINE_P51"] = math.NaN()
			if len(result.GenuinePercent) > 46 {
				actual["GENUINE_P51"] = result.GenuinePercent[46]
			}
			actual["AMPLITUDE_P52"] = math.NaN()
			if len(result.Spectrum) > 47 {
				actual["AMPLITUDE_P52"] = result.Spectrum[47]
			}
			actual["GENUINE_P52"] = math.NaN()
			if len(result.GenuinePercent) > 47 {
				actual["GENUINE_P52"] = result.GenuinePercent[47]
			}
			actual["AMPLITUDE_P53"] = math.NaN()
			if len(result.Spectrum) > 48 {
				actual["AMPLITUDE_P53"] = result.Spectrum[48]
			}
			actual["GENUINE_P53"] = math.NaN()
			if len(result.GenuinePercent) > 48 {
				actual["GENUINE_P53"] = result.GenuinePercent[48]
			}
			actual["AMPLITUDE_P54"] = math.NaN()
			if len(result.Spectrum) > 49 {
				actual["AMPLITUDE_P54"] = result.Spectrum[49]
			}
			actual["GENUINE_P54"] = math.NaN()
			if len(result.GenuinePercent) > 49 {
				actual["GENUINE_P54"] = result.GenuinePercent[49]
			}
			actual["AMPLITUDE_P55"] = math.NaN()
			if len(result.Spectrum) > 50 {
				actual["AMPLITUDE_P55"] = result.Spectrum[50]
			}
			actual["GENUINE_P55"] = math.NaN()
			if len(result.GenuinePercent) > 50 {
				actual["GENUINE_P55"] = result.GenuinePercent[50]
			}
			actual["AMPLITUDE_P56"] = math.NaN()
			if len(result.Spectrum) > 51 {
				actual["AMPLITUDE_P56"] = result.Spectrum[51]
			}
			actual["GENUINE_P56"] = math.NaN()
			if len(result.GenuinePercent) > 51 {
				actual["GENUINE_P56"] = result.GenuinePercent[51]
			}
			actual["AMPLITUDE_P57"] = math.NaN()
			if len(result.Spectrum) > 52 {
				actual["AMPLITUDE_P57"] = result.Spectrum[52]
			}
			actual["GENUINE_P57"] = math.NaN()
			if len(result.GenuinePercent) > 52 {
				actual["GENUINE_P57"] = result.GenuinePercent[52]
			}
			actual["AMPLITUDE_P58"] = math.NaN()
			if len(result.Spectrum) > 53 {
				actual["AMPLITUDE_P58"] = result.Spectrum[53]
			}
			actual["GENUINE_P58"] = math.NaN()
			if len(result.GenuinePercent) > 53 {
				actual["GENUINE_P58"] = result.GenuinePercent[53]
			}
			actual["AMPLITUDE_P59"] = math.NaN()
			if len(result.Spectrum) > 54 {
				actual["AMPLITUDE_P59"] = result.Spectrum[54]
			}
			actual["GENUINE_P59"] = math.NaN()
			if len(result.GenuinePercent) > 54 {
				actual["GENUINE_P59"] = result.GenuinePercent[54]
			}
			actual["AMPLITUDE_P60"] = math.NaN()
			if len(result.Spectrum) > 55 {
				actual["AMPLITUDE_P60"] = result.Spectrum[55]
			}
			actual["GENUINE_P60"] = math.NaN()
			if len(result.GenuinePercent) > 55 {
				actual["GENUINE_P60"] = result.GenuinePercent[55]
			}
			actual["AMPLITUDE_P61"] = math.NaN()
			if len(result.Spectrum) > 56 {
				actual["AMPLITUDE_P61"] = result.Spectrum[56]
			}
			actual["GENUINE_P61"] = math.NaN()
			if len(result.GenuinePercent) > 56 {
				actual["GENUINE_P61"] = result.GenuinePercent[56]
			}
			actual["AMPLITUDE_P62"] = math.NaN()
			if len(result.Spectrum) > 57 {
				actual["AMPLITUDE_P62"] = result.Spectrum[57]
			}
			actual["GENUINE_P62"] = math.NaN()
			if len(result.GenuinePercent) > 57 {
				actual["GENUINE_P62"] = result.GenuinePercent[57]
			}
			actual["AMPLITUDE_P63"] = math.NaN()
			if len(result.Spectrum) > 58 {
				actual["AMPLITUDE_P63"] = result.Spectrum[58]
			}
			actual["GENUINE_P63"] = math.NaN()
			if len(result.GenuinePercent) > 58 {
				actual["GENUINE_P63"] = result.GenuinePercent[58]
			}
			actual["AMPLITUDE_P64"] = math.NaN()
			if len(result.Spectrum) > 59 {
				actual["AMPLITUDE_P64"] = result.Spectrum[59]
			}
			actual["GENUINE_P64"] = math.NaN()
			if len(result.GenuinePercent) > 59 {
				actual["GENUINE_P64"] = result.GenuinePercent[59]
			}
			actual["AMPLITUDE_P65"] = math.NaN()
			if len(result.Spectrum) > 60 {
				actual["AMPLITUDE_P65"] = result.Spectrum[60]
			}
			actual["GENUINE_P65"] = math.NaN()
			if len(result.GenuinePercent) > 60 {
				actual["GENUINE_P65"] = result.GenuinePercent[60]
			}
			actual["AMPLITUDE_P66"] = math.NaN()
			if len(result.Spectrum) > 61 {
				actual["AMPLITUDE_P66"] = result.Spectrum[61]
			}
			actual["GENUINE_P66"] = math.NaN()
			if len(result.GenuinePercent) > 61 {
				actual["GENUINE_P66"] = result.GenuinePercent[61]
			}
			actual["AMPLITUDE_P67"] = math.NaN()
			if len(result.Spectrum) > 62 {
				actual["AMPLITUDE_P67"] = result.Spectrum[62]
			}
			actual["GENUINE_P67"] = math.NaN()
			if len(result.GenuinePercent) > 62 {
				actual["GENUINE_P67"] = result.GenuinePercent[62]
			}
			actual["AMPLITUDE_P68"] = math.NaN()
			if len(result.Spectrum) > 63 {
				actual["AMPLITUDE_P68"] = result.Spectrum[63]
			}
			actual["GENUINE_P68"] = math.NaN()
			if len(result.GenuinePercent) > 63 {
				actual["GENUINE_P68"] = result.GenuinePercent[63]
			}
			actual["AMPLITUDE_P69"] = math.NaN()
			if len(result.Spectrum) > 64 {
				actual["AMPLITUDE_P69"] = result.Spectrum[64]
			}
			actual["GENUINE_P69"] = math.NaN()
			if len(result.GenuinePercent) > 64 {
				actual["GENUINE_P69"] = result.GenuinePercent[64]
			}
			actual["AMPLITUDE_P70"] = math.NaN()
			if len(result.Spectrum) > 65 {
				actual["AMPLITUDE_P70"] = result.Spectrum[65]
			}
			actual["GENUINE_P70"] = math.NaN()
			if len(result.GenuinePercent) > 65 {
				actual["GENUINE_P70"] = result.GenuinePercent[65]
			}
			actual["AMPLITUDE_P71"] = math.NaN()
			if len(result.Spectrum) > 66 {
				actual["AMPLITUDE_P71"] = result.Spectrum[66]
			}
			actual["GENUINE_P71"] = math.NaN()
			if len(result.GenuinePercent) > 66 {
				actual["GENUINE_P71"] = result.GenuinePercent[66]
			}
			actual["AMPLITUDE_P72"] = math.NaN()
			if len(result.Spectrum) > 67 {
				actual["AMPLITUDE_P72"] = result.Spectrum[67]
			}
			actual["GENUINE_P72"] = math.NaN()
			if len(result.GenuinePercent) > 67 {
				actual["GENUINE_P72"] = result.GenuinePercent[67]
			}
			actual["AMPLITUDE_P73"] = math.NaN()
			if len(result.Spectrum) > 68 {
				actual["AMPLITUDE_P73"] = result.Spectrum[68]
			}
			actual["GENUINE_P73"] = math.NaN()
			if len(result.GenuinePercent) > 68 {
				actual["GENUINE_P73"] = result.GenuinePercent[68]
			}
			actual["AMPLITUDE_P74"] = math.NaN()
			if len(result.Spectrum) > 69 {
				actual["AMPLITUDE_P74"] = result.Spectrum[69]
			}
			actual["GENUINE_P74"] = math.NaN()
			if len(result.GenuinePercent) > 69 {
				actual["GENUINE_P74"] = result.GenuinePercent[69]
			}
			actual["AMPLITUDE_P75"] = math.NaN()
			if len(result.Spectrum) > 70 {
				actual["AMPLITUDE_P75"] = result.Spectrum[70]
			}
			actual["GENUINE_P75"] = math.NaN()
			if len(result.GenuinePercent) > 70 {
				actual["GENUINE_P75"] = result.GenuinePercent[70]
			}
			actual["AMPLITUDE_P76"] = math.NaN()
			if len(result.Spectrum) > 71 {
				actual["AMPLITUDE_P76"] = result.Spectrum[71]
			}
			actual["GENUINE_P76"] = math.NaN()
			if len(result.GenuinePercent) > 71 {
				actual["GENUINE_P76"] = result.GenuinePercent[71]
			}
			actual["AMPLITUDE_P77"] = math.NaN()
			if len(result.Spectrum) > 72 {
				actual["AMPLITUDE_P77"] = result.Spectrum[72]
			}
			actual["GENUINE_P77"] = math.NaN()
			if len(result.GenuinePercent) > 72 {
				actual["GENUINE_P77"] = result.GenuinePercent[72]
			}
			actual["AMPLITUDE_P78"] = math.NaN()
			if len(result.Spectrum) > 73 {
				actual["AMPLITUDE_P78"] = result.Spectrum[73]
			}
			actual["GENUINE_P78"] = math.NaN()
			if len(result.GenuinePercent) > 73 {
				actual["GENUINE_P78"] = result.GenuinePercent[73]
			}
			actual["AMPLITUDE_P79"] = math.NaN()
			if len(result.Spectrum) > 74 {
				actual["AMPLITUDE_P79"] = result.Spectrum[74]
			}
			actual["GENUINE_P79"] = math.NaN()
			if len(result.GenuinePercent) > 74 {
				actual["GENUINE_P79"] = result.GenuinePercent[74]
			}
			actual["AMPLITUDE_P80"] = math.NaN()
			if len(result.Spectrum) > 75 {
				actual["AMPLITUDE_P80"] = result.Spectrum[75]
			}
			actual["GENUINE_P80"] = math.NaN()
			if len(result.GenuinePercent) > 75 {
				actual["GENUINE_P80"] = result.GenuinePercent[75]
			}
			actual["RANK1_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PERIOD"] = result.Cycles[0].Period
			}
			actual["RANK1_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_AMPLITUDE"] = result.Cycles[0].Amplitude
			}
			actual["RANK1_PHASE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PHASE"] = result.Cycles[0].Phase
			}
			actual["RANK1_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_STRENGTH"] = result.Cycles[0].Strength
			}
			actual["RANK1_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_GENUINE_PERCENT"] = result.Cycles[0].GenuinePercent
			}
			actual["RANK1_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SEGMENTS"] = result.Cycles[0].Segments
			}
			actual["RANK1_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_AMPLITUDE"] = result.Cycles[0].FitAmplitude
			}
			actual["RANK1_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_SAMPLES"] = result.Cycles[0].FitSamples
			}
			actual["RANK1_VALUE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_VALUE"] = result.Cycles[0].Value
			}
			actual["RANK1_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SLOPE"] = result.Cycles[0].Slope
			}
			actual["RANK1_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_TOP"] = result.Cycles[0].EtaTop
			}
			actual["RANK1_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_BOTTOM"] = result.Cycles[0].EtaBottom
			}
			actual["RANK2_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_PERIOD"] = result.Cycles[1].Period
			}
			actual["RANK2_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_AMPLITUDE"] = result.Cycles[1].Amplitude
			}
			actual["RANK2_PHASE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_PHASE"] = result.Cycles[1].Phase
			}
			actual["RANK2_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_STRENGTH"] = result.Cycles[1].Strength
			}
			actual["RANK2_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_GENUINE_PERCENT"] = result.Cycles[1].GenuinePercent
			}
			actual["RANK2_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_SEGMENTS"] = result.Cycles[1].Segments
			}
			actual["RANK2_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_FIT_AMPLITUDE"] = result.Cycles[1].FitAmplitude
			}
			actual["RANK2_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_FIT_SAMPLES"] = result.Cycles[1].FitSamples
			}
			actual["RANK2_VALUE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_VALUE"] = result.Cycles[1].Value
			}
			actual["RANK2_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_SLOPE"] = result.Cycles[1].Slope
			}
			actual["RANK2_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_ETA_TOP"] = result.Cycles[1].EtaTop
			}
			actual["RANK2_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_ETA_BOTTOM"] = result.Cycles[1].EtaBottom
			}
			actual["RANK3_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_PERIOD"] = result.Cycles[2].Period
			}
			actual["RANK3_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_AMPLITUDE"] = result.Cycles[2].Amplitude
			}
			actual["RANK3_PHASE"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_PHASE"] = result.Cycles[2].Phase
			}
			actual["RANK3_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_STRENGTH"] = result.Cycles[2].Strength
			}
			actual["RANK3_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_GENUINE_PERCENT"] = result.Cycles[2].GenuinePercent
			}
			actual["RANK3_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_SEGMENTS"] = result.Cycles[2].Segments
			}
			actual["RANK3_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_FIT_AMPLITUDE"] = result.Cycles[2].FitAmplitude
			}
			actual["RANK3_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_FIT_SAMPLES"] = result.Cycles[2].FitSamples
			}
			actual["RANK3_VALUE"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_VALUE"] = result.Cycles[2].Value
			}
			actual["RANK3_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_SLOPE"] = result.Cycles[2].Slope
			}
			actual["RANK3_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_ETA_TOP"] = result.Cycles[2].EtaTop
			}
			actual["RANK3_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 3 {
				actual["RANK3_ETA_BOTTOM"] = result.Cycles[2].EtaBottom
			}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
	t.Run("case1", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 49.0
		p.MaxPeriod = 12
		p.MaxSegments = 5
		p.MinPeriod = 12
		p.PeakDistance = 2
		p.PhaseCycles = 0.0
		p.Smoothing = 1600.0
		p.TopK = 1
		p.Window = 60
		indicator, err := NewCompleteCycleScanner(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"ACCEPTED_COUNT": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ACCEPTED_COUNT, "AMPLITUDE_P12": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_AMPLITUDE_P12, "GENUINE_P12": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_GENUINE_P12, "RANK1_PERIOD": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_PERIOD, "RANK1_AMPLITUDE": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_AMPLITUDE, "RANK1_PHASE": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_PHASE, "RANK1_STRENGTH": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_STRENGTH, "RANK1_GENUINE_PERCENT": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_GENUINE_PERCENT, "RANK1_SEGMENTS": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_SEGMENTS, "RANK1_FIT_AMPLITUDE": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_FIT_AMPLITUDE, "RANK1_FIT_SAMPLES": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_FIT_SAMPLES, "RANK1_VALUE": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_VALUE, "RANK1_SLOPE": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_SLOPE, "RANK1_ETA_TOP": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_ETA_TOP, "RANK1_ETA_BOTTOM": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_RANK1_ETA_BOTTOM}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"ACCEPTED_COUNT": result.AcceptedCount}
			actual["AMPLITUDE_P12"] = math.NaN()
			if len(result.Spectrum) > 0 {
				actual["AMPLITUDE_P12"] = result.Spectrum[0]
			}
			actual["GENUINE_P12"] = math.NaN()
			if len(result.GenuinePercent) > 0 {
				actual["GENUINE_P12"] = result.GenuinePercent[0]
			}
			actual["RANK1_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PERIOD"] = result.Cycles[0].Period
			}
			actual["RANK1_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_AMPLITUDE"] = result.Cycles[0].Amplitude
			}
			actual["RANK1_PHASE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PHASE"] = result.Cycles[0].Phase
			}
			actual["RANK1_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_STRENGTH"] = result.Cycles[0].Strength
			}
			actual["RANK1_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_GENUINE_PERCENT"] = result.Cycles[0].GenuinePercent
			}
			actual["RANK1_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SEGMENTS"] = result.Cycles[0].Segments
			}
			actual["RANK1_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_AMPLITUDE"] = result.Cycles[0].FitAmplitude
			}
			actual["RANK1_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_SAMPLES"] = result.Cycles[0].FitSamples
			}
			actual["RANK1_VALUE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_VALUE"] = result.Cycles[0].Value
			}
			actual["RANK1_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SLOPE"] = result.Cycles[0].Slope
			}
			actual["RANK1_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_TOP"] = result.Cycles[0].EtaTop
			}
			actual["RANK1_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_BOTTOM"] = result.Cycles[0].EtaBottom
			}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
	t.Run("case2", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 49.0
		p.MaxPeriod = 30
		p.MaxSegments = 5
		p.MinPeriod = 5
		p.PeakDistance = 2
		p.PhaseCycles = 2.0
		p.Smoothing = 0.0
		p.TopK = 2
		p.Window = 100
		indicator, err := NewCompleteCycleScanner(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"ACCEPTED_COUNT": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ACCEPTED_COUNT, "AMPLITUDE_P5": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P5, "GENUINE_P5": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P5, "AMPLITUDE_P6": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P6, "GENUINE_P6": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P6, "AMPLITUDE_P7": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P7, "GENUINE_P7": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P7, "AMPLITUDE_P8": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P8, "GENUINE_P8": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P8, "AMPLITUDE_P9": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P9, "GENUINE_P9": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P9, "AMPLITUDE_P10": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P10, "GENUINE_P10": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P10, "AMPLITUDE_P11": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P11, "GENUINE_P11": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P11, "AMPLITUDE_P12": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P12, "GENUINE_P12": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P12, "AMPLITUDE_P13": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P13, "GENUINE_P13": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P13, "AMPLITUDE_P14": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P14, "GENUINE_P14": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P14, "AMPLITUDE_P15": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P15, "GENUINE_P15": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P15, "AMPLITUDE_P16": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P16, "GENUINE_P16": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P16, "AMPLITUDE_P17": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P17, "GENUINE_P17": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P17, "AMPLITUDE_P18": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P18, "GENUINE_P18": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P18, "AMPLITUDE_P19": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P19, "GENUINE_P19": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P19, "AMPLITUDE_P20": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P20, "GENUINE_P20": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P20, "AMPLITUDE_P21": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P21, "GENUINE_P21": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P21, "AMPLITUDE_P22": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P22, "GENUINE_P22": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P22, "AMPLITUDE_P23": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P23, "GENUINE_P23": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P23, "AMPLITUDE_P24": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P24, "GENUINE_P24": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P24, "AMPLITUDE_P25": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P25, "GENUINE_P25": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P25, "AMPLITUDE_P26": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P26, "GENUINE_P26": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P26, "AMPLITUDE_P27": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P27, "GENUINE_P27": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P27, "AMPLITUDE_P28": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P28, "GENUINE_P28": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P28, "AMPLITUDE_P29": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P29, "GENUINE_P29": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P29, "AMPLITUDE_P30": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE_P30, "GENUINE_P30": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_P30, "RANK1_PERIOD": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_PERIOD, "RANK1_AMPLITUDE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_AMPLITUDE, "RANK1_PHASE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_PHASE, "RANK1_STRENGTH": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_STRENGTH, "RANK1_GENUINE_PERCENT": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_GENUINE_PERCENT, "RANK1_SEGMENTS": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_SEGMENTS, "RANK1_FIT_AMPLITUDE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_FIT_AMPLITUDE, "RANK1_FIT_SAMPLES": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_FIT_SAMPLES, "RANK1_VALUE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_VALUE, "RANK1_SLOPE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_SLOPE, "RANK1_ETA_TOP": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_ETA_TOP, "RANK1_ETA_BOTTOM": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK1_ETA_BOTTOM, "RANK2_PERIOD": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_PERIOD, "RANK2_AMPLITUDE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_AMPLITUDE, "RANK2_PHASE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_PHASE, "RANK2_STRENGTH": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_STRENGTH, "RANK2_GENUINE_PERCENT": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_GENUINE_PERCENT, "RANK2_SEGMENTS": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_SEGMENTS, "RANK2_FIT_AMPLITUDE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_FIT_AMPLITUDE, "RANK2_FIT_SAMPLES": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_FIT_SAMPLES, "RANK2_VALUE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_VALUE, "RANK2_SLOPE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_SLOPE, "RANK2_ETA_TOP": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_ETA_TOP, "RANK2_ETA_BOTTOM": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_RANK2_ETA_BOTTOM}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"ACCEPTED_COUNT": result.AcceptedCount}
			actual["AMPLITUDE_P5"] = math.NaN()
			if len(result.Spectrum) > 0 {
				actual["AMPLITUDE_P5"] = result.Spectrum[0]
			}
			actual["GENUINE_P5"] = math.NaN()
			if len(result.GenuinePercent) > 0 {
				actual["GENUINE_P5"] = result.GenuinePercent[0]
			}
			actual["AMPLITUDE_P6"] = math.NaN()
			if len(result.Spectrum) > 1 {
				actual["AMPLITUDE_P6"] = result.Spectrum[1]
			}
			actual["GENUINE_P6"] = math.NaN()
			if len(result.GenuinePercent) > 1 {
				actual["GENUINE_P6"] = result.GenuinePercent[1]
			}
			actual["AMPLITUDE_P7"] = math.NaN()
			if len(result.Spectrum) > 2 {
				actual["AMPLITUDE_P7"] = result.Spectrum[2]
			}
			actual["GENUINE_P7"] = math.NaN()
			if len(result.GenuinePercent) > 2 {
				actual["GENUINE_P7"] = result.GenuinePercent[2]
			}
			actual["AMPLITUDE_P8"] = math.NaN()
			if len(result.Spectrum) > 3 {
				actual["AMPLITUDE_P8"] = result.Spectrum[3]
			}
			actual["GENUINE_P8"] = math.NaN()
			if len(result.GenuinePercent) > 3 {
				actual["GENUINE_P8"] = result.GenuinePercent[3]
			}
			actual["AMPLITUDE_P9"] = math.NaN()
			if len(result.Spectrum) > 4 {
				actual["AMPLITUDE_P9"] = result.Spectrum[4]
			}
			actual["GENUINE_P9"] = math.NaN()
			if len(result.GenuinePercent) > 4 {
				actual["GENUINE_P9"] = result.GenuinePercent[4]
			}
			actual["AMPLITUDE_P10"] = math.NaN()
			if len(result.Spectrum) > 5 {
				actual["AMPLITUDE_P10"] = result.Spectrum[5]
			}
			actual["GENUINE_P10"] = math.NaN()
			if len(result.GenuinePercent) > 5 {
				actual["GENUINE_P10"] = result.GenuinePercent[5]
			}
			actual["AMPLITUDE_P11"] = math.NaN()
			if len(result.Spectrum) > 6 {
				actual["AMPLITUDE_P11"] = result.Spectrum[6]
			}
			actual["GENUINE_P11"] = math.NaN()
			if len(result.GenuinePercent) > 6 {
				actual["GENUINE_P11"] = result.GenuinePercent[6]
			}
			actual["AMPLITUDE_P12"] = math.NaN()
			if len(result.Spectrum) > 7 {
				actual["AMPLITUDE_P12"] = result.Spectrum[7]
			}
			actual["GENUINE_P12"] = math.NaN()
			if len(result.GenuinePercent) > 7 {
				actual["GENUINE_P12"] = result.GenuinePercent[7]
			}
			actual["AMPLITUDE_P13"] = math.NaN()
			if len(result.Spectrum) > 8 {
				actual["AMPLITUDE_P13"] = result.Spectrum[8]
			}
			actual["GENUINE_P13"] = math.NaN()
			if len(result.GenuinePercent) > 8 {
				actual["GENUINE_P13"] = result.GenuinePercent[8]
			}
			actual["AMPLITUDE_P14"] = math.NaN()
			if len(result.Spectrum) > 9 {
				actual["AMPLITUDE_P14"] = result.Spectrum[9]
			}
			actual["GENUINE_P14"] = math.NaN()
			if len(result.GenuinePercent) > 9 {
				actual["GENUINE_P14"] = result.GenuinePercent[9]
			}
			actual["AMPLITUDE_P15"] = math.NaN()
			if len(result.Spectrum) > 10 {
				actual["AMPLITUDE_P15"] = result.Spectrum[10]
			}
			actual["GENUINE_P15"] = math.NaN()
			if len(result.GenuinePercent) > 10 {
				actual["GENUINE_P15"] = result.GenuinePercent[10]
			}
			actual["AMPLITUDE_P16"] = math.NaN()
			if len(result.Spectrum) > 11 {
				actual["AMPLITUDE_P16"] = result.Spectrum[11]
			}
			actual["GENUINE_P16"] = math.NaN()
			if len(result.GenuinePercent) > 11 {
				actual["GENUINE_P16"] = result.GenuinePercent[11]
			}
			actual["AMPLITUDE_P17"] = math.NaN()
			if len(result.Spectrum) > 12 {
				actual["AMPLITUDE_P17"] = result.Spectrum[12]
			}
			actual["GENUINE_P17"] = math.NaN()
			if len(result.GenuinePercent) > 12 {
				actual["GENUINE_P17"] = result.GenuinePercent[12]
			}
			actual["AMPLITUDE_P18"] = math.NaN()
			if len(result.Spectrum) > 13 {
				actual["AMPLITUDE_P18"] = result.Spectrum[13]
			}
			actual["GENUINE_P18"] = math.NaN()
			if len(result.GenuinePercent) > 13 {
				actual["GENUINE_P18"] = result.GenuinePercent[13]
			}
			actual["AMPLITUDE_P19"] = math.NaN()
			if len(result.Spectrum) > 14 {
				actual["AMPLITUDE_P19"] = result.Spectrum[14]
			}
			actual["GENUINE_P19"] = math.NaN()
			if len(result.GenuinePercent) > 14 {
				actual["GENUINE_P19"] = result.GenuinePercent[14]
			}
			actual["AMPLITUDE_P20"] = math.NaN()
			if len(result.Spectrum) > 15 {
				actual["AMPLITUDE_P20"] = result.Spectrum[15]
			}
			actual["GENUINE_P20"] = math.NaN()
			if len(result.GenuinePercent) > 15 {
				actual["GENUINE_P20"] = result.GenuinePercent[15]
			}
			actual["AMPLITUDE_P21"] = math.NaN()
			if len(result.Spectrum) > 16 {
				actual["AMPLITUDE_P21"] = result.Spectrum[16]
			}
			actual["GENUINE_P21"] = math.NaN()
			if len(result.GenuinePercent) > 16 {
				actual["GENUINE_P21"] = result.GenuinePercent[16]
			}
			actual["AMPLITUDE_P22"] = math.NaN()
			if len(result.Spectrum) > 17 {
				actual["AMPLITUDE_P22"] = result.Spectrum[17]
			}
			actual["GENUINE_P22"] = math.NaN()
			if len(result.GenuinePercent) > 17 {
				actual["GENUINE_P22"] = result.GenuinePercent[17]
			}
			actual["AMPLITUDE_P23"] = math.NaN()
			if len(result.Spectrum) > 18 {
				actual["AMPLITUDE_P23"] = result.Spectrum[18]
			}
			actual["GENUINE_P23"] = math.NaN()
			if len(result.GenuinePercent) > 18 {
				actual["GENUINE_P23"] = result.GenuinePercent[18]
			}
			actual["AMPLITUDE_P24"] = math.NaN()
			if len(result.Spectrum) > 19 {
				actual["AMPLITUDE_P24"] = result.Spectrum[19]
			}
			actual["GENUINE_P24"] = math.NaN()
			if len(result.GenuinePercent) > 19 {
				actual["GENUINE_P24"] = result.GenuinePercent[19]
			}
			actual["AMPLITUDE_P25"] = math.NaN()
			if len(result.Spectrum) > 20 {
				actual["AMPLITUDE_P25"] = result.Spectrum[20]
			}
			actual["GENUINE_P25"] = math.NaN()
			if len(result.GenuinePercent) > 20 {
				actual["GENUINE_P25"] = result.GenuinePercent[20]
			}
			actual["AMPLITUDE_P26"] = math.NaN()
			if len(result.Spectrum) > 21 {
				actual["AMPLITUDE_P26"] = result.Spectrum[21]
			}
			actual["GENUINE_P26"] = math.NaN()
			if len(result.GenuinePercent) > 21 {
				actual["GENUINE_P26"] = result.GenuinePercent[21]
			}
			actual["AMPLITUDE_P27"] = math.NaN()
			if len(result.Spectrum) > 22 {
				actual["AMPLITUDE_P27"] = result.Spectrum[22]
			}
			actual["GENUINE_P27"] = math.NaN()
			if len(result.GenuinePercent) > 22 {
				actual["GENUINE_P27"] = result.GenuinePercent[22]
			}
			actual["AMPLITUDE_P28"] = math.NaN()
			if len(result.Spectrum) > 23 {
				actual["AMPLITUDE_P28"] = result.Spectrum[23]
			}
			actual["GENUINE_P28"] = math.NaN()
			if len(result.GenuinePercent) > 23 {
				actual["GENUINE_P28"] = result.GenuinePercent[23]
			}
			actual["AMPLITUDE_P29"] = math.NaN()
			if len(result.Spectrum) > 24 {
				actual["AMPLITUDE_P29"] = result.Spectrum[24]
			}
			actual["GENUINE_P29"] = math.NaN()
			if len(result.GenuinePercent) > 24 {
				actual["GENUINE_P29"] = result.GenuinePercent[24]
			}
			actual["AMPLITUDE_P30"] = math.NaN()
			if len(result.Spectrum) > 25 {
				actual["AMPLITUDE_P30"] = result.Spectrum[25]
			}
			actual["GENUINE_P30"] = math.NaN()
			if len(result.GenuinePercent) > 25 {
				actual["GENUINE_P30"] = result.GenuinePercent[25]
			}
			actual["RANK1_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PERIOD"] = result.Cycles[0].Period
			}
			actual["RANK1_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_AMPLITUDE"] = result.Cycles[0].Amplitude
			}
			actual["RANK1_PHASE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PHASE"] = result.Cycles[0].Phase
			}
			actual["RANK1_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_STRENGTH"] = result.Cycles[0].Strength
			}
			actual["RANK1_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_GENUINE_PERCENT"] = result.Cycles[0].GenuinePercent
			}
			actual["RANK1_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SEGMENTS"] = result.Cycles[0].Segments
			}
			actual["RANK1_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_AMPLITUDE"] = result.Cycles[0].FitAmplitude
			}
			actual["RANK1_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_SAMPLES"] = result.Cycles[0].FitSamples
			}
			actual["RANK1_VALUE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_VALUE"] = result.Cycles[0].Value
			}
			actual["RANK1_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SLOPE"] = result.Cycles[0].Slope
			}
			actual["RANK1_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_TOP"] = result.Cycles[0].EtaTop
			}
			actual["RANK1_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_BOTTOM"] = result.Cycles[0].EtaBottom
			}
			actual["RANK2_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_PERIOD"] = result.Cycles[1].Period
			}
			actual["RANK2_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_AMPLITUDE"] = result.Cycles[1].Amplitude
			}
			actual["RANK2_PHASE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_PHASE"] = result.Cycles[1].Phase
			}
			actual["RANK2_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_STRENGTH"] = result.Cycles[1].Strength
			}
			actual["RANK2_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_GENUINE_PERCENT"] = result.Cycles[1].GenuinePercent
			}
			actual["RANK2_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_SEGMENTS"] = result.Cycles[1].Segments
			}
			actual["RANK2_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_FIT_AMPLITUDE"] = result.Cycles[1].FitAmplitude
			}
			actual["RANK2_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_FIT_SAMPLES"] = result.Cycles[1].FitSamples
			}
			actual["RANK2_VALUE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_VALUE"] = result.Cycles[1].Value
			}
			actual["RANK2_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_SLOPE"] = result.Cycles[1].Slope
			}
			actual["RANK2_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_ETA_TOP"] = result.Cycles[1].EtaTop
			}
			actual["RANK2_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_ETA_BOTTOM"] = result.Cycles[1].EtaBottom
			}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
	t.Run("case3", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 100.0
		p.MaxPeriod = 30
		p.MaxSegments = 5
		p.MinPeriod = 5
		p.PeakDistance = 2
		p.PhaseCycles = 2.0
		p.Smoothing = 1600.0
		p.TopK = 2
		p.Window = 100
		indicator, err := NewCompleteCycleScanner(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"ACCEPTED_COUNT": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ACCEPTED_COUNT, "AMPLITUDE_P5": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P5, "GENUINE_P5": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P5, "AMPLITUDE_P6": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P6, "GENUINE_P6": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P6, "AMPLITUDE_P7": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P7, "GENUINE_P7": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P7, "AMPLITUDE_P8": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P8, "GENUINE_P8": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P8, "AMPLITUDE_P9": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P9, "GENUINE_P9": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P9, "AMPLITUDE_P10": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P10, "GENUINE_P10": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P10, "AMPLITUDE_P11": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P11, "GENUINE_P11": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P11, "AMPLITUDE_P12": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P12, "GENUINE_P12": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P12, "AMPLITUDE_P13": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P13, "GENUINE_P13": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P13, "AMPLITUDE_P14": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P14, "GENUINE_P14": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P14, "AMPLITUDE_P15": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P15, "GENUINE_P15": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P15, "AMPLITUDE_P16": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P16, "GENUINE_P16": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P16, "AMPLITUDE_P17": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P17, "GENUINE_P17": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P17, "AMPLITUDE_P18": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P18, "GENUINE_P18": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P18, "AMPLITUDE_P19": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P19, "GENUINE_P19": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P19, "AMPLITUDE_P20": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P20, "GENUINE_P20": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P20, "AMPLITUDE_P21": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P21, "GENUINE_P21": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P21, "AMPLITUDE_P22": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P22, "GENUINE_P22": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P22, "AMPLITUDE_P23": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P23, "GENUINE_P23": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P23, "AMPLITUDE_P24": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P24, "GENUINE_P24": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P24, "AMPLITUDE_P25": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P25, "GENUINE_P25": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P25, "AMPLITUDE_P26": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P26, "GENUINE_P26": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P26, "AMPLITUDE_P27": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P27, "GENUINE_P27": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P27, "AMPLITUDE_P28": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P28, "GENUINE_P28": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P28, "AMPLITUDE_P29": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P29, "GENUINE_P29": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P29, "AMPLITUDE_P30": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE_P30, "GENUINE_P30": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_P30, "RANK1_PERIOD": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_PERIOD, "RANK1_AMPLITUDE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_AMPLITUDE, "RANK1_PHASE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_PHASE, "RANK1_STRENGTH": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_STRENGTH, "RANK1_GENUINE_PERCENT": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_GENUINE_PERCENT, "RANK1_SEGMENTS": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_SEGMENTS, "RANK1_FIT_AMPLITUDE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_FIT_AMPLITUDE, "RANK1_FIT_SAMPLES": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_FIT_SAMPLES, "RANK1_VALUE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_VALUE, "RANK1_SLOPE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_SLOPE, "RANK1_ETA_TOP": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_ETA_TOP, "RANK1_ETA_BOTTOM": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK1_ETA_BOTTOM, "RANK2_PERIOD": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_PERIOD, "RANK2_AMPLITUDE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_AMPLITUDE, "RANK2_PHASE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_PHASE, "RANK2_STRENGTH": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_STRENGTH, "RANK2_GENUINE_PERCENT": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_GENUINE_PERCENT, "RANK2_SEGMENTS": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_SEGMENTS, "RANK2_FIT_AMPLITUDE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_FIT_AMPLITUDE, "RANK2_FIT_SAMPLES": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_FIT_SAMPLES, "RANK2_VALUE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_VALUE, "RANK2_SLOPE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_SLOPE, "RANK2_ETA_TOP": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_ETA_TOP, "RANK2_ETA_BOTTOM": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_RANK2_ETA_BOTTOM}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"ACCEPTED_COUNT": result.AcceptedCount}
			actual["AMPLITUDE_P5"] = math.NaN()
			if len(result.Spectrum) > 0 {
				actual["AMPLITUDE_P5"] = result.Spectrum[0]
			}
			actual["GENUINE_P5"] = math.NaN()
			if len(result.GenuinePercent) > 0 {
				actual["GENUINE_P5"] = result.GenuinePercent[0]
			}
			actual["AMPLITUDE_P6"] = math.NaN()
			if len(result.Spectrum) > 1 {
				actual["AMPLITUDE_P6"] = result.Spectrum[1]
			}
			actual["GENUINE_P6"] = math.NaN()
			if len(result.GenuinePercent) > 1 {
				actual["GENUINE_P6"] = result.GenuinePercent[1]
			}
			actual["AMPLITUDE_P7"] = math.NaN()
			if len(result.Spectrum) > 2 {
				actual["AMPLITUDE_P7"] = result.Spectrum[2]
			}
			actual["GENUINE_P7"] = math.NaN()
			if len(result.GenuinePercent) > 2 {
				actual["GENUINE_P7"] = result.GenuinePercent[2]
			}
			actual["AMPLITUDE_P8"] = math.NaN()
			if len(result.Spectrum) > 3 {
				actual["AMPLITUDE_P8"] = result.Spectrum[3]
			}
			actual["GENUINE_P8"] = math.NaN()
			if len(result.GenuinePercent) > 3 {
				actual["GENUINE_P8"] = result.GenuinePercent[3]
			}
			actual["AMPLITUDE_P9"] = math.NaN()
			if len(result.Spectrum) > 4 {
				actual["AMPLITUDE_P9"] = result.Spectrum[4]
			}
			actual["GENUINE_P9"] = math.NaN()
			if len(result.GenuinePercent) > 4 {
				actual["GENUINE_P9"] = result.GenuinePercent[4]
			}
			actual["AMPLITUDE_P10"] = math.NaN()
			if len(result.Spectrum) > 5 {
				actual["AMPLITUDE_P10"] = result.Spectrum[5]
			}
			actual["GENUINE_P10"] = math.NaN()
			if len(result.GenuinePercent) > 5 {
				actual["GENUINE_P10"] = result.GenuinePercent[5]
			}
			actual["AMPLITUDE_P11"] = math.NaN()
			if len(result.Spectrum) > 6 {
				actual["AMPLITUDE_P11"] = result.Spectrum[6]
			}
			actual["GENUINE_P11"] = math.NaN()
			if len(result.GenuinePercent) > 6 {
				actual["GENUINE_P11"] = result.GenuinePercent[6]
			}
			actual["AMPLITUDE_P12"] = math.NaN()
			if len(result.Spectrum) > 7 {
				actual["AMPLITUDE_P12"] = result.Spectrum[7]
			}
			actual["GENUINE_P12"] = math.NaN()
			if len(result.GenuinePercent) > 7 {
				actual["GENUINE_P12"] = result.GenuinePercent[7]
			}
			actual["AMPLITUDE_P13"] = math.NaN()
			if len(result.Spectrum) > 8 {
				actual["AMPLITUDE_P13"] = result.Spectrum[8]
			}
			actual["GENUINE_P13"] = math.NaN()
			if len(result.GenuinePercent) > 8 {
				actual["GENUINE_P13"] = result.GenuinePercent[8]
			}
			actual["AMPLITUDE_P14"] = math.NaN()
			if len(result.Spectrum) > 9 {
				actual["AMPLITUDE_P14"] = result.Spectrum[9]
			}
			actual["GENUINE_P14"] = math.NaN()
			if len(result.GenuinePercent) > 9 {
				actual["GENUINE_P14"] = result.GenuinePercent[9]
			}
			actual["AMPLITUDE_P15"] = math.NaN()
			if len(result.Spectrum) > 10 {
				actual["AMPLITUDE_P15"] = result.Spectrum[10]
			}
			actual["GENUINE_P15"] = math.NaN()
			if len(result.GenuinePercent) > 10 {
				actual["GENUINE_P15"] = result.GenuinePercent[10]
			}
			actual["AMPLITUDE_P16"] = math.NaN()
			if len(result.Spectrum) > 11 {
				actual["AMPLITUDE_P16"] = result.Spectrum[11]
			}
			actual["GENUINE_P16"] = math.NaN()
			if len(result.GenuinePercent) > 11 {
				actual["GENUINE_P16"] = result.GenuinePercent[11]
			}
			actual["AMPLITUDE_P17"] = math.NaN()
			if len(result.Spectrum) > 12 {
				actual["AMPLITUDE_P17"] = result.Spectrum[12]
			}
			actual["GENUINE_P17"] = math.NaN()
			if len(result.GenuinePercent) > 12 {
				actual["GENUINE_P17"] = result.GenuinePercent[12]
			}
			actual["AMPLITUDE_P18"] = math.NaN()
			if len(result.Spectrum) > 13 {
				actual["AMPLITUDE_P18"] = result.Spectrum[13]
			}
			actual["GENUINE_P18"] = math.NaN()
			if len(result.GenuinePercent) > 13 {
				actual["GENUINE_P18"] = result.GenuinePercent[13]
			}
			actual["AMPLITUDE_P19"] = math.NaN()
			if len(result.Spectrum) > 14 {
				actual["AMPLITUDE_P19"] = result.Spectrum[14]
			}
			actual["GENUINE_P19"] = math.NaN()
			if len(result.GenuinePercent) > 14 {
				actual["GENUINE_P19"] = result.GenuinePercent[14]
			}
			actual["AMPLITUDE_P20"] = math.NaN()
			if len(result.Spectrum) > 15 {
				actual["AMPLITUDE_P20"] = result.Spectrum[15]
			}
			actual["GENUINE_P20"] = math.NaN()
			if len(result.GenuinePercent) > 15 {
				actual["GENUINE_P20"] = result.GenuinePercent[15]
			}
			actual["AMPLITUDE_P21"] = math.NaN()
			if len(result.Spectrum) > 16 {
				actual["AMPLITUDE_P21"] = result.Spectrum[16]
			}
			actual["GENUINE_P21"] = math.NaN()
			if len(result.GenuinePercent) > 16 {
				actual["GENUINE_P21"] = result.GenuinePercent[16]
			}
			actual["AMPLITUDE_P22"] = math.NaN()
			if len(result.Spectrum) > 17 {
				actual["AMPLITUDE_P22"] = result.Spectrum[17]
			}
			actual["GENUINE_P22"] = math.NaN()
			if len(result.GenuinePercent) > 17 {
				actual["GENUINE_P22"] = result.GenuinePercent[17]
			}
			actual["AMPLITUDE_P23"] = math.NaN()
			if len(result.Spectrum) > 18 {
				actual["AMPLITUDE_P23"] = result.Spectrum[18]
			}
			actual["GENUINE_P23"] = math.NaN()
			if len(result.GenuinePercent) > 18 {
				actual["GENUINE_P23"] = result.GenuinePercent[18]
			}
			actual["AMPLITUDE_P24"] = math.NaN()
			if len(result.Spectrum) > 19 {
				actual["AMPLITUDE_P24"] = result.Spectrum[19]
			}
			actual["GENUINE_P24"] = math.NaN()
			if len(result.GenuinePercent) > 19 {
				actual["GENUINE_P24"] = result.GenuinePercent[19]
			}
			actual["AMPLITUDE_P25"] = math.NaN()
			if len(result.Spectrum) > 20 {
				actual["AMPLITUDE_P25"] = result.Spectrum[20]
			}
			actual["GENUINE_P25"] = math.NaN()
			if len(result.GenuinePercent) > 20 {
				actual["GENUINE_P25"] = result.GenuinePercent[20]
			}
			actual["AMPLITUDE_P26"] = math.NaN()
			if len(result.Spectrum) > 21 {
				actual["AMPLITUDE_P26"] = result.Spectrum[21]
			}
			actual["GENUINE_P26"] = math.NaN()
			if len(result.GenuinePercent) > 21 {
				actual["GENUINE_P26"] = result.GenuinePercent[21]
			}
			actual["AMPLITUDE_P27"] = math.NaN()
			if len(result.Spectrum) > 22 {
				actual["AMPLITUDE_P27"] = result.Spectrum[22]
			}
			actual["GENUINE_P27"] = math.NaN()
			if len(result.GenuinePercent) > 22 {
				actual["GENUINE_P27"] = result.GenuinePercent[22]
			}
			actual["AMPLITUDE_P28"] = math.NaN()
			if len(result.Spectrum) > 23 {
				actual["AMPLITUDE_P28"] = result.Spectrum[23]
			}
			actual["GENUINE_P28"] = math.NaN()
			if len(result.GenuinePercent) > 23 {
				actual["GENUINE_P28"] = result.GenuinePercent[23]
			}
			actual["AMPLITUDE_P29"] = math.NaN()
			if len(result.Spectrum) > 24 {
				actual["AMPLITUDE_P29"] = result.Spectrum[24]
			}
			actual["GENUINE_P29"] = math.NaN()
			if len(result.GenuinePercent) > 24 {
				actual["GENUINE_P29"] = result.GenuinePercent[24]
			}
			actual["AMPLITUDE_P30"] = math.NaN()
			if len(result.Spectrum) > 25 {
				actual["AMPLITUDE_P30"] = result.Spectrum[25]
			}
			actual["GENUINE_P30"] = math.NaN()
			if len(result.GenuinePercent) > 25 {
				actual["GENUINE_P30"] = result.GenuinePercent[25]
			}
			actual["RANK1_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PERIOD"] = result.Cycles[0].Period
			}
			actual["RANK1_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_AMPLITUDE"] = result.Cycles[0].Amplitude
			}
			actual["RANK1_PHASE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_PHASE"] = result.Cycles[0].Phase
			}
			actual["RANK1_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_STRENGTH"] = result.Cycles[0].Strength
			}
			actual["RANK1_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_GENUINE_PERCENT"] = result.Cycles[0].GenuinePercent
			}
			actual["RANK1_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SEGMENTS"] = result.Cycles[0].Segments
			}
			actual["RANK1_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_AMPLITUDE"] = result.Cycles[0].FitAmplitude
			}
			actual["RANK1_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_FIT_SAMPLES"] = result.Cycles[0].FitSamples
			}
			actual["RANK1_VALUE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_VALUE"] = result.Cycles[0].Value
			}
			actual["RANK1_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_SLOPE"] = result.Cycles[0].Slope
			}
			actual["RANK1_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_TOP"] = result.Cycles[0].EtaTop
			}
			actual["RANK1_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 1 {
				actual["RANK1_ETA_BOTTOM"] = result.Cycles[0].EtaBottom
			}
			actual["RANK2_PERIOD"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_PERIOD"] = result.Cycles[1].Period
			}
			actual["RANK2_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_AMPLITUDE"] = result.Cycles[1].Amplitude
			}
			actual["RANK2_PHASE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_PHASE"] = result.Cycles[1].Phase
			}
			actual["RANK2_STRENGTH"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_STRENGTH"] = result.Cycles[1].Strength
			}
			actual["RANK2_GENUINE_PERCENT"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_GENUINE_PERCENT"] = result.Cycles[1].GenuinePercent
			}
			actual["RANK2_SEGMENTS"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_SEGMENTS"] = result.Cycles[1].Segments
			}
			actual["RANK2_FIT_AMPLITUDE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_FIT_AMPLITUDE"] = result.Cycles[1].FitAmplitude
			}
			actual["RANK2_FIT_SAMPLES"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_FIT_SAMPLES"] = result.Cycles[1].FitSamples
			}
			actual["RANK2_VALUE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_VALUE"] = result.Cycles[1].Value
			}
			actual["RANK2_SLOPE"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_SLOPE"] = result.Cycles[1].Slope
			}
			actual["RANK2_ETA_TOP"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_ETA_TOP"] = result.Cycles[1].EtaTop
			}
			actual["RANK2_ETA_BOTTOM"] = math.NaN()
			if len(result.Cycles) >= 2 {
				actual["RANK2_ETA_BOTTOM"] = result.Cycles[1].EtaBottom
			}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
}
