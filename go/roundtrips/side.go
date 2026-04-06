package roundtrips

// RoundtripSide enumerates the sides of a round-trip.
type RoundtripSide int

const (
	// Long represents a long round-trip.
	Long RoundtripSide = iota
	// Short represents a short round-trip.
	Short
)
