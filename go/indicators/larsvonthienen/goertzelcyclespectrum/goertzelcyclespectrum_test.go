// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package goertzelcyclespectrum

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 128
		p.MinPeriod = 5
		p.MaxPeriod = 50
		p.Step = 1
		p.Smoothing = 1600.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A5": expectedW128_MIN5_MAX50_S1_L1600_A5, "PH5": expectedW128_MIN5_MAX50_S1_L1600_PH5, "TO5": expectedW128_MIN5_MAX50_S1_L1600_TO5, "A6": expectedW128_MIN5_MAX50_S1_L1600_A6, "PH6": expectedW128_MIN5_MAX50_S1_L1600_PH6, "TO6": expectedW128_MIN5_MAX50_S1_L1600_TO6, "A7": expectedW128_MIN5_MAX50_S1_L1600_A7, "PH7": expectedW128_MIN5_MAX50_S1_L1600_PH7, "TO7": expectedW128_MIN5_MAX50_S1_L1600_TO7, "A8": expectedW128_MIN5_MAX50_S1_L1600_A8, "PH8": expectedW128_MIN5_MAX50_S1_L1600_PH8, "TO8": expectedW128_MIN5_MAX50_S1_L1600_TO8, "A9": expectedW128_MIN5_MAX50_S1_L1600_A9, "PH9": expectedW128_MIN5_MAX50_S1_L1600_PH9, "TO9": expectedW128_MIN5_MAX50_S1_L1600_TO9, "A10": expectedW128_MIN5_MAX50_S1_L1600_A10, "PH10": expectedW128_MIN5_MAX50_S1_L1600_PH10, "TO10": expectedW128_MIN5_MAX50_S1_L1600_TO10, "A11": expectedW128_MIN5_MAX50_S1_L1600_A11, "PH11": expectedW128_MIN5_MAX50_S1_L1600_PH11, "TO11": expectedW128_MIN5_MAX50_S1_L1600_TO11, "A12": expectedW128_MIN5_MAX50_S1_L1600_A12, "PH12": expectedW128_MIN5_MAX50_S1_L1600_PH12, "TO12": expectedW128_MIN5_MAX50_S1_L1600_TO12, "A13": expectedW128_MIN5_MAX50_S1_L1600_A13, "PH13": expectedW128_MIN5_MAX50_S1_L1600_PH13, "TO13": expectedW128_MIN5_MAX50_S1_L1600_TO13, "A14": expectedW128_MIN5_MAX50_S1_L1600_A14, "PH14": expectedW128_MIN5_MAX50_S1_L1600_PH14, "TO14": expectedW128_MIN5_MAX50_S1_L1600_TO14, "A15": expectedW128_MIN5_MAX50_S1_L1600_A15, "PH15": expectedW128_MIN5_MAX50_S1_L1600_PH15, "TO15": expectedW128_MIN5_MAX50_S1_L1600_TO15, "A16": expectedW128_MIN5_MAX50_S1_L1600_A16, "PH16": expectedW128_MIN5_MAX50_S1_L1600_PH16, "TO16": expectedW128_MIN5_MAX50_S1_L1600_TO16, "A17": expectedW128_MIN5_MAX50_S1_L1600_A17, "PH17": expectedW128_MIN5_MAX50_S1_L1600_PH17, "TO17": expectedW128_MIN5_MAX50_S1_L1600_TO17, "A18": expectedW128_MIN5_MAX50_S1_L1600_A18, "PH18": expectedW128_MIN5_MAX50_S1_L1600_PH18, "TO18": expectedW128_MIN5_MAX50_S1_L1600_TO18, "A19": expectedW128_MIN5_MAX50_S1_L1600_A19, "PH19": expectedW128_MIN5_MAX50_S1_L1600_PH19, "TO19": expectedW128_MIN5_MAX50_S1_L1600_TO19, "A20": expectedW128_MIN5_MAX50_S1_L1600_A20, "PH20": expectedW128_MIN5_MAX50_S1_L1600_PH20, "TO20": expectedW128_MIN5_MAX50_S1_L1600_TO20, "A21": expectedW128_MIN5_MAX50_S1_L1600_A21, "PH21": expectedW128_MIN5_MAX50_S1_L1600_PH21, "TO21": expectedW128_MIN5_MAX50_S1_L1600_TO21, "A22": expectedW128_MIN5_MAX50_S1_L1600_A22, "PH22": expectedW128_MIN5_MAX50_S1_L1600_PH22, "TO22": expectedW128_MIN5_MAX50_S1_L1600_TO22, "A23": expectedW128_MIN5_MAX50_S1_L1600_A23, "PH23": expectedW128_MIN5_MAX50_S1_L1600_PH23, "TO23": expectedW128_MIN5_MAX50_S1_L1600_TO23, "A24": expectedW128_MIN5_MAX50_S1_L1600_A24, "PH24": expectedW128_MIN5_MAX50_S1_L1600_PH24, "TO24": expectedW128_MIN5_MAX50_S1_L1600_TO24, "A25": expectedW128_MIN5_MAX50_S1_L1600_A25, "PH25": expectedW128_MIN5_MAX50_S1_L1600_PH25, "TO25": expectedW128_MIN5_MAX50_S1_L1600_TO25, "A26": expectedW128_MIN5_MAX50_S1_L1600_A26, "PH26": expectedW128_MIN5_MAX50_S1_L1600_PH26, "TO26": expectedW128_MIN5_MAX50_S1_L1600_TO26, "A27": expectedW128_MIN5_MAX50_S1_L1600_A27, "PH27": expectedW128_MIN5_MAX50_S1_L1600_PH27, "TO27": expectedW128_MIN5_MAX50_S1_L1600_TO27, "A28": expectedW128_MIN5_MAX50_S1_L1600_A28, "PH28": expectedW128_MIN5_MAX50_S1_L1600_PH28, "TO28": expectedW128_MIN5_MAX50_S1_L1600_TO28, "A29": expectedW128_MIN5_MAX50_S1_L1600_A29, "PH29": expectedW128_MIN5_MAX50_S1_L1600_PH29, "TO29": expectedW128_MIN5_MAX50_S1_L1600_TO29, "A30": expectedW128_MIN5_MAX50_S1_L1600_A30, "PH30": expectedW128_MIN5_MAX50_S1_L1600_PH30, "TO30": expectedW128_MIN5_MAX50_S1_L1600_TO30, "A31": expectedW128_MIN5_MAX50_S1_L1600_A31, "PH31": expectedW128_MIN5_MAX50_S1_L1600_PH31, "TO31": expectedW128_MIN5_MAX50_S1_L1600_TO31, "A32": expectedW128_MIN5_MAX50_S1_L1600_A32, "PH32": expectedW128_MIN5_MAX50_S1_L1600_PH32, "TO32": expectedW128_MIN5_MAX50_S1_L1600_TO32, "A33": expectedW128_MIN5_MAX50_S1_L1600_A33, "PH33": expectedW128_MIN5_MAX50_S1_L1600_PH33, "TO33": expectedW128_MIN5_MAX50_S1_L1600_TO33, "A34": expectedW128_MIN5_MAX50_S1_L1600_A34, "PH34": expectedW128_MIN5_MAX50_S1_L1600_PH34, "TO34": expectedW128_MIN5_MAX50_S1_L1600_TO34, "A35": expectedW128_MIN5_MAX50_S1_L1600_A35, "PH35": expectedW128_MIN5_MAX50_S1_L1600_PH35, "TO35": expectedW128_MIN5_MAX50_S1_L1600_TO35, "A36": expectedW128_MIN5_MAX50_S1_L1600_A36, "PH36": expectedW128_MIN5_MAX50_S1_L1600_PH36, "TO36": expectedW128_MIN5_MAX50_S1_L1600_TO36, "A37": expectedW128_MIN5_MAX50_S1_L1600_A37, "PH37": expectedW128_MIN5_MAX50_S1_L1600_PH37, "TO37": expectedW128_MIN5_MAX50_S1_L1600_TO37, "A38": expectedW128_MIN5_MAX50_S1_L1600_A38, "PH38": expectedW128_MIN5_MAX50_S1_L1600_PH38, "TO38": expectedW128_MIN5_MAX50_S1_L1600_TO38, "A39": expectedW128_MIN5_MAX50_S1_L1600_A39, "PH39": expectedW128_MIN5_MAX50_S1_L1600_PH39, "TO39": expectedW128_MIN5_MAX50_S1_L1600_TO39, "A40": expectedW128_MIN5_MAX50_S1_L1600_A40, "PH40": expectedW128_MIN5_MAX50_S1_L1600_PH40, "TO40": expectedW128_MIN5_MAX50_S1_L1600_TO40, "A41": expectedW128_MIN5_MAX50_S1_L1600_A41, "PH41": expectedW128_MIN5_MAX50_S1_L1600_PH41, "TO41": expectedW128_MIN5_MAX50_S1_L1600_TO41, "A42": expectedW128_MIN5_MAX50_S1_L1600_A42, "PH42": expectedW128_MIN5_MAX50_S1_L1600_PH42, "TO42": expectedW128_MIN5_MAX50_S1_L1600_TO42, "A43": expectedW128_MIN5_MAX50_S1_L1600_A43, "PH43": expectedW128_MIN5_MAX50_S1_L1600_PH43, "TO43": expectedW128_MIN5_MAX50_S1_L1600_TO43, "A44": expectedW128_MIN5_MAX50_S1_L1600_A44, "PH44": expectedW128_MIN5_MAX50_S1_L1600_PH44, "TO44": expectedW128_MIN5_MAX50_S1_L1600_TO44, "A45": expectedW128_MIN5_MAX50_S1_L1600_A45, "PH45": expectedW128_MIN5_MAX50_S1_L1600_PH45, "TO45": expectedW128_MIN5_MAX50_S1_L1600_TO45, "A46": expectedW128_MIN5_MAX50_S1_L1600_A46, "PH46": expectedW128_MIN5_MAX50_S1_L1600_PH46, "TO46": expectedW128_MIN5_MAX50_S1_L1600_TO46, "A47": expectedW128_MIN5_MAX50_S1_L1600_A47, "PH47": expectedW128_MIN5_MAX50_S1_L1600_PH47, "TO47": expectedW128_MIN5_MAX50_S1_L1600_TO47, "A48": expectedW128_MIN5_MAX50_S1_L1600_A48, "PH48": expectedW128_MIN5_MAX50_S1_L1600_PH48, "TO48": expectedW128_MIN5_MAX50_S1_L1600_TO48, "A49": expectedW128_MIN5_MAX50_S1_L1600_A49, "PH49": expectedW128_MIN5_MAX50_S1_L1600_PH49, "TO49": expectedW128_MIN5_MAX50_S1_L1600_TO49, "A50": expectedW128_MIN5_MAX50_S1_L1600_A50, "PH50": expectedW128_MIN5_MAX50_S1_L1600_PH50, "TO50": expectedW128_MIN5_MAX50_S1_L1600_TO50, "DOM": expectedW128_MIN5_MAX50_S1_L1600_DOM, "AMP": expectedW128_MIN5_MAX50_S1_L1600_AMP, "PHASE": expectedW128_MIN5_MAX50_S1_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A5"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A5"] = result.Amplitude[0]
			}
			actual["PH5"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH5"] = result.Phases[0]
			}
			actual["TO5"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO5"] = result.TroughOffsets[0]
			}
			actual["A6"] = math.NaN()
			if len(result.Amplitude) > 1 {
				actual["A6"] = result.Amplitude[1]
			}
			actual["PH6"] = math.NaN()
			if len(result.Phases) > 1 {
				actual["PH6"] = result.Phases[1]
			}
			actual["TO6"] = math.NaN()
			if len(result.TroughOffsets) > 1 {
				actual["TO6"] = result.TroughOffsets[1]
			}
			actual["A7"] = math.NaN()
			if len(result.Amplitude) > 2 {
				actual["A7"] = result.Amplitude[2]
			}
			actual["PH7"] = math.NaN()
			if len(result.Phases) > 2 {
				actual["PH7"] = result.Phases[2]
			}
			actual["TO7"] = math.NaN()
			if len(result.TroughOffsets) > 2 {
				actual["TO7"] = result.TroughOffsets[2]
			}
			actual["A8"] = math.NaN()
			if len(result.Amplitude) > 3 {
				actual["A8"] = result.Amplitude[3]
			}
			actual["PH8"] = math.NaN()
			if len(result.Phases) > 3 {
				actual["PH8"] = result.Phases[3]
			}
			actual["TO8"] = math.NaN()
			if len(result.TroughOffsets) > 3 {
				actual["TO8"] = result.TroughOffsets[3]
			}
			actual["A9"] = math.NaN()
			if len(result.Amplitude) > 4 {
				actual["A9"] = result.Amplitude[4]
			}
			actual["PH9"] = math.NaN()
			if len(result.Phases) > 4 {
				actual["PH9"] = result.Phases[4]
			}
			actual["TO9"] = math.NaN()
			if len(result.TroughOffsets) > 4 {
				actual["TO9"] = result.TroughOffsets[4]
			}
			actual["A10"] = math.NaN()
			if len(result.Amplitude) > 5 {
				actual["A10"] = result.Amplitude[5]
			}
			actual["PH10"] = math.NaN()
			if len(result.Phases) > 5 {
				actual["PH10"] = result.Phases[5]
			}
			actual["TO10"] = math.NaN()
			if len(result.TroughOffsets) > 5 {
				actual["TO10"] = result.TroughOffsets[5]
			}
			actual["A11"] = math.NaN()
			if len(result.Amplitude) > 6 {
				actual["A11"] = result.Amplitude[6]
			}
			actual["PH11"] = math.NaN()
			if len(result.Phases) > 6 {
				actual["PH11"] = result.Phases[6]
			}
			actual["TO11"] = math.NaN()
			if len(result.TroughOffsets) > 6 {
				actual["TO11"] = result.TroughOffsets[6]
			}
			actual["A12"] = math.NaN()
			if len(result.Amplitude) > 7 {
				actual["A12"] = result.Amplitude[7]
			}
			actual["PH12"] = math.NaN()
			if len(result.Phases) > 7 {
				actual["PH12"] = result.Phases[7]
			}
			actual["TO12"] = math.NaN()
			if len(result.TroughOffsets) > 7 {
				actual["TO12"] = result.TroughOffsets[7]
			}
			actual["A13"] = math.NaN()
			if len(result.Amplitude) > 8 {
				actual["A13"] = result.Amplitude[8]
			}
			actual["PH13"] = math.NaN()
			if len(result.Phases) > 8 {
				actual["PH13"] = result.Phases[8]
			}
			actual["TO13"] = math.NaN()
			if len(result.TroughOffsets) > 8 {
				actual["TO13"] = result.TroughOffsets[8]
			}
			actual["A14"] = math.NaN()
			if len(result.Amplitude) > 9 {
				actual["A14"] = result.Amplitude[9]
			}
			actual["PH14"] = math.NaN()
			if len(result.Phases) > 9 {
				actual["PH14"] = result.Phases[9]
			}
			actual["TO14"] = math.NaN()
			if len(result.TroughOffsets) > 9 {
				actual["TO14"] = result.TroughOffsets[9]
			}
			actual["A15"] = math.NaN()
			if len(result.Amplitude) > 10 {
				actual["A15"] = result.Amplitude[10]
			}
			actual["PH15"] = math.NaN()
			if len(result.Phases) > 10 {
				actual["PH15"] = result.Phases[10]
			}
			actual["TO15"] = math.NaN()
			if len(result.TroughOffsets) > 10 {
				actual["TO15"] = result.TroughOffsets[10]
			}
			actual["A16"] = math.NaN()
			if len(result.Amplitude) > 11 {
				actual["A16"] = result.Amplitude[11]
			}
			actual["PH16"] = math.NaN()
			if len(result.Phases) > 11 {
				actual["PH16"] = result.Phases[11]
			}
			actual["TO16"] = math.NaN()
			if len(result.TroughOffsets) > 11 {
				actual["TO16"] = result.TroughOffsets[11]
			}
			actual["A17"] = math.NaN()
			if len(result.Amplitude) > 12 {
				actual["A17"] = result.Amplitude[12]
			}
			actual["PH17"] = math.NaN()
			if len(result.Phases) > 12 {
				actual["PH17"] = result.Phases[12]
			}
			actual["TO17"] = math.NaN()
			if len(result.TroughOffsets) > 12 {
				actual["TO17"] = result.TroughOffsets[12]
			}
			actual["A18"] = math.NaN()
			if len(result.Amplitude) > 13 {
				actual["A18"] = result.Amplitude[13]
			}
			actual["PH18"] = math.NaN()
			if len(result.Phases) > 13 {
				actual["PH18"] = result.Phases[13]
			}
			actual["TO18"] = math.NaN()
			if len(result.TroughOffsets) > 13 {
				actual["TO18"] = result.TroughOffsets[13]
			}
			actual["A19"] = math.NaN()
			if len(result.Amplitude) > 14 {
				actual["A19"] = result.Amplitude[14]
			}
			actual["PH19"] = math.NaN()
			if len(result.Phases) > 14 {
				actual["PH19"] = result.Phases[14]
			}
			actual["TO19"] = math.NaN()
			if len(result.TroughOffsets) > 14 {
				actual["TO19"] = result.TroughOffsets[14]
			}
			actual["A20"] = math.NaN()
			if len(result.Amplitude) > 15 {
				actual["A20"] = result.Amplitude[15]
			}
			actual["PH20"] = math.NaN()
			if len(result.Phases) > 15 {
				actual["PH20"] = result.Phases[15]
			}
			actual["TO20"] = math.NaN()
			if len(result.TroughOffsets) > 15 {
				actual["TO20"] = result.TroughOffsets[15]
			}
			actual["A21"] = math.NaN()
			if len(result.Amplitude) > 16 {
				actual["A21"] = result.Amplitude[16]
			}
			actual["PH21"] = math.NaN()
			if len(result.Phases) > 16 {
				actual["PH21"] = result.Phases[16]
			}
			actual["TO21"] = math.NaN()
			if len(result.TroughOffsets) > 16 {
				actual["TO21"] = result.TroughOffsets[16]
			}
			actual["A22"] = math.NaN()
			if len(result.Amplitude) > 17 {
				actual["A22"] = result.Amplitude[17]
			}
			actual["PH22"] = math.NaN()
			if len(result.Phases) > 17 {
				actual["PH22"] = result.Phases[17]
			}
			actual["TO22"] = math.NaN()
			if len(result.TroughOffsets) > 17 {
				actual["TO22"] = result.TroughOffsets[17]
			}
			actual["A23"] = math.NaN()
			if len(result.Amplitude) > 18 {
				actual["A23"] = result.Amplitude[18]
			}
			actual["PH23"] = math.NaN()
			if len(result.Phases) > 18 {
				actual["PH23"] = result.Phases[18]
			}
			actual["TO23"] = math.NaN()
			if len(result.TroughOffsets) > 18 {
				actual["TO23"] = result.TroughOffsets[18]
			}
			actual["A24"] = math.NaN()
			if len(result.Amplitude) > 19 {
				actual["A24"] = result.Amplitude[19]
			}
			actual["PH24"] = math.NaN()
			if len(result.Phases) > 19 {
				actual["PH24"] = result.Phases[19]
			}
			actual["TO24"] = math.NaN()
			if len(result.TroughOffsets) > 19 {
				actual["TO24"] = result.TroughOffsets[19]
			}
			actual["A25"] = math.NaN()
			if len(result.Amplitude) > 20 {
				actual["A25"] = result.Amplitude[20]
			}
			actual["PH25"] = math.NaN()
			if len(result.Phases) > 20 {
				actual["PH25"] = result.Phases[20]
			}
			actual["TO25"] = math.NaN()
			if len(result.TroughOffsets) > 20 {
				actual["TO25"] = result.TroughOffsets[20]
			}
			actual["A26"] = math.NaN()
			if len(result.Amplitude) > 21 {
				actual["A26"] = result.Amplitude[21]
			}
			actual["PH26"] = math.NaN()
			if len(result.Phases) > 21 {
				actual["PH26"] = result.Phases[21]
			}
			actual["TO26"] = math.NaN()
			if len(result.TroughOffsets) > 21 {
				actual["TO26"] = result.TroughOffsets[21]
			}
			actual["A27"] = math.NaN()
			if len(result.Amplitude) > 22 {
				actual["A27"] = result.Amplitude[22]
			}
			actual["PH27"] = math.NaN()
			if len(result.Phases) > 22 {
				actual["PH27"] = result.Phases[22]
			}
			actual["TO27"] = math.NaN()
			if len(result.TroughOffsets) > 22 {
				actual["TO27"] = result.TroughOffsets[22]
			}
			actual["A28"] = math.NaN()
			if len(result.Amplitude) > 23 {
				actual["A28"] = result.Amplitude[23]
			}
			actual["PH28"] = math.NaN()
			if len(result.Phases) > 23 {
				actual["PH28"] = result.Phases[23]
			}
			actual["TO28"] = math.NaN()
			if len(result.TroughOffsets) > 23 {
				actual["TO28"] = result.TroughOffsets[23]
			}
			actual["A29"] = math.NaN()
			if len(result.Amplitude) > 24 {
				actual["A29"] = result.Amplitude[24]
			}
			actual["PH29"] = math.NaN()
			if len(result.Phases) > 24 {
				actual["PH29"] = result.Phases[24]
			}
			actual["TO29"] = math.NaN()
			if len(result.TroughOffsets) > 24 {
				actual["TO29"] = result.TroughOffsets[24]
			}
			actual["A30"] = math.NaN()
			if len(result.Amplitude) > 25 {
				actual["A30"] = result.Amplitude[25]
			}
			actual["PH30"] = math.NaN()
			if len(result.Phases) > 25 {
				actual["PH30"] = result.Phases[25]
			}
			actual["TO30"] = math.NaN()
			if len(result.TroughOffsets) > 25 {
				actual["TO30"] = result.TroughOffsets[25]
			}
			actual["A31"] = math.NaN()
			if len(result.Amplitude) > 26 {
				actual["A31"] = result.Amplitude[26]
			}
			actual["PH31"] = math.NaN()
			if len(result.Phases) > 26 {
				actual["PH31"] = result.Phases[26]
			}
			actual["TO31"] = math.NaN()
			if len(result.TroughOffsets) > 26 {
				actual["TO31"] = result.TroughOffsets[26]
			}
			actual["A32"] = math.NaN()
			if len(result.Amplitude) > 27 {
				actual["A32"] = result.Amplitude[27]
			}
			actual["PH32"] = math.NaN()
			if len(result.Phases) > 27 {
				actual["PH32"] = result.Phases[27]
			}
			actual["TO32"] = math.NaN()
			if len(result.TroughOffsets) > 27 {
				actual["TO32"] = result.TroughOffsets[27]
			}
			actual["A33"] = math.NaN()
			if len(result.Amplitude) > 28 {
				actual["A33"] = result.Amplitude[28]
			}
			actual["PH33"] = math.NaN()
			if len(result.Phases) > 28 {
				actual["PH33"] = result.Phases[28]
			}
			actual["TO33"] = math.NaN()
			if len(result.TroughOffsets) > 28 {
				actual["TO33"] = result.TroughOffsets[28]
			}
			actual["A34"] = math.NaN()
			if len(result.Amplitude) > 29 {
				actual["A34"] = result.Amplitude[29]
			}
			actual["PH34"] = math.NaN()
			if len(result.Phases) > 29 {
				actual["PH34"] = result.Phases[29]
			}
			actual["TO34"] = math.NaN()
			if len(result.TroughOffsets) > 29 {
				actual["TO34"] = result.TroughOffsets[29]
			}
			actual["A35"] = math.NaN()
			if len(result.Amplitude) > 30 {
				actual["A35"] = result.Amplitude[30]
			}
			actual["PH35"] = math.NaN()
			if len(result.Phases) > 30 {
				actual["PH35"] = result.Phases[30]
			}
			actual["TO35"] = math.NaN()
			if len(result.TroughOffsets) > 30 {
				actual["TO35"] = result.TroughOffsets[30]
			}
			actual["A36"] = math.NaN()
			if len(result.Amplitude) > 31 {
				actual["A36"] = result.Amplitude[31]
			}
			actual["PH36"] = math.NaN()
			if len(result.Phases) > 31 {
				actual["PH36"] = result.Phases[31]
			}
			actual["TO36"] = math.NaN()
			if len(result.TroughOffsets) > 31 {
				actual["TO36"] = result.TroughOffsets[31]
			}
			actual["A37"] = math.NaN()
			if len(result.Amplitude) > 32 {
				actual["A37"] = result.Amplitude[32]
			}
			actual["PH37"] = math.NaN()
			if len(result.Phases) > 32 {
				actual["PH37"] = result.Phases[32]
			}
			actual["TO37"] = math.NaN()
			if len(result.TroughOffsets) > 32 {
				actual["TO37"] = result.TroughOffsets[32]
			}
			actual["A38"] = math.NaN()
			if len(result.Amplitude) > 33 {
				actual["A38"] = result.Amplitude[33]
			}
			actual["PH38"] = math.NaN()
			if len(result.Phases) > 33 {
				actual["PH38"] = result.Phases[33]
			}
			actual["TO38"] = math.NaN()
			if len(result.TroughOffsets) > 33 {
				actual["TO38"] = result.TroughOffsets[33]
			}
			actual["A39"] = math.NaN()
			if len(result.Amplitude) > 34 {
				actual["A39"] = result.Amplitude[34]
			}
			actual["PH39"] = math.NaN()
			if len(result.Phases) > 34 {
				actual["PH39"] = result.Phases[34]
			}
			actual["TO39"] = math.NaN()
			if len(result.TroughOffsets) > 34 {
				actual["TO39"] = result.TroughOffsets[34]
			}
			actual["A40"] = math.NaN()
			if len(result.Amplitude) > 35 {
				actual["A40"] = result.Amplitude[35]
			}
			actual["PH40"] = math.NaN()
			if len(result.Phases) > 35 {
				actual["PH40"] = result.Phases[35]
			}
			actual["TO40"] = math.NaN()
			if len(result.TroughOffsets) > 35 {
				actual["TO40"] = result.TroughOffsets[35]
			}
			actual["A41"] = math.NaN()
			if len(result.Amplitude) > 36 {
				actual["A41"] = result.Amplitude[36]
			}
			actual["PH41"] = math.NaN()
			if len(result.Phases) > 36 {
				actual["PH41"] = result.Phases[36]
			}
			actual["TO41"] = math.NaN()
			if len(result.TroughOffsets) > 36 {
				actual["TO41"] = result.TroughOffsets[36]
			}
			actual["A42"] = math.NaN()
			if len(result.Amplitude) > 37 {
				actual["A42"] = result.Amplitude[37]
			}
			actual["PH42"] = math.NaN()
			if len(result.Phases) > 37 {
				actual["PH42"] = result.Phases[37]
			}
			actual["TO42"] = math.NaN()
			if len(result.TroughOffsets) > 37 {
				actual["TO42"] = result.TroughOffsets[37]
			}
			actual["A43"] = math.NaN()
			if len(result.Amplitude) > 38 {
				actual["A43"] = result.Amplitude[38]
			}
			actual["PH43"] = math.NaN()
			if len(result.Phases) > 38 {
				actual["PH43"] = result.Phases[38]
			}
			actual["TO43"] = math.NaN()
			if len(result.TroughOffsets) > 38 {
				actual["TO43"] = result.TroughOffsets[38]
			}
			actual["A44"] = math.NaN()
			if len(result.Amplitude) > 39 {
				actual["A44"] = result.Amplitude[39]
			}
			actual["PH44"] = math.NaN()
			if len(result.Phases) > 39 {
				actual["PH44"] = result.Phases[39]
			}
			actual["TO44"] = math.NaN()
			if len(result.TroughOffsets) > 39 {
				actual["TO44"] = result.TroughOffsets[39]
			}
			actual["A45"] = math.NaN()
			if len(result.Amplitude) > 40 {
				actual["A45"] = result.Amplitude[40]
			}
			actual["PH45"] = math.NaN()
			if len(result.Phases) > 40 {
				actual["PH45"] = result.Phases[40]
			}
			actual["TO45"] = math.NaN()
			if len(result.TroughOffsets) > 40 {
				actual["TO45"] = result.TroughOffsets[40]
			}
			actual["A46"] = math.NaN()
			if len(result.Amplitude) > 41 {
				actual["A46"] = result.Amplitude[41]
			}
			actual["PH46"] = math.NaN()
			if len(result.Phases) > 41 {
				actual["PH46"] = result.Phases[41]
			}
			actual["TO46"] = math.NaN()
			if len(result.TroughOffsets) > 41 {
				actual["TO46"] = result.TroughOffsets[41]
			}
			actual["A47"] = math.NaN()
			if len(result.Amplitude) > 42 {
				actual["A47"] = result.Amplitude[42]
			}
			actual["PH47"] = math.NaN()
			if len(result.Phases) > 42 {
				actual["PH47"] = result.Phases[42]
			}
			actual["TO47"] = math.NaN()
			if len(result.TroughOffsets) > 42 {
				actual["TO47"] = result.TroughOffsets[42]
			}
			actual["A48"] = math.NaN()
			if len(result.Amplitude) > 43 {
				actual["A48"] = result.Amplitude[43]
			}
			actual["PH48"] = math.NaN()
			if len(result.Phases) > 43 {
				actual["PH48"] = result.Phases[43]
			}
			actual["TO48"] = math.NaN()
			if len(result.TroughOffsets) > 43 {
				actual["TO48"] = result.TroughOffsets[43]
			}
			actual["A49"] = math.NaN()
			if len(result.Amplitude) > 44 {
				actual["A49"] = result.Amplitude[44]
			}
			actual["PH49"] = math.NaN()
			if len(result.Phases) > 44 {
				actual["PH49"] = result.Phases[44]
			}
			actual["TO49"] = math.NaN()
			if len(result.TroughOffsets) > 44 {
				actual["TO49"] = result.TroughOffsets[44]
			}
			actual["A50"] = math.NaN()
			if len(result.Amplitude) > 45 {
				actual["A50"] = result.Amplitude[45]
			}
			actual["PH50"] = math.NaN()
			if len(result.Phases) > 45 {
				actual["PH50"] = result.Phases[45]
			}
			actual["TO50"] = math.NaN()
			if len(result.TroughOffsets) > 45 {
				actual["TO50"] = result.TroughOffsets[45]
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
		p.Window = 100
		p.MinPeriod = 20
		p.MaxPeriod = 20
		p.Step = 1
		p.Smoothing = 0.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A20": expectedW100_MIN20_MAX20_S1_L0_A20, "PH20": expectedW100_MIN20_MAX20_S1_L0_PH20, "TO20": expectedW100_MIN20_MAX20_S1_L0_TO20, "DOM": expectedW100_MIN20_MAX20_S1_L0_DOM, "AMP": expectedW100_MIN20_MAX20_S1_L0_AMP, "PHASE": expectedW100_MIN20_MAX20_S1_L0_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A20"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A20"] = result.Amplitude[0]
			}
			actual["PH20"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH20"] = result.Phases[0]
			}
			actual["TO20"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO20"] = result.TroughOffsets[0]
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
		p.Window = 40
		p.MinPeriod = 5
		p.MaxPeriod = 25
		p.Step = 5
		p.Smoothing = 0.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A5": expectedW40_MIN5_MAX25_S5_L0_A5, "PH5": expectedW40_MIN5_MAX25_S5_L0_PH5, "TO5": expectedW40_MIN5_MAX25_S5_L0_TO5, "A10": expectedW40_MIN5_MAX25_S5_L0_A10, "PH10": expectedW40_MIN5_MAX25_S5_L0_PH10, "TO10": expectedW40_MIN5_MAX25_S5_L0_TO10, "A15": expectedW40_MIN5_MAX25_S5_L0_A15, "PH15": expectedW40_MIN5_MAX25_S5_L0_PH15, "TO15": expectedW40_MIN5_MAX25_S5_L0_TO15, "A20": expectedW40_MIN5_MAX25_S5_L0_A20, "PH20": expectedW40_MIN5_MAX25_S5_L0_PH20, "TO20": expectedW40_MIN5_MAX25_S5_L0_TO20, "A25": expectedW40_MIN5_MAX25_S5_L0_A25, "PH25": expectedW40_MIN5_MAX25_S5_L0_PH25, "TO25": expectedW40_MIN5_MAX25_S5_L0_TO25, "DOM": expectedW40_MIN5_MAX25_S5_L0_DOM, "AMP": expectedW40_MIN5_MAX25_S5_L0_AMP, "PHASE": expectedW40_MIN5_MAX25_S5_L0_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A5"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A5"] = result.Amplitude[0]
			}
			actual["PH5"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH5"] = result.Phases[0]
			}
			actual["TO5"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO5"] = result.TroughOffsets[0]
			}
			actual["A10"] = math.NaN()
			if len(result.Amplitude) > 1 {
				actual["A10"] = result.Amplitude[1]
			}
			actual["PH10"] = math.NaN()
			if len(result.Phases) > 1 {
				actual["PH10"] = result.Phases[1]
			}
			actual["TO10"] = math.NaN()
			if len(result.TroughOffsets) > 1 {
				actual["TO10"] = result.TroughOffsets[1]
			}
			actual["A15"] = math.NaN()
			if len(result.Amplitude) > 2 {
				actual["A15"] = result.Amplitude[2]
			}
			actual["PH15"] = math.NaN()
			if len(result.Phases) > 2 {
				actual["PH15"] = result.Phases[2]
			}
			actual["TO15"] = math.NaN()
			if len(result.TroughOffsets) > 2 {
				actual["TO15"] = result.TroughOffsets[2]
			}
			actual["A20"] = math.NaN()
			if len(result.Amplitude) > 3 {
				actual["A20"] = result.Amplitude[3]
			}
			actual["PH20"] = math.NaN()
			if len(result.Phases) > 3 {
				actual["PH20"] = result.Phases[3]
			}
			actual["TO20"] = math.NaN()
			if len(result.TroughOffsets) > 3 {
				actual["TO20"] = result.TroughOffsets[3]
			}
			actual["A25"] = math.NaN()
			if len(result.Amplitude) > 4 {
				actual["A25"] = result.Amplitude[4]
			}
			actual["PH25"] = math.NaN()
			if len(result.Phases) > 4 {
				actual["PH25"] = result.Phases[4]
			}
			actual["TO25"] = math.NaN()
			if len(result.TroughOffsets) > 4 {
				actual["TO25"] = result.TroughOffsets[4]
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
		p.Window = 40
		p.MinPeriod = 5
		p.MaxPeriod = 25
		p.Step = 5
		p.Smoothing = 1600.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A5": expectedW40_MIN5_MAX25_S5_L1600_A5, "PH5": expectedW40_MIN5_MAX25_S5_L1600_PH5, "TO5": expectedW40_MIN5_MAX25_S5_L1600_TO5, "A10": expectedW40_MIN5_MAX25_S5_L1600_A10, "PH10": expectedW40_MIN5_MAX25_S5_L1600_PH10, "TO10": expectedW40_MIN5_MAX25_S5_L1600_TO10, "A15": expectedW40_MIN5_MAX25_S5_L1600_A15, "PH15": expectedW40_MIN5_MAX25_S5_L1600_PH15, "TO15": expectedW40_MIN5_MAX25_S5_L1600_TO15, "A20": expectedW40_MIN5_MAX25_S5_L1600_A20, "PH20": expectedW40_MIN5_MAX25_S5_L1600_PH20, "TO20": expectedW40_MIN5_MAX25_S5_L1600_TO20, "A25": expectedW40_MIN5_MAX25_S5_L1600_A25, "PH25": expectedW40_MIN5_MAX25_S5_L1600_PH25, "TO25": expectedW40_MIN5_MAX25_S5_L1600_TO25, "DOM": expectedW40_MIN5_MAX25_S5_L1600_DOM, "AMP": expectedW40_MIN5_MAX25_S5_L1600_AMP, "PHASE": expectedW40_MIN5_MAX25_S5_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A5"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A5"] = result.Amplitude[0]
			}
			actual["PH5"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH5"] = result.Phases[0]
			}
			actual["TO5"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO5"] = result.TroughOffsets[0]
			}
			actual["A10"] = math.NaN()
			if len(result.Amplitude) > 1 {
				actual["A10"] = result.Amplitude[1]
			}
			actual["PH10"] = math.NaN()
			if len(result.Phases) > 1 {
				actual["PH10"] = result.Phases[1]
			}
			actual["TO10"] = math.NaN()
			if len(result.TroughOffsets) > 1 {
				actual["TO10"] = result.TroughOffsets[1]
			}
			actual["A15"] = math.NaN()
			if len(result.Amplitude) > 2 {
				actual["A15"] = result.Amplitude[2]
			}
			actual["PH15"] = math.NaN()
			if len(result.Phases) > 2 {
				actual["PH15"] = result.Phases[2]
			}
			actual["TO15"] = math.NaN()
			if len(result.TroughOffsets) > 2 {
				actual["TO15"] = result.TroughOffsets[2]
			}
			actual["A20"] = math.NaN()
			if len(result.Amplitude) > 3 {
				actual["A20"] = result.Amplitude[3]
			}
			actual["PH20"] = math.NaN()
			if len(result.Phases) > 3 {
				actual["PH20"] = result.Phases[3]
			}
			actual["TO20"] = math.NaN()
			if len(result.TroughOffsets) > 3 {
				actual["TO20"] = result.TroughOffsets[3]
			}
			actual["A25"] = math.NaN()
			if len(result.Amplitude) > 4 {
				actual["A25"] = result.Amplitude[4]
			}
			actual["PH25"] = math.NaN()
			if len(result.Phases) > 4 {
				actual["PH25"] = result.Phases[4]
			}
			actual["TO25"] = math.NaN()
			if len(result.TroughOffsets) > 4 {
				actual["TO25"] = result.TroughOffsets[4]
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
	t.Run("case4", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 200
		p.MinPeriod = 10
		p.MaxPeriod = 50
		p.Step = 10
		p.Smoothing = 1600.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A10": expectedW200_MIN10_MAX50_S10_L1600_A10, "PH10": expectedW200_MIN10_MAX50_S10_L1600_PH10, "TO10": expectedW200_MIN10_MAX50_S10_L1600_TO10, "A20": expectedW200_MIN10_MAX50_S10_L1600_A20, "PH20": expectedW200_MIN10_MAX50_S10_L1600_PH20, "TO20": expectedW200_MIN10_MAX50_S10_L1600_TO20, "A30": expectedW200_MIN10_MAX50_S10_L1600_A30, "PH30": expectedW200_MIN10_MAX50_S10_L1600_PH30, "TO30": expectedW200_MIN10_MAX50_S10_L1600_TO30, "A40": expectedW200_MIN10_MAX50_S10_L1600_A40, "PH40": expectedW200_MIN10_MAX50_S10_L1600_PH40, "TO40": expectedW200_MIN10_MAX50_S10_L1600_TO40, "A50": expectedW200_MIN10_MAX50_S10_L1600_A50, "PH50": expectedW200_MIN10_MAX50_S10_L1600_PH50, "TO50": expectedW200_MIN10_MAX50_S10_L1600_TO50, "DOM": expectedW200_MIN10_MAX50_S10_L1600_DOM, "AMP": expectedW200_MIN10_MAX50_S10_L1600_AMP, "PHASE": expectedW200_MIN10_MAX50_S10_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A10"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A10"] = result.Amplitude[0]
			}
			actual["PH10"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH10"] = result.Phases[0]
			}
			actual["TO10"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO10"] = result.TroughOffsets[0]
			}
			actual["A20"] = math.NaN()
			if len(result.Amplitude) > 1 {
				actual["A20"] = result.Amplitude[1]
			}
			actual["PH20"] = math.NaN()
			if len(result.Phases) > 1 {
				actual["PH20"] = result.Phases[1]
			}
			actual["TO20"] = math.NaN()
			if len(result.TroughOffsets) > 1 {
				actual["TO20"] = result.TroughOffsets[1]
			}
			actual["A30"] = math.NaN()
			if len(result.Amplitude) > 2 {
				actual["A30"] = result.Amplitude[2]
			}
			actual["PH30"] = math.NaN()
			if len(result.Phases) > 2 {
				actual["PH30"] = result.Phases[2]
			}
			actual["TO30"] = math.NaN()
			if len(result.TroughOffsets) > 2 {
				actual["TO30"] = result.TroughOffsets[2]
			}
			actual["A40"] = math.NaN()
			if len(result.Amplitude) > 3 {
				actual["A40"] = result.Amplitude[3]
			}
			actual["PH40"] = math.NaN()
			if len(result.Phases) > 3 {
				actual["PH40"] = result.Phases[3]
			}
			actual["TO40"] = math.NaN()
			if len(result.TroughOffsets) > 3 {
				actual["TO40"] = result.TroughOffsets[3]
			}
			actual["A50"] = math.NaN()
			if len(result.Amplitude) > 4 {
				actual["A50"] = result.Amplitude[4]
			}
			actual["PH50"] = math.NaN()
			if len(result.Phases) > 4 {
				actual["PH50"] = result.Phases[4]
			}
			actual["TO50"] = math.NaN()
			if len(result.TroughOffsets) > 4 {
				actual["TO50"] = result.TroughOffsets[4]
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
	t.Run("case5", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 20
		p.MinPeriod = 10
		p.MaxPeriod = 50
		p.Step = 10
		p.Smoothing = 1600.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A10": expectedW20_MIN10_MAX50_S10_L1600_A10, "PH10": expectedW20_MIN10_MAX50_S10_L1600_PH10, "TO10": expectedW20_MIN10_MAX50_S10_L1600_TO10, "A20": expectedW20_MIN10_MAX50_S10_L1600_A20, "PH20": expectedW20_MIN10_MAX50_S10_L1600_PH20, "TO20": expectedW20_MIN10_MAX50_S10_L1600_TO20, "A30": expectedW20_MIN10_MAX50_S10_L1600_A30, "PH30": expectedW20_MIN10_MAX50_S10_L1600_PH30, "TO30": expectedW20_MIN10_MAX50_S10_L1600_TO30, "A40": expectedW20_MIN10_MAX50_S10_L1600_A40, "PH40": expectedW20_MIN10_MAX50_S10_L1600_PH40, "TO40": expectedW20_MIN10_MAX50_S10_L1600_TO40, "A50": expectedW20_MIN10_MAX50_S10_L1600_A50, "PH50": expectedW20_MIN10_MAX50_S10_L1600_PH50, "TO50": expectedW20_MIN10_MAX50_S10_L1600_TO50, "DOM": expectedW20_MIN10_MAX50_S10_L1600_DOM, "AMP": expectedW20_MIN10_MAX50_S10_L1600_AMP, "PHASE": expectedW20_MIN10_MAX50_S10_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A10"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A10"] = result.Amplitude[0]
			}
			actual["PH10"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH10"] = result.Phases[0]
			}
			actual["TO10"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO10"] = result.TroughOffsets[0]
			}
			actual["A20"] = math.NaN()
			if len(result.Amplitude) > 1 {
				actual["A20"] = result.Amplitude[1]
			}
			actual["PH20"] = math.NaN()
			if len(result.Phases) > 1 {
				actual["PH20"] = result.Phases[1]
			}
			actual["TO20"] = math.NaN()
			if len(result.TroughOffsets) > 1 {
				actual["TO20"] = result.TroughOffsets[1]
			}
			actual["A30"] = math.NaN()
			if len(result.Amplitude) > 2 {
				actual["A30"] = result.Amplitude[2]
			}
			actual["PH30"] = math.NaN()
			if len(result.Phases) > 2 {
				actual["PH30"] = result.Phases[2]
			}
			actual["TO30"] = math.NaN()
			if len(result.TroughOffsets) > 2 {
				actual["TO30"] = result.TroughOffsets[2]
			}
			actual["A40"] = math.NaN()
			if len(result.Amplitude) > 3 {
				actual["A40"] = result.Amplitude[3]
			}
			actual["PH40"] = math.NaN()
			if len(result.Phases) > 3 {
				actual["PH40"] = result.Phases[3]
			}
			actual["TO40"] = math.NaN()
			if len(result.TroughOffsets) > 3 {
				actual["TO40"] = result.TroughOffsets[3]
			}
			actual["A50"] = math.NaN()
			if len(result.Amplitude) > 4 {
				actual["A50"] = result.Amplitude[4]
			}
			actual["PH50"] = math.NaN()
			if len(result.Phases) > 4 {
				actual["PH50"] = result.Phases[4]
			}
			actual["TO50"] = math.NaN()
			if len(result.TroughOffsets) > 4 {
				actual["TO50"] = result.TroughOffsets[4]
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
	t.Run("case6", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 252
		p.MinPeriod = 8
		p.MaxPeriod = 32
		p.Step = 8
		p.Smoothing = 160.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A8": expectedW252_MIN8_MAX32_S8_L160_A8, "PH8": expectedW252_MIN8_MAX32_S8_L160_PH8, "TO8": expectedW252_MIN8_MAX32_S8_L160_TO8, "A16": expectedW252_MIN8_MAX32_S8_L160_A16, "PH16": expectedW252_MIN8_MAX32_S8_L160_PH16, "TO16": expectedW252_MIN8_MAX32_S8_L160_TO16, "A24": expectedW252_MIN8_MAX32_S8_L160_A24, "PH24": expectedW252_MIN8_MAX32_S8_L160_PH24, "TO24": expectedW252_MIN8_MAX32_S8_L160_TO24, "A32": expectedW252_MIN8_MAX32_S8_L160_A32, "PH32": expectedW252_MIN8_MAX32_S8_L160_PH32, "TO32": expectedW252_MIN8_MAX32_S8_L160_TO32, "DOM": expectedW252_MIN8_MAX32_S8_L160_DOM, "AMP": expectedW252_MIN8_MAX32_S8_L160_AMP, "PHASE": expectedW252_MIN8_MAX32_S8_L160_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A8"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A8"] = result.Amplitude[0]
			}
			actual["PH8"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH8"] = result.Phases[0]
			}
			actual["TO8"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO8"] = result.TroughOffsets[0]
			}
			actual["A16"] = math.NaN()
			if len(result.Amplitude) > 1 {
				actual["A16"] = result.Amplitude[1]
			}
			actual["PH16"] = math.NaN()
			if len(result.Phases) > 1 {
				actual["PH16"] = result.Phases[1]
			}
			actual["TO16"] = math.NaN()
			if len(result.TroughOffsets) > 1 {
				actual["TO16"] = result.TroughOffsets[1]
			}
			actual["A24"] = math.NaN()
			if len(result.Amplitude) > 2 {
				actual["A24"] = result.Amplitude[2]
			}
			actual["PH24"] = math.NaN()
			if len(result.Phases) > 2 {
				actual["PH24"] = result.Phases[2]
			}
			actual["TO24"] = math.NaN()
			if len(result.TroughOffsets) > 2 {
				actual["TO24"] = result.TroughOffsets[2]
			}
			actual["A32"] = math.NaN()
			if len(result.Amplitude) > 3 {
				actual["A32"] = result.Amplitude[3]
			}
			actual["PH32"] = math.NaN()
			if len(result.Phases) > 3 {
				actual["PH32"] = result.Phases[3]
			}
			actual["TO32"] = math.NaN()
			if len(result.TroughOffsets) > 3 {
				actual["TO32"] = result.TroughOffsets[3]
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
	t.Run("case7", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 253
		p.MinPeriod = 10
		p.MaxPeriod = 30
		p.Step = 10
		p.Smoothing = 1600.0
		indicator, err := NewGoertzelCycleSpectrum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"A10": expectedW253_MIN10_MAX30_S10_L1600_A10, "PH10": expectedW253_MIN10_MAX30_S10_L1600_PH10, "TO10": expectedW253_MIN10_MAX30_S10_L1600_TO10, "A20": expectedW253_MIN10_MAX30_S10_L1600_A20, "PH20": expectedW253_MIN10_MAX30_S10_L1600_PH20, "TO20": expectedW253_MIN10_MAX30_S10_L1600_TO20, "A30": expectedW253_MIN10_MAX30_S10_L1600_A30, "PH30": expectedW253_MIN10_MAX30_S10_L1600_PH30, "TO30": expectedW253_MIN10_MAX30_S10_L1600_TO30, "DOM": expectedW253_MIN10_MAX30_S10_L1600_DOM, "AMP": expectedW253_MIN10_MAX30_S10_L1600_AMP, "PHASE": expectedW253_MIN10_MAX30_S10_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"DOM": result.DominantPeriod, "AMP": result.DominantAmplitude, "PHASE": result.DominantPhase}
			actual["A10"] = math.NaN()
			if len(result.Amplitude) > 0 {
				actual["A10"] = result.Amplitude[0]
			}
			actual["PH10"] = math.NaN()
			if len(result.Phases) > 0 {
				actual["PH10"] = result.Phases[0]
			}
			actual["TO10"] = math.NaN()
			if len(result.TroughOffsets) > 0 {
				actual["TO10"] = result.TroughOffsets[0]
			}
			actual["A20"] = math.NaN()
			if len(result.Amplitude) > 1 {
				actual["A20"] = result.Amplitude[1]
			}
			actual["PH20"] = math.NaN()
			if len(result.Phases) > 1 {
				actual["PH20"] = result.Phases[1]
			}
			actual["TO20"] = math.NaN()
			if len(result.TroughOffsets) > 1 {
				actual["TO20"] = result.TroughOffsets[1]
			}
			actual["A30"] = math.NaN()
			if len(result.Amplitude) > 2 {
				actual["A30"] = result.Amplitude[2]
			}
			actual["PH30"] = math.NaN()
			if len(result.Phases) > 2 {
				actual["PH30"] = result.Phases[2]
			}
			actual["TO30"] = math.NaN()
			if len(result.TroughOffsets) > 2 {
				actual["TO30"] = result.TroughOffsets[2]
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
