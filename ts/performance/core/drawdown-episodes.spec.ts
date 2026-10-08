import { DrawdownEpisode, DrawdownEpisodes } from './drawdown-episodes';

describe('DrawdownEpisodes', () => {

    // Drawdown observations (decimals) with two recovered episodes and an
    // open one at the end:
    //
    //   idx:   0     1     2     3     4     5     6     7     8
    //   dd:    0  -0.10 -0.20 -0.05    0  -0.03    0  -0.04 -0.06
    //              |--- episode 1 ---|     |-- 2 --|     |-- 3 (open)
    const DRAWDOWNS = [0.0, -0.10, -0.20, -0.05, 0.0, -0.03, 0.0, -0.04, -0.06];

    const EPISODES = [
        new DrawdownEpisode(-0.20, 1, 2, 4, true),
        new DrawdownEpisode(-0.03, 5, 5, 6, true),
        new DrawdownEpisode(-0.06, 7, 8, 8, false),
    ];

    function feed(tracker: DrawdownEpisodes, drawdowns: readonly number[]): DrawdownEpisodes {
        for (const dd of drawdowns) {
            tracker.update(dd);
        }
        return tracker;
    }

    function assertEmpty(t: DrawdownEpisodes): void {
        expect(t.episodes).toEqual([]);
        expect(t.depths).toEqual([]);
        expect(t.averageEpisodeDrawdown).toBe(0.0);
        expect(t.averageEpisodeDrawdownSquared).toBe(0.0);
        expect(t.averageEpisodeLength).toBe(0.0);
        expect(t.averageEpisodePeakToTrough).toBe(0.0);
        expect(t.averageEpisodeRecovery).toBe(0.0);
    }

    it('empty', () => {
        assertEmpty(new DrawdownEpisodes());
    });

    it('no drawdowns', () => {
        const t = feed(new DrawdownEpisodes(), [0.0, 0.0, 0.0]);
        expect(t.episodes).toEqual([]);
        expect(t.averageEpisodeDrawdown).toBe(0.0);
        expect(t.averageEpisodeDrawdownSquared).toBe(0.0);
    });

    it('episodes', () => {
        const t = feed(new DrawdownEpisodes(), DRAWDOWNS);
        expect(t.episodes).toEqual(EPISODES);
        expect(t.depths).toEqual([-0.20, -0.03, -0.06]);
    });

    it('averages', () => {
        const t = feed(new DrawdownEpisodes(), DRAWDOWNS);
        expect(t.averageEpisodeDrawdown).toBeCloseTo((0.20 + 0.03 + 0.06) / 3, 16);
        // Divided by the number of observations, not episodes.
        expect(t.averageEpisodeDrawdownSquared)
            .toBeCloseTo((0.04 + 0.0009 + 0.0036) / DRAWDOWNS.length, 16);
        // Lengths 4, 2, 2 (the open episode has no recovery observation).
        expect(t.averageEpisodeLength).toBeCloseTo(8 / 3, 15);
        // Peak to trough 2, 1, 2.
        expect(t.averageEpisodePeakToTrough).toBeCloseTo(5 / 3, 15);
        // Recovery 3, 2, 1.
        expect(t.averageEpisodeRecovery).toBeCloseTo(2.0, 15);
    });

    it('open episode progress', () => {
        const t = new DrawdownEpisodes();
        t.update(-0.05);
        expect(t.episodes).toEqual([new DrawdownEpisode(-0.05, 0, 0, 0, false)]);
        t.update(-0.02); // shallower, the trough doesn't move
        expect(t.episodes).toEqual([new DrawdownEpisode(-0.05, 0, 0, 1, false)]);
        t.update(-0.08); // new trough
        expect(t.episodes).toEqual([new DrawdownEpisode(-0.08, 0, 2, 2, false)]);
        t.update(0.0); // recovery closes the episode
        expect(t.episodes).toEqual([new DrawdownEpisode(-0.08, 0, 2, 3, true)]);
    });

    it('recalculate matches incremental', () => {
        const incremental = feed(new DrawdownEpisodes(), DRAWDOWNS);
        const rebuilt = feed(new DrawdownEpisodes(), [-0.5, -0.7, 0.0]);
        rebuilt.recalculate(DRAWDOWNS);
        expect(rebuilt.episodes).toEqual(incremental.episodes);
        expect(rebuilt.averageEpisodeDrawdown).toBe(incremental.averageEpisodeDrawdown);
        expect(rebuilt.averageEpisodeDrawdownSquared).toBe(incremental.averageEpisodeDrawdownSquared);
        expect(rebuilt.averageEpisodeLength).toBe(incremental.averageEpisodeLength);
        expect(rebuilt.averageEpisodeRecovery).toBe(incremental.averageEpisodeRecovery);
    });

    it('reset', () => {
        const t = feed(new DrawdownEpisodes(), DRAWDOWNS);
        t.reset();
        assertEmpty(t);
        t.update(-0.01);
        expect(t.episodes).toEqual([new DrawdownEpisode(-0.01, 0, 0, 0, false)]);
    });
});
