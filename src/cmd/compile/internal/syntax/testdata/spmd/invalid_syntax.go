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

// Invalid constrained varying syntax - no space allowed between qualifier and brackets
func invalidConstrainedSyntax() {
	// ERROR constrained varying requires space before type
	var a varying[]int64
	
	// ERROR constrained varying requires space before type
	var b varying[16]int64
	
	// ERROR constrained varying requires space before type
	var c varying[4]float32
	
	// ERROR constrained varying requires space before type  
	var d varying[]byte
	
	_, _, _, _ = a, b, c, d
}

// Valid constrained vs unconstrained varying syntax - demonstrating correct spacing
func validConstrainedSyntax() {
	// VALID: Constrained varying with space before type
	var constrained1 varying[16] int64    // constraint=16, elem=int64
	var constrained2 varying[4] float32   // constraint=4, elem=float32
	var universal varying[] byte          // constraint=0 (universal), elem=byte
	
	// VALID: Unconstrained varying with array element type (space after varying)
	var unconstrained1 varying [16]int64  // constraint=-1, elem=[16]int64
	var unconstrained2 varying [4]float32 // constraint=-1, elem=[4]float32
	var unconstrained3 varying []byte     // constraint=-1, elem=[]byte
	
	_, _, _, _, _, _ = constrained1, constrained2, universal, unconstrained1, unconstrained2, unconstrained3
}

func process(x varying int) {
	_ = x
}