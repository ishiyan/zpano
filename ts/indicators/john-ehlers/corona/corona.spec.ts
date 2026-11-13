import { } from 'jasmine';

import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { createIndicator } from '../../factory/factory';
import { CoronaSignalToNoiseRatio } from '../corona-signal-to-noise-ratio/corona-signal-to-noise-ratio';
import { CoronaSpectrum } from '../corona-spectrum/corona-spectrum';
import { testHigh, testInput, testLow } from '../corona-spectrum/testdata';
import { CoronaSwingPosition } from '../corona-swing-position/corona-swing-position';
import { CoronaTrendVigor } from '../corona-trend-vigor/corona-trend-vigor';
import { AdaptiveBandPass, Corona, CoronaSnapshot, Raster, TWO_PI } from './corona';
import { DefaultCoronaParams } from './params';

const engineTolerance = 5e-12;

function sameFloat(actual: number, expected: number, tolerance = engineTolerance): boolean {
    if (Number.isNaN(expected)) {
        return Number.isNaN(actual);
    }
    if (!Number.isFinite(expected)) {
        return actual === expected;
    }
    if (!Number.isFinite(actual)) {
        return false;
    }
    const bound = Math.max(tolerance, Math.abs(expected) * tolerance);
    return Math.abs(actual - expected) <= bound;
}

function sameSlice(actual: readonly number[], expected: readonly number[], tolerance = engineTolerance): boolean {
    if (actual.length !== expected.length) {
        return false;
    }
    for (let i = 0; i < expected.length; i++) {
        if (!sameFloat(actual[i], expected[i], tolerance)) {
            return false;
        }
    }
    return true;
}

function sameHeatmap(actual: Heatmap, expected: Heatmap, tolerance = engineTolerance): boolean {
    return (
        actual.time.getTime() === expected.time.getTime() &&
        sameFloat(actual.parameterFirst, expected.parameterFirst, tolerance) &&
        sameFloat(actual.parameterLast, expected.parameterLast, tolerance) &&
        sameFloat(actual.parameterResolution, expected.parameterResolution, tolerance) &&
        sameFloat(actual.valueMin, expected.valueMin, tolerance) &&
        sameFloat(actual.valueMax, expected.valueMax, tolerance) &&
        sameSlice(actual.values, expected.values, tolerance)
    );
}

