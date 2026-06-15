import { Levels, newLevel, newValueLevel } from './levels';

describe('Levels', () => {
    const time = new Date(2021, 3, 1);

    describe('newLevels', () => {
        it('stores time and levels as given', () => {
            const entries = [newLevel(105.5, 3, 0.8), newLevel(102, 1, 0.6)];
            const l = Levels.newLevels(time, entries);
            expect(l.time).toBe(time);
            expect(l.levels.length).toBe(2);
            expect(l.levels[0]).toEqual({ value: 105.5, offset: 3, strength: 0.8 });
            expect(l.levels[1]).toEqual({ value: 102, offset: 1, strength: 0.6 });
        });
    });

    describe('newValueLevel', () => {
        it('creates a value-only level with offset 0 and NaN strength', () => {
            const lv = newValueLevel(42);
            expect(lv.value).toBe(42);
            expect(lv.offset).toBe(0);
            expect(isNaN(lv.strength)).toBe(true);
        });
    });

    describe('newEmptyLevels', () => {
        it('creates a Levels with no entries', () => {
            const l = Levels.newEmptyLevels(time);
            expect(l.time).toBe(time);
            expect(l.levels).toEqual([]);
        });
    });

    describe('isEmpty', () => {
        it('returns true for a freshly created empty Levels', () => {
            expect(Levels.newEmptyLevels(time).isEmpty()).toBe(true);
        });

        it('returns false when levels are present', () => {
            const l = Levels.newLevels(time, [newValueLevel(1)]);
            expect(l.isEmpty()).toBe(false);
        });

        it('returns true when levels is undefined/null', () => {
            const l = new Levels();
            l.time = time;
            expect(l.isEmpty()).toBe(true);
        });
    });
});
