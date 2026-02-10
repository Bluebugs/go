//go:build goexperiment.spmd

// Comprehensive test cases for ISPC-based return/break control flow rules
// Tests the refined approach where return/break are allowed under uniform conditions only
package spmdtest

import (
	"lanes"
	"reduce"
)

// Test 1: Basic uniform return/break scenarios - should be ALLOWED
func testUniformReturnBreak() {
	threshold := 10
	mode := 1

	go for i := range 16 {
		// ALLOWED: Direct uniform conditions
		if threshold < 0 {
			return // OK: uniform condition allows return
		}

		if threshold > 100 {
			break // OK: uniform condition allows break
		}

		// ALLOWED: Uniform function calls
		if isShutdownRequested() { // uniform function
			return // OK: uniform condition
		}

		// ALLOWED: Uniform switch statements
		switch mode {
		case 1:
			return // OK: uniform switch allows return
		case 2:
			break // OK: uniform switch allows break
		default:
			continue // Always OK
		}
		_ = i // Use i to avoid "declared and not used" error
	}
}

// Test 2: Basic varying return/break scenarios - should generate ERRORS
func testVaryingReturnBreak() {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8}

	go for i := range len(data) {
		if data[i] < 0 { // varying condition
			return // ERROR "return statement not allowed under varying conditions in SPMD for loop"
		}

		if data[i] > 100 { // varying condition
			break // ERROR "break statement not allowed under varying conditions in SPMD for loop"
		}

		// ALLOWED: Continue under varying condition
		if data[i] == 0 {
			continue // OK: continue always allowed
		}

		switch data[i] { // varying switch - mask already altered by previous continue
		case 1:
			return // ERROR "return statement not allowed after continue in varying context in SPMD for loop"
		case 2:
			break // ERROR "break statement not allowed after continue in varying context in SPMD for loop"
		default:
			continue // OK: continue always allowed
		}
	}
}

// Test 3: Complex nested conditions - varying depth tracking
func testNestedConditions() {
	mode := 1
	data := []int{1, 2, 3, 4, 5}

	go for i := range len(data) {
		// Scenario A: Uniform outer, varying inner - return/break FORBIDDEN in inner
		if mode == 1 { // uniform condition - depth 0
			if data[i] > 3 { // varying condition - depth 1
				return // ERROR "return statement not allowed under varying conditions in SPMD for loop"
			}

			// ALLOWED: Still under uniform condition only
			if mode == 2 { // uniform condition - depth still 0
				return // OK: varying depth == 0
			}
		}

		// Scenario B: Multiple nested uniform conditions - ALLOWED
		if mode > 0 { // uniform condition - depth 0
			if mode < 10 { // uniform condition - depth still 0
				if mode != 5 { // uniform condition - depth still 0
					return // OK: all conditions uniform
				}
			}
		}

		// Scenario C: Varying outer, uniform inner - return/break FORBIDDEN everywhere
		if data[i] > 2 { // varying condition - depth 1
			if mode > 0 { // uniform condition but varying depth > 0
				break // ERROR "break statement not allowed under varying conditions in SPMD for loop"
			}

			continue // OK: continue always allowed
		}
	}
}

// Test 4: Edge cases with reduce operations
func testReduceOperationEdgeCases() {
	data := []int{1, 2, 3, 4, 5}
	uniformCondition := true

	go for i := range len(data) {
		// Edge case: reduce produces uniform result but from varying input
		varyingCondition := data[i] > 3

		if reduce.Any(varyingCondition) { // uniform result from varying input
			return // ALLOWED: uniform context as reduce.Any produces uniform result and no mask alteration earlier
		}

		// Pure uniform condition with reduce - should be OK
		if reduce.All(uniformCondition) {
			return // OK: pure uniform condition
		}

		// Mixed scenario: uniform condition tested against varying reduce result
		uniformThreshold := 2
		totalAboveThreshold := reduce.Count(data[i] > uniformThreshold)

		if totalAboveThreshold > 3 { // uniform condition on uniform result
			return // OK: uniform condition
		}
	}
}

// Test 5: Switch statement variations
func testSwitchStatementVariations() {
	mode := 1
	data := []int{1, 2, 3, 4}

	go for i := range len(data) {
		// ALLOWED: Switch on uniform expression
		switch mode + 1 { // uniform expression
		case 1:
			return // OK: uniform switch
		case 2:
			break // OK: uniform switch
		}

		// Mixed: uniform switch with varying case bodies
		switch mode { // uniform switch - OK so far
		case 1:
			if data[i] > 2 { // varying condition in case body
				return // ERROR "return statement not allowed under varying conditions in SPMD for loop"
			}
			break // OK: still under uniform switch only
		case 2:
			return // OK: uniform switch allows return
		}

		switch data[i] % 3 { // varying expression
		case 0:
			return // ERROR "return statement not allowed under varying conditions in SPMD for loop"
		case 1:
			break // ERROR "break statement not allowed under varying conditions in SPMD for loop"
		default:
			continue // OK: continue always allowed
		}
	}
}

