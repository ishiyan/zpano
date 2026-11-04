from __future__ import annotations

import csv
import json
import math
import sys
from dataclasses import asdict
from datetime import date, timedelta
from pathlib import Path
from ...indicators.lars_von_thienen.harmonic_cycle.validation import _integer, _sample
from ...indicators.lars_von_thienen.harmonic_cycle.model import phase_state
from ...indicators.lars_von_thienen.annual_cycle_model.model import (
    validate_dates, decode_annual, validate_annual_model, annual_wave)
from ...indicators.lars_von_thienen.harmonic_energy_superposition_wave.model import superposition_values, validate_superposition_model
from ...indicators.lars_von_thienen.cycle_scanner.scanner import (
    ScanConfig, scan_cycles, empty_scan, modeled_value)
from ...indicators.lars_von_thienen.dynamic_cycle_explorer.engine import centered_mean

METONIC_LAGS = (8, 11, 19, 27, 30, 38)


def phase_status(phase):
    # Explicit quadrant names avoid inventing undisclosed proprietary statuses.
    if math.isnan(phase):
        return 'unavailable'
    angle = phase % (2.0 * math.pi)
    if angle < math.pi / 2.0:
        return 'rising toward top'
    if angle < math.pi:
        return 'falling from top'
    if angle < 3.0 * math.pi / 2.0:
        return 'falling toward bottom'
    return 'rising from bottom'


def read_prices(path):
    with Path(path).open(newline='', encoding='utf-8') as stream:
        reader = csv.DictReader(stream)
        if not reader.fieldnames or not {'date','close'} <= set(reader.fieldnames):
            raise ValueError('CSV requires date,close columns')
        rows = [(date.fromisoformat(row['date']), _sample(row['close'])) for row in reader]
    validate_dates([day for day,_ in rows])
    if any(not math.isfinite(value) for _,value in rows):
        raise ValueError('CSV closes must be finite; rows are never silently dropped')
    return rows


def read_calendar(path):
    with Path(path).open(newline='', encoding='utf-8') as stream:
        reader = csv.DictReader(stream)
        if not reader.fieldnames or 'date' not in reader.fieldnames:
            raise ValueError('calendar CSV requires date column')
        dates = [date.fromisoformat(row['date']) for row in reader]
    validate_dates(dates)
    return dates


def json_clean(value):
    # JSON has no standard NaN. Unavailable report fields are null, whereas
    # numerical reference fixtures retain each language's NaN literal.
    if isinstance(value,float) and not math.isfinite(value):
        return None
    if isinstance(value,dict):
        return {key:json_clean(item) for key,item in value.items()}
    if isinstance(value,(tuple,list)):
        return [json_clean(item) for item in value]
    return value


def save_json(path, value):
    Path(path).write_text(json.dumps(json_clean(value), indent=2, allow_nan=False)+'\n', encoding='utf-8')


def load_annual(path):
    model = json.loads(Path(path).read_text(encoding='utf-8'))
    validate_annual_model(model)
    return model


def save_wave_csv(path, dates, values):
    if len(dates) != len(values):
        raise ValueError('wave and calendar lengths differ')
    with Path(path).open('w', newline='', encoding='utf-8') as stream:
        writer = csv.writer(stream)
        writer.writerow(['date','value'])
        writer.writerows((day.isoformat(), format(value,'.15g')) for day,value in zip(dates,values))


