/// RoundtripGrouping represents the grouping strategy for round-trips.
pub const RoundtripGrouping = enum(u2) {
    fill_to_fill = 0,
    flat_to_flat = 1,
    flat_to_reduced = 2,
};
