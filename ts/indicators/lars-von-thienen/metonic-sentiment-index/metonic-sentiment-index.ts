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
import { MetonicSentimentIndexParams, defaultParams } from "./params.js";
/** Lars von Thienen prepared research implementation. */
export class MetonicSentimentIndex implements Indicator {
    private _engine: Engine;
    private _mnemonic: string;
    private _description: string;
    private _barFunc: (bar: Bar) => number = barComponentValue(DefaultBarComponent);
    private _quoteFunc: (quote: Quote) => number = quoteComponentValue(DefaultQuoteComponent);
    private _tradeFunc: (trade: Trade) => number = tradeComponentValue(DefaultTradeComponent);
    constructor(params: MetonicSentimentIndexParams = defaultParams()) {
        const p = { ...params };
        const lengthError = validateInteger(p.length, "length", 1);
        if (lengthError)
            throwError(lengthError);
        const delay1Error = validateInteger(p.delay1, "delay1", 1);
        if (delay1Error)
            throwError(delay1Error);
        const delay2Error = validateInteger(p.delay2, "delay2", 1);
        if (delay2Error)
            throwError(delay2Error);
        const delay3Error = validateInteger(p.delay3, "delay3", 1);
        if (delay3Error)
            throwError(delay3Error);
        const modeError = validateInteger(p.mode, "mode", 0);
        if (modeError)
            throwError(modeError);
        if (new Set([p.delay1, p.delay2, p.delay3]).size !== 3)
            throwError("three delays must be distinct");
        if (p.mode > 1)
            throwError("mode must be 0 or 1");
        this._engine = new Engine({ p: { ...p } });
        this._engine.reset();
        const parameters = [p.length, p.delay1, p.delay2, p.delay3, p.mode].map(value => Array.isArray(value) ? value.map(item => Number(item.toPrecision(12)).toString()).join("_") : typeof value === "boolean" ? String(Number(value)) : Number(value.toPrecision(12)).toString()).join(",");
        this._mnemonic = "msent(" + parameters + ")";
        this._description = "Metonic sentiment index " + this._mnemonic;
    }
    isPrimed(): boolean { return this._engine.primed(); }
    reset(): void { this._engine.reset(); }
    get warmupBars(): number { return this._engine.warmup(); }
    metadata(): IndicatorMetadata { return buildMetadata(IndicatorIdentifier.MetonicSentimentIndex, this._mnemonic, this._description, [{ mnemonic: this._mnemonic + " value", description: this._description + " value" }, { mnemonic: this._mnemonic + " daily_sentiment", description: this._description + " daily_sentiment" }, { mnemonic: this._mnemonic + " smoothed_close", description: this._description + " smoothed_close" }, { mnemonic: this._mnemonic + " smoothed_open", description: this._description + " smoothed_open" }]); }
    update(price: number, day?: Date, openPrice: number = NaN): Snapshot { return this._engine.update(price, day, openPrice); }
    private wrap(time: Date, result: Snapshot): IndicatorOutput {
        const output: IndicatorOutput = [];
        output.push(Object.assign(new Scalar(), { time, value: result.value }));
        output.push(Object.assign(new Scalar(), { time, value: result.dailySentiment }));
        output.push(Object.assign(new Scalar(), { time, value: result.smoothedClose }));
        output.push(Object.assign(new Scalar(), { time, value: result.smoothedOpen }));
        return output;
    }
    updateScalar(sample: Scalar): IndicatorOutput { if (this._engine.p.mode === 0)
        throwError("original sentiment requires OHLC"); return this.wrap(sample.time, this.update(sample.value, sample.time)); }
    updateBar(sample: Bar): IndicatorOutput { return this.wrap(sample.time, this.update(sample.close, sample.time, sample.open)); }
    updateQuote(sample: Quote): IndicatorOutput { if (this._engine.p.mode === 0)
        throwError("original sentiment requires OHLC"); return this.wrap(sample.time, this.update(sample.mid(), sample.time)); }
    updateTrade(sample: Trade): IndicatorOutput { if (this._engine.p.mode === 0)
        throwError("original sentiment requires OHLC"); return this.wrap(sample.time, this.update(sample.price, sample.time)); }
}
