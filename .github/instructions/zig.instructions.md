---
description: 'Critical Zig patterns to avoid outdated training data. Covers breaking changes in Zig 0.14–0.15.x that cause compile errors.'
applyTo: '**/*.zig,**/build.zig,**/build.zig.zon'
---

<!-- Derived from: .opencode/skills/zig/SKILL.md (upstream: https://github.com/rudedogg/zig-skills) -->
<!-- Conventions from: .opencode/skills/zig-best-practices/SKILL.md -->
<!-- To update: when either skill is updated, diff against this file and merge relevant changes. -->
<!-- This file is a compact always-on subset. Load the full skills for deeper guidance. -->

# Zig — Coding Conventions

- 4-space indent (standard Zig formatting). Use `zig fmt` to format code.
- `camelCase` for functions, `PascalCase` for types/enums/error sets, `snake_case` for struct fields and variables.
- Prefer `const` over `var`. Immutability signals intent and enables optimizations.
- Prefer slices over raw pointers for bounds safety.
- Pass allocators explicitly — never use global state for allocation. Use `std.testing.allocator` in tests for leak detection.
- Use `defer` immediately after acquiring a resource. Use `errdefer` for cleanup on error paths only.
- Define specific error sets for functions; avoid `anyerror`. Specific errors document failure modes.
- Handle all branches in `switch`; use `else` with an error or `unreachable` for impossible cases.
- Prefer `comptime T: type` over `anytype` — explicit types produce clearer error messages. Use `anytype` only for callbacks or truly generic functions.
- Use `orelse` to provide defaults for optionals; use `.?` only when null is a program error.
- Use `std.log.scoped` for namespaced logging; define a module-level `const log` for consistent scope.
- Tests live at the bottom of source files. Use `test "descriptive name" { ... }` blocks.
- Larger cohesive files are idiomatic. Split only when a file handles genuinely separate concerns.

# Zig — Avoid Outdated Patterns (0.14–0.15.x)

Training data contains removed APIs and old syntax. These rules prevent compile errors.

## Removed Features

- **`usingnamespace`** — REMOVED. Use explicit re-exports: `pub const foo = other.foo;`
- **`async`/`await`** — REMOVED. Keywords no longer exist.
- **`@fence`** — REMOVED. Use stronger atomic orderings or RMW operations.
- **`@setCold`** — REMOVED. Use `@branchHint(.cold)` (must be first statement in block).
- **`std.BoundedArray`** — REMOVED. Use `std.ArrayList(T).initBuffer(&buffer)`.

## Build System

`root_source_file` is REMOVED from `addExecutable`/`addLibrary`/`addTest`. Use `root_module`:

```zig
// WRONG
b.addExecutable(.{ .name = "app", .root_source_file = b.path("src/main.zig"), .target = target });

// CORRECT
b.addExecutable(.{ .name = "app", .root_module = b.createModule(.{
    .root_source_file = b.path("src/main.zig"), .target = target, .optimize = optimize,
}) });
```

Module imports: use `exe.root_module.addImport("name", mod)` not `exe.addModule(...)`.

Compile-level methods (`linkSystemLibrary`, `addCSourceFiles`, `addIncludePath`, `linkLibC`) are deprecated — use `exe.root_module.*` equivalents.

## I/O API Rewrite

The entire `std.io` API changed. `GenericWriter`, `GenericReader`, `AnyWriter`, `AnyReader`, `BufferedWriter`, `FixedBufferStream` are all deprecated.

```zig
// WRONG
const stdout = std.io.getStdOut().writer();
try stdout.print("Hello\n", .{});

// CORRECT — provide buffer, access .interface, flush
var buf: [4096]u8 = undefined;
var stdout_writer = std.fs.File.stdout().writer(&buf);
const stdout = &stdout_writer.interface;
try stdout.print("Hello\n", .{});
try stdout.flush();
```

## Container Initialization

Never use `.{}` for containers. Use `.empty` or `.init`:

```zig
// WRONG
var list: std.ArrayList(u32) = .{};
var gpa: std.heap.DebugAllocator(.{}) = .{};

// CORRECT
var list: std.ArrayList(u32) = .empty;
var map: std.AutoHashMapUnmanaged(u32, u32) = .empty;
var gpa: std.heap.DebugAllocator(.{}) = .init;
```

## Renamed Types

- `std.ArrayListUnmanaged` → `std.ArrayList` (unmanaged is now the default)
- `std.heap.GeneralPurposeAllocator` → `std.heap.DebugAllocator` (old alias still works)
- `std.fmt.Formatter` → `std.fmt.Alt`
- `CountingWriter` → `std.Io.Writer.Discarding`

## Format Strings

Use `{f}` to invoke format methods (not `{}`):

```zig
// WRONG — ambiguous error
std.debug.print("{}", .{std.zig.fmtId("x")});

// CORRECT
std.debug.print("{f}", .{std.zig.fmtId("x")});
```

Format method signature changed to: `pub fn format(self: @This(), writer: *std.Io.Writer) std.Io.Writer.Error!void`

## Other Breaking Changes

- **`@export`** now takes a pointer: `@export(&foo, .{ .name = "bar" })` not `@export(foo, ...)`
- **Inline asm clobbers** are typed: `.{ .rcx = true, .r11 = true }` not `"rcx", "r11"`
- **Arithmetic on `undefined`** is now illegal
- **`sanitize_c`** type changed to `?std.zig.SanitizeC` — use `.full`, `.trap`, or `.off`

## Quick Error Fix Table

| Error | Fix |
|-------|-----|
| `no field 'root_source_file'` | Use `root_module = b.createModule(.{...})` |
| `use of undefined value` | Don't do arithmetic on `undefined` |
| `type 'f32' cannot represent integer` | Use float literal: `123_456_789.0` |
| `ambiguous format string` | Use `{f}` for format methods |

## Related Skills

This file covers only critical breaking changes. For deeper guidance, load these skills:

- **`zig`** — Full std library API reference (~30 modules), build system docs, all migration details. Load when working with specific std APIs or debugging unfamiliar compile errors.
- **`zig-best-practices`** — Design patterns: tagged unions, explicit error sets, comptime validation, memory management, logging. Load when writing new Zig code or reviewing code quality.
