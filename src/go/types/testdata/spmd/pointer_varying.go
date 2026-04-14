// -goexperiment spmd

// Test file for pointer-to-varying-struct field access and varying pointer dereference.
// Field access through *Varying[Struct] returns Varying[fieldType], allowing
// assignment of varying values to struct fields via varying pointer.
// *Varying[*T] (deref of varying pointer vector) produces Varying[T] for scatter/gather.
package spmdtest

import "lanes"

type IntPoint struct{ X, Y int }

func structPointers() {
	var v lanes.Varying[int]
	_ = v
	points := [4]IntPoint{}
	go for i := range 4 {
		pointPtr := &points[i] // &points[varyingIndex] gives Varying[*IntPoint]
		// Field access through Varying[*IntPoint] is deferred (not yet implemented).
		// For now, just verify the pointer is Varying[*IntPoint].
		_ = pointPtr
	}
	_ = points
}

// Uniform pointer field access must still produce uniform field type (no change).
func uniformPtrField() {
	p := IntPoint{1, 2}
	ptr := &p    // *IntPoint (uniform)
	ptr.X = 42   // int assignment, must compile without error
	ptr.Y = 24   // int assignment, must compile without error
	_, _ = ptr.X, ptr.Y
}

// varyingPtrDeref verifies that *Varying[*T] produces Varying[T], enabling
// scatter (LHS) and gather (RHS) through per-lane pointer vectors.
func varyingPtrDeref() {
	data := [4]int{1, 2, 3, 4}
	var result lanes.Varying[int]
	go for i := range 4 {
		var ptr lanes.Varying[*int] = &data[i]
		// Scatter: *Varying[*int] = Varying[int] (per-lane store through varying ptrs)
		*ptr = lanes.Index() * 10
		// Gather: Varying[int] = *Varying[*int] (per-lane load through varying ptrs)
		result = *ptr
	}
	_ = result
}
