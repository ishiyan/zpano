package roundtrips

// RoundtripGrouping enumerates algorithms used to group order executions
// into round-trips.
type RoundtripGrouping int

const (
	// FillToFill is the round-trip defined by (1) an order execution that
	// establishes or increases a position and (2) an offsetting execution
	// that reduces the position size.
	FillToFill RoundtripGrouping = iota
	// FlatToFlat is the round-trip defined by a sequence of order executions,
	// from a flat position to a non-zero position which may increase or
	// decrease in quantity, and back to a flat position.
	FlatToFlat
	// FlatToReduced is the round-trip defined by a sequence of order executions,
	// from a flat position to a non-zero position and an offsetting execution
	// that reduces the position size.
	FlatToReduced
)
