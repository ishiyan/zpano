import { Quote } from './quote';

describe('Quote', () => {
    describe('mid', () => {
        it('should calculate (askPrice + bidPrice) / 2', () => {
            const q = new Quote({ time: new Date(), bidPrice: 3.0, askPrice: 2.0, bidSize: 0, askSize: 0 });
            expect(q.mid()).toBe((q.askPrice + q.bidPrice) / 2);
        });
    });

    describe('weighted', () => {
        it('should calculate (askPrice*askSize + bidPrice*bidSize) / (askSize + bidSize)', () => {
            const q = new Quote({ time: new Date(), bidPrice: 3.0, askPrice: 2.0, bidSize: 5.0, askSize: 4.0 });
            const expected = (q.askPrice * q.askSize + q.bidPrice * q.bidSize) / (q.askSize + q.bidSize);
            expect(q.weighted()).toBe(expected);
        });

        it('should return 0 when total size is 0', () => {
            const q = new Quote({ time: new Date(), bidPrice: 3.0, askPrice: 2.0, bidSize: 0, askSize: 0 });
            expect(q.weighted()).toBe(0);
        });
    });

    describe('weightedMid', () => {
        it('should calculate (askPrice*bidSize + bidPrice*askSize) / (askSize + bidSize)', () => {
            const q = new Quote({ time: new Date(), bidPrice: 3.0, askPrice: 2.0, bidSize: 5.0, askSize: 4.0 });
            const expected = (q.askPrice * q.bidSize + q.bidPrice * q.askSize) / (q.askSize + q.bidSize);
            expect(q.weightedMid()).toBe(expected);
        });

        it('should return 0 when total size is 0', () => {
            const q = new Quote({ time: new Date(), bidPrice: 3.0, askPrice: 2.0, bidSize: 0, askSize: 0 });
            expect(q.weightedMid()).toBe(0);
        });
    });

    describe('spreadBp', () => {
        it('should calculate 20000 * (askPrice - bidPrice) / (askPrice + bidPrice)', () => {
            const q = new Quote({ time: new Date(), bidPrice: 3.0, askPrice: 2.0, bidSize: 0, askSize: 0 });
            const expected = 20000 * (q.askPrice - q.bidPrice) / (q.askPrice + q.bidPrice);
            expect(q.spreadBp()).toBe(expected);
        });

        it('should return 0 when mid is 0', () => {
            const q = new Quote({ time: new Date(), bidPrice: 0, askPrice: 0, bidSize: 0, askSize: 0 });
            expect(q.spreadBp()).toBe(0);
        });
    });
});
