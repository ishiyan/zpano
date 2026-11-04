import { testInput } from "./testdata.js";
import { BarComponent } from "../../../entities/bar-component.js";
import { QuoteComponent } from "../../../entities/quote-component.js";
import { TradeComponent } from "../../../entities/trade-component.js";
import { componentTripleMnemonic } from "../../core/component-triple-mnemonic.js";
import { MetonicRelativeStrengthIndex } from "./metonic-relative-strength-index.js";
import { defaultParams } from "./params.js";
describe("Lars von Thienen MetonicRelativeStrengthIndex component contract", () => {
    it("includes each nondefault extraction component", () => {
        const combinations: [
            BarComponent | undefined,
            QuoteComponent | undefined,
            TradeComponent | undefined
        ][] = [[undefined, undefined, undefined], [BarComponent.Open, undefined, undefined], [undefined, QuoteComponent.Bid, undefined], [undefined, undefined, TradeComponent.Volume], [BarComponent.Open, QuoteComponent.Bid, undefined], [BarComponent.Open, undefined, TradeComponent.Volume], [undefined, QuoteComponent.Bid, TradeComponent.Volume]];
        for (const [barComponent, quoteComponent, tradeComponent] of combinations) {
            const indicator = new MetonicRelativeStrengthIndex({ ...defaultParams(), barComponent, quoteComponent, tradeComponent });
            expect(indicator.metadata().mnemonic.endsWith(componentTripleMnemonic(barComponent, quoteComponent, tradeComponent) + ")")).toBeTrue();
        }
    });
});
describe("Lars von Thienen production Metonic delays", () => { it("uses 6939,4016,2922 calendar slots", () => { const indicator = new MetonicRelativeStrengthIndex({ ...defaultParams(), length: 1 }); for (let i = 0; i < 6945; i++) {
    const actual = indicator.update(testInput[i % testInput.length]).value;
    if (i <= 6939)
        expect(Number.isNaN(actual)).toBeTrue();
    else {
        const expected = [6939, 4016, 2922].reduce((sum, lag) => sum + (testInput[(i - lag) % testInput.length] >= testInput[(i - lag - 1) % testInput.length] ? 100 : 0), 0) / 3;
        expect(actual).toBeCloseTo(expected, 13);
    }
} }); });
