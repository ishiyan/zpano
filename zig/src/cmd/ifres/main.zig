/// ifres — command-line indicator frequency response calculator.
///
/// Reads a JSON settings file containing indicator definitions,
/// creates indicator instances, determines each indicator's warmup period,
/// and calculates the frequency response with signal length 1024.
const std = @import("std");
const indicators = @import("indicators");
const entities = @import("entities");
const Scalar = entities.Scalar;

const Io = std.Io;
const Identifier = indicators.identifier.Identifier;
const Indicator = indicators.indicator.Indicator;
const Metadata = indicators.metadata.Metadata;
const FactoryResult = indicators.factory.FactoryResult;
const frequency_response = indicators.frequency_response;
const FrequencyResponse = frequency_response.FrequencyResponse;
const Component = frequency_response.Component;

const signal_length: usize = 1024;
const max_warmup: usize = 10000;
const phase_degrees_unwrapping_limit: f64 = 179.0;

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;
    const io = init.io;

    // Parse command-line args
    var args_iter = init.minimal.args.iterate();
    _ = args_iter.next(); // skip program name
    const settings_path = args_iter.next() orelse {
        std.debug.print("usage: ifres <settings.json>\n", .{});
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

        // Create a probe instance to detect warmup period.
        const probe_result = indicators.factory.create(allocator, id, params_json) catch |err| {
            std.debug.print("error creating indicator '{s}': {}\n", .{ id_str, err });
            std.process.exit(1);
        };
        const warmup = detectWarmup(probe_result.indicator);
        probe_result.deinit(allocator);

        // Create a fresh instance for the actual calculation.
        const result = indicators.factory.create(allocator, id, params_json) catch |err| {
            std.debug.print("error creating indicator '{s}': {}\n", .{ id_str, err });
            std.process.exit(1);
        };
        defer result.deinit(allocator);

        var fr = frequency_response.calculate(
            allocator,
            signal_length,
            result.indicator,
            warmup,
            phase_degrees_unwrapping_limit,
        ) catch |err| {
            std.debug.print("error calculating frequency response for '{s}': {}\n", .{ id_str, err });
            std.process.exit(1);
        };
        defer fr.deinit(allocator);

        try printFrequencyResponse(stdout, &fr, warmup);
    }

    try stdout.flush();
}

fn detectWarmup(ind: Indicator) usize {
    var scalar = Scalar{ .time = 0, .value = 0.0 };
    for (0..max_warmup) |i| {
        if (ind.isPrimed()) {
            return i;
        }
        _ = ind.updateScalar(&scalar);
    }
    return max_warmup;
}

fn printFrequencyResponse(stdout: *Io.Writer, fr: *const FrequencyResponse, warmup: usize) !void {
    try stdout.print("=== {s} (warmup={d}) ===\n", .{ fr.label, warmup });
    try stdout.print("  Spectrum length: {d}\n", .{fr.normalized_frequency.len});

    try printComponent(stdout, "PowerPercent", fr.power_percent);
    try printComponent(stdout, "PowerDecibel", fr.power_decibel);
    try printComponent(stdout, "AmplitudePercent", fr.amplitude_percent);
    try printComponent(stdout, "AmplitudeDecibel", fr.amplitude_decibel);
    try printComponent(stdout, "PhaseDegrees", fr.phase_degrees);
    try printComponent(stdout, "PhaseDegreesUnwrapped", fr.phase_degrees_unwrapped);

    try stdout.writeByte('\n');
}

fn printComponent(stdout: *Io.Writer, name: []const u8, c: Component) !void {
    try stdout.print("  {s:<25} min={d:10.4}  max={d:10.4}", .{ name, c.min, c.max });

    const n = c.data.len;
    if (n == 0) {
        try stdout.writeByte('\n');
        return;
    }

    // Print first 3 and last 3 values as a preview.
    const preview = 3;
    if (n <= preview * 2) {
        try stdout.writeAll("  data=[");
        for (c.data, 0..) |v, i| {
            if (i > 0) try stdout.writeByte(' ');
            try stdout.print("{d:.4}", .{v});
        }
        try stdout.writeByte(']');
    } else {
        try stdout.print("  data=[{d:.4} {d:.4} {d:.4} ... {d:.4} {d:.4} {d:.4}]", .{
            c.data[0],
            c.data[1],
            c.data[2],
            c.data[n - 3],
            c.data[n - 2],
            c.data[n - 1],
        });
    }

    try stdout.writeByte('\n');
}
