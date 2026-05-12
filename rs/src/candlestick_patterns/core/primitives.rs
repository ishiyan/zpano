// ---------------------------------------------------------------------------
// Color
// ---------------------------------------------------------------------------

/// Returns true when a candlestick is white (bullish): close >= open.
pub fn is_white(o: f64, c: f64) -> bool {
    c >= o
}

/// Returns true when a candlestick is black (bearish): close < open.
pub fn is_black(o: f64, c: f64) -> bool {
    c < o
}

// ---------------------------------------------------------------------------
// Real body
// ---------------------------------------------------------------------------

/// Returns the absolute length of the real body.
pub fn real_body_len(o: f64, c: f64) -> f64 {
    if c >= o { c - o } else { o - c }
}

/// Returns the length of the real body of a white candlestick (close - open).
pub fn white_real_body(o: f64, c: f64) -> f64 {
    c - o
}

/// Returns the length of the real body of a black candlestick (open - close).
pub fn black_real_body(o: f64, c: f64) -> f64 {
    o - c
}

// ---------------------------------------------------------------------------
// Shadows
// ---------------------------------------------------------------------------

/// Returns the length of the upper shadow.
pub fn upper_shadow(o: f64, h: f64, c: f64) -> f64 {
    if c >= o { h - c } else { h - o }
}

/// Returns the length of the lower shadow.
pub fn lower_shadow(o: f64, l: f64, c: f64) -> f64 {
    if c >= o { o - l } else { c - l }
}

/// Returns the length of the upper shadow of a white candlestick.
pub fn white_upper_shadow(h: f64, c: f64) -> f64 {
    h - c
}

/// Returns the length of the upper shadow of a black candlestick.
pub fn black_upper_shadow(o: f64, h: f64) -> f64 {
    h - o
}

/// Returns the length of the lower shadow of a white candlestick.
pub fn white_lower_shadow(o: f64, l: f64) -> f64 {
    o - l
}

/// Returns the length of the lower shadow of a black candlestick.
pub fn black_lower_shadow(l: f64, c: f64) -> f64 {
    c - l
}

// ---------------------------------------------------------------------------
// Gap tests
// ---------------------------------------------------------------------------

/// Returns true when max(open1, close1) < min(open2, close2).
pub fn is_real_body_gap_up(o1: f64, c1: f64, o2: f64, c2: f64) -> bool {
    f64::max(o1, c1) < f64::min(o2, c2)
}

/// Returns true when min(open1, close1) > max(open2, close2).
pub fn is_real_body_gap_down(o1: f64, c1: f64, o2: f64, c2: f64) -> bool {
    f64::min(o1, c1) > f64::max(o2, c2)
}

/// Returns true when high of first candle < low of second candle.
pub fn is_high_low_gap_up(h1: f64, l2: f64) -> bool {
    h1 < l2
}

/// Returns true when low of first candle > high of second candle.
pub fn is_high_low_gap_down(l1: f64, h2: f64) -> bool {
    l1 > h2
}

// ---------------------------------------------------------------------------
// Enclosure tests
// ---------------------------------------------------------------------------

/// Returns true when the real body of candle 1 completely encloses the real body of candle 2.
pub fn is_real_body_encloses_real_body(o1: f64, c1: f64, o2: f64, c2: f64) -> bool {
    let (min1, max1) = if c1 > o1 { (o1, c1) } else { (c1, o1) };
    let (min2, max2) = if c2 > o2 { (o2, c2) } else { (c2, o2) };
    max1 > max2 && min1 < min2
}

/// Returns true when the real body of candle 1 encloses the open of candle 2.
pub fn is_real_body_encloses_open(o1: f64, c1: f64, o2: f64) -> bool {
    if o1 > c1 {
        o2 < o1 && o2 > c1
    } else {
        o2 > o1 && o2 < c1
    }
}

/// Returns true when the real body of candle 1 encloses the close of candle 2.
pub fn is_real_body_encloses_close(o1: f64, c1: f64, c2: f64) -> bool {
    if o1 > c1 {
        c2 < o1 && c2 > c1
    } else {
        c2 > o1 && c2 < c1
    }
}

// ---------------------------------------------------------------------------
// Misc comparisons
// ---------------------------------------------------------------------------

/// Returns true when high of candle 1 > close of candle 2.
pub fn is_high_exceeds_close(h1: f64, c2: f64) -> bool {
    h1 > c2
}

/// Returns true when candle 1 opens within the real body of candle 2
/// (with optional tolerance).
pub fn is_opens_within(o1: f64, o2: f64, c2: f64, tolerance: f64) -> bool {
    o1 >= f64::min(o2, c2) - tolerance && o1 <= f64::max(o2, c2) + tolerance
}


