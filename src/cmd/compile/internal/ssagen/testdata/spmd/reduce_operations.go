//go:build goexperiment.spmd

// Test SSA generation for reduce operations
package spmdtest

import (
	"lanes"
	"reduce"
)

// Test basic reduce operations generate builtin calls
func testBasicReduceSSA() {
	// EXPECT SSA: OpCall (to reduce.Add)
	// EXPECT SSA: OpCall (to reduce.All)
	// EXPECT SSA: OpCall (to reduce.Any)
	go for data := range 30 {
		data = data + 10

		sum := reduce.Add(data)

		condition := data > 15

		allTrue := reduce.All(condition)
		anyTrue := reduce.Any(condition)

		if allTrue || anyTrue {
			process(sum)
		}
	}
}

// Test reduce operations with different numeric types
func testNumericReduceSSA() {
	// EXPECT SSA: OpCall (to reduce.Add for different types)
	// EXPECT SSA: OpCall (to reduce.Max/Min for different types)
	var intData lanes.Varying[int32] = 25
	var floatData lanes.Varying[float32] = 3.14
	var doubleData lanes.Varying[float64] = 2.718

	var intSum int32 = reduce.Add(intData)
	var floatSum float32 = reduce.Add(floatData)
	var doubleSum float64 = reduce.Add(doubleData)

	var intMax int32 = reduce.Max(intData)
	var floatMax float32 = reduce.Max(floatData)
	var doubleMax float64 = reduce.Max(doubleData)

	var intMin int32 = reduce.Min(intData)
	var floatMin float32 = reduce.Min(floatData)
	var doubleMin float64 = reduce.Min(doubleData)

	process(intSum + lanes.Varying[int32](floatSum) + lanes.Varying[int32](doubleSum))
	process(intMax + lanes.Varying[int32](floatMax) + lanes.Varying[int32](doubleMax))
	process(intMin + lanes.Varying[int32](floatMin) + lanes.Varying[int32](doubleMin))
}

// Test bitwise reduce operations
func testBitwiseReduceSSA() {
	// EXPECT SSA: OpCall (to reduce.Or/And/Xor)
	go for i := range 32 {
		data := i << lanes.Index()

		orResult := reduce.Or(data)
		andResult := reduce.And(data)
		xorResult := reduce.Xor(data)

		process(orResult + andResult + xorResult)
	}
}

// Test reduce.From for varying-to-array conversion
func testReduceFromSSA() {
	// EXPECT SSA: OpCall (to reduce.From)
	// EXPECT SSA: array construction from varying values
	var data lanes.Varying[int32] = 5

	// Convert varying to array
	var array []int32 = reduce.From(data)

	// Use array elements
	for i := 0; i < len(array); i++ {
		_ = array[i]
	}
}

// Test reduce operations in conditional contexts
func testConditionalReduceSSA() {
	// EXPECT SSA: OpCall (to reduce within conditions)
	// EXPECT SSA: uniform control flow from reduce results
	go for i := range 10 {
		data := i + 20

		condition := data > 22

		// Uniform control flow from reduce
		if reduce.All(condition) {
			// EXPECT SSA: uniform branch, no masking needed
			_ = reduce.Add(data)
		} else if reduce.Any(condition) {
			// EXPECT SSA: mixed execution, some lanes active
			_ = reduce.Max(data)
		} else if condition {
			// EXPECT SSA: varying branch, never called, all lanes false
			process(data)
		} else {
			// EXPECT SSA: varying branch, fallback when condition false
			process(data)
		}
	}
}

// Test reduce operations in go for loop
func testReduceInGoForSSA() {
	// EXPECT SSA: OpCall (to reduce within SPMD loop)
	go for i := range 8 {
		var loopData lanes.Varying[int32] = i * 3
		var loopCondition lanes.Varying[bool] = loopData > 10

		// Early loop termination based on reduce
		if reduce.All(loopCondition) {
			// EXPECT SSA: uniform control flow affects loop
			_ = reduce.Add(loopData)
			continue
		}

		// Lane-specific processing
		if reduce.Any(loopCondition) {
			var filtered lanes.Varying[int32] = loopData
			if loopCondition {
				filtered = filtered * 2
			}
			process(filtered)
		}
	}
}

// Test constrained varying with reduce
func testConstrainedVaryingReduceSSA() {
	// EXPECT SSA: OpCall (to reduce with constrained varying)
	// EXPECT SSA: array decomposition for constrained types
	go for i := range[4] 20 {
		var constrainedData lanes.Varying[int32] = i * 10

		// Reduce constrained varying
		_ = reduce.Add(constrainedData)
		allPositive := reduce.All(constrainedData > 0)

		if allPositive {
			process(constrainedData)
		}
	}
}

// Test reduce with function calls
func testReduceWithFunctionCallsSSA() {
	// EXPECT SSA: OpCall (to SPMD function and then reduce)
	var data lanes.Varying[int32] = 15

	// Call SPMD function and reduce result
	processed := processData(data)
	_ = reduce.Add(processed)
}

func processData(input lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: function generates varying result for reduction
	return input*2 + lanes.Index()
}

// Test nested reduce operations
func testNestedReduceSSA() {
	// EXPECT SSA: multiple OpCall (to different reduce functions)
	go for i := range 32 {
		data := i + 5

		// Multiple reduce operations
		var sum int32 = reduce.Add(data)
		var max int32 = reduce.Max(data)

		// Use reduce results in computation
		var finalResult int32
		if reduce.Any(data > 0) {
			finalResult = sum + max
		} else {
			finalResult = 0
		}

		_ = finalResult
	}
}

// Test reduce with complex expressions
func testComplexReduceExpressionSSA() {
	// EXPECT SSA: OpCall (to reduce with complex varying expression)
	var a lanes.Varying[int32] = 10
	var b lanes.Varying[int32] = 20
	var c lanes.Varying[float32] = 2.5

	// Complex expression in reduce
	var complexResult int32 = reduce.Add((a + b) * lanes.Varying[int32](c))
	var complexCondition bool = reduce.All((a > b) || (lanes.Varying[int32](c) > a))

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
		data[i] = i * i
	}

	go for i := range 4 {
		// Load varying data and reduce
		loadedData := data[i*2]
		_ = reduce.Add(loadedData)
	}
}

// Helper function
func process(x lanes.Varying[int]) {
	_ = x
}
