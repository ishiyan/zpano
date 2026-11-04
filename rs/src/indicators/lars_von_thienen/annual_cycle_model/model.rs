use super::numerical;
use crate::indicators::factory::JsonValue;

pub type AnnualParams = numerical::Params;
#[derive(Clone)]
pub struct AnnualModel {
    pub fitted: numerical::Model,
    pub config: AnnualParams,
    pub dates: Option<Vec<String>>,
    pub coverage: Option<String>,
    pub fitted_as_of: Option<String>,
}
#[derive(Clone)]
pub struct PriceRecord {
    pub date: String,
    pub close: f64,
}

/// Strict ISO dates in increasing order, optionally within a single year.
pub fn validate_dates(dates: &[String], year: Option<i64>) -> Result<(), String> {
    let mut previous = "";
    for text in dates {
        let bytes = text.as_bytes();
        if bytes.len() != 10
            || bytes[4] != b'-'
            || bytes[7] != b'-'
            || bytes
                .iter()
                .enumerate()
                .any(|(i, c)| i != 4 && i != 7 && !c.is_ascii_digit())
        {
            return Err("expected ISO calendar dates".into());
        }
        let y = text[..4].parse::<i64>().map_err(|_| "invalid year")?;
        let m = text[5..7].parse::<usize>().map_err(|_| "invalid month")?;
        let d = text[8..].parse::<usize>().map_err(|_| "invalid day")?;
        let leap = y % 4 == 0 && (y % 100 != 0 || y % 400 == 0);
        let days = [
            31,
            if leap { 29 } else { 28 },
            31,
            30,
            31,
            30,
            31,
            31,
            30,
            31,
            30,
            31,
        ];
        if y < 1
            || m < 1
            || m > 12
            || d < 1
            || d > days[m - 1]
            || text.as_str() <= previous
            || year.is_some_and(|v| v != y)
        {
            return Err(
                "dates must be valid, strictly increasing and in the requested year".into(),
            );
        }
        previous = text;
    }
    Ok(())
}
/// Numerical fitting; dated completeness is enforced by decode_annual.
pub fn fit_annual_values(
    prices: &[f64],
    year: i64,
    p: &AnnualParams,
) -> Result<AnnualModel, String> {
    if year < 1
        || p.candidate_pool < 1
        || p.candidate_pool > 16
        || p.max_components < 1
        || p.max_components > 7
        || prices.len() < 6
        || prices.iter().any(|v| !v.is_finite())
    {
        return Err("invalid annual data or bounds".into());
    }
    let (fitted, error) = numerical::fit_annual_values(prices.to_vec(), year, p.clone());
    if let Some(error) = error {
        return Err(error);
    }
    Ok(AnnualModel {
        fitted,
        config: p.clone(),
        dates: None,
        coverage: None,
        fitted_as_of: None,
    })
}
/// Complete past years require an exact supplied calendar or a recorded assertion.
pub fn decode_annual(
    records: &[PriceRecord],
    year: i64,
    as_of: &str,
    calendar: Option<&[String]>,
    confirm: bool,
    p: &AnnualParams,
) -> Result<AnnualModel, String> {
    validate_dates(
        &records.iter().map(|r| r.date.clone()).collect::<Vec<_>>(),
        None,
    )?;
    validate_dates(&[as_of.into()], None)?;
    if as_of < format!("{:04}-01-01", year + 1).as_str() {
        return Err("annual decoding requires a completed year".into());
    }
    let prefix = format!("{year:04}-");
    let selected = records
        .iter()
        .filter(|r| r.date.starts_with(&prefix))
        .collect::<Vec<_>>();
    let dates = selected.iter().map(|r| r.date.clone()).collect::<Vec<_>>();
    if dates.len() < 6 || &dates[0][5..7] != "01" || &dates[dates.len() - 1][5..7] != "12" {
        return Err("annual calendar must span January through December".into());
    }
    if let Some(calendar) = calendar {
        validate_dates(calendar, Some(year))?;
        if calendar != dates {
            return Err("observed dates must match supplied full-year calendar".into());
        }
    } else if !confirm {
        return Err("provide calendar or confirm complete year".into());
    }
    let mut model = fit_annual_values(
        &selected.iter().map(|r| r.close).collect::<Vec<_>>(),
        year,
        p,
    )?;
    model.dates = Some(dates);
    model.fitted_as_of = Some(as_of.into());
    model.coverage = Some(
        if calendar.is_some() {
            "exact match to supplied full-year calendar"
        } else {
            "caller asserted complete year; session completeness unverified"
        }
        .into(),
    );
    Ok(model)
}
pub fn annual_wave(model: &AnnualModel, index: f64) -> f64 {
    numerical::annual_wave(model.fitted.clone(), index)
}
impl AnnualModel {
    pub fn validate(&self) -> Result<(), String> {
        let m = &self.fitted;
        if m.year < 1 || m.sample_count < 6 || m.components.is_empty() || m.components.len() > 7 {
            return Err("invalid annual model".into());
        }
        let mut periods = Vec::new();
        for c in &m.components {
            if !c.period.is_finite()
                || c.period <= 2.0
                || !c.amplitude.is_finite()
                || c.amplitude < 0.0
                || !c.phase.is_finite()
                || periods.contains(&c.period)
            {
                return Err("invalid or duplicate annual component".into());
            }
            periods.push(c.period);
        }
        if let Some(dates) = &self.dates {
            validate_dates(dates, Some(m.year))?;
            if dates.len() != m.sample_count as usize {
                return Err("dates must match sample_count".into());
            }
        }
        if let Some(as_of) = &self.fitted_as_of {
            validate_dates(&[as_of.clone()], None)?;
            if as_of < &format!("{:04}-01-01", m.year + 1) {
                return Err("fit predates year completion".into());
            }
        }
        Ok(())
    }
    /// Decode the version-one portable schema, preserving optional calendar metadata.
    pub fn from_json(v: &JsonValue) -> Result<Self, String> {
        if text(v, "schema")? != "lvt-annual-cycle-model"
            || number(v, "version")? != 1.0
            || text(v, "phase_convention")? != "sine radians at ordinal zero"
        {
            return Err("unsupported annual model schema".into());
        }
        let mut components = Vec::new();
        for c in array(v, "components")? {
            components.push(numerical::Component {
                period: number(c, "period")?,
                amplitude: number(c, "amplitude")?,
                phase: number(c, "phase")?,
                strength: optional_number(c, "strength", 0.0),
                genuine_percent: optional_number(c, "genuine_percent", 0.0),
            });
        }
        let fitted = numerical::Model {
            year: integer(v, "year")?,
            sample_count: integer(v, "sample_count")?,
            components,
            correlation: optional_number(v, "correlation", 0.0),
            subsets_checked: optional_number(v, "subsets_checked", 0.0) as i64,
            candidate_count: optional_number(v, "candidate_count", 0.0) as i64,
        };
        let mut config = numerical::default_params();
        if let Some(c) = v.get("config") {
            config.min_period = integer(c, "min_period")?;
            config.max_period = integer(c, "max_period")?;
            config.smoothing = number(c, "smoothing")?;
            config.genuine_threshold = number(c, "genuine_threshold")?;
            config.candidate_pool = integer(c, "candidate_pool")?;
            config.max_components = integer(c, "max_components")?;
            config.peak_distance = integer(c, "peak_distance")?;
        }
        let model = Self {
            fitted,
            config,
            dates: optional_dates(v)?,
            coverage: v
                .get("coverage")
                .and_then(JsonValue::as_str_val)
                .map(str::to_owned),
            fitted_as_of: v
                .get("fitted_as_of")
                .and_then(JsonValue::as_str_val)
                .map(str::to_owned),
        };
        model.validate()?;
        Ok(model)
    }
    pub fn to_json(&self) -> JsonValue {
        let m = &self.fitted;
        let p = &self.config;
        let mut fields = vec![
            ("schema", string("lvt-annual-cycle-model")),
            ("version", num(1.0)),
            ("research_reconstruction", JsonValue::Bool(true)),
            ("year", num(m.year as f64)),
            ("sample_count", num(m.sample_count as f64)),
            (
                "origin",
                string("first-observed-trading-day, zero-based ordinal"),
            ),
            ("phase_convention", string("sine radians at ordinal zero")),
            ("correlation", num(m.correlation)),
            ("subsets_checked", num(m.subsets_checked as f64)),
            ("candidate_count", num(m.candidate_count as f64)),
            (
                "components",
                JsonValue::Array(
                    m.components
                        .iter()
                        .map(|c| {
                            object(vec![
                                ("period", num(c.period)),
                                ("amplitude", num(c.amplitude)),
                                ("phase", num(c.phase)),
                                ("strength", num(c.strength)),
                                ("genuine_percent", num(c.genuine_percent)),
                            ])
                        })
                        .collect(),
                ),
            ),
            (
                "config",
                object(vec![
                    ("window", num(m.sample_count as f64)),
                    ("min_period", num(p.min_period as f64)),
                    ("max_period", num(p.max_period as f64)),
                    ("smoothing", num(p.smoothing)),
                    ("genuine_threshold", num(p.genuine_threshold)),
                    ("phase_cycles", num(0.0)),
                    ("top_k", num(p.candidate_pool as f64)),
                    ("peak_distance", num(p.peak_distance as f64)),
                    ("max_segments", num(5.0)),
                    ("candidate_pool", num(p.candidate_pool as f64)),
                    ("max_components", num(p.max_components as f64)),
                ]),
            ),
        ];
        if let Some(dates) = &self.dates {
            fields.push((
                "dates",
                JsonValue::Array(dates.iter().map(|d| string(d)).collect()),
            ));
            if let (Some(first), Some(last)) = (dates.first(), dates.last()) {
                fields.push(("start_date", string(first)));
                fields.push(("end_date", string(last)));
            }
        }
        if let Some(v) = &self.coverage {
            fields.push(("coverage", string(v)));
        }
        if let Some(v) = &self.fitted_as_of {
            fields.push(("fitted_as_of", string(v)));
        }
        object(fields)
    }
}
pub fn number(v: &JsonValue, key: &str) -> Result<f64, String> {
    let x = v
        .get(key)
        .and_then(JsonValue::as_f64)
        .ok_or_else(|| format!("missing numeric {key}"))?;
    if !x.is_finite() {
        return Err(format!("nonfinite {key}"));
    }
    Ok(x)
}
pub fn integer(v: &JsonValue, key: &str) -> Result<i64, String> {
    let x = number(v, key)?;
    if x.fract() != 0.0 || x < 1.0 || x > i64::MAX as f64 {
        return Err(format!("invalid integer {key}"));
    }
    Ok(x as i64)
}
pub fn optional_number(v: &JsonValue, key: &str, default: f64) -> f64 {
    v.get(key).and_then(JsonValue::as_f64).unwrap_or(default)
}
pub fn text<'a>(v: &'a JsonValue, key: &str) -> Result<&'a str, String> {
    v.get(key)
        .and_then(JsonValue::as_str_val)
        .ok_or_else(|| format!("missing text {key}"))
}
pub fn array<'a>(v: &'a JsonValue, key: &str) -> Result<&'a [JsonValue], String> {
    v.get(key)
        .and_then(JsonValue::as_array)
        .ok_or_else(|| format!("missing array {key}"))
}
pub fn optional_dates(v: &JsonValue) -> Result<Option<Vec<String>>, String> {
    if v.get("dates").is_none() {
        return Ok(None);
    }
    Ok(Some(
        array(v, "dates")?
            .iter()
            .map(|d| {
                d.as_str_val()
                    .map(str::to_owned)
                    .ok_or("invalid date".into())
            })
            .collect::<Result<_, String>>()?,
    ))
}
pub fn num(v: f64) -> JsonValue {
    if v.is_finite() {
        JsonValue::Number(v)
    } else {
        JsonValue::Null
    }
}
pub fn string(v: &str) -> JsonValue {
    JsonValue::Str(v.into())
}
pub fn object(fields: Vec<(&str, JsonValue)>) -> JsonValue {
    JsonValue::Object(fields.into_iter().map(|(k, v)| (k.into(), v)).collect())
}

pub struct AnnualWaveReader {
    model: AnnualModel,
    index: usize,
}
impl AnnualWaveReader {
    pub fn new(model: AnnualModel) -> Result<Self, String> {
        model.validate()?;
        Ok(Self { model, index: 0 })
    }
    pub fn reset(&mut self) {
        self.index = 0;
    }
    pub fn update(&mut self) -> f64 {
        let v = annual_wave(&self.model, self.index as f64);
        self.index += 1;
        v
    }
}
