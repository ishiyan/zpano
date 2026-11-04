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
import { CompleteCycleScannerParams, defaultParams } from "./params.js";
/** Lars von Thienen prepared research implementation. */
export class CompleteCycleScanner implements Indicator {
    private _engine: Engine;
    private _mnemonic: string;
    private _description: string;
    private _barFunc: (bar: Bar) => number = barComponentValue(DefaultBarComponent);
    private _quoteFunc: (quote: Quote) => number = quoteComponentValue(DefaultQuoteComponent);
    private _tradeFunc: (trade: Trade) => number = tradeComponentValue(DefaultTradeComponent);
    constructor(params: CompleteCycleScannerParams = defaultParams()) {
        const p = { ...params };
        const windowError = validateInteger(p.window, "window", 6);
        if (windowError)
            throwError(windowError);
        const minPeriodError = validateInteger(p.minPeriod, "min_period", 3);
        if (minPeriodError)
            throwError(minPeriodError);
        const maxPeriodError = validateInteger(p.maxPeriod, "max_period", 3);
        if (maxPeriodError)
            throwError(maxPeriodError);
        const smoothingError = validateReal(p.smoothing, "smoothing", 0);
        if (smoothingError)
            throwError(smoothingError);
        const genuineThresholdError = validateReal(p.genuineThreshold, "genuine_threshold", 0);
        if (genuineThresholdError)
            throwError(genuineThresholdError);
        const phaseCyclesError = validateReal(p.phaseCycles, "phase_cycles", 0);
        if (phaseCyclesError)
            throwError(phaseCyclesError);
        const topKError = validateInteger(p.topK, "top_k", 1);
        if (topKError)
            throwError(topKError);
        const peakDistanceError = validateInteger(p.peakDistance, "peak_distance", 1);
        if (peakDistanceError)
            throwError(peakDistanceError);
        const maxSegmentsError = validateInteger(p.maxSegments, "max_segments", 2);
        if (maxSegmentsError)
            throwError(maxSegmentsError);
        this._engine = new Engine({ p: { ...p } });
        this._engine.reset();
        this._barFunc = barComponentValue(p.barComponent ?? DefaultBarComponent);
        this._quoteFunc = quoteComponentValue(p.quoteComponent ?? DefaultQuoteComponent);
        this._tradeFunc = tradeComponentValue(p.tradeComponent ?? DefaultTradeComponent);
        const parameters = [p.window, p.minPeriod, p.maxPeriod, p.smoothing, p.genuineThreshold, p.phaseCycles, p.topK, p.peakDistance, p.maxSegments].map(value => Array.isArray(value) ? value.map(item => Number(item.toPrecision(12)).toString()).join("_") : typeof value === "boolean" ? String(Number(value)) : Number(value.toPrecision(12)).toString()).join(",");
        this._mnemonic = "cscan(" + parameters + componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent) + ")";
        this._description = "Complete cycle scanner " + this._mnemonic;
    }
    isPrimed(): boolean { return this._engine.primed(); }
    reset(): void { this._engine.reset(); }
    get warmupBars(): number { return this._engine.warmup(); }
    metadata(): IndicatorMetadata { return buildMetadata(IndicatorIdentifier.CompleteCycleScanner, this._mnemonic, this._description, [{ mnemonic: this._mnemonic + " spectrum", description: this._description + " spectrum" }, { mnemonic: this._mnemonic + " genuine_percent", description: this._description + " genuine_percent" }, { mnemonic: this._mnemonic + " dominant_period", description: this._description + " dominant_period" }, { mnemonic: this._mnemonic + " accepted_count", description: this._description + " accepted_count" }]); }
    update(price: number): Snapshot { return this._engine.update(price); }
    private wrap(time: Date, result: Snapshot): IndicatorOutput {
        const output: IndicatorOutput = [];
        output.push(Heatmap.newHeatmap(time, this._engine.p.minPeriod, this._engine.p.maxPeriod, 1, Math.min(...result.spectrum), Math.max(...result.spectrum), result.spectrum));
        output.push(Heatmap.newHeatmap(time, this._engine.p.minPeriod, this._engine.p.maxPeriod, 1, Math.min(...result.genuinePercent), Math.max(...result.genuinePercent), result.genuinePercent));
        output.push(Object.assign(new Scalar(), { time, value: result.dominantPeriod }));
        output.push(Object.assign(new Scalar(), { time, value: result.acceptedCount }));
        return output;
    }
    updateScalar(sample: Scalar): IndicatorOutput { return this.wrap(sample.time, this.update(sample.value)); }
    updateBar(sample: Bar): IndicatorOutput { return this.wrap(sample.time, this.update(this._barFunc(sample))); }
    updateQuote(sample: Quote): IndicatorOutput { return this.wrap(sample.time, this.update(this._quoteFunc(sample))); }
    updateTrade(sample: Trade): IndicatorOutput { return this.wrap(sample.time, this.update(this._tradeFunc(sample))); }
}
