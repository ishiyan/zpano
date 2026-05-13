const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const test_filter = b.option([]const u8, "test-filter", "Filter tests by name");
    const filters: []const []const u8 = if (test_filter) |f|
        b.allocator.dupe([]const u8, &.{f}) catch @panic("OOM")
    else
        &.{};

    // --- Library modules ---
    const conventions_mod = b.addModule("conventions", .{
        .root_source_file = b.path("src/daycounting/conventions.zig"),
        .target = target,
        .optimize = optimize,
    });

    const daycounting_mod = b.addModule("daycounting", .{
        .root_source_file = b.path("src/daycounting/daycounting.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
        },
    });

    const fractional_mod = b.addModule("fractional", .{
        .root_source_file = b.path("src/daycounting/fractional.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
            .{ .name = "daycounting", .module = daycounting_mod },
        },
    });

    const periodicity_mod = b.addModule("periodicity", .{
        .root_source_file = b.path("src/performance/periodicity.zig"),
        .target = target,
        .optimize = optimize,
    });

    _ = b.addModule("ratios", .{
        .root_source_file = b.path("src/performance/ratios.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
            .{ .name = "daycounting", .module = daycounting_mod },
            .{ .name = "fractional", .module = fractional_mod },
            .{ .name = "periodicity", .module = periodicity_mod },
        },
    });

    // --- Roundtrip library modules ---
    const execution_mod = b.addModule("execution", .{
        .root_source_file = b.path("src/roundtrips/execution.zig"),
        .target = target,
        .optimize = optimize,
    });

    const side_mod = b.addModule("side", .{
        .root_source_file = b.path("src/roundtrips/side.zig"),
        .target = target,
        .optimize = optimize,
    });

    _ = b.addModule("matching", .{
        .root_source_file = b.path("src/roundtrips/matching.zig"),
        .target = target,
        .optimize = optimize,
    });

    _ = b.addModule("grouping", .{
        .root_source_file = b.path("src/roundtrips/grouping.zig"),
        .target = target,
        .optimize = optimize,
    });

    const roundtrip_mod = b.addModule("roundtrip", .{
        .root_source_file = b.path("src/roundtrips/roundtrip.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "execution", .module = execution_mod },
            .{ .name = "side", .module = side_mod },
            .{ .name = "fractional", .module = fractional_mod },
        },
    });

    _ = b.addModule("rt_performance", .{
        .root_source_file = b.path("src/roundtrips/performance.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
            .{ .name = "fractional", .module = fractional_mod },
            .{ .name = "side", .module = side_mod },
            .{ .name = "roundtrip", .module = roundtrip_mod },
            .{ .name = "execution", .module = execution_mod },
        },
    });

    // --- Entities library modules ---
    // Individual entity modules are registered because:
    // 1. Component modules depend on their parent (bar_component → bar, etc.)
    // 2. The entities barrel (entities_mod) re-exports them via @import("bar") etc.
    // 3. Entity test targets reference them directly.
    // External consumers (indicators, CLIs) use the aggregate entities_mod barrel.
    const bar_mod = b.addModule("bar", .{
        .root_source_file = b.path("src/entities/bar.zig"),
        .target = target,
        .optimize = optimize,
    });

    const quote_mod = b.addModule("quote", .{
        .root_source_file = b.path("src/entities/quote.zig"),
        .target = target,
        .optimize = optimize,
    });

    const trade_mod = b.addModule("trade", .{
        .root_source_file = b.path("src/entities/trade.zig"),
        .target = target,
        .optimize = optimize,
    });

    const scalar_mod = b.addModule("scalar", .{
        .root_source_file = b.path("src/entities/scalar.zig"),
        .target = target,
        .optimize = optimize,
    });

    const bar_component_mod = b.addModule("bar_component", .{
        .root_source_file = b.path("src/entities/bar_component.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "bar", .module = bar_mod },
        },
    });

    const quote_component_mod = b.addModule("quote_component", .{
        .root_source_file = b.path("src/entities/quote_component.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "quote", .module = quote_mod },
        },
    });

    const trade_component_mod = b.addModule("trade_component", .{
        .root_source_file = b.path("src/entities/trade_component.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "trade", .module = trade_mod },
        },
    });

    // --- Entities aggregate module (barrel for CLIs) ---
    const entities_mod = b.addModule("entities", .{
        .root_source_file = b.path("src/entities/entities.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "bar", .module = bar_mod },
            .{ .name = "quote", .module = quote_mod },
            .{ .name = "trade", .module = trade_mod },
            .{ .name = "scalar", .module = scalar_mod },
            .{ .name = "bar_component", .module = bar_component_mod },
            .{ .name = "quote_component", .module = quote_component_mod },
            .{ .name = "trade_component", .module = trade_component_mod },
        },
    });

    // --- Indicators library module ---
    const indicators_mod = b.addModule("indicators", .{
        .root_source_file = b.path("src/indicators/indicators.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "entities", .module = entities_mod },
        },
    });

    // --- CLI executables ---
    const icalc_exe = b.addExecutable(.{
        .name = "icalc",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/cmd/icalc/main.zig"),
            .target = target,
            .optimize = optimize,
            .imports = &.{
                .{ .name = "indicators", .module = indicators_mod },
                .{ .name = "entities", .module = entities_mod },
            },
        }),
    });
    b.installArtifact(icalc_exe);

    const run_icalc = b.addRunArtifact(icalc_exe);
    run_icalc.step.dependOn(b.getInstallStep());
    if (b.args) |args| {
        run_icalc.addArgs(args);
    }
    const run_icalc_step = b.step("icalc", "Run the icalc indicator calculator");
    run_icalc_step.dependOn(&run_icalc.step);

    const ifres_exe = b.addExecutable(.{
        .name = "ifres",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/cmd/ifres/main.zig"),
            .target = target,
            .optimize = optimize,
            .imports = &.{
                .{ .name = "indicators", .module = indicators_mod },
                .{ .name = "entities", .module = entities_mod },
            },
        }),
    });
    b.installArtifact(ifres_exe);

    const run_ifres = b.addRunArtifact(ifres_exe);
    run_ifres.step.dependOn(b.getInstallStep());
    if (b.args) |args| {
        run_ifres.addArgs(args);
    }
    const run_ifres_step = b.step("ifres", "Run the ifres frequency response calculator");
    run_ifres_step.dependOn(&run_ifres.step);

    const iconf_exe = b.addExecutable(.{
        .name = "iconf",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/cmd/iconf/main.zig"),
            .target = target,
            .optimize = optimize,
            .imports = &.{
                .{ .name = "indicators", .module = indicators_mod },
                .{ .name = "entities", .module = entities_mod },
            },
        }),
    });
    b.installArtifact(iconf_exe);

    const run_iconf = b.addRunArtifact(iconf_exe);
    run_iconf.step.dependOn(b.getInstallStep());
    if (b.args) |args| {
        run_iconf.addArgs(args);
    }
    const run_iconf_step = b.step("iconf", "Run the iconf chart configuration generator");
    run_iconf_step.dependOn(&run_iconf.step);

    // --- Symbology library modules (no dependencies) ---
    _ = b.addModule("isin", .{
        .root_source_file = b.path("src/symbology/isin.zig"),
        .target = target,
        .optimize = optimize,
    });

    _ = b.addModule("cusip", .{
        .root_source_file = b.path("src/symbology/cusip.zig"),
        .target = target,
        .optimize = optimize,
    });

    _ = b.addModule("sedol", .{
        .root_source_file = b.path("src/symbology/sedol.zig"),
        .target = target,
        .optimize = optimize,
    });

    // --- Fuzzy library module (barrel re-exporting membership, operators, defuzzify) ---
    const fuzzy_mod = b.addModule("fuzzy", .{
        .root_source_file = b.path("src/fuzzy/fuzzy.zig"),
        .target = target,
        .optimize = optimize,
    });

    // --- Candlestick patterns library module (depends on fuzzy) ---
    const candlestick_patterns_mod = b.addModule("candlestick_patterns", .{
        .root_source_file = b.path("src/candlestick_patterns/candlestick_patterns.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    // --- Signals library modules (depend on fuzzy) ---
    _ = b.addModule("sig_threshold", .{
        .root_source_file = b.path("src/signals/threshold.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    _ = b.addModule("sig_crossover", .{
        .root_source_file = b.path("src/signals/crossover.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    _ = b.addModule("sig_band", .{
        .root_source_file = b.path("src/signals/band.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    _ = b.addModule("sig_histogram", .{
        .root_source_file = b.path("src/signals/histogram.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    _ = b.addModule("sig_compose", .{
        .root_source_file = b.path("src/signals/compose.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    // --- Signal ensemble library module (standalone — zero dependencies) ---
    _ = b.addModule("signal_ensemble", .{
        .root_source_file = b.path("src/signal_ensemble/signal_ensemble.zig"),
        .target = target,
        .optimize = optimize,
    });

    // --- Test modules (separate modules that share the same source files) ---
    const conventions_test_mod = b.createModule(.{
        .root_source_file = b.path("src/daycounting/conventions.zig"),
        .target = target,
        .optimize = optimize,
    });

    const daycounting_test_mod = b.createModule(.{
        .root_source_file = b.path("src/daycounting/daycounting.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
        },
    });

    const fractional_test_mod = b.createModule(.{
        .root_source_file = b.path("src/daycounting/fractional.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
            .{ .name = "daycounting", .module = daycounting_mod },
        },
    });

    const periodicity_test_mod = b.createModule(.{
        .root_source_file = b.path("src/performance/periodicity.zig"),
        .target = target,
        .optimize = optimize,
    });

    const ratios_test_mod = b.createModule(.{
        .root_source_file = b.path("src/performance/ratios.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
            .{ .name = "daycounting", .module = daycounting_mod },
            .{ .name = "fractional", .module = fractional_mod },
            .{ .name = "periodicity", .module = periodicity_mod },
        },
    });

    // --- Roundtrip test modules ---
    const execution_test_mod = b.createModule(.{
        .root_source_file = b.path("src/roundtrips/execution.zig"),
        .target = target,
        .optimize = optimize,
    });

    const side_test_mod = b.createModule(.{
        .root_source_file = b.path("src/roundtrips/side.zig"),
        .target = target,
        .optimize = optimize,
    });

    const matching_test_mod = b.createModule(.{
        .root_source_file = b.path("src/roundtrips/matching.zig"),
        .target = target,
        .optimize = optimize,
    });

    const grouping_test_mod = b.createModule(.{
        .root_source_file = b.path("src/roundtrips/grouping.zig"),
        .target = target,
        .optimize = optimize,
    });

    const roundtrip_test_mod = b.createModule(.{
        .root_source_file = b.path("src/roundtrips/roundtrip.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "execution", .module = execution_mod },
            .{ .name = "side", .module = side_mod },
            .{ .name = "fractional", .module = fractional_mod },
        },
    });

    const rt_performance_test_mod = b.createModule(.{
        .root_source_file = b.path("src/roundtrips/performance.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "conventions", .module = conventions_mod },
            .{ .name = "fractional", .module = fractional_mod },
            .{ .name = "side", .module = side_mod },
            .{ .name = "roundtrip", .module = roundtrip_mod },
            .{ .name = "execution", .module = execution_mod },
        },
    });

    // --- Symbology test modules ---
    const isin_test_mod = b.createModule(.{
        .root_source_file = b.path("src/symbology/isin.zig"),
        .target = target,
        .optimize = optimize,
    });

    const cusip_test_mod = b.createModule(.{
        .root_source_file = b.path("src/symbology/cusip.zig"),
        .target = target,
        .optimize = optimize,
    });

    const sedol_test_mod = b.createModule(.{
        .root_source_file = b.path("src/symbology/sedol.zig"),
        .target = target,
        .optimize = optimize,
    });

    // --- Entities test modules ---
    const bar_test_mod = b.createModule(.{
        .root_source_file = b.path("src/entities/bar.zig"),
        .target = target,
        .optimize = optimize,
    });

    const quote_test_mod = b.createModule(.{
        .root_source_file = b.path("src/entities/quote.zig"),
        .target = target,
        .optimize = optimize,
    });

    const trade_test_mod = b.createModule(.{
        .root_source_file = b.path("src/entities/trade.zig"),
        .target = target,
        .optimize = optimize,
    });

    const scalar_test_mod = b.createModule(.{
        .root_source_file = b.path("src/entities/scalar.zig"),
        .target = target,
        .optimize = optimize,
    });

    const bar_component_test_mod = b.createModule(.{
        .root_source_file = b.path("src/entities/bar_component.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "bar", .module = bar_mod },
        },
    });

    const quote_component_test_mod = b.createModule(.{
        .root_source_file = b.path("src/entities/quote_component.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "quote", .module = quote_mod },
        },
    });

    const trade_component_test_mod = b.createModule(.{
        .root_source_file = b.path("src/entities/trade_component.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "trade", .module = trade_mod },
        },
    });

    const indicators_test_mod = b.createModule(.{
        .root_source_file = b.path("src/indicators/indicators.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "entities", .module = entities_mod },
        },
    });

    // --- Fuzzy test module (barrel) ---
    const fuzzy_test_mod = b.createModule(.{
        .root_source_file = b.path("src/fuzzy/fuzzy.zig"),
        .target = target,
        .optimize = optimize,
    });

    // --- Signals test modules ---
    const sig_threshold_test_mod = b.createModule(.{
        .root_source_file = b.path("src/signals/threshold.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    const sig_crossover_test_mod = b.createModule(.{
        .root_source_file = b.path("src/signals/crossover.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    const sig_band_test_mod = b.createModule(.{
        .root_source_file = b.path("src/signals/band.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    const sig_histogram_test_mod = b.createModule(.{
        .root_source_file = b.path("src/signals/histogram.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    const sig_compose_test_mod = b.createModule(.{
        .root_source_file = b.path("src/signals/compose.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    // --- Candlestick patterns test module ---
    const cp_test_mod = b.createModule(.{
        .root_source_file = b.path("src/candlestick_patterns/candlestick_patterns_test.zig"),
        .target = target,
        .optimize = optimize,
        .imports = &.{
            .{ .name = "candlestick_patterns", .module = candlestick_patterns_mod },
            .{ .name = "fuzzy", .module = fuzzy_mod },
        },
    });

    // --- Signal ensemble test module (barrel) ---
    const signal_ensemble_test_mod = b.createModule(.{
        .root_source_file = b.path("src/signal_ensemble/signal_ensemble.zig"),
        .target = target,
        .optimize = optimize,
    });

    // --- Tests ---
    const conventions_tests = b.addTest(.{ .root_module = conventions_test_mod, .filters = filters });
    const daycounting_tests = b.addTest(.{ .root_module = daycounting_test_mod, .filters = filters });
    const fractional_tests = b.addTest(.{ .root_module = fractional_test_mod, .filters = filters });
    const periodicity_tests = b.addTest(.{ .root_module = periodicity_test_mod, .filters = filters });
    const ratios_tests = b.addTest(.{ .root_module = ratios_test_mod, .filters = filters });
    const execution_tests = b.addTest(.{ .root_module = execution_test_mod, .filters = filters });
    const side_tests = b.addTest(.{ .root_module = side_test_mod, .filters = filters });
    const matching_tests = b.addTest(.{ .root_module = matching_test_mod, .filters = filters });
    const grouping_tests = b.addTest(.{ .root_module = grouping_test_mod, .filters = filters });
    const roundtrip_tests = b.addTest(.{ .root_module = roundtrip_test_mod, .filters = filters });
    const rt_performance_tests = b.addTest(.{ .root_module = rt_performance_test_mod, .filters = filters });
    const isin_tests = b.addTest(.{ .root_module = isin_test_mod, .filters = filters });
    const cusip_tests = b.addTest(.{ .root_module = cusip_test_mod, .filters = filters });
    const sedol_tests = b.addTest(.{ .root_module = sedol_test_mod, .filters = filters });
    const bar_tests = b.addTest(.{ .root_module = bar_test_mod, .filters = filters });
    const quote_tests = b.addTest(.{ .root_module = quote_test_mod, .filters = filters });
    const trade_tests = b.addTest(.{ .root_module = trade_test_mod, .filters = filters });
    const scalar_tests = b.addTest(.{ .root_module = scalar_test_mod, .filters = filters });
    const bar_component_tests = b.addTest(.{ .root_module = bar_component_test_mod, .filters = filters });
    const quote_component_tests = b.addTest(.{ .root_module = quote_component_test_mod, .filters = filters });
    const trade_component_tests = b.addTest(.{ .root_module = trade_component_test_mod, .filters = filters });
    const indicators_tests = b.addTest(.{ .root_module = indicators_test_mod, .filters = filters });
    const fuzzy_tests = b.addTest(.{ .root_module = fuzzy_test_mod, .filters = filters });
    const sig_threshold_tests = b.addTest(.{ .root_module = sig_threshold_test_mod, .filters = filters });
    const sig_crossover_tests = b.addTest(.{ .root_module = sig_crossover_test_mod, .filters = filters });
    const sig_band_tests = b.addTest(.{ .root_module = sig_band_test_mod, .filters = filters });
    const sig_histogram_tests = b.addTest(.{ .root_module = sig_histogram_test_mod, .filters = filters });
    const sig_compose_tests = b.addTest(.{ .root_module = sig_compose_test_mod, .filters = filters });
    const cp_tests = b.addTest(.{ .root_module = cp_test_mod, .filters = filters });
    const signal_ensemble_tests = b.addTest(.{ .root_module = signal_ensemble_test_mod, .filters = filters });

    const run_conventions_tests = b.addRunArtifact(conventions_tests);
    const run_daycounting_tests = b.addRunArtifact(daycounting_tests);
    const run_fractional_tests = b.addRunArtifact(fractional_tests);
    const run_periodicity_tests = b.addRunArtifact(periodicity_tests);
    const run_ratios_tests = b.addRunArtifact(ratios_tests);
    const run_execution_tests = b.addRunArtifact(execution_tests);
    const run_side_tests = b.addRunArtifact(side_tests);
    const run_matching_tests = b.addRunArtifact(matching_tests);
    const run_grouping_tests = b.addRunArtifact(grouping_tests);
    const run_roundtrip_tests = b.addRunArtifact(roundtrip_tests);
    const run_rt_performance_tests = b.addRunArtifact(rt_performance_tests);
    const run_isin_tests = b.addRunArtifact(isin_tests);
    const run_cusip_tests = b.addRunArtifact(cusip_tests);
    const run_sedol_tests = b.addRunArtifact(sedol_tests);
    const run_bar_tests = b.addRunArtifact(bar_tests);
    const run_quote_tests = b.addRunArtifact(quote_tests);
    const run_trade_tests = b.addRunArtifact(trade_tests);
    const run_scalar_tests = b.addRunArtifact(scalar_tests);
    const run_bar_component_tests = b.addRunArtifact(bar_component_tests);
    const run_quote_component_tests = b.addRunArtifact(quote_component_tests);
    const run_trade_component_tests = b.addRunArtifact(trade_component_tests);
    const run_indicators_tests = b.addRunArtifact(indicators_tests);
    const run_fuzzy_tests = b.addRunArtifact(fuzzy_tests);
    const run_sig_threshold_tests = b.addRunArtifact(sig_threshold_tests);
    const run_sig_crossover_tests = b.addRunArtifact(sig_crossover_tests);
    const run_sig_band_tests = b.addRunArtifact(sig_band_tests);
    const run_sig_histogram_tests = b.addRunArtifact(sig_histogram_tests);
    const run_sig_compose_tests = b.addRunArtifact(sig_compose_tests);
    const run_cp_tests = b.addRunArtifact(cp_tests);
    const run_signal_ensemble_tests = b.addRunArtifact(signal_ensemble_tests);

    const test_step = b.step("test", "Run all tests");
    test_step.dependOn(&run_conventions_tests.step);
    test_step.dependOn(&run_daycounting_tests.step);
    test_step.dependOn(&run_fractional_tests.step);
    test_step.dependOn(&run_periodicity_tests.step);
    test_step.dependOn(&run_ratios_tests.step);
    test_step.dependOn(&run_execution_tests.step);
    test_step.dependOn(&run_side_tests.step);
    test_step.dependOn(&run_matching_tests.step);
    test_step.dependOn(&run_grouping_tests.step);
    test_step.dependOn(&run_roundtrip_tests.step);
    test_step.dependOn(&run_rt_performance_tests.step);
    test_step.dependOn(&run_isin_tests.step);
    test_step.dependOn(&run_cusip_tests.step);
    test_step.dependOn(&run_sedol_tests.step);
    test_step.dependOn(&run_bar_tests.step);
    test_step.dependOn(&run_quote_tests.step);
    test_step.dependOn(&run_trade_tests.step);
    test_step.dependOn(&run_scalar_tests.step);
    test_step.dependOn(&run_bar_component_tests.step);
    test_step.dependOn(&run_quote_component_tests.step);
    test_step.dependOn(&run_trade_component_tests.step);
    test_step.dependOn(&run_indicators_tests.step);
    test_step.dependOn(&run_fuzzy_tests.step);
    test_step.dependOn(&run_sig_threshold_tests.step);
    test_step.dependOn(&run_sig_crossover_tests.step);
    test_step.dependOn(&run_sig_band_tests.step);
    test_step.dependOn(&run_sig_histogram_tests.step);
    test_step.dependOn(&run_sig_compose_tests.step);
    test_step.dependOn(&run_cp_tests.step);
    test_step.dependOn(&run_signal_ensemble_tests.step);
}
