// -goexperiment spmd

// Test that explicit same-type casts from varying to uniform are rejected,
// while cross-type element conversions remain valid.
package spmdtest

import "lanes"

func sameTypeCastRejected(i lanes.Varying[int], vf lanes.Varying[float32]) {
	// ILLEGAL: int(Varying[int]) — same element type strips varying qualifier
	_ = int(i /* ERROR "cannot explicitly convert varying value to same uniform type" */)

	// ILLEGAL: float32(Varying[float32]) — same element type
	_ = float32(vf /* ERROR "cannot explicitly convert varying value to same uniform type" */)
}

func crossTypeCastAllowed(i lanes.Varying[int], vf lanes.Varying[float32]) {
	// LEGAL: int32(Varying[int]) — different element type, produces Varying[int32]
	var _ = int32(i)

	// LEGAL: float64(Varying[float32]) — different element type, produces Varying[float64]
	var _ = float64(vf)
}
