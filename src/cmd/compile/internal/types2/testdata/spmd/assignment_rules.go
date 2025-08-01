//go:build goexperiment.spmd

// Test assignment rules between uniform and varying types
package spmdtest

// Valid assignment patterns
func validAssignments() {
	var u_int uniform int = 42
	var v_int varying int
	
	// Uniform-to-varying broadcast (should be valid)
	v_int = u_int // OK: automatic broadcast
	
	// Same qualifier assignments
	var u_int2 uniform int = u_int  // OK: uniform to uniform
	var v_int2 varying int = v_int  // OK: varying to varying
	
	// Valid function parameter passing
	processVarying(v_int)     // OK: varying to varying parameter
	processUniform(u_int)     // OK: uniform to uniform parameter
	processVarying(u_int)     // OK: uniform to varying parameter (broadcast)
	
	_, _, _ = u_int2, v_int2, v_int
}

// Invalid assignment patterns that should generate errors
func invalidAssignments() {
	var u_int uniform int
	var v_int varying int = 42
	
	// ERROR "cannot assign varying expression to uniform variable"
	u_int = v_int
	
	// ERROR "cannot pass varying argument to uniform parameter"
	processUniform(v_int)
	
	// ERROR "cannot return varying expression from uniform function"
	_ = uniformReturner(v_int)
	
	_ = u_int
}

// Test function signature constraints
// ERROR "public functions cannot have varying parameters"
func PublicVaryingFunc(x varying int) int {
	return int(x)
}

// Private SPMD functions are allowed
func privateVaryingFunc(x varying int) varying int {
	return x * 2
}

// ERROR "functions with varying parameters cannot contain go for loops"
func invalidSPMDFunction(x varying int) varying int {
	go for i := range 10 {
		x += varying int(i)
	}
	return x
}

// Test multiple assignment rules
func multipleAssignments() {
	var u1, u2 uniform int
	var v1, v2 varying int
	
	// Valid multiple assignments
	u1, u2 = 1, 2              // OK: uniform literals to uniform
	v1, v2 = u1, u2            // OK: uniform to varying (broadcast)
	v1, v2 = v1, v2            // OK: varying to varying
	
	// ERROR "cannot assign varying expression to uniform variable"
	u1, u2 = v1, v2
	
	// Mixed assignments
	// ERROR "cannot assign varying expression to uniform variable"
	u1, v1 = v1, u1
}

// Helper functions for testing
func processUniform(x uniform int) uniform int {
	return x
}

func processVarying(x varying int) varying int {
	return x
}

func uniformReturner(x varying int) uniform int {
	// This should be an error context, not here specifically
	return 42
}