def make_cli_parser(slug):
    import argparse
    parser = argparse.ArgumentParser(description=slug + ': documented research reconstruction')
    parser.add_argument('--generate-reference-data', action='store_true')
    commands = parser.add_subparsers(dest='command')
    if slug == 'annual-cycle-pattern-decoder':
        fit = commands.add_parser('fit', help='decode a complete past year, save frozen model')
        fit.add_argument('--input', required=True)
        fit.add_argument('--year', type=int, required=True)
        fit.add_argument('--as-of', type=date.fromisoformat, default=date.today())
        fit.add_argument('--calendar', help='full-year date-column CSV for exact completeness check')
        fit.add_argument('--confirm-complete-year', action='store_true',
                         help='assert full coverage when no calendar is supplied; assertion recorded')
        fit.add_argument('--min-period', type=int, default=5)
        fit.add_argument('--max-period', type=int, default=125)
        fit.add_argument('--smoothing', type=float, default=1600.0)
        fit.add_argument('--genuine-threshold', type=float, default=49.0)
        fit.add_argument('--candidate-pool', type=int, default=10)
        fit.add_argument('--max-components', type=int, default=7)
        fit.add_argument('--peak-distance', type=int, default=2)
        fit.add_argument('--model-out', required=True)
        fit.add_argument('--wave-out', required=True)
        replay = commands.add_parser('replay', help='evaluate a frozen model on its stored dates')
        replay.add_argument('--model', required=True)
        replay.add_argument('--wave-out', required=True)
    elif slug == 'harmonic-energy-superposition-wave':
        build = commands.add_parser('build', help='average six historical annual models on target calendar')
        source = build.add_mutually_exclusive_group(required=True)
        source.add_argument('--models', nargs=6, help='six annual JSON models')
        source.add_argument('--history', help='dated price CSV; decode each required completed historical year')
        build.add_argument('--target-year', type=int, required=True)
        build.add_argument('--calendar', required=True, help='target-year session dates, no target prices')
        build.add_argument('--confirm-complete-years', action='store_true',
                           help='history mode: explicitly assert each source year has complete sessions')
        build.add_argument('--as-of', type=date.fromisoformat, default=date.today())
        build.add_argument('--normalization', choices=('rms','none'), default='rms')
        build.add_argument('--calendar-policy', choices=('extend','strict'), default='extend')
        build.add_argument('--invert', action='store_true')
        build.add_argument('--model-out', required=True)
        build.add_argument('--wave-out', required=True)
        replay = commands.add_parser('replay')
        replay.add_argument('--model', required=True)
        replay.add_argument('--wave-out', required=True)
    else:
        analyze = commands.add_parser('analyze', help='report a chosen dated range; optionally stream all rows')
        analyze.add_argument('--input', required=True)
        analyze.add_argument('--start', type=date.fromisoformat)
        analyze.add_argument('--end', type=date.fromisoformat)
        analyze.add_argument('--window', type=int, default=200, help='rolling buffer only; batch uses entire range')
        analyze.add_argument('--min-period', type=int, default=5)
        analyze.add_argument('--max-period', type=int, default=80)
        analyze.add_argument('--smoothing', type=float, default=1600.0)
        analyze.add_argument('--genuine-threshold', type=float, default=49.0)
        analyze.add_argument('--phase-cycles', type=float, default=2.0)
        analyze.add_argument('--top-k', type=int, default=3)
        analyze.add_argument('--peak-distance', type=int, default=2)
        analyze.add_argument('--max-segments', type=int, default=5)
        analyze.add_argument('--rolling', action='store_true')
        analyze.add_argument('--json-out')
        analyze.add_argument('--series-out', help='retrospective CSV: rolling fields or final-model/highlighter overlay')
        analyze.add_argument('--horizon', type=int, default=0, help='static last-model projection in bars')
        if slug == 'phasing-oscillator':
            analyze.add_argument('--short-min', type=int, default=5)
            analyze.add_argument('--short-max', type=int, default=30)
            analyze.add_argument('--long-min', type=int, default=31)
            analyze.add_argument('--long-max', type=int, default=80)
        if slug == 'cycle-phase-score':
            evaluate = commands.add_parser('phase', help='score a directly supplied sine phase')
            evaluate.add_argument('--radians', type=float, required=True)
    return parser


def scanner_report(scan, config, dates):
    return dict(schema='lvt-cycle-scan', version=1, research_reconstruction=True,
                start_date=dates[0].isoformat(), end_date=dates[-1].isoformat(),
                sample_count=len(dates), config=asdict(config),
                phase_convention='sine radians at analysis endpoint',
                validation_input='raw prices, most recent complete segments',
                accepted_count=scan.accepted_count,
                cycles=[cycle._asdict() for cycle in scan.cycles],
                spectrum=dict(periods=list(range(config.min_period,config.max_period+1)),
                              amplitude=scan.spectrum.values, genuine_percent=scan.genuine_percent))


