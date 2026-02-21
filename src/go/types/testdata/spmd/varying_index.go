// -goexperiment spmd

// Test that indexing varying types produces a clear error
package spmdtest

import "lanes"

func indexVarying() {
	var v lanes.Varying[int32]
	_ = v /* ERROR "varying types are not indexable" */ [0]
}
