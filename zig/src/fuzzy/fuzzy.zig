/// Fuzzy logic library: membership functions, operators, and defuzzification.
///
/// Re-exports all public symbols from the three sub-modules so that
/// consumers can `@import("fuzzy")` instead of importing each piece separately.

pub const membership = @import("membership.zig");
pub const operators = @import("operators.zig");
pub const defuzzify = @import("defuzzify.zig");
