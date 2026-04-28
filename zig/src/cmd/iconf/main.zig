/// iconf — command-line chart configuration generator.
///
/// Reads a JSON settings file containing indicator definitions,
/// creates indicator instances via the factory, runs embedded bar data
/// through them, and writes chart configuration files in JSON and
/// TypeScript formats.
///
/// Usage:
///     iconf <settings.json> <output-name>
const std = @import("std");
const indicators = @import("indicators");
const Bar = @import("bar").Bar;
const Scalar = @import("scalar").Scalar;

const Io = std.Io;
const Identifier = indicators.identifier.Identifier;
const Indicator = indicators.indicator.Indicator;
const OutputArray = indicators.indicator.OutputArray;
const OutputValue = indicators.indicator.OutputValue;
const Metadata = indicators.metadata.Metadata;
const Shape = indicators.shape.Shape;
const Pane = indicators.pane.Pane;
const Descriptor = indicators.descriptor.Descriptor;
const OutputDescriptor = indicators.OutputDescriptor;
const FactoryResult = indicators.factory.FactoryResult;

const line_colors = [_][]const u8{
    "#e41a1c", "#377eb8", "#4daf4a", "#984ea3", "#ff7f00",
    "#a65628", "#f781bf", "#999999", "#66c2a5", "#fc8d62",
};

const band_colors = [_][]const u8{
    "rgba(0,255,0,0.3)",   "rgba(0,0,255,0.3)",   "rgba(255,0,0,0.3)",
    "rgba(128,0,128,0.3)", "rgba(0,128,128,0.3)",
};

// --- Placement ---

const OutputPlacement = struct {
    indicator_idx: usize,
    output_idx: usize,
    mnemonic: []const u8,
    shape: Shape,
    pane: Pane,
};

// --- Chart data types ---

const ScalarVal = struct {
    time: []const u8,
    value: f64,
};

const BandVal = struct {
    time: []const u8,
    upper: f64,
    lower: f64,
};

const HeatmapVal = struct {
    time: []const u8,
    first: f64,
    last: f64,
    result: f64,
    min: f64,
    max: f64,
    values: []const f64,
};

const LineData = struct {
    name: []const u8,
    data: []const ScalarVal,
    indicator: usize,
    output: usize,
    color: []const u8,
    width: f64,
    dash: []const u8,
    interpolation: []const u8,
};

const BandData = struct {
    name: []const u8,
    data: []const BandVal,
    indicator: usize,
    output: usize,
    color: []const u8,
    legend_color: []const u8,
    interpolation: []const u8,
};

const HeatmapData = struct {
    name: []const u8,
    data: []const HeatmapVal,
    indicator: usize,
    output: usize,
    gradient: []const u8,
    invert_gradient: bool,
};

