/**
 * ifres — command-line indicator frequency response calculator (TypeScript).
 *
 * Reads a JSON settings file containing indicator definitions, creates
 * indicator instances, determines each indicator's warmup period,
 * and calculates the frequency response with signal length 1024.
 *
 * Usage:  npx tsx cmd/ifres/main.ts <settings.json>
 */

import { readFileSync } from 'node:fs';
import { type Indicator } from '../../indicators/core/indicator.js';
import { IndicatorIdentifier } from '../../indicators/core/indicator-identifier.js';
import {
    FrequencyResponse,
    type FrequencyResponseComponent,
    type FrequencyResponseFilter,
    type FrequencyResponseResult,
} from '../../indicators/core/frequency-response/frequency-response.js';
import { createIndicator } from '../../indicators/factory/factory.js';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const signalLength = 1024;
const maxWarmup = 10_000;
const phaseDegreesUnwrappingLimit = 179;

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

interface SettingsEntry {
    identifier: string;
    params: Record<string, unknown>;
}

/** Map camelCase JSON identifier string to IndicatorIdentifier enum value. */
function resolveIdentifier(name: string): IndicatorIdentifier {
    const pascal = name.charAt(0).toUpperCase() + name.slice(1);

    const value = (IndicatorIdentifier as unknown as Record<string, number>)[pascal];
    if (value === undefined) {
        throw new Error(`Unknown indicator identifier: "${name}" (tried "${pascal}")`);
    }

    return value as IndicatorIdentifier;
}

// ---------------------------------------------------------------------------
// Warmup detection
// ---------------------------------------------------------------------------

/** Feed zeros into the indicator until it is primed, returning the count. */
function detectWarmup(filter: FrequencyResponseFilter, ind: Indicator): number {
    for (let i = 0; i < maxWarmup; i++) {
        if (ind.isPrimed()) {
            return i;
        }

        filter.update(0);
    }

    return maxWarmup;
}

// ---------------------------------------------------------------------------
// Output formatting
// ---------------------------------------------------------------------------

function printFrequencyResponse(fr: FrequencyResponseResult, warmup: number): void {
    console.log(`=== ${fr.label} (warmup=${warmup}) ===`);
    console.log(`  Spectrum length: ${fr.frequencies.length}`);

    printComponent('PowerPercent', fr.powerPercent);
    printComponent('PowerDecibel', fr.powerDecibel);
    printComponent('AmplitudePercent', fr.amplitudePercent);
    printComponent('AmplitudeDecibel', fr.amplitudeDecibel);
    printComponent('PhaseDegrees', fr.phaseDegrees);
    printComponent('PhaseDegreesUnwrapped', fr.phaseDegreesUnwrapped);

    console.log();
}

function printComponent(name: string, c: FrequencyResponseComponent): void {
    const label = name.padEnd(25);
    const min = c.min.toFixed(4).padStart(10);
    const max = c.max.toFixed(4).padStart(10);

    const n = c.data.length;
    let preview = '';
    if (n === 0) {
        preview = '';
    } else if (n <= 6) {
        preview = `  data=[${c.data.map(v => v.toFixed(4)).join(', ')}]`;
    } else {
        preview = `  data=[${c.data[0].toFixed(4)} ${c.data[1].toFixed(4)} ${c.data[2].toFixed(4)} ... ${c.data[n - 3].toFixed(4)} ${c.data[n - 2].toFixed(4)} ${c.data[n - 1].toFixed(4)}]`;
    }

    console.log(`  ${label} min=${min}  max=${max}${preview}`);
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

function main(): void {
    const args = process.argv.slice(2);
    if (args.length < 1) {
        process.stderr.write('usage: ifres <settings.json>\n');
        process.exit(1);
    }

    const data = readFileSync(args[0], 'utf-8');
    const entries: SettingsEntry[] = JSON.parse(data);

    for (const e of entries) {
        const id = resolveIdentifier(e.identifier);
        const params = Object.keys(e.params).length === 0 ? undefined : (e.params as Record<string, any>);

        // Create a probe instance to determine warmup period.
        let probe: Indicator;
        try {
            probe = createIndicator(id, params);
        } catch (err) {
            process.stderr.write(`error creating indicator ${e.identifier}: ${err}\n`);
            process.exit(1);
        }

        const probeFilter = probe as unknown as FrequencyResponseFilter;
        const warmup = detectWarmup(probeFilter, probe);

        // Create a fresh instance for the actual calculation.
        const ind = createIndicator(id, params);
        const filter = ind as unknown as FrequencyResponseFilter;

        const fr = FrequencyResponse.calculate(signalLength, filter, warmup, phaseDegreesUnwrappingLimit);
        printFrequencyResponse(fr, warmup);
    }
}

main();
