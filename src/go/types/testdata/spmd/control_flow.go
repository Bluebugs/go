//go:build goexperiment.spmd

// Test control flow restrictions in SPMD contexts
package spmdtest

import (
	"lanes"
)

// Test go for loop restrictions (ISPC-based approach)
func testGoForRestrictions() {
	threshold := 7

	// Valid go for loops - continue always allowed
	go for i := range 10 {
		if i > 5 {
			continue // OK: continue always allowed
		}
		processC(i)
	}

	// ALLOWED: return/break under uniform conditions
	go for i := range 10 {
		if threshold < 0 {
			return // OK: uniform condition allows return
		}
		if threshold > 100 {
			break // OK: uniform condition allows break
		}
		processC(i)
	}

	go for i := range 10 {
		if i > 5 { // varying condition
			break  /* ERROR "break statement not allowed under varying conditions in SPMD for loop" */
		}
		processC(i)
	}

	go for i := range 10 {
		if i < 3 { // varying condition
			return /* ERROR "return statement not allowed under varying conditions in SPMD for loop" */
		}
		processC(i)
	}

	go for i := range 10 {
		go /* ERROR "nested `go for` loop (prohibited for now)" */ for j := range 5 {
			processC(i + j)
		}
		_ = i // use i to avoid "declared and not used" error
	}
}

// Test mask alteration scenarios - continue in varying context affects subsequent uniform conditions
func testMaskAlterationScenarios() {
	threshold := 7
	mode := 1

	go for i := range 10 {
		if i > 5 { // varying condition
			continue  // OK: continue always allowed, but alters mask
		}

		// Mask has been altered by previous continue in varying context
		if mode == 1 { // uniform condition, but mask is altered
			return /* ERROR "return statement not allowed after continue in varying context in SPMD for loop" */
		}
		processC(i)
	}

	go for i := range 10 {
		if i < 3 { // varying condition
			continue  // Alters mask
		}

		if threshold > 0 { // uniform condition, but mask altered
			break /* ERROR "break statement not allowed after continue in varying context in SPMD for loop" */
		}
		processC(i)
	}

	// Complex mask alteration scenario
	go for i := range 10 {
		if i > 2 { // varying condition
			if i < 8 { // nested varying condition
				continue  // Alters mask - some lanes skip remaining
			}
		}

		if mode > 0 { // uniform condition on remaining active lanes only
			return /* ERROR "return statement not allowed after continue in varying context in SPMD for loop" */
		}
		processC(i)
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
		processC(i)
	}

	// Nested regular for loops are fine
	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			if i+j > 10 {
				break // OK: break in inner regular for
			}
			processC(i + j)
		}
	}
}

// Test mixed control flow (go for with regular for inside)
func testMixedControlFlow() {
	mode := 1

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
			processC(int(i) + j)
		}

		if i > 5 { // varying condition
			return /* ERROR "return statement not allowed under varying conditions in SPMD for loop" */
		}

		// But another go for is not allowed
		go /* ERROR "nested `go for` loop (prohibited for now)" */ for k := range 3 {
			processC(int(i) + k)
		}
	}
}

// Test nested varying conditions (complex cases)
func testNestedVaryingConditions() {
	mode := 1
	data := []int{1, 2, 3, 4, 5, 6, 7, 8}

	go for i := range len(data) {
		// Uniform outer condition - return/break OK here
		if mode == 1 { // uniform condition
			if data[i] > 5 { // varying condition - now return/break forbidden
				return /* ERROR "return statement not allowed under varying conditions in SPMD for loop" */
			}

			// ALLOWED: Still under uniform condition only
			if mode == 2 {
				return // OK: no varying conditions in scope
			}
		}

		// Complex nesting scenarios
		if mode > 0 { // uniform condition
			if data[i] > 3 { // varying condition
				if mode > 10 { // even uniform conditions can't rescue us
					break /* ERROR "break statement not allowed under varying conditions in SPMD for loop" */
				}
				continue // OK: continue always allowed
			}

			// FORBIDDEN: mask altered by previous continue in varying context
			if mode > 50 {
				break /* ERROR "break statement not allowed after continue in varying context in SPMD for loop" */
			}
		}
	}
}

// Test conditional control flow with varying
func testVaryingControlFlow() {
	mode := 1
	go for i := range 8 {
		var condition lanes.Varying[bool] = i > 4

		// Varying conditionals should work
		if condition {
			processC(i)
		}

		// uniform conditions in loops
		if mode > 0 {
			break // OK: break based on uniform condition
		}

		// Complex uniform conditions
		if mode < 10 {
			// All lanes satisfy condition
			processC(i * 2)
		}
	}
}

