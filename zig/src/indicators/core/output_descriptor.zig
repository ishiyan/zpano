const shape = @import("outputs/shape.zig");
const role_mod = @import("role.zig");
const pane_mod = @import("pane.zig");

/// Classifies a single indicator output for charting / discovery.
pub const OutputDescriptor = struct {
    /// Integer representation of the output enumeration of the related indicator.
    kind: i32,
    /// The data shape of this output.
    shape: shape.Shape,
    /// The semantic role of this output.
    role: role_mod.Role,
    /// The chart pane on which this output is drawn.
    pane: pane_mod.Pane,
};
