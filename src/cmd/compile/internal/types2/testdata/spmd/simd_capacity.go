//go:build goexperiment.spmd

// Test SIMD register capacity validation
package spmdtest

import "lanes"

// Test go for loop capacity constraints
func testGoForCapacityConstraints() {
	// Valid: capacity matches typical SIMD width
	go for i := range 16 {
		var data varying int32   // Should fit in SIMD128 (4 lanes)
		process(data)
		_ = i
	}
	
	// Valid: multiple small types
	go for i := range 8 {
		var a varying int32      // 4 bytes * 4 lanes = 16 bytes
		var b varying float32    // 4 bytes * 4 lanes = 16 bytes
		process(a + varying int32(b))
		_ = i
	}
	
	go for i := range 32 {
		var a varying int64      // 8 bytes * 4 lanes = 32 bytes (over SIMD128)
		var b varying int64      // 8 bytes * 4 lanes = 32 bytes (total 64 bytes)
		process(a + b)
		_ = i
	}
	
	go for i := range 16 {
		var data varying[8] float64
		process(data)
		_ = i
	}
}

// Test SPMD function capacity constraints
func testSPMDFunctionCapacity() {
	// Valid: single varying parameter fits
	var data varying int32
	result := validSPMDFunc(data)
	_ = result
	
	// Valid: multiple small parameters
	var a varying int32
	var b varying float32
	result2 := validMultiParamFunc(a, b)
	_ = result2
	
	var large1 [16]varying int64
	var large2 [16]varying float64
	invalidLargeFunc(large1, large2)

	var large3 [4]varying[32] int64 // ERROR "constrained varying capacity exceeded"
	_ = large3
}

// Valid SPMD functions within capacity
func validSPMDFunc(data varying int32) varying int32 {
	return data * 2
}

func validMultiParamFunc(a varying int32, b varying float32) varying int32 {
	return a + varying int32(b)
}

func invalidLargeFunc(a [16]varying int64, b [16]varying float64) [16]varying int64 {
	r := [16]varying int64{}
	for i := range a {
		r[i] = a[i] + varying int64(b[i])
	}
	return r
}

// Test constrained varying capacity
func testConstrainedVaryingCapacity() {
	// Valid: constraint within capacity
	var data varying[4] int32     // 4 elements * 4 bytes = 16 bytes
	process(data)

	// Valid: smaller constraint  
	var small varying[2] int64    // 2 elements * 8 bytes = 16 bytes
	process(small)

	var array [4]varying[4] int32 // 4 elements * 4 bytes * 4 lanes = 4 * 16 bytes
	_ = array

	var large varying[32] int32   // ERROR "constrained varying capacity exceeded"  
	process(large)

	var huge varying[8] int64
	process(huge)
}

// Test mixed capacity scenarios
func testMixedCapacityScenarios() {
	go for i := range[16] 16 {
		// Valid: careful capacity management
		var small varying[16] int8       // 1 byte * 16 = 16 bytes (under 64 byte limit)
		process(small)

		var medium varying[64] int16     // ERROR "constrained varying capacity exceeded"
		var large varying[64] int32      // ERROR "constrained varying capacity exceeded"
		process(medium + varying[64] int16(large))  // ERROR "constrained varying capacity exceeded"

		_ = i
	}
}

// Test capacity with lanes.Count() constraints
func testLanesCountCapacity() {
	// Capacity should be based on actual lane count, not range size
	go for i := range 1000 {  // Large range, but processed in chunks
		// Valid: lane count determines capacity, not range size
		var data varying int32   // lanes.Count() * 4 bytes (e.g., 4 * 4 = 16 bytes)
		laneCount := lanes.Count(data)
		if laneCount > 0 {
			process(data)
		}
		_ = i
	}
}

// Helper function
func process(x varying[] int) {
	_ = x
}