//go:build goexperiment.spmd

// Test SSA generation for uniform-to-varying broadcasts
package spmdtest

import "lanes"

// Test automatic uniform-to-varying broadcast in arithmetic
func testAutomaticBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast for automatic conversion)
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted uniform)
	var uniformVal uniform int32 = 42
	var varyingVal varying int32 = 10
	
	// Uniform should be automatically broadcasted
	var result varying int32 = uniformVal + varyingVal
	var result2 varying int32 = varyingVal * uniformVal
	var result3 varying int32 = uniformVal - varyingVal
	
	process(int(result + result2 + result3))
}

// Test explicit lanes.Broadcast calls
func testExplicitBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorMul (for arithmetic with explicitly broadcasted value)
	var uniformVal uniform int32 = 100
	
	// Explicit broadcast should generate lanes.Broadcast call
	var broadcasted varying int32 = lanes.Broadcast(uniformVal, 0)
	var result varying int32 = broadcasted * varying int32(2)
	
	process(int(result))
}

// Test broadcast with different lane targets
func testLaneBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast with lane parameter)
	// EXPECT SSA: OpVectorExtractLane (for extracting specific lane)
	var varyingVal varying int32 = varying int32(lanes.Index()) + varying int32(10)
	
	// Broadcast from different lanes
	var broadcast0 varying int32 = lanes.Broadcast(varyingVal, 0) // Broadcast from lane 0
	var broadcast1 varying int32 = lanes.Broadcast(varyingVal, 1) // Broadcast from lane 1
	var broadcast2 varying int32 = lanes.Broadcast(varyingVal, 2) // Broadcast from lane 2
	
	var result varying int32 = broadcast0 + broadcast1 + broadcast2
	process(int(result))
}

// Test broadcast in conditional contexts
func testConditionalBroadcastSSA() {
	// EXPECT SSA: OpSelect (for conditional broadcast)
	// EXPECT SSA: OpCall (to lanes.Broadcast within condition)
	var uniformCondition uniform bool = true
	var uniformValue uniform int32 = 50
	var varyingValue varying int32 = 25
	
	var result varying int32
	if uniformCondition {
		// Uniform broadcast in conditional branch
		result = uniformValue + varyingValue
	} else {
		result = varyingValue * varying int32(2)
	}
	
	process(int(result))
}

// Test broadcast with function parameters
func testParameterBroadcastSSA() {
	// EXPECT SSA: OpCall (with automatic broadcast for uniform parameters)
	var uniformVal uniform int32 = 75
	var varyingVal varying int32 = 30
	
	// Call SPMD function with mixed parameters
	result := mixedParameterFunction(uniformVal, varyingVal)
	process(int(result))
}

func mixedParameterFunction(uniform uniform int32, varying varying int32) varying int32 {
	// EXPECT SSA: uniform parameter automatically broadcasted
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted uniform)
	return uniform + varying * varying int32(2)
}

// Test broadcast in go for loop context
func testGoForBroadcastSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast within loop)
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted values)
	var uniformBase uniform int32 = 100
	
	go for i := range 8 {
		// Uniform should be broadcasted within loop
		var loopResult varying int32 = uniformBase + varying int32(i)
		process(int(loopResult))
	}
}

// Test broadcast with type conversions
func testBroadcastTypeConversionSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorCvt (for type conversion of broadcasted value)
	var uniformInt uniform int32 = 42
	var varyingFloat varying float32 = 3.14
	
	// Type conversion with broadcast
	var converted varying float32 = varying float32(uniformInt) + varyingFloat
	var result varying int32 = varying int32(converted)
	
	process(int(result))
}

// Test broadcast with memory operations
func testBroadcastMemoryOpsSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorStore (for storing broadcasted values)
	var data [16]int32
	var uniformValue uniform int32 = 999
	
	go for i := range 4 {
		// Store broadcasted uniform to varying indices
		data[i] = int32(uniformValue + varying int32(i))
	}
	
	process(int(data[0]))
}

// Test broadcast with reduction operations
func testBroadcastWithReduceSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast and reduce.Add)
	var uniformVal uniform int32 = 10
	var varyingVal varying int32 = varying int32(lanes.Index()) * varying int32(5)
	
	// Combine broadcast with reduction
	var combined varying int32 = uniformVal + varyingVal
	var sum uniform int32 = reduce.Add(combined)
	
	process(int(sum))
}

// Test nested broadcast operations
func testNestedBroadcastSSA() {
	// EXPECT SSA: multiple OpCall (to lanes.Broadcast)
	// EXPECT SSA: OpVectorAdd (for nested arithmetic)
	var uniform1 uniform int32 = 20
	var uniform2 uniform int32 = 30
	var varyingVal varying int32 = 5
	
	// Nested operations with multiple broadcasts
	var result varying int32 = (uniform1 + uniform2) * varyingVal + uniform1
	process(int(result))
}

// Test broadcast in complex expressions
func testComplexBroadcastExpressionSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast for multiple uniforms)
	// EXPECT SSA: complex vector arithmetic tree
	var a uniform int32 = 5
	var b uniform float32 = 2.5
	var c varying int32 = 10
	var d varying float32 = 1.5
	
	// Complex expression requiring multiple broadcasts
	var result varying float32 = (varying float32(a) * b + varying float32(c)) / (d + varying float32(a))
	process(int(result))
}

// Helper function
func process(x int) {
	_ = x
}