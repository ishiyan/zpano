import unittest

from .drawdown_episodes import DrawdownEpisode, DrawdownEpisodes

# Drawdown observations (decimals) with two recovered episodes and an
# open one at the end:
#
#   idx:   0     1     2     3     4     5     6     7     8
#   dd:    0  -0.10 -0.20 -0.05    0  -0.03    0  -0.04 -0.06
#              |--- episode 1 ---|     |-- 2 --|     |-- 3 (open)
DRAWDOWNS = [0.0, -0.10, -0.20, -0.05, 0.0, -0.03, 0.0, -0.04, -0.06]

EPISODES = [
    DrawdownEpisode(depth=-0.20, from_idx=1, trough_idx=2, to_idx=4, recovered=True),
    DrawdownEpisode(depth=-0.03, from_idx=5, trough_idx=5, to_idx=6, recovered=True),
    DrawdownEpisode(depth=-0.06, from_idx=7, trough_idx=8, to_idx=8, recovered=False),
]


def feed(tracker: DrawdownEpisodes, drawdowns) -> DrawdownEpisodes:
    for dd in drawdowns:
        tracker.update(dd)
    return tracker


class TestDrawdownEpisodes(unittest.TestCase):

    def assert_empty(self, t: DrawdownEpisodes):
        self.assertEqual(t.episodes, [])
        self.assertEqual(t.depths, [])
        self.assertEqual(t.average_episode_drawdown, 0.0)
        self.assertEqual(t.average_episode_drawdown_squared, 0.0)
        self.assertEqual(t.average_episode_length, 0.0)
        self.assertEqual(t.average_episode_peak_to_trough, 0.0)
        self.assertEqual(t.average_episode_recovery, 0.0)

    def test_empty(self):
        self.assert_empty(DrawdownEpisodes())

    def test_no_drawdowns(self):
        t = feed(DrawdownEpisodes(), [0.0, 0.0, 0.0])
        self.assertEqual(t.episodes, [])
        self.assertEqual(t.average_episode_drawdown, 0.0)
        self.assertEqual(t.average_episode_drawdown_squared, 0.0)

    def test_episodes(self):
        t = feed(DrawdownEpisodes(), DRAWDOWNS)
        self.assertEqual(t.episodes, EPISODES)
        self.assertEqual(t.depths, [-0.20, -0.03, -0.06])

    def test_averages(self):
        t = feed(DrawdownEpisodes(), DRAWDOWNS)
        self.assertAlmostEqual(t.average_episode_drawdown, (0.20 + 0.03 + 0.06) / 3, places=16)
        # Divided by the number of observations, not episodes.
        self.assertAlmostEqual(t.average_episode_drawdown_squared,
                               (0.04 + 0.0009 + 0.0036) / len(DRAWDOWNS), places=16)
        # Lengths 4, 2, 2 (the open episode has no recovery observation).
        self.assertAlmostEqual(t.average_episode_length, 8 / 3, places=15)
        # Peak to trough 2, 1, 2.
        self.assertAlmostEqual(t.average_episode_peak_to_trough, 5 / 3, places=15)
        # Recovery 3, 2, 1.
        self.assertAlmostEqual(t.average_episode_recovery, 2.0, places=15)

    def test_open_episode_progress(self):
        t = DrawdownEpisodes()
        t.update(-0.05)
        self.assertEqual(t.episodes, [DrawdownEpisode(-0.05, 0, 0, 0, False)])
        t.update(-0.02)  # shallower, the trough doesn't move
        self.assertEqual(t.episodes, [DrawdownEpisode(-0.05, 0, 0, 1, False)])
        t.update(-0.08)  # new trough
        self.assertEqual(t.episodes, [DrawdownEpisode(-0.08, 0, 2, 2, False)])
        t.update(0.0)    # recovery closes the episode
        self.assertEqual(t.episodes, [DrawdownEpisode(-0.08, 0, 2, 3, True)])

    def test_recalculate_matches_incremental(self):
        incremental = feed(DrawdownEpisodes(), DRAWDOWNS)
        rebuilt = feed(DrawdownEpisodes(), [-0.5, -0.7, 0.0])
        rebuilt.recalculate(DRAWDOWNS)
        self.assertEqual(rebuilt.episodes, incremental.episodes)
        self.assertEqual(rebuilt.average_episode_drawdown, incremental.average_episode_drawdown)
        self.assertEqual(rebuilt.average_episode_drawdown_squared,
                         incremental.average_episode_drawdown_squared)
        self.assertEqual(rebuilt.average_episode_length, incremental.average_episode_length)
        self.assertEqual(rebuilt.average_episode_recovery, incremental.average_episode_recovery)

    def test_reset(self):
        t = feed(DrawdownEpisodes(), DRAWDOWNS)
        t.reset()
        self.assert_empty(t)
        t.update(-0.01)
        self.assertEqual(t.episodes, [DrawdownEpisode(-0.01, 0, 0, 0, False)])


if __name__ == "__main__":
    unittest.main()
