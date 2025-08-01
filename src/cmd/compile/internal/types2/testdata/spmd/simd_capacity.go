//go:build goexperiment.spmd

// Test SIMD register capacity validation
package spmdtest

import "lanes"

// Test go for loop capacity constraints
func testGoForCapacityConstraints() {
	// Valid: capacity matches typical SIMD width
	go for i := range 16 {
		var data varying int32   // Should fit in SIMD128 (4 lanes)
		process(int(data))
		_ = i
	}
	
	// Valid: multiple small types
	go for i := range 8 {
		var a varying int32      // 4 bytes * 4 lanes = 16 bytes
		var b varying float32    // 4 bytes * 4 lanes = 16 bytes
		process(int(a + varying int32(b)))
		_ = i
	}
	
	// ERROR "SIMD register capacity exceeded"
	go for i := range 32 {
		var a varying int64      // 8 bytes * 4 lanes = 32 bytes (over SIMD128)
		var b varying int64      // 8 bytes * 4 lanes = 32 bytes (total 64 bytes)
		process(int(a + b))
		_ = i
	}
	
	// ERROR "SIMD register capacity exceeded"
	go for i := range 16 {
		var data [8]varying float64  // 8*8*4 = 256 bytes (way over limit)
		process(int(data[0]))
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
	
	// ERROR "SPMD function capacity exceeded"
	var large1 varying [16]int64
	var large2 varying [16]float64
	invalidLargeFunc(large1, large2)
}

// Valid SPMD functions within capacity
func validSPMDFunc(data varying int32) varying int32 {
	return data * 2
}

func validMultiParamFunc(a varying int32, b varying float32) varying int32 {
	return a + varying int32(b)
}

// ERROR "SPMD function parameter capacity exceeded"
func invalidLargeFunc(a varying [16]int64, b varying [16]float64) varying int64 {
	return a[0] + varying int64(b[0])
}

// Test constrained varying capacity
func testConstrainedVaryingCapacity() {
	// Valid: constraint within capacity
	var data varying[4] int32     // 4 elements * 4 bytes = 16 bytes
	process(int(data[0]))
	
	// Valid: smaller constraint
	var small varying[2] int64    // 2 elements * 8 bytes = 16 bytes
	process(int(small[0]))
	
	// ERROR "constrained varying capacity exceeded"
	var large varying[32] int32   // 32 elements * 4 bytes = 128 bytes
	process(int(large[0]))
	
	// ERROR "constrained varying capacity exceeded"
	var huge varying[8] int64     // 8 elements * 8 bytes = 64 bytes
	process(int(huge[0]))
}

// Test mixed capacity scenarios
func testMixedCapacityScenarios() {
	go for i := range 16 {
		// Valid: careful capacity management
		var small varying int8       // 1 byte * 16 lanes = 16 bytes (max for SIMD128)
		process(int(small))
		
		// ERROR "SIMD register capacity exceeded"
		var medium varying int16     // 2 bytes * 16 lanes = 32 bytes (over limit)
		var large varying int32      // 4 bytes * 16 lanes = 64 bytes (way over)
		process(int(medium + varying int16(large)))
		
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
			process(int(data))
		}
		_ = i
	}
}

// Helper function
func process(x int) {
	_ = x
}