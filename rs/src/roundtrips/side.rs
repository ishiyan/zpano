/// Enumerates the sides of a round-trip.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RoundtripSide {
    Long = 0,
    Short = 1,
}
