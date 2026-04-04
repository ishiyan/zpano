const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

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

    // --- Tests ---
    const conventions_tests = b.addTest(.{ .root_module = conventions_test_mod });
    const daycounting_tests = b.addTest(.{ .root_module = daycounting_test_mod });
    const fractional_tests = b.addTest(.{ .root_module = fractional_test_mod });
    const periodicity_tests = b.addTest(.{ .root_module = periodicity_test_mod });
    const ratios_tests = b.addTest(.{ .root_module = ratios_test_mod });

    const run_conventions_tests = b.addRunArtifact(conventions_tests);
    const run_daycounting_tests = b.addRunArtifact(daycounting_tests);
    const run_fractional_tests = b.addRunArtifact(fractional_tests);
    const run_periodicity_tests = b.addRunArtifact(periodicity_tests);
    const run_ratios_tests = b.addRunArtifact(ratios_tests);

    const test_step = b.step("test", "Run all tests");
    test_step.dependOn(&run_conventions_tests.step);
    test_step.dependOn(&run_daycounting_tests.step);
    test_step.dependOn(&run_fractional_tests.step);
    test_step.dependOn(&run_periodicity_tests.step);
    test_step.dependOn(&run_ratios_tests.step);
}