def run_cli(slug, args, master_class, flatten):
    """Local CSV/JSON workflows; no WTT API calls or network dependencies."""
    if args.command == 'phase':
        if not math.isfinite(args.radians):
            raise ValueError('phase must be finite')
        print(json.dumps(dict(value=phase_state(args.radians), status=phase_status(args.radians))))
        return
    if slug == 'annual-cycle-pattern-decoder':
        if args.command == 'fit':
            rows = read_prices(args.input)
            calendar = read_calendar(args.calendar) if args.calendar else None
            model = decode_annual(rows, args.year, args.as_of, calendar, args.confirm_complete_year,
                                  min_period=args.min_period, max_period=args.max_period,
                                  smoothing=args.smoothing, genuine_threshold=args.genuine_threshold,
                                  candidate_pool=args.candidate_pool, max_components=args.max_components,
                                  peak_distance=args.peak_distance)
            save_json(args.model_out, model)
        else:
            model = load_annual(args.model)
        dates = [date.fromisoformat(day) for day in model['dates']]
        save_wave_csv(args.wave_out, dates, [annual_wave(model,i) for i in range(len(dates))])
        print(json.dumps(dict(year=model['year'], correlation=model['correlation'],
                              components=len(model['components']), sample_count=len(dates))))
        return
    if slug == 'harmonic-energy-superposition-wave':
        if args.command == 'build':
            dates = read_calendar(args.calendar)
            validate_dates(dates, args.target_year)
            if args.models:
                models = [load_annual(path) for path in args.models]
            else:
                rows = read_prices(args.history)
                models = [decode_annual(rows, args.target_year-lag, args.as_of,
                                       confirm_complete=args.confirm_complete_years)
                          for lag in METONIC_LAGS]
            # No target price is ever read. Source years alone supply fitting
            # and source-wave normalization, preventing target-year lookahead.
            values, model = superposition_values(models, args.target_year, len(dates),
                                                 args.normalization, args.calendar_policy, args.invert)
            model['dates'] = [day.isoformat() for day in dates]
            save_json(args.model_out, model)
        else:
            model = json.loads(Path(args.model).read_text(encoding='utf-8'))
            validate_superposition_model(model)
            dates = [date.fromisoformat(day) for day in model['dates']]
            validate_dates(dates, model['target_year'])
            values, _ = superposition_values(model['annual_models'], model['target_year'], len(dates),
                                              model['normalization'], model['calendar_policy'], model['invert'])
        save_wave_csv(args.wave_out, dates, values)
        print(json.dumps(dict(target_year=model['target_year'], source_years=model['source_years'],
                              sample_count=len(values))))
        return
    rows = read_prices(args.input)
    if args.start and args.end and args.start > args.end:
        raise ValueError('start must not exceed end')
    rows = [(day,value) for day,value in rows
            if (args.start is None or day >= args.start) and (args.end is None or day <= args.end)]
    if len(rows) < 6:
        raise ValueError('selected range needs >= 6 prices')
    _integer(args.horizon, 'horizon', 0)
    dates, prices = [day for day,_ in rows], [value for _,value in rows]
    common = dict(window=args.window if args.rolling else len(prices),
                  min_period=args.min_period, max_period=args.max_period,
                  smoothing=args.smoothing, genuine_threshold=args.genuine_threshold,
                  phase_cycles=args.phase_cycles, top_k=args.top_k,
                  peak_distance=args.peak_distance, max_segments=args.max_segments)
    config = ScanConfig(**common)
    scan = (empty_scan(config) if args.rolling and len(prices) < config.window else
            scan_cycles(prices[-config.window:] if args.rolling else prices, config))
    report = scanner_report(scan, config, dates[-config.window:] if args.rolling else dates)
    series = []
    if slug == 'phasing-oscillator':
        indicator = master_class(window=common['window'], short_min=args.short_min,
                                 short_max=args.short_max, long_min=args.long_min, long_max=args.long_max,
                                 smoothing=args.smoothing, genuine_threshold=args.genuine_threshold,
                                 phase_cycles=args.phase_cycles)
    else:
        indicator = master_class(**common)
    if args.rolling or slug == 'phasing-oscillator':
        for day,price in rows:
            output = indicator.update(price)
            series.append(dict(date=day.isoformat(), **flatten(indicator,output)))
        report['output'] = series[-1]
        if slug == 'phasing-oscillator':
            report['phasing_config'] = dict(short=asdict(indicator.short.config), long=asdict(indicator.long.config))
    else:
        cycle = scan.cycles[0]
        if slug == 'complete-cycle-scanner':
            report['output'] = flatten(indicator,scan)
        elif slug == 'dominant-market-vibration':
            report['output'] = cycle._asdict()
        elif slug == 'cycle-phase-score':
            report['output'] = dict(value=phase_state(cycle.phase), period=cycle.period,
                                    phase=cycle.phase, status=phase_status(cycle.phase))
        else:
            report['output'] = dict(value=cycle.value, period=cycle.period,
                                    eta_top=cycle.eta_top, eta_bottom=cycle.eta_bottom)
        # Batch historical overlay uses the final selected period/model. This is
        # retrospective, and expressly distinct from the causal --rolling CSV.
        for i,day in enumerate(dates):
            row = dict(date=day.isoformat(), fitted=modeled_value(cycle,i-(len(prices)-1)))
            if slug == 'dynamic-cycle-explorer':
                if math.isfinite(cycle.period):
                    long_length = int(cycle.period)
                    row['highlighter'] = centered_mean(prices,i,max(2,math.floor(long_length/2+0.5))) - centered_mean(prices,i,long_length)
                else:
                    row['highlighter'] = math.nan
            series.append(row)
    if slug == 'phasing-oscillator':
        # Project the two selected phase STATES, not an unrelated global
        # scanner wave. Re-analyze frozen buffers without advancing a sample.
        side_cycles = []
        for name, side in (('short',indicator.short),('long',indicator.long)):
            side_scan = scan_cycles(list(side.prices),side.config) if side.is_primed else empty_scan(side.config)
            report[name+'_scan'] = scanner_report(side_scan,side.config,dates[-side.config.window:])
            side_cycles.append(side_scan.cycles[0])
        report['projection'] = []
        for h in range(1,args.horizon+1):
            scores = [phase_state(cycle.phase+2.0*math.pi*h/cycle.period) for cycle in side_cycles]
            report['projection'].append(dict(bars_ahead=h,value=sum(scores),
                                              short_score=scores[0],long_score=scores[1]))
    elif slug == 'cycle-phase-score':
        cycle = scan.cycles[0]
        report['projection'] = [dict(bars_ahead=h,value=phase_state(cycle.phase+2.0*math.pi*h/cycle.period))
                                for h in range(1,args.horizon+1)]
    else:
        report['projection'] = [dict(bars_ahead=h,value=modeled_value(scan.cycles[0],h))
                                for h in range(1,args.horizon+1)]
    if args.series_out:
        with Path(args.series_out).open('w', newline='', encoding='utf-8') as stream:
            writer = csv.DictWriter(stream, fieldnames=list(series[0]))
            writer.writeheader()
            for row in series:
                writer.writerow({key:('' if isinstance(value,float) and math.isnan(value)
                                      else format(value,'.15g') if isinstance(value,float) else value)
                                 for key,value in row.items()})
    if args.json_out:
        save_json(args.json_out, report)
    print(json.dumps(json_clean(report), allow_nan=False))