const PaneData = struct {
    height: []const u8,
    value_format: []const u8,
    value_margin_percentage_factor: f64,
    heatmap: ?HeatmapData,
    bands: []const BandData,
    line_areas: []const ScalarVal, // unused but present in output
    horizontals: []const ScalarVal, // unused but present in output
    lines: []const LineData,
};

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const io = init.io;

    // Parse command-line args
    var args_iter = init.minimal.args.iterate();
    _ = args_iter.next(); // skip program name
    const settings_path = args_iter.next() orelse {
        std.debug.print("usage: iconf <settings.json> <output-name>\n", .{});
        std.process.exit(1);
    };
    const output_arg = args_iter.next() orelse {
        std.debug.print("usage: iconf <settings.json> <output-name>\n", .{});
        std.process.exit(1);
    };

    // Strip .json or .ts suffix from output name
    const output_name = stripSuffix(stripSuffix(output_arg, ".ts"), ".json");

    const settings_data = Io.Dir.cwd().readFileAlloc(io, settings_path, allocator, .unlimited) catch |err| {
        std.debug.print("error reading settings file: {}\n", .{err});
        std.process.exit(1);
    };
    defer allocator.free(settings_data);

    // Set up stderr writer for status messages
    var stderr_buf: [4096]u8 = undefined;
    var stderr_writer = Io.File.Writer.init(Io.File.stderr(), io, &stderr_buf);
    const stderr = &stderr_writer.interface;

    // Parse settings JSON
    const parsed = std.json.parseFromSlice(std.json.Value, allocator, settings_data, .{}) catch |err| {
        std.debug.print("error parsing settings file: {}\n", .{err});
        std.process.exit(1);
    };
    defer parsed.deinit();

    const entries = switch (parsed.value) {
        .array => |a| a.items,
        else => {
            std.debug.print("error: settings file must be a JSON array\n", .{});
            std.process.exit(1);
        },
    };

    // Create indicators from settings
    var results = std.ArrayList(FactoryResult).empty;
    defer {
        for (results.items) |r| r.deinit(allocator);
        results.deinit(allocator);
    }

    for (entries) |entry| {
        const obj = switch (entry) {
            .object => |o| o,
            else => {
                std.debug.print("error: each settings entry must be a JSON object\n", .{});
                std.process.exit(1);
            },
        };

        const id_val = obj.get("identifier") orelse {
            std.debug.print("error: settings entry missing 'identifier' field\n", .{});
            std.process.exit(1);
        };
        const id_str = switch (id_val) {
            .string => |s| s,
            else => {
                std.debug.print("error: 'identifier' must be a string\n", .{});
                std.process.exit(1);
            },
        };

        const id = Identifier.fromStr(id_str) orelse {
            std.debug.print("error: unknown indicator '{s}'\n", .{id_str});
            std.process.exit(1);
        };

        const params_val = obj.get("params");
        const params_json: []const u8 = blk: {
            if (params_val) |pv| {
                switch (pv) {
                    .object => {
                        var aw: Io.Writer.Allocating = .init(allocator);
                        std.json.Stringify.value(pv, .{}, &aw.writer) catch {
                            std.debug.print("error: failed to stringify params\n", .{});
                            std.process.exit(1);
                        };
                        break :blk aw.toOwnedSlice() catch {
                            std.debug.print("error: out of memory\n", .{});
                            std.process.exit(1);
                        };
                    },
                    else => break :blk "{}",
                }
            } else {
                break :blk "{}";
            }
        };
        defer if (params_val) |pv| switch (pv) {
            .object => allocator.free(params_json),
            else => {},
        };

        const result = indicators.factory.create(allocator, id, params_json) catch |err| {
            std.debug.print("error creating indicator '{s}': {}\n", .{ id_str, err });
            std.process.exit(1);
        };

        results.append(allocator, result) catch {
            std.debug.print("error: out of memory\n", .{});
            std.process.exit(1);
        };
    }

    // Collect output placements from descriptor registry
    var placements = std.ArrayList(OutputPlacement).empty;
    defer placements.deinit(allocator);

    for (results.items, 0..) |r, i| {
        var meta: Metadata = undefined;
        r.indicator.metadata(&meta);

        const desc = indicators.descriptorOf(meta.identifier);
        if (desc) |d| {
            for (d.outputs, 0..) |od, j| {
                var mnemonic_buf: [64]u8 = undefined;
                const mnemonic: []const u8 = if (j < meta.outputs_len)
                    meta.outputs_buf[j].mnemonic
                else blk: {
                    const m = std.fmt.bufPrint(&mnemonic_buf, "out[{d}]", .{j}) catch "out[?]";
                    break :blk m;
                };
                // We need to dupe the mnemonic since metadata is stack-local
                const mnemonic_duped = allocator.dupe(u8, mnemonic) catch {
                    std.debug.print("error: out of memory\n", .{});
                    std.process.exit(1);
                };
                placements.append(allocator, .{
                    .indicator_idx = i,
                    .output_idx = j,
                    .mnemonic = mnemonic_duped,
                    .shape = od.shape,
                    .pane = od.pane,
                }) catch {
                    std.debug.print("error: out of memory\n", .{});
                    std.process.exit(1);
                };
            }
        } else {
            // Fallback: treat all outputs as Own/Scalar
            for (meta.outputs_buf[0..meta.outputs_len], 0..) |om, j| {
                const mnemonic_duped = allocator.dupe(u8, om.mnemonic) catch {
                    std.debug.print("error: out of memory\n", .{});
                    std.process.exit(1);
                };
                placements.append(allocator, .{
                    .indicator_idx = i,
                    .output_idx = j,
                    .mnemonic = mnemonic_duped,
                    .shape = om.shape,
                    .pane = .own,
                }) catch {
                    std.debug.print("error: out of memory\n", .{});
                    std.process.exit(1);
                };
            }
        }
    }
    defer for (placements.items) |p| allocator.free(p.mnemonic);

    // Run bars through indicators, collecting all outputs
    const bars = testBars();

    // For heatmap, we need to clone values since they get overwritten.
    // Store outputs: allOutputs[barIdx] is a slice of OutputArrays per indicator.
    // We need to store scalar values, band values, heatmap values per placement.

    // Build data arrays per placement
    const PlacementData = union(enum) {
        line: std.ArrayList(ScalarVal),
        band: std.ArrayList(BandVal),
        heatmap: std.ArrayList(HeatmapVal),
    };

    var placement_data = try allocator.alloc(PlacementData, placements.items.len);
    defer allocator.free(placement_data);

    for (placements.items, 0..) |p, pi| {
        placement_data[pi] = switch (p.shape) {
            .band => .{ .band = std.ArrayList(BandVal).empty },
            .heatmap => .{ .heatmap = std.ArrayList(HeatmapVal).empty },
            else => .{ .line = std.ArrayList(ScalarVal).empty },
        };
    }
    defer for (placement_data) |*pd| {
        switch (pd.*) {
            .line => |*l| l.deinit(allocator),
            .band => |*b| b.deinit(allocator),
            .heatmap => |*h| {
                for (h.items) |item| allocator.free(item.values);
                h.deinit(allocator);
            },
        }
    };

    // We also need OHLCV bars with time strings
    var ohlcv_times: [252][]const u8 = undefined;
    for (0..252) |i| {
        ohlcv_times[i] = formatDate(allocator, bars[i].time) catch {
            std.debug.print("error: out of memory\n", .{});
            std.process.exit(1);
        };
    }
    defer for (&ohlcv_times) |t| allocator.free(t);

    // Run each bar through all indicators and collect data
    for (&bars, 0..) |*bar, bar_idx| {
        for (results.items, 0..) |r, ind_idx| {
            const output = r.indicator.updateBar(bar);

            // For each placement belonging to this indicator, extract data
            for (placements.items, 0..) |p, pi| {
                if (p.indicator_idx != ind_idx) continue;
                if (p.output_idx >= output.slice().len) continue;

                const val = output.slice()[p.output_idx];
                const time_str = ohlcv_times[bar_idx];

                switch (p.shape) {
                    .scalar, .polyline => {
                        // Treat as line
                        switch (val) {
                            .scalar => |s| {
                                if (!std.math.isNan(s.value)) {
                                    placement_data[pi].line.append(allocator, .{
                                        .time = time_str,
                                        .value = s.value,
                                    }) catch {};
                                }
                            },
                            else => {},
                        }
                    },
                    .band => {
                        switch (val) {
                            .band => |b| {
                                if (!b.isEmpty()) {
                                    placement_data[pi].band.append(allocator, .{
                                        .time = time_str,
                                        .upper = b.upper,
                                        .lower = b.lower,
                                    }) catch {};
                                }
                            },
                            else => {},
                        }
                    },
                    .heatmap => {
                        switch (val) {
                            .heatmap => |h| {
                                if (!h.isEmpty()) {
                                    // Clone values since they get overwritten
                                    const values = allocator.dupe(f64, h.valuesSlice()) catch continue;
                                    placement_data[pi].heatmap.append(allocator, .{
                                        .time = time_str,
                                        .first = h.parameter_first,
                                        .last = h.parameter_last,
                                        .result = h.parameter_resolution,
                                        .min = h.value_min,
                                        .max = h.value_max,
                                        .values = values,
                                    }) catch {
                                        allocator.free(values);
                                    };
                                }
                            },
                            else => {},
                        }
                    },
                }
            }
        }
    }

    // Group placements into price pane and indicator panes
    var price_lines = std.ArrayList(LineData).empty;
    defer price_lines.deinit(allocator);
    var price_bands = std.ArrayList(BandData).empty;
    defer price_bands.deinit(allocator);

    // own_pane_map: indicator_idx -> index in indicator_panes
    var own_pane_map = std.AutoHashMap(usize, usize).init(allocator);
    defer own_pane_map.deinit();

    // Build indicator panes dynamically
    var ind_pane_lines = std.ArrayList(std.ArrayList(LineData)).empty;
    defer {
        for (ind_pane_lines.items) |*l| l.deinit(allocator);
        ind_pane_lines.deinit(allocator);
    }
    var ind_pane_bands = std.ArrayList(std.ArrayList(BandData)).empty;
    defer {
        for (ind_pane_bands.items) |*b| b.deinit(allocator);
        ind_pane_bands.deinit(allocator);
    }
    var ind_pane_heatmaps = std.ArrayList(?HeatmapData).empty;
    defer ind_pane_heatmaps.deinit(allocator);
    var ind_pane_heights = std.ArrayList([]const u8).empty;
    defer ind_pane_heights.deinit(allocator);
    var ind_pane_vmf = std.ArrayList(f64).empty;
    defer ind_pane_vmf.deinit(allocator);

    var color_idx: usize = 0;
    var band_color_idx: usize = 0;

    for (placements.items, 0..) |p, pi| {
        if (p.pane == .price and p.shape == .scalar) {
            const data = placement_data[pi].line.items;
            price_lines.append(allocator, .{
                .name = p.mnemonic,
                .data = data,
                .indicator = p.indicator_idx,
                .output = p.output_idx,
                .color = line_colors[color_idx % line_colors.len],
                .width = 1,
                .dash = "",
                .interpolation = "natural",
            }) catch {};
            color_idx += 1;
        } else if (p.pane == .price and p.shape == .band) {
            const data = placement_data[pi].band.items;
            const color = band_colors[band_color_idx % band_colors.len];
            price_bands.append(allocator, .{
                .name = p.mnemonic,
                .data = data,
                .indicator = p.indicator_idx,
                .output = p.output_idx,
                .color = color,
                .legend_color = color,
                .interpolation = "natural",
            }) catch {};
            band_color_idx += 1;
        } else {
            // Own pane
            const gop = own_pane_map.getOrPut(p.indicator_idx) catch continue;
            if (!gop.found_existing) {
                gop.value_ptr.* = ind_pane_lines.items.len;
                ind_pane_lines.append(allocator, std.ArrayList(LineData).empty) catch {};
                ind_pane_bands.append(allocator, std.ArrayList(BandData).empty) catch {};
                ind_pane_heatmaps.append(allocator, null) catch {};
                ind_pane_heights.append(allocator, "60") catch {};
                ind_pane_vmf.append(allocator, 0.01) catch {};
            }
            const idx = gop.value_ptr.*;

            switch (p.shape) {
                .heatmap => {
                    const data = placement_data[pi].heatmap.items;
                    ind_pane_heatmaps.items[idx] = .{
                        .name = p.mnemonic,
                        .data = data,
                        .indicator = p.indicator_idx,
                        .output = p.output_idx,
                        .gradient = "Viridis",
                        .invert_gradient = false,
                    };
                    ind_pane_heights.items[idx] = "120";
                    ind_pane_vmf.items[idx] = 0;
                },
                .band => {
                    const data = placement_data[pi].band.items;
                    const color = band_colors[band_color_idx % band_colors.len];
                    ind_pane_bands.items[idx].append(allocator, .{
                        .name = p.mnemonic,
                        .data = data,
                        .indicator = p.indicator_idx,
                        .output = p.output_idx,
                        .color = color,
                        .legend_color = color,
                        .interpolation = "natural",
                    }) catch {};
                    band_color_idx += 1;
                },
                else => {
                    // Scalar or unknown -> line
                    const data = placement_data[pi].line.items;
                    ind_pane_lines.items[idx].append(allocator, .{
                        .name = p.mnemonic,
                        .data = data,
                        .indicator = p.indicator_idx,
                        .output = p.output_idx,
                        .color = line_colors[color_idx % line_colors.len],
                        .width = 1,
                        .dash = "",
                        .interpolation = "natural",
                    }) catch {};
                    color_idx += 1;
                },
            }
        }
    }

    // --- Write JSON output ---
    {
        var aw: Io.Writer.Allocating = .init(allocator);
        const w = &aw.writer;
        try writeJsonConfigFull(w, &bars, &ohlcv_times, price_lines.items, price_bands.items, ind_pane_lines.items, ind_pane_bands.items, ind_pane_heatmaps.items, ind_pane_heights.items, ind_pane_vmf.items);

        const json_data = aw.toOwnedSlice() catch {
            std.debug.print("error: out of memory\n", .{});
            std.process.exit(1);
        };
        defer allocator.free(json_data);

        const json_path = try std.fmt.allocPrint(allocator, "{s}.json", .{output_name});
        defer allocator.free(json_path);

        const cwd = Io.Dir.cwd();
        const file = cwd.createFile(io, json_path, .{}) catch |err| {
            std.debug.print("error writing {s}: {}\n", .{ json_path, err });
            std.process.exit(1);
        };
        defer file.close(io);
        file.writePositionalAll(io, json_data, 0) catch |err| {
            std.debug.print("error writing {s}: {}\n", .{ json_path, err });
            std.process.exit(1);
        };

        try stderr.print("wrote {s}\n", .{json_path});
        try stderr.flush();
    }

    // --- Write TypeScript output ---
    {
        var aw: Io.Writer.Allocating = .init(allocator);
        const w = &aw.writer;

        // Extract base name from output path
        const base_name = baseName(output_name);

        try writeTsConfig(w, allocator, base_name, &bars, &ohlcv_times, price_lines.items, price_bands.items, ind_pane_lines.items, ind_pane_bands.items, ind_pane_heatmaps.items, ind_pane_heights.items, ind_pane_vmf.items);

        const ts_data = aw.toOwnedSlice() catch {
            std.debug.print("error: out of memory\n", .{});
            std.process.exit(1);
        };
        defer allocator.free(ts_data);

        const ts_path = try std.fmt.allocPrint(allocator, "{s}.ts", .{output_name});
        defer allocator.free(ts_path);

        const cwd2 = Io.Dir.cwd();
        const file = cwd2.createFile(io, ts_path, .{}) catch |err| {
            std.debug.print("error writing {s}: {}\n", .{ ts_path, err });
            std.process.exit(1);
        };
        defer file.close(io);
        file.writePositionalAll(io, ts_data, 0) catch |err| {
            std.debug.print("error writing {s}: {}\n", .{ ts_path, err });
            std.process.exit(1);
        };

        try stderr.print("wrote {s}\n", .{ts_path});
        try stderr.flush();
    }
}

