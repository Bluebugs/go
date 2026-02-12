//go:build goexperiment.spmd

// Test SPMD function restrictions and validation
package spmdtest

import "lanes"

func /* ERROR "public functions cannot have varying parameters" */ PublicSPMDFunc(
	data lanes.Varying[int],
) lanes.Varying[int] {
	return data * 2
}

func /* ERROR "public functions cannot return varying types" */ PublicVaryingReturn() lanes.Varying[int] {
	return 42
}

// Private SPMD functions are allowed
func privateSPMDFunc(data lanes.Varying[int]) lanes.Varying[int] {
	return data * 2
}

// Non-SPMD public functions are allowed
func PublicRegularFunc(data int) int {
	return data * 2
}

func /* ERROR "functions with varying parameters cannot contain go for loops" */ invalidNestedGoFor(
	data lanes.Varying[int],
) lanes.Varying[int] {
	go for i := range 10 {
		data += i
	}
	return data
}

// Valid: functions without varying parameters can contain go for
func validGoForInNonSPMD() {
	go for i := range 10 {
		processC(i)
	}
}

// Valid: SPMD functions can call other SPMD functions
func validSPMDCalls(data lanes.Varying[int]) lanes.Varying[int] {
	result := privateSPMDFunc(data)
	return result + 1
}

// Test context restrictions for lanes.Index()
func testLanesIndexRestrictions() {
	idx := lanes /* ERROR "lanes.Index() can only be called in SPMD context" */ .Index()

	// Valid: lanes.Index() in go for loop
	go for i := range 10 {
		validIdx := lanes.Index()
		processC(validIdx) // Use first lane for regular function

		_ = i
	}

	_ = idx
}

// Test varying parameter type validation
func testVaryingParameterTypes(
	a lanes.Varying[int],          // OK
	b lanes.Varying[float32],      // OK
) lanes.Varying[int] {
	_ = b
	return a
}

// testInvalidMapKeys - not tested in go/types
// Map validation is only partially implemented in go/types (missing validateSPMDMakeType)
// This test works in types2 but is not supported here yet

// Test channels with varying types (now allowed)
func testValidChannelTypes() {
	ch := make(chan lanes.Varying[int]) // OK: channels can carry varying types
	_ = ch
}

// Test interface{} with varying
func testVaryingInterface() {
	var data lanes.Varying[int] = 42

	// OK: varying can be passed as interface{}
	var iface interface{} = data

	// Type switches with varying interface{} require explicit handling
	switch v := iface.(type) {
	case lanes.Varying[int]:
		processC(v)
	default:
		// no handling
	}
}

// Test constrained varying validation - not supported in go/types yet
// Constrained varying uses generic syntax which go/types doesn't fully support yet
// These tests are in types2 only

// Helper function
func processC(x lanes.Varying[int]) {
	_ = x
}
