//! NumPy `method="linear"` percentile.

use std::borrow::Borrow;

/// Computes the q-th percentile using NumPy's `method="linear"` definition.
///
/// `window` is any iterable of `f64` or `&f64` (a slice, `&Vec<f64>`,
/// `&VecDeque<f64>`, an array, or an iterator); it is copied and sorted
/// (stable, by `partial_cmp`, like Python's `sorted`).
///
/// # Arguments
///
/// * `window` - Input data.
/// * `q` - Percentile in the range [0, 1].
///
/// # Errors
///
/// * `"q must be between 0 and 1"` if q is outside [0, 1] (or NaN);
///   checked first.
/// * `"window must not be empty"` if the window is empty.
pub fn percentile<I>(window: I, q: f64) -> Result<f64, String>
where
    I: IntoIterator,
    I::Item: Borrow<f64>,
{
    if !(0.0..=1.0).contains(&q) {
        return Err("q must be between 0 and 1".to_string());
    }

    let mut values: Vec<f64> = window.into_iter().map(|x| *x.borrow()).collect();
    if values.is_empty() {
        return Err("window must not be empty".to_string());
    }
    values.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));
    let n = values.len();

    if n == 1 {
        return Ok(values[0]);
    }

    let idx = q * (n - 1) as f64;
    let lo = idx as usize;

    if lo >= n - 1 {
        return Ok(values[n - 1]);
    }

    let hi = lo + 1;
    let frac = idx - lo as f64;

    Ok(values[lo] + frac * (values[hi] - values[lo]))
}