def main() -> None:
    """Run one of the seven local CSV/JSON research workflows."""
    engines = {
        "complete-cycle-scanner": "complete_cycle_scanner",
        "dominant-market-vibration": "dominant_market_vibration",
        "dynamic-cycle-explorer": "dynamic_cycle_explorer",
        "cycle-phase-score": "cycle_phase_score",
        "phasing-oscillator": "phasing_oscillator",
    }
    tools = (*engines, "annual-cycle-pattern-decoder", "harmonic-energy-superposition-wave")
    if len(sys.argv) < 2 or sys.argv[1] not in tools:
        raise SystemExit("usage: python -m py.cmd.lvt <workflow> <command> [options]\n" + "\n".join(tools))
    slug = sys.argv[1]
    parser = make_cli_parser(slug)
    args = parser.parse_args(sys.argv[2:])
    if args.generate_reference_data:
        parser.error("reference fixtures are immutable; use the original prepared master to regenerate them")
    if args.command is None:
        parser.print_help()
        return
    master_class = None
    flatten = None
    if slug in engines:
        import importlib
        module = importlib.import_module("py.indicators.lars_von_thienen." + engines[slug] + ".engine")
        master_class = module.Engine
        def flatten(indicator, output):
            if slug != "complete-cycle-scanner":
                return output._asdict()
            result = {"accepted_count": output.accepted_count}
            for i, period in enumerate(range(indicator.config.min_period, indicator.config.max_period+1)):
                result["amplitude_p"+str(period)] = output.spectrum.values[i]
                result["genuine_p"+str(period)] = output.genuine_percent[i]
            for i, cycle in enumerate(output.cycles, 1):
                for field, value in cycle._asdict().items():
                    result["rank" + str(i) + "_" + field] = value
            return result
    try:
        run_cli(slug, args, master_class, flatten)
    except (ValueError, OSError) as error:
        parser.error(str(error))
