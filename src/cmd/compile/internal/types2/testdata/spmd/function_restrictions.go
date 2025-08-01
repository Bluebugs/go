//go:build goexperiment.spmd

// Test SPMD function restrictions and validation
package spmdtest

import "lanes"

// ERROR "public functions cannot have varying parameters"
func PublicSPMDFunc(data varying int) varying int {
	return data * 2
}

// ERROR "public functions cannot return varying types"
func PublicVaryingReturn() varying int {
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

// ERROR "functions with varying parameters cannot contain go for loops"
func invalidNestedGoFor(data varying int) varying int {
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
	// ERROR "lanes.Index() can only be called in SPMD context"
	idx := lanes.Index()
	
	// Valid: lanes.Index() in go for loop
	go for i := range 10 {
		validIdx := lanes.Index()
		process(validIdx[0]) // Use first lane for regular function
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
	return a + varying int(b) + c[0] + varying int(d[0])
}

// ERROR "varying map keys not supported"
func testInvalidMapKeys() {
	var vKey varying string
	m := make(map[varying string]int) // This should error
	m[vKey] = 42
	_ = m
}

// ERROR "varying channel types not supported"
func testInvalidChannelTypes() {
	ch := make(chan varying int) // This should error
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
		process(int(v[0])) // Use first lane for demonstration
	default:
		// ERROR "varying types in type switch must be handled explicitly"
	}
}

// Test constrained varying validation
func testConstrainedVarying() {
	const VALID_CONSTRAINT = 4
	var invalid_constraint int = 8
	
	var a varying[4] int         // OK: compile-time constant
	var b varying[VALID_CONSTRAINT] int // OK: named constant
	
	// ERROR "constraint must be compile-time constant"
	var c varying[invalid_constraint] int
	
	// ERROR "constraint must be positive"
	var d varying[0] int
	
	// ERROR "constraint must be positive"
	var e varying[-1] int
	
	_, _, _, _, _ = a, b, c, d, e
}

// Helper function
func process(x int) {
	_ = x
}