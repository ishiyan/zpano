use super::*;

#[test]
fn gain_to_pain_repetition_and_rolling() {
    for window in [0, 2, 4] {
        let mut m = Measures::new(1.0, 0.0, 0.0, window).unwrap();
        assert!(m.gain_to_pain_ratio().is_nan());
        for _ in 0..4 {
            m.add_return(0.1, 0.0);
            m.add_return(-0.05, 0.0);
            assert_float(
                m.gain_to_pain_ratio(),
                1.0,
                Places(13),
                "sum/sum independent of repetition",
            );
        }
        m.reset();
        m.add_return(0.1, 0.0);
        assert!(m.gain_to_pain_ratio().is_nan());
    }
}

#[test]
fn modified_information_geometric_sign() {
    let mut m = Measures::new(1.0, 0.0, 0.0, 0).unwrap();
    m.add_return(0.5, 0.05);
    m.add_return(-0.3, 0.05);
    assert!(m.active_premium() < 0.0);
    assert_float(
        m.information_ratio_modified(),
        0.04473320734100739,
        Places(13),
        "opposite arithmetic/geometric signs",
    );
}

#[test]
fn validation_before_data_fallback() {
    let mut m = Measures::new(1.0, 0.0, 0.0, 0).unwrap();
    for populated in [false, true] {
        if populated {
            m.add_return(-0.1, 0.0);
            m.add_return(0.2, 0.0);
        }
        for c in [
            -1e20,
            -1.0,
            0.0,
            1.0,
            2.0,
            f64::NAN,
            f64::INFINITY,
            f64::NEG_INFINITY,
        ] {
            let error = Err("confidence must be between 0 and 1".to_string());
            assert_eq!(m.reward_to_conditional_drawdown(c), error);
            assert_eq!(
                m.is_normal_distribution(c),
                Err("confidence must be between 0 and 1".to_string())
            );
        }
        for k in [0.0, -1.0, f64::NAN, f64::NEG_INFINITY] {
            assert_eq!(
                m.bias_ratio(k),
                Err("std_dev_multiplier must be positive".to_string())
            );
        }
    }
    assert!(Measures::new(1.0, 0.0, 0.0, 0)
        .unwrap()
        .reward_to_conditional_drawdown(0.95)
        .unwrap()
        .is_nan());
}

#[test]
fn long_tail_sums_match_compensated_python_reference() {
    let mut m = Measures::new(1.0, 0.0, 0.0, 0).unwrap();
    for _ in 0..10000 {
        m.add_return(0.1, 0.0);
        m.add_return(-0.001, 0.0);
    }
    // Each tail is constant, so its exact average gives 0.1 / 0.001 = 100.
    assert_float(
        m.rachev_ratio(0.1, 0.1).unwrap(),
        100.0,
        Places(13),
        "constant-tail Rachev",
    );
    // Python 3.14 result for this stream, including its log-equity rounding.
    assert_float(
        m.reward_to_conditional_drawdown(0.95).unwrap(),
        48.28431257909963,
        Places(13),
        "long conditional-drawdown tail",
    );
}
