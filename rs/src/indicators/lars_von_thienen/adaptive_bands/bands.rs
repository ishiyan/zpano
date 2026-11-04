// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

pub fn grid_bands(values: Vec<f64>, leveling: f64) -> (f64, f64) {
    let (mut low, mut high) = (values[(0) as usize], values[(0) as usize]);
    for value in values.clone() {
        low = (low).min(value);
        high = (high).max(value);
    }
    let step: f64 = (high - low) / 100.0;
    let fraction: f64 = leveling / 100.0;
    let (mut lower, mut upper) = (low, high);
    for k in 0..101 {
        let test: f64 = low + (step * (k as f64));
        let mut below: i64 = 0;
        for value in values.clone() {
            if value < test {
                below += 1;
            }
        }
        if ((below as f64) / ((values.len() as i64) as f64)) >= fraction {
            lower = test;
            break;
        }
    }
    for k in 0..101 {
        let test: f64 = high - (step * (k as f64));
        let mut above: i64 = 0;
        for value in values.clone() {
            if value >= test {
                above += 1;
            }
        }
        if ((above as f64) / ((values.len() as i64) as f64)) >= fraction {
            upper = test;
            break;
        }
    }
    return (lower, upper);
}
