use crate::daycounting::conventions::DayCountConvention;
use crate::daycounting::fractional::{year_frac, DateTime};
use super::roundtrip::Roundtrip;
use super::side::RoundtripSide;

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

fn slice_mean(s: &[f64]) -> f64 {
    let n = s.len();
    if n == 0 {
        return 0.0;
    }
    let sum: f64 = s.iter().sum();
    sum / n as f64
}

fn slice_std_pop(s: &[f64]) -> f64 {
    let n = s.len();
    if n == 0 {
        return 0.0;
    }
    let m = slice_mean(s);
    let sum: f64 = s.iter().map(|v| (v - m) * (v - m)).sum();
    (sum / n as f64).sqrt()
}

fn max_consecutive(bools: &[bool]) -> usize {
    let mut max_streak: usize = 0;
    let mut current: usize = 0;
    for &b in bools {
        if b {
            current += 1;
            if current > max_streak {
                max_streak = current;
            }
        } else {
            current = 0;
        }
    }
    max_streak
}

fn div_or_zero(a: f64, b: usize) -> f64 {
    if b > 0 {
        a / b as f64
    } else {
        0.0
    }
}

fn min_slice(s: &[f64]) -> f64 {
    if s.is_empty() {
        return 0.0;
    }
    let mut m = s[0];
    for &v in &s[1..] {
        if v < m {
            m = v;
        }
    }
    m
}

fn max_slice(s: &[f64]) -> f64 {
    if s.is_empty() {
        return 0.0;
    }
    let mut m = s[0];
    for &v in &s[1..] {
        if v > m {
            m = v;
        }
    }
    m
}

// ---------------------------------------------------------------------------
// RoundtripPerformance
// ---------------------------------------------------------------------------

pub struct RoundtripPerformance {
    initial_balance: f64,
    annual_risk_free_rate: f64,
    annual_target_return: f64,
    day_count_convention: DayCountConvention,

    roundtrips: Vec<Roundtrip>,
    returns_on_investments: Vec<f64>,
    sortino_downside_returns: Vec<f64>,
    returns_on_investments_annual: Vec<f64>,
    sortino_downside_returns_annual: Vec<f64>,

    first_time: Option<DateTime>,
    last_time: Option<DateTime>,
    max_net_pnl: f64,
    max_drawdown: f64,
    max_drawdown_percent: f64,

    total_commission: f64,
    gross_winning_commission: f64,
    gross_loosing_commission: f64,
    net_winning_commission: f64,
    net_loosing_commission: f64,
    gross_winning_long_commission: f64,
    gross_loosing_long_commission: f64,
    net_winning_long_commission: f64,
    net_loosing_long_commission: f64,
    gross_winning_short_commission: f64,
    gross_loosing_short_commission: f64,
    net_winning_short_commission: f64,
    net_loosing_short_commission: f64,

    net_pnl: f64,
    gross_pnl: f64,
    gross_winning_pnl: f64,
    gross_loosing_pnl: f64,
    net_winning_pnl: f64,
    net_loosing_pnl: f64,
    gross_long_pnl: f64,
    gross_short_pnl: f64,
    net_long_pnl: f64,
    net_short_pnl: f64,
    gross_long_winning_pnl: f64,
    gross_long_loosing_pnl: f64,
    net_long_winning_pnl: f64,
    net_long_loosing_pnl: f64,
    gross_short_winning_pnl: f64,
    gross_short_loosing_pnl: f64,
    net_short_winning_pnl: f64,
    net_short_loosing_pnl: f64,

    total_count: usize,
    long_count: usize,
    short_count: usize,
    gross_winning_count: usize,
    gross_loosing_count: usize,
    net_winning_count: usize,
    net_loosing_count: usize,
    gross_long_winning_count: usize,
    gross_long_loosing_count: usize,
    net_long_winning_count: usize,
    net_long_loosing_count: usize,
    gross_short_winning_count: usize,
    gross_short_loosing_count: usize,
    net_short_winning_count: usize,
    net_short_loosing_count: usize,

    duration_sec: f64,
    duration_sec_long: f64,
    duration_sec_short: f64,
    duration_sec_gross_winning: f64,
    duration_sec_gross_loosing: f64,
    duration_sec_net_winning: f64,
    duration_sec_net_loosing: f64,
    duration_sec_gross_long_winning: f64,
    duration_sec_gross_long_loosing: f64,
    duration_sec_net_long_winning: f64,
    duration_sec_net_long_loosing: f64,
    duration_sec_gross_short_winning: f64,
    duration_sec_gross_short_loosing: f64,
    duration_sec_net_short_winning: f64,
    duration_sec_net_short_loosing: f64,
    total_duration_annualized: f64,

    total_mae: f64,
    total_mfe: f64,
    total_eff: f64,
    total_eff_entry: f64,
    total_eff_exit: f64,

    roi_mean: Option<f64>,
    roi_std: Option<f64>,
    roi_tdd: Option<f64>,
    roiann_mean: Option<f64>,
    roiann_std: Option<f64>,
    roiann_tdd: Option<f64>,
}

impl RoundtripPerformance {
    pub fn new(
        initial_balance: f64,
        annual_risk_free_rate: f64,
        annual_target_return: f64,
        day_count_convention: DayCountConvention,
    ) -> Self {
        Self {
            initial_balance,
            annual_risk_free_rate,
            annual_target_return,
            day_count_convention,
            roundtrips: Vec::new(),
            returns_on_investments: Vec::new(),
            sortino_downside_returns: Vec::new(),
            returns_on_investments_annual: Vec::new(),
            sortino_downside_returns_annual: Vec::new(),
            first_time: None,
            last_time: None,
            max_net_pnl: 0.0,
            max_drawdown: 0.0,
            max_drawdown_percent: 0.0,
            total_commission: 0.0,
            gross_winning_commission: 0.0,
            gross_loosing_commission: 0.0,
            net_winning_commission: 0.0,
            net_loosing_commission: 0.0,
            gross_winning_long_commission: 0.0,
            gross_loosing_long_commission: 0.0,
            net_winning_long_commission: 0.0,
            net_loosing_long_commission: 0.0,
            gross_winning_short_commission: 0.0,
            gross_loosing_short_commission: 0.0,
            net_winning_short_commission: 0.0,
            net_loosing_short_commission: 0.0,
            net_pnl: 0.0,
            gross_pnl: 0.0,
            gross_winning_pnl: 0.0,
            gross_loosing_pnl: 0.0,
            net_winning_pnl: 0.0,
            net_loosing_pnl: 0.0,
            gross_long_pnl: 0.0,
            gross_short_pnl: 0.0,
            net_long_pnl: 0.0,
            net_short_pnl: 0.0,
            gross_long_winning_pnl: 0.0,
            gross_long_loosing_pnl: 0.0,
            net_long_winning_pnl: 0.0,
            net_long_loosing_pnl: 0.0,
            gross_short_winning_pnl: 0.0,
            gross_short_loosing_pnl: 0.0,
            net_short_winning_pnl: 0.0,
            net_short_loosing_pnl: 0.0,
            total_count: 0,
            long_count: 0,
            short_count: 0,
            gross_winning_count: 0,
            gross_loosing_count: 0,
            net_winning_count: 0,
            net_loosing_count: 0,
            gross_long_winning_count: 0,
            gross_long_loosing_count: 0,
            net_long_winning_count: 0,
            net_long_loosing_count: 0,
            gross_short_winning_count: 0,
            gross_short_loosing_count: 0,
            net_short_winning_count: 0,
            net_short_loosing_count: 0,
            duration_sec: 0.0,
            duration_sec_long: 0.0,
            duration_sec_short: 0.0,
            duration_sec_gross_winning: 0.0,
            duration_sec_gross_loosing: 0.0,
            duration_sec_net_winning: 0.0,
            duration_sec_net_loosing: 0.0,
            duration_sec_gross_long_winning: 0.0,
            duration_sec_gross_long_loosing: 0.0,
            duration_sec_net_long_winning: 0.0,
            duration_sec_net_long_loosing: 0.0,
            duration_sec_gross_short_winning: 0.0,
            duration_sec_gross_short_loosing: 0.0,
            duration_sec_net_short_winning: 0.0,
            duration_sec_net_short_loosing: 0.0,
            total_duration_annualized: 0.0,
            total_mae: 0.0,
            total_mfe: 0.0,
            total_eff: 0.0,
            total_eff_entry: 0.0,
            total_eff_exit: 0.0,
            roi_mean: None,
            roi_std: None,
            roi_tdd: None,
            roiann_mean: None,
            roiann_std: None,
            roiann_tdd: None,
        }
    }

