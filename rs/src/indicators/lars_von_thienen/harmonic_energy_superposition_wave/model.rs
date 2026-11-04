use super::super::annual_cycle_model::model::*;
use super::super::harmonic_cycle::summation::sum;
use crate::indicators::factory::JsonValue;

pub const METONIC_LAGS: [i64; 6] = [8, 11, 19, 27, 30, 38];
#[derive(Clone)]
pub struct Scale {
    pub year: i64,
    pub mean: f64,
    pub scale: f64,
    pub sample_count: i64,
}
#[derive(Clone)]
pub struct SuperpositionModel {
    pub target_year: i64,
    pub sample_count: usize,
    pub normalization: String,
    pub calendar_policy: String,
    pub invert: bool,
    pub scales: Vec<Scale>,
    pub annual_models: Vec<AnnualModel>,
    pub dates: Option<Vec<String>>,
}
/// Ordinal alignment and source-year-only normalization; no target prices are read.
pub fn superposition_values(
    models: &[AnnualModel],
    target_year: i64,
    count: usize,
    normalization: &str,
    calendar_policy: &str,
    invert: bool,
) -> Result<(Vec<f64>, SuperpositionModel), String> {
    if target_year < 39
        || count < 1
        || !["none", "rms"].contains(&normalization)
        || !["strict", "extend"].contains(&calendar_policy)
        || models.len() != 6
    {
        return Err("invalid superposition configuration".into());
    }
    let mut years = Vec::new();
    for m in models {
        m.validate()?;
        if years.contains(&m.fitted.year) {
            return Err("duplicate source year".into());
        }
        years.push(m.fitted.year);
    }
    let mut annual_models = Vec::new();
    let mut scales = Vec::new();
    for lag in METONIC_LAGS {
        let year = target_year - lag;
        let m = models
            .iter()
            .find(|m| m.fitted.year == year)
            .ok_or("require target years minus 8,11,19,27,30,38")?;
        let n = m.fitted.sample_count;
        if calendar_policy == "strict" && count != n as usize {
            return Err("strict ordinal alignment requires equal sample counts".into());
        }
        let base = (0..n).map(|i| annual_wave(m, i as f64)).collect::<Vec<_>>();
        let mean = if normalization == "rms" {
            sum(base.clone()) / n as f64
        } else {
            0.0
        };
        let scale = if normalization == "rms" {
            (sum(base.iter().map(|v| (v - mean) * (v - mean)).collect()) / n as f64).sqrt()
        } else {
            1.0
        };
        if scale == 0.0 {
            return Err("zero-energy source wave".into());
        }
        scales.push(Scale {
            year,
            mean,
            scale,
            sample_count: n,
        });
        annual_models.push(m.clone());
    }
    let values = (0..count)
        .map(|i| {
            let terms = annual_models
                .iter()
                .zip(&scales)
                .map(|(m, s)| (annual_wave(m, i as f64) - s.mean) / s.scale)
                .collect();
            (if invert { -1.0 } else { 1.0 }) * sum(terms) / 6.0
        })
        .collect();
    Ok((
        values,
        SuperpositionModel {
            target_year,
            sample_count: count,
            normalization: normalization.into(),
            calendar_policy: calendar_policy.into(),
            invert,
            scales,
            annual_models,
            dates: None,
        },
    ))
}
impl SuperpositionModel {
    pub fn from_json(v: &JsonValue) -> Result<Self, String> {
        if text(v, "schema")? != "lvt-superposition-model" || number(v, "version")? != 1.0 {
            return Err("unsupported superposition schema".into());
        }
        let models = array(v, "annual_models")?
            .iter()
            .map(AnnualModel::from_json)
            .collect::<Result<Vec<_>, _>>()?;
        let (_, mut m) = superposition_values(
            &models,
            integer(v, "target_year")?,
            integer(v, "sample_count")? as usize,
            text(v, "normalization")?,
            text(v, "calendar_policy")?,
            v.get("invert")
                .and_then(JsonValue::as_bool)
                .ok_or("invalid inversion flag")?,
        )?;
        for (key, expected) in [
            ("lags", METONIC_LAGS.to_vec()),
            (
                "source_years",
                METONIC_LAGS.iter().map(|lag| m.target_year - lag).collect(),
            ),
        ] {
            let actual = array(v, key)?;
            if actual.len() != 6
                || actual
                    .iter()
                    .zip(expected)
                    .any(|(v, n)| v.as_f64() != Some(n as f64))
            {
                return Err("superposition source years and lags disagree".into());
            }
        }
        m.dates = optional_dates(v)?;
        if let Some(d) = &m.dates {
            validate_dates(d, Some(m.target_year))?;
            if d.len() != m.sample_count {
                return Err("dates must match sample_count".into());
            }
        }
        Ok(m)
    }
    pub fn to_json(&self) -> JsonValue {
        let mut fields = vec![
            ("schema", string("lvt-superposition-model")),
            ("version", num(1.0)),
            ("research_reconstruction", JsonValue::Bool(true)),
            ("target_year", num(self.target_year as f64)),
            ("sample_count", num(self.sample_count as f64)),
            (
                "source_years",
                JsonValue::Array(
                    self.annual_models
                        .iter()
                        .map(|m| num(m.fitted.year as f64))
                        .collect(),
                ),
            ),
            (
                "lags",
                JsonValue::Array(METONIC_LAGS.iter().map(|&v| num(v as f64)).collect()),
            ),
            ("normalization", string(&self.normalization)),
            ("calendar_policy", string(&self.calendar_policy)),
            ("invert", JsonValue::Bool(self.invert)),
            (
                "annual_models",
                JsonValue::Array(
                    self.annual_models
                        .iter()
                        .map(AnnualModel::to_json)
                        .collect(),
                ),
            ),
            (
                "scales",
                JsonValue::Array(
                    self.scales
                        .iter()
                        .map(|s| {
                            object(vec![
                                ("year", num(s.year as f64)),
                                ("mean", num(s.mean)),
                                ("scale", num(s.scale)),
                                ("sample_count", num(s.sample_count as f64)),
                            ])
                        })
                        .collect(),
                ),
            ),
        ];
        if let Some(d) = &self.dates {
            fields.push((
                "dates",
                JsonValue::Array(d.iter().map(|v| string(v)).collect()),
            ));
        }
        object(fields)
    }
}
pub struct SuperpositionWaveReader {
    values: Vec<f64>,
    index: usize,
}
impl SuperpositionWaveReader {
    pub fn new(
        models: &[AnnualModel],
        year: i64,
        count: usize,
        norm: &str,
        policy: &str,
        invert: bool,
    ) -> Result<Self, String> {
        Ok(Self {
            values: superposition_values(models, year, count, norm, policy, invert)?.0,
            index: 0,
        })
    }
    pub fn reset(&mut self) {
        self.index = 0;
    }
    pub fn update(&mut self) -> Option<f64> {
        let v = self.values.get(self.index).copied()?;
        self.index += 1;
        Some(v)
    }
}
