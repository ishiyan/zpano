import { Bar } from './bar';
import { BarComponent, barComponentValue, barComponentMnemonic } from './bar-component';

describe('BarComponent', () => {
    const b = new Bar({
        time: new Date(2021, 3, 1),
        open: 2, high: 4, low: 1, close: 3, volume: 5,
    });

    describe('barComponentValue', () => {
        const tests: { component: BarComponent; expected: number }[] = [
            { component: BarComponent.Open, expected: 2 },
            { component: BarComponent.High, expected: 4 },
            { component: BarComponent.Low, expected: 1 },
            { component: BarComponent.Close, expected: 3 },
            { component: BarComponent.Volume, expected: 5 },
            { component: BarComponent.Median, expected: (1 + 4) / 2 },
            { component: BarComponent.Typical, expected: (1 + 4 + 3) / 3 },
            { component: BarComponent.Weighted, expected: (1 + 4 + 3 + 3) / 4 },
            { component: BarComponent.Average, expected: (1 + 4 + 3 + 2) / 4 },
        ];

        tests.forEach(({ component, expected }) => {
            it(`should return ${expected} for ${BarComponent[component]}`, () => {
                const fn = barComponentValue(component);
                expect(fn(b)).toBe(expected);
            });
        });

        it('should default to close for unknown component', () => {
            const fn = barComponentValue(9999 as BarComponent);
            expect(fn(b)).toBe(b.close);
        });
    });

    describe('barComponentMnemonic', () => {
        const tests: { component: BarComponent; mnemonic: string }[] = [
            { component: BarComponent.Open, mnemonic: 'o' },
            { component: BarComponent.High, mnemonic: 'h' },
            { component: BarComponent.Low, mnemonic: 'l' },
            { component: BarComponent.Close, mnemonic: 'c' },
            { component: BarComponent.Volume, mnemonic: 'v' },
            { component: BarComponent.Median, mnemonic: 'hl/2' },
            { component: BarComponent.Typical, mnemonic: 'hlc/3' },
            { component: BarComponent.Weighted, mnemonic: 'hlcc/4' },
            { component: BarComponent.Average, mnemonic: 'ohlc/4' },
        ];

        tests.forEach(({ component, mnemonic }) => {
            it(`should return '${mnemonic}' for ${BarComponent[component]}`, () => {
                expect(barComponentMnemonic(component)).toBe(mnemonic);
            });
        });

        it('should return ?? for unknown component', () => {
            expect(barComponentMnemonic(9999 as BarComponent)).toBe('??');
        });
    });
});
