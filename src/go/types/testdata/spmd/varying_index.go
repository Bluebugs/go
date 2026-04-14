// -goexperiment spmd

// Test indexing on varying types: Varying[array] is allowed, others are rejected.
package spmdtest

import "lanes"

func indexVaryingNonArray() {
	var v lanes.Varying[int32]
	_ = v /* ERROR "varying types are not indexable" */ [0]
}

func indexVaryingArray() {
	var v lanes.Varying[[4]byte]
	// Indexing Varying[array] produces Varying[elemType] — no error.
	var _ lanes.Varying[byte] = v[0]
}
