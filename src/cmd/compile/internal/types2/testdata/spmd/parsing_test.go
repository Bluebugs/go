//go:build goexperiment.spmd

// Test parsing of lanes.Varying[T] and lanes.Varying[T, N] types
package spmdtest

import "lanes"

func testParsingDifferences() {
	// Case 1: Array of unconstrained varying elements
	// Should be parsed as: ArrayType{Elem: lanes.Varying[int64]}
	var case1 [16]lanes.Varying[int64]
	_ = case1

	// Case 2: Constrained varying scalar
	// Should be parsed as: lanes.Varying[int64, 4]
	var case2 lanes.Varying[int64, 4]
	_ = case2

	// Case 3: Universal constrained varying scalar
	// Should be parsed as: lanes.Varying[int64, 0]
	var case3 lanes.Varying[int64, 0]
	_ = case3
}
