// -goexperiment spmd

// Test file for pointer-to-varying-struct field access.
// Before fix: field access through *Varying[Struct] returns uniform field type,
// causing assignment errors when trying to assign varying values.
// The annotations below capture the current (buggy) behavior.
// After fix: these annotations will be removed when field access returns varying.
package spmdtest

import "lanes"

type IntPoint struct{ X, Y int }

func structPointers() {
	var v lanes.Varying[int]
	points := [4]IntPoint{}
	go for i := range 4 {
		pointPtr := &points[i] // &points[varyingIndex] gives *Varying[IntPoint]
		// Currently: pointPtr.X is uniform int, so varying assignments fail
		pointPtr.X = v /* ERROR "cannot assign varying" */
		pointPtr.Y = v /* ERROR "cannot assign varying" */
	}
}

// Uniform pointer field access must still produce uniform field type (no change).
func uniformPtrField() {
	p := IntPoint{1, 2}
	ptr := &p    // *IntPoint (uniform)
	ptr.X = 42   // int assignment, must compile without error
	ptr.Y = 24   // int assignment, must compile without error
	_, _ = ptr.X, ptr.Y
}