#[cfg(test)]
#[allow(clippy::approx_constant)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::almost_equal;

    fn check_window(window: &[f64], description: &str, expected: &[f64]) {
        let q_values = [
            0.0, 0.01, 0.05, 0.1, 0.2, 0.25, 0.3, 0.4, 0.5, 0.6, 0.7, 0.75, 0.8, 0.9, 0.95, 0.99, 1.0,
        ];

        for (i, (&q, &e)) in q_values.iter().zip(expected.iter()).enumerate() {
            let actual = percentile(window, q).unwrap();
            assert!(
                almost_equal(actual, e, 14),
                "{description} step {i} q {q}: expected {e}, got {actual}"
            );
        }

        // Property test: monotonicity
        let mut last = f64::NEG_INFINITY;
        for q in 0..101 {
            let q = q as f64 * 0.01;
            let current = percentile(window, q).unwrap();
            assert!(
                current >= last,
                "{description} q {q}: monotonicity: last {last}, current {current}"
            );
            last = current;
        }

        // Property test: affine invariance, P_q(aX+b) = aP_q(X)+b, a > 0
        let scaled: Vec<f64> = window.iter().map(|&x| 3.0 * x + 7.0).collect();
        for q in 0..101 {
            let q = q as f64 * 0.01;
            let act = 3.0 * percentile(window, q).unwrap() + 7.0;
            let exp = percentile(&scaled, q).unwrap();
            assert!(
                almost_equal(act, exp, 13),
                "{description} q {q}: affine invariance: expected {exp}, actual {act}"
            );
        }
    }

    #[test]
    fn test_reference_dataset() {
        let win = [
            -5.453279550656607, -3.6648332058049427, 5.947309146654682, 3.525093415019491,
            -2.1778089879618197, -3.34372144267231, 1.9661750717437965, -6.265316287925733,
            3.4551208802924265, 8.836057305398743, -5.03508570740858, 8.977623036666365,
            3.3447490620074483, -8.082041288117757, -1.1632066766437443, 7.729598386550354,
            3.949069997640443, -3.4705427185977573, 4.67856326660133, -5.597300889090276,
            -8.368108609155838, -6.802087978499049, -3.197996300905894, -0.6961369259589816,
            -4.671579434184581, 6.315528068496139, -6.13411221421011, -7.410618476455994,
            -8.166704969101282, 1.971360273298263, 7.094838087480028, 2.0324248338742628,
            8.63976722271967, 4.495627221840401, 7.211026347865847, 8.586756031506326,
            0.9237201816470613, 8.75345917535514, -0.10024119842351453, -4.5245363502002505,
            -0.9644258505047869, 3.3007784679906056,
        ];
        let exp = [
            -8.368108609155838, -8.28553311673347, -8.048470147534669, -6.748410809441717,
            -5.369640782007001, -4.634818663188499, -3.6065460596427874, -1.7719680634345873,
            0.41173949161177337, 2.793437014344066, 3.821877022854157, 4.632829255411098,
            6.241884284127849, 8.501040267010728, 8.747774577723366, 8.91958108684664,
            8.977623036666365,
        ];
        check_window(&win, "reference dataset", &exp);
    }

    #[test]
    fn test_algorithm() {
        let ex = [42.0; 17];
        check_window(&[42.0], "one element", &ex);

        let ex = [
            10.0, 10.1, 10.5, 11.0, 12.0, 12.5, 13.0, 14.0, 15.0, 16.0, 17.0, 17.5, 18.0, 19.0, 19.5,
            19.9, 20.0,
        ];
        check_window(&[10.0, 20.0], "two elements", &ex);

        let ex = [
            1.0, 1.04, 1.2, 1.4, 1.8, 2.0, 2.2, 2.6, 3.0, 3.4, 3.8, 4.0, 4.2, 4.6, 4.8, 4.96, 5.0,
        ];
        check_window(&[1.0, 2.0, 3.0, 4.0, 5.0], "sorted odd number of elements", &ex);
        check_window(&[5.0, 2.0, 1.0, 4.0, 3.0], "unsorted odd number of elements", &ex);
        check_window(&[5.0, 4.0, 3.0, 2.0, 1.0], "reverse sorted odd number of elements", &ex);

        let ex = [
            1.0, 1.03, 1.15, 1.3, 1.6, 1.75, 1.9, 2.2, 2.5, 2.8, 3.1, 3.25, 3.4, 3.7, 3.85, 3.97, 4.0,
        ];
        check_window(&[1.0, 2.0, 3.0, 4.0], "sorted even number of elements", &ex);
        check_window(&[3.0, 2.0, 4.0, 1.0], "unsorted even number of elements", &ex);
        check_window(&[4.0, 3.0, 2.0, 1.0], "reverse sorted even number of elements", &ex);

        let ex = [
            1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.2, 1.6, 2.0, 2.0, 2.0, 2.0, 2.2, 2.6, 2.8, 2.96, 3.0,
        ];
        check_window(&[1.0, 1.0, 2.0, 2.0, 3.0], "duplicate elements", &ex);

        let ex = [
            1.0, 1.04, 1.2, 1.4, 1.8, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.6, 3.80, 4.4, 4.88, 5.0,
        ];
        check_window(&[1.0, 2.0, 2.0, 2.0, 5.0], "more duplicate elements", &ex);

        let ex = [2.0; 17];
        check_window(&[2.0, 2.0, 2.0, 2.0, 2.0], "equal elements", &ex);

        let ex = [
            -10.0, -9.8, -9.0, -8.0, -6.0, -5.0, -4.0, -2.0, 0.0, 2.0, 4.0, 5.0, 6.0, 8.0, 9.0, 9.8,
            10.0,
        ];
        check_window(&[-10.0, -5.0, 0.0, 5.0, 10.0], "negative elements", &ex);

        let ex = [
            -2.71828, -2.6372748, -2.313254, -1.908228, -1.098176, -0.69315, -0.439078, 0.069066,
            0.57721, 0.91201, 1.24681, 1.41421, 1.759686, 2.450638, 2.796114, 3.0724948, 3.14159,
        ];
        check_window(
            &[3.14159, -2.71828, 0.57721, 1.41421, -0.69315],
            "floating-point elements",
            &ex,
        );

        // Q out of range
        let q_err = Err("q must be between 0 and 1".to_string());
        assert_eq!(percentile([42.0], -0.01), q_err);
        assert_eq!(percentile([42.0], 1.01), q_err);

        // Empty window (q is checked first, as in Python)
        let empty: [f64; 0] = [];
        assert_eq!(percentile(empty, 42.0), q_err);
        assert_eq!(percentile(empty, 0.5), Err("window must not be empty".to_string()));
    }
}
