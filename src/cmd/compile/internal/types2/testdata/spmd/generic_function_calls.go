//go:build goexperiment.spmd

// Test for known compiler bug: function calls involving generic SPMD functions
// cause "internal compiler error: panic: unreachable"
//
// This test documents the bug and should be enabled once the issue is fixed.
// Currently all function calls to/from generic SPMD functions are commented out.

package spmdtest

import "lanes"

// Type constraint for testing
type numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

// Basic generic SPMD function definitions - these work fine
func add[T numeric](data lanes.Varying[T, 0]) T {
	var zero T
	return zero
}

func multiply[T numeric](data lanes.Varying[T, 0], factor T) lanes.Varying[T, 0] {
	// Implementation would multiply each lane by factor
	return data
}

// Non-generic SPMD functions - these work fine
func countTrue(data lanes.Varying[bool, 0]) int {
	// Would count true values across lanes
	return 0
}

func isPositive(data lanes.Varying[int, 0]) lanes.Varying[bool, 0] {
	// Would check if each lane value is positive
	var result lanes.Varying[bool, 0]
	return result
}

// KNOWN BUGS - all of these cause "internal compiler error: panic: unreachable":
// Uncomment each section once the corresponding bug is fixed

// BUG 1: Generic SPMD function calling another generic SPMD function
func sum[T numeric](data lanes.Varying[T, 0]) T {
	return add[T](data)  // ERROR: Generic calling generic with SPMD types
}

func scale[T numeric](data lanes.Varying[T, 0], factor T) lanes.Varying[T, 0] {
	return multiply[T](data, factor)  // ERROR: Generic calling generic with SPMD types
}

/*
// BUG 2: Non-generic SPMD function calling generic SPMD function
func SumInts(data lanes.Varying[int, 0]) int {
	return add(data)  // ERROR: Non-generic calling generic with SPMD types
}

func ScaleFloats(data lanes.Varying[float32, 0]) lanes.Varying[float32, 0] {
	return multiply(data, 2.0)  // ERROR: Non-generic calling generic with SPMD types
}
*/

/*
// BUG 3: Generic SPMD function calling non-generic SPMD function
func Process[T any](data lanes.Varying[T, 0], flags lanes.Varying[bool, 0]) int {
	return countTrue(flags)  // ERROR: Generic calling non-generic with SPMD types
}
*/

// WORKING PATTERNS - these compile successfully:

// Pattern 1: Generic functions without SPMD types calling each other
func regularAdd[T numeric](a T, b T) T {
	return a // simplified
}

func regularSum[T numeric](values []T) T {
	return regularAdd(values[0], values[1])  // Works: no SPMD types involved
}

// Pattern 2: Non-generic SPMD functions calling each other
func helperFunction(data lanes.Varying[bool, 0]) bool {
	return true
}

func callerFunction(data lanes.Varying[bool, 0]) bool {
	return helperFunction(data)  // Works: both non-generic with SPMD types
}

// Pattern 3: Functions with SPMD types that don't call other functions
func standalone[T numeric](data lanes.Varying[T, 0]) T {
	var result T
	// Direct implementation without function calls
	return result
}
