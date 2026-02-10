//go:build goexperiment.spmd

// Test SSA generation for uniform-to-varying broadcasts
package spmdtest

import (
	"lanes"
	"reduce"
)

// Test automatic uniform-to-varying broadcast in arithmetic
func testAutomaticBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast for automatic conversion)
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted uniform)
	var uniformVal int32 = 42
	var varyingVal lanes.Varying[int32] = 10

	// Uniform should be automatically broadcasted
	var result lanes.Varying[int32] = uniformVal + varyingVal
	var result2 lanes.Varying[int32] = varyingVal * uniformVal
	var result3 lanes.Varying[int32] = uniformVal - varyingVal

	process(result + result2 + result3)
}

// Test explicit lanes.Broadcast calls
func testExplicitBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorMul (for arithmetic with explicitly broadcasted value)
	var varyingVal lanes.Varying[int32] = 100

	// Explicit broadcast should generate lanes.Broadcast call
	var broadcasted lanes.Varying[int32] = lanes.Broadcast(varyingVal, 0)
	var result lanes.Varying[int32] = broadcasted * lanes.Varying[int32](2)

	process(result)
}

// Test broadcast with different lane targets
func testLaneBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast with lane parameter)
	// EXPECT SSA: OpVectorExtractLane (for extracting specific lane)
	var varyingVal lanes.Varying[int32] = lanes.Varying[int32](lanes.Index()) + lanes.Varying[int32](10)

	// Broadcast from different lanes
	var broadcast0 lanes.Varying[int32] = lanes.Broadcast(varyingVal, 0) // Broadcast from lane 0
	var broadcast1 lanes.Varying[int32] = lanes.Broadcast(varyingVal, 1) // Broadcast from lane 1
	var broadcast2 lanes.Varying[int32] = lanes.Broadcast(varyingVal, 2) // Broadcast from lane 2

	var result lanes.Varying[int32] = broadcast0 + broadcast1 + broadcast2
	process(result)
}

// Test broadcast in conditional contexts
func testConditionalBroadcastSSA() {
	// EXPECT SSA: OpSelect (for conditional broadcast)
	// EXPECT SSA: OpCall (to lanes.Broadcast within condition)
	var uniformCondition bool = true
	var uniformValue int32 = 50
	var varyingValue lanes.Varying[int32] = 25

	var result lanes.Varying[int32]
	if uniformCondition {
		// Uniform broadcast in conditional branch
		result = uniformValue + varyingValue
	} else {
		result = varyingValue * lanes.Varying[int32](2)
	}

	process(result)
}

// Test broadcast with function parameters
func testParameterBroadcastSSA() {
	// EXPECT SSA: OpCall (with automatic broadcast for uniform parameters)
	var uniformVal int32 = 75
	var varyingVal lanes.Varying[int32] = 30

	// Call SPMD function with mixed parameters
	result := mixedParameterFunction(uniformVal, varyingVal)
	process(result)
}

func mixedParameterFunction(u int32, v lanes.Varying[int32]) lanes.Varying[int32] {
	// EXPECT SSA: uniform parameter automatically broadcasted
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted uniform)
	return u + v*lanes.Varying[int32](2)
}

// Test broadcast in go for loop context
func testGoForBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast within loop)
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted values)
	var uniformBase int32 = 100

	go for i := range 8 {
		// Uniform should be broadcasted within loop
		var loopResult lanes.Varying[int32] = uniformBase + lanes.Varying[int32](i)
		process(loopResult)
	}
}

// Test broadcast with type conversions
func testBroadcastTypeConversionSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorCvt (for type conversion of broadcasted value)
	var uniformInt int32 = 42
	var varyingFloat lanes.Varying[float32] = 3.14

	// Type conversion with broadcast
	var converted lanes.Varying[float32] = lanes.Varying[float32](uniformInt) + varyingFloat
	var result lanes.Varying[int32] = lanes.Varying[int32](converted)

	process(result)
}

// Test broadcast with memory operations
func testBroadcastMemoryOpsSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorStore (for storing broadcasted values)
	var uniformValue int32 = 999

	go for i := range 16 {
		// Store broadcasted uniform to varying indices
		var data lanes.Varying[int32] = uniformValue + lanes.Varying[int32](i)
		process(data)
	}
}

// Test broadcast with reduction operations
func testBroadcastWithReduceSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast and reduce.Add)
	var uniformVal int32 = 10
	var varyingVal lanes.Varying[int32] = lanes.Varying[int32](lanes.Index()) * 5

	// Combine broadcast with reduction
	var combined lanes.Varying[int32] = uniformVal + varyingVal
	var sum int32 = reduce.Add(combined)

	process(sum)
}

// Test nested broadcast operations
func testNestedBroadcastSSA() {
	// EXPECT SSA: multiple OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorAdd (for nested arithmetic)
	var uniform1 int32 = 20
	var uniform2 int32 = 30
	var varyingVal lanes.Varying[int32] = 5

	// Nested operations with multiple broadcasts
	var result lanes.Varying[int32] = (uniform1 + uniform2) * varyingVal + uniform1
	process(result)
}

// Test broadcast in complex expressions
func testComplexBroadcastExpressionSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast for multiple uniforms)
	// EXPECT SSA: complex vector arithmetic tree
	var a int32 = 5
	var b float32 = 2.5
	var c lanes.Varying[int32] = 10
	var d lanes.Varying[float32] = 1.5

	// Complex expression requiring multiple broadcasts
	var result lanes.Varying[float32] = (lanes.Varying[float32](a)*b + lanes.Varying[float32](c)) / (d + lanes.Varying[float32](a))
	process(result)
}

// Helper function
func process(x lanes.Varying[int]) {
	_ = x
}
