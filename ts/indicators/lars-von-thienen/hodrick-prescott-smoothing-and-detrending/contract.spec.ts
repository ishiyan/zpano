import { BarComponent } from "../../../entities/bar-component.js";
import { QuoteComponent } from "../../../entities/quote-component.js";
import { TradeComponent } from "../../../entities/trade-component.js";
import { componentTripleMnemonic } from "../../core/component-triple-mnemonic.js";
import { HodrickPrescottSmoothingAndDetrending } from "./hodrick-prescott-smoothing-and-detrending.js";
import { defaultParams } from "./params.js";
describe("Lars von Thienen HodrickPrescottSmoothingAndDetrending component contract", () => {
    it("includes each nondefault extraction component", () => {
        const combinations: [
            BarComponent | undefined,
            QuoteComponent | undefined,
            TradeComponent | undefined
        ][] = [[undefined, undefined, undefined], [BarComponent.Open, undefined, undefined], [undefined, QuoteComponent.Bid, undefined], [undefined, undefined, TradeComponent.Volume], [BarComponent.Open, QuoteComponent.Bid, undefined], [BarComponent.Open, undefined, TradeComponent.Volume], [undefined, QuoteComponent.Bid, TradeComponent.Volume]];
        for (const [barComponent, quoteComponent, tradeComponent] of combinations) {
            const indicator = new HodrickPrescottSmoothingAndDetrending({ ...defaultParams(), barComponent, quoteComponent, tradeComponent });
            expect(indicator.metadata().mnemonic.endsWith(componentTripleMnemonic(barComponent, quoteComponent, tradeComponent) + ")")).toBeTrue();
        }
    });
});
