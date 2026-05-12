//! Streaming candlestick pattern recognition engine with fuzzy logic support.
//!
//! Usage:
//! ```ignore
//! let mut cp = CandlestickPatterns::new();
//! for bar in bars {
//!     cp.update(bar.open, bar.high, bar.low, bar.close);
//!     let result = cp.abandoned_baby(); // continuous float in [-100, +100]
//! }
//! ```
//!
//! Each pattern method returns a continuous confidence value in [-100, +100],
//! where positive values indicate bullish signals and negative values indicate
//! bearish signals. The magnitude reflects the fuzzy confidence of the match.
//! Use [`fuzzy::alpha_cut`](crate::fuzzy::alpha_cut) to convert to crisp
//! {-100, 0, +100} if needed.
//!
//! The engine inspects the most recent N bars (stored in a ring buffer)
//! and the incrementally maintained running totals for each criterion, giving
//! O(1) per bar after the warmup period.

use crate::fuzzy::{self, MembershipShape};
use super::core::{
    Criterion, CriterionState, OHLC, PatternIdentifier,
    DEFAULT_LONG_BODY, DEFAULT_VERY_LONG_BODY, DEFAULT_SHORT_BODY, DEFAULT_DOJI_BODY,
    DEFAULT_LONG_SHADOW, DEFAULT_VERY_LONG_SHADOW, DEFAULT_SHORT_SHADOW, DEFAULT_VERY_SHORT_SHADOW,
    DEFAULT_NEAR, DEFAULT_FAR, DEFAULT_EQUAL,
};
use super::patterns;

/// Minimum history size: 5-candle patterns + 10 default criterion period + 5 margin.
const MIN_HISTORY: usize = 20;

/// Options configures the CandlestickPatterns engine.
#[derive(Debug, Clone)]
pub struct Options {
    pub long_body: Option<Criterion>,
    pub very_long_body: Option<Criterion>,
    pub short_body: Option<Criterion>,
    pub doji_body: Option<Criterion>,
    pub long_shadow: Option<Criterion>,
    pub very_long_shadow: Option<Criterion>,
    pub short_shadow: Option<Criterion>,
    pub very_short_shadow: Option<Criterion>,
    pub near: Option<Criterion>,
    pub far: Option<Criterion>,
    pub equal: Option<Criterion>,
    pub fuzz_ratio: f64,
    pub shape: MembershipShape,
}

impl Default for Options {
    fn default() -> Self {
        Self {
            long_body: None,
            very_long_body: None,
            short_body: None,
            doji_body: None,
            long_shadow: None,
            very_long_shadow: None,
            short_shadow: None,
            very_short_shadow: None,
            near: None,
            far: None,
            equal: None,
            fuzz_ratio: 0.2,
            shape: MembershipShape::Sigmoid,
        }
    }
}

/// CandlestickPatterns is the candlestick pattern recognition engine.
///
/// Provides streaming bar-by-bar evaluation of 61 Japanese candlestick patterns.
/// Call `update(open, high, low, close)` for each new bar, then call any pattern
/// method to get the result for the current bar.
///
/// Pattern methods return a continuous float in [-100, +100]:
///   positive: bullish signal, negative: bearish signal, near zero: no match.
///   The magnitude reflects the fuzzy confidence of the match.
///   Hikkake and HikkakeModified may return intermediate values for
///   unconfirmed signals.
pub struct CandlestickPatterns {
    // Fuzzy configuration.
    pub(crate) fuzz_ratio: f64,
    pub(crate) shape: MembershipShape,

    // Criteria states.
    pub(crate) long_body: CriterionState,
    pub(crate) very_long_body: CriterionState,
    pub(crate) short_body: CriterionState,
    pub(crate) doji_body: CriterionState,
    pub(crate) long_shadow: CriterionState,
    pub(crate) very_long_shadow: CriterionState,
    pub(crate) short_shadow: CriterionState,
    pub(crate) very_short_shadow: CriterionState,
    pub(crate) near: CriterionState,
    pub(crate) far: CriterionState,
    pub(crate) equal: CriterionState,

