// -goexperiment spmd

// Test file for pointer-to-varying-struct field access.
// Field access through *Varying[Struct] returns Varying[fieldType], allowing
// assignment of varying values to struct fields via varying pointer.
package spmdtest

import "lanes"

type IntPoint struct{ X, Y int }

func structPointers() {
	var v lanes.Varying[int]
	points := [4]IntPoint{}
	go for i := range 4 {
		pointPtr := &points[i] // &points[varyingIndex] gives *Varying[IntPoint]
		// Field access through *Varying[IntPoint] returns Varying[int], so varying assignment is valid.
		pointPtr.X = v
		pointPtr.Y = v
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
