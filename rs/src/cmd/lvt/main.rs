//! Local CSV/JSON research workflows using the native Lars von Thienen components.
use std::{
    collections::BTreeMap,
    env, fs,
    time::{SystemTime, UNIX_EPOCH},
};
use zpano::indicators::{
    factory::JsonValue,
    lars_von_thienen::{
        annual_cycle_model::{model::*, numerical::default_params},
        cycle_scanner::scanner::{self, RankedCycle, ScanConfig, ScanOutput},
        dynamic_cycle_explorer::engine::centered_mean,
        harmonic_cycle::model::phase_state,
        harmonic_energy_superposition_wave::model::{
            METONIC_LAGS, SuperpositionModel, superposition_values,
        },
    },
};
const WORKFLOWS: [&str; 7] = [
    "complete-cycle-scanner",
    "dominant-market-vibration",
    "dynamic-cycle-explorer",
    "cycle-phase-score",
    "phasing-oscillator",
    "annual-cycle-pattern-decoder",
    "harmonic-energy-superposition-wave",
];
type Row = BTreeMap<String, JsonValue>;
struct Options(BTreeMap<String, String>);
impl Options {
    fn text(&self, key: &str) -> &str {
        self.0.get(key).map(String::as_str).unwrap_or("")
    }
    fn required(&self, key: &str) -> Result<&str, String> {
        let v = self.text(key);
        if v.is_empty() {
            Err(format!("missing --{key}"))
        } else {
            Ok(v)
        }
    }
    fn number(&self, key: &str) -> Result<f64, String> {
        let v = self
            .required(key)?
            .parse::<f64>()
            .map_err(|_| format!("invalid --{key}"))?;
        if v.is_finite() {
            Ok(v)
        } else {
            Err(format!("nonfinite --{key}"))
        }
    }
    fn integer(&self, key: &str) -> Result<i64, String> {
        let v = self.number(key)?;
        if v.fract() == 0.0 && v >= i64::MIN as f64 && v < i64::MAX as f64 {
            Ok(v as i64)
        } else {
            Err(format!("integer required for --{key}"))
        }
    }
    fn flag(&self, key: &str) -> bool {
        self.text(key) == "true"
    }
}
fn today() -> String {
    let z = (SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs()
        / 86400) as i64
        + 719468;
    let era = z / 146097;
    let doe = z - era * 146097;
    let yoe = (doe - doe / 1460 + doe / 36524 - doe / 146096) / 365;
    let mut y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = mp + if mp < 10 { 3 } else { -9 };
    if m <= 2 {
        y += 1;
    }
    format!("{y:04}-{m:02}-{d:02}")
}
fn options(args: &[String], slug: &str) -> Result<Options, String> {
    let mut o = Options(BTreeMap::new());
    for (k, v) in [
        ("window", "200"),
        ("min-period", "5"),
        (
            "max-period",
            if slug == "annual-cycle-pattern-decoder" {
                "125"
            } else {
                "80"
            },
        ),
        ("smoothing", "1600"),
        ("genuine-threshold", "49"),
        ("phase-cycles", "2"),
        ("top-k", "3"),
        ("peak-distance", "2"),
        ("max-segments", "5"),
        ("candidate-pool", "10"),
        ("max-components", "7"),
        ("horizon", "0"),
        ("short-min", "5"),
        ("short-max", "30"),
        ("long-min", "31"),
        ("long-max", "80"),
        ("normalization", "rms"),
        ("calendar-policy", "extend"),
    ] {
        o.0.insert(k.into(), v.into());
    }
    o.0.insert("as-of".into(), today());
    let flags = [
        "rolling",
        "invert",
        "confirm-complete-year",
        "confirm-complete-years",
    ];
    let extra = [
        "input",
        "calendar",
        "model",
        "models",
        "history",
        "json-out",
        "model-out",
        "wave-out",
        "series-out",
        "start",
        "end",
        "year",
        "target-year",
        "radians",
    ];
    let mut i = 0;
    while i < args.len() {
        let key = args[i]
            .strip_prefix("--")
            .ok_or("expected option beginning --")?;
        if flags.contains(&key) {
            o.0.insert(key.into(), "true".into());
        } else if key == "models" {
            if i + 6 >= args.len() || args[i + 1..i + 7].iter().any(|s| s.starts_with("--")) {
                return Err("--models requires six paths".into());
            }
            o.0.insert(key.into(), args[i + 1..i + 7].join("\0"));
            i += 6;
        } else {
            if !o.0.contains_key(key) && !extra.contains(&key) {
                return Err(format!("unknown option --{key}"));
            }
            i += 1;
            if i >= args.len() || args[i].starts_with("--") {
                return Err(format!("missing --{key} value"));
            }
            o.0.insert(key.into(), args[i].clone());
        }
        i += 1;
    }
    Ok(o)
}
fn read_csv(path: &str) -> Result<Vec<Vec<String>>, String> {
    let input = fs::read_to_string(path).map_err(|e| e.to_string())?;
    let mut rows = Vec::new();
    let mut row = Vec::new();
    let mut field = String::new();
    let mut quoted = false;
    let mut chars = input.chars().peekable();
    while let Some(c) = chars.next() {
        if c == '"' {
            if quoted && chars.peek() == Some(&'"') {
                field.push('"');
                chars.next();
            } else {
                quoted = !quoted;
            }
        } else if !quoted && (c == ',' || c == '\n') {
            row.push(field.trim_end_matches('\r').to_owned());
            field.clear();
            if c == '\n' {
                if row.iter().any(|v| !v.is_empty()) {
                    rows.push(std::mem::take(&mut row));
                } else {
                    row.clear();
                }
            }
        } else {
            field.push(c);
        }
    }
    if quoted {
        return Err("unterminated CSV quote".into());
    }
    if !field.is_empty() || !row.is_empty() {
        row.push(field.trim_end_matches('\r').to_owned());
        rows.push(row);
    }
    if rows.len() < 2 || rows.iter().any(|r| r.len() != rows[0].len()) {
        return Err("CSV requires header and consistent rows".into());
    }
    Ok(rows)
}
fn read_prices(path: &str) -> Result<Vec<PriceRecord>, String> {
    let rows = read_csv(path)?;
    let d = rows[0]
        .iter()
        .position(|s| s == "date")
        .ok_or("CSV requires date")?;
    let c = rows[0]
        .iter()
        .position(|s| s == "close")
        .ok_or("CSV requires close")?;
    let records = rows[1..]
        .iter()
        .map(|r| {
            Ok(PriceRecord {
                date: r[d].clone(),
                close: r[c].parse::<f64>().map_err(|_| "invalid close")?,
            })
        })
        .collect::<Result<Vec<_>, String>>()?;
    validate_dates(
        &records.iter().map(|r| r.date.clone()).collect::<Vec<_>>(),
        None,
    )?;
    if records.iter().any(|r| !r.close.is_finite()) {
        return Err("CSV closes must be finite".into());
    }
    Ok(records)
}
fn read_calendar(path: &str) -> Result<Vec<String>, String> {
    let rows = read_csv(path)?;
    let d = rows[0]
        .iter()
        .position(|s| s == "date")
        .ok_or("calendar requires date")?;
    let dates = rows[1..].iter().map(|r| r[d].clone()).collect::<Vec<_>>();
    validate_dates(&dates, None)?;
    Ok(dates)
}
fn escape(s: &str) -> String {
    let mut v = String::from("\"");
    for c in s.chars() {
        match c {
            '"' => v.push_str("\\\""),
            '\\' => v.push_str("\\\\"),
            '\n' => v.push_str("\\n"),
            '\r' => v.push_str("\\r"),
            '\t' => v.push_str("\\t"),
            c if c < ' ' => v.push_str(&format!("\\u{:04x}", c as u32)),
            _ => v.push(c),
        }
    }
    v.push('"');
    v
}
fn json(v: &JsonValue) -> String {
    match v {
        JsonValue::Null => "null".into(),
        JsonValue::Bool(b) => b.to_string(),
        JsonValue::Number(n) => {
            if n.is_finite() {
                n.to_string()
            } else {
                "null".into()
            }
        }
        JsonValue::Str(s) => escape(s),
        JsonValue::Array(a) => format!("[{}]", a.iter().map(json).collect::<Vec<_>>().join(",")),
        JsonValue::Object(o) => format!(
            "{{{}}}",
            o.iter()
                .map(|(k, v)| format!("{}:{}", escape(k), json(v)))
                .collect::<Vec<_>>()
                .join(",")
        ),
    }
}
fn save_json(path: &str, v: &JsonValue) -> Result<(), String> {
    fs::write(path, json(v) + "\n").map_err(|e| e.to_string())
}
fn load_json(path: &str) -> Result<JsonValue, String> {
    JsonValue::parse(&fs::read_to_string(path).map_err(|e| e.to_string())?)
}
fn csv(v: &JsonValue) -> String {
    let s = match v {
        JsonValue::Null => String::new(),
        JsonValue::Str(s) => s.clone(),
        _ => json(v),
    };
    if s.contains([',', '"', '\n', '\r']) {
        format!("\"{}\"", s.replace('"', "\"\""))
    } else {
        s
    }
}
fn save_wave(path: &str, dates: &[String], values: &[f64]) -> Result<(), String> {
    if dates.len() != values.len() {
        return Err("calendar and wave lengths differ".into());
    }
    let mut s = String::from("date,value\n");
    for (d, v) in dates.iter().zip(values) {
        s.push_str(&format!("{},{}\n", csv(&string(d)), csv(&num(*v))));
    }
    fs::write(path, s).map_err(|e| e.to_string())
}
fn save_series(path: &str, rows: &[Row]) -> Result<(), String> {
    let first = rows.first().ok_or("empty series")?;
    let mut keys = vec!["date".to_owned()];
    keys.extend(first.keys().filter(|k| *k != "date").cloned());
    let mut s = keys.join(",") + "\n";
    for row in rows {
        s.push_str(
            &keys
                .iter()
                .map(|k| csv(row.get(k).unwrap_or(&JsonValue::Null)))
                .collect::<Vec<_>>()
                .join(","),
        );
        s.push('\n');
    }
    fs::write(path, s).map_err(|e| e.to_string())
}
fn row(fields: Vec<(&str, JsonValue)>) -> Row {
    fields.into_iter().map(|(k, v)| (k.to_owned(), v)).collect()
}
fn row_json(r: Row) -> JsonValue {
    JsonValue::Object(r.into_iter().collect())
}
fn cycle(c: &RankedCycle) -> Row {
    row(vec![
        ("period", num(c.period)),
        ("amplitude", num(c.amplitude)),
        ("phase", num(c.phase)),
        ("strength", num(c.strength)),
        ("genuine_percent", num(c.genuine_percent)),
        ("segments", num(c.segments)),
        ("fit_amplitude", num(c.fit_amplitude)),
        ("fit_samples", num(c.fit_samples)),
        ("value", num(c.value)),
        ("slope", num(c.slope)),
        ("eta_top", num(c.eta_top)),
        ("eta_bottom", num(c.eta_bottom)),
    ])
}
fn config(c: &ScanConfig) -> JsonValue {
    object(vec![
        ("window", num(c.window as f64)),
        ("min_period", num(c.min_period as f64)),
        ("max_period", num(c.max_period as f64)),
        ("smoothing", num(c.smoothing)),
        ("genuine_threshold", num(c.genuine_threshold)),
        ("phase_cycles", num(c.phase_cycles)),
        ("top_k", num(c.top_k as f64)),
        ("peak_distance", num(c.peak_distance as f64)),
        ("max_segments", num(c.max_segments as f64)),
    ])
}
fn report(s: &ScanOutput, c: &ScanConfig, dates: &[String]) -> Row {
    row(vec![
        ("schema", string("lvt-cycle-scan")),
        ("version", num(1.0)),
        ("research_reconstruction", JsonValue::Bool(true)),
        ("start_date", string(&dates[0])),
        ("end_date", string(dates.last().unwrap())),
        ("sample_count", num(dates.len() as f64)),
        ("config", config(c)),
        (
            "phase_convention",
            string("sine radians at analysis endpoint"),
        ),
        (
            "validation_input",
            string("raw prices, most recent complete segments"),
        ),
        ("accepted_count", num(s.accepted_count)),
        (
            "cycles",
            JsonValue::Array(s.cycles.iter().map(|c| row_json(cycle(c))).collect()),
        ),
        (
            "spectrum",
            object(vec![
                (
                    "periods",
                    JsonValue::Array(
                        (c.min_period..=c.max_period)
                            .map(|p| num(p as f64))
                            .collect(),
                    ),
                ),
                (
                    "amplitude",
                    JsonValue::Array(s.amplitude.iter().map(|v| num(*v)).collect()),
                ),
                (
                    "genuine_percent",
                    JsonValue::Array(s.genuine_percent.iter().map(|v| num(*v)).collect()),
                ),
            ]),
        ),
    ])
}
fn analyze(values: &[f64], c: &ScanConfig) -> ScanOutput {
    if values.len() < c.window as usize {
        scanner::empty_scan(c.clone())
    } else {
        scanner::scan_cycles(
            values[values.len() - c.window as usize..].to_vec(),
            c.clone(),
        )
    }
}
fn phase_status(p: f64) -> &'static str {
    if p.is_nan() {
        return "unavailable";
    }
    let a = p.rem_euclid(2.0 * std::f64::consts::PI);
    if a < std::f64::consts::FRAC_PI_2 {
        "rising toward top"
    } else if a < std::f64::consts::PI {
        "falling from top"
    } else if a < 3.0 * std::f64::consts::FRAC_PI_2 {
        "falling toward bottom"
    } else {
        "rising from bottom"
    }
}
fn calculate(
    values: &[f64],
    slug: &str,
    c: &ScanConfig,
    short: &ScanConfig,
    long: &ScanConfig,
) -> Row {
    let s = analyze(values, c);
    let v = &s.cycles[0];
    match slug {
        "complete-cycle-scanner" => {
            let mut r = row(vec![("accepted_count", num(s.accepted_count))]);
            for (i, a) in s.amplitude.iter().enumerate() {
                r.insert(format!("amplitude_p{}", c.min_period + i as i64), num(*a));
                r.insert(
                    format!("genuine_p{}", c.min_period + i as i64),
                    num(s.genuine_percent[i]),
                );
            }
            for (i, v) in s.cycles.iter().enumerate() {
                for (k, v) in cycle(v) {
                    r.insert(format!("rank{}_{}", i + 1, k), v);
                }
            }
            r
        }
        "dominant-market-vibration" => cycle(v),
        "cycle-phase-score" => row(vec![
            ("value", num(phase_state(v.phase))),
            ("period", num(v.period)),
            ("phase", num(v.phase)),
            ("status", string(phase_status(v.phase))),
        ]),
        "phasing-oscillator" => {
            let a = analyze(values, short);
            let b = analyze(values, long);
            let x = phase_state(a.cycles[0].phase);
            let y = phase_state(b.cycles[0].phase);
            row(vec![
                ("value", num(x + y)),
                ("short_score", num(x)),
                ("long_score", num(y)),
                ("short_period", num(a.cycles[0].period)),
                ("long_period", num(b.cycles[0].period)),
            ])
        }
        _ => {
            let (mut h, mut index, mut fitted) = (f64::NAN, f64::NAN, f64::NAN);
            if v.period.is_finite() {
                let length = v.period as i64;
                let delay = length / 2;
                let buffer = values[values.len() - c.window as usize..].to_vec();
                let center = buffer.len() as i64 - 1 - delay;
                h = centered_mean(
                    buffer.clone(),
                    center,
                    2.max((v.period / 2.0 + 0.5).floor() as i64),
                ) - centered_mean(buffer, center, length);
                if h.is_finite() {
                    index = (values.len() as i64 - 1 - delay) as f64;
                    fitted = scanner::modeled_value(v.clone(), -delay as f64);
                }
            }
            row(vec![
                ("value", num(v.value)),
                ("period", num(v.period)),
                ("phase", num(v.phase)),
                ("eta_top", num(v.eta_top)),
                ("eta_bottom", num(v.eta_bottom)),
                ("highlighter", num(h)),
                ("highlighter_index", num(index)),
                ("fitted_at_highlighter", num(fitted)),
            ])
        }
    }
}
fn run(args: Vec<String>) -> Result<(), String> {
    if args.len() < 2 || args.iter().any(|a| a == "--help") {
        println!(
            "usage: lvt <workflow> <analyze|fit|build|replay|phase> [options]\n{}",
            WORKFLOWS.join("\n")
        );
        return Ok(());
    }
    let slug = &args[0];
    let command = &args[1];
    if !WORKFLOWS.contains(&slug.as_str()) {
        return Err("unknown workflow".into());
    }
    let o = options(&args[2..], slug)?;
    if slug == "cycle-phase-score" && command == "phase" {
        let p = o.number("radians")?;
        println!(
            "{}",
            json(&object(vec![
                ("value", num(phase_state(p))),
                ("status", string(phase_status(p)))
            ]))
        );
        return Ok(());
    }
    if slug == "annual-cycle-pattern-decoder" {
        let m = if command == "fit" {
            let mut p = default_params();
            p.min_period = o.integer("min-period")?;
            p.max_period = o.integer("max-period")?;
            p.smoothing = o.number("smoothing")?;
            p.genuine_threshold = o.number("genuine-threshold")?;
            p.candidate_pool = o.integer("candidate-pool")?;
            p.max_components = o.integer("max-components")?;
            p.peak_distance = o.integer("peak-distance")?;
            let calendar = if o.text("calendar").is_empty() {
                None
            } else {
                Some(read_calendar(o.text("calendar"))?)
            };
            let m = decode_annual(
                &read_prices(o.required("input")?)?,
                o.integer("year")?,
                o.text("as-of"),
                calendar.as_deref(),
                o.flag("confirm-complete-year"),
                &p,
            )?;
            o.required("wave-out")?;
            save_json(o.required("model-out")?, &m.to_json())?;
            m
        } else if command == "replay" {
            AnnualModel::from_json(&load_json(o.required("model")?)?)?
        } else {
            return Err("annual workflow supports fit or replay".into());
        };
        let dates = m.dates.as_ref().ok_or("replay requires model dates")?;
        let values = (0..dates.len())
            .map(|i| annual_wave(&m, i as f64))
            .collect::<Vec<_>>();
        save_wave(o.required("wave-out")?, dates, &values)?;
        println!(
            "{}",
            json(&object(vec![
                ("year", num(m.fitted.year as f64)),
                ("correlation", num(m.fitted.correlation)),
                ("components", num(m.fitted.components.len() as f64)),
                ("sample_count", num(values.len() as f64))
            ]))
        );
        return Ok(());
    }
    if slug == "harmonic-energy-superposition-wave" {
        let (values, m) = if command == "build" {
            let dates = read_calendar(o.required("calendar")?)?;
            let year = o.integer("target-year")?;
            validate_dates(&dates, Some(year))?;
            if o.text("models").is_empty() == o.text("history").is_empty() {
                return Err("provide models or history".into());
            }
            let models = if !o.text("models").is_empty() {
                o.text("models")
                    .split('\0')
                    .map(|p| AnnualModel::from_json(&load_json(p)?))
                    .collect::<Result<Vec<_>, String>>()?
            } else {
                let history = read_prices(o.text("history"))?;
                METONIC_LAGS
                    .iter()
                    .map(|lag| {
                        decode_annual(
                            &history,
                            year - lag,
                            o.text("as-of"),
                            None,
                            o.flag("confirm-complete-years"),
                            &default_params(),
                        )
                    })
                    .collect::<Result<Vec<_>, String>>()?
            };
            let (v, mut m) = superposition_values(
                &models,
                year,
                dates.len(),
                o.text("normalization"),
                o.text("calendar-policy"),
                o.flag("invert"),
            )?;
            m.dates = Some(dates);
            o.required("wave-out")?;
            save_json(o.required("model-out")?, &m.to_json())?;
            (v, m)
        } else if command == "replay" {
            let m = SuperpositionModel::from_json(&load_json(o.required("model")?)?)?;
            let v = superposition_values(
                &m.annual_models,
                m.target_year,
                m.sample_count,
                &m.normalization,
                &m.calendar_policy,
                m.invert,
            )?
            .0;
            (v, m)
        } else {
            return Err("superposition workflow supports build or replay".into());
        };
        save_wave(
            o.required("wave-out")?,
            m.dates.as_ref().ok_or("replay requires dates")?,
            &values,
        )?;
        println!(
            "{}",
            json(&object(vec![
                ("target_year", num(m.target_year as f64)),
                (
                    "source_years",
                    JsonValue::Array(
                        m.annual_models
                            .iter()
                            .map(|m| num(m.fitted.year as f64))
                            .collect()
                    )
                ),
                ("sample_count", num(values.len() as f64))
            ]))
        );
        return Ok(());
    }
    if command != "analyze" {
        return Err("streaming workflows support analyze".into());
    }
    for key in ["start", "end"] {
        if !o.text(key).is_empty() {
            validate_dates(&[o.text(key).into()], None)?;
        }
    }
    if !o.text("start").is_empty() && !o.text("end").is_empty() && o.text("start") > o.text("end") {
        return Err("start must not exceed end".into());
    }
    let rows = read_prices(o.required("input")?)?
        .into_iter()
        .filter(|r| {
            (o.text("start").is_empty() || r.date.as_str() >= o.text("start"))
                && (o.text("end").is_empty() || r.date.as_str() <= o.text("end"))
        })
        .collect::<Vec<_>>();
    if rows.len() < 6 {
        return Err("selected range needs >=6 prices".into());
    }
    let horizon = o.integer("horizon")?;
    if horizon < 0 {
        return Err("horizon must be nonnegative".into());
    }
    let prices = rows.iter().map(|r| r.close).collect::<Vec<_>>();
    let dates = rows.iter().map(|r| r.date.clone()).collect::<Vec<_>>();
    let c = ScanConfig {
        window: if o.flag("rolling") {
            o.integer("window")?
        } else {
            prices.len() as i64
        },
        min_period: o.integer("min-period")?,
        max_period: o.integer("max-period")?,
        smoothing: o.number("smoothing")?,
        genuine_threshold: o.number("genuine-threshold")?,
        phase_cycles: o.number("phase-cycles")?,
        top_k: o.integer("top-k")?,
        peak_distance: o.integer("peak-distance")?,
        max_segments: o.integer("max-segments")?,
    };
    if let Some(e) = c.validate() {
        return Err(e);
    }
    let short = ScanConfig {
        min_period: o.integer("short-min")?,
        max_period: o.integer("short-max")?,
        top_k: 1,
        ..c.clone()
    };
    let long = ScanConfig {
        min_period: o.integer("long-min")?,
        max_period: o.integer("long-max")?,
        top_k: 1,
        ..c.clone()
    };
    if slug == "phasing-oscillator" {
        if short.max_period >= long.min_period {
            return Err("phasing ranges must be disjoint".into());
        }
        for c in [&short, &long] {
            if let Some(e) = c.validate() {
                return Err(e);
            }
        }
    }
    let scan = analyze(&prices, &c);
    let mut r = report(
        &scan,
        &c,
        &dates[dates.len().saturating_sub(c.window as usize)..],
    );
    r.insert(
        "output".into(),
        row_json(calculate(&prices, slug, &c, &short, &long)),
    );
    if !o.text("series-out").is_empty() {
        let mut series = Vec::new();
        for (i, d) in dates.iter().enumerate() {
            let mut v = if o.flag("rolling") || slug == "phasing-oscillator" {
                calculate(&prices[..i + 1], slug, &c, &short, &long)
            } else {
                let mut v = row(vec![(
                    "fitted",
                    num(scanner::modeled_value(
                        scan.cycles[0].clone(),
                        i as f64 - (prices.len() - 1) as f64,
                    )),
                )]);
                if slug == "dynamic-cycle-explorer" {
                    let p = scan.cycles[0].period;
                    v.insert(
                        "highlighter".into(),
                        num(if p.is_finite() {
                            centered_mean(
                                prices.clone(),
                                i as i64,
                                2.max((p / 2.0 + 0.5).floor() as i64),
                            ) - centered_mean(prices.clone(), i as i64, p as i64)
                        } else {
                            f64::NAN
                        }),
                    );
                }
                v
            };
            v.insert("date".into(), string(d));
            series.push(v);
        }
        save_series(o.text("series-out"), &series)?;
    }
    let side = if slug == "phasing-oscillator" {
        let a = analyze(&prices, &short);
        let b = analyze(&prices, &long);
        r.insert(
            "short_scan".into(),
            row_json(report(
                &a,
                &short,
                &dates[dates.len().saturating_sub(short.window as usize)..],
            )),
        );
        r.insert(
            "long_scan".into(),
            row_json(report(
                &b,
                &long,
                &dates[dates.len().saturating_sub(long.window as usize)..],
            )),
        );
        r.insert(
            "phasing_config".into(),
            object(vec![("short", config(&short)), ("long", config(&long))]),
        );
        Some((a.cycles[0].clone(), b.cycles[0].clone()))
    } else {
        None
    };
    let projection = (1..=horizon)
        .map(|h| {
            let mut v = row(vec![("bars_ahead", num(h as f64))]);
            let c = &scan.cycles[0];
            if let Some((a, b)) = &side {
                let x = phase_state(a.phase + 2.0 * std::f64::consts::PI * h as f64 / a.period);
                let y = phase_state(b.phase + 2.0 * std::f64::consts::PI * h as f64 / b.period);
                v.insert("value".into(), num(x + y));
                v.insert("short_score".into(), num(x));
                v.insert("long_score".into(), num(y));
            } else {
                v.insert(
                    "value".into(),
                    num(if slug == "cycle-phase-score" {
                        phase_state(c.phase + 2.0 * std::f64::consts::PI * h as f64 / c.period)
                    } else {
                        scanner::modeled_value(c.clone(), h as f64)
                    }),
                );
            }
            row_json(v)
        })
        .collect();
    r.insert("projection".into(), JsonValue::Array(projection));
    let r = row_json(r);
    if !o.text("json-out").is_empty() {
        save_json(o.text("json-out"), &r)?;
    }
    println!("{}", json(&r));
    Ok(())
}
fn main() {
    if let Err(e) = run(env::args().skip(1).collect()) {
        eprintln!("{e}");
        std::process::exit(1);
    }
}
