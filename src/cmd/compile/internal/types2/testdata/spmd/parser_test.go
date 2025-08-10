//go:build goexperiment.spmd

// Test parser differentiation between constrained and unconstrained varying
package spmdtest

func testParserDifferentiation() {
	// These should be parsed differently:
	
	// Case 1: Array of unconstrained varying elements
	// Expected: ArrayType{Elem: SPMDType{constraint=-1, elem=int64}}
	var unconstrained [16]varying int64
	
	// Case 2: Constrained varying with scalar element type  
	// Expected: constraint=4, elem=int64
	var constrained varying[4] int64
	
	// Case 3: Universal constrained varying
	// Expected: constraint=0, elem=int64
	var universal varying[] int64

	// Case 4: Constrained varying with scalar element type  
	// Expected: constraint=4, elem=int64
	var constrainedArray [5]varying[4] int64

	var bad varying[4] [32]int32 // ERROR "constrained varying cannot have array element type; use '[n]varying[c] T' instead of 'varying[c] [n]T'"

	_ = unconstrained
	_ = constrained  
	_ = universal
	_ = constrainedArray
	_ = bad
}