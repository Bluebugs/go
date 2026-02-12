// -goexperiment spmd

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

	u_int = v_int /* ERROR "cannot assign varying expression to uniform variable" */

	processUniform(v_int /* ERROR "cannot assign varying expression to uniform variable" */)

	_ = uniformReturner(v_int)

	_ = u_int
}

// Test multiple assignment rules
func multipleAssignments() {
	var u1, u2 int
	var v1, v2 lanes.Varying[int]

	// Valid multiple assignments
	u1, u2 = 1, 2              // OK: uniform literals to uniform
	v1, v2 = u1, u2            // OK: uniform to varying (broadcast)
	v1, v2 = v1, v2            // OK: varying to varying

	u1, u2 = v1 /* ERROR "cannot assign varying expression to uniform variable" */, v2 /* ERROR "cannot assign varying expression to uniform variable" */

	// Mixed assignments
	u1, v1 = v1 /* ERROR "cannot assign varying expression to uniform variable" */, u1

	_, _, _, _ = u1, u2, v1, v2
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
