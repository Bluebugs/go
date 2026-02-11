//go:build goexperiment.spmd

// Test SIMD register capacity validation
package spmdtest

import "lanes"

// Test go for loop capacity constraints
func testGoForCapacityConstraints() {
	// Valid: capacity matches typical SIMD width
	go for i := range 16 {
		var data lanes.Varying[int32]   // Should fit in SIMD128 (4 lanes)
		processAny(data)
		_ = i
	}

	// Valid: multiple small types
	go for i := range 8 {
		var a lanes.Varying[int32]      // 4 bytes * 4 lanes = 16 bytes
		var b lanes.Varying[float32]    // 4 bytes * 4 lanes = 16 bytes
		processAny(a + lanes.Varying[int32](b))
		_ = i
	}

	go for i := range 32 {
		var a lanes.Varying[int64]      // 8 bytes * 2 lanes = 16 bytes
		var b lanes.Varying[int64]      // 8 bytes * 2 lanes = 16 bytes
		processAny(a + b)
		_ = i
	}

	go for i := range 16 {
		var data lanes.Varying[float64, 8]
		processAny(data)
		_ = i
	}
}

// Test SPMD function capacity constraints
func testSPMDFunctionCapacity() {
	// Valid: single varying parameter fits
	var data lanes.Varying[int32]
	result := validSPMDFunc(data)
	_ = result

	// Valid: multiple small parameters
	var a lanes.Varying[int32]
	var b lanes.Varying[float32]
	result2 := validMultiParamFunc(a, b)
	_ = result2

	var large1 [16]lanes.Varying[int64]
	var large2 [16]lanes.Varying[float64]
	invalidLargeFunc(large1, large2)

	var large3 [4]lanes.Varying[int64, 32] // ERROR "constrained varying capacity exceeded"
	_ = large3
}

// Valid SPMD functions within capacity
func validSPMDFunc(data lanes.Varying[int32]) lanes.Varying[int32] {
	return data * 2
}

func validMultiParamFunc(a lanes.Varying[int32], b lanes.Varying[float32]) lanes.Varying[int32] {
	return a + lanes.Varying[int32](b)
}

func invalidLargeFunc(a [16]lanes.Varying[int64], b [16]lanes.Varying[float64]) [16]lanes.Varying[int64] {
	r := [16]lanes.Varying[int64]{}
	for i := range a {
		r[i] = a[i] + lanes.Varying[int64](b[i])
	}
	return r
}

// Test constrained varying capacity
func testConstrainedVaryingCapacity() {
	// Valid: constraint within capacity
	var data lanes.Varying[int32, 4]     // 4 elements * 4 bytes = 16 bytes
	processAny(data)

	// Valid: smaller constraint
	var small lanes.Varying[int64, 2]    // 2 elements * 8 bytes = 16 bytes
	processAny(small)

	var array [4]lanes.Varying[int32, 4] // 4 elements * 4 bytes * 4 lanes = 4 * 16 bytes
	_ = array

	var large lanes.Varying[int32, 32]   // ERROR "constrained varying capacity exceeded"
	processAny(large)

	var huge lanes.Varying[int64, 8]
	processAny(huge)
}

// Test mixed capacity scenarios
func testMixedCapacityScenarios() {
	go for i := range[16] 16 {
		// Valid: careful capacity management
		var small lanes.Varying[int8, 16]       // 1 byte * 16 = 16 bytes (under 64 byte limit)
		processAny(small)

		var medium lanes.Varying[int16, 64]     // ERROR "constrained varying capacity exceeded"
		var large lanes.Varying[int32, 64]      // ERROR "constrained varying capacity exceeded"
		processAny(medium + lanes.Varying[int16, 64](large))  // ERROR "constrained varying capacity exceeded"

		_ = i
	}
}

// Test capacity with lanes.Count() constraints
func testLanesCountCapacity() {
	// Capacity should be based on actual lane count, not range size
	go for i := range 1000 {  // Large range, but processed in chunks
		// Valid: lane count determines capacity, not range size
		var data lanes.Varying[int32]   // lanes.Count() * 4 bytes (e.g., 4 * 4 = 16 bytes)
		laneCount := lanes.Count(data)
		if laneCount > 0 {
			processAny(data)
		}
		_ = i
	}
}

// Test lane count consistency across mixed element types
func testLaneCountConsistency() {
	// Mixed element sizes: compiler picks minimum lane count (2)
	go for i := range 16 {
		var a lanes.Varying[byte]    // 16 lanes naturally
		var b lanes.Varying[int64]   // 2 lanes naturally
		// Effective lane count: 2 (determined by largest element)
		_ = a
		_ = b
		_ = i
	}
}

// Helper function
func processAny(x lanes.Varying[int, 0]) {
	_ = x
}
