//go:build goexperiment.spmd

// Invalid SPMD syntax that should produce parse errors
// when GOEXPERIMENT=spmd is enabled
package spmdtest

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

// Note: With package-based types (lanes.Varying[T]), invalid type syntax
// is now handled by the type checker, not the parser. Invalid constraint
// values (e.g., negative numbers) are caught during type checking.
// See types2/testdata/spmd/ for type checking error tests.

func process(x lanes.Varying[int]) {
	_ = x
}
