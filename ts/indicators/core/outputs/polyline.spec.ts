import { Polyline } from './polyline';

describe('Polyline', () => {
    const time = new Date(2021, 3, 1);

    describe('newPolyline', () => {
        it('stores time and points as given', () => {
            const points = [
                { offset: 3, value: 10 },
                { offset: 1, value: 20 },
                { offset: 0, value: 15 },
            ];
            const p = Polyline.newPolyline(time, points);
            expect(p.time).toBe(time);
            expect(p.points.length).toBe(3);
            expect(p.points[0]).toEqual({ offset: 3, value: 10 });
            expect(p.points[2]).toEqual({ offset: 0, value: 15 });
        });
    });

    describe('newEmptyPolyline', () => {
        it('creates a polyline with no points', () => {
            const p = Polyline.newEmptyPolyline(time);
            expect(p.time).toBe(time);
            expect(p.points).toEqual([]);
        });
    });

    describe('isEmpty', () => {
        it('returns true for a freshly created empty polyline', () => {
            expect(Polyline.newEmptyPolyline(time).isEmpty()).toBe(true);
        });

        it('returns false when points are present', () => {
            const p = Polyline.newPolyline(time, [{ offset: 0, value: 1 }]);
            expect(p.isEmpty()).toBe(false);
        });

        it('returns true when points is undefined/null', () => {
            const p = new Polyline();
            p.time = time;
            expect(p.isEmpty()).toBe(true);
        });
    });
});