// ── JSON output ──────────────────────────────────────────────────────────

// (writeJsonConfig removed — using writeJsonConfigFull instead)

fn writeJsonFloat(w: anytype, v: f64) !void {
    // Go's encoding/json uses strconv.AppendFloat with 'f' format and -1 precision
    // which gives shortest representation. Zig's {d} should be similar.
    // Check if it's an integer
    if (v == @floor(v) and @abs(v) < 1e15) {
        // Integer — print without decimal point (like Go)
        try w.print("{d}", .{@as(i64, @intFromFloat(v))});
    } else {
        try w.print("{d}", .{v});
    }
}

fn writeJsonString(w: anytype, s: []const u8) !void {
    try w.writeByte('"');
    try w.writeAll(s);
    try w.writeByte('"');
}

fn baseName(path: []const u8) []const u8 {
    var i = path.len;
    while (i > 0) {
        i -= 1;
        if (path[i] == '/' or path[i] == '\\') return path[i + 1 ..];
    }
    return path;
}

fn stripSuffix(s: []const u8, suffix: []const u8) []const u8 {
    if (s.len >= suffix.len and std.mem.eql(u8, s[s.len - suffix.len ..], suffix)) {
        return s[0 .. s.len - suffix.len];
    }
    return s;
}

fn formatDate(allocator: std.mem.Allocator, timestamp: i64) ![]const u8 {
    const es = std.time.epoch.EpochSeconds{ .secs = @intCast(timestamp) };
    const ed = es.getEpochDay();
    const yd = ed.calculateYearDay();
    const md = yd.calculateMonthDay();
    return std.fmt.allocPrint(allocator, "{d:4}-{d:02}-{d:02}", .{
        @as(u16, @intCast(yd.year)),
        @as(u8, @intFromEnum(md.month)),
        @as(u8, md.day_index + 1),
    });
}

fn sanitizeVarName(allocator: std.mem.Allocator, s: []const u8) ![]const u8 {
    var buf = std.ArrayList(u8).empty;
    defer buf.deinit(allocator);

    var capitalize = false;
    for (s, 0..) |c, i| {
        if ((c >= 'a' and c <= 'z') or (c >= 'A' and c <= 'Z')) {
            if (capitalize) {
                if (c >= 'a' and c <= 'z') {
                    try buf.append(allocator, c - 'a' + 'A');
                } else {
                    try buf.append(allocator, c);
                }
                capitalize = false;
            } else {
                try buf.append(allocator, c);
            }
        } else if (c >= '0' and c <= '9') {
            if (i == 0) {
                try buf.append(allocator, '_');
            }
            try buf.append(allocator, c);
            capitalize = false;
        } else {
            capitalize = true;
        }
    }

    return try allocator.dupe(u8, buf.items);
}

// ── TypeScript output ────────────────────────────────────────────────────

