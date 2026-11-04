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
import { GoertzelCycleSpectrumParams, defaultParams } from "./params.js";
/** Lars von Thienen prepared research implementation. */
export class GoertzelCycleSpectrum implements Indicator {
    private _engine: Engine;
    private _mnemonic: string;
    private _description: string;
    private _barFunc: (bar: Bar) => number = barComponentValue(DefaultBarComponent);
    private _quoteFunc: (quote: Quote) => number = quoteComponentValue(DefaultQuoteComponent);
    private _tradeFunc: (trade: Trade) => number = tradeComponentValue(DefaultTradeComponent);
    constructor(params: GoertzelCycleSpectrumParams = defaultParams()) {
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
        const stepError = validateInteger(p.step, "step", 1);
        if (stepError)
            throwError(stepError);
        const smoothingError = validateReal(p.smoothing, "smoothing", 0);
        if (smoothingError)
            throwError(smoothingError);
        if (p.minPeriod > p.maxPeriod || (p.maxPeriod - p.minPeriod) % p.step)
            throwError("step must land on max_period");
        this._engine = new Engine({ p: { ...p } });
        this._engine.reset();
        this._barFunc = barComponentValue(p.barComponent ?? DefaultBarComponent);
        this._quoteFunc = quoteComponentValue(p.quoteComponent ?? DefaultQuoteComponent);
        this._tradeFunc = tradeComponentValue(p.tradeComponent ?? DefaultTradeComponent);
        const parameters = [p.window, p.minPeriod, p.maxPeriod, p.step, p.smoothing].map(value => Array.isArray(value) ? value.map(item => Number(item.toPrecision(12)).toString()).join("_") : typeof value === "boolean" ? String(Number(value)) : Number(value.toPrecision(12)).toString()).join(",");
        this._mnemonic = "gcs(" + parameters + componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent) + ")";
        this._description = "Goertzel cycle spectrum " + this._mnemonic;
    }
    isPrimed(): boolean { return this._engine.primed(); }
    reset(): void { this._engine.reset(); }
    get warmupBars(): number { return this._engine.warmup(); }
    metadata(): IndicatorMetadata { return buildMetadata(IndicatorIdentifier.GoertzelCycleSpectrum, this._mnemonic, this._description, [{ mnemonic: this._mnemonic + " amplitude", description: this._description + " amplitude" }, { mnemonic: this._mnemonic + " phases", description: this._description + " phases" }, { mnemonic: this._mnemonic + " trough_offsets", description: this._description + " trough_offsets" }, { mnemonic: this._mnemonic + " dominant_period", description: this._description + " dominant_period" }, { mnemonic: this._mnemonic + " dominant_amplitude", description: this._description + " dominant_amplitude" }, { mnemonic: this._mnemonic + " dominant_phase", description: this._description + " dominant_phase" }]); }
    update(price: number): Snapshot { return this._engine.update(price); }
    private wrap(time: Date, result: Snapshot): IndicatorOutput {
        const output: IndicatorOutput = [];
        output.push(Heatmap.newHeatmap(time, this._engine.p.minPeriod, this._engine.p.maxPeriod, 1 / this._engine.p.step, Math.min(...result.amplitude), Math.max(...result.amplitude), result.amplitude));
        output.push(Heatmap.newHeatmap(time, this._engine.p.minPeriod, this._engine.p.maxPeriod, 1 / this._engine.p.step, Math.min(...result.phases), Math.max(...result.phases), result.phases));
        output.push(Heatmap.newHeatmap(time, this._engine.p.minPeriod, this._engine.p.maxPeriod, 1 / this._engine.p.step, Math.min(...result.troughOffsets), Math.max(...result.troughOffsets), result.troughOffsets));
        output.push(Object.assign(new Scalar(), { time, value: result.dominantPeriod }));
        output.push(Object.assign(new Scalar(), { time, value: result.dominantAmplitude }));
        output.push(Object.assign(new Scalar(), { time, value: result.dominantPhase }));
        return output;
    }
    updateScalar(sample: Scalar): IndicatorOutput { return this.wrap(sample.time, this.update(sample.value)); }
    updateBar(sample: Bar): IndicatorOutput { return this.wrap(sample.time, this.update(this._barFunc(sample))); }
    updateQuote(sample: Quote): IndicatorOutput { return this.wrap(sample.time, this.update(this._quoteFunc(sample))); }
    updateTrade(sample: Trade): IndicatorOutput { return this.wrap(sample.time, this.update(this._tradeFunc(sample))); }
}
