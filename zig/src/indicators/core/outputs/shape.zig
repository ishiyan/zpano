/// Identifies the data shape of an indicator output.
pub const Shape = enum(u8) {
    /// Holds a time stamp and a value.
    scalar = 1,
    /// Holds a time stamp and two values representing upper and lower lines of a band.
    band = 2,
    /// Holds a time stamp and an array of values representing a heat-map column.
    heatmap = 3,
    /// Holds a time stamp and an ordered, variable-length sequence of (offset, value) points.
    polyline = 4,
    /// Holds a time stamp and a variable-length set of (value, offset, strength) entries.
    levels = 5,

    pub fn asStr(self: Shape) []const u8 {
        return switch (self) {
            .scalar => "scalar",
            .band => "band",
            .heatmap => "heatmap",
            .polyline => "polyline",
            .levels => "levels",
        };
    }
};
