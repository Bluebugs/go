// -goexperiment spmd

package spmdtest

import "lanes"

func takeUnconstrained(v lanes.Varying[int32]) {}

// VALID: constrained can pass to unconstrained parameter
func testPassConstrained() {
	var c4 lanes.Varying[int32, 4]
	takeUnconstrained(c4) // OK
	var c8 lanes.Varying[int32, 8]
	takeUnconstrained(c8) // OK
}

// VALID: assign constrained to unconstrained variable
func testAssignConstrained() {
	var c lanes.Varying[int32, 4]
	var u lanes.Varying[int32] = c // OK
	_ = u
}

// VALID: explicit conversion from constrained to unconstrained
func testExplicitConversion() {
	var c4 lanes.Varying[int32, 4]
	var u lanes.Varying[int32] = lanes.Varying[int32](c4) // OK
	_ = u
}

// VALID: generic type inference with constrained argument
func genericTakeVarying[T any](v lanes.Varying[T]) lanes.Varying[T] {
	return v
}

func testGenericInference() {
	var c4 lanes.Varying[int32, 4]
	result := genericTakeVarying(c4) // OK: infers T=int32
	_ = result
}

// STILL INVALID: two different specific constraints
func testMismatchedConstraints(a lanes.Varying[int32, 4], b lanes.Varying[int32, 8]) {
	a = b // ERROR "cannot use"
}
