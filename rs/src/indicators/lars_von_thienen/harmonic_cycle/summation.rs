/// Non-overlapping partials preserve cancellation precision.
pub fn sum(values: Vec<f64>) -> f64 {
    let mut partials: Vec<f64> = Vec::new();
    for mut value in values {
        let mut i = 0;
        for mut y in partials.clone() {
            if value.abs() < y.abs() {
                std::mem::swap(&mut value, &mut y);
            }
            let high = value + y;
            let low = y - (high - value);
            if low != 0.0 {
                partials[i] = low;
                i += 1;
            }
            value = high;
        }
        partials.truncate(i);
        partials.push(value);
    }
    partials.into_iter().rev().sum()
}