describe('Corona Engine', () => {
    const baseTime = new Date(Date.UTC(2020, 0, 1));

    it('should prime at sample 13 and hold dominant cycle when selectedBins == 0', () => {
        const engine = new Corona(DefaultCoronaParams);
        expect(engine.filterBankLength).toBe(49);
        expect(engine.minimalPeriodTimesTwo).toBe(12);
        expect(engine.maximalPeriodTimesTwo).toBe(60);

        const snapshots: CoronaSnapshot[] = testInput.map((price, i) =>
            engine.update(price, new Date(baseTime.getTime() + i * 86_400_000)),
        );
        for (let i = 0; i < 12; i++) {
            expect(snapshots[i].primed).toBe(false);
            expect(snapshots[i].maximumPower).toBe(0.0);
        }
        expect(snapshots[12].primed).toBe(true);

        let holds = 0;
        for (let i = 1; i < snapshots.length; i++) {
            if (snapshots[i].selectedBins === 0) {
                holds += 1;
                expect(snapshots[i].dominantCycle).toBe(snapshots[i - 1].dominantCycle);
                expect(snapshots[i].hasCycleEstimate).toBe(false);
            }
        }
        expect(holds).toBeGreaterThan(0);
        expect(sameFloat(snapshots[snapshots.length - 1].dominantCycle, 11.5)).toBe(true);
    });

    it('should match standalone and shared consumers across all 252 samples', () => {
        const engine = new Corona();
        const specStandalone = CoronaSpectrum.default();
        const specShared = CoronaSpectrum.default();
        const swingStandalone = CoronaSwingPosition.default();
        const swingShared = CoronaSwingPosition.default();
        const trendStandalone = CoronaTrendVigor.default();
        const trendShared = CoronaTrendVigor.default();
        const snrStandalone = CoronaSignalToNoiseRatio.default();
        const snrShared = CoronaSignalToNoiseRatio.default();

        let lastSnapshot!: CoronaSnapshot;
        for (let i = 0; i < testInput.length; i++) {
            const barTime = new Date(baseTime.getTime() + i * 86_400_000);
            const price = testInput[i];
            const snapshot = engine.update(price, barTime);
            lastSnapshot = snapshot;

            const specA = specStandalone.update(price, barTime);
            const specB = specShared.updateFromSnapshot(snapshot);
            expect(sameHeatmap(specA.heatmap, specB.heatmap)).toBe(true);
            expect(sameFloat(specA.dominantCycle, specB.dominantCycle)).toBe(true);
            expect(sameFloat(specA.dominantCycleMedian, specB.dominantCycleMedian)).toBe(true);
            expect(sameSlice(specA.power, specB.power)).toBe(true);
            expect(sameSlice(specA.relativePower, specB.relativePower)).toBe(true);

            const swingA = swingStandalone.update(price, barTime);
            const swingB = swingShared.updateFromSnapshot(snapshot);
            expect(sameHeatmap(swingA.heatmap, swingB.heatmap)).toBe(true);
            expect(sameFloat(swingA.swingPosition, swingB.swingPosition)).toBe(true);

            const trendA = trendStandalone.update(price, barTime);
            const trendB = trendShared.updateFromSnapshot(snapshot);
            expect(sameHeatmap(trendA.heatmap, trendB.heatmap)).toBe(true);
            expect(sameFloat(trendA.trendVigor, trendB.trendVigor)).toBe(true);

            const snrA = snrStandalone.update(price, testLow[i], testHigh[i], barTime);
            const snrB = snrShared.updateFromSnapshot(snapshot, testLow[i], testHigh[i]);
            expect(sameHeatmap(snrA.heatmap, snrB.heatmap)).toBe(true);
            expect(sameFloat(snrA.signalToNoiseRatio, snrB.signalToNoiseRatio)).toBe(true);
        }

        expect(engine.sampleCount).toBe(252);
        expect(() => specShared.updateFromSnapshot(lastSnapshot)).toThrow();

        const incompatible = CoronaSpectrum.fromParams({ highPassFilterCutoff: 48 });
        const freshEngine = new Corona();
        const firstSnapshot = freshEngine.update(testInput[0], baseTime);
        expect(() => incompatible.updateFromSnapshot(firstSnapshot)).toThrow();
    });

    it('should skip NaN without mutating state and return isolated snapshot arrays', () => {
        const specFirst = CoronaSpectrum.default();
        const specSecond = CoronaSpectrum.default();
        let latestFirst = specFirst.update(testInput[0], baseTime);
        let latestSecond = specSecond.update(testInput[0], baseTime);
        for (let i = 1; i < 50; i++) {
            const barTime = new Date(baseTime.getTime() + i * 86_400_000);
            if (i === 25) {
                const missing = specFirst.update(Number.NaN, barTime);
                expect(missing.heatmap.isEmpty()).toBe(true);
            }
            latestFirst = specFirst.update(testInput[i], barTime);
            latestSecond = specSecond.update(testInput[i], barTime);
            expect(sameHeatmap(latestFirst.heatmap, latestSecond.heatmap)).toBe(true);
        }
        latestFirst.heatmap.values[0] = -999.0;
        latestFirst.power[0] = -999.0;
        const afterFirst = specFirst.update(102.0, baseTime);
        const afterSecond = specSecond.update(102.0, baseTime);
        expect(sameHeatmap(afterFirst.heatmap, afterSecond.heatmap)).toBe(true);
        expect(sameSlice(afterFirst.power, afterSecond.power)).toBe(true);
    });

    it('should reject invalid parameters and overflow inputs without state mutation', () => {
        const badConfigs = [
            { minimalPeriod: 2 },
            { minimalPeriod: 0 },
            { minimalPeriod: 6.5 },
            { minimalPeriod: 6, maximalPeriod: 6 },
            { highPassFilterCutoff: 4 },
            { highPassFilterCutoff: 0 },
            { decibelsLowerThreshold: Number.NaN },
            { decibelsLowerThreshold: -1.0 },
            { decibelsLowerThreshold: 20.0, decibelsUpperThreshold: 20.0 },
            { decibelsUpperThreshold: Number.POSITIVE_INFINITY },
        ];
        for (const bad of badConfigs) {
            expect(() => new Corona(bad)).toThrow();
        }

        const zeroLower = new Corona({ decibelsLowerThreshold: 0.0 });
        expect(zeroLower.parameters.decibelsLowerThreshold).toBe(0.0);

        const engine = new Corona();
        for (let i = 0; i < 30; i++) {
            engine.update(testInput[i], baseTime);
        }
        for (const badPrice of [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, 1e308]) {
            expect(() => engine.update(badPrice, baseTime)).toThrow();
            expect(engine.sampleCount).toBe(30);
        }

        engine.reset();
        expect(engine.sampleCount).toBe(0);
        expect(engine.lastSnapshot).toBeNull();
    });

    it('should cancel sinusoid over full cycle and preserve linear displacement', () => {
        const consumer = CoronaTrendVigor.default();
        const engine = new Corona();
        for (let i = 0; i < 120; i++) {
            const price = 100.0 + 0.5 * i + 3.0 * Math.sin((TWO_PI * i) / 20.0);
            const rawSnapshot = engine.update(price, baseTime);
            const snapshot: CoronaSnapshot = {
                ...rawSnapshot,
                dominantCycle: 20.0,
                dominantCycleMedian: 20.0,
            };
            const output = consumer.updateFromSnapshot(snapshot);
            if (i < 20) {
                expect(output.heatmap.isEmpty()).toBe(true);
            } else {
                expect(output.lookback).toBe(20.0);
                expect(sameFloat(output.trendDisplacement, 10.0, 1e-12)).toBe(true);
            }
        }
        expect(consumer.isPrimed()).toBe(true);
    });

    it('should compute adaptive bandpass quadrature and 60-degree phase lead accurately', () => {
        const bandpass = new AdaptiveBandPass();
        let previousBp = 0.0;
        for (let i = 0; i < 100; i++) {
            const cycle = bandpass.update(100.0 + Math.sin((TWO_PI * i) / 20.0), 20.0);
            if (i >= 2) {
                expect(sameFloat(cycle.quadrature, (cycle.bandPass - previousBp) / (TWO_PI / 20.0), 1e-14)).toBe(true);
                expect(sameFloat(cycle.lead60, 0.5 * cycle.bandPass + (Math.sqrt(3.0) / 2.0) * cycle.quadrature, 1e-14)).toBe(true);
            }
            previousBp = cycle.bandPass;
        }
    });

    it('should produce explicit diagnostics on constant and zero prices and SNR physical amplitude ratio', () => {
        for (const price of [0.0, 100.0]) {
            const spectrum = CoronaSpectrum.default();
            const swing = CoronaSwingPosition.default();
            const trend = CoronaTrendVigor.default();
            const snr = CoronaSignalToNoiseRatio.default();
            let specOut = spectrum.update(price, baseTime);
            let swingOut = swing.update(price, baseTime);
            let trendOut = trend.update(price, baseTime);
            let snrOut = snr.updateRange(price, 2.0, baseTime);
            for (let i = 1; i < 100; i++) {
                specOut = spectrum.update(price, baseTime);
                swingOut = swing.update(price, baseTime);
                trendOut = trend.update(price, baseTime);
                snrOut = snr.updateRange(price, 2.0, baseTime);
            }
            expect(specOut.heatmap.isEmpty()).toBe(false);
            expect(swingOut.normalizedPosition).toBe(0.5);
            expect(swingOut.swingPosition).toBe(0.0);
            expect(swingOut.heatmap.values.every((v) => v === 20.0)).toBe(true);
            expect(trendOut.hasCycleEstimate).toBe(0.0);
            if (price > 0.0) {
                expect(snrOut.physicalDecibels).toBe(Number.NEGATIVE_INFINITY);
            } else {
                expect(Number.isNaN(snrOut.physicalDecibels)).toBe(true);
            }
            expect(snrOut.signalToNoiseRatio).toBe(1.0);
        }

        const consumer = CoronaSignalToNoiseRatio.default();
        const engine = new Corona();
        let output = consumer.updateRange(100.0, 2.0, baseTime);
        consumer.reset();
        engine.reset();
        for (let i = 0; i < 350; i++) {
            const snap = engine.update(100.0, baseTime);
            output = consumer.updateFromSnapshotRange({ ...snap, maximumPower: 4.0 }, 2.0);
        }
        expect(sameFloat(output.signal, 4.0, 1e-12)).toBe(true);
        expect(sameFloat(output.noise, 2.0, 1e-12)).toBe(true);
        expect(sameFloat(output.physicalDecibels, 20.0 * Math.log10(2.0), 1e-12)).toBe(true);
        expect(sameFloat(output.signalToNoiseRatio, 1.0 + output.physicalDecibels + 3.5, 1e-12)).toBe(true);
    });

    it('should align raster geometry marker endpoint and half-up rounding and factory partial params', () => {
        const raster = new Raster(50, -5.0, 5.0, 20.0);
        raster.update(1.0, 0.1, 'swing');
        const column = raster.column(baseTime);
        expect(sameFloat(column.parameterFirst, -4.8, 1e-14)).toBe(true);
        expect(sameFloat(column.parameterLast, 5.0, 1e-14)).toBe(true);
        expect(sameFloat(column.parameterResolution, 5.0, 1e-14)).toBe(true);
        expect(column.values[column.values.length - 1]).toBe(0.0);
        expect(column.values[0]).toBeGreaterThan(0.0);

        const raster2 = new Raster(50, -5.0, 5.0, 20.0);
        raster2.setValues(new Array<number>(50).fill(8.0));
        raster2.update(0.01, 0.1, 'swing');
        expect(raster2.valuesCopy()[0]).toBe(4.0);

        const specZero = createIndicator(IndicatorIdentifier.CoronaSpectrum, { minRasterValue: 0 });
        expect(specZero.metadata().mnemonic).toBe('cspect(0, 20, 6, 30, 30)');

        const snrZero = createIndicator(IndicatorIdentifier.CoronaSignalToNoiseRatio, { minParameterValue: 0 });
        expect(snrZero.metadata().mnemonic).toBe('csnr(50, 20, 0, 11, 30)');
    });
});