    // Ring buffer of recent bars: each entry is (open, high, low, close).
    // Size is the largest criterion period + 5 candles + 5 margin, floored at MIN_HISTORY.
    history: Vec<OHLC>,
    hist_size: usize,
    hist_start: usize,
    hist_len: usize,
    /// Number of bars fed so far.
    pub(crate) count: i32,

    // Stateful pattern state: hikkake_modified
    pub(crate) hikmod_pattern_result: f64,
    pub(crate) hikmod_pattern_idx: i32,
    pub(crate) hikmod_confirmed: bool,
    pub(crate) hikmod_last_signal: f64,
}

impl CandlestickPatterns {
    /// Creates a new CandlestickPatterns engine with default options.
    pub fn new() -> Self {
        Self::with_options(Options::default())
    }

    /// Creates a new CandlestickPatterns engine with the given options.
    pub fn with_options(opts: Options) -> Self {
        let fuzz_ratio = if opts.fuzz_ratio != 0.0 { opts.fuzz_ratio } else { 0.2 };
        let shape = opts.shape;
        let max_shift = 5;

        let long_body = CriterionState::new(opts.long_body.unwrap_or(DEFAULT_LONG_BODY), max_shift);
        let very_long_body = CriterionState::new(opts.very_long_body.unwrap_or(DEFAULT_VERY_LONG_BODY), max_shift);
        let short_body = CriterionState::new(opts.short_body.unwrap_or(DEFAULT_SHORT_BODY), max_shift);
        let doji_body = CriterionState::new(opts.doji_body.unwrap_or(DEFAULT_DOJI_BODY), max_shift);
        let long_shadow = CriterionState::new(opts.long_shadow.unwrap_or(DEFAULT_LONG_SHADOW), max_shift);
        let very_long_shadow = CriterionState::new(opts.very_long_shadow.unwrap_or(DEFAULT_VERY_LONG_SHADOW), max_shift);
        let short_shadow = CriterionState::new(opts.short_shadow.unwrap_or(DEFAULT_SHORT_SHADOW), max_shift);
        let very_short_shadow = CriterionState::new(opts.very_short_shadow.unwrap_or(DEFAULT_VERY_SHORT_SHADOW), max_shift);
        let near = CriterionState::new(opts.near.unwrap_or(DEFAULT_NEAR), max_shift);
        let far = CriterionState::new(opts.far.unwrap_or(DEFAULT_FAR), max_shift);
        let equal = CriterionState::new(opts.equal.unwrap_or(DEFAULT_EQUAL), max_shift);

        // History size: largest criterion period + 10, floored at MIN_HISTORY.
        let max_period = [
            &long_body, &very_long_body, &short_body, &doji_body,
            &long_shadow, &very_long_shadow, &short_shadow, &very_short_shadow,
            &near, &far, &equal,
        ].iter().map(|s| s.criterion.average_period).max().unwrap_or(0);
        let history_size = (max_period + 10).max(MIN_HISTORY);

        CandlestickPatterns {
            fuzz_ratio,
            shape,
            long_body,
            very_long_body,
            short_body,
            doji_body,
            long_shadow,
            very_long_shadow,
            short_shadow,
            very_short_shadow,
            near,
            far,
            equal,
            history: vec![OHLC { o: 0.0, h: 0.0, l: 0.0, c: 0.0 }; history_size],
            hist_size: history_size,
            hist_start: 0,
            hist_len: 0,
            count: 0,
            hikmod_pattern_result: 0.0,
            hikmod_pattern_idx: 0,
            hikmod_confirmed: false,
            hikmod_last_signal: 0.0,
        }
    }

