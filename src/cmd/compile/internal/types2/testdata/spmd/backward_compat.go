// Code that uses "uniform" and "varying" as regular identifiers
// This should parse and compile successfully in all circumstances
// With package-based types, "uniform" and "varying" are no longer keywords
package spmdTest

// Using uniform and varying as variable names
func hideVar() {
	var uniform int = 42
	var varying float32 = 3.14

	_ = uniform
	_ = varying
}

func hideFunc() {
	// Using them as function names
	uniform := func() int {
		return 42
	}

	varying := func() float32 {
		return 3.14
	}

	_ = uniform()
	_ = varying()
}

// Using them as struct field names
type Config struct {
	uniform int
	varying float32
}

// Using them as local variable names
func main() {
	// Local variables with these names
	uniform := 100
	varying := 2.71

	// Using them in function calls
	processBC(uniform, varying)

	// Using them as function calls
	hideFunc()
	hideVar()

	// Using them in struct literals and field access
	c := Config{
		uniform: 1,
		varying: 2.0,
	}

	// Accessing struct fields
	_ = c.uniform
	_ = c.varying
}

func processBC(a int, b float64) {
	_, _ = a, b
}
