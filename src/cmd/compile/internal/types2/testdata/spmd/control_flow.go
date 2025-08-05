//go:build goexperiment.spmd

// Test control flow restrictions in SPMD contexts
package spmdtest

import "reduce"

// Test go for loop restrictions (ISPC-based approach)
func testGoForRestrictions() {
	threshold := 7
	
	// Valid go for loops - continue always allowed
	go for i := range 10 {
		if i > 5 {
			continue // OK: continue always allowed
		}
		process(i)
	}
	
	// ALLOWED: return/break under uniform conditions
	go for i := range 10 {
		if threshold < 0 {
			return // OK: uniform condition allows return
		}
		if threshold > 100 {
			break // OK: uniform condition allows break
		}
		process(i)
	}
	
	// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"
	go for i := range 10 {
		if i > 5 { // varying condition
			break  // ERROR: varying condition forbids break
		}
		process(i)
	}
	
	// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"  
	go for i := range 10 {
		if i < 3 { // varying condition
			return // ERROR: varying condition forbids return
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

// Test mask alteration scenarios - continue in varying context affects subsequent uniform conditions
func testMaskAlterationScenarios() {
	threshold := 7
	mode := uniform int(1)
	
	// ERROR "return statement not allowed after continue in varying context in SPMD for loop"
	go for i := range 10 {
		if i > 5 { // varying condition
			continue  // OK: continue always allowed, but alters mask
		}
		
		// Mask has been altered by previous continue in varying context
		if mode == 1 { // uniform condition, but mask is altered
			return // ERROR: return forbidden due to mask alteration
		}
		process(i)
	}
	
	// ERROR "break statement not allowed after continue in varying context in SPMD for loop"
	go for i := range 10 {
		if i < 3 { // varying condition
			continue  // Alters mask
		}
		
		if threshold > 0 { // uniform condition, but mask altered
			break // ERROR: break forbidden due to mask alteration
		}
		process(i)
	}
	
	// Complex mask alteration scenario
	go for i := range 10 {
		if i > 2 { // varying condition
			if i < 8 { // nested varying condition
				continue  // Alters mask - some lanes skip remaining
			}
		}
		
		// ERROR "return statement not allowed after continue in varying context in SPMD for loop"
		if mode > 0 { // uniform condition on remaining active lanes only
			return // ERROR: uniform condition but altered mask
		}
		process(i)
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
	mode := uniform int(1)
	
	go for i := range 10 {
		// ALLOWED: Uniform return/break at go for level
		if mode < 0 {
			return // OK: uniform condition
		}
		
		// Regular for loop inside go for is allowed
		for j := 0; j < 5; j++ {
			if j > 2 {
				break // OK: break in regular for inside go for
			}
			process(int(i) + j)
		}
		
		// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"
		if i > 5 { // varying condition
			return // ERROR: varying condition forbids return
		}
		
		// But another go for is not allowed
		// ERROR "nested go for loops not allowed"
		go for k := range 3 {
			process(int(i) + k)
		}
	}
}

// Test nested varying conditions (complex cases)
func testNestedVaryingConditions() {
	mode := uniform int(1)
	data := []int{1, 2, 3, 4, 5, 6, 7, 8}
	
	go for i := range len(data) {
		// Uniform outer condition - return/break OK here
		if mode == 1 {
			// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"
			if data[i] > 5 { // varying condition - now return/break forbidden
				return // ERROR: enclosing varying condition
			}
			
			// ALLOWED: Still under uniform condition only  
			if mode == 2 {
				return // OK: no varying conditions in scope
			}
		}
		
		// Complex nesting scenarios
		if mode > 0 { // uniform condition
			if data[i] > 3 { // varying condition - now in varying context
				// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"
				if mode > 10 { // even uniform conditions can't rescue us
					break // ERROR: enclosing varying condition
				}
				continue // OK: continue always allowed
			}
			
			// ALLOWED: Back to uniform-only context
			if mode > 50 {
				break // OK: only uniform conditions in scope
			}
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
		
		// uniform conditions in loops
		if reduce.Any(condition) {
			break // OK: break based on reduction to uniform result
		}
		
		// Complex uniform conditions
		if reduce.All(i < 2) {
			// All lanes satisfy condition
			process(i * 2)
		}
	}
}

// Test switch statements with varying
func testVaryingSwitchStatements() {
	mode := uniform int(1)
	
	go for i := range 16 {
		// ALLOWED: Switch on uniform value
		switch mode {
		case 1:
			return // OK: uniform switch allows return
		case 2: 
			break // OK: uniform switch allows break
		default:
			continue // Always OK
		}
		
		// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"
		switch i % 4 { // varying condition
		case 0:
			return // ERROR: varying switch forbids return
		case 1:
			break // ERROR: varying switch forbids break
		default:
			continue // OK: continue always allowed
		}
		
		// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"
		var condition varying bool = i > 8
		switch condition { // varying condition
		case true:
			return // ERROR: varying switch forbids return
		case false:
			break // ERROR: varying switch forbids break
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

func testSPMDConditionalReturns(data varying int, threshold uniform int) varying int {
	// ALLOWED: Uniform conditions in SPMD functions
	if threshold < 0 {
		return data / 2  // OK: uniform condition
	}
	
	// ERROR "conditional return with varying condition not supported"
	if data > 5 { // varying condition
		return data * 2 // ERROR: varying condition in SPMD function
	}
	
	// Reduction produces uniform result but input is varying - still varying context
	if reduce.Any(data > 10) {
		return data / 2  // OK: uniform result from reduce operation
	}
	
	return data
}

// Test edge cases with reduce operations
func testReduceOperationEdgeCases() {
	data := []int{1, 2, 3, 4, 5}
	
	go for i := range len(data) {
		// Edge case: reduce produces uniform result but from varying input
		varyingCondition := data[i] > 3
		
		// This is considered uniform context since their is no alteration of the control flow mask prior to this point
		// aka no continue of the sort nor inside a if
		if reduce.Any(varyingCondition) { // uniform result with no varying context
			return // OK: pure uniform condition
		}

		// This introduce a varying context in the chain of operations
		if varyingCondition {
			// This is considered varying context since it is inside a varying context
			if reduce.Any(varyingCondition) { // uniform result in varying context
				// ERROR "break/return statement not allowed under varying conditions in SPMD for loop"
				return // ERROR: still considered varying context
			}
		}		
		
		// Pure uniform condition with reduce is OK
		uniformCondition := true
		if reduce.All(uniformCondition) {
			return // OK: pure uniform condition
		}
	}

	go for i := range len(data) {
		// Edge case: reduce produces uniform result but from varying input
		varyingCondition := data[i] > 3

		if varyingCondition {
			continue // OK: continue always allowed
		}
		// After this point, we are in a varying context as the continue above might have altered the control flow mask

		// This is considered varying context since it is inside a varying context due to the continue
		if reduce.Any(varyingCondition) { // uniform result in a varying context
			return // ERROR: still considered varying context
		}
	}
}

// Helper function
func process(x int) {
	_ = x
}