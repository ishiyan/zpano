/// icalc — command-line indicator calculator.
///
/// Reads a JSON settings file containing indicator definitions,
/// creates indicator instances, prints their metadata, then iterates
/// through embedded bar data printing bar values and all indicator
/// outputs on each iteration.
const std = @import("std");
const indicators = @import("indicators");
const entities = @import("entities");
const Bar = entities.Bar;
const Scalar = entities.Scalar;

const Io = std.Io;
const Identifier = indicators.identifier.Identifier;
const Indicator = indicators.indicator.Indicator;
const OutputArray = indicators.indicator.OutputArray;
const OutputValue = indicators.indicator.OutputValue;
const Metadata = indicators.metadata.Metadata;
const Shape = indicators.shape.Shape;
const FactoryResult = indicators.factory.FactoryResult;

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const io = init.io;

    // Parse command-line args
    var args_iter = init.minimal.args.iterate();
    _ = args_iter.next(); // skip program name
    const settings_path = args_iter.next() orelse {
        std.debug.print("usage: icalc <settings.json>\n", .{});
        std.process.exit(1);
    };

    const settings_data = Io.Dir.cwd().readFileAlloc(io, settings_path, allocator, .unlimited) catch |err| {
        std.debug.print("error reading settings file: {}\n", .{err});
        std.process.exit(1);
    };
    defer allocator.free(settings_data);

    // Set up stdout writer
    var stdout_buf: [8192]u8 = undefined;
    var file_writer = Io.File.Writer.init(Io.File.stdout(), io, &stdout_buf);
    const stdout = &file_writer.interface;

    // Parse settings JSON: array of { "identifier": "...", "params": {...} }
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

        // Stringify the params object back to JSON for the factory
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

    // Print metadata
    try printMetadata(stdout, results.items);
    try stdout.flush();

    // Generate test bars
    const bars = testBars();

    try stdout.writeAll("\n=== Bar Data & Indicator Outputs ===\n\n");

    // Iterate bars
    for (&bars, 0..) |*bar, i| {
        try stdout.print("Bar[{d:3}] {d:4}-{d:02}-{d:02}  O={d:.4} H={d:.4} L={d:.4} C={d:.4} V={d:.0}\n", .{
            i,
            dateYear(bar.time),
            dateMonth(bar.time),
            dateDay(bar.time),
            bar.open,
            bar.high,
            bar.low,
            bar.close,
            bar.volume,
        });

        for (results.items) |r| {
            var meta: Metadata = undefined;
            r.indicator.metadata(&meta);

            const output = r.indicator.updateBar(bar);

            try stdout.print("  {s:<45} primed={s:<5} ", .{
                meta.mnemonic,
                if (r.indicator.isPrimed()) "true" else "false",
            });

            try printOutput(stdout, &meta, &output);
            try stdout.writeByte('\n');
        }

        try stdout.writeByte('\n');
    }

    try stdout.flush();
}

fn printMetadata(stdout: *Io.Writer, results: []const FactoryResult) !void {
    try stdout.writeAll("=== Indicator Metadata ===\n\n");

    for (results, 0..) |r, i| {
        var meta: Metadata = undefined;
        r.indicator.metadata(&meta);

        try stdout.print("[{d}] {s}\n", .{ i, meta.mnemonic });
        try stdout.print("  Identifier:  {s}\n", .{meta.identifier.asStr()});
        try stdout.print("  Description: {s}\n", .{meta.description});
        try stdout.print("  Outputs ({d}):\n", .{meta.outputs_len});

        for (meta.outputs_buf[0..meta.outputs_len], 0..) |out, j| {
            try stdout.print("    [{d}] kind={d} shape={s} mnemonic=\"{s}\" description=\"{s}\"\n", .{
                j, out.kind, out.shape.asStr(), out.mnemonic, out.description,
            });
        }

        try stdout.writeByte('\n');
    }
}

fn printOutput(stdout: *Io.Writer, meta: *const Metadata, output: *const OutputArray) !void {
    for (output.slice(), 0..) |val, i| {
        const name: []const u8 = if (i < meta.outputs_len) meta.outputs_buf[i].mnemonic else "out[?]";

        switch (val) {
            .scalar => |s| {
                if (std.math.isNan(s.value)) {
                    try stdout.print("{s}=NaN ", .{name});
                } else {
                    try stdout.print("{s}={d:.4} ", .{ name, s.value });
                }
            },
            .band => |b| {
                if (b.isEmpty()) {
                    try stdout.print("{s}=Band(NaN) ", .{name});
                } else {
                    try stdout.print("{s}=Band({d:.4},{d:.4}) ", .{ name, b.lower, b.upper });
                }
            },
            .heatmap => |h| {
                if (h.isEmpty()) {
                    try stdout.print("{s}=Heatmap(empty) ", .{name});
                } else {
                    try stdout.print("{s}=Heatmap({d} values) ", .{ name, h.values_len });
                }
            },
            .polyline => |p| {
                if (p.isEmpty()) {
                    try stdout.print("{s}=Polyline(empty) ", .{name});
                } else {
                    try stdout.print("{s}=Polyline({d} points) ", .{ name, p.points_len });
                }
            },
        }
    }
}

// ── Date helpers (Unix epoch seconds → year/month/day) ────────────────────

fn dateYear(timestamp: i64) u16 {
    const es = std.time.epoch.EpochSeconds{ .secs = @intCast(timestamp) };
    const ed = es.getEpochDay();
    const yd = ed.calculateYearDay();
    return @intCast(yd.year);
}

fn dateMonth(timestamp: i64) u5 {
    const es = std.time.epoch.EpochSeconds{ .secs = @intCast(timestamp) };
    const ed = es.getEpochDay();
    const yd = ed.calculateYearDay();
    const md = yd.calculateMonthDay();
    return @intFromEnum(md.month);
}

fn dateDay(timestamp: i64) u5 {
    const es = std.time.epoch.EpochSeconds{ .secs = @intCast(timestamp) };
    const ed = es.getEpochDay();
    const yd = ed.calculateYearDay();
    const md = yd.calculateMonthDay();
    return md.day_index + 1;
}

// ── Embedded test data ─────────────────────────────────────────────────────

fn testBars() [252]Bar {
    const highs = testHighs();
    const lows = testLows();
    const closes = testCloses();
    const volumes = testVolumes();

    var bars: [252]Bar = undefined;
    // Base time: 2020-01-02 00:00:00 UTC
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
