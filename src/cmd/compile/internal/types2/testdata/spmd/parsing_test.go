//go:build goexperiment.spmd

// Test parsing differences between constrained and unconstrained varying
package spmdtest

func testParsingDifferences() {
	// Case 1: Array of unconstrained varying elements
	// Should be parsed as: ArrayType{Elem: SPMDType{constraint=-1, elem=int64}}
	var case1 [16]varying int64
	_ = case1
	
	// Case 2: Constrained varying scalar
	// Should be parsed as: varying[4] (constrained) applied to int64
	var case2 varying[4] int64
	_ = case2
	
	// Case 3: Universal constrained varying scalar  
	// Should be parsed as: varying[] (universal constraint) applied to int64
	var case3 varying[] int64
	_ = case3
}