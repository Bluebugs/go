//go:build goexperiment.spmd

// Test SSA generation for reduce operations
package spmdtest

import "reduce"

// Test basic reduce operations generate builtin calls
func testBasicReduceSSA() {
	// EXPECT SSA: OpCall (to reduce.Add)
	// EXPECT SSA: OpCall (to reduce.All)
	// EXPECT SSA: OpCall (to reduce.Any)
	var data varying int32 = varying int32(lanes.Index()) + varying int32(10)
	var condition varying bool = data > varying int32(15)
	
	var sum uniform int32 = reduce.Add(data)
	var allTrue uniform bool = reduce.All(condition)
	var anyTrue uniform bool = reduce.Any(condition)
	
	if allTrue || anyTrue {
		process(int(sum))
	}
}

// Test reduce operations with different numeric types
func testNumericReduceSSA() {
	// EXPECT SSA: OpCall (to reduce.Add for different types)
	// EXPECT SSA: OpCall (to reduce.Max/Min for different types)
	var intData varying int32 = 25
	var floatData varying float32 = 3.14
	var doubleData varying float64 = 2.718
	
	var intSum uniform int32 = reduce.Add(intData)
	var floatSum uniform float32 = reduce.Add(floatData)
	var doubleSum uniform float64 = reduce.Add(doubleData)
	
	var intMax uniform int32 = reduce.Max(intData)
	var floatMax uniform float32 = reduce.Max(floatData)
	var doubleMax uniform float64 = reduce.Max(doubleData)
	
	var intMin uniform int32 = reduce.Min(intData)
	var floatMin uniform float32 = reduce.Min(floatData)
	var doubleMin uniform float64 = reduce.Min(doubleData)
	
	process(int(intSum + uniform int32(floatSum) + uniform int32(doubleSum)))
	process(int(intMax + uniform int32(floatMax) + uniform int32(doubleMax)))
	process(int(intMin + uniform int32(floatMin) + uniform int32(doubleMin)))
}

// Test bitwise reduce operations
func testBitwiseReduceSSA() {
	// EXPECT SSA: OpCall (to reduce.Or/And/Xor)
	var data varying int32 = varying int32(0xFF) << varying int32(lanes.Index())
	
	var orResult uniform int32 = reduce.Or(data)
	var andResult uniform int32 = reduce.And(data)
	var xorResult uniform int32 = reduce.Xor(data)
	
	process(int(orResult + andResult + xorResult))
}

// Test reduce.From for varying-to-array conversion
func testReduceFromSSA() {
	// EXPECT SSA: OpCall (to reduce.From)
	// EXPECT SSA: array construction from varying values
	var data varying int32 = varying int32(lanes.Index()) * varying int32(5)
	
	// Convert varying to array
	var array []int32 = reduce.From(data)
	
	// Use array elements
	for i := 0; i < len(array); i++ {
		process(int(array[i]))
	}
}

// Test reduce operations in conditional contexts
func testConditionalReduceSSA() {
	// EXPECT SSA: OpCall (to reduce within conditions)
	// EXPECT SSA: uniform control flow from reduce results
	var data varying int32 = varying int32(lanes.Index()) + varying int32(20)
	var condition varying bool = data > varying int32(22)
	
	// Uniform control flow from reduce
	if reduce.All(condition) {
		// EXPECT SSA: uniform branch, no masking needed
		var sum uniform int32 = reduce.Add(data)
		process(int(sum))
	} else if reduce.Any(condition) {
		// EXPECT SSA: mixed execution, some lanes active
		var max uniform int32 = reduce.Max(data)
		process(int(max))
	} else {
		// EXPECT SSA: uniform branch, all lanes false
		process(0)
	}
}

// Test reduce operations in go for loop
func testReduceInGoForSSA() {
	// EXPECT SSA: OpCall (to reduce within SPMD loop)
	go for i := range 8 {
		var loopData varying int32 = varying int32(i) * varying int32(3)
		var loopCondition varying bool = loopData > varying int32(10)
		
		// Early loop termination based on reduce
		if reduce.All(loopCondition) {
			// EXPECT SSA: uniform control flow affects loop
			var sum uniform int32 = reduce.Add(loopData)
			process(int(sum))
			continue
		}
		
		// Lane-specific processing
		if reduce.Any(loopCondition) {
			var filtered varying int32 = loopData
			if loopCondition {
				filtered = filtered * varying int32(2)
			}
			process(int(filtered))
		}
	}
}

// Test constrained varying with reduce
func testConstrainedVaryingReduceSSA() {
	// EXPECT SSA: OpCall (to reduce with constrained varying)
	// EXPECT SSA: array decomposition for constrained types
	var constrainedData varying[4] int32
	
	// Initialize constrained data
	for i := 0; i < 4; i++ {
		constrainedData[i] = int32(i * 10)
	}
	
	// Reduce constrained varying
	var sum uniform int32 = reduce.Add(constrainedData)
	var allPositive uniform bool = reduce.All(constrainedData > varying int32(0))
	
	if allPositive {
		process(int(sum))
	}
}

// Test reduce with function calls
func testReduceWithFunctionCallsSSA() {
	// EXPECT SSA: OpCall (to SPMD function and then reduce)
	var data varying int32 = 15
	
	// Call SPMD function and reduce result
	var processed varying int32 = processData(data)
	var result uniform int32 = reduce.Add(processed)
	
	process(int(result))
}

func processData(input varying int32) varying int32 {
	// EXPECT SSA: function generates varying result for reduction
	return input * varying int32(2) + varying int32(lanes.Index())
}

// Test nested reduce operations
func testNestedReduceSSA() {
	// EXPECT SSA: multiple OpCall (to different reduce functions)
	var data varying int32 = varying int32(lanes.Index()) + varying int32(5)
	
	// Multiple reduce operations
	var sum uniform int32 = reduce.Add(data)
	var max uniform int32 = reduce.Max(data)
	var hasPositive uniform bool = reduce.Any(data > varying int32(0))
	
	// Use reduce results in computation
	var finalResult uniform int32
	if hasPositive {
		finalResult = sum + max
	} else {
		finalResult = 0
	}
	
	process(int(finalResult))
}

// Test reduce with complex expressions
func testComplexReduceExpressionSSA() {
	// EXPECT SSA: OpCall (to reduce with complex varying expression)
	var a varying int32 = 10
	var b varying int32 = 20
	var c varying float32 = 2.5
	
	// Complex expression in reduce
	var complexResult uniform int32 = reduce.Add((a + b) * varying int32(c))
	var complexCondition uniform bool = reduce.All((a > b) || (varying int32(c) > a))
	
	if complexCondition {
		process(int(complexResult))
	}
}

// Test reduce operations with memory access
func testReduceMemoryAccessSSA() {
	// EXPECT SSA: OpCall (to reduce with varying memory operations)
	// EXPECT SSA: OpVectorLoad (for varying array access)
	var data [16]int32
	for i := 0; i < 16; i++ {
		data[i] = int32(i * i)
	}
	
	go for i := range 4 {
		// Load varying data and reduce
		var loadedData varying int32 = varying int32(data[i*2])
		var sum uniform int32 = reduce.Add(loadedData)
		process(int(sum))
	}
}

// Helper function
func process(x int) {
	_ = x
}