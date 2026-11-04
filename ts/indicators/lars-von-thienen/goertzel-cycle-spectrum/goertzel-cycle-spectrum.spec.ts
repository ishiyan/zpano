import { GoertzelCycleSpectrum } from "./goertzel-cycle-spectrum.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen GoertzelCycleSpectrum", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "window": 128, "minPeriod": 5, "maxPeriod": 50, "step": 1, "smoothing": 1600.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A5": data.expectedW128_MIN5_MAX50_S1_L1600_A5, "PH5": data.expectedW128_MIN5_MAX50_S1_L1600_PH5, "TO5": data.expectedW128_MIN5_MAX50_S1_L1600_TO5, "A6": data.expectedW128_MIN5_MAX50_S1_L1600_A6, "PH6": data.expectedW128_MIN5_MAX50_S1_L1600_PH6, "TO6": data.expectedW128_MIN5_MAX50_S1_L1600_TO6, "A7": data.expectedW128_MIN5_MAX50_S1_L1600_A7, "PH7": data.expectedW128_MIN5_MAX50_S1_L1600_PH7, "TO7": data.expectedW128_MIN5_MAX50_S1_L1600_TO7, "A8": data.expectedW128_MIN5_MAX50_S1_L1600_A8, "PH8": data.expectedW128_MIN5_MAX50_S1_L1600_PH8, "TO8": data.expectedW128_MIN5_MAX50_S1_L1600_TO8, "A9": data.expectedW128_MIN5_MAX50_S1_L1600_A9, "PH9": data.expectedW128_MIN5_MAX50_S1_L1600_PH9, "TO9": data.expectedW128_MIN5_MAX50_S1_L1600_TO9, "A10": data.expectedW128_MIN5_MAX50_S1_L1600_A10, "PH10": data.expectedW128_MIN5_MAX50_S1_L1600_PH10, "TO10": data.expectedW128_MIN5_MAX50_S1_L1600_TO10, "A11": data.expectedW128_MIN5_MAX50_S1_L1600_A11, "PH11": data.expectedW128_MIN5_MAX50_S1_L1600_PH11, "TO11": data.expectedW128_MIN5_MAX50_S1_L1600_TO11, "A12": data.expectedW128_MIN5_MAX50_S1_L1600_A12, "PH12": data.expectedW128_MIN5_MAX50_S1_L1600_PH12, "TO12": data.expectedW128_MIN5_MAX50_S1_L1600_TO12, "A13": data.expectedW128_MIN5_MAX50_S1_L1600_A13, "PH13": data.expectedW128_MIN5_MAX50_S1_L1600_PH13, "TO13": data.expectedW128_MIN5_MAX50_S1_L1600_TO13, "A14": data.expectedW128_MIN5_MAX50_S1_L1600_A14, "PH14": data.expectedW128_MIN5_MAX50_S1_L1600_PH14, "TO14": data.expectedW128_MIN5_MAX50_S1_L1600_TO14, "A15": data.expectedW128_MIN5_MAX50_S1_L1600_A15, "PH15": data.expectedW128_MIN5_MAX50_S1_L1600_PH15, "TO15": data.expectedW128_MIN5_MAX50_S1_L1600_TO15, "A16": data.expectedW128_MIN5_MAX50_S1_L1600_A16, "PH16": data.expectedW128_MIN5_MAX50_S1_L1600_PH16, "TO16": data.expectedW128_MIN5_MAX50_S1_L1600_TO16, "A17": data.expectedW128_MIN5_MAX50_S1_L1600_A17, "PH17": data.expectedW128_MIN5_MAX50_S1_L1600_PH17, "TO17": data.expectedW128_MIN5_MAX50_S1_L1600_TO17, "A18": data.expectedW128_MIN5_MAX50_S1_L1600_A18, "PH18": data.expectedW128_MIN5_MAX50_S1_L1600_PH18, "TO18": data.expectedW128_MIN5_MAX50_S1_L1600_TO18, "A19": data.expectedW128_MIN5_MAX50_S1_L1600_A19, "PH19": data.expectedW128_MIN5_MAX50_S1_L1600_PH19, "TO19": data.expectedW128_MIN5_MAX50_S1_L1600_TO19, "A20": data.expectedW128_MIN5_MAX50_S1_L1600_A20, "PH20": data.expectedW128_MIN5_MAX50_S1_L1600_PH20, "TO20": data.expectedW128_MIN5_MAX50_S1_L1600_TO20, "A21": data.expectedW128_MIN5_MAX50_S1_L1600_A21, "PH21": data.expectedW128_MIN5_MAX50_S1_L1600_PH21, "TO21": data.expectedW128_MIN5_MAX50_S1_L1600_TO21, "A22": data.expectedW128_MIN5_MAX50_S1_L1600_A22, "PH22": data.expectedW128_MIN5_MAX50_S1_L1600_PH22, "TO22": data.expectedW128_MIN5_MAX50_S1_L1600_TO22, "A23": data.expectedW128_MIN5_MAX50_S1_L1600_A23, "PH23": data.expectedW128_MIN5_MAX50_S1_L1600_PH23, "TO23": data.expectedW128_MIN5_MAX50_S1_L1600_TO23, "A24": data.expectedW128_MIN5_MAX50_S1_L1600_A24, "PH24": data.expectedW128_MIN5_MAX50_S1_L1600_PH24, "TO24": data.expectedW128_MIN5_MAX50_S1_L1600_TO24, "A25": data.expectedW128_MIN5_MAX50_S1_L1600_A25, "PH25": data.expectedW128_MIN5_MAX50_S1_L1600_PH25, "TO25": data.expectedW128_MIN5_MAX50_S1_L1600_TO25, "A26": data.expectedW128_MIN5_MAX50_S1_L1600_A26, "PH26": data.expectedW128_MIN5_MAX50_S1_L1600_PH26, "TO26": data.expectedW128_MIN5_MAX50_S1_L1600_TO26, "A27": data.expectedW128_MIN5_MAX50_S1_L1600_A27, "PH27": data.expectedW128_MIN5_MAX50_S1_L1600_PH27, "TO27": data.expectedW128_MIN5_MAX50_S1_L1600_TO27, "A28": data.expectedW128_MIN5_MAX50_S1_L1600_A28, "PH28": data.expectedW128_MIN5_MAX50_S1_L1600_PH28, "TO28": data.expectedW128_MIN5_MAX50_S1_L1600_TO28, "A29": data.expectedW128_MIN5_MAX50_S1_L1600_A29, "PH29": data.expectedW128_MIN5_MAX50_S1_L1600_PH29, "TO29": data.expectedW128_MIN5_MAX50_S1_L1600_TO29, "A30": data.expectedW128_MIN5_MAX50_S1_L1600_A30, "PH30": data.expectedW128_MIN5_MAX50_S1_L1600_PH30, "TO30": data.expectedW128_MIN5_MAX50_S1_L1600_TO30, "A31": data.expectedW128_MIN5_MAX50_S1_L1600_A31, "PH31": data.expectedW128_MIN5_MAX50_S1_L1600_PH31, "TO31": data.expectedW128_MIN5_MAX50_S1_L1600_TO31, "A32": data.expectedW128_MIN5_MAX50_S1_L1600_A32, "PH32": data.expectedW128_MIN5_MAX50_S1_L1600_PH32, "TO32": data.expectedW128_MIN5_MAX50_S1_L1600_TO32, "A33": data.expectedW128_MIN5_MAX50_S1_L1600_A33, "PH33": data.expectedW128_MIN5_MAX50_S1_L1600_PH33, "TO33": data.expectedW128_MIN5_MAX50_S1_L1600_TO33, "A34": data.expectedW128_MIN5_MAX50_S1_L1600_A34, "PH34": data.expectedW128_MIN5_MAX50_S1_L1600_PH34, "TO34": data.expectedW128_MIN5_MAX50_S1_L1600_TO34, "A35": data.expectedW128_MIN5_MAX50_S1_L1600_A35, "PH35": data.expectedW128_MIN5_MAX50_S1_L1600_PH35, "TO35": data.expectedW128_MIN5_MAX50_S1_L1600_TO35, "A36": data.expectedW128_MIN5_MAX50_S1_L1600_A36, "PH36": data.expectedW128_MIN5_MAX50_S1_L1600_PH36, "TO36": data.expectedW128_MIN5_MAX50_S1_L1600_TO36, "A37": data.expectedW128_MIN5_MAX50_S1_L1600_A37, "PH37": data.expectedW128_MIN5_MAX50_S1_L1600_PH37, "TO37": data.expectedW128_MIN5_MAX50_S1_L1600_TO37, "A38": data.expectedW128_MIN5_MAX50_S1_L1600_A38, "PH38": data.expectedW128_MIN5_MAX50_S1_L1600_PH38, "TO38": data.expectedW128_MIN5_MAX50_S1_L1600_TO38, "A39": data.expectedW128_MIN5_MAX50_S1_L1600_A39, "PH39": data.expectedW128_MIN5_MAX50_S1_L1600_PH39, "TO39": data.expectedW128_MIN5_MAX50_S1_L1600_TO39, "A40": data.expectedW128_MIN5_MAX50_S1_L1600_A40, "PH40": data.expectedW128_MIN5_MAX50_S1_L1600_PH40, "TO40": data.expectedW128_MIN5_MAX50_S1_L1600_TO40, "A41": data.expectedW128_MIN5_MAX50_S1_L1600_A41, "PH41": data.expectedW128_MIN5_MAX50_S1_L1600_PH41, "TO41": data.expectedW128_MIN5_MAX50_S1_L1600_TO41, "A42": data.expectedW128_MIN5_MAX50_S1_L1600_A42, "PH42": data.expectedW128_MIN5_MAX50_S1_L1600_PH42, "TO42": data.expectedW128_MIN5_MAX50_S1_L1600_TO42, "A43": data.expectedW128_MIN5_MAX50_S1_L1600_A43, "PH43": data.expectedW128_MIN5_MAX50_S1_L1600_PH43, "TO43": data.expectedW128_MIN5_MAX50_S1_L1600_TO43, "A44": data.expectedW128_MIN5_MAX50_S1_L1600_A44, "PH44": data.expectedW128_MIN5_MAX50_S1_L1600_PH44, "TO44": data.expectedW128_MIN5_MAX50_S1_L1600_TO44, "A45": data.expectedW128_MIN5_MAX50_S1_L1600_A45, "PH45": data.expectedW128_MIN5_MAX50_S1_L1600_PH45, "TO45": data.expectedW128_MIN5_MAX50_S1_L1600_TO45, "A46": data.expectedW128_MIN5_MAX50_S1_L1600_A46, "PH46": data.expectedW128_MIN5_MAX50_S1_L1600_PH46, "TO46": data.expectedW128_MIN5_MAX50_S1_L1600_TO46, "A47": data.expectedW128_MIN5_MAX50_S1_L1600_A47, "PH47": data.expectedW128_MIN5_MAX50_S1_L1600_PH47, "TO47": data.expectedW128_MIN5_MAX50_S1_L1600_TO47, "A48": data.expectedW128_MIN5_MAX50_S1_L1600_A48, "PH48": data.expectedW128_MIN5_MAX50_S1_L1600_PH48, "TO48": data.expectedW128_MIN5_MAX50_S1_L1600_TO48, "A49": data.expectedW128_MIN5_MAX50_S1_L1600_A49, "PH49": data.expectedW128_MIN5_MAX50_S1_L1600_PH49, "TO49": data.expectedW128_MIN5_MAX50_S1_L1600_TO49, "A50": data.expectedW128_MIN5_MAX50_S1_L1600_A50, "PH50": data.expectedW128_MIN5_MAX50_S1_L1600_PH50, "TO50": data.expectedW128_MIN5_MAX50_S1_L1600_TO50, "DOM": data.expectedW128_MIN5_MAX50_S1_L1600_DOM, "AMP": data.expectedW128_MIN5_MAX50_S1_L1600_AMP, "PHASE": data.expectedW128_MIN5_MAX50_S1_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A5"] = result.amplitude[0] ?? NaN;
            actual["PH5"] = result.phases[0] ?? NaN;
            actual["TO5"] = result.troughOffsets[0] ?? NaN;
            actual["A6"] = result.amplitude[1] ?? NaN;
            actual["PH6"] = result.phases[1] ?? NaN;
            actual["TO6"] = result.troughOffsets[1] ?? NaN;
            actual["A7"] = result.amplitude[2] ?? NaN;
            actual["PH7"] = result.phases[2] ?? NaN;
            actual["TO7"] = result.troughOffsets[2] ?? NaN;
            actual["A8"] = result.amplitude[3] ?? NaN;
            actual["PH8"] = result.phases[3] ?? NaN;
            actual["TO8"] = result.troughOffsets[3] ?? NaN;
            actual["A9"] = result.amplitude[4] ?? NaN;
            actual["PH9"] = result.phases[4] ?? NaN;
            actual["TO9"] = result.troughOffsets[4] ?? NaN;
            actual["A10"] = result.amplitude[5] ?? NaN;
            actual["PH10"] = result.phases[5] ?? NaN;
            actual["TO10"] = result.troughOffsets[5] ?? NaN;
            actual["A11"] = result.amplitude[6] ?? NaN;
            actual["PH11"] = result.phases[6] ?? NaN;
            actual["TO11"] = result.troughOffsets[6] ?? NaN;
            actual["A12"] = result.amplitude[7] ?? NaN;
            actual["PH12"] = result.phases[7] ?? NaN;
            actual["TO12"] = result.troughOffsets[7] ?? NaN;
            actual["A13"] = result.amplitude[8] ?? NaN;
            actual["PH13"] = result.phases[8] ?? NaN;
            actual["TO13"] = result.troughOffsets[8] ?? NaN;
            actual["A14"] = result.amplitude[9] ?? NaN;
            actual["PH14"] = result.phases[9] ?? NaN;
            actual["TO14"] = result.troughOffsets[9] ?? NaN;
            actual["A15"] = result.amplitude[10] ?? NaN;
            actual["PH15"] = result.phases[10] ?? NaN;
            actual["TO15"] = result.troughOffsets[10] ?? NaN;
            actual["A16"] = result.amplitude[11] ?? NaN;
            actual["PH16"] = result.phases[11] ?? NaN;
            actual["TO16"] = result.troughOffsets[11] ?? NaN;
            actual["A17"] = result.amplitude[12] ?? NaN;
            actual["PH17"] = result.phases[12] ?? NaN;
            actual["TO17"] = result.troughOffsets[12] ?? NaN;
            actual["A18"] = result.amplitude[13] ?? NaN;
            actual["PH18"] = result.phases[13] ?? NaN;
            actual["TO18"] = result.troughOffsets[13] ?? NaN;
            actual["A19"] = result.amplitude[14] ?? NaN;
            actual["PH19"] = result.phases[14] ?? NaN;
            actual["TO19"] = result.troughOffsets[14] ?? NaN;
            actual["A20"] = result.amplitude[15] ?? NaN;
            actual["PH20"] = result.phases[15] ?? NaN;
            actual["TO20"] = result.troughOffsets[15] ?? NaN;
            actual["A21"] = result.amplitude[16] ?? NaN;
            actual["PH21"] = result.phases[16] ?? NaN;
            actual["TO21"] = result.troughOffsets[16] ?? NaN;
            actual["A22"] = result.amplitude[17] ?? NaN;
            actual["PH22"] = result.phases[17] ?? NaN;
            actual["TO22"] = result.troughOffsets[17] ?? NaN;
            actual["A23"] = result.amplitude[18] ?? NaN;
            actual["PH23"] = result.phases[18] ?? NaN;
            actual["TO23"] = result.troughOffsets[18] ?? NaN;
            actual["A24"] = result.amplitude[19] ?? NaN;
            actual["PH24"] = result.phases[19] ?? NaN;
            actual["TO24"] = result.troughOffsets[19] ?? NaN;
            actual["A25"] = result.amplitude[20] ?? NaN;
            actual["PH25"] = result.phases[20] ?? NaN;
            actual["TO25"] = result.troughOffsets[20] ?? NaN;
            actual["A26"] = result.amplitude[21] ?? NaN;
            actual["PH26"] = result.phases[21] ?? NaN;
            actual["TO26"] = result.troughOffsets[21] ?? NaN;
            actual["A27"] = result.amplitude[22] ?? NaN;
            actual["PH27"] = result.phases[22] ?? NaN;
            actual["TO27"] = result.troughOffsets[22] ?? NaN;
            actual["A28"] = result.amplitude[23] ?? NaN;
            actual["PH28"] = result.phases[23] ?? NaN;
            actual["TO28"] = result.troughOffsets[23] ?? NaN;
            actual["A29"] = result.amplitude[24] ?? NaN;
            actual["PH29"] = result.phases[24] ?? NaN;
            actual["TO29"] = result.troughOffsets[24] ?? NaN;
            actual["A30"] = result.amplitude[25] ?? NaN;
            actual["PH30"] = result.phases[25] ?? NaN;
            actual["TO30"] = result.troughOffsets[25] ?? NaN;
            actual["A31"] = result.amplitude[26] ?? NaN;
            actual["PH31"] = result.phases[26] ?? NaN;
            actual["TO31"] = result.troughOffsets[26] ?? NaN;
            actual["A32"] = result.amplitude[27] ?? NaN;
            actual["PH32"] = result.phases[27] ?? NaN;
            actual["TO32"] = result.troughOffsets[27] ?? NaN;
            actual["A33"] = result.amplitude[28] ?? NaN;
            actual["PH33"] = result.phases[28] ?? NaN;
            actual["TO33"] = result.troughOffsets[28] ?? NaN;
            actual["A34"] = result.amplitude[29] ?? NaN;
            actual["PH34"] = result.phases[29] ?? NaN;
            actual["TO34"] = result.troughOffsets[29] ?? NaN;
            actual["A35"] = result.amplitude[30] ?? NaN;
            actual["PH35"] = result.phases[30] ?? NaN;
            actual["TO35"] = result.troughOffsets[30] ?? NaN;
            actual["A36"] = result.amplitude[31] ?? NaN;
            actual["PH36"] = result.phases[31] ?? NaN;
            actual["TO36"] = result.troughOffsets[31] ?? NaN;
            actual["A37"] = result.amplitude[32] ?? NaN;
            actual["PH37"] = result.phases[32] ?? NaN;
            actual["TO37"] = result.troughOffsets[32] ?? NaN;
            actual["A38"] = result.amplitude[33] ?? NaN;
            actual["PH38"] = result.phases[33] ?? NaN;
            actual["TO38"] = result.troughOffsets[33] ?? NaN;
            actual["A39"] = result.amplitude[34] ?? NaN;
            actual["PH39"] = result.phases[34] ?? NaN;
            actual["TO39"] = result.troughOffsets[34] ?? NaN;
            actual["A40"] = result.amplitude[35] ?? NaN;
            actual["PH40"] = result.phases[35] ?? NaN;
            actual["TO40"] = result.troughOffsets[35] ?? NaN;
            actual["A41"] = result.amplitude[36] ?? NaN;
            actual["PH41"] = result.phases[36] ?? NaN;
            actual["TO41"] = result.troughOffsets[36] ?? NaN;
            actual["A42"] = result.amplitude[37] ?? NaN;
            actual["PH42"] = result.phases[37] ?? NaN;
            actual["TO42"] = result.troughOffsets[37] ?? NaN;
            actual["A43"] = result.amplitude[38] ?? NaN;
            actual["PH43"] = result.phases[38] ?? NaN;
            actual["TO43"] = result.troughOffsets[38] ?? NaN;
            actual["A44"] = result.amplitude[39] ?? NaN;
            actual["PH44"] = result.phases[39] ?? NaN;
            actual["TO44"] = result.troughOffsets[39] ?? NaN;
            actual["A45"] = result.amplitude[40] ?? NaN;
            actual["PH45"] = result.phases[40] ?? NaN;
            actual["TO45"] = result.troughOffsets[40] ?? NaN;
            actual["A46"] = result.amplitude[41] ?? NaN;
            actual["PH46"] = result.phases[41] ?? NaN;
            actual["TO46"] = result.troughOffsets[41] ?? NaN;
            actual["A47"] = result.amplitude[42] ?? NaN;
            actual["PH47"] = result.phases[42] ?? NaN;
            actual["TO47"] = result.troughOffsets[42] ?? NaN;
            actual["A48"] = result.amplitude[43] ?? NaN;
            actual["PH48"] = result.phases[43] ?? NaN;
            actual["TO48"] = result.troughOffsets[43] ?? NaN;
            actual["A49"] = result.amplitude[44] ?? NaN;
            actual["PH49"] = result.phases[44] ?? NaN;
            actual["TO49"] = result.troughOffsets[44] ?? NaN;
            actual["A50"] = result.amplitude[45] ?? NaN;
            actual["PH50"] = result.phases[45] ?? NaN;
            actual["TO50"] = result.troughOffsets[45] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 1", () => {
        const parameters = { ...defaultParams(), ...{ "window": 100, "minPeriod": 20, "maxPeriod": 20, "step": 1, "smoothing": 0.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A20": data.expectedW100_MIN20_MAX20_S1_L0_A20, "PH20": data.expectedW100_MIN20_MAX20_S1_L0_PH20, "TO20": data.expectedW100_MIN20_MAX20_S1_L0_TO20, "DOM": data.expectedW100_MIN20_MAX20_S1_L0_DOM, "AMP": data.expectedW100_MIN20_MAX20_S1_L0_AMP, "PHASE": data.expectedW100_MIN20_MAX20_S1_L0_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A20"] = result.amplitude[0] ?? NaN;
            actual["PH20"] = result.phases[0] ?? NaN;
            actual["TO20"] = result.troughOffsets[0] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 2", () => {
        const parameters = { ...defaultParams(), ...{ "window": 40, "minPeriod": 5, "maxPeriod": 25, "step": 5, "smoothing": 0.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A5": data.expectedW40_MIN5_MAX25_S5_L0_A5, "PH5": data.expectedW40_MIN5_MAX25_S5_L0_PH5, "TO5": data.expectedW40_MIN5_MAX25_S5_L0_TO5, "A10": data.expectedW40_MIN5_MAX25_S5_L0_A10, "PH10": data.expectedW40_MIN5_MAX25_S5_L0_PH10, "TO10": data.expectedW40_MIN5_MAX25_S5_L0_TO10, "A15": data.expectedW40_MIN5_MAX25_S5_L0_A15, "PH15": data.expectedW40_MIN5_MAX25_S5_L0_PH15, "TO15": data.expectedW40_MIN5_MAX25_S5_L0_TO15, "A20": data.expectedW40_MIN5_MAX25_S5_L0_A20, "PH20": data.expectedW40_MIN5_MAX25_S5_L0_PH20, "TO20": data.expectedW40_MIN5_MAX25_S5_L0_TO20, "A25": data.expectedW40_MIN5_MAX25_S5_L0_A25, "PH25": data.expectedW40_MIN5_MAX25_S5_L0_PH25, "TO25": data.expectedW40_MIN5_MAX25_S5_L0_TO25, "DOM": data.expectedW40_MIN5_MAX25_S5_L0_DOM, "AMP": data.expectedW40_MIN5_MAX25_S5_L0_AMP, "PHASE": data.expectedW40_MIN5_MAX25_S5_L0_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A5"] = result.amplitude[0] ?? NaN;
            actual["PH5"] = result.phases[0] ?? NaN;
            actual["TO5"] = result.troughOffsets[0] ?? NaN;
            actual["A10"] = result.amplitude[1] ?? NaN;
            actual["PH10"] = result.phases[1] ?? NaN;
            actual["TO10"] = result.troughOffsets[1] ?? NaN;
            actual["A15"] = result.amplitude[2] ?? NaN;
            actual["PH15"] = result.phases[2] ?? NaN;
            actual["TO15"] = result.troughOffsets[2] ?? NaN;
            actual["A20"] = result.amplitude[3] ?? NaN;
            actual["PH20"] = result.phases[3] ?? NaN;
            actual["TO20"] = result.troughOffsets[3] ?? NaN;
            actual["A25"] = result.amplitude[4] ?? NaN;
            actual["PH25"] = result.phases[4] ?? NaN;
            actual["TO25"] = result.troughOffsets[4] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 3", () => {
        const parameters = { ...defaultParams(), ...{ "window": 40, "minPeriod": 5, "maxPeriod": 25, "step": 5, "smoothing": 1600.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A5": data.expectedW40_MIN5_MAX25_S5_L1600_A5, "PH5": data.expectedW40_MIN5_MAX25_S5_L1600_PH5, "TO5": data.expectedW40_MIN5_MAX25_S5_L1600_TO5, "A10": data.expectedW40_MIN5_MAX25_S5_L1600_A10, "PH10": data.expectedW40_MIN5_MAX25_S5_L1600_PH10, "TO10": data.expectedW40_MIN5_MAX25_S5_L1600_TO10, "A15": data.expectedW40_MIN5_MAX25_S5_L1600_A15, "PH15": data.expectedW40_MIN5_MAX25_S5_L1600_PH15, "TO15": data.expectedW40_MIN5_MAX25_S5_L1600_TO15, "A20": data.expectedW40_MIN5_MAX25_S5_L1600_A20, "PH20": data.expectedW40_MIN5_MAX25_S5_L1600_PH20, "TO20": data.expectedW40_MIN5_MAX25_S5_L1600_TO20, "A25": data.expectedW40_MIN5_MAX25_S5_L1600_A25, "PH25": data.expectedW40_MIN5_MAX25_S5_L1600_PH25, "TO25": data.expectedW40_MIN5_MAX25_S5_L1600_TO25, "DOM": data.expectedW40_MIN5_MAX25_S5_L1600_DOM, "AMP": data.expectedW40_MIN5_MAX25_S5_L1600_AMP, "PHASE": data.expectedW40_MIN5_MAX25_S5_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A5"] = result.amplitude[0] ?? NaN;
            actual["PH5"] = result.phases[0] ?? NaN;
            actual["TO5"] = result.troughOffsets[0] ?? NaN;
            actual["A10"] = result.amplitude[1] ?? NaN;
            actual["PH10"] = result.phases[1] ?? NaN;
            actual["TO10"] = result.troughOffsets[1] ?? NaN;
            actual["A15"] = result.amplitude[2] ?? NaN;
            actual["PH15"] = result.phases[2] ?? NaN;
            actual["TO15"] = result.troughOffsets[2] ?? NaN;
            actual["A20"] = result.amplitude[3] ?? NaN;
            actual["PH20"] = result.phases[3] ?? NaN;
            actual["TO20"] = result.troughOffsets[3] ?? NaN;
            actual["A25"] = result.amplitude[4] ?? NaN;
            actual["PH25"] = result.phases[4] ?? NaN;
            actual["TO25"] = result.troughOffsets[4] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 4", () => {
        const parameters = { ...defaultParams(), ...{ "window": 200, "minPeriod": 10, "maxPeriod": 50, "step": 10, "smoothing": 1600.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A10": data.expectedW200_MIN10_MAX50_S10_L1600_A10, "PH10": data.expectedW200_MIN10_MAX50_S10_L1600_PH10, "TO10": data.expectedW200_MIN10_MAX50_S10_L1600_TO10, "A20": data.expectedW200_MIN10_MAX50_S10_L1600_A20, "PH20": data.expectedW200_MIN10_MAX50_S10_L1600_PH20, "TO20": data.expectedW200_MIN10_MAX50_S10_L1600_TO20, "A30": data.expectedW200_MIN10_MAX50_S10_L1600_A30, "PH30": data.expectedW200_MIN10_MAX50_S10_L1600_PH30, "TO30": data.expectedW200_MIN10_MAX50_S10_L1600_TO30, "A40": data.expectedW200_MIN10_MAX50_S10_L1600_A40, "PH40": data.expectedW200_MIN10_MAX50_S10_L1600_PH40, "TO40": data.expectedW200_MIN10_MAX50_S10_L1600_TO40, "A50": data.expectedW200_MIN10_MAX50_S10_L1600_A50, "PH50": data.expectedW200_MIN10_MAX50_S10_L1600_PH50, "TO50": data.expectedW200_MIN10_MAX50_S10_L1600_TO50, "DOM": data.expectedW200_MIN10_MAX50_S10_L1600_DOM, "AMP": data.expectedW200_MIN10_MAX50_S10_L1600_AMP, "PHASE": data.expectedW200_MIN10_MAX50_S10_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A10"] = result.amplitude[0] ?? NaN;
            actual["PH10"] = result.phases[0] ?? NaN;
            actual["TO10"] = result.troughOffsets[0] ?? NaN;
            actual["A20"] = result.amplitude[1] ?? NaN;
            actual["PH20"] = result.phases[1] ?? NaN;
            actual["TO20"] = result.troughOffsets[1] ?? NaN;
            actual["A30"] = result.amplitude[2] ?? NaN;
            actual["PH30"] = result.phases[2] ?? NaN;
            actual["TO30"] = result.troughOffsets[2] ?? NaN;
            actual["A40"] = result.amplitude[3] ?? NaN;
            actual["PH40"] = result.phases[3] ?? NaN;
            actual["TO40"] = result.troughOffsets[3] ?? NaN;
            actual["A50"] = result.amplitude[4] ?? NaN;
            actual["PH50"] = result.phases[4] ?? NaN;
            actual["TO50"] = result.troughOffsets[4] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 5", () => {
        const parameters = { ...defaultParams(), ...{ "window": 20, "minPeriod": 10, "maxPeriod": 50, "step": 10, "smoothing": 1600.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A10": data.expectedW20_MIN10_MAX50_S10_L1600_A10, "PH10": data.expectedW20_MIN10_MAX50_S10_L1600_PH10, "TO10": data.expectedW20_MIN10_MAX50_S10_L1600_TO10, "A20": data.expectedW20_MIN10_MAX50_S10_L1600_A20, "PH20": data.expectedW20_MIN10_MAX50_S10_L1600_PH20, "TO20": data.expectedW20_MIN10_MAX50_S10_L1600_TO20, "A30": data.expectedW20_MIN10_MAX50_S10_L1600_A30, "PH30": data.expectedW20_MIN10_MAX50_S10_L1600_PH30, "TO30": data.expectedW20_MIN10_MAX50_S10_L1600_TO30, "A40": data.expectedW20_MIN10_MAX50_S10_L1600_A40, "PH40": data.expectedW20_MIN10_MAX50_S10_L1600_PH40, "TO40": data.expectedW20_MIN10_MAX50_S10_L1600_TO40, "A50": data.expectedW20_MIN10_MAX50_S10_L1600_A50, "PH50": data.expectedW20_MIN10_MAX50_S10_L1600_PH50, "TO50": data.expectedW20_MIN10_MAX50_S10_L1600_TO50, "DOM": data.expectedW20_MIN10_MAX50_S10_L1600_DOM, "AMP": data.expectedW20_MIN10_MAX50_S10_L1600_AMP, "PHASE": data.expectedW20_MIN10_MAX50_S10_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A10"] = result.amplitude[0] ?? NaN;
            actual["PH10"] = result.phases[0] ?? NaN;
            actual["TO10"] = result.troughOffsets[0] ?? NaN;
            actual["A20"] = result.amplitude[1] ?? NaN;
            actual["PH20"] = result.phases[1] ?? NaN;
            actual["TO20"] = result.troughOffsets[1] ?? NaN;
            actual["A30"] = result.amplitude[2] ?? NaN;
            actual["PH30"] = result.phases[2] ?? NaN;
            actual["TO30"] = result.troughOffsets[2] ?? NaN;
            actual["A40"] = result.amplitude[3] ?? NaN;
            actual["PH40"] = result.phases[3] ?? NaN;
            actual["TO40"] = result.troughOffsets[3] ?? NaN;
            actual["A50"] = result.amplitude[4] ?? NaN;
            actual["PH50"] = result.phases[4] ?? NaN;
            actual["TO50"] = result.troughOffsets[4] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 6", () => {
        const parameters = { ...defaultParams(), ...{ "window": 252, "minPeriod": 8, "maxPeriod": 32, "step": 8, "smoothing": 160.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A8": data.expectedW252_MIN8_MAX32_S8_L160_A8, "PH8": data.expectedW252_MIN8_MAX32_S8_L160_PH8, "TO8": data.expectedW252_MIN8_MAX32_S8_L160_TO8, "A16": data.expectedW252_MIN8_MAX32_S8_L160_A16, "PH16": data.expectedW252_MIN8_MAX32_S8_L160_PH16, "TO16": data.expectedW252_MIN8_MAX32_S8_L160_TO16, "A24": data.expectedW252_MIN8_MAX32_S8_L160_A24, "PH24": data.expectedW252_MIN8_MAX32_S8_L160_PH24, "TO24": data.expectedW252_MIN8_MAX32_S8_L160_TO24, "A32": data.expectedW252_MIN8_MAX32_S8_L160_A32, "PH32": data.expectedW252_MIN8_MAX32_S8_L160_PH32, "TO32": data.expectedW252_MIN8_MAX32_S8_L160_TO32, "DOM": data.expectedW252_MIN8_MAX32_S8_L160_DOM, "AMP": data.expectedW252_MIN8_MAX32_S8_L160_AMP, "PHASE": data.expectedW252_MIN8_MAX32_S8_L160_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A8"] = result.amplitude[0] ?? NaN;
            actual["PH8"] = result.phases[0] ?? NaN;
            actual["TO8"] = result.troughOffsets[0] ?? NaN;
            actual["A16"] = result.amplitude[1] ?? NaN;
            actual["PH16"] = result.phases[1] ?? NaN;
            actual["TO16"] = result.troughOffsets[1] ?? NaN;
            actual["A24"] = result.amplitude[2] ?? NaN;
            actual["PH24"] = result.phases[2] ?? NaN;
            actual["TO24"] = result.troughOffsets[2] ?? NaN;
            actual["A32"] = result.amplitude[3] ?? NaN;
            actual["PH32"] = result.phases[3] ?? NaN;
            actual["TO32"] = result.troughOffsets[3] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 7", () => {
        const parameters = { ...defaultParams(), ...{ "window": 253, "minPeriod": 10, "maxPeriod": 30, "step": 10, "smoothing": 1600.0 } };
        const indicator = new GoertzelCycleSpectrum(parameters);
        const expected: Record<string, number[]> = { "A10": data.expectedW253_MIN10_MAX30_S10_L1600_A10, "PH10": data.expectedW253_MIN10_MAX30_S10_L1600_PH10, "TO10": data.expectedW253_MIN10_MAX30_S10_L1600_TO10, "A20": data.expectedW253_MIN10_MAX30_S10_L1600_A20, "PH20": data.expectedW253_MIN10_MAX30_S10_L1600_PH20, "TO20": data.expectedW253_MIN10_MAX30_S10_L1600_TO20, "A30": data.expectedW253_MIN10_MAX30_S10_L1600_A30, "PH30": data.expectedW253_MIN10_MAX30_S10_L1600_PH30, "TO30": data.expectedW253_MIN10_MAX30_S10_L1600_TO30, "DOM": data.expectedW253_MIN10_MAX30_S10_L1600_DOM, "AMP": data.expectedW253_MIN10_MAX30_S10_L1600_AMP, "PHASE": data.expectedW253_MIN10_MAX30_S10_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "DOM": result.dominantPeriod, "AMP": result.dominantAmplitude, "PHASE": result.dominantPhase };
            actual["A10"] = result.amplitude[0] ?? NaN;
            actual["PH10"] = result.phases[0] ?? NaN;
            actual["TO10"] = result.troughOffsets[0] ?? NaN;
            actual["A20"] = result.amplitude[1] ?? NaN;
            actual["PH20"] = result.phases[1] ?? NaN;
            actual["TO20"] = result.troughOffsets[1] ?? NaN;
            actual["A30"] = result.amplitude[2] ?? NaN;
            actual["PH30"] = result.phases[2] ?? NaN;
            actual["TO30"] = result.troughOffsets[2] ?? NaN;
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("skips NaN and rejects invalid observations before priming", () => {
        const indicator = new GoertzelCycleSpectrum();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