fn writeTsConfig(
    w: anytype,
    allocator: std.mem.Allocator,
    base_name: []const u8,
    bars: []const Bar,
    ohlcv_times: []const []const u8,
    price_lines: []const LineData,
    price_bands: []const BandData,
    ind_pane_lines: []const std.ArrayList(LineData),
    ind_pane_bands: []const std.ArrayList(BandData),
    ind_pane_heatmaps: []const ?HeatmapData,
    ind_pane_heights: []const []const u8,
    ind_pane_vmf: []const f64,
) !void {
    // Header
    try w.writeAll("// Auto-generated chart configuration.\n");
    try w.writeAll("// eslint-disable-next-line\n");
    try w.writeAll("import { Configuration } from '../ohlcv-chart/template/configuration';\n");
    try w.writeAll("import { Scalar, Band, Heatmap, Bar } from '../ohlcv-chart/template/types';\n\n");

    // OHLCV data array
    const ohlcv_var = try sanitizeVarName(allocator, base_name);
    defer allocator.free(ohlcv_var);
    const ohlcv_var_full = try std.fmt.allocPrint(allocator, "{s}Ohlcv", .{ohlcv_var});
    defer allocator.free(ohlcv_var_full);

    try w.print("export const {s}: Bar[] = ", .{ohlcv_var_full});
    try writeCompactOhlcvArray(w, bars, ohlcv_times);
    try w.writeAll(";\n\n");

    // Track variable names and counter
    var var_counter: usize = 0;

    // Collect line var names for price pane
    var price_line_vars = try allocator.alloc([]const u8, price_lines.len);
    defer {
        for (price_line_vars) |v| allocator.free(v);
        allocator.free(price_line_vars);
    }
    for (price_lines, 0..) |l, i| {
        var_counter += 1;
        const vn = try makeTsVarName(allocator, base_name, l.name, var_counter);
        price_line_vars[i] = vn;
        try w.print("export const {s}: Scalar[] = ", .{vn});
        try writeCompactScalarArray(w, l.data);
        try w.writeAll(";\n\n");
    }

    // Price pane band vars
    var price_band_vars = try allocator.alloc([]const u8, price_bands.len);
    defer {
        for (price_band_vars) |v| allocator.free(v);
        allocator.free(price_band_vars);
    }
    for (price_bands, 0..) |b, i| {
        var_counter += 1;
        const vn = try makeTsVarName(allocator, base_name, b.name, var_counter);
        price_band_vars[i] = vn;
        try w.print("export const {s}: Band[] = ", .{vn});
        try writeCompactBandArray(w, b.data);
        try w.writeAll(";\n\n");
    }

    // Indicator pane data arrays
    // For each indicator pane: lines, bands, heatmap
    const num_panes = ind_pane_lines.len;
    var ind_line_vars = try allocator.alloc([]const []const u8, num_panes);
    defer {
        for (ind_line_vars) |vars| {
            for (vars) |v| allocator.free(v);
            allocator.free(vars);
        }
        allocator.free(ind_line_vars);
    }
    var ind_band_vars = try allocator.alloc([]const []const u8, num_panes);
    defer {
        for (ind_band_vars) |vars| {
            for (vars) |v| allocator.free(v);
            allocator.free(vars);
        }
        allocator.free(ind_band_vars);
    }
    var ind_heatmap_vars = try allocator.alloc(?[]const u8, num_panes);
    defer {
        for (ind_heatmap_vars) |v| if (v) |s| allocator.free(s);
        allocator.free(ind_heatmap_vars);
    }

    for (0..num_panes) |pi| {
        // Lines
        const lines = ind_pane_lines[pi].items;
        var lvars = try allocator.alloc([]const u8, lines.len);
        for (lines, 0..) |l, li| {
            var_counter += 1;
            const vn = try makeTsVarName(allocator, base_name, l.name, var_counter);
            lvars[li] = vn;
            try w.print("export const {s}: Scalar[] = ", .{vn});
            try writeCompactScalarArray(w, l.data);
            try w.writeAll(";\n\n");
        }
        ind_line_vars[pi] = lvars;

        // Bands
        const bands = ind_pane_bands[pi].items;
        var bvars = try allocator.alloc([]const u8, bands.len);
        for (bands, 0..) |b, bi| {
            var_counter += 1;
            const vn = try makeTsVarName(allocator, base_name, b.name, var_counter);
            bvars[bi] = vn;
            try w.print("export const {s}: Band[] = ", .{vn});
            try writeCompactBandArray(w, b.data);
            try w.writeAll(";\n\n");
        }
        ind_band_vars[pi] = bvars;

        // Heatmap
        if (ind_pane_heatmaps[pi]) |hm| {
            var_counter += 1;
            const vn = try makeTsVarName(allocator, base_name, hm.name, var_counter);
            ind_heatmap_vars[pi] = vn;
            try w.print("export const {s}: Heatmap[] = ", .{vn});
            try writeCompactHeatmapArray(w, hm.data);
            try w.writeAll(";\n\n");
        } else {
            ind_heatmap_vars[pi] = null;
        }
    }

    // Emit configuration object
    const config_var = try sanitizeVarName(allocator, base_name);
    defer allocator.free(config_var);
    try w.print("export const {s}Config: Configuration = {{\n", .{config_var});
    try w.writeAll("  width: \"100%\",\n");
    try w.writeAll("  navigationPane: {\n");
    try w.writeAll("    height: 30, hasLine: true, hasArea: false, hasTimeAxis: true, timeTicks: 0,\n");
    try w.writeAll("  },\n");
    try w.writeAll("  heightNavigationPane: 30,\n");
    try w.writeAll("  timeAnnotationFormat: \"%Y-%m-%d\",\n");
    try w.writeAll("  axisLeft: true,\n");
    try w.writeAll("  axisRight: false,\n");
    try w.writeAll("  margin: { left: 0, top: 10, right: 20, bottom: 0 },\n");
    try w.print("  ohlcv: {{ name: \"TEST\", data: {s}, candlesticks: true }},\n", .{ohlcv_var_full});

    // Price pane
    try w.writeAll("  pricePane: {\n");
    try writeTsPane(w, "    ", "30%", ",.2f", 0.01, null, price_bands, price_band_vars, price_lines, price_line_vars, null);
    try w.writeAll("  },\n");

    // Indicator panes
    try w.writeAll("  indicatorPanes: [\n");
    for (0..num_panes) |pi| {
        try w.writeAll("    {\n");
        try writeTsPane(
            w,
            "      ",
            ind_pane_heights[pi],
            ",.2f",
            ind_pane_vmf[pi],
            if (ind_pane_heatmaps[pi]) |hm| &hm else null,
            ind_pane_bands[pi].items,
            ind_band_vars[pi],
            ind_pane_lines[pi].items,
            ind_line_vars[pi],
            ind_heatmap_vars[pi],
        );
        try w.writeAll("    },\n");
    }
    try w.writeAll("  ],\n");

    try w.writeAll("  crosshair: false,\n");
    try w.writeAll("  volumeInPricePane: true,\n");
    try w.writeAll("  menuVisible: true,\n");
    try w.writeAll("  downloadSvgVisible: true,\n");
    try w.writeAll("};\n");
}

