"""Translate existing Python reference fixtures into Go/TS/Rust/Zig source.

The expected values originate in the R scripts documented in
performance_analytics_reference_data_generation.md, with additional manual or
formula-based fixtures for measures without R reference output. This script
does not run R or calculate expected values.

Usage: python3 notes/performance/gen.py [output-root]
The default output root is the repository containing this script. Pass a
temporary output root to inspect regenerated files before replacing fixtures.
"""
import importlib
import math
import os
from pathlib import Path
import pkgutil
import re
import sys

ROOT = str(Path(__file__).resolve().parents[2])
sys.path.insert(0, ROOT)
import py.performance.reference_data as rd  # noqa: E402

HEADER = "Code generated from py/performance/reference_data/{mod}.py. DO NOT EDIT."


def modules():
    for m in sorted(pkgutil.iter_modules(rd.__path__), key=lambda m: m.name):
        mod = importlib.import_module('py.performance.reference_data.' + m.name)
        items = [(k, v) for k, v in vars(mod).items()
                 if not k.startswith('_') and isinstance(v, (list, dict))]
        yield m.name, items


def kind(v):
    """'list' | ('dict', keytype, inner-kind)."""
    if isinstance(v, list):
        return 'list'
    k = next(iter(v))
    if isinstance(k, bool):
        kt = 'bool'
    elif isinstance(k, str):
        kt = 'str'
    else:
        kt = 'num'
    for key in v:
        assert (kt == 'bool') == isinstance(key, bool)
    inner = {repr(kind(x)) for x in v.values()}
    assert len(inner) == 1, inner
    return ('dict', kt, kind(next(iter(v.values()))))


def pascal(name):
    out, prev_digit = '', False
    for p in name.split('_'):
        if not p:
            continue
        if p.isdigit():
            out += ('_' if prev_digit else '') + p
            prev_digit = True
        else:
            out += p[0].upper() + p[1:].lower()
            prev_digit = False
    return out


def snake_to_kebab(name):
    return name.replace('_', '-')


def flt(x, lang):
    x = float(x)
    if math.isnan(x):
        return {'go': 'math.NaN()', 'ts': 'NaN', 'rs': 'f64::NAN', 'zig': 'nan'}[lang]
    if math.isinf(x):
        s = '' if x > 0 else '-'
        return {'go': f'math.Inf({1 if x > 0 else -1})', 'ts': f'{s}Infinity',
                'rs': f'f64::{"INFINITY" if x > 0 else "NEG_INFINITY"}', 'zig': f'{s}inf'}[lang]
    return repr(x)


def key_lit(k, kt, lang):
    if kt == 'bool':
        return 'true' if k else 'false'
    if kt == 'str':
        return '"' + k + '"'
    return flt(k, lang)


def wrap(items, indent, per_line=4):
    lines = []
    for i in range(0, len(items), per_line):
        lines.append(indent + ', '.join(items[i:i + per_line]) + ',')
    return '\n'.join(lines)


# ---------------------------------------------------------------- Go
def go_type(k):
    if k == 'list':
        return '[]float64'
    _, kt, inner = k
    key = {'bool': 'bool', 'str': 'string', 'num': 'float64'}[kt]
    return f'map[{key}]{go_type(inner)}'


def go_val(v, k, ind):
    if k == 'list':
        return '{\n' + wrap([flt(x, 'go') for x in v], ind + '\t') + '\n' + ind + '}'
    _, kt, inner = k
    parts = []
    for key, val in v.items():
        parts.append(f'{ind}\t{key_lit(key, kt, "go")}: {go_val(val, inner, ind + chr(9))},')
    return '{\n' + '\n'.join(parts) + '\n' + ind + '}'


