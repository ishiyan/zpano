import { Quote } from './quote';
import { QuoteComponent } from './quote-component.enum';
import { quoteComponentValue, quoteComponentMnemonic } from './quote-component';

describe('QuoteComponent', () => {
    const q = new Quote({
        time: new Date(2021, 3, 1),
        askPrice: 1, bidPrice: 2, askSize: 3, bidSize: 4,
    });

    describe('quoteComponentValue', () => {
        const tests: { component: QuoteComponent; expected: number }[] = [
            { component: QuoteComponent.Bid, expected: 2 },
            { component: QuoteComponent.Ask, expected: 1 },
            { component: QuoteComponent.BidSize, expected: 4 },
            { component: QuoteComponent.AskSize, expected: 3 },
            { component: QuoteComponent.Mid, expected: (1 + 2) / 2 },
            { component: QuoteComponent.Weighted, expected: (1 * 3 + 2 * 4) / (3 + 4) },
            { component: QuoteComponent.WeightedMid, expected: (1 * 4 + 2 * 3) / (3 + 4) },
            { component: QuoteComponent.SpreadBp, expected: 10000 * 2 * (1 - 2) / (1 + 2) },
        ];

        tests.forEach(({ component, expected }) => {
            it(`should return ${expected} for ${QuoteComponent[component]}`, () => {
                const fn = quoteComponentValue(component);
                expect(fn(q)).toBe(expected);
            });
        });

        it('should default to mid for unknown component', () => {
            const fn = quoteComponentValue(9999 as QuoteComponent);
            expect(fn(q)).toBe(q.mid());
        });
    });

    describe('quoteComponentMnemonic', () => {
        const tests: { component: QuoteComponent; mnemonic: string }[] = [
            { component: QuoteComponent.Bid, mnemonic: 'b' },
            { component: QuoteComponent.Ask, mnemonic: 'a' },
            { component: QuoteComponent.BidSize, mnemonic: 'bs' },
            { component: QuoteComponent.AskSize, mnemonic: 'as' },
            { component: QuoteComponent.Mid, mnemonic: 'ba/2' },
            { component: QuoteComponent.Weighted, mnemonic: '(bbs+aas)/(bs+as)' },
            { component: QuoteComponent.WeightedMid, mnemonic: '(bas+abs)/(bs+as)' },
            { component: QuoteComponent.SpreadBp, mnemonic: 'spread bp' },
        ];

        tests.forEach(({ component, mnemonic }) => {
            it(`should return '${mnemonic}' for ${QuoteComponent[component]}`, () => {
                expect(quoteComponentMnemonic(component)).toBe(mnemonic);
            });
        });

        it('should return ?? for unknown component', () => {
            expect(quoteComponentMnemonic(9999 as QuoteComponent)).toBe('??');
        });
    });
});
