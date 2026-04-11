import { Bar } from './bar';

describe('Bar', () => {
    function bar(o: number, h: number, l: number, c: number, v: number): Bar {
        return new Bar({ time: new Date(2021, 3, 1), open: o, high: h, low: l, close: c, volume: v });
    }

    describe('median', () => {
        it('should calculate (low + high) / 2', () => {
            const b = bar(0, 3, 2, 0, 0);
            expect(b.median()).toBe((b.low + b.high) / 2);
        });
    });

    describe('typical', () => {
        it('should calculate (low + high + close) / 3', () => {
            const b = bar(0, 4, 2, 3, 0);
            expect(b.typical()).toBe((b.low + b.high + b.close) / 3);
        });
    });

    describe('weighted', () => {
        it('should calculate (low + high + 2*close) / 4', () => {
            const b = bar(0, 4, 2, 3, 0);
            expect(b.weighted()).toBe((b.low + b.high + b.close + b.close) / 4);
        });
    });

    describe('average', () => {
        it('should calculate (low + high + open + close) / 4', () => {
            const b = bar(3, 5, 2, 4, 0);
            expect(b.average()).toBe((b.low + b.high + b.open + b.close) / 4);
        });
    });

    describe('isRising', () => {
        it('should return true when open < close', () => {
            const b = bar(2, 0, 0, 3, 0);
            expect(b.isRising()).toBe(true);
        });

        it('should return false when close < open', () => {
            const b = bar(3, 0, 0, 2, 0);
            expect(b.isRising()).toBe(false);
        });

        it('should return false when open == close', () => {
            const b = bar(0, 0, 0, 0, 0);
            expect(b.isRising()).toBe(false);
        });
    });

    describe('isFalling', () => {
        it('should return false when open < close', () => {
            const b = bar(2, 0, 0, 3, 0);
            expect(b.isFalling()).toBe(false);
        });

        it('should return true when close < open', () => {
            const b = bar(3, 0, 0, 2, 0);
            expect(b.isFalling()).toBe(true);
        });

        it('should return false when open == close', () => {
            const b = bar(0, 0, 0, 0, 0);
            expect(b.isFalling()).toBe(false);
        });
    });
});
