// -goexperiment spmd

// Test array-to-constrained-varying conversion rules
package spmdtest

import "lanes"

// Valid: matching length + elem type
func validConversions() {
	_ = lanes.Varying[int32, 4]([4]int32{1, 2, 3, 4})
	_ = lanes.Varying[byte, 8]([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	_ = lanes.Varying[float32, 4]([4]float32{1.0, 2.0, 3.0, 4.0})
	_ = lanes.Varying[uint16, 2]([2]uint16{10, 20})
}

// Invalid: wrong length
func wrongLength() {
	_ = lanes.Varying[int32, 4]([3]int32{1, 2, 3})                    /* ERROR "cannot convert" */
	_ = lanes.Varying[int32, 4]([8]int32{1, 2, 3, 4, 5, 6, 7, 8})    /* ERROR "cannot convert" */
}

// Invalid: wrong elem type
func wrongElemType() {
	_ = lanes.Varying[int32, 4]([4]float32{1, 2, 3, 4})    /* ERROR "cannot convert" */
}

// Invalid: unconstrained target (constraint == -1)
func unconstrainedTarget() {
	_ = lanes.Varying[int32]([4]int32{1, 2, 3, 4})    /* ERROR "cannot convert" */
}
