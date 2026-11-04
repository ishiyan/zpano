// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
import { Bar } from '../../../entities/bar.js';
import { Quote } from '../../../entities/quote.js';
import { Trade } from '../../../entities/trade.js';
import { Scalar } from '../../../entities/scalar.js';
import { barComponentValue, DefaultBarComponent } from '../../../entities/bar-component.js';
import { quoteComponentValue, DefaultQuoteComponent } from '../../../entities/quote-component.js';
import { tradeComponentValue, DefaultTradeComponent } from '../../../entities/trade-component.js';
import { Indicator } from '../../core/indicator.js';
import { IndicatorOutput } from '../../core/indicator-output.js';
import { IndicatorMetadata } from '../../core/indicator-metadata.js';
import { IndicatorIdentifier } from '../../core/indicator-identifier.js';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic.js';
import { buildMetadata } from '../../core/build-metadata.js';
import { LineIndicator } from '../../core/line-indicator.js';
import { Heatmap } from '../../core/outputs/heatmap.js';
import { Band } from '../../core/outputs/band.js';
import { Polyline } from '../../core/outputs/polyline.js';
import { validateInteger, validateReal, throwError } from '../harmonic-cycle/validation.js';
import { Engine, Snapshot } from './engine.js';
import { DetrendedRhythmOscillatorsParams, defaultParams } from "./params.js";
/** Lars von Thienen prepared research implementation. */
export class DetrendedRhythmOscillators implements Indicator {
    private _engine: Engine;
    private _mnemonic: string;
    private _description: string;
    private _barFunc: (bar: Bar) => number = barComponentValue(DefaultBarComponent);
    private _quoteFunc: (quote: Quote) => number = quoteComponentValue(DefaultQuoteComponent);
    private _tradeFunc: (trade: Trade) => number = tradeComponentValue(DefaultTradeComponent);
    constructor(params: DetrendedRhythmOscillatorsParams = defaultParams()) {
        const p = { ...params };
        const fastLengthError = validateInteger(p.fastLength, "fast_length", 1);
        if (fastLengthError)
            throwError(fastLengthError);
        const slowLengthError = validateInteger(p.slowLength, "slow_length", 2);
        if (slowLengthError)
            throwError(slowLengthError);
        const zigPeriodError = validateInteger(p.zigPeriod, "zig_period", 5);
        if (zigPeriodError)
            throwError(zigPeriodError);
        const cycleMemoryError = validateInteger(p.cycleMemory, "cycle_memory", 3);
        if (cycleMemoryError)
            throwError(cycleMemoryError);
        if (p.fastLength >= p.slowLength)
            throwError("fast_length must be smaller than slow_length");
        this._engine = new Engine({ p: { ...p } });
        this._engine.reset();
        this._barFunc = barComponentValue(p.barComponent ?? DefaultBarComponent);
        this._quoteFunc = quoteComponentValue(p.quoteComponent ?? DefaultQuoteComponent);
        this._tradeFunc = tradeComponentValue(p.tradeComponent ?? DefaultTradeComponent);
        const parameters = [p.fastLength, p.slowLength, p.zigPeriod, p.cycleMemory, p.detrend].map(value => Array.isArray(value) ? value.map(item => Number(item.toPrecision(12)).toString()).join("_") : typeof value === "boolean" ? String(Number(value)) : Number(value.toPrecision(12)).toString()).join(",");
        this._mnemonic = "dro(" + parameters + componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent) + ")";
        this._description = "Detrended rhythm oscillators " + this._mnemonic;
    }
    isPrimed(): boolean { return this._engine.primed(); }
    reset(): void { this._engine.reset(); }
    get warmupBars(): number { return this._engine.warmup(); }
    metadata(): IndicatorMetadata { return buildMetadata(IndicatorIdentifier.DetrendedRhythmOscillators, this._mnemonic, this._description, [{ mnemonic: this._mnemonic + " value", description: this._description + " value" }, { mnemonic: this._mnemonic + " direction", description: this._description + " direction" }, { mnemonic: this._mnemonic + " cycle_length", description: this._description + " cycle_length" }, { mnemonic: this._mnemonic + " pivot_value", description: this._description + " pivot_value" }, { mnemonic: this._mnemonic + " pivot_offset", description: this._description + " pivot_offset" }]); }
    update(price: number): Snapshot { return this._engine.update(price); }
    private wrap(time: Date, result: Snapshot): IndicatorOutput {
        const output: IndicatorOutput = [];
        output.push(Object.assign(new Scalar(), { time, value: result.value }));
        output.push(Object.assign(new Scalar(), { time, value: result.direction }));
        output.push(Object.assign(new Scalar(), { time, value: result.cycleLength }));
        output.push(Object.assign(new Scalar(), { time, value: result.pivotValue }));
        output.push(Object.assign(new Scalar(), { time, value: result.pivotOffset }));
        return output;
    }
    updateScalar(sample: Scalar): IndicatorOutput { return this.wrap(sample.time, this.update(sample.value)); }
    updateBar(sample: Bar): IndicatorOutput { return this.wrap(sample.time, this.update(this._barFunc(sample))); }
    updateQuote(sample: Quote): IndicatorOutput { return this.wrap(sample.time, this.update(this._quoteFunc(sample))); }
    updateTrade(sample: Trade): IndicatorOutput { return this.wrap(sample.time, this.update(this._tradeFunc(sample))); }
}