def gen_go(out):
    d = os.path.join(out, 'go/performance/referencedata')
    os.makedirs(d, exist_ok=True)
    with open(os.path.join(d, 'doc.go'), 'w') as f:
        f.write('// Package referencedata holds expected values used by the performance\n'
                '// package tests, generated from py/performance/reference_data.\n'
                '//\n'
                '// Most values come from the PerformanceAnalytics R package applied to the\n'
                '// Bacon (2008) portfolio and benchmark returns, one value per streamed sample.\n'
                'package referencedata\n')
    for mod, items in modules():
        body = []
        for name, v in items:
            k = kind(v)
            body.append(f'// {pascal(mod)}{pascal(name)} is {mod}.{name}.\n'
                        f'var {pascal(mod)}{pascal(name)} = {go_type(k)}{go_val(v, k, "")}\n')
        text = '\n'.join(body)
        imp = 'import "math"\n\n' if 'math.' in text else ''
        with open(os.path.join(d, mod.replace('_', '') + '.go'), 'w') as f:
            f.write(f'// {HEADER.format(mod=mod)}\n\npackage referencedata\n\n{imp}{text}')


# ---------------------------------------------------------------- TS
def ts_type(k):
    if k == 'list':
        return 'readonly number[]'
    _, kt, inner = k
    key = {'bool': 'boolean', 'str': 'string', 'num': 'number'}[kt]
    return f'ReadonlyMap<{key}, {ts_type(inner)}>'


def ts_val(v, k, ind):
    if k == 'list':
        return '[\n' + wrap([flt(x, 'ts') for x in v], ind + '    ') + '\n' + ind + ']'
    _, kt, inner = k
    parts = []
    for key, val in v.items():
        parts.append(f'{ind}    [{key_lit(key, kt, "ts")}, {ts_val(val, inner, ind + "    ")}],')
    return 'new Map([\n' + '\n'.join(parts) + '\n' + ind + '])'


def gen_ts(out):
    d = os.path.join(out, 'ts/performance/reference-data')
    os.makedirs(d, exist_ok=True)
    for mod, items in modules():
        body = []
        for name, v in items:
            k = kind(v)
            body.append(f'/** {mod}.{name} */\nexport const {name}: {ts_type(k)} = {ts_val(v, k, "")};\n')
        with open(os.path.join(d, snake_to_kebab(mod) + '.ts'), 'w') as f:
            f.write(f'// {HEADER.format(mod=mod)}\n\n' + '\n'.join(body))


# ---------------------------------------------------------------- Rust
def rs_type(k):
    if k == 'list':
        return '&[f64]'
    _, kt, inner = k
    key = {'bool': 'bool', 'str': '&str', 'num': 'f64'}[kt]
    return f'&[({key}, {rs_type(inner)})]'


def rs_val(v, k, ind):
    if k == 'list':
        return '&[\n' + wrap([flt(x, 'rs') for x in v], ind + '    ') + '\n' + ind + ']'
    _, kt, inner = k
    parts = []
    for key, val in v.items():
        parts.append(f'{ind}    ({key_lit(key, kt, "rs")}, {rs_val(val, inner, ind + "    ")}),')
    return '&[\n' + '\n'.join(parts) + '\n' + ind + ']'


def gen_rs(out):
    d = os.path.join(out, 'rs/src/performance/reference_data')
    os.makedirs(d, exist_ok=True)
    mods = []
    for mod, items in modules():
        mods.append(mod)
        body = []
        for name, v in items:
            k = kind(v)
            body.append(f'/// `{mod}.{name}`\npub const {name}: {rs_type(k)} = {rs_val(v, k, "")};\n')
        with open(os.path.join(d, mod + '.rs'), 'w') as f:
            f.write(f'// {HEADER.format(mod=mod)}\n\n' + '\n'.join(body))
    with open(os.path.join(d, 'mod.rs'), 'w') as f:
        f.write('//! Expected values used by the performance tests, generated from\n'
                '//! `py/performance/reference_data`.\n'
                '#![allow(dead_code)]\n\n')
        for m in mods:
            f.write(f'pub mod {m};\n')
        f.write('\n/// Returns the value stored under `key` in a generated key/value table.\n'
                '///\n'
                '/// Panics if the key is absent.\n'
                'pub fn lookup<K: PartialEq + Copy + std::fmt::Debug, V: Copy>(pairs: &[(K, V)], key: K) -> V {\n'
                '    pairs\n'
                '        .iter()\n'
                '        .find(|(k, _)| *k == key)\n'
                '        .map(|(_, v)| *v)\n'
                '        .unwrap_or_else(|| panic!("reference data key {:?} not found", key))\n'
                '}\n')


