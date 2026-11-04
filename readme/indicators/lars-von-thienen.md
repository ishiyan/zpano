# Lars von Thienen indicators and research tools

The 20 prepared candidates are implemented in Python, Go, TypeScript, Zig and
Rust: 18 price-input streaming indicators, seven command-line workflows, and
two frozen-model readers. All implementations live in the Lars von Thienen
author folders. The prepared sources and all 100 language fixtures remain
unchanged; tests reference their original array/function names and cover all
136 parameter cases.

## Names and registered outputs

The mnemonic includes configuration and the standard component suffix. These
prefixes are distinct from the existing Ehlers and William Blau indicators.

| Prepared folder | Prefix | Registered outputs |
|---|---|---|
| cyclic-smoothed-relative-strength-index | `crsi` | value, lower, upper, band |
| cycle-swing-momentum | `csmo` | value, lower, upper, score, band |
| detrended-rhythm-oscillators | `dro` | value, direction, completed cycle length, provisional pivot value/offset |
| zero-lag-exponential-moving-average | `lzlema` | value |
| hodrick-prescott-smoothing-and-detrending | `hpsd` | endpoint trend, residual |
| goertzel-cycle-spectrum | `gcs` | amplitude/power/phase heatmaps, dominant period/amplitude/phase |
| bartels-periodicity-score | `bps` | score, genuine percentage, segment count |
| cycle-strength | `cstr` | strength, amplitude, phase |
| composite-cycle | `ccyc` | modeled value, slope, next value |
| cycle-polarity | `cpol` | polarity, short/long values and slopes |
| metonic-sentiment-index | `msent` | delayed sentiment, daily sentiment, smoothed close/open |
| metonic-relative-strength-index | `mrsi` | delayed value, underlying RSI |
| metonic-williams-percent-r | `mwillr` | delayed value, underlying Williams %R |
| complete-cycle-scanner | `cscan` | amplitude/genuine heatmaps, dominant period, accepted count |
| dominant-market-vibration | `dmv` | period, amplitude, phase, strength, genuine percentage, value, slope, top/bottom ETAs |
| dynamic-cycle-explorer | `dce` | fitted value, period, phase, ETAs, delayed highlighter/index/fit and polyline |
| cycle-phase-score | `cphase` | score, period, phase |
| phasing-oscillator | `phosc` | combined score, short/long scores and periods |

ZLEMA's public class/identifier is
`LarsVonThienenZeroLagExponentialMovingAverage`, distinguishing it from the
existing Ehlers implementation. Dominant vibration retains segment count,
fit amplitude and fit sample count in its numerical snapshot/report; the nine
registered plots fit every language's output interface. Identifiers retain the
existing numeric values, with the new author group displayed alphabetically.

Annual fitting and historical superposition have no price-input indicator
identifiers. Their readers consume already-fitted models and accept `update()`
without a price. Annual playback can extrapolate; superposition playback stops
at the declared sample count.

## Extracted components

Packages use each language's established folder spelling. Python names are
shown below; equivalent packages exist in all five languages.

| Component | Responsibility | Consumers |
|---|---|---|
| `harmonic_cycle` | Published Goertzel recurrence, sine models, slopes, phase scores, polarity/strength, validation and compensated summation | Spectrum, strength, composite, polarity, scanner, annual and superposition |
| `hodrick_prescott_smoothing_and_detrending/solver` | Complete retrospective HP solve and spectral cycle preprocessing | HP endpoint, cycle swing, Bartels, spectrum, scanner and annual fitting |
| `adaptive_bands` | Prepared empirical adaptive-band calculation | cRSI and cycle swing |
| `cyclic_smoothed_relative_strength_index/rsi` | Prepared Wilder seed/update semantics | cRSI and Metonic RSI |
| `zero_lag_exponential_moving_average/smoother` | Prepared ZLEMA lag and seed conventions | ZLEMA and Metonic sentiment |
| `bartels_periodicity_score/scorer` | Complete-segment vector coherence diagnostic | Bartels indicator, scanner and annual fitting |
| `metonic` | Calendar validation/filling and delay averaging | All three Metonic indicators |
| `cycle_scanner` | HP → integer GDFT → peaks → raw-segment validation → A/P ranking → recent phase | Scanner, dominant vibration, explorer, phase score, phasing and annual fitting |
| `annual_cycle_model` | Exhaustive signed-Pearson subset selection, portable schema, coverage validation and frozen annual playback | Annual decoder and historical superposition |
| `harmonic_energy_superposition_wave/model` | Six-year source alignment, source-only normalization and bounded playback | Superposition CLI/reader |

