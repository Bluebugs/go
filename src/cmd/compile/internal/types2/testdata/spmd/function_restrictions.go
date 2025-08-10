//go:build goexperiment.spmd

// Test SPMD function restrictions and validation
package spmdtest

import "lanes"

func PublicSPMDFunc(data varying int) varying int { // ERROR "public functions cannot have varying parameters"
	return data * 2
}

func PublicVaryingReturn() varying int { // ERROR "public functions cannot return varying types"
	return 42
}

// Private SPMD functions are allowed
func privateSPMDFunc(data varying int) varying int {
	return data * 2
}

// Non-SPMD public functions are allowed
func PublicRegularFunc(data int) int {
	return data * 2
}

func invalidNestedGoFor(data varying int) varying int { // ERROR "functions with varying parameters cannot contain go for loops"
	go for i := range 10 {
		data += varying int(i)
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
func validSPMDCalls(data varying int) varying int {
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
	a varying int,           // OK
	b varying float32,       // OK
	c varying[4] int,       // OK: constrained varying
	d varying[] byte,       // OK: universal constraint
) varying int {
	_ = d
	return a + varying int(b) + c
}

func testInvalidMapKeys() {
	var vKey varying string
	m := make(map[varying string]int) // ERROR "varying map keys not supported"
	m[vKey] = 42
	_ = m
}

// Test channels with varying types (now allowed)
func testValidChannelTypes() {
	ch := make(chan varying int) // OK: channels can carry varying types
	_ = ch
}

// Test interface{} with varying
func testVaryingInterface() {
	var data varying int = 42
	
	// OK: varying can be passed as interface{}
	var iface interface{} = data
	
	// Type switches with varying interface{} require explicit handling
	switch v := iface.(type) {
	case varying int:
		process(v)
	default:
		// no handling
	}
}

// Test constrained varying validation
func testConstrainedVarying() {
	const VALID_CONSTRAINT = 4
	var invalid_constraint int = 8
	
	var a varying[4] int         // OK: compile-time constant
	var b varying[VALID_CONSTRAINT] int // OK: named constant
	
	var c varying[invalid_constraint] int // ERROR "constraint must be compile-time constant"
	
	var d varying[0] int // ERROR "constraint must be positive"
	
	var e varying[-1] int // ERROR "constraint must be positive"
	
	_, _, _, _, _ = a, b, c, d, e
}

// Helper function
func process(x varying int) {
	_ = x
}