    pub fn reset(&mut self) {
        self.roundtrips.clear();
        self.returns_on_investments.clear();
        self.sortino_downside_returns.clear();
        self.returns_on_investments_annual.clear();
        self.sortino_downside_returns_annual.clear();
        self.first_time = None;
        self.last_time = None;
        self.max_net_pnl = 0.0;
        self.max_drawdown = 0.0;
        self.max_drawdown_percent = 0.0;
        self.total_commission = 0.0;
        self.gross_winning_commission = 0.0;
        self.gross_loosing_commission = 0.0;
        self.net_winning_commission = 0.0;
        self.net_loosing_commission = 0.0;
        self.gross_winning_long_commission = 0.0;
        self.gross_loosing_long_commission = 0.0;
        self.net_winning_long_commission = 0.0;
        self.net_loosing_long_commission = 0.0;
        self.gross_winning_short_commission = 0.0;
        self.gross_loosing_short_commission = 0.0;
        self.net_winning_short_commission = 0.0;
        self.net_loosing_short_commission = 0.0;
        self.net_pnl = 0.0;
        self.gross_pnl = 0.0;
        self.gross_winning_pnl = 0.0;
        self.gross_loosing_pnl = 0.0;
        self.net_winning_pnl = 0.0;
        self.net_loosing_pnl = 0.0;
        self.gross_long_pnl = 0.0;
        self.gross_short_pnl = 0.0;
        self.net_long_pnl = 0.0;
        self.net_short_pnl = 0.0;
        self.gross_long_winning_pnl = 0.0;
        self.gross_long_loosing_pnl = 0.0;
        self.net_long_winning_pnl = 0.0;
        self.net_long_loosing_pnl = 0.0;
        self.gross_short_winning_pnl = 0.0;
        self.gross_short_loosing_pnl = 0.0;
        self.net_short_winning_pnl = 0.0;
        self.net_short_loosing_pnl = 0.0;
        self.total_count = 0;
        self.long_count = 0;
        self.short_count = 0;
        self.gross_winning_count = 0;
        self.gross_loosing_count = 0;
        self.net_winning_count = 0;
        self.net_loosing_count = 0;
        self.gross_long_winning_count = 0;
        self.gross_long_loosing_count = 0;
        self.net_long_winning_count = 0;
        self.net_long_loosing_count = 0;
        self.gross_short_winning_count = 0;
        self.gross_short_loosing_count = 0;
        self.net_short_winning_count = 0;
        self.net_short_loosing_count = 0;
        self.duration_sec = 0.0;
        self.duration_sec_long = 0.0;
        self.duration_sec_short = 0.0;
        self.duration_sec_gross_winning = 0.0;
        self.duration_sec_gross_loosing = 0.0;
        self.duration_sec_net_winning = 0.0;
        self.duration_sec_net_loosing = 0.0;
        self.duration_sec_gross_long_winning = 0.0;
        self.duration_sec_gross_long_loosing = 0.0;
        self.duration_sec_net_long_winning = 0.0;
        self.duration_sec_net_long_loosing = 0.0;
        self.duration_sec_gross_short_winning = 0.0;
        self.duration_sec_gross_short_loosing = 0.0;
        self.duration_sec_net_short_winning = 0.0;
        self.duration_sec_net_short_loosing = 0.0;
        self.total_duration_annualized = 0.0;
        self.total_mae = 0.0;
        self.total_mfe = 0.0;
        self.total_eff = 0.0;
        self.total_eff_entry = 0.0;
        self.total_eff_exit = 0.0;
        self.roi_mean = None;
        self.roi_std = None;
        self.roi_tdd = None;
        self.roiann_mean = None;
        self.roiann_std = None;
        self.roiann_tdd = None;
    }

    pub fn add_roundtrip(&mut self, rt: Roundtrip) {
        self.total_count += 1;
        let comm = rt.commission();
        self.total_commission += comm;
        let secs = rt.duration_seconds();
        self.duration_sec += secs;
        self.total_mae += rt.maximum_adverse_excursion();
        self.total_mfe += rt.maximum_favorable_excursion();
        self.total_eff += rt.total_efficiency();
        self.total_eff_entry += rt.entry_efficiency();
        self.total_eff_exit += rt.exit_efficiency();

        let net_pnl = rt.net_pnl();
        self.net_pnl += net_pnl;
        if net_pnl > 0.0 {
            self.net_winning_count += 1;
            self.net_winning_pnl += net_pnl;
            self.net_winning_commission += comm;
            self.duration_sec_net_winning += secs;
        } else if net_pnl < 0.0 {
            self.net_loosing_count += 1;
            self.net_loosing_pnl += net_pnl;
            self.net_loosing_commission += comm;
            self.duration_sec_net_loosing += secs;
        }

        let gross_pnl = rt.gross_pnl();
        self.gross_pnl += gross_pnl;
        if gross_pnl > 0.0 {
            self.gross_winning_count += 1;
            self.gross_winning_pnl += gross_pnl;
            self.gross_winning_commission += comm;
            self.duration_sec_gross_winning += secs;
        } else if gross_pnl < 0.0 {
            self.gross_loosing_count += 1;
            self.gross_loosing_pnl += gross_pnl;
            self.gross_loosing_commission += comm;
            self.duration_sec_gross_loosing += secs;
        }

        if rt.side() == RoundtripSide::Long {
            self.gross_long_pnl += gross_pnl;
            self.net_long_pnl += net_pnl;
            self.long_count += 1;
            self.duration_sec_long += secs;
            if gross_pnl > 0.0 {
                self.gross_long_winning_count += 1;
                self.gross_long_winning_pnl += gross_pnl;
                self.gross_winning_long_commission += comm;
                self.duration_sec_gross_long_winning += secs;
            } else if gross_pnl < 0.0 {
                self.gross_long_loosing_count += 1;
                self.gross_long_loosing_pnl += gross_pnl;
                self.gross_loosing_long_commission += comm;
                self.duration_sec_gross_long_loosing += secs;
            }
            if net_pnl > 0.0 {
                self.net_long_winning_count += 1;
                self.net_long_winning_pnl += gross_pnl; // intentional: uses gross_pnl
                self.net_winning_long_commission += comm;
                self.duration_sec_net_long_winning += secs;
            } else if net_pnl < 0.0 {
                self.net_long_loosing_count += 1;
                self.net_long_loosing_pnl += gross_pnl; // intentional: uses gross_pnl
                self.net_loosing_long_commission += comm;
                self.duration_sec_net_long_loosing += secs;
            }
        } else {
            self.gross_short_pnl += gross_pnl;
            self.net_short_pnl += net_pnl;
            self.short_count += 1;
            self.duration_sec_short += secs;
            if gross_pnl > 0.0 {
                self.gross_short_winning_count += 1;
                self.gross_short_winning_pnl += gross_pnl;
                self.gross_winning_short_commission += comm;
                self.duration_sec_gross_short_winning += secs;
            } else if gross_pnl < 0.0 {
                self.gross_short_loosing_count += 1;
                self.gross_short_loosing_pnl += gross_pnl;
                self.gross_loosing_short_commission += comm;
                self.duration_sec_gross_short_loosing += secs;
            }
            if net_pnl > 0.0 {
                self.net_short_winning_count += 1;
                self.net_short_winning_pnl += gross_pnl; // intentional: uses gross_pnl
                self.net_winning_short_commission += comm;
                self.duration_sec_net_short_winning += secs;
            } else if net_pnl < 0.0 {
                self.net_short_loosing_count += 1;
                self.net_short_loosing_pnl += gross_pnl; // intentional: uses gross_pnl
                self.net_loosing_short_commission += comm;
                self.duration_sec_net_short_loosing += secs;
            }
        }

        // Update first/last times and duration
        let entry_time = *rt.entry_time();
        let exit_time = *rt.exit_time();
        let mut changed = false;
        let update_first = match self.first_time {
            None => true,
            Some(ft) => entry_time.diff_seconds(&ft) < 0.0,
        };
        if update_first {
            self.first_time = Some(entry_time);
            changed = true;
        }
        let update_last = match self.last_time {
            None => true,
            Some(lt) => exit_time.diff_seconds(&lt) > 0.0,
        };
        if update_last {
            self.last_time = Some(exit_time);
            changed = true;
        }
        if changed {
            if let (Some(ft), Some(lt)) = (self.first_time, self.last_time) {
                if let Ok(yf) = year_frac(&ft, &lt, self.day_count_convention) {
                    self.total_duration_annualized = yf;
                }
            }
        }

        let roi = net_pnl / (rt.quantity() * rt.entry_price());
        self.returns_on_investments.push(roi);
        self.roi_mean = Some(slice_mean(&self.returns_on_investments));
        self.roi_std = Some(slice_std_pop(&self.returns_on_investments));

        let downside = roi - self.annual_risk_free_rate;
        if downside < 0.0 {
            self.sortino_downside_returns.push(downside);
            let sum_sq: f64 = self.sortino_downside_returns.iter().map(|v| v * v).sum();
            let tdd = (sum_sq / self.sortino_downside_returns.len() as f64).sqrt();
            self.roi_tdd = Some(tdd);
        }

        // Calculate annualized ROI
        if let Ok(yf) = year_frac(&entry_time, &exit_time, self.day_count_convention) {
            if yf != 0.0 {
                let roiann = roi / yf;
                self.returns_on_investments_annual.push(roiann);
                self.roiann_mean = Some(slice_mean(&self.returns_on_investments_annual));
                self.roiann_std = Some(slice_std_pop(&self.returns_on_investments_annual));

                let downside_ann = roiann - self.annual_risk_free_rate;
                if downside_ann < 0.0 {
                    self.sortino_downside_returns_annual.push(downside_ann);
                    let sum_sq: f64 = self.sortino_downside_returns_annual.iter().map(|v| v * v).sum();
                    let tdd = (sum_sq / self.sortino_downside_returns_annual.len() as f64).sqrt();
                    self.roiann_tdd = Some(tdd);
                }
            }
        }

        // Calculate max drawdown
        if self.max_net_pnl < self.net_pnl {
            self.max_net_pnl = self.net_pnl;
        }
        let dd = self.max_net_pnl - self.net_pnl;
        if self.max_drawdown < dd {
            self.max_drawdown = dd;
            self.max_drawdown_percent = self.max_drawdown / (self.initial_balance + self.max_net_pnl);
        }

        self.roundtrips.push(rt);
    }

    // --- Public field accessors ---

    pub fn initial_balance(&self) -> f64 { self.initial_balance }
    pub fn annual_risk_free_rate(&self) -> f64 { self.annual_risk_free_rate }
    pub fn annual_target_return(&self) -> f64 { self.annual_target_return }
    pub fn day_count_convention(&self) -> DayCountConvention { self.day_count_convention }
    pub fn roundtrips(&self) -> &[Roundtrip] { &self.roundtrips }
    pub fn returns_on_investments(&self) -> &[f64] { &self.returns_on_investments }
    pub fn sortino_downside_returns(&self) -> &[f64] { &self.sortino_downside_returns }
    pub fn returns_on_investments_annual(&self) -> &[f64] { &self.returns_on_investments_annual }
    pub fn sortino_downside_returns_annual(&self) -> &[f64] { &self.sortino_downside_returns_annual }
    pub fn first_time(&self) -> Option<&DateTime> { self.first_time.as_ref() }
    pub fn last_time(&self) -> Option<&DateTime> { self.last_time.as_ref() }
    pub fn max_net_pnl(&self) -> f64 { self.max_net_pnl }
    pub fn max_drawdown(&self) -> f64 { self.max_drawdown }
    pub fn max_drawdown_percent(&self) -> f64 { self.max_drawdown_percent }

