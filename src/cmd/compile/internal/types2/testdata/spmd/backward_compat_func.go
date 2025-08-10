// Code that uses "uniform" and "varying" as regular identifiers
// This should parse and compile successfully in all circumstances
package spmdTest

// Using uniform and varying as variable names
func uniform() int {
	return 42
}

func varying() float32 {
	return 3.14
}

// Using them as local variable names
func main() {
	_ = uniform()
	_ = varying()
}