fn writeTsPane(
    w: anytype,
    indent: []const u8,
    height: []const u8,
    value_format: []const u8,
    vmf: f64,
    heatmap: ?*const HeatmapData,
    bands: []const BandData,
    band_vars: []const []const u8,
    lines: []const LineData,
    line_vars: []const []const u8,
    heatmap_var: ?[]const u8,
) !void {
    try w.print("{s}height: \"{s}\", valueFormat: \"{s}\", valueMarginPercentageFactor: ", .{ indent, height, value_format });
    try writeGoG(w, vmf);
    try w.writeAll(",\n");

    // Heatmap (if any)
    if (heatmap) |hm| {
        try w.print("{s}heatmap: {{ name: \"{s}\", data: {s}, indicator: {d}, output: {d}, gradient: \"{s}\", invertGradient: {s} }},\n", .{
            indent,
            hm.name,
            heatmap_var.?,
            hm.indicator,
            hm.output,
            hm.gradient,
            if (hm.invert_gradient) "true" else "false",
        });
    }

    // Bands
    try w.print("{s}bands: [", .{indent});
    if (bands.len > 0) {
        try w.writeAll("\n");
        for (bands, 0..) |b, i| {
            try w.print("{s}  {{ name: \"{s}\", data: {s}, indicator: {d}, output: {d}, color: \"{s}\", legendColor: \"{s}\", interpolation: \"{s}\" }},\n", .{
                indent,
                b.name,
                band_vars[i],
                b.indicator,
                b.output,
                b.color,
                b.legend_color,
                b.interpolation,
            });
        }
        try w.print("{s}", .{indent});
    }
    try w.writeAll("],\n");

    // LineAreas
    try w.print("{s}lineAreas: [],\n", .{indent});

    // Horizontals
    try w.print("{s}horizontals: [],\n", .{indent});

    // Lines
    try w.print("{s}lines: [", .{indent});
    if (lines.len > 0) {
        try w.writeAll("\n");
        for (lines, 0..) |l, i| {
            try w.print("{s}  {{ name: \"{s}\", data: {s}, indicator: {d}, output: {d}, color: \"{s}\", width: ", .{
                indent,
                l.name,
                line_vars[i],
                l.indicator,
                l.output,
                l.color,
            });
            try writeGoG(w, l.width);
            try w.print(", dash: \"{s}\", interpolation: \"{s}\" }},\n", .{ l.dash, l.interpolation });
        }
        try w.print("{s}", .{indent});
    }
    try w.writeAll("],\n");

    // Arrows
    try w.print("{s}arrows: [],\n", .{indent});
}

/// Write a float using Go's %g format (shortest representation, no trailing zeros).
fn writeGoG(w: anytype, v: f64) !void {
    if (v == 0) {
        try w.writeAll("0");
    } else if (v == @floor(v) and @abs(v) < 1e15) {
        try w.print("{d}", .{@as(i64, @intFromFloat(v))});
    } else {
        try w.print("{d}", .{v});
    }
}

fn makeTsVarName(allocator: std.mem.Allocator, base_name: []const u8, mnemonic: []const u8, counter: usize) ![]const u8 {
    const raw = try std.fmt.allocPrint(allocator, "{s}_{s}_{d}", .{ base_name, mnemonic, counter });
    defer allocator.free(raw);
    return sanitizeVarName(allocator, raw);
}

fn writeCompactOhlcvArray(w: anytype, bars: []const Bar, times: []const []const u8) !void {
    try w.writeByte('[');
    for (bars, 0..) |bar, i| {
        if (i > 0) try w.writeByte(',');
        try w.writeAll("{\"time\":\"");
        try w.writeAll(times[i]);
        try w.writeAll("\",\"open\":");
        try writeJsonFloat(w, bar.open);
        try w.writeAll(",\"high\":");
        try writeJsonFloat(w, bar.high);
        try w.writeAll(",\"low\":");
        try writeJsonFloat(w, bar.low);
        try w.writeAll(",\"close\":");
        try writeJsonFloat(w, bar.close);
        try w.writeAll(",\"volume\":");
        try writeJsonFloat(w, bar.volume);
        try w.writeByte('}');
    }
    try w.writeByte(']');
}

fn writeCompactScalarArray(w: anytype, data: []const ScalarVal) !void {
    try w.writeByte('[');
    for (data, 0..) |sv, i| {
        if (i > 0) try w.writeByte(',');
        try w.writeAll("{\"time\":\"");
        try w.writeAll(sv.time);
        try w.writeAll("\",\"value\":");
        try writeJsonFloat(w, sv.value);
        try w.writeByte('}');
    }
    try w.writeByte(']');
}

fn writeCompactBandArray(w: anytype, data: []const BandVal) !void {
    try w.writeByte('[');
    for (data, 0..) |bv, i| {
        if (i > 0) try w.writeByte(',');
        try w.writeAll("{\"time\":\"");
        try w.writeAll(bv.time);
        try w.writeAll("\",\"upper\":");
        try writeJsonFloat(w, bv.upper);
        try w.writeAll(",\"lower\":");
        try writeJsonFloat(w, bv.lower);
        try w.writeByte('}');
    }
    try w.writeByte(']');
}

fn writeCompactHeatmapArray(w: anytype, data: []const HeatmapVal) !void {
    try w.writeByte('[');
    for (data, 0..) |hv, i| {
        if (i > 0) try w.writeByte(',');
        try w.writeAll("{\"time\":\"");
        try w.writeAll(hv.time);
        try w.writeAll("\",\"first\":");
        try writeJsonFloat(w, hv.first);
        try w.writeAll(",\"last\":");
        try writeJsonFloat(w, hv.last);
        try w.writeAll(",\"result\":");
        try writeJsonFloat(w, hv.result);
        try w.writeAll(",\"min\":");
        try writeJsonFloat(w, hv.min);
        try w.writeAll(",\"max\":");
        try writeJsonFloat(w, hv.max);
        try w.writeAll(",\"values\":[");
        for (hv.values, 0..) |v, vi| {
            if (vi > 0) try w.writeByte(',');
            try writeJsonFloat(w, v);
        }
        try w.writeAll("]}");
    }
    try w.writeByte(']');
}

// ── Full JSON output (proper 2-space indented, matching Go's json.MarshalIndent) ──