# ---------------------------------------------------------------- Zig
def zig_type(k):
    if k == 'list':
        return '[]const f64'
    _, kt, inner = k
    key = {'bool': 'bool', 'str': '[]const u8', 'num': 'f64'}[kt]
    return f'[]const Entry({key}, {zig_type(inner)})'


def zig_val(v, k, ind):
    if k == 'list':
        return '&.{\n' + wrap([flt(x, 'zig') for x in v], ind + '    ') + '\n' + ind + '}'
    _, kt, inner = k
    parts = []
    for key, val in v.items():
        parts.append(f'{ind}    .{{ .key = {key_lit(key, kt, "zig")}, .value = {zig_val(val, inner, ind + "    ")} }},')
    return '&.{\n' + '\n'.join(parts) + '\n' + ind + '}'


ZIG_RESERVED = {'var', 'const', 'fn', 'test', 'error', 'type'}


def zig_ident(name):
    return f'@"{name}"' if name in ZIG_RESERVED else name


def gen_zig(out):
    d = os.path.join(out, 'zig/src/performance/reference_data')
    os.makedirs(d, exist_ok=True)
    mods = []
    for mod, items in modules():
        mods.append(mod)
        body = []
        for name, v in items:
            k = kind(v)
            body.append(f'/// `{mod}.{name}`\npub const {name.lower()}: {zig_type(k)} = {zig_val(v, k, "")};\n')
        text = '\n'.join(body)
        pre = 'const std = @import("std");\nconst Entry = @import("entry.zig").Entry;\n'
        if 'nan' in re.findall(r'\bnan\b', text) or re.search(r'\bnan\b', text):
            pre += 'const nan = std.math.nan(f64);\n'
        if re.search(r'\binf\b', text):
            pre += 'const inf = std.math.inf(f64);\n'
        with open(os.path.join(d, mod + '.zig'), 'w') as f:
            f.write(f'// {HEADER.format(mod=mod)}\n\n{pre}\n{text}')
    with open(os.path.join(d, 'entry.zig'), 'w') as f:
        f.write('//! Key/value entry type and lookup for the generated reference data tables.\n\n'
                'const std = @import("std");\n\n'
                '/// One key/value pair of a generated reference data table.\n'
                'pub fn Entry(comptime K: type, comptime V: type) type {\n'
                '    return struct { key: K, value: V };\n'
                '}\n\n'
                '/// Returns the value stored under `key`; panics if the key is absent.\n'
                'pub fn lookup(comptime K: type, comptime V: type, entries: []const Entry(K, V), key: K) V {\n'
                '    for (entries) |e| {\n'
                '        const equal = if (K == []const u8) std.mem.eql(u8, e.key, key) else e.key == key;\n'
                '        if (equal) return e.value;\n'
                '    }\n'
                '    @panic("reference data key not found");\n'
                '}\n')
    with open(os.path.join(d, 'reference_data.zig'), 'w') as f:
        f.write('//! Expected values used by the performance tests, generated from\n'
                '//! `py/performance/reference_data`.\n\n'
                'const entry = @import("entry.zig");\n'
                'pub const Entry = entry.Entry;\n'
                'pub const lookup = entry.lookup;\n\n')
        for m in mods:
            f.write(f'pub const {zig_ident(m)} = @import("{m}.zig");\n')


if __name__ == '__main__':
    out = sys.argv[1] if len(sys.argv) > 1 else ROOT
    gen_go(out)
    gen_ts(out)
    gen_rs(out)
    gen_zig(out)
    print('done')