    pub fn total_commission(&self) -> f64 { self.total_commission }
    pub fn gross_winning_commission(&self) -> f64 { self.gross_winning_commission }
    pub fn gross_loosing_commission(&self) -> f64 { self.gross_loosing_commission }
    pub fn net_winning_commission(&self) -> f64 { self.net_winning_commission }
    pub fn net_loosing_commission(&self) -> f64 { self.net_loosing_commission }
    pub fn gross_winning_long_commission(&self) -> f64 { self.gross_winning_long_commission }
    pub fn gross_loosing_long_commission(&self) -> f64 { self.gross_loosing_long_commission }
    pub fn net_winning_long_commission(&self) -> f64 { self.net_winning_long_commission }
    pub fn net_loosing_long_commission(&self) -> f64 { self.net_loosing_long_commission }
    pub fn gross_winning_short_commission(&self) -> f64 { self.gross_winning_short_commission }
    pub fn gross_loosing_short_commission(&self) -> f64 { self.gross_loosing_short_commission }
    pub fn net_winning_short_commission(&self) -> f64 { self.net_winning_short_commission }
    pub fn net_loosing_short_commission(&self) -> f64 { self.net_loosing_short_commission }

    // --- ROI statistics ---

    pub fn roi_mean(&self) -> Option<f64> { self.roi_mean }
    pub fn roi_std(&self) -> Option<f64> { self.roi_std }
    pub fn roi_tdd(&self) -> Option<f64> { self.roi_tdd }
    pub fn roiann_mean(&self) -> Option<f64> { self.roiann_mean }
    pub fn roiann_std(&self) -> Option<f64> { self.roiann_std }
    pub fn roiann_tdd(&self) -> Option<f64> { self.roiann_tdd }

    // --- Risk-adjusted ratios ---

    pub fn sharpe_ratio(&self) -> Option<f64> {
        let m = self.roi_mean?;
        let s = self.roi_std?;
        if s == 0.0 { return None; }
        Some(m / s)
    }

    pub fn sharpe_ratio_annual(&self) -> Option<f64> {
        let m = self.roiann_mean?;
        let s = self.roiann_std?;
        if s == 0.0 { return None; }
        Some(m / s)
    }

    pub fn sortino_ratio(&self) -> Option<f64> {
        let m = self.roi_mean?;
        let tdd = self.roi_tdd?;
        if tdd == 0.0 { return None; }
        Some((m - self.annual_risk_free_rate) / tdd)
    }

    pub fn sortino_ratio_annual(&self) -> Option<f64> {
        let m = self.roiann_mean?;
        let tdd = self.roiann_tdd?;
        if tdd == 0.0 { return None; }
        Some((m - self.annual_risk_free_rate) / tdd)
    }

    pub fn calmar_ratio(&self) -> Option<f64> {
        let m = self.roi_mean?;
        if self.max_drawdown_percent == 0.0 { return None; }
        Some(m / self.max_drawdown_percent)
    }

    pub fn calmar_ratio_annual(&self) -> Option<f64> {
        let m = self.roiann_mean?;
        if self.max_drawdown_percent == 0.0 { return None; }
        Some(m / self.max_drawdown_percent)
    }

    // --- Rate of return ---

    pub fn rate_of_return(&self) -> Option<f64> {
        if self.initial_balance == 0.0 { return None; }
        Some(self.net_pnl / self.initial_balance)
    }

    pub fn rate_of_return_annual(&self) -> Option<f64> {
        if self.total_duration_annualized == 0.0 || self.initial_balance == 0.0 {
            return None;
        }
        Some((self.net_pnl / self.initial_balance) / self.total_duration_annualized)
    }

    pub fn recovery_factor(&self) -> Option<f64> {
        let rorann = self.rate_of_return_annual()?;
        if self.max_drawdown_percent == 0.0 { return None; }
        Some(rorann / self.max_drawdown_percent)
    }

    // --- Profit ratios ---

    pub fn gross_profit_ratio(&self) -> Option<f64> {
        if self.gross_loosing_pnl == 0.0 { return None; }
        Some((self.gross_winning_pnl / self.gross_loosing_pnl).abs())
    }

    pub fn net_profit_ratio(&self) -> Option<f64> {
        if self.net_loosing_pnl == 0.0 { return None; }
        Some((self.net_winning_pnl / self.net_loosing_pnl).abs())
    }

    pub fn gross_profit_long_ratio(&self) -> Option<f64> {
        if self.gross_long_loosing_pnl == 0.0 { return None; }
        Some((self.gross_long_winning_pnl / self.gross_long_loosing_pnl).abs())
    }

    pub fn net_profit_long_ratio(&self) -> Option<f64> {
        if self.net_long_loosing_pnl == 0.0 { return None; }
        Some((self.net_long_winning_pnl / self.net_long_loosing_pnl).abs())
    }

    pub fn gross_profit_short_ratio(&self) -> Option<f64> {
        if self.gross_short_loosing_pnl == 0.0 { return None; }
        Some((self.gross_short_winning_pnl / self.gross_short_loosing_pnl).abs())
    }

    pub fn net_profit_short_ratio(&self) -> Option<f64> {
        if self.net_short_loosing_pnl == 0.0 { return None; }
        Some((self.net_short_winning_pnl / self.net_short_loosing_pnl).abs())
    }

    // --- Counts ---

    pub fn total_count(&self) -> usize { self.total_count }
    pub fn long_count(&self) -> usize { self.long_count }
    pub fn short_count(&self) -> usize { self.short_count }
    pub fn gross_winning_count(&self) -> usize { self.gross_winning_count }
    pub fn gross_loosing_count(&self) -> usize { self.gross_loosing_count }
    pub fn net_winning_count(&self) -> usize { self.net_winning_count }
    pub fn net_loosing_count(&self) -> usize { self.net_loosing_count }
    pub fn gross_long_winning_count(&self) -> usize { self.gross_long_winning_count }
    pub fn gross_long_loosing_count(&self) -> usize { self.gross_long_loosing_count }
    pub fn net_long_winning_count(&self) -> usize { self.net_long_winning_count }
    pub fn net_long_loosing_count(&self) -> usize { self.net_long_loosing_count }
    pub fn gross_short_winning_count(&self) -> usize { self.gross_short_winning_count }
    pub fn gross_short_loosing_count(&self) -> usize { self.gross_short_loosing_count }
    pub fn net_short_winning_count(&self) -> usize { self.net_short_winning_count }
    pub fn net_short_loosing_count(&self) -> usize { self.net_short_loosing_count }

    // --- Win/loss ratios ---

    pub fn gross_winning_ratio(&self) -> f64 {
        if self.total_count > 0 { self.gross_winning_count as f64 / self.total_count as f64 } else { 0.0 }
    }
    pub fn gross_loosing_ratio(&self) -> f64 {
        if self.total_count > 0 { self.gross_loosing_count as f64 / self.total_count as f64 } else { 0.0 }
    }
    pub fn net_winning_ratio(&self) -> f64 {
        if self.total_count > 0 { self.net_winning_count as f64 / self.total_count as f64 } else { 0.0 }
    }
    pub fn net_loosing_ratio(&self) -> f64 {
        if self.total_count > 0 { self.net_loosing_count as f64 / self.total_count as f64 } else { 0.0 }
    }
    pub fn gross_long_winning_ratio(&self) -> f64 {
        if self.long_count > 0 { self.gross_long_winning_count as f64 / self.long_count as f64 } else { 0.0 }
    }
    pub fn gross_long_loosing_ratio(&self) -> f64 {
        if self.long_count > 0 { self.gross_long_loosing_count as f64 / self.long_count as f64 } else { 0.0 }
    }
    pub fn net_long_winning_ratio(&self) -> f64 {
        if self.long_count > 0 { self.net_long_winning_count as f64 / self.long_count as f64 } else { 0.0 }
    }
    pub fn net_long_loosing_ratio(&self) -> f64 {
        if self.long_count > 0 { self.net_long_loosing_count as f64 / self.long_count as f64 } else { 0.0 }
    }
    pub fn gross_short_winning_ratio(&self) -> f64 {
        if self.short_count > 0 { self.gross_short_winning_count as f64 / self.short_count as f64 } else { 0.0 }
    }
    pub fn gross_short_loosing_ratio(&self) -> f64 {
        if self.short_count > 0 { self.gross_short_loosing_count as f64 / self.short_count as f64 } else { 0.0 }
    }
    pub fn net_short_winning_ratio(&self) -> f64 {
        if self.short_count > 0 { self.net_short_winning_count as f64 / self.short_count as f64 } else { 0.0 }
    }
    pub fn net_short_loosing_ratio(&self) -> f64 {
        if self.short_count > 0 { self.net_short_loosing_count as f64 / self.short_count as f64 } else { 0.0 }
    }

    // --- PnL totals ---

