//! Local CSV/JSON research workflows using native numerical components.
const std = @import("std");
const indicators = @import("indicators");
const annual = indicators.lvt_annual_model;
const superposition = indicators.lvt_superposition_model;
const scanner = indicators.lvt_cycle_scanner;
const harmonic = indicators.lvt_harmonic_cycle;
const explorer = indicators.lvt_explorer_engine;
const memory = indicators.lvt_memory;
const Value = std.json.Value;
const workflows = [_][]const u8{ "complete-cycle-scanner", "dominant-market-vibration", "dynamic-cycle-explorer", "cycle-phase-score", "phasing-oscillator", "annual-cycle-pattern-decoder", "harmonic-energy-superposition-wave" };
fn eq(a: []const u8, b: []const u8) bool {
    return std.mem.eql(u8, a, b);
}
fn number(n: f64) Value {
    return if (std.math.isFinite(n)) .{ .float = n } else .null;
}
fn value(a: std.mem.Allocator, input: anytype) !Value {
    const T = @TypeOf(input);
    return switch (@typeInfo(T)) {
        .bool => .{ .bool = input },
        .int, .comptime_int => .{ .integer = @intCast(input) },
        .float, .comptime_float => number(input),
        .optional => if (input) |v| try value(a, v) else .null,
        .array => blk: {
            var list = std.json.Array.init(a);
            for (input) |v| try list.append(try value(a, v));
            break :blk .{ .array = list };
        },
        .pointer => |p| blk: {
            if (p.size == .one) {
                if (@typeInfo(p.child) == .array and @typeInfo(p.child).array.child == u8) break :blk .{ .string = input[0..] };
                break :blk try value(a, input.*);
            }
            if (p.child == u8) break :blk .{ .string = input };
            var list = std.json.Array.init(a);
            for (input) |v| try list.append(try value(a, v));
            break :blk .{ .array = list };
        },
        .@"struct" => |s| blk: {
            var object: std.json.ObjectMap = .empty;
            inline for (s.fields) |f| try object.put(a, f.name, try value(a, @field(input, f.name)));
            break :blk .{ .object = object };
        },
        else => @compileError("unsupported JSON value"),
    };
}
const Context = struct {
    allocator: std.mem.Allocator,
    io: std.Io,
    stdout: *std.Io.Writer,
    options: std.StringHashMap([]const u8),
    fn text(self: Context, key: []const u8) []const u8 {
        return self.options.get(key) orelse "";
    }
    fn required(self: Context, key: []const u8) ![]const u8 {
        const v = self.text(key);
        if (v.len == 0) return error.MissingOption;
        return v;
    }
    fn real(self: Context, key: []const u8) !f64 {
        const n = try std.fmt.parseFloat(f64, try self.required(key));
        if (!std.math.isFinite(n)) return error.NonfiniteOption;
        return n;
    }
    fn integer(self: Context, key: []const u8) !isize {
        return std.fmt.parseInt(isize, try self.required(key), 10);
    }
    fn flag(self: Context, key: []const u8) bool {
        return eq(self.text(key), "true");
    }
    fn read(self: Context, path: []const u8) ![]u8 {
        return std.Io.Dir.cwd().readFileAlloc(self.io, path, self.allocator, .unlimited);
    }
    fn save(self: Context, path: []const u8, bytes: []const u8) !void {
        const file = try std.Io.Dir.cwd().createFile(self.io, path, .{});
        defer file.close(self.io);
        try file.writePositionalAll(self.io, bytes, 0);
    }
    fn json(self: Context, v: anytype) ![]u8 {
        var out = std.Io.Writer.Allocating.init(self.allocator);
        try std.json.Stringify.value(v, .{ .emit_null_optional_fields = false }, &out.writer);
        try out.writer.writeByte('\n');
        return out.toOwnedSlice();
    }
    fn saveJson(self: Context, path: []const u8, v: anytype) !void {
        try self.save(path, try self.json(v));
    }
    fn print(self: Context, v: anytype) !void {
        try self.stdout.writeAll(try self.json(v));
    }
    fn csv(self: Context, path: []const u8) ![][]const []const u8 {
        const bytes = try self.read(path);
        var rows: std.ArrayList([]const []const u8) = .empty;
        var row: std.ArrayList([]const u8) = .empty;
        var field: std.ArrayList(u8) = .empty;
        var quoted = false;
        var i: usize = 0;
        while (i < bytes.len) : (i += 1) {
            const c = bytes[i];
            if (c == '"') {
                if (quoted and i + 1 < bytes.len and bytes[i + 1] == '"') {
                    try field.append(self.allocator, '"');
                    i += 1;
                } else quoted = !quoted;
            } else if (!quoted and (c == ',' or c == '\n')) {
                try row.append(self.allocator, try self.allocator.dupe(u8, std.mem.trimEnd(u8, field.items, "\r")));
                field.clearRetainingCapacity();
                if (c == '\n') {
                    var nonempty = false;
                    for (row.items) |f| nonempty = nonempty or f.len > 0;
                    if (nonempty) try rows.append(self.allocator, try row.toOwnedSlice(self.allocator)) else row.clearRetainingCapacity();
                }
            } else try field.append(self.allocator, c);
        }
        if (quoted) return error.UnterminatedCsvQuote;
        if (field.items.len > 0 or row.items.len > 0) {
            try row.append(self.allocator, try self.allocator.dupe(u8, std.mem.trimEnd(u8, field.items, "\r")));
            try rows.append(self.allocator, try row.toOwnedSlice(self.allocator));
        }
        if (rows.items.len < 2) return error.EmptyCsv;
        for (rows.items) |r| if (r.len != rows.items[0].len) return error.InconsistentCsvColumns;
        return rows.toOwnedSlice(self.allocator);
    }
    fn prices(self: Context, path: []const u8) ![]annual.PriceRecord {
        const rows = try self.csv(path);
        var d: ?usize = null;
        var c: ?usize = null;
        for (rows[0], 0..) |s, i| {
            if (eq(s, "date")) d = i;
            if (eq(s, "close")) c = i;
        }
        const di = d orelse return error.DateColumnRequired;
        const ci = c orelse return error.CloseColumnRequired;
        const records = try self.allocator.alloc(annual.PriceRecord, rows.len - 1);
        const dates = try self.allocator.alloc([]const u8, records.len);
        for (rows[1..], 0..) |r, i| {
            const v = try std.fmt.parseFloat(f64, r[ci]);
            if (!std.math.isFinite(v)) return error.NonfinitePrice;
            dates[i] = r[di];
            records[i] = .{ .date = r[di], .close = v };
        }
        try annual.validateDates(dates, null);
        return records;
    }
    fn calendar(self: Context, path: []const u8) ![]const []const u8 {
        const rows = try self.csv(path);
        var column: ?usize = null;
        for (rows[0], 0..) |s, i| if (eq(s, "date")) {
            column = i;
        };
        const di = column orelse return error.DateColumnRequired;
        const dates = try self.allocator.alloc([]const u8, rows.len - 1);
        for (rows[1..], 0..) |r, i| dates[i] = r[di];
        try annual.validateDates(dates, null);
        return dates;
    }
    fn loadAnnual(self: Context, path: []const u8) !annual.Model {
        const parsed = try std.json.parseFromSlice(annual.Model, self.allocator, try self.read(path), .{ .ignore_unknown_fields = true });
        try parsed.value.validate();
        return parsed.value;
    }
    fn saveWave(self: Context, path: []const u8, dates: []const []const u8, values: []const f64) !void {
        if (dates.len != values.len) return error.CalendarLengthMismatch;
        var out = std.Io.Writer.Allocating.init(self.allocator);
        try out.writer.writeAll("date,value\n");
        for (dates, values) |d, v| {
            try out.writer.print("{s},", .{d});
            if (std.math.isFinite(v)) try out.writer.print("{d}", .{v});
            try out.writer.writeByte('\n');
        }
        try self.save(path, try out.toOwnedSlice());
    }
    fn fit(self: Context, records: []const annual.PriceRecord, year: isize, calendar_: ?[]const []const u8, confirm: bool, p: annual.Params) !annual.Model {
        var arena = std.heap.ArenaAllocator.init(self.allocator);
        defer arena.deinit();
        return memory.clone(self.allocator, try annual.decodeAnnual(arena.allocator(), records, year, self.text("as-of"), calendar_, confirm, p));
    }
    fn scan(self: Context, prices_: []f64, config: scanner.ScanConfig) !scanner.ScanOutput {
        var arena = std.heap.ArenaAllocator.init(self.allocator);
        defer arena.deinit();
        const a = arena.allocator();
        return memory.clone(self.allocator, if (prices_.len < @as(usize, @intCast(config.window))) try scanner.emptyScan(a, config) else try scanner.scanCycles(a, prices_[prices_.len - @as(usize, @intCast(config.window)) ..], config));
    }
    fn report(self: Context, scan_: scanner.ScanOutput, config: scanner.ScanConfig, dates: []const []const u8) !Value {
        var v = try value(self.allocator, .{ .schema = @as([]const u8, "lvt-cycle-scan"), .version = @as(i32, 1), .research_reconstruction = true, .start_date = dates[0], .end_date = dates[dates.len - 1], .sample_count = dates.len, .config = config, .phase_convention = @as([]const u8, "sine radians at analysis endpoint"), .validation_input = @as([]const u8, "raw prices, most recent complete segments"), .accepted_count = scan_.accepted_count, .cycles = scan_.cycles });
        const periods = try self.allocator.alloc(isize, scan_.amplitude.len);
        for (periods, 0..) |*p, i| p.* = config.min_period + @as(isize, @intCast(i));
        try v.object.put(self.allocator, "spectrum", try value(self.allocator, .{ .periods = periods, .amplitude = scan_.amplitude, .genuine_percent = scan_.genuine_percent }));
        return v;
    }
    fn calculate(self: Context, prices_: []f64, slug: []const u8, c: scanner.ScanConfig, short: scanner.ScanConfig, long: scanner.ScanConfig) !Value {
        const a = self.allocator;
        const scan_ = try self.scan(prices_, c);
        const v = scan_.cycles[0];
        if (eq(slug, "complete-cycle-scanner")) {
            var out = Value{ .object = .empty };
            try out.object.put(a, "accepted_count", number(scan_.accepted_count));
            for (scan_.amplitude, 0..) |amp, i| {
                try out.object.put(a, try std.fmt.allocPrint(a, "amplitude_p{d}", .{c.min_period + @as(isize, @intCast(i))}), number(amp));
                try out.object.put(a, try std.fmt.allocPrint(a, "genuine_p{d}", .{c.min_period + @as(isize, @intCast(i))}), number(scan_.genuine_percent[i]));
            }
            for (scan_.cycles, 0..) |cycle_, i| {
                var fields = (try value(a, cycle_)).object.iterator();
                while (fields.next()) |f| try out.object.put(a, try std.fmt.allocPrint(a, "rank{d}_{s}", .{ i + 1, f.key_ptr.* }), f.value_ptr.*);
            }
            return out;
        }
        if (eq(slug, "dominant-market-vibration")) return value(a, v);
        if (eq(slug, "cycle-phase-score")) return value(a, .{ .value = try harmonic.phaseState(a, v.phase), .period = v.period, .phase = v.phase, .status = phaseStatus(v.phase) });
        if (eq(slug, "phasing-oscillator")) {
            const x = (try self.scan(prices_, short)).cycles[0];
            const y = (try self.scan(prices_, long)).cycles[0];
            const xs = try harmonic.phaseState(a, x.phase);
            const ys = try harmonic.phaseState(a, y.phase);
            return value(a, .{ .value = xs + ys, .short_score = xs, .long_score = ys, .short_period = x.period, .long_period = y.period });
        }
        var highlighter = std.math.nan(f64);
        var index = std.math.nan(f64);
        var fitted = std.math.nan(f64);
        if (std.math.isFinite(v.period)) {
            const length: isize = @intFromFloat(v.period);
            const delay = @divTrunc(length, 2);
            const buffer = prices_[prices_.len - @as(usize, @intCast(c.window)) ..];
            const center = @as(isize, @intCast(buffer.len)) - 1 - delay;
            highlighter = (try explorer.centeredMean(a, buffer, center, @max(2, @as(isize, @intFromFloat(@floor(v.period / 2 + 0.5)))))) - (try explorer.centeredMean(a, buffer, center, length));
            if (std.math.isFinite(highlighter)) {
                index = @floatFromInt(@as(isize, @intCast(prices_.len)) - 1 - delay);
                fitted = try scanner.modeledValue(a, v, -@as(f64, @floatFromInt(delay)));
            }
        }
        return value(a, .{ .value = v.value, .period = v.period, .phase = v.phase, .eta_top = v.eta_top, .eta_bottom = v.eta_bottom, .highlighter = highlighter, .highlighter_index = index, .fitted_at_highlighter = fitted });
    }
};
fn phaseStatus(p: f64) []const u8 {
    if (std.math.isNan(p)) return "unavailable";
    const angle = @mod(p, 2 * std.math.pi);
    return if (angle < std.math.pi / 2.0) "rising toward top" else if (angle < std.math.pi) "falling from top" else if (angle < 3 * std.math.pi / 2.0) "falling toward bottom" else "rising from bottom";
}
fn today(a: std.mem.Allocator, io: std.Io) ![]const u8 {
    const z = @divFloor(std.Io.Clock.real.now(io).toSeconds(), 86400) + 719468;
    const era = @divFloor(z, 146097);
    const doe = z - era * 146097;
    const yoe = @divFloor(doe - @divFloor(doe, 1460) + @divFloor(doe, 36524) - @divFloor(doe, 146096), 365);
    var y = yoe + era * 400;
    const doy = doe - (365 * yoe + @divFloor(yoe, 4) - @divFloor(yoe, 100));
    const mp = @divFloor(5 * doy + 2, 153);
    const d = doy - @divFloor(153 * mp + 2, 5) + 1;
    const m = mp + @as(i64, if (mp < 10) 3 else -9);
    if (m <= 2) y += 1;
    return std.fmt.allocPrint(a, "{d:0>4}-{d:0>2}-{d:0>2}", .{ y, m, d });
}
fn run(ctx: *Context, args: []const []const u8) !void {
    const a = ctx.allocator;
    if (args.len < 2) {
        try ctx.stdout.writeAll("usage: lvt <workflow> <analyze|fit|build|replay|phase> [options]\n");
        for (workflows) |w| try ctx.stdout.print("{s}\n", .{w});
        return;
    }
    for (args) |s| if (eq(s, "--help")) {
        try run(ctx, &.{});
        return;
    };
    const slug = args[0];
    const command = args[1];
    var known = false;
    for (workflows) |w| known = known or eq(w, slug);
    if (!known) return error.UnknownWorkflow;
    const defaults = .{ .window = "200", .@"min-period" = "5", .@"max-period" = "80", .smoothing = "1600", .@"genuine-threshold" = "49", .@"phase-cycles" = "2", .@"top-k" = "3", .@"peak-distance" = "2", .@"max-segments" = "5", .@"candidate-pool" = "10", .@"max-components" = "7", .horizon = "0", .@"short-min" = "5", .@"short-max" = "30", .@"long-min" = "31", .@"long-max" = "80", .normalization = "rms", .@"calendar-policy" = "extend" };
    inline for (@typeInfo(@TypeOf(defaults)).@"struct".fields) |f| try ctx.options.put(f.name, @field(defaults, f.name));
    if (eq(slug, "annual-cycle-pattern-decoder")) try ctx.options.put("max-period", "125");
    try ctx.options.put("as-of", try today(a, ctx.io));
    const flags = [_][]const u8{ "rolling", "invert", "confirm-complete-year", "confirm-complete-years" };
    const extra = [_][]const u8{ "input", "calendar", "model", "models", "history", "json-out", "model-out", "wave-out", "series-out", "start", "end", "year", "target-year", "radians" };
    var i: usize = 2;
    while (i < args.len) : (i += 1) {
        if (!std.mem.startsWith(u8, args[i], "--")) return error.InvalidOption;
        const key = args[i][2..];
        var is_flag = false;
        for (flags) |f| is_flag = is_flag or eq(f, key);
        if (is_flag) {
            try ctx.options.put(key, "true");
            continue;
        }
        if (eq(key, "models")) {
            if (i + 6 >= args.len) return error.SixModelPathsRequired;
            for (args[i + 1 .. i + 7]) |p| if (std.mem.startsWith(u8, p, "--")) return error.SixModelPathsRequired;
            try ctx.options.put(key, try std.mem.join(a, "\x00", args[i + 1 .. i + 7]));
            i += 6;
            continue;
        }
        var valid = ctx.options.contains(key);
        for (extra) |e| valid = valid or eq(e, key);
        if (!valid) return error.UnknownOption;
        i += 1;
        if (i >= args.len or std.mem.startsWith(u8, args[i], "--")) return error.MissingOptionValue;
        try ctx.options.put(key, args[i]);
    }
    if (eq(slug, "cycle-phase-score") and eq(command, "phase")) {
        const p = try ctx.real("radians");
        try ctx.print(try value(a, .{ .value = try harmonic.phaseState(a, p), .status = phaseStatus(p) }));
        return;
    }
    if (eq(slug, "annual-cycle-pattern-decoder")) {
        var model: annual.Model = undefined;
        if (eq(command, "fit")) {
            const p = annual.Params{ .min_period = try ctx.integer("min-period"), .max_period = try ctx.integer("max-period"), .smoothing = try ctx.real("smoothing"), .genuine_threshold = try ctx.real("genuine-threshold"), .candidate_pool = try ctx.integer("candidate-pool"), .max_components = try ctx.integer("max-components"), .peak_distance = try ctx.integer("peak-distance") };
            const calendar = if (ctx.text("calendar").len > 0) try ctx.calendar(ctx.text("calendar")) else null;
            model = try ctx.fit(try ctx.prices(try ctx.required("input")), try ctx.integer("year"), calendar, ctx.flag("confirm-complete-year"), p);
            _ = try ctx.required("wave-out");
            try ctx.saveJson(try ctx.required("model-out"), model);
        } else if (eq(command, "replay")) {
            model = try ctx.loadAnnual(try ctx.required("model"));
        } else return error.InvalidAnnualCommand;
        const dates = model.dates orelse return error.ModelDatesRequired;
        const values = try a.alloc(f64, dates.len);
        for (values, 0..) |*v, j| v.* = try annual.annualWave(a, model, @floatFromInt(j));
        try ctx.saveWave(try ctx.required("wave-out"), dates, values);
        try ctx.print(.{ .year = model.year, .correlation = model.correlation, .components = model.components.len, .sample_count = values.len });
        return;
    }
    if (eq(slug, "harmonic-energy-superposition-wave")) {
        var result: superposition.Result = undefined;
        if (eq(command, "build")) {
            const dates = try ctx.calendar(try ctx.required("calendar"));
            const year = try ctx.integer("target-year");
            try annual.validateDates(dates, year);
            if ((ctx.text("models").len > 0) == (ctx.text("history").len > 0)) return error.ModelsOrHistoryRequired;
            var models: [6]annual.Model = undefined;
            if (ctx.text("models").len > 0) {
                var paths = std.mem.splitScalar(u8, ctx.text("models"), 0);
                for (&models) |*m| m.* = try ctx.loadAnnual(paths.next() orelse return error.SixModelPathsRequired);
            } else {
                const records = try ctx.prices(ctx.text("history"));
                for (superposition.metonic_lags, 0..) |lag, j| models[j] = try ctx.fit(records, year - lag, null, ctx.flag("confirm-complete-years"), .{});
            }
            result = try superposition.superpositionValues(a, &models, year, dates.len, ctx.text("normalization"), ctx.text("calendar-policy"), ctx.flag("invert"));
            result.model.dates = dates;
            _ = try ctx.required("wave-out");
            try ctx.saveJson(try ctx.required("model-out"), result.model);
        } else if (eq(command, "replay")) {
            const parsed = try std.json.parseFromSlice(superposition.Model, a, try ctx.read(try ctx.required("model")), .{ .ignore_unknown_fields = true });
            try parsed.value.validate(a);
            result = try superposition.superpositionValues(a, &parsed.value.annual_models, parsed.value.target_year, parsed.value.sample_count, parsed.value.normalization, parsed.value.calendar_policy, parsed.value.invert);
            result.model.dates = parsed.value.dates;
        } else return error.InvalidSuperpositionCommand;
        try ctx.saveWave(try ctx.required("wave-out"), result.model.dates orelse return error.ModelDatesRequired, result.values);
        try ctx.print(.{ .target_year = result.model.target_year, .source_years = result.model.source_years, .sample_count = result.values.len });
        return;
    }
    if (!eq(command, "analyze")) return error.InvalidAnalyzeCommand;
    for ([_][]const u8{ "start", "end" }) |k| if (ctx.text(k).len > 0) try annual.validateDates(&.{ctx.text(k)}, null);
    if (ctx.text("start").len > 0 and ctx.text("end").len > 0 and std.mem.order(u8, ctx.text("start"), ctx.text("end")) == .gt) return error.InvalidRange;
    const records = try ctx.prices(try ctx.required("input"));
    var prices: std.ArrayList(f64) = .empty;
    var dates: std.ArrayList([]const u8) = .empty;
    for (records) |r| {
        if ((ctx.text("start").len > 0 and std.mem.order(u8, r.date, ctx.text("start")) == .lt) or (ctx.text("end").len > 0 and std.mem.order(u8, r.date, ctx.text("end")) == .gt)) continue;
        try prices.append(a, r.close);
        try dates.append(a, r.date);
    }
    if (prices.items.len < 6) return error.InsufficientPrices;
    const horizon = try ctx.integer("horizon");
    if (horizon < 0) return error.InvalidHorizon;
    const c = scanner.ScanConfig{ .window = if (ctx.flag("rolling")) try ctx.integer("window") else @intCast(prices.items.len), .min_period = try ctx.integer("min-period"), .max_period = try ctx.integer("max-period"), .smoothing = try ctx.real("smoothing"), .genuine_threshold = try ctx.real("genuine-threshold"), .phase_cycles = try ctx.real("phase-cycles"), .top_k = try ctx.integer("top-k"), .peak_distance = try ctx.integer("peak-distance"), .max_segments = try ctx.integer("max-segments") };
    try c.validate(a);
    var short = c;
    short.min_period = try ctx.integer("short-min");
    short.max_period = try ctx.integer("short-max");
    short.top_k = 1;
    var long = c;
    long.min_period = try ctx.integer("long-min");
    long.max_period = try ctx.integer("long-max");
    long.top_k = 1;
    if (eq(slug, "phasing-oscillator")) {
        if (short.max_period >= long.min_period) return error.OverlappingPhasingRanges;
        try short.validate(a);
        try long.validate(a);
    }
    const scan_ = try ctx.scan(prices.items, c);
    const window = @min(prices.items.len, @as(usize, @intCast(c.window)));
    var report = try ctx.report(scan_, c, dates.items[dates.items.len - window ..]);
    try report.object.put(a, "output", try ctx.calculate(prices.items, slug, c, short, long));
    if (ctx.text("series-out").len > 0) {
        var rows = std.json.Array.init(a);
        for (dates.items, 0..) |d, j| {
            var row: Value = undefined;
            if (ctx.flag("rolling") or eq(slug, "phasing-oscillator")) {
                row = try ctx.calculate(prices.items[0 .. j + 1], slug, c, short, long);
            } else {
                row = try value(a, .{ .fitted = try scanner.modeledValue(a, scan_.cycles[0], @as(f64, @floatFromInt(j)) - @as(f64, @floatFromInt(prices.items.len - 1))) });
                if (eq(slug, "dynamic-cycle-explorer")) {
                    const p = scan_.cycles[0].period;
                    const h = if (std.math.isFinite(p)) (try explorer.centeredMean(a, prices.items, @intCast(j), @max(2, @as(isize, @intFromFloat(@floor(p / 2 + 0.5)))))) - (try explorer.centeredMean(a, prices.items, @intCast(j), @intFromFloat(p))) else std.math.nan(f64);
                    try row.object.put(a, "highlighter", number(h));
                }
            }
            try row.object.put(a, "date", .{ .string = d });
            try rows.append(row);
        }
        try saveSeries(ctx, ctx.text("series-out"), rows.items);
    }
    var side: [2]scanner.RankedCycle = undefined;
    if (eq(slug, "phasing-oscillator")) {
        const x = try ctx.scan(prices.items, short);
        const y = try ctx.scan(prices.items, long);
        side = .{ x.cycles[0], y.cycles[0] };
        try report.object.put(a, "short_scan", try ctx.report(x, short, dates.items[dates.items.len - window ..]));
        try report.object.put(a, "long_scan", try ctx.report(y, long, dates.items[dates.items.len - window ..]));
        try report.object.put(a, "phasing_config", try value(a, .{ .short = short, .long = long }));
    }
    var projections = std.json.Array.init(a);
    var h: isize = 1;
    while (h <= horizon) : (h += 1) {
        const offset: @TypeOf(@as(f64, 0)) = @floatFromInt(h);
        const v = scan_.cycles[0];
        if (eq(slug, "phasing-oscillator")) {
            const x = try harmonic.phaseState(a, side[0].phase + 2 * std.math.pi * offset / side[0].period);
            const y = try harmonic.phaseState(a, side[1].phase + 2 * std.math.pi * offset / side[1].period);
            try projections.append(try value(a, .{ .bars_ahead = h, .value = x + y, .short_score = x, .long_score = y }));
        } else {
            try projections.append(try value(a, .{ .bars_ahead = h, .value = if (eq(slug, "cycle-phase-score")) try harmonic.phaseState(a, v.phase + 2 * std.math.pi * offset / v.period) else try scanner.modeledValue(a, v, offset) }));
        }
    }
    try report.object.put(a, "projection", .{ .array = projections });
    if (ctx.text("json-out").len > 0) try ctx.saveJson(ctx.text("json-out"), report);
    try ctx.print(report);
}
fn csvField(writer: *std.Io.Writer, v: Value) !void {
    switch (v) {
        .null => {},
        .string => |s| {
            try writer.writeByte('"');
            for (s) |c| {
                try writer.writeByte(c);
                if (c == '"') try writer.writeByte('"');
            }
            try writer.writeByte('"');
        },
        else => try std.json.Stringify.value(v, .{}, writer),
    }
}
fn saveSeries(ctx: *Context, path: []const u8, rows: []Value) !void {
    if (rows.len == 0) return error.EmptySeries;
    const a = ctx.allocator;
    var keys: std.ArrayList([]const u8) = .empty;
    for (rows[0].object.keys()) |k| if (!eq(k, "date")) try keys.append(a, k);
    std.mem.sort([]const u8, keys.items, {}, struct {
        fn less(_: void, x: []const u8, y: []const u8) bool {
            return std.mem.order(u8, x, y) == .lt;
        }
    }.less);
    var out = std.Io.Writer.Allocating.init(a);
    try out.writer.writeAll("date");
    for (keys.items) |k| try out.writer.print(",{s}", .{k});
    try out.writer.writeByte('\n');
    for (rows) |r| {
        try csvField(&out.writer, r.object.get("date") orelse .null);
        for (keys.items) |k| {
            try out.writer.writeByte(',');
            try csvField(&out.writer, r.object.get(k) orelse .null);
        }
        try out.writer.writeByte('\n');
    }
    try ctx.save(path, try out.toOwnedSlice());
}
pub fn main(init: std.process.Init) !void {
    var arena = std.heap.ArenaAllocator.init(init.gpa);
    defer arena.deinit();
    const a = arena.allocator();
    var args: std.ArrayList([]const u8) = .empty;
    var iter = init.minimal.args.iterate();
    _ = iter.next();
    while (iter.next()) |s| try args.append(a, s);
    var buffer: [8192]u8 = undefined;
    var output = std.Io.File.Writer.init(std.Io.File.stdout(), init.io, &buffer);
    var ctx = Context{ .allocator = a, .io = init.io, .stdout = &output.interface, .options = std.StringHashMap([]const u8).init(a) };
    run(&ctx, args.items) catch |err| {
        std.debug.print("lvt: {s}\n", .{@errorName(err)});
        std.process.exit(1);
    };
    try output.interface.flush();
}
