//go:build goexperiment.spmd

// Test SPMD function restrictions and validation
package spmdtest

import "lanes"

func PublicSPMDFunc( // ERROR "public functions cannot have varying parameters"
	data lanes.Varying[int],
) lanes.Varying[int] {
	return data * 2
}

func PublicVaryingReturn() lanes.Varying[int] { // ERROR "public functions cannot return varying types"
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

func invalidNestedGoFor( // ERROR "functions with varying parameters cannot contain go for loops"
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
		process(i)
	}
}

// Valid: SPMD functions can call other SPMD functions
func validSPMDCalls(data lanes.Varying[int]) lanes.Varying[int] {
	result := privateSPMDFunc(data)
	return result + 1
}

// Test context restrictions for lanes.Index()
func testLanesIndexRestrictions() {
	idx := lanes.Index() // ERROR "lanes.Index() can only be called in SPMD context"

	// Valid: lanes.Index() in go for loop
	go for i := range 10 {
		validIdx := lanes.Index()
		process(validIdx) // Use first lane for regular function

		_ = i
	}

	_ = idx
}

// Test varying parameter type validation
func testVaryingParameterTypes(
	a lanes.Varying[int],          // OK
	b lanes.Varying[float32],      // OK
	c lanes.Varying[int, 4],       // OK: constrained varying
	d lanes.Varying[byte, 0],      // OK: universal constraint
) lanes.Varying[int] {
	_ = d
	return a + c
}

func testInvalidMapKeys() {
	var vKey lanes.Varying[string]
	m := make(map[lanes.Varying[string]]int) // ERROR "varying map keys not supported"
	m[vKey] = 42
	_ = m
}

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
		process(v)
	default:
		// no handling
	}
}

// Test constrained varying validation
func testConstrainedVarying() {
	var a lanes.Varying[int, 4]         // OK: compile-time constant

	var c lanes.Varying[int, -1] // ERROR "lanes.Varying constraint must be non-negative"

	_, _ = a, c
}

// Helper function
func process(x lanes.Varying[int]) {
	_ = x
}
