// Package signalensemble provides weighted blending of multiple signal sources.
// Combines n independent signal sources (each producing values in [0, 1])
// into a single blended confidence using one of seven aggregation methods.
// Supports delayed feedback and online weight learning.
package signalensemble