Indicator-specific state remains embedded: cRSI's vibration recurrence,
cycle-swing endpoint difference, DRO pivot confirmation and completed-cycle
median, fixed-period model composition, explorer highlighter alignment,
Metonic sentiment and Williams calculations, and the entity/metadata adapters.
The seven prepared copies of `cycle_core.py` are replaced by native components
in the library; they are retained in the untouched conversion sources.

## Command-line workflows

Build/run prefixes from the repository root:

| Language | Command |
|---|---|
| Python | `python3 -m py.cmd.lvt` |
| Go, from `go/` | `go run ./cmd/lvt` |
| TypeScript, from `ts/`, after `npm run build` | `node dist/cmd/lvt/main.js` |
| Rust, from `rs/` | `cargo run --bin lvt --` |
| Zig, from `zig/` | `zig build lvt --` |

Append one of the following workflows and options to the language prefix:

```text
complete-cycle-scanner analyze --input prices.csv --json-out scan.json
                         --rolling --window 200 --series-out scan.csv --horizon 20
 dominant-market-vibration analyze --input prices.csv --json-out vibration.json
 dynamic-cycle-explorer analyze --input prices.csv --series-out explorer.csv --horizon 20
 cycle-phase-score analyze --input prices.csv --json-out phase.json
 cycle-phase-score phase --radians 1.5
 phasing-oscillator analyze --input prices.csv --short-min 5 --short-max 30
                    --long-min 31 --long-max 80 --series-out phasing.csv
 annual-cycle-pattern-decoder fit --input prices.csv --calendar sessions.csv
                    --year 2020 --as-of 2021-01-01 --model-out annual.json --wave-out annual.csv
 annual-cycle-pattern-decoder replay --model annual.json --wave-out annual-replayed.csv
 harmonic-energy-superposition-wave build --models model-2013.json model-2010.json
                    model-2002.json model-1994.json model-1991.json model-1983.json
                    --target-year 2021 --calendar target-sessions.csv
                    --model-out superposition.json --wave-out superposition.csv
 harmonic-energy-superposition-wave build --history history.csv --confirm-complete-years
                    --target-year 2021 --calendar target-sessions.csv --as-of 2022-01-01
                    --model-out superposition.json --wave-out superposition.csv
 harmonic-energy-superposition-wave replay --model superposition.json --wave-out replay.csv
```

CSV input requires strictly increasing ISO `date` and finite `close` columns;
a calendar requires a `date` column. Quoted CSV fields are supported. Numeric
NaNs become JSON `null` and empty CSV fields. Scanner options include
`--start`, `--end`, `--min-period`, `--max-period`, `--smoothing`,
`--genuine-threshold`, `--phase-cycles`, `--top-k`, `--peak-distance` and
`--max-segments`. Annual options include `--candidate-pool` and
`--max-components`. Superposition also supports `--normalization none|rms`,
`--calendar-policy strict|extend` and `--invert`.

Each prepared optional folder supplies runnable examples. Their histories,
dates and full-year calendars are synthetic; use the folder's
`example-notes.md` when interpreting the saved reports.

All 18 registered indicators are also in each `icalc/settings.json`; the normal
factory and `icalc` routes exercise their metadata and entity outputs.

## Numerical and lifecycle contracts