    /// Feeds a new OHLC bar into the engine.
    ///
    /// After calling this, all pattern methods reflect the state including this bar.
    pub fn update(&mut self, o: f64, h: f64, l: f64, c: f64) {
        let bar = OHLC { o, h, l, c };
        if self.hist_len == self.hist_size {
            self.history[self.hist_start] = bar;
            self.hist_start = (self.hist_start + 1) % self.hist_size;
        } else {
            let idx = (self.hist_start + self.hist_len) % self.hist_size;
            self.history[idx] = bar;
            self.hist_len += 1;
        }
        self.long_body.push(o, h, l, c);
        self.very_long_body.push(o, h, l, c);
        self.short_body.push(o, h, l, c);
        self.doji_body.push(o, h, l, c);
        self.long_shadow.push(o, h, l, c);
        self.very_long_shadow.push(o, h, l, c);
        self.short_shadow.push(o, h, l, c);
        self.very_short_shadow.push(o, h, l, c);
        self.near.push(o, h, l, c);
        self.far.push(o, h, l, c);
        self.equal.push(o, h, l, c);
        self.count += 1;

        // Update stateful patterns.
        self.hikmod_confirmed = false;
        self.hikmod_last_signal = 0.0;
        self.hikkake_modified_update();
    }

    /// Returns the number of bars fed so far.
    pub fn count(&self) -> i32 {
        self.count
    }

    // ------------------------------------------------------------------
    // Helper: get bar at position relative to end.
    // shift=1 means the most recent bar, shift=2 the one before, etc.
    // ------------------------------------------------------------------

    /// Get OHLC of a bar. shift=1 is most recent, shift=2 is one before, etc.
    pub(crate) fn bar(&self, shift: usize) -> OHLC {
        let idx = (self.hist_start + self.hist_len - shift) % self.hist_size;
        self.history[idx]
    }

    /// Checks if we have sufficient bars for a pattern requiring n_candles
    /// plus the maximum average_period of the given criteria.
    pub(crate) fn enough(&self, n_candles: usize, criteria: &[&CriterionState]) -> bool {
        let avail = if self.hist_len >= n_candles { self.hist_len - n_candles } else { return false };
        for cs in criteria {
            if avail < cs.criterion.average_period {
                return false;
            }
        }
        true
    }

    // ------------------------------------------------------------------
    // Criterion average helpers (shift is from the end, 1-based)
    // ------------------------------------------------------------------

    /// Gets the criterion average value at a given shift from the most recent bar.
    ///
    /// TA-Lib convention: the average uses the `period` bars BEFORE the reference bar
    /// (excluding the reference bar itself).
    pub(crate) fn avg_cs(&self, cs: &CriterionState, shift: usize) -> f64 {
        let b = self.bar(shift);
        cs.avg(shift, b.o, b.h, b.l, b.c)
    }

    // ------------------------------------------------------------------
    // Fuzzy membership helpers
    // ------------------------------------------------------------------

    /// Fuzzy 'value < avg': degree of membership.
    pub(crate) fn mu_less(&self, value: f64, cs: &CriterionState, shift: usize) -> f64 {
        let avg = self.avg_cs(cs, shift);
        let w = if avg > 0.0 { self.fuzz_ratio * avg } else { 0.0 };
        fuzzy::mu_less(value, avg, w, self.shape)
    }

    /// Fuzzy 'value > avg': degree of membership.
    pub(crate) fn mu_greater(&self, value: f64, cs: &CriterionState, shift: usize) -> f64 {
        let avg = self.avg_cs(cs, shift);
        let w = if avg > 0.0 { self.fuzz_ratio * avg } else { 0.0 };
        fuzzy::mu_greater(value, avg, w, self.shape)
    }

    /// Fuzzy 'value ≈ target ± avg': degree of closeness.
    pub(crate) fn mu_near_value(&self, value: f64, target: f64, cs: &CriterionState, shift: usize) -> f64 {
        let avg = self.avg_cs(cs, shift);
        let w = if avg > 0.0 { self.fuzz_ratio * avg } else { 0.0 };
        fuzzy::mu_near(value, target, w, self.shape)
    }

