//go:build goexperiment.spmd

// Test control flow restrictions in SPMD contexts
package spmdtest

import "reduce"

// Test go for loop restrictions
func testGoForRestrictions() {
	// Valid go for loops
	go for i := range 10 {
		if i > 5 {
			continue // OK: continue allowed in go for
		}
		process(i)
	}
	
	// ERROR "break statement not allowed in go for loop"
	go for i := range 10 {
		if i > 5 {
			break
		}
		process(i)
	}
	
	// ERROR "nested go for loops not allowed"
	go for i := range 10 {
		go for j := range 5 {
			process(i + j)
		}
	}
}

// Test that regular for loops work normally
func testRegularForLoops() {
	// Regular for loops should work normally everywhere
	for i := 0; i < 10; i++ {
		if i > 5 {
			break    // OK: break allowed in regular for
		}
		if i%2 == 0 {
			continue // OK: continue allowed in regular for
		}
		process(i)
	}
	
	// Nested regular for loops are fine
	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			if i+j > 10 {
				break // OK: break in inner regular for
			}
			process(i + j)
		}
	}
}

// Test mixed control flow (go for with regular for inside)
func testMixedControlFlow() {
	go for i := range 10 {
		// Regular for loop inside go for is allowed
		for j := 0; j < 5; j++ {
			if j > 2 {
				break // OK: break in regular for inside go for
			}
			process(int(i) + j)
		}
		
		// But another go for is not allowed
		// ERROR "nested go for loops not allowed"
		go for k := range 3 {
			process(int(i) + k)
		}
	}
}

// Test conditional control flow with varying
func testVaryingControlFlow() {
	go for i := range 8 {
		var condition varying bool = i > 4
		
		// Varying conditionals should work
		if condition {
			process(i)
		}
		
		// Varying conditions in loops
		if reduce.Any(condition) {
			continue // OK: continue based on reduction
		}
		
		// Complex varying conditions
		if reduce.All(i < 2) {
			// All lanes satisfy condition
			process(i * 2)
		}
	}
}

// Test switch statements with varying
func testVaryingSwitchStatements() {
	go for i := range 16 {
		// Switch on varying value
		switch i % 4 {
		case 0:
			process(i)
		case 1:
			process(i * 2)
		default:
			process(i * 3)
		}
		
		// Switch with varying bool
		var condition varying bool = i > 8
		switch condition {
		case true:
			process(i)
		case false:
			process(-i)
		}
	}
}

// Test select statements (should be restricted)
func testSelectRestrictions() {
	ch1 := make(chan int)
	ch2 := make(chan int)
	
	// Regular select should work outside SPMD context
	select {
	case val := <-ch1:
		process(val)
	case ch2 <- 42:
		// sent
	default:
		// default case
	}
	
	// ERROR "select statements not supported in SPMD context"
	go for i := range 4 {
		select {
		case val := <-ch1:
			process(val + int(i))
		default:
			process(int(i))
		}
	}
}

// Test goto restrictions in SPMD context
func testGotoRestrictions() {
	// Regular goto should work outside SPMD
	goto regularLabel
	process(1)
regularLabel:
	process(2)
	
	// ERROR "goto statements not supported in SPMD context"
	go for i := range 4 {
		if i > 2 {
			goto spmdLabel
		}
		process(i)
	spmdLabel:
		process(i * 2)
	}
}

// Test return statements in SPMD functions
func testSPMDReturns(data varying int) varying int {
	// Simple return is OK
	return data * 2
}

func testSPMDConditionalReturns(data varying int) varying int {
	// Conditional returns with varying conditions
	if reduce.Any(data > 10) {
		return data / 2  // OK: uniform control flow based on reduction
	}
	
	// ERROR "conditional return with varying condition not supported"
	if data > 5 {
		return data * 2
	}
	
	return data
}

// Helper function
func process(x int) {
	_ = x
}