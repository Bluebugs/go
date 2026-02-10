//go:build goexperiment.spmd

// Test parser differentiation between constrained and unconstrained varying
package spmdtest

import "lanes"

func testParserDifferentiation() {
	// These should be parsed differently:

	// Case 1: Array of unconstrained varying elements
	// Expected: ArrayType{Elem: lanes.Varying[int64]}
	var unconstrained [16]lanes.Varying[int64]

	// Case 2: Constrained varying with scalar element type
	// Expected: lanes.Varying[int64, 4]
	var constrained lanes.Varying[int64, 4]

	// Case 3: Universal constrained varying
	// Expected: lanes.Varying[int64, 0]
	var universal lanes.Varying[int64, 0]

	// Case 4: Array of constrained varying with scalar element type
	// Expected: [5]lanes.Varying[int64, 4]
	var constrainedArray [5]lanes.Varying[int64, 4]

	_ = unconstrained
	_ = constrained
	_ = universal
	_ = constrainedArray
}
