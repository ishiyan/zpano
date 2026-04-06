/// RoundtripMatching represents the matching strategy for round-trips.
pub const RoundtripMatching = enum(u1) {
    fifo = 0,
    lifo = 1,
};