    /// Fuzzy 'value >= threshold' with explicit width (no criterion).
    pub(crate) fn mu_ge_raw(&self, value: f64, threshold: f64, width: f64) -> f64 {
        fuzzy::mu_greater_equal(value, threshold, width, self.shape)
    }

    /// Fuzzy 'value > threshold' with explicit width (no criterion).
    pub(crate) fn mu_gt_raw(&self, value: f64, threshold: f64, width: f64) -> f64 {
        fuzzy::mu_greater(value, threshold, width, self.shape)
    }

    /// Fuzzy 'value < threshold' with explicit width (no criterion).
    pub(crate) fn mu_lt_raw(&self, value: f64, threshold: f64, width: f64) -> f64 {
        fuzzy::mu_less(value, threshold, width, self.shape)
    }

    /// Raw fuzzy direction ∈ [-1, +1].
    pub(crate) fn mu_direction_raw(&self, o: f64, c: f64, shift: usize) -> f64 {
        let avg = self.avg_cs(&self.short_body, shift);
        fuzzy::mu_direction(o, c, avg, 2.0)
    }

    /// Fuzzy degree of bullishness ∈ [0, 1].
    pub(crate) fn mu_bullish(&self, o: f64, c: f64, shift: usize) -> f64 {
        let d = self.mu_direction_raw(o, c, shift);
        if d > 0.0 { d } else { 0.0 }
    }

    /// Fuzzy degree of bearishness ∈ [0, 1].
    pub(crate) fn mu_bearish(&self, o: f64, c: f64, shift: usize) -> f64 {
        let d = self.mu_direction_raw(o, c, shift);
        if -d > 0.0 { -d } else { 0.0 }
    }

    // -----------------------------------------------------------------------
    // Stateful pattern: hikkake_modified_update
    // -----------------------------------------------------------------------

    /// hikkake_modified_update is called from update() to track stateful hikkake_modified pattern.
    fn hikkake_modified_update(&mut self) {
        if self.count < 4 { return; }

        let b1 = self.bar(4);
        let b2 = self.bar(3);
        let b3 = self.bar(2);
        let b4 = self.bar(1);

        // Check for new pattern.
        if b2.h < b1.h && b2.l > b1.l
            && b3.h < b2.h && b3.l > b2.l
        {
            let near_avg = self.avg_cs(&self.near, 3);
            // Bullish: 4th breaks low, 2nd close near its low.
            if b4.h < b3.h && b4.l < b3.l && b2.c <= b2.l + near_avg {
                self.hikmod_pattern_result = 100.0;
                self.hikmod_pattern_idx = self.count;
                return;
            }
            // Bearish: 4th breaks high, 2nd close near its high.
            if b4.h > b3.h && b4.l > b3.l && b2.c >= b2.h - near_avg {
                self.hikmod_pattern_result = -100.0;
                self.hikmod_pattern_idx = self.count;
                return;
            }
        }

        // No new pattern — check for confirmation.
        if self.hikmod_pattern_result != 0.0 && self.count <= self.hikmod_pattern_idx + 3 {
            let shift3rd = (self.count - self.hikmod_pattern_idx + 2) as usize;
            let b3rd = self.bar(shift3rd);

            if self.hikmod_pattern_result > 0.0 && b4.c > b3rd.h {
                self.hikmod_last_signal = 200.0;
                self.hikmod_pattern_result = 0.0;
                self.hikmod_pattern_idx = 0;
                self.hikmod_confirmed = true;
                return;
            }
            if self.hikmod_pattern_result < 0.0 && b4.c < b3rd.l {
                self.hikmod_last_signal = -200.0;
                self.hikmod_pattern_result = 0.0;
                self.hikmod_pattern_idx = 0;
                self.hikmod_confirmed = true;
                return;
            }
        }

        // If we passed the 3-bar window, reset.
        if self.hikmod_pattern_result != 0.0 && self.count > self.hikmod_pattern_idx + 3 {
            self.hikmod_pattern_result = 0.0;
            self.hikmod_pattern_idx = 0;
        }
    }

