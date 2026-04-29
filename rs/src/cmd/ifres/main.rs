/// ifres — command-line indicator frequency response calculator.
///
/// Reads a JSON settings file containing indicator definitions,
/// creates indicator instances, determines each indicator's warmup period,
/// and calculates the frequency response with signal length 1024.
///
/// Usage: ifres <settings.json>

use portf::entities::scalar::Scalar;
use portf::indicators::core::frequency_response::{self, Component, FrequencyResponse, Updater};
use portf::indicators::core::identifier::Identifier;
use portf::indicators::core::indicator::Indicator;
use portf::indicators::core::metadata::Metadata;
use portf::indicators::factory::{create_indicator, JsonValue};

use std::env;
use std::fs;
use std::process;

const SIGNAL_LENGTH: usize = 1024;
const MAX_WARMUP: usize = 10000;
const PHASE_DEGREES_UNWRAPPING_LIMIT: f64 = 179.0;

/// Adapts a boxed Indicator to the frequency_response Updater trait.
struct IndicatorUpdater {
    ind: Box<dyn Indicator>,
}

impl Updater for IndicatorUpdater {
    fn metadata(&self) -> Metadata {
        self.ind.metadata()
    }

    fn update(&mut self, sample: f64) -> f64 {
        let s = Scalar::new(0, sample);
        let output = self.ind.update_scalar(&s);
        // Extract the first scalar value from output.
        if let Some(first) = output.first() {
            if let Some(scalar) = first.downcast_ref::<Scalar>() {
                return scalar.value;
            }
        }
        f64::NAN
    }
}

fn detect_warmup(updater: &mut IndicatorUpdater) -> usize {
    for i in 0..MAX_WARMUP {
        if updater.ind.is_primed() {
            return i;
        }
        updater.update(0.0);
    }
    MAX_WARMUP
}

fn print_component(name: &str, c: &Component) {
    print!("  {:<25} min={:10.4}  max={:10.4}", name, c.min, c.max);

    let n = c.data.len();
    if n == 0 {
        println!();
        return;
    }

    let preview = 3;
    if n <= preview * 2 {
        print!("  data={:?}", c.data);
    } else {
        print!(
            "  data=[{:.4} {:.4} {:.4} ... {:.4} {:.4} {:.4}]",
            c.data[0], c.data[1], c.data[2],
            c.data[n - 3], c.data[n - 2], c.data[n - 1]
        );
    }

    println!();
}

fn print_frequency_response(fr: &FrequencyResponse, warmup: usize) {
    println!("=== {} (warmup={}) ===", fr.label, warmup);
    println!("  Spectrum length: {}", fr.normalized_frequency.len());

    print_component("PowerPercent", &fr.power_percent);
    print_component("PowerDecibel", &fr.power_decibel);
    print_component("AmplitudePercent", &fr.amplitude_percent);
    print_component("AmplitudeDecibel", &fr.amplitude_decibel);
    print_component("PhaseDegrees", &fr.phase_degrees);
    print_component("PhaseDegreesUnwrapped", &fr.phase_degrees_unwrapped);

    println!();
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() < 2 {
        eprintln!("usage: ifres <settings.json>");
        process::exit(1);
    }

    let data = fs::read_to_string(&args[1]).unwrap_or_else(|e| {
        eprintln!("error reading settings file: {}", e);
        process::exit(1);
    });

    let parsed = JsonValue::parse(&data).unwrap_or_else(|e| {
        eprintln!("error parsing settings file: {}", e);
        process::exit(1);
    });

    let entries = parsed.as_array().unwrap_or_else(|| {
        eprintln!("error: settings file must be a JSON array");
        process::exit(1);
    });

    for entry in entries {
        let id_str = entry.get("identifier")
            .and_then(|v| v.as_str_val())
            .unwrap_or_else(|| {
                eprintln!("error: entry missing 'identifier' string");
                process::exit(1);
            });

        let ident = Identifier::from_str(id_str).unwrap_or_else(|| {
            eprintln!("error: unknown indicator identifier: {}", id_str);
            process::exit(1);
        });

        let params_json = match entry.get("params") {
            Some(p) => json_value_to_string(p),
            None => "{}".to_string(),
        };

        // Create a probe instance to determine warmup period.
        let probe = create_indicator(ident, &params_json).unwrap_or_else(|e| {
            eprintln!("error creating indicator {}: {}", id_str, e);
            process::exit(1);
        });

        let mut probe_updater = IndicatorUpdater { ind: probe };
        let warmup = detect_warmup(&mut probe_updater);

        // Create a fresh instance for actual calculation.
        let ind = create_indicator(ident, &params_json).unwrap_or_else(|e| {
            eprintln!("error creating indicator {}: {}", id_str, e);
            process::exit(1);
        });

        let mut updater = IndicatorUpdater { ind };

        let fr = frequency_response::calculate(
            SIGNAL_LENGTH,
            &mut updater,
            warmup,
            PHASE_DEGREES_UNWRAPPING_LIMIT,
        ).unwrap_or_else(|e| {
            eprintln!("error calculating frequency response for {}: {}", id_str, e);
            process::exit(1);
        });

        print_frequency_response(&fr, warmup);
    }
}

// ---------------------------------------------------------------------------
// JSON value to string (for passing params to factory)
// ---------------------------------------------------------------------------

fn json_value_to_string(val: &JsonValue) -> String {
    match val {
        JsonValue::Null => "null".to_string(),
        JsonValue::Bool(b) => b.to_string(),
        JsonValue::Number(n) => {
            if *n == (*n as i64) as f64 && n.abs() < 1e15 {
                format!("{}", *n as i64)
            } else {
                format!("{}", n)
            }
        }
        JsonValue::Str(s) => format!("{:?}", s),
        JsonValue::Array(arr) => {
            let items: Vec<String> = arr.iter().map(json_value_to_string).collect();
            format!("[{}]", items.join(","))
        }
        JsonValue::Object(pairs) => {
            let items: Vec<String> = pairs.iter()
                .map(|(k, v)| format!("{:?}:{}", k, json_value_to_string(v)))
                .collect();
            format!("{{{}}}", items.join(","))
        }
    }
}
