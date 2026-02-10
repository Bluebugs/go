// Code that uses "uniform" and "varying" as regular identifiers
// This should parse and compile successfully in all circumstances
// With package-based types, "uniform" and "varying" are no longer keywords
package spmdTest

// Using uniform and varying as function names
func uniform() int {
	return 42
}

func varying() float32 {
	return 3.14
}

// Using them as local variable names
func mainBCF() {
	_ = uniform()
	_ = varying()
}
