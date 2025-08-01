//go:build !goexperiment.spmd

// Code that uses "uniform" and "varying" as regular identifiers
// This should parse successfully when GOEXPERIMENT=spmd is disabled
package spmdtest

// These should be valid identifiers when SPMD is disabled
var uniform int = 42
var varying float32 = 3.14

// Using them as function names  
func uniform() int {
	return 42
}

func varying() float32 {
	return 3.14
}

// Using them as struct field names
type Config struct {
	uniform int
	varying float32
}

func main() {
	// Using uniform and varying as regular variable names
	uniform = 100
	varying = 2.71
	
	// Using them in function calls
	process(uniform, varying)
	
	// Using them as function calls
	_ = uniform()
	_ = varying()
	
	c := Config{
		uniform: 1,
		varying: 2.0,
	}
	
	_ = c
}

func process(a int, b float32) {
	_, _ = a, b
}