fn writeJsonConfigFull(
    w: anytype,
    bars: []const Bar,
    ohlcv_times: []const []const u8,
    price_lines: []const LineData,
    price_bands: []const BandData,
    ind_pane_lines: []const std.ArrayList(LineData),
    ind_pane_bands: []const std.ArrayList(BandData),
    ind_pane_heatmaps: []const ?HeatmapData,
    ind_pane_heights: []const []const u8,
    ind_pane_vmf: []const f64,
) !void {
    try w.writeAll("{\n");
    try w.writeAll("  \"width\": \"100%\",\n");
    try w.writeAll("  \"navigationPane\": {\n");
    try w.writeAll("    \"height\": 30,\n");
    try w.writeAll("    \"hasLine\": true,\n");
    try w.writeAll("    \"hasArea\": false,\n");
    try w.writeAll("    \"hasTimeAxis\": true,\n");
    try w.writeAll("    \"timeTicks\": 0\n");
    try w.writeAll("  },\n");
    try w.writeAll("  \"heightNavigationPane\": 30,\n");
    try w.writeAll("  \"timeAnnotationFormat\": \"%Y-%m-%d\",\n");
    try w.writeAll("  \"axisLeft\": true,\n");
    try w.writeAll("  \"axisRight\": false,\n");
    try w.writeAll("  \"margin\": {\n");
    try w.writeAll("    \"left\": 0,\n");
    try w.writeAll("    \"top\": 10,\n");
    try w.writeAll("    \"right\": 20,\n");
    try w.writeAll("    \"bottom\": 0\n");
    try w.writeAll("  },\n");

    // OHLCV
    try w.writeAll("  \"ohlcv\": {\n");
    try w.writeAll("    \"name\": \"TEST\",\n");
    try w.writeAll("    \"data\": [\n");
    for (bars, 0..) |bar, i| {
        try w.writeAll("      {\n");
        try w.print("        \"time\": \"{s}\",\n", .{ohlcv_times[i]});
        try writeJsonField(w, "        ", "open", bar.open, true);
        try writeJsonField(w, "        ", "high", bar.high, true);
        try writeJsonField(w, "        ", "low", bar.low, true);
        try writeJsonField(w, "        ", "close", bar.close, true);
        try writeJsonField(w, "        ", "volume", bar.volume, false);
        if (i < bars.len - 1) {
            try w.writeAll("      },\n");
        } else {
            try w.writeAll("      }\n");
        }
    }
    try w.writeAll("    ],\n");
    try w.writeAll("    \"candlesticks\": true\n");
    try w.writeAll("  },\n");

    // Price pane
    try w.writeAll("  \"pricePane\": {\n");
    try w.writeAll("    \"height\": \"30%\",\n");
    try w.writeAll("    \"valueFormat\": \",.2f\",\n");
    try w.writeAll("    \"valueMarginPercentageFactor\": 0.01,\n");
    try writeJsonBands(w, "    ", price_bands, true);
    try writeJsonEmptyArray(w, "    ", "lineAreas", true);
    try writeJsonEmptyArray(w, "    ", "horizontals", true);
    try writeJsonLines(w, "    ", price_lines, true);
    try writeJsonEmptyArray(w, "    ", "arrows", false);
    try w.writeAll("  },\n");

    // Indicator panes
    try w.writeAll("  \"indicatorPanes\": [\n");
    for (0..ind_pane_lines.len) |pi| {
        try w.writeAll("    {\n");
        try w.print("      \"height\": \"{s}\",\n", .{ind_pane_heights[pi]});
        try w.writeAll("      \"valueFormat\": \",.2f\",\n");
        try w.print("      \"valueMarginPercentageFactor\": ", .{});
        try writeJsonFloat(w, ind_pane_vmf[pi]);
        try w.writeAll(",\n");

        // Heatmap
        if (ind_pane_heatmaps[pi]) |hm| {
            try w.writeAll("      \"heatmap\": {\n");
            try w.print("        \"name\": \"{s}\",\n", .{hm.name});
            try w.writeAll("        \"data\": [\n");
            for (hm.data, 0..) |hv, hi| {
                try w.writeAll("          {\n");
                try w.print("            \"time\": \"{s}\",\n", .{hv.time});
                try writeJsonField(w, "            ", "first", hv.first, true);
                try writeJsonField(w, "            ", "last", hv.last, true);
                try writeJsonField(w, "            ", "result", hv.result, true);
                try writeJsonField(w, "            ", "min", hv.min, true);
                try writeJsonField(w, "            ", "max", hv.max, true);
                // values array
                try w.writeAll("            \"values\": [\n");
                for (hv.values, 0..) |v, vi| {
                    try w.writeAll("              ");
                    try writeJsonFloat(w, v);
                    if (vi < hv.values.len - 1) {
                        try w.writeAll(",\n");
                    } else {
                        try w.writeAll("\n");
                    }
                }
                try w.writeAll("            ]\n");
                if (hi < hm.data.len - 1) {
                    try w.writeAll("          },\n");
                } else {
                    try w.writeAll("          }\n");
                }
            }
            try w.writeAll("        ],\n");
            try w.print("        \"indicator\": {d},\n", .{hm.indicator});
            try w.print("        \"output\": {d},\n", .{hm.output});
            try w.print("        \"gradient\": \"{s}\",\n", .{hm.gradient});
            try w.print("        \"invertGradient\": {s}\n", .{if (hm.invert_gradient) "true" else "false"});
            try w.writeAll("      },\n");
        }

        try writeJsonBands(w, "      ", ind_pane_bands[pi].items, true);
        try writeJsonEmptyArray(w, "      ", "lineAreas", true);
        try writeJsonEmptyArray(w, "      ", "horizontals", true);
        try writeJsonLines(w, "      ", ind_pane_lines[pi].items, true);
        try writeJsonEmptyArray(w, "      ", "arrows", false);

        if (pi < ind_pane_lines.len - 1) {
            try w.writeAll("    },\n");
        } else {
            try w.writeAll("    }\n");
        }
    }
    try w.writeAll("  ],\n");

    try w.writeAll("  \"crosshair\": false,\n");
    try w.writeAll("  \"volumeInPricePane\": true,\n");
    try w.writeAll("  \"menuVisible\": true,\n");
    try w.writeAll("  \"downloadSvgVisible\": true\n");
    try w.writeAll("}");
}

fn writeJsonField(w: anytype, indent: []const u8, name: []const u8, value: f64, comma: bool) !void {
    try w.print("{s}\"{s}\": ", .{ indent, name });
    try writeJsonFloat(w, value);
    if (comma) try w.writeByte(',');
    try w.writeByte('\n');
}

fn writeJsonBands(w: anytype, indent: []const u8, bands: []const BandData, comma: bool) !void {
    if (bands.len == 0) {
        try w.print("{s}\"bands\": []", .{indent});
        if (comma) try w.writeByte(',');
        try w.writeByte('\n');
        return;
    }
    try w.print("{s}\"bands\": [\n", .{indent});
    for (bands, 0..) |b, i| {
        try w.print("{s}  {{\n", .{indent});
        try w.print("{s}    \"name\": \"{s}\",\n", .{ indent, b.name });
        try w.print("{s}    \"data\": [\n", .{indent});
        for (b.data, 0..) |bv, bi| {
            try w.print("{s}      {{\n", .{indent});
            try w.print("{s}        \"time\": \"{s}\",\n", .{ indent, bv.time });
            try w.print("{s}        \"upper\": ", .{indent});
            try writeJsonFloat(w, bv.upper);
            try w.writeAll(",\n");
            try w.print("{s}        \"lower\": ", .{indent});
            try writeJsonFloat(w, bv.lower);
            try w.writeByte('\n');
            if (bi < b.data.len - 1) {
                try w.print("{s}      }},\n", .{indent});
            } else {
                try w.print("{s}      }}\n", .{indent});
            }
        }
        try w.print("{s}    ],\n", .{indent});
        try w.print("{s}    \"indicator\": {d},\n", .{ indent, b.indicator });
        try w.print("{s}    \"output\": {d},\n", .{ indent, b.output });
        try w.print("{s}    \"color\": \"{s}\",\n", .{ indent, b.color });
        try w.print("{s}    \"legendColor\": \"{s}\",\n", .{ indent, b.legend_color });
        try w.print("{s}    \"interpolation\": \"{s}\"\n", .{ indent, b.interpolation });
        if (i < bands.len - 1) {
            try w.print("{s}  }},\n", .{indent});
        } else {
            try w.print("{s}  }}\n", .{indent});
        }
    }
    try w.print("{s}]", .{indent});
    if (comma) try w.writeByte(',');
    try w.writeByte('\n');
}