    pub fn total_gross_pnl(&self) -> f64 { self.gross_pnl }
    pub fn total_net_pnl(&self) -> f64 { self.net_pnl }
    pub fn winning_gross_pnl(&self) -> f64 { self.gross_winning_pnl }
    pub fn loosing_gross_pnl(&self) -> f64 { self.gross_loosing_pnl }
    pub fn winning_net_pnl(&self) -> f64 { self.net_winning_pnl }
    pub fn loosing_net_pnl(&self) -> f64 { self.net_loosing_pnl }
    pub fn winning_gross_long_pnl(&self) -> f64 { self.gross_long_winning_pnl }
    pub fn loosing_gross_long_pnl(&self) -> f64 { self.gross_long_loosing_pnl }
    pub fn winning_net_long_pnl(&self) -> f64 { self.net_long_winning_pnl }
    pub fn loosing_net_long_pnl(&self) -> f64 { self.net_long_loosing_pnl }
    pub fn winning_gross_short_pnl(&self) -> f64 { self.gross_short_winning_pnl }
    pub fn loosing_gross_short_pnl(&self) -> f64 { self.gross_short_loosing_pnl }
    pub fn winning_net_short_pnl(&self) -> f64 { self.net_short_winning_pnl }
    pub fn loosing_net_short_pnl(&self) -> f64 { self.net_short_loosing_pnl }

    // --- Average PnL ---

    pub fn average_gross_pnl(&self) -> f64 { div_or_zero(self.gross_pnl, self.total_count) }
    pub fn average_net_pnl(&self) -> f64 { div_or_zero(self.net_pnl, self.total_count) }
    pub fn average_gross_long_pnl(&self) -> f64 { div_or_zero(self.gross_long_pnl, self.long_count) }
    pub fn average_net_long_pnl(&self) -> f64 { div_or_zero(self.net_long_pnl, self.long_count) }
    pub fn average_gross_short_pnl(&self) -> f64 { div_or_zero(self.gross_short_pnl, self.short_count) }
    pub fn average_net_short_pnl(&self) -> f64 { div_or_zero(self.net_short_pnl, self.short_count) }
    pub fn average_winning_gross_pnl(&self) -> f64 { div_or_zero(self.gross_winning_pnl, self.gross_winning_count) }
    pub fn average_loosing_gross_pnl(&self) -> f64 { div_or_zero(self.gross_loosing_pnl, self.gross_loosing_count) }
    pub fn average_winning_net_pnl(&self) -> f64 { div_or_zero(self.net_winning_pnl, self.net_winning_count) }
    pub fn average_loosing_net_pnl(&self) -> f64 { div_or_zero(self.net_loosing_pnl, self.net_loosing_count) }
    pub fn average_winning_gross_long_pnl(&self) -> f64 { div_or_zero(self.gross_long_winning_pnl, self.gross_long_winning_count) }
    pub fn average_loosing_gross_long_pnl(&self) -> f64 { div_or_zero(self.gross_long_loosing_pnl, self.gross_long_loosing_count) }
    pub fn average_winning_net_long_pnl(&self) -> f64 { div_or_zero(self.net_long_winning_pnl, self.net_long_winning_count) }
    pub fn average_loosing_net_long_pnl(&self) -> f64 { div_or_zero(self.net_long_loosing_pnl, self.net_long_loosing_count) }
    pub fn average_winning_gross_short_pnl(&self) -> f64 { div_or_zero(self.gross_short_winning_pnl, self.gross_short_winning_count) }
    pub fn average_loosing_gross_short_pnl(&self) -> f64 { div_or_zero(self.gross_short_loosing_pnl, self.gross_short_loosing_count) }
    pub fn average_winning_net_short_pnl(&self) -> f64 { div_or_zero(self.net_short_winning_pnl, self.net_short_winning_count) }
    pub fn average_loosing_net_short_pnl(&self) -> f64 { div_or_zero(self.net_short_loosing_pnl, self.net_short_loosing_count) }

    // --- Average win/loss ratios ---

    pub fn average_gross_winning_loosing_ratio(&self) -> f64 {
        let l = self.average_loosing_gross_pnl();
        if l != 0.0 { self.average_winning_gross_pnl() / l } else { 0.0 }
    }
    pub fn average_net_winning_loosing_ratio(&self) -> f64 {
        let l = self.average_loosing_net_pnl();
        if l != 0.0 { self.average_winning_net_pnl() / l } else { 0.0 }
    }
    pub fn average_gross_winning_loosing_long_ratio(&self) -> f64 {
        let l = self.average_loosing_gross_long_pnl();
        if l != 0.0 { self.average_winning_gross_long_pnl() / l } else { 0.0 }
    }
    pub fn average_net_winning_loosing_long_ratio(&self) -> f64 {
        let l = self.average_loosing_net_long_pnl();
        if l != 0.0 { self.average_winning_net_long_pnl() / l } else { 0.0 }
    }
    pub fn average_gross_winning_loosing_short_ratio(&self) -> f64 {
        let l = self.average_loosing_gross_short_pnl();
        if l != 0.0 { self.average_winning_gross_short_pnl() / l } else { 0.0 }
    }
    pub fn average_net_winning_loosing_short_ratio(&self) -> f64 {
        let l = self.average_loosing_net_short_pnl();
        if l != 0.0 { self.average_winning_net_short_pnl() / l } else { 0.0 }
    }

    // --- Profit PnL ratios ---

    pub fn gross_profit_pnl_ratio(&self) -> f64 {
        if self.gross_pnl != 0.0 { self.gross_winning_pnl / self.gross_pnl } else { 0.0 }
    }
    pub fn net_profit_pnl_ratio(&self) -> f64 {
        if self.net_pnl != 0.0 { self.net_winning_pnl / self.net_pnl } else { 0.0 }
    }
    pub fn gross_profit_pnl_long_ratio(&self) -> f64 {
        if self.gross_long_pnl != 0.0 { self.gross_long_winning_pnl / self.gross_long_pnl } else { 0.0 }
    }
    pub fn net_profit_pnl_long_ratio(&self) -> f64 {
        if self.net_long_pnl != 0.0 { self.net_long_winning_pnl / self.net_long_pnl } else { 0.0 }
    }
    pub fn gross_profit_pnl_short_ratio(&self) -> f64 {
        if self.gross_short_pnl != 0.0 { self.gross_short_winning_pnl / self.gross_short_pnl } else { 0.0 }
    }
    pub fn net_profit_pnl_short_ratio(&self) -> f64 {
        if self.net_short_pnl != 0.0 { self.net_short_winning_pnl / self.net_short_pnl } else { 0.0 }
    }

    // --- Average duration ---

