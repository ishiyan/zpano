use super::range_entity::RangeEntity;

/// Criterion defines how to measure a candlestick feature relative to recent history.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Criterion {
    pub entity: RangeEntity,
    pub average_period: usize,
    pub factor: f64,
}

impl Criterion {
    /// Creates a new Criterion.
    pub const fn new(entity: RangeEntity, average_period: usize, factor: f64) -> Self {
        Self { entity, average_period, factor }
    }

    /// Computes the threshold value from a running total of candle contributions.
    /// When average_period > 0, divides total by the period (doubled for Shadows).
    /// When average_period == 0, uses the current candle's range value directly.
    pub fn average_value_from_total(&self, total: f64, o: f64, h: f64, l: f64, c: f64) -> f64 {
        if self.average_period > 0 {
            if self.entity == RangeEntity::Shadows {
                return self.factor * total / (self.average_period as f64 * 2.0);
            }
            return self.factor * total / self.average_period as f64;
        }
        self.factor * candle_range_value(self.entity, o, h, l, c)
    }

    /// Returns the contribution of a single candle to the running total,
    /// based on the criterion's entity type.
    pub fn candle_contribution(&self, o: f64, h: f64, l: f64, c: f64) -> f64 {
        match self.entity {
            RangeEntity::RealBody => {
                if c >= o { c - o } else { o - c }
            }
            RangeEntity::HighLow => h - l,
            RangeEntity::Shadows => {
                if c >= o { h - c + o - l } else { h - o + c - l }
            }
        }
    }
}

/// Computes the range value of a candle for a given RangeEntity type.
pub fn candle_range_value(entity: RangeEntity, o: f64, h: f64, l: f64, c: f64) -> f64 {
    match entity {
        RangeEntity::RealBody => {
            if c >= o { c - o } else { o - c }
        }
        RangeEntity::HighLow => h - l,
        RangeEntity::Shadows => {
            // Average of upper and lower shadow.
            if c >= o {
                (h - c + o - l) / 2.0
            } else {
                (h - o + c - l) / 2.0
            }
        }
    }
}
