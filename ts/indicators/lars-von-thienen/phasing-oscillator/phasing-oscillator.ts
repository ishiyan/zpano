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
import { PhasingOscillatorParams, defaultParams } from "./params.js";
/** Lars von Thienen prepared research implementation. */
export class PhasingOscillator implements Indicator {
    private _engine: Engine;
    private _mnemonic: string;
    private _description: string;
    private _barFunc: (bar: Bar) => number = barComponentValue(DefaultBarComponent);
    private _quoteFunc: (quote: Quote) => number = quoteComponentValue(DefaultQuoteComponent);
    private _tradeFunc: (trade: Trade) => number = tradeComponentValue(DefaultTradeComponent);
    constructor(params: PhasingOscillatorParams = defaultParams()) {
        const p = { ...params };
        const windowError = validateInteger(p.window, "window", 6);
        if (windowError)
            throwError(windowError);
        const shortMinError = validateInteger(p.shortMin, "short_min", 3);
        if (shortMinError)
            throwError(shortMinError);
        const shortMaxError = validateInteger(p.shortMax, "short_max", 3);
        if (shortMaxError)
            throwError(shortMaxError);
        const longMinError = validateInteger(p.longMin, "long_min", 3);
        if (longMinError)
            throwError(longMinError);
        const longMaxError = validateInteger(p.longMax, "long_max", 3);
        if (longMaxError)
            throwError(longMaxError);
        const smoothingError = validateReal(p.smoothing, "smoothing", 0);
        if (smoothingError)
            throwError(smoothingError);
        const genuineThresholdError = validateReal(p.genuineThreshold, "genuine_threshold", 0);
        if (genuineThresholdError)
            throwError(genuineThresholdError);
        const phaseCyclesError = validateReal(p.phaseCycles, "phase_cycles", 0);
        if (phaseCyclesError)
            throwError(phaseCyclesError);
        if (p.shortMax >= p.longMin)
            throwError("short_max must be smaller than long_min");
        this._engine = new Engine({ p: { ...p } });
        this._engine.reset();
        this._barFunc = barComponentValue(p.barComponent ?? DefaultBarComponent);
        this._quoteFunc = quoteComponentValue(p.quoteComponent ?? DefaultQuoteComponent);
        this._tradeFunc = tradeComponentValue(p.tradeComponent ?? DefaultTradeComponent);
        const parameters = [p.window, p.shortMin, p.shortMax, p.longMin, p.longMax, p.smoothing, p.genuineThreshold, p.phaseCycles].map(value => Array.isArray(value) ? value.map(item => Number(item.toPrecision(12)).toString()).join("_") : typeof value === "boolean" ? String(Number(value)) : Number(value.toPrecision(12)).toString()).join(",");
        this._mnemonic = "phosc(" + parameters + componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent) + ")";
        this._description = "Phasing oscillator " + this._mnemonic;
    }
    isPrimed(): boolean { return this._engine.primed(); }
    reset(): void { this._engine.reset(); }
    get warmupBars(): number { return this._engine.warmup(); }
    metadata(): IndicatorMetadata { return buildMetadata(IndicatorIdentifier.PhasingOscillator, this._mnemonic, this._description, [{ mnemonic: this._mnemonic + " value", description: this._description + " value" }, { mnemonic: this._mnemonic + " short_score", description: this._description + " short_score" }, { mnemonic: this._mnemonic + " long_score", description: this._description + " long_score" }, { mnemonic: this._mnemonic + " short_period", description: this._description + " short_period" }, { mnemonic: this._mnemonic + " long_period", description: this._description + " long_period" }]); }
    update(price: number): Snapshot { return this._engine.update(price); }
    private wrap(time: Date, result: Snapshot): IndicatorOutput {
        const output: IndicatorOutput = [];
        output.push(Object.assign(new Scalar(), { time, value: result.value }));
        output.push(Object.assign(new Scalar(), { time, value: result.shortScore }));
        output.push(Object.assign(new Scalar(), { time, value: result.longScore }));
        output.push(Object.assign(new Scalar(), { time, value: result.shortPeriod }));
        output.push(Object.assign(new Scalar(), { time, value: result.longPeriod }));
        return output;
    }
    updateScalar(sample: Scalar): IndicatorOutput { return this.wrap(sample.time, this.update(sample.value)); }
    updateBar(sample: Bar): IndicatorOutput { return this.wrap(sample.time, this.update(this._barFunc(sample))); }
    updateQuote(sample: Quote): IndicatorOutput { return this.wrap(sample.time, this.update(this._quoteFunc(sample))); }
    updateTrade(sample: Trade): IndicatorOutput { return this.wrap(sample.time, this.update(this._tradeFunc(sample))); }
}
