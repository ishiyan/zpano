/// Minimal JSON parser for flat objects with number/bool/string/null values.
/// Supports arrays and nested objects for cmd tool settings parsing.
/// Zero external dependencies.

#[derive(Debug, Clone, PartialEq)]
pub enum JsonValue {
    Null,
    Bool(bool),
    Number(f64),
    Str(String),
    Array(Vec<JsonValue>),
    Object(Vec<(String, JsonValue)>),
}

impl JsonValue {
    pub fn parse(input: &str) -> Result<JsonValue, String> {
        let mut chars = input.trim().chars().peekable();
        let val = parse_value(&mut chars)?;
        // skip trailing whitespace
        skip_ws(&mut chars);
        Ok(val)
    }

    pub fn get(&self, key: &str) -> Option<&JsonValue> {
        match self {
            JsonValue::Object(pairs) => pairs.iter().find(|(k, _)| k == key).map(|(_, v)| v),
            _ => None,
        }
    }

    pub fn as_f64(&self) -> Option<f64> {
        match self {
            JsonValue::Number(n) => Some(*n),
            _ => None,
        }
    }

    pub fn as_str_val(&self) -> Option<&str> {
        match self {
            JsonValue::Str(s) => Some(s),
            _ => None,
        }
    }

    pub fn as_bool(&self) -> Option<bool> {
        match self {
            JsonValue::Bool(b) => Some(*b),
            _ => None,
        }
    }

    pub fn as_array(&self) -> Option<&[JsonValue]> {
        match self {
            JsonValue::Array(a) => Some(a),
            _ => None,
        }
    }

    pub fn as_object(&self) -> Option<&[(String, JsonValue)]> {
        match self {
            JsonValue::Object(o) => Some(o),
            _ => None,
        }
    }
}

type Chars<'a> = std::iter::Peekable<std::str::Chars<'a>>;

fn skip_ws(chars: &mut Chars) {
    while let Some(&c) = chars.peek() {
        if c.is_ascii_whitespace() {
            chars.next();
        } else {
            break;
        }
    }
}

fn parse_value(chars: &mut Chars) -> Result<JsonValue, String> {
    skip_ws(chars);
    match chars.peek() {
        None => Err("unexpected end of input".into()),
        Some(&'"') => parse_string(chars).map(JsonValue::Str),
        Some(&'{') => parse_object(chars),
        Some(&'[') => parse_array(chars),
        Some(&'t') => parse_literal(chars, "true", JsonValue::Bool(true)),
        Some(&'f') => parse_literal(chars, "false", JsonValue::Bool(false)),
        Some(&'n') => parse_literal(chars, "null", JsonValue::Null),
        Some(&c) if c == '-' || c.is_ascii_digit() => parse_number(chars),
        Some(&c) => Err(format!("unexpected character: '{}'", c)),
    }
}

fn parse_literal(chars: &mut Chars, expected: &str, value: JsonValue) -> Result<JsonValue, String> {
    for ec in expected.chars() {
        match chars.next() {
            Some(c) if c == ec => {}
            _ => return Err(format!("expected '{}'", expected)),
        }
    }
    Ok(value)
}

fn parse_number(chars: &mut Chars) -> Result<JsonValue, String> {
    let mut s = String::new();
    if let Some(&'-') = chars.peek() {
        s.push('-');
        chars.next();
    }
    while let Some(&c) = chars.peek() {
        if c.is_ascii_digit() || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-' {
            // avoid consuming '-' or '+' unless after 'e'/'E'
            if (c == '-' || c == '+') && !s.ends_with('e') && !s.ends_with('E') {
                break;
            }
            s.push(c);
            chars.next();
        } else {
            break;
        }
    }
    s.parse::<f64>()
        .map(JsonValue::Number)
        .map_err(|e| format!("invalid number '{}': {}", s, e))
}

fn parse_string(chars: &mut Chars) -> Result<String, String> {
    match chars.next() {
        Some('"') => {}
        _ => return Err("expected '\"'".into()),
    }
    let mut s = String::new();
    loop {
        match chars.next() {
            None => return Err("unterminated string".into()),
            Some('"') => return Ok(s),
            Some('\\') => match chars.next() {
                Some('"') => s.push('"'),
                Some('\\') => s.push('\\'),
                Some('/') => s.push('/'),
                Some('n') => s.push('\n'),
                Some('r') => s.push('\r'),
                Some('t') => s.push('\t'),
                Some('b') => s.push('\u{0008}'),
                Some('f') => s.push('\u{000C}'),
                Some('u') => {
                    let mut hex = String::with_capacity(4);
                    for _ in 0..4 {
                        match chars.next() {
                            Some(c) => hex.push(c),
                            None => return Err("unterminated unicode escape".into()),
                        }
                    }
                    let code = u32::from_str_radix(&hex, 16)
                        .map_err(|_| format!("invalid unicode escape: \\u{}", hex))?;
                    let ch = char::from_u32(code)
                        .ok_or_else(|| format!("invalid unicode code point: {}", code))?;
                    s.push(ch);
                }
                Some(c) => return Err(format!("invalid escape: \\{}", c)),
                None => return Err("unterminated escape".into()),
            },
            Some(c) => s.push(c),
        }
    }
}

