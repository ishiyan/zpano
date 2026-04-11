import { Trade } from './trade';
import { TradeComponent } from './trade-component.enum';
import { tradeComponentValue, tradeComponentMnemonic } from './trade-component';

describe('TradeComponent', () => {
    const t = new Trade({
        time: new Date(2021, 3, 1),
        price: 1, volume: 2,
    });

    describe('tradeComponentValue', () => {
        const tests: { component: TradeComponent; expected: number }[] = [
            { component: TradeComponent.Price, expected: 1 },
            { component: TradeComponent.Volume, expected: 2 },
        ];

        tests.forEach(({ component, expected }) => {
            it(`should return ${expected} for ${TradeComponent[component]}`, () => {
                const fn = tradeComponentValue(component);
                expect(fn(t)).toBe(expected);
            });
        });

        it('should default to price for unknown component', () => {
            const fn = tradeComponentValue(9999 as TradeComponent);
            expect(fn(t)).toBe(t.price);
        });
    });

    describe('tradeComponentMnemonic', () => {
        const tests: { component: TradeComponent; mnemonic: string }[] = [
            { component: TradeComponent.Price, mnemonic: 'p' },
            { component: TradeComponent.Volume, mnemonic: 'v' },
        ];

        tests.forEach(({ component, mnemonic }) => {
            it(`should return '${mnemonic}' for ${TradeComponent[component]}`, () => {
                expect(tradeComponentMnemonic(component)).toBe(mnemonic);
            });
        });

        it('should return ?? for unknown component', () => {
            expect(tradeComponentMnemonic(9999 as TradeComponent)).toBe('??');
        });
    });
});
