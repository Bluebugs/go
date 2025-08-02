//go:build goexperiment.spmd

// Test SSA generation for mask propagation through control flow
package spmdtest

import "reduce"

// Test basic conditional mask propagation
func testBasicMaskPropagationSSA() {
	// EXPECT SSA: OpAnd (for mask intersection with condition)
	// EXPECT SSA: OpSelect (for conditional execution)
	var data varying int32 = 42
	var condition varying bool = data > varying int32(20)
	
	var result varying int32
	if condition {
		// EXPECT SSA: operations masked with condition
		result = data * varying int32(2)
	} else {
		// EXPECT SSA: operations masked with negated condition
		result = data + varying int32(10)
	}
	
	process(int(result))
}

// Test nested conditional mask propagation
func testNestedMaskPropagationSSA() {
	// EXPECT SSA: multiple OpAnd for nested mask combinations
	// EXPECT SSA: OpNot (for condition negation)
	var data varying int32 = 30
	var outer varying bool = data > varying int32(10)
	var inner varying bool = data < varying int32(50)
	
	var result varying int32 = data
	if outer {
		// EXPECT SSA: mask = currentMask & outer
		if inner {
			// EXPECT SSA: mask = currentMask & outer & inner
			result = data * varying int32(3)
		} else {
			// EXPECT SSA: mask = currentMask & outer & !inner
			result = data * varying int32(2)
		}
	} else {
		// EXPECT SSA: mask = currentMask & !outer
		result = data + varying int32(5)
	}
	
	process(int(result))
}

// Test switch statement mask propagation
func testSwitchMaskPropagationSSA() {
	// EXPECT SSA: OpEq (for case comparisons)
	// EXPECT SSA: OpOr (for combining case masks)
	// EXPECT SSA: OpSelect (for case execution)
	var data varying int32 = 15
	var selector varying int32 = data % varying int32(3)
	
	var result varying int32
	switch selector {
	case varying int32(0):
		// EXPECT SSA: mask = currentMask & (selector == 0)
		result = data * varying int32(10)
	case varying int32(1):
		// EXPECT SSA: mask = currentMask & (selector == 1)
		result = data * varying int32(20)
	default:
		// EXPECT SSA: mask = currentMask & !(selector == 0 || selector == 1)
		result = data * varying int32(30)
	}
	
	process(int(result))
}

// Test for loop with continue mask propagation
func testForLoopMaskPropagationSSA() {
	// EXPECT SSA: OpPhi (for continue mask tracking)
	// EXPECT SSA: OpOr (for accumulating continue conditions)
	// EXPECT SSA: OpAndNot (for excluding continued lanes)
	for i := 0; i < 10; i++ {
		var condition varying bool = varying int32(i)%varying int32(2) == varying int32(0)
		
		if condition {
			// EXPECT SSA: continue mask updated
			continue
		}
		
		// EXPECT SSA: operations executed with !continue mask
		var data varying int32 = varying int32(i) * varying int32(5)
		process(int(data))
	}
}

// Test mask propagation with reduce operations
func testReduceMaskPropagationSSA() {
	// EXPECT SSA: OpCall (to reduce.All/reduce.Any)
	// EXPECT SSA: uniform result from reduce affects control flow
	var data varying int32 = 25
	var condition varying bool = data > varying int32(20)
	
	// Uniform control flow from reduce
	if reduce.All(condition) {
		// EXPECT SSA: uniform branch, no mask needed
		process(100)
	} else if reduce.Any(condition) {
		// EXPECT SSA: mixed execution, varying operations still masked
		var result varying int32 = data * varying int32(2)
		process(int(result))
	} else {
		// EXPECT SSA: uniform branch, no mask needed
		process(0)
	}
}

// Test complex mask propagation in go for loop
func testGoForMaskPropagationSSA() {
	// EXPECT SSA: OpPhi (for loop mask)
	// EXPECT SSA: OpAnd (for condition intersection)
	// EXPECT SSA: OpSelect (for masked operations)
	go for i := range 16 {
		var condition1 varying bool = i > varying int32(4)
		var condition2 varying bool = i < varying int32(12)
		var combinedCond varying bool = condition1 && condition2
		
		if combinedCond {
			// EXPECT SSA: mask = loopMask & combinedCond
			var data varying int32 = varying int32(i) * varying int32(3)
			process(int(data))
		}
		
		if condition1 {
			// EXPECT SSA: mask = loopMask & condition1
			var other varying int32 = varying int32(i) + varying int32(10)
			process(int(other))
		}
	}
}

// Test mask propagation with function calls
func testFunctionCallMaskPropagationSSA() {
	// EXPECT SSA: OpCall (with current mask passed to SPMD functions)
	var data varying int32 = 35
	var condition varying bool = data > varying int32(30)
	
	if condition {
		// EXPECT SSA: SPMD call with condition mask
		result := maskedSPMDFunction(data)
		process(int(result))
	}
	
	// EXPECT SSA: SPMD call with full mask
	result2 := maskedSPMDFunction(data)
	process(int(result2))
}

func maskedSPMDFunction(value varying int32) varying int32 {
	// EXPECT SSA: function receives mask parameter
	// EXPECT SSA: all operations use received mask
	return value * varying int32(4) + varying int32(7)
}

// Test mask propagation with early exit conditions
func testEarlyExitMaskPropagationSSA() {
	// EXPECT SSA: OpAnd (for break condition mask)
	// EXPECT SSA: OpOr (for accumulated exit mask)
	var counter varying int32 = 0
	
	for i := 0; i < 20; i++ {
		counter = counter + varying int32(1)
		
		// Varying break condition
		var shouldBreak varying bool = counter > varying int32(10)
		if reduce.Any(shouldBreak) {
			// EXPECT SSA: partial lane exit
			break
		}
		
		process(int(counter))
	}
}

// Test mask propagation through select statements
func testSelectMaskPropagationSSA() {
	// Note: select with varying should generate error in type checker
	// This tests mask handling for uniform select in SPMD context
	ch := make(chan int, 1)
	ch <- 42
	
	var data varying int32 = 20
	
	// Uniform select in SPMD context
	select {
	case val := <-ch:
		// EXPECT SSA: uniform operations don't need masking
		uniformResult := val * 2
		
		// EXPECT SSA: varying operations use current mask
		varyingResult := data + varying int32(uniformResult)
		process(int(varyingResult))
	default:
		// EXPECT SSA: varying operations use current mask
		defaultResult := data * varying int32(3)
		process(int(defaultResult))
	}
}

// Helper function
func process(x int) {
	_ = x
}