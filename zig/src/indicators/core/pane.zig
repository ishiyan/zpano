/// Identifies the chart pane an indicator output is drawn on.
pub const Pane = enum(u8) {
    /// The primary price pane.
    price = 1,
    /// A dedicated sub-pane for this indicator.
    own = 2,
    /// Drawing on the parent indicator's pane.
    overlay_on_parent = 3,

    pub fn asStr(self: Pane) []const u8 {
        return switch (self) {
            .price => "price",
            .own => "own",
            .overlay_on_parent => "overlayOnParent",
        };
    }
};
