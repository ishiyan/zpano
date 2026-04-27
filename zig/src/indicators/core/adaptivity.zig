/// Classifies whether an indicator adapts its parameters to market conditions.
pub const Adaptivity = enum(u8) {
    /// Fixed parameters.
    static_ = 1,
    /// Adapts parameters to market conditions.
    adaptive = 2,

    pub fn asStr(self: Adaptivity) []const u8 {
        return switch (self) {
            .static_ => "static",
            .adaptive => "adaptive",
        };
    }
};