fn writeJsonLines(w: anytype, indent: []const u8, lines: []const LineData, comma: bool) !void {
    if (lines.len == 0) {
        try w.print("{s}\"lines\": []", .{indent});
        if (comma) try w.writeByte(',');
        try w.writeByte('\n');
        return;
    }
    try w.print("{s}\"lines\": [\n", .{indent});
    for (lines, 0..) |l, i| {
        try w.print("{s}  {{\n", .{indent});
        try w.print("{s}    \"name\": \"{s}\",\n", .{ indent, l.name });
        try w.print("{s}    \"data\": [\n", .{indent});
        for (l.data, 0..) |sv, si| {
            try w.print("{s}      {{\n", .{indent});
            try w.print("{s}        \"time\": \"{s}\",\n", .{ indent, sv.time });
            try w.print("{s}        \"value\": ", .{indent});
            try writeJsonFloat(w, sv.value);
            try w.writeByte('\n');
            if (si < l.data.len - 1) {
                try w.print("{s}      }},\n", .{indent});
            } else {
                try w.print("{s}      }}\n", .{indent});
            }
        }
        try w.print("{s}    ],\n", .{indent});
        try w.print("{s}    \"indicator\": {d},\n", .{ indent, l.indicator });
        try w.print("{s}    \"output\": {d},\n", .{ indent, l.output });
        try w.print("{s}    \"color\": \"{s}\",\n", .{ indent, l.color });
        try w.print("{s}    \"width\": ", .{indent});
        try writeJsonFloat(w, l.width);
        try w.writeAll(",\n");
        try w.print("{s}    \"dash\": \"{s}\",\n", .{ indent, l.dash });
        try w.print("{s}    \"interpolation\": \"{s}\"\n", .{ indent, l.interpolation });
        if (i < lines.len - 1) {
            try w.print("{s}  }},\n", .{indent});
        } else {
            try w.print("{s}  }}\n", .{indent});
        }
    }
    try w.print("{s}]", .{indent});
    if (comma) try w.writeByte(',');
    try w.writeByte('\n');
}

fn writeJsonEmptyArray(w: anytype, indent: []const u8, name: []const u8, comma: bool) !void {
    try w.print("{s}\"{s}\": []", .{ indent, name });
    if (comma) try w.writeByte(',');
    try w.writeByte('\n');
}

// ── Embedded test data ─────────────────────────────────────────────────────

fn testBars() [252]Bar {
    const highs = testHighs();
    const lows = testLows();
    const closes = testCloses();
    const volumes = testVolumes();

    var bars: [252]Bar = undefined;
    const base_time: i64 = 1577923200; // 2020-01-02T00:00:00Z
    const day_seconds: i64 = 86400;

    for (0..252) |i| {
        const open_price = if (i > 0) closes[i - 1] else closes[0];
        bars[i] = .{
            .time = base_time + @as(i64, @intCast(i)) * day_seconds,
            .open = open_price,
            .high = highs[i],
            .low = lows[i],
            .close = closes[i],
            .volume = volumes[i],
        };
    }

    return bars;
}

fn testHighs() [252]f64 {
    return .{
        93.25,  94.94,  96.375,  96.19,   96,      94.72,  95,     93.72,   92.47,   92.75,
        96.25,  99.625, 99.125,  92.75,   91.315,  93.25,  93.405, 90.655,  91.97,   92.25,
        90.345, 88.5,   88.25,   85.5,    84.44,   84.75,  84.44,  89.405,  88.125,  89.125,
        87.155, 87.25,  87.375,  88.97,   90,      89.845, 86.97,  85.94,   84.75,   85.47,
        84.47,  88.5,   89.47,   90,      92.44,   91.44,  92.97,  91.72,   91.155,  91.75,
        90,     88.875, 89,      85.25,   83.815,  85.25,  86.625, 87.94,   89.375,  90.625,
        90.75,  88.845, 91.97,   93.375,  93.815,  94.03,  94.03,  91.815,  92,      91.94,
        89.75,  88.75,  86.155,  84.875,  85.94,   99.375, 103.28, 105.375, 107.625, 105.25,
        104.5,  105.5,  106.125, 107.94,  106.25,  107,    108.75, 110.94,  110.94,  114.22,
        123,    121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113,     118.315,
        116.87, 116.75, 113.87,  114.62,  115.31,  116,    121.69, 119.87,  120.87,  116.75,
        116.5,  116,    118.31,  121.5,   122,     121.44, 125.75, 127.75,  124.19,  124.44,
        125.75, 124.69, 125.31,  132,     131.31,  132.25, 133.88, 133.5,   135.5,   137.44,
        138.69, 139.19, 138.5,   138.13,  137.5,   138.88, 132.13, 129.75,  128.5,   125.44,
        125.12, 126.5,  128.69,  126.62,  126.69,  126,    123.12, 121.87,  124,     127,
        124.44, 122.5,  123.75,  123.81,  124.5,   127.87, 128.56, 129.63,  124.87,  124.37,
        124.87, 123.62, 124.06,  125.87,  125.19,  125.62, 126,    128.5,   126.75,  129.75,
        132.69, 133.94, 136.5,   137.69,  135.56,  133.56, 135,    132.38,  131.44,  130.88,
        129.63, 127.25, 127.81,  125,     126.81,  124.75, 122.81, 122.25,  121.06,  120,
        123.25, 122.75, 119.19,  115.06,  116.69,  114.87, 110.87, 107.25,  108.87,  109,
        108.5,  113.06, 93,      94.62,   95.12,   96,     95.56,  95.31,   99,      98.81,
        96.81,  95.94,  94.44,   92.94,   93.94,   95.5,   97.06,  97.5,    96.25,   96.37,
        95,     94.87,  98.25,   105.12,  108.44,  109.87, 105,    106,     104.94,  104.5,
        104.44, 106.31, 112.87,  116.5,   119.19,  121,    122.12, 111.94,  112.75,  110.19,
        107.94, 109.69, 111.06,  110.44,  110.12,  110.31, 110.44, 110,     110.75,  110.5,
        110.5,  109.5,
    };
}

fn testLows() [252]f64 {
    return .{
        90.75,  91.405, 94.25,   93.5,   92.815,  93.5,   92,      89.75,   89.44,  90.625,
        92.75,  96.315, 96.03,   88.815, 86.75,   90.94,  88.905,  88.78,   89.25,  89.75,
        87.5,   86.53,  84.625,  82.28,  81.565,  80.875, 81.25,   84.065,  85.595, 85.97,
        84.405, 85.095, 85.5,    85.53,  87.875,  86.565, 84.655,  83.25,   82.565, 83.44,
        82.53,  85.065, 86.875,  88.53,  89.28,   90.125, 90.75,   89,      88.565, 90.095,
        89,     86.47,  84,      83.315, 82,      83.25,  84.75,   85.28,   87.19,  88.44,
        88.25,  87.345, 89.28,   91.095, 89.53,   91.155, 92,      90.53,   89.97,  88.815,
        86.75,  85.065, 82.03,   81.5,   82.565,  96.345, 96.47,   101.155, 104.25, 101.75,
        101.72, 101.72, 103.155, 105.69, 103.655, 104,    105.53,  108.53,  108.75, 107.75,
        117,    118,    116,     118.5,  116.53,  116.25, 114.595, 110.875, 110.5,  110.72,
        112.62, 114.19, 111.19,  109.44, 111.56,  112.44, 117.5,   116.06,  116.56, 113.31,
        112.56, 114,    114.75,  118.87, 119,     119.75, 122.62,  123,     121.75, 121.56,
        123.12, 122.19, 122.75,  124.37, 128,     129.5,  130.81,  130.63,  132.13, 133.88,
        135.38, 135.75, 136.19,  134.5,  135.38,  133.69, 126.06,  126.87,  123.5,  122.62,
        122.75, 123.56, 125.81,  124.62, 124.37,  121.81, 118.19,  118.06,  117.56, 121,
        121.12, 118.94, 119.81,  121,    122,     124.5,  126.56,  123.5,   121.25, 121.06,
        122.31, 121,    120.87,  122.06, 122.75,  122.69, 122.87,  125.5,   124.25, 128,
        128.38, 130.69, 131.63,  134.38, 132,     131.94, 131.94,  129.56,  123.75, 126,
        126.25, 124.37, 121.44,  120.44, 121.37,  121.69, 120,     119.62,  115.5,  116.75,
        119.06, 119.06, 115.06,  111.06, 113.12,  110,    105,     104.69,  103.87, 104.69,
        105.44, 107,    89,      92.5,   92.12,   94.62,  92.81,   94.25,   96.25,  96.37,
        93.69,  93.5,   90,      90.19,  90.5,    92.12,  94.12,   94.87,   93,     93.87,
        93,     92.62,  93.56,   98.37,  104.44,  106,    101.81,  104.12,  103.37, 102.12,
        102.25, 103.37, 107.94,  112.5,  115.44,  115.5,  112.25,  107.56,  106.56, 106.87,
        104.5,  105.75, 108.62,  107.75, 108.06,  108,    108.19,  108.12,  109.06, 108.75,
        108.56, 106.62,
    };
}

