// -goexperiment spmd

// Test that indexing varying types produces a clear error
package spmdtest

import "lanes"

func indexVarying() {
	var v lanes.Varying[int32]
	_ = v[0] /* ERROR "cannot index.*varying types are not indexable" */

	var c lanes.Varying[uint16, 8]
	_ = c[0] /* ERROR "cannot index.*varying types are not indexable" */

	var u lanes.Varying[byte, 0]
	_ = u[0] /* ERROR "cannot index.*varying types are not indexable" */
}
