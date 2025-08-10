//go:build goexperiment.spmd

// Test SSA generation for SPMD function calls
package spmdtest

// Test SPMD function calls get mask-first parameter insertion
func testSPMDFunctionCallSSA() {
	// EXPECT SSA: OpCall (with mask as first parameter)
	// EXPECT SSA: OpPhi (for mask parameter in callee)
	var data varying int32 = 42
	
	// Call to SPMD function should insert mask as first parameter
	result := spmdMultiply(data, varying int32(2))
	process(result)
}

// SPMD function that should receive mask as first parameter
func spmdMultiply(a varying int32, b varying int32) varying int32 {
	// EXPECT SSA: function signature includes mask parameter first
	// EXPECT SSA: OpVectorMul with mask applied via OpSelect
	return a * b
}

// Test calling SPMD function from within go for loop
func testSPMDCallFromGoForSSA() {
	// EXPECT SSA: OpCall (with current loop mask passed)
	// EXPECT SSA: OpAnd (for combining loop mask with call mask)
	go for i := range 8 {
		// Mask from go for should be passed to SPMD function
		result := spmdProcess(i)
		process(result)
	}
}

func spmdProcess(value varying int32) varying int32 {
	// EXPECT SSA: OpSelect (for masked execution)
	return value * varying int32(3) + varying int32(1)
}

// Test conditional SPMD function calls
func testConditionalSPMDCallSSA() {
	// EXPECT SSA: OpAnd (for combining condition mask with call mask)
	// EXPECT SSA: OpSelect (for conditional call execution)
	go for data := range 100 {	
		var result varying int32
		if data > 50 {
			// Call should be predicated with condition mask
			result = spmdDouble(data)
		} else {
			result = data
		}
	
		process(result)
	}
}

func spmdDouble(value varying int32) varying int32 {
	// EXPECT SSA: function receives mask for predicated execution
	return value * varying int32(2)
}

// Test SPMD function with multiple varying parameters
func testMultiParameterSPMDCallSSA() {
	// EXPECT SSA: OpCall (with mask first, then multiple varying params)
	var a varying int32 = 10
	var b varying int32 = 20
	var c varying float32 = 3.14
	
	result := complexSPMDFunc(a, b, c)
	process(result)
}

func complexSPMDFunc(x varying int32, y varying int32, z varying float32) varying int32 {
	// EXPECT SSA: mask parameter first, then x, y, z parameters
	// EXPECT SSA: all operations masked with function mask
	var converted varying int32 = varying int32(z)
	return (x + y) * converted
}

// Test SPMD function calling another SPMD function
func testChainedSPMDCallsSSA() {
	// EXPECT SSA: mask propagation through call chain
	var data varying int32 = 5
	
	result := spmdLevel1(data)
	process(result)
}

func spmdLevel1(value varying int32) varying int32 {
	// EXPECT SSA: OpCall (passing through received mask)
	return spmdLevel2(value * 2)
}

func spmdLevel2(value varying int32) varying int32 {
	// EXPECT SSA: OpCall (passing through received mask) 
	return spmdLevel3(value + 10)
}

func spmdLevel3(value varying int32) varying int32 {
	// EXPECT SSA: all operations use received mask
	return value / 3
}

// Test SPMD function with early return
func testSPMDEarlyReturnSSA() {
	// EXPECT SSA: OpSelect (for conditional return with mask)
	var data varying int32 = 25
	
	result := spmdConditionalReturn(data)
	process(result)
}

func spmdConditionalReturn(value varying int32) varying int32 {
	// EXPECT SSA: OpAnd (for combining mask with condition)
	// EXPECT SSA: OpSelect (for masked return)
	if reduce.All(value > 20) {
		// Early return should respect mask
		return value * 2
	}
	return value + 1
}

// Test non-SPMD function calling SPMD function 
func testNonSPMDToSPMDCallSSA() {
	// EXPECT SSA: OpCall (with default mask for non-SPMD context)
	var uniformData uniform int32 = 42
	
	// Non-SPMD function should create initial mask for SPMD call
	var varyingData varying int32 = varying int32(uniformData)
	result := spmdFromNonSPMD(varyingData)
	process(result)
}

func spmdFromNonSPMD(value varying int32) varying int32 {
	// EXPECT SSA: receives mask from non-SPMD caller (all lanes active)
	return value * 3
}

// Helper function
func process(x varying int) {
	_ = x
}