- HP's zero lambda returns the identity trend; spectral zero lambda uses raw
  prices. Rolling HP publishes the current endpoint and does not revise
  previously emitted values.
- Goertzel phase is a sine phase at the window start, scanner phase is at the
  analysis endpoint, and annual phase is at ordinal zero.
- Scanner validation uses the latest complete **raw-price** segments, strict
  `G > threshold`, full-window A/P ranking and a separately detrended recent
  phase suffix. Phasing uses independent, disjoint short/long scans.
- Prepared RSI consumes L actual changes, seeds flat history at 100 and permits
  length one. Prepared ZLEMA preserves its own lag/seeding semantics.
- Metonic delays are **6939, 4016 and 2922 calendar days**. Missing calendar
  days must be filled with the complete previous OHLC payload before updating.
  An unavailable underlying value still occupies a calendar slot. Default
  252-bar composites remain NaN; synthetic short delays are test cases.
- NaN observations skip advancement. Invalid observations are rejected before
  updating. Reset retains configuration. Construction copies caller parameters.
- Annual fitting examines all subsets up to seven components from a pool of
  at most 16, maximizes signed Pearson correlation and uses `1e-12` ties. A
  past year must match a supplied calendar exactly or carry the caller's
  explicit completeness assertion.
- Version-one model schemas are `lvt-annual-cycle-model` and
  `lvt-superposition-model`. Source years are target minus **8,11,19,27,30,38**.
  Only source-year waves are normalized; target prices do not enter fitting.

Fixtures store 15 significant digits. Arithmetic-filter assertions use
`max(1e-13, abs(expected) * 5e-15)`, retaining the exact NaN mask and allowing
only saved-decimal rounding at larger magnitudes. Spectral/phase assertions
use the skill's `1e-10` allowance for trigonometric numerical drift. The Go
fixture audit measured a maximum absolute difference of approximately
`1.8e-11` in phasing, versus about `5.1e-13` in large arithmetic values and
`5.7e-14` in cRSI. Universal 13-decimal absolute agreement is therefore not
claimed for spectral/phase outputs. No fixture values were changed to conceal
these differences.

## Verification

Validation passed the indicator suites in all five languages, every `icalc`
program, and all seven CLI workflows in every language. The full TypeScript
suite passed 23,287 specs; the full Zig suite passed 2,095 tests. Fifty
cross-language frozen-model replay checks passed with maximum absolute drift
of `1.021405182655144e-14`. This replay result does not change the spectral
fixture precision limits described above.

The Zig calendar helper has direct tests using the unchanged prepared OHLC
arrays: gaps repeat every field of the preceding payload, and duplicate or
reversed days are rejected. Its `dailyRecord(Payload)` type supplies the same
record type to `fillCalendarDays()` and its returned owned slice.

## Ownership

Go public indicators hold `sync.RWMutex`; internal numerical engines and model
helpers are caller-confined. Python, TypeScript, Rust and Zig follow the
repository's single-caller convention. Returned Python snapshots use immutable
records; changing an emitted snapshot does not configure an indicator.

Zig indicators own their state and scratch arenas: call `deinit()`. Update
commits the scratch state only after success. Snapshot and heatmap/polyline
slices are borrowed and valid until the next update/reset/deinit; copy them if
retaining history. Metadata takes an out-pointer and copies per-output text.
The single-line wrapper repairs its stored slices after relocation.

Zig annual fitting uses a temporary workspace and returns owned components;
`Model.deinit(allocator)` releases components and the optional date list, while
date strings remain borrowed from the input. Parsed JSON models remain tied to
the parser's lifetime. Both frozen readers have `deinit()`; the annual reader
owns a cloned model arena and the superposition reader owns its wave array.
Standalone scanner batch results use a caller-provided arena/workspace.

DRO's source-derived files retain **MPL-2.0** notices. Other numerical research
ports retain the prepared source's **CC BY 4.0** attribution. These algorithms
are the documented research reconstructions, with no claim of proprietary WTT
platform equivalence; consult each indicator's copied `description.md`.
