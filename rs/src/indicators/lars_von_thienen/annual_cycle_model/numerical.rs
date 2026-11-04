use super::super::cycle_scanner::scanner as cyclescanner;
use super::super::cycle_scanner::scanner::RankedCycle;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Params {
    pub min_period: i64,
    pub max_period: i64,
    pub smoothing: f64,
    pub genuine_threshold: f64,
    pub candidate_pool: i64,
    pub max_components: i64,
    pub peak_distance: i64,
}
impl Params {}
pub fn default_params() -> Params {
    return Params {
        min_period: 5,
        max_period: 125,
        smoothing: 1600.0,
        genuine_threshold: 49.0,
        candidate_pool: 10,
        max_components: 7,
        peak_distance: 2,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Component {
    pub period: f64,
    pub amplitude: f64,
    pub phase: f64,
    pub strength: f64,
    pub genuine_percent: f64,
}
impl Component {}
#[derive(Clone, Default)]
pub struct Model {
    pub year: i64,
    pub sample_count: i64,
    pub components: Vec<Component>,
    pub correlation: f64,
    pub subsets_checked: i64,
    pub candidate_count: i64,
}
impl Model {}
pub fn pearson(left: Vec<f64>, right: Vec<f64>) -> f64 {
    if ((left.len() as i64) < 1) || ((left.len() as i64) != (right.len() as i64)) {
        panic!("{:?}", "correlation requires nonempty equal-length vectors");
    }
    let (a, b) = (
        (harmoniccycle::sum(left.clone()) / ((left.len() as i64) as f64)),
        (harmoniccycle::sum(right.clone()) / ((right.len() as i64) as f64)),
    );
    let (mut xx, mut yy, mut xy) = (
        vec![Default::default(); (left.len() as i64) as usize].clone(),
        vec![Default::default(); (left.len() as i64) as usize].clone(),
        vec![Default::default(); (left.len() as i64) as usize].clone(),
    );
    for i in 0..(left.len() as i64) {
        let (x, y) = ((left[(i) as usize] - a), (right[(i) as usize] - b));
        xx[(i) as usize] = x * x;
        yy[(i) as usize] = y * y;
        xy[(i) as usize] = x * y;
    }
    let (x, y) = (
        harmoniccycle::sum(xx.clone()),
        harmoniccycle::sum(yy.clone()),
    );
    if (x == 0.0) || (y == 0.0) {
        return f64::NAN;
    }
    return (harmoniccycle::sum(xy.clone()) / (x * y).sqrt()).clamp(-1.0, 1.0);
}
pub fn choose_subset(
    target: Vec<f64>,
    waves: Vec<Vec<f64>>,
    periods: Vec<f64>,
    max_components: i64,
) -> (Vec<i64>, f64, i64, Option<String>) {
    if ((((((waves.len() as i64) < 1) || ((waves.len() as i64) > 16))
        || ((waves.len() as i64) != (periods.len() as i64)))
        || ((target.len() as i64) < 6))
        || (max_components < 1))
        || (max_components > 7)
    {
        return (
            Vec::new(),
            0.0,
            0,
            Some("invalid annual subset inputs".into()),
        );
    }
    let n: i64 = target.len() as i64;
    let mean: f64 = harmoniccycle::sum(target.clone()) / (n as f64);
    let mut tc: Vec<f64> = vec![Default::default(); (n) as usize].clone();
    let mut energy_values: Vec<f64> = vec![Default::default(); (n) as usize].clone();
    for (__index, value) in target.clone().into_iter().enumerate() {
        let i = __index as i64;
        tc[(i) as usize] = value - mean;
        energy_values[(i) as usize] = tc[(i) as usize] * tc[(i) as usize];
    }
    let energy: f64 = harmoniccycle::sum(energy_values.clone());
    if energy == 0.0 {
        return (
            Vec::new(),
            0.0,
            0,
            Some("invalid annual subset inputs".into()),
        );
    }
    let mut centered: Vec<Vec<f64>> =
        vec![Default::default(); (waves.len() as i64) as usize].clone();
    for (__index, wave) in waves.clone().into_iter().enumerate() {
        let i = __index as i64;
        if (wave.len() as i64) != n {
            return (
                Vec::new(),
                0.0,
                0,
                Some("invalid annual subset inputs".into()),
            );
        }
        let mean: f64 = harmoniccycle::sum(wave.clone()) / (n as f64);
        centered[(i) as usize] = vec![Default::default(); (n) as usize].clone();
        for (__index, value) in wave.clone().into_iter().enumerate() {
            let j = __index as i64;
            centered[(i) as usize][(j) as usize] = value - mean;
        }
    }
    let mut cov: Vec<f64> = vec![Default::default(); (waves.len() as i64) as usize].clone();
    let mut gram: Vec<Vec<f64>> = vec![Default::default(); (waves.len() as i64) as usize].clone();
    for (__index, left) in centered.clone().into_iter().enumerate() {
        let i = __index as i64;
        let mut products: Vec<f64> = vec![Default::default(); (n) as usize].clone();
        for k in 0..n {
            products[(k) as usize] = tc[(k) as usize] * left[(k) as usize];
        }
        cov[(i) as usize] = harmoniccycle::sum(products.clone());
        gram[(i) as usize] = vec![Default::default(); (waves.len() as i64) as usize].clone();
        for (__index, right) in centered.clone().into_iter().enumerate() {
            let j = __index as i64;
            for k in 0..n {
                products[(k) as usize] = left[(k) as usize] * right[(k) as usize];
            }
            gram[(i) as usize][(j) as usize] = harmoniccycle::sum(products.clone());
        }
    }
    let mut best: Vec<i64> = vec![].clone();
    let mut best_score: f64 = f64::NEG_INFINITY;
    let mut best_periods: Vec<f64> = vec![].clone();
    let mut checked: i64 = 0;
    {
        let mut count: i64 = 1;
        while count <= (max_components).min(waves.len() as i64) {
            {
                let mut mask: i64 = 1;
                while mask < (1 << (waves.len() as i64)) {
                    let mut subset: Vec<i64> = vec![].clone();
                    {
                        let mut i: i64 = 0;
                        while i < (waves.len() as i64) {
                            if (mask & (1 << i)) != 0 {
                                subset = {
                                    let mut __values = subset.clone();
                                    __values.push(i);
                                    __values
                                }
                                .clone();
                            }
                            i += 1;
                        }
                    }
                    if (subset.len() as i64) != count {
                        mask += 1;
                        continue;
                    }
                    checked += 1;
                    let mut numerators: Vec<f64> =
                        vec![Default::default(); (count) as usize].clone();
                    let mut denominators: Vec<f64> = vec![].clone();
                    let mut key: Vec<f64> = vec![Default::default(); (count) as usize].clone();
                    for (__index, i) in subset.clone().into_iter().enumerate() {
                        let k = __index as i64;
                        numerators[(k) as usize] = cov[(i) as usize];
                        key[(k) as usize] = periods[(i) as usize];
                        for j in subset.clone() {
                            denominators = {
                                let mut __values = denominators.clone();
                                __values.push(gram[(i) as usize][(j) as usize]);
                                __values
                            }
                            .clone();
                        }
                    }
                    let denominator: f64 = harmoniccycle::sum(denominators.clone());
                    if denominator <= 0.0 {
                        mask += 1;
                        continue;
                    }
                    let score = (harmoniccycle::sum(numerators.clone())
                        / (energy * denominator).sqrt())
                    .clamp(-1.0, 1.0);
                    key.sort_by(|a, b| a.total_cmp(b));
                    let mut smaller = (best.len() as i64) > count;
                    if (best.len() as i64) == count {
                        for i in 0..(key.len() as i64) {
                            if key[(i) as usize] != best_periods[(i) as usize] {
                                smaller = key[(i) as usize] < best_periods[(i) as usize];
                                break;
                            }
                        }
                    }
                    if ((score as f64) > (best_score + 1e-12))
                        || ((((score as f64) - best_score).abs() <= 1e-12) && smaller)
                    {
                        best = {
                            let mut __values = vec![].clone();
                            __values.extend(subset.clone());
                            __values
                        }
                        .clone();
                        best_score = score as f64;
                        best_periods = key.clone();
                    }
                    mask += 1;
                }
            }
            count += 1;
        }
    }
    if ((best.len() as i64) < 1) || (best_score <= 0.0) {
        return (
            Vec::new(),
            0.0,
            0,
            Some("invalid annual subset inputs".into()),
        );
    }
    return (best.clone(), best_score, checked, None);
}
pub fn modulo(mut value: f64) -> f64 {
    value = value % (2.0 * std::f64::consts::PI);
    if value < 0.0 {
        value += 2.0 * std::f64::consts::PI;
    }
    return value;
}
pub fn fit_annual_values(
    prices: Vec<f64>,
    year: i64,
    p: Params,
) -> (Model, Option<String>) {
    if ((((year < 1) || (p.candidate_pool < 1)) || (p.candidate_pool > 16))
        || (p.max_components < 1))
        || (p.max_components > 7)
    {
        return (
            Model {
                ..Default::default()
            }
            .clone(),
            Some("invalid annual bounds".to_string()),
        );
    }
    let config = cyclescanner::ScanConfig {
        window: (prices.len() as i64),
        min_period: p.min_period,
        max_period: p.max_period,
        smoothing: p.smoothing,
        genuine_threshold: p.genuine_threshold,
        phase_cycles: 0.0,
        top_k: p.candidate_pool,
        peak_distance: p.peak_distance,
        max_segments: 5,
    }
    .clone();
    {
        let err = config.validate();
        if err.is_some() {
            return (
                Model {
                    ..Default::default()
                }
                .clone(),
                err,
            );
        }
    }
    let scan = cyclescanner::scan_cycles(prices.clone(), config.clone()).clone();
    let mut cycles: Vec<RankedCycle> = Vec::<cyclescanner::RankedCycle>::new().clone();
    for cycle in scan.cycles.clone() {
        if !(cycle.period).is_nan() {
            cycles = {
                let mut __values = cycles.clone();
                __values.push(cycle.clone());
                __values
            }
            .clone();
        }
    }
    if (cycles.len() as i64) < 1 {
        return (
            Model {
                ..Default::default()
            }
            .clone(),
            Some("no accepted annual cycles".to_string()),
        );
    }
    let n: i64 = prices.len() as i64;
    let mut waves: Vec<Vec<f64>> = vec![Default::default(); (cycles.len() as i64) as usize].clone();
    let mut periods: Vec<f64> = vec![Default::default(); (cycles.len() as i64) as usize].clone();
    for (__index, cycle) in cycles.clone().into_iter().enumerate() {
        let i = __index as i64;
        periods[(i) as usize] = cycle.period;
        waves[(i) as usize] = vec![Default::default(); (n) as usize].clone();
        for j in 0..n {
            waves[(i) as usize][(j) as usize] =
                cyclescanner::modeled_value(cycle.clone(), (j - (n - 1)) as f64);
        }
    }
    let (subset, correlation, checked, err) = choose_subset(
        hodrickprescottsmoothinganddetrending::cycle_data(prices.clone(), p.smoothing).clone(),
        waves.clone(),
        periods.clone(),
        p.max_components,
    );
    if err.is_some() {
        return (
            Model {
                ..Default::default()
            }
            .clone(),
            err,
        );
    }
    let mut components: Vec<Component> =
        vec![Default::default(); (subset.len() as i64) as usize].clone();
    for (__index, index) in subset.into_iter().enumerate() {
        let i = __index as i64;
        let cycle = cycles[(index) as usize].clone();
        components[(i) as usize] = Component {
            period: cycle.period,
            amplitude: cycle.fit_amplitude,
            phase: modulo(
                cycle.phase - (((2.0 * std::f64::consts::PI) * ((n - 1) as f64)) / cycle.period),
            ),
            strength: cycle.strength,
            genuine_percent: cycle.genuine_percent,
        }
        .clone();
    }
    {
        let mut i: i64 = 1;
        while i < (components.len() as i64) {
            {
                let mut j: i64 = i;
                while (j > 0)
                    && (components[(j) as usize].period < components[(j - 1) as usize].period)
                {
                    (components[(j) as usize], components[(j - 1) as usize]) = (
                        components[(j - 1) as usize].clone(),
                        components[(j) as usize].clone(),
                    );
                    j -= 1;
                }
            }
            i += 1;
        }
    }
    return (
        Model {
            year: year,
            sample_count: n,
            components: components.clone(),
            correlation: correlation,
            subsets_checked: checked,
            candidate_count: (cycles.len() as i64),
        }
        .clone(),
        None,
    );
}
pub fn annual_wave(model: Model, index: f64) -> f64 {
    let mut values: Vec<f64> =
        vec![Default::default(); (model.components.len() as i64) as usize].clone();
    for (__index, c) in model.components.clone().into_iter().enumerate() {
        let i = __index as i64;
        values[(i) as usize] =
            c.amplitude * ((((2.0 * std::f64::consts::PI) * index) / c.period) + c.phase).sin();
    }
    return harmoniccycle::sum(values.clone());
}