// Test 6: Function call variations affecting conditions
func testFunctionCallConditions() {
	data := []int{1, 2, 3, 4}

	go for i := range len(data) {
		// ALLOWED: Uniform function call
		if isShutdownRequested() { // uniform function
			return // OK: uniform condition
		}

		if isNegative(data[i]) { // function taking varying input and returning varying result
			return // ERROR "return statement not allowed under varying conditions in SPMD for loop"
		}

		// ALLOWED: Uniform result from uniform function
		if getMode() > 5 { // uniform function returning uniform
			break // OK: uniform condition
		}
	}
}

// Test 7: Complex masking scenarios that should remain as continue-only
func testComplexMaskingScenarios() {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8}

	go for i := range len(data) {
		// Complex varying conditions that require masking
		if data[i] > 3 && data[i] < 7 { // varying condition
			if data[i]%2 == 0 { // nested varying condition
				// These require complex per-lane mask tracking
				continue // OK: continue handles masking automatically
			}

			if data[i] == 5 { // mask altered by previous continue
				return // ERROR "return statement not allowed after continue in varying context in SPMD for loop"
			}
		}

		// Multi-level varying nesting
		condition1 := data[i] > 2
		condition2 := data[i] < 6

		if condition1 { // varying
			if condition2 { // varying nested in varying - mask already altered by previous continue
				return // ERROR "return statement not allowed after continue in varying context in SPMD for loop"
			}
		}
	}
}

// Test 8: Mask alteration scenarios - continue in varying context affects subsequent uniform conditions
func testMaskAlterationScenarios() {
	mode := 1
	data := []int{1, 2, 3, 4, 5, 6, 7, 8}

	// FORBIDDEN: Return after continue in varying context
	go for i := range len(data) {
		if data[i] < 0 { // varying condition
			continue  // OK: continue always allowed, but alters mask
		}

		if mode == 1 { // uniform condition, but mask is altered
			return // ERROR "return statement not allowed after continue in varying context in SPMD for loop"
		}

		process(data[i])
	}

	// FORBIDDEN: Break after continue in varying context
	go for i := range len(data) {
		if data[i] > 10 { // varying condition
			continue  // Alters mask
		}

		if mode > 0 { // uniform condition, but mask altered
			break // ERROR "break statement not allowed after continue in varying context in SPMD for loop"
		}

		process(data[i])
	}

	// FORBIDDEN: Complex mask alteration with nested conditions
	go for i := range len(data) {
		if data[i] > 3 { // varying condition - depth 1
			if data[i] < 7 { // nested varying condition - depth 2
				continue  // Alters mask - some lanes skip remaining
			}
		}

		if mode > 5 { // uniform condition on remaining active lanes only
			return // ERROR "return statement not allowed after continue in varying context in SPMD for loop"
		}

		process(data[i])
	}

	// ALLOWED: No continue before uniform condition
	go for i := range len(data) {
		if data[i] > 100 { // varying condition, but no continue
			// Just some varying processing, no continue
			process(data[i] * 2)
		}

		// This is fine as the varying condition above did not alter the mask by using any continue
		if mode == 2 { // uniform condition leading to uniform context
			return // OK: clean uniform context, no prior mask alteration
		}

		process(data[i])
	}

	// ALLOWED: Continue in uniform context doesn't alter mask
	go for i := range len(data) {
		if mode < 0 { // uniform condition
			continue  // OK: continue in uniform context doesn't alter mask
		}

		// ALLOWED: No mask alteration occurred
		if mode > 10 { // uniform condition, no mask alteration
			return // OK: clean uniform context
		}

		process(data[i])
	}
}

// Test 9: Edge case - labeled breaks (if supported)
func testLabeledBreaks() {
	threshold := 5
	data := []int{1, 2, 3, 4}

	// ALLOWED: Labeled break under uniform condition
outerUniform:
	go for i := range len(data) {
		if threshold < 0 { // uniform condition
			break outerUniform // OK: uniform condition allows labeled break
		}

	outerVarying:
		for j := 0; j < 3; j++ { // regular for loop
			if data[i] > j { // varying condition
				break outerVarying // ERROR "break statement not allowed under varying conditions in SPMD for loop"
			}
		}
	}
}

// Helper functions for testing
func isShutdownRequested() bool {
	return false
}

func isNegative(x lanes.Varying[int]) lanes.Varying[bool] {
	return x < 0
}

func getMode() int {
	return 1
}

func process(x lanes.Varying[int]) {
	_ = x // Simulate processing
}