    // -----------------------------------------------------------------------
    // Pattern methods (delegated to patterns/ files)
    // -----------------------------------------------------------------------

    pub fn abandoned_baby(&self) -> f64 { patterns::abandoned_baby::abandoned_baby(self) }
    pub fn advance_block(&self) -> f64 { patterns::advance_block::advance_block(self) }
    pub fn belt_hold(&self) -> f64 { patterns::belt_hold::belt_hold(self) }
    pub fn breakaway(&self) -> f64 { patterns::breakaway::breakaway(self) }
    pub fn closing_marubozu(&self) -> f64 { patterns::closing_marubozu::closing_marubozu(self) }
    pub fn concealing_baby_swallow(&self) -> f64 { patterns::concealing_baby_swallow::concealing_baby_swallow(self) }
    pub fn counterattack(&self) -> f64 { patterns::counterattack::counterattack(self) }
    pub fn dark_cloud_cover(&self) -> f64 { patterns::dark_cloud_cover::dark_cloud_cover(self) }
    pub fn doji(&self) -> f64 { patterns::doji::doji(self) }
    pub fn doji_star(&self) -> f64 { patterns::doji_star::doji_star(self) }
    pub fn dragonfly_doji(&self) -> f64 { patterns::dragonfly_doji::dragonfly_doji(self) }
    pub fn engulfing(&self) -> f64 { patterns::engulfing::engulfing(self) }
    pub fn evening_doji_star(&self) -> f64 { patterns::evening_doji_star::evening_doji_star(self) }
    pub fn evening_star(&self) -> f64 { patterns::evening_star::evening_star(self) }
    pub fn gravestone_doji(&self) -> f64 { patterns::gravestone_doji::gravestone_doji(self) }
    pub fn hammer(&self) -> f64 { patterns::hammer::hammer(self) }
    pub fn hanging_man(&self) -> f64 { patterns::hanging_man::hanging_man(self) }
    pub fn harami(&self) -> f64 { patterns::harami::harami(self) }
    pub fn harami_cross(&self) -> f64 { patterns::harami_cross::harami_cross(self) }
    pub fn high_wave(&self) -> f64 { patterns::high_wave::high_wave(self) }
    pub fn hikkake(&self) -> f64 { patterns::hikkake::hikkake(self) }
    pub fn hikkake_modified(&self) -> f64 { patterns::hikkake_modified::hikkake_modified(self) }
    pub fn homing_pigeon(&self) -> f64 { patterns::homing_pigeon::homing_pigeon(self) }
    pub fn identical_three_crows(&self) -> f64 { patterns::identical_three_crows::identical_three_crows(self) }
    pub fn in_neck(&self) -> f64 { patterns::in_neck::in_neck(self) }
    pub fn inverted_hammer(&self) -> f64 { patterns::inverted_hammer::inverted_hammer(self) }
    pub fn kicking(&self) -> f64 { patterns::kicking::kicking(self) }
    pub fn kicking_by_length(&self) -> f64 { patterns::kicking_by_length::kicking_by_length(self) }
    pub fn ladder_bottom(&self) -> f64 { patterns::ladder_bottom::ladder_bottom(self) }
    pub fn long_legged_doji(&self) -> f64 { patterns::long_legged_doji::long_legged_doji(self) }
    pub fn long_line(&self) -> f64 { patterns::long_line::long_line(self) }
    pub fn marubozu(&self) -> f64 { patterns::marubozu::marubozu(self) }
    pub fn matching_low(&self) -> f64 { patterns::matching_low::matching_low(self) }
    pub fn mat_hold(&self) -> f64 { patterns::mat_hold::mat_hold(self) }
    pub fn morning_doji_star(&self) -> f64 { patterns::morning_doji_star::morning_doji_star(self) }
    pub fn morning_star(&self) -> f64 { patterns::morning_star::morning_star(self) }
    pub fn on_neck(&self) -> f64 { patterns::on_neck::on_neck(self) }
    pub fn piercing(&self) -> f64 { patterns::piercing::piercing(self) }
    pub fn rickshaw_man(&self) -> f64 { patterns::rickshaw_man::rickshaw_man(self) }
    pub fn rising_falling_three_methods(&self) -> f64 { patterns::rising_falling_three_methods::rising_falling_three_methods(self) }
    pub fn separating_lines(&self) -> f64 { patterns::separating_lines::separating_lines(self) }
    pub fn shooting_star(&self) -> f64 { patterns::shooting_star::shooting_star(self) }
    pub fn short_line(&self) -> f64 { patterns::short_line::short_line(self) }
    pub fn spinning_top(&self) -> f64 { patterns::spinning_top::spinning_top(self) }
    pub fn stalled(&self) -> f64 { patterns::stalled::stalled(self) }
    pub fn stick_sandwich(&self) -> f64 { patterns::stick_sandwich::stick_sandwich(self) }
    pub fn takuri(&self) -> f64 { patterns::takuri::takuri(self) }
    pub fn tasuki_gap(&self) -> f64 { patterns::tasuki_gap::tasuki_gap(self) }
    pub fn three_black_crows(&self) -> f64 { patterns::three_black_crows::three_black_crows(self) }
    pub fn three_inside(&self) -> f64 { patterns::three_inside::three_inside(self) }
    pub fn three_line_strike(&self) -> f64 { patterns::three_line_strike::three_line_strike(self) }
    pub fn three_outside(&self) -> f64 { patterns::three_outside::three_outside(self) }
    pub fn three_stars_in_the_south(&self) -> f64 { patterns::three_stars_in_the_south::three_stars_in_the_south(self) }
    pub fn three_white_soldiers(&self) -> f64 { patterns::three_white_soldiers::three_white_soldiers(self) }
    pub fn thrusting(&self) -> f64 { patterns::thrusting::thrusting(self) }
    pub fn tristar(&self) -> f64 { patterns::tristar::tristar(self) }
    pub fn two_crows(&self) -> f64 { patterns::two_crows::two_crows(self) }
    pub fn unique_three_river(&self) -> f64 { patterns::unique_three_river::unique_three_river(self) }
    pub fn up_down_gap_side_by_side_white_lines(&self) -> f64 { patterns::up_down_gap_side_by_side_white_lines::up_down_gap_side_by_side_white_lines(self) }
    pub fn upside_gap_two_crows(&self) -> f64 { patterns::upside_gap_two_crows::upside_gap_two_crows(self) }
    pub fn x_side_gap_three_methods(&self) -> f64 { patterns::x_side_gap_three_methods::x_side_gap_three_methods(self) }