    pub fn average_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec, self.total_count) }
    pub fn average_gross_winning_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_gross_winning, self.gross_winning_count) }
    pub fn average_gross_loosing_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_gross_loosing, self.gross_loosing_count) }
    pub fn average_net_winning_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_net_winning, self.net_winning_count) }
    pub fn average_net_loosing_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_net_loosing, self.net_loosing_count) }
    pub fn average_long_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_long, self.long_count) }
    pub fn average_short_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_short, self.short_count) }
    pub fn average_gross_winning_long_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_gross_long_winning, self.gross_long_winning_count) }
    pub fn average_gross_loosing_long_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_gross_long_loosing, self.gross_long_loosing_count) }
    pub fn average_net_winning_long_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_net_long_winning, self.net_long_winning_count) }
    pub fn average_net_loosing_long_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_net_long_loosing, self.net_long_loosing_count) }
    pub fn average_gross_winning_short_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_gross_short_winning, self.gross_short_winning_count) }
    pub fn average_gross_loosing_short_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_gross_short_loosing, self.gross_short_loosing_count) }
    pub fn average_net_winning_short_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_net_short_winning, self.net_short_winning_count) }
    pub fn average_net_loosing_short_duration_seconds(&self) -> f64 { div_or_zero(self.duration_sec_net_short_loosing, self.net_short_loosing_count) }

    // --- Min/max duration helpers ---

    fn filter_duration_seconds<F: Fn(&Roundtrip) -> bool>(&self, filter: F) -> Vec<f64> {
        self.roundtrips.iter()
            .filter(|r| filter(r))
            .map(|r| r.duration_seconds())
            .collect()
    }

    pub fn minimum_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|_| true))
    }
    pub fn maximum_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|_| true))
    }
    pub fn minimum_long_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.side() == RoundtripSide::Long))
    }
    pub fn maximum_long_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.side() == RoundtripSide::Long))
    }
    pub fn minimum_short_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.side() == RoundtripSide::Short))
    }
    pub fn maximum_short_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.side() == RoundtripSide::Short))
    }
    pub fn minimum_gross_winning_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.gross_pnl() > 0.0))
    }
    pub fn maximum_gross_winning_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.gross_pnl() > 0.0))
    }
    pub fn minimum_gross_loosing_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.gross_pnl() < 0.0))
    }
    pub fn maximum_gross_loosing_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.gross_pnl() < 0.0))
    }
    pub fn minimum_net_winning_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.net_pnl() > 0.0))
    }
    pub fn maximum_net_winning_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.net_pnl() > 0.0))
    }
    pub fn minimum_net_loosing_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.net_pnl() < 0.0))
    }
    pub fn maximum_net_loosing_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.net_pnl() < 0.0))
    }
    pub fn minimum_gross_winning_long_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.gross_pnl() > 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn maximum_gross_winning_long_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.gross_pnl() > 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn minimum_gross_loosing_long_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.gross_pnl() < 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn maximum_gross_loosing_long_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.gross_pnl() < 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn minimum_net_winning_long_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.net_pnl() > 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn maximum_net_winning_long_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.net_pnl() > 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn minimum_net_loosing_long_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.net_pnl() < 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn maximum_net_loosing_long_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.net_pnl() < 0.0 && r.side() == RoundtripSide::Long))
    }
    pub fn minimum_gross_winning_short_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.gross_pnl() > 0.0 && r.side() == RoundtripSide::Short))
    }
    pub fn maximum_gross_winning_short_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.gross_pnl() > 0.0 && r.side() == RoundtripSide::Short))
    }
    pub fn minimum_gross_loosing_short_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.gross_pnl() < 0.0 && r.side() == RoundtripSide::Short))
    }
    pub fn maximum_gross_loosing_short_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.gross_pnl() < 0.0 && r.side() == RoundtripSide::Short))
    }
    pub fn minimum_net_winning_short_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.net_pnl() > 0.0 && r.side() == RoundtripSide::Short))
    }
    pub fn maximum_net_winning_short_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.net_pnl() > 0.0 && r.side() == RoundtripSide::Short))
    }
    pub fn minimum_net_loosing_short_duration_seconds(&self) -> f64 {
        min_slice(&self.filter_duration_seconds(|r| r.net_pnl() < 0.0 && r.side() == RoundtripSide::Short))
    }
    pub fn maximum_net_loosing_short_duration_seconds(&self) -> f64 {
        max_slice(&self.filter_duration_seconds(|r| r.net_pnl() < 0.0 && r.side() == RoundtripSide::Short))
    }

    // --- MAE / MFE / efficiency ---

    pub fn average_maximum_adverse_excursion(&self) -> f64 { div_or_zero(self.total_mae, self.total_count) }
    pub fn average_maximum_favorable_excursion(&self) -> f64 { div_or_zero(self.total_mfe, self.total_count) }
    pub fn average_entry_efficiency(&self) -> f64 { div_or_zero(self.total_eff_entry, self.total_count) }
    pub fn average_exit_efficiency(&self) -> f64 { div_or_zero(self.total_eff_exit, self.total_count) }
    pub fn average_total_efficiency(&self) -> f64 { div_or_zero(self.total_eff, self.total_count) }

    // filtered average helper
    fn filtered_avg<FF, FV>(&self, field: FF, filter: FV, count: usize) -> f64
    where
        FF: Fn(&Roundtrip) -> f64,
        FV: Fn(&Roundtrip) -> bool,
    {
        if count == 0 { return 0.0; }
        let sum: f64 = self.roundtrips.iter().filter(|r| filter(r)).map(|r| field(r)).sum();
        sum / count as f64
    }

    pub fn average_maximum_adverse_excursion_gross_winning(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_adverse_excursion(), |r| r.gross_pnl() > 0.0, self.gross_winning_count)
    }
    pub fn average_maximum_adverse_excursion_gross_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_adverse_excursion(), |r| r.gross_pnl() < 0.0, self.gross_loosing_count)
    }
    pub fn average_maximum_adverse_excursion_net_winning(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_adverse_excursion(), |r| r.net_pnl() > 0.0, self.net_winning_count)
    }
    pub fn average_maximum_adverse_excursion_net_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_adverse_excursion(), |r| r.net_pnl() < 0.0, self.net_loosing_count)
    }
    pub fn average_maximum_favorable_excursion_gross_winning(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_favorable_excursion(), |r| r.gross_pnl() > 0.0, self.gross_winning_count)
    }
    pub fn average_maximum_favorable_excursion_gross_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_favorable_excursion(), |r| r.gross_pnl() < 0.0, self.gross_loosing_count)
    }
    pub fn average_maximum_favorable_excursion_net_winning(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_favorable_excursion(), |r| r.net_pnl() > 0.0, self.net_winning_count)
    }
    pub fn average_maximum_favorable_excursion_net_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.maximum_favorable_excursion(), |r| r.net_pnl() < 0.0, self.net_loosing_count)
    }
    pub fn average_entry_efficiency_gross_winning(&self) -> f64 {
        self.filtered_avg(|r| r.entry_efficiency(), |r| r.gross_pnl() > 0.0, self.gross_winning_count)
    }
    pub fn average_entry_efficiency_gross_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.entry_efficiency(), |r| r.gross_pnl() < 0.0, self.gross_loosing_count)
    }
    pub fn average_entry_efficiency_net_winning(&self) -> f64 {
        self.filtered_avg(|r| r.entry_efficiency(), |r| r.net_pnl() > 0.0, self.net_winning_count)
    }
    pub fn average_entry_efficiency_net_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.entry_efficiency(), |r| r.net_pnl() < 0.0, self.net_loosing_count)
    }
    pub fn average_exit_efficiency_gross_winning(&self) -> f64 {
        self.filtered_avg(|r| r.exit_efficiency(), |r| r.gross_pnl() > 0.0, self.gross_winning_count)
    }
    pub fn average_exit_efficiency_gross_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.exit_efficiency(), |r| r.gross_pnl() < 0.0, self.gross_loosing_count)
    }
    pub fn average_exit_efficiency_net_winning(&self) -> f64 {
        self.filtered_avg(|r| r.exit_efficiency(), |r| r.net_pnl() > 0.0, self.net_winning_count)
    }
    pub fn average_exit_efficiency_net_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.exit_efficiency(), |r| r.net_pnl() < 0.0, self.net_loosing_count)
    }
    pub fn average_total_efficiency_gross_winning(&self) -> f64 {
        self.filtered_avg(|r| r.total_efficiency(), |r| r.gross_pnl() > 0.0, self.gross_winning_count)
    }
    pub fn average_total_efficiency_gross_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.total_efficiency(), |r| r.gross_pnl() < 0.0, self.gross_loosing_count)
    }
    pub fn average_total_efficiency_net_winning(&self) -> f64 {
        self.filtered_avg(|r| r.total_efficiency(), |r| r.net_pnl() > 0.0, self.net_winning_count)
    }
    pub fn average_total_efficiency_net_loosing(&self) -> f64 {
        self.filtered_avg(|r| r.total_efficiency(), |r| r.net_pnl() < 0.0, self.net_loosing_count)
    }

    // --- Consecutive streaks ---

    pub fn max_consecutive_gross_winners(&self) -> usize {
        let bools: Vec<bool> = self.roundtrips.iter().map(|r| r.gross_pnl() > 0.0).collect();
        max_consecutive(&bools)
    }
    pub fn max_consecutive_gross_loosers(&self) -> usize {
        let bools: Vec<bool> = self.roundtrips.iter().map(|r| r.gross_pnl() < 0.0).collect();
        max_consecutive(&bools)
    }
    pub fn max_consecutive_net_winners(&self) -> usize {
        let bools: Vec<bool> = self.roundtrips.iter().map(|r| r.net_pnl() > 0.0).collect();
        max_consecutive(&bools)
    }
    pub fn max_consecutive_net_loosers(&self) -> usize {
        let bools: Vec<bool> = self.roundtrips.iter().map(|r| r.net_pnl() < 0.0).collect();
        max_consecutive(&bools)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::roundtrips::execution::{Execution, OrderSide};

    const EPSILON: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool {
        (a - b).abs() < eps
    }

    fn assert_almost(name: &str, got: f64, expected: f64, eps: f64) {
        assert!(almost_equal(got, expected, eps),
            "{}: expected {}, got {}", name, expected, got);
    }

    fn assert_opt_almost(name: &str, got: Option<f64>, expected: f64, eps: f64) {
        assert!(got.is_some(), "{}: expected {}, got None", name, expected);
        assert!(almost_equal(got.unwrap(), expected, eps),
            "{}: expected {}, got {}", name, expected, got.unwrap());
    }

    fn assert_opt_none(name: &str, got: Option<f64>) {
        assert!(got.is_none(), "{}: expected None, got {:?}", name, got);
    }

    fn exec(side: OrderSide, price: f64, comm: f64, high: f64, low: f64,
            yr: i32, mo: i32, dy: i32, hr: i32, mi: i32, se: i32) -> Execution {
        Execution {
            side,
            price,
            commission_per_unit: comm,
            unrealized_price_high: high,
            unrealized_price_low: low,
            datetime: DateTime::new(yr, mo, dy, hr, mi, se),
        }
    }

    fn make_rt1() -> Roundtrip {
        Roundtrip::new(
            &exec(OrderSide::Buy, 50.0, 0.01, 56.0, 48.0, 2024, 1, 1, 9, 30, 0),
            &exec(OrderSide::Sell, 55.0, 0.02, 57.0, 49.0, 2024, 1, 5, 16, 0, 0),
            100.0)
    }
    fn make_rt2() -> Roundtrip {
        Roundtrip::new(
            &exec(OrderSide::Sell, 80.0, 0.03, 85.0, 72.0, 2024, 2, 1, 10, 0, 0),
            &exec(OrderSide::Buy, 72.0, 0.02, 83.0, 70.0, 2024, 2, 10, 15, 30, 0),
            200.0)
    }
    fn make_rt3() -> Roundtrip {
        Roundtrip::new(
            &exec(OrderSide::Buy, 60.0, 0.005, 62.0, 53.0, 2024, 3, 1, 9, 30, 0),
            &exec(OrderSide::Sell, 54.0, 0.005, 61.0, 52.0, 2024, 3, 3, 16, 0, 0),
            150.0)
    }
    fn make_rt4() -> Roundtrip {
        Roundtrip::new(
            &exec(OrderSide::Sell, 40.0, 0.01, 42.0, 39.0, 2024, 4, 1, 10, 0, 0),
            &exec(OrderSide::Buy, 45.0, 0.01, 46.0, 38.0, 2024, 4, 5, 15, 0, 0),
            300.0)
    }
    fn make_rt5() -> Roundtrip {
        Roundtrip::new(
            &exec(OrderSide::Buy, 100.0, 0.02, 112.0, 98.0, 2024, 5, 1, 9, 0, 0),
            &exec(OrderSide::Sell, 110.0, 0.02, 115.0, 99.0, 2024, 5, 15, 16, 0, 0),
            50.0)
    }
    fn make_rt6() -> Roundtrip {
        Roundtrip::new(
            &exec(OrderSide::Sell, 90.0, 0.015, 92.0, 84.0, 2024, 6, 1, 10, 0, 0),
            &exec(OrderSide::Buy, 82.0, 0.015, 93.0, 80.0, 2024, 6, 20, 15, 0, 0),
            100.0)
    }

    fn make_all_rts() -> Vec<Roundtrip> {
        vec![make_rt1(), make_rt2(), make_rt3(), make_rt4(), make_rt5(), make_rt6()]
    }

    fn make_perf_with_all() -> RoundtripPerformance {
        let mut perf = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        for rt in make_all_rts() {
            perf.add_roundtrip(rt);
        }
        perf
    }

    // ====================== Initial state ======================

    #[test] fn test_init_initial_balance() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("initial_balance", p.initial_balance(), 100000.0, EPSILON);
    }
    #[test] fn test_init_annual_risk_free_rate() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("annual_risk_free_rate", p.annual_risk_free_rate(), 0.0, EPSILON);
    }
    #[test] fn test_init_total_count_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_eq!(p.total_count(), 0);
    }
    #[test] fn test_init_roi_mean_none() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("roi_mean", p.roi_mean());
    }
    #[test] fn test_init_roi_std_none() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("roi_std", p.roi_std());
    }
    #[test] fn test_init_roi_tdd_none() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("roi_tdd", p.roi_tdd());
    }
    #[test] fn test_init_sharpe_none() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("sharpe", p.sharpe_ratio());
    }
    #[test] fn test_init_sortino_none() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("sortino", p.sortino_ratio());
    }
    #[test] fn test_init_calmar_none() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("calmar", p.calmar_ratio());
    }
    #[test] fn test_init_roundtrips_empty() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_eq!(p.roundtrips().len(), 0);
    }
    #[test] fn test_init_total_gross_pnl_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("total_gross_pnl", p.total_gross_pnl(), 0.0, EPSILON);
    }
    #[test] fn test_init_total_net_pnl_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("total_net_pnl", p.total_net_pnl(), 0.0, EPSILON);
    }
    #[test] fn test_init_max_drawdown_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("max_drawdown", p.max_drawdown(), 0.0, EPSILON);
    }
    #[test] fn test_init_average_net_pnl_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("average_net_pnl", p.average_net_pnl(), 0.0, EPSILON);
    }

    // ====================== Reset ======================

    #[test] fn test_reset_total_count_zero() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        p.add_roundtrip(make_rt3());
        p.reset();
        assert_eq!(p.total_count(), 0);
    }
    #[test] fn test_reset_total_net_pnl_zero() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        p.add_roundtrip(make_rt3());
        p.reset();
        assert_almost("total_net_pnl", p.total_net_pnl(), 0.0, EPSILON);
    }
    #[test] fn test_reset_roi_mean_none() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        p.add_roundtrip(make_rt3());
        p.reset();
        assert_opt_none("roi_mean", p.roi_mean());
    }
    #[test] fn test_reset_roundtrips_empty() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        p.add_roundtrip(make_rt3());
        p.reset();
        assert_eq!(p.roundtrips().len(), 0);
    }
    #[test] fn test_reset_returns_empty() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        p.add_roundtrip(make_rt3());
        p.reset();
        assert_eq!(p.returns_on_investments().len(), 0);
    }
    #[test] fn test_reset_max_drawdown_zero() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        p.add_roundtrip(make_rt3());
        p.reset();
        assert_almost("max_drawdown", p.max_drawdown(), 0.0, EPSILON);
    }

    // ====================== Single long winner ======================

    #[test] fn test_single_winner_total_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.total_count(), 1);
    }
    #[test] fn test_single_winner_long_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.long_count(), 1);
    }
    #[test] fn test_single_winner_short_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.short_count(), 0);
    }
    #[test] fn test_single_winner_gross_winning_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.gross_winning_count(), 1);
    }
    #[test] fn test_single_winner_gross_loosing_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.gross_loosing_count(), 0);
    }
    #[test] fn test_single_winner_net_winning_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.net_winning_count(), 1);
    }
    #[test] fn test_single_winner_net_loosing_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.net_loosing_count(), 0);
    }
    #[test] fn test_single_winner_total_gross_pnl() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_almost("total_gross_pnl", p.total_gross_pnl(), 500.0, EPSILON);
    }
    #[test] fn test_single_winner_total_net_pnl() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_almost("total_net_pnl", p.total_net_pnl(), 497.0, EPSILON);
    }
    #[test] fn test_single_winner_total_commission() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_almost("total_commission", p.total_commission(), 3.0, EPSILON);
    }
    #[test] fn test_single_winner_roi_mean() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_almost("roi_mean", p.roi_mean(), 0.0994, EPSILON);
    }
    #[test] fn test_single_winner_roi_std_zero() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_almost("roi_std", p.roi_std(), 0.0, EPSILON);
    }
    #[test] fn test_single_winner_roi_tdd_none() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("roi_tdd", p.roi_tdd());
    }
    #[test] fn test_single_winner_sharpe_none() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("sharpe", p.sharpe_ratio());
    }
    #[test] fn test_single_winner_sortino_none() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("sortino", p.sortino_ratio());
    }
    #[test] fn test_single_winner_calmar_none() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("calmar", p.calmar_ratio());
    }
    #[test] fn test_single_winner_max_drawdown_zero() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_almost("max_drawdown", p.max_drawdown(), 0.0, EPSILON);
    }
    #[test] fn test_single_winner_rate_of_return() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_almost("rate_of_return", p.rate_of_return(), 0.00497, EPSILON);
    }
    #[test] fn test_single_winner_gross_winning_ratio() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_almost("gross_winning_ratio", p.gross_winning_ratio(), 1.0, EPSILON);
    }
    #[test] fn test_single_winner_net_winning_ratio() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_almost("net_winning_ratio", p.net_winning_ratio(), 1.0, EPSILON);
    }
    #[test] fn test_single_winner_gross_profit_ratio_none() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("gross_profit_ratio", p.gross_profit_ratio());
    }
    #[test] fn test_single_winner_net_profit_ratio_none() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("net_profit_ratio", p.net_profit_ratio());
    }
    #[test] fn test_single_winner_avg_mae() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        let rt = make_rt1();
        let mae = rt.maximum_adverse_excursion();
        p.add_roundtrip(rt);
        assert_almost("avg_mae", p.average_maximum_adverse_excursion(), mae, EPSILON);
    }
    #[test] fn test_single_winner_avg_mfe() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        let rt = make_rt1();
        let mfe = rt.maximum_favorable_excursion();
        p.add_roundtrip(rt);
        assert_almost("avg_mfe", p.average_maximum_favorable_excursion(), mfe, EPSILON);
    }
    #[test] fn test_single_winner_avg_entry_eff() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        let rt = make_rt1();
        let ee = rt.entry_efficiency();
        p.add_roundtrip(rt);
        assert_almost("avg_entry_eff", p.average_entry_efficiency(), ee, EPSILON);
    }
    #[test] fn test_single_winner_avg_exit_eff() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        let rt = make_rt1();
        let xe = rt.exit_efficiency();
        p.add_roundtrip(rt);
        assert_almost("avg_exit_eff", p.average_exit_efficiency(), xe, EPSILON);
    }
    #[test] fn test_single_winner_avg_total_eff() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        let rt = make_rt1();
        let te = rt.total_efficiency();
        p.add_roundtrip(rt);
        assert_almost("avg_total_eff", p.average_total_efficiency(), te, EPSILON);
    }
    #[test] fn test_single_winner_avg_duration() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_almost("avg_duration", p.average_duration_seconds(), 369000.0, EPSILON);
    }
    #[test] fn test_single_winner_consecutive_gross_winners() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.max_consecutive_gross_winners(), 1);
    }
    #[test] fn test_single_winner_consecutive_gross_loosers() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_eq!(p.max_consecutive_gross_loosers(), 0);
    }

    // ====================== Single long loser ======================

    #[test] fn test_single_looser_total_net_pnl() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_almost("total_net_pnl", p.total_net_pnl(), -901.5, EPSILON);
    }
    #[test] fn test_single_looser_max_drawdown() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_almost("max_drawdown", p.max_drawdown(), 901.5, EPSILON);
    }
    #[test] fn test_single_looser_max_drawdown_percent() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_almost("max_drawdown_pct", p.max_drawdown_percent(), 0.009015, EPSILON);
    }
    #[test] fn test_single_looser_calmar() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_opt_almost("calmar", p.calmar_ratio(), -11.11111111111111, 1e-10);
    }
    #[test] fn test_single_looser_roi_mean() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_opt_almost("roi_mean", p.roi_mean(), -0.10016666666666667, EPSILON);
    }
    #[test] fn test_single_looser_roi_tdd() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_opt_almost("roi_tdd", p.roi_tdd(), 0.10016666666666667, EPSILON);
    }
    #[test] fn test_single_looser_sortino() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_opt_almost("sortino", p.sortino_ratio(), -1.0, EPSILON);
    }
    #[test] fn test_single_looser_gross_loosing_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_eq!(p.gross_loosing_count(), 1);
    }
    #[test] fn test_single_looser_net_loosing_count() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt3());
        assert_eq!(p.net_loosing_count(), 1);
    }

    // ====================== Multiple mixed (all 6) - counts ======================

    #[test] fn test_mixed_total_count() { assert_eq!(make_perf_with_all().total_count(), 6); }
    #[test] fn test_mixed_long_count() { assert_eq!(make_perf_with_all().long_count(), 3); }
    #[test] fn test_mixed_short_count() { assert_eq!(make_perf_with_all().short_count(), 3); }
    #[test] fn test_mixed_gross_winning_count() { assert_eq!(make_perf_with_all().gross_winning_count(), 4); }
    #[test] fn test_mixed_gross_loosing_count() { assert_eq!(make_perf_with_all().gross_loosing_count(), 2); }
    #[test] fn test_mixed_net_winning_count() { assert_eq!(make_perf_with_all().net_winning_count(), 4); }
    #[test] fn test_mixed_net_loosing_count() { assert_eq!(make_perf_with_all().net_loosing_count(), 2); }
    #[test] fn test_mixed_gross_long_winning_count() { assert_eq!(make_perf_with_all().gross_long_winning_count(), 2); }
    #[test] fn test_mixed_gross_long_loosing_count() { assert_eq!(make_perf_with_all().gross_long_loosing_count(), 1); }
    #[test] fn test_mixed_net_long_winning_count() { assert_eq!(make_perf_with_all().net_long_winning_count(), 2); }
    #[test] fn test_mixed_net_long_loosing_count() { assert_eq!(make_perf_with_all().net_long_loosing_count(), 1); }
    #[test] fn test_mixed_gross_short_winning_count() { assert_eq!(make_perf_with_all().gross_short_winning_count(), 2); }
    #[test] fn test_mixed_gross_short_loosing_count() { assert_eq!(make_perf_with_all().gross_short_loosing_count(), 1); }
    #[test] fn test_mixed_net_short_winning_count() { assert_eq!(make_perf_with_all().net_short_winning_count(), 2); }
    #[test] fn test_mixed_net_short_loosing_count() { assert_eq!(make_perf_with_all().net_short_loosing_count(), 1); }

    // ====================== Multiple mixed - PnL totals ======================

    #[test] fn test_mixed_total_gross_pnl() {
        assert_almost("total_gross_pnl", make_perf_with_all().total_gross_pnl(), 1000.0, EPSILON);
    }
    #[test] fn test_mixed_total_net_pnl() {
        assert_almost("total_net_pnl", make_perf_with_all().total_net_pnl(), 974.5, EPSILON);
    }
    #[test] fn test_mixed_winning_gross_pnl() {
        assert_almost("winning_gross_pnl", make_perf_with_all().winning_gross_pnl(), 3400.0, EPSILON);
    }
    #[test] fn test_mixed_loosing_gross_pnl() {
        assert_almost("loosing_gross_pnl", make_perf_with_all().loosing_gross_pnl(), -2400.0, EPSILON);
    }
    #[test] fn test_mixed_winning_net_pnl() {
        assert_almost("winning_net_pnl", make_perf_with_all().winning_net_pnl(), 3382.0, EPSILON);
    }
    #[test] fn test_mixed_loosing_net_pnl() {
        assert_almost("loosing_net_pnl", make_perf_with_all().loosing_net_pnl(), -2407.5, EPSILON);
    }
    #[test] fn test_mixed_winning_gross_long_pnl() {
        assert_almost("winning_gross_long_pnl", make_perf_with_all().winning_gross_long_pnl(), 1000.0, EPSILON);
    }
    #[test] fn test_mixed_loosing_gross_long_pnl() {
        assert_almost("loosing_gross_long_pnl", make_perf_with_all().loosing_gross_long_pnl(), -900.0, EPSILON);
    }
    #[test] fn test_mixed_winning_gross_short_pnl() {
        assert_almost("winning_gross_short_pnl", make_perf_with_all().winning_gross_short_pnl(), 2400.0, EPSILON);
    }
    #[test] fn test_mixed_loosing_gross_short_pnl() {
        assert_almost("loosing_gross_short_pnl", make_perf_with_all().loosing_gross_short_pnl(), -1500.0, EPSILON);
    }

    // ====================== Multiple mixed - commission ======================

    #[test] fn test_mixed_total_commission() {
        assert_almost("total_commission", make_perf_with_all().total_commission(), 25.5, EPSILON);
    }
    #[test] fn test_mixed_gross_winning_commission() {
        assert_almost("gross_winning_commission", make_perf_with_all().gross_winning_commission(), 18.0, EPSILON);
    }
    #[test] fn test_mixed_gross_loosing_commission() {
        assert_almost("gross_loosing_commission", make_perf_with_all().gross_loosing_commission(), 7.5, EPSILON);
    }
    #[test] fn test_mixed_net_winning_commission() {
        assert_almost("net_winning_commission", make_perf_with_all().net_winning_commission(), 18.0, EPSILON);
    }
    #[test] fn test_mixed_net_loosing_commission() {
        assert_almost("net_loosing_commission", make_perf_with_all().net_loosing_commission(), 7.5, EPSILON);
    }

    // ====================== Multiple mixed - average PnL ======================

    #[test] fn test_mixed_average_gross_pnl() {
        assert_almost("avg_gross_pnl", make_perf_with_all().average_gross_pnl(), 1000.0 / 6.0, EPSILON);
    }
    #[test] fn test_mixed_average_net_pnl() {
        assert_almost("avg_net_pnl", make_perf_with_all().average_net_pnl(), 974.5 / 6.0, EPSILON);
    }
    #[test] fn test_mixed_average_winning_gross_pnl() {
        assert_almost("avg_win_gross_pnl", make_perf_with_all().average_winning_gross_pnl(), 3400.0 / 4.0, EPSILON);
    }
    #[test] fn test_mixed_average_loosing_gross_pnl() {
        assert_almost("avg_loose_gross_pnl", make_perf_with_all().average_loosing_gross_pnl(), -2400.0 / 2.0, EPSILON);
    }
    #[test] fn test_mixed_average_winning_net_pnl() {
        assert_almost("avg_win_net_pnl", make_perf_with_all().average_winning_net_pnl(), 3382.0 / 4.0, EPSILON);
    }
    #[test] fn test_mixed_average_loosing_net_pnl() {
        assert_almost("avg_loose_net_pnl", make_perf_with_all().average_loosing_net_pnl(), -2407.5 / 2.0, EPSILON);
    }
    #[test] fn test_mixed_average_gross_long_pnl() {
        assert_almost("avg_gross_long_pnl", make_perf_with_all().average_gross_long_pnl(), 100.0 / 3.0, EPSILON);
    }
    #[test] fn test_mixed_average_gross_short_pnl() {
        assert_almost("avg_gross_short_pnl", make_perf_with_all().average_gross_short_pnl(), 300.0, EPSILON);
    }

    // ====================== Multiple mixed - win/loss ratios ======================

    #[test] fn test_mixed_gross_winning_ratio() {
        assert_almost("gwr", make_perf_with_all().gross_winning_ratio(), 4.0 / 6.0, EPSILON);
    }
    #[test] fn test_mixed_gross_loosing_ratio() {
        assert_almost("glr", make_perf_with_all().gross_loosing_ratio(), 2.0 / 6.0, EPSILON);
    }
    #[test] fn test_mixed_net_winning_ratio() {
        assert_almost("nwr", make_perf_with_all().net_winning_ratio(), 4.0 / 6.0, EPSILON);
    }
    #[test] fn test_mixed_net_loosing_ratio() {
        assert_almost("nlr", make_perf_with_all().net_loosing_ratio(), 2.0 / 6.0, EPSILON);
    }
    #[test] fn test_mixed_gross_long_winning_ratio() {
        assert_almost("glwr", make_perf_with_all().gross_long_winning_ratio(), 2.0 / 3.0, EPSILON);
    }
    #[test] fn test_mixed_gross_short_winning_ratio() {
        assert_almost("gswr", make_perf_with_all().gross_short_winning_ratio(), 2.0 / 3.0, EPSILON);
    }

    // ====================== Multiple mixed - profit ratios ======================

    #[test] fn test_mixed_gross_profit_ratio() {
        assert_opt_almost("gpr", make_perf_with_all().gross_profit_ratio(), 1.4166666666666667, EPSILON);
    }
    #[test] fn test_mixed_net_profit_ratio() {
        assert_opt_almost("npr", make_perf_with_all().net_profit_ratio(), 1.4047767393561785, EPSILON);
    }
    #[test] fn test_mixed_gross_profit_long_ratio() {
        assert_opt_almost("gplr", make_perf_with_all().gross_profit_long_ratio(), 1.1111111111111112, EPSILON);
    }
    #[test] fn test_mixed_gross_profit_short_ratio() {
        assert_opt_almost("gpsr", make_perf_with_all().gross_profit_short_ratio(), 1.6, EPSILON);
    }

    // ====================== Multiple mixed - profit PnL ratios ======================

    #[test] fn test_mixed_gross_profit_pnl_ratio() {
        assert_almost("gppr", make_perf_with_all().gross_profit_pnl_ratio(), 3.4, EPSILON);
    }
    #[test] fn test_mixed_net_profit_pnl_ratio() {
        assert_almost("nppr", make_perf_with_all().net_profit_pnl_ratio(), 3382.0 / 974.5, EPSILON);
    }

    // ====================== Multiple mixed - average win/loss ratio ======================

    #[test] fn test_mixed_avg_gross_winning_loosing_ratio() {
        assert_almost("agwlr", make_perf_with_all().average_gross_winning_loosing_ratio(), 850.0 / -1200.0, EPSILON);
    }
    #[test] fn test_mixed_avg_net_winning_loosing_ratio() {
        assert_almost("anwlr", make_perf_with_all().average_net_winning_loosing_ratio(), 845.5 / -1203.75, EPSILON);
    }

    // ====================== Multiple mixed - ROI statistics ======================

    #[test] fn test_mixed_roi_mean() {
        assert_opt_almost("roi_mean", make_perf_with_all().roi_mean(), 0.026877314814814812, EPSILON);
    }
    #[test] fn test_mixed_roi_std() {
        assert_opt_almost("roi_std", make_perf_with_all().roi_std(), 0.0991356544050762, EPSILON);
    }
    #[test] fn test_mixed_roi_tdd() {
        assert_opt_almost("roi_tdd", make_perf_with_all().roi_tdd(), 0.11354208715518468, EPSILON);
    }
    #[test] fn test_mixed_roiann_mean() {
        assert_opt_almost("roiann_mean", make_perf_with_all().roiann_mean(), -1.7233887909446202, 1e-12);
    }
    #[test] fn test_mixed_roiann_std() {
        assert_opt_almost("roiann_std", make_perf_with_all().roiann_std(), 8.73138705463156, 1e-12);
    }
    #[test] fn test_mixed_roiann_tdd() {
        assert_opt_almost("roiann_tdd", make_perf_with_all().roiann_tdd(), 13.751365296707874, 1e-12);
    }

    // ====================== Multiple mixed - risk-adjusted ratios ======================

    #[test] fn test_mixed_sharpe_ratio() {
        assert_opt_almost("sharpe", make_perf_with_all().sharpe_ratio(), 0.27111653194916085, EPSILON);
    }
    #[test] fn test_mixed_sharpe_ratio_annual() {
        assert_opt_almost("sharpe_ann", make_perf_with_all().sharpe_ratio_annual(), -0.1973785814512082, 1e-12);
    }
    #[test] fn test_mixed_sortino_ratio() {
        assert_opt_almost("sortino", make_perf_with_all().sortino_ratio(), 0.23671675841293985, EPSILON);
    }
    #[test] fn test_mixed_sortino_ratio_annual() {
        assert_opt_almost("sortino_ann", make_perf_with_all().sortino_ratio_annual(), -0.1253249225629404, 1e-12);
    }
    #[test] fn test_mixed_calmar_ratio() {
        assert_opt_almost("calmar", make_perf_with_all().calmar_ratio(), 1.139698624091381, 1e-12);
    }
    #[test] fn test_mixed_calmar_ratio_annual() {
        assert_opt_almost("calmar_ann", make_perf_with_all().calmar_ratio_annual(), -73.07812731097131, 1e-10);
    }

    // ====================== Multiple mixed - rate of return ======================

    #[test] fn test_mixed_rate_of_return() {
        assert_opt_almost("ror", make_perf_with_all().rate_of_return(), 0.009745, EPSILON);
    }
    #[test] fn test_mixed_rate_of_return_annual() {
        assert_opt_almost("ror_ann", make_perf_with_all().rate_of_return_annual(), 0.020786693247353695, 1e-12);
    }
    #[test] fn test_mixed_recovery_factor() {
        assert_opt_almost("recovery", make_perf_with_all().recovery_factor(), 0.8814335009522727, 1e-12);
    }

    // ====================== Multiple mixed - drawdown ======================

    #[test] fn test_mixed_max_net_pnl() {
        assert_almost("max_net_pnl", make_perf_with_all().max_net_pnl(), 2087.0, EPSILON);
    }
    #[test] fn test_mixed_max_drawdown() {
        assert_almost("max_drawdown", make_perf_with_all().max_drawdown(), 2407.5, EPSILON);
    }
    #[test] fn test_mixed_max_drawdown_percent() {
        let expected = 2407.5 / (100000.0 + 2087.0);
        assert_almost("max_dd_pct", make_perf_with_all().max_drawdown_percent(), expected, EPSILON);
    }

    // ====================== Multiple mixed - duration ======================

    #[test] fn test_mixed_avg_duration() {
        assert_almost("avg_dur", make_perf_with_all().average_duration_seconds(), 770100.0, EPSILON);
    }
    #[test] fn test_mixed_avg_long_duration() {
        assert_almost("avg_long_dur", make_perf_with_all().average_long_duration_seconds(), 600000.0, EPSILON);
    }
    #[test] fn test_mixed_avg_short_duration() {
        assert_almost("avg_short_dur", make_perf_with_all().average_short_duration_seconds(), 940200.0, EPSILON);
    }
    #[test] fn test_mixed_avg_gross_winning_duration() {
        assert_almost("avg_gw_dur", make_perf_with_all().average_gross_winning_duration_seconds(), 1015200.0, EPSILON);
    }
    #[test] fn test_mixed_avg_gross_loosing_duration() {
        assert_almost("avg_gl_dur", make_perf_with_all().average_gross_loosing_duration_seconds(), 279900.0, EPSILON);
    }
    #[test] fn test_mixed_min_duration() {
        assert_almost("min_dur", make_perf_with_all().minimum_duration_seconds(), 196200.0, EPSILON);
    }
    #[test] fn test_mixed_max_duration() {
        assert_almost("max_dur", make_perf_with_all().maximum_duration_seconds(), 1659600.0, EPSILON);
    }
    #[test] fn test_mixed_min_long_duration() {
        assert_almost("min_long_dur", make_perf_with_all().minimum_long_duration_seconds(), 196200.0, EPSILON);
    }
    #[test] fn test_mixed_max_long_duration() {
        assert_almost("max_long_dur", make_perf_with_all().maximum_long_duration_seconds(), 1234800.0, EPSILON);
    }
    #[test] fn test_mixed_min_short_duration() {
        assert_almost("min_short_dur", make_perf_with_all().minimum_short_duration_seconds(), 363600.0, EPSILON);
    }
    #[test] fn test_mixed_max_short_duration() {
        assert_almost("max_short_dur", make_perf_with_all().maximum_short_duration_seconds(), 1659600.0, EPSILON);
    }

    // ====================== Multiple mixed - MAE/MFE/efficiency ======================

    #[test] fn test_mixed_avg_mae() {
        let rts = make_all_rts();
        let sum: f64 = rts.iter().map(|r| r.maximum_adverse_excursion()).sum();
        let p = make_perf_with_all();
        assert_almost("avg_mae", p.average_maximum_adverse_excursion(), sum / 6.0, EPSILON);
    }
    #[test] fn test_mixed_avg_mfe() {
        let rts = make_all_rts();
        let sum: f64 = rts.iter().map(|r| r.maximum_favorable_excursion()).sum();
        let p = make_perf_with_all();
        assert_almost("avg_mfe", p.average_maximum_favorable_excursion(), sum / 6.0, EPSILON);
    }
    #[test] fn test_mixed_avg_entry_eff() {
        let rts = make_all_rts();
        let sum: f64 = rts.iter().map(|r| r.entry_efficiency()).sum();
        let p = make_perf_with_all();
        assert_almost("avg_entry_eff", p.average_entry_efficiency(), sum / 6.0, EPSILON);
    }
    #[test] fn test_mixed_avg_exit_eff() {
        let rts = make_all_rts();
        let sum: f64 = rts.iter().map(|r| r.exit_efficiency()).sum();
        let p = make_perf_with_all();
        assert_almost("avg_exit_eff", p.average_exit_efficiency(), sum / 6.0, EPSILON);
    }
    #[test] fn test_mixed_avg_total_eff() {
        let rts = make_all_rts();
        let sum: f64 = rts.iter().map(|r| r.total_efficiency()).sum();
        let p = make_perf_with_all();
        assert_almost("avg_total_eff", p.average_total_efficiency(), sum / 6.0, EPSILON);
    }

    // ====================== Multiple mixed - consecutive ======================

    #[test] fn test_mixed_consecutive_gross_winners() {
        assert_eq!(make_perf_with_all().max_consecutive_gross_winners(), 2);
    }
    #[test] fn test_mixed_consecutive_gross_loosers() {
        assert_eq!(make_perf_with_all().max_consecutive_gross_loosers(), 2);
    }
    #[test] fn test_mixed_consecutive_net_winners() {
        assert_eq!(make_perf_with_all().max_consecutive_net_winners(), 2);
    }
    #[test] fn test_mixed_consecutive_net_loosers() {
        assert_eq!(make_perf_with_all().max_consecutive_net_loosers(), 2);
    }

    // ====================== Multiple mixed - time tracking ======================

    #[test] fn test_mixed_first_time() {
        let p = make_perf_with_all();
        let ft = p.first_time().unwrap();
        assert_eq!(ft.year, 2024);
        assert_eq!(ft.month, 1);
        assert_eq!(ft.day, 1);
        assert_eq!(ft.hour, 9);
        assert_eq!(ft.minute, 30);
        assert_eq!(ft.second, 0);
    }
    #[test] fn test_mixed_last_time() {
        let p = make_perf_with_all();
        let lt = p.last_time().unwrap();
        assert_eq!(lt.year, 2024);
        assert_eq!(lt.month, 6);
        assert_eq!(lt.day, 20);
        assert_eq!(lt.hour, 15);
        assert_eq!(lt.minute, 0);
        assert_eq!(lt.second, 0);
    }

    // ====================== Edge cases ======================

    #[test] fn test_edge_zero_balance_ror_none() {
        let p = RoundtripPerformance::new(0.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("ror", p.rate_of_return());
    }
    #[test] fn test_edge_no_rt_avg_gross_pnl_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("avg_gross_pnl", p.average_gross_pnl(), 0.0, EPSILON);
    }
    #[test] fn test_edge_no_rt_avg_net_pnl_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("avg_net_pnl", p.average_net_pnl(), 0.0, EPSILON);
    }
    #[test] fn test_edge_no_rt_gross_winning_ratio_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("gwr", p.gross_winning_ratio(), 0.0, EPSILON);
    }
    #[test] fn test_edge_no_rt_avg_duration_zero() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_almost("avg_dur", p.average_duration_seconds(), 0.0, EPSILON);
    }
    #[test] fn test_edge_sharpe_none_single_point() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("sharpe", p.sharpe_ratio());
    }
    #[test] fn test_edge_ror_annual_none_zero_duration() {
        let p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        assert_opt_none("ror_ann", p.rate_of_return_annual());
    }
    #[test] fn test_edge_recovery_factor_none_no_drawdown() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        p.add_roundtrip(make_rt1());
        assert_opt_none("recovery", p.recovery_factor());
    }

    // ====================== Incremental ======================

    #[test] fn test_incremental_roi_list_length() {
        let mut p = RoundtripPerformance::new(100000.0, 0.0, 0.0, DayCountConvention::Raw);
        let rts = make_all_rts();
        for (i, rt) in rts.into_iter().enumerate() {
            p.add_roundtrip(rt);
            assert_eq!(p.returns_on_investments().len(), i + 1);
        }
    }
    #[test] fn test_incremental_roi_values() {
        let expected_rois = vec![
            0.0994,
            0.099375,
            -0.10016666666666667,
            -0.1255,
            0.0996,
            0.08855555555555556,
        ];
        let p = make_perf_with_all();
        for (i, &expected) in expected_rois.iter().enumerate() {
            assert!(almost_equal(p.returns_on_investments()[i], expected, EPSILON),
                "ROI[{}]: expected {}, got {}", i, expected, p.returns_on_investments()[i]);
        }
    }
    #[test] fn test_incremental_sortino_downside_count() {
        let p = make_perf_with_all();
        assert_eq!(p.sortino_downside_returns().len(), 2);
    }
}
