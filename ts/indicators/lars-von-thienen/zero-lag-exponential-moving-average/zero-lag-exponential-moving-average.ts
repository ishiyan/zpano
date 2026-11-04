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
import { LarsVonThienenZeroLagExponentialMovingAverageParams, defaultParams } from "./params.js";
/** Lars von Thienen prepared research implementation. */
export class LarsVonThienenZeroLagExponentialMovingAverage extends LineIndicator {
    private _engine: Engine;
    constructor(params: LarsVonThienenZeroLagExponentialMovingAverageParams = defaultParams()) {
        const p = { ...params };
        super();
        const lengthError = validateInteger(p.length, "length", 1);
        if (lengthError)
            throwError(lengthError);
        this._engine = new Engine({ p: { ...p } });
        this._engine.reset();
        this.barComponent = p.barComponent;
        this.quoteComponent = p.quoteComponent;
        this.tradeComponent = p.tradeComponent;
        const parameters = [p.length].map(value => Array.isArray(value) ? value.map(item => Number(item.toPrecision(12)).toString()).join("_") : typeof value === "boolean" ? String(Number(value)) : Number(value.toPrecision(12)).toString()).join(",");
        this.mnemonic = "lzlema(" + parameters + componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent) + ")";
        this.description = "Zero lag exponential moving average " + this.mnemonic;
    }
    isPrimed(): boolean { return this._engine.primed(); }
    reset(): void { this._engine.reset(); }
    get warmupBars(): number { return this._engine.warmup(); }
    metadata(): IndicatorMetadata { return buildMetadata(IndicatorIdentifier.LarsVonThienenZeroLagExponentialMovingAverage, this.mnemonic, this.description, [{ mnemonic: this.mnemonic + " value", description: this.description + " value" }]); }
    update(price: number): number { return this._engine.update(price).value; }
}