fn testCloses() [252]f64 {
    return .{
        91.5,    94.815,  94.375,  95.095, 93.78,   94.625,  92.53,   92.75,   90.315,  92.47,
        96.125,  97.25,   98.5,    89.875, 91,      92.815,  89.155,  89.345,  91.625,  89.875,
        88.375,  87.625,  84.78,   83,     83.5,    81.375,  84.44,   89.25,   86.375,  86.25,
        85.25,   87.125,  85.815,  88.97,  88.47,   86.875,  86.815,  84.875,  84.19,   83.875,
        83.375,  85.5,    89.19,   89.44,  91.095,  90.75,   91.44,   89,      91,      90.5,
        89.03,   88.815,  84.28,   83.5,   82.69,   84.75,   85.655,  86.19,   88.94,   89.28,
        88.625,  88.5,    91.97,   91.5,   93.25,   93.5,    93.155,  91.72,   90,      89.69,
        88.875,  85.19,   83.375,  84.875, 85.94,   97.25,   99.875,  104.94,  106,     102.5,
        102.405, 104.595, 106.125, 106,    106.065, 104.625, 108.625, 109.315, 110.5,   112.75,
        123,     119.625, 118.75,  119.25, 117.94,  116.44,  115.19,  111.875, 110.595, 118.125,
        116,     116,     112,     113.75, 112.94,  116,     120.5,   116.62,  117,     115.25,
        114.31,  115.5,   115.87,  120.69, 120.19,  120.75,  124.75,  123.37,  122.94,  122.56,
        123.12,  122.56,  124.62,  129.25, 131,     132.25,  131,     132.81,  134,     137.38,
        137.81,  137.88,  137.25,  136.31, 136.25,  134.63,  128.25,  129,     123.87,  124.81,
        123,     126.25,  128.38,  125.37, 125.69,  122.25,  119.37,  118.5,   123.19,  123.5,
        122.19,  119.31,  123.31,  121.12, 123.37,  127.37,  128.5,   123.87,  122.94,  121.75,
        124.44,  122,     122.37,  122.94, 124,     123.19,  124.56,  127.25,  125.87,  128.86,
        132,     130.75,  134.75,  135,    132.38,  133.31,  131.94,  130,     125.37,  130.13,
        127.12,  125.19,  122,     125,    123,     123.5,   120.06,  121,     117.75,  119.87,
        122,     119.19,  116.37,  113.5,  114.25,  110,     105.06,  107,     107.87,  107,
        107.12,  107,     91,      93.94,  93.87,   95.5,    93,      94.94,   98.25,   96.75,
        94.81,   94.37,   91.56,   90.25,  93.94,   93.62,   97,      95,      95.87,   94.06,
        94.62,   93.75,   98,      103.94, 107.87,  106.06,  104.5,   105,     104.19,  103.06,
        103.42,  105.27,  111.87,  116,    116.62,  118.28,  113.37,  109,     109.7,   109.25,
        107,     109.19,  110,     109.2,  110.12,  108,     108.62,  109.75,  109.81,  109,
        108.75,  107.87,
    };
}

fn testVolumes() [252]f64 {
    return .{
        4077500,  4955900,  4775300,  4155300,  4593100,  3631300,  3382800,  4954200,  4500000,  3397500,
        4204500,  6321400,  10203600, 19043900, 11692000, 9553300,  8920300,  5970900,  5062300,  3705600,
        5865600,  5603000,  5811900,  8483800,  5995200,  5408800,  5430500,  6283800,  5834800,  4515500,
        4493300,  4346100,  3700300,  4600200,  4557200,  4323600,  5237500,  7404100,  4798400,  4372800,
        3872300,  10750800, 5804800,  3785500,  5014800,  3507700,  4298800,  4842500,  3952200,  3304700,
        3462000,  7253900,  9753100,  5953000,  5011700,  5910800,  4916900,  4135000,  4054200,  3735300,
        2921900,  2658400,  4624400,  4372200,  5831600,  4268600,  3059200,  4495500,  3425000,  3630800,
        4168100,  5966900,  7692800,  7362500,  6581300,  19587700, 10378600, 9334700,  10467200, 5671400,
        5645000,  4518600,  4519500,  5569700,  4239700,  4175300,  4995300,  4776600,  4190000,  6035300,
        12168900, 9040800,  5780300,  4320800,  3899100,  3221400,  3455500,  4304200,  4703900,  8316300,
        10553900, 6384800,  7163300,  7007800,  5114100,  5263800,  6666100,  7398400,  5575000,  4852300,
        4298100,  4900500,  4887700,  6964800,  4679200,  9165000,  6469800,  6792000,  4423800,  5231900,
        4565600,  6235200,  5225900,  8261400,  5912500,  3545600,  5714500,  6653900,  6094500,  4799200,
        5050800,  5648900,  4726300,  5585600,  5124800,  7630200,  14311600, 8793600,  8874200,  6966600,
        5525500,  6515500,  5291900,  5711700,  4327700,  4568000,  6859200,  5757500,  7367000,  6144100,
        4052700,  5849700,  5544700,  5032200,  4400600,  4894100,  5140000,  6610900,  7585200,  5963100,
        6045500,  8443300,  6464700,  6248300,  4357200,  4774700,  6216900,  6266900,  5584800,  5284500,
        7554500,  7209500,  8424800,  5094500,  4443600,  4591100,  5658400,  6094100,  14862200, 7544700,
        6985600,  8093000,  7590000,  7451300,  7078000,  7105300,  8778800,  6643900,  10563900, 7043100,
        6438900,  8057700,  14240000, 17872300, 7831100,  8277700,  15017800, 14183300, 13921100, 9683000,
        9187300,  11380500, 69447300, 26673600, 13768400, 11371600, 9872200,  9450500,  11083300, 9552800,
        11108400, 10374200, 16701900, 13741900, 8523600,  9551900,  8680500,  7151700,  9673100,  6264700,
        8541600,  8358000,  18720800, 19683100, 13682500, 10668100, 9710600,  3113100,  5682000,  5763600,
        5340000,  6220800,  14680500, 9933000,  11329500, 8145300,  16644700, 12593800, 7138100,  7442300,
        9442300,  7123600,  7680600,  4839800,  4775500,  4008800,  4533600,  3741100,  4084800,  2685200,
        3438000,  2870500,
    };
}
