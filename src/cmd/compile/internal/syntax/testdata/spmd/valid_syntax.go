//go:build goexperiment.spmd

// Valid SPMD syntax that should parse successfully
// when GOEXPERIMENT=spmd is enabled
package spmdtest

import (
	"lanes"
	"reduce"
)

// Valid uniform and varying variable declarations
func validDeclarations() {
	var a uniform int = 42
	var b varying float32 = 3.14
	var c varying[4] int
	var d varying[] byte
	
	// Valid function parameter types
	func localFunc(x uniform int, y varying float32) varying int {
		return y + varying float32(x)
	}
	
	_ = localFunc
	_, _, _, _ = a, b, c, d
}

// Valid go for loop syntax
func validGoFor() {
	// Basic go for with range
	go for i := range 10 {
		process(i)
	}
	
	// go for with slice
	data := []int{1, 2, 3, 4}
	go for _, value := range data {
		process(value)
	}
	
	// go for with constrained range
	go for i := range[4] 16 {
		process(i)
	}
}

// Valid constrained varying types
func validConstraints() {
	const LANES = 4
	
	var a varying[4] int
	var b varying[LANES] float32
	var c varying[] byte // universal constraint
	
	_, _, _ = a, b, c
}

// Valid built-in function usage
func validBuiltins() {
	go for i := range 8 {
		var data varying int = i
		
		// lanes functions
		count := lanes.Count(data)
		index := lanes.Index()
		broadcast := lanes.Broadcast(42, 0)
		rotated := lanes.Rotate(data, 1)
		
		// reduce functions  
		sum := reduce.Add(data)
		all := reduce.All(data > 5)
		any := reduce.Any(data > 5)
		
		_, _, _, _, _, _, _ = count, index, broadcast, rotated, sum, all, any
	}
}

// Valid control flow
func validControlFlow() {
	go for i := range 10 {
		if i > 5 {
			continue // valid in go for
		}
		// Note: break is not allowed in go for loops
		
		process(i)
	}
}

func process(x varying int) {
	// Valid SPMD function with varying parameter
	_ = x
}