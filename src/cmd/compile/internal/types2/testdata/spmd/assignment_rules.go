//go:build goexperiment.spmd

// Test assignment rules between uniform and varying types
package spmdtest

import "lanes"

// Valid assignment patterns
func validAssignments() {
	var u_int int = 42
	var v_int lanes.Varying[int]

	// Uniform-to-varying broadcast (should be valid)
	v_int = u_int // OK: automatic broadcast

	// Same qualifier assignments
	var u_int2 int = u_int                    // OK: uniform to uniform
	var v_int2 lanes.Varying[int] = v_int     // OK: varying to varying

	// Valid function parameter passing
	processVarying(v_int)     // OK: varying to varying parameter
	processUniform(u_int)     // OK: uniform to uniform parameter
	processVarying(u_int)     // OK: uniform to varying parameter (broadcast)

	_, _, _ = u_int2, v_int2, v_int
}

// Invalid assignment patterns that should generate errors
func invalidAssignments() {
	var u_int int
	var v_int lanes.Varying[int] = 42

	u_int = v_int // ERROR "cannot use v_int (variable of type lanes.Varying[int]) as int value in assignment: cannot assign varying expression to uniform variable"

	processUniform(v_int) // ERROR "cannot use v_int (variable of type lanes.Varying[int]) as int value in argument to processUniform: cannot assign varying expression to uniform variable"

	_ = uniformReturner(v_int)

	_ = u_int
}

// Test function signature constraints
func PublicVaryingFunc(x lanes.Varying[int]) int { // ERROR "public functions cannot have varying parameters"
	return 0
}

// Private SPMD functions are allowed
func privateVaryingFunc(x lanes.Varying[int]) lanes.Varying[int] {
	return x * 2
}

func invalidSPMDFunction(x lanes.Varying[int]) lanes.Varying[int] { // ERROR "functions with varying parameters cannot contain go for loops"
	go for i := range 10 {
		x += i
	}
	return x
}

// Test multiple assignment rules
func multipleAssignments() {
	var u1, u2 int
	var v1, v2 lanes.Varying[int]

	// Valid multiple assignments
	u1, u2 = 1, 2              // OK: uniform literals to uniform
	v1, v2 = u1, u2            // OK: uniform to varying (broadcast)
	v1, v2 = v1, v2            // OK: varying to varying

	u1, u2 = v1, v2            // ERROR "cannot use v1 (variable of type lanes.Varying[int]) as int value in assignment: cannot assign varying expression to uniform variable"
	                           // ERROR "cannot use v2 (variable of type lanes.Varying[int]) as int value in assignment: cannot assign varying expression to uniform variable"

	// Mixed assignments
	u1, v1 = v1, u1            // ERROR "cannot use v1 (variable of type lanes.Varying[int]) as int value in assignment: cannot assign varying expression to uniform variable"
}

// Helper functions for testing
func processUniform(x int) int {
	return x
}

func processVarying(x lanes.Varying[int]) lanes.Varying[int] {
	return x
}

func uniformReturner(x lanes.Varying[int]) int {
	return 42
}
