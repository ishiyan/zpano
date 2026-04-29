/// Enumerates algorithms used to match the offsetting order executions
/// in a round-trip.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RoundtripMatching {
    /// Matches offsetting order executions in First In First Out order.
    Fifo = 0,
    /// Matches offsetting order executions in Last In First Out order.
    Lifo = 1,
}