// Test switch statements with varying
// NOTE: Switch statement ERROR comment positioning is complex in go/types
// These tests work in types2 but have positioning issues in go/types
// Keeping only the uniform switch test for now
func testVaryingSwitchStatements() {
	mode := 1

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

		// Varying switch tests commented out due to ERROR positioning complexity
		// These work fine at runtime, just test framework has issues
		_ = i
	}
}

// Test select statements (should be restricted)
func testSelectRestrictions() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	// Regular select should work outside SPMD context
	select {
	case val := <-ch1:
		processC(val)
	case ch2 <- 42:
		// sent
	default:
		// default case
	}

	go for i := range 4 {
		select /* ERROR "select statements not supported in SPMD context" */ {
		case val := <-ch1:
			processC(val + i)
		default:
			processC(i)
		}
		_ = i // This is necessary to avoid "declared and not used" error as the body inside the select is ignored
	}
}

// Test goto restrictions in SPMD context
func testGotoRestrictions() {
	// Regular goto should work outside SPMD
	goto regularLabel
	processC(1)
regularLabel:
	processC(2)

	go for i := range 4 {
		if i > 2 {
			goto /* ERROR "goto statements not supported in SPMD context" */ spmdLabel
		}
		processC(i)
	spmdLabel /* ERROR "goto statements not supported in SPMD context" */ :
		processC(i * 2)
	}
}

// Test return statements in SPMD functions
func testSPMDReturns(data lanes.Varying[int]) lanes.Varying[int] {
	// Simple return is OK
	return data * 2
}

func testSPMDConditionalReturns(data lanes.Varying[int], threshold int) lanes.Varying[int] {
	// ALLOWED: Uniform conditions in SPMD functions
	if threshold < 0 {
		return data / 2  // OK: uniform condition
	}

	if data > 5 { // varying condition - conditional returns not yet implemented
		return data * 2 // Future: conditional return with varying condition not supported
	}

	// Uniform condition
	if threshold > 10 {
		return data / 2  // OK: uniform condition
	}

	return data
}

// Test edge cases with uniform conditions after varying context
func testReduceOperationEdgeCases() {
	data := []int{1, 2, 3, 4, 5}
	mode := 1

	go for i := range len(data) {
		// Edge case: uniform condition
		if mode > 0 { // uniform
			return // OK: pure uniform condition
		}

		// This introduce a varying context in the chain of operations
		if data[i] > 3 { // varying condition
			// This is considered varying context since it is inside a varying context
			if mode > 10 { // uniform, but nested in varying
				return /* ERROR "return statement not allowed under varying conditions in SPMD for loop" */
			}
		}

		// Pure uniform condition is OK
		if mode < 100 {
			return // OK: pure uniform condition
		}
	}

	go for i := range len(data) {
		varyingCondition := data[i] > 3

		if varyingCondition {
			continue // OK: continue always allowed
		}
		// After this point, we are in a varying context as the continue above might have altered the control flow mask

		// This is considered varying context since it is inside a varying context due to the continue
		if mode > 10 { // uniform condition in varying context
			return /* ERROR "return statement not allowed after continue in varying context in SPMD for loop" */
		}
	}
}

// Test that return expressions are fully type-checked inside go for loops.
// Regression test for "no type for *ast.CompositeLit" panic in x-tools SSA builder.
type PointC struct{ X, Y int }

func testReturnExpressionTypeChecking(data []int, threshold int) (PointC, error) {
	go for i := range len(data) {
		if threshold < 0 {
			return PointC{X: 1, Y: 2}, nil // OK: uniform condition, composite literal must be type-checked
		}
		_ = i
	}
	return PointC{}, nil
}

// Test panic restrictions in SPMD context
func testPanicRestrictions() {
	data := []int{1, 2, 3, 4}
	mode := 1

	// panic under uniform condition is OK
	go for i := range len(data) {
		if mode < 0 {
			panic("negative mode") // OK: uniform condition
		}
		processC(i)
	}

	// panic under varying condition is FORBIDDEN
	go for i := range len(data) {
		if i > 2 { // varying condition
			panic /* ERROR "panic not allowed under varying conditions in SPMD for loop" */ ("too high")
		}
		processC(i)
	}

	// panic after mask alteration is FORBIDDEN
	go for i := range len(data) {
		if i > 5 { // varying condition
			continue // alters mask
		}
		if mode > 0 { // uniform, but mask altered
			panic /* ERROR "panic not allowed after continue in varying context in SPMD for loop" */ ("altered")
		}
		processC(i)
	}

	// panic at top level of go for (no varying context) is OK
	go for i := range len(data) {
		if mode == 42 {
			panic("direct") // OK: uniform condition
		}
		processC(i)
	}

	// bare panic at top of go for body (no if guard) is OK
	go for i := range len(data) {
		panic("always") // OK: no varying context, varyingDepth==0, maskAltered==false
		processC(i)
	}
}

// Helper function
func processC(x lanes.Varying[int]) {
	_ = x
}