    // -----------------------------------------------------------------------
    // Evaluate
    // -----------------------------------------------------------------------

    /// Evaluates a single pattern by its identifier.
    ///
    /// Returns the pattern signal: +100 bullish, -100 bearish, 0 no match.
    pub fn evaluate(&self, id: PatternIdentifier) -> f64 {
        match id {
            PatternIdentifier::AbandonedBaby => self.abandoned_baby(),
            PatternIdentifier::AdvanceBlock => self.advance_block(),
            PatternIdentifier::BeltHold => self.belt_hold(),
            PatternIdentifier::Breakaway => self.breakaway(),
            PatternIdentifier::ClosingMarubozu => self.closing_marubozu(),
            PatternIdentifier::ConcealingBabySwallow => self.concealing_baby_swallow(),
            PatternIdentifier::Counterattack => self.counterattack(),
            PatternIdentifier::DarkCloudCover => self.dark_cloud_cover(),
            PatternIdentifier::Doji => self.doji(),
            PatternIdentifier::DojiStar => self.doji_star(),
            PatternIdentifier::DragonflyDoji => self.dragonfly_doji(),
            PatternIdentifier::Engulfing => self.engulfing(),
            PatternIdentifier::EveningDojiStar => self.evening_doji_star(),
            PatternIdentifier::EveningStar => self.evening_star(),
            PatternIdentifier::GravestoneDoji => self.gravestone_doji(),
            PatternIdentifier::Hammer => self.hammer(),
            PatternIdentifier::HangingMan => self.hanging_man(),
            PatternIdentifier::Harami => self.harami(),
            PatternIdentifier::HaramiCross => self.harami_cross(),
            PatternIdentifier::HighWave => self.high_wave(),
            PatternIdentifier::Hikkake => self.hikkake(),
            PatternIdentifier::HikkakeModified => self.hikkake_modified(),
            PatternIdentifier::HomingPigeon => self.homing_pigeon(),
            PatternIdentifier::IdenticalThreeCrows => self.identical_three_crows(),
            PatternIdentifier::InNeck => self.in_neck(),
            PatternIdentifier::InvertedHammer => self.inverted_hammer(),
            PatternIdentifier::Kicking => self.kicking(),
            PatternIdentifier::KickingByLength => self.kicking_by_length(),
            PatternIdentifier::LadderBottom => self.ladder_bottom(),
            PatternIdentifier::LongLeggedDoji => self.long_legged_doji(),
            PatternIdentifier::LongLine => self.long_line(),
            PatternIdentifier::Marubozu => self.marubozu(),
            PatternIdentifier::MatchingLow => self.matching_low(),
            PatternIdentifier::MatHold => self.mat_hold(),
            PatternIdentifier::MorningDojiStar => self.morning_doji_star(),
            PatternIdentifier::MorningStar => self.morning_star(),
            PatternIdentifier::OnNeck => self.on_neck(),
            PatternIdentifier::Piercing => self.piercing(),
            PatternIdentifier::RickshawMan => self.rickshaw_man(),
            PatternIdentifier::RisingFallingThreeMethods => self.rising_falling_three_methods(),
            PatternIdentifier::SeparatingLines => self.separating_lines(),
            PatternIdentifier::ShootingStar => self.shooting_star(),
            PatternIdentifier::ShortLine => self.short_line(),
            PatternIdentifier::SpinningTop => self.spinning_top(),
            PatternIdentifier::Stalled => self.stalled(),
            PatternIdentifier::StickSandwich => self.stick_sandwich(),
            PatternIdentifier::Takuri => self.takuri(),
            PatternIdentifier::TasukiGap => self.tasuki_gap(),
            PatternIdentifier::ThreeBlackCrows => self.three_black_crows(),
            PatternIdentifier::ThreeInside => self.three_inside(),
            PatternIdentifier::ThreeLineStrike => self.three_line_strike(),
            PatternIdentifier::ThreeOutside => self.three_outside(),
            PatternIdentifier::ThreeStarsInTheSouth => self.three_stars_in_the_south(),
            PatternIdentifier::ThreeWhiteSoldiers => self.three_white_soldiers(),
            PatternIdentifier::Thrusting => self.thrusting(),
            PatternIdentifier::Tristar => self.tristar(),
            PatternIdentifier::TwoCrows => self.two_crows(),
            PatternIdentifier::UniqueThreeRiver => self.unique_three_river(),
            PatternIdentifier::UpDownGapSideBySideWhiteLines => self.up_down_gap_side_by_side_white_lines(),
            PatternIdentifier::UpsideGapTwoCrows => self.upside_gap_two_crows(),
            PatternIdentifier::XSideGapThreeMethods => self.x_side_gap_three_methods(),
        }
    }
}
