//go:build goexperiment.spmd

// Invalid SPMD syntax that should produce parse errors
// when GOEXPERIMENT=spmd is enabled
package spmdtest

// Invalid qualifier syntax
func invalidQualifiers() {
	// ERROR qualified type must specify exactly one of uniform or varying
	var a uniform varying int
	
	// ERROR qualified type must specify exactly one of uniform or varying  
	var b varying uniform int
	
	// ERROR type qualifiers must come before the type
	var c int uniform
	
	// ERROR type qualifiers must come before the type
	var d int varying
}

// Invalid go for syntax
func invalidGoFor() {
	// ERROR 'go for' requires range clause
	go for i := 0; i < 10; i++ {
		process(i)
	}
	
	// ERROR missing range expression
	go for i := range {
		process(i)
	}
	
	// Note: break and nested go for restrictions are enforced in type checking,
	// not parsing. See types2/testdata/spmd/ for those tests.
}

// Invalid constraint syntax
func invalidConstraints() {
	var n int = 4
	
	// ERROR constraint must be compile-time constant
	var a varying[n] int
	
	// ERROR constraint must be positive integer
	var b varying[0] int
	
	// ERROR constraint must be positive integer
	var c varying[-1] int
	
	_, _, _ = a, b, c
}

func process(x varying int) {
	_ = x
}