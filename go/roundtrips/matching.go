package roundtrips

// RoundtripMatching enumerates algorithms used to match the offsetting
// order executions in a round-trip.
type RoundtripMatching int

const (
	// FIFO matches offsetting order executions in First In First Out order.
	FIFO RoundtripMatching = iota
	// LIFO matches offsetting order executions in Last In First Out order.
	LIFO
)