fn parse_object(chars: &mut Chars) -> Result<JsonValue, String> {
    chars.next(); // consume '{'
    skip_ws(chars);
    let mut pairs = Vec::new();
    if let Some(&'}') = chars.peek() {
        chars.next();
        return Ok(JsonValue::Object(pairs));
    }
    loop {
        skip_ws(chars);
        let key = parse_string(chars)?;
        skip_ws(chars);
        match chars.next() {
            Some(':') => {}
            _ => return Err("expected ':'".into()),
        }
        let val = parse_value(chars)?;
        pairs.push((key, val));
        skip_ws(chars);
        match chars.peek() {
            Some(&',') => {
                chars.next();
            }
            Some(&'}') => {
                chars.next();
                return Ok(JsonValue::Object(pairs));
            }
            _ => return Err("expected ',' or '}'".into()),
        }
    }
}

fn parse_array(chars: &mut Chars) -> Result<JsonValue, String> {
    chars.next(); // consume '['
    skip_ws(chars);
    let mut items = Vec::new();
    if let Some(&']') = chars.peek() {
        chars.next();
        return Ok(JsonValue::Array(items));
    }
    loop {
        let val = parse_value(chars)?;
        items.push(val);
        skip_ws(chars);
        match chars.peek() {
            Some(&',') => {
                chars.next();
            }
            Some(&']') => {
                chars.next();
                return Ok(JsonValue::Array(items));
            }
            _ => return Err("expected ',' or ']'".into()),
        }
    }
}

// ── Helper functions for the factory ─────────────────────────────────────────

pub fn has_key(v: &JsonValue, key: &str) -> bool {
    v.get(key).is_some()
}

pub fn is_empty_object(v: &JsonValue) -> bool {
    matches!(v, JsonValue::Object(pairs) if pairs.is_empty())
}

pub fn get_f64(v: &JsonValue, key: &str) -> Option<f64> {
    v.get(key).and_then(|v| v.as_f64())
}

pub fn get_usize(v: &JsonValue, key: &str) -> Option<usize> {
    v.get(key).and_then(|v| v.as_f64()).map(|n| n as usize)
}

pub fn get_i32(v: &JsonValue, key: &str) -> Option<i32> {
    v.get(key).and_then(|v| v.as_f64()).map(|n| n as i32)
}

pub fn get_i64(v: &JsonValue, key: &str) -> Option<i64> {
    v.get(key).and_then(|v| v.as_f64()).map(|n| n as i64)
}

pub fn get_bool(v: &JsonValue, key: &str) -> Option<bool> {
    v.get(key).and_then(|v| v.as_bool())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_empty_object() {
        let v = JsonValue::parse("{}").unwrap();
        assert!(is_empty_object(&v));
    }

    #[test]
    fn test_parse_simple_object() {
        let v = JsonValue::parse(r#"{"length": 14, "active": true}"#).unwrap();
        assert_eq!(get_usize(&v, "length"), Some(14));
        assert_eq!(get_bool(&v, "active"), Some(true));
    }

    #[test]
    fn test_parse_float() {
        let v = JsonValue::parse(r#"{"factor": 0.991}"#).unwrap();
        assert!((get_f64(&v, "factor").unwrap() - 0.991).abs() < 1e-15);
    }

    #[test]
    fn test_parse_negative() {
        let v = JsonValue::parse(r#"{"val": -3.5}"#).unwrap();
        assert!((get_f64(&v, "val").unwrap() - (-3.5)).abs() < 1e-15);
    }

    #[test]
    fn test_parse_array() {
        let v = JsonValue::parse(r#"[1, 2, 3]"#).unwrap();
        let arr = v.as_array().unwrap();
        assert_eq!(arr.len(), 3);
    }

    #[test]
    fn test_parse_nested() {
        let v = JsonValue::parse(r#"{"a": {"b": 1}}"#).unwrap();
        let inner = v.get("a").unwrap();
        assert_eq!(get_f64(inner, "b"), Some(1.0));
    }

    #[test]
    fn test_parse_string_value() {
        let v = JsonValue::parse(r#"{"name": "hello"}"#).unwrap();
        assert_eq!(v.get("name").unwrap().as_str_val(), Some("hello"));
    }

    #[test]
    fn test_parse_null() {
        let v = JsonValue::parse(r#"{"x": null}"#).unwrap();
        assert_eq!(v.get("x"), Some(&JsonValue::Null));
    }

    #[test]
    fn test_has_key() {
        let v = JsonValue::parse(r#"{"smoothingFactor": 0.5}"#).unwrap();
        assert!(has_key(&v, "smoothingFactor"));
        assert!(!has_key(&v, "length"));
    }

    #[test]
    fn test_settings_format() {
        let input = r#"[
            {"identifier": "simpleMovingAverage", "params": {"length": 14}},
            {"identifier": "trueRange", "params": {}}
        ]"#;
        let v = JsonValue::parse(input).unwrap();
        let arr = v.as_array().unwrap();
        assert_eq!(arr.len(), 2);
        assert_eq!(arr[0].get("identifier").unwrap().as_str_val(), Some("simpleMovingAverage"));
        let params = arr[0].get("params").unwrap();
        assert_eq!(get_usize(params, "length"), Some(14));
        assert!(is_empty_object(arr[1].get("params").unwrap()));
    }
}
