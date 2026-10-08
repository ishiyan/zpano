//! Expected values used by the performance tests, generated from
//! `py/performance/reference_data`.
#![allow(dead_code)]

pub mod active_premium;
pub mod adjusted_sharpe_ratio;
pub mod appraisal_ratio;
pub mod bernardo_ledoit_ratio;
pub mod bias_ratio;
pub mod burke_ratio;
pub mod calmar_ratio;
pub mod cdar;
pub mod cdar_alpha;
pub mod cdar_beta;
pub mod cumulative_geometric_return;
pub mod d_ratio;
pub mod down_number_ratio;
pub mod down_percentage_ratio;
pub mod downside_capture_ratio;
pub mod downside_deviation;
pub mod downside_frequency;
pub mod downside_potential;
pub mod downside_sharpe_ratio;
pub mod drawdown_average;
pub mod drawdown_average_length;
pub mod drawdown_average_peak_to_trough;
pub mod drawdown_average_recovery;
pub mod drawdown_deviation;
pub mod drawdowns_cumulative;
pub mod drawdowns_high_watermark;
pub mod es;
pub mod fama_beta;
pub mod gain_to_pain_ratio;
pub mod geometric_mean_return;
pub mod hurst_exponent;
pub mod information_ratio;
pub mod jarque_bera_normality_test_statistic;
pub mod jensen_alpha;
pub mod jensen_alpha_alternative;
pub mod jensen_alpha_modified;
pub mod k_ratio;
pub mod kappa_ratio;
pub mod kelly_ratio;
pub mod kurtosis;
pub mod loss_rate;
pub mod m_squared;
pub mod m_squared_excess;
pub mod m_squared_sortino;
pub mod martin_ratio;
pub mod mean_loss_return;
pub mod mean_non_zero_return;
pub mod mean_win_return;
pub mod min_drawdowns_cumulative;
pub mod modigliani;
pub mod omega_excess_return;
pub mod omega_ratio;
pub mod omega_sharpe_ratio;
pub mod overall_capture_ratio;
pub mod pain_index;
pub mod pain_ratio;
pub mod probabilistic_sharpe_ratio;
pub mod prospect_ratio;
pub mod rachev_ratio;
pub mod semi_deviation;
pub mod sfm_alpha;
pub mod sfm_beta;
pub mod sfm_beta_bear;
pub mod sfm_beta_bull;
pub mod sfm_r2;
pub mod sfm_risk_premium;
pub mod sharpe_ratio;
pub mod skewness;
pub mod skewness_kurtosis_ratio;
pub mod sortino_ratio;
pub mod specific_risk;
pub mod sterling_ratio;
pub mod systematic_risk;
pub mod tail_ratio;
pub mod timing_ratio;
pub mod total_risk;
pub mod tracking_error;
pub mod treynor_ratio;
pub mod treynor_ratio_modified;
pub mod ulcer_index;
pub mod up_number_ratio;
pub mod up_percentage_ratio;
pub mod upside_capture_ratio;
pub mod upside_frequency;
pub mod upside_potential_ratio;
pub mod upside_risk;
pub mod var;
pub mod volatility_skewness;
pub mod win_rate;

/// Returns the value stored under `key` in a generated key/value table.
///
/// Panics if the key is absent.
pub fn lookup<K: PartialEq + Copy + std::fmt::Debug, V: Copy>(pairs: &[(K, V)], key: K) -> V {
    pairs
        .iter()
        .find(|(k, _)| *k == key)
        .map(|(_, v)| *v)
        .unwrap_or_else(|| panic!("reference data key {:?} not found", key))
}
