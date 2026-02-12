//go:build goexperiment.spmd

// Test SSA generation for SPMD function calls
package spmdtest

import (
	"lanes"
	"reduce"
)

// Test SPMD function calls get mask annotation
func testSPMDFunctionCallSSA() {
	// EXPECT SSA: OpSPMDCallSetMask (mask annotation before SPMD call)
	// EXPECT SSA: OpSPMDFuncEntryMask (in callee, receives implicit mask)
	var data lanes.Varying[int32] = 42

	// Call to SPMD function should emit OpSPMDCallSetMask
	result := spmdMultiply(data, lanes.Varying[int32](2))
	process(result)
}

// SPMD function that receives mask via OpSPMDFuncEntryMask
func spmdMultiply(a lanes.Varying[int32], b lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives implicit mask parameter)
	// EXPECT SSA: OpSPMDMul (varying multiplication under mask)
	return a * b
}

// Test calling SPMD function from within go for loop
func testSPMDCallFromGoForSSA() {
	// EXPECT SSA: OpSPMDCallSetMask (current loop mask passed to callee)
	go for i := range 8 {
		// Mask from go for should be passed via OpSPMDCallSetMask
		result := spmdProcess(i)
		process(result)
	}
}

func spmdProcess(value lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives implicit mask)
	// EXPECT SSA: OpSPMDMul (varying multiplication)
	// EXPECT SSA: OpSPMDAdd (varying addition)
	return value*lanes.Varying[int32](3) + lanes.Varying[int32](1)
}

// Test conditional SPMD function calls
func testConditionalSPMDCallSSA() {
	// EXPECT SSA: OpSPMDCallSetMask (call under varying condition mask)
	// EXPECT SSA: OpSPMDSelect (merge results from both branches)
	go for data := range 100 {
		var result lanes.Varying[int32]
		if data > 50 {
			// Call should be predicated with condition mask via OpSPMDCallSetMask
			result = spmdDouble(data)
		} else {
			result = data
		}

		process(result)
	}
}

func spmdDouble(value lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives mask for predicated execution)
	// EXPECT SSA: OpSPMDMul (varying multiplication)
	return value * lanes.Varying[int32](2)
}

// Test SPMD function with multiple varying parameters
func testMultiParameterSPMDCallSSA() {
	// EXPECT SSA: OpSPMDCallSetMask (mask annotation for multi-param SPMD call)
	var a lanes.Varying[int32] = 10
	var b lanes.Varying[int32] = 20
	var c lanes.Varying[float32] = 3.14

	result := complexSPMDFunc(a, b, c)
	process(result)
}

func complexSPMDFunc(x lanes.Varying[int32], y lanes.Varying[int32], z lanes.Varying[float32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives implicit mask)
	// EXPECT SSA: OpSPMDAdd (x + y)
	var converted lanes.Varying[int32] = lanes.Varying[int32](z)
	return (x + y) * converted
}

// Test SPMD function calling another SPMD function (chained calls)
func testChainedSPMDCallsSSA() {
	// EXPECT SSA: OpSPMDCallSetMask (mask propagation through call chain)
	var data lanes.Varying[int32] = 5

	result := spmdLevel1(data)
	process(result)
}

func spmdLevel1(value lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives mask from caller)
	// EXPECT SSA: OpSPMDCallSetMask (passes mask to spmdLevel2)
	return spmdLevel2(value * 2)
}

func spmdLevel2(value lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives mask from spmdLevel1)
	// EXPECT SSA: OpSPMDCallSetMask (passes mask to spmdLevel3)
	return spmdLevel3(value + 10)
}

func spmdLevel3(value lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives mask)
	// EXPECT SSA: OpSPMDDiv (varying division under mask)
	return value / 3
}

// Test SPMD function with early return
func testSPMDEarlyReturnSSA() {
	// EXPECT SSA: OpSPMDCallSetMask (mask annotation for call)
	var data lanes.Varying[int32] = 25

	result := spmdConditionalReturn(data)
	process(result)
}

func spmdConditionalReturn(value lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives implicit mask)
	// EXPECT SSA: OpSPMDMaskAllTrue (reduce.All check)
	if reduce.All(value > 20) {
		// Early return should respect mask
		return value * 2
	}
	return value + 1
}

// Test non-SPMD function calling SPMD function
func testNonSPMDToSPMDCallSSA() {
	// EXPECT SSA: OpSPMDCallSetMask (non-SPMD caller creates all-true mask)
	var uniformData int32 = 42

	// Non-SPMD function should create initial mask for SPMD call
	var varyingData lanes.Varying[int32] = lanes.Varying[int32](uniformData)
	result := spmdFromNonSPMD(varyingData)
	process(result)
}

func spmdFromNonSPMD(value lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: OpSPMDFuncEntryMask (receives mask from non-SPMD caller)
	// EXPECT SSA: OpSPMDMul (varying multiplication)
	return value * 3
}

// Helper function
func process(x lanes.Varying[int]) {
	_ = x
}
